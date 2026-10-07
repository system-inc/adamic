// share.c: publish only after the complete reachable graph is permanently marked.
#include "adamic.h"
#include <pthread.h>
#include <stdlib.h>

// Serializing graph preparation avoids a second publisher mistaking a partly marked graph for a
// complete one. Retain/release never take this lock.
static pthread_mutex_t sharing = PTHREAD_MUTEX_INITIALIZER;
typedef struct { void **values; size_t count; size_t capacity; void **seen; size_t seen_count; size_t seen_capacity; } pending_values;
static size_t seen_slot(void *value, size_t capacity) {
#ifdef ADAMIC_TARGET_WASI
	// Widen before the 64-bit hash shifts; wasm32 pointers have only 32 bits.
	uint64_t hash = (uintptr_t)value >> 3;
#else
	uintptr_t hash = (uintptr_t)value >> 3;
#endif
	hash ^= hash >> 33; hash *= UINT64_C(0xff51afd7ed558ccd); hash ^= hash >> 33;
	return (size_t)hash & (capacity - 1);
}
static bool visited(pending_values *pending, void *value) {
	if (pending->seen_count * 2 >= pending->seen_capacity) {
		size_t capacity = pending->seen_capacity == 0 ? 64 : pending->seen_capacity * 2;
		void **seen = calloc(capacity, sizeof *seen);
		if (seen == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
		for (size_t index = 0; index < pending->seen_capacity; index++) {
			void *old = pending->seen[index];
			if (old == NULL) { continue; }
			size_t slot = seen_slot(old, capacity);
			while (seen[slot] != NULL) { slot = (slot + 1) & (capacity - 1); }
			seen[slot] = old;
		}
		free(pending->seen); pending->seen = seen; pending->seen_capacity = capacity;
	}
	size_t slot = seen_slot(value, pending->seen_capacity);
	while (pending->seen[slot] != NULL) {
		if (pending->seen[slot] == value) { return true; }
		slot = (slot + 1) & (pending->seen_capacity - 1);
	}
	pending->seen[slot] = value; pending->seen_count++;
	return false;
}
static void append(pending_values *pending, void *value) {
	if (value == NULL) { return; }
	if (pending->count == pending->capacity) {
		size_t capacity = pending->capacity == 0 ? 64 : pending->capacity * 2;
		void **values = realloc(pending->values, capacity * sizeof *values);
		if (values == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
		pending->values = values;
		pending->capacity = capacity;
	}
	pending->values[pending->count++] = value;
}

void adamic_share(void *value) {
	if (value == NULL) { return; }
	pthread_mutex_lock(&sharing);
	pending_values pending = {0};
	append(&pending, value);
	while (pending.count != 0) {
		adamic_heap *heap = pending.values[--pending.count];
		bool shared = adamic_is_shared(heap);
		// Immutable shared leaves cannot gain children. Containers can have gained fresh descendants
		// through unique reuse before the caller retained them again, so always revisit containers.
		if (shared && (heap->kind == adamic_kind_string || heap->kind == adamic_kind_number || heap->kind == adamic_kind_boolean)) { continue; }
		if (visited(&pending, heap)) { continue; }
		// Static scalar/closure identities need no marking. Region objects also have a zero count,
		// but their children must be visited and their storage remains alive through the join.
		if (adamic_reference_count(heap) == 0 && (heap->slab & ADAMIC_REGION_VALUE) == 0) { continue; }
		if (!shared) {
			__atomic_fetch_or(&heap->references, ADAMIC_SHARED, __ATOMIC_RELAXED);
			heap->slab |= ADAMIC_SHARED_HEADER;
		}
		switch (heap->kind) {
		case adamic_kind_async_frame: case adamic_kind_async_promise: case adamic_kind_async_reaction:
			adamic_panic("async values cannot cross worker threads", 39);
		case adamic_kind_string: append(&pending, ((adamic_string *)heap)->owner); break;
		case adamic_kind_object: {
			adamic_object *object = (adamic_object *)heap;
			for (size_t i = 0; i < object->shape->count; i++) {
				if (object->shape->references[i]) { append(&pending, object->slots[i].reference); }
			}
			break;
		}
		case adamic_kind_array: {
			adamic_array *array = (adamic_array *)heap;
			if (array->references) {
				for (size_t i = 0; i < array->length; i++) { append(&pending, array->elements[i].reference); }
			}
			append(&pending, array->properties);
			break;
		}
		case adamic_kind_map: {
			adamic_map *map = (adamic_map *)heap;
			for (size_t i = 0; i < map->used; i++) {
				if (map->entries[i].deleted) { continue; }
				if (map->reference_keys) { append(&pending, map->entries[i].key.reference); }
				if (map->reference_values) { append(&pending, map->entries[i].value.reference); }
			}
			break;
		}
		case adamic_kind_cell: {
			adamic_cell *cell = (adamic_cell *)heap;
			if (cell->references) { append(&pending, cell->value.reference); }
			break;
		}
		case adamic_kind_closure: {
			adamic_closure *closure = (adamic_closure *)heap;
			for (size_t i = 0; i < closure->count; i++) { append(&pending, closure->cells[i]); }
			break;
		}
		case adamic_kind_number: case adamic_kind_boolean: break;
		case adamic_kind_typed_array: case adamic_kind_typed_array_iterator:
		case adamic_kind_weak: case adamic_kind_map_iterator:
			adamic_panic("compiler bug: non-shareable value reached a task", sizeof "compiler bug: non-shareable value reached a task" - 1);
		}
	}
	free(pending.values);
	free(pending.seen);
	pthread_mutex_unlock(&sharing);
}
