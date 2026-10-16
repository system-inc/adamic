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

func runWASIUnit(t *testing.T, unit int) {
	t.Helper()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("WASI integration is opt-in: set ADAMIC_TEST_WASI=1 and WASI_SYSROOT, with the WASI SDK's clang beside it and Node 24 on PATH")
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		t.Skip("WASI toolchain missing: WASI_SYSROOT is not set")
	}
	for _, tool := range []string{wasiCompiler(), "node", "go"} {
		if _, err := toolchain.Path(tool); err != nil {
			if toolchain.Sandboxed() {
				t.Fatal(err)
			}
			t.Skipf("WASI toolchain missing: %s: %v", tool, err)
		}
	}
	phase := time.Now()
	directory := t.TempDir()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	flags := wasiFlags(sysroot)
	probe := filepath.Join(directory, "probe.c")
	writeWASIFile(t, probe, "#include <stdlib.h>\nint main(void) { void *p = malloc(16); free(p); return 0; }\n")
	if output, err := toolchain.Command(wasiCompiler(), append(append([]string{}, flags...), probe, "-o", filepath.Join(directory, "probe.wasm"))...).CombinedOutput(); err != nil {
		t.Skipf("WASI toolchain missing or unusable (sysroot, linker, builtins): %v\n%s", err, output)
	}
	if output, err := toolchain.Command("node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e",
		"import {WASI} from 'node:wasi'; import {registerHooks} from 'node:module'; new WASI({version:'preview1'}); if (!registerHooks || Number(process.versions.node.split('.')[0]) < 24) process.exit(1)").CombinedOutput(); err != nil {
		t.Skipf("WASI toolchain missing: Node 24 with node:wasi and source oracle hooks required: %v\n%s", err, output)
	}
	t.Logf("WASI toolchain probe phase %.3fs", time.Since(phase).Seconds())
	// The runtime's objects come from the strict-wasi-runtime product, built ahead; the counted variant is the
	// request benchmark's.
	runtimeVariant := func(t *testing.T, variant string) (string, []string) {
		t.Helper()
		started := time.Now()
		defer func() {
			t.Logf("WASI runtime fetch phase %s %.3fs", variant, time.Since(started).Seconds())
		}()
		return strictWASIRuntimeVariant(t, strictWASIRuntime(t), variant)
	}
	var objects []string
	directory, objects = runtimeVariant(t, "plain")
	t.Logf("strict C11 runtime: %d translation units compiled", len(objects))
	phase = time.Now()
	// The shared stage 0 product, built ahead by TestProduct_WASICompiler: a runner never builds it (#5qykzj5).
	compiler := buildcache.Adamic(t)
	t.Logf("WASI compiler fetch phase %.3fs", time.Since(phase).Seconds())
	generate := func(t *testing.T, fixture string) string {
		t.Helper()
		started := time.Now()
		defer func() { t.Logf("WASI generate phase %.3fs", time.Since(started).Seconds()) }()
		return string(wasiCommand(t, repository, compiler, "c", fixture))
	}
	link := func(t *testing.T, source, output string, extra ...string) {
		t.Helper()
		started := time.Now()
		defer func() { t.Logf("WASI link phase %.3fs", time.Since(started).Seconds()) }()
		arguments := append(append([]string{}, flags...), "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function", "-Wno-unused-parameter", "-Wno-self-assign",
			"-I", directory, "-Wl,-z,stack-size=131072", "-Wl,--export=__stack_low", "-o", output, source)
		arguments = append(arguments, extra...)
		arguments = append(arguments, objects...)
		arguments = append(arguments, "-lm")
		command := toolchain.CommandIn(repository, wasiCompiler(), arguments...)
		if data, err := command.CombinedOutput(); err != nil {
			t.Fatalf("WASI link: %v\n%s", err, data)
		}
	}

	passed := 0
	shard := testShard{unit, 36}
	for index, fixture := range wasiFixtures {
		if !shard.owns(index) {
			continue
		}
		t.Run(fixture, func(t *testing.T) {
			scratch := t.TempDir()
			source := filepath.Join(scratch, "main.c")
			writeWASIFile(t, source, generate(t, fixture))
			module := filepath.Join(scratch, "main.wasm")
			link(t, source, module)
			working := repository
			if strings.HasSuffix(fixture, "read_files.a") {
				working = filepath.Join(repository, "internal/oracle/testdata")
			}
			if strings.HasSuffix(fixture, "wasm/io.a") {
				working = scratch
				if err := os.Mkdir(filepath.Join(scratch, "files"), 0o755); err != nil {
					t.Fatal(err)
				}
				writeWASIFile(t, filepath.Join(scratch, "files/input.txt"), "héllo 世界\n")
			}
			started := time.Now()
			node := observeWASI(t, working, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(repository, fixture))
			t.Logf("WASI source Node phase %.3fs", time.Since(started).Seconds())
			started = time.Now()
			wasm := observeWASI(t, working, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "internal/native/wasm/run.mjs"), module)
			t.Logf("WASI execution phase %.3fs", time.Since(started).Seconds())
			if !bytes.Equal(node.stdout, wasm.stdout) || !bytes.Equal(node.stderr, wasm.stderr) || node.exit != wasm.exit {
				t.Fatalf("Node exit=%d stdout=%q stderr=%q; WASI exit=%d stdout=%q stderr=%q", node.exit, node.stdout, node.stderr, wasm.exit, wasm.stdout, wasm.stderr)
			}
			passed++
		})
	}
	t.Logf("WASI oracle: %d/%d equivalent", passed, len(wasiFixtures))
	if !shard.owns(len(wasiFixtures)) {
		return
	}
	t.Run("requests", func(t *testing.T) {
		// Counters are for the memory probe only; oracle command stderr stays untouched.
		directory, objects = runtimeVariant(t, "counted")
		scratch := t.TempDir()
		code := generate(t, "internal/native/wasm/request.a")
		if !strings.Contains(code, "adamic_region_end(") {
			t.Fatal("request fixture must actually use a statement region")
		}
		handler := wasiRequestHandler(code)
		if len(handler) != 2 {
			t.Fatal("cannot find the handler in the existing executable C backend")
		}
		writeWASIFile(t, filepath.Join(scratch, "program.c"), code)
		module := filepath.Join(scratch, "request.wasm")
		link(t, filepath.Join(repository, "internal/native/wasm/request-abi.c"), module, "-I", scratch,
			"-DADAMIC_HANDLER="+handler[1], "-mexec-model=reactor", "-Wl,--export=malloc", "-Wl,--export=free", "-Wl,--export=adamic_release")
		output := wasiCommand(t, repository, "node", "--disable-warning=ExperimentalWarning", "internal/native/wasm/request-host.mjs", module, "internal/native/wasm/request.a")
		t.Logf("request benchmark: %s", output)
		if destination := os.Getenv("ADAMIC_WASI_ARTIFACT"); destination != "" {
			contents, err := os.ReadFile(module)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(destination, contents, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func writeWASIFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func wasiCommand(t *testing.T, directory, name string, arguments ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := programCommand(ctx, directory, name, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, arguments, err, output)
	}
	return output
}

type wasiObservation struct {
	stdout, stderr []byte
	exit           int
}

func observeWASI(t *testing.T, directory, name string, arguments ...string) wasiObservation {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := programCommand(ctx, directory, name, arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exit := 0
	if err != nil {
		var failed *exec.ExitError
		if !errors.As(err, &failed) || ctx.Err() != nil {
			t.Fatalf("%s %v: %v", name, arguments, err)
		}
		exit = failed.ExitCode()
	}
	return wasiObservation{stdout.Bytes(), stderr.Bytes(), exit}
}

// TestProduct_WASICompiler builds the stage 0 compiler the WASI units run, so the tree build builds it ahead. It is the
// one product test that builds it (buildcache.AdamicOwner): every other package's finds it, so it has one key.
func TestProduct_WASICompiler(t *testing.T) {
	t.Parallel()
	buildcache.Adamic(t)
}

// wasiFlags are what the WASI units compile the runtime and link every module with, without the generated-C unused
// warnings exemptions, which only a module's own C gets.
func wasiFlags(sysroot string) []string {
	return []string{"--target=wasm32-wasi", "--sysroot=" + sysroot, "-DADAMIC_TARGET_WASI=1",
		"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O2", "-ffp-contract=off", "-fno-optimize-sibling-calls"}
}

// strictWASIRuntimeVariants are the strict runtime's two builds: plain, which every fixture's module links, and
// counted (-DADAMIC_COUNT), which only the request benchmark's memory probe links.
var strictWASIRuntimeVariants = map[string][]string{"plain": nil, "counted": {"-DADAMIC_COUNT"}}

// strictWASIRuntime is every translation unit of the runtime compiled with wasiFlags, plain and counted, a product
// built ahead by TestProduct_StrictWASIRuntime: a directory per variant holding the runtime's files beside their
// objects. Every unit used to compile it into a cache keyed by the test binary's hash, which no product declares, so a
// runner that reads products never found it.
func strictWASIRuntime(t *testing.T) string {
	t.Helper()
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		t.Fatal("WASI_SYSROOT missing")
	}
	flags := wasiFlags(sysroot)
	variants := slices.Sorted(maps.Keys(strictWASIRuntimeVariants))
	inputs := buildcache.Inputs{
		Name:      "strict-wasi-runtime",
		Files:     []string{"internal/native/runtime", "internal/native/wasm_test.go", "internal/native/split_units_test.go", "internal/native/library.go"},
		Flags:     slices.Clone(flags),
		Toolchain: []string{buildcache.Tool(wasiCompiler(), "--version")},
	}
	for _, variant := range variants {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("variant %s %s", variant, strings.Join(strictWASIRuntimeVariants[variant], " ")))
	}
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	// A planted runtime (ADAMIC_WASI_RUNTIME, for isolated mutants) lies outside the tree, so its content keys the
	// product. It is built for this machine alone.
	if alternate := os.Getenv("ADAMIC_WASI_RUNTIME"); alternate != "" {
		if files, err = readRuntime(os.DirFS(alternate), "."); err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			inputs.Flags = append(inputs.Flags, fmt.Sprintf("runtime %s %x", file.name, sha256.Sum256(file.contents)))
		}
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		for _, variant := range variants {
			destination := filepath.Join(directory, variant)
			if err := os.Mkdir(destination, 0o755); err != nil {
				return err
			}
			var sources []string
			for _, file := range files {
				if err := os.WriteFile(filepath.Join(destination, file.name), file.contents, 0o644); err != nil {
					return err
				}
				if strings.HasSuffix(file.name, ".c") {
					sources = append(sources, file.name)
				}
			}
			// Each unit compiles from inside the variant's directory, mapped, so no object names where it was built.
			compileFlags := append(append(slices.Clone(flags), strictWASIRuntimeVariants[variant]...), prefixMaps(destination, "/adamic-runtime")...)
			if err := eachInParallel(len(sources), func(index int) error {
				command := toolchain.CommandIn(destination, wasiCompiler(), append(slices.Clone(compileFlags), "-c", sources[index], "-o", sources[index]+".o")...)
				if output, err := command.CombinedOutput(); err != nil {
					return fmt.Errorf("compile %s %s: %w\n%s", variant, sources[index], err, output)
				}
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// strictWASIRuntimeVariant is one variant of the strict runtime product: its directory, which holds the headers a
// module includes, and its objects in the runtime's order.
func strictWASIRuntimeVariant(t *testing.T, product, variant string) (string, []string) {
	t.Helper()
	if _, ok := strictWASIRuntimeVariants[variant]; !ok {
		t.Fatalf("no strict WASI runtime variant %q", variant)
	}
	directory := filepath.Join(product, variant)
	objects, err := filepath.Glob(filepath.Join(directory, "*.c.o"))
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) == 0 {
		t.Fatalf("strict WASI runtime %s holds no objects", variant)
	}
	return directory, objects
}

// TestProduct_StrictWASIRuntime builds the strict runtime the WASI units link, where they run, so the tree build
// builds it ahead.
func TestProduct_StrictWASIRuntime(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1")
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("WASI toolchain missing: WASI_SYSROOT is not set")
	}
	strictWASIRuntime(t)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit00(t *testing.T) {
	runWASIUnit(t, 0)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit01(t *testing.T) {
	runWASIUnit(t, 1)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit02(t *testing.T) {
	runWASIUnit(t, 2)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit03(t *testing.T) {
	runWASIUnit(t, 3)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit04(t *testing.T) {
	runWASIUnit(t, 4)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit05(t *testing.T) {
	runWASIUnit(t, 5)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit06(t *testing.T) {
	runWASIUnit(t, 6)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit07(t *testing.T) {
	runWASIUnit(t, 7)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit08(t *testing.T) {
	runWASIUnit(t, 8)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit09(t *testing.T) {
	runWASIUnit(t, 9)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit10(t *testing.T) {
	runWASIUnit(t, 10)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit11(t *testing.T) {
	runWASIUnit(t, 11)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit12(t *testing.T) {
	runWASIUnit(t, 12)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit13(t *testing.T) {
	runWASIUnit(t, 13)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit14(t *testing.T) {
	runWASIUnit(t, 14)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit15(t *testing.T) {
	runWASIUnit(t, 15)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit16(t *testing.T) {
	runWASIUnit(t, 16)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit17(t *testing.T) {
	runWASIUnit(t, 17)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit18(t *testing.T) {
	runWASIUnit(t, 18)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit19(t *testing.T) {
	runWASIUnit(t, 19)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit20(t *testing.T) {
	runWASIUnit(t, 20)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit21(t *testing.T) {
	runWASIUnit(t, 21)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit22(t *testing.T) {
	runWASIUnit(t, 22)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit23(t *testing.T) {
	runWASIUnit(t, 23)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit24(t *testing.T) {
	runWASIUnit(t, 24)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit25(t *testing.T) {
	runWASIUnit(t, 25)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit26(t *testing.T) {
	runWASIUnit(t, 26)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit27(t *testing.T) {
	runWASIUnit(t, 27)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit28(t *testing.T) {
	runWASIUnit(t, 28)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit29(t *testing.T) {
	runWASIUnit(t, 29)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit30(t *testing.T) {
	runWASIUnit(t, 30)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit31(t *testing.T) {
	runWASIUnit(t, 31)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit32(t *testing.T) {
	runWASIUnit(t, 32)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit33(t *testing.T) {
	runWASIUnit(t, 33)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit34(t *testing.T) {
	runWASIUnit(t, 34)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestWASIUnit35(t *testing.T) {
	runWASIUnit(t, 35)
}

// Every WASI build in this file runs wasiCompiler: none names "clang", which toolchain resolves to the box's own clang
// rather than the SDK's beside WASI_SYSROOT, so a build of one and a key of the other could each work and disagree
// (#jzc9yne). Mutant: one site back to "clang".
func TestTheWASIBuildsNameNoBareClang(t *testing.T) {
	t.Parallel()
	files := token.NewFileSet()
	parsed, err := parser.ParseFile(files, "wasm_test.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING && literal.Value == `"clang"` {
			t.Errorf("%s names bare clang: run wasiCompiler, the WASI SDK's", files.Position(literal.Pos()))
		}
		return true
	})
}

// The ABI adapter locates the source-qualified handler, independent of ordinals.
func wasiRequestHandler(code string) []string {
	return regexp.MustCompile(`static adamic_string \* (adamic_function_[A-Za-z0-9_]*handleRequest_[a-f0-9]{32})\(`).FindStringSubmatch(code)
}
