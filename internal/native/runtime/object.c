// object.c: plain objects, each carrying its shape.

#include "adamic.h"

#include <string.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(sizeof *object + shape->count * sizeof object->slots[0], adamic_kind_object);
	object->shape = shape;
	object->class = NULL;
	object->prototype = NULL;
	object->frozen = false;
	object->sealed = false;
	object->nonextensible = false;
	object->has_captured_stack = false;
	object->captured_stack.reference = NULL;
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	return object;
}

adamic_object *adamic_object_copy(const adamic_object *source) {
	const adamic_shape *shape = source->class == NULL ? source->shape : source->class->public_shape;
	adamic_object *object = adamic_object_new(shape);
	for (size_t position = 0; position < shape->count; position++) {
		size_t index = adamic_public_index(shape, position);
		// Captured stack is non-enumerable, even when it replaced an own field.
		if (source->has_captured_stack && strcmp(shape->names[index], "stack") == 0) { continue; }
		adamic_slot_cache cache = {0};
		const adamic_accessor *accessor = adamic_accessor_find(source, shape->names[index]);
		object->slots[index] = accessor == NULL ? *adamic_object_field(source, shape->names[index], &cache) : adamic_accessor_get((adamic_object *)source, shape->names[index]);
		if (shape->references[index] && accessor == NULL) {
			adamic_retain(object->slots[index].reference);
		}
	}
	return object;
}

// adamic_object_has is object.hasOwnProperty(name): one of the shape's own names, not a method on a
// prototype. A shape's names are C strings, so the lengths have to agree before the bytes do.
bool adamic_object_has(const adamic_object *object, const adamic_string *name) {
	if (object->has_captured_stack && name->length == 5 && memcmp(name->bytes, "stack", 5) == 0) { return true; }
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

// A field made as undefined alone holds NULL, whereas number | undefined holds a packed number.
// The shape decides which union member is live; reading NULL's bits as a double would produce 0.
adamic_maybe_number adamic_object_maybe_number(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_value *slot = adamic_object_field(object, name, cache);
	adamic_value *own = object->class != NULL && object->class->is_static ? adamic_object_find(object, name, cache) : slot;
	if (object->shape->references[(size_t)(own - object->slots)]) {
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
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	size_t slot = packed >> 48;
	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)object->shape) {
		slot = object->shape->count;
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(object->shape->names[index], name) == 0) {
				slot = index;
				break;
			}
		}
		adamic_slot_cache_store(cache, object->shape, slot);
	}
	if (slot == object->shape->count) { return NULL; }
	return &((adamic_object *)object)->slots[slot];
}

// Frame text is intentionally unspecified. Capturing still creates a real own,
// writable, non-enumerable string property; no V8 frame text enters the oracle.
void adamic_error_capture_stack(adamic_object *target) {
	adamic_object_check_data_write(target, "stack");
	adamic_retain(&adamic_string_empty);
	if (target->has_captured_stack) { adamic_release(target->captured_stack.reference); }
	target->captured_stack.reference = &adamic_string_empty;
	target->has_captured_stack = true;
}

adamic_string *adamic_error_read_stack(const adamic_object *target) {
	adamic_slot_cache cache = {0};
	adamic_value *slot = adamic_object_optional_field(target, "stack", &cache);
	return slot == NULL ? NULL : adamic_retain(slot->reference);
}
