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
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Private UTF-16 cache access, kept in string_index.c with its lifetime.
const uint16_t *adamic_string_unit_view(const adamic_string *string);

static size_t sequence(unsigned char lead);
static unsigned decode(const unsigned char *bytes, size_t size);

#include "string_build_impl.h"

#include "string_decode_impl.h"

#include "string_trim_impl.h"

#include "string_walk_impl.h"

#include "string_builder_impl.h"

#include "string_slice_impl.h"

#include "string_repeat_impl.h"

#include "string_search_impl.h"

#include "string_replace_impl.h"

adamic_string *adamic_string_at_relative(const adamic_string *string, double index) {
	// string.at(index), as array.at: a fraction truncates, NaN is 0, a negative counts from the end.
	index = isnan(index) ? 0 : trunc(index);
	if (index < 0) {
		index += adamic_string_length(string);
	}
	return adamic_string_at(string, index);
}

bool adamic_string_starts_with(const adamic_string *string, const adamic_string *search) {
	if (halves_pairs(search)) {
		return affix(string, search, true);
	}
	return search->length <= string->length && memcmp(string->bytes, search->bytes, search->length) == 0;
}

bool adamic_string_ends_with(const adamic_string *string, const adamic_string *search) {
	if (halves_pairs(search)) {
		return affix(string, search, false);
	}
	return search->length <= string->length && memcmp(string->bytes + string->length - search->length, search->bytes, search->length) == 0;
}

#include "string_split_impl.h"

// inspect_quote is the quote util.inspect gives one string (Node's strEscape): a single quote, unless
// the string holds one; then a double quote, unless it holds one too; then a backtick, unless it holds
// a backtick or `${`; and else a single quote again, the string's own single quotes escaped.
static char inspect_quote(const unsigned char *bytes, size_t length) {
	if (memchr(bytes, '\'', length) == NULL) {
		return '\'';
	}
	if (memchr(bytes, '"', length) == NULL) {
		return '"';
	}
	if (memchr(bytes, '`', length) != NULL) {
		return '\'';
	}
	for (size_t i = 1; i < length; i++) {
		if (bytes[i - 1] == '$' && bytes[i] == '{') {
			return '\'';
		}
	}
	return '`';
}

// adamic_string_inspect_received is util.inspect(value) of a string, cut to 128 units and `...` as
// Node's ERR_INVALID_ARG_VALUE cuts its Received preview, a string the caller owns. Node v24.19.0,
// measured over every unit 0x00 to 0xFF: \b \t \n \f \r are letters; every other C0 control, DEL and
// the C1 controls U+0080 to U+009F are \xHH in uppercase, \v as \x0B; a backslash is doubled; a lone
// surrogate is \uXXXX in lowercase, where a pair stays raw; and every other unit is raw. A string past
// 76 units (formatPrimitive's test: inspect's default breakLength of 80, less 4, and its kMinLineLength
// of 16, which 76 passes) is split after each line feed, each piece quoted on its own and joined by
// ` +\n  `. The cut counts units of the quoted text, so it can fall inside an escape or between a
// pair's halves, as Node's does.
adamic_string *adamic_string_inspect_received(const adamic_string *value) {
	static const char hex[] = "0123456789ABCDEF";
	// The walk takes a step only while it has written at most 128 units, and a unit takes at most 3
	// bytes: a raw 1- to 3-byte character is one unit, a raw 4-byte one two, and an escape, a quote and
	// a join a byte a unit. Its last step is at most 6 bytes, a `\uXXXX` escape or a join and the next
	// piece's opening quote, and then the piece's closing quote, so it writes at most 3 * 128 + 6 + 1.
	// The `...` is added to the string, never to shown.
	enum { most_bytes = 3 * 128 + 6 + 1 };
	_Static_assert(sizeof " +\n  " - 1 + 1 <= 6 && sizeof "\\uXXXX" - 1 <= 6, "a last step is at most 6 bytes");
	char shown[most_bytes];
	size_t at = 0, units = 0;
	const unsigned char *bytes = (const unsigned char *)value->bytes;
	bool split = adamic_string_length(value) > 76;
	size_t start = 0;
	do {
		size_t end = value->length;
		const unsigned char *line = split ? memchr(bytes + start, '\n', value->length - start) : NULL;
		if (line != NULL && (size_t)(line - bytes) + 1 < value->length) {
			end = (size_t)(line - bytes) + 1;
		}
		char quote = inspect_quote(bytes + start, end - start);
		if (start > 0) {
			memcpy(shown + at, " +\n  ", 5);
			at += 5;
			units += 5;
		}
		shown[at++] = quote;
		units++;
		size_t i = start;
		while (i < end && units <= 128) {
			size_t size = sequence(bytes[i]);
			unsigned point = decode(bytes + i, size);
			if (point == '\\' || point == (unsigned char)quote) {
				shown[at++] = '\\';
				shown[at++] = (char)point;
				units += 2;
			} else if (point == '\b' || point == '\t' || point == '\n' || point == '\f' || point == '\r') {
				shown[at++] = '\\';
				shown[at++] = point == '\b' ? 'b' : point == '\t' ? 't' : point == '\n' ? 'n' : point == '\f' ? 'f' : 'r';
				units += 2;
			} else if (point < 0x20 || (point >= 0x7f && point <= 0x9f)) {
				shown[at++] = '\\';
				shown[at++] = 'x';
				shown[at++] = hex[point >> 4];
				shown[at++] = hex[point & 15];
				units += 4;
			} else if (point >= 0xd800 && point <= 0xdfff) {
				// Only a lone surrogate decodes to one: a pair is stored as its 4-byte code point.
				static const char lower[] = "0123456789abcdef";
				shown[at++] = '\\';
				shown[at++] = 'u';
				for (int shift = 12; shift >= 0; shift -= 4) {
					shown[at++] = lower[(point >> (unsigned)shift) & 15];
				}
				units += 6;
			} else {
				memcpy(shown + at, bytes + i, size);
				at += size;
				units += size == 4 ? 2 : 1;
			}
			i += size;
		}
		if (i < end) {
			break;
		}
		shown[at++] = quote;
		units++;
		start = end;
	} while (start < value->length && units <= 128);
	adamic_string *inspection = adamic_decode_utf8((const unsigned char *)shown, at);
	if (adamic_string_length(inspection) <= 128) {
		return inspection;
	}
	adamic_string *cut = adamic_string_slice(inspection, 0, 128, true);
	adamic_release(inspection);
	static adamic_string dots = ADAMIC_STRING("...");
	adamic_string *preview = adamic_string_concat(2, (adamic_string *const[]){cut, &dots});
	adamic_release(cut);
	return preview;
}
