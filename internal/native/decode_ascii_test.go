package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: expands a large byte oracle corpus and runs complete runtime builds serially.
func TestDecodeASCII(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	run := func(t *testing.T, name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, name, args...)
		command.Dir = repository
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
		return output
	}
	corpus := filepath.Join(directory, "corpus.bin")
	t.Log(string(run(t, "node", "internal/native/decode_ascii/corpus.mjs", corpus)))
	runtimeDirectory := filepath.Join(repository, "internal/native/runtime")
	if alternate := os.Getenv("ADAMIC_DECODE_RUNTIME"); alternate != "" {
		runtimeDirectory = alternate
	}
	sources, err := filepath.Glob(filepath.Join(runtimeDirectory, "*.c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) == 0 {
		t.Fatal("no runtime sources")
	}
	for _, target := range []string{"native", "wasi"} {
		t.Run(target, func(t *testing.T) {
			compiler := "clang"
			flags := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O2", "-ffp-contract=off", "-fno-optimize-sibling-calls", "-I", runtimeDirectory}
			if target == "wasi" {
				if os.Getenv("ADAMIC_TEST_WASI") != "1" {
					t.Skip("set ADAMIC_TEST_WASI=1")
				}
				sysroot := os.Getenv("WASI_SYSROOT")
				if sysroot == "" {
					t.Fatal("WASI_SYSROOT missing")
				}
				compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin/clang")
				flags = append(flags, "--target=wasm32-wasi", "--sysroot="+sysroot, "-DADAMIC_TARGET_WASI=1", "-Wl,-z,stack-size=1048576")
			} else {
				flags = append(flags, "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all")
			}
			binary := filepath.Join(directory, "decode-"+target)
			args := append(flags, "internal/native/decode_ascii/probe.c", "internal/native/decode_ascii/baseline.c")
			args = append(args, sources...)
			args = append(args, "-lm", "-o", binary)
			run(t, compiler, args...)
			var output []byte
			if target == "wasi" {
				output = run(t, "node", "--disable-warning=ExperimentalWarning", "oracle/wasi.mjs", binary, corpus)
			} else {
				output = run(t, binary, corpus)
			}
			if !strings.Contains(string(output), "Node and baseline identical") {
				t.Fatalf("missing completed corpus: %s", output)
			}
			t.Log(string(output))
		})
	}
}
