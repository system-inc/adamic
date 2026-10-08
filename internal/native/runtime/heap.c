// heap.c: allocation, counting and freeing for every heap value (docs/memory.md).

#include "adamic.h"
#include "count.h"

#include <stdint.h>
#include <stdlib.h>

// Small heap values come from size classes, 16 bytes apart up to 256, carved from 64 KB chunks: each
// chunk holds one class's slots, its freed ones on a list threaded through them, and each class keeps
// a list of the chunks it can take a slot from. Allocating is taking a slot and freeing is putting it
// back on its chunk's list, a few instructions each; a value's class is in its header, since a
// string's length can shrink after it's made. A larger value comes from malloc. A chunk that empties
// goes to spares any class can carve, so memory one class used can serve another: a program that
// builds its values at one size, then another, peaks at what it holds at once, not at the sum of
// every size's peak.
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


// A free slot holds the next free slot of its chunk.
typedef struct free_slot {
	struct free_slot *next;
} free_slot;

// A chunk is 64 KB from malloc, numbered in the order chunks are made; a value's header holds its
// chunk's number (adamic_heap.slab), so a slot finds its chunk with one load. A chunk starts with
// this, and its slots of one class follow.
typedef struct chunk {
	// The class's chunks with a slot to give, which this is on when listed.
	struct chunk *previous;
	struct chunk *next;
	bool listed;
	// Slots freed and not yet taken again, and those never taken yet, from fresh to end.
	free_slot *free;
	char *fresh;
	char *end;
	size_t live;
	size_t class;
	uint32_t number;
} chunk;

#define FIRST_SLOT ((sizeof(chunk) + GRANULE - 1) / GRANULE * GRANULE)

static chunk *giving[CLASSES];
static chunk *spares;
static chunk **chunks;
static size_t chunk_count;
static size_t chunk_capacity;

static void list_chunk(chunk *each) {
	each->previous = NULL;
	each->next = giving[each->class];
	if (each->next != NULL) {
		each->next->previous = each;
	}
	giving[each->class] = each;
	each->listed = true;
}

static void unlist_chunk(chunk *each) {
	if (each->previous != NULL) {
		each->previous->next = each->next;
	} else {
		giving[each->class] = each->next;
	}
	if (each->next != NULL) {
		each->next->previous = each->previous;
	}
	each->listed = false;
}

// new_chunk gives a class a chunk: a spare one, whatever class it held, or a new one from malloc.
static chunk *new_chunk(size_t class) {
	chunk *each = spares;
	if (each != NULL) {
		spares = each->next;
	} else {
		if (chunk_count == chunk_capacity) {
			chunk_capacity = chunk_capacity == 0 ? 64 : chunk_capacity * 2;
			chunk **grown = realloc(chunks, chunk_capacity * sizeof *grown);
			if (grown == NULL || chunk_count >= UINT32_MAX) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			chunks = grown;
		}
		each = malloc(CHUNK);
		if (each == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		each->number = (uint32_t)chunk_count;
		chunks[chunk_count++] = each;
	}
	each->free = NULL;
	each->fresh = (char *)each + FIRST_SLOT;
	each->end = (char *)each + CHUNK;
	each->live = 0;
	each->class = class;
	POISON(each->fresh, (size_t)(each->end - each->fresh));
	list_chunk(each);
	return each;
}

static void *take(size_t class, uint32_t *number) {
	size_t size = (class + 1) * GRANULE;
	chunk *each = giving[class];
	if (each == NULL) {
		each = new_chunk(class);
	}
	void *slot;
	if (each->free != NULL) {
		slot = each->free;
		UNPOISON(slot, size);
		each->free = ((free_slot *)slot)->next;
	} else {
		slot = each->fresh;
		each->fresh += size;
		UNPOISON(slot, size);
	}
	each->live++;
	*number = each->number;
	if (each->free == NULL && each->fresh + size > each->end) {
		// Full: off the list until a slot comes back.
		unlist_chunk(each);
	}
	return slot;
}

static void give(void *slot, uint32_t number) {
	chunk *each = chunks[number];
	size_t class = each->class;
	size_t size = (class + 1) * GRANULE;
	((free_slot *)slot)->next = each->free;
	each->free = slot;
	// The whole slot, the link too, so a use after the free is caught where there's a sanitizer, even a
	// retain or release, which touches the first word; take unpoisons a slot before reading its link.
	POISON(slot, size);
	each->live--;
	if (!each->listed) {
		list_chunk(each);
	}
	if (each->live == 0 && !(giving[class] == each && each->next == NULL)) {
		// Empty, and not the only chunk its class has to give from (which stays, so a class that takes
		// and gives back one value at a time doesn't trade a chunk back and forth): to the spares, which
		// any class can carve. Chunks are kept for the program's life, as malloc keeps what it's given
		// back, but shared: a program peaks at the chunks it holds at once, not at every class's peak.
		unlist_chunk(each);
		each->next = spares;
		spares = each;
	}
}

void *adamic_allocate(size_t size, enum adamic_kind kind) {
	adamic_heap *heap;
	uint32_t slab = 0;
	if (SLABS && size <= CLASSES * GRANULE) {
		uint32_t number;
		heap = take((size - 1) / GRANULE, &number);
		slab = number + 1;
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

// deallocate gives a value's memory back: to its chunk, or to free.
static void deallocate(adamic_heap *heap) {
	adamic_object_collection_integrity_forget(heap);
	if (heap->slab == 0) {
		free(heap);
		return;
	}
	give(heap, heap->slab - 1);
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
		// A shared slice lets go of the string whose bytes it reads.
		let_go(((adamic_string *)value)->owner);
		break;
	case adamic_kind_object: {
		adamic_object *object = value;
		adamic_object_free_children(object, let_go);
		break;
	}
	case adamic_kind_array: {
		adamic_array *array = value;
		if (array->references) {
			for (size_t index = 0; index < array->length; index++) {
				let_go(array->elements[index].reference);
			}
		}
		let_go(array->properties);
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
		// Only an unfinished iterator still holds a place in the entries.
		adamic_map_iterator *iterator = value;
		adamic_map_iterator_close(iterator);
		let_go(iterator->map);
		break;
	}
	}
	// Anything weak that pointed here now points at nothing, before the memory can be anything else.
	adamic_weak_forget(value);
	deallocate(value);
	ADAMIC_COUNT_FREE();
}

// Keep destruction out of the common release path: null, immortal and still-shared values need
// no destruction registers or queue access. The list still drains iteratively, including children.
__attribute__((noinline)) static void release_last(void *value) {
	list(value);
	if (draining) {
		return;
	}
	draining = true;
	while (freeing_count > 0) {
		free_one(freeing[--freeing_count]);
	}
	draining = false;
}

void adamic_release(void *value) {
	ADAMIC_COUNT_RELEASE();
	adamic_heap *heap = value;
	if (heap == NULL || heap->references == 0 || --heap->references != 0) {
		return;
	}
	release_last(value);
}
