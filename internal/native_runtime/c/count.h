// count.h: what a counted build counts (native.Options.Count, which defines ADAMIC_COUNT). Counts are
// deterministic, unlike time, so they can be measured anywhere: the oracle's counts table holds every
// fixture's, and borrow inference and reuse in place are measured against it (docs/memory.md).
//
// An allocation is a heap value made (adamic_allocate) and a free is one freed; the buffers behind an
// array's elements, a map's table or a string being built aren't heap values and aren't counted as
// allocations. Their growth is counted instead (#42q54q5): a growth is one regrowth of such a buffer,
// a realloc that enlarges it or a rebuild into a larger one. A buffer's first allocation isn't one: the
// value it belongs to is already an allocation, or it's scratch freed in the same call. A retain or a
// release is a call to adamic_retain or adamic_release, whatever it's given: each one is work the
// compiler asked for, even on an immortal constant or undefined. Outside a counted build every hook is
// nothing.

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
	_Atomic size_t graph_regions;
	_Atomic size_t graph_merges;
	// graph_searched is how many counted edges merges' searches read (graph_regions.c's between).
	_Atomic size_t graph_searched;
	// graph_moves is how many values graph adoption moved into graph storage, a copy and a free each
	// (graph_regions.c's adamic_graph_adopt); a value born there (adamic_heap_graph_allocate) isn't
	// one. A harness reads it, never a counts line, so every line keeps its columns.
	_Atomic size_t graph_moves;
	// region_blocks is the blocks region.c asked malloc for and region_blocks_freed the ones it gave
	// back: once every thread's cache is let go (adamic_region_thread_end), they're equal (#yecmp1d).
	_Atomic size_t region_blocks;
	_Atomic size_t region_blocks_freed;
	// heap_chunks_released is the spare chunks heap.c gave back to the kernel at a quiet point
	// (#vea9pph), written on a line of its own when there are any, never a column of the counts line.
	_Atomic size_t heap_chunks_released;
	// weak_entries is the entries WeakMaps and WeakSets hold (weak_collection.c), and weak_entries_peak
	// the most they ever held at once, written on a line of their own when there were any.
	_Atomic size_t weak_entries;
	_Atomic size_t weak_entries_peak;
	// join_capacity_peak is a buffer's size rather than a count: the largest capacity in bytes
	// adamic_array_join's buffer ever reached (array.c). #sbvs9d4's pin reads it to hold the join's
	// running length check, since resident memory moves with the box. Never printed.
	_Atomic size_t join_capacity_peak;
	// growths is the buffers' regrowths, as the header's first paragraph has them (ADAMIC_COUNT_GROWTH
	// at each site). A harness reads it through adamic_count_snapshot, never a counts line, so every
	// line keeps its columns.
	_Atomic size_t growths;
} adamic_counts;

extern adamic_counts adamic_counted;

void adamic_count_allocation(void);

// adamic_count_report writes the counts to stderr, as one line the oracle reads:
// adamic: counts: allocations 3 frees 3 retains 7 releases 10 peak 2 regions 0
void adamic_count_report(void);

#ifdef ADAMIC_COUNT
// adamic_count_window is the counts a harness subtracts: two snapshots taken around one call, the later
// less the earlier, are the allocations, frees and growths that call made. It's a measurement, never a
// proof that a loop's body doesn't allocate (#42q54q5).
typedef struct adamic_count_window {
	size_t allocations;
	size_t frees;
	size_t growths;
} adamic_count_window;

// adamic_count_snapshot reads the counts, each relaxed. They're process-wide by design, not per thread:
// besides the program's own thread only the signal thread and the pool's workers ever run, and an idle
// worker makes and frees nothing, so the difference across a synchronous call is that call's own work,
// a parallelMap's workers included.
adamic_count_window adamic_count_snapshot(void);
#endif

#ifdef ADAMIC_COUNT
#define ADAMIC_COUNT_ALLOCATION() adamic_count_allocation()
#define ADAMIC_COUNT_FREE() (atomic_fetch_add_explicit(&adamic_counted.frees, 1, memory_order_relaxed), atomic_fetch_sub_explicit(&adamic_counted.live, 1, memory_order_relaxed))
#define ADAMIC_COUNT_RETAIN() atomic_fetch_add_explicit(&adamic_counted.retains, 1, memory_order_relaxed)
#define ADAMIC_COUNT_RELEASE() atomic_fetch_add_explicit(&adamic_counted.releases, 1, memory_order_relaxed)
#define ADAMIC_COUNT_REPORT() adamic_count_report()
#define ADAMIC_COUNT_REGION(count) (atomic_fetch_add_explicit(&adamic_counted.regions, (count), memory_order_relaxed), atomic_fetch_sub_explicit(&adamic_counted.live, (count), memory_order_relaxed))
// ADAMIC_COUNT_GROWTH counts one regrowth of a buffer a value owns (the header's first paragraph).
#define ADAMIC_COUNT_GROWTH() atomic_fetch_add_explicit(&adamic_counted.growths, 1, memory_order_relaxed)
// ADAMIC_COUNT_JOIN_CAPACITY raises join_capacity_peak to a join buffer's capacity, as
// adamic_count_allocation raises peak.
#define ADAMIC_COUNT_JOIN_CAPACITY(capacity) do { \
	size_t adamic_count_join_capacity = (capacity); \
	size_t adamic_count_join_peak = atomic_load_explicit(&adamic_counted.join_capacity_peak, memory_order_relaxed); \
	while (adamic_count_join_peak < adamic_count_join_capacity && !atomic_compare_exchange_weak_explicit(&adamic_counted.join_capacity_peak, &adamic_count_join_peak, adamic_count_join_capacity, memory_order_relaxed, memory_order_relaxed)) {} \
} while (0)
#else
#define ADAMIC_COUNT_ALLOCATION() ((void)0)
#define ADAMIC_COUNT_FREE() ((void)0)
#define ADAMIC_COUNT_RETAIN() ((void)0)
#define ADAMIC_COUNT_RELEASE() ((void)0)
#define ADAMIC_COUNT_REPORT() ((void)0)
#define ADAMIC_COUNT_REGION(count) ((void)(count))
#define ADAMIC_COUNT_GROWTH() ((void)0)
#define ADAMIC_COUNT_JOIN_CAPACITY(capacity) ((void)(capacity))
#endif

#endif
