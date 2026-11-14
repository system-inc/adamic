// parse.c: Number.parseInt and Number.parseFloat, exactly as Node answers them.
//
// parseInt follows V8 (src/numbers/conversions.cc: StringToIntHelper::DetectRadixInternal,
// NumberParseIntHelper and InternalStringToIntDouble; copyright the V8 project authors, BSD-style
// license). Radix 10 goes through strtod and is exact; a power of two is rounded exactly; every other
// radix accumulates as V8 does, which ECMAScript lets an implementation approximate past 2^53 and
// Adamic must therefore match rather than improve. The sweep in parse_test.go holds both to Node.

#include "adamic.h"

#include <math.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

// to_int32 is ECMAScript's ToInt32: NaN and the infinities are 0, a fraction truncates, and the rest
// wraps modulo 2^32 into a signed 32-bit integer.
static int32_t to_int32(double value) {
	if (!isfinite(value)) {
		return 0;
	}
	double wrapped = fmod(trunc(value), 4294967296.0);
	if (wrapped < 0) {
		wrapped += 4294967296.0;
	}
	if (wrapped >= 2147483648.0) {
		wrapped -= 4294967296.0;
	}
	return (int32_t)wrapped;
}

static bool is_digit(unsigned char character, int radix) {
	return (character >= '0' && character <= '9' && character < '0' + radix) ||
		(radix > 10 && character >= 'a' && character < 'a' + radix - 10) ||
		(radix > 10 && character >= 'A' && character < 'A' + radix - 10);
}

static int digit_value(unsigned char character) {
	if (character >= '0' && character <= '9') {
		return character - '0';
	}
	if (character >= 'a' && character <= 'z') {
		return character - 'a' + 10;
	}
	return character - 'A' + 10;
}

static double signed_zero(bool negative) {
	return negative ? -0.0 : 0.0;
}

// base_ten is V8's HandleBaseTenCase: the decimal digits through strtod, which rounds correctly. More
// than 309 significant digits is past the largest double, so only that many are kept.
static double base_ten(const unsigned char *current, const unsigned char *end) {
	char buffer[311];
	size_t length = 0;
	while (current != end && *current >= '0' && *current <= '9') {
		if (length <= 309) {
			buffer[length++] = (char)*current;
		}
		current++;
	}
	buffer[length] = '\0';
	return strtod(buffer, NULL);
}

// power_of_two is V8's InternalStringToIntDouble, trailing junk allowed: the digits gathered into 53
// bits, the rest rounded half to even, a nonzero tail counting as above half.
static double power_of_two(const unsigned char *current, const unsigned char *end, int radix_log_2) {
	int radix = 1 << radix_log_2;
	while (*current == '0') {
		current++;
		if (current == end) {
			return 0;
		}
	}
	int64_t number = 0;
	int exponent = 0;
	do {
		if (!is_digit(*current, radix)) {
			break;
		}
		number = number * radix + digit_value(*current);
		int overflow = (int)(number >> 53);
		if (overflow != 0) {
			int overflow_bits_count = 1;
			while (overflow > 1) {
				overflow_bits_count++;
				overflow >>= 1;
			}
			int dropped_bits_mask = (1 << overflow_bits_count) - 1;
			int dropped_bits = (int)number & dropped_bits_mask;
			number >>= overflow_bits_count;
			exponent = overflow_bits_count;
			bool zero_tail = true;
			for (;;) {
				current++;
				if (current == end || !is_digit(*current, radix)) {
					break;
				}
				zero_tail = zero_tail && *current == '0';
				// Capped, so a huge run of digits can't overflow the exponent.
				if (exponent <= 1024) {
					exponent += radix_log_2;
				}
			}
			int middle_value = 1 << (overflow_bits_count - 1);
			if (dropped_bits > middle_value || (dropped_bits == middle_value && ((number & 1) != 0 || !zero_tail))) {
				number++;
			}
			if ((number & ((int64_t)1 << 53)) != 0) {
				exponent++;
				number >>= 1;
			}
			break;
		}
		current++;
	} while (current != end);
	if (exponent == 0) {
		return (double)number;
	}
	return ldexp((double)number, exponent);
}

// generic is V8's HandleGenericCase, for every other radix: parts of at most 32 bits, each folded
// into a double, which loses precision past about 2^56 exactly as V8 does.
static double generic(const unsigned char *current, const unsigned char *end, int radix) {
	double result = 0;
	bool done = false;
	do {
		uint32_t part = 0, multiplier = 1;
		for (;;) {
			if (!is_digit(*current, radix)) {
				done = true;
				break;
			}
			const uint32_t maximum_multiplier = UINT32_MAX / 36;
			uint32_t next = multiplier * (uint32_t)radix;
			if (next > maximum_multiplier) {
				break;
			}
			part = part * (uint32_t)radix + (uint32_t)digit_value(*current);
			multiplier = next;
			current++;
			if (current == end) {
				done = true;
				break;
			}
		}
		result = result * multiplier + part;
	} while (!done);
	return result;
}

double adamic_number_parse_int(const adamic_string *text, double radix_value) {
	int radix = to_int32(radix_value);
	if (radix != 0 && (radix < 2 || radix > 36)) {
		return NAN;
	}
	adamic_string *trimmed = adamic_string_trim_sides((adamic_string *)text, true, false);
	const unsigned char *current = (const unsigned char *)trimmed->bytes;
	const unsigned char *end = current + trimmed->length;
	double result = NAN;
	bool negative = false;
	bool leading_zero = false;
	// DetectRadixInternal: a sign, then 0x when the radix is 0 or 16, then leading zeros.
	if (current == end) {
		goto done;
	}
	if (*current == '+' || *current == '-') {
		negative = *current == '-';
		if (++current == end) {
			goto done;
		}
	}
	if ((radix == 0 || radix == 16) && *current == '0') {
		if (++current == end) {
			result = signed_zero(negative);
			goto done;
		}
		if (*current == 'x' || *current == 'X') {
			radix = 16;
			if (++current == end) {
				goto done;
			}
		} else {
			leading_zero = true;
		}
	}
	if (radix == 0) {
		radix = 10;
	}
	while (*current == '0') {
		leading_zero = true;
		if (++current == end) {
			result = signed_zero(negative);
			goto done;
		}
	}
	if (!is_digit(*current, radix)) {
		if (leading_zero) {
			result = signed_zero(negative);
		}
		goto done;
	}
	if (radix == 10) {
		result = base_ten(current, end);
	} else if ((radix & (radix - 1)) == 0) {
		int radix_log_2 = 0;
		while ((1 << radix_log_2) < radix) {
			radix_log_2++;
		}
		result = power_of_two(current, end, radix_log_2);
	} else {
		result = generic(current, end, radix);
	}
	result = negative ? -result : result;
done:
	adamic_release(trimmed);
	return result;
}

double adamic_number_parse_float(const adamic_string *text) {
	// The longest prefix, after leading space, that is a decimal literal (StrDecimalLiteral, with no
	// separators): a sign, then Infinity, or digits with perhaps a fraction and an exponent. An
	// exponent with no digits isn't part of it, so "1e" is 1. strtod rounds what's found correctly.
	adamic_string *trimmed = adamic_string_trim_sides((adamic_string *)text, true, false);
	const char *start = trimmed->bytes, *end = trimmed->bytes + trimmed->length;
	const char *cursor = start;
	double result = NAN;
	bool negative = false;
	if (cursor != end && (*cursor == '+' || *cursor == '-')) {
		negative = *cursor == '-';
		cursor++;
	}
	if (end - cursor >= 8 && memcmp(cursor, "Infinity", 8) == 0) {
		result = negative ? -INFINITY : INFINITY;
		adamic_release(trimmed);
		return result;
	}
	size_t digits = 0;
	while (cursor != end && *cursor >= '0' && *cursor <= '9') {
		cursor++;
		digits++;
	}
	if (cursor != end && *cursor == '.') {
		cursor++;
		while (cursor != end && *cursor >= '0' && *cursor <= '9') {
			cursor++;
			digits++;
		}
	}
	if (digits > 0) {
		if (cursor != end && (*cursor == 'e' || *cursor == 'E')) {
			const char *exponent = cursor + 1;
			if (exponent != end && (*exponent == '+' || *exponent == '-')) {
				exponent++;
			}
			if (exponent != end && *exponent >= '0' && *exponent <= '9') {
				while (exponent != end && *exponent >= '0' && *exponent <= '9') {
					exponent++;
				}
				cursor = exponent;
			}
		}
		size_t length = (size_t)(cursor - start);
		char *literal = malloc(length + 1);
		if (literal == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		memcpy(literal, start, length);
		literal[length] = '\0';
		result = strtod(literal, NULL);
		free(literal);
	}
	adamic_release(trimmed);
	return result;
}
