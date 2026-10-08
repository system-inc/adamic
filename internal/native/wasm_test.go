package native

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Not parallel: this opt-in integration probe builds one complete runtime and runs a request
// benchmark. Keep native tests on the native clang PATH, and run this separately with WASI clang.
func TestWASI(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("WASI integration is opt-in: set ADAMIC_TEST_WASI=1 and WASI_SYSROOT, with WASI clang and Node 24 on PATH")
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		t.Skip("WASI toolchain missing: WASI_SYSROOT is not set")
	}
	for _, tool := range []string{"clang", "node", "go"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("WASI toolchain missing: %s: %v", tool, err)
		}
	}
	// The toolchain is pinned: a clang other than wasi-sdk's fails here, by name, rather than skipping.
	clang, err := WASIClang()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"--target=wasm32-wasi", "--sysroot=" + sysroot, "-DADAMIC_TARGET_WASI=1",
		"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O2", "-ffp-contract=off", "-fno-optimize-sibling-calls"}
	probe := filepath.Join(directory, "probe.c")
	writeWASIFile(t, probe, "#include <stdlib.h>\nint main(void) { void *p = malloc(16); free(p); return 0; }\n")
	if output, err := exec.Command(clang, append(append([]string{}, flags...), probe, "-o", filepath.Join(directory, "probe.wasm"))...).CombinedOutput(); err != nil {
		t.Skipf("WASI toolchain missing or unusable (sysroot, linker, builtins): %v\n%s", err, output)
	}
	if output, err := exec.Command("node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e",
		"import {WASI} from 'node:wasi'; import {registerHooks} from 'node:module'; new WASI({version:'preview1'}); if (!registerHooks || Number(process.versions.node.split('.')[0]) < 24) process.exit(1)").CombinedOutput(); err != nil {
		t.Skipf("WASI toolchain missing: Node 24 with node:wasi and source oracle hooks required: %v\n%s", err, output)
	}
	// Compile every translation unit without the generated-C unused warnings exemptions.
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	if alternate := os.Getenv("ADAMIC_WASI_RUNTIME"); alternate != "" {
		files, err = readRuntime(os.DirFS(alternate), ".")
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var objects []string
	for _, file := range files {
		if !strings.HasSuffix(file.name, ".c") {
			continue
		}
		object := filepath.Join(directory, file.name+".o")
		wasiCommand(t, repository, clang, append(append([]string{}, flags...), "-c", filepath.Join(directory, file.name), "-o", object)...)
		objects = append(objects, object)
	}
	t.Logf("strict C11 runtime: %d translation units compiled", len(objects))
	compiler := filepath.Join(directory, "adamic")
	wasiCommand(t, repository, "go", "build", "-o", compiler, "./cmd/adamic")
	generate := func(t *testing.T, fixture string) string {
		t.Helper()
		return string(wasiCommand(t, repository, compiler, "c", fixture))
	}
	link := func(t *testing.T, source, output string, extra ...string) {
		t.Helper()
		arguments := append(append([]string{}, flags...), "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function", "-Wno-unused-parameter", "-Wno-self-assign",
			"-I", directory, "-Wl,-z,stack-size=131072", "-Wl,--export=__stack_low", "-o", output, source)
		arguments = append(arguments, extra...)
		arguments = append(arguments, objects...)
		arguments = append(arguments, "-lm")
		wasiCommand(t, repository, clang, arguments...)
	}
	fixtures := []string{
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
	passed := 0
	for _, fixture := range fixtures {
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
			node := observeWASI(t, working, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(repository, fixture))
			wasm := observeWASI(t, working, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "internal/native/wasm/run.mjs"), module)
			if !bytes.Equal(node.stdout, wasm.stdout) || !bytes.Equal(node.stderr, wasm.stderr) || node.exit != wasm.exit {
				t.Fatalf("Node exit=%d stdout=%q stderr=%q; WASI exit=%d stdout=%q stderr=%q", node.exit, node.stdout, node.stderr, wasm.exit, wasm.stdout, wasm.stderr)
			}
			passed++
		})
	}
	t.Logf("WASI oracle: %d/%d equivalent", passed, len(fixtures))
	t.Run("requests", func(t *testing.T) {
		// Counters are for the memory probe only; oracle command stderr stays untouched.
		for _, file := range files {
			if strings.HasSuffix(file.name, ".c") {
				wasiCommand(t, repository, clang, append(append([]string{}, flags...), "-DADAMIC_COUNT", "-c", filepath.Join(directory, file.name), "-o", filepath.Join(directory, file.name+".o"))...)
			}
		}
		scratch := t.TempDir()
		code := generate(t, "internal/native/wasm/request.a")
		if !strings.Contains(code, "adamic_region_end(") {
			t.Fatal("request fixture must actually use a statement region")
		}
		handler := regexp.MustCompile(`static adamic_string \* (adamic_function_[0-9]+_handleRequest)\(`).FindStringSubmatch(code)
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
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
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
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
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
