#include "adamic.h"
#include "count.h"
#include <assert.h>
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void number(double value) {
	if (value == 0 && signbit(value)) { puts("-0"); return; }
	char buffer[ADAMIC_NUMBER_FORMAT_MAX];
	size_t length = adamic_number_format(value, buffer);
	printf("%.*s\n", (int)length, buffer);
}
static void show(const adamic_typed_array *array) {
	number((double)array->length);
	for (size_t index = 0; index < array->length; index++) {
		number(adamic_typed_array_get_integer(array, (int64_t)index).number);
	}
}
static void sort(adamic_typed_array *array) {
	size_t before = adamic_counted.allocations;
	assert(adamic_typed_array_sort(array) == array);
	assert(adamic_counted.allocations == before);
	show(array);
}
int main(int argc, char **argv) {
	assert(argc == 2);
	enum adamic_typed_array_kind kind = (enum adamic_typed_array_kind)atoi(argv[1]);
	const double edges[] = {0, -0.0, NAN, 10, 2, -INFINITY, INFINITY, -1, 2, NAN, -0.0, 0, -1, 255, 65535, 2147483647, -2147483648.0};
	adamic_typed_array *array = adamic_typed_array_new(kind, sizeof edges / sizeof *edges);
	for (size_t index = 0; index < array->length; index++) { adamic_typed_array_set_integer(array, (int64_t)index, edges[index]); }
	sort(array); sort(array); adamic_release(array);
	array = adamic_typed_array_new(kind, 6);
	const double values[] = {77, 10, 2, 3, 2, 66};
	for (size_t index = 0; index < 6; index++) { adamic_typed_array_set_integer(array, (int64_t)index, values[index]); }
	adamic_typed_array *view = adamic_typed_array_subarray(array, 1, -1, true);
	adamic_typed_array *nested = adamic_typed_array_subarray(view, 0, 0, false);
	sort(nested); show(array); show(view);
	adamic_release(view); adamic_release(array);
	sort(nested); adamic_release(nested);
	const size_t lengths[] = {0, 1, 2, 3, 7, 16, 31, 32, 33, 127, 128, 129, 1025};
	for (size_t at = 0; at < sizeof lengths / sizeof *lengths; at++) {
		array = adamic_typed_array_new(kind, (double)lengths[at]);
		uint32_t state = 42;
		for (size_t index = 0; index < array->length; index++) {
			state = state * UINT32_C(1664525) + UINT32_C(1013904223);
			adamic_typed_array_set_integer(array, (int64_t)index, (double)(state % 10001) - 5000);
		}
		sort(array); adamic_release(array);
	}
	if (kind == adamic_typed_array_float64) {
		const uint64_t first = UINT64_C(0x7ff8000000000001), second = UINT64_C(0x7ff8000000000002), third = UINT64_C(0x7ff8000000000003);
		array = adamic_typed_array_new(kind, 5);
		double *data = array->data;
		memcpy(&data[0], &first, sizeof first); data[1] = 2;
		memcpy(&data[2], &second, sizeof second); data[3] = 1;
		memcpy(&data[4], &third, sizeof third);
		sort(array);
		for (size_t index = 2; index < 5; index++) {
			uint64_t bits; memcpy(&bits, &data[index], sizeof bits);
			printf("%016" PRIx64 "\n", bits);
		}
		adamic_release(array);
	}
	assert(adamic_counted.allocations == adamic_counted.frees);
	puts("balanced");
	return 0;
}
