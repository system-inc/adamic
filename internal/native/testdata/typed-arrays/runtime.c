#include "adamic.h"
#include "count.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const double edges[] = {0, 65535, 65536, 1.5, 4294967297.0, 300, -1, 255, 256, NAN, INFINITY, -INFINITY, -0.0,
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
// These controls are independent so one index rejection cannot mask another.
static bool inline_controls(const char *mode) {
    bool conversion = strcmp(mode, "conversion") == 0;
    bool fractional = strcmp(mode, "fractional") == 0;
    bool nan = strcmp(mode, "nan") == 0;
    bool zero = strcmp(mode, "negative-zero") == 0;
    bool bounds = strcmp(mode, "bounds") == 0;
    bool integer = strcmp(mode, "integer-bounds") == 0;
    bool write = strcmp(mode, "bounds-write") == 0;
    bool integer_write = strcmp(mode, "integer-bounds-write") == 0;
    if (!(conversion || fractional || nan || zero || bounds || integer || write || integer_write)) { return false; }
    for (int kind = adamic_typed_array_uint8; kind <= adamic_typed_array_uint16; kind++) {
        adamic_typed_array *array = adamic_typed_array_new((enum adamic_typed_array_kind)kind, 3);
        for (size_t i = 0; i < 3; i++) { adamic_typed_array_set(array, (double)i, (double)i + 1); }
        if (conversion) {
            for (size_t i = 0; i < sizeof edges / sizeof *edges; i++) {
                adamic_typed_array_set(array, 0, edges[i]);
                number(adamic_typed_array_get(array, 0).number);
                adamic_typed_array_set_integer(array, 0, edges[i]);
                number(adamic_typed_array_get_integer(array, 0).number);
            }
        }
        if (zero) {
            adamic_typed_array_set(array, -0.0, 27);
            adamic_maybe_number value = adamic_typed_array_get(array, -0.0);
            if (value.present) { number(value.number); } else { puts("undefined"); }
            adamic_typed_array_set(array, 0, -0.0);
            number(adamic_typed_array_get_integer(array, 0).number);
        }
        if (fractional || nan || bounds) {
            const double indices[] = {0.5, -0.5, 1.5, NAN, -1, 3, INFINITY, -INFINITY, 1e100};
            size_t first = fractional ? 0 : nan ? 3 : 4;
            size_t last = fractional ? 3 : nan ? 4 : sizeof indices / sizeof *indices;
            for (size_t i = first; i < last; i++) {
                adamic_maybe_number value = adamic_typed_array_get(array, indices[i]);
                if (value.present) { number(value.number); } else { assert(value.number == 0); puts("undefined"); }
            }
        }
        if (integer) {
            const int64_t indices[] = {-1, 3, INT64_MIN, INT64_MAX};
            for (size_t i = 0; i < sizeof indices / sizeof *indices; i++) {
                adamic_maybe_number value = adamic_typed_array_get_integer(array, indices[i]);
                if (value.present) { number(value.number); } else { assert(value.number == 0); puts("undefined"); }
            }
        }
        if (write) { adamic_typed_array_set(array, 3, 27); }
        if (integer_write) { adamic_typed_array_set_integer(array, 3, 27); }
        adamic_release(array);
    }
    assert(adamic_counted.allocations == adamic_counted.frees);
    return true;
}
int main(int argc, char **argv) {
    if (argc > 1 && inline_controls(argv[1])) { return 0; }
	if (argc > 1) {
		enum adamic_typed_array_kind kind = strstr(argv[1], "16") != NULL ? adamic_typed_array_uint16 : adamic_typed_array_uint8;
		adamic_typed_array *array = adamic_typed_array_new(kind, 3);
		double value = argc > 2 ? strtod(argv[2], NULL) : 3;
		if (strncmp(argv[1], "write", 5) == 0) { adamic_typed_array_set(array, value, 1); }
		if (strncmp(argv[1], "check", 5) == 0) { adamic_typed_array_check_write(array, value); }
		if (strcmp(argv[1], "length") == 0) { adamic_typed_array *other = adamic_typed_array_new(adamic_typed_array_uint8, value); adamic_release(other); }
		if (strcmp(argv[1], "offset") == 0) { adamic_typed_array_set_from(array, array, value, true); }
		if (strcmp(argv[1], "kind") == 0) { adamic_typed_array *other = adamic_typed_array_new(adamic_typed_array_int32, 0); adamic_typed_array_set_from(array, other, 0, false); adamic_release(other); }
		adamic_release(array); return 0;
	}
	for (int kind = adamic_typed_array_uint8; kind <= adamic_typed_array_uint16; kind++) { suite((enum adamic_typed_array_kind)kind); }
	assert(adamic_counted.allocations == adamic_counted.frees);
	puts("balanced");
	return 0;
}
