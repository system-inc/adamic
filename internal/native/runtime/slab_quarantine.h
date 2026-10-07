// slab_quarantine.h: a bounded first-in, first-out hold on freed slab slots, for sanitized slab
// builds only (slab_quarantine.c).
//
// It is on only where the address sanitizer and ADAMIC_SLABS are both on, the build the slab lane
// tests with. Everywhere else ADAMIC_SLAB_QUARANTINE is 0 and heap.c compiles as if this file did
// not exist. ADAMIC_SLAB_QUARANTINE_OFF turns it off in that build too, so a test can show what
// the quarantine catches by running the same program without it.

#ifndef ADAMIC_SLAB_QUARANTINE_H
#define ADAMIC_SLAB_QUARANTINE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#if defined(__has_feature)
#if __has_feature(address_sanitizer)
#define ADAMIC_SLAB_QUARANTINE_SANITIZED 1
#endif
#endif
#if defined(__SANITIZE_ADDRESS__)
#define ADAMIC_SLAB_QUARANTINE_SANITIZED 1
#endif

#if defined(ADAMIC_SLAB_QUARANTINE_SANITIZED) && defined(ADAMIC_SLABS) && !defined(ADAMIC_SLAB_QUARANTINE_OFF)
#define ADAMIC_SLAB_QUARANTINE 1
#else
#define ADAMIC_SLAB_QUARANTINE 0
#endif

#if ADAMIC_SLAB_QUARANTINE
// adamic_slab_quarantine holds a freed slot of size bytes from chunk number, poisoned. Once the
// hold is full, taking a slot in evicts the oldest: it is unpoisoned, written back through slot
// and number, and true is returned, and the caller gives it to its chunk's free list. Until then
// it returns false and the caller gives nothing back.
bool adamic_slab_quarantine(void **slot, uint32_t *number, size_t size);
#endif

#endif
