#include "harness.h"
#include <signal.h>
#include <stdatomic.h>

_Static_assert(ATOMIC_BOOL_LOCK_FREE == 2, "signal fixture requires lock-free bool atomics");

static pthread_t caller;
static atomic_flag sender_selected = ATOMIC_FLAG_INIT;
static atomic_flag writer_selected = ATOMIC_FLAG_INIT;
static _Atomic bool writer_ready;
static _Atomic bool writer_released;
static _Thread_local bool hold_writer;
static adamic_string line = ADAMIC_STRING("signal worker line");

// These hooks are linked only into a cloned runtime. All gates are relaxed: they force the
// conflicting accesses to occur, but never add a happens-before edge that could hide the race.
void adamic_test_signal_writer(void) {
 if (!hold_writer) { return; }
 hold_writer = false;
 atomic_store_explicit(&writer_ready, true, memory_order_relaxed);
 while (!atomic_load_explicit(&writer_released, memory_order_relaxed)) {}
}
void adamic_test_signal_release(void) {
 atomic_store_explicit(&writer_released, true, memory_order_relaxed);
}
static void prepare(void) { caller = pthread_self(); }
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 bool sender = !pthread_equal(caller, pthread_self()) &&
  !atomic_flag_test_and_set_explicit(&sender_selected, memory_order_relaxed);
 if (sender) {
  while (!atomic_load_explicit(&writer_ready, memory_order_relaxed)) {}
  sigset_t signal;
  sigemptyset(&signal);
  sigaddset(&signal, SIGTERM);
  pthread_sigmask(SIG_UNBLOCK, &signal, NULL);
  raise(SIGTERM);
 } else if (!atomic_flag_test_and_set_explicit(&writer_selected, memory_order_relaxed)) {
  hold_writer = true;
  // The hook pauses after the real buffer write while this call still owns output_lock.
  // The safe handler releases this writer, then its loop waits for the completed newline.
  adamic_write_line(adamic_stdout, &line);
 }
 // No second sender, process exit, or further buffer write can mask the one handler race.
 for (;;) { atomic_signal_fence(memory_order_seq_cst); }
}
