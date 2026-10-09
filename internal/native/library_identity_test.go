package native

import (
	"path/filepath"
	"testing"
)

// Object readers must see JSON's actual empty layout, while the six constructor
// identities must have the complete closure layout read by identity comparisons.
func TestGenericValueLibraryIdentityViews(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
int main(void) {
 adamic_object *json = adamic_library_identity(4);
 if (json->heap.kind != adamic_kind_object || json->heap.references != 0 || json->heap.slab != 0 || json->shape == NULL || json->shape->count != 0) return 1;
 adamic_array *keys = adamic_plain_object_keys(json);
 printf("object %zu", keys->length);
 adamic_release(keys);
 for (size_t i = 0; i < 7; i++) {
  if (i == 4) continue;
  adamic_closure *identity = adamic_library_identity(i);
  if (identity->heap.kind != adamic_kind_closure || identity->count != 0 || identity->source_identity != 0 || adamic_reference_identity(identity) != (uintptr_t)identity) return 2;
  printf(" function");
 }
 printf("\n");
 return 0;
}
`
	want := runWithInput(t, "", "node", "--eval", `console.log(typeof JSON + " " + Object.keys(JSON).length + " " + [Number,String,Boolean,Object,Array,RegExp].map(value => typeof value).join(" "));`)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "library-identities")
		if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: %q; Node %q", sanitize, got, want)
		}
	}
}
