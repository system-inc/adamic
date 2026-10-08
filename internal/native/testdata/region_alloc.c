#include "adamic.h"
#include "count.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static const bool references[] = {true, false, false, false};
static const adamic_shape shapes[] = {{2, NULL, references, NULL}, {3, NULL, references, NULL}, {4, NULL, references, NULL}};
static adamic_string text = ADAMIC_STRING("outside");
static adamic_object *items[8192];

int main(void) {
 adamic_region region = ADAMIC_REGION;
 adamic_object *outside = adamic_object_new(&shapes[0]);
 outside->slots[1].number = 7;
 adamic_weak *kept = adamic_weak_of(outside);
 for (int round = 0; round < 2; round++) {
  adamic_object *first = adamic_object_new_in(&region, &shapes[0]);
  if (first->slots[0].reference != NULL || first->slots[1].number != 0) { return 1; }
  adamic_weak *weak_first = adamic_weak_of(first);
  for (size_t i = 0; i < 8192; i++) {
   const adamic_shape *shape = &shapes[i % 3];
   adamic_object *object = adamic_object_new_filled_in(&region, shape);
   if ((uintptr_t)object % 16 != 0) { fprintf(stderr,"unaligned regional object\n"); return 2; }
   if (object->heap.references != 0 || object->heap.kind != adamic_kind_object ||
       object->heap.slab != ADAMIC_REGION_VALUE || object->shape != shape || object->class != NULL || object->frozen) { return 3; }
   object->slots[0].reference = adamic_string_concat(2, (adamic_string *const[]){&text, &text});
   adamic_region_hold(&region, object->slots[0].reference);
   for (size_t slot = 1; slot < shape->count; slot++) { object->slots[slot].number = (double)i; }
   items[i] = object;
  }
  if (!adamic_region_contains(&region,items[8191]) || adamic_region_contains(&region,outside) ||
      adamic_region_contains(&region,region.next)) {
   fprintf(stderr,"invalid regional cursor extent\n"); return 4;
  }
  if (!adamic_region_contains(&region,first)) {
   fprintf(stderr,"completed block extent was lost\n"); return 14;
  }
  size_t sum = 0;
  for (size_t i = 0; i < 8192; i++) { sum += (size_t)items[i]->slots[1].number; }
  if (sum != 8192 * 8191 / 2) { fprintf(stderr,"regional allocations alias\n"); return 5; }
  adamic_weak *weak_last = adamic_weak_of(items[8191]);
  // A shape larger than the 1 MiB block cap must still get aligned storage.
  bool *large_references = calloc(140000, sizeof *large_references);
  if (large_references == NULL) { return 6; }
  adamic_shape large_shape = {140000, NULL, large_references, NULL};
  adamic_object *large = adamic_object_new_in(&region, &large_shape);
  if ((uintptr_t)large % 16 != 0 || !adamic_region_contains(&region,large)) { return 7; }
  large->slots[139999].number = 11;
  if (!adamic_region_contains(&region,items[8191])) { return 8; }
  adamic_region_end(&region);
  free(large_references);
  if (region.blocks != NULL || region.next != NULL || region.end != NULL || region.holds_outside || region.count != 0) { return 9; }
  if (adamic_weak_target(weak_first) != NULL || adamic_weak_target(weak_last) != NULL || adamic_weak_target(kept) != outside) { return 10; }
  adamic_release(weak_first);
  adamic_release(weak_last);
#ifdef ADAMIC_COUNT
  if (adamic_counted.live != 2) { fprintf(stderr,"regional children were not released\n"); return 11; }
#endif
 }
 adamic_region_end(&region);
 adamic_release(outside);
 if (adamic_weak_target(kept) != NULL) { return 12; }
 adamic_release(kept);
#ifdef ADAMIC_COUNT
 if (adamic_counted.live != 0 || adamic_counted.allocations != adamic_counted.frees + adamic_counted.regions) { return 13; }
#endif
 puts("regions clean");
 return 0;
}
