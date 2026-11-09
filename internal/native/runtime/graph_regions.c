// graph_regions.c: count only references crossing a graph region's boundary.
//
// An edge between two members of one region carries no count; an edge from a member to a member of
// another region carries one, on the target and its region, as any outside holder's does. Adoption
// makes a region of one: a new allocation keeps the counts its member children already hold, so a
// copy of a long-lived graph's array that's dropped frees at its last release instead of joining the
// graph until it dies (#5nrs3q2). Only a store merges regions (adamic_graph_hold, _take).
//
// Two invariants hold after every merge: no counted edge lies inside a region, and the counted edges
// between regions form no cycle, since a cycle of regions holding each other by counted edges never
// reaches zero. Adoption keeps both, since a new allocation's edges run only to older members. A union
// of two regions would close a cycle through any third region on a counted path between them, so
// adamic_graph_merge unites every region on such a path, either way round, and un-counts every edge
// the union brings inside (graph_regions_entries.a: a store joined node and its entries array, and the
// [key, node] pair between them kept a cycle of counted edges until it was merged too).
#include "graph_regions.h"
#include "count.h"

#include <stdlib.h>
#include <stdio.h>
#include <string.h>

typedef struct graph_region {
	struct graph_region *parent;
	struct graph_region *next_record;
	struct graph_region *record_tail;
	size_t size;
	size_t references;
	adamic_heap *head;
	adamic_heap *tail;
	// boundary lists the members holding a counted edge into another region, the ones a merge
	// walks. A member whose every member child has come inside leaves it at the next merge.
	adamic_heap **boundary;
	size_t boundary_count;
	size_t boundary_capacity;
#ifdef ADAMIC_COUNT
	// counted_out is how many counted edges run from this region's members to other regions, kept
	// by every path that makes or ends one, and checked against the members at every merge and free.
	size_t counted_out;
#endif
	bool shared;
	// Scratch for one merge at a time: searched and reaches for its search, merging for its union.
	bool searched;
	bool reaches;
	bool merging;
} graph_region;

static void fail(const char *message, size_t size) {
	adamic_panic(message, size);
}
#define FAIL(message) fail(message, sizeof(message) - 1)

static adamic_heap *owner(void *value) {
	adamic_heap *heap = value;
#ifdef ADAMIC_CANONICAL_CLOSURES
	if (heap != NULL && heap->kind == adamic_kind_cell && ((adamic_cell *)heap)->owner != NULL) {
		heap = ((adamic_cell *)heap)->owner;
	}
#endif
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

// Adoption into a region, graph or Program, moves the allocation and frees the original, so it takes
// only the kinds whose every pointer at or into the allocation it re-aims. adamic_region_unmovable
// names the kind it can't move, or says NULL. The switch names every kind with no default, so a kind
// added to adamic.h fails the build here until someone decides whether it can move.
const char *adamic_region_unmovable(const adamic_heap *heap) {
	switch (heap->kind) {
	case adamic_kind_object:
	case adamic_kind_array:
	case adamic_kind_map:
	case adamic_kind_cell:
	case adamic_kind_map_iterator:
	case adamic_kind_closure:
#ifdef ADAMIC_CANONICAL_CLOSURES
	case adamic_kind_environment:
#endif
		return NULL;
	case adamic_kind_string: return "a string";
	case adamic_kind_number: return "a number box";
	case adamic_kind_boolean: return "a boolean box";
	case adamic_kind_weak: return "a Weak";
	case adamic_kind_typed_array: return "a typed array";
	case adamic_kind_typed_array_iterator: return "a typed array iterator";
	case adamic_kind_async_frame: return "an async frame";
	case adamic_kind_async_promise: return "a promise";
	case adamic_kind_async_reaction: return "an async reaction";
#ifdef ADAMIC_CANONICAL_CLOSURES
	case adamic_kind_generator: return "a generator";
#endif
	}
	return "an unknown kind";
}

// unadoptable is graph adoption's refusal of a movable kind that something still names by address, or
// NULL. Unmovable kinds are adamic_region_unmovable's.
static const char *unadoptable(adamic_heap *heap) {
	// Without canonical closures nothing names a movable kind's address, so there's nothing to check.
	(void)heap;
#ifdef ADAMIC_CANONICAL_CLOSURES
	// Its frame's cache names its address and holds no count on it.
	if (heap->kind == adamic_kind_closure && ((adamic_closure *)heap)->canonical_owner != NULL) { return "compiler bug: a canonical closure can't be adopted into a graph region after it's cached"; }
	if (heap->kind == adamic_kind_environment) {
		adamic_environment *environment = (adamic_environment *)heap;
		if (environment->functions != NULL) { return "compiler bug: an environment can't be adopted into a graph region after it caches a closure"; }
		for (size_t i = 0; i < environment->count; i++) {
			if (adamic_weak_held(&environment->cells[i])) { return "compiler bug: an environment can't be adopted into a graph region while a Weak names one of its cells"; }
		}
	}
#endif
	return NULL;
}

void *adamic_graph_adopt(void *value, size_t bytes) {
	adamic_heap *heap = owner(value);
	if (heap == NULL || heap != value || (heap->slab & (ADAMIC_SHARED_HEADER | ADAMIC_REGION_VALUE)) != 0 || heap->references != 1 || adamic_graph_is(heap)) {
		FAIL("compiler bug: graph adoption needs a new allocation");
	}
	const char *unmovable = adamic_region_unmovable(heap);
	if (unmovable != NULL) {
		char message[96];
		int length = snprintf(message, sizeof message, "compiler bug: %s can't be adopted into a graph region", unmovable);
		fail(message, (size_t)length);
	}
	const char *refused = unadoptable(heap);
	if (refused != NULL) { fail(refused, strlen(refused)); }
	// A Weak names the address the move frees.
	if (adamic_weak_held(heap)) { FAIL("compiler bug: a value can't be adopted into a graph region while a Weak names it"); }
	if (heap->kind == adamic_kind_object && bytes != adamic_object_allocation_size((adamic_object *)heap)) {
		FAIL("compiler bug: graph adoption truncates object storage");
	}
	heap = adamic_heap_graph_storage(value, bytes);
	if (heap->kind == adamic_kind_object) {
		adamic_object *object = (adamic_object *)heap;
		for (size_t i = 0; i < object->shape->count; i++) {
			if (object->shape->references[i] && object->slots[i].reference == value) { object->slots[i].reference = heap; }
		}
	}
#ifdef ADAMIC_CANONICAL_CLOSURES
	// Interior cells must point at the environment's final address.
	if (heap->kind == adamic_kind_environment) {
		adamic_environment *environment = (adamic_environment *)heap;
  environment->cells = (adamic_cell *)(void *)(environment + 1);
		for (size_t i = 0; i < environment->count; i++) { environment->cells[i].owner = &environment->heap; }
	}
#endif
	return heap;
}

static void children(void *value, void (*visit)(void *));

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

static void boundary_add(graph_region *region, adamic_heap *heap) {
	if (region->boundary_count == region->boundary_capacity) {
		size_t capacity = region->boundary_capacity == 0 ? 4 : region->boundary_capacity * 2;
		if (capacity < region->boundary_capacity || capacity > SIZE_MAX / sizeof *region->boundary) { FAIL("graph region count overflow"); }
		adamic_heap **grown = realloc(region->boundary, capacity * sizeof *grown);
		if (grown == NULL) { FAIL("out of memory"); }
		region->boundary = grown;
		region->boundary_capacity = capacity;
	}
	region->boundary[region->boundary_count++] = heap;
}

// Scratch for one merge at a time, per thread, as adoption's and free's are.
static _Thread_local graph_region **merge_list;
static _Thread_local size_t merge_count, merge_capacity;
static void merge_add(graph_region *region) {
	if (merge_count == merge_capacity) {
		size_t capacity = merge_capacity == 0 ? 8 : merge_capacity * 2;
		if (capacity < merge_capacity || capacity > SIZE_MAX / sizeof *merge_list) { FAIL("graph region count overflow"); }
		graph_region **grown = realloc(merge_list, capacity * sizeof *grown);
		if (grown == NULL) { FAIL("out of memory"); }
		merge_list = grown;
		merge_capacity = capacity;
	}
	merge_list[merge_count++] = region;
}

// each_edge visits every counted edge out of a region: each member child of its boundary members
// whose root is another region, or a member with no region.
static _Thread_local graph_region *edge_from;
static _Thread_local void (*edge_visit)(adamic_heap *, graph_region *);
static void edge_child(void *value) {
	adamic_heap *heap = member(value);
	if (heap == NULL) { return; }
	graph_region *root = region_of(heap);
	if (root != edge_from) { edge_visit(heap, root); }
}
static void each_edge(graph_region *region, void (*visit)(adamic_heap *, graph_region *)) {
	graph_region *saved_from = edge_from;
	void (*saved_visit)(adamic_heap *, graph_region *) = edge_visit;
	edge_from = region;
	edge_visit = visit;
	for (size_t i = 0; i < region->boundary_count; i++) { children(region->boundary[i], edge_child); }
	edge_from = saved_from;
	edge_visit = saved_visit;
}

// between finds the regions on a counted path from from to the target (a root, or a member with no
// region) and marks them merging. The search visits regions, never members, and remembers them, so
// each is visited once: first every region from reaches, then, to a fixed point, which of them reach
// the target.
static _Thread_local graph_region **search_list;
static _Thread_local size_t search_count, search_capacity;
static _Thread_local graph_region *search_target_root, *search_at;
static _Thread_local adamic_heap *search_target_lone;
static _Thread_local bool search_changed;
static void search_add(graph_region *region) {
	if (search_count == search_capacity) {
		size_t capacity = search_capacity == 0 ? 8 : search_capacity * 2;
		if (capacity < search_capacity || capacity > SIZE_MAX / sizeof *search_list) { FAIL("graph region count overflow"); }
		graph_region **grown = realloc(search_list, capacity * sizeof *grown);
		if (grown == NULL) { FAIL("out of memory"); }
		search_list = grown;
		search_capacity = capacity;
	}
	region->searched = true;
	search_list[search_count++] = region;
}
static void search_out(adamic_heap *heap, graph_region *root) {
	if ((search_target_root != NULL && root == search_target_root) || (root == NULL && heap == search_target_lone)) {
		search_at->reaches = true;
	} else if (root != NULL && !root->searched) {
		search_add(root);
	}
}
static void search_back(adamic_heap *heap, graph_region *root) {
	(void)heap;
	if (!search_at->reaches && root != NULL && root->searched && root->reaches) {
		search_at->reaches = true;
		search_changed = true;
	}
}
static void between(graph_region *from, graph_region *target_root, adamic_heap *target_lone) {
	if (from == NULL) { return; }
	search_target_root = target_root;
	search_target_lone = target_lone;
	search_count = 0;
	search_add(from);
	for (size_t i = 0; i < search_count; i++) {
		search_at = search_list[i];
		each_edge(search_at, search_out);
	}
	for (search_changed = true; search_changed;) {
		search_changed = false;
		for (size_t i = 0; i < search_count; i++) {
			search_at = search_list[i];
			each_edge(search_at, search_back);
		}
	}
	for (size_t i = 0; i < search_count; i++) {
		graph_region *region = search_list[i];
		if (region != from && region->reaches && !region->merging) {
			region->merging = true;
			merge_add(region);
		}
		region->searched = region->reaches = false;
	}
#ifdef ADAMIC_COUNT
	adamic_counted.graph_searched += search_count;
#endif
	search_count = 0;
	search_target_root = search_at = NULL;
	search_target_lone = NULL;
}

// uncount takes the count off an edge the union brings inside, from the target and its region.
static _Thread_local adamic_heap *merging_one, *merging_two;
static void uncount(adamic_heap *heap, graph_region *root) {
	if ((root != NULL && root->merging) || (root == NULL && (heap == merging_one || heap == merging_two))) {
		if (heap->references == 0 || (root != NULL && root->references == 0)) { FAIL("compiler bug: graph released without ownership"); }
		heap->references--;
		if (root != NULL) { root->references--; }
#ifdef ADAMIC_COUNT
		edge_from->counted_out--;
#endif
	}
}

static _Thread_local graph_region *crossing_root;
static _Thread_local size_t crossing;
static void count_crossing(void *value) {
	adamic_heap *heap = member(value);
	if (heap != NULL && region_of(heap) != crossing_root) { crossing++; }
}

// prune keeps the boundary members that still hold an edge out of the region, and in the counted
// build checks that those edges are exactly the counted ones: a counted edge left inside the region
// would hold it alive forever.
static void prune(graph_region *region) {
	crossing_root = region;
	size_t total = 0, kept = 0;
	for (size_t i = 0; i < region->boundary_count; i++) {
		crossing = 0;
		children(region->boundary[i], count_crossing);
		total += crossing;
		if (crossing != 0) { region->boundary[kept++] = region->boundary[i]; }
	}
	region->boundary_count = kept;
	crossing_root = NULL;
#ifdef ADAMIC_COUNT
	if (total != region->counted_out) { FAIL("compiler bug: a counted edge lies inside a graph region"); }
#else
	(void)total;
#endif
}

// unite joins two regions, or a region and a member with none, and returns the root. A member with
// no region holds no counted edge, since adoption gives one to any member that does.
static graph_region *unite(graph_region *a, adamic_heap *one, graph_region *b, adamic_heap *two) {
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
		for (size_t i = 0; i < b->boundary_count; i++) { boundary_add(a, b->boundary[i]); }
		free(b->boundary);
		b->boundary = NULL;
		b->boundary_count = b->boundary_capacity = 0;
#ifdef ADAMIC_COUNT
		a->counted_out += b->counted_out;
		b->counted_out = 0;
#endif
	}
#ifdef ADAMIC_COUNT
	adamic_counted.graph_merges++;
#endif
	return a;
}

#ifdef ADAMIC_COUNT
// acyclic checks, in the counted build, that no counted path leads from root back to itself: the
// one place a merge could have left a cycle of regions.
static _Thread_local graph_region *acyclic_root;
static void acyclic_out(adamic_heap *heap, graph_region *root) {
	(void)heap;
	if (root == acyclic_root) { FAIL("compiler bug: counted edges between graph regions close a cycle"); }
	if (root != NULL && !root->searched) { search_add(root); }
}
static void acyclic(graph_region *root) {
	acyclic_root = root;
	search_count = 0;
	search_add(root);
	for (size_t i = 0; i < search_count; i++) { each_edge(search_list[i], acyclic_out); }
	for (size_t i = 0; i < search_count; i++) { search_list[i]->searched = false; }
	search_count = 0;
	acyclic_root = NULL;
}
#endif

void adamic_graph_merge(void *left, void *right) {
	adamic_heap *one = member(left), *two = member(right);
	if (one == NULL || two == NULL) { FAIL("compiler bug: merging non-graph values"); }
	graph_region *a = region_of(one), *b = region_of(two);
	if (one == two || (a != NULL && a == b)) { return; }
	if ((a != NULL && a->shared) || (b != NULL && b->shared)) {
		FAIL("merging shared graph regions is not yet supported");
	}
	// Every region on a counted path between the two, either way round, joins them: left out, it would
	// hold the union and be held by it, a cycle of counted edges no region could see.
	merge_count = 0;
	if (a != NULL) { a->merging = true; }
	if (b != NULL) { b->merging = true; }
	between(b, a, a == NULL ? one : NULL);
	between(a, b, b == NULL ? two : NULL);
	for (size_t i = 0; i < merge_count; i++) {
		if (merge_list[i]->shared) { FAIL("merging shared graph regions is not yet supported"); }
	}
	// Then every edge among them is inside: its count goes, read while each root still names its side.
	merging_one = a == NULL ? one : NULL;
	merging_two = b == NULL ? two : NULL;
	if (a != NULL) { each_edge(a, uncount); }
	if (b != NULL) { each_edge(b, uncount); }
	for (size_t i = 0; i < merge_count; i++) { each_edge(merge_list[i], uncount); }
	merging_one = merging_two = NULL;
	if (a != NULL) { a->merging = false; }
	if (b != NULL) { b->merging = false; }
	for (size_t i = 0; i < merge_count; i++) { merge_list[i]->merging = false; }
	graph_region *root = unite(a, one, b, two);
	for (size_t i = 0; i < merge_count; i++) {
		if (find(merge_list[i]) != root) { root = unite(root, NULL, merge_list[i], NULL); }
	}
	merge_count = 0;
	prune(root);
#ifdef ADAMIC_COUNT
	acyclic(root);
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
		graph_region *region = region_of(from);
		if (from == to || (region != NULL && region == region_of(to))) { return; }
		// A counted edge out of the region, which only a member with a region holds.
		if (region == NULL) { FAIL("compiler bug: graph edge crosses regions"); }
#ifdef ADAMIC_COUNT
		region->counted_out--;
#endif
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
#ifdef ADAMIC_CANONICAL_CLOSURES
	case adamic_kind_environment: {
		adamic_environment *environment = value;
		for (size_t i = 0; i < environment->count; i++) {
			if (environment->cells[i].references) { visit(environment->cells[i].value.reference); }
		}
		break;
	}
#endif
	case adamic_kind_map_iterator:
		visit(((adamic_map_iterator *)value)->map);
		break;
	default: break;
	}
}

// Scratch for one adoption, deallocation or report at a time. Graph regions never cross threads
// (share.c refuses them), but every task builds its own, so each thread has its own scratch.
static _Thread_local adamic_heap *joining;
static _Thread_local size_t joined;
static void count_owned(void *value) {
	adamic_heap *heap = member(value);
	if (heap == NULL || heap == joining) { return; }
	graph_region *root = region_of(heap);
	if (root != NULL && root->shared) { FAIL("merging shared graph regions is not yet supported"); }
	joined++;
}

// The new member keeps the count each member child already holds, as a counted edge into that
// child's region, and becomes a region of one on its boundary. It joins nothing until a store does,
// and it can't close a cycle of regions: its edges run only to members older than it, which no
// counted path leads back from. The only unions are hold's and take's, both through
// adamic_graph_merge; Program adoption, relocation (adamic_heap_graph_storage) and share.c, which
// refuses graph values, join no regions.
void *adamic_graph_adopt_owned(void *value, size_t bytes) {
	if (adamic_graph_is(value)) { return value; }
	joining = adamic_graph_adopt(value, bytes);
	joined = 0;
	children(joining, count_owned);
	if (joined != 0) {
		graph_region *region = new_region(joining);
		boundary_add(region, joining);
#ifdef ADAMIC_COUNT
		region->counted_out = joined;
#endif
	}
	void *result = joining;
	joining = NULL;
	return result;
}

// What a slot gives up becomes the statement's: an edge inside the region carries no count, so the
// statement takes one; a counted edge out of the region hands its count over.
void *adamic_graph_escape(void *holder, void *value) {
	adamic_heap *from = member(holder), *to = member(value);
	if (from == NULL || to == NULL) { return value; }
	graph_region *region = region_of(from);
	if (from == to || (region != NULL && region == region_of(to))) { return adamic_retain(value); }
	if (region == NULL) { FAIL("compiler bug: graph edge crosses regions"); }
#ifdef ADAMIC_COUNT
	region->counted_out--;
#endif
	return value;
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
	case adamic_kind_object: return adamic_object_size(((adamic_object *)heap)->shape->count);
	case adamic_kind_array: return sizeof(adamic_array) + ((adamic_array *)heap)->capacity * sizeof(adamic_value);
	case adamic_kind_map: {
		adamic_map *map = (adamic_map *)heap;
		return sizeof *map + map->capacity * sizeof(adamic_map_entry) + map->bucket_count * sizeof(size_t);
	}
	case adamic_kind_closure: return sizeof(adamic_closure) + ((adamic_closure *)heap)->count * sizeof(adamic_cell *);
	case adamic_kind_cell: return sizeof(adamic_cell);
#ifdef ADAMIC_CANONICAL_CLOSURES
	case adamic_kind_environment: return sizeof(adamic_environment) + ((adamic_environment *)heap)->count * sizeof(adamic_cell);
#endif
	case adamic_kind_map_iterator: return sizeof(adamic_map_iterator);
	default: FAIL("compiler bug: unsupported graph allocation"); return 0;
	}
}

// The outside count remains in each member's existing header. Its high bit is
// borrowed as a diagnostic mark only during this pass, then restored.
#define MARK_BIT ((size_t)1 << (sizeof(size_t) * 8 - 1))
static _Thread_local adamic_heap **mark_work;
static _Thread_local size_t mark_count;
static _Thread_local graph_region *mark_root;
static void mark(void *value) {
	adamic_heap *heap = member(value);
	if (heap == NULL) { return; }
	// A counted edge into another region: outside this one's report.
	if (region_of(heap) != mark_root) { return; }
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

static _Thread_local void (*outside_release)(void *);
static _Thread_local graph_region *freeing_root;
static _Thread_local adamic_heap *freeing_lone;
static _Thread_local size_t freeing_crossing;
static void release_outside(void *value) {
	adamic_heap *heap = member(value);
	if (heap != NULL) {
		if (heap == freeing_lone || (freeing_root != NULL && region_of(heap) == freeing_root)) { return; }
		// A counted edge into another region, released like any outside hold. A member with no
		// region holds none.
		if (freeing_root == NULL) { FAIL("compiler bug: graph edge crosses regions"); }
		freeing_crossing++;
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
#ifdef ADAMIC_CANONICAL_CLOSURES
		if (each->kind == adamic_kind_environment) {
			adamic_environment *environment = (adamic_environment *)each;
			for (size_t i = 0; i < environment->count; i++) { adamic_weak_forget(&environment->cells[i]); }
		}
#endif
	}
	freeing_crossing = 0;
	for (adamic_heap *each = head; each != NULL; each = adamic_graph_header_of(each)->next) {
		adamic_heap_free_children(each, release_outside);
	}
#ifdef ADAMIC_COUNT
	if (root != NULL && freeing_crossing != root->counted_out) { FAIL("compiler bug: a counted edge lies inside a graph region"); }
#endif
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
		free(root->boundary);
		free(root);
		root = next;
	}
}
