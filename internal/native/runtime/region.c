// region.c: regions, the arenas of docs/memory.md. A region lives for one statement and holds the
// objects that statement makes and nothing reaches after it: they're bump-allocated from the
// region's blocks and let go of together when the statement ends, instead of one free at a time.
//
// An object in a region is immortal while the region lives (its count is 0, as a constant's is), so
// retain and release pass over it. When the region ends, it releases what its objects hold outside
// it (an immortal child is passed over), tells any Weak that pointed at one that it's gone, and frees
// its blocks. The blocks are ordinary malloc blocks, so a use after the region ended (an escape the
// compiler missed) is a use after free to AddressSanitizer.
//
// A thread keeps the blocks its regions free, up to ADAMIC_REGION_BLOCK_CACHE bytes, and the next
// region it opens takes them back before mapping new ones (#yecmp1d, Adamic's ruling): malloc handed
// out pages the kernel faulted in again for every region, a third of bench/trees' time in the kernel.
// A standard block is a mapping of its own, and one the cache doesn't keep is unmapped, so what a
// thread holds after its last region is the cache and nothing more. Kept in malloc's arena instead,
// the cached blocks held glibc's top from trimming and the process kept about 50 MB a thread while
// idle (Codex 6af450b5, memory/probe-region-rss), against the ruling's at most the cap. A
// block is taken and given on one thread, since a region is a stack value its own function opens and
// ends, and a parallelMap task, a closure, is never handed its caller's region (only a direct call's
// region version takes one). The cache is off under AddressSanitizer, so a use after a region ends is
// still a use after free there.
//
// Past 1 MB a region grows by 2 MB blocks, each aligned to 2 MB and madvised for a huge page, on
// Linux (#cgn2cwt): trees' regions took 1,969 more 1 MB blocks after the ladder, about 2.3 GB that
// faulted in a 4 KB page at a time, 626K faults and most of its system time. One fault backs a whole
// huge block. A thread keeps one huge block idle when the cap holds it whole (cached_huge), so a
// cap below 2 MB never keeps one and what a thread keeps idle is still the cap.

// mmap's MAP_ANONYMOUS under -std=c11 on glibc (the block cache below).
#define _DEFAULT_SOURCE
#include "adamic.h"
#include "count.h"

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

// size is what the block holds for objects and used how much of it is taken. turn is the standard
// size, 4 KB doubling to 1 MB and then 2 MB where huge blocks are, of the turn the block was made in,
// header included, which the next turn doubles whether the block was mapped or malloc's. mapping is
// the whole mapping a mapped block is, its turn, or 0 for a block malloc made.
struct adamic_region_block {
	struct adamic_region_block *next;
	size_t size;
	size_t used;
	size_t turn;
	size_t mapping;
	_Alignas(16) unsigned char bytes[];
};

// ALIGN is every object's alignment in a block, and object_size an object's place in one, rounded
// up to it, so the end can walk a block from object to object.
#define ALIGN 16

static size_t object_size(size_t count) {
	size_t size = adamic_object_size(count);
	return (size + ALIGN - 1) & ~(size_t)(ALIGN - 1);
}

// ADAMIC_REGION_BLOCK_CACHE is what a thread keeps after its last region ends: a build constant, never
// read from the environment, so a program behaves the same whatever shell runs it. At 1 MB, trees runs
// 3.12 s against main's 3.78 under glibc as shipped and 3.22 against 3.51 with its thresholds pinned,
// faults 627K against 750K (Runtime Performance on Home, #yecmp1d). The 1.91 s a malloc'd cache showed
// came from glibc keeping freed arenas warm, about 50 MB a thread, which is the retention this rules out.
// trees' regions outgrow 1 MB, so most of its remaining time is allocation, not the cache (Adamic's
// ruling). A larger cap for a target is a build input in its unit key, after a measurement.
#ifndef ADAMIC_REGION_BLOCK_CACHE
#define ADAMIC_REGION_BLOCK_CACHE ((size_t)1048576)
#endif
#if defined(__has_feature)
#if __has_feature(address_sanitizer)
#define REGION_BLOCK_CACHE_OFF 1
#endif
#endif
#if defined(__SANITIZE_ADDRESS__)
#define REGION_BLOCK_CACHE_OFF 1
#endif
// A target with no mmap keeps malloc's blocks and no cache.
#if defined(ADAMIC_TARGET_WASI) || defined(_WIN32)
#define REGION_BLOCK_CACHE_OFF 1
#endif
#ifndef REGION_BLOCK_CACHE_OFF
#include <sys/mman.h>
#endif

// Huge blocks use Linux's transparent huge pages, which the house boxes give only to a mapping that
// asks for them (THP at madvise). Not under any sanitizer, so a sanitized build runs the blocks it
// always has; AddressSanitizer already turns the cache off.
#if defined(__has_feature)
#if __has_feature(thread_sanitizer) || __has_feature(memory_sanitizer)
#define REGION_BLOCK_HUGE_OFF 1
#endif
#endif
#if defined(__SANITIZE_THREAD__)
#define REGION_BLOCK_HUGE_OFF 1
#endif
#if defined(__linux__) && !defined(__wasm32__) && !defined(REGION_BLOCK_CACHE_OFF) && !defined(REGION_BLOCK_HUGE_OFF) && defined(MADV_HUGEPAGE)
#define REGION_BLOCK_HUGE 1
#endif

// One list per standard block size, a mapping of 4 KB doubling to 1 MB. A block made for an object
// bigger than its turn's size is malloc's, and goes back to it.
#define BLOCK_SIZES 9
#define BLOCK_LARGEST ((size_t)1048576)
// BLOCK_HUGE, the turn after BLOCK_LARGEST, is one huge page on x86-64 and on arm64 with 4 KB pages.
#define BLOCK_HUGE ((size_t)2097152)
static _Thread_local adamic_region_block *cached[BLOCK_SIZES];
static _Thread_local size_t cached_bytes;
#ifdef REGION_BLOCK_HUGE
// cached_huge is the one huge block a thread keeps, counted in cached_bytes with the standard lists.
static _Thread_local adamic_region_block *cached_huge;
#endif

static void count_block(bool made) {
#ifdef ADAMIC_COUNT
	atomic_fetch_add_explicit(made ? &adamic_counted.region_blocks : &adamic_counted.region_blocks_freed, 1, memory_order_relaxed);
#else
	(void)made;
#endif
}

#ifndef REGION_BLOCK_CACHE_OFF
static int size_index(size_t mapping) {
	for (int index = 0; index < BLOCK_SIZES; index++) {
		if (mapping == ((size_t)4096 << index)) { return index; }
	}
	return -1;
}
#endif

#ifdef REGION_BLOCK_HUGE
// huge_map is a BLOCK_HUGE mapping aligned to BLOCK_HUGE, or NULL. The kernel backs only an aligned
// 2 MB stretch with a huge page, so twice the size is mapped, which holds one such stretch, and the
// head and tail around it are unmapped. The madvise asks for the huge page; where the kernel can't
// give one, the block is still ordinary memory.
static void *huge_map(void) {
	unsigned char *mapped = mmap(NULL, 2 * BLOCK_HUGE, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
	if (mapped == MAP_FAILED) { return NULL; }
	uintptr_t start = (uintptr_t)mapped;
	size_t head = (size_t)(((start + BLOCK_HUGE - 1) & ~(uintptr_t)(BLOCK_HUGE - 1)) - start);
	if (head != 0) { munmap(mapped, head); }
	munmap(mapped + head + BLOCK_HUGE, BLOCK_HUGE - head);
	madvise(mapped + head, BLOCK_HUGE, MADV_HUGEPAGE);
	return mapped + head;
}
#endif

// block_take is a block for at least size bytes of objects, whose turn is a mapping of total bytes:
// the cache's, a new mapping, or, for an object bigger than the turn holds, malloc's.
static adamic_region_block *block_take(size_t total, size_t size) {
	adamic_region_block *block;
#ifdef REGION_BLOCK_HUGE
	if (total == BLOCK_HUGE && total - sizeof *block >= size) {
		if (cached_huge != NULL) {
			block = cached_huge;
			cached_huge = NULL;
			cached_bytes -= BLOCK_HUGE;
			return block;
		}
		void *mapped = huge_map();
		// As for a standard block, a failed mapping leaves this one to malloc.
		if (mapped != NULL) {
			block = mapped;
			block->turn = block->mapping = total;
			block->size = total - sizeof *block;
			count_block(true);
			return block;
		}
	}
#endif
#ifndef REGION_BLOCK_CACHE_OFF
	int index = size_index(total);
	if (index >= 0 && total - sizeof *block >= size) {
		if (cached[index] != NULL) {
			block = cached[index];
			cached[index] = block->next;
			cached_bytes -= total;
			return block;
		}
		void *mapped = mmap(NULL, total, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
		// A process can run out of mappings (Linux's vm.max_map_count) before memory: then this block is
		// malloc's, uncached, as on a target with no mmap.
		if (mapped != MAP_FAILED) {
			block = mapped;
			block->turn = block->mapping = total;
			block->size = total - sizeof *block;
			count_block(true);
			return block;
		}
	}
#endif
	size_t capacity = total - sizeof *block < size ? size : total - sizeof *block;
	block = malloc(sizeof *block + capacity);
	if (block == NULL) { return NULL; }
	block->turn = total;
	block->mapping = 0;
	block->size = capacity;
	count_block(true);
	return block;
}

static void block_free(adamic_region_block *block) {
	count_block(false);
#ifndef REGION_BLOCK_CACHE_OFF
	if (block->mapping != 0) {
		munmap(block, block->mapping);
		return;
	}
#endif
	free(block);
}

// block_give keeps a standard mapped block within the cap and frees the rest. A huge block is kept in
// the one slot when it's empty and the cap holds the block whole, and unmapped otherwise. A malloc
// block (mapping 0, no size of its own) is always freed.
static void block_give(adamic_region_block *block) {
#ifdef REGION_BLOCK_HUGE
	if (block->mapping == BLOCK_HUGE && cached_huge == NULL && cached_bytes + BLOCK_HUGE <= ADAMIC_REGION_BLOCK_CACHE) {
		cached_huge = block;
		cached_bytes += BLOCK_HUGE;
		return;
	}
#endif
#ifndef REGION_BLOCK_CACHE_OFF
	int index = size_index(block->mapping);
	if (index >= 0 && cached_bytes + block->mapping <= ADAMIC_REGION_BLOCK_CACHE) {
		block->next = cached[index];
		cached[index] = block;
		cached_bytes += block->mapping;
		return;
	}
#endif
	block_free(block);
}

size_t adamic_region_cached(void) {
	return cached_bytes;
}

void adamic_region_thread_end(void) {
	for (int index = 0; index < BLOCK_SIZES; index++) {
		while (cached[index] != NULL) {
			adamic_region_block *next = cached[index]->next;
			block_free(cached[index]);
			cached[index] = next;
		}
	}
#ifdef REGION_BLOCK_HUGE
	if (cached_huge != NULL) {
		block_free(cached_huge);
		cached_huge = NULL;
	}
#endif
	cached_bytes = 0;
}

static inline void *allocate_in(adamic_region *region, size_t size) {
	adamic_region_block *block = region->blocks;
	if (block == NULL || block->size - block->used < size) {
		// Each turn twice the last, 4 KB up to 1 MB, header included, and 2 MB from then on where huge
		// blocks are. A block holds at least the object: one bigger than its turn is malloc's, made to fit.
		size_t total = block == NULL ? 4096 : block->turn * 2;
		if (total > BLOCK_LARGEST) {
#ifdef REGION_BLOCK_HUGE
			total = BLOCK_HUGE;
#else
			total = BLOCK_LARGEST;
#endif
		}
		adamic_region_block *fresh = block_take(total, size);
		if (fresh == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		fresh->next = block;
		fresh->used = 0;
		region->blocks = fresh;
		block = fresh;
	}
	void *value = (void *)(block->bytes + block->used);
	block->used += size;
 region->count++;
 ADAMIC_COUNT_ALLOCATION();
 return value;
}


static inline adamic_object *object_new(adamic_region *region, const adamic_shape *shape, bool zero) {
	adamic_shape_check(shape);
	if (region == NULL) {
		return adamic_object_new(shape);
	}
	size_t size = object_size(shape->count);
 adamic_object *object = allocate_in(region, size);
	object->heap.references = 0;
	object->heap.kind = adamic_kind_object;
	object->heap.slab = ADAMIC_REGION_VALUE;
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->tuple = false;
	object->write_order = NULL;
	if (zero) {
		memset(object->slots, 0, shape->count * sizeof object->slots[0]);
		// Incremental constructors may throw or later store outside children.
		region->holds_outside = true;
	}
	memset(adamic_object_initialized(object), 1, shape->count);
	memset(adamic_object_field_types(object), 0, shape->count);
	return object;
}

adamic_object *adamic_object_new_in(adamic_region *region, const adamic_shape *shape) {
	return object_new(region, shape, true);
}

adamic_object *adamic_object_new_filled_in(adamic_region *region, const adamic_shape *shape) {
	return object_new(region, shape, false);
}

#ifdef ADAMIC_CANONICAL_CLOSURES
adamic_environment *adamic_environment_new_in(adamic_region *region, size_t count) {
 size_t size = sizeof(adamic_environment) + count * sizeof(adamic_cell);
 size = (size + ALIGN - 1) & ~(size_t)(ALIGN - 1);
 adamic_environment *environment = allocate_in(region, size);
 adamic_environment_init(environment, (adamic_cell *)(void *)(environment + 1), count);
 environment->heap.slab = ADAMIC_REGION_VALUE;
 region->holds_outside = true;
 return environment;
}
#endif

size_t adamic_region_capacity(const adamic_region *region) {
	size_t total = 0;
	for (const adamic_region_block *block = region->blocks; block != NULL; block = block->next) { total += block->size; }
	return total;
}

bool adamic_region_contains(const adamic_region *region, const void *value) {
	// Integer ranges, not relational comparisons of pointers into unrelated allocations.
	uintptr_t address = (uintptr_t)value;
	for (const adamic_region_block *block = region->blocks; block != NULL; block = block->next) {
		uintptr_t start = (uintptr_t)block->bytes;
		if (address >= start && address - start < block->used) {
			return true;
		}
	}
	return false;
}

void adamic_region_end(adamic_region *region) {
	// No object/slot walk at all when every child is regional or immortal. Otherwise let go while
	// every block still exists: a child can be in another block of the same region.
	if (region->holds_outside) {
		for (adamic_region_block *block = region->blocks; block != NULL; block = block->next) {
			for (size_t offset = 0; offset < block->used;) {
                adamic_heap *value = (adamic_heap *)(void *)(block->bytes + offset);
#ifdef ADAMIC_CANONICAL_CLOSURES
                if (value->kind == adamic_kind_environment) {
                    adamic_environment *environment = (adamic_environment *)value;
                    adamic_environment_end(environment);
                    size_t size = sizeof *environment + environment->count * sizeof(adamic_cell);
                    offset += (size + ALIGN - 1) & ~(size_t)(ALIGN - 1);
                    continue;
                }
#endif
                adamic_object *object = (adamic_object *)value;
                adamic_object_free_children(object, adamic_release);
                offset += object_size(object->shape->count);
			}
		}
	}
	// Weak is independent of strong-child cleanup. The table walks its handles once, rather than
	// asking it about every object in a program that may have no Weak at all.
	adamic_weak_forget_region(region);
	for (adamic_region_block *block = region->blocks; block != NULL;) {
		adamic_region_block *next = block->next;
		block_give(block);
		block = next;
	}
	ADAMIC_COUNT_REGION(region->count);
	region->blocks = NULL;
	region->count = 0;
	region->holds_outside = false;
}
