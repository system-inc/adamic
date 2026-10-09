// Private string split implementation, included only by string.c.

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
