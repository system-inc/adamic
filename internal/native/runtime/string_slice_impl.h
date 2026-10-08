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

// One ASCII unit has only 128 possible values. Static initialization publishes all
// immutable bytes and immortal headers before any worker can index a string.
static adamic_string *ascii_character(unsigned char value) {
	static const char bytes[128] = {
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
		32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
		64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
		80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95,
		96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111,
		112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127,
	};
#define ASCII_CHARACTER(index) {{0, adamic_kind_string, 0}, 1, &bytes[index], 2, ADAMIC_LITERAL_INDEX, NULL, 0}
	static adamic_string characters[128] = {
		ASCII_CHARACTER(0), ASCII_CHARACTER(1), ASCII_CHARACTER(2), ASCII_CHARACTER(3), ASCII_CHARACTER(4), ASCII_CHARACTER(5), ASCII_CHARACTER(6), ASCII_CHARACTER(7),
		ASCII_CHARACTER(8), ASCII_CHARACTER(9), ASCII_CHARACTER(10), ASCII_CHARACTER(11), ASCII_CHARACTER(12), ASCII_CHARACTER(13), ASCII_CHARACTER(14), ASCII_CHARACTER(15),
		ASCII_CHARACTER(16), ASCII_CHARACTER(17), ASCII_CHARACTER(18), ASCII_CHARACTER(19), ASCII_CHARACTER(20), ASCII_CHARACTER(21), ASCII_CHARACTER(22), ASCII_CHARACTER(23),
		ASCII_CHARACTER(24), ASCII_CHARACTER(25), ASCII_CHARACTER(26), ASCII_CHARACTER(27), ASCII_CHARACTER(28), ASCII_CHARACTER(29), ASCII_CHARACTER(30), ASCII_CHARACTER(31),
		ASCII_CHARACTER(32), ASCII_CHARACTER(33), ASCII_CHARACTER(34), ASCII_CHARACTER(35), ASCII_CHARACTER(36), ASCII_CHARACTER(37), ASCII_CHARACTER(38), ASCII_CHARACTER(39),
		ASCII_CHARACTER(40), ASCII_CHARACTER(41), ASCII_CHARACTER(42), ASCII_CHARACTER(43), ASCII_CHARACTER(44), ASCII_CHARACTER(45), ASCII_CHARACTER(46), ASCII_CHARACTER(47),
		ASCII_CHARACTER(48), ASCII_CHARACTER(49), ASCII_CHARACTER(50), ASCII_CHARACTER(51), ASCII_CHARACTER(52), ASCII_CHARACTER(53), ASCII_CHARACTER(54), ASCII_CHARACTER(55),
		ASCII_CHARACTER(56), ASCII_CHARACTER(57), ASCII_CHARACTER(58), ASCII_CHARACTER(59), ASCII_CHARACTER(60), ASCII_CHARACTER(61), ASCII_CHARACTER(62), ASCII_CHARACTER(63),
		ASCII_CHARACTER(64), ASCII_CHARACTER(65), ASCII_CHARACTER(66), ASCII_CHARACTER(67), ASCII_CHARACTER(68), ASCII_CHARACTER(69), ASCII_CHARACTER(70), ASCII_CHARACTER(71),
		ASCII_CHARACTER(72), ASCII_CHARACTER(73), ASCII_CHARACTER(74), ASCII_CHARACTER(75), ASCII_CHARACTER(76), ASCII_CHARACTER(77), ASCII_CHARACTER(78), ASCII_CHARACTER(79),
		ASCII_CHARACTER(80), ASCII_CHARACTER(81), ASCII_CHARACTER(82), ASCII_CHARACTER(83), ASCII_CHARACTER(84), ASCII_CHARACTER(85), ASCII_CHARACTER(86), ASCII_CHARACTER(87),
		ASCII_CHARACTER(88), ASCII_CHARACTER(89), ASCII_CHARACTER(90), ASCII_CHARACTER(91), ASCII_CHARACTER(92), ASCII_CHARACTER(93), ASCII_CHARACTER(94), ASCII_CHARACTER(95),
		ASCII_CHARACTER(96), ASCII_CHARACTER(97), ASCII_CHARACTER(98), ASCII_CHARACTER(99), ASCII_CHARACTER(100), ASCII_CHARACTER(101), ASCII_CHARACTER(102), ASCII_CHARACTER(103),
		ASCII_CHARACTER(104), ASCII_CHARACTER(105), ASCII_CHARACTER(106), ASCII_CHARACTER(107), ASCII_CHARACTER(108), ASCII_CHARACTER(109), ASCII_CHARACTER(110), ASCII_CHARACTER(111),
		ASCII_CHARACTER(112), ASCII_CHARACTER(113), ASCII_CHARACTER(114), ASCII_CHARACTER(115), ASCII_CHARACTER(116), ASCII_CHARACTER(117), ASCII_CHARACTER(118), ASCII_CHARACTER(119),
		ASCII_CHARACTER(120), ASCII_CHARACTER(121), ASCII_CHARACTER(122), ASCII_CHARACTER(123), ASCII_CHARACTER(124), ASCII_CHARACTER(125), ASCII_CHARACTER(126), ASCII_CHARACTER(127),
	};
#undef ASCII_CHARACTER
	return &characters[value];
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

