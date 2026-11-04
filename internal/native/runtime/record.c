// record.c: own string data properties, using Map's counted ordered hash table. Each is enumerable,
// writable and configurable, but for the attributes Object.defineProperty took away (adamic.h).
#include "adamic.h"
#include "graph_regions.h"

#include <math.h>
#include <stddef.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

// Wrappers let the ordinary iterative object cleanup release the table and iterator snapshots.
// Their fields are implementation storage, never the record's observable own properties.
// Tuple names 0 and 1, and key at slot 2, reuse offsets already covered by the compiler's
// global field-layout proof (fields.go). No new unchecked field offset is introduced.
static const char *const record_names[] = {"0"};
static const bool record_references[] = {true};
static const adamic_field_kind record_kinds[] = {adamic_field_reference};
// One wrapper shape for each adamic_field_kind the record's values have, alike in layout, so the shape
// says whether its table holds counted references, numbers (number | undefined packed) or booleans
// (boolean | undefined packed) without a header bit, a slot or a field offset of its own: a run-time
// key read boxes a value by it (adamic_dynamic_property_keyed). The three are alike in content and told
// apart by address alone, so a linker's identical-data folding (--icf=all, -fmerge-all-constants) would
// read every record as one kind; the records harness makes one of each and reads its kind back.
static const adamic_shape record_number_shape = {1, record_names, record_references, NULL, record_kinds};
static const adamic_shape record_boolean_shape = {1, record_names, record_references, NULL, record_kinds};
static const adamic_shape record_reference_shape = {1, record_names, record_references, NULL, record_kinds};
static const char *const iteration_names[] = {"0", "1", "key"};
static const bool iteration_references[] = {true, true, false};
static const adamic_field_kind iteration_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_number};
static const adamic_shape iteration_shape = {3, iteration_names, iteration_references, &adamic_host_methods, iteration_kinds};

static adamic_map *table(const adamic_record *record) {
	return record->slots[0].reference;
}

adamic_record *adamic_record_new(adamic_field_kind values) {
	const adamic_shape *shape = values == adamic_field_number ? &record_number_shape : values == adamic_field_boolean ? &record_boolean_shape : values == adamic_field_reference ? &record_reference_shape : NULL;
	if (shape == NULL) {
		static const char message[] = "compiler bug: a record of unknown value kind";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_record *record = adamic_object_new(shape);
	record->slots[0].reference = adamic_map_new(true, values == adamic_field_reference);
	return record;
}

adamic_value *adamic_record_get_own(const adamic_record *record, const adamic_string *key) {
	return adamic_map_get(table(record), (adamic_value){.reference = (void *)key});
}

bool adamic_record_has_own(const adamic_record *record, const adamic_string *key) {
	return adamic_record_get_own(record, key) != NULL;
}

static bool named(const adamic_string *key, const char *name) {
	size_t length = strlen(name);
	return key->length == length && memcmp(key->bytes, name, length) == 0;
}

// Object.getOwnPropertyNames(Object.prototype), observed on Node v24.19.0. This is called
// only on an own miss. Length dispatch precedes every byte comparison; no strlen or allocation.
static bool prototype_member(const adamic_string *key) {
	switch (key->length) {
	case 7:
		return memcmp(key->bytes, "valueOf", 7) == 0;
	case 8:
		return memcmp(key->bytes, "toString", 8) == 0;
	case 9:
		return memcmp(key->bytes, "__proto__", 9) == 0;
	case 11:
		return memcmp(key->bytes, "constructor", 11) == 0;
	case 13:
		return memcmp(key->bytes, "isPrototypeOf", 13) == 0;
	case 14:
		return memcmp(key->bytes, "hasOwnProperty", 14) == 0 ||
			memcmp(key->bytes, "toLocaleString", 14) == 0;
	case 16:
		return memcmp(key->bytes, "__defineGetter__", 16) == 0 ||
			memcmp(key->bytes, "__defineSetter__", 16) == 0 ||
			memcmp(key->bytes, "__lookupGetter__", 16) == 0 ||
			memcmp(key->bytes, "__lookupSetter__", 16) == 0;
	case 20:
		return memcmp(key->bytes, "propertyIsEnumerable", 20) == 0;
	default:
		return false;
	}
}

static void check_missing_member(const adamic_string *key) {
	if (prototype_member(key)) {
		static const char prefix[] = "record member '";
		static const char suffix[] = "' is missing; records hold own keys only";
		// A recognized member is at most 20 bytes. Build the diagnostic without allocating.
		char message[sizeof prefix - 1 + 20 + sizeof suffix - 1];
		memcpy(message, prefix, sizeof prefix - 1);
		memcpy(message + sizeof prefix - 1, key->bytes, key->length);
		memcpy(message + sizeof prefix - 1 + key->length, suffix, sizeof suffix - 1);
		adamic_panic(message, sizeof prefix - 1 + key->length + sizeof suffix - 1);
	}
}

adamic_value *adamic_record_get(const adamic_record *record, const adamic_string *key) {
	adamic_value *value = adamic_record_get_own(record, key);
	if (value == NULL) {
		check_missing_member(key);
	}
	return value;
}

bool adamic_record_has(const adamic_record *record, const adamic_string *key) {
	return adamic_record_get(record, key) != NULL;
}

// entry_of is the table entry holding a slot adamic_map_get gave.
static adamic_map_entry *entry_of(adamic_value *slot) {
	return (adamic_map_entry *)((char *)slot - offsetof(adamic_map_entry, value));
}

// let_go releases the key and value a refused write was handed, as adamic_map_set lets go of the
// key it already has and the value it replaces.
static void let_go(adamic_map *map, adamic_string *key, adamic_value value) {
	if (adamic_graph_is(map)) { adamic_graph_drop(map, key); } else { adamic_release(key); }
	if (map->reference_values) {
		if (adamic_graph_is(map)) { adamic_graph_drop(map, value.reference); } else { adamic_release(value.reference); }
	}
}

void adamic_record_define(adamic_record *record, adamic_string *key, adamic_value value) {
	adamic_map_set(table(record), (adamic_value){.reference = key}, value);
}

// same_number is SameValue on numbers: NaN is NaN, and +0 isn't -0.
static bool same_number(double left, double right) {
	if (isnan(left) || isnan(right)) {
		return isnan(left) && isnan(right);
	}
	return left == right && signbit(left) == signbit(right);
}

// same_value is SameValue on two values a record's slots hold. A reference is the same object, or a
// string, number or boolean of the same value, as a union slot boxes them: JavaScript compares those
// by value.
static bool same_value(const adamic_map *map, adamic_record_scalar scalar, adamic_value left, adamic_value right) {
	if (map->reference_values) {
		const adamic_heap *a = left.reference;
		const adamic_heap *b = right.reference;
		if (a == b) {
			return true;
		}
		if (a == NULL || b == NULL || a->kind != b->kind) {
			return false;
		}
		switch (a->kind) {
		case adamic_kind_string:
			return adamic_string_equal((const adamic_string *)a, (const adamic_string *)b);
		case adamic_kind_number:
			return same_number(((const adamic_number_box *)a)->number, ((const adamic_number_box *)b)->number);
		case adamic_kind_boolean:
			return ((const adamic_boolean_box *)a)->boolean == ((const adamic_boolean_box *)b)->boolean;
		default:
			return false;
		}
	}
	switch (scalar) {
	case adamic_record_booleans:
		return left.boolean == right.boolean;
	case adamic_record_maybe_numbers: {
		adamic_maybe_number a = adamic_maybe_number_unpack(left.number);
		adamic_maybe_number b = adamic_maybe_number_unpack(right.number);
		return a.present == b.present && (!a.present || same_number(a.number, b.number));
	}
	case adamic_record_numbers:
		return same_number(left.number, right.number);
	case adamic_record_references:
		break;
	}
	static const char message[] = "compiler bug: a record of scalars defined as one of references";
	adamic_panic(message, sizeof message - 1);
}

bool adamic_record_define_hidden(adamic_record *record, adamic_string *key, adamic_value value, adamic_record_scalar scalar) {
	adamic_map *map = table(record);
	adamic_value *slot = adamic_map_get(map, (adamic_value){.reference = key});
	if (slot != NULL && entry_of(slot)->fixed) {
		// Node's ValidateAndApplyPropertyDescriptor: a non-configurable, non-writable property takes a
		// define only of the value it has, which changes nothing.
		bool same = same_value(map, scalar, *slot, value);
		let_go(map, key, value);
		return same;
	}
	if (slot != NULL) {
		// Set finds the same entry and replaces its value there, so the key keeps its place, and it
		// stays configurable, as the descriptor doesn't say otherwise.
		adamic_map_entry *entry = entry_of(slot);
		entry->hidden = true;
		entry->readonly = true;
		adamic_map_set(map, (adamic_value){.reference = key}, value);
		return true;
	}
	adamic_map_set(map, (adamic_value){.reference = key}, value);
	// A key set didn't have goes last, with every attribute the descriptor leaves out false.
	adamic_map_entry *entry = &map->entries[map->used - 1];
	entry->hidden = true;
	entry->readonly = true;
	entry->fixed = true;
	return true;
}

void adamic_record_set(adamic_record *record, adamic_string *key, adamic_value value) {
	if (named(key, "__proto__")) {
		static const char message[] = "NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_record_define(record, key, value);
}

bool adamic_record_delete(adamic_record *record, const adamic_string *key) {
	// JavaScript delete succeeds even when no own property was there.
	(void)adamic_map_delete(table(record), (adamic_value){.reference = (void *)key});
	return true;
}

bool adamic_record_set_checked(adamic_record *record, adamic_string *key, adamic_value value) {
	if (named(key, "__proto__")) {
		static const char message[] = "NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_map *map = table(record);
	adamic_value *slot = adamic_map_get(map, (adamic_value){.reference = key});
	if (slot != NULL && entry_of(slot)->readonly) {
		let_go(map, key, value);
		return false;
	}
	adamic_map_set(map, (adamic_value){.reference = key}, value);
	return true;
}

bool adamic_record_delete_checked(adamic_record *record, const adamic_string *key) {
	adamic_map *map = table(record);
	adamic_value *slot = adamic_map_get(map, (adamic_value){.reference = (void *)key});
	if (slot != NULL && entry_of(slot)->fixed) {
		return false;
	}
	// JavaScript delete succeeds even when no own property was there.
	(void)adamic_map_delete(map, (adamic_value){.reference = (void *)key});
	return true;
}

// The messages are Node's (v24): an assignment to a read-only property, and a delete of a
// non-configurable one, in strict code, which every module is.
void adamic_record_refused(const adamic_string *key, bool deleting) {
	static const char assign_prefix[] = "Cannot assign to read only property '";
	static const char assign_suffix[] = "' of object '#<Object>'";
	static const char delete_prefix[] = "Cannot delete property '";
	static const char delete_suffix[] = "' of #<Object>";
	const char *prefix = deleting ? delete_prefix : assign_prefix;
	const char *suffix = deleting ? delete_suffix : assign_suffix;
	size_t prefix_length = deleting ? sizeof delete_prefix - 1 : sizeof assign_prefix - 1;
	size_t suffix_length = deleting ? sizeof delete_suffix - 1 : sizeof assign_suffix - 1;
	adamic_string *text = adamic_string_allocate(prefix_length + key->length + suffix_length);
	char *bytes = (char *)text->bytes;
	memcpy(bytes, prefix, prefix_length);
	memcpy(bytes + prefix_length, key->bytes, key->length);
	memcpy(bytes + prefix_length + key->length, suffix, suffix_length);
	adamic_object *error = adamic_error_new(text);
	adamic_release(text);
	static const char name[] = "TypeError";
	adamic_string *type_name = adamic_string_allocate(sizeof name - 1);
	memcpy((char *)type_name->bytes, name, sizeof name - 1);
	adamic_release(error->slots[0].reference);
	error->slots[0].reference = type_name;
	adamic_thrown = error;
}

size_t adamic_record_size(const adamic_record *record) {
	return table(record)->count;
}

// ECMA-262 array indices: canonical decimal strings in [0, 2^32 - 2]. Length is bounded
// before arithmetic, so arbitrarily long keys cannot overflow. -0 and 01 are ordinary strings.
static bool array_index(const adamic_string *key, uint32_t *value) {
	if (key->length == 0 || key->length > 10 || (key->length > 1 && key->bytes[0] == '0')) {
		return false;
	}
	uint64_t number = 0;
	for (size_t index = 0; index < key->length; index++) {
		unsigned char digit = (unsigned char)key->bytes[index];
		if (digit < '0' || digit > '9') {
			return false;
		}
		number = number * 10 + (digit - '0');
	}
	if (number >= UINT32_MAX) {
		return false;
	}
	*value = (uint32_t)number;
	return true;
}

typedef struct indexed_key {
	adamic_string *key;
	uint32_t index;
} indexed_key;

static int compare_indices(const void *left, const void *right) {
	uint32_t a = ((const indexed_key *)left)->index;
	uint32_t b = ((const indexed_key *)right)->index;
	return a < b ? -1 : a > b ? 1 : 0;
}

adamic_array *adamic_record_keys(const adamic_record *record) {
	const adamic_map *map = table(record);
	adamic_array *keys = adamic_array_new(map->count, true);
	size_t count = 0;
	for (size_t index = 0; index < map->used; index++) {
		uint32_t number;
		if (!map->entries[index].deleted && !map->entries[index].hidden && array_index(map->entries[index].key.reference, &number)) {
			count++;
		}
	}
	indexed_key *indices = NULL;
	if (count > 0) {
		if (count > SIZE_MAX / sizeof *indices) {
			adamic_panic("out of memory", sizeof "out of memory" - 1);
		}
		indices = malloc(count * sizeof *indices);
		if (indices == NULL) {
			adamic_panic("out of memory", sizeof "out of memory" - 1);
		}
	}
	size_t written = 0;
	for (size_t index = 0; index < map->used; index++) {
		const adamic_map_entry *entry = &map->entries[index];
		uint32_t number;
		if (!entry->deleted && !entry->hidden && array_index(entry->key.reference, &number)) {
			indices[written++] = (indexed_key){entry->key.reference, number};
		}
	}
	if (count > 1) {
		qsort(indices, count, sizeof *indices, compare_indices);
	}
	for (size_t index = 0; index < count; index++) {
		adamic_array_push(keys, (adamic_value){.reference = adamic_retain(indices[index].key)});
	}
	free(indices);
	for (size_t index = 0; index < map->used; index++) {
		const adamic_map_entry *entry = &map->entries[index];
		uint32_t number;
		if (!entry->deleted && !entry->hidden && !array_index(entry->key.reference, &number)) {
			adamic_array_push(keys, (adamic_value){.reference = adamic_retain(entry->key.reference)});
		}
	}
	return keys;
}

adamic_record_iterator *adamic_record_iterate(adamic_record *record) {
	adamic_record_iterator *iterator = adamic_object_new(&iteration_shape);
	iterator->slots[0].reference = adamic_retain(record);
	iterator->slots[1].reference = adamic_record_keys(record);
	iterator->slots[2].number = 0;
	return iterator;
}

bool adamic_record_iterator_next(adamic_record_iterator *iterator, adamic_string **key, adamic_value *value) {
	const adamic_record *record = iterator->slots[0].reference;
	const adamic_array *keys = iterator->slots[1].reference;
	size_t next = (size_t)iterator->slots[2].number;
	while (next < keys->length) {
		adamic_string *candidate = keys->elements[next++].reference;
		iterator->slots[2].number = (double)next;
		const adamic_value *slot = adamic_record_get_own(record, candidate);
		if (slot != NULL) {
			*key = candidate;
			*value = *slot;
			return true;
		}
	}
	return false;
}

bool adamic_record_is(const adamic_object *object) {
	return object->shape == &record_number_shape || object->shape == &record_boolean_shape || object->shape == &record_reference_shape;
}

adamic_field_kind adamic_record_values(const adamic_record *record) {
	if (record->shape == &record_number_shape) return adamic_field_number;
	if (record->shape == &record_boolean_shape) return adamic_field_boolean;
	if (record->shape == &record_reference_shape) return adamic_field_reference;
	static const char message[] = "compiler bug: a record of unknown value kind";
	adamic_panic(message, sizeof message - 1);
}

adamic_value adamic_record_view(const adamic_record *record, const char *name, unsigned char wanted, const char *type, const char *expression) {
 adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
 const adamic_value *slot = adamic_record_get_own(record, &key);
 const adamic_heap *value = slot == NULL ? NULL : slot->reference;
 if (wanted == 10 && slot != NULL) return *slot;
 if (value != NULL) {
  if (wanted == 1 && value->kind == adamic_kind_number) return (adamic_value){.number = ((const adamic_number_box *)value)->number};
  if (wanted == 2 && value->kind == adamic_kind_boolean) return (adamic_value){.boolean = ((const adamic_boolean_box *)value)->boolean};
  if (wanted == 3 && value->kind == adamic_kind_string) return *slot;
 }
 const char *actual = value == NULL ? "undefined" : value->kind == adamic_kind_number ? "number" : value->kind == adamic_kind_boolean ? "boolean" : value->kind == adamic_kind_string ? "string" : "object";
 size_t capacity = strlen(expression) + strlen(type) + strlen(actual) + 100;
 char *message = malloc(capacity);
 if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
 (void)snprintf(message, capacity, "field read failed: %s expected %s, found %s", expression, type, actual);
 adamic_panic(message, strlen(message));
}
