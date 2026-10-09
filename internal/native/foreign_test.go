package native

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// Foreign payload slots stay private to the bridge under main's object layout.
func TestForeignObjectLifetime(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <assert.h>
static int pointer, releases;
static void dropped(void *value) { assert(value == &pointer); releases++; }
static const adamic_foreign_kind kind = {"test", dropped};
int main(void) {
 adamic_object *object = adamic_foreign_new(&pointer, &kind);
 assert(object->shape->count == 0 && object->class == NULL);
 assert(object->frozen && !object->tuple);
 assert(adamic_foreign_pointer(object) == &pointer);
 adamic_array *keys = adamic_object_keys(object);
 assert(keys->length == 0);
 adamic_release(keys);
 adamic_retain(object);
 adamic_release(object);
 assert(releases == 0);
 adamic_release(object);
 assert(releases == 1);
 assert(adamic_foreign_pointer(NULL) == NULL);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "foreign")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("foreign lifetime: %v\n%s", err, output)
	}
}
