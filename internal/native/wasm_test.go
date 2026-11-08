package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: this opt-in integration probe builds one complete runtime and runs a request
// benchmark. Native tests keep the native clang; every WASI build here runs wasiCompiler, the SDK's.
var wasiFixtures = []string{
	"internal/load/testdata/0.1/compile/01_hello.ts",
	"internal/load/testdata/0.1/compile/02_fizzbuzz.ts",
	"internal/load/testdata/0.1/compile/03_shapes.ts",
	"internal/load/testdata/0.1/compile/04_closures.ts",
	"internal/load/testdata/0.1/compile/05_wordcount.ts",
	"internal/load/testdata/0.1/compile/06_stack.ts",
	"internal/load/testdata/0.1/compile/07_modules/main.ts",
	"internal/load/testdata/0.1/compile/08_results.ts",
	"internal/load/testdata/0.1/compile/09_tree.ts",
	"internal/load/testdata/0.1/compile/10_unicode.ts",
	"internal/oracle/testdata/strings.a", "internal/oracle/testdata/numbers.a",
	"internal/oracle/testdata/bitwise_sweep.a", "internal/oracle/testdata/functions.a",
	"internal/oracle/testdata/regions.a", "internal/oracle/testdata/regions_throw.a",
	"internal/oracle/testdata/weak_parent.a", "internal/oracle/testdata/weak_narrowed.a",
	"internal/oracle/testdata/reuse_weak_after_reuse.a", "internal/oracle/testdata/reuse_weak_during_spread.a",
	"internal/oracle/testdata/normalize.a", "internal/oracle/testdata/maps_and_text.a",
	"internal/oracle/testdata/stack_overflow.a", "internal/oracle/testdata/stack_forever.a",
	"internal/oracle/testdata/stack_tail_call.a", "internal/oracle/testdata/panic.a",
	"internal/oracle/testdata/json_stringify_numbers.a", "internal/oracle/testdata/regexp_cycle_collections.a",
	"internal/oracle/testdata/library_object_keys.a", "internal/oracle/testdata/library_array_flat.a",
	"internal/oracle/testdata/read_files.a", "internal/native/wasm/io.a",
	"internal/oracle/testdata/closures_throw.a",
	"internal/oracle/testdata/write_stdout_order.a",
	"internal/oracle/testdata/write_stderr_order.a",
}

// wasiCompiler is the clang every WASI build in this file runs and keys its products by: the WASI SDK's, beside
// WASI_SYSROOT (compilerName), as the module emission and decode-ascii-wasi already were, never the box's
// tools/bin/clang that toolchain finds for "clang" (#jzc9yne). A path passes toolchain.Path unchanged.
func wasiCompiler() string {
	return compilerName(Options{Target: "wasm32-wasi"})
}

func runWASIUnit(test *testing.T, unit int) {
	test.Helper()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip(
			"WASI integration is opt-in: set ADAMIC_TEST_WASI=1 and WASI_SYSROOT, with the WASI SDK's clang beside it and Node 24 on PATH",
		)
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		test.Skip("WASI toolchain missing: WASI_SYSROOT is not set")
	}
	for _, tool := range []string{wasiCompiler(), "node", "go"} {
		if _, pathError := toolchain.Path(tool); pathError != nil {
			if toolchain.Sandboxed() {
				test.Fatal(pathError)
			}
			test.Skipf("WASI toolchain missing: %s: %v", tool, pathError)
		}
	}
	// The toolchain is pinned: a clang other than wasi-sdk 27's fails here, by name, rather than skipping.
	if _, wasiClangError := WasiClang(); wasiClangError != nil {
		test.Fatal(wasiClangError)
	}
	phase := time.Now()
	directory := test.TempDir()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	flags := wasiFlags(sysroot)
	probe := filepath.Join(directory, "probe.c")
	writeWasiFile(test, probe, "#include <stdlib.h>\nint main(void) { void *p = malloc(16); free(p); return 0; }\n")
	if output, combinedOutputError := toolchain.Command(
		wasiCompiler(),
		append(append([]string{}, flags...), probe, "-o", filepath.Join(directory, "probe.wasm"))...,
	).CombinedOutput(); combinedOutputError != nil {
		test.Skipf(
			"WASI toolchain missing or unusable (sysroot, linker, builtins): %v\n%s",
			combinedOutputError,
			output,
		)
	}
	if output, combinedOutputError := toolchain.Command(
		"node",
		"--disable-warning=ExperimentalWarning",
		"--input-type=module",
		"-e",
		"import {WASI} from 'node:wasi'; import {registerHooks} from 'node:module'; new WASI({version:'preview1'}); if (!registerHooks || Number(process.versions.node.split('.')[0]) < 24) process.exit(1)",
	).CombinedOutput(); combinedOutputError != nil {
		test.Skipf(
			"WASI toolchain missing: Node 24 with node:wasi and source oracle hooks required: %v\n%s",
			combinedOutputError,
			output,
		)
	}
	test.Logf("WASI toolchain probe phase %.3fs", time.Since(phase).Seconds())
	// The runtime's objects come from the strict-wasi-runtime product, built ahead; the counted variant is the
	// request benchmark's.
	runtimeVariant := func(test *testing.T, variant string) (string, []string) {
		test.Helper()
		started := time.Now()
		defer func() {
			test.Logf("WASI runtime fetch phase %s %.3fs", variant, time.Since(started).Seconds())
		}()
		return strictWasiRuntimeVariant(test, strictWasiRuntime(test), variant)
	}
	var objects []string
	directory, objects = runtimeVariant(test, "plain")
	test.Logf("strict C11 runtime: %d translation units compiled", len(objects))
	phase = time.Now()
	// The shared stage 0 product, built ahead by TestProduct_WASICompiler: a runner never builds it (#5qykzj5).
	compiler := buildcache.Adamic(test)
	test.Logf("WASI compiler fetch phase %.3fs", time.Since(phase).Seconds())
	generate := func(test *testing.T, fixture string) string {
		test.Helper()
		started := time.Now()
		defer func() { test.Logf("WASI generate phase %.3fs", time.Since(started).Seconds()) }()
		return string(wasiCommand(test, repository, compiler, "c", fixture))
	}
	link := func(test *testing.T, source, output string, extra ...string) {
		test.Helper()
		started := time.Now()
		defer func() { test.Logf("WASI link phase %.3fs", time.Since(started).Seconds()) }()
		arguments := append(
			append([]string{}, flags...),
			"-Wno-unused-variable",
			"-Wno-unused-but-set-variable",
			"-Wno-unused-function",
			"-Wno-unused-parameter",
			"-Wno-self-assign",
			"-I",
			directory,
			"-Wl,-z,stack-size=131072",
			"-Wl,--export=__stack_low",
			"-o",
			output,
			source,
		)
		arguments = append(arguments, extra...)
		arguments = append(arguments, objects...)
		arguments = append(arguments, "-lm")
		command := toolchain.CommandIn(repository, wasiCompiler(), arguments...)
		if data, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("WASI link: %v\n%s", combinedOutputError, data)
		}
	}

	passed := 0
	shard := testShard{unit, 36}
	for index, fixture := range wasiFixtures {
		if !shard.owns(index) {
			continue
		}
		test.Run(fixture, func(test *testing.T) {
			scratch := test.TempDir()
			source := filepath.Join(scratch, "main.c")
			writeWasiFile(test, source, generate(test, fixture))
			module := filepath.Join(scratch, "main.wasm")
			link(test, source, module)
			working := repository
			if strings.HasSuffix(fixture, "read_files.a") {
				working = filepath.Join(repository, "internal/oracle/testdata")
			}
			if strings.HasSuffix(fixture, "wasm/io.a") {
				working = scratch
				if createDirectoryError := os.Mkdir(
					filepath.Join(scratch, "files"),
					0o755,
				); createDirectoryError != nil {
					test.Fatal(createDirectoryError)
				}
				writeWasiFile(test, filepath.Join(scratch, "files/input.txt"), "héllo 世界\n")
			}
			started := time.Now()
			node := observeWasi(
				test,
				working,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "oracle/node.mjs"),
				filepath.Join(repository, fixture),
			)
			test.Logf("WASI source Node phase %.3fs", time.Since(started).Seconds())
			started = time.Now()
			wasm := observeWasi(
				test,
				working,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "internal/native/wasm/run.mjs"),
				module,
			)
			test.Logf("WASI execution phase %.3fs", time.Since(started).Seconds())
			if !bytes.Equal(node.stdout, wasm.stdout) ||
				!bytes.Equal(node.stderr, wasm.stderr) ||
				node.exit != wasm.exit {
				test.Fatalf(
					"Node exit=%d stdout=%q stderr=%q; WASI exit=%d stdout=%q stderr=%q",
					node.exit,
					node.stdout,
					node.stderr,
					wasm.exit,
					wasm.stdout,
					wasm.stderr,
				)
			}
			passed++
		})
	}
	test.Logf("WASI oracle: %d/%d equivalent", passed, len(wasiFixtures))
	if !shard.owns(len(wasiFixtures)) {
		return
	}
	test.Run("requests", func(test *testing.T) {
		// Counters are for the memory probe only; oracle command stderr stays untouched.
		directory, objects = runtimeVariant(test, "counted")
		scratch := test.TempDir()
		code := generate(test, "internal/native/wasm/request.a")
		if !strings.Contains(code, "adamic_region_end(") {
			test.Fatal("request fixture must actually use a statement region")
		}
		handler := wasiRequestHandler(code)
		if len(handler) != 2 {
			test.Fatal("cannot find the handler in the existing executable C backend")
		}
		writeWasiFile(test, filepath.Join(scratch, "program.c"), code)
		module := filepath.Join(scratch, "request.wasm")
		link(
			test,
			filepath.Join(repository, "internal/native/wasm/request-abi.c"),
			module,
			"-I",
			scratch,
			"-DADAMIC_HANDLER="+handler[1],
			"-mexec-model=reactor",
			"-Wl,--export=malloc",
			"-Wl,--export=free",
			"-Wl,--export=adamic_release",
		)
		output := wasiCommand(
			test,
			repository,
			"node",
			"--disable-warning=ExperimentalWarning",
			"internal/native/wasm/request-host.mjs",
			module,
			"internal/native/wasm/request.a",
		)
		test.Logf("request benchmark: %s", output)
		if destination := os.Getenv("ADAMIC_WASI_ARTIFACT"); destination != "" {
			contents, readError := os.ReadFile(module)
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(destination, contents, 0o644); writeError != nil {
				test.Fatal(writeError)
			}
		}
	})
}

func writeWasiFile(test *testing.T, path, contents string) {
	test.Helper()
	if writeError := os.WriteFile(path, []byte(contents), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
}

func wasiCommand(test *testing.T, directory, name string, arguments ...string) []byte {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := programCommand(timeoutContext, directory, name, arguments...)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("%s %v: %v\n%s", name, arguments, combinedOutputError, output)
	}
	return output
}

type wasiObservation struct {
	stdout, stderr []byte
	exit           int
}

func observeWasi(test *testing.T, directory, name string, arguments ...string) wasiObservation {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := programCommand(timeoutContext, directory, name, arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	exit := 0
	if runError != nil {
		var failed *exec.ExitError
		if !errors.As(runError, &failed) || timeoutContext.Err() != nil {
			test.Fatalf("%s %v: %v", name, arguments, runError)
		}
		exit = failed.ExitCode()
	}
	return wasiObservation{stdout.Bytes(), stderr.Bytes(), exit}
}

// TestProduct_WASICompiler builds the stage 0 compiler the WASI units run, so the tree build builds it ahead. It is the
// one product test that builds it (buildcache.AdamicOwner): every other package's finds it, so it has one key.
func TestProduct_WASICompiler(test *testing.T) {
	test.Parallel()
	buildcache.Adamic(test)
}

// wasiFlags are what the WASI units compile the runtime and link every module with, without the generated-C unused
// warnings exemptions, which only a module's own C gets.
func wasiFlags(sysroot string) []string {
	return []string{
		"--target=wasm32-wasi",
		"--sysroot=" + sysroot,
		"-DADAMIC_TARGET_WASI=1",
		"-std=c11",
		"-Wall",
		"-Wextra",
		"-Werror",
		"-pedantic",
		"-O2",
		"-ffp-contract=off",
		"-fno-optimize-sibling-calls",
	}
}

// strictWasiRuntimeVariants are the strict runtime's two builds: plain, which every fixture's module links, and
// counted (-DADAMIC_COUNT), which only the request benchmark's memory probe links.
var strictWasiRuntimeVariants = map[string][]string{"plain": nil, "counted": {"-DADAMIC_COUNT"}}

// strictWasiRuntime is every translation unit of the runtime compiled with wasiFlags, plain and counted, a product
// built ahead by TestProduct_StrictWASIRuntime: a directory per variant holding the runtime's files beside their
// objects. Every unit used to compile it into a cache keyed by the test binary's hash, which no product declares, so a
// runner that reads products never found it.
func strictWasiRuntime(test *testing.T) string {
	test.Helper()
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		test.Fatal("WASI_SYSROOT missing")
	}
	// The runtime compiles only with wasi-sdk 27's clang, refused by name otherwise, and its version keys the product.
	if _, wasiClangError := WasiClang(); wasiClangError != nil {
		test.Fatal(wasiClangError)
	}
	flags := wasiFlags(sysroot)
	variants := slices.Sorted(maps.Keys(strictWasiRuntimeVariants))
	inputs := buildcache.Inputs{
		Name: "strict-wasi-runtime",
		Files: []string{
			"internal/native/runtime",
			"internal/native/wasm_test.go",
			"internal/native/split_units_test.go",
			"internal/native/library.go",
		},
		Flags:     slices.Clone(flags),
		Toolchain: []string{buildcache.Tool(wasiCompiler(), "--version")},
	}
	for _, variant := range variants {
		inputs.Flags = append(
			inputs.Flags,
			fmt.Sprintf("variant %s %s", variant, strings.Join(strictWasiRuntimeVariants[variant], " ")),
		)
	}
	files, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		test.Fatal(readRuntimeError)
	}
	// A planted runtime (ADAMIC_WASI_RUNTIME, for isolated mutants) lies outside the tree, so its content keys the
	// product. It is built for this machine alone.
	if alternate := os.Getenv("ADAMIC_WASI_RUNTIME"); alternate != "" {
		if files, readRuntimeError = readRuntime(os.DirFS(alternate), "."); readRuntimeError != nil {
			test.Fatal(readRuntimeError)
		}
		for _, file := range files {
			inputs.Flags = append(inputs.Flags, fmt.Sprintf("runtime %s %x", file.name, sha256.Sum256(file.contents)))
		}
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		for _, variant := range variants {
			destination := filepath.Join(directory, variant)
			if createDirectoryError := os.Mkdir(destination, 0o755); createDirectoryError != nil {
				return createDirectoryError
			}
			var sources []string
			for _, file := range files {
				if writeError := os.WriteFile(
					filepath.Join(destination, file.name),
					file.contents,
					0o644,
				); writeError != nil {
					return writeError
				}
				if strings.HasSuffix(file.name, ".c") {
					sources = append(sources, file.name)
				}
			}
			// Each unit compiles from inside the variant's directory, mapped, so no object names where it was built.
			compileFlags := append(
				append(slices.Clone(flags), strictWasiRuntimeVariants[variant]...),
				prefixMaps(destination, "/adamic-runtime")...,
			)
			if eachInParallelError := eachInParallel(len(sources), func(index int) error {
				command := toolchain.CommandIn(
					destination,
					wasiCompiler(),
					append(slices.Clone(compileFlags), "-c", sources[index], "-o", sources[index]+".o")...,
				)
				if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
					return fmt.Errorf("compile %s %s: %w\n%s", variant, sources[index], combinedOutputError, output)
				}
				return nil
			}); eachInParallelError != nil {
				return eachInParallelError
			}
		}
		return nil
	})
}

// strictWasiRuntimeVariant is one variant of the strict runtime product: its directory, which holds the headers a
// module includes, and its objects in the runtime's order.
func strictWasiRuntimeVariant(test *testing.T, product, variant string) (string, []string) {
	test.Helper()
	if _, ok := strictWasiRuntimeVariants[variant]; !ok {
		test.Fatalf("no strict WASI runtime variant %q", variant)
	}
	directory := filepath.Join(product, variant)
	objects, globError := filepath.Glob(filepath.Join(directory, "*.c.o"))
	if globError != nil {
		test.Fatal(globError)
	}
	if len(objects) == 0 {
		test.Fatalf("strict WASI runtime %s holds no objects", variant)
	}
	return directory, objects
}

// TestProduct_StrictWasiRuntime builds the strict runtime the WASI units link, where they run, so the tree build
// builds it ahead.
func TestProduct_StrictWasiRuntime(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip("set ADAMIC_TEST_WASI=1")
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		test.Skip("WASI toolchain missing: WASI_SYSROOT is not set")
	}
	strictWasiRuntime(test)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit00(test *testing.T) {
	runWASIUnit(test, 0)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit01(test *testing.T) {
	runWASIUnit(test, 1)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit02(test *testing.T) {
	runWASIUnit(test, 2)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit03(test *testing.T) {
	runWASIUnit(test, 3)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit04(test *testing.T) {
	runWASIUnit(test, 4)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit05(test *testing.T) {
	runWASIUnit(test, 5)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit06(test *testing.T) {
	runWASIUnit(test, 6)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit07(test *testing.T) {
	runWASIUnit(test, 7)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit08(test *testing.T) {
	runWASIUnit(test, 8)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit09(test *testing.T) {
	runWASIUnit(test, 9)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit10(test *testing.T) {
	runWASIUnit(test, 10)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit11(test *testing.T) {
	runWASIUnit(test, 11)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit12(test *testing.T) {
	runWASIUnit(test, 12)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit13(test *testing.T) {
	runWASIUnit(test, 13)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit14(test *testing.T) {
	runWASIUnit(test, 14)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit15(test *testing.T) {
	runWASIUnit(test, 15)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit16(test *testing.T) {
	runWASIUnit(test, 16)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit17(test *testing.T) {
	runWASIUnit(test, 17)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit18(test *testing.T) {
	runWASIUnit(test, 18)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit19(test *testing.T) {
	runWASIUnit(test, 19)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit20(test *testing.T) {
	runWASIUnit(test, 20)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit21(test *testing.T) {
	runWASIUnit(test, 21)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit22(test *testing.T) {
	runWASIUnit(test, 22)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit23(test *testing.T) {
	runWASIUnit(test, 23)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit24(test *testing.T) {
	runWASIUnit(test, 24)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit25(test *testing.T) {
	runWASIUnit(test, 25)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit26(test *testing.T) {
	runWASIUnit(test, 26)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit27(test *testing.T) {
	runWASIUnit(test, 27)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit28(test *testing.T) {
	runWASIUnit(test, 28)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit29(test *testing.T) {
	runWASIUnit(test, 29)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit30(test *testing.T) {
	runWASIUnit(test, 30)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit31(test *testing.T) {
	runWASIUnit(test, 31)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit32(test *testing.T) {
	runWASIUnit(test, 32)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit33(test *testing.T) {
	runWASIUnit(test, 33)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit34(test *testing.T) {
	runWASIUnit(test, 34)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWasiUnit35(test *testing.T) {
	runWASIUnit(test, 35)
}

// Every WASI build in this file runs wasiCompiler: none names "clang", which toolchain resolves to the box's own clang
// rather than the SDK's beside WASI_SYSROOT, so a build of one and a key of the other could each work and disagree
// (#jzc9yne). Mutant: one site back to "clang".
func TestTheWasiBuildsNameNoBareClang(test *testing.T) {
	test.Parallel()
	files := token.NewFileSet()
	parsed, parseFileError := parser.ParseFile(files, "wasm_test.go", nil, parser.SkipObjectResolution)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING && literal.Value == `"clang"` {
			test.Errorf("%s names bare clang: run wasiCompiler, the WASI SDK's", files.Position(literal.Pos()))
		}
		return true
	})
}

// The ABI adapter locates the source-qualified handler, independent of ordinals.
func wasiRequestHandler(code string) []string {
	return regexp.MustCompile(
		`static adamic_string \* (adamic_function_[A-Za-z0-9_]*handleRequest_[a-f0-9]{32})\(`,
	).FindStringSubmatch(code)
}
