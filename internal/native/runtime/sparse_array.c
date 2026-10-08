// Sparse presence travels with the array and is released by ordinary array cleanup.
#include "adamic.h"
#include <stdlib.h>
#include <string.h>

// Reuse the runtime's wrapper slot layout, also used by records. These are
// internal properties, never observable array own keys.
static const char *const sparse_names[] = {"0"};
static const bool sparse_references[] = {true};
static const adamic_shape sparse_shape = {1, sparse_names, sparse_references, NULL};

bool adamic_array_has_index(const adamic_array *array, size_t index) {
 if (array->properties == NULL || array->properties->shape != &sparse_shape) return true;
 adamic_map *presence = array->properties->slots[0].reference;
 return adamic_map_get(presence, (adamic_value){.number = (double)index}) != NULL;
}

void adamic_array_mark_index(adamic_array *array, size_t index) {
 if (array->properties == NULL || array->properties->shape != &sparse_shape) return;
 adamic_map *presence = array->properties->slots[0].reference;
 adamic_map_set(presence, (adamic_value){.number = (double)index}, (adamic_value){.number = 1});
}

// The host array_holes.c owns sparse construction in this scratch proof.
