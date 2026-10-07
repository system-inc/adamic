#include <stdatomic.h>
_Thread_local int local;
_Atomic int counter;
int step(void) { local++; return atomic_fetch_add_explicit(&counter, 1, memory_order_relaxed) + local; }
