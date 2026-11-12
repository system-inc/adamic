package native

import (
	"path/filepath"
	"testing"
)

// Error construction tags remain strings to checked and dynamic readers (#z4a9p9e).
func TestErrorConstructionBytesReadAsStrings(t *testing.T) {
	t.Parallel()
	const code = `#include "adamic.h"
#include "view_unions_mixed.h"
#include <stdio.h>
static adamic_string message = ADAMIC_STRING("x");
int main(void) {
 adamic_object *error = adamic_error_new(&message);
 if (adamic_object_field_types(error)[0] != adamic_error_unwritten || adamic_object_field_types(error)[1] != adamic_error_unwritten) return 1;
 if (adamic_object_field_type(error, 0) != adamic_ir_String || adamic_object_field_type(error, 1) != adamic_ir_String) return 2;
 adamic_slot_cache cache = {0};
 adamic_value name = adamic_object_view(error, "name", &cache, adamic_ir_String, "string", "error.name");
 adamic_object_view_write(error, "name", &cache, adamic_ir_String, "string", "error.name");
 adamic_view_union_value snapshot = adamic_object_view_union_snapshot(error, "name", &cache, "error.name", "string | number", false);
 if (snapshot.kind != adamic_view_union_string || snapshot.payload.reference != name.reference) return 3;
 adamic_heap *dynamic = adamic_dynamic_property(&error->heap, "name");
 if (dynamic != name.reference) return 4;
 adamic_release(dynamic);
 adamic_release(error);
 puts("construction bytes read as strings");
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "error-bytes")
	if err := Build(code, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != "construction bytes read as strings\n" {
		t.Fatalf("got %q", got)
	}
}
