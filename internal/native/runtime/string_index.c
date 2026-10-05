// string_index.c: where a UTF-16 unit is in a string's UTF-8 bytes, in constant time.
//
// JavaScript indexes strings by UTF-16 unit, and Adamic keeps them as UTF-8 (WTF-8), so text[index],
// charCodeAt and slice must find the byte a unit starts at. Walking from the start each time makes a
// tokenizer's loop over a non-ASCII file quadratic. So a string keeps its length in units once it's
// known (an ASCII string's units are its bytes, and need nothing more), and a long non-ASCII string
// built at runtime gets an index the first time it's asked: a checkpoint every STEP units, and a
// cursor at the last code point found. A position just ahead of the cursor (the next index in a loop)
// is walked to from the cursor, and any other from the checkpoint below it, so either is at most STEP
// code points away. The index is built once, in one pass, and freed with the string.
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

struct adamic_string_index {
	// The last code point found: its first unit, and its byte offset.
	size_t cursor_unit;
	size_t cursor_offset;
	// checkpoints[k] is where unit k * STEP is: the byte offset of the code point holding it, shifted
	// left once, and 1 when that unit is the low half of a surrogate pair, whose code point starts a
	// unit earlier.
	size_t count;
	uint32_t checkpoints[];
};

static size_t width(unsigned char lead) {
	return lead < 0x80 ? 1 : lead < 0xe0 ? 2 : lead < 0xf0 ? 3 : 4;
}

size_t adamic_string_units(const adamic_string *string) {
	if (string->units == 0) {
		size_t units = 0;
		for (size_t offset = 0; offset < string->length;) {
			size_t size = width((unsigned char)string->bytes[offset]);
			units += size == 4 ? 2 : 1;
			offset += size;
		}
		// A cache, not a change: the string means the same with it or without.
		((adamic_string *)string)->units = units + 1;
	}
	return string->units - 1;
}

static struct adamic_string_index *build(adamic_string *string, size_t units) {
	size_t count = units / STEP + 1;
	struct adamic_string_index *index = malloc(sizeof *index + count * sizeof index->checkpoints[0]);
	if (index == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	index->cursor_unit = 0;
	index->cursor_offset = 0;
	index->count = count;
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
	string->index = index;
	return index;
}

char adamic_literal_mark;

// usable is a string's index, built the first time it's wanted, or NULL for a string that has none: a
// short one, an ASCII one, or one made on the stack. A literal (marked by ADAMIC_STRING) is immortal,
// so its index is built once and kept for as long as the program runs.
static struct adamic_string_index *usable(const adamic_string *string, size_t units) {
	struct adamic_string_index *index = string->index;
	if (index != NULL && index != ADAMIC_LITERAL_INDEX) {
		return index;
	}
	bool literal = index == ADAMIC_LITERAL_INDEX;
	// Offsets must fit a checkpoint, shifted: every string V8 allows does, by a factor of two.
	if ((literal || string->heap.references != 0) && units != string->length && string->length >= MINIMUM && string->length <= UINT32_MAX >> 1) {
		return build((adamic_string *)string, units);
	}
	return NULL;
}

size_t adamic_string_units_before(const adamic_string *string, size_t offset) {
	size_t units = adamic_string_units(string);
	if (units == string->length) {
		return offset;
	}
	struct adamic_string_index *index = usable(string, units);
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
	size_t units = adamic_string_units(string);
	if (units == string->length) {
		// Every code point is one byte.
		*low = false;
		return unit;
	}
	struct adamic_string_index *index = usable(string, units);
	size_t start = 0, offset = 0;
	if (index != NULL) {
		if (CURSOR && unit >= index->cursor_unit && unit - index->cursor_unit < STEP) {
			start = index->cursor_unit;
			offset = index->cursor_offset;
		} else {
			uint32_t checkpoint = index->checkpoints[unit / STEP];
			start = unit / STEP * STEP - (checkpoint & 1);
			offset = checkpoint >> 1;
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
	if (index != NULL) {
		index->cursor_unit = start;
		index->cursor_offset = offset;
	}
	*low = unit != start;
	return offset;
}

void adamic_string_free_index(adamic_string *string) {
	free(string->index);
}
