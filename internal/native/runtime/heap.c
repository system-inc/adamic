// heap.c: allocation, counting and freeing for every heap value (docs/memory.md).

#include "adamic.h"

#include <stdlib.h>

void *adamic_allocate(size_t size, enum adamic_kind kind) {
	adamic_heap *heap = malloc(size);
	if (heap == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	heap->references = 1;
	heap->kind = kind;
	return heap;
}

void *adamic_retain(void *value) {
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
	}
	free(value);
}

void adamic_release(void *value) {
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
