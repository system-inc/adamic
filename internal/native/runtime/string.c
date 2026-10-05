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
	// A lone high surrogate followed by a lone low one is a character again, as in JavaScript, and
	// WTF-8 writes it as UTF-8's four bytes, so equal strings stay equal byte for byte.
	char *bytes = (char *)string->bytes;
	size_t kept = 0;
	for (size_t at = 0; at < length;) {
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
	string->length = kept;
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
	adamic_string piece = {{0, adamic_kind_string}, build->length, build->bytes};
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
	builder build = {NULL, 0, 0};
	units walk = units_start(string);
	unsigned unit;
	double position = 0;
	while (position < to && units_next(&walk, &unit)) {
		if (position >= from) {
			// A whole supplementary character is copied as its four bytes; half of one becomes a
			// lone surrogate.
			if (walk.second && position + 1 < to && position >= from) {
				builder_add(&build, string->bytes + walk.offset, walk.size);
				units_next(&walk, &unit);
				position += 2;
				continue;
			}
			builder_unit(&build, unit);
		}
		position++;
	}
	return builder_finish(&build);
}

adamic_maybe_number adamic_string_code_point_at(const adamic_string *string, double position) {
	position = isnan(position) ? 0 : trunc(position);
	adamic_maybe_number missing = {false, 0};
	if (position < 0) {
		return missing;
	}
	units walk = units_start(string);
	unsigned unit;
	double index = 0;
	while (units_next(&walk, &unit)) {
		if (index == position) {
			if (walk.second) {
				// The high half of a pair: the whole code point.
				adamic_maybe_number found = {true, (double)decode((const unsigned char *)string->bytes + walk.offset, walk.size)};
				return found;
			}
			adamic_maybe_number found = {true, (double)unit};
			return found;
		}
		index++;
	}
	return missing;
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

adamic_string *adamic_string_repeat(const adamic_string *string, double count) {
	count = isnan(count) ? 0 : trunc(count);
	if (count < 0 || isinf(count)) {
		char message[96];
		char number[ADAMIC_NUMBER_FORMAT_MAX];
		size_t length = adamic_number_format(count, number);
		int written = snprintf(message, sizeof message, "RangeError: Invalid count value: %.*s", (int)length, number);
		adamic_panic(message, (size_t)written);
	}
	builder build = {NULL, 0, 0};
	for (double index = 0; index < count; index++) {
		builder_add(&build, string->bytes, string->length);
	}
	return builder_finish(&build);
}

adamic_string *adamic_string_pad(const adamic_string *string, double target, const adamic_string *fill, bool at_start) {
	double length = adamic_string_length(string);
	target = isnan(target) ? 0 : trunc(target);
	if (target <= length || fill->length == 0) {
		return adamic_retain((adamic_string *)string);
	}
	// The fill, repeated and cut to exactly the missing number of UTF-16 units.
	double missing = target - length;
	double fill_length = adamic_string_length(fill);
	adamic_string *repeated = adamic_string_repeat(fill, ceil(missing / fill_length));
	adamic_string *padding = adamic_string_slice(repeated, 0, missing, true);
	adamic_release(repeated);
	adamic_string *parts[2] = {padding, (adamic_string *)string};
	if (!at_start) {
		parts[0] = (adamic_string *)string;
		parts[1] = padding;
	}
	adamic_string *padded = adamic_string_concat(2, parts);
	adamic_release(padding);
	return padded;
}

double adamic_string_index_of(const adamic_string *string, const adamic_string *search) {
	if (search->length == 0) {
		return 0;
	}
	for (size_t offset = 0; offset + search->length <= string->length;) {
		if (memcmp(string->bytes + offset, search->bytes, search->length) == 0) {
			adamic_string prefix = {{0, adamic_kind_string}, offset, string->bytes};
			return adamic_string_length(&prefix);
		}
		offset += sequence((unsigned char)string->bytes[offset]);
	}
	return -1;
}

bool adamic_string_starts_with(const adamic_string *string, const adamic_string *search) {
	return search->length <= string->length && memcmp(string->bytes, search->bytes, search->length) == 0;
}

bool adamic_string_ends_with(const adamic_string *string, const adamic_string *search) {
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
