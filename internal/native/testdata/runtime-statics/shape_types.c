#include "harness.h"
// Each worker owns an eight-item chunk, whose callbacks make 256 exercise calls.
// Rendezvous before each of those calls on fresh metadata, so registration writes
// keep overlapping instead of turning into eight already-registered fast paths.
static const char *const names[] = {"x"};
static const bool references[] = {false};
static const adamic_field_kind kinds[] = {adamic_field_number};
static adamic_shape shapes[256];
static const int types[] = {1};
static adamic_shape_types metadata[256];
static pthread_barrier_t registrations;
static _Thread_local size_t visits;
static void prepare(void) {
 for (size_t index = 0; index < 256; index++) {
  shapes[index] = (adamic_shape){1, names, references, NULL, kinds};
  metadata[index] = (adamic_shape_types){&shapes[index], types, NULL, false};
 }
 if (pthread_barrier_init(&registrations, NULL, 4) != 0) { abort(); }
}
static void cleanup(void) { pthread_barrier_destroy(&registrations); }
static double exercise(size_t index) {
 size_t which = visits < 256 ? visits : index % 256;
 if (visits++ < 256) {
  // This orders earlier rounds, never the new round's registry accesses.
  pthread_barrier_wait(&registrations);
 }
 adamic_register_shape_types(&metadata[which]);
 adamic_object *object = adamic_object_new(&shapes[which]);
 object->slots[0].number = (double)which;
 adamic_heap *value = adamic_dynamic_property(&object->heap, "x");
 // A registry missing this shape (a lost push) panics; one that links an entry to itself never returns.
 if (value == NULL || value->kind != adamic_kind_number || ((adamic_number_box *)value)->number != (double)which) { abort(); }
 adamic_release(value);
 adamic_release(object);
 return 1;
}
