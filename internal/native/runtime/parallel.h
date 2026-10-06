// Private pool lifecycle and observations used by C harnesses.
#ifndef ADAMIC_PARALLEL_H
#define ADAMIC_PARALLEL_H
#include <stddef.h>
size_t adamic_parallel_threads(void);
size_t adamic_parallel_workers(void);
void adamic_parallel_shutdown(void);
#endif
