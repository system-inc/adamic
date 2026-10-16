// count.h: what a counted build counts (native.Options.Count, which defines ADAMIC_COUNT). Counts are
// deterministic, unlike time, so they can be measured anywhere: the oracle's counts table holds every
// fixture's, and borrow inference and reuse in place are measured against it (docs/memory.md).
//
// An allocation is a heap value made (adamic_allocate) and a free is one freed; the buffers behind an
// array's elements or a map's table aren't heap values and aren't counted. A retain or a release is
// a call to adamic_retain or adamic_release, whatever it's given: each one is work the compiler asked
// for, even on an immortal constant or undefined. Outside a counted build every hook is nothing.

#ifndef ADAMIC_COUNT_H
#define ADAMIC_COUNT_H

#include <stddef.h>
#include <stdatomic.h>

typedef struct adamic_counts {
	_Atomic size_t allocations;
	_Atomic size_t frees;
	_Atomic size_t retains;
	_Atomic size_t releases;
	// live is allocations less frees, and peak the most that were ever live at once.
	_Atomic size_t live;
	_Atomic size_t peak;
	// regions is the values let go of with their region rather than freed one at a time (region.c):
	// a finished program's allocations are its frees and its regions.
	_Atomic size_t regions;
} adamic_counts;

extern adamic_counts adamic_counted;

void adamic_count_allocation(void);

// adamic_count_report writes the counts to stderr, as one line the oracle reads:
// adamic: counts: allocations 3 frees 3 retains 7 releases 10 peak 2 regions 0
void adamic_count_report(void);

#ifdef ADAMIC_COUNT
#define ADAMIC_COUNT_ALLOCATION() adamic_count_allocation()
#define ADAMIC_COUNT_FREE() (atomic_fetch_add_explicit(&adamic_counted.frees, 1, memory_order_relaxed), atomic_fetch_sub_explicit(&adamic_counted.live, 1, memory_order_relaxed))
#define ADAMIC_COUNT_RETAIN() atomic_fetch_add_explicit(&adamic_counted.retains, 1, memory_order_relaxed)
#define ADAMIC_COUNT_RELEASE() atomic_fetch_add_explicit(&adamic_counted.releases, 1, memory_order_relaxed)
#define ADAMIC_COUNT_REPORT() adamic_count_report()
#define ADAMIC_COUNT_REGION(count) (atomic_fetch_add_explicit(&adamic_counted.regions, (count), memory_order_relaxed), atomic_fetch_sub_explicit(&adamic_counted.live, (count), memory_order_relaxed))
#else
#define ADAMIC_COUNT_ALLOCATION() ((void)0)
#define ADAMIC_COUNT_FREE() ((void)0)
#define ADAMIC_COUNT_RETAIN() ((void)0)
#define ADAMIC_COUNT_RELEASE() ((void)0)
#define ADAMIC_COUNT_REPORT() ((void)0)
#define ADAMIC_COUNT_REGION(count) ((void)(count))
#endif

#endif
