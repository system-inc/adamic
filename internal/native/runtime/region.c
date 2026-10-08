// region.c: regions, the arenas of docs/memory.md. A region lives for one statement and holds the
// objects that statement makes and nothing reaches after it: they're bump-allocated from the
// region's blocks and let go of together when the statement ends, instead of one free at a time.
//
// An object in a region is immortal while the region lives (its count is 0, as a constant's is), so
// retain and release pass over it. When the region ends, it releases what its objects hold outside
// it (an immortal child is passed over), tells any Weak that pointed at one that it's gone, and frees
// its blocks. The blocks are ordinary malloc blocks, so a use after the region ended (an escape the
// compiler missed) is a use after free to AddressSanitizer.

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
	size_t size = sizeof(adamic_object) + count * sizeof(adamic_value);
	return (size + ALIGN - 1) & ~(size_t)(ALIGN - 1);
}

// Keep allocation in each entry point: an out-of-line helper adds a call per object.
static inline adamic_object *object_new(adamic_region *region, const adamic_shape *shape, bool zero) {
	if (region == NULL) {
		return adamic_object_new(shape);
	}
	size_t size = object_size(shape->count);
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
		adamic_region_block *fresh = malloc(sizeof *fresh + capacity);
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
	adamic_object *object = (adamic_object *)(void *)(block->bytes + block->used);
	block->used += size;
	object->heap.references = 0;
	object->heap.kind = adamic_kind_object;
	object->heap.slab = ADAMIC_REGION_VALUE;
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	if (zero) {
		memset(object->slots, 0, shape->count * sizeof object->slots[0]);
		// An incremental initializer can throw or write heap children later. Until proven otherwise
		// the region must walk it, including the still-zero slots on its exceptional exit.
		region->holds_outside = true;
	}
	region->count++;
	ADAMIC_COUNT_ALLOCATION();
	return object;
}

adamic_object *adamic_object_new_in(adamic_region *region, const adamic_shape *shape) {
	return object_new(region, shape, true);
}

adamic_object *adamic_object_new_filled_in(adamic_region *region, const adamic_shape *shape) {
	return object_new(region, shape, false);
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
				adamic_object *object = (adamic_object *)(void *)(block->bytes + offset);
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
		free(block);
		block = next;
	}
	ADAMIC_COUNT_REGION(region->count);
	region->blocks = NULL;
	region->count = 0;
	region->holds_outside = false;
}
