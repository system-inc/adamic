// slab_quarantine.h: a bounded first-in, first-out hold on freed slab slots, for sanitized slab
// builds only (slab_quarantine.c).
//
// It is on only in a sanitized build with ADAMIC_SLABS: the address sanitizer's, the build the slab
// lane tests with, where it catches stale reads, and the thread sanitizer's, where it does nothing
// but run, so its own thread safety meets TSan. Everywhere else ADAMIC_SLAB_QUARANTINE is 0 and
// heap.c compiles as if this file did not exist. ADAMIC_SLAB_QUARANTINE_OFF turns it off in those
// builds too, so a test can show what the quarantine catches by running the same program without it.

#ifndef ADAMIC_SLAB_QUARANTINE_H
#define ADAMIC_SLAB_QUARANTINE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#if defined(__has_feature)
#if __has_feature(address_sanitizer)
#define ADAMIC_SLAB_QUARANTINE_ADDRESS 1
#endif
#if __has_feature(thread_sanitizer)
#define ADAMIC_SLAB_QUARANTINE_THREAD 1
#endif
#endif
#if defined(__SANITIZE_ADDRESS__)
#define ADAMIC_SLAB_QUARANTINE_ADDRESS 1
#endif
#if defined(__SANITIZE_THREAD__)
#define ADAMIC_SLAB_QUARANTINE_THREAD 1
#endif

#if (defined(ADAMIC_SLAB_QUARANTINE_ADDRESS) || defined(ADAMIC_SLAB_QUARANTINE_THREAD)) && defined(ADAMIC_SLABS) && !defined(ADAMIC_SLAB_QUARANTINE_OFF)
#define ADAMIC_SLAB_QUARANTINE 1
#else
#define ADAMIC_SLAB_QUARANTINE 0
#endif

#if ADAMIC_SLAB_QUARANTINE
// adamic_slab_quarantine holds a freed slot of size bytes from chunk number, poisoned, in the
// calling thread's queue. Once that queue is full, taking a slot in evicts its oldest: it is
// unpoisoned, written back through slot and number, and true is returned, and the caller gives it
// back to its chunk. Until then it returns false and the caller gives nothing back.
bool adamic_slab_quarantine(void **slot, uint32_t *number, size_t size);

// adamic_slab_quarantine_drain gives every slot the calling thread holds, oldest first and
// unpoisoned, to give, then frees the thread's queue. A thread calls it as it ends.
void adamic_slab_quarantine_drain(void (*give)(void *slot, uint32_t number));
#endif

#endif
