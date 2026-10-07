// string_share.c: a slice of a string that reads its parent's bytes instead of copying them.
//
// A string is immutable, so a slice can point into the bytes of the string it was cut from, as a Go
// string does, and hold a reference to that string so the bytes outlive it. The slice is a string like
// any other: its own length, and its own caches (units and index, string_index.c), which start empty
// and are counted from its own bytes, so a checkpoint or a cursor is an offset into the slice, never
// into its parent. A slice of a slice points at the first one's owner, so no chain grows.
//
// Nothing ever writes into a string's bytes once it's made. If strings are ever reused in place (an
// append to a string only one holder has), a shared slice's bytes aren't its own to write, nor are
// the bytes of a string some slice reads (its count is more than one, so it isn't reused), and reuse
// would have to reset the cached units and index, which describe the old bytes.
//
// Sharing keeps the whole owner alive for as long as the slice is, so a short slice of a long string
// is copied instead: a slice shares only when it is at least SHARE_MINIMUM bytes, below which a copy
// is as cheap as the header, and at least 1/SHARE_FRACTION of its owner, so it never keeps more than
// SHARE_FRACTION times its own bytes alive.

#include "adamic.h"

#include <string.h>

#ifndef SHARE_MINIMUM
#define SHARE_MINIMUM 64
#endif
#ifndef SHARE_FRACTION
#define SHARE_FRACTION 4
#endif

adamic_string *adamic_string_share(const adamic_string *string, size_t offset, size_t size) {
	if (offset == 0 && size == string->length) {
		// The whole string is itself.
		return adamic_retain((adamic_string *)string);
	}
	const adamic_string *owner = string->owner != NULL ? string->owner : string;
	if (size < SHARE_MINIMUM || size < owner->length / SHARE_FRACTION) {
		adamic_string *copy = adamic_string_allocate(size);
		if (size > 0) {
			memcpy((char *)copy->bytes, string->bytes + offset, size);
		}
		if (string->units == string->length + 1) {
			copy->units = size + 1;
		}
		return copy;
	}
	adamic_string *shared = adamic_allocate(sizeof *shared, adamic_kind_string);
	shared->length = size;
	shared->bytes = string->bytes + offset;
	shared->units = string->units == string->length + 1 ? size + 1 : 0;
	shared->index = NULL;
	shared->capacity = 0;
	// A constant's bytes last as long as the program, and need no one held for them.
	shared->owner = adamic_reference_count(&owner->heap) == 0 ? NULL : adamic_retain((adamic_string *)owner);
	return shared;
}
