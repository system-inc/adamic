// closure.c: closures and the cells they capture.

#include "adamic.h"
#include "graph_regions.h"

adamic_cell *adamic_cell_new(adamic_value value, bool references) {
	adamic_cell *cell = adamic_allocate(sizeof *cell, adamic_kind_cell);
	cell->references = references;
	cell->ready = true;
#ifdef ADAMIC_CANONICAL_CLOSURES
	cell->owner = NULL;
#endif
	cell->value = value;
	return cell;
}

adamic_closure *(adamic_closure_new)(adamic_code code, size_t count) {
	adamic_closure *closure = adamic_allocate(sizeof *closure + count * sizeof closure->cells[0], adamic_kind_closure);
	closure->code = code;
	closure->bound_method_storage.counted = false;
	closure->bound_method_storage.code = NULL;
	closure->counted = false;
#ifdef ADAMIC_CLOSURE_RECEIVERS
 closure->receiver = false;
#endif
#ifdef ADAMIC_CLOSURE_DISCARDS
	// The made function's result says otherwise where it's a reference: MakeClosure's emission, and
	// the runtime's own closures that return one (map_set.c).
	closure->discard = adamic_discard_none;
#endif
	closure->count = count;
	closure->source_length = 0;
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

#ifdef ADAMIC_CANONICAL_CLOSURES
adamic_environment *adamic_environment_new(size_t count) {
 adamic_environment *environment = adamic_allocate(sizeof *environment + count * sizeof(adamic_cell), adamic_kind_environment);
 adamic_heap heap = environment->heap;
 adamic_environment_init(environment, (adamic_cell *)(void *)(environment + 1), count);
 environment->heap = heap;
 return environment;
}
void adamic_environment_init(adamic_environment *environment, adamic_cell *cells, size_t count) {
 environment->heap = (adamic_heap){.kind = adamic_kind_environment};
 environment->count = count;
 environment->functions = NULL;
 environment->cells = cells;
 adamic_environment_initialize_cells(&environment->heap, cells, count);
}
void adamic_environment_end(adamic_environment *environment) {
 adamic_environment_drop_cells(environment->cells, environment->count, adamic_release);
}

void adamic_environment_initialize_cells(adamic_heap *owner, adamic_cell *cells, size_t count) {
	for (size_t index = 0; index < count; index++) {
		cells[index] = (adamic_cell){.heap = {.kind = adamic_kind_cell}, .owner = owner};
	}
}

void adamic_environment_drop_cells(adamic_cell *cells, size_t count, void (*drop)(void *)) {
	for (size_t index = 0; index < count; index++) {
		if (cells[index].references) { drop(cells[index].value.reference); }
	}
}

// The cache of canonical closures lives on an environment, so a cell some other frame owns (an
// async frame's, a generator's) would have its words read and written as one. Emission keeps those
// cells out of canonical closures; this says so out loud if it ever doesn't.
static adamic_environment *canonical_owner(const adamic_cell *identity) {
	if (identity->owner == NULL || identity->owner->kind != adamic_kind_environment) {
		adamic_panic("compiler bug: a canonical closure over a cell no environment owns", sizeof "compiler bug: a canonical closure over a cell no environment owns" - 1);
	}
	return (adamic_environment *)identity->owner;
}

// A live value is unique for its code and activation. If no owner kept it, a
// later reference can recreate it: no surviving reference can observe that gap.
// Captured cells keep the frame alive; its cache never keeps a closure alive.
static adamic_closure *canonical(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[], bool graph ADAMIC_CANONICAL_DISCARD) {
	adamic_environment *owner = canonical_owner(identity);
	for (adamic_closure *held = owner->functions; held != NULL; held = held->canonical_next) {
#ifdef ADAMIC_CLOSURE_CONVENTION
        if (held->counted) { continue; }
#endif
		if (held->code == code) {
			return adamic_retain(held);
		}
	}
	adamic_closure *closure = adamic_closure_new(code, count);
#ifdef ADAMIC_CLOSURE_DISCARDS
	closure->discard = discard;
#endif
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = adamic_retain(cells[index]);
	}
	if (graph) { closure = adamic_graph_adopt_owned(closure, sizeof *closure + count * sizeof closure->cells[0]); }
	closure->canonical_owner = owner;
	closure->canonical_next = owner->functions;
	if (owner->functions != NULL) {
		owner->functions->canonical_previous = closure;
	}
	owner->functions = closure;
	return closure;
}


#ifdef ADAMIC_CLOSURE_CONVENTION
static adamic_closure *counted_canonical(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[], bool graph ADAMIC_CANONICAL_DISCARD) {
 adamic_environment *owner = canonical_owner(identity);
 for (adamic_closure *held = owner->functions; held != NULL; held = held->canonical_next) {
 if (held->counted && held->counted_code == code) { return adamic_retain(held); }
 }
 adamic_closure *closure = adamic_counted_closure_new(code, count);
#ifdef ADAMIC_CLOSURE_DISCARDS
 closure->discard = discard;
#endif
 for (size_t index = 0; index < count; index++) { closure->cells[index] = adamic_retain(cells[index]); }
 if (graph) { closure = adamic_graph_adopt_owned(closure, sizeof *closure + count * sizeof closure->cells[0]); }
 closure->canonical_owner = owner;
 closure->canonical_next = owner->functions;
 if (owner->functions != NULL) { owner->functions->canonical_previous = closure; }
 owner->functions = closure;
 return closure;
}
#endif

adamic_closure *(adamic_closure_canonical)(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[] ADAMIC_CANONICAL_DISCARD) {
 return canonical(identity, code, count, cells, false
#ifdef ADAMIC_CLOSURE_DISCARDS
 , discard
#endif
 );
}
adamic_closure *(adamic_closure_canonical_graph)(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[] ADAMIC_CANONICAL_DISCARD) {
 return canonical(identity, code, count, cells, true
#ifdef ADAMIC_CLOSURE_DISCARDS
 , discard
#endif
 );
}
#ifdef ADAMIC_CLOSURE_CONVENTION
adamic_closure *(adamic_counted_closure_canonical)(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[] ADAMIC_CANONICAL_DISCARD) {
 return counted_canonical(identity, code, count, cells, false
#ifdef ADAMIC_CLOSURE_DISCARDS
 , discard
#endif
 );
}
adamic_closure *(adamic_counted_closure_canonical_graph)(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[] ADAMIC_CANONICAL_DISCARD) {
 return counted_canonical(identity, code, count, cells, true
#ifdef ADAMIC_CLOSURE_DISCARDS
 , discard
#endif
 );
}
#endif

void adamic_closure_uncache(adamic_closure *closure) {
	adamic_environment *owner = closure->canonical_owner;
	if (owner == NULL) {
		return;
	}
	if (closure->canonical_previous != NULL) {
		closure->canonical_previous->canonical_next = closure->canonical_next;
	} else {
		owner->functions = closure->canonical_next;
	}
	if (closure->canonical_next != NULL) {
		closure->canonical_next->canonical_previous = closure->canonical_previous;
	}
}

#endif

// Fresh on every bind, including a target without captures. The ordinary closure
// free_children walks these owning cells, so graph and program regions see them too.
adamic_closure *adamic_bound_closure_new(adamic_code code, size_t count) {
 return (adamic_closure_new)(code, count);
}

adamic_value adamic_bound_closure_call(adamic_closure *self, adamic_value *arguments, size_t argument_count, size_t packed_size) {
 adamic_closure *target = self->cells[0]->value.reference;
 if (target == NULL) {
#ifdef ADAMIC_CLOSURE_CONVENTION
  return adamic_method_call(self->bound_method, self->cells[1]->value.reference, arguments, argument_count);
#else
  if (self->bound_method_storage.counted) {
   return self->bound_method_storage.counted_code(self->cells[1]->value.reference, arguments, argument_count);
  }
  return self->bound_method_storage.code(self->cells[1]->value.reference, arguments);
#endif
 }
#ifdef ADAMIC_CLOSURE_RECEIVERS
 return adamic_closure_receiver_call(target, self->cells[1]->value.reference, arguments, argument_count, packed_size);
#else
 (void)packed_size;
 return adamic_closure_call(target, arguments, argument_count);
#endif
}

// Both ordinary and bound closures own cells. The cells, in turn, release their
// reference-valued contents; bound calls only borrow those contents.
void adamic_closure_free_children(adamic_closure *closure, void (*drop)(void *)) {
 for (size_t index = 0; index < closure->count; index++) { drop(closure->cells[index]); }
}
