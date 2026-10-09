// Copyright 2019 the V8 project authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// The notice above is V8's, from the top of src/builtins/math.tq, kept verbatim as its license asks.
//
// hypot.c: Math.hypot, as V8 computes it (src/builtins/math.tq, MathHypot, V8 13.6.233.17).
//
// V8 has a fast path for up to three arguments (FastMathHypot) and a loop for more. They give the
// same bits: the fast path is the loop unrolled, its compensation term written out. So this is the
// loop: every value scaled by the largest, the squares added with Kahan's compensated summation, and
// the square root scaled back. C's hypot is a different algorithm and differs in the last bit.

#include "adamic.h"

#include <math.h>

// Fused multiply-adds would change the compensated sum's bits; see ieee754.c.
#ifndef ADAMIC_FUSED_RUNTIME
#pragma STDC FP_CONTRACT OFF
#endif

double adamic_math_hypot(size_t count, const double *values) {
	// An infinity wins over NaN (Math.hypot(NaN, Infinity) is Infinity), and NaN over the rest.
	bool hasNaN = false;
	double max = 0;
	for (size_t index = 0; index < count; index++) {
		if (isnan(values[index])) {
			hasNaN = true;
		} else if (fabs(values[index]) > max) {
			max = fabs(values[index]);
		}
	}
	if (isinf(max)) {
		return INFINITY;
	}
	if (hasNaN) {
		return NAN;
	}
	// Every value is ±0, or there are none: +0, never -0.
	if (max == 0) {
		return 0;
	}
	// Normalized to the largest, so the squares can't overflow.
	double sum = 0;
	double compensation = 0;
	for (size_t index = 0; index < count; index++) {
		double scaled = fabs(values[index]) / max;
		double summand = scaled * scaled - compensation;
		double preliminary = sum + summand;
		compensation = (preliminary - sum) - summand;
		sum = preliminary;
	}
	return sqrt(sum) * max;
}
