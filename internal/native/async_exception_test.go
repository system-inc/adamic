package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestAsyncUnexpectedExceptionStops(t *testing.T) {
	t.Parallel()
	e := emitter{function: &ir.Function{Name: "proved_safe"}, asyncSlots: map[int]int{}}
	e.checkThrown()
	source := "#include \"adamic.h\"\nint main(void) { adamic_thrown = (adamic_object *)1;\n" + e.out.String() + "return 0; }\n"
	binary := filepath.Join(t.TempDir(), "unexpected-exception")
	if err := Build(source, binary, Options{}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || !strings.Contains(string(output), "proved_safe: exception where the flow graph proved none") {
		t.Fatalf("want named exit 70, got %v: %s", err, output)
	}
}

func TestAsyncUnhandledNonReferenceReasonStops(t *testing.T) {
	t.Parallel()
	source := `#include "async.h"
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    adamic_async_promise *promise = adamic_async_new();
    adamic_async_settle(promise, (adamic_value){.number = 1}, false, true);
    adamic_release(promise);
    adamic_async_run();
    return 0;
}
`
	binary := filepath.Join(t.TempDir(), "non-reference-rejection")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	want := "adamic: panic: compiler bug: an unhandled rejection whose reason is not a reference\n"
	if !ok || exit.ExitCode() != 70 || string(output) != want {
		t.Fatalf("want named exit 70, got %v: %s", err, output)
	}
}
