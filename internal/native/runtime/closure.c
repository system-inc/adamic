// closure.c: closures and the cells they capture.

#include "adamic.h"

adamic_cell *adamic_cell_new(adamic_value value, bool references) {
	adamic_cell *cell = adamic_allocate(sizeof *cell, adamic_kind_cell);
	cell->references = references;
	cell->ready = true;
	cell->owner = NULL;
	cell->value = value;
	return cell;
}

adamic_closure *adamic_closure_new(adamic_code code, size_t count) {
	adamic_closure *closure = adamic_allocate(sizeof *closure + count * sizeof closure->cells[0], adamic_kind_closure);
	closure->code = code;
	closure->receiver = false;
	closure->count = count;
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = NULL;
	}
	return closure;
}

adamic_environment *adamic_environment_new(size_t count) {
	adamic_environment *environment = adamic_allocate(sizeof *environment + count * sizeof environment->cells[0], adamic_kind_environment);
	environment->count = count;
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
