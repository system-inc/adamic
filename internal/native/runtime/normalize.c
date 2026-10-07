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
	if (point < 0x300) {
		return 0;
	}
	// Small direct-mapped caches avoid repeating binary searches in patterned text. Keys
	// are checked even when their slots collide; thread-local storage keeps callers independent.
	static _Thread_local uint32_t cached_points[64];
	static _Thread_local uint8_t cached_classes[64];
	size_t slot = (point ^ (point >> 8)) & 63;
	if (cached_points[slot] == point) {
		return cached_classes[slot];
	}
	size_t low = 0, high = COUNT(normalize_classes);
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (normalize_classes[middle].last < point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	uint8_t class = low < COUNT(normalize_classes) && normalize_classes[low].first <= point ? normalize_classes[low].class : 0;
	cached_points[slot] = point;
	cached_classes[slot] = class;
	return class;
}

static const normalize_mapping *mapping_of(uint32_t point) {
	if (point < 0xa0) {
		return NULL;
	}
	static _Thread_local uint32_t cached_points[64];
	static _Thread_local const normalize_mapping *cached_mappings[64];
	size_t slot = (point ^ (point >> 8)) & 63;
	if (cached_points[slot] == point) {
		return cached_mappings[slot];
	}
	size_t low = 0, high = COUNT(normalize_mappings);
	while (low < high) {
		size_t middle = low + (high - low) / 2;
		if (normalize_mappings[middle].point < point) {
			low = middle + 1;
		} else {
			high = middle;
		}
	}
	const normalize_mapping *mapping = low < COUNT(normalize_mappings) && normalize_mappings[low].point == point ? &normalize_mappings[low] : NULL;
	cached_points[slot] = point;
	cached_mappings[slot] = mapping;
	return mapping;
}

// composite is what first and second compose to, or 0 when they don't.
static uint32_t composite(uint32_t first, uint32_t second) {
	if (first >= hangul_l && first < hangul_l + hangul_l_count && second >= hangul_v && second < hangul_v + hangul_v_count) {
		return hangul_s + ((first - hangul_l) * hangul_v_count + (second - hangul_v)) * hangul_t_count;
	}
	if (first >= hangul_s && first < hangul_s + hangul_s_count && (first - hangul_s) % hangul_t_count == 0 && second > hangul_t && second < hangul_t + hangul_t_count) {
		return first + (second - hangul_t);
	}
	static _Thread_local normalize_pair cached_pairs[64];
	size_t slot = ((first * 33 + second) ^ (first >> 8) ^ (second >> 4)) & 63;
	if (cached_pairs[slot].first == first && cached_pairs[slot].second == second) {
		return cached_pairs[slot].composite;
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
	uint32_t joined = low < COUNT(normalize_pairs) && normalize_pairs[low].first == first && normalize_pairs[low].second == second ? normalize_pairs[low].composite : 0;
	cached_pairs[slot] = (normalize_pair){first, second, joined};
	return joined;
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

// Small runs use insertion sort. Long runs use a stable counting sort, so scrambled marks
// cost linear work rather than quadratic work. A starter always ends a run.
static void reorder(points *list) {
	for (size_t first = 0; first < list->count;) {
		if (combining_class(list->items[first]) == 0) {
			first++;
			continue;
		}
		size_t end = first + 1;
		bool ordered = true;
		uint8_t previous = combining_class(list->items[first]);
		while (end < list->count) {
			uint8_t class = combining_class(list->items[end]);
			if (class == 0) {
				break;
			}
			ordered = ordered && previous <= class;
			previous = class;
			end++;
		}
		if (!ordered && end - first <= 64) {
			for (size_t index = first + 1; index < end; index++) {
				uint32_t point = list->items[index];
				uint8_t class = combining_class(point);
				size_t at = index;
				while (at > first && combining_class(list->items[at - 1]) > class) {
					list->items[at] = list->items[at - 1];
					at--;
				}
				list->items[at] = point;
			}
		} else if (!ordered) {
			size_t offsets[256] = {0};
			for (size_t index = first; index < end; index++) {
				offsets[combining_class(list->items[index])]++;
			}
			size_t total = 0;
			for (size_t class = 0; class < COUNT(offsets); class++) {
				size_t count = offsets[class];
				offsets[class] = total;
				total += count;
			}
			uint32_t *sorted = malloc(total * sizeof *sorted);
			if (sorted == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			for (size_t index = first; index < end; index++) {
				uint32_t point = list->items[index];
				sorted[offsets[combining_class(point)]++] = point;
			}
			memcpy(list->items + first, sorted, total * sizeof *sorted);
			free(sorted);
		}
		first = end;
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

// Quick check without another Unicode table. Decomposed forms require no applicable mapping
// and canonical order. Composed forms accept only starters that normalize individually to themselves
// and cannot join their neighbour. Marks conservatively take the streaming path: a precomposed
// starter's decomposition can interact with them even when no direct composition pair exists.
static bool normalized(const adamic_string *string, bool compatibility, bool composed) {
	points expanded = {NULL, 0, 0};
	uint32_t cached = UINT32_MAX, starter = 0, self_join = 0, leading = 0;
	bool same = false, has_starter = false, result = true;
	uint8_t previous = 0, cached_class = 0;
	for (size_t at = 0; at < string->length;) {
		size_t size = size_at((unsigned char)string->bytes[at]);
		uint32_t point = decode_at((const unsigned char *)string->bytes + at, size);
		uint8_t class = point == cached ? cached_class : combining_class(point);
		if ((class != 0 && previous > class) || (composed && class != 0)) {
			result = false;
			break;
		}
		if (point != cached) {
			if (composed) {
				expanded.count = 0;
				decompose(&expanded, point, compatibility);
				leading = expanded.items[0];
				reorder(&expanded);
				compose(&expanded);
				same = expanded.count == 1 && expanded.items[0] == point;
			} else {
				const normalize_mapping *mapping = mapping_of(point);
				same = !(point >= hangul_s && point < hangul_s + hangul_s_count) &&
					(mapping == NULL || (mapping->compatibility && !compatibility));
			}
			cached = point;
			cached_class = class;
			self_join = composed ? composite(point, leading) : 0;
		}
		if (!same || (composed && has_starter && (starter == point ? self_join : composite(starter, leading)) != 0)) {
			result = false;
			break;
		}
		starter = point;
		has_starter = true;
		previous = class;
		at += size;
	}
	free(expanded.items);
	return result;
}

// A batch ends only at a starter. Its last starter stays pending when it could join the next
// one. Growing the buffer never ends a combining run.
// A first pass counts bytes and UTF-16 units; a second writes straight into one exact allocation.
// This also checks V8's length limit before allocating or encoding an oversized result.
typedef struct normalization {
	points segment;
	bool composed;
	char *out;
	size_t length;
	size_t units;
} normalization;

static void emit_point(normalization *state, uint32_t point) {
	char scratch[4];
	size_t size = encode(point, state->out == NULL ? scratch : state->out + state->length);
	state->length += size;
	state->units += point >= 0x10000 ? 2 : 1;
	if (state->units > ADAMIC_STRING_MAX_UNITS) {
		adamic_string_check_length((double)state->units);
	}
}

static void finish_segment(normalization *state, bool keep_starter) {
	reorder(&state->segment);
	if (state->composed) {
		compose(&state->segment);
	}
	size_t count = state->segment.count;
	bool kept = keep_starter && state->composed && count != 0 && combining_class(state->segment.items[count - 1]) == 0;
	for (size_t index = 0; index < count - (kept ? 1 : 0); index++) {
		emit_point(state, state->segment.items[index]);
	}
	if (kept) {
		state->segment.items[0] = state->segment.items[count - 1];
	}
	state->segment.count = kept ? 1 : 0;
}

static void accept_point(normalization *state, uint32_t point) {
	if (combining_class(point) == 0 && state->segment.count >= 64) {
		finish_segment(state, true);
	}
	points_add(&state->segment, point);
}

// Repeat an already normalized block of starters. Copy by doubling from bytes already written,
// rather than encode eighteen separate points for each occurrence of a compatibility ligature.
static void emit_blocks(normalization *state, const char *bytes, size_t length, size_t units, size_t count) {
	if (count == 0) {
		return;
	}
	if (count > (ADAMIC_STRING_MAX_UNITS - state->units) / units) {
		adamic_string_check_length((double)ADAMIC_STRING_MAX_UNITS + 1);
	}
	size_t size = length * count;
	if (state->out != NULL) {
		char *out = state->out + state->length;
		memcpy(out, bytes, length);
		for (size_t written = length; written < size;) {
			size_t next = size - written < written ? size - written : written;
			memcpy(out + written, out, next);
			written += next;
		}
	}
	state->length += size;
	state->units += units * count;
}

// A repeated prefix need only be normalized once when its decomposition starts with a
// starter and successive copies cannot compose across their boundary. Keep the final copy
// decomposed, not precomposed: later marks may have to reorder into its last canonical run.
static size_t repeated_prefix(normalization *state, const adamic_string *string, bool compatibility) {
	if (string->length < 64) {
		return 0;
	}
	size_t length = 0;
	for (size_t index = 0; index < 8 && length < string->length; index++) {
		length += size_at((unsigned char)string->bytes[length]);
		size_t count = 1, maximum = string->length / length;
		while (count < maximum &&
			memcmp(string->bytes, string->bytes + length * count, length) == 0) {
			count++;
		}
		if (count < 8) {
			continue;
		}
		points raw = {NULL, 0, 0}, ordered = {NULL, 0, 0};
		for (size_t at = 0; at < length;) {
			size_t size = size_at((unsigned char)string->bytes[at]);
			decompose(&raw, decode_at((const unsigned char *)string->bytes + at, size), compatibility);
			at += size;
		}
		for (size_t index = 0; index < raw.count; index++) {
			points_add(&ordered, raw.items[index]);
		}
		reorder(&ordered);
		if (state->composed) {
			compose(&ordered);
		}
		uint32_t first = raw.items[0], last = ordered.items[ordered.count - 1];
		bool separate = combining_class(first) == 0 &&
			(!state->composed || combining_class(last) != 0 || composite(last, first) == 0);
		if (separate) {
			char *bytes = malloc(ordered.count * 4);
			if (bytes == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			size_t written = 0, units = 0;
			for (size_t index = 0; index < ordered.count; index++) {
				written += encode(ordered.items[index], bytes + written);
				units += ordered.items[index] >= 0x10000 ? 2 : 1;
			}
			emit_blocks(state, bytes, written, units, count - 1);
			for (size_t index = 0; index < raw.count; index++) {
				accept_point(state, raw.items[index]);
			}
			free(bytes);
		}
		free(raw.items);
		free(ordered.items);
		if (separate) {
			return length * count;
		}
	}
	return 0;
}

static normalization stream(const adamic_string *string, bool compatibility, bool composed, char *out) {
	normalization state = {{NULL, 0, 0}, composed, out, 0, 0};
	points expanded = {NULL, 0, 0}, ordered = {NULL, 0, 0};
	uint32_t cached = UINT32_MAX;
	bool starters = false, separate = false;
	char *bytes = NULL;
	size_t length = 0, units = 0, capacity = 0;
	for (size_t at = repeated_prefix(&state, string, compatibility); at < string->length;) {
		size_t size = size_at((unsigned char)string->bytes[at]);
		uint32_t point = decode_at((const unsigned char *)string->bytes + at, size);
		if (point != cached) {
			expanded.count = 0;
			decompose(&expanded, point, compatibility);
			starters = true;
			for (size_t index = 0; index < expanded.count; index++) {
				starters = starters && combining_class(expanded.items[index]) == 0;
			}
			if (starters) {
				ordered.count = 0;
				for (size_t index = 0; index < expanded.count; index++) {
					points_add(&ordered, expanded.items[index]);
				}
				if (composed) {
					compose(&ordered);
				}
				separate = !composed || composite(ordered.items[ordered.count - 1], expanded.items[0]) == 0;
				if (expanded.count * 4 > capacity) {
					capacity = expanded.count * 4;
					char *grown = realloc(bytes, capacity);
					if (grown == NULL) {
						static const char message[] = "out of memory";
						adamic_panic(message, sizeof message - 1);
					}
					bytes = grown;
				}
				length = 0, units = 0;
				for (size_t index = 0; index < ordered.count; index++) {
					length += encode(ordered.items[index], bytes + length);
					units += ordered.items[index] >= 0x10000 ? 2 : 1;
				}
			}
			cached = point;
		}
		for (size_t index = 0; index < expanded.count; index++) {
			accept_point(&state, expanded.items[index]);
		}
		size_t next = at + size;
		// Keep the final starter pending for marks or a starter arriving in the next block.
		// The first block above settles any composition with what preceded this input point.
		if (starters && separate) {
			size_t count = 0;
			while (next + size <= string->length && memcmp(string->bytes + at, string->bytes + next, size) == 0) {
				count++;
				next += size;
			}
			if (count != 0) {
				// Settle composition with the preceding batch before taking the bulk path.
				if (composed) {
					reorder(&state.segment);
					compose(&state.segment);
				}
				if (state.segment.items[state.segment.count - 1] != ordered.items[ordered.count - 1]) {
					next = at + size;
				} else {
					finish_segment(&state, false);
					emit_blocks(&state, bytes, length, units, count - 1);
					for (size_t index = 0; index < expanded.count; index++) {
						accept_point(&state, expanded.items[index]);
					}
				}
			}
		}
		at = next;
	}
	finish_segment(&state, false);
	free(bytes);
	free(expanded.items);
	free(ordered.items);
	free(state.segment.items);
	state.segment = (points){NULL, 0, 0};
	return state;
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
	if (normalized(string, compatibility, composed)) {
		return adamic_retain((adamic_string *)string);
	}
	normalization sized = stream(string, compatibility, composed, NULL);
	adamic_string *result = adamic_string_allocate(sized.length);
	stream(string, compatibility, composed, (char *)result->bytes);
	result->units = sized.units + 1;
	return result;
}
