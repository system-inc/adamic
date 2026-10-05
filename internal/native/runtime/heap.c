// heap.c: allocation, counting and freeing for every heap value (docs/memory.md).

#include "adamic.h"
#include "count.h"

#include <stdint.h>
#include <stdlib.h>

// Small heap values come from size classes, 16 bytes apart up to 256: each class keeps a list of its
// free slots, threaded through them, and takes a chunk of slots at a time from malloc when the list is
// empty. Allocating is taking the list's first slot and freeing is putting a slot back on its class's
// list, a few instructions each; a value's class is in its header, since a string's length can shrink
// after it's made. A larger value comes from malloc. Chunks are kept for the program's life, so its
// memory is its peak, as a free list's always is.
//
// Under the address sanitizer every value comes from malloc and goes back to free, so ASan sees each
// use after a free and LeakSanitizer each value never freed, as the oracle needs. ADAMIC_SLABS turns
// the classes on there too, to test them: a free slot is then poisoned, so a use after a free is
// still caught, though a leak isn't, since a chunk stays reachable.
#if defined(__has_feature)
#if __has_feature(address_sanitizer)
#define SANITIZED 1
#endif
#endif
#if defined(__SANITIZE_ADDRESS__)
#define SANITIZED 1
#endif
#ifndef SANITIZED
#define SANITIZED 0
#endif

#if SANITIZED
#include <sanitizer/asan_interface.h>
#define POISON(slot, size) ASAN_POISON_MEMORY_REGION((slot), (size))
#define UNPOISON(slot, size) ASAN_UNPOISON_MEMORY_REGION((slot), (size))
#else
#define POISON(slot, size) ((void)(slot), (void)(size))
#define UNPOISON(slot, size) ((void)(slot), (void)(size))
#endif

#if SANITIZED && !defined(ADAMIC_SLABS)
#define SLABS 0
#else
#define SLABS 1
#endif

#define GRANULE 16
#define CLASSES 16
#define CHUNK 65536

// A free slot holds the next free slot of its class.
typedef struct free_slot {
	struct free_slot *next;
} free_slot;

static free_slot *free_slots[CLASSES];

// refill gives a class a chunk's worth of free slots.
static void refill(size_t class) {
	size_t size = (class + 1) * GRANULE;
	char *chunk = malloc(CHUNK);
	if (chunk == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t offset = CHUNK / size * size; offset >= size; offset -= size) {
		free_slot *slot = (free_slot *)(chunk + offset - size);
		slot->next = free_slots[class];
		free_slots[class] = slot;
		POISON(slot, size);
	}
}

void *adamic_allocate(size_t size, enum adamic_kind kind) {
	adamic_heap *heap;
	uint32_t slab = 0;
	if (SLABS && size <= CLASSES * GRANULE) {
		size_t class = (size - 1) / GRANULE;
		if (free_slots[class] == NULL) {
			refill(class);
		}
		free_slot *slot = free_slots[class];
		UNPOISON(slot, (class + 1) * GRANULE);
		free_slots[class] = slot->next;
		heap = (adamic_heap *)slot;
		slab = (uint32_t)class + 1;
	} else {
		heap = malloc(size);
		if (heap == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
	}
	heap->references = 1;
	heap->kind = kind;
	heap->slab = slab;
	ADAMIC_COUNT_ALLOCATION();
	return heap;
}

// deallocate gives a value's memory back: to its class's free list, or to free.
static void deallocate(adamic_heap *heap) {
	if (heap->slab == 0) {
		free(heap);
		return;
	}
	size_t class = heap->slab - 1;
	free_slot *slot = (free_slot *)heap;
	slot->next = free_slots[class];
	free_slots[class] = slot;
	// The whole slot, the link too, so a use after the free is caught where there's a sanitizer, even a
	// retain or release, which touches the first word; allocate unpoisons a slot before reading its link.
	POISON(slot, (class + 1) * GRANULE);
}

void *adamic_retain(void *value) {
	ADAMIC_COUNT_RETAIN();
	adamic_heap *heap = value;
	if (heap != NULL && heap->references != 0) {
		heap->references++;
	}
	return value;
}

// Freeing works from a list, never by recursion, so letting go of a chain a million long can't
// overflow the stack: a value whose count reaches zero is listed, and while the list has anything on
// it, the next one is taken, its children are let go (perhaps listing them), and it's freed.
static void **freeing;
static size_t freeing_count;
static size_t freeing_capacity;
static bool draining;

static void list(void *value) {
	if (freeing_count == freeing_capacity) {
		size_t capacity = freeing_capacity == 0 ? 64 : freeing_capacity * 2;
		void **grown = realloc(freeing, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		freeing = grown;
		freeing_capacity = capacity;
	}
	freeing[freeing_count++] = value;
}

// let_go drops one reference and lists the value if that was its last.
static void let_go(void *value) {
	adamic_heap *heap = value;
	if (heap != NULL && heap->references != 0 && --heap->references == 0) {
		list(value);
	}
}

static void free_one(void *value) {
	adamic_heap *heap = value;
	switch (heap->kind) {
	case adamic_kind_string:
		adamic_string_free_index(value);
		break;
	case adamic_kind_object: {
		adamic_object *object = value;
		for (size_t index = 0; index < object->shape->count; index++) {
			if (object->shape->references[index]) {
				let_go(object->slots[index].reference);
			}
		}
		break;
	}
	case adamic_kind_array: {
		adamic_array *array = value;
		if (array->references) {
			for (size_t index = 0; index < array->length; index++) {
				let_go(array->elements[index].reference);
			}
		}
		free(array->elements);
		break;
	}
	case adamic_kind_map:
		adamic_map_free_children(value, let_go);
		break;
	case adamic_kind_cell: {
		adamic_cell *cell = value;
		if (cell->references) {
			let_go(cell->value.reference);
		}
		break;
	}
	case adamic_kind_closure: {
		adamic_closure *closure = value;
		for (size_t index = 0; index < closure->count; index++) {
			let_go(closure->cells[index]);
		}
		break;
	}
	case adamic_kind_number:
	case adamic_kind_boolean:
		break;
	case adamic_kind_weak:
		adamic_weak_dropped(value);
		break;
	case adamic_kind_map_iterator: {
		// The iteration is over: the map may compact again.
		adamic_map_iterator *iterator = value;
		iterator->map->iterating--;
		let_go(iterator->map);
		break;
	}
	}
	// Anything weak that pointed here now points at nothing, before the memory can be anything else.
	adamic_weak_forget(value);
	deallocate(value);
	ADAMIC_COUNT_FREE();
}

void adamic_release(void *value) {
	ADAMIC_COUNT_RELEASE();
	let_go(value);
	if (draining) {
		// An outer release is already working through the list.
		return;
	}
	draining = true;
	while (freeing_count > 0) {
		free_one(freeing[--freeing_count]);
	}
	draining = false;
}
