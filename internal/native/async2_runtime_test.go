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
func async2RuntimeGuard(t *testing.T, mode, file, old, replacement, diagnostic string) {
	t.Helper()
	options := Options{Sanitize: true}
	files, flags, compiler, version, err := runtimeInputs("", options, []string{"-DADAMIC_CANONICAL_CLOSURES=1"})
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("async2-runtime")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("testdata/async2/runtime.c")
	if err != nil {
		t.Fatal(err)
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
					t.Fatalf("missing mutant seam %q", old)
				}
				files[index].contents = []byte(strings.Replace(text, old, replacement, 1))
				found = true
			}
			if !found {
				t.Fatalf("missing runtime file %s", file)
			}
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			t.Fatal(err)
		}
		binary := buildHarnessWithLibrary(t, string(source), flags, compiler, library)
		command, release := boundedrun.Command(5*time.Second, binary, mode)
		leaks := "1"
		if mode == "share" {
			leaks = "0"
		} // The refusal intentionally exits with a live frame.
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks="+leaks+":halt_on_error=1")
		logPath := filepath.Join(t.TempDir(), "run.log")
		log, err := os.Create(logPath)
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout, command.Stderr = log, log
		err = command.Run()
		release()
		if closeErr := log.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		output, readErr := os.ReadFile(logPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if mode == "share" && !mutant {
			if command.ProcessState == nil || command.ProcessState.ExitCode() != 70 || !strings.Contains(string(output), "async values cannot cross worker thread") {
				t.Fatalf("share guard: %v\n%s", err, output)
			}
		} else if mutant {
			if err == nil || !strings.Contains(string(output), diagnostic) {
				t.Fatalf("mutant survived or failed for another reason: %v\n%s", err, output)
			}
			t.Logf("%s mutant caught by %s", mode, diagnostic)
		} else if err != nil || len(output) != 0 {
			t.Fatalf("control: %v\n%s", err, output)
		}
	}
}

func TestAsync2RuntimeCancel(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "cancel", "async.c", "if (break_cycle && cleanup != NULL) cleanup(entry->frame);", "if (false && break_cycle && cleanup != NULL) cleanup(entry->frame);", "Assertion")
}
func TestAsync2RuntimeTeardown(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "exit", "async.c", "if (break_cycle && cleanup != NULL) cleanup(entry->frame);", "if (false && break_cycle && cleanup != NULL) cleanup(entry->frame);", "Assertion")
}
func TestAsync2RuntimeForget(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "forget", "async.c", "void adamic_async_forget_cleanup(adamic_async_frame *frame) { (void)take_cleanup(frame); }", "void adamic_async_forget_cleanup(adamic_async_frame *frame) { (void)frame; }", "Assertion")
}
func TestAsync2RuntimeEnvironment(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "environment", "closure.c", "if (cells[index].references) { drop(cells[index].value.reference); }", "if (false && cells[index].references) { drop(cells[index].value.reference); }", "Assertion")
}
func TestAsync2RuntimeShare(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "share", "share.c", "value = ((adamic_cell *)value)->owner;", "(void)heap;", "Assertion")
}

func TestAsync2RuntimeOwnership(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "environment", "heap.c", "heap = ((adamic_cell *)heap)->owner;", "(void)heap;", "Assertion")
}
func TestAsync2RuntimeInitialization(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "environment", "closure.c", ".owner = owner", ".owner = (owner == NULL ? owner : NULL)", "Assertion")
}

func TestAsync2RuntimeFreeChildren(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "free", "async.c", "adamic_async_forget_cleanup(frame);", "(void)frame;", "Assertion")
}

func TestAsync2RuntimeRelease(t *testing.T) {
	t.Parallel()
	async2RuntimeGuard(t, "environment", "heap.c", "heap = ((adamic_cell *)heap)->owner;\n\t}\n#endif\n\tif (heap == NULL)", "(void)heap;\n\t}\n#endif\n\tif (heap == NULL)", "Assertion")
}
