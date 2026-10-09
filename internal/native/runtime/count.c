// count.c: the counts a counted build keeps (count.h).

#include "count.h"

#include <stdio.h>
#include <unistd.h>

adamic_counts adamic_counted;

void adamic_count_allocation(void) {
	atomic_fetch_add_explicit(&adamic_counted.allocations, 1, memory_order_relaxed);
	size_t live = atomic_fetch_add_explicit(&adamic_counted.live, 1, memory_order_relaxed) + 1;
	size_t peak = atomic_load_explicit(&adamic_counted.peak, memory_order_relaxed);
	while (peak < live && !atomic_compare_exchange_weak_explicit(&adamic_counted.peak, &peak, live, memory_order_relaxed, memory_order_relaxed)) {}

}

void adamic_count_report(void) {
	char line[200];
	int length = snprintf(line, sizeof line, "adamic: counts: allocations %zu frees %zu retains %zu releases %zu peak %zu regions %zu\n",
		adamic_counted.allocations, adamic_counted.frees, adamic_counted.retains, adamic_counted.releases, adamic_counted.peak, adamic_counted.regions);
	if (length > 0 && (size_t)length < sizeof line) {
		// Best effort, as a panic's message is: stderr may be what failed.
		(void)!write(2, line, (size_t)length);
	}
}

#ifdef ADAMIC_COUNT
// A finished program reports once main has returned, after it let go of its globals. A panic reports
// from adamic_panic, since it leaves by _exit and runs no destructor.
__attribute__((destructor)) static void report_at_exit(void) {
	adamic_count_report();
}
#endif
