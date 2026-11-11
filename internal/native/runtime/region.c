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
// region it opens takes them back before asking malloc (#yecmp1d, Adamic's ruling): malloc handed out
// pages the kernel faulted in again for every region, a third of bench/trees' time in the kernel. A
// block is taken and given on one thread, since a region is a stack value its own function opens and
// ends, and a parallelMap task, a closure, is never handed its caller's region (only a direct call's
// region version takes one). The cache is off under AddressSanitizer, so a use after a region ends is
// still a use after free there.

#include "adamic.h"
#include "count.h"

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

struct adamic_region_block {
	struct adamic_region_block *next;
	size_t size;
	size_t used;
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
// read from the environment, so a program behaves the same whatever shell runs it. 1 MB took 93% of
// trees' minor faults and 37% of its wall on Home; 16 MB bought 6 more points for 15 MB more per thread
// (Runtime Performance, #yecmp1d).
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

// One list per standard block size, 4 KB doubling to 1 MB. A block of another size, one made for an
// object bigger than its turn's size, goes back to malloc.
#define BLOCK_SIZES 9
static _Thread_local adamic_region_block *cached[BLOCK_SIZES];
static _Thread_local size_t cached_bytes;

#ifndef REGION_BLOCK_CACHE_OFF
static int size_index(size_t capacity) {
	for (int index = 0; index < BLOCK_SIZES; index++) {
		if (capacity == ((size_t)4096 << index)) { return index; }
	}
	return -1;
}
#endif

static adamic_region_block *block_take(size_t capacity) {
#ifndef REGION_BLOCK_CACHE_OFF
	int index = size_index(capacity);
	if (index >= 0 && cached[index] != NULL) {
		adamic_region_block *block = cached[index];
		cached[index] = block->next;
		cached_bytes -= capacity;
		return block;
	}
#endif
	adamic_region_block *block = malloc(sizeof *block + capacity);
#ifdef ADAMIC_COUNT
	if (block != NULL) { atomic_fetch_add_explicit(&adamic_counted.region_blocks, 1, memory_order_relaxed); }
#endif
	return block;
}

static void block_free(adamic_region_block *block) {
#ifdef ADAMIC_COUNT
	atomic_fetch_add_explicit(&adamic_counted.region_blocks_freed, 1, memory_order_relaxed);
#endif
	free(block);
}

static void block_give(adamic_region_block *block) {
#ifndef REGION_BLOCK_CACHE_OFF
	int index = size_index(block->size);
	if (index >= 0 && cached_bytes + block->size <= ADAMIC_REGION_BLOCK_CACHE) {
		block->next = cached[index];
		cached[index] = block;
		cached_bytes += block->size;
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
	cached_bytes = 0;
}

static inline void *allocate_in(adamic_region *region, size_t size) {
	adamic_region_block *block = region->blocks;
	if (block == NULL || block->size - block->used < size) {
		// Each block twice the last, from 4 KB up to 1 MB, and at least as big as the object.
		size_t capacity = block == NULL ? 4096 : block->size * 2;
		if (capacity > 1048576) {
			capacity = 1048576;
		}
		if (capacity < size) {
			capacity = size;
		}
		adamic_region_block *fresh = block_take(capacity);
		if (fresh == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		fresh->next = block;
		fresh->size = capacity;
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
