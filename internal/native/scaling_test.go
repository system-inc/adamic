package native

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
)

// Partition by build so the two worker counts and pointer guard share one binary.
func parallelScaling(test *testing.T, options Options) {
	test.Helper()
	binary := parallelHarness(test, "scaling.c", options)
	for _, threads := range []string{"1", "4"} {
		out, _ := parallelRun(test, binary, threads)
		if out != "scaling clean\n" {
			test.Fatal(out)
		}
		parallelLeaks(test, binary, options, threads)
	}
	command := exec.Command(binary, "high_pointer")
	command.Env = parallelEnvironment("1", false)
	out, err := command.CombinedOutput()
	failure, ok := err.(*exec.ExitError)
	if !ok || failure.ExitCode() != 70 || !strings.Contains(string(out), "shape address exceeds 48 bits") {
		test.Fatalf("pointer guard: %v: %s", err, out)
	}
}

func TestParallelScaling_asan(test *testing.T) {
	test.Parallel()
	parallelScaling(test, Options{Sanitize: true})
}

func TestParallelScaling_asan_slabs(test *testing.T) {
	test.Parallel()
	parallelScaling(test, Options{Sanitize: true, slabs: true})
}

func TestParallelScaling_count(test *testing.T) {
	test.Parallel()
	parallelScaling(test, Options{Count: true})
}

func TestParallelScaling_release(test *testing.T) {
	test.Parallel()
	parallelScaling(test, Options{})
}

func TestParallelScaling_tsan(test *testing.T) {
	test.Parallel()
	if !parallelTSan(test) {
		test.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelScaling(test, Options{ThreadSanitize: true})
}

// The snapshot hook stops every builder before publication, forcing the CAS
// losing path without adding synchronization or instrumentation to the runtime.
func scalingSnapshot(test *testing.T, harness, file string, rewrite func(string) string, options Options) string {
	test.Helper()
	directory := test.TempDir()
	files, err := fs.ReadDir(runtime, "runtime")
	if err != nil {
		test.Fatal(err)
	}
	for _, entry := range files {
		data, err := fs.ReadFile(runtime, "runtime/"+entry.Name())
		if err != nil {
			test.Fatal(err)
		}
		if entry.Name() == file {
			data = []byte(rewrite(string(data)))
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); err != nil {
			test.Fatal(err)
		}
	}
	library, err := RuntimeLibrary(directory, options)
	if err != nil {
		test.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join("testdata", "parallel", harness))
	if err != nil {
		test.Fatal(err)
	}
	path := filepath.Join(directory, "main.c")
	if err := os.WriteFile(path, source, 0600); err != nil {
		test.Fatal(err)
	}
	name := "harness"
	if options.ThreadSanitize {
		name += "-tsan"
	}
	binary := filepath.Join(directory, name)
	arguments := append(Flags(options), "-I", filepath.Dir(library), path, "-o", binary)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		test.Fatalf("snapshot must compile: %v\n%s", err, output)
	}
	return binary
}

func cacheBuilderGate(source string) string {
	return strings.Replace(
		source,
		"struct adamic_string_index *expected = index;",
		"extern void adamic_test_cache_built(void); adamic_test_cache_built(); struct adamic_string_index *expected = index;",
		1,
	)
}

// One top-level test per build: each builds a whole runtime from an edited snapshot, uncached,
// and all five in one leaf ran 60.7 s cold on a 4-CPU share.
func parallelCachePublication(test *testing.T, options Options) {
	binary := scalingSnapshot(test, "cache_race.c", "string_index.c", cacheBuilderGate, options)
	output, _ := parallelRun(test, binary, "4")
	if output != "cache publication clean\n" {
		test.Fatal(output)
	}
	parallelLeaks(test, binary, options, "4")
}

func TestParallelCachePublication_asan(test *testing.T) {
	test.Parallel()
	parallelCachePublication(test, Options{Sanitize: true})
}

func TestParallelCachePublication_asan_slabs(test *testing.T) {
	test.Parallel()
	parallelCachePublication(test, Options{Sanitize: true, slabs: true})
}

func TestParallelCachePublication_count(test *testing.T) {
	test.Parallel()
	parallelCachePublication(test, Options{Count: true})
}

func TestParallelCachePublication_release(test *testing.T) {
	test.Parallel()
	parallelCachePublication(test, Options{})
}

func TestParallelCachePublication_tsan(test *testing.T) {
	test.Parallel()
	if !parallelTSan(test) {
		test.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelCachePublication(test, Options{ThreadSanitize: true})
}

func TestParallelScalingGuardMutants(test *testing.T) {
	test.Parallel()
	test.Run("pointer_guard", func(test *testing.T) {
		binary := scalingSnapshot(test, "scaling.c", "object.c", func(source string) string {
			return strings.Replace(source, "if ((pointer & ~ADAMIC_SLOT_SHAPE_MASK) != 0)", "if (false)", 1)
		}, Options{})
		command := exec.Command(binary, "high_pointer")
		command.Env = parallelEnvironment("1", false)
		output, err := command.CombinedOutput()
		failure, ok := err.(*exec.ExitError)
		if !ok || failure.ExitCode() == 70 || strings.Contains(string(output), "shape address exceeds 48 bits") {
			test.Fatalf("pointer mutant did not defeat the guard assertion: %v: %s", err, output)
		}
		test.Logf("pointer_guard caught by expected panic status/message: %v", err)
	})
	test.Run("loser_free", func(test *testing.T) {
		if goruntime.GOOS != "linux" {
			test.Skip("LeakSanitizer proof is Linux-only; Darwin positive control runs under leaks")
		}
		binary := scalingSnapshot(test, "cache_race.c", "string_index.c", func(source string) string {
			return strings.Replace(
				cacheBuilderGate(source),
				"free(candidate->view);\n\tfree(candidate);",
				"/* mutant: leak the losing copy */",
				1,
			)
		}, Options{Sanitize: true})
		command := exec.Command(binary)
		command.Env = parallelEnvironment("4", true)
		output, err := command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "LeakSanitizer: detected memory leaks") {
			test.Fatalf("loser free mutant not caught: %v\n%s", err, output)
		}
		test.Logf("loser_free caught by LeakSanitizer: %v", err)
	})
}

// Each build retains the complete four-reader cache-pair stress test.
func parallelSlotPairs(test *testing.T, options Options) {
	test.Helper()
	binary := parallelHarness(test, "cache_pairs.c", options)
	out, _ := parallelRun(test, binary, "4")
	if out != "cache pairs clean\n" {
		test.Fatal(out)
	}
	parallelLeaks(test, binary, options, "4")
}

func TestParallelSlotPairs_asan(test *testing.T) {
	test.Parallel()
	parallelSlotPairs(test, Options{Sanitize: true})
}

func TestParallelSlotPairs_asan_slabs(test *testing.T) {
	test.Parallel()
	parallelSlotPairs(test, Options{Sanitize: true, slabs: true})
}

func TestParallelSlotPairs_count(test *testing.T) {
	test.Parallel()
	parallelSlotPairs(test, Options{Count: true})
}

func TestParallelSlotPairs_release(test *testing.T) {
	test.Parallel()
	parallelSlotPairs(test, Options{})
}

func TestParallelSlotPairs_tsan(test *testing.T) {
	test.Parallel()
	if !parallelTSan(test) {
		test.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelSlotPairs(test, Options{ThreadSanitize: true})
}
