// object.c: plain objects, each carrying its shape.

#include "adamic.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(sizeof *object + shape->count * (sizeof object->slots[0] + sizeof(size_t) + 1), adamic_kind_object);
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->dynamic_shape = false;
	object->dynamic_types = NULL;
	for (size_t index = 0; index < shape->count; index++) { adamic_object_orders(object)[index] = index + 1; }
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	memset(adamic_object_initialized(object), 1, shape->count);
	return object;
}

// Both fixed and reserved descriptors copy through the same source-index path.
static void copy_fields(const adamic_object *source, adamic_object *object, const adamic_shape *shape, const char *expression) {
	for (size_t position = 0; position < shape->count; position++) {
		size_t index = adamic_public_index(shape, position);
		adamic_slot_cache cache = {NULL, 0};
		const adamic_accessor *accessor = adamic_accessor_find(source, shape->names[index]);
		if (accessor == NULL) {
			adamic_value *slot = adamic_object_field(source, shape->names[index], &cache);
			if (expression != NULL && adamic_object_orders(source)[cache.index] != 0) slot = adamic_object_read(source, shape->names[index], &cache, expression);
			object->slots[index] = *slot;
			adamic_object_initialized(object)[index] = adamic_object_initialized(source)[cache.index];
			if (source->class == NULL) adamic_object_orders(object)[index] = adamic_object_orders(source)[cache.index];
		} else {
			object->slots[index] = adamic_accessor_get((adamic_object *)source, shape->names[index]);
			adamic_object_orders(object)[index] = index + 1;
		}
		if (shape->references[index] && accessor == NULL) {
			adamic_retain(object->slots[index].reference);
		}
	}
}

adamic_object *adamic_object_copy_checked(const adamic_object *source, const char *expression) {
	if (source->dynamic_shape) { return adamic_object_copy_reserving_checked(source, NULL, expression); }
	const adamic_shape *shape = source->class == NULL ? source->shape : source->class->public_shape;
	adamic_object *object = adamic_object_new(shape);
	copy_fields(source, object, shape, expression);
	return object;
}

adamic_object *adamic_object_copy(const adamic_object *source) { return adamic_object_copy_checked(source, NULL); }

// adamic_object_has is object.hasOwnProperty(name): one of the shape's own names, not a method on a
// prototype. A shape's names are C strings, so the lengths have to agree before the bytes do.
bool adamic_object_has(const adamic_object *object, const adamic_string *name) {
	for (size_t index = 0; index < object->shape->count; index++) {
		const char *field = object->shape->names[index];
		size_t length = strlen(field);
		if (adamic_object_orders(object)[index] != 0 && length == name->length && memcmp(field, name->bytes, length) == 0) {
			return true;
		}
	}
	return false;
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
	if (object->dynamic_shape || cache->shape != shape) {
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

// A field made as undefined alone holds NULL, whereas number | undefined holds a packed number.
// The shape decides which union member is live; reading NULL's bits as a double would produce 0.
adamic_maybe_number adamic_object_maybe_number(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_value *slot = adamic_object_field(object, name, cache);
	if (object->shape->references[cache->index]) {
		if (slot->reference != NULL) {
			static const char message[] = "compiler bug: a numeric field holds a reference";
			adamic_panic(message, sizeof message - 1);
		}
		return (adamic_maybe_number){false, 0.0};
	}
	return adamic_maybe_number_unpack(slot->number);
}

// Cache absence too, with count as the index, without adding a field to the object's shape.
adamic_value *adamic_object_optional_find(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	cache->shape = object->shape;
	cache->index = object->shape->count;
	for (size_t index = 0; index < object->shape->count; index++) {
		if (strcmp(object->shape->names[index], name) == 0) {
			cache->index = index;
			break;
		}
	}
	if (cache->index == object->shape->count || adamic_object_orders(object)[cache->index] == 0) {
		return NULL;
	}
	return &((adamic_object *)object)->slots[cache->index];
}

// The caller supplies the source expression, so both backends name the same failed read.
adamic_value *adamic_object_read(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression) {
	adamic_value *slot = adamic_object_optional_field(object, name, cache);
	if (slot != NULL && object->class != NULL && object->class->is_static) {
		size_t flag = object->class->static_flags[cache->index];
		if (flag != 0 && object->slots[flag - 1].number == 0 && object->class->static_parent != 0) {
			return adamic_object_read(object->slots[object->class->static_parent - 1].reference, name, cache, expression);
		}
	}
	if (slot == NULL || !adamic_object_initialized(object)[cache->index]) {
		size_t capacity = strlen(name) + strlen(expression) + sizeof "read before assignment: field '' in ";
		char *message = malloc(capacity);
		if (message == NULL) {
			static const char failure[] = "out of memory";
			adamic_panic(failure, sizeof failure - 1);
		}
		int length = snprintf(message, capacity, "read before assignment: field '%s' in %s", name, expression);
		adamic_panic(message, (size_t)length);
	}
	return slot;
}

void adamic_object_set_initialized(adamic_object *object, const char *name, bool initialized) {
	adamic_slot_cache cache = {NULL, 0};
	(void)adamic_object_field(object, name, &cache);
	adamic_object_initialized(object)[cache.index] = initialized;
	adamic_object_present(object, cache.index);
}

void adamic_object_absent(adamic_object *object, size_t index) {
	adamic_object_orders(object)[index] = 0;
}

void adamic_object_present(adamic_object *object, size_t index) {
	size_t *orders = adamic_object_orders(object);
	if (orders[index] != 0) return;
	size_t last = 0;
	for (size_t at = 0; at < object->shape->count; at++) {
		if (orders[at] > last) last = orders[at];
	}
	orders[index] = last + 1;
}

bool adamic_object_delete(adamic_object *object, const adamic_string *key) {
	for (size_t index = 0; index < object->shape->count; index++) {
		const char *name = object->shape->names[index];
		if (strlen(name) != key->length || memcmp(name, key->bytes, key->length) != 0) continue;
		adamic_object_check_write(object, name);
		if (object->shape->references[index]) {
			adamic_release(object->slots[index].reference);
			object->slots[index].reference = NULL;
		} else {
			object->slots[index].number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
		}
		adamic_object_absent(object, index);
		return true;
	}
	return true;
}

adamic_object *adamic_object_copy_reserving(const adamic_object *source, const adamic_shape *reserved) {
	return adamic_object_copy_reserving_checked(source, reserved, NULL);
}

adamic_object *adamic_object_copy_reserving_checked(const adamic_object *source, const adamic_shape *reserved, const char *expression) {
	const adamic_shape *base = source == NULL ? NULL : (source->class == NULL ? source->shape : source->class->public_shape);
	size_t count = base == NULL ? 0 : base->count;
	size_t extra = reserved == NULL ? 0 : reserved->count;
	if (base != NULL && !source->dynamic_shape) {
		bool adds = false;
		for (size_t index = 0; index < extra; index++) {
			bool found = false;
			for (size_t at = 0; at < count; at++) {
				if (strcmp(base->names[at], reserved->names[index]) == 0) found = true;
			}
			if (!found) adds = true;
		}
		if (!adds) return adamic_object_copy_checked(source, expression);
	}
	size_t capacity = count + extra;
	size_t size = sizeof(adamic_object) + capacity * (sizeof(adamic_value) + sizeof(size_t) + 1);
	// The byte tail must not misalign the descriptor or its pointer/int arrays.
	size_t alignment = _Alignof(adamic_shape);
	size = (size + alignment - 1) / alignment * alignment;
	adamic_object *object = adamic_allocate(size + sizeof(adamic_shape) + capacity * (sizeof(const char *) + sizeof(int) + sizeof(bool)), adamic_kind_object);
	// Reserve capacity for values and orders before the descriptor and its arrays.
	adamic_shape *shape = (adamic_shape *)(void *)((unsigned char *)object + size);
	const char **names = (const char **)(void *)(shape + 1);
	int *types = (int *)(void *)(names + capacity);
	bool *references = (bool *)(void *)(types + capacity);
	for (size_t index = 0; index < count; index++) {
		names[index] = base->names[index];
		references[index] = base->references[index];
		types[index] = source->dynamic_shape ? source->dynamic_types[index] : adamic_shape_type(base, index);
	}
	for (size_t index = 0; index < extra; index++) {
		bool found = false;
		for (size_t at = 0; at < count; at++) {
			if (strcmp(names[at], reserved->names[index]) == 0) found = true;
		}
		if (!found) {
			names[count] = reserved->names[index];
			references[count] = reserved->references[index];
		types[count++] = adamic_shape_type(reserved, index);
		}
	}
	*shape = (adamic_shape){count, names, references, base == NULL ? NULL : base->methods};
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->dynamic_shape = true;
	object->dynamic_types = types;
	memset(object->slots, 0, count * sizeof(adamic_value));
	memset(adamic_object_orders(object), 0, count * sizeof(size_t));
	memset(adamic_object_initialized(object), 1, count);
	for (size_t index = base == NULL ? 0 : base->count; index < count; index++) {
		if (!references[index]) object->slots[index].number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
	}
	if (source != NULL) copy_fields(source, object, base, expression);
	return object;
}
