package markdownblocks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

const testMarkdownTableLayoutShards = 8

type tableLayoutProducts struct{ main, list, document, sanitized, release, backend string }

func tableLayoutShard(key string) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int(h.Sum64() % testMarkdownTableLayoutShards)
}

var tableLayoutSharedProducts tableLayoutProducts
var tableLayoutInputs []auditInput
var tableLayoutSetupElapsed time.Duration

// Every selected shard prepares once per process, before its own deadline.
var tableLayoutOnce sync.Once

func tableLayoutReady(test *testing.T) {
	test.Helper()
	tableLayoutOnce.Do(func() {
		// Setup has no deadline; Loom's kill covers the whole unit.
		started := time.Now()
		tableLayoutSharedProducts = tableLayoutBuild(test)
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		tableLayoutInputs, _ = blockCorpus(test, root, "whitespace")
		tableLayoutSetupElapsed = time.Since(started)
	})
	if tableLayoutInputs == nil {
		test.Fatal("table layout setup failed")
	}
}

func testMarkdownTableLayout(test *testing.T) {
	tableLayoutReady(test)
	test.Logf("shared setup %.6fs; before shard deadline", tableLayoutSetupElapsed.Seconds())
}

func TestMarkdownTableLayout_Setup(test *testing.T) {
	test.Parallel()
	testMarkdownTableLayout(test)
}

func tableLayoutPartition(inputs []auditInput) [][]auditInput {
	shards := make([][]auditInput, testMarkdownTableLayoutShards)
	for _, input := range inputs {
		shard := tableLayoutShard(input.Name)
		shards[shard] = append(shards[shard], input)
	}
	return shards
}

func TestMarkdownTableLayoutUnion(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := blockCorpus(test, root, "whitespace")
	shards := tableLayoutPartition(inputs)
	expected, actual := make(map[auditInput]int), make(map[auditInput]int)
	for _, input := range inputs {
		expected[input]++
	}
	total := 0
	for shard, cases := range shards {
		for _, input := range cases {
			if tableLayoutShard(input.Name) != shard {
				test.Fatal("case assigned to wrong shard")
			}
			actual[input]++
			total++
		}
	}
	if len(shards) != testMarkdownTableLayoutShards || total != len(inputs) || len(actual) != len(expected) {
		test.Fatal("shard union differs from live enumeration")
	}
	for input, count := range expected {
		if actual[input] != count {
			test.Fatalf("case %q not covered exactly once per enumerated occurrence", input.Name)
		}
	}
	test.Logf(
		"union %d cases; planted-failure target %q in TestMarkdownTableLayout_%03d",
		total,
		inputs[0].Name,
		tableLayoutShard(inputs[0].Name),
	)
}

func tableLayoutRunShard(test *testing.T, shard int) {
	test.Helper()
	tableLayoutReady(test)
	timeoutContext, cancel := context.WithTimeout(test.Context(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { test.Logf("shard cases %.6fs; setup excluded", time.Since(started).Seconds()) }()
	shards := tableLayoutPartition(tableLayoutInputs)
	if len(shards) != testMarkdownTableLayoutShards || shard < 0 || shard >= len(shards) {
		test.Fatal("invalid shard enumeration")
	}
	fixture := tableLayoutFixture(
		test,
		timeoutContext,
		shards[shard],
		tableLayoutSharedProducts,
		tableLayoutInputs[0].Name,
	)
	var caught [3]atomic.Int64
	tableLayoutChecks(test, timeoutContext, fixture, tableLayoutSharedProducts, &caught)
	for index := range caught {
		if caught[index].Load() == 0 {
			test.Errorf("table mutant %d survived shard %03d", index, shard)
		}
	}
}

func TestMarkdownTableLayout_000(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 0) }

func TestMarkdownTableLayout_001(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 1) }

func TestMarkdownTableLayout_002(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 2) }

func TestMarkdownTableLayout_003(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 3) }

func TestMarkdownTableLayout_004(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 4) }

func TestMarkdownTableLayout_005(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 5) }

func TestMarkdownTableLayout_006(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 6) }

func TestMarkdownTableLayout_007(test *testing.T) { test.Parallel(); tableLayoutRunShard(test, 7) }

func tableLayoutBuild(test *testing.T) tableLayoutProducts {
	return tableLayoutBuildProduct(test, "")
}

func tableLayoutBuildProduct(test *testing.T, target string) tableLayoutProducts {
	test.Helper()
	var products tableLayoutProducts
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	products.main, absolutePathError = filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs := buildcache.Inputs{
		Name:      "markdown table lowered",
		Files:     tableLayoutBuildFiles(test),
		Toolchain: []string{runtime.Version()},
	}
	loweredDirectory := buildcache.Product(test, inputs, func(directory string) error {
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
	source, readError := os.ReadFile(filepath.Join(loweredDirectory, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	backend, readError := os.ReadFile(filepath.Join(loweredDirectory, "program.mjs"))
	if readError != nil {
		test.Fatal(readError)
	}
	products.backend = string(backend)
	if target == "lowered" {
		return products
	}
	var workers sync.WaitGroup
	defer workers.Wait()
	builds := make([]*fixtureTask[string], 0, 2)
	for _, sanitize := range []bool{true, false} {
		if target != "" && target != fmt.Sprintf("native-%t", sanitize) {
			continue
		}
		modeInputs := inputs
		modeInputs.Name = fmt.Sprintf("markdown table native %t", sanitize)
		// native.Build splits by ADAMIC_NATIVE_SPLIT when its options don't, so the switch is an input.
		modeInputs.Flags = append(
			native.Flags(native.Options{Sanitize: sanitize}),
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		)
		modeInputs.Toolchain = []string{runtime.Version(), buildcache.Tool("clang", "--version")}
		builds = append(builds, startFixtureTask(&workers, func() (string, error) {
			directory := buildcache.Product(test, modeInputs, func(directory string) error {
				return native.Build(
					string(source),
					filepath.Join(directory, "program"),
					native.Options{Sanitize: sanitize},
				)
			})
			return filepath.Join(directory, "program"), nil
		}))
	}
	goProducts := oracles.TableLayoutGoProducts(test, root, target)
	products.list, products.document = goProducts.List, goProducts.Document
	if target == "" {
		products.sanitized = builds[0].await(test)
		products.release = builds[1].await(test)
	} else if target == "native-true" {
		products.sanitized = builds[0].await(test)
	} else if target == "native-false" {
		products.release = builds[0].await(test)
	}
	if target == "" && (products.sanitized == "" || products.release == "") {
		test.Fatal("native product build failed")
	}
	return products
}

func tableLayoutFixture(
	test *testing.T,
	callerContext context.Context,
	inputs []auditInput,
	products tableLayoutProducts,
	planted string,
) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	files := 0
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
	goBinary := products.list
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {
		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var tableLayoutExecuteError error
			mutantWant, tableLayoutExecuteError = tableLayoutExecute(
				callerContext,
				nil,
				goBinary,
				fullCases,
				mutantNativeCases,
				filepath.Join(dir, "mutant-canonical.txt"),
			)
			if tableLayoutExecuteError != nil {
				return run{}, tableLayoutExecuteError
			}
		}
		return tableLayoutExecute(callerContext, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.document
	main := products.main
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
		return tableLayoutExecute(callerContext, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	var program *ir.Program
	want := listTask.await(test)
	clean(test, "Go list fixtures", want)
	// Opt-in negative control: corrupt exactly one oracle line, after fixture generation.
	// The real document byte comparison must reject it in exactly its owning shard.
	if os.Getenv("ADAMIC_TABLE_LAYOUT_PLANT") == "1" {
		for index, input := range inputs {
			if input.Name == planted {
				lines := bytes.Split(want.stdout, []byte("\n"))
				lines[index] = append([]byte("planted-disagreement:"), lines[index]...)
				want.stdout = bytes.Join(lines, []byte("\n"))
			}
		}
	}
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
		return tableLayoutExecute(callerContext, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(
		&workers,
		func() (run, error) { return tableLayoutNode(callerContext, main, nativeCases) },
	)
	backendPath := filepath.Join(dir, "program.mjs")
	write(test, backendPath, []byte(products.backend))
	backendTask := startFixtureTask(
		&workers,
		func() (run, error) { return tableLayoutNode(callerContext, backendPath, nativeCases) },
	)
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.sanitized
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return tableLayoutExecute(callerContext, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.release
		return tableLayoutExecute(callerContext, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(
		&workers,
		func() (run, error) {
			return tableLayoutExecute(callerContext, nil, "node", script, fork, canonicalCases)
		},
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
		if !bytes.Equal(side.result.stdout, want.stdout) {
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
	leakCommand, leakArguments := binary, []string{nativeCases}
	leakEnvironment := []string{"ASAN_OPTIONS=detect_leaks=1"}
	if runtime.GOOS == "darwin" {
		leakCommand = "leaks"
		leakArguments = []string{"--atExit", "--", products.release, nativeCases}
		leakEnvironment = nil
	}
	leakResult, tableLayoutExecuteError := tableLayoutExecute(
		callerContext,
		leakEnvironment,
		leakCommand,
		leakArguments...,
	)
	if tableLayoutExecuteError != nil {
		test.Fatal(tableLayoutExecuteError)
	}
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
	return &layoutFixture{
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
		program:           program,
	}
}

func tableLayoutChecks(
	test *testing.T,
	callerContext context.Context,
	fixture *layoutFixture,
	products tableLayoutProducts,
	caught *[3]atomic.Int64,
) {
	root, nativeCases, canonicalCases := fixture.root, fixture.nativeCases, fixture.canonicalCases
	main, fork, goLayout, script := fixture.main, fixture.fork, fixture.goLayout, fixture.script
	inputs, files := fixture.inputs, fixture.files
	want := run{stdout: fixture.want}
	mutantFile := "tables.ts"
	mutations := []struct{ name, from, to string }{
		{"table minimum", "widths.push(3)", "widths.push(4)"},
		{"table center", "Math.floor(spaces / 2)", "Math.ceil(spaces / 2)"},
		{"table alignment", "align === 'right' ? spaces", "align === 'right' ? 0"},
	}
	for mutationIndex, mutation := range mutations {
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
				content = []byte(
					strings.Replace(string(content), "../markdowninline/inline.ts", "./markdowninline/inline.ts", 1),
				)
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
		result, tableLayoutNodeError := tableLayoutNode(callerContext, mutantMain, nativeCases)
		if tableLayoutNodeError != nil {
			test.Fatal(tableLayoutNodeError)
		}
		clean(test, "source Node layout mutant", result)
		if bytes.Equal(result.stdout, want.stdout) {
			continue
		}
		caught[mutationIndex].Add(1)
		offset := firstDifference(string(result.stdout), string(want.stdout))
		index := bytes.Count(want.stdout[:offset], []byte("\n"))
		test.Logf("output-only source Node mutant caught by %q at byte %d", inputs[index].Name, offset)
	}
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
				result, tableLayoutExecuteError := tableLayoutExecute(callerContext, nil, side.command, side.args...)
				if tableLayoutExecuteError != nil {
					test.Fatal(tableLayoutExecuteError)
				}
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
	data, readError := os.ReadFile(nativeCases)
	if readError != nil {
		test.Fatal(readError)
	}
	test.Logf(
		"%d physical files, %d generated, %d whole-document contexts; %d native list frames, %d composed word nodes; Go/source/native/backend/original doc bytes identical",
		files,
		len(inputs)-files,
		len(inputs),
		bytes.Count(data, []byte("\nL\t")),
		bytes.Count(data, []byte("\nW\t")),
	)
	test.Logf("%d native HTML frames", bytes.Count(data, []byte("\nH\t")))
	test.Logf("%d native code frames", bytes.Count(data, []byte("\nC\t")))
	test.Logf("%d native table frames", bytes.Count(data, []byte("\nT\t")))
	test.Logf("%d native quote frames", bytes.Count(data, []byte("\nQ\t")))
}

// Build keys include source and compiler inputs, never this test's shard wrappers.
func tableLayoutBuildFiles(test *testing.T) []string {
	test.Helper()
	files := []string{"internal", "cohere", "go.mod", "go.work"}
	for _, pattern := range []string{"*.ts", "testdata/*.ts", "../markdowninline/*.ts"} {
		matches, globError := filepath.Glob(pattern)
		if globError != nil {
			test.Fatal(globError)
		}
		for _, path := range matches {
			files = append(files, filepath.ToSlash(filepath.Clean(filepath.Join("stage1/cohere/markdownblocks", path))))
		}
	}
	return files
}
