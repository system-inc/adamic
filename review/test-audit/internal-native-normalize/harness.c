#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void put_hex(const adamic_string *string) {
	static const char digits[] = "0123456789abcdef";
	for (size_t index = 0; index < string->length; index++) {
		unsigned char byte = (unsigned char)string->bytes[index];
		putchar(digits[byte >> 4]);
		putchar(digits[byte & 15]);
	}
}

// encode writes a code point as WTF-8: a surrogate as its three bytes, as the runtime stores one.
static size_t encode(uint32_t point, char *out) {
	if (point < 0x80) {
		out[0] = (char)point;
		return 1;
	}
	if (point < 0x800) {
		out[0] = (char)(0xc0 | (point >> 6));
		out[1] = (char)(0x80 | (point & 0x3f));
		return 2;
	}
	if (point < 0x10000) {
		out[0] = (char)(0xe0 | (point >> 12));
		out[1] = (char)(0x80 | ((point >> 6) & 0x3f));
		out[2] = (char)(0x80 | (point & 0x3f));
		return 3;
	}
	out[0] = (char)(0xf0 | (point >> 18));
	out[1] = (char)(0x80 | ((point >> 12) & 0x3f));
	out[2] = (char)(0x80 | ((point >> 6) & 0x3f));
	out[3] = (char)(0x80 | (point & 0x3f));
	return 4;
}

// text is the string of these code points, made the way the runtime makes strings: a lone high
// surrogate met by a lone low one becomes the pair's character, as in JavaScript.
static adamic_string *text(const uint32_t *points, size_t count) {
	adamic_string *parts[8];
	char bytes[8][4];
	adamic_string pieces[8];
	for (size_t index = 0; index < count; index++) {
		size_t size = encode(points[index], bytes[index]);
		adamic_string piece = {{0, adamic_kind_string, 0}, size, bytes[index], 0, NULL, NULL, 0};
		pieces[index] = piece;
		parts[index] = &pieces[index];
	}
	return adamic_string_concat(count, parts);
}

static adamic_string forms[4] = {ADAMIC_STRING("NFC"), ADAMIC_STRING("NFD"), ADAMIC_STRING("NFKC"), ADAMIC_STRING("NFKD")};

static void answer(const uint32_t *points, size_t count) {
	adamic_string *string = text(points, count);
	for (size_t form = 0; form < 4; form++) {
		adamic_string *normalized = adamic_string_normalize(string, &forms[form]);
		putchar(' ');
		put_hex(normalized);
		adamic_release(normalized);
	}
	adamic_release(string);
}

static const uint32_t alphabet[] = {0x61, 0x41, 0x300, 0x301, 0x302, 0x323, 0x338, 0x345, 0x1100, 0x1161, 0x11a8, 0xac00, 0xac01, 0x9c7, 0x9be, 0x212b, 0x344, 0xf73, 0x1d15e, 0xfb01, 0x1e9b, 0xd800, 0xdc00, 0xc5};
#define LETTERS (sizeof alphabet / sizeof alphabet[0])

int main(int count, char **arguments) {
	static char buffer[1 << 20];
	setvbuf(stdout, buffer, _IOFBF, sizeof buffer);
	if (count >= 2 && strcmp(arguments[1], "points") == 0) {
		uint32_t first = count == 4 ? (uint32_t)strtoul(arguments[2], NULL, 10) : 0;
		uint32_t last = count == 4 ? (uint32_t)strtoul(arguments[3], NULL, 10) : 0x110000;
		for (uint32_t point = first; point < last; point++) {
			printf("%x", point);
			answer((uint32_t[]){point}, 1);
			answer((uint32_t[]){'a', point, 0x301}, 3);
			answer((uint32_t[]){point, 0x301, 0x323}, 3);
			answer((uint32_t[]){0x1100, point}, 2);
			answer((uint32_t[]){0xac00, point}, 2);
			putchar('\n');
		}
		return 0;
	}
	// Every string of one to four letters of the alphabet.
	for (size_t length = 1; length <= 4; length++) {
		size_t total = 1;
		for (size_t index = 0; index < length; index++) {
			total *= LETTERS;
		}
		for (size_t number = 0; number < total; number++) {
			uint32_t points[4];
			size_t rest = number;
			for (size_t index = length; index-- > 0;) {
				points[index] = alphabet[rest % LETTERS];
				rest /= LETTERS;
			}
			adamic_string *string = text(points, length);
			put_hex(string);
			adamic_release(string);
			answer(points, length);
			putchar('\n');
		}
	}
	return 0;
}
