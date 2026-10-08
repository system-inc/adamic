// Private string search implementation, included only by string.c.

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
	return adamic_string_last_index_of_from(string, search, INFINITY);
}

// Port of V8 String::LastIndexOf (Node 24.19.0, src/objects/string.cc), with the
// existing WTF-8 byte search and surrogate-half fallback in place of flat UTF-16 storage.
double adamic_string_last_index_of_from(const adamic_string *string, const adamic_string *search, double position) {
	size_t length = adamic_string_units(string);
	size_t from = length;
	if (!isnan(position) && position < (double)length) {
		from = position <= 0 ? 0 : (size_t)trunc(position);
	}
	if (search->length == 0) {
		return (double)from;
	}
	if (!halves_pairs(search)) {
		// A whole-character needle matches by bytes, as indexOf does. Look backward without
		// decoding the whole haystack into a temporary UTF-16 buffer for each search.
		if (search->length > string->length) {
			return -1;
		}
		size_t start = string->length - search->length;
		if (from < length) {
			bool low;
			size_t bound = adamic_string_locate(string, from, &low);
			if (start > bound) {
				start = bound;
			}
		}
		for (size_t offset = start + 1; offset-- > 0;) {
			if (string->bytes[offset] == search->bytes[0] &&
				memcmp(string->bytes + offset, search->bytes, search->length) == 0) {
				return (double)adamic_string_units_before(string, offset);
			}
		}
		return -1;
	}
	size_t haystack_count, needle_count;
	unsigned *haystack = to_units(string, &haystack_count), *needle = to_units(search, &needle_count);
	double found = -1;
	if (needle_count <= haystack_count) {
		size_t start = haystack_count - needle_count;
		if (start > from) {
			start = from;
		}
		for (size_t at = start + 1; at-- > 0;) {
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

