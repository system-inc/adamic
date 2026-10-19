// sort_undefined.c: sort on an array of number | undefined, as V8 does it (SortIndexedProperties):
// the numbers are taken out and sorted by the comparator, which never sees undefined, then written
// back first and every undefined after them.

#include "adamic.h"

void adamic_array_sort_undefined_last(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
	size_t length = array->length;
	adamic_array *defined = adamic_array_new(length, false);
	for (size_t index = 0; index < length; index++) {
		if (adamic_maybe_number_unpack(array->elements[index].number).present) {
			adamic_array_push(defined, array->elements[index]);
		}
	}
	// The comparator may change the array it sorts; what it sorts is this copy, which nothing else sees.
	adamic_array_sort(defined, compare, context);
	if (adamic_exception_pending) {
		// A comparator threw: nothing is written back, as V8 writes nothing back then.
		adamic_release(defined);
		return;
	}
	// Written back index by index, as adamic_array_sort does: past a length the comparator shrank, the
	// array grows again.
	adamic_value undefined = {.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0.0})};
	for (size_t index = 0; index < length; index++) {
		adamic_value value = index < defined->length ? defined->elements[index] : undefined;
		if (index < array->length) {
			array->elements[index] = value;
		} else {
			adamic_array_push(array, value);
		}
	}
	adamic_release(defined);
}
