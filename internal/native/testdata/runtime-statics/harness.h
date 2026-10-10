#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "parallel.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void prepare(void);
static void cleanup(void);
static double exercise(size_t index);
static pthread_mutex_t arrivals_lock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t arrivals_changed = PTHREAD_COND_INITIALIZER;
static size_t arrivals;
static _Thread_local bool arrived;

static adamic_value work(adamic_closure *self, adamic_value *arguments) {
 (void)self;
 size_t threads = adamic_parallel_threads();
 // Force real overlap on every worker. The rendezvous precedes the state accesses,
 // so it cannot repair a missing atomic or thread-local declaration in exercise.
 if (!arrived && threads > 1) {
  arrived = true;
  pthread_mutex_lock(&arrivals_lock);
  arrivals++;
  pthread_cond_broadcast(&arrivals_changed);
  while (arrivals < threads) { pthread_cond_wait(&arrivals_changed, &arrivals_lock); }
  pthread_mutex_unlock(&arrivals_lock);
 }
 double sum = 0;
 for (size_t repeat = 0; repeat < 32; repeat++) { sum += exercise((size_t)arguments[0].number); }
 return (adamic_value){.number = sum};
}

int main(int argc, char **argv) {
 adamic_start(argc, argv);
 prepare();
 adamic_array *items = adamic_array_new(256, false);
 for (size_t index = 0; index < 256; index++) { adamic_array_push(items, (adamic_value){.number = (double)index}); }
 adamic_closure *callback = adamic_closure_new(work, 0);
 adamic_array *results = adamic_parallel_map(items, callback, false);
 if (results == NULL || results->length != items->length) { abort(); }
 for (size_t index = 0; index < results->length; index++) {
  if (results->elements[index].number <= 0) { abort(); }
 }
 adamic_release(results); adamic_release(callback); adamic_release(items);
 adamic_parallel_shutdown();
 cleanup();
 puts("runtime statics clean");
 return 0;
}
