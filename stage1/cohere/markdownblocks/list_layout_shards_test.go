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
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/tracked"
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

var listLayoutBuildOnce sync.Once
var listLayoutOnce sync.Once
var listLayoutShared listLayoutProducts
var listLayoutBuckets [][]auditInput
var listLayoutContexts sync.Map

type listLayoutManifest struct {
	Paths      []string
	JavaScript []byte
	Mutants    []string
}

func listLayoutSetupInputs(test *testing.T) buildcache.Inputs {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	files := append([]string{
		"go.mod",
		"internal/load",
		"internal/lower",
		"internal/ir",
		"internal/javascript",
		"internal/native/runtime",
	}, listLayoutSources(test, root, false)...)
	return buildcache.Inputs{Name: "markdown-list-layout-shared-v4", Files: files, Flags: []string{
		"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
	}, Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
}

// listLayoutSourceRoots are the directories whose sources key the list layout's products, each with the repository
// that tracks it (cohere is a submodule, its own repository) and its root inside that repository.
var listLayoutSourceRoots = []struct{ repository, root string }{
	{".", "stage1/cohere/markdownblocks"},
	{".", "stage1/cohere/markdowninline"},
	{".", "internal/native"},
	{"cohere", ""},
}

// listLayoutSources is the compiler and port sources that key the list layout's products, repository-relative: every
// tracked .go file under listLayoutSourceRoots (tests too when withTests), and the .ts files of markdownblocks and
// markdowninline. The listing is git's, run as key-time work (buildcache.Keying), never a walk of the checkout: walked
// under the trace, every directory under cohere and the ports entered the product's read set, untracked ones too, and
// a listing the tree's checkout and the key check's assembled source valued differently moved the key, so Adam's key
// check refused markdownblocks and voided its shards (#kfjqsnh, Oct 19; tsprinter's corpus had the same shape,
// #0rx19ha).
func listLayoutSources(test *testing.T, root string, withTests bool) []string {
	test.Helper()
	var sources []string
	for _, source := range listLayoutSourceRoots {
		var paths []string
		var err error
		buildcache.Keying(
			func() { paths, err = tracked.PathsUnder(filepath.Join(root, source.repository), source.root) },
		)
		if err != nil {
			test.Fatal(err)
		}
		ports := source.root == "stage1/cohere/markdownblocks" || source.root == "stage1/cohere/markdowninline"
		for _, path := range paths {
			goSource := strings.HasSuffix(path, ".go") && (withTests || !strings.HasSuffix(path, "_test.go"))
			if goSource || (ports && strings.HasSuffix(path, ".ts")) {
				sources = append(sources, filepath.ToSlash(filepath.Join(source.repository, path)))
			}
		}
	}
	return sources
}

func buildListLayoutSetup(test *testing.T) {
	test.Helper()
	listLayoutBuildOnce.Do(func() {
		// Setup carries no deadline of its own: it is the build phase's to make fast, and Loom's 90 s kill
		// still covers the unit (@system_adamic's ruling). Shards start their own deadline after it.
		started := time.Now()
		defer func() { test.Logf("Markdown list layout setup: %.3fs", time.Since(started).Seconds()) }()
		// Reads: internal/native/*.go
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		directory := buildcache.Product(test, listLayoutSetupInputs(test), func(directory string) error {
			products := prepareListLayoutProducts(test, root, directory)
			mutants := make([]string, len(products.mutants))
			for index, path := range products.mutants {
				relative, relativePathError := filepath.Rel(directory, path)
				if relativePathError != nil {
					return relativePathError
				}
				mutants[index] = relative
			}
			manifest := listLayoutManifest{Paths: []string{
				products.goList,
				products.goLayout,
				products.sanitized,
				products.release,
				products.canary,
			}, JavaScript: products.javascript, Mutants: mutants}
			// Go binaries are stored inside this product; native products are named by their keys, which a runner finds
			// in its own cache (#tqrqx60).
			for index := 0; index < 2; index++ {
				relative, relativePathError := filepath.Rel(directory, manifest.Paths[index])
				if relativePathError != nil {
					return relativePathError
				}
				manifest.Paths[index] = relative
			}
			for index := 2; index < len(manifest.Paths); index++ {
				manifest.Paths[index] = buildcache.Relative(manifest.Paths[index])
			}
			data, encodeError := json.Marshal(manifest)
			if encodeError != nil {
				return encodeError
			}
			return os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0644)
		})
		listLayoutOnce.Do(func() { loadListLayoutSetup(test, directory) })
	})
}

// Not parallel: publish the shared build product before parallel corpus shards resume.
func TestMarkdownListLayout_Setup(test *testing.T) { buildListLayoutSetup(test) }

// TestProduct_MarkdownListLayout builds the shared products every list layout shard reads (238 s inside
// TestMarkdownListLayout_Setup on Oct 10: the lowered program, its sanitized and release natives, the canary and the
// Go oracles), so Workshop builds them once and the shards and Setup only read them.
func TestProduct_MarkdownListLayout(test *testing.T) { test.Parallel(); buildListLayoutSetup(test) }

// A shard prepares the shared products itself, once per process: it fetches them when they are built
// and builds them on a miss, so it never needs another top-level test to have run first.
func listLayoutSetup(test *testing.T) listLayoutProducts {
	test.Helper()
	buildListLayoutSetup(test)
	if listLayoutShared.sanitized == "" {
		test.Fatal("shared setup unavailable")
	}
	return listLayoutShared
}

// One deadline spans the entire unit; every child inherits its remaining time.
func listLayoutDeadline(test *testing.T, unit string) func() {
	test.Helper()
	started := time.Now()
	timeoutContext, cancel := context.WithTimeout(test.Context(), 90*time.Second)
	listLayoutContexts.Store(test, timeoutContext)
	timer := time.AfterFunc(
		90*time.Second,
		func() { cancel(); panic("cooked: Markdown list layout " + unit + " exceeded 90s") },
	)
	return func() {
		timer.Stop()
		cancel()
		listLayoutContexts.Delete(test)
		test.Logf("Markdown list layout %s: %.3fs", unit, time.Since(started).Seconds())
	}
}

func TestMarkdownListLayoutUnion(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := listLayoutInputs(test, root)
	shards := []func(*testing.T){
		TestMarkdownListLayout_000,
		TestMarkdownListLayout_001,
		TestMarkdownListLayout_002,
		TestMarkdownListLayout_003,
		TestMarkdownListLayout_004,
		TestMarkdownListLayout_005,
		TestMarkdownListLayout_006,
		TestMarkdownListLayout_007,
		TestMarkdownListLayout_008,
		TestMarkdownListLayout_009,
		TestMarkdownListLayout_010,
		TestMarkdownListLayout_011,
		TestMarkdownListLayout_012,
		TestMarkdownListLayout_013,
		TestMarkdownListLayout_014,
		TestMarkdownListLayout_015,
	}
	if len(shards) != testMarkdownListLayoutShards {
		test.Fatalf("enumerated %d shards, want %d", len(shards), testMarkdownListLayoutShards)
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
			test.Fatalf("case %d covered %d times", index, n)
		}
	}
	if caught != 1 || catcher != listLayoutBucket(inputs[planted].Name) {
		test.Fatalf("planted disagreement caught by %d shards", caught)
	}
	test.Logf(
		"union=%d; planted failure %q caught by TestMarkdownListLayout_%03d",
		len(inputs),
		inputs[planted].Name,
		catcher,
	)
}

func runListLayoutShard(test *testing.T, shard int) {
	test.Helper()
	products := listLayoutSetup(test)
	finish := listLayoutDeadline(test, fmt.Sprintf("shard-%03d", shard))
	defer finish()
	fixture := listLayoutShardFixture(test, products, listLayoutBuckets[shard])
	for index, mutant := range products.mutants {
		result := listLayoutNode(test, mutant, fixture.mutantNativeCases)
		clean(test, "source Node layout mutant", result)
		if index == 0 {
			resultNative := listLayoutExecute(
				test,
				[]string{"ASAN_OPTIONS=detect_leaks=0"},
				products.canary,
				fixture.mutantNativeCases,
			)
			clean(test, "native layout canary", resultNative)
			equal(test, "edited native canary equals source Node", resultNative.stdout, result.stdout)
		}
		// The shard containing a fixed witness owns each mutant must-fail assertion.
		if bytes.Equal(result.stdout, fixture.mutantWant) {
			if shard == listLayoutBucket(listLayoutMutantControls[index].Name) {
				test.Fatalf("list mutant %d survived", index)
			}
		} else {
			test.Logf("list mutant %d caught", index)
		}
	}
}

func TestMarkdownListLayout_000(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 0)
}
func TestMarkdownListLayout_001(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 1)
}
func TestMarkdownListLayout_002(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 2)
}
func TestMarkdownListLayout_003(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 3)
}
func TestMarkdownListLayout_004(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 4)
}
func TestMarkdownListLayout_005(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 5)
}
func TestMarkdownListLayout_006(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 6)
}
func TestMarkdownListLayout_007(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 7)
}
func TestMarkdownListLayout_008(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 8)
}
func TestMarkdownListLayout_009(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 9)
}
func TestMarkdownListLayout_010(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 10)
}
func TestMarkdownListLayout_011(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 11)
}
func TestMarkdownListLayout_012(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 12)
}
func TestMarkdownListLayout_013(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 13)
}
func TestMarkdownListLayout_014(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 14)
}
func TestMarkdownListLayout_015(test *testing.T) {
	test.Parallel()
	runListLayoutShard(test, 15)
}

func prepareListLayoutProducts(test *testing.T, root, dir string) listLayoutProducts {
	test.Helper()
	var products listLayoutProducts
	if createDirectoryError := os.MkdirAll(dir, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	cohere := filepath.Join(root, "cohere")
	buildGo := func(name, driver string, bridge bool) string {
		main := filepath.Join(cohere, "cmd", name, "main.go")
		absDriver, absolutePathError := filepath.Abs(driver)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		replace := map[string]string{main: absDriver}
		if bridge {
			absBridge, absolutePathError := filepath.Abs("testdata/list_bridge.go")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = absBridge
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		overlayPath := filepath.Join(dir, name+".json")
		write(test, overlayPath, overlay)
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		binary := filepath.Join(dir, name)
		command, ctx := listLayoutCommand(test, "go", append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-overlay="+overlayPath,
			"-o",
			binary,
			main,
		)...)
		command.Dir = cohere
		// Read through childguard.Capture for listLayoutExecuteResult's reason. go build writes nothing until it
		// fails, so of childguard.Run's windows only the 30-minute first-output one could have fired, and the unit's
		// own ceiling ends a hung build sooner.
		if stdout, stderr, captureError := childguard.Capture(ctx, command); captureError != nil {
			test.Fatalf("Go bridge: %v\n%s%s", captureError, stdout, stderr)
		}
		return binary
	}
	// Every build here runs one after another, never two at once: each product's traced window holds the whole
	// process's reads, so a build running beside it puts its reads into that product's set.
	// markdown-list-layout-lowered and -canary-lowered, built at once, each recorded a set reading the other, a cycle a
	// runner's lookup met and memoized as a miss: Adam's key check refused markdownblocks and voided its shards
	// (#kfjqsnh, Oct 19).
	products.goList = buildGo("adamic_markdown_lists", "testdata/list_go.go", true)
	products.goLayout = buildGo("adamic_markdown_doclayout", "testdata/document_go.go", false)
	files := append([]string{
		"internal/load",
		"internal/lower",
		"internal/ir",
		"internal/javascript",
		"go.mod",
		"internal/native/runtime",
	}, listLayoutSources(test, root, true)...)
	nativeFiles := []string{"internal/native/runtime"}
	entries, readDirectoryError := os.ReadDir(filepath.Join(root, "internal/native"))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			nativeFiles = append(nativeFiles, filepath.Join("internal/native", entry.Name()))
		}
	}
	emit := func(main, name string) string {
		return buildcache.Product(test, buildcache.Inputs{
			Name:      name,
			Files:     files,
			Flags:     []string{main},
			Toolchain: []string{runtime.Version()},
		}, func(directory string) error {
			program, loweredResultError := loweredResult(main)
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
	}
	buildNative := func(emitted, name string, sanitize bool) string {
		source, readError := os.ReadFile(filepath.Join(emitted, "program.c"))
		if readError != nil {
			test.Fatal(readError)
		}
		sum := sha256.Sum256(source)
		// v3: built alone (#kfjqsnh), so a v2 set, which may hold a build that ran beside it, isn't one this reads; v2
		// was native.Build's, whose binary names no build directory, unlike v1's.
		product := buildcache.Product(test, buildcache.Inputs{
			Name:  name + "-v3",
			Files: nativeFiles,
			Flags: append(
				native.Flags(native.Options{Sanitize: sanitize}),
				fmt.Sprintf("source=%x", sum),
				"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
			),
			Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.Version()},
		}, func(directory string) error {
			if writeError := os.WriteFile(filepath.Join(directory, "main.c"), source, 0644); writeError != nil {
				return writeError
			}
			// native.Build compiles in a directory of its own and gives clang no absolute path, so the binary names no
			// build directory: given <build directory>/main.c, a sanitized binary kept it in its globals' metadata,
			// random per build (landable-14b's tree, Oct 11: markdown-list-layout-sanitized and -canary-sanitized).
			if buildError := native.Build(
				string(source),
				filepath.Join(directory, "binary"),
				native.Options{Sanitize: sanitize},
			); buildError != nil {
				return fmt.Errorf("native build: %w", buildError)
			}
			return nil
		})
		return filepath.Join(product, "binary")
	}
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	emitted := emit(main, "markdown-list-layout-lowered-v2")
	script, readError := os.ReadFile(filepath.Join(emitted, "program.mjs"))
	if readError != nil {
		test.Fatal(readError)
	}
	products.javascript = script
	products.sanitized = buildNative(emitted, "markdown-list-layout-sanitized", true)
	products.release = buildNative(emitted, "markdown-list-layout-release", false)
	mutations := []struct{ from, to string }{
		{"frame.sibling % 2 === 0 ? '- ' : '* '", "frame.sibling % 2 === 0 ? '+ ' : '* '"},
		{"'[x] '", "'[X] '"},
		{"999999999", "999999998"},
	}
	for index, mutation := range mutations {
		scratch := filepath.Join(dir, fmt.Sprintf("mutant-%d", index))
		if createDirectoryError := os.MkdirAll(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
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
			if name == "lists.ts" {
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
		products.mutants = append(products.mutants, mutantMain)
		if index == 0 { // Cache keys name the stable mutation rather than its temporary path.
			// The mutant is a deterministic transformation of the hashed repository files.
			product := buildcache.Product(test, buildcache.Inputs{
				Name:      "markdown-list-layout-canary-lowered-v2",
				Files:     files,
				Flags:     []string{mutation.from, mutation.to},
				Toolchain: []string{runtime.Version()},
			}, func(directory string) error {
				program, loweredResultError := loweredResult(mutantMain)
				if loweredResultError != nil {
					return loweredResultError
				}
				return os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0644)
			})
			source, readError := os.ReadFile(filepath.Join(product, "program.c"))
			if readError != nil {
				test.Fatal(readError)
			}
			sum := sha256.Sum256(source)
			flags := native.Flags(native.Options{Sanitize: true})
			for index, flag := range flags {
				if flag == "-O1" {
					flags[index] = "-O0"
				}
			}
			// v3: built alone (#kfjqsnh); v2 was native.BuildFlags', which names no build directory, unlike v1's.
			built := buildcache.Product(test, buildcache.Inputs{
				Name:      "markdown-list-layout-canary-sanitized-v3",
				Files:     nativeFiles,
				Flags:     append(append([]string{}, flags...), fmt.Sprintf("source=%x", sum)),
				Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.Version()},
			}, func(directory string) error {
				if writeError := os.WriteFile(filepath.Join(directory, "main.c"), source, 0644); writeError != nil {
					return writeError
				}
				if buildFlagsError := native.BuildFlags(
					string(source),
					filepath.Join(directory, "binary"),
					native.Options{Sanitize: true},
					flags,
				); buildFlagsError != nil {
					return fmt.Errorf("mutant build: %w", buildFlagsError)
				}
				return nil
			})
			products.canary = filepath.Join(built, "binary")
		}
	}
	return products
}

func listLayoutShardFixture(test *testing.T, products listLayoutProducts, inputs []auditInput) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	// This last cumulative corpus contains every earlier layout milestone's inputs.

	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
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
	cohere := filepath.Join(root, "cohere")
	goBinary := products.goList
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {
		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var listLayoutExecuteResultError error
			mutantWant, listLayoutExecuteResultError = listLayoutExecuteResult(
				test,
				nil,
				goBinary,
				fullCases,
				mutantNativeCases,
				filepath.Join(dir, "mutant-canonical.txt"),
			)
			if listLayoutExecuteResultError != nil {
				return run{}, listLayoutExecuteResultError
			}
		}
		return listLayoutExecuteResult(test, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	goLayout := products.goLayout
	main, absolutePathError := filepath.Abs("testdata/list_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}

	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
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
		return listLayoutExecuteResult(test, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
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
		keep = filepath.Join(keep, filepath.Base(test.Name()))
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
		return listLayoutExecuteResult(test, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(
		&workers,
		func() (run, error) { return listLayoutNodeResult(test, main, nativeCases) },
	)
	backendPath := filepath.Join(dir, "program.mjs")
	write(test, backendPath, products.javascript)
	backendTask := startFixtureTask(
		&workers,
		func() (run, error) { return listLayoutNodeResult(test, backendPath, nativeCases) },
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
		return listLayoutExecuteResult(test, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, resultError := releaseBuild.result()
		if resultError != nil {
			return run{}, resultError
		}
		return listLayoutExecuteResult(test, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(
		&workers,
		func() (run, error) { return listLayoutExecuteResult(test, nil, "node", script, fork, canonicalCases) },
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
	switch runtime.GOOS {
	case "linux":
		report := listLayoutExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, nativeCases)
		if report.exitCode != 0 {
			test.Fatalf("leak check: exit %d\n%s", report.exitCode, report.stderr)
		}
	case "darwin":
		report := listLayoutExecute(test, nil, "leaks", "--atExit", "--", products.release, nativeCases)
		if report.exitCode != 0 {
			test.Fatal(string(report.stdout))
		}
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
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
	}
}

// A child started inside a shard inherits the shard's deadline; a setup child has none, since setup is
// off the clock. Cancellation includes spawned compilers either way. It returns the command's context with it, for
// childguard.Capture to end the reads by.
func listLayoutCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	ctx := test.Context()
	if deadline, ok := listLayoutContexts.Load(test); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(deadline.(context.Context), 90*time.Second)
		test.Cleanup(cancel)
	}
	command := exec.CommandContext(ctx, name, arguments...)
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
	return command, ctx
}

// listLayoutExecuteResult reads a run's output through childguard.Capture, to its end with no timer deciding when:
// exec's own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a
// second after the child exited, so a run that had finished could fail its shard under load (#n3gk22h's shape).
func listLayoutExecuteResult(test *testing.T, environment []string, name string, arguments ...string) (run, error) {
	test.Helper()
	command, ctx := listLayoutCommand(test, name, arguments...)
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
func listLayoutExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	result, listLayoutExecuteResultError := listLayoutExecuteResult(test, environment, name, arguments...)
	if listLayoutExecuteResultError != nil {
		test.Fatal(listLayoutExecuteResultError)
	}
	return result
}
func listLayoutNodeResult(test *testing.T, path string, arguments ...string) (run, error) {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return listLayoutExecuteResult(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}
func listLayoutNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	result, listLayoutNodeResultError := listLayoutNodeResult(test, path, arguments...)
	if listLayoutNodeResultError != nil {
		test.Fatal(listLayoutNodeResultError)
	}
	return result
}

func loadListLayoutSetup(test *testing.T, directory string) {
	test.Helper()
	data, readError := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var manifest listLayoutManifest
	if decodeError := json.Unmarshal(data, &manifest); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(manifest.Paths) != 5 {
		test.Fatal("incomplete shared setup manifest")
	}
	listLayoutShared = listLayoutProducts{
		goList:     filepath.Join(directory, manifest.Paths[0]),
		goLayout:   filepath.Join(directory, manifest.Paths[1]),
		sanitized:  buildcache.Absolute(manifest.Paths[2]),
		release:    buildcache.Absolute(manifest.Paths[3]),
		canary:     buildcache.Absolute(manifest.Paths[4]),
		javascript: manifest.JavaScript,
	}
	for _, path := range manifest.Mutants {
		listLayoutShared.mutants = append(listLayoutShared.mutants, filepath.Join(directory, path))
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := listLayoutInputs(test, root)
	listLayoutBuckets = make([][]auditInput, testMarkdownListLayoutShards)
	for _, input := range inputs {
		shard := listLayoutBucket(input.Name)
		listLayoutBuckets[shard] = append(listLayoutBuckets[shard], input)
	}
}

var listLayoutMutantControls = []auditInput{
	{Name: "generated/list-layout/mutant-control/unordered", Text: "- a\n- b\n"},
	{Name: "generated/list-layout/mutant-control/task", Text: "- [x] a\n"},
	{Name: "generated/list-layout/mutant-control/cap", Text: "999999999. a\n999999999. b\n"},
}

func listLayoutInputs(test *testing.T, root string) ([]auditInput, int) {
	test.Helper()
	inputs, files := blockCorpus(test, root, "whitespace")
	return append(inputs, listLayoutMutantControls...), files
}
