package native

import "testing"

// A cache is shared by callers with different shapes. Both the own-closure and
// prototype-method cases must survive a miss, repeated hits and a shape change.
func TestCachedCalleeHitAndFallback(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <assert.h>
static adamic_value call(adamic_object *self, adamic_value *args) {
 (void)self; (void)args; return (adamic_value){.number = 29};
}
static const char *const names[] = {"other", "run"};
static const bool refs[] = {true, true};
static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_reference};
#ifdef ADAMIC_CLOSURE_CONVENTION
static const adamic_method_entry code[] = {{.counted = false, .code = call}};
#else
static const adamic_method_entry code[] = {call};
#endif
static const adamic_methods methods = {1, names + 1, code};
static const adamic_shape fields = {2, names, refs, NULL, kinds};
static const adamic_shape prototype = {1, names, refs, &methods, kinds};
int main(void) {
 adamic_slot_cache cache = {0}, field_cache = {0};
 adamic_object *own = adamic_object_new(&fields);
 adamic_object *method = adamic_object_new(&prototype);
 // A non-owning sentinel identifies the selected closure slot without calling it.
 own->slots[1].reference = own;
 for (size_t round = 0; round < 4; round++) {
  for (size_t hit = 0; hit < 3; hit++) {
   adamic_method_entry entry = {0};
   assert(adamic_object_callee(own, "run", &cache, &entry) == (adamic_closure *)own);
   assert(adamic_object_field(own, "run", &field_cache) == &own->slots[1]);
  }
  for (size_t hit = 0; hit < 3; hit++) {
   adamic_method_entry entry = {0};
   assert(adamic_object_callee(method, "run", &cache, &entry) == NULL);
#ifdef ADAMIC_CLOSURE_CONVENTION
   assert(adamic_method_call(entry, method, NULL, 0).number == 29);
#else
   assert(entry(method, NULL).number == 29);
#endif
   assert(adamic_object_field(method, "other", &field_cache) == &method->slots[0]);
  }
 }
 own->slots[1].reference = NULL;
 adamic_release(own); adamic_release(method);
 return 0;
}
`
	graphRuntimeGuard(t, source, "object.c", "return object->slots[slot].reference;\n\t}\n\t*method = shape->methods->code[slot - shape->count];\n\treturn NULL;\n}\n\n// A field", "return NULL;\n\t}\n\t*method = shape->methods->code[slot - shape->count];\n\treturn NULL;\n}\n\n// A field", "Assertion")
}
