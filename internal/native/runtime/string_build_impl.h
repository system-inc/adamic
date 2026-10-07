// Private string build implementation, included only by string.c.

// allocate makes a string of length bytes, references 1, its bytes right after it in one block.
static adamic_string *allocate(size_t length) {
	adamic_string *string = adamic_allocate(sizeof *string + length, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	string->units = 0;
	string->index = NULL;
	string->owner = NULL;
	string->capacity = length;
	return string;
}

adamic_string *adamic_string_allocate(size_t length) {
	return allocate(length);
}

adamic_string *adamic_string_from_number(double value) {
	char buffer[ADAMIC_NUMBER_FORMAT_MAX];
	size_t length = adamic_number_format(value, buffer);
	adamic_string *string = allocate(length);
	memcpy((char *)string->bytes, buffer, length);
	return string;
}

// V8's longest string, in UTF-16 units (String::kMaxLength on 64-bit): past it, every way of making
// a string throws RangeError: Invalid string length, and so Adamic panics there too.
void adamic_string_check_length(double units) {
	if (units > ADAMIC_STRING_MAX_UNITS) {
		static const char message[] = "RangeError: Invalid string length";
		adamic_panic(message, sizeof message - 1);
	}
}

adamic_string *adamic_string_concat(size_t count, adamic_string *const parts[]) {
	size_t length = 0;
	for (size_t index = 0; index < count; index++) {
		if (parts[index]->length > SIZE_MAX - length) {
			static const char message[] = "string too long";
			adamic_panic(message, sizeof message - 1);
		}
		length += parts[index]->length;
	}
	// A string's UTF-16 units are never more than its bytes, so only a long one needs counting.
	if (length > ADAMIC_STRING_MAX_UNITS) {
		double units = 0;
		for (size_t index = 0; index < count; index++) {
			units += adamic_string_length(parts[index]);
		}
		adamic_string_check_length(units);
	}
	adamic_string *string = allocate(length);
	if (count == 1) {
		// One piece is a copy, and its bytes may be a builder's (a stack piece, from fromCharCode or
		// slice), which can hold halves of a pair side by side: every byte is looked at.
		if (length > 0) {
			memcpy((char *)string->bytes, parts[0]->bytes, length);
		}
		string->length = adamic_string_join_halves((char *)string->bytes, 0, length);
		return string;
	}
	size_t written = 0;
	for (size_t index = 0; index < count; index++) {
		written = adamic_string_put((char *)string->bytes, written, parts[index]);
	}
	string->length = written;
	return string;
}

size_t adamic_string_put(char *bytes, size_t written, const adamic_string *part) {
	// A string's own halves are joined already, so halves of a pair can meet only where two pieces
	// do: a lone high surrogate the bytes so far end with, and a lone low one the part begins with.
	// The part's first three bytes go in first, and only those six are looked at.
	size_t head = part->length < 3 ? part->length : 3;
	if (head > 0) {
		memcpy(bytes + written, part->bytes, head);
	}
	written = adamic_string_join_halves(bytes, written >= 3 ? written - 3 : 0, written + head);
	if (part->length > head) {
		memcpy(bytes + written, part->bytes + head, part->length - head);
	}
	return written + part->length - head;
}

size_t adamic_string_join_halves(char *bytes, size_t from, size_t length) {
	// A lone high surrogate followed by a lone low one is a character again, as in JavaScript, and
	// WTF-8 writes it as UTF-8's four bytes, so equal strings stay equal byte for byte.
	size_t kept = from;
	for (size_t at = from; at < length;) {
		unsigned char *here = (unsigned char *)bytes + at;
		if (at + 6 <= length && here[0] == 0xed && here[1] >= 0xa0 && here[1] <= 0xaf && here[3] == 0xed && here[4] >= 0xb0 && here[4] <= 0xbf) {
			unsigned high = decode(here, 3), low = decode(here + 3, 3);
			unsigned point = 0x10000 + ((high - 0xd800) << 10) + (low - 0xdc00);
			bytes[kept++] = (char)(0xf0 | (point >> 18));
			bytes[kept++] = (char)(0x80 | ((point >> 12) & 0x3f));
			bytes[kept++] = (char)(0x80 | ((point >> 6) & 0x3f));
			bytes[kept++] = (char)(0x80 | (point & 0x3f));
			at += 6;
			continue;
		}
		bytes[kept++] = bytes[at++];
	}
	return kept;
}

int adamic_string_equal(const adamic_string *left, const adamic_string *right) {
	// Either may be undefined (a string | undefined): undefined is equal only to itself.
	if (left == NULL || right == NULL) {
		return left == right;
	}
	return left->length == right->length && (left->length == 0 || memcmp(left->bytes, right->bytes, left->length) == 0);
}

adamic_string adamic_string_empty = ADAMIC_STRING("");
adamic_string adamic_string_true = ADAMIC_STRING("true");
adamic_string adamic_string_false = ADAMIC_STRING("false");

