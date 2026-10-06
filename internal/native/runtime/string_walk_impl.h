// Private string walk implementation, included only by string.c.

size_t adamic_string_next(const adamic_string *string, size_t offset) {
	return sequence((unsigned char)string->bytes[offset]);
}

adamic_string *adamic_string_slice_bytes(const adamic_string *string, size_t offset, size_t size) {
	return adamic_string_share(string, offset, size);
}

// A view of a string as UTF-16 code units, walked from the start. Each step yields the unit's value
// and where its code point's bytes are; a supplementary code point is two steps over the same bytes.
typedef struct units {
	const adamic_string *string;
	size_t offset;   // bytes of the current code point
	size_t size;     // its byte length
	bool second;     // on the low half of a surrogate pair
} units;

static bool units_next(units *walk, unsigned *unit) {
	if (walk->second) {
		unsigned point = decode((const unsigned char *)walk->string->bytes + walk->offset, walk->size);
		*unit = 0xdc00 + ((point - 0x10000) & 0x3ff);
		walk->second = false;
		walk->offset += walk->size;
		walk->size = 0;
		return true;
	}
	walk->offset += walk->size;
	if (walk->offset >= walk->string->length) {
		walk->size = 0;
		return false;
	}
	walk->size = sequence((unsigned char)walk->string->bytes[walk->offset]);
	unsigned point = decode((const unsigned char *)walk->string->bytes + walk->offset, walk->size);
	if (walk->size == 4) {
		// The high half now, and the low half from the same bytes on the next step.
		*unit = 0xd800 + ((point - 0x10000) >> 10);
		walk->second = true;
		return true;
	}
	*unit = point;
	return true;
}

static units units_start(const adamic_string *string) {
	units walk = {string, 0, 0, false};
	return walk;
}

