// number.c: numbers as text, exactly as JavaScript writes them.

#include "adamic.h"
#include "library_errors.h"

#include <stdint.h>

#include <math.h>
#include <stdio.h>
#include <stdlib.h>

static size_t append(char *buffer, size_t length, const char *text) {
	while (*text != '\0') {
		buffer[length++] = *text++;
	}
	return length;
}

size_t adamic_number_format(double value, char buffer[ADAMIC_NUMBER_FORMAT_MAX]) {
	if (isnan(value)) {
		return append(buffer, 0, "NaN");
	}
	if (value == 0) {
		// Both zeros: String(-0) is "0".
		return append(buffer, 0, "0");
	}
	size_t length = 0;
	if (value < 0) {
		buffer[length++] = '-';
		value = -value;
	}
	if (isinf(value)) {
		return append(buffer, length, "Infinity");
	}
	if (value < 9007199254740992.0 && value == (double)(uint64_t)value) {
		// A whole number below 2^53: every integer there is a double, and so are its neighbours, so no
		// shorter digits read back as it, and its shortest form is its own digits, written directly
		// rather than searched for. 2^53 and past, 2^60 is 1152921504606847000, so those search.
		char reversed[16];
		int count = 0;
		uint64_t whole = (uint64_t)value;
		do {
			reversed[count++] = (char)('0' + whole % 10);
			whole /= 10;
		} while (whole != 0);
		while (count > 0) {
			buffer[length++] = reversed[--count];
		}
		return length;
	}

	char digits[18];
	int point;
	// The fewest digits that read back, and the closest of those, from V8's own algorithm (dtoa.c).
	// Asking printf for the closest p-digit decimal at each p, as this once did, isn't the same:
	// just below a power of two the doubles are twice as dense, and the closest 16-digit decimal to
	// 2 ** 976 doesn't read back where a farther one does, so it wrote 17 digits where Node writes 16.
	int count = adamic_number_shortest_digits(value, digits, &point);

	if (count <= point && point <= 21) {
		// An integer: the digits, then zeros out to the point.
		for (int index = 0; index < count; index++) {
			buffer[length++] = digits[index];
		}
		for (int index = count; index < point; index++) {
			buffer[length++] = '0';
		}
	} else if (0 < point && point <= 21) {
		// The point falls inside the digits.
		for (int index = 0; index < count; index++) {
			if (index == point) {
				buffer[length++] = '.';
			}
			buffer[length++] = digits[index];
		}
	} else if (-6 < point && point <= 0) {
		// Small: "0.", zeros, then the digits.
		length = append(buffer, length, "0.");
		for (int index = point; index < 0; index++) {
			buffer[length++] = '0';
		}
		for (int index = 0; index < count; index++) {
			buffer[length++] = digits[index];
		}
	} else {
		// Exponential: d.ddd, then e, a sign always, and the exponent.
		buffer[length++] = digits[0];
		if (count > 1) {
			buffer[length++] = '.';
			for (int index = 1; index < count; index++) {
				buffer[length++] = digits[index];
			}
		}
		int exponent = point - 1;
		buffer[length++] = 'e';
		buffer[length++] = exponent < 0 ? '-' : '+';
		length += (size_t)snprintf(buffer + length, ADAMIC_NUMBER_FORMAT_MAX - length, "%d", abs(exponent));
	}
	return length;
}

double adamic_power(double base, double exponent) {
	// This is V8's math::pow (src/numbers/ieee754.cc), case for case, so ** gives Node's bits and not
	// just a close answer. Where JavaScript and C disagree (ECMA-262, Number::exponentiate), a NaN
	// exponent is NaN even for base 1, and a base of magnitude 1 to an infinite power is NaN.
	if (isnan(exponent)) {
		return NAN;
	}
	if (isinf(exponent) && (base == 1 || base == -1)) {
		return NAN;
	}
	// V8 special-cases these to match its optimizing compilers, and its answers differ from the C
	// library's: for x ** 0.5, macOS's pow and sqrt disagree in the last bit (found by the sweep in
	// math_test.go, 4503599627370495.5 ** 0.5).
	if (exponent == 2) {
		return base * base;
	}
	if (exponent == 0.5) {
		// +0 so that -0 ** 0.5 is +0; and -Infinity ** 0.5 is Infinity, where sqrt says NaN.
		return isinf(base) ? INFINITY : sqrt(base + 0);
	}
	// Then V8 calls the platform's pow, the same library a native Adamic program links on the same
	// platform.
	return pow(base, exponent);
}

adamic_string *adamic_number_to_fixed(double value, double digits) {
	// ToIntegerOrInfinity first: NaN is 0, and a fraction truncates, so toFixed(-0.5) is toFixed(0).
	digits = isnan(digits) ? 0 : trunc(digits);
	if (!(digits >= 0 && digits <= 100)) {
		static const char message[] = "toFixed() digits argument must be between 0 and 100";
		adamic_library_throw("RangeError", message, sizeof message - 1);
		return NULL;
	}
	int places = (int)digits;
	if (isnan(value)) {
		return adamic_string_concat(1, (adamic_string *const[]){&(adamic_string)ADAMIC_STRING("NaN")});
	}
	if (fabs(value) >= 1e21) {
		return adamic_string_from_number(value);
	}
	// The exact decimal value: every double has a finite expansion, at most 1074 digits after the
	// point, and printf writes it exactly when asked for that many. JavaScript then rounds half up on
	// the exact value (the larger n on a tie), where printf's own rounding would round half to even.
	char exact[1500];
	bool negative = value < 0;
	int length = snprintf(exact, sizeof exact, "%.1100f", fabs(value));
	int point = 0;
	while (exact[point] != '.') {
		point++;
	}
	// Keep the integer digits and places fraction digits, then round on the next one.
	int keep = point + 1 + places;
	bool up = keep < length && exact[keep] >= '5';
	char digits_text[1500];
	int count = 0;
	for (int index = 0; index < keep; index++) {
		if (exact[index] != '.') {
			digits_text[count++] = exact[index];
		}
	}
	if (up) {
		int index = count - 1;
		while (index >= 0 && digits_text[index] == '9') {
			digits_text[index--] = '0';
		}
		if (index >= 0) {
			digits_text[index]++;
		} else {
			// Every digit carried: 9.99 becomes 10.00, one digit longer.
			for (int move = count; move > 0; move--) {
				digits_text[move] = digits_text[move - 1];
			}
			digits_text[0] = '1';
			count++;
			point++;
		}
	}
	char result[1600];
	int written = 0;
	// -0 and values that round to zero keep the sign JavaScript gives them: only a value below zero
	// is written with a minus, and (-0.001).toFixed(2) is "-0.00".
	if (negative) {
		result[written++] = '-';
	}
	int integer_digits = count - places;
	for (int index = 0; index < count; index++) {
		if (index == integer_digits) {
			result[written++] = '.';
		}
		result[written++] = digits_text[index];
	}
	adamic_string text = {{0, adamic_kind_string, 0}, (size_t)written, result, 0, NULL, NULL, 0};
	return adamic_string_concat(1, (adamic_string *const[]){&text});
}
