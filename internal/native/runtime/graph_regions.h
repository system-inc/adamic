// Storage-prefix seam only. This unit does not implement graph regions.
#ifndef ADAMIC_GRAPH_REGIONS_H
#define ADAMIC_GRAPH_REGIONS_H
#include "adamic.h"

typedef struct adamic_graph_header {
    struct graph_region *region;
    adamic_heap *next;
} adamic_graph_header;
static inline adamic_graph_header *adamic_graph_header_of(void *value) {
    return (adamic_graph_header *)value - 1;
}
// Move a fresh counted allocation into two-word prefix storage. No count event.
void *adamic_heap_graph_storage(void *value, size_t size);
void *adamic_heap_program_storage(void *value, size_t size);
// Separate child release from storage destruction for the Program end passes.
void adamic_heap_free_children(void *value, void (*drop)(void *));
void adamic_heap_free_storage(void *value, uint32_t slab);
#endif
