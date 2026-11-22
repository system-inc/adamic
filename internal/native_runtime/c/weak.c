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
//
// A handle is also a WeakMap's or WeakSet's key (weak_collection.c, weak.h): it names each collection
// holding an entry for it, and when its target is freed, forget removes those entries too. A value a
// handle was made for carries ADAMIC_WEAK_NAMED in its header, so freeing any other unshared value
// skips the table and its lock (heap.c's free_one).

#include "weak.h"

#include <stdint.h>
#include <stdlib.h>
#include <pthread.h>

static uintptr_t hide(const void *pointer) {
	return adamic_weak_hide(pointer);
}

static void *reveal(uintptr_t hidden) {
	return adamic_weak_reveal(hidden);
}

// The side table: open addressing on the target's address, holding hidden handles, a removed entry
// left as a tombstone so a probe goes past it.
static uintptr_t *table;
static size_t table_capacity;
static _Atomic size_t table_count;
static pthread_mutex_t table_lock = PTHREAD_MUTEX_INITIALIZER;
static size_t table_used;
#define EMPTY ((uintptr_t)0)
#define TOMBSTONE ((uintptr_t)1)

// region_handles is how many handles point at a statement region's value, which only adamic_weak_of
// makes: a region's end asks the table only while there's one (adamic_weak_forget_region). It's the
// process's, not a thread's, since a region can end on a thread other than the one that made a handle
// to its value (region_end_test.go's ThreadSanitizer pin). Changed under the lock, read without it.
static _Atomic size_t region_handles;

// releasing is what forget lets go of once the lock is free: the values and handle counts of the
// entries it removed. A release never runs under the lock, since a handle's last one takes it
// (adamic_weak_dropped). Inside the heap's drain a release only queues the value (heap.c's
// release_last), so forget from free_one frees nothing itself.
static _Thread_local void **releasing;
static _Thread_local size_t releasing_count;
static _Thread_local size_t releasing_capacity;

static size_t slot_of(const void *target) {
	return adamic_weak_mix(target, table_capacity);
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

// markable says whether a value's header can carry the named bit: writable heap storage that isn't
// shared. A count-zero value with no flag can be a constant in read-only static storage, where a write
// would fault, and a shared one's header another thread may read; neither is ever marked. A graph or
// Program member counts zero while only its region holds it, and its storage is the heap's.
static bool markable(const adamic_heap *heap) {
	if ((heap->slab & ADAMIC_SHARED_HEADER) != 0) { return false; }
	return heap->references != 0 || (heap->slab & (ADAMIC_GRAPH_FLAG | ADAMIC_PROGRAM_FLAG)) != 0;
}

// region_value says whether the target is a statement region's value: one of its own, or a cell inside
// a region's environment, which has no flag of its own and counts on its environment.
static bool region_value(const adamic_heap *heap) {
#ifdef ADAMIC_CANONICAL_CLOSURES
	if (heap->kind == adamic_kind_cell && ((const adamic_cell *)heap)->owner != NULL) {
		heap = ((const adamic_cell *)heap)->owner;
	}
#endif
	return (heap->slab & ADAMIC_REGION_VALUE) != 0;
}

// handle_of is the target's handle, retained for the caller, made if it has none. Under the lock.
static adamic_weak *handle_of(void *target) {
	adamic_heap *heap = target;
	if (markable(heap) && (heap->slab & ADAMIC_WEAK_NAMED) == 0) {
		heap->slab |= ADAMIC_WEAK_NAMED;
	}
	uintptr_t *entry = find(target);
	if (entry != NULL) {
		return adamic_retain(reveal(*entry));
	}
	adamic_weak *handle = adamic_allocate(sizeof *handle, adamic_kind_weak);
	handle->target = hide(target);
	handle->collections = NULL;
	handle->collection_count = 0;
	handle->collection_capacity = 0;
	handle->region = region_value(heap);
	if (handle->region) {
		region_handles++;
	}
	insert(handle);
	return handle;
}

adamic_weak *adamic_weak_of(void *target) {
	if (target == NULL) {
		return NULL;
	}
	pthread_mutex_lock(&table_lock);
	adamic_weak *handle = handle_of(target);
	pthread_mutex_unlock(&table_lock);
	return handle;
}

adamic_weak *adamic_weak_key(void *key) {
	if (key == NULL) {
		return NULL;
	}
	const adamic_heap *heap = key;
	// A region's value goes with its region, whose end walks no collection, and a shared key could be
	// freed on another thread while its collection's thread reads the entries.
	if ((heap->slab & ADAMIC_REGION_VALUE) != 0) {
		static const char message[] = "a statement region's value can't be a WeakMap or WeakSet key";
		adamic_panic(message, sizeof message - 1);
	}
	if ((heap->slab & ADAMIC_SHARED_HEADER) != 0) {
		static const char message[] = "a value shared with parallel work can't be a WeakMap or WeakSet key";
		adamic_panic(message, sizeof message - 1);
	}
	return adamic_weak_of(key);
}

adamic_weak *adamic_weak_key_find(const void *key) {
	if (key == NULL) {
		return NULL;
	}
	// A value adamic_weak_key would have marked and didn't has no handle, and no entry anywhere: the
	// common miss costs a header read. Anything it can't mark, a constant above all, asks the table.
	const adamic_heap *heap = key;
	if ((heap->slab & ADAMIC_WEAK_NAMED) == 0 && markable(heap)) {
		return NULL;
	}
	pthread_mutex_lock(&table_lock);
	uintptr_t *entry = find(key);
	adamic_weak *handle = entry == NULL ? NULL : reveal(*entry);
	pthread_mutex_unlock(&table_lock);
	return handle;
}

void adamic_weak_lock(void) {
	pthread_mutex_lock(&table_lock);
}

void adamic_weak_unlock(void) {
	pthread_mutex_unlock(&table_lock);
}

void adamic_weak_name(adamic_weak *handle, adamic_weak_map *map) {
	if (handle->collection_count == handle->collection_capacity) {
		size_t capacity = handle->collection_capacity == 0 ? 2 : handle->collection_capacity * 2;
		uintptr_t *grown = realloc(handle->collections, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		handle->collections = grown;
		handle->collection_capacity = capacity;
	}
	handle->collections[handle->collection_count++] = hide(map);
}

void adamic_weak_unname(adamic_weak *handle, const adamic_weak_map *map) {
	for (size_t index = 0; index < handle->collection_count; index++) {
		if (handle->collections[index] == hide(map)) {
			handle->collections[index] = handle->collections[--handle->collection_count];
			return;
		}
	}
	adamic_invariant("a weak collection's entry had no name on its key's handle");
}

static void release_later(void *value) {
	if (value == NULL) {
		return;
	}
	if (releasing_count == releasing_capacity) {
		size_t capacity = releasing_capacity == 0 ? 16 : releasing_capacity * 2;
		void **grown = realloc(releasing, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		releasing = grown;
		releasing_capacity = capacity;
	}
	releasing[releasing_count++] = value;
}

// release_pending lets go of what forget listed, the lock free. A release that frees outside the
// heap's drain can free another key, whose forget lists more and lets go of the whole list itself, so
// each value is taken off before it's let go of and none is let go of twice.
static void release_pending(void) {
	while (releasing_count > 0) {
		adamic_release(releasing[--releasing_count]);
	}
}

// forget_collections removes every entry naming the handle, listing each value and handle count to
// let go of, and drops the names. Under the lock.
static void forget_collections(adamic_weak *handle) {
	for (size_t index = 0; index < handle->collection_count; index++) {
		adamic_weak_map *map = reveal(handle->collections[index]);
		adamic_weak_map_entry *entry = adamic_weak_map_entry_of(map, handle);
		if (entry == NULL) {
			adamic_invariant("a weak handle named a collection without its entry");
		}
		if (map->reference_values) {
			release_later(entry->value.reference);
		}
		release_later(handle);
		adamic_weak_map_erase(map, entry);
	}
	free(handle->collections);
	handle->collections = NULL;
	handle->collection_count = 0;
	handle->collection_capacity = 0;
}

void *adamic_weak_target(const adamic_weak *handle) {
	pthread_mutex_lock(&table_lock);
	void *target = handle == NULL || handle->target == 0 ? NULL : reveal(handle->target);
	pthread_mutex_unlock(&table_lock);
	return target;
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
	pthread_mutex_lock(&table_lock);
	uintptr_t *entry = find(target);
	if (entry != NULL) {
		adamic_weak *handle = reveal(*entry);
		handle->target = 0;
		remove_entry(entry);
		forget_collections(handle);
	}
	pthread_mutex_unlock(&table_lock);
	release_pending();
}

void adamic_weak_forget_region(const adamic_region *region) {
	// Only adamic_weak_of makes a handle to a region's value (adamic_weak_key refuses one), so with
	// none on this thread, no handle points into the region, and no collection names its values.
	if (atomic_load_explicit(&region_handles, memory_order_relaxed) == 0 || table_count == 0) {
		return;
	}
	// The table and every handle target are shared, including unrelated targets.
	// Hold the same lock as readers, insertion, and removal for the entire scan.
	pthread_mutex_lock(&table_lock);
	for (size_t slot = 0; slot < table_capacity; slot++) {
		uintptr_t entry = table[slot];
		if (entry == EMPTY || entry == TOMBSTONE) {
			continue;
		}
		adamic_weak *handle = reveal(entry);
		if (adamic_region_contains(region, reveal(handle->target))) {
			handle->target = 0;
			if (handle->region) {
				handle->region = false;
				region_handles--;
			}
			remove_entry(&table[slot]);
			// Removing the last entry frees the table and resets capacity to zero; the loop stops.
		}
	}
	pthread_mutex_unlock(&table_lock);
}

void adamic_weak_dropped(adamic_weak *handle) {
	pthread_mutex_lock(&table_lock);
	if (handle->target != 0) {
		remove_entry(find(reveal(handle->target)));
		if (handle->region) {
			handle->region = false;
			region_handles--;
		}
	}
	// Every name is an entry holding a count on the handle, so a dropped one names nothing.
	free(handle->collections);
	pthread_mutex_unlock(&table_lock);
}

bool adamic_weak_held(const void *target) {
	// A program with no Weak has no table, and pays one comparison.
	if (atomic_load_explicit(&table_count, memory_order_relaxed) == 0) { return false; }
	pthread_mutex_lock(&table_lock);
	bool held = find(target) != NULL;
	pthread_mutex_unlock(&table_lock);
	return held;
}

bool adamic_weak_named(const void *target) {
	if (atomic_load_explicit(&table_count, memory_order_relaxed) == 0) { return false; }
	pthread_mutex_lock(&table_lock);
	uintptr_t *entry = find(target);
	bool named = entry != NULL && ((adamic_weak *)reveal(*entry))->collection_count != 0;
	pthread_mutex_unlock(&table_lock);
	return named;
}

void adamic_weak_thread_end(void) {
	free(releasing);
	releasing = NULL;
	releasing_capacity = 0;
}
