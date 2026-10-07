// maybe.c: number | undefined and boolean | undefined, where the pair itself is seen.

#include "adamic.h"

#include <math.h>
#include <stdint.h>
#include <string.h>

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

double adamic_maybe_number_pack(adamic_maybe_number value) {
	uint64_t bits = ADAMIC_UNDEFINED_BITS;
	if (value.present) {
		if (!isnan(value.number)) {
			return value.number;
		}
		// NaN, -NaN, 0/0's NaN, any payload: all one quiet NaN, never the reserved one.
		bits = 0x7ff8000000000000u;
	}
	double packed;
	memcpy(&packed, &bits, sizeof packed);
	return packed;
}

adamic_maybe_number adamic_maybe_number_unpack(double packed) {
	uint64_t bits;
	memcpy(&bits, &packed, sizeof bits);
	if (bits == ADAMIC_UNDEFINED_BITS) {
		return (adamic_maybe_number){false, 0.0};
	}
	return (adamic_maybe_number){true, packed};
}

uint8_t adamic_maybe_boolean_pack(adamic_maybe_boolean value) {
 return value.present ? (uint8_t)value.boolean : 2;
}

adamic_maybe_boolean adamic_maybe_boolean_unpack(uint8_t packed) {
 return (adamic_maybe_boolean){packed != 2, packed == 1};
}
