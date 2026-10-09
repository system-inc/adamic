#include "harness.h"
// Every worker's first item lands on the same metadata (the pool's chunks start at multiples of 8),
// so four workers register one shape's types at once, then read through the registry.
static const char *const names[] = {"x"};
static const bool references[] = {false};
static const adamic_shape shapes[8] = {
 {1, names, references, NULL}, {1, names, references, NULL}, {1, names, references, NULL}, {1, names, references, NULL},
 {1, names, references, NULL}, {1, names, references, NULL}, {1, names, references, NULL}, {1, names, references, NULL},
};
static const int types[] = {1};
static adamic_shape_types metadata[8] = {
 {&shapes[0], types, NULL, false}, {&shapes[1], types, NULL, false}, {&shapes[2], types, NULL, false}, {&shapes[3], types, NULL, false},
 {&shapes[4], types, NULL, false}, {&shapes[5], types, NULL, false}, {&shapes[6], types, NULL, false}, {&shapes[7], types, NULL, false},
};
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 size_t which = index % 8;
 adamic_register_shape_types(&metadata[which]);
 adamic_object *object = adamic_object_new(&shapes[which]);
 object->slots[0].number = (double)which;
 adamic_heap *value = adamic_dynamic_property(&object->heap, "x");
 // A registry missing this shape (a lost push) panics; one that links an entry to itself never returns.
 if (value == NULL || value->kind != adamic_kind_number || ((adamic_number_box *)value)->number != (double)which) { abort(); }
 adamic_release(value);
 adamic_release(object);
 return 1;
}
