package markdownblocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

const testMarkdownWhitespaceLayoutShards = 24

type whitespaceLayoutProducts struct {
	goBinary, goLayout, sanitized, release string
	backend                                []byte
	source                                 string
	inputs                                 buildcache.Inputs
}

func whitespaceLayoutShard(name string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int(h.Sum64() % testMarkdownWhitespaceLayoutShards)
}

var whitespaceLayoutOnce sync.Once
var whitespaceLayoutShared whitespaceLayoutProducts

func whitespaceLayoutSetup(t *testing.T, ctx context.Context) whitespaceLayoutProducts {
	whitespaceLayoutOnce.Do(func() {
		started := time.Now()
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		whitespaceLayoutShared = whitespaceLayoutBuild(t, ctx, root)
		t.Logf("TestMarkdownWhitespaceLayout (Go and lowered products): %.3fs", time.Since(started).Seconds())
	})
	if whitespaceLayoutShared.goBinary == "" {
		t.Fatal("whitespace layout setup failed")
	}
	return whitespaceLayoutShared
}

var whitespaceLayoutNativeOnce sync.Once
var whitespaceLayoutNativeShared whitespaceLayoutProducts

func whitespaceLayoutNativeSetup(t *testing.T, ctx context.Context, products whitespaceLayoutProducts) whitespaceLayoutProducts {
	whitespaceLayoutNativeOnce.Do(func() {
		started := time.Now()
		inputs, source := products.inputs, products.source
		var builds sync.WaitGroup
		for _, sanitize := range []bool{true, false} {
			builds.Add(1)
			go func() {
				defer builds.Done()
				options := native.Options{Sanitize: sanitize, Split: true, Jobs: 2}
				nativeInputs := inputs
				nativeInputs.Name = fmt.Sprintf("%s-native-%t", inputs.Name, sanitize)
				nativeInputs.Flags = append(append([]string{}, inputs.Flags...), native.Flags(options)...)
				nativeInputs.Flags = append(nativeInputs.Flags, "Split=true", "Jobs=2", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
				nativeInputs.Toolchain = append([]string{runtime.Version()}, buildcache.Tool("clang", "--version"))
				dir := buildcache.Product(t, nativeInputs, func(dir string) error {
					return whitespaceLayoutNativeBuild(ctx, source, filepath.Join(dir, "port"), options)
				})
				if sanitize {
					products.sanitized = filepath.Join(dir, "port")
				} else {
					products.release = filepath.Join(dir, "port")
				}
			}()
		}

		builds.Wait()
		whitespaceLayoutNativeShared = products
		t.Logf("TestMarkdownWhitespaceLayoutNative (setup): %.3fs", time.Since(started).Seconds())
	})
	if whitespaceLayoutNativeShared.sanitized == "" || whitespaceLayoutNativeShared.release == "" {
		t.Fatal("whitespace native setup failed")
	}
	return whitespaceLayoutNativeShared
}

func whitespaceLayoutEnumerate(t *testing.T) ([]auditInput, [testMarkdownWhitespaceLayoutShards][]int) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := blockCorpus(t, root, "whitespace")
	var shards [testMarkdownWhitespaceLayoutShards][]int
	for index, input := range inputs {
		shard := whitespaceLayoutShard(input.Name)
		shards[shard] = append(shards[shard], index)
	}
	return inputs, shards
}

// Not parallel: publishes shared build products before parallel shards are released.
func TestMarkdownWhitespaceLayout_Setup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	deadline := time.AfterFunc(90*time.Second, func() { panic("TestMarkdownWhitespaceLayout_Setup exceeded 90s") })
	defer deadline.Stop()
	defer close(whitespaceLayoutReady)
	directory := buildcache.Product(t, whitespaceLayoutSetupInputs(t), func(directory string) error {
		var workers sync.WaitGroup
		defer workers.Wait()
		policyTask := startFixtureTask(&workers, func() (whitespaceLayoutProducts, error) { return whitespaceLayoutPolicySetup(t, ctx), nil })
		products := whitespaceLayoutNativeSetup(t, ctx, whitespaceLayoutSetup(t, ctx))
		policy := policyTask.await(t)
		if policy.sanitized == "" {
			return fmt.Errorf("whitespace policy setup failed")
		}
		if err := os.WriteFile(filepath.Join(directory, "layout.mjs"), products.backend, 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "policy.mjs"), policy.backend, 0644); err != nil {
			return err
		}
		paths := map[string]string{"go": products.goBinary, "goLayout": products.goLayout, "sanitized": products.sanitized, "release": products.release, "policy": policy.sanitized}
		data, err := json.Marshal(paths)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "products.json"), data, 0644)
	})
	whitespaceLayoutPrepared, whitespaceLayoutPolicyPrepared = whitespaceLayoutReadProducts(t, directory)
}

func TestMarkdownWhitespaceLayoutUnion(t *testing.T) {
	t.Parallel()

	tree, err := parser.ParseFile(token.NewFileSet(), "whitespace_layout_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared := make([]bool, testMarkdownWhitespaceLayoutShards)
	for _, declaration := range tree.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "TestMarkdownWhitespaceLayout_Setup" {
			continue
		}
		if !ok || !strings.HasPrefix(function.Name.Name, "TestMarkdownWhitespaceLayout_") {
			continue
		}
		index, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestMarkdownWhitespaceLayout_"))
		if err != nil || index < 0 || index >= len(declared) || declared[index] {
			t.Fatal("shard function enumeration mismatch")
		}
		declared[index] = true
	}
	for index, present := range declared {
		if !present {
			t.Fatalf("missing shard function %03d", index)
		}
	}
	inputs, shards := whitespaceLayoutEnumerate(t)
	seen := make([]int, len(inputs))
	planted, plantedShard, total := 0, -1, 0
	if len(inputs) == 0 {
		t.Fatal("empty whitespace corpus")
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
			if !whitespaceLayoutBytesEqual(actual, expected) {
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
	if len(shards) != testMarkdownWhitespaceLayoutShards || total != len(inputs) {
		t.Fatal("whitespace shard enumeration mismatch")
	}
	if planted != 1 {
		t.Fatalf("planted disagreement caught %d times", planted)
	}
	t.Logf("union %d cases, each exactly once; planted disagreement %q caught only by shard-%03d", total, inputs[0].Name, plantedShard)
}

func whitespaceLayoutBytesEqual(actual, expected []byte) bool { return bytes.Equal(actual, expected) }

func whitespaceLayoutRunShard(t *testing.T, shard int) {
	products, policyProducts := whitespaceLayoutReadyProducts(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("own cases: %.3fs", time.Since(started).Seconds()) }()
	inputs, shards := whitespaceLayoutEnumerate(t)
	cases := make([]auditInput, 0, len(shards[shard]))
	for _, index := range shards[shard] {
		cases = append(cases, inputs[index])
	}
	fixture := whitespaceLayoutFixture(t, ctx, cases, products)
	whitespaceLayoutPolicy(t, ctx, fixture.nativeCases, fixture.fork, policyProducts, false)
	t.Logf("shard-%03d: %d cases", shard, len(cases))
}

func TestMarkdownWhitespaceLayout_000(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 0) }
func TestMarkdownWhitespaceLayout_001(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 1) }
func TestMarkdownWhitespaceLayout_002(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 2) }
func TestMarkdownWhitespaceLayout_003(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 3) }
func TestMarkdownWhitespaceLayout_004(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 4) }
func TestMarkdownWhitespaceLayout_005(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 5) }
func TestMarkdownWhitespaceLayout_006(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 6) }
func TestMarkdownWhitespaceLayout_007(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 7) }
func TestMarkdownWhitespaceLayout_008(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 8) }
func TestMarkdownWhitespaceLayout_009(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 9) }
func TestMarkdownWhitespaceLayout_010(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 10) }
func TestMarkdownWhitespaceLayout_011(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 11) }

func whitespaceLayoutBuild(t *testing.T, ctx context.Context, root string) whitespaceLayoutProducts {
	return whitespaceLayoutBuildProbe(t, ctx, root, "list_probe.ts", true)
}

func whitespaceLayoutBuildProbe(t *testing.T, ctx context.Context, root, probe string, goOracles bool) whitespaceLayoutProducts {
	main, err := filepath.Abs(filepath.Join("testdata", probe))
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:      "markdown-whitespace-layout-lowered-" + probe,
		Files:     whitespaceLayoutBuildFiles(),
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
	products := whitespaceLayoutProducts{backend: backend, source: string(source), inputs: inputs}
	if !goOracles {
		return products
	}
	// GoBuild is not on this base. Cache the unchanged Go overlay build commands as products.
	dir := filepath.Join(artifactDirectory, "whitespace-products")
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
		goInputs := whitespaceLayoutSetupInputs(t)
		goInputs.Name = "markdown-whitespace-layout-go-" + mode.name
		goInputs.Flags = append(goInputs.Flags, "go build", "overlay="+mode.driver, "CGO_ENABLED="+os.Getenv("CGO_ENABLED"), "GOFLAGS="+os.Getenv("GOFLAGS"))
		product := buildcache.Product(t, goInputs, func(product string) error {
			command := whitespaceLayoutCommand(ctx, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(product, "oracle"), mainPath)
			command.Dir = cohere
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode.name, err, output)
			}
			return nil
		})
		binary := filepath.Join(product, "oracle")
		if mode.name == "lists" {
			products.goBinary = binary
		} else {
			products.goLayout = binary
		}
	}
	return products
}

func whitespaceLayoutMutants(t *testing.T, ctx context.Context, fixture *layoutFixture) {
	main, nativeCases, inputs := fixture.main, fixture.nativeCases, fixture.inputs
	want := run{stdout: fixture.want}
	mutations := []struct{ name, from, to string }{
		{"preserved newline", "if(proseWrap === 'preserve' && frame.value === '\\n')", "if(proseWrap === 'never' && frame.value === '\\n')"},
		{"space removal", "return arena.text(canBeSpace ? ' ' : '')", "return arena.text('')"},
		{"ordered syntax", "value.endsWith('.')", "value.endsWith(':')"},
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
			if name == "whitespace.ts" {
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
		result := whitespaceLayoutNode(t, ctx, mutantMain, nativeCases)
		clean(t, "source Node layout mutant", result)
		if whitespaceLayoutBytesEqual(result.stdout, want.stdout) {
			t.Fatalf("whitespace mutant %s survived shard", mutation.name)
		}
		offset := firstDifference(string(result.stdout), string(want.stdout))
		index := bytes.Count(want.stdout[:offset], []byte("\n"))
		t.Logf("output-only source Node mutant %s caught by %q at byte %d", mutation.name, inputs[index].Name, offset)
	}
}

func whitespaceLayoutFixture(t *testing.T, ctx context.Context, inputs []auditInput, products whitespaceLayoutProducts) *layoutFixture {
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
		return whitespaceLayoutExecuteResult(ctx, nil, goBinary, cases, nativeCases, canonicalCases)
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
		return whitespaceLayoutExecuteResult(ctx, nil, "node", markdownScript, fork, cases, "fork", "off-only")
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
		return whitespaceLayoutExecuteResult(ctx, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) { return whitespaceLayoutNodeResult(ctx, main, nativeCases) })
	backendPath := filepath.Join(dir, "program.mjs")
	write(t, backendPath, products.backend)
	backendTask := startFixtureTask(&workers, func() (run, error) { return whitespaceLayoutNodeResult(ctx, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.sanitized
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return whitespaceLayoutExecuteResult(ctx, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.release
		return whitespaceLayoutExecuteResult(ctx, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) {
		return whitespaceLayoutExecuteResult(ctx, nil, "node", script, fork, canonicalCases)
	})
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
		if !whitespaceLayoutBytesEqual(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			t.Fatalf("%s output byte %d in %s\ngot %q\nwant %q", side.name, offset, inputs[index].Name, actual[index], expected[index])
		}
	}
	leakResult := whitespaceLayoutExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, nativeCases)
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
				result := whitespaceLayoutExecute(t, ctx, nil, side.command, side.args...)
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

func whitespaceLayoutPolicy(t *testing.T, ctx context.Context, cases, fork string, products whitespaceLayoutProducts, mutants bool) {
	t.Helper()
	raw, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	var expected strings.Builder
	count := 0
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "S\t") {
			continue
		}
		fields := strings.Split(line, "\t")
		expected.WriteString(fields[22] + "\n")
		count++
	}
	want := []byte(expected.String())
	main, err := filepath.Abs("testdata/whitespace_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary := products.sanitized
	answer := whitespaceLayoutExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, cases)
	backend := filepath.Join(t.TempDir(), "policy.mjs")
	write(t, backend, products.backend)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"whitespace actual original fork", whitespaceLayoutExecute(t, ctx, nil, "node", "testdata/whitespace_library.mjs", fork, cases)},
		{"whitespace native", answer}, {"whitespace source Node", whitespaceLayoutNode(t, ctx, main, cases)}, {"whitespace backend", whitespaceLayoutNode(t, ctx, backend, cases)},
	} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, want)
	}
	clean(t, "whitespace native leak check", whitespaceLayoutExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases))
	if !mutants {
		return
	}
	for _, m := range []struct{ name, from, to string }{
		{"policy CJ spacing tie", "return spaces > empties", "return spaces >= empties"},
		{"policy Korean pair", "return (previous === 'k-letter'", "return (previous === 'non-cjk'"},
		{"policy single line", "kind === 'tableCell' ||", "kind === 'missing-tableCell' ||"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"whitespace.ts", "whitespaceCodec.ts", "codec.ts", "document.ts", "commandStack.ts", "width.ts", "widthTables.ts", "widthRuneRanges.ts", "emojiMatcher.ts", "testdata/whitespace_probe.ts"} {
				b, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "whitespace.ts" {
					if strings.Count(string(b), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), b)
			}
			result := whitespaceLayoutNode(t, ctx, filepath.Join(scratch, "testdata/whitespace_probe.ts"), cases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want) {
				t.Fatal("survived")
			}
			t.Log("caught by Go policy document classification")
		})
	}
	t.Logf("%d whitespace nodes, six modes each, match Go and unchanged original private fork functions", count)
}

func whitespaceLayoutBuildFiles() []string {
	files := []string{"internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod"}
	for _, directory := range []string{".", "../markdowninline", "testdata"} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".ts") || strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go")) {
				files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/markdownblocks", directory, entry.Name())))
			}
		}
	}
	return files
}

var whitespaceLayoutPolicyOnce sync.Once
var whitespaceLayoutPolicyProducts whitespaceLayoutProducts

func whitespaceLayoutPolicySetup(t *testing.T, ctx context.Context) whitespaceLayoutProducts {
	whitespaceLayoutPolicyOnce.Do(func() {
		started := time.Now()
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		products := whitespaceLayoutBuildProbe(t, ctx, root, "whitespace_probe.ts", false)
		inputs := products.inputs
		inputs.Name += "-native-sanitized"
		inputs.Flags = append(append([]string{}, inputs.Flags...), native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 2})...)
		inputs.Flags = append(inputs.Flags, "Split=true", "Jobs=2", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		dir := buildcache.Product(t, inputs, func(dir string) error {
			return whitespaceLayoutNativeBuild(ctx, products.source, filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true, Jobs: 2})
		})
		products.sanitized = filepath.Join(dir, "port")
		whitespaceLayoutPolicyProducts = products
		t.Logf("TestMarkdownWhitespaceLayout (policy setup): %.3fs", time.Since(started).Seconds())
	})
	return whitespaceLayoutPolicyProducts
}
func TestMarkdownWhitespaceLayoutMutants(t *testing.T) {
	t.Parallel()
	products, policyProducts := whitespaceLayoutReadyProducts(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	inputs, _ := whitespaceLayoutEnumerate(t)
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
	nativeCases, canonical := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	want := whitespaceLayoutExecute(t, ctx, nil, products.goBinary, cases, nativeCases, canonical)
	clean(t, "Go mutant fixtures", want)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(root, "cohere/internal/format/prettier/bundles")
	}
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	fixture := &layoutFixture{main: main, nativeCases: nativeCases, inputs: inputs, want: want.stdout}
	whitespaceLayoutMutants(t, ctx, fixture)
	whitespaceLayoutPolicy(t, ctx, nativeCases, fork, policyProducts, true)
}

func TestMarkdownWhitespaceLayout_012(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 12) }
func TestMarkdownWhitespaceLayout_013(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 13) }
func TestMarkdownWhitespaceLayout_014(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 14) }
func TestMarkdownWhitespaceLayout_015(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 15) }
func TestMarkdownWhitespaceLayout_016(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 16) }
func TestMarkdownWhitespaceLayout_017(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 17) }
func TestMarkdownWhitespaceLayout_018(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 18) }
func TestMarkdownWhitespaceLayout_019(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 19) }
func TestMarkdownWhitespaceLayout_020(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 20) }
func TestMarkdownWhitespaceLayout_021(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 21) }
func TestMarkdownWhitespaceLayout_022(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 22) }
func TestMarkdownWhitespaceLayout_023(t *testing.T) { t.Parallel(); whitespaceLayoutRunShard(t, 23) }

// The setup test is the only builder. Filtered shard runs require its cached
// manifest; they fail on a miss instead of doing everyone's work inside a leaf.
var whitespaceLayoutReady = make(chan struct{})
var whitespaceLayoutPrepared, whitespaceLayoutPolicyPrepared whitespaceLayoutProducts

func whitespaceLayoutSetupInputs(t *testing.T) buildcache.Inputs {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	files := whitespaceLayoutBuildFiles()
	files = append(files, "stage1/cohere/markdownblocks/whitespace_layout_split_test.go", "cohere/internal", "cohere/go.mod", "cohere/go.sum", "go.work")
	return buildcache.Inputs{Name: "markdown-whitespace-layout-setup-v2", Files: files,
		Flags:     []string{root, os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK"), os.Getenv("ADAMIC_NATIVE_SPLIT"), os.Getenv("GOFLAGS"), os.Getenv("CGO_ENABLED"), "Split=true", "Jobs=2"},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}}
}

func whitespaceLayoutReadyProducts(t *testing.T) (whitespaceLayoutProducts, whitespaceLayoutProducts) {
	t.Helper()
	// When setup is selected, wait for its own top-level test to publish. With
	// -run selecting one shard, read a product made by a prior setup-only run.
	selection := flag.Lookup("test.run").Value.String()
	selected, err := regexp.MatchString(selection, "TestMarkdownWhitespaceLayout_Setup")
	if err != nil {
		t.Fatal(err)
	}
	if selected {
		<-whitespaceLayoutReady
		if whitespaceLayoutPrepared.goBinary == "" {
			t.Fatal("whitespace layout setup failed")
		}
		return whitespaceLayoutPrepared, whitespaceLayoutPolicyPrepared
	}
	directory := buildcache.Product(t, whitespaceLayoutSetupInputs(t), func(string) error {
		return fmt.Errorf("shared setup missing; run go test -run '^TestMarkdownWhitespaceLayout_Setup$' -timeout 90s first")
	})
	return whitespaceLayoutReadProducts(t, directory)
}

func whitespaceLayoutReadProducts(t *testing.T, directory string) (whitespaceLayoutProducts, whitespaceLayoutProducts) {
	data, err := os.ReadFile(filepath.Join(directory, "products.json"))
	if err != nil {
		t.Fatal(err)
	}
	var paths map[string]string
	if err := json.Unmarshal(data, &paths); err != nil {
		t.Fatal(err)
	}
	backend, err := os.ReadFile(filepath.Join(directory, "layout.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := os.ReadFile(filepath.Join(directory, "policy.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return whitespaceLayoutProducts{goBinary: paths["go"], goLayout: paths["goLayout"], sanitized: paths["sanitized"], release: paths["release"], backend: backend}, whitespaceLayoutProducts{sanitized: paths["policy"], backend: policy}
}

func whitespaceLayoutCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func whitespaceLayoutExecuteResult(ctx context.Context, environment []string, name string, args ...string) (run, error) {
	command := whitespaceLayoutCommand(ctx, name, args...)
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
func whitespaceLayoutExecute(t *testing.T, ctx context.Context, environment []string, name string, args ...string) run {
	t.Helper()
	result, err := whitespaceLayoutExecuteResult(ctx, environment, name, args...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func whitespaceLayoutNodeResult(ctx context.Context, path string, args ...string) (run, error) {
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		return run{}, err
	}
	return whitespaceLayoutExecuteResult(ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...)
}
func whitespaceLayoutNode(t *testing.T, ctx context.Context, path string, args ...string) run {
	t.Helper()
	result, err := whitespaceLayoutNodeResult(ctx, path, args...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// Run native.Build in a child test process so its compiler grandchildren share
// a process group that the setup context can kill without changing native code.
type whitespaceLayoutBuildRequest struct {
	Source, Output string
	Options        native.Options
}

func whitespaceLayoutNativeBuild(ctx context.Context, source, output string, options native.Options) error {
	sourcePath := output + ".c"
	if err := os.WriteFile(sourcePath, []byte(source), 0644); err != nil {
		return err
	}
	defer os.Remove(sourcePath)
	request, err := json.Marshal(whitespaceLayoutBuildRequest{Source: sourcePath, Output: output, Options: options})
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	command := whitespaceLayoutCommand(ctx, executable, "-test.run=^TestMarkdownWhitespaceLayoutBuildWorker$", "-test.timeout=90s")
	command.Env = append(os.Environ(), "ADAMIC_WHITESPACE_BUILD="+string(request))
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("native build: %w\n%s", err, output)
	}
	return nil
}
func TestMarkdownWhitespaceLayoutBuildWorker(t *testing.T) {
	t.Parallel()
	request := os.Getenv("ADAMIC_WHITESPACE_BUILD")
	if request == "" {
		return
	}
	var build whitespaceLayoutBuildRequest
	if err := json.Unmarshal([]byte(request), &build); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(build.Source)
	if err != nil {
		t.Fatal(err)
	}
	if err := native.Build(string(source), build.Output, build.Options); err != nil {
		t.Fatal(err)
	}
}
