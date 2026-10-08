// Fill multiple slabs without remote work, then free them from a second thread.
#include "adamic.h"
#include <pthread.h>
#include <stdint.h>
#include <stdio.h>

extern size_t adamic_test_remote_drains(void);
static adamic_cell *items[6000];

static uint32_t fill(void) {
	uint32_t maximum = 0;
	for (size_t i = 0; i < sizeof items / sizeof items[0]; i++) {
		items[i] = adamic_cell_new((adamic_value){.number = (double)i}, false);
		if (items[i]->heap.slab > maximum) { maximum = items[i]->heap.slab; }
	}
	return maximum;
}

static void *remote_free(void *unused) {
	(void)unused;
	for (size_t i = 0; i < sizeof items / sizeof items[0]; i++) { adamic_release(items[i]); }
	adamic_heap_thread_end();
	return NULL;
}

int main(void) {
	uint32_t first = fill();
	if (adamic_test_remote_drains() != 0) { puts("serial scan"); return 1; }
	pthread_t thread;
	if (pthread_create(&thread, NULL, remote_free, NULL) != 0) { return 2; }
	if (pthread_join(thread, NULL) != 0) { return 3; }
	uint32_t second = fill();
	if (second > first || adamic_test_remote_drains() == 0) { puts("remote slots lost"); return 4; }
	for (size_t i = 0; i < sizeof items / sizeof items[0]; i++) {
		if (items[i]->value.number != (double)i) { return 5; }
		adamic_release(items[i]);
	}
	puts("serial skip; remote reuse");
	return 0;
}
