// Copyright 2022 the V8 project authors. All rights reserved.
// BSD-3-Clause, reproduced in THIRD_PARTY_NOTICES.md.
// Port of src/builtins/string-iswellformed.tq and string-towellformed.tq.
// Adamic's canonical WTF-8 joins every paired surrogate into a four-byte point.
// Thus every three-byte surrogate is unpaired, and replacement preserves byte length.
#include "adamic.h"

#include <string.h>

static size_t string_wellformed_width(unsigned char lead) {
	return lead < 0x80 ? 1 : lead < 0xe0 ? 2 : lead < 0xf0 ? 3 : 4;
}

static bool string_wellformed_unpaired(const unsigned char *bytes) {
	return bytes[0] == 0xed && bytes[1] >= 0xa0;
}

bool adamic_string_is_well_formed(const adamic_string *string) {
	for (size_t at = 0; at < string->length;) {
		const unsigned char *bytes = (const unsigned char *)string->bytes + at;
		size_t width = string_wellformed_width(bytes[0]);
		if (width == 3 && string_wellformed_unpaired(bytes)) { return false; }
		at += width;
	}
	return true;
}

adamic_string *adamic_string_to_well_formed(const adamic_string *string) {
	if (adamic_string_is_well_formed(string)) { return adamic_retain((adamic_string *)string); }
	adamic_string *result = adamic_string_allocate(string->length);
	for (size_t at = 0; at < string->length;) {
		const unsigned char *bytes = (const unsigned char *)string->bytes + at;
		size_t width = string_wellformed_width(bytes[0]);
		if (width == 3 && string_wellformed_unpaired(bytes)) {
			memcpy((char *)result->bytes + at, "\xef\xbf\xbd", 3);
		} else { memcpy((char *)result->bytes + at, bytes, width); }
		at += width;
	}
	return result;
}
