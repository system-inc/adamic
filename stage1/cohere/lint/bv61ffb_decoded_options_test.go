package lint

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testDecodedOptionsAndMutantShards = 6

// The fixture and its six assertions are pinned, rather than a growing corpus.
func decodedOptionsCases() []string {
	return []string{"Node", "emitted JavaScript", "native", "Go count", "mutant Node", "mutant emitted JavaScript"}
}

func decodedOptionsSlice(shard int) []int {
	var indices []int
	for index := range decodedOptionsCases() {
		if index%testDecodedOptionsAndMutantShards == shard {
			indices = append(indices, index)
		}
	}
	return indices
}

func TestDecodedOptionsAndMutant_000(test *testing.T) { test.Parallel(); decodedOptionsShard(test, 0) }
func TestDecodedOptionsAndMutant_001(test *testing.T) { test.Parallel(); decodedOptionsShard(test, 1) }
func TestDecodedOptionsAndMutant_002(test *testing.T) { test.Parallel(); decodedOptionsShard(test, 2) }
func TestDecodedOptionsAndMutant_003(test *testing.T) { test.Parallel(); decodedOptionsShard(test, 3) }
func TestDecodedOptionsAndMutant_004(test *testing.T) { test.Parallel(); decodedOptionsShard(test, 4) }
func TestDecodedOptionsAndMutant_005(test *testing.T) { test.Parallel(); decodedOptionsShard(test, 5) }

func TestDecodedOptionsAndMutantUnion(test *testing.T) {
	test.Parallel()
	// Read actual top-level registrations: a missing, repeated, or misnumbered
	// wrapper must fail even if the slice helper itself still covers every case.
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "bv61ffb_decoded_options_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	registrations := map[int]int{}
	const prefix = "TestDecodedOptionsAndMutant_"
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		suffix := strings.TrimPrefix(function.Name.Name, prefix)
		if suffix == "Setup" {
			continue
		}
		shard, parseError := strconv.Atoi(suffix)
		if parseError != nil ||
			shard < 0 ||
			shard >= testDecodedOptionsAndMutantShards ||
			suffix != fmt.Sprintf("%03d", shard) {
			test.Fatalf("invalid shard name %s", function.Name.Name)
		}
		registrations[shard]++
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "decodedOptionsShard" {
				return true
			}
			calls++
			if len(call.Args) != 2 {
				test.Fatalf("bad wrapper %s", function.Name.Name)
			}
			number, ok := call.Args[1].(*ast.BasicLit)
			if !ok || number.Value != strconv.Itoa(shard) {
				test.Fatalf("wrapper %s runs wrong slice", function.Name.Name)
			}
			return true
		})
		if calls != 1 {
			test.Fatalf("wrapper %s calls runner %d times", function.Name.Name, calls)
		}
	}
	if len(registrations) != testDecodedOptionsAndMutantShards {
		test.Fatalf("registered %d shards, want %d", len(registrations), testDecodedOptionsAndMutantShards)
	}
	seen := make([]int, len(decodedOptionsCases()))
	for shard := range testDecodedOptionsAndMutantShards {
		if registrations[shard] != 1 {
			test.Fatalf("shard %03d registered %d times", shard, registrations[shard])
		}
		for _, index := range decodedOptionsSlice(shard) {
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			test.Fatalf("case %s covered %d times", decodedOptionsCases()[index], count)
		}
	}
	test.Logf("union: %d cases, each exactly once across %d top-level shards", len(seen), len(registrations))
}

// Both the live shards and the planted-failure test use this exact evaluator.
func decodedOptionsCheck(index int, got, want []byte) error {
	name := decodedOptionsCases()[index]
	if index == 3 {
		if !bytes.Equal(got, []byte("0\n")) {
			return fmt.Errorf("JSON catch option did not override the legacy default: %q", got)
		}
	} else if index >= 4 {
		if bytes.Equal(got, want) {
			return fmt.Errorf("ignored decoded-option mutant survived on %s", name)
		}
	} else if diff := difference(got, want); diff != "" {
		return fmt.Errorf("%s: %s", name, diff)
	}
	return nil
}

func TestDecodedOptionsAndMutantPlantedFailure(test *testing.T) {
	test.Parallel()
	planted := 1
	var caught []int
	for shard := range testDecodedOptionsAndMutantShards {
		failures := 0
		for _, index := range decodedOptionsSlice(shard) {
			want := []byte("case catch\n0 findings\n")
			got := append([]byte(nil), want...)
			if index == 3 {
				got = []byte("0\n")
			}
			if index >= 4 {
				got = []byte("case catch\n1 finding\n")
			}
			if decodedOptionsCheckError := decodedOptionsCheck(index, got, want); decodedOptionsCheckError != nil {
				test.Fatalf("unplanted case %s: %v", decodedOptionsCases()[index], decodedOptionsCheckError)
			}
			if index == planted {
				got = []byte("case catch\nplanted disagreement\n")
			}
			if decodedOptionsCheck(index, got, want) != nil {
				failures++
			}
		}
		if failures > 0 {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != planted%testDecodedOptionsAndMutantShards {
		test.Fatalf("planted disagreement caught by shards %v", caught)
	}
	test.Logf("planted case %s caught by exactly shard-%03d", decodedOptionsCases()[planted], caught[0])
}

func decodedOptionsInputs(test *testing.T, name string) buildcache.Inputs {
	test.Helper()
	prepareRegistry(test, ".")
	files := []string{
		"go.mod",
		"go.work",
		"stage1/cohere/lint/bv61ffb_decoded_options_test.go",
		"stage1/cohere/lint/shared_test.go",
		"stage1/cohere/lint/registry",
		"oracle/adamic.mjs",
		"cohere/TypeScript/tsc",
		"cohere/TypeScript-shim",
	}
	for _, file := range portFiles(test) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	// Hash implementation sources, excluding unrelated tests and evidence trees.
	// Reads: internal/native/runtime/*.[ch]
	walkError := filepath.WalkDir(filepath.Join(
		repository,
		"internal",
	), func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() && (entry.Name() == "testdata" || entry.Name() == "performance") {
			return filepath.SkipDir
		}
		extension := filepath.Ext(path)
		if !entry.IsDir() &&
			(extension == ".go" || extension == ".c" || extension == ".h") &&
			!strings.HasSuffix(path, "_test.go") {
			relative, relativePathError := filepath.Rel(repository, path)
			if relativePathError != nil {
				return relativePathError
			}
			files = append(files, filepath.ToSlash(relative))
		}
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	sort.Strings(files)
	return buildcache.Inputs{Name: name, Files: files, Flags: []string{
		"Sanitize=true",
		"Target=native",
		"Split=true",
		"Jobs=4",
		"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		"GOOS=" + runtime.GOOS,
		"GOARCH=" + runtime.GOARCH,
		"source-root=" + packageDirectory,
		"GOFLAGS=" + os.Getenv("GOFLAGS"),
		"GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"),
		"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
	}, Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
}

type decodedOptionsProducts struct {
	Directory, Changed, Module, ChangedModule, Binary, Oracle, Manifest string
	Want                                                                []byte
}

var decodedOptionsReady *decodedOptionsProducts
var decodedOptionsOnce sync.Once

// Retained for TestMain compatibility. Each selected leaf owns preparation.
func decodedOptionsBeforeTests() error { return nil }

func decodedOptionsPrepare(test *testing.T) *decodedOptionsProducts {
	test.Helper()
	decodedOptionsOnce.Do(func() { decodedOptionsSetupTest(test) })
	if decodedOptionsReady == nil {
		test.Fatal("decoded-options product preparation failed")
	}
	return decodedOptionsReady
}

func decodedOptionsSetupTest(test *testing.T) {
	test.Helper()
	started := time.Now()
	cancelContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	path := filepath.Join(sharedDirectory, "decoded-options-setup.json")
	// Both original and mutant source copies live in the content-addressed
	// product, rather than in a setup test's TempDir.
	lowered := decodedOptionsFoundLowered(test, false)
	// Node runs the changed port, so it reads the copy whose imports name this checkout.
	changed, resolvedError := buildcache.Resolved(decodedOptionsFoundLowered(test, true))
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	inputs, build := decodedOptionsNative(test)
	nativeProduct := buildcache.BuiltElsewhere(
		test,
		"decoded options' native scanner",
		inputs,
		decodedOptionsNativeOwner,
		build,
	)
	// The fixture and manifest belong to TestMain's directory and remain valid
	// after the shard that prepares them finishes, until the process exits.
	fixture := filepath.Join(filepath.Dir(path), "catch.ts")
	if writeError := os.WriteFile(fixture, []byte("try { work(); } catch(e) {}\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	manifestPath := filepath.Join(filepath.Dir(path), "decoded-options-manifest.txt")
	row := fixture + "\tno-empty\t\t\tfalse\t{\"AllowEmptyCatch\":true}\n"
	if writeError := os.WriteFile(manifestPath, []byte(row), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	ready := decodedOptionsProducts{
		Directory: packageDirectory, Changed: filepath.Join(changed, "port"),
		Module: filepath.Join(lowered, "lint.mjs"), ChangedModule: filepath.Join(changed, "lint.mjs"),
		Binary: filepath.Join(nativeProduct, "scanner"), Oracle: lintGoOracleProduct(test), Manifest: manifestPath,
	}
	ready.Want = decodedOptionsExecute(test, cancelContext, ready.Oracle, "--manifest", manifestPath)
	decodedOptionsReady = &ready
	test.Logf("decoded-options shared preparation: %s; own-work deadline starts afterward", time.Since(started))
}

// Each product here has one owner, the product test that builds it, and every other reader finds it by name, after its
// owner in cloud/fast-gate/product-order.json. A product two tests each build is settled twice in one tree, under two
// read sets, and the tree keeps only the newer: the native scanner, built where the decoded-options preparation also
// ran the Go oracle, voided block 165 that way.
const (
	decodedOptionsLoweredOwner       = "stage1/cohere/lint TestProduct_DecodedOptionsLowered"
	decodedOptionsMutantLoweredOwner = "stage1/cohere/lint TestProduct_DecodedOptionsMutantLowered"
	decodedOptionsNativeOwner        = "stage1/cohere/lint TestProduct_DecodedOptionsNative"
)

// decodedOptionsNativeProduct is the sanitized native scanner, built from the lowered port it finds by name. Its owner,
// TestProduct_DecodedOptionsNative, builds it and nothing else: no lowering, no oracle run.
func decodedOptionsNativeProduct(test *testing.T) string {
	test.Helper()
	inputs, build := decodedOptionsNative(test)
	return buildcache.Product(test, inputs, build)
}

func decodedOptionsNative(test *testing.T) (buildcache.Inputs, func(output string) error) {
	test.Helper()
	lowered := decodedOptionsFoundLowered(test, false)
	inputs := decodedOptionsInputs(test, "lint-decoded-options-native-setup-v2")
	inputs.Flags = append(inputs.Flags, "registry=no-empty")
	return inputs, func(output string) error {
		source, readError := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if readError != nil {
			return readError
		}
		return native.Build(
			string(source),
			filepath.Join(output, "scanner"),
			native.Options{Sanitize: true, Split: true, Jobs: 4},
		)
	}
}

// decodedOptionsSetupLowered builds a lowered port, the original or the mutant: only its owner calls it.
func decodedOptionsSetupLowered(test *testing.T, mutated bool) string {
	test.Helper()
	inputs, build := decodedOptionsLowered(test, mutated)
	return buildcache.Product(test, inputs, build)
}

// decodedOptionsFoundLowered is a lowered port for every reader but its owner.
func decodedOptionsFoundLowered(test *testing.T, mutated bool) string {
	test.Helper()
	inputs, build := decodedOptionsLowered(test, mutated)
	what, owner := "decoded options' lowered port", decodedOptionsLoweredOwner
	if mutated {
		what, owner = "decoded options' lowered mutant port", decodedOptionsMutantLoweredOwner
	}
	return buildcache.BuiltElsewhere(test, what, inputs, owner, build)
}

func decodedOptionsLowered(test *testing.T, mutated bool) (buildcache.Inputs, func(output string) error) {
	test.Helper()
	name := "lint-decoded-options-lowered-setup-v2"
	if mutated {
		name += "-ignored-allowemptycatch"
	}
	inputs := decodedOptionsInputs(test, name)
	inputs.Flags = append(inputs.Flags, "registry=no-empty")
	return inputs, func(output string) error {
		from, to := "", ""
		if mutated {
			from, to = "'allowemptycatch'", "'ignored-allowemptycatch'"
		}
		directory := copyPort(test, filepath.Join(output, "port"), from, to, "main.ts")
		// Keep the original main/options decoder and no-empty implementation. Only
		// unused factories are omitted; Node and the Go oracle retain the full registry.
		entries, readDirectoryError := os.ReadDir(filepath.Join(directory, "rules"))
		if readDirectoryError != nil {
			return readDirectoryError
		}
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "no-empty" {
				if removeError := os.RemoveAll(filepath.Join(directory, "rules", entry.Name())); removeError != nil {
					return removeError
				}
			}
		}
		prepareRegistry(test, directory)
		program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
		if loadError != nil {
			return loadError
		}
		lowered, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(
			filepath.Join(output, "lint.c"),
			[]byte(native.C(lowered)),
			0644,
		); writeError != nil {
			return writeError
		}
		if writeError := os.WriteFile(
			filepath.Join(output, "lint.mjs"),
			[]byte(javascript.JavaScript(lowered)),
			0644,
		); writeError != nil {
			return writeError
		}
		// The port's imports name this checkout for the build's own lowering; the product names it <repository>
		// (#tqrqx60).
		return buildcache.RelativeFiles(output)
	}
}

func decodedOptionsExecute(test *testing.T, callerContext context.Context, name string, args ...string) []byte {
	test.Helper()
	command := lintCommand(callerContext, "", name, args...)
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = output, &stderr
	runError := command.Run()
	if callerContext.Err() != nil {
		test.Fatalf("cooked: shard exceeded 90s: %v", callerContext.Err())
	}
	if runError != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return data
}

func decodedOptionsShard(test *testing.T, shard int) {
	test.Helper()
	products := decodedOptionsPrepare(test)
	// No build, registry generation, oracle preparation, or lock acquisition is
	// allowed below this boundary. Each leaf owns only its case deadline.
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, index := range decodedOptionsSlice(shard) {
		var got []byte
		switch index {
		case 0:
			got = decodedOptionsExecute(
				test,
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(products.Directory, "main.ts"),
				"--manifest",
				products.Manifest,
			)
		case 1:
			got = decodedOptionsExecute(
				test,
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				products.Module,
				"--manifest",
				products.Manifest,
			)
		case 2:
			got = decodedOptionsExecute(test, timeoutContext, products.Binary, "--manifest", products.Manifest)
		case 3:
			got = decodedOptionsExecute(
				test,
				timeoutContext,
				products.Oracle,
				"--manifest",
				products.Manifest,
				"--count",
			)
		case 4:
			got = decodedOptionsExecute(
				test,
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(products.Changed, "main.ts"),
				"--manifest",
				products.Manifest,
			)
		case 5:
			got = decodedOptionsExecute(
				test,
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				products.ChangedModule,
				"--manifest",
				products.Manifest,
			)
		default:
			test.Fatalf("unhandled case %d", index)
		}
		if decodedOptionsCheckError := decodedOptionsCheck(index, got, products.Want); decodedOptionsCheckError != nil {
			test.Fatal(decodedOptionsCheckError)
		}
		if index >= 4 {
			test.Logf(
				"ignored decoded-option mutant caught on %s: %s",
				decodedOptionsCases()[index],
				difference(got, products.Want),
			)
		}
	}
	elapsed := time.Since(started)
	test.Logf("shard-%03d cases only: %.6fs; cooked=%t", shard, elapsed.Seconds(), elapsed >= 60*time.Second)
	if elapsed >= 60*time.Second {
		test.Fatal("cooked: shard exceeded 60s budget")
	}
}

// Optional preparation probe. No shard requires this test to run first.
func TestDecodedOptionsAndMutant_Setup(test *testing.T) {
	test.Parallel()
	decodedOptionsPrepare(test)
}
