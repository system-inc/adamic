// count.c: the counts a counted build keeps (count.h).

#include "count.h"

#include <stdio.h>
#include <unistd.h>

adamic_counts adamic_counted;

void adamic_count_allocation(void) {
	adamic_counted.allocations++;
	adamic_counted.live++;
	if (adamic_counted.live > adamic_counted.peak) {
		adamic_counted.peak = adamic_counted.live;
	}
}

void adamic_count_report(void) {
	char line[256];
	int length = snprintf(line, sizeof line, "adamic: counts: allocations %zu frees %zu retains %zu releases %zu peak %zu regions %zu%s\n",
		adamic_counted.allocations, adamic_counted.frees, adamic_counted.retains, adamic_counted.releases, adamic_counted.peak, adamic_counted.regions, adamic_counted.termination == NULL ? "" : adamic_counted.termination);
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
