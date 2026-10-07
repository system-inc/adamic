// slab_quarantine.c: freed slab slots wait here before they can be taken again (slab_quarantine.h).
//
// The address sanitizer poisons a freed slot, but heap.c's take unpoisons it the moment it hands
// the slot out again, and its free list gives back the slot freed last. So with slabs on, a value
// freed and then read through a stale pointer is caught only until the next value of its class is
// made; after that the stale pointer reads the new value as if it were its own. malloc under ASan
// avoids this with its own quarantine, and this is the same for slots: a freed slot stays poisoned
// here, first in, first out, and reaches its chunk's free list only when it is evicted. A stale use
// is then caught for as long as the slot is held.
//
// The hold is 4096 slots. A slot is at most 256 bytes (heap.c's 16 classes of 16 bytes), so it
// keeps at most 1 MB from reuse. Held slots still count as live in their chunks, so a chunk with
// one held never goes to the spares; that is the price of the hold, and it is bounded by it.
// Like the rest of heap.c this follows the single-threaded counting model.

#include "slab_quarantine.h"

#if ADAMIC_SLAB_QUARANTINE

#include <sanitizer/asan_interface.h>

#define HELD 4096

typedef struct held_slot {
	void *slot;
	uint32_t number;
	size_t size;
} held_slot;

// A ring: next is the oldest once the ring is full, and the place the next slot goes.
static held_slot held[HELD];
static size_t next;

bool adamic_slab_quarantine(void **slot, uint32_t *number, size_t size) {
	ASAN_POISON_MEMORY_REGION(*slot, size);
	held_slot oldest = held[next];
	held[next] = (held_slot){*slot, *number, size};
	next = (next + 1) % HELD;
	if (oldest.slot == NULL) {
		return false;
	}
	// heap.c's give writes the free list's link into the slot before it poisons it again.
	ASAN_UNPOISON_MEMORY_REGION(oldest.slot, oldest.size);
	*slot = oldest.slot;
	*number = oldest.number;
	return true;
}

#else
/* C11 forbids an empty translation unit under -pedantic. */
typedef int adamic_slab_quarantine_not_linked;
#endif
