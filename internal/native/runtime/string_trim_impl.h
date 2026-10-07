// Private string trim implementation, included only by string.c.

// is_space is JavaScript's WhiteSpace and LineTerminator, which trim removes (ECMA-262).
static bool is_space(unsigned point) {
	switch (point) {
	case 0x09: case 0x0a: case 0x0b: case 0x0c: case 0x0d: case 0x20: case 0xa0: case 0x1680:
	case 0x2028: case 0x2029: case 0x202f: case 0x205f: case 0x3000: case 0xfeff:
		return true;
	}
	return point >= 0x2000 && point <= 0x200a;
}

adamic_string *adamic_string_trim(adamic_string *string) {
	return adamic_string_trim_sides(string, true, true);
}

adamic_string *adamic_string_trim_sides(adamic_string *string, bool at_start, bool at_end) {
	size_t start = 0, end = string->length;
	while (at_start && start < end) {
		size_t size = sequence((unsigned char)string->bytes[start]);
		if (!is_space(decode((const unsigned char *)string->bytes + start, size))) {
			break;
		}
		start += size;
	}
	while (at_end && end > start) {
		// Back up to the start of the last code point: continuation bytes are 10xxxxxx.
		size_t lead = end - 1;
		while (lead > start && ((unsigned char)string->bytes[lead] & 0xc0) == 0x80) {
			lead--;
		}
		if (!is_space(decode((const unsigned char *)string->bytes + lead, end - lead))) {
			break;
		}
		end = lead;
	}
	return adamic_string_share(string, start, end - start);
}

