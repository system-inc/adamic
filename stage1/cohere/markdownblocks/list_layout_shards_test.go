package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
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

const testMarkdownListLayoutShards = 16

type listLayoutProducts struct {
	goList, goLayout, sanitized, release, canary string
	javascript                                   []byte
	mutants                                      []string
}

func listLayoutBucket(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(binary.LittleEndian.Uint64(sum[:8]) % testMarkdownListLayoutShards)
}

var listLayoutOnce sync.Once
var listLayoutShared listLayoutProducts
var listLayoutBuckets [][]auditInput

func listLayoutSetup(t *testing.T) listLayoutProducts {
	t.Helper()
	listLayoutOnce.Do(func() {
		started := time.Now()
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		inputs, _ := blockCorpus(t, root, "whitespace")
		listLayoutBuckets = make([][]auditInput, testMarkdownListLayoutShards)
		for _, input := range inputs {
			shard := listLayoutBucket(input.Name)
			listLayoutBuckets[shard] = append(listLayoutBuckets[shard], input)
		}
		listLayoutShared = prepareListLayoutProducts(t, root)
		t.Logf("TestMarkdownListLayout (setup): %.3fs", time.Since(started).Seconds())
	})
	if listLayoutShared.sanitized == "" {
		t.Fatal("shared layout setup failed")
	}
	return listLayoutShared
}

func TestMarkdownListLayoutUnion(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := blockCorpus(t, root, "whitespace")
	shards := []func(*testing.T){TestMarkdownListLayout_000, TestMarkdownListLayout_001, TestMarkdownListLayout_002, TestMarkdownListLayout_003, TestMarkdownListLayout_004, TestMarkdownListLayout_005, TestMarkdownListLayout_006, TestMarkdownListLayout_007, TestMarkdownListLayout_008, TestMarkdownListLayout_009, TestMarkdownListLayout_010, TestMarkdownListLayout_011, TestMarkdownListLayout_012, TestMarkdownListLayout_013, TestMarkdownListLayout_014, TestMarkdownListLayout_015}
	if len(shards) != testMarkdownListLayoutShards {
		t.Fatalf("enumerated %d shards, want %d", len(shards), testMarkdownListLayoutShards)
	}
	buckets := make([][]int, testMarkdownListLayoutShards)
	for index, input := range inputs {
		shard := listLayoutBucket(input.Name)
		buckets[shard] = append(buckets[shard], index)
	}
	seen := make([]int, len(inputs))
	caught := 0
	planted := len(inputs) / 2
	catcher := -1
	for shard, bucket := range buckets {
		for _, index := range bucket {
			seen[index]++
			got := []byte(inputs[index].Text)
			want := append([]byte(nil), got...)
			if index == planted {
				want = append(want, '!')
			}
			if !bytes.Equal(got, want) {
				caught++
				catcher = shard
			}
		}
	}
	for index, n := range seen {
		if n != 1 {
			t.Fatalf("case %d covered %d times", index, n)
		}
	}
	if caught != 1 || catcher != listLayoutBucket(inputs[planted].Name) {
		t.Fatalf("planted disagreement caught by %d shards", caught)
	}
	t.Logf("union=%d; planted failure %q caught by TestMarkdownListLayout_%03d", len(inputs), inputs[planted].Name, catcher)
}

func runListLayoutShard(t *testing.T, shard int) {
	t.Helper()
	t.Parallel()
	products := listLayoutSetup(t)
	fixture := listLayoutShardFixture(t, products, listLayoutBuckets[shard])
	for i, mutant := range products.mutants {
		result := listLayoutNode(t, mutant, fixture.mutantNativeCases)
		clean(t, "source Node layout mutant", result)
		if i == 0 {
			resultNative := listLayoutExecute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.canary, fixture.mutantNativeCases)
			clean(t, "native layout canary", resultNative)
			equal(t, "edited native canary equals source Node", resultNative.stdout, result.stdout)
		}
		// Shard zero owns the must-fail assertion; all shards retain byte observations.
		if bytes.Equal(result.stdout, fixture.mutantWant) {
			if shard == 0 {
				t.Fatalf("list mutant %d survived", i)
			}
		} else {
			t.Logf("list mutant %d caught", i)
		}
	}
}

func TestMarkdownListLayout_000(t *testing.T) { runListLayoutShard(t, 0) }
func TestMarkdownListLayout_001(t *testing.T) { runListLayoutShard(t, 1) }
func TestMarkdownListLayout_002(t *testing.T) { runListLayoutShard(t, 2) }
func TestMarkdownListLayout_003(t *testing.T) { runListLayoutShard(t, 3) }
func TestMarkdownListLayout_004(t *testing.T) { runListLayoutShard(t, 4) }
func TestMarkdownListLayout_005(t *testing.T) { runListLayoutShard(t, 5) }
func TestMarkdownListLayout_006(t *testing.T) { runListLayoutShard(t, 6) }
func TestMarkdownListLayout_007(t *testing.T) { runListLayoutShard(t, 7) }
func TestMarkdownListLayout_008(t *testing.T) { runListLayoutShard(t, 8) }
func TestMarkdownListLayout_009(t *testing.T) { runListLayoutShard(t, 9) }
func TestMarkdownListLayout_010(t *testing.T) { runListLayoutShard(t, 10) }
func TestMarkdownListLayout_011(t *testing.T) { runListLayoutShard(t, 11) }
func TestMarkdownListLayout_012(t *testing.T) { runListLayoutShard(t, 12) }
func TestMarkdownListLayout_013(t *testing.T) { runListLayoutShard(t, 13) }
func TestMarkdownListLayout_014(t *testing.T) { runListLayoutShard(t, 14) }
func TestMarkdownListLayout_015(t *testing.T) { runListLayoutShard(t, 15) }

func prepareListLayoutProducts(t *testing.T, root string) listLayoutProducts {
	t.Helper()
	var p listLayoutProducts
	dir := filepath.Join(artifactDirectory, "list-layout-shared")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	buildGo := func(name, driver string, bridge bool) string {
		main := filepath.Join(cohere, "cmd", name, "main.go")
		absDriver, err := filepath.Abs(driver)
		if err != nil {
			t.Fatal(err)
		}
		replace := map[string]string{main: absDriver}
		if bridge {
			absBridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = absBridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			t.Fatal(err)
		}
		overlayPath := filepath.Join(dir, name+".json")
		write(t, overlayPath, overlay)
		binary := filepath.Join(dir, name)
		command := listLayoutCommand(t, "go", "build", "-overlay="+overlayPath, "-o", binary, main)
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			t.Fatalf("Go bridge: %v\n%s", err, output)
		}
		return binary
	}
	var workers sync.WaitGroup
	defer workers.Wait()
	goListTask := startFixtureTask(&workers, func() (string, error) { return buildGo("adamic_markdown_lists", "testdata/list_go.go", true), nil })
	goLayoutTask := startFixtureTask(&workers, func() (string, error) {
		return buildGo("adamic_markdown_doclayout", "testdata/document_go.go", false), nil
	})
	files := []string{"internal/load", "internal/lower", "internal/ir", "internal/javascript", "go.mod", "internal/native/runtime"}
	// Hash compiler and port source, excluding unrelated evidence and TypeScript's fixture tree.
	for _, directory := range []string{"stage1/cohere/markdownblocks", "stage1/cohere/markdowninline", "internal/native", "cohere"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			extension := filepath.Ext(path)
			if extension == ".go" || ((directory == "stage1/cohere/markdownblocks" || directory == "stage1/cohere/markdowninline") && extension == ".ts") {
				relative, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				files = append(files, relative)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	nativeFiles := []string{"internal/native/runtime"}
	entries, err := os.ReadDir(filepath.Join(root, "internal/native"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			nativeFiles = append(nativeFiles, filepath.Join("internal/native", entry.Name()))
		}
	}
	emit := func(main, name string) string {
		return buildcache.Product(t, buildcache.Inputs{Name: name, Files: files, Flags: []string{main}, Toolchain: []string{runtime.Version()}}, func(directory string) error {
			program, err := loweredResult(main)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
		})
	}
	buildNative := func(emitted, name string, sanitize bool) string {
		source, err := os.ReadFile(filepath.Join(emitted, "program.c"))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(source)
		product := buildcache.Product(t, buildcache.Inputs{Name: name, Files: nativeFiles, Flags: append(native.Flags(native.Options{Sanitize: sanitize}), fmt.Sprintf("source=%x", sum), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")), Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.Version()}}, func(directory string) error {
			options := native.Options{Sanitize: sanitize}
			library, err := native.RuntimeLibraryForSource("", string(source), options)
			if err != nil {
				return err
			}
			path := filepath.Join(directory, "main.c")
			if err := os.WriteFile(path, source, 0644); err != nil {
				return err
			}
			args := append(native.Flags(options), "-I", filepath.Dir(library), "-o", filepath.Join(directory, "binary"), path)
			args = append(args, native.RuntimeLinkFlags(library)...)
			args = append(args, "-lm")
			output, err := combinedOutput(listLayoutCommand(t, "clang", args...))
			if err != nil {
				return fmt.Errorf("native build: %w\n%s", err, output)
			}
			return nil
		})
		return filepath.Join(product, "binary")
	}
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	baselineTask := startFixtureTask(&workers, func() (listLayoutProducts, error) {
		emitted := emit(main, "markdown-list-layout-lowered")
		data, err := os.ReadFile(filepath.Join(emitted, "program.mjs"))
		if err != nil {
			return listLayoutProducts{}, err
		}
		sanitizedTask := startFixtureTask(&workers, func() (string, error) { return buildNative(emitted, "markdown-list-layout-sanitized", true), nil })
		release := buildNative(emitted, "markdown-list-layout-release", false)
		return listLayoutProducts{javascript: data, sanitized: sanitizedTask.await(t), release: release}, nil
	})
	mutations := []struct{ from, to string }{{"frame.sibling % 2 === 0 ? '- ' : '* '", "frame.sibling % 2 === 0 ? '+ ' : '* '"}, {"'[x] '", "'[X] '"}, {"999999999", "999999998"}}
	for i, mutation := range mutations {
		scratch := filepath.Join(dir, fmt.Sprintf("mutant-%d", i))
		if err := os.MkdirAll(filepath.Join(scratch, "testdata"), 0755); err != nil {
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
			if name == "lists.ts" {
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
		p.mutants = append(p.mutants, mutantMain)
		if i == 0 { // Cache keys name the stable mutation rather than its temporary path.
			// The mutant is a deterministic transformation of the hashed repository files.
			product := buildcache.Product(t, buildcache.Inputs{Name: "markdown-list-layout-canary-lowered", Files: files, Flags: []string{mutation.from, mutation.to}, Toolchain: []string{runtime.Version()}}, func(directory string) error {
				program, err := loweredResult(mutantMain)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0644)
			})
			source, err := os.ReadFile(filepath.Join(product, "program.c"))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(source)
			flags := native.Flags(native.Options{Sanitize: true})
			for index, flag := range flags {
				if flag == "-O1" {
					flags[index] = "-O0"
				}
			}
			built := buildcache.Product(t, buildcache.Inputs{Name: "markdown-list-layout-canary-sanitized", Files: nativeFiles, Flags: append(append([]string{}, flags...), fmt.Sprintf("source=%x", sum)), Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.Version()}}, func(directory string) error {
				library, err := native.RuntimeLibrary("", native.Options{Sanitize: true})
				if err != nil {
					return err
				}
				path := filepath.Join(directory, "main.c")
				if err := os.WriteFile(path, source, 0644); err != nil {
					return err
				}
				arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), "-o", filepath.Join(directory, "binary"), path)
				arguments = append(arguments, native.RuntimeLinkFlags(library)...)
				arguments = append(arguments, "-lm")
				output, err := combinedOutput(listLayoutCommand(t, "clang", arguments...))
				if err != nil {
					return fmt.Errorf("mutant build: %w\n%s", err, output)
				}
				return nil
			})
			p.canary = filepath.Join(built, "binary")
		}
	}
	baseline := baselineTask.await(t)
	p.javascript, p.sanitized, p.release = baseline.javascript, baseline.sanitized, baseline.release
	p.goList, p.goLayout = goListTask.await(t), goLayoutTask.await(t)
	return p
}

func listLayoutShardFixture(t *testing.T, products listLayoutProducts, inputs []auditInput) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	// This last cumulative corpus contains every earlier layout milestone's inputs.

	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	files := 0
	for _, input := range inputs {
		if !strings.HasPrefix(input.Name, "generated/") {
			files++
		}
	}
	fullInputs := inputs
	selection := gatesample.Selection{}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
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
	cohere := filepath.Join(root, "cohere")
	goBinary := products.goList
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {
		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var err error
			mutantWant, err = listLayoutExecuteResult(t, nil, goBinary, fullCases, mutantNativeCases, filepath.Join(dir, "mutant-canonical.txt"))
			if err != nil {
				return run{}, err
			}
		}
		return listLayoutExecuteResult(t, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.goLayout
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}

	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
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
		return listLayoutExecuteResult(t, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
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
		keep = filepath.Join(keep, filepath.Base(t.Name()))
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
		return listLayoutExecuteResult(t, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) { return listLayoutNodeResult(t, main, nativeCases) })
	backendPath := filepath.Join(dir, "program.mjs")
	write(t, backendPath, products.javascript)
	backendTask := startFixtureTask(&workers, func() (run, error) { return listLayoutNodeResult(t, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := sanitizedBuild.result()
		if err != nil {
			return run{}, err
		}
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return listLayoutExecuteResult(t, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := releaseBuild.result()
		if err != nil {
			return run{}, err
		}
		return listLayoutExecuteResult(t, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) { return listLayoutExecuteResult(t, nil, "node", script, fork, canonicalCases) })
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
	switch runtime.GOOS {
	case "linux":
		report := listLayoutExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, nativeCases)
		if report.exitCode != 0 {
			t.Fatalf("leak check: exit %d\n%s", report.exitCode, report.stderr)
		}
	case "darwin":
		report := listLayoutExecute(t, nil, "leaks", "--atExit", "--", products.release, nativeCases)
		if report.exitCode != 0 {
			t.Fatal(string(report.stdout))
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
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
		inputs: inputs, files: files, want: want.stdout,
	}
}

// Each explicit child has a portable deadline; cancellation includes spawned compilers.
func listLayoutCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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

func listLayoutExecuteResult(t *testing.T, environment []string, name string, arguments ...string) (run, error) {
	t.Helper()
	command := listLayoutCommand(t, name, arguments...)
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
func listLayoutExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := listLayoutExecuteResult(t, environment, name, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func listLayoutNodeResult(t *testing.T, path string, arguments ...string) (run, error) {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		return run{}, err
	}
	return listLayoutExecuteResult(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}
func listLayoutNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	result, err := listLayoutNodeResult(t, path, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
