#include "array_maybe_boolean.h"

// SortIndexedProperties never calls the comparator for undefined. The original receiver
// stays visible to comparator side effects; only the collected defined values are sorted.
void adamic_array_sort_maybe_boolean(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
	size_t length = array->length;
	adamic_array *defined = adamic_array_new(length, false);
	for (size_t index = 0; index < length; index++) {
		if (adamic_maybe_boolean_unpack(array->elements[index].maybe_boolean).present) {
			adamic_array_push(defined, array->elements[index]);
		}
	}
	adamic_array_sort(defined, compare, context);
	if (adamic_thrown == NULL) {
		adamic_value missing = {.maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){false, false})};
		for (size_t index = 0; index < length; index++) {
			adamic_value value = index < defined->length ? defined->elements[index] : missing;
			if (index < array->length) {
				array->elements[index] = value;
			} else {
				adamic_array_push(array, value);
			}
		}
	}
	adamic_release(defined);
}
