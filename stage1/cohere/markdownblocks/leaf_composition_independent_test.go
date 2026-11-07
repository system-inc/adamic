package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

type leafCompositionAnswer struct{ Name, Off string }
type leafCompositionFixture struct {
	*layoutFixture
	outputs map[string][][]byte
	answers []leafCompositionAnswer
}

const testMarkdownLeafCompositionShards = 7
const leafCompositionCorpusShards = testMarkdownLeafCompositionShards - 3

func leafCompositionShard(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(sum[0]) % leafCompositionCorpusShards
}

func leafCompositionNative(test *testing.T, source string, sanitize bool) string {
	sum := sha256.Sum256([]byte(source))
	directory := buildcache.Product(test, buildcache.Inputs{
		Name:  fmt.Sprintf("markdown-leaf-native-%t", sanitize),
		Files: []string{"internal/native"},
		Flags: []string{
			fmt.Sprintf("source=%x sanitize=%t split=%s", sum, sanitize, os.Getenv("ADAMIC_NATIVE_SPLIT")),
		},
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		return native.Build(source, filepath.Join(directory, "port"), native.Options{Sanitize: sanitize})
	})
	return filepath.Join(directory, "port")
}

var leafCompositionPrepared *leafCompositionFixture
var leafCompositionBuilds leafCompositionProducts

func testMarkdownLeafShards(test *testing.T) { leafCompositionFixtureFor(test) }

// Preparation is process-wide and precedes every shard's work deadline.
var leafCompositionOnce sync.Once

func TestMarkdownLeafComposition_Setup(test *testing.T) {
	test.Parallel()
	leafCompositionFixtureFor(test)
}

func leafCompositionFixtureFor(test *testing.T) *leafCompositionFixture {
	test.Helper()
	configureMarkdownMemory(test)
	leafCompositionOnce.Do(func() {
		started := time.Now()
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		all, _ := blockCorpus(test, root, "whitespace")
		products := prepareLeafCompositionGo(test, root)
		prepareLeafCompositionLowered(test, &products)
		products.sanitized = leafCompositionNative(test, products.source, true)
		products.release = leafCompositionNative(test, products.source, false)
		fixture := &leafCompositionFixture{layoutFixture: &layoutFixture{inputs: all}}
		mutantDirectory, temporaryDirectoryError := os.MkdirTemp(artifactDirectory, "leaf-composition-mutants-")
		if temporaryDirectoryError != nil {
			test.Fatal(temporaryDirectoryError)
		}
		var batch bytes.Buffer
		for _, input := range all {
			if encodeError := json.NewEncoder(&batch).Encode(input); encodeError != nil {
				test.Fatal(encodeError)
			}
		}
		cases := filepath.Join(mutantDirectory, "cases.jsonl")
		write(test, cases, batch.Bytes())
		mutantCases := filepath.Join(mutantDirectory, "native.txt")
		want := leafCompositionExecute(
			test,
			nil,
			products.goList,
			cases,
			mutantCases,
			filepath.Join(mutantDirectory, "canonical.txt"),
		)
		clean(test, "full Go mutant fixtures", want)
		fixture.mutantInputs, fixture.mutantWant, fixture.mutantNativeCases, fixture.main = all, want.stdout, mutantCases, products.main
		fixture.want = want.stdout
		poisonLeafCompositionCases(test, mutantCases)
		leafCompositionPrepared, leafCompositionBuilds = fixture, products
		test.Logf("TestMarkdownLeafComposition (setup): %.3fs", time.Since(started).Seconds())
	})
	if leafCompositionPrepared == nil {
		test.Fatal("leaf composition preparation failed")
	}
	return leafCompositionPrepared
}

func leafCompositionDeadline(test *testing.T) *time.Timer {
	test.Helper()
	name := test.Name()
	return time.AfterFunc(90*time.Second, func() { panic("cooked: " + name + " exceeded 90s") })
}

func TestMarkdownLeafCompositionUnion(test *testing.T) {
	test.Parallel()
	testMarkdownLeafCompositionUnion(test)
}

func TestMarkdownLeafCompositionPlantedDisagreement(test *testing.T) {
	test.Parallel()
	testMarkdownLeafCompositionUnion(test)
}

func testMarkdownLeafCompositionUnion(test *testing.T) {
	configureMarkdownMemory(test)
	deadline := leafCompositionDeadline(test)
	defer deadline.Stop()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	all, _ := blockCorpus(test, root, "whitespace")
	seen := make([]int, len(all))
	caught := 0
	plantedShard := -1
	for shard := 0; shard < leafCompositionCorpusShards; shard++ {
		for index, input := range all {
			if leafCompositionShard(input.Name) != shard {
				continue
			}
			seen[index]++
			actual, expected := []byte(input.Text), []byte(input.Text)
			if index == 0 {
				actual = append(append([]byte(nil), actual...), '!')
			}
			if !bytes.Equal(actual, expected) {
				caught++
				plantedShard = shard
			}
		}
	}
	for index, count := range seen {
		if count != 1 {
			test.Fatalf("union case %d occurs %d times", index, count)
		}
	}
	if caught != 1 || plantedShard != leafCompositionShard(all[0].Name) {
		test.Fatalf("planted disagreement caught %d times", caught)
	}
	if len(leafCompositionTopLevelShards) != testMarkdownLeafCompositionShards {
		test.Fatal("shard enumeration mismatch")
	}
	test.Logf(
		"union=%d each exactly once; planted disagreement caught by TestMarkdownLeafComposition_%03d",
		len(all),
		plantedShard,
	)
}

func runLeafCompositionShard(test *testing.T, shard int) {
	configureMarkdownMemory(test)
	fixture := leafCompositionFixtureFor(test)
	started := time.Now()
	deadline := leafCompositionDeadline(test)
	defer deadline.Stop()
	defer func() { test.Logf("shard elapsed %.3fs", time.Since(started).Seconds()) }()
	if shard >= leafCompositionCorpusShards {
		testLeafCompositionMutation(test, fixture.layoutFixture, shard-leafCompositionCorpusShards)
		return
	}
	var inputs []auditInput
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	all, files := blockCorpus(test, root, "whitespace")
	physical := 0
	for index, input := range all {
		if leafCompositionShard(input.Name) == shard {
			inputs = append(inputs, input)
			if index < files {
				physical++
			}
		}
	}
	shardFixture := buildLeafCompositionFixture(test, inputs, physical, leafCompositionBuilds)
	expected := bytes.Split(bytes.TrimSuffix(shardFixture.want, []byte("\n")), []byte("\n"))
	for index, input := range inputs {
		for side, output := range shardFixture.outputs {
			if !bytes.Equal(expected[index], output[index]) {
				test.Fatalf("%s case %s differs: got %q want %q", side, input.Name, output[index], expected[index])
			}
		}
		answer := shardFixture.answers[index]
		encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
		if answer.Name != input.Name || encode.Replace(answer.Off) != string(expected[index]) {
			test.Fatalf("original Markdown parser/layout disagreed in %s", input.Name)
		}
	}
	test.Logf("cases=%d", len(inputs))
}
func TestMarkdownLeafComposition_000(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 0)
}
func TestMarkdownLeafComposition_001(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 1)
}
func TestMarkdownLeafComposition_002(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 2)
}
func TestMarkdownLeafComposition_003(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 3)
}
func TestMarkdownLeafComposition_004(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 4)
}
func TestMarkdownLeafComposition_005(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 5)
}
func TestMarkdownLeafComposition_006(test *testing.T) {
	test.Parallel()
	runLeafCompositionShard(test, 6)
}

var leafCompositionTopLevelShards = [...]func(*testing.T){
	TestMarkdownLeafComposition_000,
	TestMarkdownLeafComposition_001,
	TestMarkdownLeafComposition_002,
	TestMarkdownLeafComposition_003,
	TestMarkdownLeafComposition_004,
	TestMarkdownLeafComposition_005,
	TestMarkdownLeafComposition_006,
}

func testLeafCompositionMutation(test *testing.T, fixture *layoutFixture, shard int) {
	main := fixture.main
	mutantFile := "leaves.ts"
	mutations := []struct{ name, from, to string }{
		{
			"strong delimiter",
			"frame.kind === 'strong'\n                ? '**'",
			"frame.kind === 'strong'\n                ? '*'",
		},
		{
			"reference collapse",
			"frame.referenceType === 'collapsed'\n                  ? '[]'",
			"frame.referenceType === 'collapsed'\n                  ? ''",
		},
		{"footnote indentation", "arena.align('    ',", "arena.align('   ',"},
	}
	for index, mutation := range mutations {
		if index != shard {
			continue
		}
		test.Run(mutation.name, func(test *testing.T) {
			nativeCases, inputs := fixture.mutantNativeCases, fixture.mutantInputs
			want := run{stdout: fixture.mutantWant}
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if createDirectoryError := os.Mkdir(
				filepath.Join(scratch, "markdowninline"),
				0755,
			); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, name := range []string{"inline.ts", "classes.ts"} {
				content, readError := os.ReadFile(filepath.Join("../markdowninline", name))
				if readError != nil {
					test.Fatal(readError)
				}
				write(test, filepath.Join(scratch, "markdowninline", name), content)
			}
			for _, name := range []string{
				"document.ts",
				"commandStack.ts",
				"codec.ts",
				"lists.ts",
				"quotes.ts",
				"tables.ts",
				"codeblocks.ts",
				"htmlblocks.ts",
				"width.ts",
				"widthTables.ts",
				"widthRuneRanges.ts",
				"emojiMatcher.ts",
				"structure.ts",
				"root.ts",
				"leaves.ts",
				"children.ts",
				"splitText.ts",
				"textTokens.ts",
				"textClasses.ts",
				"preservedLabel.ts",
				"whitespace.ts",
				"whitespaceCodec.ts",
			} {
				content, readError := os.ReadFile(name)
				if readError != nil {
					test.Fatal(readError)
				}
				if name == "leaves.ts" {
					content = []byte(strings.Replace(
						string(content),
						"../markdowninline/inline.ts",
						"./markdowninline/inline.ts",
						1,
					))
				}
				if name == mutantFile {
					if strings.Count(string(content), mutation.from) != 1 {
						test.Fatal("mutation site count")
					}
					content = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
				}
				write(test, filepath.Join(scratch, name), content)
			}
			content, readError := os.ReadFile(main)
			if readError != nil {
				test.Fatal(readError)
			}
			content = []byte(
				strings.Replace(string(content), "../../markdowninline/inline.ts", "../markdowninline/inline.ts", 1),
			)
			mutantMain := filepath.Join(scratch, "testdata/list_probe.ts")
			write(test, mutantMain, content)
			result := leafCompositionNode(test, mutantMain, nativeCases)
			clean(test, "source Node layout mutant", result)

			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("list mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			test.Logf("output-only source Node mutant caught by %q at byte %d", inputs[index].Name, offset)
		})
	}
}

func buildLeafCompositionFixture(
	test *testing.T,
	inputs []auditInput,
	files int,
	products leafCompositionProducts,
) *leafCompositionFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	root := products.root
	fullInputs := inputs
	selection := gatesample.Selection{}
	dir := test.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if encodeError := encoder.Encode(input); encodeError != nil {
			test.Fatal(encodeError)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(test, cases, batch.Bytes())
	fullCases := cases
	if selection.Sample {
		fullCases = filepath.Join(dir, "mutant-cases.jsonl")
		var fullBatch bytes.Buffer
		fullEncoder := json.NewEncoder(&fullBatch)
		for _, input := range fullInputs {
			if encodeError := fullEncoder.Encode(input); encodeError != nil {
				test.Fatal(encodeError)
			}
		}
		write(test, fullCases, fullBatch.Bytes())
	}
	goBinary := products.goList
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {

		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var leafCompositionExecuteResultError error
			mutantWant, leafCompositionExecuteResultError = leafCompositionExecuteResult(
				test,
				nil,
				goBinary,
				fullCases,
				mutantNativeCases,
				filepath.Join(dir, "mutant-canonical.txt"),
			)
			if leafCompositionExecuteResultError != nil {
				return run{}, leafCompositionExecuteResultError
			}
		}
		return leafCompositionExecuteResult(test, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	main, fork, script, goLayout := products.main, products.fork, products.script, products.goLayout
	markdownScript, absolutePathError := filepath.Abs("testdata/library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	libraryTask := startFixtureTask(&workers, func() (run, error) {
		return leafCompositionExecuteResult(test, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	source := products.source
	sanitizedBuild := startFixtureTask(&workers, func() (string, error) { return products.sanitized, nil })
	releaseBuild := startFixtureTask(&workers, func() (string, error) { return products.release, nil })
	want := listTask.await(test)
	clean(test, "Go list fixtures", want)
	if selection.Sample {
		clean(test, "full Go mutant fixtures", mutantWant)
	} else {
		mutantWant = want
	}
	// Width slots are retained only for the canonical Go/fork doc protocol.
	// Poison every text-width field in the native protocol to prove no oracle service remains.
	protocolPaths := []string{nativeCases}
	if selection.Sample {
		protocolPaths = append(protocolPaths, mutantNativeCases)
	}
	for _, protocolPath := range protocolPaths {
		poisonLeafCompositionCases(test, protocolPath)
	}
	if keep := os.Getenv("ADAMIC_MARKDOWNLISTS_KEEP"); keep != "" {
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		write(test, filepath.Join(keep, "cases.jsonl"), batch.Bytes())
		write(test, filepath.Join(keep, "want.txt"), want.stdout)
		for _, name := range []string{"native.txt", "canonical.txt"} {
			value, readError := os.ReadFile(filepath.Join(dir, name))
			if readError != nil {
				test.Fatal(readError)
			}
			write(test, filepath.Join(keep, name), value)
		}
	}
	docTask := startFixtureTask(&workers, func() (run, error) {
		return leafCompositionExecuteResult(test, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(
		&workers,
		func() (run, error) { return leafCompositionNodeResult(test, main, nativeCases) },
	)
	backendPath := products.backend
	backendTask := startFixtureTask(
		&workers,
		func() (run, error) { return leafCompositionNodeResult(test, backendPath, nativeCases) },
	)
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary, resultError := sanitizedBuild.result()
		if resultError != nil {
			return run{}, resultError
		}
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return leafCompositionExecuteResult(test, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, resultError := releaseBuild.result()
		if resultError != nil {
			return run{}, resultError
		}
		return leafCompositionExecuteResult(test, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) {
		return leafCompositionExecuteResult(test, nil, "node", script, fork, canonicalCases)
	})
	outputs := map[string][][]byte{}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Go document layout", docTask.await(test)}, {"source Node lists", sourceTask.await(test)},
		{"backend lists", backendTask.await(test)}, {"native lists", nativeTask.await(test)},
		{"original document printer", originalTask.await(test)}, {"release block layout", releaseTask.await(test)},
	} {
		clean(test, side.name, side.result)
		// Enforce the stream terminator too; per-case byte comparisons run in shards.
		if len(side.result.stdout) == 0 || side.result.stdout[len(side.result.stdout)-1] != '\n' {
			test.Fatalf("%s missing terminator", side.name)
		}
		lines := bytes.Split(bytes.TrimSuffix(side.result.stdout, []byte("\n")), []byte("\n"))
		if len(lines) != len(inputs) {
			test.Fatalf("%s lost a document", side.name)
		}
		outputs[side.name] = lines
	}
	binary := sanitizedBuild.await(test)
	if report := leafCompositionLeaks(test, source, binary, nativeCases); report != "" {
		test.Fatal(report)
	}
	sourceAnswers := auditResults(test, "original Markdown parser/layout", libraryTask.await(test))
	if len(sourceAnswers) != len(inputs) {
		test.Fatal("original Markdown oracle lost a document")
	}
	answers := make([]leafCompositionAnswer, len(sourceAnswers))
	for index, answer := range sourceAnswers {
		answers[index] = leafCompositionAnswer{answer.Name, answer.Off}
	}

	return &leafCompositionFixture{outputs: outputs, answers: answers, layoutFixture: &layoutFixture{
		selection:         selection,
		mutantNativeCases: mutantNativeCases,
		mutantInputs:      fullInputs,
		mutantWant:        mutantWant.stdout,
		root:              root,
		directory:         dir,
		nativeCases:       nativeCases,
		canonicalCases:    canonicalCases,
		main:              main,
		fork:              fork,
		goLayout:          goLayout,
		script:            script,
		inputs:            inputs,
		files:             files,
		want:              want.stdout,
	}}
}

func leafCompositionLeaks(test *testing.T, source, sanitized string, arguments ...string) string {
	var report run
	switch runtime.GOOS {
	case "linux":
		report = leafCompositionExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
	case "darwin":
		binary := leafCompositionNative(test, source, false)
		report = leafCompositionExecute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
	}
	if report.exitCode == 0 {
		return ""
	}
	return fmt.Sprintf("exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
}

type leafCompositionProducts struct{ root, main, fork, script, goList, goLayout, source, backend, sanitized, release string }

func leafCompositionGoProduct(test *testing.T, root, name, driverName string) string {
	test.Helper()
	return oracles.LeafCompositionGoProduct(test, root, name, driverName)
}

func prepareLeafCompositionGo(test *testing.T, root string) leafCompositionProducts {
	products := leafCompositionProducts{root: root}
	products.goList = leafCompositionGoProduct(test, root, "adamic_markdown_lists", "list_go.go")
	products.goLayout = leafCompositionGoProduct(test, root, "adamic_markdown_doclayout", "document_go.go")
	var absolutePathError error
	cohere := filepath.Join(root, "cohere")
	products.main, absolutePathError = filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	products.script, absolutePathError = filepath.Abs("testdata/document_library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	products.fork = os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if products.fork == "" {
		products.fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	return products
}

func prepareLeafCompositionLowered(test *testing.T, products *leafCompositionProducts) {
	loweredDirectory := buildcache.Product(test, buildcache.Inputs{
		Name:      "markdown-leaf-lowered",
		Files:     leafCompositionLoweredFiles(test, products.root),
		Toolchain: []string{runtime.Version()},
	}, func(directory string) error {
		program, loweredResultError := loweredResult(products.main)
		if loweredResultError != nil {
			return loweredResultError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "program.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	data, readError := os.ReadFile(filepath.Join(loweredDirectory, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	products.source = string(data)
	products.backend = filepath.Join(loweredDirectory, "program.mjs")

}

func poisonLeafCompositionCases(test *testing.T, protocolPath string) {

	raw, readError := os.ReadFile(protocolPath)
	if readError != nil {
		test.Fatal(readError)
	}
	lines := strings.Split(string(raw), "\n")
	for index, line := range lines {
		fields := strings.Split(line, "\t")
		switch fields[0] {
		case "D":
			if fields[1] == "t" {
				fields[3] = "-777"
			}
		case "W":
			fields[3] = "-777"
		case "H":
			fields[2] = ""
		case "C":
			fields[4], fields[5], fields[6] = "-777", "-777", ""
		case "T":
			fields[2] = ""
			rows := strings.Split(fields[3], ":")
			for r, row := range rows {
				cells := strings.Split(row, ";")
				for c, cell := range cells {
					pair := strings.Split(cell, ",")
					if len(pair) == 2 {
						pair[1] = "-777"
						cells[c] = strings.Join(pair, ",")
					}
				}
				rows[r] = strings.Join(cells, ";")
			}
			fields[3] = strings.Join(rows, ":")
		}
		lines[index] = strings.Join(fields, "\t")
	}
	write(test, protocolPath, []byte(strings.Join(lines, "\n")))
}

// Preparation has no local deadline; the unit runner bounds the whole process. Output is read through
// childguard.Capture, to its end with no timer deciding when: exec's own copying under a one-second WaitDelay
// reported exec.ErrWaitDelay whenever Node's reads hadn't reached EOF a second after it exited, and that failed
// TestMarkdownRootLayout_005 under load in chg_7jt635t4's run (#n3gk22h's shape).
func leafCompositionExecuteResult(
	test *testing.T,
	environment []string,
	name string,
	arguments ...string,
) (run, error) {
	test.Helper()
	cancelContext, cancel := context.WithCancel(context.Background())
	test.Cleanup(cancel)
	command := exec.CommandContext(cancelContext, name, arguments...)
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(cancelContext, command)
	if captureError != nil {
		if _, ok := captureError.(*exec.ExitError); !ok {
			return run{}, fmt.Errorf("running %s: %v", name, captureError)
		}
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, nil
}
func leafCompositionExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	result, leafCompositionExecuteResultError := leafCompositionExecuteResult(test, environment, name, arguments...)
	if leafCompositionExecuteResultError != nil {
		test.Fatal(leafCompositionExecuteResultError)
	}
	return result
}
func leafCompositionNodeResult(test *testing.T, path string, arguments ...string) (run, error) {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return leafCompositionExecuteResult(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}
func leafCompositionNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	result, leafCompositionNodeResultError := leafCompositionNodeResult(test, path, arguments...)
	if leafCompositionNodeResultError != nil {
		test.Fatal(leafCompositionNodeResultError)
	}
	return result
}

// The compiler product depends on TypeScript and compiler inputs, not on test wrappers.
func leafCompositionLoweredFiles(test *testing.T, root string) []string {
	test.Helper()
	files := []string{
		"stage1/cohere/markdownblocks/testdata/list_probe.ts",
		"internal",
		"cohere/TypeScript/tsc",
		"cohere/TypeScript-shim",
		"cohere/go.mod",
		"cohere/go.sum",
		"go.mod",
		"go.work",
	}
	for _, directory := range []string{"stage1/cohere/markdownblocks", "stage1/cohere/markdowninline"} {
		matches, globError := filepath.Glob(filepath.Join(root, directory, "*.ts"))
		if globError != nil {
			test.Fatal(globError)
		}
		for _, path := range matches {
			relative, relativePathError := filepath.Rel(root, path)
			if relativePathError != nil {
				test.Fatal(relativePathError)
			}
			files = append(files, filepath.ToSlash(relative))
		}
	}
	return files
}
