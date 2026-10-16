package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Build a real shared string and remove one owner before reading the other.
// The mutant must build and fail at runtime, not at clang's type checks.
func TestReleaseSharedValueAndUnsafeMutant(t *testing.T) {
	t.Parallel()
	checked, err := load.Load([]string{"testdata/release_fastpath.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	node := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", "testdata/release_fastpath.a")
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
			if mutant && (file.name == "heap.c" || file.name == "adamic.h") {
				// Route count two through the slow path, then wrongly free its remaining owner.
				old, changed := "return count == 1 ? heap : NULL;", "return count <= 2 ? heap : NULL;"
				if file.name == "adamic.h" {
					old, changed = "if (count > 1)", "if (count > 2)"
				}
				if strings.Count(string(data), old) != 1 {
					t.Fatal("mutant lost its unique anchor")
				}
				data = []byte(strings.Replace(string(data), old, changed, 1))
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
		command.Env = append(os.Environ(), "ADAMIC_THREADS=1", "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "LSAN_OPTIONS=exitcode=23", "UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
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

// Check emitted calls, not the spelling of inline in the header. ASan must retain
// instrumentation inside the wrappers; release and TSan must keep their fast paths.
func TestOwnershipInlining(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") {
			if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	source := filepath.Join(directory, "ownership.c")
	if err := os.WriteFile(source, []byte(`#include "adamic.h"
void *hold(void *value) { return adamic_retain(value); }
void drop(void *value) { adamic_release(value); }
`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, options := range []Options{{}, {Sanitize: true}, {Sanitize: true, Count: true}, {ThreadSanitize: true}} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		arguments := append(Flags(options), "-S", "-emit-llvm", "-o", "-", source)
		output, err := exec.CommandContext(ctx, "clang", arguments...).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("%+v clang: %v %s", options, err, output)
		}
		for _, name := range []string{"retain", "release"} {
			call := regexp.MustCompile(`call [^\n]*@adamic_` + name + `\(`)
			if got := call.Match(output); got != options.Sanitize {
				t.Errorf("%+v: %s call present=%t, want %t", options, name, got, options.Sanitize)
			}
			if options.Sanitize {
				definition := regexp.MustCompile(`(?ms)^define [^\n]*@adamic_` + name + `\([^\n]*\n.*?^}`).Find(output)
				if !bytes.Contains(definition, []byte("@__asan_report_load")) {
					t.Errorf("%+v: %s ownership wrapper lost ASan instrumentation", options, name)
				}
			}
		}
	}
}
