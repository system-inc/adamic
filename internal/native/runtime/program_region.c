// Runtime half of step 06. Membership is a compiler decision; no tracing or merges.
#include "graph_regions.h"
#include "count.h"
#include <stddef.h>
#include <stdio.h>

#ifdef ADAMIC_PROGRAM_REGION
static adamic_heap *members;

void *adamic_program_adopt_owned(void *value, size_t size) {
 adamic_heap *heap = value;
 if (adamic_program_is(heap)) { return value; }
 if (heap == NULL || heap->references != 1 || adamic_weak_held(heap) ||
     (heap->slab & (ADAMIC_REGION_VALUE | ADAMIC_SHARED_HEADER | ADAMIC_GRAPH_FLAG)) != 0) {
  adamic_panic("compiler bug: Program adoption needs a fresh counted allocation", sizeof "compiler bug: Program adoption needs a fresh counted allocation" - 1);
 }
 // Adoption moves the allocation, so only the kinds it can re-aim may join (graph_regions.c).
 const char *unmovable = adamic_region_unmovable(heap);
 if (unmovable != NULL) {
  char message[96];
  int length = snprintf(message, sizeof message, "compiler bug: %s can't be adopted into the Program region", unmovable);
  adamic_panic(message, (size_t)length);
 }
#ifdef ADAMIC_CANONICAL_CLOSURES
 if (heap->kind == adamic_kind_environment) {
  adamic_environment *environment = (adamic_environment *)heap;
  for (size_t i = 0; i < environment->count; i++) {
   if (adamic_weak_held(&environment->cells[i])) {
    adamic_panic("compiler bug: Program adoption needs a fresh counted allocation", sizeof "compiler bug: Program adoption needs a fresh counted allocation" - 1);
   }
  }
 }
 if ((heap->kind == adamic_kind_environment && ((adamic_environment *)heap)->functions != NULL) ||
     (heap->kind == adamic_kind_closure && ((adamic_closure *)heap)->canonical_owner != NULL)) {
  adamic_panic("compiler bug: Program adoption must precede canonical caching", sizeof "compiler bug: Program adoption must precede canonical caching" - 1);
 }
#endif
 heap = adamic_heap_program_storage(heap, size);
 adamic_graph_header_of(heap)->next = members;
 members = heap;
#ifdef ADAMIC_CANONICAL_CLOSURES
 // Interior cells share this indivisible member and take the header fast path too.
 if (heap->kind == adamic_kind_environment) {
  adamic_environment *environment = (adamic_environment *)heap;
  environment->cells = (adamic_cell *)(void *)(environment + 1);
  for (size_t i = 0; i < environment->count; i++) {
   environment->cells[i].owner = &environment->heap;
   environment->cells[i].heap.slab |= ADAMIC_PROGRAM_FLAG;
  }
 }
#endif
 return heap;
}

// The switch between Programs (#64pg02g). A run that builds a second Program (tsc's watch mode,
// or --build over several projects) would otherwise keep every Program it ever built until exit.
// Lowering calls adamic_program_region_begin at the entry of each function that allocates the
// Program root type, so the latch only asks whether this is the first Program or a later one,
// and saturates at 2 rather than wrapping back to adoption after 2^32 rebuilds. With no begin
// at all the generation stays 0 and every member site adopts, as before the switch.
//
// Single-threaded by construction, so a plain static and no atomic: a Program-region program
// refuses parallelMap (lower/program_region.go), and members refuse sharing (share.c).
//
// Adoption decides once, at allocation, and never re-adopts. A value adopted at generation 0 or 1
// is never later passed to graph adoption; graph_regions.c's union-find treats a member as
// outside, as it always has, so no merge walks into one. Under-inclusion is the proven case
// (program_region_core_test.go's membership mutants): an edge into a member needs no count, and
// a member's edge out is a real retain, released only when adamic_program_region_end frees it.
// So the first Program's members are shared by every later Program until exit, and a later
// graph region or counted value a member points into is held until exit too. That is intended
// retention, and the boundary can never free early, wherever it falls.
static unsigned program_generation;

void adamic_program_region_begin(void) {
 if (program_generation < 2) { program_generation++; }
}

void *adamic_program_adopt_or_fallback(void *value, size_t size, bool graph) {
 if (program_generation < 2) { return adamic_program_adopt_owned(value, size); }
 return graph ? adamic_graph_adopt_owned(value, size) : value;
}

void adamic_program_region_end(void) {
 // Keep every member alive throughout invalidation and counted-child destruction.
 for (adamic_heap *each = members; each != NULL; each = adamic_graph_header_of(each)->next) {
  adamic_weak_forget(each);
#ifdef ADAMIC_CANONICAL_CLOSURES
  if (each->kind == adamic_kind_environment) {
   adamic_environment *environment = (adamic_environment *)each;
   for (size_t i = 0; i < environment->count; i++) { adamic_weak_forget(&environment->cells[i]); }
  }
#endif
 }
 for (adamic_heap *each = members; each != NULL; each = adamic_graph_header_of(each)->next) {
  adamic_heap_free_children(each, adamic_release);
 }
 while (members != NULL) {
  adamic_heap *each = members;
  members = adamic_graph_header_of(each)->next;
  adamic_heap_free_storage(each, each->slab);
 }
}
#else
void *adamic_program_adopt_owned(void *value, size_t size) {
 (void)value; (void)size;
 adamic_panic("Program region emission requires ADAMIC_PROGRAM_REGION", sizeof "Program region emission requires ADAMIC_PROGRAM_REGION" - 1);
}
void adamic_program_region_begin(void) {}
void *adamic_program_adopt_or_fallback(void *value, size_t size, bool graph) {
 (void)graph;
 return adamic_program_adopt_owned(value, size);
}
void adamic_program_region_end(void) {}
#endif
