// math.c: JavaScript's Math, where C's library means something else (ECMA-262, the Math object).

#include "adamic.h"

#include <math.h>

double adamic_math_round(double value) {
	if (isnan(value) || isinf(value)) {
		return value;
	}
	// Half rounds up, toward +Infinity: Math.round(2.5) is 3 and Math.round(-2.5) is -2. The
	// difference from the floor is exact for every double, so 0.49999999999999994 stays 0, where
	// floor(value + 0.5) would round it up.
	double floored = floor(value);
	double rounded = value - floored >= 0.5 ? floored + 1 : floored;
	// From -0.5 up to -0, the answer is -0.
	if (rounded == 0 && signbit(value)) {
		return -0.0;
	}
	return rounded;
}

double adamic_math_sign(double value) {
	if (isnan(value) || value == 0) {
		return value;
	}
	return value > 0 ? 1 : -1;
}

double adamic_math_max(double left, double right) {
	if (isnan(left) || isnan(right)) {
		return NAN;
	}
	if (left == 0 && right == 0) {
		return signbit(left) && signbit(right) ? -0.0 : 0.0;
	}
	return left > right ? left : right;
}

double adamic_math_min(double left, double right) {
	if (isnan(left) || isnan(right)) {
		return NAN;
	}
	if (left == 0 && right == 0) {
		return signbit(left) || signbit(right) ? -0.0 : 0.0;
	}
	return left < right ? left : right;
}
