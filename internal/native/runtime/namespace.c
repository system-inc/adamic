// Namespaces share one ordered export table and a hidden readonly-member table.
#include "namespace.h"
#include <stdio.h>
#include <string.h>
static const char *const names[] = {"0", "1"};
static const bool references[] = {true, true};
static const adamic_shape shape = {2, names, references, NULL};
bool adamic_namespace_is(const adamic_object *object) { return object != NULL && object->shape == &shape; }
adamic_object *adamic_namespace_new(void) {
 adamic_object *object = adamic_object_new(&shape);
 object->slots[0].reference = adamic_map_new(true, true);
 object->slots[1].reference = adamic_map_new(true, false);
 return object;
}
void adamic_namespace_install(adamic_object *object, adamic_string *key, adamic_value value, bool readonly) {
 adamic_record_define(object, key, value);
 if (readonly) adamic_map_set(object->slots[1].reference, (adamic_value){.reference = key}, (adamic_value){.boolean = true});
}
void adamic_namespace_check_write(const adamic_object *object, const adamic_string *key) {
 if (!adamic_namespace_is(object)) return;
 if (adamic_map_get(object->slots[1].reference, (adamic_value){.reference = (void *)key}) != NULL) {
  char message[512];
  int length = snprintf(message, sizeof message, "namespace write failed: %.*s is readonly", (int)(key->length < 400 ? key->length : 400), key->bytes);
  adamic_panic(message, (size_t)length);
 }
}
adamic_value adamic_namespace_read(const adamic_object *object, adamic_string *key, unsigned char wanted) {
 if (object == NULL) adamic_panic("namespace read on undefined", sizeof "namespace read on undefined" - 1);
 const adamic_value *slot;
 if (adamic_namespace_is(object)) slot = adamic_record_get_own(object, key);
 else {
  adamic_slot_cache cache = {NULL, 0};
  // Generated constant keys are NUL terminated.
  slot = adamic_object_optional_field(object, key->bytes, &cache);
  if (slot != NULL) {
   if (wanted == 10) {
    unsigned char actual = adamic_object_field_types(object)[cache.index];
    if (actual == 1) return (adamic_value){.reference = adamic_box_number(slot->number)};
    if (actual == 2) return (adamic_value){.reference = slot->boolean ? &adamic_box_true : &adamic_box_false};
   } else if (adamic_object_field_types(object)[cache.index] != 10 && (wanted <= 2 || wanted == 7 || wanted == 9)) return *slot;
  }
 }
 const adamic_heap *value = slot == NULL ? NULL : slot->reference;
 if (wanted == 10) return (adamic_value){.reference = adamic_retain((void *)value)};
 if (wanted == 7 && (value == NULL || value->kind == adamic_kind_number)) return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){value != NULL, value == NULL ? 0 : ((const adamic_number_box *)value)->number})};
 if (wanted == 9 && (value == NULL || value->kind == adamic_kind_boolean)) return (adamic_value){.maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){value != NULL, value != NULL && ((const adamic_boolean_box *)value)->boolean})};
 if (wanted == 1 && value != NULL && value->kind == adamic_kind_number) return (adamic_value){.number = ((const adamic_number_box *)value)->number};
 if (wanted == 2 && value != NULL && value->kind == adamic_kind_boolean) return (adamic_value){.boolean = ((const adamic_boolean_box *)value)->boolean};
 if (((wanted >= 3 && wanted <= 6) || wanted == 8) && (value == NULL || value->kind == (wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : wanted == 6 ? adamic_kind_map : adamic_kind_closure))) return (adamic_value){.reference = adamic_retain((void *)value)};
 char message[512];
 int length = snprintf(message, sizeof message, "namespace read failed: %.*s has an unexpected runtime type", (int)(key->length < 400 ? key->length : 400), key->bytes);
 adamic_panic(message, (size_t)length);
}
