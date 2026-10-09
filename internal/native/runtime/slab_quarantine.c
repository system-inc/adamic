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
// Each thread holds the slots it frees in its own queue of 4096, made the first time it frees one
// and drained when it ends (heap.c's adamic_heap_thread_end). A slot is at most 256 bytes (heap.c's
// 16 classes of 16 bytes), so a thread keeps at most 1 MB from reuse. Held slots still count as
// live in their chunks, so a chunk with one held never goes to the spares; that is the price of the
// hold, and it is bounded by it. An evicted slot goes back through heap.c's give, to its chunk if
// this thread owns it and to its owner's remote list if not, as any free does.
//
// Per thread rather than one queue under a lock, because the allocator is per thread already and a
// lock would order every free against every other thread's: in a TSan build that order is a
// happens-before edge the program never had, which hides the races the build is there to find.
// A held slot is on no free list and no remote list, so no thread can take it, and its poison is
// in the sanitizer's shadow, which every thread reads; a stale read from any thread is caught for as
// long as the freeing thread holds it, and one busy thread's frees never push out another's.

#include "slab_quarantine.h"

#if ADAMIC_SLAB_QUARANTINE

#include <stdlib.h>

#if defined(ADAMIC_SLAB_QUARANTINE_ADDRESS)
#include <sanitizer/asan_interface.h>
#define POISON(slot, size) ASAN_POISON_MEMORY_REGION((slot), (size))
#define UNPOISON(slot, size) ASAN_UNPOISON_MEMORY_REGION((slot), (size))
#else
#define POISON(slot, size) ((void)(slot), (void)(size))
#define UNPOISON(slot, size) ((void)(slot), (void)(size))
#endif

#define HELD 4096

typedef struct held_slot {
	void *slot;
	uint32_t number;
	size_t size;
} held_slot;

// A ring per thread: next is the oldest once the ring is full, and the place the next slot goes.
static _Thread_local held_slot *held;
static _Thread_local size_t next;

bool adamic_slab_quarantine(void **slot, uint32_t *number, size_t size) {
	if (held == NULL) {
		held = calloc(HELD, sizeof *held);
		if (held == NULL) {
			// Holding nothing loses only the extra catch, never a slot.
			return true;
		}
	}
	POISON(*slot, size);
	held_slot oldest = held[next];
	held[next] = (held_slot){*slot, *number, size};
	next = (next + 1) % HELD;
	if (oldest.slot == NULL) {
		return false;
	}
	// heap.c's give writes the free list's link into the slot before it poisons it again.
	UNPOISON(oldest.slot, oldest.size);
	*slot = oldest.slot;
	*number = oldest.number;
	return true;
}

void adamic_slab_quarantine_drain(void (*give)(void *slot, uint32_t number)) {
	if (held == NULL) {
		return;
	}
	for (size_t index = 0; index < HELD; index++) {
		held_slot each = held[(next + index) % HELD];
		if (each.slot != NULL) {
			UNPOISON(each.slot, each.size);
			give(each.slot, each.number);
		}
	}
	free(held);
	held = NULL;
	next = 0;
}

#else
/* C11 forbids an empty translation unit under -pedantic. */
typedef int adamic_slab_quarantine_not_linked;
#endif
