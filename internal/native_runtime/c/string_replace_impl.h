// Private string replace implementation, included only by string.c.

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

