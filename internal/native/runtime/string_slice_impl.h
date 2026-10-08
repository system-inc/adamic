// Private string slice implementation, included only by string.c.

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
	if (!(from < to)) {
		return adamic_retain(&adamic_string_empty);
	}
	// Whole code points are their bytes, in one piece, shared with the string where that's worth it
	// (string_share.c). A slice that starts on the low half of a pair begins with that half, and one
	// that ends between the halves ends with the high one, each a lone surrogate, which isn't in the
	// string's bytes, so that slice is built.
	size_t first = (size_t)from, last = (size_t)to;
	if (string->units == string->length + 1) {
		return adamic_string_share(string, first, last - first);
	}
	bool low;
	size_t offset = adamic_string_locate(string, first, &low);
	if (low) {
		offset += 4;
	}
	bool ends_low = false;
	size_t stop = last < (size_t)length ? adamic_string_locate(string, last, &ends_low) : string->length;
	size_t middle = stop > offset ? stop - offset : 0;
	if (!low && !ends_low) {
		adamic_string *slice = adamic_string_share(string, offset, middle);
		slice->units = last - first + 1;
		return slice;
	}
	builder build = {NULL, 0, 0};
	if (low) {
		builder_unit(&build, unit_at(string, offset - 4, true));
	}
	if (middle > 0) {
		builder_add(&build, string->bytes + offset, middle);
	}
	if (ends_low) {
		builder_unit(&build, unit_at(string, stop, false));
	}
	adamic_string *slice = builder_finish(&build);
	slice->units = last - first + 1;
	return slice;
}

// One ASCII unit has only 128 possible values. Its immutable bytes and header can live
// with the runtime, so indexing does not mint a header or pin a source for one byte.
// Like the other string caches, initialization follows the single-threaded counting model.
static adamic_string *ascii_character(unsigned char value) {
	static char bytes[128];
	static adamic_string characters[128];
	adamic_string *character = &characters[value];
	if (character->units == 0) {
		bytes[value] = (char)value;
		*character = (adamic_string){{0, adamic_kind_string, 0}, 1, &bytes[value], 2, ADAMIC_LITERAL_INDEX, NULL, 0};
	}
	return character;
}

adamic_string *adamic_string_at(const adamic_string *string, double index) {
	// As for an array: an index is an integer from 0 up to the length in UTF-16 units, and anything
	// else (negative, a fraction, NaN, past the end) is a property the string doesn't have. Half of a
	// supplementary character is a lone surrogate, as slice makes it.
	if (!(index >= 0) || index != trunc(index) || index >= adamic_string_length(string)) {
		return NULL;
	}
	if (string->units == string->length + 1) {
		return ascii_character((unsigned char)string->bytes[(size_t)index]);
	}
	// An ASCII unit of a non-ASCII source is the same immutable character. Long sources
	// already carry a direct UTF-16 view; short ones keep the existing bounded walk.
	double unit = adamic_string_char_code_at(string, index);
	if (unit < 128) {
		return ascii_character((unsigned char)unit);
	}
	return adamic_string_slice(string, index, index + 1, true);
}

adamic_maybe_number adamic_string_code_point_at(const adamic_string *string, double position) {
	position = isnan(position) ? 0 : trunc(position);
	adamic_maybe_number missing = {false, 0};
	size_t known = adamic_string_known_units(string);
	size_t length = known != 0 ? known - 1 : adamic_string_units(string);
	if (position < 0 || position >= (double)length) {
		return missing;
	}
	if (length == string->length) {
		adamic_maybe_number found = {true, (double)(unsigned char)string->bytes[(size_t)position]};
		return found;
	}
	const uint16_t *view = adamic_string_unit_view(string);
	if (view != NULL) {
		size_t at = (size_t)position;
		unsigned point = view[at];
		if (point >= 0xd800 && point <= 0xdbff && at + 1 < length && view[at + 1] >= 0xdc00 && view[at + 1] <= 0xdfff) {
			point = 0x10000 + ((point - 0xd800) << 10) + (view[at + 1] - 0xdc00);
		}
		adamic_maybe_number found = {true, (double)point};
		return found;
	}
	// At the high half of a pair, the whole code point; anywhere else, the unit.
	bool low;
	size_t offset = adamic_string_locate(string, (size_t)position, &low);
	size_t size = sequence((unsigned char)string->bytes[offset]);
	unsigned point = decode((const unsigned char *)string->bytes + offset, size);
	if (low) {
		point = 0xdc00 + ((point - 0x10000) & 0x3ff);
	}
	adamic_maybe_number found = {true, (double)point};
	return found;
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

