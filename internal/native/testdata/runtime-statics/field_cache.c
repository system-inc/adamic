#include "harness.h"
static const char *const first_names[] = {"key", "value"};
static const char *const second_names[] = {"value", "key"};
static const bool references[] = {false, false};
static const adamic_field_kind shapes_kinds_0[] = {adamic_field_number, adamic_field_number};
static const adamic_field_kind shapes_kinds_1[] = {adamic_field_number, adamic_field_number};
static const adamic_shape shapes[] = {{2, first_names, references, NULL, shapes_kinds_0}, {2, second_names, references, NULL, shapes_kinds_1}};
static adamic_slot_cache cache;
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 adamic_object *object = adamic_object_new(&shapes[index % 2]);
 object->slots[index % 2].number = 17;
 object->slots[1 - index % 2].number = 23;
 double key = adamic_object_field(object, "key", &cache)->number;
 if (key != 17) { abort(); }
 adamic_release(object);
 return 1;
}
