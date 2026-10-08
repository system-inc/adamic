// library_language.c: fixed plain-object enumeration and opaque built-in identities.

#include "adamic.h"

#include <string.h>

// An array index is a canonical decimal string in [0, 2^32 - 2]. In particular, 00, -0 and
// 4294967295 are ordinary string keys. Check before multiplying so arbitrarily long names cannot
// wrap an integer into an index.
static bool array_index(const char *name, uint32_t *value) {
	if (*name == '\0' || (name[0] == '0' && name[1] != '\0')) {
		return false;
	}
	uint64_t number = 0;
	for (const unsigned char *cursor = (const unsigned char *)name; *cursor != '\0'; cursor++) {
		if (*cursor < '0' || *cursor > '9') {
			return false;
		}
		number = number * 10 + (*cursor - '0');
		if (number >= UINT32_MAX) {
			return false;
		}
	}
	*value = (uint32_t)number;
	return true;
}

adamic_array *adamic_plain_object_keys(const adamic_object *object) {
	const adamic_shape *shape = object->shape;
	adamic_array *keys = adamic_array_new_typed(shape->count, true, &adamic_json_string_schema);
	// Insertion-sort just the integer keys, then append ordinary strings in shape order. Runtime
	// shapes retain insertion order even through a narrower static type and a spread.
	for (size_t position = 0; position < shape->count; position++) {
		uint32_t index;
		const char *name = shape->names[position];
		if (!array_index(name, &index)) {
			continue;
		}
		adamic_string *key = adamic_string_allocate(strlen(name));
		memcpy((char *)key->bytes, name, key->length);
		size_t slot = keys->length;
		adamic_array_push(keys, (adamic_value){.reference = key});
		while (slot > 0) {
			uint32_t previous;
			const adamic_string *before = keys->elements[slot - 1].reference;
			// Keys' allocated bytes aren't NUL-terminated. Accumulate their already-proved digits.
			previous = 0;
			for (size_t digit = 0; digit < before->length; digit++) {
				previous = previous * 10 + (uint32_t)(before->bytes[digit] - '0');
			}
			if (previous < index) {
				break;
			}
			keys->elements[slot] = keys->elements[slot - 1];
			slot--;
		}
		keys->elements[slot].reference = key;
	}
	for (size_t position = 0; position < shape->count; position++) {
		uint32_t index;
		const char *name = shape->names[position];
		if (array_index(name, &index)) {
			continue;
		}
		adamic_string *key = adamic_string_allocate(strlen(name));
		memcpy((char *)key->bytes, name, key->length);
		adamic_array_push(keys, (adamic_value){.reference = key});
	}
	return keys;
}

// These are identities, not implementations of the constructors. Lowering prevents calls and
// property reads through them. The kind still gives typeof and boxed identity comparisons their
// real JavaScript answer, and a zero count makes them immortal.
void *adamic_library_identity(size_t index) {
	static adamic_heap identities[] = {
		{0, adamic_kind_closure, 0}, {0, adamic_kind_closure, 0},
		{0, adamic_kind_closure, 0}, {0, adamic_kind_closure, 0},
		{0, adamic_kind_object, 0}, {0, adamic_kind_closure, 0},
		{0, adamic_kind_closure, 0},
	};
	return &identities[index];
}
