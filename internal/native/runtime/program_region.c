// Runtime half of step 06. Membership is a compiler decision; no tracing or merges.
#include "graph_regions.h"
#include "count.h"
#include <stddef.h>

#ifdef ADAMIC_PROGRAM_REGION
static adamic_heap *members;

void *adamic_program_adopt_owned(void *value, size_t size) {
 adamic_heap *heap = value;
 if (adamic_program_is(heap)) { return value; }
 if (heap == NULL || heap->references != 1 || adamic_weak_held(heap) ||
     (heap->slab & (ADAMIC_REGION_VALUE | ADAMIC_SHARED_HEADER)) != 0) {
  adamic_panic("compiler bug: Program adoption needs a fresh counted allocation", sizeof "compiler bug: Program adoption needs a fresh counted allocation" - 1);
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
  for (size_t i = 0; i < environment->count; i++) {
   environment->cells[i].owner = environment;
   environment->cells[i].heap.slab |= ADAMIC_PROGRAM_FLAG;
  }
 }
#endif
 return heap;
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
void adamic_program_region_end(void) {}
#endif
