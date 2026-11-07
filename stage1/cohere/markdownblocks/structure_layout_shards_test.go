package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
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
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
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

func structureLayoutEnumeration(test *testing.T) *structureLayoutState {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := blockCorpus(test, root, "whitespace")
	state := &structureLayoutState{inputs: inputs, shards: make([][]auditInput, testMarkdownStructureLayoutShards)}
	for _, input := range inputs {
		index := structureLayoutShard(input)
		state.shards[index] = append(state.shards[index], input)
	}
	state.products.root = root
	state.products.plantedName = inputs[0].Name
	return state
}

func TestMarkdownStructureLayoutUnion(test *testing.T) {
	test.Parallel()
	configureMarkdownMemory(test)
	deadline := structureLayoutDeadline(test, test.Name())
	defer deadline.Stop()
	state := structureLayoutEnumeration(test)
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
				test.Fatalf("misassigned case %q", input.Name)
			}
			got[key(input)]++
			total++
			if input.Name == state.products.plantedName {
				if bytes.Equal([]byte(input.Text), append([]byte(input.Text), '!')) {
					test.Fatal("planted byte disagreement survived")
				}
				detections++
				plantedShard = index
			}
		}
	}
	if len(state.shards) != testMarkdownStructureLayoutShards || total != len(state.inputs) || len(got) != len(want) {
		test.Fatal("shard union differs from live enumeration")
	}
	for key, count := range want {
		if got[key] != count {
			test.Fatalf("case %q: union count %d, enumeration count %d", key, got[key], count)
		}
	}
	if detections != 1 {
		test.Fatalf("planted failure caught %d times", detections)
	}
	test.Logf(
		"union: %d cases, each exactly once; planted byte disagreement belongs to TestMarkdownStructureLayout_%03d",
		total,
		plantedShard,
	)
}

func structureLayoutShared(test *testing.T) *structureLayoutState {
	structureLayoutOnce.Do(func() {
		started := time.Now()
		// Shared setup has no deadline; Loom bounds the whole unit.
		state := structureLayoutEnumeration(test)
		plantedName := state.products.plantedName
		state.products = buildStructureLayoutProducts(test, state.products.root)
		state.products.plantedName = plantedName
		structureLayoutComplete = state
		test.Logf("TestMarkdownStructureLayout_Setup: %.3fs", time.Since(started).Seconds())
	})
	if structureLayoutComplete == nil {
		test.Fatal("structure layout setup failed")
	}
	return structureLayoutComplete
}

func runStructureLayoutShard(test *testing.T, index int) {
	configureMarkdownMemory(test)
	// Each shard prepares shared products before its own deadline starts.
	state := structureLayoutShared(test)
	deadline := structureLayoutDeadline(test, test.Name())
	defer deadline.Stop()
	if index < 0 || index >= len(state.shards) {
		test.Fatal("shard outside enumeration")
	}
	fixture := structureLayoutFixture(test, state.shards[index], state.products)
	structureLayoutMutants(test, fixture)
	test.Logf("union shard: %d cases", len(state.shards[index]))
}
func TestMarkdownStructureLayout_000(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 0)
}
func TestMarkdownStructureLayout_001(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 1)
}
func TestMarkdownStructureLayout_002(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 2)
}
func TestMarkdownStructureLayout_003(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 3)
}
func TestMarkdownStructureLayout_004(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 4)
}
func TestMarkdownStructureLayout_005(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 5)
}
func TestMarkdownStructureLayout_006(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 6)
}
func TestMarkdownStructureLayout_007(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 7)
}
func TestMarkdownStructureLayout_008(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 8)
}
func TestMarkdownStructureLayout_009(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 9)
}
func TestMarkdownStructureLayout_010(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 10)
}
func TestMarkdownStructureLayout_011(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 11)
}
func TestMarkdownStructureLayout_012(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 12)
}
func TestMarkdownStructureLayout_013(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 13)
}
func TestMarkdownStructureLayout_014(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 14)
}
func TestMarkdownStructureLayout_015(test *testing.T) {
	test.Parallel()
	runStructureLayoutShard(test, 15)
}

func buildStructureLayoutProducts(test *testing.T, root string) structureLayoutProducts {
	return structureLayoutBuildProduct(test, root, "")
}

func structureLayoutBuildProduct(test *testing.T, root, target string) structureLayoutProducts {
	p := structureLayoutProducts{root: root}
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	files := []string{
		"stage1/cohere/markdownblocks/testdata/list_probe.ts",
		"stage1/cohere/markdowninline",
		"internal",
		"cohere/TypeScript",
		"cohere/TypeScript-shim",
		"go.mod",
	}
	sources, globError := filepath.Glob("*.ts")
	if globError != nil {
		test.Fatal(globError)
	}
	for _, source := range sources {
		files = append(files, "stage1/cohere/markdownblocks/"+source)
	}
	lowerDirectory := buildcache.Product(test, buildcache.Inputs{
		Name:      "markdown-structure-lowered",
		Files:     files,
		Flags:     []string{main},
		Toolchain: []string{runtime.Version()},
	}, func(dir string) error {
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
	source, readError := os.ReadFile(filepath.Join(lowerDirectory, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	p.javascript, readError = os.ReadFile(filepath.Join(lowerDirectory, "program.mjs"))
	if readError != nil {
		test.Fatal(readError)
	}
	if target == "lowered" {
		return p
	}
	if runtime.GOOS == "darwin" {
		p.program = lowered(test, main)
	}
	if target == "" || target == "native" {
		options := native.Options{Sanitize: true}
		flags := append(native.Flags(options), native.Flags(native.Options{})...)
		flags = append(
			flags,
			fmt.Sprintf("source=%x", sha256.Sum256(source)),
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		)
		nativeDirectory := buildcache.Product(test, buildcache.Inputs{
			Name:      "markdown-structure-native-modes",
			Files:     files,
			Flags:     flags,
			Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}, func(dir string) error {
			var workers sync.WaitGroup
			sanitized := startFixtureTask(&workers, func() (struct{}, error) {
				return struct{}{}, native.Build(
					string(source),
					filepath.Join(dir, "sanitized"),
					native.Options{Sanitize: true},
				)
			})
			release := startFixtureTask(&workers, func() (struct{}, error) {
				return struct{}{}, native.Build(string(source), filepath.Join(dir, "release"), native.Options{})
			})
			workers.Wait()
			if _, resultError := sanitized.result(); resultError != nil {
				return resultError
			}
			_, resultError := release.result()
			return resultError
		})
		p.sanitized = filepath.Join(nativeDirectory, "sanitized")
		p.release = filepath.Join(nativeDirectory, "release")
	}
	if target == "native" {
		return p
	}
	goProducts := oracles.StructureLayoutGoProducts(test, root, target)
	p.goList, p.goLayout = goProducts.List, goProducts.Document
	return p
}

func structureLayoutMutants(test *testing.T, fixture *layoutFixture) {
	main := fixture.main
	mutantFile := "structure.ts"
	mutations := []struct{ name, from, to string }{
		{"heading depth", "'#'.repeat(depth)", "'#'.repeat(depth + 1)"},
		{"sentence separator", "parts.push(child.doc)", "parts.push(arena.text(''))"},
		{
			"paragraph fill tail",
			"for(let index = 1; index < node.children.length",
			"for(let index = 2; index < node.children.length",
		},
	}
	for _, mutation := range mutations {
		func() {
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
			result := onNode(test, mutantMain, nativeCases)
			clean(test, "source Node layout mutant", result)
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatalf("structure mutant %s survived shard", mutation.name)
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			test.Logf(
				"output-only source Node mutant %s caught by %q at byte %d",
				mutation.name,
				inputs[index].Name,
				offset,
			)
		}()
	}
}

func structureLayoutFixture(test *testing.T, inputs []auditInput, products structureLayoutProducts) *layoutFixture {
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
			var executeResultError error
			mutantWant, executeResultError = executeResult(
				test,
				nil,
				goBinary,
				fullCases,
				mutantNativeCases,
				filepath.Join(dir, "mutant-canonical.txt"),
			)
			if executeResultError != nil {
				return run{}, executeResultError
			}
		}
		return executeResult(test, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.goLayout
	docBuild := startFixtureTask(&workers, func() (struct{}, error) { return struct{}{}, nil })
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
		return executeResult(test, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	program := products.program
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
		if _, resultError := docBuild.result(); resultError != nil {
			return run{}, resultError
		}
		return executeResult(test, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) {
		result, onNodeResultError := onNodeResult(test, main, nativeCases)
		if os.Getenv("ADAMIC_STRUCTURE_LAYOUT_PLANT") == "1" {
			for index, input := range inputs {
				if input.Name == products.plantedName {
					lines := bytes.Split(result.stdout, []byte("\n"))
					lines[index] = append(lines[index], '!')
					result.stdout = bytes.Join(lines, []byte("\n"))
				}
			}
		}
		return result, onNodeResultError
	})
	backendPath := filepath.Join(dir, "program.mjs")
	write(test, backendPath, products.javascript)
	backendTask := startFixtureTask(
		&workers,
		func() (run, error) { return onNodeResult(test, backendPath, nativeCases) },
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
		return executeResult(test, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, resultError := releaseBuild.result()
		if resultError != nil {
			return run{}, resultError
		}
		return executeResult(test, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(
		&workers,
		func() (run, error) { return executeResult(test, nil, "node", script, fork, canonicalCases) },
	)
	goResult := docTask.await(test)
	clean(test, "Go document layout", goResult)
	equal(test, "Go document layout", goResult.stdout, want.stdout)
	answer := nativeTask.await(test)
	binary := sanitizedBuild.await(test)
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
	if report := leaks(test, program, binary, nativeCases); report != "" {
		test.Fatal(report)
	}
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

// An independent wall deadline includes silent children and kills the process at
// the slack limit instead of waiting for an over-budget unit to finish.
func structureLayoutDeadline(test *testing.T, name string) *time.Timer {
	return time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s cooked: over budget at 90s\n", name)
		os.Exit(124)
	})
}

// Setup has no deadline; cancellation still kills the compiler process group.
func structureLayoutCommand(name string, arguments ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := markdownLayoutSetupContext(context.Background())
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command, cancel
}
