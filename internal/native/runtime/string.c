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
#include "library_errors.h"

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
