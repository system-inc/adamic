// Snapshot of input.c decoder from area/runtime 4d86c305, kept as an independent regression oracle.
// input.c: what a program reads from outside, opened in 0.2, its arguments and files, and the one
// way it writes back besides its output: writeTextFile.
//
// Both arrive as bytes, and a program sees strings, so both are decoded exactly as Node decodes them:
// UTF-8 by the WHATWG decoder's rules, every invalid sequence one U+FFFD, and a byte-order mark kept
// as U+FEFF. The oracle holds that to Node with a sweep over malformed sequences.

// O_CLOEXEC is POSIX.1-2008's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"

#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#ifdef ADAMIC_TARGET_WASI
#include <sys/stat.h>
#endif

// new_string makes a string of length bytes, references 1, its bytes right after it in one block, the
// same layout string.c makes, so the heap frees it the same way.
static adamic_string *new_string(size_t length) {
	adamic_string *string = adamic_allocate(sizeof *string + length, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	string->units = 0;
	string->index = NULL;
	string->owner = NULL;
	string->capacity = length;
	return string;
}

// decode_step reads one code point of WHATWG UTF-8 at offset, and returns how many bytes it took.
// An invalid sequence is U+FFFD and takes only the bytes before the one that broke it, so that byte
// is read again as the start of what follows; a sequence cut off by the end is one U+FFFD.
static size_t decode_step(const unsigned char *bytes, size_t length, size_t offset, unsigned *point) {
	unsigned char lead = bytes[offset];
	if (lead < 0x80) {
		*point = lead;
		return 1;
	}
	unsigned char lower = 0x80, upper = 0xbf;
	size_t needed;
	unsigned value;
	if (lead >= 0xc2 && lead <= 0xdf) {
		needed = 1;
		value = lead & 0x1f;
	} else if (lead >= 0xe0 && lead <= 0xef) {
		// E0 can't begin an overlong form, and ED can't begin a surrogate.
		if (lead == 0xe0) {
			lower = 0xa0;
		}
		if (lead == 0xed) {
			upper = 0x9f;
		}
		needed = 2;
		value = lead & 0x0f;
	} else if (lead >= 0xf0 && lead <= 0xf4) {
		// F0 can't begin an overlong form, and F4 can't go past U+10FFFF.
		if (lead == 0xf0) {
			lower = 0x90;
		}
		if (lead == 0xf4) {
			upper = 0x8f;
		}
		needed = 3;
		value = lead & 0x07;
	} else {
		*point = 0xfffd;
		return 1;
	}
	size_t seen = 1;
	while (seen <= needed) {
		if (offset + seen >= length) {
			*point = 0xfffd;
			return seen;
		}
		unsigned char byte = bytes[offset + seen];
		if (byte < lower || byte > upper) {
			*point = 0xfffd;
			return seen;
		}
		lower = 0x80;
		upper = 0xbf;
		value = (value << 6) | (byte & 0x3f);
		seen++;
	}
	*point = value;
	return seen;
}

// encoded_size is how many bytes UTF-8 writes a code point in.
static size_t encoded_size(unsigned point) {
	return point < 0x80 ? 1 : point < 0x800 ? 2 : point < 0x10000 ? 3 : 4;
}

// decode makes a string of bytes decoded as WHATWG UTF-8, which the caller owns. What it makes is
// always valid UTF-8, so no lone surrogate can come in from outside.
static adamic_string *decode(const unsigned char *bytes, size_t length) {
	size_t size = 0;
	unsigned point;
	for (size_t offset = 0; offset < length;) {
		offset += decode_step(bytes, length, offset, &point);
		size += encoded_size(point);
	}
	adamic_string *string = new_string(size);
	unsigned char *cursor = (unsigned char *)string->bytes;
	for (size_t offset = 0; offset < length;) {
		offset += decode_step(bytes, length, offset, &point);
		switch (encoded_size(point)) {
		case 1:
			*cursor++ = (unsigned char)point;
			break;
		case 2:
			*cursor++ = (unsigned char)(0xc0 | (point >> 6));
			*cursor++ = (unsigned char)(0x80 | (point & 0x3f));
			break;
		case 3:
			*cursor++ = (unsigned char)(0xe0 | (point >> 12));
			*cursor++ = (unsigned char)(0x80 | ((point >> 6) & 0x3f));
			*cursor++ = (unsigned char)(0x80 | (point & 0x3f));
			break;
		default:
			*cursor++ = (unsigned char)(0xf0 | (point >> 18));
			*cursor++ = (unsigned char)(0x80 | ((point >> 12) & 0x3f));
			*cursor++ = (unsigned char)(0x80 | ((point >> 6) & 0x3f));
			*cursor++ = (unsigned char)(0x80 | (point & 0x3f));
			break;
		}
	}
	return string;
}

adamic_string *baseline_decode_utf8(const unsigned char *bytes, size_t length) {
	return decode(bytes, length);
}
