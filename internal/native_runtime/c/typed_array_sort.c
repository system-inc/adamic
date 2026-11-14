// typed_array_sort.c: stable numeric sorting over the receiver's unboxed range.
#include "adamic.h"

#include <stdlib.h>
#include <string.h>

static bool before_float64(double left, double right) {
	if (isnan(left)) { return false; }
	if (isnan(right)) { return true; }
	if (left == 0 && right == 0) { return signbit(left) && !signbit(right); }
	return left < right;
}

#define NUMERIC_BEFORE(left, right) ((left) < (right))

// Specialize the merge loop so each element keeps its own width and comparisons inline.
// Choosing the left run on equality preserves even the order of equal NaN payloads.
#define SORT_ELEMENTS(name, element, before) \
static void name(element *data, size_t length) { \
	if (length < 2) { return; } \
	element *temporary = malloc(length * sizeof *data); \
	if (temporary == NULL) { \
		static const char message[] = "out of memory"; \
		adamic_panic(message, sizeof message - 1); \
	} \
	element *source = data, *destination = temporary; \
	for (size_t run = 1; run < length; run = run > length / 2 ? length : run * 2) { \
		for (size_t base = 0; base < length;) { \
			size_t middle = base + (run < length - base ? run : length - base); \
			size_t end = middle + (run < length - middle ? run : length - middle); \
			size_t left = base, right = middle, out = base; \
			while (left < middle && right < end) { \
				if (before(source[right], source[left])) { destination[out++] = source[right++]; } \
				else { destination[out++] = source[left++]; } \
			} \
			memcpy(destination + out, source + left, (middle - left) * sizeof *data); \
			out += middle - left; \
			memcpy(destination + out, source + right, (end - right) * sizeof *data); \
			base = end; \
		} \
		element *swapped = source; source = destination; destination = swapped; \
	} \
	if (source != data) { memcpy(data, source, length * sizeof *data); } \
	free(temporary); \
}

SORT_ELEMENTS(sort_uint8, uint8_t, NUMERIC_BEFORE)
SORT_ELEMENTS(sort_uint16, uint16_t, NUMERIC_BEFORE)
SORT_ELEMENTS(sort_int32, int32_t, NUMERIC_BEFORE)
SORT_ELEMENTS(sort_float64, double, before_float64)

#undef SORT_ELEMENTS
#undef NUMERIC_BEFORE

adamic_typed_array *adamic_typed_array_sort(adamic_typed_array *array) {
	// A view's data and length describe only its range, not its owner's whole buffer.
	adamic_typed_array *range = array;
	switch (range->kind) {
	case adamic_typed_array_uint8: sort_uint8(range->data, range->length); break;
	case adamic_typed_array_uint16: sort_uint16(range->data, range->length); break;
	case adamic_typed_array_int32: sort_int32(range->data, range->length); break;
	case adamic_typed_array_float64: sort_float64(range->data, range->length); break;
	default: adamic_unreachable();
	}
	return array;
}
