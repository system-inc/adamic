// array_from.c: Array.from({ length }, callback)'s length.

#include "adamic.h"

#include <math.h>

size_t adamic_array_from_length(double length) {
	// ToLength: NaN and anything below 1 are 0, a fraction truncates. Array.from then makes the array
	// with new Array(length), which throws past 2^32 - 1.
	length = isnan(length) ? 0 : trunc(length);
	if (length <= 0) {
		return 0;
	}
	if (length > 4294967295.0) {
		static const char message[] = "RangeError: Invalid array length";
		adamic_panic(message, sizeof message - 1);
	}
	return (size_t)length;
}
