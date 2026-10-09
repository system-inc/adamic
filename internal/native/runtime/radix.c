// Copyright 2011 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// The notice above is V8's, from the top of src/numbers/conversions.cc, kept verbatim as its license
// asks. The order the builtin checks things in is from src/builtins/number.tq, whose notice is:
//
// Copyright 2019 the V8 project authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// radix.c: number.toString(radix), byte for byte as Node writes it (V8 13.6.233.17, the V8 in
// Node 24): V8's DoubleToRadixStringView, ported to C with its logic and comments kept, and
// NumberPrototypeToString's checks before it.
//
// The digits after the point stop where the double's own precision does (half the distance to the
// next double), and the last one rounds to even, carrying back through the digits already written.
// That's V8's algorithm, not the exact expansion: (0.1).toString(3) stops after 34 digits.

#include "adamic.h"
#include "library_errors.h"

#include <math.h>
#include <stdint.h>
#include <string.h>

// The math here (a multiply, then a subtraction) is rounded step by step, as V8's is; see ieee754.c.
#ifndef ADAMIC_FUSED_RUNTIME
#pragma STDC FP_CONTRACT OFF
#endif

// V8's kDoubleToRadixMaxChars (src/numbers/conversions.h): 1,074 binary digits after the point for
// the smallest double, 1,024 before it for the largest, a sign and a point, with room.
enum { radix_buffer_size = 2200 };

// Double::Exponent (src/base/numbers/double.h): the exponent of the double read as an integer
// significand, so it's above zero exactly when the double is 2^53 or more (and finite).
static int radix_double_exponent(double value) {
	uint64_t bits;
	memcpy(&bits, &value, sizeof bits);
	int biased = (int)((bits >> 52) & 0x7FF);
	return biased == 0 ? -1074 : biased - 1075;
}

// Double::NextDouble, for a value that's positive and finite.
static double radix_next_double(double value) {
	uint64_t bits;
	memcpy(&bits, &value, sizeof bits);
	bits++;
	memcpy(&value, &bits, sizeof value);
	return value;
}

// DoubleToRadixStringView. value is finite and not zero; radix is 2 to 36. Returns where the text
// starts in buffer, and its length.
static const char *double_to_radix(double value, int radix, char buffer[radix_buffer_size], size_t *length) {
	// Character array used for conversion.
	static const char chars[] = "0123456789abcdefghijklmnopqrstuvwxyz";

	size_t integer_cursor = radix_buffer_size / 2;
	size_t fraction_cursor = integer_cursor;

	bool negative = value < 0;
	if (negative) {
		value = -value;
	}

	// Split the value into an integer part and a fractional part.
	double integer = floor(value);
	double fraction = value - integer;
	// We only compute fractional digits up to the input double's precision.
	double delta = 0.5 * (radix_next_double(value) - value);
	// If the delta rounded down to zero, use the minimum (denormal) delta value. (V8 skips the loop
	// instead when the processor flushes denormals to zero, which a native Adamic program never asks
	// it to.)
	if (delta <= 0) {
		delta = radix_next_double(0.0);
	}
	if (fraction >= delta) {
		// Insert decimal point.
		buffer[fraction_cursor++] = '.';
		do {
			// Shift up by one digit.
			fraction *= radix;
			delta *= radix;
			// Write digit.
			int digit = (int)fraction;
			buffer[fraction_cursor++] = chars[digit];
			// Calculate remainder.
			fraction -= digit;
			// Round to even.
			if (fraction > 0.5 || (fraction == 0.5 && (digit & 1))) {
				if (fraction + delta > 1) {
					// We need to back trace already written digits in case of carry-over.
					while (true) {
						fraction_cursor--;
						if (fraction_cursor == radix_buffer_size / 2) {
							// Carry over to the integer part.
							integer += 1;
							break;
						}
						char c = buffer[fraction_cursor];
						// Reconstruct digit.
						digit = c > '9' ? (c - 'a' + 10) : (c - '0');
						if (digit + 1 < radix) {
							buffer[fraction_cursor++] = chars[digit + 1];
							break;
						}
					}
					break;
				}
			}
		} while (fraction >= delta);
	}

	// Compute integer digits. Fill unrepresented digits with zero.
	while (radix_double_exponent(integer / radix) > 0) {
		integer /= radix;
		buffer[--integer_cursor] = '0';
	}
	do {
		// V8's Modulo is fmod, away from Windows.
		double remainder = fmod(integer, radix);
		buffer[--integer_cursor] = chars[(int)remainder];
		integer = (integer - remainder) / radix;
	} while (integer > 0);

	// Add sign and terminate string.
	if (negative) {
		buffer[--integer_cursor] = '-';
	}
	*length = fraction_cursor - integer_cursor;
	return buffer + integer_cursor;
}

static adamic_string *radix_result(const char *text, size_t length) {
	adamic_string result = {{0, adamic_kind_string, 0}, length, text, 0, NULL, NULL, 0};
	return adamic_string_concat(1, (adamic_string *const[]){&result});
}

// NumberPrototypeToString with a radix: the radix is checked first, before NaN (unlike
// toExponential), so (NaN).toString(1) is a RangeError.
adamic_string *adamic_number_to_radix(double value, double radix) {
	// ToIntegerOrInfinity: NaN is 0, and a fraction truncates.
	radix = isnan(radix) ? 0 : trunc(radix);
	if (radix < 2 || radix > 36) {
		static const char message[] = "toString() radix argument must be between 2 and 36";
		adamic_library_throw("RangeError", message, sizeof message - 1);
		return NULL;
	}
	if (radix == 10) {
		return adamic_string_from_number(value);
	}
	if (value == 0) {
		return radix_result("0", 1);
	}
	if (isnan(value)) {
		return radix_result("NaN", 3);
	}
	if (isinf(value)) {
		return value < 0 ? radix_result("-Infinity", 9) : radix_result("Infinity", 8);
	}
	char buffer[radix_buffer_size];
	size_t length;
	const char *text = double_to_radix(value, (int)radix, buffer, &length);
	return radix_result(text, length);
}
