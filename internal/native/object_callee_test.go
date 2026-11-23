package native

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// An unpublished slot does not supply a callee, even with initialized storage.
// Publishing that same slot makes it visible through the same lookup cache.
func TestObjectCalleeAbsentSlot(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <stdio.h>

static adamic_value greet(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	(void)arguments;
	return (adamic_value){0};
}

int main(void) {
	static const char *const names[] = {"greet"};
	static const bool references[] = {true};
	static const adamic_field_kind kinds[] = {adamic_field_reference};
	static const adamic_shape shape = {1, names, references, NULL, kinds};

	adamic_object *object = adamic_object_construct(&shape);
	adamic_closure *closure = adamic_closure_new(greet, 0);
	object->slots[0].reference = closure;
	adamic_object_initialized(object)[0] = 1;

	adamic_slot_cache cache = {0};
	adamic_method_entry method = NULL;
	if (adamic_object_callee(object, "greet", &cache, &method) != NULL || method != NULL) {
		fputs("unpublished greet slot supplied a callee\n", stderr);
		return 1;
	}
	adamic_object_publish(object, 0);
	if (adamic_object_callee(object, "greet", &cache, &method) != closure || method != NULL) {
		fputs("published greet slot did not supply its closure\n", stderr);
		return 2;
	}

	adamic_release(object);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "object-callee")
	if err := Build(harness, binary, native_runtime.Options{}); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := recordOutput(exec.CommandContext(t.Context(), binary))
	if err != nil || stdout != "" || stderr != "" {
		t.Fatalf("callee lookup: %v\nstdout %q\nstderr %q", err, stdout, stderr)
	}
}
