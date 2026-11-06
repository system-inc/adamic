package native

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

// mapSetRuntimeRun runs one mode of testdata/map_set/runtime.c, a one-entry map's values() stepped by
// hand against the runtime with no compiler lowering. The build is sanitized and counted, as the
// generator harness's is, so a retain of what isn't a pointer stops under the sanitizers and a leak
// shows in the counts on every platform.
func mapSetRuntimeRun(t *testing.T, mode string) (exit int, stdout, stderr string) {
	t.Helper()
	files, flags, compiler, version, err := runtimeInputs("", Options{Sanitize: true, Count: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("map-set-runtime")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("testdata/map_set/runtime.c")
	if err != nil {
		t.Fatal(err)
	}
	library, err := cachedRuntime(files, flags, compiler, version, cache)
	if err != nil {
		t.Fatal(err)
	}
	binary := buildHarnessWithLibrary(t, string(source), flags, compiler, library)
	command, release := boundedrun.Command(5*time.Second, binary, mode)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1")
	var output, diagnostics strings.Builder
	command.Stdout, command.Stderr = &output, &diagnostics
	_ = command.Run()
	release()
	exit = -1
	if command.ProcessState != nil {
		exit = command.ProcessState.ExitCode()
	}
	return exit, output.String(), diagnostics.String()
}

// A counted element steps once through adamic_iterator_result, its result holding its own reference,
// and then the iterator is done, with nothing leaked. It's the control for the stop below: the same
// harness, with a representation a collection holds.
func TestCollectionIteratorStepsCountedElement(t *testing.T) {
	t.Parallel()
	exit, stdout, stderr := mapSetRuntimeRun(t, "string")
	if exit != 0 || stdout != "7\ndone\n" {
		t.Fatalf("exit %d, stdout %q\n%s", exit, stdout, stderr)
	}
	if leaked := generatorLeaked(stderr); leaked != "" {
		t.Fatalf("%s\n%s", leaked, stderr)
	}
}

// A collection's element is a number, a boolean, a maybe-number or counted, and any other
// representation stops as a compiler bug rather than be retained as the pointer it isn't (map_set.c's
// collection_reference, #ak92fcp). The lowering keeps every other one out of a collection, so only
// this harness hands the runtime a MaybeBoolean element. Mutant, run by hand on map_set.c:
// collection_reference's old body, which calls everything but Number, Boolean and MaybeNumber
// counted; the step then retains address 1 and the sanitizer stops it without these words
// (internal/oracle/testdata/iterator_result_builder/mutants.md).
func TestCollectionElementOfAnotherRepresentationStops(t *testing.T) {
	t.Parallel()
	exit, stdout, stderr := mapSetRuntimeRun(t, "maybe_boolean")
	want := "adamic: panic: compiler bug: a collection element of representation 9 is neither"
	if exit != 70 || stdout != "" || !strings.Contains(stderr, want) {
		t.Fatalf("exit %d, stdout %q, want %q on stderr\n%s", exit, stdout, want, stderr)
	}
}
