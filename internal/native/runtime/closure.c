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
	closure->count = count;
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = NULL;
	}
	return closure;
}

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
  if (cells[index].references) drop(cells[index].value.reference);
 }
}
