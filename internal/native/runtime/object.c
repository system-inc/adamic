// object.c: plain objects, each carrying its shape.

#include "adamic.h"
#include "view_unions_mixed.h"
#include "namespace.h"

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
	object->write_order = NULL;
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	memset(adamic_object_initialized(object), 1, shape->count);
	memset(adamic_object_field_types(object), 0, shape->count);
	return object;
}

// Fixed and dynamic descriptors share one copy of slot state.
static void copy_fields(const adamic_object *source, adamic_object *object, const adamic_shape *shape, const char *expression) {
	for (size_t position = 0; position < shape->count; position++) {
		size_t index = adamic_public_index(shape, position);
		adamic_slot_cache cache = {NULL, 0};
		const adamic_accessor *accessor = adamic_accessor_find(source, shape->names[index]);
		if (accessor == NULL) {
			adamic_value *slot = adamic_object_field(source, shape->names[index], &cache);
			if (expression != NULL && adamic_object_present(source, cache.index)) slot = adamic_object_read(source, shape->names[index], &cache, expression);
			object->slots[index] = *slot;
			adamic_object_initialized(object)[index] = adamic_object_initialized(source)[cache.index];
			adamic_object_field_types(object)[index] = adamic_object_field_types(source)[cache.index];
			if (source->write_order != NULL) object->write_order[index] = source->write_order[cache.index];
			else adamic_object_publish(object, index);
		} else {
			object->slots[index] = adamic_accessor_get((adamic_object *)source, shape->names[index]);
			adamic_object_initialized(object)[index] = 1;
			adamic_object_field_types(object)[index] = (unsigned char)accessor->type;
			adamic_object_publish(object, index);
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
	adamic_object *object = source->write_order == NULL ? adamic_object_new(shape) : adamic_object_construct(shape);
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
		if (length == name->length && memcmp(field, name->bytes, length) == 0) {
			return adamic_object_present(object, index);
		}
	}
	return false;
}

adamic_value *adamic_object_find(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	name = adamic_error_field_name(object, name);
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

adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method) {
	if (adamic_namespace_is(object)) {
  adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
  const adamic_value *slot = adamic_record_get_own(object, &key);
  if (slot == NULL || slot->reference == NULL || ((const adamic_heap *)slot->reference)->kind != adamic_kind_closure) adamic_panic("namespace member is not a function", sizeof "namespace member is not a function" - 1);
  return slot->reference;
 }
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
	name = adamic_error_field_name(object, name);
	cache->shape = object->shape;
	cache->index = object->shape->count;
	for (size_t index = 0; index < object->shape->count; index++) {
		if (strcmp(object->shape->names[index], name) == 0) {
			cache->index = index;
			break;
		}
	}
	if (cache->index == object->shape->count || !adamic_object_present(object, cache->index)) {
		return NULL;
	}
	return &((adamic_object *)object)->slots[cache->index];
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
	adamic_slot_cache cache = {NULL, 0};
	(void)adamic_object_field(object, name, &cache);
	adamic_object_initialized(object)[cache.index] = initialized;
	adamic_object_publish(object, cache.index);
}

void adamic_object_absent(adamic_object *object, size_t index) {
 if (object->write_order == NULL) {
  size_t offset = sizeof(adamic_object) + object->shape->count * (sizeof(adamic_value) + 2);
  offset = (offset + _Alignof(size_t) - 1) & ~(size_t)(_Alignof(size_t) - 1);
  object->write_order = (size_t *)(void *)((unsigned char *)object + offset);
  for (size_t at = 0; at < object->shape->count; at++) object->write_order[at] = at;
 }
 object->write_order[index] = SIZE_MAX;
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
	size_t size = adamic_object_size(capacity);
	// The byte tail must not misalign the descriptor or its pointer/int arrays.
	size_t alignment = _Alignof(adamic_shape);
	size = (size + alignment - 1) / alignment * alignment;
	adamic_object *object = adamic_allocate(size + sizeof(struct adamic_dynamic_shape) + capacity * (sizeof(const char *) + sizeof(int) + sizeof(bool)), adamic_kind_object);
	// Reserve capacity for values and orders before the descriptor and its arrays.
	struct adamic_dynamic_shape *descriptor = (struct adamic_dynamic_shape *)(void *)((unsigned char *)object + size);
	adamic_shape *shape = &descriptor->shape;
	const char **names = (const char **)(void *)(descriptor + 1);
	int *types = (int *)(void *)(names + capacity);
	bool *references = (bool *)(void *)(types + capacity);
	for (size_t index = 0; index < count; index++) {
		names[index] = base->names[index];
		references[index] = base->references[index];
		types[index] = source->dynamic_shape ? ((const struct adamic_dynamic_shape *)(const void *)base)->types[index] : adamic_shape_storage(base, index);
	}
	for (size_t index = 0; index < extra; index++) {
		bool found = false;
		for (size_t at = 0; at < count; at++) {
			if (strcmp(names[at], reserved->names[index]) == 0) found = true;
		}
		if (!found) {
			names[count] = reserved->names[index];
			references[count] = reserved->references[index];
		types[count++] = adamic_shape_storage(reserved, index);
		}
	}
	*shape = (adamic_shape){count, names, references, base == NULL ? NULL : base->methods};
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->tuple = false;
	object->dynamic_shape = true;
	descriptor->types = types;
	memset(object->slots, 0, count * sizeof(adamic_value));
	size_t order_offset = sizeof(adamic_object) + count * (sizeof(adamic_value) + 2);
	order_offset = (order_offset + _Alignof(size_t) - 1) & ~(size_t)(_Alignof(size_t) - 1);
	object->write_order = (size_t *)(void *)((unsigned char *)object + order_offset);
	for (size_t index = 0; index < count; index++) object->write_order[index] = SIZE_MAX;
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
	unsigned char actual = adamic_object_field_types(owner)[cache->index];
	// Boxed unions and packed maybe-numbers have a real runtime tag. Convert only
	// after that tag proves which payload is live; never interpret a pointer as a number.
	if (actual == 10 && slot->reference != NULL) {
		const adamic_heap *boxed = slot->reference;
		if (wanted == 1 && boxed->kind == adamic_kind_number) { return (adamic_value){.number = ((const adamic_number_box *)boxed)->number}; }
		if (wanted == 2 && boxed->kind == adamic_kind_boolean) { return (adamic_value){.boolean = ((const adamic_boolean_box *)boxed)->boolean}; }
		if (wanted >= 3 && wanted <= 6) {
			enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : wanted == 6 ? adamic_kind_map : adamic_kind_closure;
			if (boxed->kind == kind) { return *slot; }
		}
	}
	if (actual == 7 && wanted == 1) {
		adamic_maybe_number unpacked = adamic_maybe_number_unpack(slot->number);
		if (unpacked.present) { return (adamic_value){.number = unpacked.number}; }
	}
	if (actual == wanted && wanted >= 1 && (wanted <= 6 || wanted == 8)) {
		if (wanted <= 2) { return *slot; }
		const adamic_heap *reference = slot->reference;
		enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : wanted == 6 ? adamic_kind_map : adamic_kind_closure;
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
	adamic_value *slot = adamic_object_optional_find(object, name, cache);
	if (slot == NULL && cache->index < object->shape->count) { slot = &object->slots[cache->index]; }
	if (slot != NULL) {
		size_t index = adamic_slot_index(object, slot);
		unsigned char actual = adamic_object_field_types(object)[index];
		bool reference_write = (wanted >= 3 && wanted <= 6) || wanted == 8 || wanted == 10;
		if (actual == wanted || (actual == 13 && reference_write && object->shape->references[index]) || (actual == 10 && wanted <= 2) || (actual == 7 && wanted == 1)) { return; }
	}
	(void)adamic_object_view(object, name, cache, wanted, type, expression);
}

adamic_maybe_boolean adamic_object_maybe_boolean(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_value *slot = adamic_object_field(object, name, cache);
	if (object->shape->references[cache->index]) {
		if (slot->reference != NULL) {
			static const char message[] = "compiler bug: a boolean field holds a reference";
			adamic_panic(message, sizeof message - 1);
		}
		return (adamic_maybe_boolean){false, false};
	}
	return adamic_maybe_boolean_unpack(slot->maybe_boolean);
}

// Optional presence is distinct from initialized storage and its live representation.
adamic_value adamic_object_optional_view(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
    adamic_view_union_value snapshot = adamic_object_view_union_snapshot(object, name, cache, expression, type, true);
    if (snapshot.kind == adamic_view_union_undefined) {
        if (wanted == 7) return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0})};
        if (wanted == 9) return (adamic_value){.maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){false, false})};
        if (wanted == 3) return (adamic_value){.reference = NULL};
    }
    if (wanted == 7 && snapshot.kind == adamic_view_union_number) return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){true, snapshot.payload.number})};
    if (wanted == 9 && snapshot.kind == adamic_view_union_boolean) return (adamic_value){.maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){true, snapshot.payload.boolean})};
    if (wanted == 3 && snapshot.kind == adamic_view_union_string) return snapshot.payload;
    return adamic_object_view(object, name, cache, wanted == 7 ? 1 : wanted == 9 ? 2 : wanted, type, expression);
}

static const char *view_storage_name(unsigned char type) {
    switch (type) {
    case 1: return "number";
    case 2: return "boolean";
    case 3: return "string";
    case 7: return "number | undefined";
    case 9: return "boolean | undefined";
    case 10: return "union";
    case 13: return "undefined";
    default: return "unsupported representation";
    }
}

// Validate the incoming value before touching the destination or publishing presence.
// The source slot keeps its representation even through a wider checked alias.
void adamic_object_view_store(adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_value value, unsigned char wanted) {
    adamic_value *slot = adamic_object_optional_find(object, name, cache);
    if (slot == NULL && cache->index < object->shape->count) slot = &object->slots[cache->index];
    unsigned char actual = slot == NULL ? 0 : adamic_object_field_types(object)[cache->index];
    if (slot != NULL && actual == 0) actual = (object->dynamic_shape ? ((const struct adamic_dynamic_shape *)(const void *)object->shape)->types[cache->index] : adamic_shape_type(object->shape, cache->index)) & 255;
    unsigned char incoming = wanted;
    if (wanted == 7) {
        adamic_maybe_number number = adamic_maybe_number_unpack(value.number);
        incoming = number.present ? 1 : 13;
        value.number = number.number;
    } else if (wanted == 9) {
        adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(value.maybe_boolean);
        incoming = boolean.present ? 2 : 13;
        value.boolean = boolean.boolean;
    } else if (wanted == 3 && value.reference == NULL) incoming = 13;
    bool fits = (actual == 3 && incoming == 13 && adamic_object_optional_storage(object, cache->index)) || (actual == incoming && actual >= 1 && actual <= 3) || (actual == 7 && (incoming == 1 || incoming == 13)) || (actual == 9 && (incoming == 2 || incoming == 13)) || (actual == 10 && (incoming == 1 || incoming == 2 || incoming == 3 || incoming == 13));
    if (!fits) {
        const char *expected = slot == NULL ? "missing storage" : view_storage_name(actual);
        const char *found = view_storage_name(incoming);
        size_t capacity = strlen(name) + strlen(expected) + strlen(found) + 100;
        char *message = malloc(capacity);
        if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
        int length = snprintf(message, capacity, "field write failed: %s expected %s, found %s", name, expected, found);
        adamic_panic(message, (size_t)length);
    }
    if (actual == 7) value.number = adamic_maybe_number_pack((adamic_maybe_number){incoming != 13, value.number});
    if (actual == 9) value.maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){incoming != 13, value.boolean});
    if (actual == 10) {
        if (incoming == 1) value.reference = adamic_box_number(value.number);
        else if (incoming == 2) value.reference = value.boolean ? &adamic_box_true : &adamic_box_false;
        else if (incoming == 13) value.reference = NULL;
    }
    if (object->shape->references[cache->index]) adamic_release(slot->reference);
    *slot = value;
    adamic_object_field_types(object)[cache->index] = actual;
    adamic_object_initialized(object)[cache->index] = 1;
    adamic_object_publish(object, cache->index);
}
