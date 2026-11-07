package markdownblocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

const testMarkdownQuoteLayoutShards = 12

type quoteLayoutProducts struct {
	goBinary, goLayout, sanitized, release string
	backend                                []byte
	source                                 string
	inputs                                 buildcache.Inputs
}

func quoteLayoutShard(name string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int(h.Sum64() % testMarkdownQuoteLayoutShards)
}

var quoteLayoutOnce sync.Once
var quoteLayoutShared quoteLayoutProducts

func quoteLayoutSetup(test *testing.T) quoteLayoutProducts {
	quoteLayoutOnce.Do(func() {
		started := time.Now()
		// Shared setup has no deadline; Loom bounds the whole unit.
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		quoteLayoutShared = quoteLayoutBuild(test, root)
		test.Logf("TestMarkdownQuoteLayout (setup): %.3fs", time.Since(started).Seconds())
	})
	if quoteLayoutShared.goBinary == "" {
		test.Fatal("quote layout setup failed")
	}
	return quoteLayoutShared
}

var quoteLayoutNativeOnce sync.Once
var quoteLayoutNativeShared quoteLayoutProducts

// Shared products are also prepared by independently selected shards.
func TestMarkdownQuoteLayoutNative(test *testing.T) {
	test.Parallel()
	quoteLayoutNativeSetup(test, quoteLayoutSetup(test))
}

func quoteLayoutNativeSetup(test *testing.T, products quoteLayoutProducts) quoteLayoutProducts {
	quoteLayoutNativeOnce.Do(func() {
		started := time.Now()
		// Shared setup has no deadline; Loom bounds the whole unit.
		buildProducts := products
		var builds sync.WaitGroup
		for _, sanitize := range []bool{true, false} {
			builds.Add(1)
			go func() {
				defer builds.Done()
				dir := quoteLayoutNativeProduct(test, buildProducts, sanitize)
				if sanitize {
					products.sanitized = filepath.Join(dir, "port")
				} else {
					products.release = filepath.Join(dir, "port")
				}
			}()
		}

		builds.Wait()
		quoteLayoutNativeShared = products
		test.Logf("TestMarkdownQuoteLayoutNative (setup): %.3fs", time.Since(started).Seconds())
	})
	if quoteLayoutNativeShared.sanitized == "" || quoteLayoutNativeShared.release == "" {
		test.Fatal("quote native setup failed")
	}
	return quoteLayoutNativeShared
}

var quoteLayoutReadyOnce sync.Once
var quoteLayoutReadyShared quoteLayoutProducts

// Setup exercises the same preparation used by independently selected shards.
func TestMarkdownQuoteLayout_Setup(test *testing.T) {
	test.Parallel()
	quoteLayoutReady(test)
}

func quoteLayoutReady(test *testing.T) quoteLayoutProducts {
	test.Helper()
	quoteLayoutReadyOnce.Do(func() {
		started := time.Now()
		// Shared setup has no deadline; Loom bounds the whole unit.
		quoteLayoutReadyShared = quoteLayoutNativeSetup(test, quoteLayoutSetup(test))
		test.Logf("TestMarkdownQuoteLayout_Setup: %.3fs", time.Since(started).Seconds())
	})
	if quoteLayoutReadyShared.goBinary == "" ||
		quoteLayoutReadyShared.goLayout == "" ||
		quoteLayoutReadyShared.sanitized == "" ||
		quoteLayoutReadyShared.release == "" {
		test.Fatal("quote shared setup failed")
	}
	return quoteLayoutReadyShared
}

func quoteLayoutEnumerate(test *testing.T) ([]auditInput, [testMarkdownQuoteLayoutShards][]int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := blockCorpus(test, root, "whitespace")
	var shards [testMarkdownQuoteLayoutShards][]int
	for index, input := range inputs {
		shard := quoteLayoutShard(input.Name)
		shards[shard] = append(shards[shard], index)
	}
	return inputs, shards
}

func TestMarkdownQuoteLayoutUnion(test *testing.T) {
	test.Parallel()
	inputs, shards := quoteLayoutEnumerate(test)
	seen := make([]int, len(inputs))
	planted, plantedShard, total := 0, -1, 0
	if len(inputs) == 0 {
		test.Fatal("empty quote corpus")
	}
	for shard, indices := range shards {
		for _, index := range indices {
			seen[index]++
			total++
			expected := []byte(inputs[index].Text)
			actual := append([]byte(nil), expected...)
			if index == 0 {
				actual = append(actual, 0)
			}
			if !quoteLayoutBytesEqual(actual, expected) {
				planted++
				plantedShard = shard
			}
		}
	}
	for index, count := range seen {
		if count != 1 {
			test.Fatalf("case %q covered %d times", inputs[index].Name, count)
		}
	}
	if len(shards) != testMarkdownQuoteLayoutShards || total != len(inputs) {
		test.Fatal("quote shard enumeration mismatch")
	}
	if planted != 1 {
		test.Fatalf("planted disagreement caught %d times", planted)
	}
	test.Logf(
		"union %d cases, each exactly once; planted disagreement %q caught only by shard-%03d",
		total,
		inputs[0].Name,
		plantedShard,
	)
}

func quoteLayoutBytesEqual(actual, expected []byte) bool { return bytes.Equal(actual, expected) }

func quoteLayoutRunShard(test *testing.T, shard int) {
	test.Helper()
	// Each shard prepares shared products itself before its own clock starts.
	products := quoteLayoutReady(test)
	deadline := time.AfterFunc(90*time.Second, func() { panic("cooked: " + test.Name() + " exceeded 90s") })
	defer deadline.Stop()
	inputs, shards := quoteLayoutEnumerate(test)
	cases := make([]auditInput, 0, len(shards[shard]))
	for _, index := range shards[shard] {
		cases = append(cases, inputs[index])
	}
	fixture := quoteLayoutFixture(test, cases, products)
	quoteLayoutMutants(test, fixture)
	test.Logf("shard-%03d: %d cases", shard, len(cases))
}

func TestMarkdownQuoteLayout_000(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 0) }
func TestMarkdownQuoteLayout_001(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 1) }
func TestMarkdownQuoteLayout_002(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 2) }
func TestMarkdownQuoteLayout_003(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 3) }
func TestMarkdownQuoteLayout_004(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 4) }
func TestMarkdownQuoteLayout_005(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 5) }
func TestMarkdownQuoteLayout_006(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 6) }
func TestMarkdownQuoteLayout_007(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 7) }
func TestMarkdownQuoteLayout_008(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 8) }
func TestMarkdownQuoteLayout_009(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 9) }
func TestMarkdownQuoteLayout_010(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 10) }
func TestMarkdownQuoteLayout_011(test *testing.T) { test.Parallel(); quoteLayoutRunShard(test, 11) }

func quoteLayoutBuild(test *testing.T, root string) quoteLayoutProducts {
	return quoteLayoutBuildProduct(test, root, "")
}

func quoteLayoutBuildProduct(test *testing.T, root, target string) quoteLayoutProducts {
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs := buildcache.Inputs{
		Name: "markdown-quote-layout-lowered",
		// What lowering reads (whitespaceLayoutBuildFiles, which TestLoweringInputsCoverCompilerDependencies keeps
		// complete); the native builds keep the same Files.
		Files:     whitespaceLayoutBuildFiles(),
		Flags:     []string{"C and JavaScript"},
		Toolchain: []string{runtime.Version()},
	}
	loweredDir := buildcache.Product(test, inputs, func(dir string) error {
		program, loweredResultError := loweredResult(main)
		if loweredResultError != nil {
			return loweredResultError
		}
		if writeError := os.WriteFile(
			filepath.Join(dir, "program.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	source, readError := os.ReadFile(filepath.Join(loweredDir, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	backend, readError := os.ReadFile(filepath.Join(loweredDir, "program.mjs"))
	if readError != nil {
		test.Fatal(readError)
	}
	products := quoteLayoutProducts{backend: backend, source: string(source), inputs: inputs}
	if target == "lowered" {
		return products
	}
	goProducts := oracles.QuoteLayoutGoProducts(test, root, target)
	products.goBinary, products.goLayout = goProducts.List, goProducts.Document
	return products
}

func quoteLayoutMutants(test *testing.T, fixture *layoutFixture) {
	main, nativeCases, inputs := fixture.main, fixture.nativeCases, fixture.inputs
	want := run{stdout: fixture.want}
	mutations := []struct{ name, from, to string }{
		{"quote marker", "arena.text('> ')", "arena.text('>> ')"},
		{"quote alignment", "arena.align('> ',", "arena.align('',"},
		{"quote blank line", "!overlapping && !definitions", "!overlapping && definitions"},
	}
	for _, mutation := range mutations {
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
				content = []byte(
					strings.Replace(string(content), "../markdowninline/inline.ts", "./markdowninline/inline.ts", 1),
				)
			}
			if name == "quotes.ts" {
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
		result := quoteLayoutOnNode(test, mutantMain, nativeCases)
		clean(test, "source Node layout mutant", result)
		if quoteLayoutBytesEqual(result.stdout, want.stdout) {
			test.Fatalf("quote mutant %s survived shard", mutation.name)
		}
		offset := firstDifference(string(result.stdout), string(want.stdout))
		index := bytes.Count(want.stdout[:offset], []byte("\n"))
		test.Logf(
			"output-only source Node mutant %s caught by %q at byte %d",
			mutation.name,
			inputs[index].Name,
			offset,
		)
	}
}

func quoteLayoutFixture(test *testing.T, inputs []auditInput, products quoteLayoutProducts) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	// This last cumulative corpus contains every earlier layout milestone's inputs.

	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	dir := test.TempDir()
	if createDirectoryError := os.MkdirAll(dir, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if encodeError := encoder.Encode(input); encodeError != nil {
			test.Fatal(encodeError)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(test, cases, batch.Bytes())
	goBinary := products.goBinary
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	listTask := startFixtureTask(&workers, func() (run, error) {
		return quoteLayoutExecuteResult(test, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.goLayout
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(root, "cohere/internal/format/prettier/bundles")
	}
	script, absolutePathError := filepath.Abs("testdata/document_library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	markdownScript, absolutePathError := filepath.Abs("testdata/library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	libraryTask := startFixtureTask(&workers, func() (run, error) {
		return quoteLayoutExecuteResult(test, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	want := listTask.await(test)
	clean(test, "Go list fixtures", want)
	// Width slots are retained only for the canonical Go/fork doc protocol.
	// Poison every text-width field in the native protocol to prove no oracle service remains.
	protocolPaths := []string{nativeCases}
	for _, protocolPath := range protocolPaths {
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
		return quoteLayoutExecuteResult(test, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(
		&workers,
		func() (run, error) { return quoteLayoutOnNodeResult(test, main, nativeCases) },
	)
	backendPath := filepath.Join(dir, "program.mjs")
	write(test, backendPath, products.backend)
	backendTask := startFixtureTask(
		&workers,
		func() (run, error) { return quoteLayoutOnNodeResult(test, backendPath, nativeCases) },
	)
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.sanitized
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return quoteLayoutExecuteResult(test, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.release
		return quoteLayoutExecuteResult(test, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(
		&workers,
		func() (run, error) { return quoteLayoutExecuteResult(test, nil, "node", script, fork, canonicalCases) },
	)
	goResult := docTask.await(test)
	clean(test, "Go document layout", goResult)
	equal(test, "Go document layout", goResult.stdout, want.stdout)
	answer := nativeTask.await(test)
	binary := products.sanitized
	for _, side := range []struct {
		name   string
		result run
	}{
		{"source Node lists", sourceTask.await(test)},
		{"backend lists", backendTask.await(test)},
		{"native lists", answer},
	} {
		clean(test, side.name, side.result)
		if !quoteLayoutBytesEqual(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			test.Fatalf(
				"%s output byte %d in %s\ngot %q\nwant %q",
				side.name,
				offset,
				inputs[index].Name,
				actual[index],
				expected[index],
			)
		}
	}
	leakResult := quoteLayoutExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, nativeCases)
	clean(test, "native leak check", leakResult)
	original := originalTask.await(test)
	clean(test, "original document printer", original)
	equal(test, "original document printer", original.stdout, want.stdout)
	sourceAnswers := auditResults(test, "original Markdown parser/layout", libraryTask.await(test))
	if len(sourceAnswers) != len(inputs) {
		test.Fatal("original Markdown oracle lost a document")
	}
	encodeOutput := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	wantLines := strings.Split(strings.TrimSuffix(string(want.stdout), "\n"), "\n")
	for index, answer := range sourceAnswers {
		if answer.Name != inputs[index].Name || encodeOutput.Replace(answer.Off) != wantLines[index] {
			test.Fatalf(
				"original Markdown parser/layout disagreed with Go in %s: got %q, Go encoded %q",
				inputs[index].Name,
				answer.Off,
				wantLines[index],
			)
		}
	}
	test.Logf("original fork full Markdown parsing/layout off agrees on all %d source documents", len(inputs))
	release := releaseTask.await(test)
	clean(test, "release block layout", release)
	equal(test, "release block layout", release.stdout, want.stdout)
	// Repeat timings only on request; all corpus and mutant comparisons ran above.
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := products.release
		runner, absolutePathError := filepath.Abs(filepath.Join(root, "oracle/node.mjs"))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{
			{"Go document layout", goLayout, []string{canonicalCases}},
			{"native block/layout component", fast, []string{nativeCases}},
			{
				"source Node block/layout component",
				"node",
				[]string{"--disable-warning=ExperimentalWarning", runner, main, nativeCases},
			},
			{"original Node document layout", "node", []string{script, fork, canonicalCases}},
		} {
			var elapsed time.Duration
			for round := 0; round < 3; round++ {
				start := time.Now()
				result := quoteLayoutExecute(test, nil, side.command, side.args...)
				elapsed += time.Since(start)
				clean(test, side.name, result)
				equal(test, side.name, result.stdout, want.stdout)
			}
			test.Logf(
				"component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded",
				side.name,
				float64(len(inputs)*3)/elapsed.Seconds(),
				elapsed.Seconds(),
			)
		}
	}
	return &layoutFixture{
		mutantNativeCases: nativeCases, mutantInputs: inputs, mutantWant: want.stdout,
		root: root, directory: dir, nativeCases: nativeCases, canonicalCases: canonicalCases,
		main: main, fork: fork, goLayout: goLayout, script: script,
		inputs: inputs, want: want.stdout,
	}
}

// quoteLayoutCommand gives children their own deadline and process group so
// cancellation reaches the compilers spawned by Go as well as the Go process. It returns the command's context with
// it, for childguard.Capture to end the reads by.
func quoteLayoutCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(test.Context(), 90*time.Second)
	test.Cleanup(cancel)
	return quoteLayoutContextCommand(timeoutContext, name, arguments...), timeoutContext
}

// Setup keeps process-group cancellation without a deadline of its own.
func quoteLayoutSetupCommand(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	test.Cleanup(cancel)
	return quoteLayoutContextCommand(ctx, name, arguments...)
}

func quoteLayoutContextCommand(callerContext context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

// quoteLayoutExecuteResult reads a run's output through childguard.Capture, to its end with no timer deciding when:
// exec's own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a
// second after the child exited, so a run that had finished could fail its shard under load (#n3gk22h's shape).
func quoteLayoutExecuteResult(test *testing.T, environment []string, name string, arguments ...string) (run, error) {
	test.Helper()
	command, ctx := quoteLayoutCommand(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		return run{}, fmt.Errorf("running %s: %w", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, nil
}

func quoteLayoutExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	result, quoteLayoutExecuteResultError := quoteLayoutExecuteResult(test, environment, name, arguments...)
	if quoteLayoutExecuteResultError != nil {
		test.Fatal(quoteLayoutExecuteResultError)
	}
	return result
}

func quoteLayoutOnNodeResult(test *testing.T, path string, arguments ...string) (run, error) {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return quoteLayoutExecuteResult(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

func quoteLayoutOnNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	result, quoteLayoutOnNodeResultError := quoteLayoutOnNodeResult(test, path, arguments...)
	if quoteLayoutOnNodeResultError != nil {
		test.Fatal(quoteLayoutOnNodeResultError)
	}
	return result
}

func quoteLayoutNativeProduct(test *testing.T, products quoteLayoutProducts, sanitize bool) string {
	options := native.Options{Sanitize: sanitize}
	nativeInputs := products.inputs
	nativeInputs.Name = fmt.Sprintf("markdown-quote-layout-native-%t", sanitize)
	nativeInputs.Flags = append(append([]string{}, products.inputs.Flags...), native.Flags(options)...)
	nativeInputs.Flags = append(nativeInputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	nativeInputs.Toolchain = append([]string{runtime.Version()}, buildcache.Tool("clang", "--version"))
	dir := buildcache.Product(
		test,
		nativeInputs,
		func(dir string) error { return native.Build(products.source, filepath.Join(dir, "port"), options) },
	)
	return dir
}
