// string_append.c: text += more, in linear time.
//
// A string is immutable to the program, but when a local holds the only reference to a string, no one
// else can see it change: appending to it can write after its last byte, if there's room, instead of
// copying everything before. When there isn't room, the new string gets twice what it needs, so a
// loop that appends n bytes copies each byte a constant number of times on average, and allocates
// about log n times, where copying every time is quadratic. The emitter calls this for an assignment
// text = text + ..., of a local nothing else can write while the parts are evaluated (emit.go).
//
// The reference count is what says "only": a string an array, a map, an object, another local, a
// shared slice (string_share.c) or a caller holds has a count above one, and a constant's is 0. Only a
// string with bytes of its own has room (capacity); a shared slice's are its owner's, and never
// written.

#include "adamic.h"

#include <stdint.h>
#include <string.h>

// grown makes a string of length bytes with room for capacity, references 1, laid out as string.c's
// allocate lays them out.
static adamic_string *grown(size_t length, size_t capacity) {
	adamic_string *string = adamic_allocate(sizeof *string + capacity, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	string->units = 0;
	string->index = NULL;
	string->owner = NULL;
	string->capacity = capacity;
	return string;
}

adamic_string *adamic_string_append(adamic_string *string, size_t count, adamic_string *const parts[]) {
	size_t length = string->length, added = 0;
	for (size_t index = 0; index < count; index++) {
		if (parts[index]->length > SIZE_MAX / 2 - length - added) {
			static const char message[] = "string too long";
			adamic_panic(message, sizeof message - 1);
		}
		added += parts[index]->length;
	}
	// A string's UTF-16 units are never more than its bytes, so only a long one needs counting.
	if (length + added > ADAMIC_STRING_MAX_UNITS) {
		double units = adamic_string_length(string);
		for (size_t index = 0; index < count; index++) {
			units += adamic_string_length(parts[index]);
		}
		adamic_string_check_length(units);
	}
	// text += text can't be written in place: joining halves where the pieces meet would rewrite the
	// bytes the copy of text is still reading. Room is never as much as the length today (a string
	// grows to twice what it held before), so this guards a change to how strings grow.
	bool itself = false;
	for (size_t index = 0; index < count; index++) {
		itself = itself || parts[index] == string;
	}
	// Room is capacity against the length after: a shared slice, or a slice of a constant, has a count
	// of 1 and a capacity of 0 below its length, and capacity - length would wrap around to nearly all
	// of memory and write into bytes that are its owner's. length + added can't overflow, as the loop
	// above held it under SIZE_MAX / 2.
	adamic_string *result = string;
	if (string->heap.references == 1 && !itself && length + added <= string->capacity) {
		// The bytes are about to change, so what was cached about the old ones goes: the length in
		// units, and the position index (string_index.c).
		adamic_string_free_index(string);
		string->index = NULL;
		string->units = 0;
	} else {
		size_t capacity = length + added;
		if (capacity < 2 * length) {
			capacity = 2 * length;
		}
		result = grown(length, capacity);
		if (length > 0) {
			memcpy((char *)result->bytes, string->bytes, length);
		}
	}
	// A part that is the string itself is read from its own bytes, untouched, since that append
	// made a new string.
	size_t written = length;
	for (size_t index = 0; index < count; index++) {
		written = adamic_string_put((char *)result->bytes, written, parts[index]);
	}
	result->length = written;
	if (result != string) {
		adamic_release(string);
	}
	return result;
}
