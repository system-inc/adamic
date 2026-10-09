package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
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

func leafCompositionNative(t *testing.T, source string, sanitize bool) string {
	sum := sha256.Sum256([]byte(source))
	directory := buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprintf("markdown-leaf-native-%t", sanitize), Files: []string{"internal/native"}, Flags: []string{fmt.Sprintf("source=%x sanitize=%t split=%s", sum, sanitize, os.Getenv("ADAMIC_NATIVE_SPLIT"))}, Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")}}, func(directory string) error {
		return native.Build(source, filepath.Join(directory, "port"), native.Options{Sanitize: sanitize})
	})
	return filepath.Join(directory, "port")
}

var leafCompositionPrepared *leafCompositionFixture
var leafCompositionBuilds leafCompositionProducts

func testMarkdownLeafShards(t *testing.T) { leafCompositionFixtureFor(t) }

// Not parallel: publishes shared Go oracle products before the other setup phases and parallel shards.
func TestMarkdownLeafComposition_SetupGo(t *testing.T) {
	configureMarkdownMemory(t)
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	leafCompositionBuilds = prepareLeafCompositionGo(t, root)
}

// Not parallel: publishes lowered source consumed by native setup and every parallel shard.
func TestMarkdownLeafComposition_SetupLowered(t *testing.T) {
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	if leafCompositionBuilds.main == "" {
		t.Fatal("select all TestMarkdownLeafComposition_Setup tests; Go setup is not ready")
	}
	prepareLeafCompositionLowered(t, &leafCompositionBuilds)
}

// Not parallel: publishes the shared sanitized native product before fixture setup and parallel shards.
func TestMarkdownLeafComposition_SetupNativeSanitized(t *testing.T) {
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	if leafCompositionBuilds.source == "" {
		t.Fatal("lowered setup is not ready; select all TestMarkdownLeafComposition_Setup tests")
	}
	leafCompositionBuilds.sanitized = leafCompositionNative(t, leafCompositionBuilds.source, true)
}

// Not parallel: publishes the shared release native product before fixture setup and parallel shards.
func TestMarkdownLeafComposition_SetupNativeRelease(t *testing.T) {
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	if leafCompositionBuilds.source == "" {
		t.Fatal("lowered setup is not ready; select all TestMarkdownLeafComposition_Setup tests")
	}
	leafCompositionBuilds.release = leafCompositionNative(t, leafCompositionBuilds.source, false)
}

// Not parallel: publishes the shared leaf-composition fixture before parallel shards resume.
func TestMarkdownLeafComposition_Setup(t *testing.T) {
	configureMarkdownMemory(t)
	started := time.Now()
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	all, _ := blockCorpus(t, root, "whitespace")
	if leafCompositionBuilds.source == "" {
		t.Fatal("select all TestMarkdownLeafComposition_Setup tests; lowered setup is not ready")
	}
	products := leafCompositionBuilds
	if products.sanitized == "" || products.release == "" {
		t.Fatal("native setup is not ready; select all TestMarkdownLeafComposition_Setup tests")
	}
	fixture := &leafCompositionFixture{layoutFixture: &layoutFixture{inputs: all}}
	mutantDir, err := os.MkdirTemp(artifactDirectory, "leaf-composition-mutants-")
	if err != nil {
		t.Fatal(err)
	}
	var batch bytes.Buffer
	for _, input := range all {
		if err := json.NewEncoder(&batch).Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(mutantDir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	mutantCases := filepath.Join(mutantDir, "native.txt")
	want := leafCompositionExecute(t, nil, products.goList, cases, mutantCases, filepath.Join(mutantDir, "canonical.txt"))
	clean(t, "full Go mutant fixtures", want)
	fixture.mutantInputs, fixture.mutantWant, fixture.mutantNativeCases, fixture.main = all, want.stdout, mutantCases, products.main
	fixture.want = want.stdout
	poisonLeafCompositionCases(t, mutantCases)
	leafCompositionPrepared, leafCompositionBuilds = fixture, products
	t.Logf("TestMarkdownLeafComposition (setup): %.3fs", time.Since(started).Seconds())
}

// Selection must include _Setup when running shards: Go's serial phase finishes
// shared preparation before any t.Parallel shard resumes. Never build on a miss here.
func leafCompositionFixtureFor(t *testing.T) *leafCompositionFixture {
	t.Helper()
	if leafCompositionPrepared == nil {
		t.Fatal("leaf composition setup is not ready: select TestMarkdownLeafComposition_Setup together with the shard")
	}
	return leafCompositionPrepared
}

func leafCompositionDeadline(t *testing.T) *time.Timer {
	t.Helper()
	name := t.Name()
	return time.AfterFunc(90*time.Second, func() { panic("cooked: " + name + " exceeded 90s") })
}

func TestMarkdownLeafCompositionUnion(t *testing.T) {
	t.Parallel()
	configureMarkdownMemory(t)
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	all, _ := blockCorpus(t, root, "whitespace")
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
			t.Fatalf("union case %d occurs %d times", index, count)
		}
	}
	if caught != 1 || plantedShard != leafCompositionShard(all[0].Name) {
		t.Fatalf("planted disagreement caught %d times", caught)
	}
	if len(leafCompositionTopLevelShards) != testMarkdownLeafCompositionShards {
		t.Fatal("shard enumeration mismatch")
	}
	t.Logf("union=%d each exactly once; planted disagreement caught by TestMarkdownLeafComposition_%03d", len(all), plantedShard)
}

func runLeafCompositionShard(t *testing.T, shard int) {
	configureMarkdownMemory(t)
	fixture := leafCompositionFixtureFor(t)
	started := time.Now()
	deadline := leafCompositionDeadline(t)
	defer deadline.Stop()
	defer func() { t.Logf("shard elapsed %.3fs", time.Since(started).Seconds()) }()
	if shard >= leafCompositionCorpusShards {
		testLeafCompositionMutation(t, fixture.layoutFixture, shard-leafCompositionCorpusShards)
		return
	}
	var inputs []auditInput
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	all, files := blockCorpus(t, root, "whitespace")
	physical := 0
	for index, input := range all {
		if leafCompositionShard(input.Name) == shard {
			inputs = append(inputs, input)
			if index < files {
				physical++
			}
		}
	}
	shardFixture := buildLeafCompositionFixture(t, inputs, physical, leafCompositionBuilds)
	expected := bytes.Split(bytes.TrimSuffix(shardFixture.want, []byte("\n")), []byte("\n"))
	for index, input := range inputs {
		for side, output := range shardFixture.outputs {
			if !bytes.Equal(expected[index], output[index]) {
				t.Fatalf("%s case %s differs: got %q want %q", side, input.Name, output[index], expected[index])
			}
		}
		answer := shardFixture.answers[index]
		encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
		if answer.Name != input.Name || encode.Replace(answer.Off) != string(expected[index]) {
			t.Fatalf("original Markdown parser/layout disagreed in %s", input.Name)
		}
	}
	t.Logf("cases=%d", len(inputs))
}
func TestMarkdownLeafComposition_000(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 0) }
func TestMarkdownLeafComposition_001(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 1) }
func TestMarkdownLeafComposition_002(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 2) }
func TestMarkdownLeafComposition_003(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 3) }
func TestMarkdownLeafComposition_004(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 4) }
func TestMarkdownLeafComposition_005(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 5) }
func TestMarkdownLeafComposition_006(t *testing.T) { t.Parallel(); runLeafCompositionShard(t, 6) }

var leafCompositionTopLevelShards = [...]func(*testing.T){TestMarkdownLeafComposition_000, TestMarkdownLeafComposition_001, TestMarkdownLeafComposition_002, TestMarkdownLeafComposition_003, TestMarkdownLeafComposition_004, TestMarkdownLeafComposition_005, TestMarkdownLeafComposition_006}

func testLeafCompositionMutation(t *testing.T, fixture *layoutFixture, shard int) {
	main := fixture.main
	mutantFile := "leaves.ts"
	mutations := []struct{ name, from, to string }{
		{"strong delimiter", "frame.kind === 'strong'\n                ? '**'", "frame.kind === 'strong'\n                ? '*'"},
		{"reference collapse", "frame.referenceType === 'collapsed'\n                  ? '[]'", "frame.referenceType === 'collapsed'\n                  ? ''"},
		{"footnote indentation", "arena.align('    ',", "arena.align('   ',"},
	}
	for index, mutation := range mutations {
		if index != shard {
			continue
		}
		t.Run(mutation.name, func(t *testing.T) {
			nativeCases, inputs := fixture.mutantNativeCases, fixture.mutantInputs
			want := run{stdout: fixture.mutantWant}
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(scratch, "markdowninline"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"inline.ts", "classes.ts"} {
				content, err := os.ReadFile(filepath.Join("../markdowninline", name))
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(scratch, "markdowninline", name), content)
			}
			for _, name := range []string{"document.ts", "commandStack.ts", "codec.ts", "lists.ts", "quotes.ts", "tables.ts", "codeblocks.ts", "htmlblocks.ts", "width.ts", "widthTables.ts", "widthRuneRanges.ts", "emojiMatcher.ts", "structure.ts", "root.ts", "leaves.ts", "children.ts", "splitText.ts", "textTokens.ts", "textClasses.ts", "preservedLabel.ts", "whitespace.ts", "whitespaceCodec.ts"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == "leaves.ts" {
					content = []byte(strings.Replace(string(content), "../markdowninline/inline.ts", "./markdowninline/inline.ts", 1))
				}
				if name == mutantFile {
					if strings.Count(string(content), mutation.from) != 1 {
						t.Fatal("mutation site count")
					}
					content = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
				}
				write(t, filepath.Join(scratch, name), content)
			}
			content, err := os.ReadFile(main)
			if err != nil {
				t.Fatal(err)
			}
			content = []byte(strings.Replace(string(content), "../../markdowninline/inline.ts", "../markdowninline/inline.ts", 1))
			mutantMain := filepath.Join(scratch, "testdata/list_probe.ts")
			write(t, mutantMain, content)
			result := leafCompositionNode(t, mutantMain, nativeCases)
			clean(t, "source Node layout mutant", result)

			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("list mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Logf("output-only source Node mutant caught by %q at byte %d", inputs[index].Name, offset)
		})
	}
}

func buildLeafCompositionFixture(t *testing.T, inputs []auditInput, files int, products leafCompositionProducts) *leafCompositionFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	root := products.root
	fullInputs := inputs
	selection := gatesample.Selection{}
	dir := t.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if err := encoder.Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	fullCases := cases
	if selection.Sample {
		fullCases = filepath.Join(dir, "mutant-cases.jsonl")
		var fullBatch bytes.Buffer
		fullEncoder := json.NewEncoder(&fullBatch)
		for _, input := range fullInputs {
			if err := fullEncoder.Encode(input); err != nil {
				t.Fatal(err)
			}
		}
		write(t, fullCases, fullBatch.Bytes())
	}
	goBinary := products.goList
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {

		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var err error
			mutantWant, err = leafCompositionExecuteResult(t, nil, goBinary, fullCases, mutantNativeCases, filepath.Join(dir, "mutant-canonical.txt"))
			if err != nil {
				return run{}, err
			}
		}
		return leafCompositionExecuteResult(t, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	main, fork, script, goLayout := products.main, products.fork, products.script, products.goLayout
	markdownScript, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	libraryTask := startFixtureTask(&workers, func() (run, error) {
		return leafCompositionExecuteResult(t, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	source := products.source
	sanitizedBuild := startFixtureTask(&workers, func() (string, error) { return products.sanitized, nil })
	releaseBuild := startFixtureTask(&workers, func() (string, error) { return products.release, nil })
	want := listTask.await(t)
	clean(t, "Go list fixtures", want)
	if selection.Sample {
		clean(t, "full Go mutant fixtures", mutantWant)
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
		poisonLeafCompositionCases(t, protocolPath)
	}
	if keep := os.Getenv("ADAMIC_MARKDOWNLISTS_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "cases.jsonl"), batch.Bytes())
		write(t, filepath.Join(keep, "want.txt"), want.stdout)
		for _, name := range []string{"native.txt", "canonical.txt"} {
			value, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(keep, name), value)
		}
	}
	docTask := startFixtureTask(&workers, func() (run, error) {
		return leafCompositionExecuteResult(t, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) { return leafCompositionNodeResult(t, main, nativeCases) })
	backendPath := products.backend
	backendTask := startFixtureTask(&workers, func() (run, error) { return leafCompositionNodeResult(t, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := sanitizedBuild.result()
		if err != nil {
			return run{}, err
		}
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return leafCompositionExecuteResult(t, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := releaseBuild.result()
		if err != nil {
			return run{}, err
		}
		return leafCompositionExecuteResult(t, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) { return leafCompositionExecuteResult(t, nil, "node", script, fork, canonicalCases) })
	outputs := map[string][][]byte{}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Go document layout", docTask.await(t)}, {"source Node lists", sourceTask.await(t)},
		{"backend lists", backendTask.await(t)}, {"native lists", nativeTask.await(t)},
		{"original document printer", originalTask.await(t)}, {"release block layout", releaseTask.await(t)},
	} {
		clean(t, side.name, side.result)
		// Enforce the stream terminator too; per-case byte comparisons run in shards.
		if len(side.result.stdout) == 0 || side.result.stdout[len(side.result.stdout)-1] != '\n' {
			t.Fatalf("%s missing terminator", side.name)
		}
		lines := bytes.Split(bytes.TrimSuffix(side.result.stdout, []byte("\n")), []byte("\n"))
		if len(lines) != len(inputs) {
			t.Fatalf("%s lost a document", side.name)
		}
		outputs[side.name] = lines
	}
	binary := sanitizedBuild.await(t)
	if report := leafCompositionLeaks(t, source, binary, nativeCases); report != "" {
		t.Fatal(report)
	}
	sourceAnswers := auditResults(t, "original Markdown parser/layout", libraryTask.await(t))
	if len(sourceAnswers) != len(inputs) {
		t.Fatal("original Markdown oracle lost a document")
	}
	answers := make([]leafCompositionAnswer, len(sourceAnswers))
	for index, answer := range sourceAnswers {
		answers[index] = leafCompositionAnswer{answer.Name, answer.Off}
	}

	return &leafCompositionFixture{outputs: outputs, answers: answers, layoutFixture: &layoutFixture{
		selection: selection, mutantNativeCases: mutantNativeCases, mutantInputs: fullInputs, mutantWant: mutantWant.stdout,
		root: root, directory: dir, nativeCases: nativeCases, canonicalCases: canonicalCases,
		main: main, fork: fork, goLayout: goLayout, script: script,
		inputs: inputs, files: files, want: want.stdout,
	}}
}

func leafCompositionLeaks(t *testing.T, source, sanitized string, arguments ...string) string {
	var report run
	switch runtime.GOOS {
	case "linux":
		report = leafCompositionExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
	case "darwin":
		binary := leafCompositionNative(t, source, false)
		report = leafCompositionExecute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
	if report.exitCode == 0 {
		return ""
	}
	return fmt.Sprintf("exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
}

type leafCompositionProducts struct{ root, main, fork, script, goList, goLayout, source, backend, sanitized, release string }

func prepareLeafCompositionGo(t *testing.T, root string) leafCompositionProducts {
	p := leafCompositionProducts{root: root}
	cohere := filepath.Join(root, "cohere")
	bridge, err := filepath.Abs("testdata/list_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name, driver string
		target       *string
	}{
		{"adamic_markdown_lists", "list_go.go", &p.goList}, {"adamic_markdown_doclayout", "document_go.go", &p.goLayout},
	} {
		driver, err := filepath.Abs(filepath.Join("testdata", item.driver))
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(cohere, "cmd", item.name, "main.go")
		inputs := buildcache.Inputs{Name: "markdown-leaf-go-" + item.name, Files: []string{
			"stage1/cohere/markdownblocks/testdata/" + item.driver,
			"stage1/cohere/markdownblocks/testdata/list_bridge.go",
			"cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim",
			"cohere/go.mod", "cohere/go.sum", "cohere/go.work", "cohere/go.work.sum", "go.mod", "go.work",
		}, Flags: []string{"overlay=" + item.driver, "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH}}
		product := buildcache.Product(t, inputs, func(out string) error {
			replacements := map[string]string{main: driver}
			if item.name == "adamic_markdown_lists" {
				replacements[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
			}
			overlay, err := json.Marshal(map[string]any{"Replace": replacements})
			if err != nil {
				return err
			}
			overlayPath := filepath.Join(out, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			command := leafCompositionCommand(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(out, "oracle"), main)
			command.Dir = cohere
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("Go bridge: %v\n%s", err, output)
			}
			return nil
		})
		*item.target = filepath.Join(product, "oracle")
	}
	p.main, err = filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	p.script, err = filepath.Abs("testdata/document_library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	p.fork = os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if p.fork == "" {
		p.fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	return p
}

func prepareLeafCompositionLowered(t *testing.T, p *leafCompositionProducts) {
	loweredDirectory := buildcache.Product(t, buildcache.Inputs{Name: "markdown-leaf-lowered", Files: leafCompositionLoweredFiles(t, p.root), Toolchain: []string{runtime.Version()}}, func(directory string) error {
		program, err := loweredResult(p.main)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	data, err := os.ReadFile(filepath.Join(loweredDirectory, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	p.source = string(data)
	p.backend = filepath.Join(loweredDirectory, "program.mjs")

}

func poisonLeafCompositionCases(t *testing.T, protocolPath string) {

	raw, err := os.ReadFile(protocolPath)
	if err != nil {
		t.Fatal(err)
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
	write(t, protocolPath, []byte(strings.Join(lines, "\n")))
}

// Keep deadlines local to this test rather than changing the shared child helpers.
func leafCompositionCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = time.Second
	return command
}
func leafCompositionExecuteResult(t *testing.T, environment []string, name string, arguments ...string) (run, error) {
	t.Helper()
	command := leafCompositionCommand(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return run{}, fmt.Errorf("running %s: %v", name, err)
		}
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
}
func leafCompositionExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := leafCompositionExecuteResult(t, environment, name, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func leafCompositionNodeResult(t *testing.T, path string, arguments ...string) (run, error) {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		return run{}, err
	}
	return leafCompositionExecuteResult(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}
func leafCompositionNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	result, err := leafCompositionNodeResult(t, path, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// The compiler product depends on TypeScript and compiler inputs, not on test wrappers.
func leafCompositionLoweredFiles(t *testing.T, root string) []string {
	t.Helper()
	files := []string{"stage1/cohere/markdownblocks/testdata/list_probe.ts", "internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/go.mod", "cohere/go.sum", "go.mod", "go.work"}
	for _, directory := range []string{"stage1/cohere/markdownblocks", "stage1/cohere/markdowninline"} {
		matches, err := filepath.Glob(filepath.Join(root, directory, "*.ts"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range matches {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, filepath.ToSlash(relative))
		}
	}
	return files
}
