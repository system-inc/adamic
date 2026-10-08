// String.fromCharCode and String.fromCodePoint, as ECMAScript defines them.
//
// Each writes its units as WTF-8, a lone surrogate as three bytes, and hands the bytes to
// adamic_string_concat, which joins a high surrogate followed by a low one into the four bytes of the
// character they make together, as JavaScript sees them: String.fromCharCode(0xd83c, 0xdf0d) is 🌍.

#include "adamic.h"

#include <math.h>
#include <stdlib.h>
#include <string.h>

// codes is a growing buffer of WTF-8.
typedef struct codes {
	char *bytes;
	size_t length;
	size_t capacity;
} codes;

static void codes_add(codes *buffer, const char *bytes, size_t size) {
	if (buffer->length + size > buffer->capacity) {
		size_t capacity = buffer->capacity == 0 ? 32 : buffer->capacity;
		while (capacity < buffer->length + size) {
			capacity *= 2;
		}
		char *grown = realloc(buffer->bytes, capacity);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		buffer->bytes = grown;
		buffer->capacity = capacity;
	}
	memcpy(buffer->bytes + buffer->length, bytes, size);
	buffer->length += size;
}

// codes_point appends one code point, a surrogate as WTF-8's three bytes.
static void codes_point(codes *buffer, unsigned point) {
	char bytes[4];
	if (point < 0x80) {
		bytes[0] = (char)point;
		codes_add(buffer, bytes, 1);
	} else if (point < 0x800) {
		bytes[0] = (char)(0xc0 | (point >> 6));
		bytes[1] = (char)(0x80 | (point & 0x3f));
		codes_add(buffer, bytes, 2);
	} else if (point < 0x10000) {
		bytes[0] = (char)(0xe0 | (point >> 12));
		bytes[1] = (char)(0x80 | ((point >> 6) & 0x3f));
		bytes[2] = (char)(0x80 | (point & 0x3f));
		codes_add(buffer, bytes, 3);
	} else {
		bytes[0] = (char)(0xf0 | (point >> 18));
		bytes[1] = (char)(0x80 | ((point >> 12) & 0x3f));
		bytes[2] = (char)(0x80 | ((point >> 6) & 0x3f));
		bytes[3] = (char)(0x80 | (point & 0x3f));
		codes_add(buffer, bytes, 4);
	}
}

static adamic_string *codes_finish(codes *buffer) {
	adamic_string piece = {{0, adamic_kind_string, 0}, buffer->length, buffer->bytes == NULL ? "" : buffer->bytes, 0, NULL, NULL, 0};
	adamic_string *string = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(buffer->bytes);
	return string;
}

// adamic_to_uint16 is ECMAScript's ToUint16: NaN and the infinities are 0, anything else is truncated
// and taken modulo 2^16. fmod is exact, so this is exactly the specification's arithmetic.
static unsigned adamic_to_uint16(double value) {
	if (!isfinite(value)) {
		return 0;
	}
	double modulo = fmod(trunc(value), 65536.0);
	if (modulo < 0) {
		modulo += 65536.0;
	}
	return (unsigned)modulo;
}

adamic_string *adamic_string_from_char_codes(size_t count, const double values[]) {
	codes buffer = {NULL, 0, 0};
	for (size_t index = 0; index < count; index++) {
		codes_point(&buffer, adamic_to_uint16(values[index]));
	}
	return codes_finish(&buffer);
}

adamic_string *adamic_string_from_code_points(size_t count, const double values[]) {
	codes buffer = {NULL, 0, 0};
	for (size_t index = 0; index < count; index++) {
		double value = values[index];
		// Invalid points raise through the same pending word as an explicit throw, so catches
		// and finallies release the caller's references on the way out.
		if (!(value >= 0 && value <= 0x10ffff && value == trunc(value))) {
			free(buffer.bytes);
			static adamic_string prefix = ADAMIC_STRING("Invalid code point ");
			adamic_string *written = adamic_string_from_number(value);
			adamic_string *message = adamic_string_concat(2, (adamic_string *const[]){&prefix, written});
			adamic_thrown = adamic_builtin_error_new(3, message);
			adamic_release(message);
			adamic_release(written);
			return NULL;
		}
		codes_point(&buffer, (unsigned)value);
	}
	return codes_finish(&buffer);
}
