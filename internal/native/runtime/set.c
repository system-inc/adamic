// set.c: Set, held as a Map whose values aren't used (map.c), so it has a Map's insertion order,
// SameValueZero keys and live iteration, as JavaScript's Set does.

#include "adamic.h"

void adamic_set_add_all(adamic_map *set, const adamic_array *values) {
	for (size_t index = 0; index < values->length; index++) {
		adamic_value element = values->elements[index];
		if (set->reference_keys) {
			adamic_graph_hold(set, element.reference);
		}
		adamic_map_set(set, element, (adamic_value){.number = 0});
	}
}

adamic_array *adamic_set_values(const adamic_map *set) {
	adamic_array *values = adamic_array_new(set->count, set->reference_keys);
	for (size_t index = 0; index < set->used; index++) {
		if (set->entries[index].deleted) {
			continue;
		}
		adamic_value element = set->entries[index].key;
		if (set->reference_keys) {
			adamic_retain(element.reference);
		}
		adamic_array_push(values, element);
	}
	return values;
}
