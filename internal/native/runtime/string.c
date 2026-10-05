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

static size_t sequence(unsigned char lead);
static unsigned decode(const unsigned char *bytes, size_t size);

// allocate makes a string of length bytes, references 1, its bytes right after it in one block.
static adamic_string *allocate(size_t length) {
	adamic_string *string = adamic_allocate(sizeof *string + length, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	string->units = 0;
	string->index = NULL;
	string->owner = NULL;
	string->capacity = length;
	return string;
}

adamic_string *adamic_string_allocate(size_t length) {
	return allocate(length);
}

adamic_string *adamic_string_from_number(double value) {
	char buffer[ADAMIC_NUMBER_FORMAT_MAX];
	size_t length = adamic_number_format(value, buffer);
	adamic_string *string = allocate(length);
	memcpy((char *)string->bytes, buffer, length);
	return string;
}

// V8's longest string, in UTF-16 units (String::kMaxLength on 64-bit): past it, every way of making
// a string throws RangeError: Invalid string length, and so Adamic panics there too.
void adamic_string_check_length(double units) {
	if (units > ADAMIC_STRING_MAX_UNITS) {
		static const char message[] = "RangeError: Invalid string length";
		adamic_panic(message, sizeof message - 1);
	}
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
	// A string's UTF-16 units are never more than its bytes, so only a long one needs counting.
	if (length > ADAMIC_STRING_MAX_UNITS) {
		double units = 0;
		for (size_t index = 0; index < count; index++) {
			units += adamic_string_length(parts[index]);
		}
		adamic_string_check_length(units);
	}
	adamic_string *string = allocate(length);
	if (count == 1) {
		// One piece is a copy, and its bytes may be a builder's (a stack piece, from fromCharCode or
		// slice), which can hold halves of a pair side by side: every byte is looked at.
		if (length > 0) {
			memcpy((char *)string->bytes, parts[0]->bytes, length);
		}
		string->length = adamic_string_join_halves((char *)string->bytes, 0, length);
		return string;
	}
	size_t written = 0;
	for (size_t index = 0; index < count; index++) {
		written = adamic_string_put((char *)string->bytes, written, parts[index]);
	}
	string->length = written;
	return string;
}

size_t adamic_string_put(char *bytes, size_t written, const adamic_string *part) {
	// A string's own halves are joined already, so halves of a pair can meet only where two pieces
	// do: a lone high surrogate the bytes so far end with, and a lone low one the part begins with.
	// The part's first three bytes go in first, and only those six are looked at.
	size_t head = part->length < 3 ? part->length : 3;
	if (head > 0) {
		memcpy(bytes + written, part->bytes, head);
	}
	written = adamic_string_join_halves(bytes, written >= 3 ? written - 3 : 0, written + head);
	if (part->length > head) {
		memcpy(bytes + written, part->bytes + head, part->length - head);
	}
	return written + part->length - head;
}

size_t adamic_string_join_halves(char *bytes, size_t from, size_t length) {
	// A lone high surrogate followed by a lone low one is a character again, as in JavaScript, and
	// WTF-8 writes it as UTF-8's four bytes, so equal strings stay equal byte for byte.
	size_t kept = from;
	for (size_t at = from; at < length;) {
		unsigned char *here = (unsigned char *)bytes + at;
		if (at + 6 <= length && here[0] == 0xed && here[1] >= 0xa0 && here[1] <= 0xaf && here[3] == 0xed && here[4] >= 0xb0 && here[4] <= 0xbf) {
			unsigned high = decode(here, 3), low = decode(here + 3, 3);
			unsigned point = 0x10000 + ((high - 0xd800) << 10) + (low - 0xdc00);
			bytes[kept++] = (char)(0xf0 | (point >> 18));
			bytes[kept++] = (char)(0x80 | ((point >> 12) & 0x3f));
			bytes[kept++] = (char)(0x80 | ((point >> 6) & 0x3f));
			bytes[kept++] = (char)(0x80 | (point & 0x3f));
			at += 6;
			continue;
		}
		bytes[kept++] = bytes[at++];
	}
	return kept;
}

int adamic_string_equal(const adamic_string *left, const adamic_string *right) {
	// Either may be undefined (a string | undefined): undefined is equal only to itself.
	if (left == NULL || right == NULL) {
		return left == right;
	}
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
	return (double)adamic_string_units(string);
}

// unit_at is the UTF-16 unit at an index below the length: a code point of the BMP or a lone
// surrogate, or one half of a pair.
static unsigned unit_at(const adamic_string *string, size_t index) {
	bool low;
	size_t offset = adamic_string_locate(string, index, &low);
	size_t size = sequence((unsigned char)string->bytes[offset]);
	unsigned point = decode((const unsigned char *)string->bytes + offset, size);
	if (size == 4) {
		return low ? 0xdc00 + ((point - 0x10000) & 0x3ff) : 0xd800 + ((point - 0x10000) >> 10);
	}
	return point;
}

double adamic_string_char_code_at(const adamic_string *string, double position) {
	// ToIntegerOrInfinity: NaN is 0, and a fraction truncates.
	position = isnan(position) ? 0 : trunc(position);
	if (position < 0 || position >= (double)adamic_string_units(string)) {
		return NAN;
	}
	return (double)unit_at(string, (size_t)position);
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
	return adamic_string_trim_sides(string, true, true);
}

adamic_string *adamic_string_trim_sides(adamic_string *string, bool at_start, bool at_end) {
	size_t start = 0, end = string->length;
	while (at_start && start < end) {
		size_t size = sequence((unsigned char)string->bytes[start]);
		if (!is_space(decode((const unsigned char *)string->bytes + start, size))) {
			break;
		}
		start += size;
	}
	while (at_end && end > start) {
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
	return adamic_string_share(string, start, end - start);
}

size_t adamic_string_next(const adamic_string *string, size_t offset) {
	return sequence((unsigned char)string->bytes[offset]);
}

adamic_string *adamic_string_slice_bytes(const adamic_string *string, size_t offset, size_t size) {
	return adamic_string_share(string, offset, size);
}

// A view of a string as UTF-16 code units, walked from the start. Each step yields the unit's value
// and where its code point's bytes are; a supplementary code point is two steps over the same bytes.
typedef struct units {
	const adamic_string *string;
	size_t offset;   // bytes of the current code point
	size_t size;     // its byte length
	bool second;     // on the low half of a surrogate pair
} units;

static bool units_next(units *walk, unsigned *unit) {
	if (walk->second) {
		unsigned point = decode((const unsigned char *)walk->string->bytes + walk->offset, walk->size);
		*unit = 0xdc00 + ((point - 0x10000) & 0x3ff);
		walk->second = false;
		walk->offset += walk->size;
		walk->size = 0;
		return true;
	}
	walk->offset += walk->size;
	if (walk->offset >= walk->string->length) {
		walk->size = 0;
		return false;
	}
	walk->size = sequence((unsigned char)walk->string->bytes[walk->offset]);
	unsigned point = decode((const unsigned char *)walk->string->bytes + walk->offset, walk->size);
	if (walk->size == 4) {
		// The high half now, and the low half from the same bytes on the next step.
		*unit = 0xd800 + ((point - 0x10000) >> 10);
		walk->second = true;
		return true;
	}
	*unit = point;
	return true;
}

static units units_start(const adamic_string *string) {
	units walk = {string, 0, 0, false};
	return walk;
}

// A growing buffer for building strings, WTF-8 included.
typedef struct builder {
	char *bytes;
	size_t length;
	size_t capacity;
} builder;

static void builder_add(builder *build, const char *bytes, size_t size) {
	if (build->length + size > build->capacity) {
		size_t capacity = build->capacity == 0 ? 32 : build->capacity;
		while (capacity < build->length + size) {
			capacity *= 2;
		}
		char *grown = realloc(build->bytes, capacity);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		build->bytes = grown;
		build->capacity = capacity;
	}
	if (size > 0) {
		memcpy(build->bytes + build->length, bytes, size);
	}
	build->length += size;
}

// builder_unit appends one UTF-16 code unit: a BMP character, or a lone surrogate as WTF-8's three
// bytes (ED A0 80 to ED BF BF), which no valid UTF-8 contains, so it can never be mistaken.
static void builder_unit(builder *build, unsigned unit) {
	char bytes[3];
	if (unit < 0x80) {
		bytes[0] = (char)unit;
		builder_add(build, bytes, 1);
	} else if (unit < 0x800) {
		bytes[0] = (char)(0xc0 | (unit >> 6));
		bytes[1] = (char)(0x80 | (unit & 0x3f));
		builder_add(build, bytes, 2);
	} else {
		bytes[0] = (char)(0xe0 | (unit >> 12));
		bytes[1] = (char)(0x80 | ((unit >> 6) & 0x3f));
		bytes[2] = (char)(0x80 | (unit & 0x3f));
		builder_add(build, bytes, 3);
	}
}

static adamic_string *builder_finish(builder *build) {
	adamic_string piece = {{0, adamic_kind_string}, build->length, build->bytes, 0, NULL, NULL, 0};
	adamic_string *string = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(build->bytes);
	return string;
}

// clamp_index is ECMAScript's relative index: ToIntegerOrInfinity, negative counts from the end, then
// clamped to [0, length].
static double clamp_index(double index, double length) {
	index = isnan(index) ? 0 : trunc(index);
	if (index < 0) {
		index = length + index < 0 ? 0 : length + index;
	}
	return index > length ? length : index;
}

adamic_string *adamic_string_slice(const adamic_string *string, double start, double end, bool has_end) {
	double length = adamic_string_length(string);
	double from = clamp_index(start, length);
	double to = has_end ? clamp_index(end, length) : length;
	if (!(from < to)) {
		return allocate(0);
	}
	// Whole code points are their bytes, in one piece, shared with the string where that's worth it
	// (string_share.c). A slice that starts on the low half of a pair begins with that half, and one
	// that ends between the halves ends with the high one, each a lone surrogate, which isn't in the
	// string's bytes, so that slice is built.
	size_t first = (size_t)from, last = (size_t)to;
	bool low;
	size_t offset = adamic_string_locate(string, first, &low);
	if (low) {
		offset += 4;
	}
	bool ends_low = false;
	size_t stop = last < (size_t)length ? adamic_string_locate(string, last, &ends_low) : string->length;
	size_t middle = stop > offset ? stop - offset : 0;
	if (!low && !ends_low) {
		return adamic_string_share(string, offset, middle);
	}
	builder build = {NULL, 0, 0};
	if (low) {
		builder_unit(&build, unit_at(string, first));
	}
	if (middle > 0) {
		builder_add(&build, string->bytes + offset, middle);
	}
	if (ends_low) {
		builder_unit(&build, unit_at(string, last - 1));
	}
	return builder_finish(&build);
}

adamic_string *adamic_string_at(const adamic_string *string, double index) {
	// As for an array: an index is an integer from 0 up to the length in UTF-16 units, and anything
	// else (negative, a fraction, NaN, past the end) is a property the string doesn't have. Half of a
	// supplementary character is a lone surrogate, as slice makes it.
	if (!(index >= 0) || index != trunc(index) || index >= adamic_string_length(string)) {
		return NULL;
	}
	return adamic_string_slice(string, index, index + 1, true);
}

adamic_maybe_number adamic_string_code_point_at(const adamic_string *string, double position) {
	position = isnan(position) ? 0 : trunc(position);
	adamic_maybe_number missing = {false, 0};
	if (position < 0 || position >= (double)adamic_string_units(string)) {
		return missing;
	}
	// At the high half of a pair, the whole code point; anywhere else, the unit.
	bool low;
	size_t offset = adamic_string_locate(string, (size_t)position, &low);
	size_t size = sequence((unsigned char)string->bytes[offset]);
	unsigned point = low || size != 4 ? unit_at(string, (size_t)position) : decode((const unsigned char *)string->bytes + offset, size);
	adamic_maybe_number found = {true, (double)point};
	return found;
}

int adamic_string_compare(const adamic_string *left, const adamic_string *right) {
	// UTF-16 code unit order, as JavaScript's < compares: not UTF-8's byte order, which puts U+FFFF
	// before an emoji where UTF-16 puts it after.
	units left_walk = units_start(left), right_walk = units_start(right);
	for (;;) {
		unsigned left_unit, right_unit;
		bool left_more = units_next(&left_walk, &left_unit);
		bool right_more = units_next(&right_walk, &right_unit);
		if (!left_more || !right_more) {
			return left_more ? 1 : right_more ? -1 : 0;
		}
		if (left_unit != right_unit) {
			return left_unit < right_unit ? -1 : 1;
		}
	}
}

// repeat_unchecked is count copies of string, count already a whole number in range.
static adamic_string *repeat_unchecked(const adamic_string *string, double count) {
	builder build = {NULL, 0, 0};
	for (double index = 0; index < count; index++) {
		builder_add(&build, string->bytes, string->length);
	}
	return builder_finish(&build);
}

adamic_string *adamic_string_repeat(const adamic_string *string, double count) {
	count = isnan(count) ? 0 : trunc(count);
	if (count < 0 || isinf(count)) {
		char message[96];
		char number[ADAMIC_NUMBER_FORMAT_MAX];
		size_t length = adamic_number_format(count, number);
		int written = snprintf(message, sizeof message, "RangeError: Invalid count value: %.*s", (int)length, number);
		adamic_panic(message, (size_t)written);
	}
	// Checked before any of it is built, as V8 does: an empty string repeats to itself however many
	// times, and anything longer than V8's longest string is refused.
	if (string->length == 0 || count == 0) {
		return adamic_retain((adamic_string *)&adamic_string_empty);
	}
	adamic_string_check_length(adamic_string_length(string) * count);
	return repeat_unchecked(string, count);
}

adamic_string *adamic_string_pad(const adamic_string *string, double target, const adamic_string *fill, bool at_start) {
	double length = adamic_string_length(string);
	target = isnan(target) ? 0 : trunc(target);
	if (target <= length || fill->length == 0) {
		return adamic_retain((adamic_string *)string);
	}
	adamic_string_check_length(target);
	// The fill, repeated and cut to exactly the missing number of UTF-16 units: whole fills, then the
	// start of one more, so nothing on the way is longer than the result.
	double missing = target - length;
	double fill_length = adamic_string_length(fill);
	double whole = floor(missing / fill_length);
	adamic_string *repeated = repeat_unchecked(fill, whole);
	adamic_string *rest = adamic_string_slice(fill, 0, missing - whole * fill_length, true);
	adamic_string *parts[3] = {repeated, rest, (adamic_string *)string};
	if (!at_start) {
		parts[0] = (adamic_string *)string;
		parts[1] = repeated;
		parts[2] = rest;
	}
	adamic_string *padded = adamic_string_concat(3, parts);
	adamic_release(repeated);
	adamic_release(rest);
	return padded;
}

// Searching by WTF-8 bytes finds exactly what searching by UTF-16 units does, with one exception: a
// search that begins with a lone low surrogate or ends with a lone high one can match half of a
// supplementary character, which UTF-16 sees and the bytes (four of them, whole) don't. Only those
// searches go unit by unit.
static bool halves_pairs(const adamic_string *search) {
	units walk = units_start(search);
	unsigned unit, first = 0, last = 0;
	bool any = false;
	while (units_next(&walk, &unit)) {
		if (!any) {
			first = unit;
		}
		last = unit;
		any = true;
	}
	return any && ((first >= 0xdc00 && first <= 0xdfff) || (last >= 0xd800 && last <= 0xdbff));
}

// to_units is a string's UTF-16 code units, in a buffer the caller frees.
static unsigned *to_units(const adamic_string *string, size_t *count) {
	// A string's units are never more than its bytes.
	unsigned *buffer = malloc((string->length + 1) * sizeof *buffer);
	if (buffer == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	units walk = units_start(string);
	*count = 0;
	while (units_next(&walk, &buffer[*count])) {
		(*count)++;
	}
	return buffer;
}

// unit_index_of is the first UTF-16 index at or after from where search's units are, or -1.
static double unit_index_of(const unsigned *haystack, size_t haystack_count, const unsigned *needle, size_t needle_count, size_t from) {
	for (size_t at = from; at + needle_count <= haystack_count; at++) {
		if (memcmp(haystack + at, needle, needle_count * sizeof *needle) == 0) {
			return (double)at;
		}
	}
	return -1;
}

double adamic_string_index_of(const adamic_string *string, const adamic_string *search) {
	return adamic_string_index_of_at(string, search, 0);
}

double adamic_string_index_of_at(const adamic_string *string, const adamic_string *search, size_t from) {
	if (search->length == 0) {
		return (double)from;
	}
	if (halves_pairs(search)) {
		size_t haystack_count, needle_count;
		unsigned *haystack = to_units(string, &haystack_count), *needle = to_units(search, &needle_count);
		double found = unit_index_of(haystack, haystack_count, needle, needle_count, from);
		free(haystack);
		free(needle);
		return found;
	}
	size_t start = 0;
	if (from > 0) {
		if (from >= adamic_string_units(string)) {
			return -1;
		}
		// From the low half of a pair, the search can begin only at the next code point: it doesn't
		// begin with a low half, or halves_pairs would have said so.
		bool low;
		start = adamic_string_locate(string, from, &low);
		if (low) {
			start += 4;
		}
	}
	for (size_t offset = start; offset + search->length <= string->length;) {
		if (memcmp(string->bytes + offset, search->bytes, search->length) == 0) {
			return (double)adamic_string_units_before(string, offset);
		}
		offset += sequence((unsigned char)string->bytes[offset]);
	}
	return -1;
}

// affix reports whether search's units are string's first (at_start) or last ones, unit by unit.
static bool affix(const adamic_string *string, const adamic_string *search, bool at_start) {
	size_t haystack_count, needle_count;
	unsigned *haystack = to_units(string, &haystack_count), *needle = to_units(search, &needle_count);
	bool found = needle_count <= haystack_count &&
		memcmp(haystack + (at_start ? 0 : haystack_count - needle_count), needle, needle_count * sizeof *needle) == 0;
	free(haystack);
	free(needle);
	return found;
}

double adamic_string_last_index_of(const adamic_string *string, const adamic_string *search) {
	size_t haystack_count, needle_count;
	unsigned *haystack = to_units(string, &haystack_count), *needle = to_units(search, &needle_count);
	double found = -1;
	if (needle_count <= haystack_count) {
		for (size_t at = haystack_count - needle_count + 1; at-- > 0;) {
			if (memcmp(haystack + at, needle, needle_count * sizeof *needle) == 0) {
				found = (double)at;
				break;
			}
		}
	}
	free(haystack);
	free(needle);
	return found;
}

// pieces collects the strings a result is joined from, each owned until the join.
typedef struct pieces {
	adamic_string **items;
	size_t count;
	size_t capacity;
} pieces;

static void pieces_add(pieces *list, adamic_string *piece) {
	if (list->count == list->capacity) {
		list->capacity = list->capacity == 0 ? 8 : list->capacity * 2;
		adamic_string **grown = realloc(list->items, list->capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		list->items = grown;
	}
	list->items[list->count++] = piece;
}

// pieces_join joins the pieces (halves of a pair glued where they meet) and lets go of them.
static adamic_string *pieces_join(pieces *list) {
	adamic_string *joined = adamic_string_concat(list->count, list->items);
	for (size_t index = 0; index < list->count; index++) {
		adamic_release(list->items[index]);
	}
	free(list->items);
	return joined;
}

// units_piece is the string of the UTF-16 units from one index up to another, made from the units
// already in hand, so it costs what it holds: slicing the string itself would walk it from the
// start each time, and replaceAll, which cuts a piece per match, would be quadratic. The halves of a
// pair that both fall inside are glued back into one character by the builder's finish.
static adamic_string *units_piece(const unsigned *units, size_t from, size_t to) {
	builder build = {NULL, 0, 0};
	for (size_t index = from; index < to; index++) {
		builder_unit(&build, units[index]);
	}
	return builder_finish(&build);
}

// substitution is ECMAScript's GetSubstitution for a match of a string pattern, from start to end
// in UTF-16 units: $$ is $, $& the match, $` what precedes it and $' what follows. A string pattern
// has no groups, so $1 and $< are themselves, like every other character.
static adamic_string *substitution(const unsigned *units, const adamic_string *replacement, size_t start, size_t end, size_t length) {
	if (memchr(replacement->bytes, '$', replacement->length) == NULL) {
		return adamic_retain((adamic_string *)replacement);
	}
	pieces list = {NULL, 0, 0};
	size_t literal = 0;
	for (size_t at = 0; at < replacement->length; at++) {
		if (replacement->bytes[at] != '$' || at + 1 == replacement->length) {
			continue;
		}
		adamic_string *expanded = NULL;
		switch (replacement->bytes[at + 1]) {
		case '$':
			expanded = adamic_string_slice_bytes(replacement, at, 1);
			break;
		case '&':
			expanded = units_piece(units, start, end);
			break;
		case '`':
			expanded = units_piece(units, 0, start);
			break;
		case '\'':
			expanded = units_piece(units, end, length);
			break;
		}
		if (expanded == NULL) {
			continue;
		}
		pieces_add(&list, adamic_string_slice_bytes(replacement, literal, at - literal));
		pieces_add(&list, expanded);
		at++;
		literal = at + 1;
	}
	pieces_add(&list, adamic_string_slice_bytes(replacement, literal, replacement->length - literal));
	return pieces_join(&list);
}

adamic_string *adamic_string_replace(const adamic_string *string, const adamic_string *search, const adamic_string *replacement, bool all) {
	// By UTF-16 units throughout: an empty search matches between every unit, the halves of a pair
	// included, as JavaScript's does.
	size_t haystack_count, needle_count;
	unsigned *haystack = to_units(string, &haystack_count), *needle = to_units(search, &needle_count);
	pieces list = {NULL, 0, 0};
	size_t kept = 0;
	for (double found = unit_index_of(haystack, haystack_count, needle, needle_count, 0); found >= 0;) {
		size_t end = (size_t)found + needle_count;
		pieces_add(&list, units_piece(haystack, kept, (size_t)found));
		pieces_add(&list, substitution(haystack, replacement, (size_t)found, end, haystack_count));
		kept = end;
		if (!all) {
			break;
		}
		// The next search starts past the match, or one unit on from an empty one.
		size_t next = (size_t)found + (needle_count == 0 ? 1 : needle_count);
		found = next > haystack_count ? -1 : unit_index_of(haystack, haystack_count, needle, needle_count, next);
	}
	pieces_add(&list, units_piece(haystack, kept, haystack_count));
	free(haystack);
	free(needle);
	return pieces_join(&list);
}

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

adamic_array *adamic_string_code_points(const adamic_string *string) {
	adamic_array *array = adamic_array_new(0, true);
	for (size_t offset = 0; offset < string->length;) {
		size_t size = sequence((unsigned char)string->bytes[offset]);
		adamic_array_push(array, (adamic_value){.reference = adamic_string_slice_bytes(string, offset, size)});
		offset += size;
	}
	return array;
}

adamic_array *adamic_string_split(const adamic_string *string, const adamic_string *separator) {
	adamic_array *parts = adamic_array_new(0, true);
	if (separator->length == 0) {
		// An empty separator splits into UTF-16 code units, an emoji into its two halves.
		units walk = units_start(string);
		unsigned unit;
		while (units_next(&walk, &unit)) {
			builder build = {NULL, 0, 0};
			builder_unit(&build, unit);
			adamic_array_push(parts, (adamic_value){.reference = builder_finish(&build)});
		}
		return parts;
	}
	if (halves_pairs(separator)) {
		// Unit by unit, each part cut by UTF-16 index, so a half of a pair stays a lone surrogate.
		size_t haystack_count, needle_count;
		unsigned *haystack = to_units(string, &haystack_count), *needle = to_units(separator, &needle_count);
		size_t from = 0;
		for (;;) {
			double found = unit_index_of(haystack, haystack_count, needle, needle_count, from);
			double end = found < 0 ? (double)haystack_count : found;
			adamic_array_push(parts, (adamic_value){.reference = adamic_string_slice(string, (double)from, end, true)});
			if (found < 0) {
				break;
			}
			from = (size_t)found + needle_count;
		}
		free(haystack);
		free(needle);
		return parts;
	}
	size_t start = 0;
	for (size_t at = 0; at + separator->length <= string->length;) {
		if (memcmp(string->bytes + at, separator->bytes, separator->length) == 0) {
			adamic_array_push(parts, (adamic_value){.reference = adamic_string_slice_bytes(string, start, at - start)});
			at += separator->length;
			start = at;
		} else {
			at += sequence((unsigned char)string->bytes[at]);
		}
	}
	adamic_array_push(parts, (adamic_value){.reference = adamic_string_slice_bytes(string, start, string->length - start)});
	return parts;
}
