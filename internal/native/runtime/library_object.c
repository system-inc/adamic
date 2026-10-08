// library_object.c: static Object methods over proven, fixed shapes.
#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

bool adamic_object_is(const adamic_heap *left, const adamic_heap *right) {
 if (left == NULL || right == NULL) return left == right;
 if (left->kind != right->kind) return false;
 if (left->kind == adamic_kind_number) {
  double a = ((const adamic_number_box *)left)->number;
  double b = ((const adamic_number_box *)right)->number;
  if (isnan(a) && isnan(b)) return true;
  return a == b && (a != 0 || signbit(a) == signbit(b));
 }
 return adamic_union_equal(left, right);
}

bool adamic_object_is_frozen(const adamic_heap *value) {
 if (value == NULL) return true;
 if (value->kind == adamic_kind_object) return ((const adamic_object *)value)->frozen;
 return value->kind == adamic_kind_string || value->kind == adamic_kind_number || value->kind == adamic_kind_boolean;
}

adamic_object *adamic_object_freeze(adamic_object *object) {
 object->frozen = true;
 return adamic_retain(object);
}

void adamic_object_check_write(const adamic_object *object, const char *name) {
 // Private fields are internal slots, unaffected by Object.freeze in JavaScript.
 if (!object->frozen || name[0] == '#') return;
 const char prefix[] = "TypeError: Cannot assign to read only property '";
 const char suffix[] = "' of object '#<Object>'";
 size_t length = sizeof prefix - 1 + strlen(name) + sizeof suffix - 1;
 char *message = malloc(length + 1);
 if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
 (void)snprintf(message, length + 1, "%s%s%s", prefix, name, suffix);
 adamic_panic(message, length);
}

bool adamic_object_has_own(const adamic_object *object, const adamic_string *key) {
 if (object->dictionary != NULL) return adamic_record_has_own(object, key);
 for (size_t index = 0; index < object->shape->count; index++) {
  const char *name = object->shape->names[index];
  if (name[0] != '#' && strlen(name) == key->length && memcmp(name, key->bytes, key->length) == 0) return true;
 }
 return false;
}

// Array indices are canonical decimal strings from 0 to 2^32 - 2, not general integers.
// In particular 01, -0, 4294967295 and 1e0 retain insertion order.
static bool array_index(const char *name, uint32_t *value) {
 if (*name == '\0' || (name[0] == '0' && name[1] != '\0')) return false;
 uint64_t number = 0;
 for (const char *at = name; *at != '\0'; at++) {
  if (*at < '0' || *at > '9') return false;
  number = number * 10 + (unsigned)(*at - '0');
  if (number >= UINT32_MAX) return false;
 }
 *value = (uint32_t)number;
 return true;
}

static int compare_names(const char *left, const char *right) {
 uint32_t a = 0, b = 0;
 bool ai = array_index(left, &a), bi = array_index(right, &b);
 if (ai != bi) return ai ? -1 : 1;
 if (!ai) return 0;
 return a < b ? -1 : a > b ? 1 : 0;
}

static size_t *ordered(const adamic_object *object) {
 size_t count = object->shape->count;
 size_t *indices = malloc((count == 0 ? 1 : count) * sizeof *indices);
 if (indices == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
 // Stable insertion sort preserves order of ordinary strings.
 for (size_t index = 0; index < count; index++) {
  size_t place = index;
  while (place > 0 && compare_names(object->shape->names[index], object->shape->names[indices[place - 1]]) < 0) {
   indices[place] = indices[place - 1];
   place--;
  }
  indices[place] = index;
 }
 return indices;
}

static adamic_string *key_string(const char *name) {
 size_t length = strlen(name);
 adamic_string *key = adamic_string_allocate(length);
 memcpy((char *)key->bytes, name, length);
 return key;
}

adamic_array *adamic_object_keys(const adamic_object *object) {
 if (object->dictionary != NULL) return adamic_record_keys(object);
 // Class descriptors hide private storage and track static own-property presence.
 if (object->class != NULL) return adamic_class_object_keys(object);
 size_t *indices = ordered(object);
 adamic_array *keys = adamic_array_new(object->shape->count, true);
 for (size_t at = 0; at < object->shape->count; at++) {
  const char *name = object->shape->names[indices[at]];
  if (name[0] == '#') continue;
  adamic_array_push(keys, (adamic_value){.reference = key_string(name)});
 }
 free(indices);
 return keys;
}

adamic_array *adamic_object_values(const adamic_object *object, bool references, bool entries) {
 if (object->dictionary != NULL) return adamic_record_values(object, entries);
 size_t *indices = ordered(object);
 adamic_array *values = adamic_array_new(object->shape->count, entries || references);
 static const char *const names[] = {"0", "1"};
 static const bool number_references[] = {true, false};
 static const bool string_references[] = {true, true};
 static const adamic_shape number_pair = {2, names, number_references, NULL};
 static const adamic_shape string_pair = {2, names, string_references, NULL};
 for (size_t at = 0; at < object->shape->count; at++) {
  size_t index = indices[at];
  const char *name = object->shape->names[index];
  if (name[0] == '#') continue;
  adamic_value value = object->slots[index];
  if (references) adamic_retain(value.reference);
  if (entries) {
   adamic_object *pair = adamic_object_new(references ? &string_pair : &number_pair);
   pair->slots[0].reference = key_string(name);
   pair->slots[1] = value;
   value.reference = pair;
  }
  adamic_array_push(values, value);
 }
 free(indices);
 return values;
}

void adamic_object_assign(adamic_object *target, const adamic_object *source) {
 size_t *indices = ordered(source);
 for (size_t at = 0; at < source->shape->count; at++) {
  size_t index = indices[at];
  const char *name = source->shape->names[index];
  adamic_object_check_write(target, name);
  adamic_slot_cache cache = {NULL, 0};
  adamic_value *slot = adamic_object_field(target, name, &cache);
  adamic_value value = source->slots[index];
  if (source->shape->references[index]) {
   adamic_retain(value.reference);
   adamic_release(slot->reference);
  }
  *slot = value;
 }
 free(indices);
}
