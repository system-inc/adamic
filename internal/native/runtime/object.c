// object.c: plain objects, each carrying its shape.

#include "adamic.h"
#include "view_unions_mixed.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(adamic_object_size(shape->count), adamic_kind_object);
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->tuple = false;
	object->dynamic_shape = false;
	object->dynamic_types = NULL;
	for (size_t index = 0; index < shape->count; index++) { adamic_object_orders(object)[index] = index + 1; }
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	memset(adamic_object_initialized(object), 1, shape->count);
	memset(adamic_object_field_types(object), 0, shape->count);
	return object;
}

// field_owned is adamic_object_field that also says which object holds the slot: a constructor's
// static field can live in a parent constructor until an own write shadows it (class_static.c).
// Per-slot metadata is indexed from that object's slots, never from the shared packed cache.
static adamic_value *field_owned(const adamic_object *object, const char *name, adamic_slot_cache *cache, const adamic_object **owner) {
	while (object->class != NULL && object->class->is_static) {
		adamic_value *slot = adamic_object_find(object, name, cache);
		size_t flag = object->class->static_flags[adamic_slot_index(object, slot)];
		if (flag == 0 || object->slots[flag - 1].number != 0 || object->class->static_parent == 0) {
			*owner = object;
			return slot;
		}
		object = object->slots[object->class->static_parent - 1].reference;
	}
	*owner = object;
	return adamic_object_data_field(object, name, cache);
}

static adamic_value *adamic_object_read_mode(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner);

// Fixed and dynamic descriptors share one copy of slot state.
static void copy_fields(const adamic_object *source, adamic_object *object, const adamic_shape *shape, const char *expression) {
	for (size_t position = 0; position < shape->count; position++) {
		size_t index = adamic_public_index(shape, position);
		adamic_slot_cache cache = {0};
		const adamic_accessor *accessor = adamic_accessor_find(source, shape->names[index]);
		if (accessor == NULL) {
			// The slot found says where the field is, not the cache: another thread may store a
			// different shape's slot in a shared cache (the packed cache, adamic.h).
			const adamic_object *owner = source;
			adamic_value *found = field_owned(source, shape->names[index], &cache, &owner);
			size_t at = adamic_slot_index(owner, found);
			if (expression != NULL && adamic_object_orders(owner)[at] != 0) {
				found = adamic_object_read_mode(source, shape->names[index], &cache, expression, NULL, &owner);
				at = adamic_slot_index(owner, found);
			}
			object->slots[index] = *found;
			adamic_object_initialized(object)[index] = adamic_object_initialized(owner)[at];
			adamic_object_field_types(object)[index] = adamic_object_field_types(owner)[at];
			if (source->class == NULL) adamic_object_orders(object)[index] = adamic_object_orders(owner)[at];
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
	if (adamic_record_is(source)) adamic_panic("NotYet: spread of record storage", sizeof "NotYet: spread of record storage" - 1);
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
	if (adamic_record_is(object)) return adamic_record_has_own(object, name);
	for (size_t index = 0; index < object->shape->count; index++) {
		const char *field = object->shape->names[index];
		size_t length = strlen(field);
		if (adamic_object_orders(object)[index] != 0 && length == name->length && memcmp(field, name->bytes, length) == 0) {
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
			// A dynamic shape's descriptor lives in its object, so its address can come back as
			// another object's layout: such a lookup is never cached.
			if (!object->dynamic_shape) { adamic_slot_cache_store(cache, object->shape, index); }
			return &mutable->slots[index];
		}
	}
	static const char message[] = "compiler bug: a field the checker proved is there is missing";
	adamic_panic(message, sizeof message - 1);
}

adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method) {
	const adamic_shape *shape = object->shape;
	// The cache's index counts the fields, then the methods after them.
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	size_t slot = packed >> 48;
	if (object->dynamic_shape || (packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)shape) {
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
		if (!object->dynamic_shape) { adamic_slot_cache_store(cache, shape, slot); }
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
	// A dynamic shape's descriptor lives in its object, so its address can come back as another
	// object's layout: never trust or fill the cache for one.
	if (object->dynamic_shape || (packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)object->shape) {
		slot = object->shape->count;
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(object->shape->names[index], name) == 0) {
				slot = index;
				break;
			}
		}
		if (!object->dynamic_shape) { adamic_slot_cache_store(cache, object->shape, slot); }
	}
	if (slot == object->shape->count || adamic_object_orders(object)[slot] == 0) { return NULL; }
	return &((adamic_object *)object)->slots[slot];
}

// Both public reads use the worker's readiness check; only their diagnostics differ.
static adamic_value *adamic_object_read_mode(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner) {
	adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
	if (slot != NULL && object->class != NULL && object->class->is_static) {
		size_t flag = object->class->static_flags[adamic_slot_index(object, slot)];
		if (flag != 0 && object->slots[flag - 1].number == 0 && object->class->static_parent != 0) {
			return adamic_object_read_mode(object->slots[object->class->static_parent - 1].reference, name, cache, expression, expected, owner);
		}
	}
	if (slot == NULL || !adamic_object_initialized(object)[adamic_slot_index(object, slot)]) {
		size_t capacity = strlen(name) + strlen(expression) + (expected == NULL ? 0 : strlen(expected)) + 100;
		char *message = malloc(capacity);
		if (message == NULL) {
			static const char failure[] = "out of memory";
			adamic_panic(failure, sizeof failure - 1);
		}
		int length = expected != NULL ? snprintf(message, capacity, "field read failed: %s is not initialized; expected %s, found %s", expression, expected, slot == NULL ? "missing" : "uninitialized") : snprintf(message, capacity, "read before assignment: field '%s' in %s", name, expression);
		adamic_panic(message, (size_t)length);
	}
	if (owner != NULL) { *owner = object; }
	return slot;
}

adamic_view_union_value adamic_object_view_union_snapshot(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *declared, bool absent) {
    adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
    if (object != NULL && slot == NULL && absent) {
        return (adamic_view_union_value){adamic_view_union_undefined, {.reference = NULL}};
    }
    const adamic_object *owner = NULL;
    slot = adamic_object_read_mode(object, name, cache, expression, declared, &owner);
    size_t index = adamic_slot_index(owner, slot);
    unsigned char storage = adamic_object_field_types(owner)[index];
    adamic_view_union_value value = {adamic_view_union_unknown, *slot};
    if (storage == adamic_rep_number) { value.kind = adamic_view_union_number; }
    else if (storage == adamic_rep_boolean) { value.kind = adamic_view_union_boolean; }
    else if (storage == adamic_rep_maybe_number) {
        adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
        value.kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;
        value.payload.number = number.number;
    } else if (storage == adamic_rep_maybe_boolean) {
        adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot->maybe_boolean);
        value.kind = boolean.present ? adamic_view_union_boolean : adamic_view_union_undefined;
        value.payload.boolean = boolean.boolean;
    } else if (storage == adamic_rep_null) { value.kind = adamic_view_union_null; }
    else if (storage == adamic_rep_undefined) { value.kind = adamic_view_union_undefined; }
    else if ((storage >= adamic_rep_string && storage <= adamic_rep_map) || storage == adamic_rep_closure || storage == adamic_rep_union || storage == adamic_rep_weak) {
        value = adamic_view_union_heap(slot->reference);
        adamic_view_union_kind expected = storage == adamic_rep_string ? adamic_view_union_string : storage == adamic_rep_object || storage == adamic_rep_weak ? adamic_view_union_object : storage == adamic_rep_array ? adamic_view_union_array : storage == adamic_rep_map ? adamic_view_union_map : storage == adamic_rep_closure ? adamic_view_union_function : value.kind;
        if (value.kind != adamic_view_union_undefined && value.kind != expected) { value.kind = adamic_view_union_unknown; }
    }
    return value;
}

adamic_value *adamic_object_read(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression) {
	return adamic_object_read_mode(object, name, cache, expression, NULL, NULL);
}

void adamic_object_set_initialized(adamic_object *object, const char *name, bool initialized) {
	adamic_slot_cache cache = {0};
	// The object's own field: readiness belongs to the object being initialized, never a parent's.
	adamic_value *slot = adamic_object_find(object, name, &cache);
	adamic_object_initialized(object)[adamic_slot_index(object, slot)] = initialized;
	adamic_object_present(object, adamic_slot_index(object, slot));
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
	size_t size = sizeof(adamic_object) + capacity * (sizeof(adamic_value) + sizeof(size_t) + 2);
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
	object->tuple = false;
	object->dynamic_shape = true;
	object->dynamic_types = types;
	memset(object->slots, 0, count * sizeof(adamic_value));
	memset(adamic_object_orders(object), 0, count * sizeof(size_t));
	memset(adamic_object_initialized(object), 1, count);
	memset(adamic_object_field_types(object), 0, count);
	for (size_t index = base == NULL ? 0 : base->count; index < count; index++) {
		if (!references[index]) object->slots[index].number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
	}
	if (source != NULL) copy_fields(source, object, base, expression);
	return object;
}

// Required-field contract checks use the shared readiness bitmap. Representation evidence is
// checked before reading any union member, including before following a possible reference.
adamic_value adamic_object_view(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
	if (object != NULL && adamic_record_is(object)) return adamic_record_view(object, name, wanted, type, expression);
	const adamic_object *owner = NULL;
	adamic_value *slot = adamic_object_read_mode(object, name, cache, expression, type, &owner);
	unsigned char actual = adamic_object_field_types(owner)[adamic_slot_index(owner, slot)];
	// Boxed unions and packed maybe-numbers have a real runtime tag. Convert only
	// after that tag proves which payload is live; never interpret a pointer as a number.
	if (actual == 10 && slot->reference != NULL) {
		const adamic_heap *boxed = slot->reference;
		if (wanted == 1 && boxed->kind == adamic_kind_number) { return (adamic_value){.number = ((const adamic_number_box *)boxed)->number}; }
		if (wanted == 2 && boxed->kind == adamic_kind_boolean) { return (adamic_value){.boolean = ((const adamic_boolean_box *)boxed)->boolean}; }
		if (wanted >= 3 && wanted <= 6) {
			enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : adamic_kind_map;
			if (boxed->kind == kind) { return *slot; }
		}
	}
	if (actual == 7 && wanted == 1) {
		adamic_maybe_number unpacked = adamic_maybe_number_unpack(slot->number);
		if (unpacked.present) { return (adamic_value){.number = unpacked.number}; }
	}
	if (actual == wanted && wanted >= 1 && wanted <= 6) {
		if (wanted <= 2) { return *slot; }
		const adamic_heap *reference = slot->reference;
		enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : adamic_kind_map;
		if (reference != NULL && reference->kind == kind) { return *slot; }
	}
	const char *found = actual == 1 ? "number" : actual == 2 ? "boolean" : actual == 3 ? "string" : actual == 4 ? "object" : actual == 5 ? "array" : actual == 6 ? "Map" : actual == 7 ? "number" : actual == 8 ? "function" : actual == 11 ? "object" : "unsupported representation";
	if (actual >= 3 && actual <= 6 && slot->reference == NULL) { found = "nullish"; }
	if (actual == 7 && !adamic_maybe_number_unpack(slot->number).present) { found = "nullish"; }
	if (actual == 10) {
		const adamic_heap *boxed = slot->reference;
		found = boxed == NULL ? "nullish" : boxed->kind == adamic_kind_number ? "number" : boxed->kind == adamic_kind_boolean ? "boolean" : boxed->kind == adamic_kind_string ? "string" : boxed->kind == adamic_kind_object ? "object" : boxed->kind == adamic_kind_array ? "array" : boxed->kind == adamic_kind_map ? "Map" : "function";
	}
	size_t capacity = strlen(expression) + strlen(type) + strlen(found) + 100;
	char *message = malloc(capacity);
	if (message == NULL) {
		static const char oom[] = "out of memory";
		adamic_panic(oom, sizeof oom - 1);
	}
	int length = snprintf(message, capacity, "field read failed: %s is not a %s; expected %s, found %s", expression, type, type, found);
	adamic_panic(message, (size_t)length);
}

void adamic_view_literal_failure(const char *expression, const char *expected, unsigned char type, adamic_value value) {
	const adamic_string *text = type == 3 ? value.reference : type == 1 ? adamic_string_from_number(value.number) : value.boolean ? &adamic_string_true : &adamic_string_false;
	const char *kind = type == 3 ? "string" : type == 1 ? "number" : "boolean";
	size_t capacity = strlen(expression) + strlen(expected) + text->length + 100;
	char *message = malloc(capacity);
	if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
	int prefix = snprintf(message, capacity, "field read failed: %s expected %s, found %s ", expression, expected, kind);
	memcpy(message + prefix, text->bytes, text->length);
	adamic_panic(message, (size_t)prefix + text->length);
}

// Writes through an asserted view must not reinterpret or release a differently typed slot.
void adamic_object_view_write(adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
	adamic_value *slot = adamic_object_optional_field(object, name, cache);
	if (slot != NULL) {
		size_t index = adamic_slot_index(object, slot);
		unsigned char actual = adamic_object_field_types(object)[index];
		bool reference_write = (wanted >= 3 && wanted <= 6) || wanted == 8 || wanted == 10;
		if (actual == wanted || (actual == 13 && reference_write && object->shape->references[index]) || (actual == 10 && wanted <= 2) || (actual == 7 && wanted == 1)) { return; }
	}
	(void)adamic_object_view(object, name, cache, wanted, type, expression);
}

adamic_maybe_boolean adamic_object_maybe_boolean(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	const adamic_object *owner = object;
	adamic_value *slot = field_owned(object, name, cache, &owner);
	if (owner->shape->references[adamic_slot_index(owner, slot)]) {
		if (slot->reference != NULL) {
			static const char message[] = "compiler bug: a boolean field holds a reference";
			adamic_panic(message, sizeof message - 1);
		}
		return (adamic_maybe_boolean){false, false};
	}
	return adamic_maybe_boolean_unpack(slot->maybe_boolean);
}
