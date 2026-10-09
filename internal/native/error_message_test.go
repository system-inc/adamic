package native

import (
	"path/filepath"
	"testing"
)

// Direct own-property queries on Error remain NotYet in lowering. Probe the runtime
// directly to distinguish the inherited empty message from an explicitly supplied empty one.
func TestErrorUndefinedMessageHasNoOwnProperty(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include "view_representations.h"
#include <stdio.h>
static adamic_string empty = ADAMIC_STRING("");
static adamic_string text = ADAMIC_STRING("m");
static adamic_string key = ADAMIC_STRING("message");
int main(void) {
    adamic_string *messages[] = {NULL, &empty, &text, NULL};
    adamic_slot_cache cache = {NULL, 0};
    adamic_slot_cache optional_cache = {NULL, 0};
    const char *kinds[] = {"Error", "RangeError", "TypeError"};
    const adamic_class *classes[] = {&adamic_error_class, &adamic_range_error_class, &adamic_type_error_class};
    for (size_t kind = 0; kind < 3; kind++) {
    for (size_t i = 0; i < 4; i++) {
        adamic_object *error = adamic_error_new_kind(messages[i], kinds[kind]);
        if (error->class != classes[kind]) return 3;
        adamic_string *message = adamic_object_field(error, "message", &cache)->reference;
        printf("[%.*s] %s %s\n", (int)message->length, message->bytes,
            adamic_object_has(error, &key) ? "true" : "false",
            adamic_object_has_own(error, &key) ? "true" : "false");
        adamic_value *optional = adamic_object_optional_field(error, "message", &optional_cache);
        if (optional == NULL || optional->reference != message) return 1;
        adamic_value checked = adamic_object_view(error, "message", &cache, adamic_rep_string, "string", "error.message");
        if (checked.reference != message) return 2;
        adamic_release(error);
    }
    }
    return 0;
}
`
	const oracle = `for (const Constructor of [Error, RangeError, TypeError]) for (const error of [new Constructor(), new Constructor(''), new Constructor('m'), new Constructor(undefined)]) console.log('[' + error.message + '] ' + error.hasOwnProperty('message') + ' ' + Object.hasOwn(error, 'message'));`
	want := runWithInput(t, "", "node", "--eval", oracle)
	const oneKind = "[] false false\n[] true true\n[m] true true\n[] false false\n"
	if want != oneKind+oneKind+oneKind {
		t.Fatalf("Node changed: %q", want)
	}
	binary := filepath.Join(t.TempDir(), "error-message")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("native %q; Node %q", got, want)
	}
}
