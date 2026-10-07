// typed_array.c: fixed-width numeric storage and counted, mutable views.
#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

_Static_assert(sizeof(double) == 8, "Float64Array requires 8-byte doubles");

static size_t width(enum adamic_typed_array_kind kind) {
	switch (kind) {
	case adamic_typed_array_uint8: return 1;
	case adamic_typed_array_int32: return 4;
	case adamic_typed_array_float64: return 8;
	}
	adamic_unreachable();
}

static void range_error(void) {
	static const char message[] = "RangeError: offset is out of bounds";
	adamic_panic(message, sizeof message - 1);
}

static adamic_typed_array *allocate(enum adamic_typed_array_kind kind, size_t length) {
	size_t size = width(kind);
	if (length > SIZE_MAX / size) {
		static const char message[] = "RangeError: Invalid typed array length";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_typed_array *array = adamic_allocate(sizeof *array, adamic_kind_typed_array);
	array->kind = kind;
	array->length = length;
	array->owner = NULL;
	// A zero-length owner still has a valid base for empty views; never add to NULL.
	array->data = calloc(length == 0 ? 1 : length, size);
	if (array->data == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return array;
}

adamic_typed_array *adamic_typed_array_new(enum adamic_typed_array_kind kind, double length) {
	// ToIndex truncates before checking the sign: -0.5 is zero, -1 is invalid.
	double integer = isnan(length) ? 0 : trunc(length);
	if (!(integer >= 0) || integer > 9007199254740991.0 || integer >= (double)SIZE_MAX) {
		static const char message[] = "RangeError: Invalid typed array length";
		adamic_panic(message, sizeof message - 1);
	}
	return allocate(kind, (size_t)integer);
}

static double read_element(const adamic_typed_array *array, size_t index) {
	switch (array->kind) {
	case adamic_typed_array_uint8: return ((const uint8_t *)array->data)[index];
	case adamic_typed_array_int32: return ((const int32_t *)array->data)[index];
	case adamic_typed_array_float64: return ((const double *)array->data)[index];
	}
	adamic_unreachable();
}

static void write_element(adamic_typed_array *array, size_t index, double value) {
	if (array->kind == adamic_typed_array_float64) {
		((double *)array->data)[index] = value;
		return;
	}
	// Avoid undefined float-to-integer casts and implementation-defined signed wrap.
	double modulus = array->kind == adamic_typed_array_uint8 ? 256.0 : 4294967296.0;
	double modulo = isfinite(value) ? fmod(trunc(value), modulus) : 0;
	if (modulo < 0) { modulo += modulus; }
	if (array->kind == adamic_typed_array_uint8) {
		((uint8_t *)array->data)[index] = (uint8_t)modulo;
	} else {
		double signed_value = modulo >= 2147483648.0 ? modulo - 4294967296.0 : modulo;
		((int32_t *)array->data)[index] = (int32_t)signed_value;
	}
}

adamic_typed_array *adamic_typed_array_from_numbers(enum adamic_typed_array_kind kind, const adamic_array *numbers) {
	adamic_typed_array *array = allocate(kind, numbers->length);
	for (size_t index = 0; index < numbers->length; index++) {
		write_element(array, index, numbers->elements[index].number);
	}
	return array;
}

static bool valid_index(const adamic_typed_array *array, double index) {
	return index >= 0 && index < (double)array->length && index == trunc(index);
}

adamic_maybe_number adamic_typed_array_get(const adamic_typed_array *array, double index) {
	if (!valid_index(array, index)) { return (adamic_maybe_number){false, 0}; }
	return (adamic_maybe_number){true, read_element(array, (size_t)index)};
}

void adamic_typed_array_check_write(const adamic_typed_array *array, double index) {
	if (!valid_index(array, index)) {
		char number[ADAMIC_NUMBER_FORMAT_MAX], length[ADAMIC_NUMBER_FORMAT_MAX], message[128];
		size_t number_size = adamic_number_format(index, number);
		size_t length_size = adamic_number_format((double)array->length, length);
		int written = snprintf(message, sizeof message, "index %.*s is outside an array of length %.*s", (int)number_size, number, (int)length_size, length);
		adamic_panic(message, (size_t)written);
	}
}

void adamic_typed_array_set(adamic_typed_array *array, double index, double value) {
	adamic_typed_array_check_write(array, index);
	write_element(array, (size_t)index, value);
}

double adamic_typed_array_length(const adamic_typed_array *array) { return (double)array->length; }

static size_t relative(double index, size_t length) {
	double integer = isnan(index) ? 0 : trunc(index);
	if (integer < 0) { integer += (double)length; }
	return integer <= 0 ? 0 : integer >= (double)length ? length : (size_t)integer;
}

adamic_typed_array *adamic_typed_array_fill(adamic_typed_array *array, double value, double start, double end, bool has_start, bool has_end) {
	size_t from = has_start ? relative(start, array->length) : 0;
	size_t to = has_end ? relative(end, array->length) : array->length;
	for (size_t index = from; index < to; index++) { write_element(array, index, value); }
	return array;
}

void adamic_typed_array_set_from(adamic_typed_array *array, const adamic_typed_array *source, double offset, bool has_offset) {
	if (array->kind != source->kind) {
		static const char message[] = "typed array set requires the same element type";
		adamic_panic(message, sizeof message - 1);
	}
	double integer = !has_offset || isnan(offset) ? 0 : trunc(offset);
	if (!(integer >= 0) || integer > (double)array->length) { range_error(); }
	size_t from = (size_t)integer;
	if (source->length > array->length - from) { range_error(); }
	if (source->length > 0) {
		// Same kind means the byte copy preserves conversion and all Float64 bits.
		size_t size = width(array->kind);
		memmove((unsigned char *)array->data + from * size, source->data, source->length * size);
	}
}

adamic_typed_array *adamic_typed_array_subarray(const adamic_typed_array *array, double start, double end, bool has_end) {
	size_t from = relative(start, array->length);
	size_t to = has_end ? relative(end, array->length) : array->length;
	adamic_typed_array *view = adamic_allocate(sizeof *view, adamic_kind_typed_array);
	view->kind = array->kind;
	view->length = to > from ? to - from : 0;
	view->data = (unsigned char *)array->data + from * width(array->kind);
	view->owner = adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array);
	return view;
}

adamic_typed_array_iterator *adamic_typed_array_iterate(adamic_typed_array *array) {
	adamic_typed_array_iterator *iterator = adamic_allocate(sizeof *iterator, adamic_kind_typed_array_iterator);
	iterator->array = adamic_retain(array);
	iterator->next = 0;
	return iterator;
}

bool adamic_typed_array_iterator_next(adamic_typed_array_iterator *iterator, double *value) {
	if (iterator->next >= iterator->array->length) { return false; }
	*value = read_element(iterator->array, iterator->next++);
	return true;
}
