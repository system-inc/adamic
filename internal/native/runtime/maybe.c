// maybe.c: number | undefined and boolean | undefined, where the pair itself is seen.

#include "adamic.h"

adamic_string adamic_string_undefined = ADAMIC_STRING("undefined");

adamic_string *adamic_string_from_maybe_number(adamic_maybe_number value) {
	if (!value.present) {
		return &adamic_string_undefined;
	}
	return adamic_string_from_number(value.number);
}

bool adamic_maybe_number_equal(adamic_maybe_number left, adamic_maybe_number right) {
	// NaN is never ===, present or not; a missing value is === only to another.
	return left.present == right.present && (!left.present || left.number == right.number);
}

bool adamic_maybe_boolean_equal(adamic_maybe_boolean left, adamic_maybe_boolean right) {
	return left.present == right.present && (!left.present || left.boolean == right.boolean);
}
