// heap.c: allocation, counting and freeing for every heap value (docs/memory.md).

#ifdef ADAMIC_TARGET_WASI
// Keep the inline path and also supply the request host's ownership export.
#define adamic_release adamic_release_inline
#endif
#include "adamic.h"
#ifdef ADAMIC_TARGET_WASI
#undef adamic_release
void adamic_release(void *value) { adamic_release_inline(value); }
#endif
#include "async.h"
#include "count.h"
#include "graph_regions.h"
#include "slab_quarantine.h"

#include <stdint.h>
#include <stdlib.h>
#include <stdatomic.h>
#include <string.h>

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
// still caught, though a leak isn't, since a chunk stays reachable. The oracle runs every fixture that
// way as well, so a mistake in the classes themselves meets the sanitizers. ADAMIC_MALLOC turns them
// off in a build without sanitizers, for macOS's leaks tool, which for the same reason can't see a
// value leaked into a chunk.
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

#if (SANITIZED && !defined(ADAMIC_SLABS)) || defined(ADAMIC_MALLOC)
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
	size_t owner;
	_Atomic(struct remote_slot *) remote;
	struct chunk *owned_next;
} chunk;

#define FIRST_SLOT ((sizeof(chunk) + GRANULE - 1) / GRANULE * GRANULE)

static _Thread_local chunk *giving[CLASSES];
static _Thread_local chunk *spares;
static _Thread_local chunk *owned_chunks;

#include "heap_parallel.h"

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
		each = malloc(CHUNK);
		if (each == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		each->owner = thread_number();
		atomic_init(&each->remote, NULL);
		each->owned_next = owned_chunks;
		owned_chunks = each;
		register_chunk(each);
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

static void drain_remote(chunk *each);

static void *take(size_t class, uint32_t *number) {
	// Full chunks may have remote frees waiting without being on a giving list.
	if (giving[class] == NULL) {
		for (chunk *each = owned_chunks; each != NULL; each = each->owned_next) {
			drain_remote(each);
		}
	} else {
		drain_remote(giving[class]);
	}
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

static void give_local(void *slot, chunk *each) {
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

// A remote free does not touch any owner-only list or live count. Queue nodes are out of line so
// ASan can poison the entire freed slot, its first word included, exactly as on a local free.
static void give(void *slot, uint32_t number) {
	chunk *each = find_chunk(number);
	if (each->owner == thread_number()) {
		give_local(slot, each);
		return;
	}
	ADAMIC_TSAN_PAUSE(adamic_tsan_remote_free);
	remote_slot *node = malloc(sizeof *node);
	if (node == NULL) { adamic_panic("out of memory", 13); }
	node->slot = slot;
	POISON(slot, (each->class + 1) * GRANULE);
	remote_slot *head = atomic_load_explicit(&each->remote, memory_order_relaxed);
	do { node->next = head; } while (!atomic_compare_exchange_weak_explicit(&each->remote, &head, node, memory_order_release, memory_order_relaxed));
}

static void drain_remote(chunk *each) {
	remote_slot *node = atomic_exchange_explicit(&each->remote, NULL, memory_order_acquire);
	while (node != NULL) {
		remote_slot *next = node->next;
		UNPOISON(node->slot, (each->class + 1) * GRANULE);
		give_local(node->slot, each);
		free(node);
		node = next;
	}
}

static void *allocate_storage(size_t size, enum adamic_kind kind) {
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
	return heap;
}

void *adamic_allocate(size_t size, enum adamic_kind kind) {
	ADAMIC_COUNT_ALLOCATION();
	return allocate_storage(size, kind);
}

// deallocate gives a value's memory back: to its chunk, or to free.
static void deallocate(adamic_heap *heap) {
	uint32_t slab = heap->slab & ~ADAMIC_SHARED_HEADER;
	if (slab == 0) {
		free(heap);
		return;
	}
#if ADAMIC_SLAB_QUARANTINE
	// Sanitized with slabs, a freed slot waits poisoned on this thread before it can be taken again,
	// so a stale pointer can't read the next value made in it as its own; the slot it evicts is given
	// back, to its chunk here or to its owner's remote list.
	void *slot = heap;
	uint32_t number = slab - 1;
	if (!adamic_slab_quarantine(&slot, &number, (find_chunk(number)->class + 1) * GRANULE)) {
		return;
	}
	give(slot, number);
#else
	give(heap, slab - 1);
#endif
}

// Move a still-new allocation into storage with its graph-only prefix. This is
// one logical allocation, and must happen before aliases or owned slots exist.
void *adamic_heap_graph_storage(void *value, size_t size) {
	adamic_heap *old = value;
	adamic_heap *storage = allocate_storage(size + sizeof(adamic_graph_header), old->kind);
	uint32_t slab = storage->slab;
	adamic_graph_header *prefix = (adamic_graph_header *)storage;
	adamic_heap *heap = (adamic_heap *)(prefix + 1);
	memcpy(heap, old, size);
	heap->slab = slab | ADAMIC_GRAPH_FLAG;
	*prefix = (adamic_graph_header){NULL, NULL};
	deallocate(old);
	return heap;
}

// An environment's interior cell has no count of its own; its environment holds it. Such a cell's
// count is zero and a graph object's header has ADAMIC_GRAPH_FLAG, so adamic_retain and
// adamic_release (adamic.h) bring both here and never count them inline.
static adamic_heap *counted_heap(void *value) {
	adamic_heap *heap = value;
	if (heap != NULL && heap->kind == adamic_kind_cell && ((adamic_cell *)heap)->owner != NULL) {
		heap = ((adamic_cell *)heap)->owner;
	}
	return heap;
}

void *adamic_retain_slow(void *value) {
	adamic_heap *heap = counted_heap(value);
	if (adamic_graph_is(heap)) {
		// Graph regions are never shared yet (share.c refuses them), so their counts stay plain.
		adamic_graph_retain(heap);
	} else if (heap != NULL) {
		size_t count = __atomic_load_n(&heap->references, __ATOMIC_RELAXED);
		// Clang's native intptr_t conversion makes shared counts negative. One test covers both
		// sharing and zero, so the ordinary positive-count path keeps one branch and plain stores.
		if ((intptr_t)count > 0) { heap->references = count + 1; }
		else if (count != 0 && count != ADAMIC_SHARED) { ADAMIC_TSAN_PAUSE(adamic_tsan_shared_count); __atomic_fetch_add(&heap->references, 1, __ATOMIC_RELAXED); }
	}
	return value;
}

// Freeing works from a list, never by recursion, so letting go of a chain a million long can't
// overflow the stack: a value whose count reaches zero is listed, and while the list has anything on
// it, the next one is taken, its children are let go (perhaps listing them), and it's freed.
static _Thread_local void **freeing;
static _Thread_local size_t freeing_count;
static _Thread_local size_t freeing_capacity;
static _Thread_local bool draining;

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

// drop_reference lets go of one count and says what to free when it was the last: the value, or for
// an interior cell its environment. It never touches the freeing queue itself.
static adamic_heap *drop_reference(void *value) {
	adamic_heap *heap = counted_heap(value);
	if (heap == NULL) { return NULL; }
	if (adamic_graph_is(heap)) {
		// A member reaching zero does not mean its region is unowned: only the region's last
		// outside release frees it.
		return adamic_graph_release_last(heap) ? heap : NULL;
	}
	size_t count = __atomic_load_n(&heap->references, __ATOMIC_RELAXED);
	if ((intptr_t)count > 0) {
		heap->references = count - 1;
		return count == 1 ? heap : NULL;
	}
	if (count != 0 && count != ADAMIC_SHARED &&
		__atomic_fetch_sub(&heap->references, 1, __ATOMIC_RELEASE) == ADAMIC_SHARED + 1) {
		__atomic_thread_fence(__ATOMIC_ACQUIRE);
		return heap;
	}
	return NULL;
}

// Children are queued for an outer drain, never freed recursively.
static void let_go(void *value) {
	adamic_heap *last = drop_reference(value);
	if (last != NULL) { list(last); }
}

void adamic_heap_free_children(void *value, void (*let_go)(void *)) {
	adamic_heap *heap = value;
	switch (heap->kind) {
	case adamic_kind_async_frame:
	case adamic_kind_async_promise:
	case adamic_kind_async_reaction:
		adamic_async_free_children(value, let_go);
		break;
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
		free(array->json_elements);
		break;
	}
	case adamic_kind_typed_array: {
		adamic_typed_array *array = value;
		if (array->owner != NULL) {
			let_go(array->owner);
		} else {
			free(array->data);
		}
		break;
	}
	case adamic_kind_typed_array_iterator:
		let_go(((adamic_typed_array_iterator *)value)->array);
		break;
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
	case adamic_kind_environment: {
		adamic_environment *environment = value;
		adamic_environment_drop_cells(environment->cells, environment->count, let_go);
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
		if (!iterator->exhausted) {
			iterator->map->iterating--;
		}
		let_go(iterator->map);
		break;
	}
	}
}

void adamic_heap_free_storage(void *value, uint32_t slab) {
	adamic_heap *heap = value;
	if (adamic_graph_is(heap)) {
		heap = (adamic_heap *)adamic_graph_header_of(heap);
	}
	heap->slab = slab & ~ADAMIC_GRAPH_FLAG;
	deallocate(heap);
	ADAMIC_COUNT_FREE();
}

static void free_one(void *value) {
	if (adamic_graph_is(value)) {
		adamic_graph_free(value, let_go);
		return;
	}
	adamic_heap_free_children(value, let_go);
	// Anything weak that pointed here now points at nothing, before the memory can be anything else.
	adamic_weak_forget(value);
	adamic_heap_free_storage(value, ((adamic_heap *)value)->slab);
}

// Keep destruction out of the common release path: null, immortal and still-shared values need
// no destruction registers or queue access. The list still drains iteratively, including children.
__attribute__((noinline)) static void release_last(void *value) {
	list(value);
	if (draining) {
		return;
	}
	draining = true;
	while (freeing_count > 0) { free_one(freeing[--freeing_count]); }
	draining = false;
}

void adamic_release_slow(void *value) {
	adamic_heap *last = drop_reference(value);
	if (last != NULL) { release_last(last); }
}

void adamic_heap_thread_end(void) {
#if ADAMIC_SLAB_QUARANTINE
	// The slots this thread holds go back before its remote frees are drained, so its own come home.
	adamic_slab_quarantine_drain(give);
#endif
	for (chunk *each = owned_chunks; each != NULL; each = each->owned_next) { drain_remote(each); }
	free(freeing);
	freeing = NULL;
	freeing_capacity = 0;
}

// Called only after all workers join. No owner can allocate or publish remote frees now.
void adamic_heap_end(void) {
	adamic_heap_thread_end();
	size_t count = atomic_load_explicit(&chunk_count, memory_order_relaxed);
	for (size_t index = 0; index < count; index++) {
		chunk *each = find_chunk((uint32_t)index);
		remote_slot *node = atomic_load_explicit(&each->remote, memory_order_relaxed);
		while (node != NULL) { remote_slot *next = node->next; free(node); node = next; }
		free(each);
	}
	for (size_t page = 0; page < (count + PAGE_SIZE - 1) / PAGE_SIZE; page++) {
		free(atomic_load_explicit(&chunk_pages[page], memory_order_relaxed));
	}
}

__attribute__((constructor)) static void heap_cleanup_at_exit(void) {
	// Registered before main, so the lazily registered pool join runs before this cleanup.
	atexit(adamic_heap_end);
}
