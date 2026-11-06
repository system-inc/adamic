// Packed number | undefined keys keep undefined separate from every present NaN.
#include "adamic.h"
#include "graph_regions.h"
#include <string.h>

bool adamic_map_maybe_key_equal(double left, double right) {
	adamic_maybe_number a = adamic_maybe_number_unpack(left);
	adamic_maybe_number b = adamic_maybe_number_unpack(right);
	return a.present == b.present && (!a.present || a.number == b.number || (isnan(a.number) && isnan(b.number)));
}

uint64_t adamic_map_maybe_key_hash(double key) {
	adamic_maybe_number value = adamic_maybe_number_unpack(key);
	if (!value.present) {
		return ADAMIC_UNDEFINED_BITS;
	}
	return adamic_map_number_hash(value.number);
}

// V8 13.6.233.17: Object::GetSimpleHash (src/objects/objects-inl.h),
// ComputeUnseededHash and ComputeLongHash (src/utils/utils.h). Unsigned arithmetic
// wraps exactly as V8's does; check the signed range before converting a double.
uint64_t adamic_map_number_hash(double number) {
	if (isnan(number)) {
		return 0x3fffffffu;
	}
	if (number >= INT32_MIN && number <= INT32_MAX && trunc(number) == number) {
		uint32_t hash = (uint32_t)(int32_t)number;
		hash = ~hash + (hash << 15);
		hash ^= hash >> 12;
		hash += hash << 2;
		hash ^= hash >> 4;
		hash *= 2057;
		hash ^= hash >> 16;
		return hash & 0x3fffffffu;
	}
	uint64_t hash;
	memcpy(&hash, &number, sizeof hash);
	hash = ~hash + (hash << 18);
	hash ^= hash >> 31;
	hash *= 21;
	hash ^= hash >> 11;
	hash += hash << 6;
	hash ^= hash >> 22;
	return hash & 0x3fffffffu;
}

static const char *const iterator_names[] = {"next"};
static const bool iterator_references[] = {true};
static const adamic_field_kind iterator_kinds[] = {adamic_field_reference};
static const adamic_shape iterator_shape = {1, iterator_names, iterator_references, &adamic_host_methods, iterator_kinds};
static const char *const state_names[] = {"iterator", "part", "key", "value", "set"};
static const bool state_references[] = {true, false, false, false, false};
static const adamic_field_kind state_kinds[] = {adamic_field_reference, adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_boolean};
static const adamic_shape state_shape = {5, state_names, state_references, &adamic_host_methods, state_kinds};
// An iterator's result holds value, then done, the order Node's { value, done } has, so its keys
// enumerate as Node's do (#7nw4kgj).
static const char *const result_names[] = {"value", "done"};
static const bool result_reference[] = {true, false};
static const bool result_scalar[] = {false, false};
static const adamic_field_kind result_kinds[][2] = {{adamic_field_number, adamic_field_boolean}, {adamic_field_boolean, adamic_field_boolean}, {adamic_field_reference, adamic_field_boolean}};
static const adamic_shape result_shapes[] = {{2, result_names, result_scalar, NULL, result_kinds[0]}, {2, result_names, result_scalar, NULL, result_kinds[1]}, {2, result_names, result_reference, NULL, result_kinds[2]}};
// An iterator's result is a plain object on Node, as the iterator itself isn't, so no result shape is
// marked host (adamic.h), where the iterator's and its state's are, and Object.prototype's names
// answer on a result by a run-time key (union.c). Each is registered as a literal's is, with zero
// scalar tags, no evidence: a result's own representation bytes, written below, vary with what it
// holds, and union.c's dynamic_slot reads those first. An entries() pair is an array on Node, a
// tuple here, and marked host as well.
static const int result_types[2];
static adamic_shape_types result_metadata[] = {{&result_shapes[0], result_types, NULL, false}, {&result_shapes[1], result_types, NULL, false}, {&result_shapes[2], result_types, NULL, false}};
static const char *const pair_names[] = {"0", "1"};
static const bool pair_references[][2] = {{false, false}, {false, true}, {true, false}, {true, true}};
static const adamic_field_kind pair_kinds[][2] = {{adamic_field_number, adamic_field_number}, {adamic_field_number, adamic_field_boolean}, {adamic_field_number, adamic_field_reference}, {adamic_field_boolean, adamic_field_number}, {adamic_field_boolean, adamic_field_boolean}, {adamic_field_boolean, adamic_field_reference}, {adamic_field_reference, adamic_field_number}, {adamic_field_reference, adamic_field_boolean}, {adamic_field_reference, adamic_field_reference}};
static const adamic_shape pair_shapes[] = {{2, pair_names, pair_references[0], &adamic_host_methods, pair_kinds[0]}, {2, pair_names, pair_references[0], &adamic_host_methods, pair_kinds[1]}, {2, pair_names, pair_references[1], &adamic_host_methods, pair_kinds[2]}, {2, pair_names, pair_references[0], &adamic_host_methods, pair_kinds[3]}, {2, pair_names, pair_references[0], &adamic_host_methods, pair_kinds[4]}, {2, pair_names, pair_references[1], &adamic_host_methods, pair_kinds[5]}, {2, pair_names, pair_references[2], &adamic_host_methods, pair_kinds[6]}, {2, pair_names, pair_references[2], &adamic_host_methods, pair_kinds[7]}, {2, pair_names, pair_references[3], &adamic_host_methods, pair_kinds[8]}};

// adamic_iterator_result (adamic.h) is the one place a step result is built, a collection's and a
// string's alike. At done a number or a boolean kind's value holds the packed absent maybe-number, as
// its representation byte then says MaybeNumber, which union.c's dynamic_slot reads first.
adamic_object *adamic_iterator_result(adamic_field_kind kind, bool present, adamic_value element, unsigned char element_tag) {
	adamic_object *result = adamic_object_new(&result_shapes[kind]);
	adamic_register_shape_types(&result_metadata[kind]);
	result->slots[1].boolean = !present;
	adamic_object_field_types(result)[1] = adamic_ir_Boolean;
	adamic_object_field_types(result)[0] = element_tag;
	if (present) {
		result->slots[0] = element;
	} else if (kind == adamic_field_reference) {
		result->slots[0].reference = NULL;
	} else {
		result->slots[0].number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
	}
	return result;
}

// A collection's element is one of ir.Type's scalar representations Number, Boolean and MaybeNumber,
// or counted. Anything else, MaybeBoolean among them, would be retained as a pointer it isn't, so it
// stops as a compiler bug: the lowering keeps every other representation out of a collection.
static bool collection_reference(int type) {
	if (type == adamic_ir_Number || type == adamic_ir_Boolean || type == adamic_ir_MaybeNumber) {
		return false;
	}
	if (adamic_ir_type_reference((enum adamic_ir_type)type)) {
		return true;
	}
	char message[160];
	int length = snprintf(message, sizeof message, "compiler bug: a collection element of representation %d is neither a number, a boolean, a maybe-number nor counted", type);
	adamic_panic(message, length < 0 ? 0 : (size_t)length < sizeof message ? (size_t)length : sizeof message - 1);
}

static adamic_field_kind collection_kind(int type) {
 return collection_reference(type) ? adamic_field_reference : type == adamic_ir_Boolean ? adamic_field_boolean : adamic_field_number;
}

static adamic_code_function collection_next;

static adamic_value collection_next(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	adamic_object *state = self->cells[0]->value.reference;
	adamic_map_iterator *iterator = state->slots[0].reference;
	int part = (int)state->slots[1].number;
	int key_type = (int)state->slots[2].number;
	int value_type = (int)state->slots[3].number;
	bool set = state->slots[4].boolean;
	bool reference = part == 3 || collection_reference(part == 1 || set ? key_type : value_type);
	adamic_field_kind result_kind = part == 3 ? adamic_field_reference : collection_kind(part == 1 || set ? key_type : value_type);
	adamic_value key, value;
	bool present = adamic_map_iterator_next(iterator, &key, &value);
	unsigned char element_tag = reference ? (part == 3 ? 4 : (unsigned char)(part == 1 || set ? key_type : value_type)) : (unsigned char)(!present || (part == 1 || set ? key_type : value_type) != 2 ? 7 : 2);
	if (!present) {
		return (adamic_value){.reference = adamic_iterator_result(result_kind, false, (adamic_value){0}, element_tag)};
	}
	if (set) { value = key; value_type = key_type; }
	adamic_value element;
	if (part == 3) {
		int shape = collection_kind(key_type) * 3 + collection_kind(value_type);
		adamic_object *pair = adamic_object_new(&pair_shapes[shape]);
		pair->tuple = true;
		pair->slots[0] = key;
		pair->slots[1] = value;
		adamic_object_field_types(pair)[0] = (unsigned char)key_type;
		adamic_object_field_types(pair)[1] = (unsigned char)value_type;
		if (collection_reference(key_type)) { adamic_retain(key.reference); }
		if (collection_reference(value_type)) { adamic_retain(value.reference); }
		element.reference = pair;
	} else {
		int type = part == 1 || set ? key_type : value_type;
		element = part == 1 || set ? key : value;
		// The map lends its element, and the result is given one of its own.
		if (reference) { adamic_retain(element.reference); }
		if (type == adamic_ir_Number) {
			// IteratorResult<number, undefined>.value reads the same packed word until narrowed.
			element.number = adamic_maybe_number_pack((adamic_maybe_number){true, element.number});
		}
	}
	return (adamic_value){.reference = adamic_iterator_result(result_kind, true, element, element_tag)};
}

bool adamic_collection_iterator_is(const void *value) {
	if (value == NULL || ((const adamic_heap *)value)->kind != adamic_kind_object) {
		return false;
	}
	const adamic_object *object = value;
	if (object->shape != &iterator_shape) {
		return false;
	}
	const adamic_closure *next = object->slots[0].reference;
	return next != NULL && next->code == collection_next;
}

adamic_object *adamic_collection_iterator(adamic_map *collection, int part, int key, int value, bool set) {
	adamic_object *state = adamic_object_new(&state_shape);
	state->slots[0].reference = adamic_map_iterate(collection);
	state->slots[1].number = part;
	state->slots[2].number = key;
	state->slots[3].number = value;
	state->slots[4].boolean = set;
	if (adamic_graph_is(collection)) { state = adamic_graph_adopt_owned(state, adamic_object_size(state->shape->count)); }
	adamic_closure *next = adamic_closure_new(collection_next, 1);
#ifdef ADAMIC_CLOSURE_DISCARDS
	// Its IteratorResult comes back owned.
	next->discard = adamic_discard_reference;
#endif
	next->cells[0] = adamic_cell_new((adamic_value){.reference = state}, true);
	if (adamic_graph_is(collection)) {
		next->cells[0] = adamic_graph_adopt_owned(next->cells[0], sizeof(adamic_cell));
		next = adamic_graph_adopt_owned(next, sizeof *next + sizeof(adamic_cell *));
	}
	adamic_object *object = adamic_object_new(&iterator_shape);
	object->slots[0].reference = next;
	if (adamic_graph_is(collection)) { object = adamic_graph_adopt_owned(object, adamic_object_size(object->shape->count)); }
	return object;
}

static const char *const string_state_names[] = {"string", "offset"};
static const bool string_state_references[] = {true, false};
static const adamic_field_kind string_state_kinds[] = {adamic_field_reference, adamic_field_number};
static const adamic_shape string_state_shape = {2, string_state_names, string_state_references, &adamic_host_methods, string_state_kinds};

static adamic_code_function string_next;

static adamic_value string_next(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	adamic_object *state = self->cells[0]->value.reference;
	adamic_string *string = state->slots[0].reference;
	size_t offset = (size_t)state->slots[1].number;
	if (offset >= string->length) {
		return (adamic_value){.reference = adamic_iterator_result(adamic_field_reference, false, (adamic_value){0}, adamic_ir_String)};
	}
	size_t size = adamic_string_next(string, offset);
	state->slots[1].number = (double)(offset + size);
	adamic_value element = {.reference = adamic_string_slice_bytes(string, offset, size)};
	return (adamic_value){.reference = adamic_iterator_result(adamic_field_reference, true, element, adamic_ir_String)};
}

adamic_object *adamic_string_iterator(adamic_string *string) {
	adamic_object *state = adamic_object_new(&string_state_shape);
	adamic_retain(string);
	state->slots[0].reference = string;
	adamic_closure *next = adamic_closure_new(string_next, 1);
	next->cells[0] = adamic_cell_new((adamic_value){.reference = state}, true);
	adamic_object *object = adamic_object_new(&iterator_shape);
	object->slots[0].reference = next;
	return object;
}
