// object.c: plain objects, each carrying its shape.

#include "adamic.h"

#include <string.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(sizeof *object + shape->count * sizeof object->slots[0], adamic_kind_object);
	object->shape = shape;
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

adamic_value *adamic_object_find(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_object *mutable = (adamic_object *)object;
	for (size_t index = 0; index < object->shape->count; index++) {
		if (strcmp(object->shape->names[index], name) == 0) {
			cache->shape = object->shape;
			cache->index = index;
			return &mutable->slots[index];
		}
	}
	static const char message[] = "compiler bug: a field the checker proved is there is missing";
	adamic_panic(message, sizeof message - 1);
}

adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method) {
	const adamic_shape *shape = object->shape;
	// The cache's index counts the fields, then the methods after them.
	if (cache->shape != shape) {
		bool found = false;
		for (size_t index = 0; index < shape->count && !found; index++) {
			if (strcmp(shape->names[index], name) == 0) {
				cache->index = index;
				found = true;
			}
		}
		for (size_t index = 0; shape->methods != NULL && index < shape->methods->count && !found; index++) {
			if (strcmp(shape->methods->names[index], name) == 0) {
				cache->index = shape->count + index;
				found = true;
			}
		}
		if (!found) {
			static const char message[] = "compiler bug: a method the checker proved is there is missing";
			adamic_panic(message, sizeof message - 1);
		}
		cache->shape = shape;
	}
	if (cache->index < shape->count) {
		return object->slots[cache->index].reference;
	}
	*method = shape->methods->code[cache->index - shape->count];
	return NULL;
}
