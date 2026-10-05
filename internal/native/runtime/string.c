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
