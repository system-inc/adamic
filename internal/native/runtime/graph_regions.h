// Dynamic graph regions, distinct from the statement arenas in region.c.
#ifndef ADAMIC_GRAPH_REGIONS_H
#define ADAMIC_GRAPH_REGIONS_H

#include "adamic.h"

// Graph and opt-in Program allocations carry this two-word prefix. Ordinary layouts are unchanged.
// ADAMIC_GRAPH_FLAG or ADAMIC_PROGRAM_FLAG marks the prefix; low slab bits retain allocator identity.
typedef struct adamic_graph_header {
	struct graph_region *region;
	adamic_heap *next;
} adamic_graph_header;
static inline bool adamic_graph_is(const void *value) {
	return value != NULL && (((const adamic_heap *)value)->slab & ADAMIC_GRAPH_FLAG) != 0;
}
static inline adamic_graph_header *adamic_graph_header_of(void *value) {
	return (adamic_graph_header *)value - 1;
}

// Adoption is only for a new, fully initialized allocation, before any owned
// reference is stored in it. size is its allocation size, excluding buffers.
void *adamic_graph_adopt(void *value, size_t size);
void *adamic_heap_graph_storage(void *value, size_t size);
void adamic_graph_merge(void *left, void *right);
void adamic_graph_retain(void *value);
bool adamic_graph_release_last(void *value);
void adamic_graph_free(void *value, void (*release)(void *));
// A borrowed reference stored internally joins regions, without counting.
void *adamic_graph_hold(void *holder, void *value);
void adamic_graph_drop(void *holder, void *value);
// Design-only shared state: merging distinct shared regions must panic.
void adamic_graph_mark_shared(void *value);

// Private heap seams: children are released before any member storage is freed.
void adamic_heap_free_children(void *value, void (*release)(void *));
void adamic_heap_free_storage(void *value, uint32_t slab);

#endif
