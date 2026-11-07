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
func TestReleaseSharedValueAndUnsafeMutant(test *testing.T) {
	test.Parallel()
	checked, loadError := load.Load([]string{"testdata/release_fastpath.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	node := exec.Command(
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		"testdata/release_fastpath.a",
	)
	expected, combinedOutputError := node.CombinedOutput()
	if combinedOutputError != nil || string(expected) != "200 xxxx\n" {
		test.Fatalf("Node: %v %s", combinedOutputError, expected)
	}
	for _, mutant := range []bool{false, true} {
		directory := test.TempDir()
		files, readRuntimeError := readRuntime(runtime, "runtime")
		if readRuntimeError != nil {
			test.Fatal(readRuntimeError)
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
					test.Fatal("mutant lost its unique anchor")
				}
				data = []byte(strings.Replace(string(data), old, changed, 1))
			}
			if writeError := os.WriteFile(filepath.Join(directory, file.name), data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
		}
		library, runtimeLibraryError := RuntimeLibrary(directory, Options{Sanitize: true})
		if runtimeLibraryError != nil {
			test.Fatal(runtimeLibraryError)
		}
		main := filepath.Join(directory, "main.c")
		if writeError := os.WriteFile(main, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		binary := filepath.Join(directory, "program")
		arguments := append(Flags(Options{Sanitize: true}), "-I", filepath.Dir(library), "-o", binary, main)
		arguments = append(arguments, RuntimeLinkFlags(library)...)
		arguments = append(arguments, "-lm")
		if output, combinedOutputError := exec.Command(
			"clang",
			arguments...,
		).CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("clang: %v %s", combinedOutputError, output)
		}
		command := exec.Command(binary)
		command.Env = append(
			os.Environ(),
			"ADAMIC_THREADS=1",
			"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1",
			"LSAN_OPTIONS=exitcode=23",
			"UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1",
		)
		output, combinedOutputError := command.CombinedOutput()
		if mutant {
			if combinedOutputError == nil || !bytes.Contains(output, []byte("AddressSanitizer: heap-use-after-free")) {
				test.Fatalf("unsafe mutant survived: %v %s", combinedOutputError, output)
			}
			test.Log("unsafe last-owner mutant compiles; ASan catches heap-use-after-free on dynamic shared text")
		} else if combinedOutputError != nil || !bytes.Equal(output, expected) {
			test.Fatalf("native and leak check: %v %s; Node %s", combinedOutputError, output, expected)
		}
	}
}

// Check emitted calls, not the spelling of inline in the header. ASan must retain
// instrumentation inside the wrappers; release and TSan must keep their fast paths.
func TestOwnershipInlining(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	files, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		test.Fatal(readRuntimeError)
	}
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") {
			if writeError := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0600); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	source := filepath.Join(directory, "ownership.c")
	if writeError := os.WriteFile(source, []byte(`#include "adamic.h"
void *hold(void *value) { return adamic_retain(value); }
void drop(void *value) { adamic_release(value); }
`), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	for _, options := range []Options{{}, {Sanitize: true}, {Sanitize: true, Count: true}, {ThreadSanitize: true}} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		arguments := append(Flags(options), "-S", "-emit-llvm", "-o", "-", source)
		output, combinedOutputError := exec.CommandContext(ctx, "clang", arguments...).CombinedOutput()
		cancel()
		if combinedOutputError != nil {
			test.Fatalf("%+v clang: %v %s", options, combinedOutputError, output)
		}
		for _, name := range []string{"retain", "release"} {
			call := regexp.MustCompile(`call [^\n]*@adamic_` + name + `\(`)
			if got := call.Match(output); got != options.Sanitize {
				test.Errorf("%+v: %s call present=%t, want %t", options, name, got, options.Sanitize)
			}
			if options.Sanitize {
				definition := regexp.MustCompile(`(?ms)^define [^\n]*@adamic_` + name + `\([^\n]*\n.*?^}`).Find(output)
				if !bytes.Contains(definition, []byte("@__asan_report_load")) {
					test.Errorf("%+v: %s ownership wrapper lost ASan instrumentation", options, name)
				}
			}
		}
	}
	// A file that includes the header and retains nothing builds under the strict flags
	// harnesses use (decode_ascii_cache_test.go), which don't pass -Wno-unused-function.
	quiet := filepath.Join(directory, "quiet.c")
	if writeError := os.WriteFile(
		quiet,
		[]byte("#include \"adamic.h\"\nint quiet(void) { return 0; }\n"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	for _, sanitizers := range []string{"-fsanitize=address,undefined", "-fsanitize=thread", ""} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		arguments := []string{
			"-std=c11",
			"-Wall",
			"-Wextra",
			"-Werror",
			"-pedantic",
			"-fsyntax-only",
			"-I",
			directory,
			quiet,
		}
		if sanitizers != "" {
			arguments = append([]string{sanitizers}, arguments...)
		}
		output, combinedOutputError := exec.CommandContext(ctx, "clang", arguments...).CombinedOutput()
		cancel()
		if combinedOutputError != nil {
			test.Errorf(
				"%q: the header alone doesn't build under -Werror: %v %s",
				sanitizers,
				combinedOutputError,
				output,
			)
		}
	}
}
