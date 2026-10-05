// number.c: numbers as text, exactly as JavaScript writes them.

#include "adamic.h"

#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// shortest_digits finds the fewest significant digits that read back as value (positive and
// finite), and the decimal exponent of the first one: value is 0.d1d2d3... times 10^point.
//
// ECMAScript (Number::toString) wants as few digits as read back, then of those the closest, then
// the even. printf's %e rounds correctly, so its p-digit answer is the closest p digits, ties to
// even. When it reads back, nothing of p digits is closer, and it's the answer. When it doesn't,
// another p-digit decimal still may: the one on value's other side, one unit away in the last digit.
// That happens where the doubles' spacing changes, at a power of two, where the double below is
// half as far as the double above (2 ** 89 is 6.189700196426902e+26, whose closest 16 digits read back
// as the double below). That's plain rather than fast; the oracle's sweep is what says it's right,
// and a faster algorithm has to pass the same sweep before it replaces this.
static int shortest_digits(double value, char digits[18], int *point) {
	char scientific[40];
	int count = 0, exponent = 0;
	for (int precision = 1; precision <= 17; precision++) {
		snprintf(scientific, sizeof scientific, "%.*e", precision - 1, value);
		count = 0;
		const char *cursor = scientific;
		digits[count++] = *cursor++;
		if (*cursor == '.') {
			cursor++;
			while (*cursor != 'e') {
				digits[count++] = *cursor++;
			}
		}
		exponent = atoi(cursor + 1);
		if (strtod(scientific, NULL) == value) {
			break;
		}
		// The neighbor across value: up a unit in the last digit when the closest read back low, down
		// a unit when it read back high, carrying or borrowing through the digits.
		char neighbor[18];
		int neighbor_exponent = exponent;
		memcpy(neighbor, digits, (size_t)count);
		if (strtod(scientific, NULL) < value) {
			int at = count - 1;
			while (at >= 0 && neighbor[at] == '9') {
				neighbor[at--] = '0';
			}
			if (at < 0) {
				// 9.99 up a unit is 10.0, written 1.00 a power of ten higher.
				neighbor[0] = '1';
				neighbor_exponent++;
			} else {
				neighbor[at]++;
			}
		} else {
			int at = count - 1;
			while (at >= 0 && neighbor[at] == '0') {
				neighbor[at--] = '9';
			}
			neighbor[at]--;
			if (neighbor[0] == '0') {
				// Below a power of ten the p-digit decimals are ten times closer: the one under 1.00
				// is 9.99 a power of ten lower.
				memset(neighbor, '9', (size_t)count);
				neighbor_exponent--;
			}
		}
		char candidate[40];
		int length = 0;
		candidate[length++] = neighbor[0];
		if (count > 1) {
			candidate[length++] = '.';
			memcpy(candidate + length, neighbor + 1, (size_t)count - 1);
			length += count - 1;
		}
		snprintf(candidate + length, sizeof candidate - (size_t)length, "e%d", neighbor_exponent);
		if (strtod(candidate, NULL) == value) {
			memcpy(digits, neighbor, (size_t)count);
			exponent = neighbor_exponent;
			break;
		}
	}
	// Trailing zeros aren't digits of the shortest form: 1.00e6 is 1e6.
	while (count > 1 && digits[count - 1] == '0') {
		count--;
	}
	*point = exponent + 1;
	return count;
}

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

	char digits[18];
	int point;
	int count = shortest_digits(value, digits, &point);

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
		static const char message[] = "RangeError: toFixed() digits argument must be between 0 and 100";
		adamic_panic(message, sizeof message - 1);
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
	adamic_string text = {{0, adamic_kind_string}, (size_t)written, result, 0, NULL};
	return adamic_string_concat(1, (adamic_string *const[]){&text});
}
