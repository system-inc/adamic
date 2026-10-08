// map.c: Map, as JavaScript has it. Up to four entries live inline and are searched linearly.
// Larger maps keep insertion order in one array and a hash index finds their keys.
// Keys compare with SameValueZero: NaN equals NaN, +0 equals -0, and an object, an array,
// a map or a function is its own key, found by identity. Deleting leaves a
// tombstone, and while an iteration is open a full map grows rather than compacting them away, so
// the iteration keeps its place; entries added during it are visited, and deleted ones aren't, as
// ECMA-262 requires.

#include "adamic.h"
#include "graph_regions.h"

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
	map->key_contract = map->value_contract = 0;
	map->contract_name = "uncertified Map";
	map->key_type = map->value_type = 0;
	map->count = 0;
	map->used = 0;
	map->capacity = 4;
	map->entries = map->small;
	map->bucket_count = 0;
	map->buckets = NULL;
	map->intrinsic_set = false;
	map->string_keys = string_keys;
	map->reference_keys = string_keys;
	map->boolean_keys = false;
	map->maybe_number_keys = false;
	map->reference_values = reference_values;
	map->iterating = 0;
	return map;
}

adamic_map *adamic_map_new_booleans(bool reference_values) {
	adamic_map *map = adamic_map_new(false, reference_values);
	map->boolean_keys = true;
	return map;
}

adamic_map *adamic_map_new_maybe_numbers(bool reference_values) {
	adamic_map *map = adamic_map_new(false, reference_values);
	map->maybe_number_keys = true;
	return map;
}

adamic_map *adamic_map_new_identity(bool reference_values) {
	adamic_map *map = adamic_map_new(false, reference_values);
	map->reference_keys = true;
	return map;
}

static uint64_t boxed_key_hash(const adamic_heap *key) {
 if (key == NULL) return ADAMIC_UNDEFINED_BITS;
 if (key->kind == adamic_kind_number) return adamic_map_number_hash(((const adamic_number_box *)key)->number);
 if (key->kind == adamic_kind_boolean) return ((const adamic_boolean_box *)key)->boolean ? 0x9e3779b97f4a7c15ull : 0x7f4a7c159e3779b9ull;
 if (key->kind == adamic_kind_string) {
  const adamic_string *string = (const adamic_string *)key;
  uint64_t hash = 14695981039346656037ull;
  for (size_t index=0;index<string->length;index++) hash=(hash^(unsigned char)string->bytes[index])*1099511628211ull;
  return hash;
 }
 uint64_t bits=(uint64_t)(uintptr_t)key;
 return (bits^(bits>>4)^(bits>>29))*1099511628211ull;
}

static bool boxed_key_equal(const adamic_heap *left, const adamic_heap *right) {
 if (left != NULL && right != NULL && left->kind == adamic_kind_number && right->kind == adamic_kind_number) {
  double a=((const adamic_number_box *)left)->number,b=((const adamic_number_box *)right)->number;
  return a == b || (isnan(a) && isnan(b));
 }
 return adamic_union_equal(left,right);
}

static uint64_t hash_key(const adamic_map *map, adamic_value key) {
 if (map->key_type == 10) return boxed_key_hash(key.reference);
 if (map->key_type == 9) return key.maybe_boolean;
	if (map->maybe_number_keys) {
		return adamic_map_maybe_key_hash(key.number);
	}
	uint64_t hash = 14695981039346656037ull;
	if (map->string_keys) {
		const adamic_string *string = key.reference;
		if (string == NULL) {
			// undefined, which a string | undefined key holds as NULL. It hashes as the empty string
			// does, and same_key tells the two apart: adamic_string_equal holds undefined equal only
			// to itself.
			return hash;
		}
		for (size_t index = 0; index < string->length; index++) {
			hash = (hash ^ (unsigned char)string->bytes[index]) * 1099511628211ull;
		}
		return hash;
	}
	if (map->boolean_keys) {
		return key.boolean ? 0x9e3779b97f4a7c15ull : 0x7f4a7c159e3779b9ull;
	}
	if (map->reference_keys) {
		// By identity: the address, its low bits (alignment, always zero) mixed up into the rest.
		uint64_t bits = (uint64_t)(uintptr_t)key.reference;
		return (bits ^ (bits >> 4) ^ (bits >> 29)) * 1099511628211ull;
	}
	return adamic_map_number_hash(key.number);
}

static bool same_key(const adamic_map *map, adamic_value left, adamic_value right) {
 if (map->key_type == 10) return boxed_key_equal(left.reference,right.reference);
 if (map->key_type == 9) return left.maybe_boolean == right.maybe_boolean;
	if (map->maybe_number_keys) {
		return adamic_map_maybe_key_equal(left.number, right.number);
	}
	if (map->string_keys) {
		return adamic_string_equal(left.reference, right.reference);
	}
	if (map->boolean_keys) {
		return left.boolean == right.boolean;
	}
	if (map->reference_keys) {
		return left.reference == right.reference;
	}
	return left.number == right.number || (isnan(left.number) && isnan(right.number));
}

// find is the entry index for a key, or SIZE_MAX.
static size_t find(const adamic_map *map, adamic_value key) {
	if (map->bucket_count == 0) {
		for (size_t index = 0; index < map->used; index++) {
			const adamic_map_entry *entry = &map->entries[index];
			if (!entry->deleted && same_key(map, entry->key, key)) {
				return index;
			}
		}
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

// Without an active iterator, entry positions can move. Copying transfers ownership:
// the same references remain in the map, without extra retains or releases.
static void compact(adamic_map *map) {
	size_t live = 0;
	for (size_t index = 0; index < map->used; index++) {
		if (!map->entries[index].deleted) {
			map->entries[live++] = map->entries[index];
		}
	}
	map->used = live;
}

static void return_inline(adamic_map *map) {
	if (map->iterating != 0 || map->count > 4) {
		return;
	}
	compact(map);
	if (map->entries != map->small) {
		memcpy(map->small, map->entries, map->used * sizeof *map->entries);
		free(map->entries);
		free(map->buckets);
		map->entries = map->small;
		map->capacity = 4;
		map->buckets = NULL;
		map->bucket_count = 0;
	}
}

// rebuild compacts away tombstones, unless an iteration is open, and rehashes into enough buckets
// for the entries to grow.
static void rebuild(adamic_map *map, size_t entries_needed) {
	if (map->iterating == 0) {
		compact(map);
	}
	if (entries_needed <= 4 && map->entries == map->small) {
		return;
	}
	if (entries_needed > map->capacity) {
		size_t capacity = map->capacity == 0 ? 8 : map->capacity;
		while (capacity < entries_needed) {
			capacity *= 2;
		}
		bool inline_entries = map->entries == map->small;
		adamic_map_entry *grown = inline_entries ? malloc(capacity * sizeof *grown) : realloc(map->entries, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		if (inline_entries) {
			// Promotion preserves historical positions, including tombstones an iterator needs.
			memcpy(grown, map->entries, map->used * sizeof *grown);
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
		if (map->entries[index].deleted) {
			// Kept for an open iteration; its key was let go, and nothing finds it.
			continue;
		}
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
 if (map->key_type == 10 && key.reference != NULL && ((const adamic_heap *)key.reference)->kind == adamic_kind_number) {
  double number=((const adamic_number_box *)key.reference)->number;
  if (number == 0 && signbit(number)) {
   // Do not alter a box shared with the program's original -0 value.
   adamic_graph_drop(map,key.reference);
   key.reference=adamic_graph_take(map,adamic_box_number(0));
  }
 }
	size_t index = find(map, key);
	if (index != SIZE_MAX) {
		// The key it already has stays; the one passed in is let go, and so is the old value.
		if (map->reference_keys) {
			if (adamic_graph_is(map)) { adamic_graph_drop(map, key.reference); } else { adamic_release(key.reference); }
		}
		if (map->reference_values) {
			if (adamic_graph_is(map)) { adamic_graph_drop(map, map->entries[index].value.reference); } else { adamic_release(map->entries[index].value.reference); }
		}
		map->entries[index].value = value;
		return;
	}
	if (map->used == map->capacity) {
		// Full: grow when most entries are live, and when most are tombstones, compact in place, unless
		// an iteration is open, which needs every entry where it is.
		size_t needed;
		if (map->bucket_count == 0) {
			needed = map->iterating == 0 && map->count < 4 ? 4 : 8;
		} else {
			needed = map->iterating > 0 || map->count * 2 > map->capacity ? map->capacity * 2 : map->capacity;
		}
		rebuild(map, needed);
	}
	adamic_map_entry *entry = &map->entries[map->used];
	if (!map->reference_keys && !map->boolean_keys && map->key_type != 9 && key.number == 0) {
		// Map.prototype.set stores -0 as +0 (ECMA-262), so iterating gives back +0: 1 / key is Infinity.
		key.number = 0;
	}
	entry->key = key;
	entry->value = value;
	entry->deleted = false;
	if (map->bucket_count != 0) {
		size_t mask = map->bucket_count - 1;
		size_t bucket = hash_key(map, key) & mask;
		while (map->buckets[bucket] != 0) {
			bucket = (bucket + 1) & mask;
		}
		map->buckets[bucket] = map->used + 1;
	}
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
	if (map->reference_keys) {
		if (adamic_graph_is(map)) { adamic_graph_drop(map, entry->key.reference); } else { adamic_release(entry->key.reference); }
	}
	if (map->reference_values) {
		if (adamic_graph_is(map)) { adamic_graph_drop(map, entry->value.reference); } else { adamic_release(entry->value.reference); }
	}
	map->count--;
	return_inline(map);
	return true;
}

void adamic_map_free_children(adamic_map *map, void (*let_go)(void *)) {
	for (size_t index = 0; index < map->used; index++) {
		if (map->entries[index].deleted) {
			continue;
		}
		if (map->reference_keys) {
			let_go(map->entries[index].key.reference);
		}
		if (map->reference_values) {
			let_go(map->entries[index].value.reference);
		}
	}
	if (map->entries != map->small) {
		free(map->entries);
	}
	free(map->buckets);
}

adamic_map_iterator *adamic_map_iterate(adamic_map *map) {
	adamic_map_iterator *iterator = adamic_allocate(sizeof *iterator, adamic_kind_map_iterator);
	iterator->map = adamic_retain(map);
	iterator->next = 0;
	iterator->exhausted = false;
	map->iterating++;
	if (adamic_graph_is(map)) { iterator = adamic_graph_adopt_owned(iterator, sizeof *iterator); }
	return iterator;
}

bool adamic_map_iterator_next(adamic_map_iterator *iterator, adamic_value *key, adamic_value *value) {
	if (iterator->exhausted) {
		return false;
	}
	// used is read each time, so an entry added since the last step is still ahead.
	while (iterator->next < iterator->map->used) {
		const adamic_map_entry *entry = &iterator->map->entries[iterator->next++];
		if (!entry->deleted) {
			*key = entry->key;
			*value = entry->value;
			return true;
		}
	}
	// A held iterator that reached done no longer needs stable entry positions.
	// exhausted also tells its eventual free not to drop the count a second time.
	adamic_map_iterator_close(iterator);
	return false;
}

void adamic_map_iterator_close(adamic_map_iterator *iterator) {
	if (!iterator->exhausted) {
		iterator->exhausted = true;
		iterator->map->iterating--;
		return_inline(iterator->map);
	}
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
		if (map->reference_keys) {
			adamic_retain(entry->key.reference);
		}
		if (map->reference_values) {
			adamic_retain(entry->value.reference);
		}
		if (adamic_graph_is(map)) { tuple = adamic_graph_adopt_owned(tuple, sizeof *tuple + pair->count * sizeof(adamic_value)); }
		adamic_array_push(entries, (adamic_value){.reference = tuple});
	}
	return entries;
}

void adamic_map_clear(adamic_map *map) {
	for (size_t index = 0; index < map->used; index++) {
		adamic_map_entry *entry = &map->entries[index];
		if (entry->deleted) {
			continue;
		}
		entry->deleted = true;
		if (map->reference_keys) {
			if (adamic_graph_is(map)) { adamic_graph_drop(map, entry->key.reference); } else { adamic_release(entry->key.reference); }
		}
		if (map->reference_values) {
			if (adamic_graph_is(map)) { adamic_graph_drop(map, entry->value.reference); } else { adamic_release(entry->value.reference); }
		}
	}
	map->count = 0;
	return_inline(map);
}

// listed is an array of every live entry's key, or every value, in insertion order, each reference
// retained for the array.
static adamic_array *listed(const adamic_map *map, bool keys) {
	bool references = keys ? map->reference_keys : map->reference_values;
	adamic_array *array = adamic_array_new(map->count, references);
	for (size_t index = 0; index < map->used; index++) {
		const adamic_map_entry *entry = &map->entries[index];
		if (entry->deleted) {
			continue;
		}
		adamic_value value = keys ? entry->key : entry->value;
		if (references) {
			adamic_retain(value.reference);
		}
		adamic_array_push(array, value);
	}
	return array;
}

adamic_array *adamic_map_keys(const adamic_map *map) {
	return listed(map, true);
}

adamic_array *adamic_map_values(const adamic_map *map) {
	return listed(map, false);
}

void adamic_map_add_pairs(adamic_map *map, const adamic_array *pairs) {
	static adamic_slot_cache key_cache, value_cache;
	for (size_t index = 0; index < pairs->length; index++) {
		const adamic_object *pair = pairs->elements[index].reference;
		adamic_value key = *adamic_object_field(pair, "0", &key_cache);
		adamic_value value = *adamic_object_field(pair, "1", &value_cache);
		if (map->reference_keys) {
			adamic_graph_hold(map, key.reference);
		}
		if (map->reference_values) {
			adamic_graph_hold(map, value.reference);
		}
		adamic_map_set(map, key, value);
	}
}
