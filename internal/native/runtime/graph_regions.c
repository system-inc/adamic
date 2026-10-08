// graph_regions.c: count only references crossing a graph region's boundary.
#include "graph_regions.h"
#include "count.h"

#include <stdlib.h>
#include <stdio.h>

typedef struct graph_region {
	struct graph_region *parent;
	struct graph_region *next_record;
	struct graph_region *record_tail;
	size_t size;
	size_t references;
	adamic_heap *head;
	adamic_heap *tail;
	bool shared;
} graph_region;

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

static adamic_heap *member(void *value) {
	adamic_heap *heap = owner(value);
	return adamic_graph_is(heap) ? heap : NULL;
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
	if (heap == NULL || heap != value || heap->references != 1 || adamic_graph_is(heap)) {
		FAIL("compiler bug: graph adoption needs a new allocation");
	}
	bool inline_map = heap->kind == adamic_kind_map && ((adamic_map *)heap)->entries == ((adamic_map *)heap)->small;
	heap = adamic_heap_graph_storage(value, bytes);
	if (inline_map) { ((adamic_map *)heap)->entries = ((adamic_map *)heap)->small; }
	// Interior cells must point at the environment's final address.
	if (heap->kind == adamic_kind_environment) {
		adamic_environment *environment = (adamic_environment *)heap;
		for (size_t i = 0; i < environment->count; i++) { environment->cells[i].owner = environment; }
	}
	return heap;
}

static graph_region *region_of(adamic_heap *heap) {
	graph_region *region = adamic_graph_header_of(heap)->region;
	return region == NULL ? NULL : find(region);
}

static graph_region *new_region(adamic_heap *heap) {
	graph_region *region = calloc(1, sizeof *region);
	if (region == NULL) { FAIL("out of memory"); }
	*region = (graph_region){.parent = region, .record_tail = region,
		.size = 1, .references = heap->references, .head = heap, .tail = heap};
	adamic_graph_header_of(heap)->region = region;
#ifdef ADAMIC_COUNT
	adamic_counted.graph_regions++;
#endif
	return region;
}

void adamic_graph_merge(void *left, void *right) {
	adamic_heap *one = member(left), *two = member(right);
	if (one == NULL || two == NULL) { FAIL("compiler bug: merging non-graph values"); }
	graph_region *a = region_of(one), *b = region_of(two);
	if (one == two || (a != NULL && a == b)) { return; }
	if ((a != NULL && a->shared) || (b != NULL && b->shared)) {
		FAIL("merging shared graph regions is not yet supported");
	}
	if (a == NULL && b != NULL) {
		adamic_heap *swap = one; one = two; two = swap;
		a = b; b = NULL;
	}
	if (a == NULL) { a = new_region(one); }
	if (b == NULL) {
		if (SIZE_MAX - a->references < two->references || a->size == SIZE_MAX) { FAIL("graph region count overflow"); }
		a->references += two->references;
		a->size++;
		adamic_graph_header_of(a->tail)->next = two;
		a->tail = two;
		adamic_graph_header_of(two)->region = a;
	} else {
		if (a->size < b->size) { graph_region *swap = a; a = b; b = swap; }
		if (SIZE_MAX - a->references < b->references || SIZE_MAX - a->size < b->size) { FAIL("graph region count overflow"); }
		a->references += b->references;
		a->size += b->size;
		adamic_graph_header_of(a->tail)->next = b->head;
		a->tail = b->tail;
		a->record_tail->next_record = b;
		a->record_tail = b->record_tail;
		b->parent = a;
	}
#ifdef ADAMIC_COUNT
	adamic_counted.graph_merges++;
#endif
}

void adamic_graph_mark_shared(void *value) {
	adamic_heap *heap = member(value);
	if (heap == NULL) { FAIL("compiler bug: sharing a non-graph value"); }
	graph_region *root = region_of(heap);
	if (root == NULL) { root = new_region(heap); }
	root->shared = true;
}

void adamic_graph_retain(void *value) {
	adamic_heap *heap = member(value);
	graph_region *root = region_of(heap);
	if (heap->references == SIZE_MAX || (root != NULL && root->references == SIZE_MAX)) { FAIL("graph region count overflow"); }
	heap->references++;
	if (root != NULL) { root->references++; }
}

void *adamic_graph_hold(void *holder, void *value) {
	if (value == NULL) { return member(holder) == NULL ? adamic_retain(NULL) : NULL; }
	if (member(holder) != NULL && member(value) != NULL) {
		adamic_graph_merge(holder, value);
		return value;
	}
	return adamic_retain(value);
}

void adamic_graph_drop(void *holder, void *value) {
	if (value == NULL) { return; }
	adamic_heap *from = member(holder), *to = member(value);
	if (from != NULL && to != NULL) {
		if (from != to && (region_of(from) == NULL || region_of(from) != region_of(to))) { FAIL("compiler bug: graph edge crosses regions"); }
		return;
	}
	adamic_release(value);
}

// Enumerate immediate owned slots for initialization and counting diagnostics.
// This does not follow the graph or decide what is freed.
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

static adamic_heap *joining;
static void join_owned(void *value) {
	if (member(value) != NULL) {
		adamic_graph_merge(joining, value);
		adamic_release(value);
	}
}

void *adamic_graph_adopt_owned(void *value, size_t bytes) {
	if (adamic_graph_is(value)) { return value; }
	joining = adamic_graph_adopt(value, bytes);
	children(joining, join_owned);
	void *result = joining;
	joining = NULL;
	return result;
}

void *adamic_graph_escape(void *holder, void *value) {
	return member(holder) != NULL && member(value) != NULL ? adamic_retain(value) : value;
}

void *adamic_graph_take(void *holder, void *value) {
	if (member(holder) != NULL && member(value) != NULL) {
		adamic_graph_merge(holder, value);
		adamic_release(value);
	}
	return value;
}

bool adamic_graph_counted(const void *value) {
	return !adamic_graph_is(value);
}

#ifdef ADAMIC_COUNT
static size_t bytes(adamic_heap *heap) {
	switch (heap->kind) {
	case adamic_kind_object: return sizeof(adamic_object) + ((adamic_object *)heap)->shape->count * (sizeof(adamic_value) + 2);
	case adamic_kind_array: return sizeof(adamic_array) + ((adamic_array *)heap)->capacity * sizeof(adamic_value);
	case adamic_kind_map: {
		adamic_map *map = (adamic_map *)heap;
		return sizeof *map + map->capacity * sizeof(adamic_map_entry) + map->bucket_count * sizeof(size_t);
	}
	case adamic_kind_closure: return sizeof(adamic_closure) + ((adamic_closure *)heap)->count * sizeof(adamic_cell *);
	case adamic_kind_cell: return sizeof(adamic_cell);
	case adamic_kind_environment: return sizeof(adamic_environment) + ((adamic_environment *)heap)->count * sizeof(adamic_cell);
	case adamic_kind_map_iterator: return sizeof(adamic_map_iterator);
	default: FAIL("compiler bug: unsupported graph allocation"); return 0;
	}
}

// The outside count remains in each member's existing header. Its high bit is
// borrowed as a diagnostic mark only during this pass, then restored.
#define MARK_BIT ((size_t)1 << (sizeof(size_t) * 8 - 1))
static adamic_heap **mark_work;
static size_t mark_count;
static graph_region *mark_root;
static void mark(void *value) {
	adamic_heap *heap = member(value);
	if (heap == NULL) { return; }
	if (region_of(heap) != mark_root) { FAIL("compiler bug: graph edge crosses regions"); }
	if ((heap->references & MARK_BIT) != 0) { return; }
	heap->references |= MARK_BIT;
	mark_work[mark_count++] = heap;
}

static void report(graph_region *root, adamic_heap *lone) {
	size_t count = root == NULL ? 1 : root->size;
	mark_root = root;
	mark_count = 0;
	mark_work = malloc(count * sizeof *mark_work);
	if (mark_work == NULL) { FAIL("out of memory"); }
	adamic_heap *head = root == NULL ? lone : root->head;
	for (adamic_heap *each = head; each != NULL; each = adamic_graph_header_of(each)->next) {
		if (each->references != 0) { mark(each); }
	}
	while (mark_count != 0) { children(mark_work[--mark_count], mark); }
	size_t total = 0, reachable = 0, reachable_bytes = 0, records = 0;
	for (adamic_heap *each = head; each != NULL; each = adamic_graph_header_of(each)->next) {
		size_t size = bytes(each);
		total += size;
		if ((each->references & MARK_BIT) != 0) { reachable++; reachable_bytes += size; }
		each->references &= ~MARK_BIT;
	}
	for (graph_region *each = root; each != NULL; each = each->next_record) { records++; }
	fprintf(stderr, "adamic: graph region: live %zu bytes %zu reachable %zu bytes %zu unreachable %zu bytes %zu metadata %zu\n",
		count, total, reachable, reachable_bytes, count - reachable, total - reachable_bytes,
		count * sizeof(adamic_graph_header) + records * sizeof(graph_region));
	free(mark_work);
	mark_work = NULL;
	mark_root = NULL;
}
#endif

bool adamic_graph_release_last(void *value) {
	adamic_heap *heap = member(value);
	graph_region *root = region_of(heap);
	if (heap->references == 0 || (root != NULL && root->references == 0)) { FAIL("compiler bug: graph released without ownership"); }
#ifdef ADAMIC_COUNT
	if ((root == NULL && heap->references == 1) || (root != NULL && root->references == 1)) { report(root, heap); }
#endif
	heap->references--;
	return root == NULL ? heap->references == 0 : --root->references == 0;
}

static void (*outside_release)(void *);
static graph_region *freeing_root;
static adamic_heap *freeing_lone;
static void release_outside(void *value) {
	adamic_heap *heap = member(value);
	if (heap != NULL) {
		if (heap != freeing_lone && (freeing_root == NULL || region_of(heap) != freeing_root)) { FAIL("compiler bug: graph edge crosses regions"); }
		return;
	}
	outside_release(value);
}

void adamic_graph_free(void *value, void (*release)(void *)) {
	adamic_heap *heap = member(value);
	graph_region *root = region_of(heap);
	if ((root != NULL && root->references != 0) || (root == NULL && heap->references != 0)) { FAIL("compiler bug: freeing a graph still owned outside"); }
	freeing_root = root;
	freeing_lone = root == NULL ? heap : NULL;
	outside_release = release;
	adamic_heap *head = root == NULL ? heap : root->head;
	for (adamic_heap *each = head; each != NULL; each = adamic_graph_header_of(each)->next) {
		adamic_weak_forget(each);
		if (each->kind == adamic_kind_environment) {
			adamic_environment *environment = (adamic_environment *)each;
			for (size_t i = 0; i < environment->count; i++) { adamic_weak_forget(&environment->cells[i]); }
		}
	}
	for (adamic_heap *each = head; each != NULL; each = adamic_graph_header_of(each)->next) {
		adamic_heap_free_children(each, release_outside);
	}
	freeing_root = NULL;
	freeing_lone = NULL;
	outside_release = NULL;
	while (head != NULL) {
		adamic_heap *next = adamic_graph_header_of(head)->next;
		adamic_heap_free_storage(head, head->slab);
		head = next;
	}
	while (root != NULL) {
		graph_region *next = root->next_record;
		free(root);
		root = next;
	}
}
