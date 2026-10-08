#include "harness.h"
#include <stdatomic.h>

_Static_assert(ATOMIC_BOOL_LOCK_FREE == 2, "fixture requires lock-free bool atomics");
static pthread_t caller;
static atomic_flag writer_selected = ATOMIC_FLAG_INIT;
static _Atomic bool writer_ready;
static _Atomic bool writer_released;
static _Thread_local bool hold_writer;
static adamic_string line = ADAMIC_STRING("signal worker line");

// Hold one actual buffer writer while it owns output_lock. The exit hook releases
// it before the protected shutdown, or after the mutant's unlocked flush.
void adamic_test_exit_writer(void) {
 if (!hold_writer) { return; }
 hold_writer = false;
 atomic_store_explicit(&writer_ready, true, memory_order_relaxed);
 while (!atomic_load_explicit(&writer_released, memory_order_relaxed)) {}
}
void adamic_test_exit_release(void) {
 atomic_store_explicit(&writer_released, true, memory_order_relaxed);
}
static void prepare(void) { caller = pthread_self(); }
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 if (pthread_equal(caller, pthread_self())) {
  while (!atomic_load_explicit(&writer_ready, memory_order_relaxed)) {}
  exit(0);
 }
 if (!atomic_flag_test_and_set_explicit(&writer_selected, memory_order_relaxed)) {
  hold_writer = true;
  adamic_write_line(adamic_stdout, &line);
 }
 return 1;
}
