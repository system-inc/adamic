package native

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// Copies retain the same canonical immutable shape after the original object is
// gone. Repeating a replacement must not grow a chain of distinct cached layouts.
func TestObjectSpreadExtendedLayouts(t *testing.T) {
	const source = `#include "object_spread_extend.h"
static const char *const base_names[] = {"count"};
static const bool base_refs[] = {false};
static const adamic_shape base = {1, base_names, base_refs, NULL};
static const char *const extra_names[] = {"text"};
static const bool extra_refs[] = {true};
static const adamic_shape extra = {1, extra_names, extra_refs, NULL};
int main(void) {
 adamic_object *first = adamic_object_new(&base);
 first->slots[0].number = 7;
 adamic_object *held = adamic_object_spread_extend(first, &extra);
 adamic_release(first);
 const adamic_shape *canonical = held->shape;
 for (size_t iteration = 0; iteration < 1000; iteration++) {
  adamic_object *copy = adamic_object_copy(held);
  adamic_release(held);
  if (copy->slots[0].number != 7) return 1;
  if (adamic_object_spread_field_shape(copy->shape, "count") != &base) return 2;
  if (adamic_object_spread_field_shape(copy->shape, "text") != &extra) return 3;
  held = adamic_object_spread_extend(copy, &extra);
  adamic_release(copy);
  if (held->shape != canonical) return 4;
 }
 adamic_release(held);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "spread-layouts")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("layout lifetime and canonicalization: %v: %s", err, output)
	}
}
