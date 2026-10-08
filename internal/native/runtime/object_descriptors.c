// Copyright 2015 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license in THIRD_PARTY_NOTICES.md.
// Port of the data-property part of ValidateAndApplyPropertyDescriptor in
// V8 13.6.233 src/objects/js-objects.cc. Lowering excludes accessors, proxies,
// coercing keys and writes whose value would invalidate a statically typed slot.
#include "adamic.h"

#include <stdlib.h>
#include <string.h>

// The table does not own its receiver. Heap and region destruction both remove
// its entries through the object's child-release hook. Only scalar values are
// stored outside the shape, so this cannot introduce a reference cycle.
typedef struct descriptor {
 adamic_object *object;
 adamic_string *key;
 size_t slot;
 bool writable, enumerable, configurable;
 adamic_heap *value;
 struct descriptor *next;
} descriptor;
static descriptor *descriptors;

static descriptor *find(const adamic_object *object, const adamic_string *key) {
 for (descriptor *entry = descriptors; entry != NULL; entry = entry->next) {
  if (entry->object == object && adamic_string_equal(entry->key, key)) return entry;
 }
 return NULL;
}
static size_t slot_named(const adamic_object *object, const adamic_string *key) {
 for (size_t index = 0; index < object->shape->count; index++) {
  const char *name = object->shape->names[index];
  if (name[0] != '#' && strlen(name) == key->length && memcmp(name, key->bytes, key->length) == 0) return index;
 }
 return SIZE_MAX;
}
static bool truthy(const adamic_heap *value) {
 if (value == NULL) return false;
 switch (value->kind) {
 case adamic_kind_number: {
  double number = ((const adamic_number_box *)value)->number;
  return number != 0 && !isnan(number);
 }
 case adamic_kind_boolean: return ((const adamic_boolean_box *)value)->boolean;
 case adamic_kind_string: return ((const adamic_string *)value)->length != 0;
 default: return true;
 }
}
static adamic_string *failure(const adamic_string *key, bool extensible) {
 const char *prefix = extensible ? "Cannot redefine property: " : "Cannot define property ";
 const char *suffix = extensible ? "" : ", object is not extensible";
 size_t start = strlen(prefix), end = strlen(suffix);
 adamic_string *message = adamic_string_allocate(start + key->length + end);
 memcpy((char *)message->bytes, prefix, start);
 memcpy((char *)message->bytes + start, key->bytes, key->length);
 memcpy((char *)message->bytes + start + key->length, suffix, end);
 return message;
}
static bool same_slot(const adamic_object *object, size_t slot, const adamic_heap *value, int representation) {
 if (representation == 1) {
  if (value == NULL || value->kind != adamic_kind_number) return false;
  double left = object->slots[slot].number, right = ((const adamic_number_box *)value)->number;
  if (isnan(left) && isnan(right)) return true;
  return left == right && (left != 0 || signbit(left) == signbit(right));
 }
 if (representation == 2) return value != NULL && value->kind == adamic_kind_boolean && object->slots[slot].boolean == ((const adamic_boolean_box *)value)->boolean;
 return adamic_object_is(object->slots[slot].reference, value);
}
adamic_string *adamic_object_define_data(adamic_object *object, const adamic_string *key,
 const adamic_heap *value, const adamic_heap *writable_value, const adamic_heap *enumerable_value,
 const adamic_heap *configurable_value, unsigned mask, int representation) {
 descriptor *entry = find(object, key);
 size_t slot = entry != NULL ? entry->slot : slot_named(object, key);
 bool exists = entry != NULL || slot != SIZE_MAX;
 bool writable = truthy(writable_value), enumerable = truthy(enumerable_value), configurable = truthy(configurable_value);
 bool old_writable = exists && (entry == NULL || entry->writable) && !object->frozen;
 bool old_enumerable = exists && (entry == NULL || entry->enumerable);
 bool old_configurable = exists && (entry == NULL || entry->configurable) && !object->sealed && !object->frozen;
 if (!exists && !adamic_object_is_extensible((const adamic_heap *)object)) return failure(key, false);
 if (exists && !old_configurable) {
  if (((mask & 8) && configurable) || ((mask & 4) && enumerable != old_enumerable)) return failure(key, true);
  if (!old_writable) {
   if ((mask & 2) && writable) return failure(key, true);
   if ((mask & 1) && !(slot != SIZE_MAX ? same_slot(object, slot, value, representation) : adamic_object_is(entry->value, value))) return failure(key, true);
  }
 }
 if (exists && mask == 0) return NULL;
 if (entry == NULL) {
  entry = malloc(sizeof *entry);
  if (entry == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
  *entry = (descriptor){object, adamic_retain((void *)key), slot, old_writable, old_enumerable, old_configurable, NULL, NULL};
  descriptor **tail = &descriptors;
  while (*tail != NULL) tail = &(*tail)->next;
  *tail = entry;
 }
 if (mask & 2) entry->writable = writable;
 if (mask & 4) entry->enumerable = enumerable;
 if (mask & 8) entry->configurable = configurable;
 if (mask & 1) {
  if (slot == SIZE_MAX) {
   adamic_retain((void *)value);
   adamic_release(entry->value);
   entry->value = (adamic_heap *)value;
  } else if (representation == 1) object->slots[slot].number = ((const adamic_number_box *)value)->number;
  else if (representation == 2) object->slots[slot].boolean = ((const adamic_boolean_box *)value)->boolean;
  else {
   adamic_retain((void *)value);
   adamic_release(object->slots[slot].reference);
   object->slots[slot].reference = (void *)value;
  }
 }
 return NULL;
}
bool adamic_object_descriptor_has(const adamic_object *object, const adamic_string *key) { return find(object, key) != NULL; }
bool adamic_object_enumerable(const adamic_object *object, const adamic_string *key) {
 descriptor *entry = find(object, key);
 if (entry != NULL) return entry->enumerable;
 return slot_named(object, key) != SIZE_MAX;
}
bool adamic_object_descriptor_writable(const adamic_object *object, const char *key) {
 for (descriptor *entry = descriptors; entry != NULL; entry = entry->next) {
  if (entry->object == object && strlen(key) == entry->key->length && memcmp(key, entry->key->bytes, entry->key->length) == 0) return entry->writable;
 }
 return true;
}
bool adamic_object_descriptor_integrity(const adamic_object *object, bool frozen) {
 if (adamic_object_is_extensible((const adamic_heap *)object)) return false;
 for (size_t slot = 0; slot < object->shape->count; slot++) {
  const char *name = object->shape->names[slot];
  if (name[0] == '#') continue;
  descriptor *found = NULL;
  for (descriptor *entry = descriptors; entry != NULL; entry = entry->next) if (entry->object == object && entry->slot == slot) { found = entry; break; }
  if (!object->sealed && !object->frozen && (found == NULL || found->configurable)) return false;
  if (frozen && !object->frozen && (found == NULL || found->writable)) return false;
 }
 for (descriptor *entry = descriptors; entry != NULL; entry = entry->next) {
  if (entry->object != object || entry->slot != SIZE_MAX) continue;
  if (!object->sealed && !object->frozen && entry->configurable) return false;
  if (frozen && !object->frozen && entry->writable) return false;
 }
 return true;
}
static adamic_string *name_string(const char *name) {
 size_t length = strlen(name);
 adamic_string *key = adamic_string_allocate(length);
 memcpy((char *)key->bytes, name, length);
 return key;
}
static bool index_key(const adamic_string *key, uint32_t *index) {
 if (key->length == 0 || (key->length > 1 && key->bytes[0] == '0')) return false;
 uint64_t value = 0;
 for (size_t at = 0; at < key->length; at++) {
  unsigned char byte = (unsigned char)key->bytes[at];
  if (byte < '0' || byte > '9') return false;
  value = value * 10 + byte - '0';
  if (value >= UINT32_MAX) return false;
 }
 *index = (uint32_t)value;
 return true;
}
static int order(const adamic_string *left, const adamic_string *right) {
 uint32_t a = 0, b = 0;
 bool ai = index_key(left, &a), bi = index_key(right, &b);
 if (ai != bi) return ai ? -1 : 1;
 if (!ai) return 0;
 return a < b ? -1 : a > b ? 1 : 0;
}
adamic_array *adamic_object_descriptor_names(const adamic_object *object, bool all) {
 bool altered = false;
 for (descriptor *entry = descriptors; entry != NULL; entry = entry->next) if (entry->object == object) { altered = true; break; }
 if (!altered) return NULL;
 adamic_array *keys = adamic_array_new(object->shape->count, true);
 for (size_t slot = 0; slot < object->shape->count; slot++) {
  if (object->shape->names[slot][0] == '#') continue;
  adamic_string *key = name_string(object->shape->names[slot]);
  if (all || adamic_object_enumerable(object, key)) adamic_array_push(keys, (adamic_value){.reference = key});
  else adamic_release(key);
 }
 for (descriptor *entry = descriptors; entry != NULL; entry = entry->next) {
  if (entry->object == object && entry->slot == SIZE_MAX && (all || entry->enumerable)) adamic_array_push(keys, (adamic_value){.reference = adamic_retain(entry->key)});
 }
 for (size_t at = 1; at < keys->length; at++) {
  adamic_value value = keys->elements[at];
  size_t place = at;
  while (place > 0 && order(value.reference, keys->elements[place - 1].reference) < 0) { keys->elements[place] = keys->elements[place - 1]; place--; }
  keys->elements[place] = value;
 }
 return keys;
}
void adamic_object_descriptors_free(adamic_object *object, void (*release)(void *)) {
 descriptor **at = &descriptors;
 while (*at != NULL) {
  descriptor *entry = *at;
  if (entry->object != object) { at = &entry->next; continue; }
  *at = entry->next;
  if (entry->key->heap.references != 0) release(entry->key);
  if (entry->value != NULL && entry->value->references != 0) release(entry->value);
  free(entry);
 }
}
adamic_object *adamic_object_type_error(adamic_string *message) {
 return adamic_builtin_error_new(1, message);
}
