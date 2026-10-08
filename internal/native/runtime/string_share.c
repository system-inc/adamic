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
// Sharing retains the ultimate owner, including its spare append capacity. Copy only when that
// storage exceeds SHARE_FRACTION times the view's own header and bytes. Including the header
// lets short views share small parents, without a minimum substring size. Constants pin nothing.

#include "adamic.h"

#include <string.h>

#ifndef SHARE_FRACTION
#define SHARE_FRACTION 8
#endif

adamic_string *adamic_string_share(const adamic_string *string, size_t offset, size_t size) {
	if (size == 0) {
		return adamic_retain(&adamic_string_empty);
	}
	const adamic_string *owner = string->owner != NULL ? string->owner : string;
	// A zero-count stack piece is borrowed, unlike a marked literal. It has no count a view
	// can keep, so even a whole slice must copy. Built literal indexes remain non-NULL.
	bool borrowed = owner->heap.references == 0 && owner->index == NULL;
	if (!borrowed && offset == 0 && size == string->length) {
		return adamic_retain((adamic_string *)string);
	}
	size_t storage = owner->capacity > owner->length ? owner->capacity : owner->length;
	// Avoid multiplication and addition overflow. ceil((header + storage) / fraction)
	// is at most header + view bytes exactly when the retained-storage bound holds.
	size_t minimum = storage / SHARE_FRACTION + sizeof *owner / SHARE_FRACTION;
	size_t remainder = storage % SHARE_FRACTION + sizeof *owner % SHARE_FRACTION;
	minimum += remainder / SHARE_FRACTION + (remainder % SHARE_FRACTION != 0);
	bool oversized = minimum > sizeof *owner && size < minimum - sizeof *owner;
	if (borrowed || (owner->heap.references != 0 && oversized)) {
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
