package native

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
)

var primitiveCases = []string{
	"version-template", "version-add", "declaration-array", "valueOf-first",
	"string-fallback", "objects-throw", "noncallable-skipped", "absent-throw",
	"call-throws", "get-throws", "lookup-after-mutation", "undefined-result",
	"null-result", "boolean-result", "negative-zero-result", "string-result", "number-hint",
	"function-results-throw", "getter-mutates-next", "second-get-throws",
}

func primitiveRun(binary string, index int, environment ...string) primitiveExecution {
	command := exec.Command(binary, fmt.Sprint(index))
	command.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if err != nil {
		code = -1
		if command.ProcessState != nil {
			code = command.ProcessState.ExitCode()
		}
	}
	return primitiveExecution{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: code}
}

func primitiveOracle(t *testing.T, index int) []byte {
	t.Helper()
	got, err := exec.Command("node", "testdata/to-primitive/node.a", fmt.Sprint(index)).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, got)
	}
	if index == 5 || index == 7 || index == 17 {
		if !bytes.HasPrefix(got, []byte("TypeError: Cannot convert object to primitive value\n")) {
			t.Fatalf("Node's TypeError changed: %s", got)
		}
	}
	return got
}

func TestOrdinaryToPrimitive(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/to-primitive/runtime.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, slabs := range []bool{false, true} {
		options := Options{Sanitize: true, slabs: slabs, Count: true}
		binary := filepath.Join(t.TempDir(), "primitive")
		if err := Build(string(source), binary, options); err != nil {
			t.Fatal(err)
		}
		t.Logf("clang flags: %s", strings.Join(Flags(options), " "))
		for index, name := range primitiveCases {
			t.Run(fmt.Sprintf("slabs-%t/%s", slabs, name), func(t *testing.T) {
				want := primitiveOracle(t, index)
				got := primitiveRun(binary, index, "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				if got.ExitCode != 0 || !bytes.Equal(got.Stdout, want) {
					t.Fatalf("C differs from Node: exit %d\n%s%s\nNode: %s", got.ExitCode, got.Stdout, got.Stderr, want)
				}
				if report := primitiveUnbalanced(got); report != "" {
					t.Fatal(report)
				}
				if report := primitiveLeaks(t, binary, fmt.Sprint(index)); report != "" {
					t.Fatal(report)
				}
			})
		}
	}
}

// Mutants use separate runtime archives; none modifies the checkout or fails compilation.
func primitiveMutant(t *testing.T, source, old, changed string) string {
	t.Helper()
	directory := t.TempDir()
	entries, err := os.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join("runtime", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == "exceptions.c" {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor is no longer unique")
			}
			data = []byte(strings.Replace(string(data), old, changed, 1))
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	options := Options{Sanitize: true, Count: true}
	library, err := RuntimeLibrary(directory, options)
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "primitive")
	arguments := append(Flags(options), "-I", filepath.Dir(library), "-o", binary, main)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if got, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("mutant must compile: %v %s", err, got)
	}
	return binary
}

func TestOrdinaryToPrimitiveMutants(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/to-primitive/runtime.c")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("wrong-order", func(t *testing.T) {
		binary := primitiveMutant(t, string(source), "if (hint == adamic_hint_string)", "if (hint != adamic_hint_string)")
		for index, name := range primitiveCases {
			t.Run(name, func(t *testing.T) {
				got := primitiveRun(binary, index, "ASAN_OPTIONS=detect_leaks=1")
				if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
					t.Fatalf("behavior mutant must finish leak-clean: %+v", got)
				}
				if bytes.Equal(got.Stdout, primitiveOracle(t, index)) {
					t.Fatal("wrong order survived")
				}
				t.Log("wrong-order mutant caught by Node output and lookup/call trace")
			})
		}
	})
	t.Run("missing-TypeError", func(t *testing.T) {
		binary := primitiveMutant(t, string(source), "adamic_thrown = error;", "adamic_release(error);")
		for _, index := range []int{5, 7, 17} {
			got := primitiveRun(binary, index, "ASAN_OPTIONS=detect_leaks=1")
			if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
				t.Fatalf("behavior mutant must finish leak-clean: %+v", got)
			}
			if bytes.Equal(got.Stdout, primitiveOracle(t, index)) {
				t.Fatal("missing TypeError survived")
			}
			t.Logf("missing-TypeError mutant caught by %s", primitiveCases[index])
		}
	})
	t.Run("leaked-object-result", func(t *testing.T) {
		binary := primitiveMutant(t, string(source), "adamic_release(result.value.reference);", "adamic_release(NULL);")
		got := primitiveRun(binary, 5, "ASAN_OPTIONS=detect_leaks=1", "LSAN_OPTIONS=use_stacks=0:use_registers=0")
		if !bytes.Contains(got.Stderr, []byte("LeakSanitizer: detected memory leaks")) {
			t.Fatalf("leaked object result survived: %+v", got)
		}
		t.Log("rejected-result release mutant caught by LeakSanitizer")
	})
}

// These two runtime ABI guards have no source-level conversion counterpart.
func TestOrdinaryToPrimitiveProtocolGuards(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/to-primitive/runtime.c")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "primitive")
	if err := Build(string(source), binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	t.Run("terminal-invalid-hint", func(t *testing.T) {
		got := primitiveRun(binary, 20)
		if got.ExitCode != 70 || !bytes.HasPrefix(got.Stderr, []byte("adamic: panic: compiler bug: invalid primitive hint\n")) {
			t.Fatalf("invalid hint must be terminal: %+v", got)
		}
		old := "if (hint != adamic_hint_default && hint != adamic_hint_number && hint != adamic_hint_string)"
		mutant := primitiveMutant(t, string(source), old, "if (false)")
		got = primitiveRun(mutant, 20, "ASAN_OPTIONS=detect_leaks=1")
		if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
			t.Fatalf("guard mutant must finish cleanly: %+v", got)
		}
		t.Log("missing invalid-hint guard caught by expected terminal panic")
	})
	t.Run("preserve-pending-error", func(t *testing.T) {
		const want = "Error: kept\n\nafter catch\n"
		got := primitiveRun(binary, 21, "ASAN_OPTIONS=detect_leaks=1")
		if got.ExitCode != 0 || string(got.Stdout) != want || primitiveUnbalanced(got) != "" {
			t.Fatalf("pending error must prevent lookup: %+v", got)
		}
		old := "if (adamic_thrown != NULL) { return empty; }\n\tif (hint"
		mutant := primitiveMutant(t, string(source), old, "if (hint")
		got = primitiveRun(mutant, 21, "ASAN_OPTIONS=detect_leaks=1")
		if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
			t.Fatalf("guard mutant must finish cleanly: %+v", got)
		}
		if string(got.Stdout) == want {
			t.Fatal("pending-error entry guard mutant survived")
		}
		t.Log("missing pending-error guard caught by an unexpected Get in the trace")
	})
}

// Main keeps its leak gates in the oracle package. These C seam witnesses check
// their own counts and use the same OS leak tools without importing that harness.
type primitiveExecution struct {
	Stdout, Stderr []byte
	ExitCode       int
}

var primitiveCounts = regexp.MustCompile(`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`)

func primitiveUnbalanced(run primitiveExecution) string {
	match := primitiveCounts.FindSubmatch(run.Stderr)
	if run.ExitCode != 0 || match == nil {
		return fmt.Sprintf("unfinished counted witness: %+v", run)
	}
	allocations, _ := strconv.ParseUint(string(match[1]), 10, 64)
	frees, _ := strconv.ParseUint(string(match[2]), 10, 64)
	regions, _ := strconv.ParseUint(string(match[6]), 10, 64)
	if allocations != frees+regions {
		return fmt.Sprintf("heap values leaked: %d", allocations-frees-regions)
	}
	return ""
}
func primitiveLeaks(t *testing.T, binary string, arguments ...string) string {
	t.Helper()
	if goruntime.GOOS != "darwin" {
		return ""
	} // Linux runs LSan in primitiveRun.
	command := exec.Command("leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("leaks: %v %s", err, output)
	}
	return ""
}
