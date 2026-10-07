// Integer fast paths. Included by adamic.h after its value declarations.
#ifndef ADAMIC_INTEGER_H
#define ADAMIC_INTEGER_H

uint32_t adamic_to_uint32_slow(double value);

static inline uint32_t adamic_to_uint32(double value) {
	// A cast truncates fractions too, exactly as ToInt32 does. Guard before casting: NaN,
	// infinities and out-of-range doubles must never reach a C integer conversion.
	if (fabs(value) < 2147483648.0) {
		return (uint32_t)(int32_t)value;
	}
	// INT32_MIN and arithmetic crossing int32 use the wider cast. Within 2^53,
	// truncation to int64 is exact and the unsigned conversion supplies the low 32 bits.
	if (fabs(value) <= 9007199254740992.0) {
		return (uint32_t)(int64_t)value;
	}
	return adamic_to_uint32_slow(value);
}

static inline double adamic_remainder(double left, double right) {
	// All integers through 2^53 fit exactly in int64. Range guards precede both casts,
	// excluding NaN and infinities; equality rejects fractions. No signed overflow is possible.
	if (fabs(left) <= 9007199254740992.0 && fabs(right) <= 9007199254740992.0 && right != 0.0) {
		int64_t dividend = (int64_t)left, divisor = (int64_t)right;
		if ((double)dividend == left && (double)divisor == right) {
			int64_t remainder = dividend % divisor;
			return remainder == 0 ? copysign(0.0, left) : (double)remainder;
		}
	}
	return fmod(left, right);
}

// adamic_signed_bits is 32 bits read as a two's complement integer, which is what ToInt32 gives.
static inline double adamic_signed_bits(uint32_t bits) {
	return bits >= 0x80000000u ? (double)bits - 4294967296.0 : (double)bits;
}

// adamic_shift_count is a shift's right operand: ToUint32, then its low five bits.
static inline unsigned adamic_shift_count(double count) {
	return (unsigned)(adamic_to_uint32(count) & 31u);
}

static inline double adamic_bitwise_and(double left, double right) {
	return adamic_signed_bits(adamic_to_uint32(left) & adamic_to_uint32(right));
}

static inline double adamic_bitwise_or(double left, double right) {
	return adamic_signed_bits(adamic_to_uint32(left) | adamic_to_uint32(right));
}

static inline double adamic_bitwise_xor(double left, double right) {
	return adamic_signed_bits(adamic_to_uint32(left) ^ adamic_to_uint32(right));
}

static inline double adamic_bitwise_not(double value) {
	return adamic_signed_bits(~adamic_to_uint32(value));
}

static inline double adamic_shift_left(double left, double right) {
	return adamic_signed_bits((uint32_t)(adamic_to_uint32(left) << adamic_shift_count(right)));
}

// adamic_shift_right is >>, which keeps the sign: the bits shifted in are the sign bit's.
static inline uint32_t adamic_shift_right_bits(uint32_t bits, unsigned count) {
	uint32_t shifted = bits >> count;
	if ((bits & 0x80000000u) != 0 && count > 0) {
		shifted |= ~(UINT32_MAX >> count);
	}
	return shifted;
}

static inline double adamic_shift_right(double left, double right) {
	return adamic_signed_bits(adamic_shift_right_bits(adamic_to_uint32(left), adamic_shift_count(right)));
}

// adamic_shift_right_unsigned is >>>, whose result is ToUint32's, from 0 to 2^32 - 1.
static inline double adamic_shift_right_unsigned(double left, double right) {
	return (double)(adamic_to_uint32(left) >> adamic_shift_count(right));
}

#endif
