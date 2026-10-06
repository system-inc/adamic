// bitwise.c: ECMAScript's number bitwise operations, without undefined C conversions or shifts.

#include "adamic.h"

#include <stdint.h>

// to_uint32 is ECMA-262 ToUint32. fmod keeps the conversion to uint32_t in range even when the
// original double is far outside it; non-finite values and both zeroes become zero first.
static uint32_t to_uint32(double value) {
	if (!isfinite(value) || value == 0) {
		return 0;
	}
	double reduced = fmod(trunc(value), 4294967296.0);
	if (reduced < 0) {
		reduced += 4294967296.0;
	}
	return (uint32_t)reduced;
}

// from_int32_bits gives the signed interpretation as an exactly representable double, without an
// implementation-defined uint32_t to int32_t conversion.
static double from_int32_bits(uint32_t value) {
	if (value >= UINT32_C(0x80000000)) {
		return (double)value - 4294967296.0;
	}
	return (double)value;
}

double adamic_bitwise_and(double left, double right) {
	return from_int32_bits(to_uint32(left) & to_uint32(right));
}

double adamic_bitwise_or(double left, double right) {
	return from_int32_bits(to_uint32(left) | to_uint32(right));
}

double adamic_bitwise_xor(double left, double right) {
	return from_int32_bits(to_uint32(left) ^ to_uint32(right));
}

double adamic_bitwise_not(double value) {
	return from_int32_bits(~to_uint32(value));
}

double adamic_shift_left(double left, double right) {
	uint32_t count = to_uint32(right) & UINT32_C(31);
	return from_int32_bits(to_uint32(left) << count);
}

double adamic_shift_right(double left, double right) {
	uint32_t value = to_uint32(left);
	uint32_t count = to_uint32(right) & UINT32_C(31);
	if (count != 0 && (value & UINT32_C(0x80000000)) != 0) {
		value = (value >> count) | (UINT32_MAX << (32 - count));
	} else {
		value >>= count;
	}
	return from_int32_bits(value);
}

double adamic_shift_right_unsigned(double left, double right) {
	uint32_t count = to_uint32(right) & UINT32_C(31);
	return (double)(to_uint32(left) >> count);
}
