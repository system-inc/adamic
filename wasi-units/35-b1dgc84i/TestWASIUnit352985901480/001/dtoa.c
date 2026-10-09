// Copyright 2011 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// The notice above is V8's, from the top of src/base/numbers/bignum.cc, bignum-dtoa.cc and dtoa.cc
// and src/numbers/conversions.cc (each carries the same one), kept verbatim as its license asks.
// The two builtins at the end are from src/builtins/builtins-number.cc, whose notice is:
//
// Copyright 2016 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// dtoa.c: number.toExponential(digits) and number.toPrecision(digits), byte for byte as Node writes
// them (V8 13.6.233.17, the V8 in Node 24).
//
// The digits come from V8's BignumDtoa: exact arithmetic on big integers, so a digit is never a
// guess, and the last one rounds half up on the exact value, as V8 does (printf would round half to
// even: (2.5).toPrecision(1) is "3" in JavaScript and "2" from printf). V8's DoubleToAscii first
// tries FastDtoa, which answers only when it can prove its digits are the exact ones, and falls back
// to BignumDtoa when it can't; so BignumDtoa alone gives the same digits, and that's what's here.
// The C follows V8's functions one for one, under V8's names in snake case, so each can be read
// against its original; Bignum's methods take the bignum as their first argument.

#include "adamic.h"

#include <math.h>
#include <stdint.h>
#include <string.h>

// V8's kMaxFractionDigits (src/numbers/conversions.h): the most digits toExponential, toFixed and
// toPrecision will write.
enum { max_fraction_digits = 100 };

// ---- src/base/numbers/double.h: the parts of class Double that BignumDtoa uses.

enum {
	double_physical_significand_size = 52,
	double_exponent_bias = 0x3FF + double_physical_significand_size,
	double_denormal_exponent = -double_exponent_bias + 1,
};
static const uint64_t double_significand_mask = 0x000FFFFFFFFFFFFFu;
static const uint64_t double_exponent_mask = 0x7FF0000000000000u;
static const uint64_t double_hidden_bit = 0x0010000000000000u;

static uint64_t double_bits(double value) {
	uint64_t bits;
	memcpy(&bits, &value, sizeof bits);
	return bits;
}

static bool double_is_denormal(double value) {
	return (double_bits(value) & double_exponent_mask) == 0;
}

static int double_exponent(double value) {
	if (double_is_denormal(value)) {
		return double_denormal_exponent;
	}
	int biased = (int)((double_bits(value) & double_exponent_mask) >> double_physical_significand_size);
	return biased - double_exponent_bias;
}

static uint64_t double_significand(double value) {
	uint64_t significand = double_bits(value) & double_significand_mask;
	return double_is_denormal(value) ? significand : significand + double_hidden_bit;
}

// ---- src/base/numbers/bignum.h and bignum.cc: class Bignum, the methods BignumDtoa uses. (The
// rest, AddBignum, MultiplyByPowerOfTen and the string conversions, BignumDtoa never calls.)

// 3584 = 128 * 28. We can represent 2^3584 > 10^1000 accurately.
enum {
	bignum_max_significant_bits = 3584,
	bignum_chunk_size = 32,
	// With bigit size of 28 we loose some bits, but a double still fits easily into two chunks, and
	// more importantly we can use the Comba multiplication.
	bignum_bigit_size = 28,
	bignum_bigit_capacity = bignum_max_significant_bits / bignum_bigit_size,
};
static const uint32_t bignum_bigit_mask = (1u << bignum_bigit_size) - 1;

// The value is bigits times 2^(exponent * bignum_bigit_size).
typedef struct bignum {
	uint32_t bigits[bignum_bigit_capacity];
	int used_digits;
	int exponent;
} bignum;

static void bignum_init(bignum *number) {
	memset(number, 0, sizeof *number);
}

// Bignums cannot grow: V8 sizes them for every double, and so stops if one ever would.
static void bignum_ensure_capacity(int size) {
	if (size > bignum_bigit_capacity) {
		adamic_unreachable();
	}
}

static int bignum_bigit_length(const bignum *number) {
	return number->used_digits + number->exponent;
}

static void bignum_zero(bignum *number) {
	for (int i = 0; i < number->used_digits; ++i) {
		number->bigits[i] = 0;
	}
	number->used_digits = 0;
	number->exponent = 0;
}

static void bignum_clamp(bignum *number) {
	while (number->used_digits > 0 && number->bigits[number->used_digits - 1] == 0) {
		number->used_digits--;
	}
	if (number->used_digits == 0) {
		number->exponent = 0;
	}
}

static void bignum_assign_uint16(bignum *number, uint16_t value) {
	bignum_zero(number);
	if (value == 0) {
		return;
	}
	bignum_ensure_capacity(1);
	number->bigits[0] = value;
	number->used_digits = 1;
}

static void bignum_assign_uint64(bignum *number, uint64_t value) {
	const int uint64_size = 64;
	bignum_zero(number);
	if (value == 0) {
		return;
	}
	int needed_bigits = uint64_size / bignum_bigit_size + 1;
	bignum_ensure_capacity(needed_bigits);
	for (int i = 0; i < needed_bigits; ++i) {
		number->bigits[i] = (uint32_t)(value & bignum_bigit_mask);
		value = value >> bignum_bigit_size;
	}
	number->used_digits = needed_bigits;
	bignum_clamp(number);
}

static void bignum_assign_bignum(bignum *number, const bignum *other) {
	number->exponent = other->exponent;
	for (int i = 0; i < other->used_digits; ++i) {
		number->bigits[i] = other->bigits[i];
	}
	for (int i = other->used_digits; i < number->used_digits; ++i) {
		number->bigits[i] = 0;
	}
	number->used_digits = other->used_digits;
}

static void bignum_align(bignum *number, const bignum *other) {
	if (number->exponent > other->exponent) {
		int zero_digits = number->exponent - other->exponent;
		bignum_ensure_capacity(number->used_digits + zero_digits);
		for (int i = number->used_digits - 1; i >= 0; --i) {
			number->bigits[i + zero_digits] = number->bigits[i];
		}
		for (int i = 0; i < zero_digits; ++i) {
			number->bigits[i] = 0;
		}
		number->used_digits += zero_digits;
		number->exponent -= zero_digits;
	}
}

static void bignum_bigits_shift_left(bignum *number, int shift_amount) {
	uint32_t carry = 0;
	for (int i = 0; i < number->used_digits; ++i) {
		uint32_t new_carry = number->bigits[i] >> (bignum_bigit_size - shift_amount);
		number->bigits[i] = ((number->bigits[i] << shift_amount) + carry) & bignum_bigit_mask;
		carry = new_carry;
	}
	if (carry != 0) {
		number->bigits[number->used_digits] = carry;
		number->used_digits++;
	}
}

static void bignum_shift_left(bignum *number, int shift_amount) {
	if (number->used_digits == 0) {
		return;
	}
	number->exponent += shift_amount / bignum_bigit_size;
	int local_shift = shift_amount % bignum_bigit_size;
	bignum_ensure_capacity(number->used_digits + 1);
	bignum_bigits_shift_left(number, local_shift);
}

// Precondition: number >= other.
static void bignum_subtract_bignum(bignum *number, const bignum *other) {
	bignum_align(number, other);
	int offset = other->exponent - number->exponent;
	uint32_t borrow = 0;
	int i;
	for (i = 0; i < other->used_digits; ++i) {
		uint32_t difference = number->bigits[i + offset] - other->bigits[i] - borrow;
		number->bigits[i + offset] = difference & bignum_bigit_mask;
		borrow = difference >> (bignum_chunk_size - 1);
	}
	while (borrow != 0) {
		uint32_t difference = number->bigits[i + offset] - borrow;
		number->bigits[i + offset] = difference & bignum_bigit_mask;
		borrow = difference >> (bignum_chunk_size - 1);
		++i;
	}
	bignum_clamp(number);
}

static void bignum_multiply_by_uint32(bignum *number, uint32_t factor) {
	if (factor == 1) {
		return;
	}
	if (factor == 0) {
		bignum_zero(number);
		return;
	}
	if (number->used_digits == 0) {
		return;
	}
	uint64_t carry = 0;
	for (int i = 0; i < number->used_digits; ++i) {
		uint64_t product = (uint64_t)factor * number->bigits[i] + carry;
		number->bigits[i] = (uint32_t)(product & bignum_bigit_mask);
		carry = product >> bignum_bigit_size;
	}
	while (carry != 0) {
		bignum_ensure_capacity(number->used_digits + 1);
		number->bigits[number->used_digits] = (uint32_t)(carry & bignum_bigit_mask);
		number->used_digits++;
		carry >>= bignum_bigit_size;
	}
}

static void bignum_times_10(bignum *number) {
	bignum_multiply_by_uint32(number, 10);
}

static void bignum_multiply_by_uint64(bignum *number, uint64_t factor) {
	if (factor == 1) {
		return;
	}
	if (factor == 0) {
		bignum_zero(number);
		return;
	}
	uint64_t carry = 0;
	uint64_t low = factor & 0xFFFFFFFF;
	uint64_t high = factor >> 32;
	for (int i = 0; i < number->used_digits; ++i) {
		uint64_t product_low = low * number->bigits[i];
		uint64_t product_high = high * number->bigits[i];
		uint64_t tmp = (carry & bignum_bigit_mask) + product_low;
		number->bigits[i] = (uint32_t)(tmp & bignum_bigit_mask);
		carry = (carry >> bignum_bigit_size) + (tmp >> bignum_bigit_size) + (product_high << (32 - bignum_bigit_size));
	}
	while (carry != 0) {
		bignum_ensure_capacity(number->used_digits + 1);
		number->bigits[number->used_digits] = (uint32_t)(carry & bignum_bigit_mask);
		number->used_digits++;
		carry >>= bignum_bigit_size;
	}
}

static void bignum_square(bignum *number) {
	int product_length = 2 * number->used_digits;
	bignum_ensure_capacity(product_length);
	// Comba multiplication: compute each column separately. V8 stops if the accumulator could
	// overflow, which takes more digits than a bignum can have.
	if ((1 << (2 * (bignum_chunk_size - bignum_bigit_size))) <= number->used_digits) {
		adamic_unreachable();
	}
	uint64_t accumulator = 0;
	// First shift the digits so we don't overwrite them.
	int copy_offset = number->used_digits;
	for (int i = 0; i < number->used_digits; ++i) {
		number->bigits[copy_offset + i] = number->bigits[i];
	}
	// We have two loops to avoid some 'if's in the loop.
	for (int i = 0; i < number->used_digits; ++i) {
		int bigit_index1 = i;
		int bigit_index2 = 0;
		while (bigit_index1 >= 0) {
			uint32_t chunk1 = number->bigits[copy_offset + bigit_index1];
			uint32_t chunk2 = number->bigits[copy_offset + bigit_index2];
			accumulator += (uint64_t)chunk1 * chunk2;
			bigit_index1--;
			bigit_index2++;
		}
		number->bigits[i] = (uint32_t)accumulator & bignum_bigit_mask;
		accumulator >>= bignum_bigit_size;
	}
	for (int i = number->used_digits; i < product_length; ++i) {
		int bigit_index1 = number->used_digits - 1;
		int bigit_index2 = i - bigit_index1;
		while (bigit_index2 < number->used_digits) {
			uint32_t chunk1 = number->bigits[copy_offset + bigit_index1];
			uint32_t chunk2 = number->bigits[copy_offset + bigit_index2];
			accumulator += (uint64_t)chunk1 * chunk2;
			bigit_index1--;
			bigit_index2++;
		}
		number->bigits[i] = (uint32_t)accumulator & bignum_bigit_mask;
		accumulator >>= bignum_bigit_size;
	}
	number->used_digits = product_length;
	number->exponent *= 2;
	bignum_clamp(number);
}

static void bignum_assign_power_uint16(bignum *number, uint16_t base, int power_exponent) {
	if (power_exponent == 0) {
		bignum_assign_uint16(number, 1);
		return;
	}
	bignum_zero(number);
	int shifts = 0;
	// We expect base to be in range 2-32, and most often to be 10. It does not make much sense to
	// implement different algorithms for counting the bits.
	while ((base & 1) == 0) {
		base >>= 1;
		shifts++;
	}
	int bit_size = 0;
	int tmp_base = base;
	while (tmp_base != 0) {
		tmp_base >>= 1;
		bit_size++;
	}
	int final_size = bit_size * power_exponent;
	// 1 extra bigit for the shifting, and one for rounded final_size.
	bignum_ensure_capacity(final_size / bignum_bigit_size + 2);

	// Left to Right exponentiation.
	int mask = 1;
	while (power_exponent >= mask) {
		mask <<= 1;
	}
	// The mask is now pointing to the bit above the most significant 1-bit of power_exponent.
	// Get rid of first 1-bit;
	mask >>= 2;
	uint64_t this_value = base;

	bool delayed_multipliciation = false;
	const uint64_t max_32bits = 0xFFFFFFFF;
	while (mask != 0 && this_value <= max_32bits) {
		this_value = this_value * this_value;
		// Verify that there is enough space in this_value to perform the multiplication. The first
		// bit_size bits must be 0.
		if ((power_exponent & mask) != 0) {
			uint64_t base_bits_mask = ~(((uint64_t)1 << (64 - bit_size)) - 1);
			bool high_bits_zero = (this_value & base_bits_mask) == 0;
			if (high_bits_zero) {
				this_value *= base;
			} else {
				delayed_multipliciation = true;
			}
		}
		mask >>= 1;
	}
	bignum_assign_uint64(number, this_value);
	if (delayed_multipliciation) {
		bignum_multiply_by_uint32(number, base);
	}

	// Now do the same thing as a bignum.
	while (mask != 0) {
		bignum_square(number);
		if ((power_exponent & mask) != 0) {
			bignum_multiply_by_uint32(number, base);
		}
		mask >>= 1;
	}

	// And finally add the saved shifts.
	bignum_shift_left(number, shifts * power_exponent);
}

static uint32_t bignum_bigit_at(const bignum *number, int index) {
	if (index >= bignum_bigit_length(number)) {
		return 0;
	}
	if (index < number->exponent) {
		return 0;
	}
	return number->bigits[index - number->exponent];
}

static int bignum_compare(const bignum *a, const bignum *b) {
	int bigit_length_a = bignum_bigit_length(a);
	int bigit_length_b = bignum_bigit_length(b);
	if (bigit_length_a < bigit_length_b) {
		return -1;
	}
	if (bigit_length_a > bigit_length_b) {
		return +1;
	}
	int lowest = a->exponent < b->exponent ? a->exponent : b->exponent;
	for (int i = bigit_length_a - 1; i >= lowest; --i) {
		uint32_t bigit_a = bignum_bigit_at(a, i);
		uint32_t bigit_b = bignum_bigit_at(b, i);
		if (bigit_a < bigit_b) {
			return -1;
		}
		if (bigit_a > bigit_b) {
			return +1;
		}
		// Otherwise they are equal up to this digit. Try the next digit.
	}
	return 0;
}

static bool bignum_equal(const bignum *a, const bignum *b) {
	return bignum_compare(a, b) == 0;
}

static bool bignum_less_equal(const bignum *a, const bignum *b) {
	return bignum_compare(a, b) <= 0;
}

static bool bignum_less(const bignum *a, const bignum *b) {
	return bignum_compare(a, b) < 0;
}

// Returns compare(a + b, c).
static int bignum_plus_compare(const bignum *a, const bignum *b, const bignum *c) {
	if (bignum_bigit_length(a) < bignum_bigit_length(b)) {
		return bignum_plus_compare(b, a, c);
	}
	if (bignum_bigit_length(a) + 1 < bignum_bigit_length(c)) {
		return -1;
	}
	if (bignum_bigit_length(a) > bignum_bigit_length(c)) {
		return +1;
	}
	// The exponent encodes 0-bigits. So if there are more 0-digits in 'a' than 'b' has digits, then
	// the bigit-length of 'a'+'b' must be equal to the one of 'a'.
	if (a->exponent >= bignum_bigit_length(b) && bignum_bigit_length(a) < bignum_bigit_length(c)) {
		return -1;
	}

	uint32_t borrow = 0;
	// Starting at min_exponent all digits are == 0. So no need to compare them.
	int min_exponent = a->exponent;
	if (b->exponent < min_exponent) {
		min_exponent = b->exponent;
	}
	if (c->exponent < min_exponent) {
		min_exponent = c->exponent;
	}
	for (int i = bignum_bigit_length(c) - 1; i >= min_exponent; --i) {
		uint32_t chunk_a = bignum_bigit_at(a, i);
		uint32_t chunk_b = bignum_bigit_at(b, i);
		uint32_t chunk_c = bignum_bigit_at(c, i);
		uint32_t sum = chunk_a + chunk_b;
		if (sum > chunk_c + borrow) {
			return +1;
		} else {
			borrow = chunk_c + borrow - sum;
			if (borrow > 1) {
				return -1;
			}
			borrow <<= bignum_bigit_size;
		}
	}
	if (borrow == 0) {
		return 0;
	}
	return -1;
}

static void bignum_subtract_times(bignum *number, const bignum *other, int factor) {
	if (factor < 3) {
		for (int i = 0; i < factor; ++i) {
			bignum_subtract_bignum(number, other);
		}
		return;
	}
	uint32_t borrow = 0;
	int exponent_diff = other->exponent - number->exponent;
	for (int i = 0; i < other->used_digits; ++i) {
		uint64_t product = (uint64_t)factor * other->bigits[i];
		uint64_t remove = borrow + product;
		uint32_t difference = number->bigits[i + exponent_diff] - (uint32_t)(remove & bignum_bigit_mask);
		number->bigits[i + exponent_diff] = difference & bignum_bigit_mask;
		borrow = (uint32_t)((difference >> (bignum_chunk_size - 1)) + (remove >> bignum_bigit_size));
	}
	for (int i = other->used_digits + exponent_diff; i < number->used_digits; ++i) {
		if (borrow == 0) {
			return;
		}
		uint32_t difference = number->bigits[i] - borrow;
		number->bigits[i] = difference & bignum_bigit_mask;
		borrow = difference >> (bignum_chunk_size - 1);
	}
	bignum_clamp(number);
}

// Precondition: number / other < 16. Returns number / other and leaves number % other in number.
static uint16_t bignum_divide_modulo_int_bignum(bignum *number, const bignum *other) {
	// Easy case: if we have less digits than the divisor than the result is 0. Note: this handles
	// the case where this == 0, too.
	if (bignum_bigit_length(number) < bignum_bigit_length(other)) {
		return 0;
	}

	bignum_align(number, other);

	uint16_t result = 0;

	// Start by removing multiples of 'other' until both numbers have the same number of digits.
	while (bignum_bigit_length(number) > bignum_bigit_length(other)) {
		// This naive approach is extremely inefficient if `this` divided by other is big. This
		// function is implemented for doubleToString where the result should be small (less than 10).
		result += (uint16_t)number->bigits[number->used_digits - 1];
		bignum_subtract_times(number, other, (int)number->bigits[number->used_digits - 1]);
	}

	uint32_t this_bigit = number->bigits[number->used_digits - 1];
	uint32_t other_bigit = other->bigits[other->used_digits - 1];

	if (other->used_digits == 1) {
		// Shortcut for easy (and common) case.
		int quotient = (int)(this_bigit / other_bigit);
		number->bigits[number->used_digits - 1] = this_bigit - other_bigit * (uint32_t)quotient;
		result += (uint16_t)quotient;
		bignum_clamp(number);
		return result;
	}

	int division_estimate = (int)(this_bigit / (other_bigit + 1));
	result += (uint16_t)division_estimate;
	bignum_subtract_times(number, other, division_estimate);

	if (other_bigit * (uint32_t)(division_estimate + 1) > this_bigit) {
		// No need to even try to subtract. Even if other's remaining digits were 0 another
		// subtraction would be too much.
		return result;
	}

	while (bignum_less_equal(other, number)) {
		bignum_subtract_bignum(number, other);
		result++;
	}
	return result;
}

// ---- src/base/numbers/bignum-dtoa.cc

enum bignum_dtoa_mode {
	bignum_dtoa_shortest,
	bignum_dtoa_precision,
};

static int normalized_exponent(uint64_t significand, int exponent) {
	while ((significand & double_hidden_bit) == 0) {
		significand = significand << 1;
		exponent = exponent - 1;
	}
	return exponent;
}

// Generates the shortest digits that read back as v; the delta bignums are the distance to v's
// neighbors' midpoints.
static void generate_shortest_digits(bignum *numerator, bignum *denominator, bignum *delta_minus, bignum *delta_plus, bool is_even, char *buffer, int *length) {
	// Small optimization: if delta_minus and delta_plus are the same just reuse one of the two
	// bignums.
	if (bignum_equal(delta_minus, delta_plus)) {
		delta_plus = delta_minus;
	}
	*length = 0;
	while (true) {
		uint16_t digit = bignum_divide_modulo_int_bignum(numerator, denominator);
		// digit = numerator / denominator (integer division). numerator = numerator % denominator.
		buffer[(*length)++] = (char)(digit + '0');

		// Can we stop already? If the remainder of the division is less than the distance to the
		// lower boundary we can stop. In this case we simply round down (discarding the remainder).
		// Similarly we test if we can round up (using the upper boundary).
		bool in_delta_room_minus;
		bool in_delta_room_plus;
		if (is_even) {
			in_delta_room_minus = bignum_less_equal(numerator, delta_minus);
		} else {
			in_delta_room_minus = bignum_less(numerator, delta_minus);
		}
		if (is_even) {
			in_delta_room_plus = bignum_plus_compare(numerator, delta_plus, denominator) >= 0;
		} else {
			in_delta_room_plus = bignum_plus_compare(numerator, delta_plus, denominator) > 0;
		}
		if (!in_delta_room_minus && !in_delta_room_plus) {
			// Prepare for next iteration.
			bignum_times_10(numerator);
			bignum_times_10(delta_minus);
			// We optimized delta_plus to be equal to delta_minus (if they share the same value). So
			// don't multiply delta_plus if they point to the same object.
			if (delta_minus != delta_plus) {
				bignum_times_10(delta_plus);
			}
		} else if (in_delta_room_minus && in_delta_room_plus) {
			// Let's see if 2*numerator < denominator. If yes, then the next digit would be < 5 and we
			// can round down.
			int compare = bignum_plus_compare(numerator, numerator, denominator);
			if (compare < 0) {
				// Remaining digits are less than .5. -> Round down (== do nothing).
			} else if (compare > 0) {
				// Remaining digits are more than .5 of denominator. -> Round up. Note that the last
				// digit could not be a '9' as otherwise the whole loop would have stopped earlier.
				buffer[(*length) - 1]++;
			} else {
				// Halfway case. Round towards even.
				if ((buffer[(*length) - 1] - '0') % 2 != 0) {
					buffer[(*length) - 1]++;
				}
			}
			return;
		} else if (in_delta_room_minus) {
			// Round down (== do nothing).
			return;
		} else {
			// in_delta_room_plus: round up. Note again that the last digit could not be '9' since
			// the loop would have stopped earlier.
			buffer[(*length) - 1]++;
			return;
		}
	}
}

// Let v = numerator / denominator < 10. Then we generate 'count' digits of d = v / 10^count, the
// last one rounded half up on the exact remainder.
static void generate_counted_digits(int count, int *decimal_point, bignum *numerator, bignum *denominator, char *buffer, int *length) {
	for (int i = 0; i < count - 1; ++i) {
		uint16_t digit = bignum_divide_modulo_int_bignum(numerator, denominator);
		// digit = numerator / denominator (integer division). numerator = numerator % denominator.
		buffer[i] = (char)(digit + '0');
		// Prepare for next iteration.
		bignum_times_10(numerator);
	}
	// Generate the last digit.
	uint16_t digit = bignum_divide_modulo_int_bignum(numerator, denominator);
	if (bignum_plus_compare(numerator, numerator, denominator) >= 0) {
		digit++;
	}
	buffer[count - 1] = (char)(digit + '0');
	// Correct bad digits (in case we had a sequence of '9's). Propagate the carry until we hat a
	// non-'9' or til we reach the first digit.
	for (int i = count - 1; i > 0; --i) {
		if (buffer[i] != '0' + 10) {
			break;
		}
		buffer[i] = '0';
		buffer[i - 1]++;
	}
	if (buffer[0] == '0' + 10) {
		// Propagate a carry past the top place.
		buffer[0] = '1';
		(*decimal_point)++;
	}
	*length = count;
}

// Returns an estimation of k such that 10^(k-1) <= v < 10^k where v = f * 2^exponent and
// 2^52 <= f < 2^53. v is hence a normalized double with the given exponent. The result might
// undershoot by 1, in which case 10^k <= v < 10^k+1. Note: this property holds for v's upper
// boundary m+ too.
static int estimate_power(int exponent) {
	const double one_over_log2_10 = 0.30102999566398114; // 1/lg(10)
	// For doubles len(f) == 53 (don't forget the hidden bit).
	const int significand_size = 53;
	double estimate = ceil((exponent + significand_size - 1) * one_over_log2_10 - 1e-10);
	return (int)estimate;
}

// See comments for initial_scaled_start_values.
static void initial_scaled_start_values_positive_exponent(double v, int estimated_power, bool need_boundary_deltas, bignum *numerator, bignum *denominator, bignum *delta_minus, bignum *delta_plus) {
	// A positive exponent implies a positive power.
	// v = f * 2^e with e > 0. And let p = estimated_power. Then v / 10^p = (f * 2^e) / 10^p. Let
	// numerator = f * 2^e and denominator = 10^p.
	bignum_assign_uint64(numerator, double_significand(v));
	bignum_shift_left(numerator, double_exponent(v));
	bignum_assign_power_uint16(denominator, 10, estimated_power);

	if (need_boundary_deltas) {
		// Introduce a common denominator so that the deltas to the boundaries are integers.
		bignum_shift_left(denominator, 1);
		bignum_shift_left(numerator, 1);
		// Let v + delta_plus = (f+1) * 2^e be the upper boundary. With the common denominator
		// delta_plus = 2^e.
		bignum_assign_uint16(delta_plus, 1);
		bignum_shift_left(delta_plus, double_exponent(v));
		// Same for delta_minus (with adjustments below if f == 2^p-1).
		bignum_assign_uint16(delta_minus, 1);
		bignum_shift_left(delta_minus, double_exponent(v));

		// If the significand (without the hidden bit) is 0, then the lower boundary is closer than
		// just half a ulp (unit in the last place). There is only one exception: if the next lower
		// number is a denormal then the distance is 1 ulp. This cannot be the case for exponent >= 0
		// (but we have to test it in the other function where exponent < 0).
		if ((double_bits(v) & double_significand_mask) == 0) {
			// The lower boundary is closer at half the distance of "normal" numbers. Increase the
			// common denominator and adapt all but the delta_minus.
			bignum_shift_left(denominator, 1); // *2
			bignum_shift_left(numerator, 1);   // *2
			bignum_shift_left(delta_plus, 1);  // *2
		}
	}
}

// See comments for initial_scaled_start_values.
static void initial_scaled_start_values_negative_exponent_positive_power(double v, int estimated_power, bool need_boundary_deltas, bignum *numerator, bignum *denominator, bignum *delta_minus, bignum *delta_plus) {
	uint64_t significand = double_significand(v);
	int exponent = double_exponent(v);
	// v = f * 2^e with e < 0, and with estimated_power >= 0. This means that e is close to 0 (have
	// a look at how estimated_power is computed).

	bignum_assign_uint64(numerator, significand);
	bignum_assign_power_uint16(denominator, 10, estimated_power);
	bignum_shift_left(denominator, -exponent);

	if (need_boundary_deltas) {
		// Introduce a common denominator so that the deltas to the boundaries are integers.
		bignum_shift_left(denominator, 1);
		bignum_shift_left(numerator, 1);
		// Let v + delta_plus = (f+1) * 2^e be the upper boundary. With the common denominator
		// delta_plus = 1.
		bignum_assign_uint16(delta_plus, 1);
		// Same for delta_minus (with adjustments below if f == 2^p-1).
		bignum_assign_uint16(delta_minus, 1);

		// If the significand (without the hidden bit) is 0, then the lower boundary is closer than
		// just one ulp (unit in the last place). There is only one exception: if the next lower
		// number is a denormal then the distance is 1 ulp. Since the exponent is close to zero
		// (see estimated_power) the next lower number is never a denormal when the exponent is
		// negative.
		if ((double_bits(v) & double_significand_mask) == 0) {
			// The lower boundary is closer at half the distance of "normal" numbers. Increase the
			// denominator and adapt all but the delta_minus.
			bignum_shift_left(denominator, 1); // *2
			bignum_shift_left(numerator, 1);   // *2
			bignum_shift_left(delta_plus, 1);  // *2
		}
	}
}

// See comments for initial_scaled_start_values.
static void initial_scaled_start_values_negative_exponent_negative_power(double v, int estimated_power, bool need_boundary_deltas, bignum *numerator, bignum *denominator, bignum *delta_minus, bignum *delta_plus) {
	const uint64_t minimal_normalized_exponent = 0x0010000000000000u;
	uint64_t significand = double_significand(v);
	int exponent = double_exponent(v);
	// Instead of multiplying the denominator with 10^estimated_power we multiply all values
	// (numerator and deltas) by 10^-estimated_power.

	// Use numerator as temporary container for power_ten.
	bignum *power_ten = numerator;
	bignum_assign_power_uint16(power_ten, 10, -estimated_power);

	if (need_boundary_deltas) {
		// Since power_ten == numerator we must make a copy of 10^estimated_power before we
		// multiply the numerator.
		bignum_assign_bignum(delta_plus, power_ten);
		bignum_assign_bignum(delta_minus, power_ten);
	}

	// numerator = significand * 2 * 10^-estimated_power since v = significand * 2^exponent this is
	// equivalent to v * 10^-estimated_power * 2^-exponent * 2. Remember: numerator has been
	// abused as power_ten. So no need to assign it to itself.
	bignum_multiply_by_uint64(numerator, significand);

	// denominator = 2 * 2^-exponent with exponent < 0.
	bignum_assign_uint16(denominator, 1);
	bignum_shift_left(denominator, -exponent);

	if (need_boundary_deltas) {
		// Introduce a common denominator so that the deltas to the boundaries are integers.
		bignum_shift_left(numerator, 1);
		bignum_shift_left(denominator, 1);
		// With this shift the boundaries have their correct value, since delta_plus = 10^-estimated_power,
		// and delta_minus = 10^-estimated_power. These assignments have been done earlier.

		// The special case where the lower boundary is twice as close. This time we have to look
		// out for the exception too.
		uint64_t v_bits = double_bits(v);
		if ((v_bits & double_significand_mask) == 0 && (v_bits & double_exponent_mask) != minimal_normalized_exponent) {
			bignum_shift_left(numerator, 1);   // *2
			bignum_shift_left(denominator, 1); // *2
			bignum_shift_left(delta_plus, 1);  // *2
		}
	}
}

// Let v = significand * 2^exponent. Computes v / 10^estimated_power exactly, as a ratio of two
// bignums, numerator / denominator, and with need_boundary_deltas the distances to v's neighbors'
// midpoints over the same denominator.
static void initial_scaled_start_values(double v, int estimated_power, bool need_boundary_deltas, bignum *numerator, bignum *denominator, bignum *delta_minus, bignum *delta_plus) {
	if (double_exponent(v) >= 0) {
		initial_scaled_start_values_positive_exponent(v, estimated_power, need_boundary_deltas, numerator, denominator, delta_minus, delta_plus);
	} else if (estimated_power >= 0) {
		initial_scaled_start_values_negative_exponent_positive_power(v, estimated_power, need_boundary_deltas, numerator, denominator, delta_minus, delta_plus);
	} else {
		initial_scaled_start_values_negative_exponent_negative_power(v, estimated_power, need_boundary_deltas, numerator, denominator, delta_minus, delta_plus);
	}
}

// This routine multiplies numerator/denominator so that its values lies in the range 1-10. That is
// after a call to this function we have: 1 <= (numerator + delta_plus) / denominator < 10. Let
// numerator the input before modification and numerator' the argument after modification, then
// the output-parameter decimal_point is such that numerator / denominator * 10^estimated_power ==
// numerator' / denominator' * 10^(decimal_point - 1). In some cases estimated_power was too low,
// and this is already the case. We then simply adjust the power so that 10^(k-1) <= v < 10^k (with
// k == estimated_power) but do not touch the numerator or denominator. Otherwise the routine
// multiplies the numerator and the deltas by 10.
static void fixup_multiply_10(int estimated_power, bool is_even, int *decimal_point, bignum *numerator, bignum *denominator, bignum *delta_minus, bignum *delta_plus) {
	bool in_range;
	if (is_even) {
		// For IEEE doubles half-way cases (in decimal system numbers ending with 5) are rounded to
		// the closest floating-point number with even significand.
		in_range = bignum_plus_compare(numerator, delta_plus, denominator) >= 0;
	} else {
		in_range = bignum_plus_compare(numerator, delta_plus, denominator) > 0;
	}
	if (in_range) {
		// Since numerator + delta_plus >= denominator we already have 1 <= numerator/denominator < 10.
		// Simply adjust the estimated_power.
		*decimal_point = estimated_power + 1;
	} else {
		*decimal_point = estimated_power;
		bignum_times_10(numerator);
		if (bignum_equal(delta_minus, delta_plus)) {
			bignum_times_10(delta_minus);
			bignum_assign_bignum(delta_plus, delta_minus);
		} else {
			bignum_times_10(delta_minus);
			bignum_times_10(delta_plus);
		}
	}
}

// v is positive and finite. The digits go to buffer (terminated), and v is 0.d1d2... times
// 10^decimal_point.
static void bignum_dtoa(double v, enum bignum_dtoa_mode mode, int requested_digits, char *buffer, int *length, int *decimal_point) {
	uint64_t significand = double_significand(v);
	bool is_even = (significand & 1) == 0;
	int exponent = double_exponent(v);
	int estimated_power = estimate_power(normalized_exponent(significand, exponent));

	bignum numerator, denominator, delta_minus, delta_plus;
	bignum_init(&numerator);
	bignum_init(&denominator);
	bignum_init(&delta_minus);
	bignum_init(&delta_plus);
	// Make sure the bignum can grow large enough. The smallest double equals 4e-324. In this case
	// the denominator needs fewer than 324*4 binary digits. The maximum double is 1.7976931348623157e308
	// which needs fewer than 308*4 binary digits.
	bool need_boundary_deltas = mode == bignum_dtoa_shortest;
	initial_scaled_start_values(v, estimated_power, need_boundary_deltas, &numerator, &denominator, &delta_minus, &delta_plus);
	// We now have v = (numerator / denominator) * 10^estimated_power.
	fixup_multiply_10(estimated_power, is_even, decimal_point, &numerator, &denominator, &delta_minus, &delta_plus);
	// We now have v = (numerator / denominator) * 10^(decimal_point-1), and 1 <= (numerator + delta_plus) / denominator < 10
	if (mode == bignum_dtoa_shortest) {
		generate_shortest_digits(&numerator, &denominator, &delta_minus, &delta_plus, is_even, buffer, length);
	} else {
		generate_counted_digits(requested_digits, decimal_point, &numerator, &denominator, buffer, length);
	}
	buffer[*length] = '\0';
}

// ---- src/base/numbers/dtoa.cc: DoubleToAscii, for a value that is finite and not negative.

static void double_to_ascii(double v, enum bignum_dtoa_mode mode, int requested_digits, char *buffer, int *length, int *point) {
	if (v == 0) {
		buffer[0] = '0';
		buffer[1] = '\0';
		*length = 1;
		*point = 1;
		return;
	}
	if (mode == bignum_dtoa_precision && requested_digits == 0) {
		buffer[0] = '\0';
		*length = 0;
		return;
	}
	bignum_dtoa(v, mode, requested_digits, buffer, length, point);
}

// adamic_number_shortest_digits is DoubleToAscii in its shortest mode, what V8's DoubleToCString
// (Number::toString) takes its digits from: the fewest digits that read back as value, the closest
// of those. value is positive and finite; digits holds at most 17 and a terminating NUL, and value
// is 0.d1d2d3... times 10^point.
int adamic_number_shortest_digits(double value, char digits[18], int *point) {
	int length = 0;
	double_to_ascii(value, bignum_dtoa_shortest, 0, digits, &length, point);
	return length;
}

// ---- src/numbers/conversions.cc

// The longest text either function writes: a sign, 101 digits, a point, and "e+308" or so, with
// room to spare.
enum { conversion_buffer_size = 128 };

static int append_text(char *buffer, int length, const char *text) {
	while (*text != '\0') {
		buffer[length++] = *text++;
	}
	return length;
}

static int append_padding(char *buffer, int length, char character, int count) {
	for (int index = 0; index < count; index++) {
		buffer[length++] = character;
	}
	return length;
}

static int append_decimal_integer(char *buffer, int length, int value) {
	char digits[12];
	int count = 0;
	do {
		digits[count++] = (char)('0' + value % 10);
		value /= 10;
	} while (value != 0);
	while (count > 0) {
		buffer[length++] = digits[--count];
	}
	return length;
}

// CreateExponentialRepresentation
static int create_exponential_representation(const char *decimal_rep, int rep_length, int exponent, bool negative, int significant_digits, char *buffer) {
	bool negative_exponent = false;
	if (exponent < 0) {
		negative_exponent = true;
		exponent = -exponent;
	}
	int length = 0;
	if (negative) {
		buffer[length++] = '-';
	}
	buffer[length++] = decimal_rep[0];
	if (significant_digits != 1) {
		buffer[length++] = '.';
		for (int index = 1; index < rep_length; index++) {
			buffer[length++] = decimal_rep[index];
		}
		length = append_padding(buffer, length, '0', significant_digits - rep_length);
	}
	buffer[length++] = 'e';
	buffer[length++] = negative_exponent ? '-' : '+';
	return append_decimal_integer(buffer, length, exponent);
}

// DoubleToExponentialStringView. f is the digits after the point, or -1 when there was no argument.
static int double_to_exponential(double value, int f, char *buffer) {
	bool negative = false;
	if (value < 0) {
		value = -value;
		negative = true;
	}
	int decimal_point = 0;
	char decimal_rep[max_fraction_digits + 1 + 1];
	int decimal_rep_length;
	if (f == -1) {
		double_to_ascii(value, bignum_dtoa_shortest, 0, decimal_rep, &decimal_rep_length, &decimal_point);
		f = decimal_rep_length - 1;
	} else {
		double_to_ascii(value, bignum_dtoa_precision, f + 1, decimal_rep, &decimal_rep_length, &decimal_point);
	}
	int exponent = decimal_point - 1;
	return create_exponential_representation(decimal_rep, decimal_rep_length, exponent, negative, f + 1, buffer);
}

// DoubleToPrecisionStringView
static int double_to_precision(double value, int p, char *buffer) {
	bool negative = false;
	if (value < 0) {
		value = -value;
		negative = true;
	}
	int decimal_point = 0;
	char decimal_rep[max_fraction_digits + 1];
	int decimal_rep_length;
	double_to_ascii(value, bignum_dtoa_precision, p, decimal_rep, &decimal_rep_length, &decimal_point);

	int exponent = decimal_point - 1;
	if (exponent < -6 || exponent >= p) {
		return create_exponential_representation(decimal_rep, decimal_rep_length, exponent, negative, p, buffer);
	}
	// Use fixed notation.
	int length = 0;
	if (negative) {
		buffer[length++] = '-';
	}
	if (decimal_point <= 0) {
		length = append_text(buffer, length, "0.");
		length = append_padding(buffer, length, '0', -decimal_point);
		for (int index = 0; index < decimal_rep_length; index++) {
			buffer[length++] = decimal_rep[index];
		}
		length = append_padding(buffer, length, '0', p - decimal_rep_length);
	} else {
		int m = decimal_rep_length < decimal_point ? decimal_rep_length : decimal_point;
		for (int index = 0; index < m; index++) {
			buffer[length++] = decimal_rep[index];
		}
		length = append_padding(buffer, length, '0', decimal_point - decimal_rep_length);
		if (decimal_point < p) {
			buffer[length++] = '.';
			int extra = negative ? 2 : 1;
			if (decimal_rep_length > decimal_point) {
				int len = decimal_rep_length - decimal_point;
				int n = p - (length - extra);
				if (len < n) {
					n = len;
				}
				for (int index = 0; index < n; index++) {
					buffer[length++] = decimal_rep[decimal_point + index];
				}
			}
			length = append_padding(buffer, length, '0', extra + (p - length));
		}
	}
	return length;
}

// ---- src/builtins/builtins-number.cc: the order things are checked in, which decides what
// (NaN).toExponential(1000) is ("NaN", not a RangeError).

static adamic_string *conversion_result(const char *text, int length) {
	adamic_string result = {{0, adamic_kind_string, 0}, (size_t)length, text, 0, NULL, NULL, 0};
	return adamic_string_concat(1, (adamic_string *const[]){&result});
}

static adamic_string *conversion_special(double value) {
	if (isnan(value)) {
		return conversion_result("NaN", 3);
	}
	return value < 0 ? conversion_result("-Infinity", 9) : conversion_result("Infinity", 8);
}

// NumberPrototypeToExponential. Without an argument the digits are the shortest that read back.
adamic_string *adamic_number_to_exponential(double value, double fraction_digits, bool has_digits) {
	// ToIntegerOrInfinity: NaN is 0, and a fraction truncates.
	double digits = has_digits && !isnan(fraction_digits) ? trunc(fraction_digits) : 0;
	if (isnan(value) || isinf(value)) {
		return conversion_special(value);
	}
	if (digits < 0 || digits > max_fraction_digits) {
		static const char message[] = "RangeError: toExponential() argument must be between 0 and 100";
		adamic_panic(message, sizeof message - 1);
	}
	char buffer[conversion_buffer_size];
	int length = double_to_exponential(value, has_digits ? (int)digits : -1, buffer);
	return conversion_result(buffer, length);
}

// NumberPrototypeToPrecision. Without an argument it's ToString.
adamic_string *adamic_number_to_precision(double value, double precision, bool has_precision) {
	if (!has_precision) {
		return adamic_string_from_number(value);
	}
	double digits = isnan(precision) ? 0 : trunc(precision);
	if (isnan(value) || isinf(value)) {
		return conversion_special(value);
	}
	if (digits < 1 || digits > max_fraction_digits) {
		static const char message[] = "RangeError: toPrecision() argument must be between 1 and 100";
		adamic_panic(message, sizeof message - 1);
	}
	char buffer[conversion_buffer_size];
	int length = double_to_precision(value, (int)digits, buffer);
	return conversion_result(buffer, length);
}
