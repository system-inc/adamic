#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
static size_t live, peak, requested, peak_requested, rounded, peak_rounded;
void *__real_adamic_allocate(size_t, enum adamic_kind);
void __real_adamic_weak_forget(void *);
static size_t slot(size_t size) { return size <= 256 ? (size + 15) / 16 * 16 : size; }
void *__wrap_adamic_allocate(size_t size, enum adamic_kind kind) {
 void *value = __real_adamic_allocate(size, kind);
 if (kind == adamic_kind_closure) {
  live++; requested += size; rounded += slot(size);
  if (live > peak) peak = live;
  if (requested > peak_requested) peak_requested = requested;
  if (rounded > peak_rounded) peak_rounded = rounded;
 }
 return value;
}
void __wrap_adamic_weak_forget(void *value) {
 const adamic_heap *heap = value;
 if (heap->kind == adamic_kind_closure) {
  const adamic_closure *closure = value;
  size_t size = sizeof *closure + closure->count * sizeof closure->cells[0];
  live--; requested -= size; rounded -= slot(size);
 }
 __real_adamic_weak_forget(value);
}
__attribute__((destructor)) static void report(void) {
 fprintf(stderr, "peak_closures=%zu peak_requested=%zu peak_slab_slots=%zu final_live=%zu header=%zu\n", peak, peak_requested, peak_rounded, live, sizeof(adamic_closure));
}
