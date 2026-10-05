// number.c: numbers as text, exactly as JavaScript writes them.

#include "adamic.h"

#include <math.h>
#include <stdio.h>
#include <stdlib.h>

// shortest_digits finds the fewest significant digits that read back as value (positive and
// finite), and the decimal exponent of the first one: value is 0.d1d2d3... times 10^point.
//
// printf's %e rounds correctly, so its p-digit answer is the p-digit decimal closest to value, and
// on an exact tie it takes the even one. The first p that round-trips through strtod is therefore
// ECMAScript's choice: as few digits as possible, then the closest, then the even (ECMA-262,
// Number::toString). That's plain rather than fast; the oracle's sweep is what says it's right, and
// a faster algorithm has to pass the same sweep before it replaces this.
static int shortest_digits(double value, char digits[18], int *point) {
	char scientific[40];
	for (int precision = 1; precision <= 17; precision++) {
		snprintf(scientific, sizeof scientific, "%.*e", precision - 1, value);
		if (strtod(scientific, NULL) == value) {
			break;
		}
	}
	int count = 0;
	const char *cursor = scientific;
	digits[count++] = *cursor++;
	if (*cursor == '.') {
		cursor++;
		while (*cursor != 'e') {
			digits[count++] = *cursor++;
		}
	}
	*point = atoi(cursor + 1) + 1;
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
