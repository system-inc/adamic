package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Build a real shared string and remove one owner before reading the other.
// The mutant must build and fail at runtime, not at clang's type checks.
func TestReleaseSharedValueAndUnsafeMutant(t *testing.T) {
	checked, err := load.Load([]string{"testdata/release_fastpath.ts"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	node := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", "testdata/release_fastpath.ts")
	expected, err := node.CombinedOutput()
	if err != nil || string(expected) != "200 xxxx\n" {
		t.Fatalf("Node: %v %s", err, expected)
	}
	for _, mutant := range []bool{false, true} {
		directory := t.TempDir()
		files, err := readRuntime(runtime, "runtime")
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			data := file.contents
			if mutant && file.name == "heap.c" {
				old := "heap == NULL || heap->references == 0 || --heap->references != 0"
				if strings.Count(string(data), old) != 1 {
					t.Fatal("mutant lost its unique anchor")
				}
				data = []byte(strings.Replace(string(data), old, "heap == NULL || heap->references == 0 || --heap->references > 1", 1))
			}
			if err := os.WriteFile(filepath.Join(directory, file.name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		library, err := RuntimeLibrary(directory, Options{Sanitize: true})
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(directory, "main.c")
		if err := os.WriteFile(main, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(directory, "program")
		arguments := append(Flags(Options{Sanitize: true}), "-I", filepath.Dir(library), "-o", binary, main)
		arguments = append(arguments, RuntimeLinkFlags(library)...)
		arguments = append(arguments, "-lm")
		if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
			t.Fatalf("clang: %v %s", err, output)
		}
		command := exec.Command(binary)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
		output, err := command.CombinedOutput()
		if mutant {
			if err == nil || !bytes.Contains(output, []byte("AddressSanitizer: heap-use-after-free")) {
				t.Fatalf("unsafe mutant survived: %v %s", err, output)
			}
			t.Log("unsafe last-owner mutant compiles; ASan catches heap-use-after-free on dynamic shared text")
		} else if err != nil || !bytes.Equal(output, expected) {
			t.Fatalf("native and leak check: %v %s; Node %s", err, output, expected)
		}
	}
}
