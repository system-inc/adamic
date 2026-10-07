// graph_regions.c: count only references crossing a graph region's boundary.
#include "graph_regions.h"
#include "count.h"

#include <stdlib.h>
#include <stdio.h>

typedef struct graph_member graph_member;
typedef struct graph_region {
	struct graph_region *parent;
	size_t size;
	size_t references;
	graph_member *head;
	graph_member *tail;
	bool shared;
} graph_region;

struct graph_member {
	// Each singleton's record survives union: no member pointer needs rewriting.
	graph_region region;
	graph_member *next;
	adamic_heap *value;
	size_t bytes;
	uint32_t slab;
#ifdef ADAMIC_COUNT
	size_t outside;
	bool marked;
	graph_member *work;
#endif
};

// As for Weak's table, hidden pointers must not make a leaked graph reachable
// to LeakSanitizer. The table supplies identity, never ownership.
#define HIDDEN ((uintptr_t)0xa5a5a5a5a5a5a5a0u)
static uintptr_t *members;
static size_t capacity;
static size_t used;
static size_t live;
static uint32_t vacant;

static void fail(const char *message, size_t size) {
	adamic_panic(message, size);
}
#define FAIL(message) fail(message, sizeof(message) - 1)

static adamic_heap *owner(void *value) {
	adamic_heap *heap = value;
	if (heap != NULL && heap->kind == adamic_kind_cell && ((adamic_cell *)heap)->owner != NULL) {
		heap = &((adamic_cell *)heap)->owner->heap;
	}
	return heap;
}

static graph_member *member(void *value) {
	adamic_heap *heap = owner(value);
	if (!adamic_graph_is(heap)) { return NULL; }
	return (graph_member *)(members[heap->slab - 1] ^ HIDDEN);
}

static graph_region *find(graph_region *region) {
	graph_region *root = region;
	while (root->parent != root) { root = root->parent; }
	while (region->parent != region) {
		graph_region *next = region->parent;
		region->parent = root;
		region = next;
	}
	return root;
}

void *adamic_graph_adopt(void *value, size_t bytes) {
	adamic_heap *heap = owner(value);
	if (heap == NULL || heap != value || heap->references != 1) {
		FAIL("compiler bug: graph adoption needs a new allocation");
	}
	graph_member *each = calloc(1, sizeof *each);
	if (each == NULL) { FAIL("out of memory"); }
	each->region = (graph_region){&each->region, 1, 1, each, each, false};
	each->value = heap;
	each->bytes = bytes;
	each->slab = heap->slab;
#ifdef ADAMIC_COUNT
	each->outside = 1;
	adamic_counted.graph_regions++;
#endif
	uint32_t slot;
	if (vacant != 0) {
		slot = vacant - 1;
		vacant = (uint32_t)members[slot];
	} else {
		if (used == UINT32_MAX) { FAIL("too many graph members"); }
		if (used == capacity) {
			size_t grown = capacity == 0 ? 64 : capacity * 2;
			uintptr_t *table = realloc(members, grown * sizeof *table);
			if (table == NULL) { FAIL("out of memory"); }
			members = table;
			capacity = grown;
		}
		slot = (uint32_t)used++;
	}
	members[slot] = (uintptr_t)each ^ HIDDEN;
	live++;
	heap->references = SIZE_MAX;
	heap->slab = slot + 1;
	return value;
}

void adamic_graph_merge(void *left, void *right) {
	graph_member *one = member(left), *two = member(right);
	if (one == NULL || two == NULL) { FAIL("compiler bug: merging non-graph values"); }
	graph_region *a = find(&one->region), *b = find(&two->region);
	if (a == b) { return; }
	if (a->shared || b->shared) { FAIL("merging shared graph regions is not yet supported"); }
	if (a->size < b->size) { graph_region *swap = a; a = b; b = swap; }
	if (SIZE_MAX - a->references < b->references || SIZE_MAX - a->size < b->size) {
		FAIL("graph region count overflow");
	}
	a->references += b->references;
	a->size += b->size;
	a->tail->next = b->head;
	a->tail = b->tail;
	b->parent = a;
#ifdef ADAMIC_COUNT
	adamic_counted.graph_merges++;
#endif
}

void adamic_graph_mark_shared(void *value) {
	graph_member *each = member(value);
	if (each == NULL) { FAIL("compiler bug: sharing a non-graph value"); }
	find(&each->region)->shared = true;
}

void adamic_graph_retain(void *value) {
	graph_member *each = member(value);
	graph_region *root = find(&each->region);
	if (root->references == 0 || root->references == SIZE_MAX) { FAIL("graph region count overflow"); }
	root->references++;
#ifdef ADAMIC_COUNT
	each->outside++;
#endif
}

void *adamic_graph_hold(void *holder, void *value) {
	if (value == NULL) { return NULL; }
	if (member(holder) != NULL && member(value) != NULL) {
		adamic_graph_merge(holder, value);
		return value;
	}
	return adamic_retain(value);
}

void adamic_graph_drop(void *holder, void *value) {
	if (value == NULL) { return; }
	graph_member *from = member(holder), *to = member(value);
	if (from != NULL && to != NULL) {
		if (find(&from->region) != find(&to->region)) { FAIL("compiler bug: graph edge crosses regions"); }
		return;
	}
	adamic_release(value);
}

#ifdef ADAMIC_COUNT
// This enumerates current strong edges for diagnostics only. It never frees.
static void children(void *value, void (*visit)(void *)) {
	adamic_heap *heap = value;
	switch (heap->kind) {
	case adamic_kind_object: {
		adamic_object *object = value;
		for (size_t i = 0; i < object->shape->count; i++) {
			if (object->shape->references[i]) { visit(object->slots[i].reference); }
		}
		break;
	}
	case adamic_kind_array: {
		adamic_array *array = value;
		if (array->references) {
			for (size_t i = 0; i < array->length; i++) { visit(array->elements[i].reference); }
		}
		visit(array->properties);
		break;
	}
	case adamic_kind_map: {
		adamic_map *map = value;
		for (size_t i = 0; i < map->used; i++) {
			if (map->entries[i].deleted) { continue; }
			if (map->reference_keys) { visit(map->entries[i].key.reference); }
			if (map->reference_values) { visit(map->entries[i].value.reference); }
		}
		break;
	}
	case adamic_kind_closure: {
		adamic_closure *closure = value;
		for (size_t i = 0; i < closure->count; i++) { visit(closure->cells[i]); }
		break;
	}
	case adamic_kind_cell: {
		adamic_cell *cell = value;
		if (cell->references) { visit(cell->value.reference); }
		break;
	}
	case adamic_kind_environment: {
		adamic_environment *environment = value;
		for (size_t i = 0; i < environment->count; i++) {
			if (environment->cells[i].references) { visit(environment->cells[i].value.reference); }
		}
		break;
	}
	case adamic_kind_map_iterator:
		visit(((adamic_map_iterator *)value)->map);
		break;
	default: break;
	}
}

static size_t bytes(graph_member *each) {
	size_t total = each->bytes;
	if (each->value->kind == adamic_kind_array) {
		total += ((adamic_array *)each->value)->capacity * sizeof(adamic_value);
	} else if (each->value->kind == adamic_kind_map) {
		adamic_map *map = (adamic_map *)each->value;
		total += map->capacity * sizeof(adamic_map_entry) + map->bucket_count * sizeof(size_t);
	}
	return total;
}

static graph_member *mark_work;
static graph_region *mark_root;
static void mark(void *value) {
	graph_member *each = member(value);
	if (each == NULL) { return; }
	if (find(&each->region) != mark_root) { FAIL("compiler bug: graph edge crosses regions"); }
	if (each->marked) { return; }
	each->marked = true;
	each->work = mark_work;
	mark_work = each;
}

static void report(graph_region *root) {
	mark_root = root;
	mark_work = NULL;
	for (graph_member *each = root->head; each != NULL; each = each->next) {
		each->marked = false;
	}
	for (graph_member *each = root->head; each != NULL; each = each->next) {
		if (each->outside != 0) { mark(each->value); }
	}
	while (mark_work != NULL) {
		graph_member *each = mark_work;
		mark_work = each->work;
		children(each->value, mark);
	}
	size_t total = 0, reachable = 0, reachable_bytes = 0;
	for (graph_member *each = root->head; each != NULL; each = each->next) {
		size_t size = bytes(each);
		total += size;
		if (each->marked) { reachable++; reachable_bytes += size; }
	}
	fprintf(stderr, "adamic: graph region: live %zu bytes %zu reachable %zu bytes %zu unreachable %zu bytes %zu metadata %zu\n",
		root->size, total, reachable, reachable_bytes, root->size - reachable, total - reachable_bytes, root->size * sizeof(graph_member));
	mark_root = NULL;
}
#endif

bool adamic_graph_release_last(void *value) {
	graph_member *each = member(value);
	graph_region *root = find(&each->region);
	if (root->references == 0) { FAIL("compiler bug: graph region released without ownership"); }
#ifdef ADAMIC_COUNT
	if (each->outside == 0) { FAIL("compiler bug: graph member released without ownership"); }
	if (root->references == 1) { report(root); }
	each->outside--;
#endif
	return --root->references == 0;
}

static void (*outside_release)(void *);
static graph_region *freeing_root;
static void release_outside(void *value) {
	graph_member *each = member(value);
	if (each != NULL) {
		if (find(&each->region) != freeing_root) { FAIL("compiler bug: graph edge crosses regions"); }
		return;
	}
	outside_release(value);
}

void adamic_graph_free(void *value, void (*release)(void *)) {
	graph_region *root = find(&member(value)->region);
	if (root->references != 0) { FAIL("compiler bug: freeing a graph region still owned outside"); }
	freeing_root = root;
	outside_release = release;
	for (graph_member *each = root->head; each != NULL; each = each->next) {
		adamic_weak_forget(each->value);
		if (each->value->kind == adamic_kind_environment) {
			adamic_environment *environment = (adamic_environment *)each->value;
			for (size_t i = 0; i < environment->count; i++) { adamic_weak_forget(&environment->cells[i]); }
		}
	}
	for (graph_member *each = root->head; each != NULL; each = each->next) {
		adamic_heap_free_children(each->value, release_outside);
	}
	graph_member *each = root->head;
	freeing_root = NULL;
	outside_release = NULL;
	while (each != NULL) {
		graph_member *next = each->next;
		uint32_t slot = each->value->slab;
		adamic_heap_free_storage(each->value, each->slab);
		members[slot - 1] = vacant;
		vacant = slot;
		free(each);
		live--;
		each = next;
	}
	if (live == 0) {
		free(members);
		members = NULL;
		capacity = used = 0;
		vacant = 0;
	}
}
