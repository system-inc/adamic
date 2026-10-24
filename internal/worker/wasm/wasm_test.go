package wasm

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: builds several reactors and runs the lifetime and independent mutant probes.
func workersWASMProduct(t *testing.T) string {
	t.Helper()
	return buildcache.Product(t, integrationInputs(t, "workersWASMProduct"), func(directory string) error {
		root, err := filepath.Abs("../../..")
		if err != nil {
			t.Fatal(err)
		}
		write := func(path, text string) {
			t.Helper()
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
		}
		run := commandRunner(t, root)
		compiler := buildcache.AdamicBuiltElsewhere(t)
		fixtures := map[string]string{
			"driver":  "cmd/adamic/testdata/wasi/request.a",
			"counted": "internal/native/wasm/request.a",
			"grow":    "internal/worker/wasm/grow.a",
			"panic":   "internal/worker/wasm/panic.a",
		}
		for name, source := range fixtures {
			arguments := []string{"build", "--target", "wasm32-wasi", source, "-o", filepath.Join(directory, name+".wasm")}
			if name != "panic" {
				arguments = append(arguments, "--count")
			}
			run(compiler, arguments...)
		}
		inputs := []string{"", "request", "echo", "reassign", "世界 🌍", strings.Repeat("x", 1024), "a\x00b", "\ufeffbom"}
		requests := make([]string, 1000)
		for index := range requests {
			requests[index] = inputs[index%len(inputs)]
		}
		data, err := json.Marshal(requests)
		if err != nil {
			t.Fatal(err)
		}
		write(filepath.Join(directory, "requests.json"), string(data))
		run("node", "--disable-warning=ExperimentalWarning", "internal/worker/wasm/reference.mjs",
			filepath.Join(directory, "driver.wasm"), filepath.Join(root, "cmd/adamic/testdata/wasi/request-host.mjs"), filepath.Join(directory, "requests.json"), filepath.Join(directory, "expected.json"))
		write(filepath.Join(directory, "grow-requests.json"), `["世界🌍"]`)
		run("node", "--disable-warning=ExperimentalWarning", "internal/worker/wasm/reference.mjs",
			filepath.Join(directory, "grow.wasm"), filepath.Join(root, "cmd/adamic/testdata/wasi/request-host.mjs"), filepath.Join(directory, "grow-requests.json"), filepath.Join(directory, "grow-expected.json"))
		// Also execute the complete original driver's source-oracle host unchanged.
		run("node", "--disable-warning=ExperimentalWarning", "cmd/adamic/testdata/wasi/request-host.mjs", filepath.Join(directory, "driver.wasm"), filepath.Join(root, "cmd/adamic/testdata/wasi/request.a"), filepath.Join(directory, "driver-stdout"))
		generated := filepath.Join(directory, "generated")
		run("node", "internal/worker/wasm/generate.mjs", generated)
		binary, err := os.ReadFile(filepath.Join(directory, "panic.wasm"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(generated, "handler.wasm"), binary, 0644); err != nil {
			t.Fatal(err)
		}

		// The oracle helper scripts name this build directory; only their JSON
		// observations are portable inputs to the fresh harness.
		for _, name := range []string{"expected.json.mjs", "grow-expected.json.mjs"} {
			if err := os.Remove(filepath.Join(directory, name)); err != nil {
				return err
			}
		}
		return nil
	})
}

// Not parallel: prepares shared WASI module inputs before the harnesses.
func TestProduct_WorkersWASMInputs(t *testing.T) { workersWASMProduct(t) }

func TestWorkersWASM(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	product := workersWASMProduct(t)
	directory := t.TempDir()
	if err := os.CopyFS(directory, os.DirFS(product)); err != nil {
		t.Fatal(err)
	}
	run := commandRunner(t, root)
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// A small real Wasm trap probe observes recreation independently of proc_exit.
	trapSource := filepath.Join(directory, "trap.c")
	write(trapSource, `
#include <stddef.h>
static unsigned char input[32];
static unsigned char output[1];
static unsigned calls;
__attribute__((export_name("_initialize"))) void initialize(void) { calls = 0; }
__attribute__((export_name("malloc"))) void *allocate(size_t size) { (void)size; return input; }
__attribute__((export_name("free"))) void deallocate(void *pointer) { (void)pointer; }
__attribute__((export_name("adamic_request"))) void *request(unsigned char *bytes, size_t length) {
 calls++;
 if (length && bytes[0] == 't') __builtin_trap();
 output[0] = (unsigned char)('0' + calls);
 return output;
}
__attribute__((export_name("adamic_response_bytes"))) void *response_bytes(void *value) { return value; }
__attribute__((export_name("adamic_response_length"))) size_t response_length(void *value) { (void)value; return 1; }
__attribute__((export_name("adamic_release"))) void release(void *value) { (void)value; }
`)
	run(wasiClang(t), "--target=wasm32-unknown-unknown", "-nostdlib", "-Wl,--no-entry", "-Wl,--export-memory", trapSource, "-o", filepath.Join(directory, "trap.wasm"))
	t.Log(string(run("node", "internal/worker/wasm/harness.mjs", directory)))
	// Mutants operate on independent copies, never on the checked-in host.
	mutants := []struct{ name, file, before, after, mode, diagnosis string }{
		{"stderr", "wasi.mjs", "else logger.error(line);", "else { /* swallowed stderr */ }", "panic", "panic stderr lost"},
		{"stale-view", "host.mjs", "const response = api.adamic_request(input, bytes.length) >>> 0;", "const stale = api.memory.buffer;\n        const response = api.adamic_request(input, bytes.length) >>> 0;", "compare", "detached"},
		{"release", "host.mjs", "api.adamic_release(response);", "/* omitted release */", "lifetime", "warmup retained counted values"},
		{"dead-instance", "host.mjs", "instance = undefined;", "/* kept terminated instance */", "panic", "terminated instance reused"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			scratch := t.TempDir()
			for _, name := range []string{"harness.mjs", "host.mjs", "wasi.mjs", "bridge.mjs"} {
				data, err := os.ReadFile(filepath.Join(root, "internal/worker/wasm", name))
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if name == mutant.file {
					if !strings.Contains(source, mutant.before) {
						t.Fatal("mutant target missing")
					}
					source = strings.ReplaceAll(source, mutant.before, mutant.after)
					if mutant.name == "stale-view" {
						source = strings.Replace(source, "new Uint8Array(api.memory.buffer, pointer, length)", "new Uint8Array(stale, pointer, length)", 1)
					}
				}
				write(filepath.Join(scratch, name), source)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, "node", filepath.Join(scratch, "harness.mjs"), directory, mutant.mode)
			output, err := command.CombinedOutput()
			if err == nil || ctx.Err() != nil || !strings.Contains(string(output), mutant.diagnosis) {
				t.Fatalf("mutant must fail its intended check: %v\n%s", err, output)
			}
			t.Logf("caught %s: %s", mutant.name, mutant.diagnosis)
		})
	}
}
