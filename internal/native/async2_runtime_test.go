package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

// Each leaf builds one control and one runtime mutant, without compiler lowering.
func async2RuntimeGuard(test *testing.T, mode, file, old, replacement, diagnostic string) {
	test.Helper()
	options := Options{Sanitize: true}
	files, flags, compiler, version, runtimeInputsError := runtimeInputs(
		"",
		options,
		[]string{"-DADAMIC_CANONICAL_CLOSURES=1"},
	)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	cache, cacheDirectoryError := cacheDirectory("async2-runtime")
	if cacheDirectoryError != nil {
		test.Fatal(cacheDirectoryError)
	}
	source, readError := os.ReadFile("testdata/async2/runtime.c")
	if readError != nil {
		test.Fatal(readError)
	}
	for _, mutant := range []bool{false, true} {
		if mutant {
			found := false
			for index := range files {
				if files[index].name != file {
					continue
				}
				text := string(files[index].contents)
				if !strings.Contains(text, old) {
					test.Fatalf("missing mutant seam %q", old)
				}
				files[index].contents = []byte(strings.Replace(text, old, replacement, 1))
				found = true
			}
			if !found {
				test.Fatalf("missing runtime file %s", file)
			}
		}
		library, cachedRuntimeError := cachedRuntime(files, flags, compiler, version, cache)
		if cachedRuntimeError != nil {
			test.Fatal(cachedRuntimeError)
		}
		binary := buildHarnessWithLibrary(test, string(source), flags, compiler, library)
		command, release := boundedrun.Command(5*time.Second, binary, mode)
		// A mode that ends at a runtime stop, and the words it stops with.
		stops := map[string]string{
			"share":     "async values cannot cross worker thread",
			"canonical": "compiler bug: a canonical closure over a cell no environment owns",
		}
		leaks := "1"
		if _, stop := stops[mode]; stop {
			leaks = "0"
		} // The refusal intentionally exits with a live frame.
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks="+leaks+":halt_on_error=1")
		logPath := filepath.Join(test.TempDir(), "run.log")
		log, createError := os.Create(logPath)
		if createError != nil {
			test.Fatal(createError)
		}
		command.Stdout, command.Stderr = log, log
		runError := command.Run()
		release()
		if closeError := log.Close(); closeError != nil {
			test.Fatal(closeError)
		}
		output, readError := os.ReadFile(logPath)
		if readError != nil {
			test.Fatal(readError)
		}
		if stop, stops := stops[mode]; stops && !mutant {
			if command.ProcessState == nil ||
				command.ProcessState.ExitCode() != 70 ||
				!strings.Contains(string(output), stop) {
				test.Fatalf("%s guard: %v\n%s", mode, runError, output)
			}
		} else if mutant {
			if runError == nil || !strings.Contains(string(output), diagnostic) {
				test.Fatalf("mutant survived or failed for another reason: %v\n%s", runError, output)
			}
			test.Logf("%s mutant caught by %s", mode, diagnostic)
		} else if runError != nil || len(output) != 0 {
			test.Fatalf("control: %v\n%s", runError, output)
		}
	}
}

func TestAsync2RuntimeCancel(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"cancel",
		"async.c",
		"if (break_cycle && cleanup != NULL) cleanup(entry->frame);",
		"if (false && break_cycle && cleanup != NULL) cleanup(entry->frame);",
		"Assertion",
	)
}
func TestAsync2RuntimeTeardown(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"exit",
		"async.c",
		"if (break_cycle && cleanup != NULL) cleanup(entry->frame);",
		"if (false && break_cycle && cleanup != NULL) cleanup(entry->frame);",
		"Assertion",
	)
}
func TestAsync2RuntimeForget(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"forget",
		"async.c",
		"void adamic_async_forget_cleanup(adamic_async_frame *frame) { (void)take_cleanup(frame); }",
		"void adamic_async_forget_cleanup(adamic_async_frame *frame) { (void)frame; }",
		"Assertion",
	)
}
func TestAsync2RuntimeEnvironment(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"environment",
		"closure.c",
		"if (cells[index].references) { drop(cells[index].value.reference); }",
		"if (false && cells[index].references) { drop(cells[index].value.reference); }",
		"Assertion",
	)
}

// A canonical closure's cell must be an environment's: the cache is read and written through it as
// one, so an async frame's or a generator's cell would have its words overwritten (Memory, on
// #qk8rztp S1a). Mutant: canonical() takes the owner unchecked, and the harness's call returns.
func TestAsync2RuntimeCanonicalOwner(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"canonical",
		"closure.c",
		"\tadamic_environment *owner = canonical_owner(identity);",
		"\tadamic_environment *owner = (adamic_environment *)identity->owner;",
		"Assertion",
	)
}

// counted_canonical() takes its owner through the same helper, and this harness doesn't define
// ADAMIC_CLOSURE_CONVENTION, so its call site isn't built here. Mutant: the helper's check itself
// dropped, which both callers share, and the harness's call returns.
func TestAsync2RuntimeCanonicalOwnerHelper(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"canonical",
		"closure.c",
		"if (identity->owner == NULL || identity->owner->kind != adamic_kind_environment) {",
		"if (false) {",
		"Assertion",
	)
}

func TestAsync2RuntimeShare(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(test, "share", "share.c", "value = ((adamic_cell *)value)->owner;", "(void)heap;", "Assertion")
}

func TestAsync2RuntimeOwnership(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"environment",
		"heap.c",
		"heap = ((adamic_cell *)heap)->owner;",
		"(void)heap;",
		"Assertion",
	)
}
func TestAsync2RuntimeInitialization(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"environment",
		"closure.c",
		".owner = owner",
		".owner = (owner == NULL ? owner : NULL)",
		"Assertion",
	)
}

func TestAsync2RuntimeFreeChildren(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(test, "free", "async.c", "adamic_async_forget_cleanup(frame);", "(void)frame;", "Assertion")
}

func TestAsync2RuntimeRelease(test *testing.T) {
	test.Parallel()
	async2RuntimeGuard(
		test,
		"environment",
		"heap.c",
		"heap = ((adamic_cell *)heap)->owner;\n\t}\n#endif\n\tif (heap == NULL || adamic_program_is(heap))",
		"(void)heap;\n\t}\n#endif\n\tif (heap == NULL || adamic_program_is(heap))",
		"Assertion",
	)
}
