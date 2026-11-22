// weak_collection.c: WeakMap and WeakSet, collections that don't keep their keys alive (adamic.h).
//
// An entry is keyed by the key's weak handle (weak.c), never by the key's address, and holds a count on
// the handle, so the handle's address can't be reused while the entry lives, and a new value made at a
// freed key's address gets a new handle and finds nothing here. The table is open addressing on the
// handle's address, a removed entry left as a tombstone so a probe goes past it, as the side table is.
//
// When a key is freed, weak.c's forget removes its entries from every collection its handle names and
// lets go of their values and handle counts. When a collection is freed, it takes its name off each
// key's handle first, so no handle ever names a freed collection. Both happen under weak.c's table
// lock, and every change here to the entries or the names is made under it too, so forget always finds
// the entries the names say are there. Nothing is let go of under the lock.
//
// A collection is never shared (share.c refuses it, and adamic_weak_key refuses a shared key), so its
// entries are read without the lock: only its own thread reaches it, and only its own thread can free
// one of its keys.

#include "weak.h"

#include <stdint.h>
#include <stdlib.h>

#ifdef ADAMIC_COUNT
// The entries every collection holds, and the most there ever were at once, written as their own line
// at exit (count.c).
static void count_entries(size_t added, size_t removed) {
	if (removed != 0) {
		atomic_fetch_sub_explicit(&adamic_counted.weak_entries, removed, memory_order_relaxed);
	}
	if (added == 0) {
		return;
	}
	size_t live = atomic_fetch_add_explicit(&adamic_counted.weak_entries, added, memory_order_relaxed) + added;
	size_t peak = atomic_load_explicit(&adamic_counted.weak_entries_peak, memory_order_relaxed);
	while (peak < live && !atomic_compare_exchange_weak_explicit(&adamic_counted.weak_entries_peak, &peak, live, memory_order_relaxed, memory_order_relaxed)) {}
}

size_t adamic_weak_entries_live(void) {
	return atomic_load_explicit(&adamic_counted.weak_entries, memory_order_relaxed);
}
#define COUNT_ENTRIES(added, removed) count_entries((added), (removed))
#else
#define COUNT_ENTRIES(added, removed) ((void)0)
#endif

adamic_weak_map *adamic_weak_map_new(bool set, bool reference_values) {
	adamic_weak_map *map = adamic_allocate(sizeof *map, adamic_kind_weak_map);
	map->set = set;
	// A set has no values, so none is counted.
	map->reference_values = reference_values && !set;
	map->count = 0;
	map->used = 0;
	map->capacity = 0;
	map->entries = NULL;
	return map;
}

adamic_weak_map_entry *adamic_weak_map_entry_of(adamic_weak_map *map, const adamic_weak *key) {
	if (map->capacity == 0 || key == NULL) {
		return NULL;
	}
	for (size_t slot = adamic_weak_mix(key, map->capacity);; slot = (slot + 1) & (map->capacity - 1)) {
		adamic_weak *each = map->entries[slot].key;
		if (each == NULL) {
			return NULL;
		}
		if (each == key) {
			return &map->entries[slot];
		}
	}
}

void adamic_weak_map_erase(adamic_weak_map *map, adamic_weak_map_entry *entry) {
	entry->key = ADAMIC_WEAK_MAP_TOMBSTONE;
	entry->value = (adamic_value){0};
	map->count--;
	COUNT_ENTRIES(0, 1);
}

// insert puts a new entry in, grown (or rebuilt without tombstones) to stay at most half full. Under
// the lock, since forget may read the table of any collection a handle names.
static void insert(adamic_weak_map *map, adamic_weak *key, adamic_value value) {
	if ((map->used + 1) * 2 > map->capacity) {
		size_t capacity = map->capacity == 0 ? 8 : (map->count + 1) * 4 > map->capacity ? map->capacity * 2 : map->capacity;
		adamic_weak_map_entry *old = map->entries;
		size_t old_capacity = map->capacity;
		map->entries = calloc(capacity, sizeof *map->entries);
		if (map->entries == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		map->capacity = capacity;
		map->used = 0;
		for (size_t slot = 0; slot < old_capacity; slot++) {
			adamic_weak *each = old[slot].key;
			if (each != NULL && each != ADAMIC_WEAK_MAP_TOMBSTONE) {
				size_t into = adamic_weak_mix(each, map->capacity);
				while (map->entries[into].key != NULL) {
					into = (into + 1) & (map->capacity - 1);
				}
				map->entries[into] = old[slot];
				map->used++;
			}
		}
		free(old);
	}
	size_t slot = adamic_weak_mix(key, map->capacity);
	while (map->entries[slot].key != NULL && map->entries[slot].key != ADAMIC_WEAK_MAP_TOMBSTONE) {
		slot = (slot + 1) & (map->capacity - 1);
	}
	if (map->entries[slot].key == NULL) {
		map->used++;
	}
	map->entries[slot] = (adamic_weak_map_entry){key, value};
	map->count++;
	COUNT_ENTRIES(1, 0);
}

bool adamic_weak_map_lookup(adamic_weak_map *map, const adamic_weak *key, adamic_value *value) {
	adamic_weak_map_entry *entry = adamic_weak_map_entry_of(map, key);
	if (entry == NULL) {
		return false;
	}
	if (value != NULL) {
		*value = entry->value;
	}
	return true;
}

void adamic_weak_map_store(adamic_weak_map *map, adamic_weak *key, adamic_value value) {
	if (key == NULL) {
		adamic_invariant("a weak collection's key had no handle");
	}
	if (map->set) {
		value = (adamic_value){0};
	}
	// A region's value is freed with its region whatever counts it, so a collection that outlived the
	// statement would hold freed memory.
	if (map->reference_values && value.reference != NULL && (((adamic_heap *)value.reference)->slab & ADAMIC_REGION_VALUE) != 0) {
		static const char message[] = "a WeakMap can't hold a statement region's value";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_weak_map_entry *entry = adamic_weak_map_entry_of(map, key);
	if (entry != NULL) {
		// The entry keeps its handle count, so the caller's is the extra one; the old value goes.
		adamic_value old = entry->value;
		entry->value = value;
		if (map->reference_values) {
			adamic_release(old.reference);
		}
		adamic_release(key);
		return;
	}
	adamic_weak_lock();
	// A handle whose target is already gone names nothing forget would ever remove: storing under it
	// would hold the value until the collection goes.
	bool gone = key->target == 0;
	if (!gone) {
		adamic_weak_name(key, map);
		insert(map, key, value);
	}
	adamic_weak_unlock();
	if (gone) {
		if (map->reference_values) {
			adamic_release(value.reference);
		}
		adamic_release(key);
	}
}

bool adamic_weak_map_remove(adamic_weak_map *map, const adamic_weak *key) {
	adamic_weak_map_entry *entry = adamic_weak_map_entry_of(map, key);
	if (entry == NULL) {
		return false;
	}
	adamic_weak *handle = entry->key;
	adamic_value value = entry->value;
	adamic_weak_lock();
	adamic_weak_unname(handle, map);
	adamic_weak_map_erase(map, entry);
	adamic_weak_unlock();
	if (map->reference_values) {
		adamic_release(value.reference);
	}
	adamic_release(handle);
	return true;
}

void adamic_weak_map_free_children(adamic_weak_map *map, void (*let_go)(void *)) {
	// The names go first, under the lock, so a key freed after this never reaches the freed
	// collection; then, the lock free, every value and handle count, read from the entries still here.
	adamic_weak_lock();
	for (size_t slot = 0; slot < map->capacity; slot++) {
		adamic_weak *each = map->entries[slot].key;
		if (each != NULL && each != ADAMIC_WEAK_MAP_TOMBSTONE) {
			adamic_weak_unname(each, map);
		}
	}
	adamic_weak_unlock();
	for (size_t slot = 0; slot < map->capacity; slot++) {
		adamic_weak *each = map->entries[slot].key;
		if (each != NULL && each != ADAMIC_WEAK_MAP_TOMBSTONE) {
			if (map->reference_values) {
				let_go(map->entries[slot].value.reference);
			}
			let_go(each);
		}
	}
	COUNT_ENTRIES(0, map->count);
	free(map->entries);
	map->entries = NULL;
	map->count = 0;
	map->used = 0;
	map->capacity = 0;
}
