// utf8Length and utf8At from 'adamic': a string's UTF-8, read in place.
//
// A string is stored as its UTF-8 already, so the view copies nothing. The one difference is a lone
// surrogate, which UTF-8 can't hold: Adamic stores one as WTF-8's three bytes (ED A0 80 to ED BF BF),
// and the WHATWG encoder (TextEncoder, Node's Buffer) writes it as U+FFFD's three, EF BF BD. The two
// are the same length, so the view reads those three bytes as U+FFFD's, and agrees with Node's.

#include "adamic.h"

#include <math.h>
#include <stdlib.h>
#include <string.h>

double adamic_utf8_length(const adamic_string *text) {
	return (double)text->length;
}

double adamic_utf8_at(const adamic_string *text, double index) {
	if (!(index >= 0 && index < (double)text->length && index == trunc(index))) {
		// Past the end, before the start or between bytes: a panic, as Go's text[index] is, in words
		// both backends write the same.
		adamic_string *written = adamic_string_from_number(index);
		adamic_string *size = adamic_string_from_number((double)text->length);
		static const char before[] = "RangeError: utf8At index ";
		static const char middle[] = " is not a byte of a text of ";
		static const char after[] = " bytes";
		size_t length = sizeof before - 1 + written->length + sizeof middle - 1 + size->length + sizeof after - 1;
		char *message = malloc(length);
		if (message == NULL) {
			static const char failed[] = "out of memory";
			adamic_panic(failed, sizeof failed - 1);
		}
		char *cursor = message;
		memcpy(cursor, before, sizeof before - 1);
		cursor += sizeof before - 1;
		memcpy(cursor, written->bytes, written->length);
		cursor += written->length;
		memcpy(cursor, middle, sizeof middle - 1);
		cursor += sizeof middle - 1;
		memcpy(cursor, size->bytes, size->length);
		cursor += size->length;
		memcpy(cursor, after, sizeof after - 1);
		adamic_panic(message, length);
	}
	const unsigned char *bytes = (const unsigned char *)text->bytes;
	size_t at = (size_t)index;
	// ED never continues a sequence, so an ED one or two bytes back leads the sequence this byte is in;
	// with a second byte from A0 to BF, that sequence is a lone surrogate (a pair is four bytes).
	for (size_t back = 0; back < 3 && back <= at; back++) {
		size_t lead = at - back;
		if (bytes[lead] == 0xed && lead + 2 < text->length && bytes[lead + 1] >= 0xa0 && bytes[lead + 1] <= 0xbf) {
			static const unsigned char replacement[] = {0xef, 0xbf, 0xbd};
			return replacement[back];
		}
		if (bytes[lead] >= 0xc0) {
			break;
		}
	}
	return bytes[at];
}
