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
func parallelScaling(t *testing.T, options Options) {
	t.Helper()
	binary := parallelHarness(t, "scaling.c", options)
	for _, threads := range []string{"1", "4"} {
		out, _ := parallelRun(t, binary, threads)
		if out != "scaling clean\n" {
			t.Fatal(out)
		}
		parallelLeaks(t, binary, options, threads)
	}
	command := exec.Command(binary, "high_pointer")
	command.Env = parallelEnvironment("1", false)
	out, err := command.CombinedOutput()
	failure, ok := err.(*exec.ExitError)
	if !ok || failure.ExitCode() != 70 || !strings.Contains(string(out), "shape address exceeds 48 bits") {
		t.Fatalf("pointer guard: %v: %s", err, out)
	}
}

func TestParallelScaling_asan(t *testing.T) {
	t.Parallel()
	parallelScaling(t, Options{Sanitize: true})
}

func TestParallelScaling_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelScaling(t, Options{Sanitize: true, slabs: true})
}

func TestParallelScaling_count(t *testing.T) {
	t.Parallel()
	parallelScaling(t, Options{Count: true})
}

func TestParallelScaling_release(t *testing.T) {
	t.Parallel()
	parallelScaling(t, Options{})
}

func TestParallelScaling_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelScaling(t, Options{ThreadSanitize: true})
}

// The snapshot hook stops every builder before publication, forcing the CAS
// losing path without adding synchronization or instrumentation to the runtime.
func scalingSnapshot(t *testing.T, harness, file string, rewrite func(string) string, options Options) string {
	t.Helper()
	directory := t.TempDir()
	files, err := fs.ReadDir(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range files {
		data, err := fs.ReadFile(runtime, "runtime/"+entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			data = []byte(rewrite(string(data)))
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	library, err := RuntimeLibrary(directory, options)
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join("testdata", "parallel", harness))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "main.c")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
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
		t.Fatalf("snapshot must compile: %v\n%s", err, output)
	}
	return binary
}

func cacheBuilderGate(source string) string {
	return strings.Replace(source, "struct adamic_string_index *expected = index;", "extern void adamic_test_cache_built(void); adamic_test_cache_built(); struct adamic_string_index *expected = index;", 1)
}

// One top-level test per build: each builds a whole runtime from an edited snapshot, uncached,
// and all five in one leaf ran 60.7 s cold on a 4-CPU share.
func parallelCachePublication(t *testing.T, options Options) {
	binary := scalingSnapshot(t, "cache_race.c", "string_index.c", cacheBuilderGate, options)
	output, _ := parallelRun(t, binary, "4")
	if output != "cache publication clean\n" {
		t.Fatal(output)
	}
	parallelLeaks(t, binary, options, "4")
}

func TestParallelCachePublication_asan(t *testing.T) {
	t.Parallel()
	parallelCachePublication(t, Options{Sanitize: true})
}

func TestParallelCachePublication_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelCachePublication(t, Options{Sanitize: true, slabs: true})
}

func TestParallelCachePublication_count(t *testing.T) {
	t.Parallel()
	parallelCachePublication(t, Options{Count: true})
}

func TestParallelCachePublication_release(t *testing.T) {
	t.Parallel()
	parallelCachePublication(t, Options{})
}

func TestParallelCachePublication_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelCachePublication(t, Options{ThreadSanitize: true})
}

func TestParallelScalingGuardMutants(t *testing.T) {
	t.Parallel()
	t.Run("pointer_guard", func(t *testing.T) {
		binary := scalingSnapshot(t, "scaling.c", "object.c", func(source string) string {
			return strings.Replace(source, "if ((pointer & ~ADAMIC_SLOT_SHAPE_MASK) != 0)", "if (false)", 1)
		}, Options{})
		command := exec.Command(binary, "high_pointer")
		command.Env = parallelEnvironment("1", false)
		output, err := command.CombinedOutput()
		failure, ok := err.(*exec.ExitError)
		if !ok || failure.ExitCode() == 70 || strings.Contains(string(output), "shape address exceeds 48 bits") {
			t.Fatalf("pointer mutant did not defeat the guard assertion: %v: %s", err, output)
		}
		t.Logf("pointer_guard caught by expected panic status/message: %v", err)
	})
	t.Run("loser_free", func(t *testing.T) {
		if goruntime.GOOS != "linux" {
			t.Skip("LeakSanitizer proof is Linux-only; Darwin positive control runs under leaks")
		}
		binary := scalingSnapshot(t, "cache_race.c", "string_index.c", func(source string) string {
			return strings.Replace(cacheBuilderGate(source), "free(candidate->view);\n\tfree(candidate);", "/* mutant: leak the losing copy */", 1)
		}, Options{Sanitize: true})
		command := exec.Command(binary)
		command.Env = parallelEnvironment("4", true)
		output, err := command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "LeakSanitizer: detected memory leaks") {
			t.Fatalf("loser free mutant not caught: %v\n%s", err, output)
		}
		t.Logf("loser_free caught by LeakSanitizer: %v", err)
	})
}

// Each build retains the complete four-reader cache-pair stress test.
func parallelSlotPairs(t *testing.T, options Options) {
	t.Helper()
	binary := parallelHarness(t, "cache_pairs.c", options)
	out, _ := parallelRun(t, binary, "4")
	if out != "cache pairs clean\n" {
		t.Fatal(out)
	}
	parallelLeaks(t, binary, options, "4")
}

func TestParallelSlotPairs_asan(t *testing.T) {
	t.Parallel()
	parallelSlotPairs(t, Options{Sanitize: true})
}

func TestParallelSlotPairs_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelSlotPairs(t, Options{Sanitize: true, slabs: true})
}

func TestParallelSlotPairs_count(t *testing.T) {
	t.Parallel()
	parallelSlotPairs(t, Options{Count: true})
}

func TestParallelSlotPairs_release(t *testing.T) {
	t.Parallel()
	parallelSlotPairs(t, Options{})
}

func TestParallelSlotPairs_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelSlotPairs(t, Options{ThreadSanitize: true})
}
