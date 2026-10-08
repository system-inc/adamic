// Private pool lifecycle and observations used by C harnesses.
#ifndef ADAMIC_PARALLEL_H
#define ADAMIC_PARALLEL_H
#include <stddef.h>
#include "tsan_test.h"
// Implemented by the runtime lane's shared-heap and worker-stack support.
void adamic_share(void *value);
void adamic_heap_thread_end(void);
void adamic_stack_thread_start(void);
size_t adamic_parallel_threads(void);
size_t adamic_parallel_workers(void);
size_t adamic_parallel_grain(size_t items, size_t threads);
void adamic_parallel_shutdown(void);
#endif
