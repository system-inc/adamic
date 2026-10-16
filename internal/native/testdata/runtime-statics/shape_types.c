#include "harness.h"
// Four workers rendezvous for 64 rounds, each pushing its own fresh metadata.
// Sharing an entry would let a late worker acquire its registered flag and skip
// the push, hiding a missing registry lock from ThreadSanitizer.
static const char *const names[] = {"x"};
static const bool references[] = {false};
static adamic_shape shapes[256];
static const int types[] = {1};
static adamic_shape_types metadata[256];
static pthread_barrier_t registrations;
static size_t next_worker;
static _Thread_local size_t worker;
static _Thread_local size_t visits;
static void prepare(void) {
 for (size_t index = 0; index < 256; index++) {
  shapes[index] = (adamic_shape){1, names, references, NULL};
  metadata[index] = (adamic_shape_types){&shapes[index], types, NULL, false};
 }
 if (pthread_barrier_init(&registrations, NULL, 4) != 0) { abort(); }
}
static void cleanup(void) { pthread_barrier_destroy(&registrations); }
static double exercise(size_t index) {
 if (visits == 0) {
  worker = __atomic_fetch_add(&next_worker, 1, __ATOMIC_RELAXED);
  if (worker >= 4) { abort(); }
 }
 size_t which = visits < 64 ? visits * 4 + worker : index % 256;
 if (visits++ < 64) {
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
