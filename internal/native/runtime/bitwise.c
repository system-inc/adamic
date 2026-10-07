// JavaScript's bitwise operators: &, |, ^, ~, <<, >> and >>>, on numbers.
//
// Each converts its operands with ECMAScript's ToInt32 or ToUint32 and gives a number back. The
// conversions are done exactly, as the specification states them: NaN and the infinities are 0, and
// anything else is truncated and taken modulo 2^32, which fmod does without rounding. Turning 32 bits
// into a signed value is arithmetic too, never a C conversion of an out-of-range value, which C leaves
// to the implementation.

#include "adamic.h"

#include <math.h>
#include <stdint.h>

// The uncommon conversion stays out of line; common machine integers are handled in integer.h.
uint32_t adamic_to_uint32_slow(double value) {
	if (!isfinite(value)) {
		return 0;
	}
	double modulo = fmod(trunc(value), 4294967296.0);
	if (modulo < 0) {
		modulo += 4294967296.0;
	}
	return (uint32_t)modulo;
}
