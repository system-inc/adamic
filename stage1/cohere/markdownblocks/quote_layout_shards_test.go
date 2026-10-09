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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
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

func quoteLayoutSetup(t *testing.T) quoteLayoutProducts {
	quoteLayoutOnce.Do(func() {
		started := time.Now()
		deadline := time.AfterFunc(75*time.Second, func() { panic("cooked: TestMarkdownQuoteLayout (setup) exceeded 75s") })
		defer deadline.Stop()
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		quoteLayoutShared = quoteLayoutBuild(t, root)
		t.Logf("TestMarkdownQuoteLayout (setup): %.3fs", time.Since(started).Seconds())
	})
	if quoteLayoutShared.goBinary == "" {
		t.Fatal("quote layout setup failed")
	}
	return quoteLayoutShared
}

var quoteLayoutNativeOnce sync.Once
var quoteLayoutNativeShared quoteLayoutProducts

// Not parallel: sanitized and release build products initialized before parallel shards.
func TestMarkdownQuoteLayoutNative(t *testing.T) {
	quoteLayoutNativeSetup(t, quoteLayoutSetup(t))
}

func quoteLayoutNativeSetup(t *testing.T, products quoteLayoutProducts) quoteLayoutProducts {
	quoteLayoutNativeOnce.Do(func() {
		started := time.Now()
		deadline := time.AfterFunc(75*time.Second, func() { panic("cooked: TestMarkdownQuoteLayoutNative (setup) exceeded 75s") })
		defer deadline.Stop()
		inputs, source := products.inputs, products.source
		var builds sync.WaitGroup
		for _, sanitize := range []bool{true, false} {
			builds.Add(1)
			go func() {
				defer builds.Done()
				options := native.Options{Sanitize: sanitize}
				nativeInputs := inputs
				nativeInputs.Name = fmt.Sprintf("markdown-quote-layout-native-%t", sanitize)
				nativeInputs.Flags = append(append([]string{}, inputs.Flags...), native.Flags(options)...)
				nativeInputs.Flags = append(nativeInputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
				nativeInputs.Toolchain = append([]string{runtime.Version()}, buildcache.Tool("clang", "--version"))
				dir := buildcache.Product(t, nativeInputs, func(dir string) error { return native.Build(string(source), filepath.Join(dir, "port"), options) })
				if sanitize {
					products.sanitized = filepath.Join(dir, "port")
				} else {
					products.release = filepath.Join(dir, "port")
				}
			}()
		}

		builds.Wait()
		quoteLayoutNativeShared = products
		t.Logf("TestMarkdownQuoteLayoutNative (setup): %.3fs", time.Since(started).Seconds())
	})
	if quoteLayoutNativeShared.sanitized == "" || quoteLayoutNativeShared.release == "" {
		t.Fatal("quote native setup failed")
	}
	return quoteLayoutNativeShared
}

func quoteLayoutEnumerate(t *testing.T) ([]auditInput, [testMarkdownQuoteLayoutShards][]int) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := blockCorpus(t, root, "whitespace")
	var shards [testMarkdownQuoteLayoutShards][]int
	for index, input := range inputs {
		shard := quoteLayoutShard(input.Name)
		shards[shard] = append(shards[shard], index)
	}
	return inputs, shards
}

func TestMarkdownQuoteLayoutUnion(t *testing.T) {
	t.Parallel()
	inputs, shards := quoteLayoutEnumerate(t)
	seen := make([]int, len(inputs))
	planted, plantedShard, total := 0, -1, 0
	if len(inputs) == 0 {
		t.Fatal("empty quote corpus")
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
			t.Fatalf("case %q covered %d times", inputs[index].Name, count)
		}
	}
	if len(shards) != testMarkdownQuoteLayoutShards || total != len(inputs) {
		t.Fatal("quote shard enumeration mismatch")
	}
	if planted != 1 {
		t.Fatalf("planted disagreement caught %d times", planted)
	}
	t.Logf("union %d cases, each exactly once; planted disagreement %q caught only by shard-%03d", total, inputs[0].Name, plantedShard)
}

func quoteLayoutBytesEqual(actual, expected []byte) bool { return bytes.Equal(actual, expected) }

func quoteLayoutRunShard(t *testing.T, shard int) {
	t.Helper()
	deadline := time.AfterFunc(75*time.Second, func() { panic("cooked: " + t.Name() + " exceeded 75s") })
	defer deadline.Stop()
	products := quoteLayoutNativeSetup(t, quoteLayoutSetup(t))
	inputs, shards := quoteLayoutEnumerate(t)
	cases := make([]auditInput, 0, len(shards[shard]))
	for _, index := range shards[shard] {
		cases = append(cases, inputs[index])
	}
	fixture := quoteLayoutFixture(t, cases, products)
	quoteLayoutMutants(t, fixture)
	t.Logf("shard-%03d: %d cases", shard, len(cases))
}

func TestMarkdownQuoteLayout_000(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 0) }
func TestMarkdownQuoteLayout_001(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 1) }
func TestMarkdownQuoteLayout_002(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 2) }
func TestMarkdownQuoteLayout_003(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 3) }
func TestMarkdownQuoteLayout_004(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 4) }
func TestMarkdownQuoteLayout_005(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 5) }
func TestMarkdownQuoteLayout_006(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 6) }
func TestMarkdownQuoteLayout_007(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 7) }
func TestMarkdownQuoteLayout_008(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 8) }
func TestMarkdownQuoteLayout_009(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 9) }
func TestMarkdownQuoteLayout_010(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 10) }
func TestMarkdownQuoteLayout_011(t *testing.T) { t.Parallel(); quoteLayoutRunShard(t, 11) }

func quoteLayoutBuild(t *testing.T, root string) quoteLayoutProducts {
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:      "markdown-quote-layout-lowered",
		Files:     []string{"stage1/cohere/markdownblocks", "stage1/cohere/markdowninline", "internal/load", "internal/lower", "internal/ir", "internal/native", "internal/javascript", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod"},
		Flags:     []string{"C and JavaScript", os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")},
		Toolchain: []string{runtime.Version()},
	}
	loweredDir := buildcache.Product(t, inputs, func(dir string) error {
		program, err := loweredResult(main)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	source, err := os.ReadFile(filepath.Join(loweredDir, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	backend, err := os.ReadFile(filepath.Join(loweredDir, "program.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	products := quoteLayoutProducts{backend: backend, source: string(source), inputs: inputs}
	// GoBuild is not on this base; preserve the original Go overlay builds.
	dir := filepath.Join(artifactDirectory, "quote-products")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	for _, mode := range []struct{ name, driver, command string }{
		{"lists", "list_go.go", "adamic_markdown_lists"},
		{"layout", "document_go.go", "adamic_markdown_doclayout"},
	} {
		driver, err := filepath.Abs(filepath.Join("testdata", mode.driver))
		if err != nil {
			t.Fatal(err)
		}
		mainPath := filepath.Join(cohere, "cmd", mode.command, "main.go")
		replace := map[string]string{mainPath: driver}
		if mode.name == "lists" {
			bridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			t.Fatal(err)
		}
		overlayPath := filepath.Join(dir, mode.name+".json")
		write(t, overlayPath, overlay)
		binary := filepath.Join(dir, mode.name)
		command := quoteLayoutCommand(t, "go", "build", "-overlay="+overlayPath, "-o", binary, mainPath)
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			t.Fatalf("Go %s: %v\n%s", mode.name, err, output)
		}
		if mode.name == "lists" {
			products.goBinary = binary
		} else {
			products.goLayout = binary
		}
	}
	return products
}

func quoteLayoutMutants(t *testing.T, fixture *layoutFixture) {
	main, nativeCases, inputs := fixture.main, fixture.nativeCases, fixture.inputs
	want := run{stdout: fixture.want}
	mutations := []struct{ name, from, to string }{
		{"quote marker", "arena.text('> ')", "arena.text('>> ')"},
		{"quote alignment", "arena.align('> ',", "arena.align('',"},
		{"quote blank line", "!overlapping && !definitions", "!overlapping && definitions"},
	}
	for _, mutation := range mutations {
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
			if name == "quotes.ts" {
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
		result := quoteLayoutOnNode(t, mutantMain, nativeCases)
		clean(t, "source Node layout mutant", result)
		if quoteLayoutBytesEqual(result.stdout, want.stdout) {
			t.Fatalf("quote mutant %s survived shard", mutation.name)
		}
		offset := firstDifference(string(result.stdout), string(want.stdout))
		index := bytes.Count(want.stdout[:offset], []byte("\n"))
		t.Logf("output-only source Node mutant %s caught by %q at byte %d", mutation.name, inputs[index].Name, offset)
	}
}

func quoteLayoutFixture(t *testing.T, inputs []auditInput, products quoteLayoutProducts) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	// This last cumulative corpus contains every earlier layout milestone's inputs.

	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if err := encoder.Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	goBinary := products.goBinary
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	listTask := startFixtureTask(&workers, func() (run, error) {
		return quoteLayoutExecuteResult(t, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.goLayout
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(root, "cohere/internal/format/prettier/bundles")
	}
	script, err := filepath.Abs("testdata/document_library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	markdownScript, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	libraryTask := startFixtureTask(&workers, func() (run, error) {
		return quoteLayoutExecuteResult(t, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	want := listTask.await(t)
	clean(t, "Go list fixtures", want)
	// Width slots are retained only for the canonical Go/fork doc protocol.
	// Poison every text-width field in the native protocol to prove no oracle service remains.
	protocolPaths := []string{nativeCases}
	for _, protocolPath := range protocolPaths {
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
		return quoteLayoutExecuteResult(t, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) { return quoteLayoutOnNodeResult(t, main, nativeCases) })
	backendPath := filepath.Join(dir, "program.mjs")
	write(t, backendPath, products.backend)
	backendTask := startFixtureTask(&workers, func() (run, error) { return quoteLayoutOnNodeResult(t, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.sanitized
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return quoteLayoutExecuteResult(t, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.release
		return quoteLayoutExecuteResult(t, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) { return quoteLayoutExecuteResult(t, nil, "node", script, fork, canonicalCases) })
	goResult := docTask.await(t)
	clean(t, "Go document layout", goResult)
	equal(t, "Go document layout", goResult.stdout, want.stdout)
	answer := nativeTask.await(t)
	binary := products.sanitized
	for _, side := range []struct {
		name   string
		result run
	}{{"source Node lists", sourceTask.await(t)}, {"backend lists", backendTask.await(t)}, {"native lists", answer}} {
		clean(t, side.name, side.result)
		if !quoteLayoutBytesEqual(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			t.Fatalf("%s output byte %d in %s\ngot %q\nwant %q", side.name, offset, inputs[index].Name, actual[index], expected[index])
		}
	}
	leakResult := quoteLayoutExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, nativeCases)
	clean(t, "native leak check", leakResult)
	original := originalTask.await(t)
	clean(t, "original document printer", original)
	equal(t, "original document printer", original.stdout, want.stdout)
	sourceAnswers := auditResults(t, "original Markdown parser/layout", libraryTask.await(t))
	if len(sourceAnswers) != len(inputs) {
		t.Fatal("original Markdown oracle lost a document")
	}
	encodeOutput := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	wantLines := strings.Split(strings.TrimSuffix(string(want.stdout), "\n"), "\n")
	for index, answer := range sourceAnswers {
		if answer.Name != inputs[index].Name || encodeOutput.Replace(answer.Off) != wantLines[index] {
			t.Fatalf("original Markdown parser/layout disagreed with Go in %s: got %q, Go encoded %q", inputs[index].Name, answer.Off, wantLines[index])
		}
	}
	t.Logf("original fork full Markdown parsing/layout off agrees on all %d source documents", len(inputs))
	release := releaseTask.await(t)
	clean(t, "release block layout", release)
	equal(t, "release block layout", release.stdout, want.stdout)
	// Repeat timings only on request; all corpus and mutant comparisons ran above.
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := products.release
		runner, err := filepath.Abs(filepath.Join(root, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"Go document layout", goLayout, []string{canonicalCases}}, {"native block/layout component", fast, []string{nativeCases}}, {"source Node block/layout component", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, nativeCases}}, {"original Node document layout", "node", []string{script, fork, canonicalCases}}} {
			var elapsed time.Duration
			for round := 0; round < 3; round++ {
				start := time.Now()
				result := quoteLayoutExecute(t, nil, side.command, side.args...)
				elapsed += time.Since(start)
				clean(t, side.name, result)
				equal(t, side.name, result.stdout, want.stdout)
			}
			t.Logf("component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
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
// cancellation reaches the compilers spawned by Go as well as the Go process.
func quoteLayoutCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func quoteLayoutExecuteResult(t *testing.T, environment []string, name string, arguments ...string) (run, error) {
	t.Helper()
	command := quoteLayoutCommand(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, fmt.Errorf("running %s: %w", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
}

func quoteLayoutExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := quoteLayoutExecuteResult(t, environment, name, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func quoteLayoutOnNodeResult(t *testing.T, path string, arguments ...string) (run, error) {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		return run{}, err
	}
	return quoteLayoutExecuteResult(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

func quoteLayoutOnNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	result, err := quoteLayoutOnNodeResult(t, path, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
