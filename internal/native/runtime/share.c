// share.c: publish only after the complete reachable graph is permanently marked.
#include "adamic.h"
#include <pthread.h>
#include <stdlib.h>

// Serializing graph preparation avoids a second publisher mistaking a partly marked graph for a
// complete one. The lock is outside the element loop. Retain/release never take it.
static pthread_mutex_t sharing = PTHREAD_MUTEX_INITIALIZER;
typedef struct { void **values; size_t count; size_t capacity; } pending_values;
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
		if (adamic_is_shared(heap)) { continue; }
		// Static scalar/closure identities need no marking. Region objects also have a zero count,
		// but their children must be visited and their storage remains alive through the join.
		if (adamic_reference_count(heap) == 0 && (heap->slab & ADAMIC_REGION_VALUE) == 0) { continue; }
		if (heap->kind == adamic_kind_string) { adamic_string_prepare_shared((adamic_string *)heap); }
		__atomic_fetch_or(&heap->slab, ADAMIC_SHARED, __ATOMIC_RELAXED);
		switch (heap->kind) {
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
		case adamic_kind_weak: case adamic_kind_map_iterator:
			adamic_panic("compiler bug: non-shareable value reached a task", sizeof "compiler bug: non-shareable value reached a task" - 1);
		}
	}
	free(pending.values);
	pthread_mutex_unlock(&sharing);
}
