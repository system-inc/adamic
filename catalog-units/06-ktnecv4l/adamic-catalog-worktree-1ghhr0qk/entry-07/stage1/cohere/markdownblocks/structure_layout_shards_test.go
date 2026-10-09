package markdownblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const testMarkdownStructureLayoutShards = 16

type structureLayoutProducts struct {
	root, goList, goLayout, sanitized, release, plantedName string
	javascript                                              []byte
	program                                                 *ir.Program
}

func structureLayoutShard(input auditInput) int {
	sum := sha256.Sum256([]byte(input.Name + "\x00" + input.Text))
	return int(binary.LittleEndian.Uint64(sum[:8]) % testMarkdownStructureLayoutShards)
}

type structureLayoutState struct {
	inputs   []auditInput
	shards   [][]auditInput
	products structureLayoutProducts
}

var structureLayoutOnce sync.Once
var structureLayoutComplete *structureLayoutState

func structureLayoutEnumeration(t *testing.T) *structureLayoutState {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := blockCorpus(t, root, "whitespace")
	state := &structureLayoutState{inputs: inputs, shards: make([][]auditInput, testMarkdownStructureLayoutShards)}
	for _, input := range inputs {
		index := structureLayoutShard(input)
		state.shards[index] = append(state.shards[index], input)
	}
	state.products.root = root
	state.products.plantedName = inputs[0].Name
	return state
}

func TestMarkdownStructureLayoutUnion(t *testing.T) {
	parallelMarkdown(t)
	deadline := structureLayoutDeadline(t, t.Name())
	defer deadline.Stop()
	state := structureLayoutEnumeration(t)
	// Compare multiplicities by the complete stable key, including any duplicate
	// corpus entries, rather than assuming corpus names or totals are fixed.
	want, got := map[string]int{}, map[string]int{}
	key := func(input auditInput) string { return input.Name + "\x00" + input.Text }
	for _, input := range state.inputs {
		want[key(input)]++
	}
	total, detections, plantedShard := 0, 0, -1
	for index, inputs := range state.shards {
		for _, input := range inputs {
			if structureLayoutShard(input) != index {
				t.Fatalf("misassigned case %q", input.Name)
			}
			got[key(input)]++
			total++
			if input.Name == state.products.plantedName {
				if bytes.Equal([]byte(input.Text), append([]byte(input.Text), '!')) {
					t.Fatal("planted byte disagreement survived")
				}
				detections++
				plantedShard = index
			}
		}
	}
	if len(state.shards) != testMarkdownStructureLayoutShards || total != len(state.inputs) || len(got) != len(want) {
		t.Fatal("shard union differs from live enumeration")
	}
	for key, count := range want {
		if got[key] != count {
			t.Fatalf("case %q: union count %d, enumeration count %d", key, got[key], count)
		}
	}
	if detections != 1 {
		t.Fatalf("planted failure caught %d times", detections)
	}
	t.Logf("union: %d cases, each exactly once; planted byte disagreement belongs to TestMarkdownStructureLayout_%03d", total, plantedShard)
}

func structureLayoutShared(t *testing.T) *structureLayoutState {
	structureLayoutOnce.Do(func() {
		started := time.Now()
		deadline := structureLayoutDeadline(t, "TestMarkdownStructureLayout (setup)")
		defer deadline.Stop()
		state := structureLayoutEnumeration(t)
		plantedName := state.products.plantedName
		state.products = buildStructureLayoutProducts(t, state.products.root)
		state.products.plantedName = plantedName
		structureLayoutComplete = state
		t.Logf("TestMarkdownStructureLayout (setup): %.3fs", time.Since(started).Seconds())
	})
	if structureLayoutComplete == nil {
		t.Fatal("structure layout setup failed")
	}
	return structureLayoutComplete
}

func runStructureLayoutShard(t *testing.T, index int) {
	parallelMarkdown(t)
	deadline := structureLayoutDeadline(t, t.Name())
	defer deadline.Stop()
	state := structureLayoutShared(t)
	if index < 0 || index >= len(state.shards) {
		t.Fatal("shard outside enumeration")
	}
	fixture := structureLayoutFixture(t, state.shards[index], state.products)
	structureLayoutMutants(t, fixture)
	t.Logf("union shard: %d cases", len(state.shards[index]))
}
func TestMarkdownStructureLayout_000(t *testing.T) { runStructureLayoutShard(t, 0) }
func TestMarkdownStructureLayout_001(t *testing.T) { runStructureLayoutShard(t, 1) }
func TestMarkdownStructureLayout_002(t *testing.T) { runStructureLayoutShard(t, 2) }
func TestMarkdownStructureLayout_003(t *testing.T) { runStructureLayoutShard(t, 3) }
func TestMarkdownStructureLayout_004(t *testing.T) { runStructureLayoutShard(t, 4) }
func TestMarkdownStructureLayout_005(t *testing.T) { runStructureLayoutShard(t, 5) }
func TestMarkdownStructureLayout_006(t *testing.T) { runStructureLayoutShard(t, 6) }
func TestMarkdownStructureLayout_007(t *testing.T) { runStructureLayoutShard(t, 7) }
func TestMarkdownStructureLayout_008(t *testing.T) { runStructureLayoutShard(t, 8) }
func TestMarkdownStructureLayout_009(t *testing.T) { runStructureLayoutShard(t, 9) }
func TestMarkdownStructureLayout_010(t *testing.T) { runStructureLayoutShard(t, 10) }
func TestMarkdownStructureLayout_011(t *testing.T) { runStructureLayoutShard(t, 11) }
func TestMarkdownStructureLayout_012(t *testing.T) { runStructureLayoutShard(t, 12) }
func TestMarkdownStructureLayout_013(t *testing.T) { runStructureLayoutShard(t, 13) }
func TestMarkdownStructureLayout_014(t *testing.T) { runStructureLayoutShard(t, 14) }
func TestMarkdownStructureLayout_015(t *testing.T) { runStructureLayoutShard(t, 15) }

func buildStructureLayoutProducts(t *testing.T, root string) structureLayoutProducts {
	p := structureLayoutProducts{root: root}
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	files := []string{"stage1/cohere/markdownblocks/testdata/list_probe.ts", "stage1/cohere/markdowninline", "internal", "cohere/TypeScript", "cohere/TypeScript-shim", "go.mod"}
	sources, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sources {
		files = append(files, "stage1/cohere/markdownblocks/"+source)
	}
	lowerDir := buildcache.Product(t, buildcache.Inputs{Name: "markdown-structure-lowered", Files: files, Flags: []string{main}, Toolchain: []string{runtime.Version()}}, func(dir string) error {
		program, err := loweredResult(main)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	source, err := os.ReadFile(filepath.Join(lowerDir, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	p.javascript, err = os.ReadFile(filepath.Join(lowerDir, "program.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "darwin" {
		p.program = lowered(t, main)
	}
	options := native.Options{Sanitize: true}
	flags := append(native.Flags(options), native.Flags(native.Options{})...)
	flags = append(flags, fmt.Sprintf("source=%x", sha256.Sum256(source)), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	nativeDir := buildcache.Product(t, buildcache.Inputs{Name: "markdown-structure-native-modes", Files: files, Flags: flags, Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH}}, func(dir string) error {
		var workers sync.WaitGroup
		sanitized := startFixtureTask(&workers, func() (struct{}, error) {
			return struct{}{}, native.Build(string(source), filepath.Join(dir, "sanitized"), native.Options{Sanitize: true})
		})
		release := startFixtureTask(&workers, func() (struct{}, error) {
			return struct{}{}, native.Build(string(source), filepath.Join(dir, "release"), native.Options{})
		})
		workers.Wait()
		if _, err := sanitized.result(); err != nil {
			return err
		}
		_, err := release.result()
		return err
	})
	p.sanitized = filepath.Join(nativeDir, "sanitized")
	p.release = filepath.Join(nativeDir, "release")
	// These products are shared by independent top-level tests, so their lifetime
	// belongs to TestMain rather than whichever test first wins sync.Once.
	dir, err := os.MkdirTemp(artifactDirectory, "structure-go-")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	for _, item := range []struct {
		command, driver string
		bridge          bool
	}{{"adamic_markdown_lists", "list_go.go", true}, {"adamic_markdown_doclayout", "document_go.go", false}} {
		driver, err := filepath.Abs(filepath.Join("testdata", item.driver))
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(cohere, "cmd", item.command, "main.go")
		replace := map[string]string{main: driver}
		if item.bridge {
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
		overlayPath := filepath.Join(dir, item.command+".json")
		write(t, overlayPath, overlay)
		binary := filepath.Join(dir, item.command)
		command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", binary, main)
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			t.Fatalf("Go bridge: %v\n%s", err, output)
		}
		if item.bridge {
			p.goList = binary
		} else {
			p.goLayout = binary
		}
	}
	return p
}

func structureLayoutMutants(t *testing.T, fixture *layoutFixture) {
	main := fixture.main
	mutantFile := "structure.ts"
	mutations := []struct{ name, from, to string }{
		{"heading depth", "'#'.repeat(depth)", "'#'.repeat(depth + 1)"},
		{"sentence separator", "parts.push(child.doc)", "parts.push(arena.text(''))"},
		{"paragraph fill tail", "for(let index = 1; index < node.children.length", "for(let index = 2; index < node.children.length"},
	}
	for _, mutation := range mutations {
		func() {
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
				t.Fatalf("structure mutant %s survived shard", mutation.name)
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Logf("output-only source Node mutant %s caught by %q at byte %d", mutation.name, inputs[index].Name, offset)
		}()
	}
}

func structureLayoutFixture(t *testing.T, inputs []auditInput, products structureLayoutProducts) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	root := products.root
	files := 0
	for _, input := range inputs {
		if !strings.HasPrefix(input.Name, "generated/") {
			files++
		}
	}
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
			mutantWant, err = executeResult(t, nil, goBinary, fullCases, mutantNativeCases, filepath.Join(dir, "mutant-canonical.txt"))
			if err != nil {
				return run{}, err
			}
		}
		return executeResult(t, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.goLayout
	docBuild := startFixtureTask(&workers, func() (struct{}, error) { return struct{}{}, nil })
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
		return executeResult(t, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	program := products.program
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
		if _, err := docBuild.result(); err != nil {
			return run{}, err
		}
		return executeResult(t, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) {
		result, err := onNodeResult(t, main, nativeCases)
		if os.Getenv("ADAMIC_STRUCTURE_LAYOUT_PLANT") == "1" {
			for index, input := range inputs {
				if input.Name == products.plantedName {
					lines := bytes.Split(result.stdout, []byte("\n"))
					lines[index] = append(lines[index], '!')
					result.stdout = bytes.Join(lines, []byte("\n"))
				}
			}
		}
		return result, err
	})
	backendPath := filepath.Join(dir, "program.mjs")
	write(t, backendPath, products.javascript)
	backendTask := startFixtureTask(&workers, func() (run, error) { return onNodeResult(t, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := sanitizedBuild.result()
		if err != nil {
			return run{}, err
		}
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return executeResult(t, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := releaseBuild.result()
		if err != nil {
			return run{}, err
		}
		return executeResult(t, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) { return executeResult(t, nil, "node", script, fork, canonicalCases) })
	goResult := docTask.await(t)
	clean(t, "Go document layout", goResult)
	equal(t, "Go document layout", goResult.stdout, want.stdout)
	answer := nativeTask.await(t)
	binary := sanitizedBuild.await(t)
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

// An independent wall deadline includes silent children and kills the process at
// the slack limit instead of waiting for an over-budget unit to finish.
func structureLayoutDeadline(t *testing.T, name string) *time.Timer {
	return time.AfterFunc(75*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s cooked: over budget at 75s\n", name)
		os.Exit(124)
	})
}
