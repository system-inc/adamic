package native

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// A checked field read handed a receiver of another heap kind stops at the receiver's kind, before
// reading a shape the receiver doesn't have (#zpzm9ta). The harness hands adamic_object_view a
// string. Mutant (kind_check_removed): object.c without the receiver's kind check, under which the
// string's length is read as its shape and the read goes through it, a sanitizer crash and never
// the clean exit 70 naming the string.
func TestObjectViewStopsAtANonObjectReceiver(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include "view_representations.h"

static adamic_string text = ADAMIC_STRING("ab");

int main(void) {
	adamic_slot_cache cache = {0};
	(void)adamic_object_view((const adamic_object *)(const void *)&text, "name", &cache, adamic_rep_string, "Named", "value.name");
	adamic_heap_end();
	return 0;
}
`
	const want = "adamic: panic: field read failed: value.name is not a Named; expected Named, found string\n"
	flags := native_runtime.Flags(native_runtime.Options{Sanitize: true})
	compiler, library := runtimeWithFlags(t, flags)
	flags = append(flags, "-I", filepath.Dir(library))
	for _, mutant := range []bool{false, true} {
		name := "guarded"
		archive := library
		if mutant {
			name = "kind_check_removed"
			const guard = "\tif (object != NULL && object->heap.kind != adamic_kind_object) { view_failure(type, expression, heap_kind_name(object->heap.kind)); }\n"
			archive = runtimeSourceMutant(t, compiler, library, flags, "object.c", guard, "")
		}
		t.Run(name, func(t *testing.T) {
			binary := buildHarnessWithLibrary(t, harness, flags, compiler, archive)
			stdout, stderr, err := recordRun(binary)
			exit, exited := err.(*exec.ExitError)
			matches := exited && exit.ExitCode() == 70 && stdout == "" && stderr == want
			if !mutant {
				if !matches {
					t.Fatalf("want exit 70 and %q, got %v\nstdout %q\nstderr %s", want, err, stdout, stderr)
				}
				return
			}
			if matches || err == nil {
				t.Fatalf("receiver kind mutant must fail other than the pin's exit 70, got %v\nstdout %q\nstderr %s", err, stdout, stderr)
			}
			t.Logf("receiver kind mutant compiled; the exit-70 pin rejected it: %v %s", err, firstSanitizerLine(stderr))
		})
	}
}
