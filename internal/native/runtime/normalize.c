// normalize.c: String.prototype.normalize, NFC, NFD, NFKC and NFKD, as Node does them (UAX #15).
//
// A string is decomposed (canonically, or with compatibility mappings too for the K forms), put in
// canonical order (each run of combining marks stably sorted by combining class), and for NFC and
// NFKC composed again, each mark joining the last starter unless something between blocks it. The
// tables are Unicode's, for the version Node carries, written by normalize_generate.go. Hangul
// syllables decompose and compose by arithmetic. A lone surrogate is a starter with nothing to
// decompose or compose, and passes through as its WTF-8 bytes, as in Node.

#include "adamic.h"
#include "library_errors.h"

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef struct normalize_class {
	uint32_t first;
	uint32_t last;
	uint8_t class;
} normalize_class;

typedef struct normalize_mapping {
	uint32_t point;
	bool compatibility;
	uint16_t start;
	uint8_t count;
} normalize_mapping;

typedef struct normalize_pair {
	uint32_t first;
	uint32_t second;
	uint32_t composite;
} normalize_pair;

#include "normalize_tables.h"

#define COUNT(table) (sizeof table / sizeof table[0])

// Hangul syllables, by arithmetic (The Unicode Standard, 3.12).
enum {
	hangul_s = 0xac00,
	hangul_l = 0x1100,
	hangul_v = 0x1161,
	hangul_t = 0x11a7,
	hangul_l_count = 19,
	hangul_v_count = 21,
	hangul_t_count = 28,
	hangul_n_count = hangul_v_count * hangul_t_count,
	hangul_s_count = hangul_l_count * hangul_n_count,
};

static uint8_t combining_class(uint32_t point) {
	size_t low = 0, high = COUNT(normalize_classes);
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (normalize_classes[middle].last < point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	return low < COUNT(normalize_classes) && normalize_classes[low].first <= point ? normalize_classes[low].class : 0;
}

static const normalize_mapping *mapping_of(uint32_t point) {
	size_t low = 0, high = COUNT(normalize_mappings);
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (normalize_mappings[middle].point < point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	return low < COUNT(normalize_mappings) && normalize_mappings[low].point == point ? &normalize_mappings[low] : NULL;
}

// composite is what first and second compose to, or 0 when they don't.
static uint32_t composite(uint32_t first, uint32_t second) {
	if (first >= hangul_l && first < hangul_l + hangul_l_count && second >= hangul_v && second < hangul_v + hangul_v_count) {
		return hangul_s + ((first - hangul_l) * hangul_v_count + (second - hangul_v)) * hangul_t_count;
	}
	if (first >= hangul_s && first < hangul_s + hangul_s_count && (first - hangul_s) % hangul_t_count == 0 && second > hangul_t && second < hangul_t + hangul_t_count) {
		return first + (second - hangul_t);
	}
	size_t low = 0, high = COUNT(normalize_pairs);
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		const normalize_pair *pair = &normalize_pairs[middle];
		if (pair->first < first || (pair->first == first && pair->second < second)) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	if (low < COUNT(normalize_pairs) && normalize_pairs[low].first == first && normalize_pairs[low].second == second) {
		return normalize_pairs[low].composite;
	}
	return 0;
}

// A growing list of code points.
typedef struct points {
	uint32_t *items;
	size_t count;
	size_t capacity;
} points;

static void points_add(points *list, uint32_t point) {
	if (list->count == list->capacity) {
		size_t capacity = list->capacity == 0 ? 64 : list->capacity * 2;
		uint32_t *grown = realloc(list->items, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		list->items = grown;
		list->capacity = capacity;
	}
	list->items[list->count++] = point;
}

// decompose appends a code point's full decomposition: its mapping's, each decomposed in turn.
static void decompose(points *list, uint32_t point, bool compatibility) {
	if (point >= hangul_s && point < hangul_s + hangul_s_count) {
		uint32_t index = point - hangul_s;
		points_add(list, hangul_l + index / hangul_n_count);
		points_add(list, hangul_v + (index % hangul_n_count) / hangul_t_count);
		if (index % hangul_t_count != 0) {
			points_add(list, hangul_t + index % hangul_t_count);
		}
		return;
	}
	const normalize_mapping *mapping = mapping_of(point);
	if (mapping == NULL || (mapping->compatibility && !compatibility)) {
		points_add(list, point);
		return;
	}
	for (size_t index = 0; index < mapping->count; index++) {
		decompose(list, normalize_mapped[mapping->start + index], compatibility);
	}
}

// reorder puts every run of combining marks in canonical order: stably, by combining class.
static void reorder(points *list) {
	for (size_t index = 1; index < list->count; index++) {
		uint32_t point = list->items[index];
		uint8_t class = combining_class(point);
		if (class == 0) {
			continue;
		}
		size_t at = index;
		while (at > 0) {
			uint8_t before = combining_class(list->items[at - 1]);
			if (before == 0 || before <= class) {
				break;
			}
			list->items[at] = list->items[at - 1];
			at--;
		}
		list->items[at] = point;
	}
}

// compose joins each code point to the last starter when nothing between blocks it: a mark is blocked
// by anything between of class 0 or of a class at least its own, so a starter joins only the starter
// right before it.
static void compose(points *list) {
	size_t written = 0;
	bool has_starter = false;
	size_t starter = 0;
	uint8_t last_class = 0;
	for (size_t index = 0; index < list->count; index++) {
		uint32_t point = list->items[index];
		uint8_t class = combining_class(point);
		if (has_starter) {
			bool adjacent = written - 1 == starter;
			if (adjacent || (last_class != 0 && last_class < class)) {
				uint32_t joined = composite(list->items[starter], point);
				if (joined != 0) {
					list->items[starter] = joined;
					continue;
				}
			}
		}
		if (class == 0) {
			has_starter = true;
			starter = written;
		}
		last_class = class;
		list->items[written++] = point;
	}
	list->count = written;
}

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

// encode writes a code point as UTF-8, or a lone surrogate as WTF-8's three bytes, which the same
// arithmetic gives.
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

static bool form_is(const adamic_string *form, const char *name) {
	size_t length = strlen(name);
	return form->length == length && memcmp(form->bytes, name, length) == 0;
}

adamic_string *adamic_string_normalize(const adamic_string *string, const adamic_string *form) {
	bool compatibility, composed;
	if (form_is(form, "NFC")) {
		compatibility = false, composed = true;
	} else if (form_is(form, "NFD")) {
		compatibility = false, composed = false;
	} else if (form_is(form, "NFKC")) {
		compatibility = true, composed = true;
	} else if (form_is(form, "NFKD")) {
		compatibility = true, composed = false;
	} else {
		static const char message[] = "RangeError: The normalization form should be one of NFC, NFD, NFKC, NFKD.";
		adamic_uncaught_library_error(message, sizeof message - 1);
	}
	// ASCII is the same in every form.
	bool ascii = true;
	for (size_t at = 0; at < string->length && ascii; at++) {
		ascii = (unsigned char)string->bytes[at] < 0x80;
	}
	if (ascii) {
		return adamic_retain((adamic_string *)string);
	}
	points list = {NULL, 0, 0};
	for (size_t at = 0; at < string->length;) {
		size_t size = size_at((unsigned char)string->bytes[at]);
		decompose(&list, decode_at((const unsigned char *)string->bytes + at, size), compatibility);
		at += size;
	}
	reorder(&list);
	if (composed) {
		compose(&list);
	}
	size_t length = 0;
	char scratch[4];
	for (size_t index = 0; index < list.count; index++) {
		length += encode(list.items[index], scratch);
	}
	adamic_string *result = adamic_string_allocate(length);
	char *out = (char *)result->bytes;
	for (size_t index = 0; index < list.count; index++) {
		out += encode(list.items[index], out);
	}
	free(list.items);
	// Decomposing can lengthen a string (U+FDFA is eighteen code points) past V8's longest.
	if (length > ADAMIC_STRING_MAX_UNITS) {
		adamic_string_check_length(adamic_string_length(result));
	}
	return result;
}
