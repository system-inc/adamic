#include "adamic.h"
#include "count.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const double edges[] = {300, -1, 255, 256, NAN, INFINITY, -INFINITY, -0.0,
	2147483648.0, 4294967296.0, -2147483649.0, 1.9, -1.9, 255.9, -256.9,
	4294967295.9, -4294967296.9, 1e100, -1e100, 0x1p53, 0x1p-1074};

static void number(double value) {
	if (value == 0 && signbit(value)) { puts("-0"); return; }
	char buffer[ADAMIC_NUMBER_FORMAT_MAX];
	size_t size = adamic_number_format(value, buffer);
	printf("%.*s\n", (int)size, buffer);
}
static void show(const adamic_typed_array *array) {
	number(adamic_typed_array_length(array));
	for (size_t index = 0; index < array->length; index++) {
		adamic_maybe_number value = adamic_typed_array_get(array, (double)index);
		assert(value.present);
		number(value.number);
	}
}
static adamic_typed_array *sequence(enum adamic_typed_array_kind kind) {
	adamic_typed_array *array = adamic_typed_array_new(kind, 6);
	for (size_t index = 0; index < 6; index++) { adamic_typed_array_set(array, (double)index, (double)index + 1); }
	return array;
}
static void suite(enum adamic_typed_array_kind kind) {
	adamic_array *input = adamic_array_new(sizeof edges / sizeof *edges, false);
	for (size_t index = 0; index < sizeof edges / sizeof *edges; index++) {
		adamic_array_push(input, (adamic_value){.number = edges[index]});
	}
	size_t before = adamic_counted.allocations;
	adamic_typed_array *array = adamic_typed_array_from_numbers(kind, input);
	assert(adamic_counted.allocations == before + 1);
	assert(array->owner == NULL);
	show(array);
	for (size_t index = 0; index < sizeof edges / sizeof *edges; index++) {
		adamic_typed_array_set(array, (double)index, edges[index]);
	}
	show(array);
	adamic_typed_array *copied = adamic_typed_array_new(kind, (double)array->length);
	adamic_typed_array_set_from(copied, array, 0, false); show(copied); adamic_release(copied);
	copied = adamic_typed_array_new(kind, 1);
	for (size_t index = 0; index < sizeof edges / sizeof *edges; index++) {
		adamic_typed_array_fill(copied, edges[index], 0, 0, false, false); show(copied);
	}
	adamic_release(copied);
	const double absent[] = {-1, 0.5, NAN, INFINITY, -INFINITY, (double)array->length, 1e100};
	for (size_t index = 0; index < sizeof absent / sizeof *absent; index++) {
		adamic_maybe_number value = adamic_typed_array_get(array, absent[index]);
		assert(!value.present && value.number == 0);
		puts("undefined");
	}
	number(adamic_typed_array_get(array, -0.0).number);
	adamic_release(array);
	adamic_release(input);
	const double lengths[] = {0, NAN, -0.0, -0.9, 3.9};
	for (size_t index = 0; index < sizeof lengths / sizeof *lengths; index++) {
		array = adamic_typed_array_new(kind, lengths[index]); show(array); adamic_release(array);
	}
	array = sequence(kind);
	adamic_typed_array *view = adamic_typed_array_subarray(array, 1, -1, true);
	adamic_typed_array_set(view, 0, 300); show(array);
	adamic_typed_array_set(array, 2, -1); show(view);
	adamic_typed_array *nested = adamic_typed_array_subarray(view, 1, 0, false);
	show(nested);
	adamic_release(view);
	adamic_release(array); show(nested);
	adamic_typed_array_set(nested, 0, 42); show(nested);
	adamic_release(nested);
	const double indexes[] = {-INFINITY, -99, -6, -2.9, -0.0, NAN, 0.9, 2.9, 6, 99, INFINITY};
	for (size_t first = 0; first < sizeof indexes / sizeof *indexes; first++) {
		for (size_t last = 0; last < sizeof indexes / sizeof *indexes; last++) {
			array = sequence(kind);
			assert(adamic_typed_array_fill(array, 300, indexes[first], indexes[last], true, true) == array);
			show(array);
			view = adamic_typed_array_subarray(array, indexes[first], indexes[last], true);
			show(view); adamic_release(view); adamic_release(array);
		}
	}
	array = sequence(kind);
	adamic_typed_array_fill(array, -1, 99, 99, false, false); show(array);
	adamic_typed_array_fill(array, -0.0, -2, 0, true, false); show(array);
	adamic_release(array);
	const double offsets[] = {0, -0.5, NAN, 1.9, 2};
	for (size_t index = 0; index < sizeof offsets / sizeof *offsets; index++) {
		array = sequence(kind);
		view = adamic_typed_array_subarray(array, 0, 4, true);
		adamic_typed_array_set_from(array, view, offsets[index], true); show(array);
		adamic_release(view); adamic_release(array);
	}
	array = sequence(kind);
	view = adamic_typed_array_subarray(array, 2, 0, false);
	adamic_typed_array_set_from(array, view, 0, false); show(array);
	adamic_typed_array_set_from(view, view, 0, true); show(view);
	adamic_typed_array *empty = adamic_typed_array_subarray(array, 6, 0, false);
	adamic_typed_array_set_from(array, empty, 6, true);
	adamic_typed_array_iterator *iterator = adamic_typed_array_iterate(array);
	double value;
	assert(adamic_typed_array_iterator_next(iterator, &value)); number(value);
	adamic_typed_array_set(view, 0, 77);
	adamic_release(empty); adamic_release(view); adamic_release(array);
	while (adamic_typed_array_iterator_next(iterator, &value)) { number(value); }
	assert(!adamic_typed_array_iterator_next(iterator, &value));
	adamic_release(iterator);
	array = sequence(kind); iterator = adamic_typed_array_iterate(array);
	adamic_release(array); adamic_release(iterator); // Early exit owns no lingering count.
	array = adamic_typed_array_new(kind, 0);
	view = adamic_typed_array_subarray(array, 0, 0, false);
	adamic_release(array); show(view); adamic_release(view);
}
int main(int argc, char **argv) {
	if (argc > 1) {
		adamic_typed_array *array = adamic_typed_array_new(adamic_typed_array_uint8, 3);
		double value = argc > 2 ? strtod(argv[2], NULL) : 3;
		if (strcmp(argv[1], "write") == 0) { adamic_typed_array_set(array, value, 1); }
		if (strcmp(argv[1], "check") == 0) { adamic_typed_array_check_write(array, value); }
		if (strcmp(argv[1], "length") == 0) { adamic_typed_array *other = adamic_typed_array_new(adamic_typed_array_uint8, value); adamic_release(other); }
		if (strcmp(argv[1], "offset") == 0) { adamic_typed_array_set_from(array, array, value, true); }
		if (strcmp(argv[1], "kind") == 0) { adamic_typed_array *other = adamic_typed_array_new(adamic_typed_array_int32, 0); adamic_typed_array_set_from(array, other, 0, false); adamic_release(other); }
		adamic_release(array); return 0;
	}
	for (int kind = adamic_typed_array_uint8; kind <= adamic_typed_array_float64; kind++) { suite((enum adamic_typed_array_kind)kind); }
	assert(adamic_counted.allocations == adamic_counted.frees);
	puts("balanced");
	return 0;
}
