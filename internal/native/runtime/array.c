// array.c: arrays.

#include "adamic.h"

#include <math.h>

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
