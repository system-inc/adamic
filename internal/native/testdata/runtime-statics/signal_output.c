#include "harness.h"
#include <signal.h>
static pthread_t caller;
static int stop_signal;
static bool on_worker;
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
 pthread_sigmask(SIG_UNBLOCK, &signals, NULL);
 bool is_caller = pthread_equal(caller, pthread_self());
 for (size_t repeat = 0; repeat < 10000; repeat++) {
  adamic_write_line(adamic_stdout, &line);
  if (repeat == 128 && is_caller != on_worker) {
   if (stop_signal == 0) { exit(0); }
   raise(stop_signal);
  }
 }
 return 1;
}
