// weak.h: what weak.c and weak_collection.c share, private to the two.
//
// A weak collection's entries hold counts on key handles, and each handle names, hidden, the
// collections holding an entry for it, so freeing a key (weak.c's forget) finds every entry to remove
// without a walk over every collection. The names and the entries agree whenever weak.c's table lock
// is free: every change to either is made under it.

#ifndef ADAMIC_WEAK_H
#define ADAMIC_WEAK_H

#include "adamic.h"

#include <stdint.h>

// ADAMIC_WEAK_HIDDEN is the mask: its low bits are clear, so a hidden heap address keeps its
// alignment, and 0 (empty) and 1 (a tombstone) are never one.
#define ADAMIC_WEAK_HIDDEN ((uintptr_t)0xa5a5a5a5a5a5a5a0u)

static inline uintptr_t adamic_weak_hide(const void *pointer) {
	return (uintptr_t)pointer ^ ADAMIC_WEAK_HIDDEN;
}

static inline void *adamic_weak_reveal(uintptr_t hidden) {
	return (void *)(hidden ^ ADAMIC_WEAK_HIDDEN);
}

// adamic_weak_mix spreads an address over a table's slots, the side table's and a collection's alike.
static inline size_t adamic_weak_mix(const void *address, size_t capacity) {
#ifdef ADAMIC_TARGET_WASI
	// Widen before the 64-bit hash shifts; wasm32 pointers have only 32 bits.
	uint64_t key = (uintptr_t)address;
#else
	uintptr_t key = (uintptr_t)address;
#endif
	key ^= key >> 33;
	key *= 0xff51afd7ed558ccdu;
	key ^= key >> 33;
	return (size_t)key & (capacity - 1);
}

struct adamic_weak {
	adamic_heap heap;
	// target is the target's address, hidden, or 0 once it's freed.
	uintptr_t target;
	// collections names, hidden, each collection holding an entry for this handle, once each, so the
	// leak check counts nothing alive through them. Read and written only under the table lock.
	uintptr_t *collections;
	size_t collection_count;
	size_t collection_capacity;
	// region says the target is a statement region's value, counted in weak.c's region_handles until
	// the target or the handle goes.
	bool region;
};

typedef struct adamic_weak_map_entry {
	// key is the entry's handle, one count the entry's, or NULL for a slot never used, or
	// ADAMIC_WEAK_MAP_TOMBSTONE for one whose entry was removed, so a probe goes past it.
	adamic_weak *key;
	adamic_value value;
} adamic_weak_map_entry;
#define ADAMIC_WEAK_MAP_TOMBSTONE ((adamic_weak *)(uintptr_t)1)

struct adamic_weak_map {
	adamic_heap heap;
	// set is a WeakSet, which keeps no values.
	bool set;
	bool reference_values;
	// count is the entries, used the slots holding an entry or a tombstone, capacity the slots.
	size_t count;
	size_t used;
	size_t capacity;
	adamic_weak_map_entry *entries;
};

// weak.c: the table lock, which guards the side table, every handle's target and collections, and the
// entries of every collection a handle names.
void adamic_weak_lock(void);
void adamic_weak_unlock(void);
// Under the lock: the map starts or stops holding an entry for the handle.
void adamic_weak_name(adamic_weak *handle, adamic_weak_map *map);
void adamic_weak_unname(adamic_weak *handle, const adamic_weak_map *map);

// weak_collection.c, under the lock: the map's entry for the handle, or NULL, and its removal, which
// lets go of nothing: the caller lets go of the entry's value and handle count once the lock is free.
adamic_weak_map_entry *adamic_weak_map_entry_of(adamic_weak_map *map, const adamic_weak *key);
void adamic_weak_map_erase(adamic_weak_map *map, adamic_weak_map_entry *entry);

#endif
