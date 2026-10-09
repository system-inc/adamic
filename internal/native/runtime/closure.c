// closure.c: closures and the cells they capture.

#include "adamic.h"
#include "async.h"
#include "graph_regions.h"

adamic_cell *adamic_cell_new(adamic_value value, bool references) {
	adamic_cell *cell = adamic_allocate(sizeof *cell, adamic_kind_cell);
	cell->references = references;
	cell->ready = true;
	cell->owner = NULL;
	cell->value = value;
	return cell;
}

adamic_closure *(adamic_closure_new)(adamic_code code, size_t count) {
	adamic_closure *closure = adamic_allocate(sizeof *closure + count * sizeof closure->cells[0], adamic_kind_closure);
	closure->code = code;
#ifdef ADAMIC_CLOSURE_CONVENTION
 closure->counted = false;
#endif
#ifdef ADAMIC_CLOSURE_RECEIVERS
 closure->receiver = false;
#endif
	closure->count = count;
#ifdef ADAMIC_CANONICAL_CLOSURES
	closure->canonical_owner = NULL;
	closure->canonical_previous = NULL;
	closure->canonical_next = NULL;
#endif
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = NULL;
	}
	return closure;
}

adamic_environment *adamic_environment_new(size_t count) {
	adamic_environment *environment = adamic_allocate(sizeof *environment + count * sizeof environment->cells[0], adamic_kind_environment);
	environment->count = count;
	environment->functions = NULL;
	adamic_environment_initialize_cells(&environment->heap, environment->cells, count);
	return environment;
}

void adamic_environment_initialize_cells(adamic_heap *owner, adamic_cell *cells, size_t count) {
 for (size_t index = 0; index < count; index++) {
  cells[index] = (adamic_cell){.heap = {.kind = adamic_kind_cell}, .owner = owner};
 }
}
void adamic_environment_drop_cells(adamic_cell *cells, size_t count, void (*drop)(void *)) {
 for (size_t index = 0; index < count; index++) {
  if (cells[index].references) drop(cells[index].value.reference);
 }
}

#ifdef ADAMIC_CANONICAL_CLOSURES
// canonical_functions is the cache of canonical closures over one activation's cells: an
// environment's, or an async frame's when the cells are the async function's own locals.
static adamic_closure **canonical_functions(adamic_heap *owner) {
	switch (owner->kind) {
	case adamic_kind_environment: return &((adamic_environment *)owner)->functions;
	case adamic_kind_async_frame: return &((adamic_async_frame *)owner)->functions;
	default: {
		static const char message[] = "compiler bug: a canonical function's cell has no environment or async frame";
		adamic_panic(message, sizeof message - 1);
	}
	}
}

static adamic_closure *canonical_insert(adamic_heap *owner, adamic_closure *closure, size_t count, adamic_cell *const cells[], bool graph) {
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = adamic_retain(cells[index]);
	}
	// Adoption moves the allocation and joins its owned graph edges. Publish only
	// the final address in the weak canonical cache. Cache hits never adopt again.
	if (graph) {
		closure = adamic_graph_adopt_owned(closure, sizeof *closure + count * sizeof closure->cells[0]);
	}
	adamic_closure **functions = canonical_functions(owner);
	closure->canonical_owner = owner;
	closure->canonical_next = *functions;
	if (*functions != NULL) {
		(*functions)->canonical_previous = closure;
	}
	*functions = closure;
	return closure;
}

// A live value is unique for its code and activation. If no owner kept it, a
// later reference can recreate it: no surviving reference can observe that gap.
// Captured cells keep the frame alive; its cache never keeps a closure alive.
static adamic_closure *canonical(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[], bool graph) {
	adamic_heap *owner = identity->owner;
	for (adamic_closure *held = *canonical_functions(owner); held != NULL; held = held->canonical_next) {
#ifdef ADAMIC_CLOSURE_CONVENTION
		if (held->counted) { continue; }
#endif
		if (held->code == code) {
			return adamic_retain(held);
		}
	}
	return canonical_insert(owner, adamic_closure_new(code, count), count, cells, graph);
}

adamic_closure *(adamic_closure_canonical)(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[]) {
	return canonical(identity, code, count, cells, false);
}

adamic_closure *(adamic_closure_canonical_graph)(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[]) {
	return canonical(identity, code, count, cells, true);
}

#ifdef ADAMIC_CLOSURE_CONVENTION
static adamic_closure *counted_canonical(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[], bool graph) {
	adamic_heap *owner = identity->owner;
	for (adamic_closure *held = *canonical_functions(owner); held != NULL; held = held->canonical_next) {
		if (held->counted && held->counted_code == code) { return adamic_retain(held); }
	}
	return canonical_insert(owner, adamic_counted_closure_new(code, count), count, cells, graph);
}
adamic_closure *(adamic_counted_closure_canonical)(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[]) {
	return counted_canonical(identity, code, count, cells, false);
}

adamic_closure *(adamic_counted_closure_canonical_graph)(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[]) {
	return counted_canonical(identity, code, count, cells, true);
}

#endif

void adamic_closure_uncache(adamic_closure *closure) {
	if (closure->canonical_owner == NULL) {
		return;
	}
	if (closure->canonical_previous != NULL) {
		closure->canonical_previous->canonical_next = closure->canonical_next;
	} else {
		*canonical_functions(closure->canonical_owner) = closure->canonical_next;
	}
	if (closure->canonical_next != NULL) {
		closure->canonical_next->canonical_previous = closure->canonical_previous;
	}
}
#endif
