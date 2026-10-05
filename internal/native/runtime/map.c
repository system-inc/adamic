// map.c: Map, as JavaScript has it. Entries keep insertion order in one array, and a hash index
// finds them. Keys compare with SameValueZero: NaN equals NaN, and +0 equals -0. Deleting leaves a
// tombstone, so an iteration in progress keeps its place; entries added during an iteration are
// visited, and deleted ones aren't, as ECMA-262 requires.

#include "adamic.h"

#include <math.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

static void *allocate_zeroed(size_t count, size_t size) {
	void *memory = calloc(count, size);
	if (memory == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return memory;
}

adamic_map *adamic_map_new(bool string_keys, bool reference_values) {
	adamic_map *map = adamic_allocate(sizeof *map, adamic_kind_map);
	map->count = 0;
	map->used = 0;
	map->capacity = 0;
	map->entries = NULL;
	map->bucket_count = 0;
	map->buckets = NULL;
	map->string_keys = string_keys;
	map->reference_values = reference_values;
	return map;
}

static uint64_t hash_key(const adamic_map *map, adamic_value key) {
	uint64_t hash = 14695981039346656037ull;
	if (map->string_keys) {
		const adamic_string *string = key.reference;
		for (size_t index = 0; index < string->length; index++) {
			hash = (hash ^ (unsigned char)string->bytes[index]) * 1099511628211ull;
		}
		return hash;
	}
	double number = key.number;
	if (number == 0) {
		number = 0; // -0 and +0 are one key
	}
	if (isnan(number)) {
		return 0x7ff8000000000000ull; // every NaN is one key
	}
	uint64_t bits;
	memcpy(&bits, &number, sizeof bits);
	return (bits ^ (bits >> 29)) * 1099511628211ull;
}

static bool same_key(const adamic_map *map, adamic_value left, adamic_value right) {
	if (map->string_keys) {
		return adamic_string_equal(left.reference, right.reference);
	}
	return left.number == right.number || (isnan(left.number) && isnan(right.number));
}

// find is the entry index for a key, or SIZE_MAX.
static size_t find(const adamic_map *map, adamic_value key) {
	if (map->bucket_count == 0) {
		return SIZE_MAX;
	}
	size_t mask = map->bucket_count - 1;
	for (size_t bucket = hash_key(map, key) & mask;; bucket = (bucket + 1) & mask) {
		size_t slot = map->buckets[bucket];
		if (slot == 0) {
			return SIZE_MAX;
		}
		const adamic_map_entry *entry = &map->entries[slot - 1];
		if (!entry->deleted && same_key(map, entry->key, key)) {
			return slot - 1;
		}
	}
}

// rebuild compacts away tombstones and rehashes into enough buckets for the live entries to grow.
static void rebuild(adamic_map *map, size_t entries_needed) {
	size_t live = 0;
	for (size_t index = 0; index < map->used; index++) {
		if (!map->entries[index].deleted) {
			map->entries[live++] = map->entries[index];
		}
	}
	map->used = live;
	if (entries_needed > map->capacity) {
		size_t capacity = map->capacity == 0 ? 8 : map->capacity;
		while (capacity < entries_needed) {
			capacity *= 2;
		}
		adamic_map_entry *grown = realloc(map->entries, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		map->entries = grown;
		map->capacity = capacity;
	}
	free(map->buckets);
	map->bucket_count = 16;
	while (map->bucket_count < map->capacity * 2) {
		map->bucket_count *= 2;
	}
	map->buckets = allocate_zeroed(map->bucket_count, sizeof *map->buckets);
	size_t mask = map->bucket_count - 1;
	for (size_t index = 0; index < map->used; index++) {
		size_t bucket = hash_key(map, map->entries[index].key) & mask;
		while (map->buckets[bucket] != 0) {
			bucket = (bucket + 1) & mask;
		}
		map->buckets[bucket] = index + 1;
	}
}

adamic_value *adamic_map_get(const adamic_map *map, adamic_value key) {
	size_t index = find(map, key);
	return index == SIZE_MAX ? NULL : &map->entries[index].value;
}

void adamic_map_set(adamic_map *map, adamic_value key, adamic_value value) {
	size_t index = find(map, key);
	if (index != SIZE_MAX) {
		// The key it already has stays; the one passed in is let go, and so is the old value.
		if (map->string_keys) {
			adamic_release(key.reference);
		}
		if (map->reference_values) {
			adamic_release(map->entries[index].value.reference);
		}
		map->entries[index].value = value;
		return;
	}
	if (map->used == map->capacity) {
		// Full: grow when most entries are live, and when most are tombstones, compact in place.
		size_t needed = map->count * 2 > map->capacity ? map->capacity * 2 : map->capacity;
		rebuild(map, needed == 0 ? 8 : needed);
	}
	adamic_map_entry *entry = &map->entries[map->used];
	entry->key = key;
	entry->value = value;
	entry->deleted = false;
	size_t mask = map->bucket_count - 1;
	size_t bucket = hash_key(map, key) & mask;
	while (map->buckets[bucket] != 0) {
		bucket = (bucket + 1) & mask;
	}
	map->buckets[bucket] = map->used + 1;
	map->used++;
	map->count++;
}

bool adamic_map_delete(adamic_map *map, adamic_value key) {
	size_t index = find(map, key);
	if (index == SIZE_MAX) {
		return false;
	}
	adamic_map_entry *entry = &map->entries[index];
	entry->deleted = true;
	if (map->string_keys) {
		adamic_release(entry->key.reference);
	}
	if (map->reference_values) {
		adamic_release(entry->value.reference);
	}
	map->count--;
	return true;
}

void adamic_map_free_children(adamic_map *map, void (*let_go)(void *)) {
	for (size_t index = 0; index < map->used; index++) {
		if (map->entries[index].deleted) {
			continue;
		}
		if (map->string_keys) {
			let_go(map->entries[index].key.reference);
		}
		if (map->reference_values) {
			let_go(map->entries[index].value.reference);
		}
	}
	free(map->entries);
	free(map->buckets);
}

adamic_array *adamic_map_entries(const adamic_map *map, const adamic_shape *pair) {
	adamic_array *entries = adamic_array_new(map->count, true);
	for (size_t index = 0; index < map->used; index++) {
		const adamic_map_entry *entry = &map->entries[index];
		if (entry->deleted) {
			continue;
		}
		// [key, value], an object whose fields are named "0" and "1".
		adamic_object *tuple = adamic_object_new(pair);
		tuple->slots[0] = entry->key;
		tuple->slots[1] = entry->value;
		if (map->string_keys) {
			adamic_retain(entry->key.reference);
		}
		if (map->reference_values) {
			adamic_retain(entry->value.reference);
		}
		adamic_array_push(entries, (adamic_value){.reference = tuple});
	}
	return entries;
}
