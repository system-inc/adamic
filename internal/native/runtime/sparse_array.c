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

adamic_array *adamic_array_holes(double length, bool references) {
 if (!(length >= 0) || length > 4294967295.0 || length != trunc(length)) adamic_panic("RangeError: Invalid array length", sizeof "RangeError: Invalid array length" - 1);
 adamic_array *array = adamic_array_new((size_t)length, references);
 if (length > 0) memset(array->elements, 0, (size_t)length * sizeof *array->elements);
 array->length = (size_t)length;
 array->properties = adamic_object_new(&sparse_shape);
 array->properties->slots[0].reference = adamic_map_new(false, false);
 return array;
}
