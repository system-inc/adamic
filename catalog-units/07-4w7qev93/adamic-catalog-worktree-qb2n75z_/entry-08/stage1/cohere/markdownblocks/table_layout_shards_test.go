package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testMarkdownTableLayoutShards = 8

type tableLayoutProducts struct{ main, list, document, sanitized, release, backend string }

func tableLayoutShard(key string) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int(h.Sum64() % testMarkdownTableLayoutShards)
}

var tableLayoutOnce sync.Once
var tableLayoutSharedProducts tableLayoutProducts
var tableLayoutInputs []auditInput

func tableLayoutSetup(t *testing.T) {
	t.Helper()
	tableLayoutOnce.Do(func() {
		tableLayoutSharedProducts = tableLayoutBuild(t)
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		tableLayoutInputs, _ = blockCorpus(t, root, "whitespace")
	})
	if tableLayoutInputs == nil {
		t.Fatal("table layout setup failed")
	}
}

func testMarkdownTableLayout(t *testing.T) {
	started := time.Now()
	tableLayoutSetup(t)
	t.Logf("TestMarkdownTableLayout (setup) %.6fs", time.Since(started).Seconds())
}

func tableLayoutPartition(inputs []auditInput) [][]auditInput {
	shards := make([][]auditInput, testMarkdownTableLayoutShards)
	for _, input := range inputs {
		shard := tableLayoutShard(input.Name)
		shards[shard] = append(shards[shard], input)
	}
	return shards
}

func TestMarkdownTableLayoutUnion(t *testing.T) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := blockCorpus(t, root, "whitespace")
	shards := tableLayoutPartition(inputs)
	expected, actual := make(map[auditInput]int), make(map[auditInput]int)
	for _, input := range inputs {
		expected[input]++
	}
	total := 0
	for shard, cases := range shards {
		for _, input := range cases {
			if tableLayoutShard(input.Name) != shard {
				t.Fatal("case assigned to wrong shard")
			}
			actual[input]++
			total++
		}
	}
	if len(shards) != testMarkdownTableLayoutShards || total != len(inputs) || len(actual) != len(expected) {
		t.Fatal("shard union differs from live enumeration")
	}
	for input, count := range expected {
		if actual[input] != count {
			t.Fatalf("case %q not covered exactly once per enumerated occurrence", input.Name)
		}
	}
	t.Logf("union %d cases; planted-failure target %q in TestMarkdownTableLayout_%03d", total, inputs[0].Name, tableLayoutShard(inputs[0].Name))
}

func tableLayoutRunShard(t *testing.T, shard int) {
	t.Helper()
	t.Parallel()
	tableLayoutSetup(t)
	shards := tableLayoutPartition(tableLayoutInputs)
	if len(shards) != testMarkdownTableLayoutShards || shard < 0 || shard >= len(shards) {
		t.Fatal("invalid shard enumeration")
	}
	fixture := tableLayoutFixture(t, shards[shard], tableLayoutSharedProducts, tableLayoutInputs[0].Name)
	var caught [3]atomic.Int64
	tableLayoutChecks(t, fixture, tableLayoutSharedProducts, &caught)
	for index := range caught {
		if caught[index].Load() == 0 {
			t.Errorf("table mutant %d survived shard %03d", index, shard)
		}
	}
}

func TestMarkdownTableLayout_000(t *testing.T) { tableLayoutRunShard(t, 0) }

func TestMarkdownTableLayout_001(t *testing.T) { tableLayoutRunShard(t, 1) }

func TestMarkdownTableLayout_002(t *testing.T) { tableLayoutRunShard(t, 2) }

func TestMarkdownTableLayout_003(t *testing.T) { tableLayoutRunShard(t, 3) }

func TestMarkdownTableLayout_004(t *testing.T) { tableLayoutRunShard(t, 4) }

func TestMarkdownTableLayout_005(t *testing.T) { tableLayoutRunShard(t, 5) }

func TestMarkdownTableLayout_006(t *testing.T) { tableLayoutRunShard(t, 6) }

func TestMarkdownTableLayout_007(t *testing.T) { tableLayoutRunShard(t, 7) }

func tableLayoutBuild(t *testing.T) tableLayoutProducts {
	t.Helper()
	var products tableLayoutProducts
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	products.main, err = filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "markdown table lowered", Files: tableLayoutBuildFiles(t), Toolchain: []string{runtime.Version()}}
	loweredDirectory := buildcache.Product(t, inputs, func(directory string) error {
		program, err := loweredResult(products.main)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	source, err := os.ReadFile(filepath.Join(loweredDirectory, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	backend, err := os.ReadFile(filepath.Join(loweredDirectory, "program.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	products.backend = string(backend)
	var workers sync.WaitGroup
	defer workers.Wait()
	builds := make([]*fixtureTask[string], 0, 2)
	for _, sanitize := range []bool{true, false} {
		modeInputs := inputs
		modeInputs.Name = fmt.Sprintf("markdown table native %t", sanitize)
		modeInputs.Flags = native.Flags(native.Options{Sanitize: sanitize})
		modeInputs.Toolchain = []string{runtime.Version(), buildcache.Tool("clang", "--version")}
		builds = append(builds, startFixtureTask(&workers, func() (string, error) {
			directory := buildcache.Product(t, modeInputs, func(directory string) error {
				return native.Build(string(source), filepath.Join(directory, "program"), native.Options{Sanitize: sanitize})
			})
			return filepath.Join(directory, "program"), nil
		}))
	}
	directory := filepath.Join(artifactDirectory, "table-layout-builds")
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	for _, mode := range []string{"list", "document"} {
		driver := "list_go.go"
		command := "adamic_markdown_lists"
		if mode == "document" {
			driver = "document_go.go"
			command = "adamic_markdown_doclayout"
		}
		driverPath, err := filepath.Abs(filepath.Join("testdata", driver))
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(cohere, "cmd", command, "main.go")
		replacements := map[string]string{main: driverPath}
		if mode == "list" {
			bridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replacements[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replacements})
		if err != nil {
			t.Fatal(err)
		}
		overlayPath := filepath.Join(directory, mode+".json")
		write(t, overlayPath, overlay)
		binary := filepath.Join(directory, mode)
		build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", binary, main)
		build.Dir = cohere
		if output, err := combinedOutput(build); err != nil {
			t.Fatalf("Go %s: %v\n%s", mode, err, output)
		}
		if mode == "list" {
			products.list = binary
		} else {
			products.document = binary
		}
	}
	products.sanitized = builds[0].await(t)
	products.release = builds[1].await(t)
	if products.sanitized == "" || products.release == "" {
		t.Fatal("native product build failed")
	}
	return products
}

func tableLayoutFixture(t *testing.T, inputs []auditInput, products tableLayoutProducts, planted string) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	files := 0
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
	goBinary := products.list
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {
		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var err error
			mutantWant, err = executeResult(t, nil, goBinary, fullCases, mutantNativeCases, filepath.Join(dir, "mutant-canonical.txt"))
			if err != nil {
				return run{}, err
			}
		}
		return executeResult(t, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.document
	main := products.main
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
		return executeResult(t, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	var program *ir.Program
	want := listTask.await(t)
	clean(t, "Go list fixtures", want)
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
		return executeResult(t, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) { return onNodeResult(t, main, nativeCases) })
	backendPath := filepath.Join(dir, "program.mjs")
	write(t, backendPath, []byte(products.backend))
	backendTask := startFixtureTask(&workers, func() (run, error) { return onNodeResult(t, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.sanitized
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return executeResult(t, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary := products.release
		return executeResult(t, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) { return executeResult(t, nil, "node", script, fork, canonicalCases) })
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
		if !bytes.Equal(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			t.Fatalf("%s output byte %d in %s\ngot %q\nwant %q", side.name, offset, inputs[index].Name, actual[index], expected[index])
		}
	}
	if report := leaks(t, program, binary, nativeCases); report != "" {
		t.Fatal(report)
	}
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
	return &layoutFixture{
		selection: selection, mutantNativeCases: mutantNativeCases, mutantInputs: fullInputs, mutantWant: mutantWant.stdout,
		root: root, directory: dir, nativeCases: nativeCases, canonicalCases: canonicalCases,
		main: main, fork: fork, goLayout: goLayout, script: script,
		inputs: inputs, files: files, want: want.stdout, program: program,
	}
}

func tableLayoutChecks(t *testing.T, fixture *layoutFixture, products tableLayoutProducts, caught *[3]atomic.Int64) {
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
		result := onNode(t, mutantMain, nativeCases)
		clean(t, "source Node layout mutant", result)
		if bytes.Equal(result.stdout, want.stdout) {
			continue
		}
		caught[mutationIndex].Add(1)
		offset := firstDifference(string(result.stdout), string(want.stdout))
		index := bytes.Count(want.stdout[:offset], []byte("\n"))
		t.Logf("output-only source Node mutant caught by %q at byte %d", inputs[index].Name, offset)
	}
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
				result := execute(t, nil, side.command, side.args...)
				elapsed += time.Since(start)
				clean(t, side.name, result)
				equal(t, side.name, result.stdout, want.stdout)
			}
			t.Logf("component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
		}
	}
	data, err := os.ReadFile(nativeCases)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d physical files, %d generated, %d whole-document contexts; %d native list frames, %d composed word nodes; Go/source/native/backend/original doc bytes identical", files, len(inputs)-files, len(inputs), bytes.Count(data, []byte("\nL\t")), bytes.Count(data, []byte("\nW\t")))
	t.Logf("%d native HTML frames", bytes.Count(data, []byte("\nH\t")))
	t.Logf("%d native code frames", bytes.Count(data, []byte("\nC\t")))
	t.Logf("%d native table frames", bytes.Count(data, []byte("\nT\t")))
	t.Logf("%d native quote frames", bytes.Count(data, []byte("\nQ\t")))
}

// Build keys include source and compiler inputs, never this test's shard wrappers.
func tableLayoutBuildFiles(t *testing.T) []string {
	t.Helper()
	files := []string{"internal", "cohere", "go.mod", "go.work"}
	for _, pattern := range []string{"*.ts", "testdata/*.ts", "../markdowninline/*.ts"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range matches {
			files = append(files, filepath.ToSlash(filepath.Clean(filepath.Join("stage1/cohere/markdownblocks", path))))
		}
	}
	return files
}
