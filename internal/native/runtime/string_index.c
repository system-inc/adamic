// string_index.c: where a UTF-16 unit is in a string's UTF-8 bytes, in constant time.
//
// JavaScript indexes strings by UTF-16 unit, and Adamic keeps them as UTF-8 (WTF-8), so text[index],
// charCodeAt and slice must find the byte a unit starts at. Walking from the start each time makes a
// tokenizer's loop over a non-ASCII file quadratic. So a string keeps its length in units once it's
// known (an ASCII string's units are its bytes, and need nothing more), and a long non-ASCII string
// built at runtime gets an index the first time it's asked: a checkpoint every STEP units, and a
// cursor at the last code point found. Nearby positions are walked to in either direction from
// that cursor, with backward reads choosing the closer of the cursor and checkpoint, so at most STEP
// code points are visited. The checkpoints are built once, in one pass, and freed with the string.
// Indexed strings also decode a compact UTF-16 view once for direct unit reads.
//
// A string is immutable, so nothing cached can go stale while it lives, and a string's memory comes
// back from allocate with nothing cached. The program's literals are immortal, and so is the index a
// long one gets; stack pieces cache their length but never get an index, which nothing would free.

#include "adamic.h"

#include <stdint.h>
#include <stdlib.h>

// STEP is the units between checkpoints, and MINIMUM the shortest string, in bytes, worth an index:
// below it a walk from the start is already a few steps. CURSOR says whether the index keeps a cursor.
// Measured on a tokenizer's loop and on scattered reads (2,000,000 units, -O2): STEP 16, 32, 64 and
// 128 were all linear, 32 was the quickest at scattered reads, and the cursor made the loop about a
// fifth quicker. The checkpoints cost 4 bytes per 32 units, an eighth of a byte per unit.
#ifndef STEP
#define STEP 32
#endif
#ifndef MINIMUM
#define MINIMUM 64
#endif
#ifndef CURSOR
#define CURSOR 1
#endif



static size_t width(unsigned char lead) {
	return lead < 0x80 ? 1 : lead < 0xe0 ? 2 : lead < 0xf0 ? 3 : 4;
}

static struct adamic_string_index *shared_index(const adamic_string *string, size_t units);

static struct adamic_string_index *cached(const adamic_string *string) {
	return adamic_is_shared(&string->heap) ?
		__atomic_load_n(&string->index, __ATOMIC_ACQUIRE) : string->index;
}

size_t adamic_string_units(const adamic_string *string) {
	if (string->units != 0) { return string->units - 1; }
	bool shared = adamic_is_shared(&string->heap);
	if (shared) {
		struct adamic_string_index *index = cached(string);
		if (index != NULL && index != ADAMIC_LITERAL_INDEX) { return index->units; }
	}
	size_t units = 0;
	for (size_t offset = 0; offset < string->length;) {
		size_t size = width((unsigned char)string->bytes[offset]);
		units += size == 4 ? 2 : 1;
		offset += size;
	}
	if (shared) { return shared_index(string, units)->units; }
	// Immortal literals may be read by any worker. Their fields remain immutable.
	if (adamic_reference_count(&string->heap) != 0) {
		((adamic_string *)string)->units = units + 1;
	}
	return units;
}

static struct adamic_string_index *build(adamic_string *string, size_t units) {
	bool indexed = units != string->length && string->length >= MINIMUM && string->length <= UINT32_MAX >> 1;
	size_t count = indexed ? units / STEP + 1 : 0;
	struct adamic_string_index *index = malloc(sizeof *index + count * sizeof index->checkpoints[0]);
	if (index == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	index->units = units;
	index->cursor_unit = 0;
	index->cursor_offset = 0;
	index->count = count;
	index->view = NULL;
	if (!indexed) { return index; }
	size_t checkpoint = 0, unit = 0;
	for (size_t offset = 0; offset < string->length;) {
		size_t size = width((unsigned char)string->bytes[offset]);
		size_t next = unit + (size == 4 ? 2 : 1);
		// Every checkpoint this code point holds: its first unit, or the low half of a pair.
		for (; checkpoint < count && checkpoint * STEP < next; checkpoint++) {
			index->checkpoints[checkpoint] = (uint32_t)(offset << 1 | (checkpoint * STEP != unit));
		}
		unit = next;
		offset += size;
	}
	// A checkpoint at the very end (units a multiple of STEP) is the end.
	for (; checkpoint < count; checkpoint++) {
		index->checkpoints[checkpoint] = (uint32_t)(string->length << 1);
	}
	index->view = malloc(units * sizeof *index->view);
	if (index->view == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	size_t at = 0;
	for (size_t offset = 0; offset < string->length;) {
		const unsigned char *bytes = (const unsigned char *)string->bytes + offset;
		size_t size = width(bytes[0]);
		unsigned point = size == 1 ? bytes[0] : size == 2 ?
			((unsigned)(bytes[0] & 0x1f) << 6) | (bytes[1] & 0x3f) : size == 3 ?
			((unsigned)(bytes[0] & 0x0f) << 12) | ((unsigned)(bytes[1] & 0x3f) << 6) | (bytes[2] & 0x3f) :
			((unsigned)(bytes[0] & 0x07) << 18) | ((unsigned)(bytes[1] & 0x3f) << 12) | ((unsigned)(bytes[2] & 0x3f) << 6) | (bytes[3] & 0x3f);
		if (size == 4) {
			index->view[at++] = (uint16_t)(0xd800 + ((point - 0x10000) >> 10));
			index->view[at++] = (uint16_t)(0xdc00 + ((point - 0x10000) & 0x3ff));
		} else {
			index->view[at++] = (uint16_t)point;
		}
		offset += size;
	}
	return index;
}

// Publish a complete immutable index with one CAS. A competing builder owns and frees
// its losing copy. The string's plain units field is never changed after publication.
static struct adamic_string_index *shared_index(const adamic_string *string, size_t units) {
	struct adamic_string_index *index = cached(string);
	if (index != NULL && index != ADAMIC_LITERAL_INDEX) { return index; }
	struct adamic_string_index *candidate = build((adamic_string *)string, units);
	struct adamic_string_index *expected = index;
	if (__atomic_compare_exchange_n(&((adamic_string *)string)->index, &expected, candidate,
		false, __ATOMIC_RELEASE, __ATOMIC_ACQUIRE)) { return candidate; }
	free(candidate->view);
	free(candidate);
	return expected;
}

char adamic_literal_mark;

// usable is a string's index, built the first time it's wanted, or NULL for a string that has none: a
// short one, an ASCII one, or one made on the stack. A literal (marked by ADAMIC_STRING) is immortal,
// so its index is built once and kept for as long as the program runs.
static struct adamic_string_index *usable(const adamic_string *string, size_t units) {
	struct adamic_string_index *index = cached(string);
	if (index != NULL && index != ADAMIC_LITERAL_INDEX) {
		return index->count != 0 ? index : NULL;
	}
	if (adamic_is_shared(&string->heap)) {
		index = shared_index(string, units);
		return index->count != 0 ? index : NULL;
	}

	// Offsets must fit a checkpoint, shifted: every string V8 allows does, by a factor of two.
	if ((adamic_reference_count(&string->heap) != 0 && !adamic_is_shared(&string->heap)) && units != string->length && string->length >= MINIMUM && string->length <= UINT32_MAX >> 1) {
		index = build((adamic_string *)string, units);
		((adamic_string *)string)->index = index;
		return index;
	}
	return NULL;
}

size_t adamic_string_units_before(const adamic_string *string, size_t offset) {
	size_t units = string->units != 0 ? string->units - 1 : adamic_string_units(string);
	if (units == string->length) {
		return offset;
	}
	// Keep the established-index path local; usable also handles building and unindexed strings.
	struct adamic_string_index *index = cached(string);
	if (index == NULL || index == ADAMIC_LITERAL_INDEX || index->count == 0) {
		index = usable(string, units);
	}
	size_t start = 0, at = 0;
	if (index != NULL) {
		// The last checkpoint at or before offset, by its offset (checkpoints never go backward), and
		// on from there.
		size_t low = 1, high = index->count;
		while (low < high) {
			size_t middle = low + (high - low) / 2;
			if ((index->checkpoints[middle] >> 1) <= offset) {
				low = middle + 1;
			} else {
				high = middle;
			}
		}
		uint32_t checkpoint = index->checkpoints[low - 1];
		start = (low - 1) * STEP - (checkpoint & 1);
		at = checkpoint >> 1;
	}
	while (at < offset) {
		size_t size = width((unsigned char)string->bytes[at]);
		start += size == 4 ? 2 : 1;
		at += size;
	}
	return start;
}

size_t adamic_string_locate(const adamic_string *string, size_t unit, bool *low) {
	size_t units = string->units != 0 ? string->units - 1 : adamic_string_units(string);
	if (units == string->length) {
		// Every code point is one byte.
		*low = false;
		return unit;
	}
	// Keep the established-index path local; usable also handles building and unindexed strings.
	struct adamic_string_index *index = cached(string);
	if (index == NULL || index == ADAMIC_LITERAL_INDEX || index->count == 0) {
		index = usable(string, units);
	}
	size_t start = 0, offset = 0;
	if (index != NULL) {
		// Sequential reads need no checkpoint load. A near forward read remains bounded by STEP.
		if (CURSOR && !adamic_is_shared(&string->heap) && unit >= index->cursor_unit && unit - index->cursor_unit < STEP) {
			start = index->cursor_unit;
			offset = index->cursor_offset;
		} else {
			uint32_t checkpoint = index->checkpoints[unit / STEP];
			start = unit / STEP * STEP - (checkpoint & 1);
			offset = checkpoint >> 1;
			if (CURSOR && !adamic_is_shared(&string->heap) && index->cursor_unit > unit && index->cursor_unit - unit <= unit - start) {
				start = index->cursor_unit;
				offset = index->cursor_offset;
				// The cursor names a code point's first unit. Step back over continuation bytes,
				// subtracting two units for a supplementary point and one for a lone surrogate.
				while (start > unit) {
					do {
						offset--;
					} while (((unsigned char)string->bytes[offset] & 0xc0) == 0x80);
					start -= width((unsigned char)string->bytes[offset]) == 4 ? 2 : 1;
				}
			}
		}
	}
	for (;;) {
		size_t size = width((unsigned char)string->bytes[offset]);
		size_t next = start + (size == 4 ? 2 : 1);
		if (unit < next) {
			break;
		}
		start = next;
		offset += size;
	}
	if (index != NULL && !adamic_is_shared(&string->heap)) {
		index->cursor_unit = start;
		index->cursor_offset = offset;
	}
	*low = unit != start;
	return offset;
}

// Private to the string runtime: the direct unit view, built on the first long-string read.
// Short strings and stack pieces keep their allocation-free walk.
const uint16_t *adamic_string_unit_view(const adamic_string *string) {
	struct adamic_string_index *index = cached(string);
	if (index == NULL || index == ADAMIC_LITERAL_INDEX) {
		index = usable(string, adamic_string_units(string));
	}
	return index != NULL ? index->view : NULL;
}

void adamic_string_free_index(adamic_string *string) {
	if (string->index == NULL || string->index == ADAMIC_LITERAL_INDEX) {
		return;
	}
	free(string->index->view);
	free(string->index);
}

void adamic_string_prepare_shared(adamic_string *string) {
	size_t units = adamic_string_units(string);
	(void)usable(string, units);
}

// RegExp borrows the same immutable view published by the shared string index.
// Short strings and ownerless stack pieces use the caller's temporary decoder.
const uint16_t *adamic_string_utf16_view(adamic_string *string) {
	return adamic_string_unit_view(string);
}
