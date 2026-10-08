// object.c: plain objects, each carrying its shape.

#include "adamic.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(adamic_object_size(shape->count), adamic_kind_object);
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->real_type = "record";
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	memset(adamic_object_initialized(object), 1, shape->count);
	memset(adamic_object_field_types(object), 0, shape->count);
	memset(adamic_object_contracts(object), 0, shape->count * sizeof(unsigned int));
	return object;
}

adamic_object *adamic_object_copy_checked(const adamic_object *source, const char *expression) {
	const adamic_shape *shape = source->class == NULL ? source->shape : source->class->public_shape;
	adamic_object *object = adamic_object_new(shape);
	for (size_t position = 0; position < shape->count; position++) {
		size_t index = adamic_public_index(shape, position);
		adamic_slot_cache cache = {NULL, 0};
		const adamic_accessor *accessor = adamic_accessor_find(source, shape->names[index]);
		object->slots[index] = accessor == NULL ? *(expression == NULL ? adamic_object_field(source, shape->names[index], &cache) : adamic_object_read(source, shape->names[index], &cache, expression)) : adamic_accessor_get((adamic_object *)source, shape->names[index]);
		if (accessor == NULL) {
			adamic_object_initialized(object)[index] = adamic_object_initialized(source)[cache.index];
			adamic_object_field_types(object)[index] = adamic_object_field_types(source)[cache.index];
			adamic_object_contracts(object)[index] = adamic_object_contracts(source)[cache.index];
		}
		if (shape->references[index] && accessor == NULL) {
			adamic_retain(object->slots[index].reference);
		}
	}
	return object;
}

adamic_object *adamic_object_copy(const adamic_object *source) { return adamic_object_copy_checked(source, NULL); }

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
	if (cache->index == object->shape->count) {
		return NULL;
	}
	return &((adamic_object *)object)->slots[cache->index];
}

// Both public reads use the worker's readiness check; only their diagnostics differ.
static adamic_value *adamic_object_read_mode(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner) {
	adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
	if (slot != NULL && object->class != NULL && object->class->is_static) {
		size_t flag = object->class->static_flags[cache->index];
		if (flag != 0 && object->slots[flag - 1].number == 0 && object->class->static_parent != 0) {
			return adamic_object_read_mode(object->slots[object->class->static_parent - 1].reference, name, cache, expression, expected, owner);
		}
	}
	if (slot == NULL || !adamic_object_initialized(object)[cache->index]) {
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

adamic_value *adamic_object_read(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression) {
	return adamic_object_read_mode(object, name, cache, expression, NULL, NULL);
}

void adamic_object_set_initialized(adamic_object *object, const char *name, bool initialized) {
	adamic_slot_cache cache = {NULL, 0};
	(void)adamic_object_field(object, name, &cache);
	adamic_object_initialized(object)[cache.index] = initialized;
}

// Required-field contract checks use the shared readiness bitmap. Representation evidence is
// checked before reading any union member, including before following a possible reference.
adamic_value adamic_object_view(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
	const adamic_object *owner = NULL;
	adamic_value *slot = adamic_object_read_mode(object, name, cache, expression, type, &owner);
	unsigned char actual = adamic_object_field_types(owner)[cache->index];
	// Undefined has no object payload. Its contract checks the storage tag before
	// reading a null reference, keeping an explicit null (tag 12) distinct.
	if (wanted == 13 && (actual == 13 ||
	    (((actual >= 3 && actual <= 6) || (actual == 8 || actual == 10 || actual == 11)) && slot->reference == NULL) ||
	    (actual == 7 && !adamic_maybe_number_unpack(slot->number).present) ||
	    (actual == 9 && !adamic_maybe_boolean_unpack(slot->maybe_boolean).present))) {
		return (adamic_value){.reference = NULL};
	}
	// Boxed unions and packed maybe-numbers have a real runtime tag. Convert only
	// after that tag proves which payload is live; never interpret a pointer as a number.
	if (actual == 10 && slot->reference != NULL) {
		const adamic_heap *boxed = slot->reference;
		if (wanted == 1 && boxed->kind == adamic_kind_number) { return (adamic_value){.number = ((const adamic_number_box *)boxed)->number}; }
		if (wanted == 2 && boxed->kind == adamic_kind_boolean) { return (adamic_value){.boolean = ((const adamic_boolean_box *)boxed)->boolean}; }
		if ((wanted >= 3 && wanted <= 6) || wanted == 8) {
			enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : wanted == 8 ? adamic_kind_closure : adamic_kind_map;
			if (boxed->kind == kind) { return *slot; }
		}
	}
	if (actual == 9 && wanted == 2) {
		adamic_maybe_boolean unpacked = adamic_maybe_boolean_unpack(slot->maybe_boolean);
		if (unpacked.present) { return (adamic_value){.boolean = unpacked.boolean}; }
	}
	if (actual == 7 && wanted == 1) {
		adamic_maybe_number unpacked = adamic_maybe_number_unpack(slot->number);
		if (unpacked.present) { return (adamic_value){.number = unpacked.number}; }
	}
	if (actual == wanted && ((wanted >= 1 && wanted <= 6) || wanted == 8)) {
		if (wanted <= 2) { return *slot; }
		const adamic_heap *reference = slot->reference;
		enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : wanted == 8 ? adamic_kind_closure : adamic_kind_map;
		if (reference != NULL && reference->kind == kind) { return *slot; }
	}
	const char *found = actual == 1 ? "number" : actual == 2 ? "boolean" : actual == 3 ? "string" : actual == 4 ? "object" : actual == 5 ? "array" : actual == 6 ? "Map" : actual == 7 ? "number" : actual == 8 ? "function" : actual == 9 ? "boolean" : actual == 11 ? "object" : actual == 12 ? "null" : actual == 13 ? "nullish" : "unsupported representation";
	if (actual >= 3 && actual <= 6 && slot->reference == NULL) { found = "nullish"; }
	if (actual == 7 && !adamic_maybe_number_unpack(slot->number).present) { found = "nullish"; }
	if (actual == 9 && !adamic_maybe_boolean_unpack(slot->maybe_boolean).present) { found = "nullish"; }
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
		unsigned char actual = adamic_object_field_types(object)[cache->index];
		if (actual == wanted || (actual == 10 && wanted <= 2) || (actual == 7 && wanted == 1)) { return; }
	}
	(void)adamic_object_view(object, name, cache, wanted, type, expression);
}

// Optional views share the required-field validator after proving presence. An
// absent slot or undefined payload must never be interpreted as numeric bits.
adamic_value adamic_object_optional_view(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression, bool absent, bool optional) {
 adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
 bool missing = slot == NULL && (absent || optional);
 if (slot != NULL && adamic_object_initialized(object)[cache->index]) {
  unsigned char actual = adamic_object_field_types(object)[cache->index];
  missing = actual == 13 || (actual >= 3 && actual <= 6 && slot->reference == NULL) || (actual == 7 && !adamic_maybe_number_unpack(slot->number).present) || (actual == 9 && !adamic_maybe_boolean_unpack(slot->maybe_boolean).present) || (actual == 10 && slot->reference == NULL);
 }
 if (missing) {
  if (wanted == 7) { return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0.0})}; }
  return (adamic_value){.reference = NULL};
 }
 unsigned char base = wanted == 7 ? 1 : wanted == 9 ? 2 : wanted;
 adamic_value result = adamic_object_view(object, name, cache, base, type, expression);
 if (wanted == 9) { result.reference = result.boolean ? &adamic_box_true : &adamic_box_false; }
 if (wanted == 7) { result.number = adamic_maybe_number_pack((adamic_maybe_number){true, result.number}); }
 return result;
}

// The write is admitted only by the real shape and its declared slot contract,
// never by the view's type or the old payload's physical representation.
void adamic_object_checked_write(adamic_object *object, const char *name, adamic_slot_cache *cache, const unsigned int *allowed, size_t count, const char *where) {
 adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
 if (slot != NULL) {
  unsigned int contract = adamic_object_contracts(object)[cache->index];
  for (size_t i = 0; i < count; i++) { if (contract != 0 && allowed[i] == contract) { return; } }
 }
 const char *real = object == NULL ? "undefined" : object->real_type;
 size_t capacity = strlen(name) + strlen(real) + strlen(where) + 100;
 char *message = malloc(capacity);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int length = snprintf(message, capacity, "field write failed: property '%s' on %s at %s has no compatible declared slot", name, real, where);
 adamic_panic(message, (size_t)length);
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
