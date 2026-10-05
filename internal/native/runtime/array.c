// array.c: arrays.

#include "adamic.h"

#include <math.h>
#include <stdio.h>

#include <stdlib.h>
#include <string.h>

adamic_array *adamic_array_new(size_t capacity, bool references) {
	adamic_array *array = adamic_allocate(sizeof *array, adamic_kind_array);
	array->length = 0;
	array->capacity = capacity;
	array->references = references;
	array->elements = NULL;
	if (capacity > 0) {
		array->elements = malloc(capacity * sizeof *array->elements);
		if (array->elements == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
	}
	return array;
}

void adamic_array_push(adamic_array *array, adamic_value value) {
	if (array->length == array->capacity) {
		size_t capacity = array->capacity == 0 ? 4 : array->capacity * 2;
		adamic_value *grown = realloc(array->elements, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		array->elements = grown;
		array->capacity = capacity;
	}
	array->elements[array->length++] = value;
}

// adamic_array_join builds the string in one buffer, growing it as it goes.
adamic_string *adamic_array_join(const adamic_array *array, const adamic_string *separator, enum adamic_join kind) {
	size_t length = 0, capacity = 64;
	char *buffer = malloc(capacity);
	if (buffer == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t index = 0; index < array->length; index++) {
		char number[ADAMIC_NUMBER_FORMAT_MAX];
		const char *bytes = NULL;
		size_t size = 0;
		switch (kind) {
		case adamic_join_numbers:
			size = adamic_number_format(array->elements[index].number, number);
			bytes = number;
			break;
		case adamic_join_booleans:
			bytes = array->elements[index].boolean ? "true" : "false";
			size = strlen(bytes);
			break;
		case adamic_join_strings: {
			const adamic_string *string = array->elements[index].reference;
			bytes = string->bytes;
			size = string->length;
			break;
		}
		}
		size_t needed = length + size + (index > 0 ? separator->length : 0);
		if (needed > capacity) {
			while (capacity < needed) {
				capacity *= 2;
			}
			char *grown = realloc(buffer, capacity);
			if (grown == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			buffer = grown;
		}
		if (index > 0 && separator->length > 0) {
			memcpy(buffer + length, separator->bytes, separator->length);
			length += separator->length;
		}
		if (size > 0) {
			memcpy(buffer + length, bytes, size);
			length += size;
		}
	}
	adamic_string piece = {{0, adamic_kind_string}, length, buffer};
	adamic_string *joined = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(buffer);
	return joined;
}

adamic_array *adamic_array_slice(const adamic_array *array, double start, double end, bool has_end) {
	// ECMAScript's relative indexes: ToIntegerOrInfinity, negative from the end, clamped.
	double length = (double)array->length;
	start = isnan(start) ? 0 : trunc(start);
	start = start < 0 ? (length + start < 0 ? 0 : length + start) : (start > length ? length : start);
	if (has_end) {
		end = isnan(end) ? 0 : trunc(end);
		end = end < 0 ? (length + end < 0 ? 0 : length + end) : (end > length ? length : end);
	} else {
		end = length;
	}
	size_t from = (size_t)start, to = end > start ? (size_t)end : from;
	adamic_array *sliced = adamic_array_new(to - from, array->references);
	for (size_t index = from; index < to; index++) {
		adamic_value value = array->elements[index];
		if (array->references) {
			adamic_retain(value.reference);
		}
		adamic_array_push(sliced, value);
	}
	return sliced;
}

// adamic_array_sort sorts in place, stably (ECMA-262 requires it since 2019), by merging. compare is
// the program's comparator through an adapter that gives -1, 0 or 1, NaN read as 0.
void adamic_array_sort(adamic_array *array, int (*compare)(adamic_value, adamic_value)) {
	size_t length = array->length;
	if (length < 2) {
		return;
	}
	adamic_value *scratch = malloc(length * sizeof *scratch);
	if (scratch == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t width = 1; width < length; width *= 2) {
		for (size_t left = 0; left < length; left += 2 * width) {
			size_t middle = left + width < length ? left + width : length;
			size_t right = left + 2 * width < length ? left + 2 * width : length;
			size_t from_left = left, from_right = middle, out = left;
			while (from_left < middle && from_right < right) {
				// Take from the right only when it's strictly smaller: that keeps equal elements in
				// their original order.
				if (compare(array->elements[from_right], array->elements[from_left]) < 0) {
					scratch[out++] = array->elements[from_right++];
				} else {
					scratch[out++] = array->elements[from_left++];
				}
			}
			while (from_left < middle) {
				scratch[out++] = array->elements[from_left++];
			}
			while (from_right < right) {
				scratch[out++] = array->elements[from_right++];
			}
		}
		memcpy(array->elements, scratch, length * sizeof *scratch);
	}
	free(scratch);
}

adamic_value *adamic_array_at(const adamic_array *array, double index) {
	// An array index is an integer from 0 up to the length; anything else (negative, a fraction,
	// NaN, past the end) is a property the array doesn't have, which reads as undefined.
	if (!(index >= 0) || index >= (double)array->length || index != trunc(index)) {
		return NULL;
	}
	return &array->elements[(size_t)index];
}

adamic_value *adamic_array_at_relative(const adamic_array *array, double index) {
	// array.at(index): ToIntegerOrInfinity, so NaN is 0 and a fraction truncates, then a negative
	// index counts from the end. Anywhere outside the array is undefined.
	index = isnan(index) ? 0 : trunc(index);
	if (index < 0) {
		index += (double)array->length;
	}
	if (!(index >= 0) || index >= (double)array->length) {
		return NULL;
	}
	return &array->elements[(size_t)index];
}

double adamic_array_index_of(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool same_value_zero) {
	for (size_t index = 0; index < array->length; index++) {
		adamic_value element = array->elements[index];
		bool equal = false;
		switch (equality) {
		case adamic_equal_numbers:
			// === never finds NaN; SameValueZero (includes) does. Both take 0 and -0 as equal.
			equal = element.number == value.number || (same_value_zero && isnan(element.number) && isnan(value.number));
			break;
		case adamic_equal_booleans:
			equal = element.boolean == value.boolean;
			break;
		case adamic_equal_strings:
			equal = adamic_string_equal(element.reference, value.reference);
			break;
		case adamic_equal_identity:
			equal = element.reference == value.reference;
			break;
		}
		if (equal) {
			return (double)index;
		}
	}
	return -1;
}

adamic_array *adamic_array_reverse(adamic_array *array) {
	for (size_t left = 0, right = array->length; left + 1 < right; left++, right--) {
		adamic_value swapped = array->elements[left];
		array->elements[left] = array->elements[right - 1];
		array->elements[right - 1] = swapped;
	}
	return array;
}

adamic_array *adamic_array_concat(size_t count, adamic_array *const arrays[]) {
	size_t length = 0;
	for (size_t which = 0; which < count; which++) {
		length += arrays[which]->length;
	}
	adamic_array *joined = adamic_array_new(length, arrays[0]->references);
	for (size_t which = 0; which < count; which++) {
		for (size_t index = 0; index < arrays[which]->length; index++) {
			adamic_value value = arrays[which]->elements[index];
			if (joined->references) {
				adamic_retain(value.reference);
			}
			adamic_array_push(joined, value);
		}
	}
	return joined;
}

void adamic_array_set(adamic_array *array, double index, adamic_value value) {
	// 0.1 writes only at an index the array has: JavaScript would grow the array, or leave a hole, and
	// a hole is something 0.1 can't hold. push is how to append.
	adamic_value *slot = adamic_array_at(array, index);
	if (slot == NULL) {
		char number[ADAMIC_NUMBER_FORMAT_MAX], length[ADAMIC_NUMBER_FORMAT_MAX], message[128];
		size_t number_size = adamic_number_format(index, number);
		size_t length_size = adamic_number_format((double)array->length, length);
		int written = snprintf(message, sizeof message, "index %.*s is outside an array of length %.*s", (int)number_size, number, (int)length_size, length);
		adamic_panic(message, (size_t)written);
	}
	if (array->references) {
		void *old = slot->reference;
		slot->reference = value.reference;
		adamic_release(old);
		return;
	}
	*slot = value;
}
