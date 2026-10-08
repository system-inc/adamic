#include "array_union.h"

// Undefined sorts after all defined values without a comparator call. The temporary owns
// every collected value across callbacks which may mutate the receiver or throw.
void adamic_array_sort_union(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
	size_t length = array->length;
	adamic_array *defined = adamic_array_new(length, true);
	for (size_t index = 0; index < length; index++) {
		if (array->elements[index].reference != NULL) {
			adamic_value value = {.reference = adamic_retain(array->elements[index].reference)};
			adamic_array_push(defined, value);
		}
	}
	adamic_array_sort(defined, compare, context);
	if (adamic_thrown == NULL) {
		for (size_t index = 0; index < length; index++) {
			adamic_value value = {.reference = index < defined->length ? adamic_retain(defined->elements[index].reference) : NULL};
			if (index < array->length) {
				adamic_release(array->elements[index].reference);
				array->elements[index] = value;
			} else {
				adamic_array_push(array, value);
			}
		}
	}
	adamic_release(defined);
}
