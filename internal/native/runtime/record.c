// record.c: own enumerable string properties, using Map's counted ordered hash table.
#include "adamic.h"

#include <stdlib.h>
#include <string.h>

// The iterator alone uses private wrapper slots. Records themselves use the ordinary
// object's declared shape and dictionary pointer, never a second object representation.
static const adamic_shape record_shape = {0, NULL, NULL, NULL};
static const char *const iteration_names[] = {"0", "1", "key"};
static const bool iteration_references[] = {true, true, false};
static const adamic_shape iteration_shape = {3, iteration_names, iteration_references, NULL};

static adamic_map *table(const adamic_record *record) {
	return record->dictionary;
}

adamic_record *adamic_record_new_shaped(const adamic_shape *shape, bool reference_values) {
	adamic_record *record = adamic_object_new(shape);
	record->dictionary = adamic_map_new(true, reference_values);
	return record;
}

adamic_record *adamic_record_new(bool reference_values) {
	return adamic_record_new_shaped(&record_shape, reference_values);
}

// A spread makes a new object, retaining both stores and their common key order.
adamic_record *adamic_record_copy(const adamic_record *source) {
	adamic_record *copy = adamic_record_new_shaped(source->shape, table(source)->reference_values);
	for (size_t index = 0; index < source->shape->count; index++) {
		copy->slots[index] = source->slots[index];
		if (source->shape->references[index]) { adamic_retain(copy->slots[index].reference); }
	}
	memcpy(adamic_object_initialized(copy), adamic_object_initialized(source), source->shape->count);
	memcpy(adamic_object_field_types(copy), adamic_object_field_types(source), source->shape->count);
	const adamic_map *map = table(source);
	for (size_t index = 0; index < map->used; index++) {
		const adamic_map_entry *entry = &map->entries[index];
		if (entry->deleted) { continue; }
		adamic_value value = entry->value;
		if (map->reference_values) { adamic_retain(value.reference); }
		adamic_map_set(table(copy), (adamic_value){.reference = adamic_retain(entry->key.reference)}, value);
	}
	return copy;
}

static adamic_value *fixed_slot(const adamic_record *record, const adamic_string *key) {
	for (size_t index = 0; index < record->shape->count; index++) {
		const char *name = record->shape->names[index];
		if (strlen(name) == key->length && memcmp(name, key->bytes, key->length) == 0) {
			return &((adamic_record *)record)->slots[index];
		}
	}
	return NULL;
}

adamic_value *adamic_record_get_own(const adamic_record *record, const adamic_string *key) {
	adamic_value *entry = adamic_map_get(table(record), (adamic_value){.reference = (void *)key});
	if (entry == NULL) { return NULL; }
	adamic_value *fixed = fixed_slot(record, key);
	return fixed == NULL ? entry : fixed;
}

adamic_value *adamic_object_dictionary_field(const adamic_object *object, const char *name) {
	const adamic_string key = {.length = strlen(name), .bytes = name};
	return adamic_record_get_own(object, &key);
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

void adamic_record_define(adamic_record *record, adamic_string *key, adamic_value value) {
	adamic_value *fixed = fixed_slot(record, key);
	if (fixed != NULL) {
		if (table(record)->reference_values) { adamic_release(fixed->reference); }
		*fixed = value;
		// A zero-valued token owns the name and its insertion position, not the field's value.
		value = (adamic_value){.reference = NULL};
	}
	adamic_map_set(table(record), (adamic_value){.reference = key}, value);
}

void adamic_record_set(adamic_record *record, adamic_string *key, adamic_value value) {
	if (named(key, "__proto__")) {
		static const char message[] = "NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_record_define(record, key, value);
}

// Fixed string views use the same table when a named optional field is first written.
adamic_value *adamic_record_write_field(adamic_record *record, const char *name) {
	const adamic_string borrowed = {.length = strlen(name), .bytes = name};
	adamic_value *slot = adamic_record_get_own(record, &borrowed);
	if (slot != NULL) { return slot; }
	adamic_string *key = adamic_string_allocate(borrowed.length);
	memcpy((char *)key->bytes, name, borrowed.length);
	adamic_record_set(record, key, (adamic_value){.reference = NULL});
	return adamic_record_get_own(record, &borrowed);
}

bool adamic_record_delete(adamic_record *record, const adamic_string *key) {
	// Delete clears a fixed value too; reinsertion gets a new table position.
	adamic_value *fixed = fixed_slot(record, key);
	if (fixed != NULL) {
		if (table(record)->reference_values) { adamic_release(fixed->reference); }
		*fixed = (adamic_value){.reference = NULL};
	}
	// JavaScript delete succeeds even when no own property was there.
	(void)adamic_map_delete(table(record), (adamic_value){.reference = (void *)key});
	return true;
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
		if (!map->entries[index].deleted && array_index(map->entries[index].key.reference, &number)) {
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
		if (!entry->deleted && array_index(entry->key.reference, &number)) {
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
		if (!entry->deleted && !array_index(entry->key.reference, &number)) {
			adamic_array_push(keys, (adamic_value){.reference = adamic_retain(entry->key.reference)});
		}
	}
	return keys;
}

// The same ordered keys drive both reflection APIs; a fixed field's token is
// never interpreted as its value. The returned arrays own their values and pairs.
adamic_array *adamic_record_values(const adamic_record *record, bool entries) {
	adamic_array *keys = adamic_record_keys(record);
	bool references = table(record)->reference_values;
	adamic_array *values = adamic_array_new(keys->length, entries || references);
	static const char *const names[] = {"0", "1"};
	static const bool scalar_references[] = {true, false};
	static const bool reference_references[] = {true, true};
	static const adamic_shape scalar_pair = {2, names, scalar_references, NULL};
	static const adamic_shape reference_pair = {2, names, reference_references, NULL};
	for (size_t index = 0; index < keys->length; index++) {
		adamic_string *key = keys->elements[index].reference;
		adamic_value value = *adamic_record_get_own(record, key);
		if (references) { adamic_retain(value.reference); }
		if (entries) {
			adamic_object *pair = adamic_object_new(references ? &reference_pair : &scalar_pair);
			pair->slots[0].reference = adamic_retain(key);
			pair->slots[1] = value;
			value.reference = pair;
		}
		adamic_array_push(values, value);
	}
	adamic_release(keys);
	return values;
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
