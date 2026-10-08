// for_in.c: runtime key snapshots, with presence separate from undefined values.
#include "adamic.h"
#include <string.h>

// Only compiler-created undefined-spread shapes carry this tag. Their final slot
// owns an ordered list of real keys; preceding slots may include absent storage.
const adamic_methods adamic_for_in_metadata = {0, NULL, NULL};

static bool tracked(const adamic_object *object) {
 return object->class == NULL && object->shape->methods == &adamic_for_in_metadata;
}
static adamic_array *presence(const adamic_object *object) {
 return object->slots[object->shape->count - 1].reference;
}
static adamic_string *name_string(const char *name) {
 size_t length = strlen(name);
 adamic_string *key = adamic_string_allocate(length);
 memcpy((char *)key->bytes, name, length);
 return key;
}
static bool contains(const adamic_array *keys, const adamic_string *key) {
 for (size_t i = 0; i < keys->length; i++) {
  if (adamic_string_equal(keys->elements[i].reference, key)) return true;
 }
 return false;
}
static bool index_key(const adamic_string *key, uint32_t *result) {
 if (key->length == 0 || key->length > 10 || (key->length > 1 && key->bytes[0] == '0')) return false;
 uint64_t number = 0;
 for (size_t i = 0; i < key->length; i++) {
  unsigned char c = (unsigned char)key->bytes[i];
  if (c < '0' || c > '9') return false;
  number = number * 10 + c - '0';
 }
 if (number >= UINT32_MAX) return false;
 *result = (uint32_t)number;
 return true;
}
static void order_keys(adamic_array *keys) {
 // Stable insertion sort puts canonical indices first without disturbing strings.
 for (size_t i = 1; i < keys->length; i++) {
  adamic_value value = keys->elements[i];
  uint32_t index;
  if (!index_key(value.reference, &index)) continue;
  size_t place = i;
  while (place > 0) {
   uint32_t previous;
   if (index_key(keys->elements[place - 1].reference, &previous) && previous <= index) break;
   keys->elements[place] = keys->elements[place - 1];
   place--;
  }
  keys->elements[place] = value;
 }
}
static adamic_array *copy_keys(const adamic_array *source) {
 adamic_array *keys = adamic_array_new(source->length, true);
 for (size_t i = 0; i < source->length; i++) {
  adamic_array_push(keys, (adamic_value){.reference = adamic_retain(source->elements[i].reference)});
 }
 return keys;
}
void adamic_for_in_initialize(adamic_object *object, size_t absent) {
 adamic_array *keys = adamic_array_new(object->shape->count - 1 - absent, true);
 for (size_t i = absent; i + 1 < object->shape->count; i++) {
  adamic_array_push(keys, (adamic_value){.reference = name_string(object->shape->names[i])});
 }
 object->slots[object->shape->count - 1].reference = keys;
}
void adamic_for_in_write(adamic_object *object, const char *name) {
 if (!tracked(object)) return;
 adamic_string *key = name_string(name);
 adamic_array *keys = presence(object);
 if (contains(keys, key)) { adamic_release(key); return; }
 adamic_array_push(keys, (adamic_value){.reference = key});
}
adamic_object *adamic_for_in_copy(const adamic_object *source) {
 if (!tracked(source)) return adamic_object_copy(source);
 // Copy slots by index: the hidden final slot is not a JavaScript property,
 // and its storage name must not collide with a real user's property lookup.
 adamic_object *object = adamic_object_new(source->shape);
 for (size_t i = 0; i + 1 < source->shape->count; i++) {
  object->slots[i] = source->slots[i];
  if (source->shape->references[i]) adamic_retain(object->slots[i].reference);
 }
 object->slots[source->shape->count - 1].reference = copy_keys(presence(source));
 return object;
}
// RegExp arrays use a fixed storage shape for both ordinary match properties and
// indices.groups. Storage for a disabled indices result is not an own property.
static bool array_property_present(const adamic_object *object, size_t slot) {
 const adamic_shape *shape = object->shape;
 if (shape->count == 4 && strcmp(shape->names[0], "index") == 0 && strcmp(shape->names[1], "input") == 0 && strcmp(shape->names[2], "groups") == 0 && strcmp(shape->names[3], "indices") == 0) {
  if (object->slots[1].reference == NULL) return slot == 2;
  if (slot == 3) return object->slots[3].reference != NULL;
 }
 return true;
}
static bool array_has_property(const adamic_object *object, const adamic_string *key) {
 if (object == NULL) return false;
 for (size_t i = 0; i < object->shape->count; i++) {
  const char *name = object->shape->names[i];
  if (strlen(name) == key->length && memcmp(name, key->bytes, key->length) == 0) return array_property_present(object, i);
 }
 return false;
}
bool adamic_for_in_own(const adamic_heap *value, const adamic_string *key) {
 if (value == NULL) return false;
 uint32_t index;
 switch (value->kind) {
 case adamic_kind_object: {
  const adamic_object *object = (const adamic_object *)value;
  if (tracked(object)) return contains(presence(object), key);
  if (object->class != NULL && object->class->is_static) {
   for (size_t i = 0; i < object->shape->count; i++) {
    const char *name = object->shape->names[i];
    if (strlen(name) != key->length || memcmp(name, key->bytes, key->length) != 0) continue;
    size_t flag = object->class->static_flags[i];
    if (flag == 0 || object->slots[flag - 1].number != 0) return true;
    break;
   }
   if (object->class->static_parent != 0) return adamic_for_in_own(object->slots[object->class->static_parent - 1].reference, key);
  }
  return adamic_object_has(object, key);
 }
 case adamic_kind_array: {
  const adamic_array *array = (const adamic_array *)value;
  if (index_key(key, &index) && index < array->length) return true;
  return array_has_property(array->properties, key);
 }
 case adamic_kind_string:
  return index_key(key, &index) && index < adamic_string_length((const adamic_string *)value);
 case adamic_kind_typed_array:
  return index_key(key, &index) && index < ((const adamic_typed_array *)value)->length;
 default: return false;
 }
}
static bool static_shadow(const adamic_object *object, const adamic_string *key) {
 for (size_t i = 0; i < object->shape->count; i++) {
  const char *name = object->shape->names[i];
  if (strlen(name) != key->length || memcmp(name, key->bytes, key->length) != 0) continue;
  size_t flag = object->class->static_flags[i];
  return flag == 0 || object->slots[flag - 1].number != 0;
 }
 return false;
}
adamic_array *adamic_for_in_keys(const adamic_heap *value) {
 if (value != NULL && value->kind == adamic_kind_object) {
  const adamic_object *object = (const adamic_object *)value;
  if (!tracked(object)) {
   adamic_array *keys = adamic_class_object_keys(object);
   if (object->class != NULL && object->class->is_static && object->class->static_parent != 0) {
    adamic_array *inherited = adamic_for_in_keys(object->slots[object->class->static_parent - 1].reference);
    for (size_t i = 0; i < inherited->length; i++) {
     const adamic_string *key = inherited->elements[i].reference;
     if (!static_shadow(object, key) && !contains(keys, key)) adamic_array_push(keys, (adamic_value){.reference = adamic_retain((void *)key)});
    }
    adamic_release(inherited);
   }
   return keys;
  }
  adamic_array *keys = copy_keys(presence(object));
  order_keys(keys);
  return keys;
 }
 size_t length = 0;
 const adamic_object *properties = NULL;
 if (value != NULL) {
  if (value->kind == adamic_kind_array) {
   const adamic_array *array = (const adamic_array *)value;
   length = array->length;
   properties = array->properties;
  } else if (value->kind == adamic_kind_string) {
   length = (size_t)adamic_string_length((const adamic_string *)value);
  } else if (value->kind == adamic_kind_typed_array) {
   length = ((const adamic_typed_array *)value)->length;
  }
 }
 adamic_array *keys = adamic_array_new(length, true);
 for (size_t i = 0; i < length; i++) {
  adamic_array_push(keys, (adamic_value){.reference = adamic_string_from_number((double)i)});
 }
 if (properties != NULL) {
  for (size_t i = 0; i < properties->shape->count; i++) {
   if (!array_property_present(properties, i)) continue;
   adamic_string *key = name_string(properties->shape->names[i]);
   if (!contains(keys, key)) adamic_array_push(keys, (adamic_value){.reference = key});
   else adamic_release(key);
  }
  order_keys(keys);
 }
 return keys;
}

// Preserve Object.keys' own-only semantics while hiding enumeration metadata.
adamic_array *adamic_for_in_object_keys(const adamic_object *object) {
 if (!tracked(object)) return adamic_class_object_keys(object);
 adamic_array *keys = copy_keys(presence(object));
 order_keys(keys);
 return keys;
}
