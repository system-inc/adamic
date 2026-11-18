#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// A shared string's readers each keep their own cursor (string_index.c). Four workers read one shared
// non-ASCII string's units forward and back from different places, each offset checked against a walk of
// the bytes made before sharing; then one thread's cursor for a freed index must not be taken for a new
// index at the same address. The text mixes one-, two-, three- and four-byte code points, so an offset
// read from the wrong place lands mid-character.

static adamic_string *text;
static size_t units;
static size_t *offsets; // offsets[unit]: the byte its code point starts at
static bool *lows;      // lows[unit]: the low half of a surrogate pair

static void fill(adamic_string *string, const char *pattern, size_t pattern_length) {
	for (size_t i = 0; i < string->length; i += pattern_length) { memcpy((char *)string->bytes + i, pattern, pattern_length); }
}

static void walk(const adamic_string *string) {
	units = 0;
	for (size_t offset = 0; offset < string->length;) {
		unsigned char lead = (unsigned char)string->bytes[offset];
		size_t size = lead < 0x80 ? 1 : lead < 0xe0 ? 2 : lead < 0xf0 ? 3 : 4;
		offsets[units] = offset;
		lows[units++] = false;
		if (size == 4) {
			offsets[units] = offset;
			lows[units++] = true;
		}
		offset += size;
	}
}

static void check(const adamic_string *string, size_t unit) {
	bool low = false;
	size_t offset = adamic_string_locate(string, unit, &low);
	if (offset != offsets[unit] || low != lows[unit]) {
		fprintf(stderr, "unit %zu: offset %zu low %d, want %zu low %d\n", unit, offset, low, offsets[unit], lows[unit]);
		abort();
	}
}

static void *read_text(void *argument) {
	size_t first = (size_t)argument * units / 4;
	for (size_t step = 0; step < units; step++) { check(text, (first + step) % units); }
	for (size_t step = units; step > 0; step--) { check(text, (first + step - 1) % units); }
	return NULL;
}

int main(void) {
	// "aé中😀": 1 + 2 + 3 + 4 bytes, 5 units, repeated.
	static const char pattern[] = "a\xc3\xa9\xe4\xb8\xad\xf0\x9f\x98\x80";
	text = adamic_string_allocate((sizeof pattern - 1) * 4096);
	fill(text, pattern, sizeof pattern - 1);
	offsets = malloc(text->length * sizeof *offsets);
	lows = malloc(text->length * sizeof *lows);
	if (offsets == NULL || lows == NULL) { abort(); }
	walk(text);
	adamic_share(text);
	pthread_t readers[4];
	for (size_t i = 0; i < 4; i++) { if (pthread_create(&readers[i], NULL, read_text, (void *)i) != 0) { abort(); } }
	for (size_t i = 0; i < 4; i++) { pthread_join(readers[i], NULL); }
	adamic_release(text);

	// Reuse: this thread's cursor sits at unit 1000 of a shared all-"é" string's index. The index is freed
	// with the string, and a string with the same units and bytes but every offset one byte later ("中",
	// then "é"s, then "a") gets an index at the same address. Its unit 1001 is at byte 2003, where a cursor
	// taken from the old index would start at byte 2000, inside a character.
	adamic_string *old = adamic_string_allocate(2 * 4096);
	fill(old, "\xc3\xa9", 2);
	adamic_share(old);
	adamic_string_prepare_shared(old);
	walk(old);
	check(old, 1000);
	const struct adamic_string_index *freed = old->index;
	adamic_release(old);
	adamic_string *next = adamic_string_allocate(2 * 4096);
	fill(next, "\xc3\xa9", 2);
	memcpy((char *)next->bytes, "\xe4\xb8\xad", 3);
	for (size_t i = 3; i + 1 < next->length; i += 2) { memcpy((char *)next->bytes + i, "\xc3\xa9", 2); }
	((char *)next->bytes)[next->length - 1] = 'a';
	adamic_share(next);
	adamic_string_prepare_shared(next);
	if (next->index != freed) {
		// A sanitizer's quarantine keeps the address from coming back; the release build is the witness.
		adamic_release(next);
		free(offsets);
		free(lows);
		puts("shared cursors clean, index address not reused");
		return 0;
	}
	walk(next);
	check(next, 1001);
	adamic_release(next);
	free(offsets);
	free(lows);
	puts("shared cursors clean");
	return 0;
}
