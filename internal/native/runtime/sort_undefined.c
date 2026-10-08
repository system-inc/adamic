// sort_undefined.c: SortIndexedProperties for optional numbers, optional booleans and unions.
// Defined values are collected before callbacks, sorted, then written before undefined.

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
	if (adamic_thrown != NULL) {
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
