// string.c: Adamic's strings. Immutable, reference counted, and never collected.
//
// A string the program spells out is a static constant with references 0: immortal, never counted,
// never freed. A string built at runtime starts at 1, and the last release frees it. Whoever holds a
// reference releases it exactly once; the compiler inserts every retain and release, so nothing here
// has to guess who's still looking.
//
// The bytes are UTF-8. JavaScript's strings are UTF-16 to a program (length, indexes, <), and the
// operations that see the difference arrive with the program that needs them (docs/0.1.md, program
// 10), each held to Node by the oracle.

#include "adamic.h"

#include <math.h>

#include <stdint.h>
#include <string.h>

// allocate makes a string of length bytes, references 1, its bytes right after it in one block.
static adamic_string *allocate(size_t length) {
	adamic_string *string = adamic_allocate(sizeof *string + length, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	return string;
}

adamic_string *adamic_string_from_number(double value) {
	char buffer[ADAMIC_NUMBER_FORMAT_MAX];
	size_t length = adamic_number_format(value, buffer);
	adamic_string *string = allocate(length);
	memcpy((char *)string->bytes, buffer, length);
	return string;
}

adamic_string *adamic_string_concat(size_t count, adamic_string *const parts[]) {
	size_t length = 0;
	for (size_t index = 0; index < count; index++) {
		if (parts[index]->length > SIZE_MAX - length) {
			static const char message[] = "string too long";
			adamic_panic(message, sizeof message - 1);
		}
		length += parts[index]->length;
	}
	adamic_string *string = allocate(length);
	char *cursor = (char *)string->bytes;
	for (size_t index = 0; index < count; index++) {
		if (parts[index]->length > 0) {
			memcpy(cursor, parts[index]->bytes, parts[index]->length);
		}
		cursor += parts[index]->length;
	}
	return string;
}

int adamic_string_equal(const adamic_string *left, const adamic_string *right) {
	return left->length == right->length && (left->length == 0 || memcmp(left->bytes, right->bytes, left->length) == 0);
}

adamic_string adamic_string_empty = ADAMIC_STRING("");
adamic_string adamic_string_true = ADAMIC_STRING("true");
adamic_string adamic_string_false = ADAMIC_STRING("false");

// UTF-16 to the program, UTF-8 underneath. A code point of 1 to 3 UTF-8 bytes is one UTF-16 code
// unit, and one of 4 bytes is two, a surrogate pair. A lone surrogate is stored as the 3 bytes
// WTF-8 gives it, and is one unit too, so the same rule counts it.

// sequence is how many bytes the code point starting with this byte takes.
static size_t sequence(unsigned char lead) {
	if (lead < 0x80) {
		return 1;
	}
	if (lead < 0xe0) {
		return 2;
	}
	if (lead < 0xf0) {
		return 3;
	}
	return 4;
}

// decode reads the code point at bytes, given its sequence length.
static unsigned decode(const unsigned char *bytes, size_t size) {
	switch (size) {
	case 1:
		return bytes[0];
	case 2:
		return ((unsigned)(bytes[0] & 0x1f) << 6) | (bytes[1] & 0x3f);
	case 3:
		return ((unsigned)(bytes[0] & 0x0f) << 12) | ((unsigned)(bytes[1] & 0x3f) << 6) | (bytes[2] & 0x3f);
	}
	return ((unsigned)(bytes[0] & 0x07) << 18) | ((unsigned)(bytes[1] & 0x3f) << 12) | ((unsigned)(bytes[2] & 0x3f) << 6) | (bytes[3] & 0x3f);
}

double adamic_string_length(const adamic_string *string) {
	size_t units = 0;
	for (size_t offset = 0; offset < string->length;) {
		size_t size = sequence((unsigned char)string->bytes[offset]);
		units += size == 4 ? 2 : 1;
		offset += size;
	}
	return (double)units;
}

double adamic_string_char_code_at(const adamic_string *string, double position) {
	// ToIntegerOrInfinity: NaN is 0, and a fraction truncates.
	position = isnan(position) ? 0 : trunc(position);
	if (position < 0) {
		return NAN;
	}
	size_t unit = 0;
	for (size_t offset = 0; offset < string->length;) {
		size_t size = sequence((unsigned char)string->bytes[offset]);
		unsigned point = decode((const unsigned char *)string->bytes + offset, size);
		if (size == 4) {
			if ((double)unit == position) {
				return (double)(0xd800 + ((point - 0x10000) >> 10));
			}
			if ((double)(unit + 1) == position) {
				return (double)(0xdc00 + ((point - 0x10000) & 0x3ff));
			}
			unit += 2;
		} else {
			if ((double)unit == position) {
				return (double)point;
			}
			unit++;
		}
		offset += size;
	}
	return NAN;
}

// is_space is JavaScript's WhiteSpace and LineTerminator, which trim removes (ECMA-262).
static bool is_space(unsigned point) {
	switch (point) {
	case 0x09: case 0x0a: case 0x0b: case 0x0c: case 0x0d: case 0x20: case 0xa0: case 0x1680:
	case 0x2028: case 0x2029: case 0x202f: case 0x205f: case 0x3000: case 0xfeff:
		return true;
	}
	return point >= 0x2000 && point <= 0x200a;
}

adamic_string *adamic_string_trim(adamic_string *string) {
	size_t start = 0, end = string->length;
	while (start < end) {
		size_t size = sequence((unsigned char)string->bytes[start]);
		if (!is_space(decode((const unsigned char *)string->bytes + start, size))) {
			break;
		}
		start += size;
	}
	while (end > start) {
		// Back up to the start of the last code point: continuation bytes are 10xxxxxx.
		size_t lead = end - 1;
		while (lead > start && ((unsigned char)string->bytes[lead] & 0xc0) == 0x80) {
			lead--;
		}
		if (!is_space(decode((const unsigned char *)string->bytes + lead, end - lead))) {
			break;
		}
		end = lead;
	}
	adamic_string piece = {{0, adamic_kind_string}, end - start, string->bytes + start};
	return adamic_string_concat(1, (adamic_string *const[]){&piece});
}

size_t adamic_string_next(const adamic_string *string, size_t offset) {
	return sequence((unsigned char)string->bytes[offset]);
}

adamic_string *adamic_string_slice_bytes(const adamic_string *string, size_t offset, size_t size) {
	adamic_string piece = {{0, adamic_kind_string}, size, string->bytes + offset};
	return adamic_string_concat(1, (adamic_string *const[]){&piece});
}
