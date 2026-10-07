// object.c: plain objects, each carrying its shape.

#include "adamic.h"

#include <string.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(sizeof *object + shape->count * sizeof object->slots[0], adamic_kind_object);
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	return object;
}

adamic_object *adamic_object_copy(const adamic_object *source) {
	adamic_object *object = adamic_object_new(source->shape);
	for (size_t index = 0; index < source->shape->count; index++) {
		object->slots[index] = source->slots[index];
		if (source->shape->references[index]) {
			adamic_retain(object->slots[index].reference);
		}
	}
	return object;
}

// adamic_object_has is object.hasOwnProperty(name): one of the shape's own names, not a method on a
// prototype. A shape's names are C strings, so the lengths have to agree before the bytes do.
bool adamic_object_has(const adamic_object *object, const adamic_string *name) {
	for (size_t index = 0; index < object->shape->count; index++) {
		const char *field = object->shape->names[index];
		size_t length = strlen(field);
		if (length == name->length && memcmp(field, name->bytes, length) == 0) {
			return true;
		}
	}
	return false;
}

void adamic_slot_cache_store(adamic_slot_cache *cache, const adamic_shape *shape, size_t index) {
	uintptr_t pointer = (uintptr_t)shape;
	if ((pointer & ~ADAMIC_SLOT_SHAPE_MASK) != 0) {
		adamic_panic("shape address exceeds 48 bits", sizeof "shape address exceeds 48 bits" - 1);
	}
	uint64_t packed = index <= UINT16_MAX ? pointer | ((uint64_t)index << 48) : 0;
	__atomic_store_n(&cache->packed, packed, __ATOMIC_RELAXED);
	ADAMIC_TSAN_PAUSE(adamic_tsan_cache_publication);
}

adamic_value *adamic_object_find(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_object *mutable = (adamic_object *)object;
	for (size_t index = 0; index < object->shape->count; index++) {
		if (strcmp(object->shape->names[index], name) == 0) {
			adamic_slot_cache_store(cache, object->shape, index);
			return &mutable->slots[index];
		}
	}
	static const char message[] = "compiler bug: a field the checker proved is there is missing";
	adamic_panic(message, sizeof message - 1);
}

adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method) {
	const adamic_shape *shape = object->shape;
	// The cache's index counts the fields, then the methods after them.
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	size_t slot = packed >> 48;
	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)shape) {
		bool found = false;
		for (size_t index = 0; index < shape->count && !found; index++) {
			if (strcmp(shape->names[index], name) == 0) {
				slot = index;
				found = true;
			}
		}
		for (size_t index = 0; shape->methods != NULL && index < shape->methods->count && !found; index++) {
			if (strcmp(shape->methods->names[index], name) == 0) {
				slot = shape->count + index;
				found = true;
			}
		}
		if (!found) {
			static const char message[] = "compiler bug: a method the checker proved is there is missing";
			adamic_panic(message, sizeof message - 1);
		}
		adamic_slot_cache_store(cache, shape, slot);
	}
	if (slot < shape->count) {
		return object->slots[slot].reference;
	}
	*method = shape->methods->code[slot - shape->count];
	return NULL;
}
