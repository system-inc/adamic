package wasm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: full lifetime checks and independent mutant processes share the toolchain.
func TestGeneratedCrossing(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1 for generated crossing integration")
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		t.Skip("missing WASI_SYSROOT")
	}
	for _, tool := range []string{"go", "node", "clang"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("missing %s: %v", tool, err)
		}
	}
	if output, err := exec.Command("node", "-e", "if(Number(process.versions.node.split('.')[0])<24)process.exit(1)").CombinedOutput(); err != nil {
		t.Skipf("missing Node 24: %v %s", err, output)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	probe := filepath.Join(directory, "probe.c")
	write(probe, "#include <stdlib.h>\nint main(void) { void *p = malloc(16); free(p); return 0; }\n")
	if output, err := exec.Command("clang", "--target=wasm32-wasi", "--sysroot="+sysroot, probe, "-o", filepath.Join(directory, "probe.wasm")).CombinedOutput(); err != nil {
		t.Skipf("missing usable WASI clang, linker, sysroot or builtins: %v\n%s", err, output)
	}
	run := func(name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
		return output
	}
	compiler := filepath.Join(directory, "adamic")
	run("go", "build", "-o", compiler, "./cmd/adamic")
	run(compiler, "build", "--target", "wasm32-wasi", "internal/worker/wasm/crossing.a", "-o", filepath.Join(directory, "crossing.wasm"), "--reactor", "--count", "--abi-json", filepath.Join(directory, "abi.json"))
	generated := filepath.Join(directory, "crossing.mjs")
	run("node", "internal/worker/wasm/generate-crossing.mjs", filepath.Join(directory, "abi.json"), generated)
	trapSource := filepath.Join(directory, "trap.c")
	write(trapSource, `
#include <stddef.h>
static unsigned calls;
__attribute__((export_name("_initialize"))) void initialize(void) { calls = 0; }
__attribute__((export_name("adamic_export_numberValue"))) double number_value(double value) {
 calls++; if (value < 0) __builtin_trap(); return calls;
}
__attribute__((export_name("adamic_alloc"))) unsigned allocate(unsigned size) { (void)size; return 0; }
__attribute__((export_name("adamic_free"))) void deallocate(unsigned pointer) { (void)pointer; }
__attribute__((export_name("adamic_result_bytes"))) unsigned result_bytes(unsigned handle) { (void)handle; return 0; }
__attribute__((export_name("adamic_result_length"))) unsigned result_length(unsigned handle) { (void)handle; return 0; }
__attribute__((export_name("adamic_result_release"))) void result_release(unsigned handle) { (void)handle; }
`)
	run("clang", "--target=wasm32-unknown-unknown", "-nostdlib", "-Wl,--no-entry", "-Wl,--export-memory", trapSource, "-o", filepath.Join(directory, "trap.wasm"))
	write(filepath.Join(directory, "trap.json"), `{"version":1,"exports":[{"name":"numberValue","parameters":[{"name":"value","type":{"kind":"number"}}],"returns":{"kind":"number"}}]}`)
	run("node", "internal/worker/wasm/generate-crossing.mjs", filepath.Join(directory, "trap.json"), filepath.Join(directory, "trap-crossing.mjs"))
	reference := filepath.Join(root, "internal/native/wasm/exports.mjs")
	t.Log(string(run("node", "internal/worker/wasm/crossing-harness.mjs", directory, reference, generated)))
	mutants := []struct{ name, file, before, after, mode, diagnosis string }{
		{"nested-alignment", "crossing.mjs", `encoder0.boolean(argument0["inner"]["flag"]);
encoder0.number(argument0["inner"]["score"]);`, `const nestedOrigin = encoder0.offset;
encoder0.boolean(argument0["inner"]["flag"]);
const padding = (8 - (encoder0.offset - nestedOrigin) % 8) % 8;
encoder0.reserve(padding + 8);
encoder0.offset += padding;
new DataView(encoder0.buffer.buffer).setFloat64(encoder0.offset, argument0["inner"]["score"], true);
encoder0.offset += 8;`, "compare", "recordValue input bytes differ"},
		{"lone-surrogate", "wtf8.mjs", "const unit = text.charCodeAt(index);", "let unit = text.charCodeAt(index);\n    if ((unit >= 0xdc00 && unit < 0xe000) || (unit >= 0xd800 && unit < 0xdc00 && !(text.charCodeAt(index + 1) >= 0xdc00 && text.charCodeAt(index + 1) < 0xe000))) unit = 0xfffd;", "compare", "input bytes differ"},
		{"result-release", "crossing.mjs", "context.invoke('adamic_result_release', handle);", "; /* result handle leaked */", "lifetime", "memory grew after warmup"},
		{"stale-view", "crossing.mjs", "const returned = context.invoke(\"adamic_export_grow\", input0, bytes0.length, argument1);", "const stale = api.memory.buffer;\nconst returned = context.invoke(\"adamic_export_grow\", input0, bytes0.length, argument1);", "growth", "detached"},
		{"input-double-free", "crossing.mjs", "context.invoke('adamic_free', input0);", "context.invoke('adamic_free', input0); context.invoke('adamic_free', input0);", "compare", "input freed twice"},
		{"input-not-freed", "crossing.mjs", "context.invoke('adamic_free', input0);", "; /* input leaked */", "compare", "input allocation leaked"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			scratch := t.TempDir()
			for _, name := range []string{"crossing.mjs", "wtf8.mjs", "crossing-runtime.mjs", "wasi.mjs"} {
				data, err := os.ReadFile(filepath.Join(directory, name))
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if name == mutant.file {
					if !strings.Contains(source, mutant.before) {
						t.Fatalf("mutant target missing: %s", mutant.before)
					}
					source = strings.ReplaceAll(source, mutant.before, mutant.after)
					if mutant.name == "stale-view" {
						start := strings.Index(source, "const stale = api.memory.buffer;")
						source = source[:start] + strings.Replace(source[start:], "new Uint8Array(api.memory.buffer, pointer, length)", "new Uint8Array(stale, pointer, length)", 1)
					}
				}
				write(filepath.Join(scratch, name), source)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, "node", filepath.Join(root, "internal/worker/wasm/crossing-harness.mjs"), directory, reference, filepath.Join(scratch, "crossing.mjs"), mutant.mode)
			output, err := command.CombinedOutput()
			if err == nil || ctx.Err() != nil || !strings.Contains(string(output), mutant.diagnosis) {
				t.Fatalf("mutant must fail its intended check: %v\n%s", err, output)
			}
			t.Logf("caught %s: %s", mutant.name, mutant.diagnosis)
		})
	}
}
