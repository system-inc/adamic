// construction.c: fixed storage whose own properties appear only on actual writes.
#include "adamic.h"

#include <stdlib.h>
#include <string.h>

static size_t construction_size(const adamic_shape *shape) {
 size_t per_field = sizeof(adamic_value) + 2 + sizeof(size_t);
 if (shape->count > (SIZE_MAX - sizeof(adamic_object) - _Alignof(size_t)) / per_field) {
  adamic_panic("out of memory", sizeof "out of memory" - 1);
 }
 size_t before_order = sizeof(adamic_object) + shape->count * (sizeof(adamic_value) + 2);
 size_t aligned = (before_order + _Alignof(size_t) - 1) & ~(size_t)(_Alignof(size_t) - 1);
 return aligned + shape->count * sizeof(size_t);
}

static void construction_start(adamic_object *object, const adamic_shape *shape) {
 object->shape = shape;
 object->class = NULL;
 object->frozen = false;
 memset(object->slots, 0, shape->count * sizeof(adamic_value));
 memset(adamic_object_initialized(object), 0, shape->count);
 memset(adamic_object_field_types(object), 0, shape->count);
 size_t order_offset = construction_size(shape) - shape->count * sizeof(size_t);
 object->write_order = (size_t *)(void *)((unsigned char *)object + order_offset);
 for (size_t index = 0; index < shape->count; index++) object->write_order[index] = SIZE_MAX;
}

bool adamic_object_present(const adamic_object *object, size_t index) {
 return object->write_order == NULL || object->write_order[index] != SIZE_MAX;
}

void adamic_object_publish(adamic_object *object, size_t index) {
 if (object->write_order == NULL || object->write_order[index] != SIZE_MAX) return;
 size_t order = 0;
 for (size_t previous = 0; previous < object->shape->count; previous++) {
  if (object->write_order[previous] != SIZE_MAX && object->write_order[previous] >= order) order = object->write_order[previous] + 1;
 }
 object->write_order[index] = order;
}

adamic_object *adamic_object_construct(const adamic_shape *shape) {
 adamic_object *object = adamic_allocate(construction_size(shape), adamic_kind_object);
 construction_start(object, shape);
 return object;
}

adamic_array *adamic_node_array_new(size_t capacity, bool references, const adamic_shape *extras) {
 size_t metadata_size = construction_size(extras);
 if (metadata_size > SIZE_MAX - sizeof(adamic_array) || capacity > SIZE_MAX / sizeof(adamic_value)) {
  adamic_panic("out of memory", sizeof "out of memory" - 1);
 }
 adamic_array *array = adamic_allocate(sizeof *array + metadata_size, adamic_kind_array);
 array->length = 0;
 array->capacity = capacity;
 array->references = references;
 array->properties = NULL;
 array->elements = capacity == 0 ? NULL : malloc(capacity * sizeof(adamic_value));
 if (capacity != 0 && array->elements == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
 array->metadata = (adamic_object *)(void *)(array + 1);
 // Interior storage is not independently retained or freed. The array releases its children.
 array->metadata->heap = (adamic_heap){0, adamic_kind_object, 0};
 construction_start(array->metadata, extras);
 return array;
}

adamic_array *adamic_node_array_keys(const adamic_array *array) {
 adamic_array *keys = adamic_array_new(array->length, true);
 for (size_t index = 0; index < array->length; index++) {
  adamic_array_push(keys, (adamic_value){.reference = adamic_string_from_number((double)index)});
 }
 if (array->metadata != NULL) {
  adamic_array *extras = adamic_object_keys(array->metadata);
  for (size_t index = 0; index < extras->length; index++) adamic_array_push(keys, (adamic_value){.reference = adamic_retain(extras->elements[index].reference)});
  adamic_release(extras);
 }
 return keys;
}
