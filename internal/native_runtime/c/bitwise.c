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

// to_uint32 is ToUint32.
static uint32_t to_uint32(double value) {
	if (!isfinite(value)) {
		return 0;
	}
	double modulo = fmod(trunc(value), 4294967296.0);
	if (modulo < 0) {
		modulo += 4294967296.0;
	}
	return (uint32_t)modulo;
}

// signed_value is 32 bits read as a two's complement integer, which is what ToInt32 gives.
static double signed_value(uint32_t bits) {
	return bits >= 0x80000000u ? (double)bits - 4294967296.0 : (double)bits;
}

// shift_count is a shift's right operand: ToUint32, then its low five bits.
static unsigned shift_count(double count) {
	return (unsigned)(to_uint32(count) & 31u);
}

double adamic_bitwise_and(double left, double right) {
	return signed_value(to_uint32(left) & to_uint32(right));
}

double adamic_bitwise_or(double left, double right) {
	return signed_value(to_uint32(left) | to_uint32(right));
}

double adamic_bitwise_xor(double left, double right) {
	return signed_value(to_uint32(left) ^ to_uint32(right));
}

double adamic_bitwise_not(double value) {
	return signed_value(~to_uint32(value));
}

double adamic_shift_left(double left, double right) {
	return signed_value((uint32_t)(to_uint32(left) << shift_count(right)));
}

// adamic_shift_right is >>, which keeps the sign: the bits shifted in are the sign bit's.
double adamic_shift_right(double left, double right) {
	uint32_t bits = to_uint32(left);
	unsigned count = shift_count(right);
	uint32_t shifted = bits >> count;
	if ((bits & 0x80000000u) != 0 && count > 0) {
		shifted |= ~(UINT32_MAX >> count);
	}
	return signed_value(shifted);
}

// adamic_shift_right_unsigned is >>>, whose result is ToUint32's, from 0 to 2^32 - 1.
double adamic_shift_right_unsigned(double left, double right) {
	return (double)(to_uint32(left) >> shift_count(right));
}
