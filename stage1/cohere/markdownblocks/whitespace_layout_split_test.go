package markdownblocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
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

func whitespaceLayoutSetup(test *testing.T, ctx context.Context) whitespaceLayoutProducts {
	whitespaceLayoutOnce.Do(func() {
		started := time.Now()
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		whitespaceLayoutShared = whitespaceLayoutBuild(test, ctx, root)
		test.Logf("TestMarkdownWhitespaceLayout (Go and lowered products): %.3fs", time.Since(started).Seconds())
	})
	if whitespaceLayoutShared.goBinary == "" {
		test.Fatal("whitespace layout setup failed")
	}
	return whitespaceLayoutShared
}

var whitespaceLayoutNativeOnce sync.Once
var whitespaceLayoutNativeShared whitespaceLayoutProducts

func whitespaceLayoutNativeSetup(
	test *testing.T,
	ctx context.Context,
	products whitespaceLayoutProducts,
) whitespaceLayoutProducts {
	whitespaceLayoutNativeOnce.Do(func() {
		started := time.Now()
		buildProducts := products
		var builds sync.WaitGroup
		for _, sanitize := range []bool{true, false} {
			builds.Add(1)
			go func() {
				defer builds.Done()
				dir := whitespaceLayoutNativeProduct(test, ctx, buildProducts, sanitize)
				if sanitize {
					products.sanitized = filepath.Join(dir, "port")
				} else {
					products.release = filepath.Join(dir, "port")
				}
			}()
		}

		builds.Wait()
		whitespaceLayoutNativeShared = products
		test.Logf("TestMarkdownWhitespaceLayoutNative (setup): %.3fs", time.Since(started).Seconds())
	})
	if whitespaceLayoutNativeShared.sanitized == "" || whitespaceLayoutNativeShared.release == "" {
		test.Fatal("whitespace native setup failed")
	}
	return whitespaceLayoutNativeShared
}

func whitespaceLayoutEnumerate(test *testing.T) ([]auditInput, [testMarkdownWhitespaceLayoutShards][]int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := blockCorpus(test, root, "whitespace")
	var shards [testMarkdownWhitespaceLayoutShards][]int
	for index, input := range inputs {
		shard := whitespaceLayoutShard(input.Name)
		shards[shard] = append(shards[shard], index)
	}
	return inputs, shards
}

func TestMarkdownWhitespaceLayout_Setup(test *testing.T) {
	test.Parallel()
	whitespaceLayoutReadyProducts(test)
}

// Preparation has no deadline; Loom bounds the complete unit.
func whitespaceLayoutPrepare(test *testing.T) {
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	directory := buildcache.Product(test, whitespaceLayoutSetupInputs(test), func(directory string) error {
		var workers sync.WaitGroup
		defer workers.Wait()
		policyTask := startFixtureTask(
			&workers,
			func() (whitespaceLayoutProducts, error) { return whitespaceLayoutPolicySetup(test, ctx), nil },
		)
		products := whitespaceLayoutNativeSetup(test, ctx, whitespaceLayoutSetup(test, ctx))
		policy := policyTask.await(test)
		if policy.sanitized == "" {
			return fmt.Errorf("whitespace policy setup failed")
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "layout.mjs"),
			products.backend,
			0644,
		); writeError != nil {
			return writeError
		}
		if writeError := os.WriteFile(filepath.Join(directory, "policy.mjs"), policy.backend, 0644); writeError != nil {
			return writeError
		}
		// Each is another product, named by its key so a runner finds it in its own cache (#tqrqx60).
		paths := map[string]string{
			"go":        products.goBinary,
			"goLayout":  products.goLayout,
			"sanitized": products.sanitized,
			"release":   products.release,
			"policy":    policy.sanitized,
		}
		for name, path := range paths {
			paths[name] = buildcache.Relative(path)
		}
		data, encodeError := json.Marshal(paths)
		if encodeError != nil {
			return encodeError
		}
		return os.WriteFile(filepath.Join(directory, "products.json"), data, 0644)
	})
	whitespaceLayoutPrepared, whitespaceLayoutPolicyPrepared = whitespaceLayoutReadProducts(test, directory)
}

func TestMarkdownWhitespaceLayoutUnion(test *testing.T) {
	test.Parallel()

	tree, parseFileError := parser.ParseFile(token.NewFileSet(), "whitespace_layout_split_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
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
		index, parseError := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestMarkdownWhitespaceLayout_"))
		if parseError != nil || index < 0 || index >= len(declared) || declared[index] {
			test.Fatal("shard function enumeration mismatch")
		}
		declared[index] = true
	}
	for index, present := range declared {
		if !present {
			test.Fatalf("missing shard function %03d", index)
		}
	}
	inputs, shards := whitespaceLayoutEnumerate(test)
	seen := make([]int, len(inputs))
	planted, plantedShard, total := 0, -1, 0
	if len(inputs) == 0 {
		test.Fatal("empty whitespace corpus")
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
			test.Fatalf("case %q covered %d times", inputs[index].Name, count)
		}
	}
	if len(shards) != testMarkdownWhitespaceLayoutShards || total != len(inputs) {
		test.Fatal("whitespace shard enumeration mismatch")
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

func whitespaceLayoutBytesEqual(actual, expected []byte) bool { return bytes.Equal(actual, expected) }

func whitespaceLayoutRunShard(test *testing.T, shard int) {
	products, policyProducts := whitespaceLayoutReadyProducts(test)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { test.Logf("own cases: %.3fs", time.Since(started).Seconds()) }()
	inputs, shards := whitespaceLayoutEnumerate(test)
	cases := make([]auditInput, 0, len(shards[shard]))
	for _, index := range shards[shard] {
		cases = append(cases, inputs[index])
	}
	fixture := whitespaceLayoutFixture(test, ctx, cases, products)
	whitespaceLayoutPolicy(test, ctx, fixture.nativeCases, fixture.fork, policyProducts)
	test.Logf("shard-%03d: %d cases", shard, len(cases))
}

func TestMarkdownWhitespaceLayout_000(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 0)
}
func TestMarkdownWhitespaceLayout_001(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 1)
}
func TestMarkdownWhitespaceLayout_002(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 2)
}
func TestMarkdownWhitespaceLayout_003(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 3)
}
func TestMarkdownWhitespaceLayout_004(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 4)
}
func TestMarkdownWhitespaceLayout_005(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 5)
}
func TestMarkdownWhitespaceLayout_006(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 6)
}
func TestMarkdownWhitespaceLayout_007(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 7)
}
func TestMarkdownWhitespaceLayout_008(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 8)
}
func TestMarkdownWhitespaceLayout_009(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 9)
}
func TestMarkdownWhitespaceLayout_010(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 10)
}
func TestMarkdownWhitespaceLayout_011(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 11)
}

func whitespaceLayoutBuild(test *testing.T, ctx context.Context, root string) whitespaceLayoutProducts {
	return whitespaceLayoutBuildProbe(test, ctx, root, "list_probe.ts", true)
}

func whitespaceLayoutBuildProbe(
	test *testing.T,
	ctx context.Context,
	root, probe string,
	goOracles bool,
) whitespaceLayoutProducts {
	target := ""
	if !goOracles {
		target = "lowered"
	}
	return whitespaceLayoutBuildProduct(test, ctx, root, probe, target)
}

func whitespaceLayoutBuildProduct(
	test *testing.T,
	ctx context.Context,
	root, probe, target string,
) whitespaceLayoutProducts {
	main, absolutePathError := filepath.Abs(filepath.Join("testdata", probe))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs := buildcache.Inputs{
		Name:      "markdown-whitespace-layout-lowered-" + probe,
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
	products := whitespaceLayoutProducts{backend: backend, source: string(source), inputs: inputs}
	if target == "lowered" {
		return products
	}
	goProducts := oracles.WhitespaceLayoutGoProducts(test, ctx, root, target)
	products.goBinary, products.goLayout = goProducts.List, goProducts.Document
	return products
}

// whitespaceLayoutMutations are the output-only layout mutants of whitespace.ts, each run on source Node over the whole
// whitespace corpus.
var whitespaceLayoutMutations = []struct{ name, from, to string }{
	{
		"preserved newline",
		"if(proseWrap === 'preserve' && frame.value === '\\n')",
		"if(proseWrap === 'never' && frame.value === '\\n')",
	},
	{"space removal", "return arena.text(canBeSpace ? ' ' : '')", "return arena.text('')"},
	{"ordered syntax", "value.endsWith('.')", "value.endsWith(':')"},
}

func whitespaceLayoutMutant(
	test *testing.T,
	ctx context.Context,
	fixture *layoutFixture,
	mutation struct{ name, from, to string },
) {
	test.Helper()
	main, nativeCases, inputs := fixture.main, fixture.nativeCases, fixture.inputs
	want := run{stdout: fixture.want}
	scratch := test.TempDir()
	if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if createDirectoryError := os.Mkdir(filepath.Join(scratch, "markdowninline"), 0755); createDirectoryError != nil {
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
		if name == "whitespace.ts" {
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
	result := whitespaceLayoutNode(test, ctx, mutantMain, nativeCases)
	clean(test, "source Node layout mutant", result)
	if whitespaceLayoutMutantSurvivedError := whitespaceLayoutMutantSurvived(
		result.stdout,
		want.stdout,
	); whitespaceLayoutMutantSurvivedError != nil {
		test.Fatalf("whitespace mutant %s: %v", mutation.name, whitespaceLayoutMutantSurvivedError)
	}
	offset := firstDifference(string(result.stdout), string(want.stdout))
	index := bytes.Count(want.stdout[:offset], []byte("\n"))
	test.Logf("output-only source Node mutant %s caught by %q at byte %d", mutation.name, inputs[index].Name, offset)
}

func whitespaceLayoutFixture(
	test *testing.T,
	ctx context.Context,
	inputs []auditInput,
	products whitespaceLayoutProducts,
) *layoutFixture {
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
		return whitespaceLayoutExecuteResult(ctx, nil, goBinary, cases, nativeCases, canonicalCases)
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
		return whitespaceLayoutExecuteResult(ctx, nil, "node", markdownScript, fork, cases, "fork", "off-only")
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
		return whitespaceLayoutExecuteResult(ctx, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(
		&workers,
		func() (run, error) { return whitespaceLayoutNodeResult(ctx, main, nativeCases) },
	)
	backendPath := filepath.Join(dir, "program.mjs")
	write(test, backendPath, products.backend)
	backendTask := startFixtureTask(
		&workers,
		func() (run, error) { return whitespaceLayoutNodeResult(ctx, backendPath, nativeCases) },
	)
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
		if !whitespaceLayoutBytesEqual(side.result.stdout, want.stdout) {
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
	leakResult := whitespaceLayoutExecute(test, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, nativeCases)
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
				result := whitespaceLayoutExecute(test, ctx, nil, side.command, side.args...)
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

// whitespaceLayoutPolicyWant is Go's policy document classification, field 22 of every S row of cases, and the
// number of whitespace nodes it names.
func whitespaceLayoutPolicyWant(test *testing.T, cases string) ([]byte, int) {
	test.Helper()
	raw, readError := os.ReadFile(cases)
	if readError != nil {
		test.Fatal(readError)
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
	return []byte(expected.String()), count
}

func whitespaceLayoutPolicy(
	test *testing.T,
	ctx context.Context,
	cases, fork string,
	products whitespaceLayoutProducts,
) {
	test.Helper()
	want, count := whitespaceLayoutPolicyWant(test, cases)
	main, absolutePathError := filepath.Abs("testdata/whitespace_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	binary := products.sanitized
	answer := whitespaceLayoutExecute(test, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, cases)
	backend := filepath.Join(test.TempDir(), "policy.mjs")
	write(test, backend, products.backend)
	for _, side := range []struct {
		name   string
		result run
	}{
		{
			"whitespace actual original fork",
			whitespaceLayoutExecute(test, ctx, nil, "node", "testdata/whitespace_library.mjs", fork, cases),
		},
		{"whitespace native", answer},
		{"whitespace source Node", whitespaceLayoutNode(test, ctx, main, cases)},
		{"whitespace backend", whitespaceLayoutNode(test, ctx, backend, cases)},
	} {
		clean(test, side.name, side.result)
		equal(test, side.name, side.result.stdout, want)
	}
	clean(
		test,
		"whitespace native leak check",
		whitespaceLayoutExecute(test, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases),
	)
	test.Logf("%d whitespace nodes, six modes each, match Go and unchanged original private fork functions", count)
}

// whitespaceLayoutPolicyMutations are the output-only policy mutants of whitespace.ts, each run on source Node.
var whitespaceLayoutPolicyMutations = []struct{ name, from, to string }{
	{"policy CJ spacing tie", "return spaces > empties", "return spaces >= empties"},
	{"policy Korean pair", "return (previous === 'k-letter'", "return (previous === 'non-cjk'"},
	{"policy single line", "kind === 'tableCell' ||", "kind === 'missing-tableCell' ||"},
}

func whitespaceLayoutPolicyMutant(
	test *testing.T,
	ctx context.Context,
	cases string,
	m struct{ name, from, to string },
) {
	test.Helper()
	want, _ := whitespaceLayoutPolicyWant(test, cases)
	scratch := test.TempDir()
	if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for _, file := range []string{
		"whitespace.ts",
		"whitespaceCodec.ts",
		"codec.ts",
		"document.ts",
		"commandStack.ts",
		"width.ts",
		"widthTables.ts",
		"widthRuneRanges.ts",
		"emojiMatcher.ts",
		"testdata/whitespace_probe.ts",
	} {
		b, readError := os.ReadFile(file)
		if readError != nil {
			test.Fatal(readError)
		}
		if file == "whitespace.ts" {
			if strings.Count(string(b), m.from) != 1 {
				test.Fatal("mutation anchor")
			}
			b = []byte(strings.Replace(string(b), m.from, m.to, 1))
		}
		write(test, filepath.Join(scratch, file), b)
	}
	result := whitespaceLayoutNode(test, ctx, filepath.Join(scratch, "testdata/whitespace_probe.ts"), cases)
	clean(test, m.name, result)
	if whitespaceLayoutMutantSurvivedError := whitespaceLayoutMutantSurvived(
		result.stdout,
		want,
	); whitespaceLayoutMutantSurvivedError != nil {
		test.Fatalf("%s: %v", m.name, whitespaceLayoutMutantSurvivedError)
	}
	test.Logf("%s caught by Go policy document classification", m.name)
}

// loweringCompilerInputs are the in-repository sources the compiler that lowers a port links, and the module files
// that pin them, by the precise directories `go list -deps` names (all of cohere/internal would hash 92 MB of lint for
// about 5.6 MB used). TestLoweringInputsCoverCompilerDependencies keeps the list complete and every entry needed.
var loweringCompilerInputs = []string{
	"internal", "go.mod", "go.work", "bridge/tsgo", "cohere/TypeScript/tsc", "cohere/TypeScript-shim",
}

func whitespaceLayoutBuildFiles() []string {
	files := append([]string(nil), loweringCompilerInputs...)
	for _, directory := range []string{".", "../markdowninline", "testdata"} {
		entries, readDirectoryError := os.ReadDir(directory)
		if readDirectoryError != nil {
			panic(readDirectoryError)
		}
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".ts") ||
				strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go")) {
				files = append(
					files,
					filepath.ToSlash(filepath.Join("stage1/cohere/markdownblocks", directory, entry.Name())),
				)
			}
		}
	}
	return files
}

// The lowered ports are keyed by whitespaceLayoutBuildFiles, so it must cover every in-repository package the
// compiler links and every module file that pins one, or an untraced cache could serve a stale port. Each listed
// compiler input, dropped alone, must leave a gap, so the list carries nothing the compiler doesn't use.
func TestLoweringInputsCoverCompilerDependencies(test *testing.T) {
	test.Parallel()
	root := markdownLayoutProductRoot(test)
	// toolchain's go, under its environment, in the repository root with PWD set to it, so go.work resolves there.
	command := toolchain.CommandIn(
		root,
		"go",
		"list",
		"-deps",
		"-f",
		"{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}",
		"./internal/load",
		"./internal/lower",
		"./internal/native",
		"./internal/javascript",
	)
	listing, combinedOutputError := combinedOutput(command)
	if combinedOutputError != nil {
		test.Fatalf("go list: %v\n%s", combinedOutputError, listing)
	}
	needed := map[string]bool{"go.work": true}
	inside := func(path string) (string, bool) {
		relative, relativePathError := filepath.Rel(root, path)
		return filepath.ToSlash(
			relative,
		), relativePathError == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
	}
	for _, line := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
		directory, goMod, _ := strings.Cut(line, "\t")
		if relative, ok := inside(directory); ok {
			needed[relative] = true
		}
		if relative, ok := inside(goMod); ok && goMod != "" {
			needed[relative] = true
			if _, statError := os.Stat(filepath.Join(filepath.Dir(goMod), "go.sum")); statError == nil {
				needed[filepath.ToSlash(filepath.Join(filepath.Dir(relative), "go.sum"))] = true
			}
		}
	}
	if gaps := loweringInputGaps(whitespaceLayoutBuildFiles(), needed); len(gaps) > 0 {
		test.Fatalf("whitespaceLayoutBuildFiles misses lowering inputs: %s", strings.Join(gaps, ", "))
	}
	for index, dropped := range loweringCompilerInputs {
		planted := append(append([]string(nil), loweringCompilerInputs[:index]...), loweringCompilerInputs[index+1:]...)
		if len(loweringInputGaps(planted, needed)) == 0 {
			test.Fatalf("dropping %s leaves no gap: it keys nothing the compiler links", dropped)
		}
	}
	test.Logf(
		"%d in-repository compiler directories and module files covered; each of %d declared inputs needed",
		len(needed),
		len(loweringCompilerInputs),
	)
}

// loweringInputGaps returns the needed paths no declared path holds, sorted.
func loweringInputGaps(declared []string, needed map[string]bool) []string {
	var gaps []string
	for path := range needed {
		covered := false
		for _, file := range declared {
			if path == file || strings.HasPrefix(path, file+"/") {
				covered = true
				break
			}
		}
		if !covered {
			gaps = append(gaps, path)
		}
	}
	sort.Strings(gaps)
	return gaps
}

var whitespaceLayoutPolicyOnce sync.Once
var whitespaceLayoutPolicyProducts whitespaceLayoutProducts

func whitespaceLayoutPolicySetup(test *testing.T, ctx context.Context) whitespaceLayoutProducts {
	whitespaceLayoutPolicyOnce.Do(func() {
		started := time.Now()
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		products := whitespaceLayoutBuildProbe(test, ctx, root, "whitespace_probe.ts", false)
		inputs := products.inputs
		inputs.Name += "-native-sanitized"
		inputs.Flags = append(
			append([]string{}, inputs.Flags...),
			native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 2})...,
		)
		inputs.Flags = append(
			inputs.Flags,
			"Split=true",
			"Jobs=2",
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		)
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		dir := buildcache.Product(test, inputs, func(dir string) error {
			return whitespaceLayoutNativeBuild(
				ctx,
				products.source,
				filepath.Join(dir, "port"),
				native.Options{Sanitize: true, Split: true, Jobs: 2},
			)
		})
		products.sanitized = filepath.Join(dir, "port")
		whitespaceLayoutPolicyProducts = products
		test.Logf("TestMarkdownWhitespaceLayout (policy setup): %.3fs", time.Since(started).Seconds())
	})
	return whitespaceLayoutPolicyProducts
}

// TestMarkdownWhitespaceLayoutMutants' parts, each its own top-level test so Loom's planner can cut between them; in
// one test the three layout mutants, three policy mutants and the full-corpus policy check ran one after another. The
// layout mutants come first, then the policy mutants, then the policy check.
const whitespaceLayoutMutantsShards = 7

// whitespaceLayoutMutantsAnswers holds what every part compares against, made once per process: the corpus Go encoded
// for the port and Go's layout of it.
var whitespaceLayoutMutantsAnswers struct {
	once              sync.Once
	nativeCases, want []byte
	inputs            []auditInput
}

func whitespaceLayoutMutantsExpected(
	test *testing.T,
	ctx context.Context,
	products whitespaceLayoutProducts,
) ([]byte, []byte, []auditInput) {
	test.Helper()
	whitespaceLayoutMutantsAnswers.once.Do(func() {
		inputs, _ := whitespaceLayoutEnumerate(test)
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
		nativeCases, canonical := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
		want := whitespaceLayoutExecute(test, ctx, nil, products.goBinary, cases, nativeCases, canonical)
		clean(test, "Go mutant fixtures", want)
		// Go writes the port's cases into the first caller's temporary directory, gone when that test ends, so their
		// bytes are kept.
		encoded, readError := os.ReadFile(nativeCases)
		if readError != nil {
			test.Fatal(readError)
		}
		whitespaceLayoutMutantsAnswers.want, whitespaceLayoutMutantsAnswers.inputs = want.stdout, inputs
		whitespaceLayoutMutantsAnswers.nativeCases = encoded
	})
	if whitespaceLayoutMutantsAnswers.nativeCases == nil {
		test.Fatal("the Go mutant fixtures failed earlier in this process")
	}
	return whitespaceLayoutMutantsAnswers.nativeCases, whitespaceLayoutMutantsAnswers.want, whitespaceLayoutMutantsAnswers.inputs
}

// whitespaceLayoutMutantSurvived is a mutant's verdict: one whose output still equals Go's was missed.
func whitespaceLayoutMutantSurvived(output, want []byte) error {
	if bytes.Equal(output, want) {
		return fmt.Errorf("survived")
	}
	return nil
}

func whitespaceLayoutMutantsShard(test *testing.T, index int) {
	test.Helper()
	products, policyProducts := whitespaceLayoutReadyProducts(test)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	encoded, want, inputs := whitespaceLayoutMutantsExpected(test, ctx, products)
	nativeCases := filepath.Join(test.TempDir(), "native.txt")
	write(test, nativeCases, encoded)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(root, "cohere/internal/format/prettier/bundles")
	}
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	layout, policy := len(whitespaceLayoutMutations), len(whitespaceLayoutPolicyMutations)
	switch {
	case index < layout:
		fixture := &layoutFixture{main: main, nativeCases: nativeCases, inputs: inputs, want: want}
		whitespaceLayoutMutant(test, ctx, fixture, whitespaceLayoutMutations[index])
	case index < layout+policy:
		whitespaceLayoutPolicyMutant(test, ctx, nativeCases, whitespaceLayoutPolicyMutations[index-layout])
	default:
		whitespaceLayoutPolicy(test, ctx, nativeCases, fork, policyProducts)
	}
}

// whitespaceLayoutMutantsEnumeration reads the top-level wrappers themselves, so a missing, duplicated or miswired
// part can't go unrun while the count still agrees.
func whitespaceLayoutMutantsEnumeration(test *testing.T) []int {
	test.Helper()
	if parts := len(whitespaceLayoutMutations) +
		len(whitespaceLayoutPolicyMutations) +
		1; parts != whitespaceLayoutMutantsShards {
		test.Fatalf(
			"%d whitespace mutant parts, but whitespaceLayoutMutantsShards is %d",
			parts,
			whitespaceLayoutMutantsShards,
		)
	}
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "whitespace_layout_split_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	var plan []int
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestMarkdownWhitespaceLayoutMutants_") {
			continue
		}
		shard, parseError := strconv.Atoi(
			strings.TrimPrefix(function.Name.Name, "TestMarkdownWhitespaceLayoutMutants_"),
		)
		if parseError != nil {
			test.Fatal(parseError)
		}
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "whitespaceLayoutMutantsShard" {
				return true
			}
			if len(call.Args) != 2 {
				test.Fatalf("%s: invalid shard call", function.Name.Name)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				test.Fatalf("%s: shard index must be a literal", function.Name.Name)
			}
			if index, parseError := strconv.Atoi(literal.Value); parseError != nil || index != shard {
				test.Fatalf("%s runs part %s", function.Name.Name, literal.Value)
			}
			calls++
			return true
		})
		if calls != 1 {
			test.Fatalf("%s runs %d parts, want one", function.Name.Name, calls)
		}
		plan = append(plan, shard)
	}
	return plan
}

func TestMarkdownWhitespaceLayoutMutantsUnion(test *testing.T) {
	test.Parallel()
	visits := make([]int, whitespaceLayoutMutantsShards)
	for _, index := range whitespaceLayoutMutantsEnumeration(test) {
		if index < 0 || index >= len(visits) {
			test.Fatalf("TestMarkdownWhitespaceLayoutMutants_%03d names no part", index)
		}
		visits[index]++
	}
	for index, count := range visits {
		if count != 1 {
			test.Fatalf("part %d run by %d top-level tests, want one", index, count)
		}
	}
	test.Logf("union: all %d whitespace mutant parts run, each by exactly one top-level test", len(visits))
}

func TestMarkdownWhitespaceLayoutMutantsPlantedFailure(test *testing.T) {
	test.Parallel()
	// Plant a survivor in one mutant's output. Exactly its owner must reject it; every other mutant differs from Go
	// and stays accepted. The last part, the policy check, compares rather than mutates.
	var caught []int
	for _, shard := range whitespaceLayoutMutantsEnumeration(test) {
		if shard == whitespaceLayoutMutantsShards-1 {
			continue
		}
		output := []byte("wrong bytes")
		if shard == 4 {
			output = []byte("Go layout bytes")
		}
		if whitespaceLayoutMutantSurvived(output, []byte("Go layout bytes")) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 4 {
		test.Fatalf("planted survivor caught by %v, want [4]", caught)
	}
	test.Log("planted survivor caught exactly by TestMarkdownWhitespaceLayoutMutants_004")
}

func TestMarkdownWhitespaceLayoutMutants_000(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 0)
}

func TestMarkdownWhitespaceLayoutMutants_001(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 1)
}

func TestMarkdownWhitespaceLayoutMutants_002(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 2)
}

func TestMarkdownWhitespaceLayoutMutants_003(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 3)
}

func TestMarkdownWhitespaceLayoutMutants_004(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 4)
}

func TestMarkdownWhitespaceLayoutMutants_005(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 5)
}

func TestMarkdownWhitespaceLayoutMutants_006(test *testing.T) {
	test.Parallel()
	whitespaceLayoutMutantsShard(test, 6)
}

func TestMarkdownWhitespaceLayout_012(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 12)
}
func TestMarkdownWhitespaceLayout_013(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 13)
}
func TestMarkdownWhitespaceLayout_014(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 14)
}
func TestMarkdownWhitespaceLayout_015(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 15)
}
func TestMarkdownWhitespaceLayout_016(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 16)
}
func TestMarkdownWhitespaceLayout_017(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 17)
}
func TestMarkdownWhitespaceLayout_018(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 18)
}
func TestMarkdownWhitespaceLayout_019(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 19)
}
func TestMarkdownWhitespaceLayout_020(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 20)
}
func TestMarkdownWhitespaceLayout_021(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 21)
}
func TestMarkdownWhitespaceLayout_022(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 22)
}
func TestMarkdownWhitespaceLayout_023(test *testing.T) {
	test.Parallel()
	whitespaceLayoutRunShard(test, 23)
}

// Independently selected shards prepare the shared products once per process.
var whitespaceLayoutReadyOnce sync.Once
var whitespaceLayoutPrepared, whitespaceLayoutPolicyPrepared whitespaceLayoutProducts

func whitespaceLayoutSetupInputs(test *testing.T) buildcache.Inputs {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	files := whitespaceLayoutBuildFiles()
	files = append(files, "stage1/cohere/markdownblocks/whitespace_layout_split_test.go", "cohere/internal")
	return buildcache.Inputs{Name: "markdown-whitespace-layout-setup-v2", Files: files,
		Flags: []string{
			root,
			os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK"),
			os.Getenv("ADAMIC_NATIVE_SPLIT"),
			os.Getenv("GOFLAGS"),
			os.Getenv("CGO_ENABLED"),
			"Split=true",
			"Jobs=2",
		},
		Toolchain: []string{
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("go", "version"),
		}}
}

func whitespaceLayoutReadyProducts(test *testing.T) (whitespaceLayoutProducts, whitespaceLayoutProducts) {
	test.Helper()
	whitespaceLayoutReadyOnce.Do(func() { whitespaceLayoutPrepare(test) })
	if whitespaceLayoutPrepared.goBinary == "" {
		test.Fatal("whitespace layout setup failed")
	}
	return whitespaceLayoutPrepared, whitespaceLayoutPolicyPrepared
}

func whitespaceLayoutReadProducts(
	test *testing.T,
	directory string,
) (whitespaceLayoutProducts, whitespaceLayoutProducts) {
	data, readError := os.ReadFile(filepath.Join(directory, "products.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var paths map[string]string
	if decodeError := json.Unmarshal(data, &paths); decodeError != nil {
		test.Fatal(decodeError)
	}
	for name, path := range paths {
		paths[name] = buildcache.Absolute(path)
	}
	backend, readError := os.ReadFile(filepath.Join(directory, "layout.mjs"))
	if readError != nil {
		test.Fatal(readError)
	}
	policy, readError := os.ReadFile(filepath.Join(directory, "policy.mjs"))
	if readError != nil {
		test.Fatal(readError)
	}
	return whitespaceLayoutProducts{
		goBinary:  paths["go"],
		goLayout:  paths["goLayout"],
		sanitized: paths["sanitized"],
		release:   paths["release"],
		backend:   backend,
	}, whitespaceLayoutProducts{sanitized: paths["policy"], backend: policy}
}

func whitespaceLayoutCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

// whitespaceLayoutExecuteResult reads a run's output through childguard.Capture, to its end with no timer deciding
// when: exec's own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached
// EOF a second after the child exited, so a run that had finished could fail its shard under load (#n3gk22h's shape).
func whitespaceLayoutExecuteResult(
	ctx context.Context,
	environment []string,
	name string,
	args ...string,
) (run, error) {
	command := whitespaceLayoutCommand(ctx, name, args...)
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
func whitespaceLayoutExecute(
	test *testing.T,
	ctx context.Context,
	environment []string,
	name string,
	args ...string,
) run {
	test.Helper()
	result, whitespaceLayoutExecuteResultError := whitespaceLayoutExecuteResult(ctx, environment, name, args...)
	if whitespaceLayoutExecuteResultError != nil {
		test.Fatal(whitespaceLayoutExecuteResultError)
	}
	return result
}
func whitespaceLayoutNodeResult(ctx context.Context, path string, args ...string) (run, error) {
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return whitespaceLayoutExecuteResult(
		ctx,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...,
	)
}
func whitespaceLayoutNode(test *testing.T, ctx context.Context, path string, args ...string) run {
	test.Helper()
	result, whitespaceLayoutNodeResultError := whitespaceLayoutNodeResult(ctx, path, args...)
	if whitespaceLayoutNodeResultError != nil {
		test.Fatal(whitespaceLayoutNodeResultError)
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
	if writeError := os.WriteFile(sourcePath, []byte(source), 0644); writeError != nil {
		return writeError
	}
	defer os.Remove(sourcePath)
	request, encodeError := json.Marshal(
		whitespaceLayoutBuildRequest{Source: sourcePath, Output: output, Options: options},
	)
	if encodeError != nil {
		return encodeError
	}
	executable, executableError := os.Executable()
	if executableError != nil {
		return executableError
	}
	command := whitespaceLayoutCommand(
		ctx,
		executable,
		"-test.run=^TestMarkdownWhitespaceLayoutBuildWorker$",
		"-test.timeout=0",
	)
	command.Env = append(os.Environ(), "ADAMIC_WHITESPACE_BUILD="+string(request))
	// Read through childguard.Capture for whitespaceLayoutExecuteResult's reason. The worker starts compilers of its
	// own, and Capture kills the group once it exits, so none of them holds a pipe after it.
	if stdout, stderr, captureError := childguard.Capture(ctx, command); captureError != nil {
		return fmt.Errorf("native build: %w\n%s%s", captureError, stdout, stderr)
	}
	return nil
}
func TestMarkdownWhitespaceLayoutBuildWorker(test *testing.T) {
	test.Parallel()
	request := os.Getenv("ADAMIC_WHITESPACE_BUILD")
	if request == "" {
		return
	}
	var build whitespaceLayoutBuildRequest
	if decodeError := json.Unmarshal([]byte(request), &build); decodeError != nil {
		test.Fatal(decodeError)
	}
	source, readError := os.ReadFile(build.Source)
	if readError != nil {
		test.Fatal(readError)
	}
	if buildError := native.Build(string(source), build.Output, build.Options); buildError != nil {
		test.Fatal(buildError)
	}
}

func whitespaceLayoutNativeProduct(
	test *testing.T,
	ctx context.Context,
	products whitespaceLayoutProducts,
	sanitize bool,
) string {
	options := native.Options{Sanitize: sanitize, Split: true, Jobs: 2}
	nativeInputs := products.inputs
	nativeInputs.Name = fmt.Sprintf("%s-native-%t", products.inputs.Name, sanitize)
	nativeInputs.Flags = append(append([]string{}, products.inputs.Flags...), native.Flags(options)...)
	nativeInputs.Flags = append(
		nativeInputs.Flags,
		"Split=true",
		"Jobs=2",
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
	)
	nativeInputs.Toolchain = append([]string{runtime.Version()}, buildcache.Tool("clang", "--version"))
	dir := buildcache.Product(test, nativeInputs, func(dir string) error {
		return whitespaceLayoutNativeBuild(ctx, products.source, filepath.Join(dir, "port"), options)
	})
	return dir
}
