// case.c: toUpperCase and toLowerCase, as Node does them.
//
// JavaScript's case mapping is locale-independent and full: a code point can map to more than one
// (ß uppercases to SS, İ lowercases to i and a combining dot), and Σ lowercases to ς at the end of a
// word (SpecialCasing.txt's Final_Sigma). The tables are Unicode's, for the version Node carries,
// written by case_generate.go into case_tables.h. A lone surrogate has no case and passes through as
// the WTF-8 bytes it is.

#include "adamic.h"

#include <stdint.h>
#include <string.h>

typedef struct case_run {
	uint32_t start;
	uint16_t count;
	uint8_t stride;
	int32_t delta;
} case_run;

typedef struct case_full {
	uint32_t point;
	uint8_t count;
	uint32_t mapped[3];
} case_full;

typedef struct case_span {
	uint32_t first;
	uint32_t last;
} case_span;

#include "case_tables.h"

#define COUNT(table) (sizeof table / sizeof table[0])

// simple is a code point's one-to-one mapping in a table of runs, or the code point itself.
static uint32_t simple(const case_run *runs, size_t count, uint32_t point) {
	// The last run starting at or before point is the only one that can hold it.
	size_t low = 0, high = count;
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (runs[middle].start <= point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	if (low == 0) {
		return point;
	}
	const case_run *run = &runs[low - 1];
	uint32_t offset = point - run->start;
	if (offset % run->stride != 0 || offset / run->stride >= run->count) {
		return point;
	}
	return (uint32_t)((int32_t)point + run->delta);
}

// full is a code point's mapping to more than one code point, or NULL.
static const case_full *full(const case_full *mappings, size_t count, uint32_t point) {
	size_t low = 0, high = count;
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (mappings[middle].point < point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	return low < count && mappings[low].point == point ? &mappings[low] : NULL;
}

static bool within(const case_span *spans, size_t count, uint32_t point) {
	size_t low = 0, high = count;
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (spans[middle].last < point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	return low < count && spans[low].first <= point;
}

// A code point's bytes: 1 to 4 of UTF-8, or 3 of WTF-8 for a lone surrogate.
static size_t size_at(unsigned char lead) {
	return lead < 0x80 ? 1 : lead < 0xe0 ? 2 : lead < 0xf0 ? 3 : 4;
}

static uint32_t decode_at(const unsigned char *bytes, size_t size) {
	switch (size) {
	case 1:
		return bytes[0];
	case 2:
		return ((uint32_t)(bytes[0] & 0x1f) << 6) | (bytes[1] & 0x3f);
	case 3:
		return ((uint32_t)(bytes[0] & 0x0f) << 12) | ((uint32_t)(bytes[1] & 0x3f) << 6) | (bytes[2] & 0x3f);
	}
	return ((uint32_t)(bytes[0] & 0x07) << 18) | ((uint32_t)(bytes[1] & 0x3f) << 12) | ((uint32_t)(bytes[2] & 0x3f) << 6) | (bytes[3] & 0x3f);
}

static size_t encoded_size(uint32_t point) {
	return point < 0x80 ? 1 : point < 0x800 ? 2 : point < 0x10000 ? 3 : 4;
}

static size_t encode(uint32_t point, char *out) {
	size_t size = encoded_size(point);
	switch (size) {
	case 1:
		out[0] = (char)point;
		break;
	case 2:
		out[0] = (char)(0xc0 | (point >> 6));
		out[1] = (char)(0x80 | (point & 0x3f));
		break;
	case 3:
		out[0] = (char)(0xe0 | (point >> 12));
		out[1] = (char)(0x80 | ((point >> 6) & 0x3f));
		out[2] = (char)(0x80 | (point & 0x3f));
		break;
	default:
		out[0] = (char)(0xf0 | (point >> 18));
		out[1] = (char)(0x80 | ((point >> 12) & 0x3f));
		out[2] = (char)(0x80 | ((point >> 6) & 0x3f));
		out[3] = (char)(0x80 | (point & 0x3f));
	}
	return size;
}

// cased_beyond is Final_Sigma's test, as ICU (and so Node) makes it: walking away from the sigma,
// forward or back, skip every case-ignorable code point, and report whether the first other one is
// cased. A code point that is both, like ʰ, is skipped: Node lowercases ʰΣ to ʰσ, not ʰς.
static bool cased_beyond(const adamic_string *string, size_t offset, size_t size, bool forward) {
	const unsigned char *bytes = (const unsigned char *)string->bytes;
	if (forward) {
		for (size_t at = offset + size; at < string->length;) {
			size_t step = size_at(bytes[at]);
			uint32_t point = decode_at(bytes + at, step);
			if (!within(case_ignorable, COUNT(case_ignorable), point)) {
				return within(case_cased, COUNT(case_cased), point);
			}
			at += step;
		}
		return false;
	}
	for (size_t end = offset; end > 0;) {
		// Back up to the start of the code point before end: continuation bytes are 10xxxxxx.
		size_t at = end - 1;
		while (at > 0 && (bytes[at] & 0xc0) == 0x80) {
			at--;
		}
		uint32_t point = decode_at(bytes + at, end - at);
		if (!within(case_ignorable, COUNT(case_ignorable), point)) {
			return within(case_cased, COUNT(case_cased), point);
		}
		end = at;
	}
	return false;
}

// map_point writes what one code point maps to into out (room for 12 bytes) and returns how many bytes.
static size_t map_point(const adamic_string *string, size_t offset, size_t size, bool upper, char *out) {
	const unsigned char *bytes = (const unsigned char *)string->bytes + offset;
	uint32_t point = decode_at(bytes, size);
	if (point >= 0xd800 && point <= 0xdfff) {
		// A lone surrogate: its own WTF-8 bytes, unchanged.
		memcpy(out, bytes, size);
		return size;
	}
	if (!upper && point == 0x3a3 && cased_beyond(string, offset, size, false) && !cased_beyond(string, offset, size, true)) {
		return encode(0x3c2, out);
	}
	const case_full *mapping = upper ? full(case_full_upper, COUNT(case_full_upper), point) : full(case_full_lower, COUNT(case_full_lower), point);
	if (mapping != NULL) {
		size_t written = 0;
		for (size_t index = 0; index < mapping->count; index++) {
			written += encode(mapping->mapped[index], out + written);
		}
		return written;
	}
	return encode(upper ? simple(case_upper, COUNT(case_upper), point) : simple(case_lower, COUNT(case_lower), point), out);
}

static adamic_string *convert(const adamic_string *string, bool upper) {
	// ASCII maps within ASCII, byte for byte; anything else goes through the tables, measured first so
	// the result is allocated once.
	bool ascii = true;
	for (size_t at = 0; at < string->length && ascii; at++) {
		ascii = (unsigned char)string->bytes[at] < 0x80;
	}
	if (ascii) {
		adamic_string *result = adamic_string_allocate(string->length);
		char *out = (char *)result->bytes;
		for (size_t at = 0; at < string->length; at++) {
			char character = string->bytes[at];
			if (upper && character >= 'a' && character <= 'z') {
				character = (char)(character - 'a' + 'A');
			} else if (!upper && character >= 'A' && character <= 'Z') {
				character = (char)(character - 'A' + 'a');
			}
			out[at] = character;
		}
		return result;
	}
	char scratch[12];
	size_t length = 0;
	for (size_t at = 0; at < string->length;) {
		size_t size = size_at((unsigned char)string->bytes[at]);
		length += map_point(string, at, size, upper, scratch);
		at += size;
	}
	adamic_string *result = adamic_string_allocate(length);
	char *out = (char *)result->bytes;
	for (size_t at = 0; at < string->length;) {
		size_t size = size_at((unsigned char)string->bytes[at]);
		out += map_point(string, at, size, upper, out);
		at += size;
	}
	// A mapping can lengthen a string (ß to SS) past V8's longest; units are never more than bytes.
	if (length > ADAMIC_STRING_MAX_UNITS) {
		adamic_string_check_length(adamic_string_length(result));
	}
	return result;
}

adamic_string *adamic_string_to_upper(const adamic_string *string) {
	return convert(string, true);
}

adamic_string *adamic_string_to_lower(const adamic_string *string) {
	return convert(string, false);
}
