#include "harness.h"
#include <stdatomic.h>

_Static_assert(ATOMIC_BOOL_LOCK_FREE == 2, "fixture requires lock-free bool atomics");
_Static_assert(ATOMIC_INT_LOCK_FREE == 2, "fixture requires lock-free int atomics");
static _Atomic unsigned selected;
static _Atomic bool writer_ready;
static _Atomic bool writer_released;
static _Thread_local unsigned role;
static _Thread_local bool completed;
static adamic_string decomposed = ADAMIC_STRING("é");
static adamic_string composed = ADAMIC_STRING("é");
static adamic_string nfc = ADAMIC_STRING("NFC");
static adamic_string nfd = ADAMIC_STRING("NFD");
static bool mappings;

// Only cloned runtime snapshots call these hooks. Relaxed gates force a real cache
// write followed by another thread's access, without ordering those accesses for TSan.
void adamic_test_normalization_before(void) {
 if (role == 2 && !completed) {
  while (!atomic_load_explicit(&writer_ready, memory_order_relaxed)) {}
 }
}
void adamic_test_normalization_after(void) {
 if (completed) { return; }
 completed = true;
 if (role == 1) {
  atomic_store_explicit(&writer_ready, true, memory_order_relaxed);
  while (!atomic_load_explicit(&writer_released, memory_order_relaxed)) {}
 } else if (role == 2) {
  atomic_store_explicit(&writer_released, true, memory_order_relaxed);
 }
}
static void prepare(void) {
 const char *kind = getenv("ADAMIC_NORMALIZATION_KIND");
 mappings = kind != NULL && strcmp(kind, "normalization_mappings") == 0;
}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 unsigned ticket = atomic_fetch_add_explicit(&selected, 1, memory_order_relaxed);
 if (ticket >= 2) { return 1; }
 role = ticket + 1;
 adamic_string *actual = adamic_string_normalize(mappings ? &composed : &decomposed, mappings ? &nfd : &nfc);
 if (!completed || !adamic_string_equal(actual, mappings ? &decomposed : &composed)) { abort(); }
 adamic_release(actual);
 return 1;
}
