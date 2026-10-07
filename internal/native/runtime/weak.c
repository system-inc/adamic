// weak.c: Weak<Target>, a reference that doesn't count (docs/memory.md).
//
// A weak slot holds a counted handle, never the target: the handle is shared by every weak slot that
// points at one target, and it points at the target until the target is freed, then at nothing. A
// side table from each target to its handle is how freeing finds the handle; it's consulted only
// while any handle exists, so a program with no weak reference pays nothing for the feature.
//
// The table's pointers to handles and each handle's pointer to its target are kept hidden (xored with
// a mask), so the leak check, which counts anything a global can reach as alive, still sees a leaked
// handle or target: only a real reference keeps one reachable.

#include "adamic.h"

#include <stdint.h>
#include <stdlib.h>

// HIDDEN is the mask: its low bits are clear, so a hidden heap address keeps its alignment, and 0
// (empty) and 1 (a tombstone) are never one.
#define HIDDEN ((uintptr_t)0xa5a5a5a5a5a5a5a0u)

struct adamic_weak {
	adamic_heap heap;
	// target is the target's address, hidden, or 0 once it's freed.
	uintptr_t target;
};

static uintptr_t hide(const void *pointer) {
	return (uintptr_t)pointer ^ HIDDEN;
}

static void *reveal(uintptr_t hidden) {
	return (void *)(hidden ^ HIDDEN);
}

// The side table: open addressing on the target's address, holding hidden handles, a removed entry
// left as a tombstone so a probe goes past it.
static uintptr_t *table;
static size_t table_capacity;
static size_t table_count;
static size_t table_used;
#define EMPTY ((uintptr_t)0)
#define TOMBSTONE ((uintptr_t)1)

static size_t slot_of(const void *target) {
#ifdef ADAMIC_TARGET_WASI
	// Widen before the 64-bit hash shifts; wasm32 pointers have only 32 bits.
	uint64_t key = (uintptr_t)target;
#else
	uintptr_t key = (uintptr_t)target;
#endif
	key ^= key >> 33;
	key *= 0xff51afd7ed558ccdu;
	key ^= key >> 33;
	return (size_t)key & (table_capacity - 1);
}

static uintptr_t *find(const void *target) {
	if (table_capacity == 0) {
		return NULL;
	}
	for (size_t slot = slot_of(target);; slot = (slot + 1) & (table_capacity - 1)) {
		uintptr_t entry = table[slot];
		if (entry == EMPTY) {
			return NULL;
		}
		if (entry != TOMBSTONE && ((adamic_weak *)reveal(entry))->target == hide(target)) {
			return &table[slot];
		}
	}
}

static void insert(adamic_weak *handle) {
	if ((table_used + 1) * 2 > table_capacity) {
		// Grown (or rebuilt without tombstones) to stay at most half full.
		size_t capacity = table_capacity == 0 ? 16 : (table_count + 1) * 4 > table_capacity ? table_capacity * 2 : table_capacity;
		uintptr_t *old = table;
		size_t old_capacity = table_capacity;
		table = calloc(capacity, sizeof *table);
		if (table == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		table_capacity = capacity;
		table_used = 0;
		for (size_t slot = 0; slot < old_capacity; slot++) {
			if (old[slot] != EMPTY && old[slot] != TOMBSTONE) {
				size_t into = slot_of(reveal(((adamic_weak *)reveal(old[slot]))->target));
				while (table[into] != EMPTY) {
					into = (into + 1) & (table_capacity - 1);
				}
				table[into] = old[slot];
				table_used++;
			}
		}
		free(old);
	}
	size_t slot = slot_of(reveal(handle->target));
	while (table[slot] != EMPTY && table[slot] != TOMBSTONE) {
		slot = (slot + 1) & (table_capacity - 1);
	}
	if (table[slot] == EMPTY) {
		table_used++;
	}
	table[slot] = hide(handle);
	table_count++;
}

static void remove_entry(uintptr_t *entry) {
	*entry = TOMBSTONE;
	table_count--;
	if (table_count == 0) {
		// Nothing weak is left: the table goes, and freeing stops asking it.
		free(table);
		table = NULL;
		table_capacity = 0;
		table_used = 0;
	}
}

adamic_weak *adamic_weak_of(void *target) {
	if (target == NULL) {
		return NULL;
	}
	uintptr_t *entry = find(target);
	if (entry != NULL) {
		return adamic_retain(reveal(*entry));
	}
	adamic_weak *handle = adamic_allocate(sizeof *handle, adamic_kind_weak);
	handle->target = hide(target);
	insert(handle);
	return handle;
}

void *adamic_weak_target(const adamic_weak *handle) {
	return handle == NULL || handle->target == 0 ? NULL : reveal(handle->target);
}

void *adamic_weak_target_present(const adamic_weak *handle) {
	void *target = adamic_weak_target(handle);
	if (target == NULL) {
		static const char message[] = "a weak reference was read after what it pointed to was freed";
		adamic_panic(message, sizeof message - 1);
	}
	return target;
}

void adamic_weak_forget(void *target) {
	if (table_count == 0) {
		return;
	}
	uintptr_t *entry = find(target);
	if (entry != NULL) {
		((adamic_weak *)reveal(*entry))->target = 0;
		remove_entry(entry);
	}
}

void adamic_weak_dropped(adamic_weak *handle) {
	if (handle->target != 0) {
		remove_entry(find(reveal(handle->target)));
	}
}

bool adamic_weak_held(const void *target) {
	// A program with no Weak has no table, and pays one comparison.
	return table_count > 0 && find(target) != NULL;
}
