// Packed number | undefined keys keep undefined separate from every present NaN.
#include "adamic.h"
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
static const adamic_shape iterator_shape = {1, iterator_names, iterator_references, NULL};
static const char *const state_names[] = {"iterator", "part", "key", "value", "set"};
static const bool state_references[] = {true, false, false, false, false};
static const adamic_shape state_shape = {5, state_names, state_references, NULL};
static const char *const result_names[] = {"done", "value"};
static const bool result_reference[] = {false, true};
static const bool result_scalar[] = {false, false};
static const adamic_shape result_shapes[] = {{2, result_names, result_scalar, NULL}, {2, result_names, result_reference, NULL}};
static const char *const pair_names[] = {"0", "1"};
static const bool pair_references[][2] = {{false, false}, {false, true}, {true, false}, {true, true}};
static const adamic_shape pair_shapes[] = {{2, pair_names, pair_references[0], NULL}, {2, pair_names, pair_references[1], NULL}, {2, pair_names, pair_references[2], NULL}, {2, pair_names, pair_references[3], NULL}};

// These are ir.Type's scalar representations; all other accepted collection elements are counted.
static bool collection_reference(int type) {
	return type != 1 && type != 2 && type != 7;
}

static adamic_value collection_next(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	adamic_object *state = self->cells[0]->value.reference;
	adamic_map_iterator *iterator = state->slots[0].reference;
	int part = (int)state->slots[1].number;
	int key_type = (int)state->slots[2].number;
	int value_type = (int)state->slots[3].number;
	bool set = state->slots[4].boolean;
	bool reference = part == 3 || collection_reference(part == 1 || set ? key_type : value_type);
	adamic_object *result = adamic_object_new(&result_shapes[reference ? 1 : 0]);
	adamic_value key, value;
	bool present = adamic_map_iterator_next(iterator, &key, &value);
	result->slots[0].boolean = !present;
	adamic_object_field_types(result)[0] = 2;
	adamic_object_field_types(result)[1] = reference ? (part == 3 ? 4 : (unsigned char)(part == 1 || set ? key_type : value_type)) : (unsigned char)(!present || (part == 1 || set ? key_type : value_type) != 2 ? 7 : 2);
	if (!present) {
		if (!reference) {
			result->slots[1].number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
		}
		return (adamic_value){.reference = result};
	}
	if (set) { value = key; value_type = key_type; }
	if (part == 3) {
		int shape = (collection_reference(key_type) ? 2 : 0) + (collection_reference(value_type) ? 1 : 0);
		adamic_object *pair = adamic_object_new(&pair_shapes[shape]);
		pair->slots[0] = key;
		pair->slots[1] = value;
		adamic_object_field_types(pair)[0] = (unsigned char)key_type;
		adamic_object_field_types(pair)[1] = (unsigned char)value_type;
		if (collection_reference(key_type)) { adamic_retain(key.reference); }
		if (collection_reference(value_type)) { adamic_retain(value.reference); }
		result->slots[1].reference = pair;
	} else {
		int type = part == 1 || set ? key_type : value_type;
		adamic_value element = part == 1 || set ? key : value;
		if (reference) { adamic_retain(element.reference); }
		if (type == 1) {
			// IteratorResult<number, undefined>.value reads the same packed word until narrowed.
			element.number = adamic_maybe_number_pack((adamic_maybe_number){true, element.number});
		}
		result->slots[1] = element;
	}
	return (adamic_value){.reference = result};
}

adamic_object *adamic_collection_iterator(adamic_map *collection, int part, int key, int value, bool set) {
	adamic_object *state = adamic_object_new(&state_shape);
	state->slots[0].reference = adamic_map_iterate(collection);
	state->slots[1].number = part;
	state->slots[2].number = key;
	state->slots[3].number = value;
	state->slots[4].boolean = set;
	adamic_closure *next = adamic_closure_new(collection_next, 1);
	next->cells[0] = adamic_cell_new((adamic_value){.reference = state}, true);
	adamic_object *object = adamic_object_new(&iterator_shape);
	object->slots[0].reference = next;
	return object;
}
