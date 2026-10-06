// Private string decode implementation, included only by string.c.

// UTF-16 to the program, UTF-8 underneath. A code point of 1 to 3 UTF-8 bytes is one UTF-16 code
// unit, and one of 4 bytes is two, a surrogate pair. A lone surrogate is stored as the 3 bytes
// WTF-8 gives it, and is one unit too, so the same rule counts it.

// sequence is how many bytes the code point starting with this byte takes.
static size_t sequence(unsigned char lead) {
	if (lead < 0x80) {
		return 1;
	}
	if (lead < 0xe0) {
		return 2;
	}
	if (lead < 0xf0) {
		return 3;
	}
	return 4;
}

// decode reads the code point at bytes, given its sequence length.
static unsigned decode(const unsigned char *bytes, size_t size) {
	switch (size) {
	case 1:
		return bytes[0];
	case 2:
		return ((unsigned)(bytes[0] & 0x1f) << 6) | (bytes[1] & 0x3f);
	case 3:
		return ((unsigned)(bytes[0] & 0x0f) << 12) | ((unsigned)(bytes[1] & 0x3f) << 6) | (bytes[2] & 0x3f);
	}
	return ((unsigned)(bytes[0] & 0x07) << 18) | ((unsigned)(bytes[1] & 0x3f) << 12) | ((unsigned)(bytes[2] & 0x3f) << 6) | (bytes[3] & 0x3f);
}

// unit_at is the unit at an already located code point, or one half of a pair.
static unsigned unit_at(const adamic_string *string, size_t offset, bool low) {
	size_t size = sequence((unsigned char)string->bytes[offset]);
	unsigned point = decode((const unsigned char *)string->bytes + offset, size);
	if (size == 4) {
		return low ? 0xdc00 + ((point - 0x10000) & 0x3ff) : 0xd800 + ((point - 0x10000) >> 10);
	}
	return point;
}

double adamic_string_char_code(const adamic_string *string, double position) {
	// ToIntegerOrInfinity: NaN is 0, and a fraction truncates.
	position = isnan(position) ? 0 : trunc(position);
	size_t length = string->units != 0 ? string->units - 1 : adamic_string_units(string);
	if (position < 0 || position >= (double)length) {
		return NAN;
	}
	if (length == string->length) {
		return (double)(unsigned char)string->bytes[(size_t)position];
	}
	const uint16_t *bmp = adamic_string_bmp_view(string);
	if (bmp != NULL) {
		return (double)bmp[(size_t)position];
	}
	bool low;
	size_t offset = adamic_string_locate(string, (size_t)position, &low);
	return (double)unit_at(string, offset, low);
}

