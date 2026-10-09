// Calls whose arguments include a spread, Math.max(...values): the arguments, already evaluated in
// order into an array of numbers, folded as the calls fold their arguments.

#include "adamic.h"

#include <math.h>
#include <stdlib.h>

double adamic_math_max_of(const adamic_array *values) {
	// Math.max() is -Infinity, and max(-Infinity, x) is x, NaN and -0 included.
	double result = -HUGE_VAL;
	for (size_t index = 0; index < values->length; index++) {
		result = adamic_math_max(result, values->elements[index].number);
	}
	return result;
}

double adamic_math_min_of(const adamic_array *values) {
	double result = HUGE_VAL;
	for (size_t index = 0; index < values->length; index++) {
		result = adamic_math_min(result, values->elements[index].number);
	}
	return result;
}

// numbers copies an array's numbers into memory of their own, for the functions that take a C array.
static double *numbers(const adamic_array *values) {
	double *copied = malloc((values->length == 0 ? 1 : values->length) * sizeof *copied);
	if (copied == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t index = 0; index < values->length; index++) {
		copied[index] = values->elements[index].number;
	}
	return copied;
}

double adamic_math_hypot_of(const adamic_array *values) {
	if (values->length == 0) {
		return 0.0;
	}
	double *copied = numbers(values);
	double result = adamic_math_hypot(values->length, copied);
	free(copied);
	return result;
}

adamic_string *adamic_string_from_char_codes_of(const adamic_array *values) {
	double *copied = numbers(values);
	adamic_string *result = adamic_string_from_char_codes(values->length, copied);
	free(copied);
	return result;
}

adamic_string *adamic_string_from_code_points_of(const adamic_array *values) {
	// A RangeError panics from inside, where the copy would be let go of at exit anyway.
	double *copied = numbers(values);
	adamic_string *result = adamic_string_from_code_points(values->length, copied);
	free(copied);
	return result;
}
