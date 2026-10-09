#include "harness.h"
#include <signal.h>
#include <stdatomic.h>
static pthread_t caller;
static int stop_signal;
static bool on_worker;
static atomic_flag stop_selected = ATOMIC_FLAG_INIT;
static adamic_string line = ADAMIC_STRING("signal worker line");
static void prepare(void) {
 caller = pthread_self();
 stop_signal = atoi(getenv("ADAMIC_STOP_SIGNAL"));
 on_worker = getenv("ADAMIC_STOP_WORKER") != NULL;
}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 // Explicitly unblock on a pool worker too: safety must not depend on pool masks.
 sigset_t signals;
 sigemptyset(&signals);
 sigaddset(&signals, SIGTERM); sigaddset(&signals, SIGINT); sigaddset(&signals, SIGHUP);
 if (stop_signal != 0) { sigaddset(&signals, stop_signal); }
 pthread_sigmask(SIG_UNBLOCK, &signals, NULL);
 bool is_caller = pthread_equal(caller, pthread_self());
 // One sender keeps a second worker from terminating the process while TSan is
 // reporting the first handler's race. Relaxed selection does not order writes.
 bool sender = is_caller != on_worker &&
  !atomic_flag_test_and_set_explicit(&stop_selected, memory_order_relaxed);
 for (size_t repeat = 0; repeat < 10000; repeat++) {
  adamic_write_line(adamic_stdout, &line);
  if (repeat == 128 && sender) {
   if (stop_signal == 0) { exit(0); }
   raise(stop_signal);
  }
 }
 return 1;
}
