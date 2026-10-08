#include "view_unions_object_primitive.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_heap *adamic_object_primitive_view(const adamic_object *object, const char *field, adamic_slot_cache *cache, const adamic_object_primitive_member *members, size_t count, bool undefined, bool allow_absent, bool optional, const char *declared, const char *expression) {
 if (object == NULL && optional) return NULL;
 adamic_value *raw = object == NULL ? NULL : adamic_object_optional_field(object, field, cache);
 if (object != NULL && raw == NULL && allow_absent && undefined) return NULL;
 // Force the common read to report presence and readiness before inspecting any
 // payload. Static inherited slots need the owner's normalized probe.
 if (raw == NULL || !adamic_object_initialized(object)[cache->index]) {
  (void)adamic_object_view(object, field, cache, 4, declared, expression);
 }
 unsigned char actual = adamic_object_field_types(object)[cache->index];
 if (object->class != NULL && object->class->is_static) actual = 0;
 unsigned char type = actual;
 if (actual == adamic_rep_union && raw->reference != NULL) {
  const adamic_heap *heap = raw->reference;
  type = heap->kind == adamic_kind_string ? adamic_rep_string : heap->kind == adamic_kind_object ? adamic_rep_object : heap->kind == adamic_kind_number ? adamic_rep_number : heap->kind == adamic_kind_boolean ? adamic_rep_boolean : heap->kind == adamic_kind_array ? adamic_rep_array : heap->kind == adamic_kind_map ? adamic_rep_map : heap->kind == adamic_kind_closure ? adamic_rep_closure : 0;
 }
 bool absent = actual == adamic_rep_undefined || (actual == adamic_rep_union && raw->reference == NULL) || (actual >= adamic_rep_string && actual <= adamic_rep_map && raw->reference == NULL);
 if (absent && undefined) return NULL;
 if (!absent) for (size_t index = 0; index < count; index++) {
  const adamic_object_primitive_member *member = &members[index];
  if (member->type != type) continue;
  adamic_value value = adamic_object_view(object, field, cache, type, declared, expression);
  if (member->literal && !(type == adamic_rep_number ? value.number == member->value.number : type == adamic_rep_boolean ? value.boolean == member->value.boolean : adamic_string_equal(value.reference, member->value.reference))) continue;
  if (type == adamic_rep_number) return adamic_box_number(value.number);
  if (type == adamic_rep_boolean) return value.boolean ? (adamic_heap *)&adamic_box_true : (adamic_heap *)&adamic_box_false;
  return adamic_retain(value.reference);
 }
 const char *found = absent ? "undefined" : type == adamic_rep_number ? "number" : type == adamic_rep_boolean ? "boolean" : type == adamic_rep_string ? "string" : type == adamic_rep_object ? "object" : type == adamic_rep_array ? "array" : type == adamic_rep_map ? "Map" : type == adamic_rep_closure ? "function" : type == adamic_rep_null ? "null" : "unsupported representation";
 size_t capacity = strlen(expression) + 2 * strlen(declared) + strlen(found) + 100;
 char *message = malloc(capacity);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int length = snprintf(message, capacity, "field read failed: %s matches no member of %s; expected %s, found %s", expression, declared, declared, found);
 adamic_panic(message, (size_t)length);
}
