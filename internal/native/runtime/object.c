// object.c: plain objects, each carrying its shape.

#include "adamic.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_object *object = adamic_allocate(sizeof *object + shape->count * (sizeof object->slots[0] + 2), adamic_kind_object);
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	memset(adamic_object_initialized(object), 1, shape->count);
	memset(adamic_object_field_types(object), 0, shape->count);
	return object;
}

adamic_object *adamic_object_copy_checked(const adamic_object *source, const char *expression) {
	if (adamic_record_is(source)) adamic_panic("NotYet: spread of record storage", sizeof "NotYet: spread of record storage" - 1);
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
	if (adamic_record_is(object)) return adamic_record_has_own(object, name);
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

adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method) {
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
			enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : adamic_kind_map;
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
		unsigned char actual = adamic_object_field_types(object)[cache->index];
		if (actual == wanted || (actual == 10 && wanted <= 2) || (actual == 7 && wanted == 1) || (actual == 9 && wanted == 2)) { return; }
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
	if (adamic_object_field_types(object)[cache->index] == 2) { return (adamic_maybe_boolean){true, slot->boolean}; }
	return adamic_maybe_boolean_unpack(slot->maybe_boolean);
}

static bool adamic_contract_proven(size_t count, const int *proofs, int source) {
 for (size_t index = 0; index < count; index++) { if (proofs[index] == source) { return true; } }
 return false;
}

// Compare allocation declarations, not a snapshot of a mutable object's values.
// This preserves a nested literal contract even when the candidate arrived through
// a broader source view. Getters and unavailable allocation metadata fail closed.
static bool adamic_contract_object(const adamic_field_contract *contract, const adamic_object *value) {
 if (value->shape->contracts == NULL) { return false; }
 if (adamic_contract_proven(contract->field_proof_count, contract->field_proofs, value->shape->contracts[-1].type_id)) { return true; }
 if (!contract->structural) { return false; }
 for (size_t index = 0; index < contract->field_count; index++) {
  adamic_slot_cache cache = {NULL, 0};
  adamic_value *slot = adamic_object_optional_field(value, contract->field_names[index], &cache);
  if (slot == NULL) { if (contract->field_optional[index]) { continue; } return false; }
  if (!adamic_object_initialized(value)[cache.index] || adamic_accessor_find(value, contract->field_names[index]) != NULL) { return false; }
  const adamic_field_contract *actual = &value->shape->contracts[cache.index];
  const adamic_field_contract *expected = &contract->field_contracts[index];
  if (!adamic_contract_proven(expected->field_proof_count, expected->field_proofs, actual->type_id)) { return false; }
 }
 return true;
}

// Check the incoming value before ownership changes or a store. The old payload is never
// evidence for a literal domain: the immutable actual shape carries the declaration.
void adamic_object_check_contract(adamic_object *object, const char *name, unsigned char kind, adamic_value value, int source_type, const char *expression) {
 adamic_slot_cache cache = {NULL, 0};
 adamic_value *slot = adamic_object_optional_field(object, name, &cache);
 const adamic_field_contract *contract = slot == NULL || object->shape->contracts == NULL ? NULL : &object->shape->contracts[cache.index];
 adamic_check_contract(contract, kind, value, source_type, expression);
}

void adamic_check_contract(const adamic_field_contract *contract, unsigned char kind, adamic_value value, int source_type, const char *expression) {
 bool present = true;
 if (kind >= 3 && kind <= 6) { present = value.reference != NULL; }
 if (kind == 7) { present = adamic_maybe_number_unpack(value.number).present; }
 if (kind == 9) { present = adamic_maybe_boolean_unpack(value.maybe_boolean).present; }
 bool valid = contract != NULL && contract->kind != 0 && (contract->kind == kind || (contract->kind == 1 && kind == 7) || ((contract->kind == 2 || contract->kind == 9) && (kind == 2 || kind == 9)) || (!present && contract->nullable && contract->kind>=3 && contract->kind<=6 && kind>=3 && kind<=6)) && (present || contract->nullable) && (!contract->nullish_only || !present);
 if (valid && present && contract->count != 0) {
  valid = false;
  for (size_t index = 0; index < contract->count; index++) {
   adamic_value allowed = contract->allowed[index];
   if ((kind == 1 || kind == 7) && (kind == 7 ? adamic_maybe_number_unpack(value.number).number : value.number) == allowed.number) { valid = true; }
   if ((kind == 2 || kind == 9) && (kind == 9 ? adamic_maybe_boolean_unpack(value.maybe_boolean).boolean : value.boolean) == allowed.boolean) { valid = true; }
   if (kind == 3 && adamic_string_equal(value.reference, allowed.reference)) { valid = true; }
  }
 }
 if (valid && present && contract->reference) {
  valid = adamic_contract_proven(contract->write_proof_count, contract->write_proofs, source_type);
  if (!valid && kind == 4 && ((const adamic_heap *)value.reference)->kind == adamic_kind_object) { valid = adamic_contract_object(contract, value.reference); }
  if (!valid && kind == 5 && ((const adamic_heap *)value.reference)->kind == adamic_kind_array) { valid = adamic_contract_proven(contract->field_proof_count,contract->field_proofs,((const adamic_array *)value.reference)->allocation_type); }
  if (!valid && kind == 6 && ((const adamic_heap *)value.reference)->kind == adamic_kind_map) { valid = adamic_contract_proven(contract->field_proof_count,contract->field_proofs,((const adamic_map *)value.reference)->allocation_type); }
 }
 if (valid) { return; }
 const char *expected = contract == NULL || contract->declared == NULL ? "unavailable field contract" : contract->declared;
 const char *got = kind == 1 ? "number" : kind == 2 ? "boolean" : kind == 3 ? "string" : kind == 4 ? "object" : kind == 5 ? "array" : kind == 6 ? "Map" : "unsupported representation";
 const adamic_string *text = NULL;
 if (!present) { got = "undefined"; }
 else if (kind == 1 || kind == 7) { text = adamic_string_from_number(kind == 7 ? adamic_maybe_number_unpack(value.number).number : value.number); }
 else if (kind == 2 || kind == 9) { got = (kind == 9 ? adamic_maybe_boolean_unpack(value.maybe_boolean).boolean : value.boolean) ? "true" : "false"; }
 else if (kind == 3) { text = value.reference; }
 size_t capacity = strlen(expression) + strlen(expected) + strlen(got) + (text == NULL ? 0 : text->length) + 100;
 char *message = malloc(capacity);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int prefix = snprintf(message, capacity, "write failed: %s expects %s, got %s", expression, expected, text == NULL ? got : "");
 if (text != NULL) { memcpy(message + prefix, text->bytes, text->length); }
 adamic_panic(message, (size_t)prefix + (text == NULL ? 0 : text->length));
}
