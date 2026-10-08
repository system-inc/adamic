#include "view_unions_object_primitive.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_heap *adamic_object_primitive_view(const adamic_object *object, const char *field, adamic_slot_cache *cache, const adamic_object_primitive_member *members, size_t count, bool undefined, const char *declared, const char *expression) {
 adamic_value *raw = object == NULL ? NULL : adamic_object_optional_field(object, field, cache);
 // Force the common read to report presence and readiness before inspecting any
 // payload. Static inherited slots need the owner's normalized probe.
 if (raw == NULL || !adamic_object_initialized(object)[cache->index]) {
  (void)adamic_object_view(object, field, cache, 4, declared, expression);
 }
 unsigned char actual = adamic_object_field_types(object)[cache->index];
 if (object->class != NULL && object->class->is_static) actual = 0;
 unsigned char type = actual;
 if (actual == 10 && raw->reference != NULL) {
  const adamic_heap *heap = raw->reference;
  type = heap->kind == adamic_kind_string ? 3 : heap->kind == adamic_kind_object ? 4 : heap->kind == adamic_kind_number ? 1 : heap->kind == adamic_kind_boolean ? 2 : heap->kind == adamic_kind_array ? 5 : heap->kind == adamic_kind_map ? 6 : heap->kind == adamic_kind_closure ? 8 : 0;
 }
 bool absent = actual == 13 || (actual == 10 && raw->reference == NULL) || (actual >= 3 && actual <= 6 && raw->reference == NULL);
 if (absent && undefined) return NULL;
 if (!absent) for (size_t index = 0; index < count; index++) {
  const adamic_object_primitive_member *member = &members[index];
  if (member->type != type) continue;
  adamic_value value = adamic_object_view(object, field, cache, type, declared, expression);
  if (member->literal && !(type == 1 ? value.number == member->value.number : type == 2 ? value.boolean == member->value.boolean : adamic_string_equal(value.reference, member->value.reference))) continue;
  if (type == 1) return adamic_box_number(value.number);
  if (type == 2) return value.boolean ? (adamic_heap *)&adamic_box_true : (adamic_heap *)&adamic_box_false;
  return adamic_retain(value.reference);
 }
 const char *found = absent ? "undefined" : type == 1 ? "number" : type == 2 ? "boolean" : type == 3 ? "string" : type == 4 ? "object" : type == 5 ? "array" : type == 6 ? "Map" : type == 8 ? "function" : type == 12 ? "null" : "unsupported representation";
 size_t capacity = strlen(expression) + 2 * strlen(declared) + strlen(found) + 100;
 char *message = malloc(capacity);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int length = snprintf(message, capacity, "field read failed: %s matches no member of %s; expected %s, found %s", expression, declared, declared, found);
 adamic_panic(message, (size_t)length);
}
