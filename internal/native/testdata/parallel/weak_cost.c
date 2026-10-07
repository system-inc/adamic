#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <stdio.h>
#include <time.h>

int main(int argc, char **argv) {
	(void)argv;
	adamic_heap *value = adamic_box_number(7);
	adamic_weak *weak = argc > 1 ? adamic_weak_of(value) : NULL;
	struct timespec before, after;
	size_t held = 0;
	clock_gettime(CLOCK_MONOTONIC, &before);
	for (size_t index = 0; index < 20000000; index++) { held += adamic_weak_held(value); }
	clock_gettime(CLOCK_MONOTONIC, &after);
	if (held != (weak == NULL ? 0 : 20000000)) { return 3; }
	double seconds = (double)(after.tv_sec - before.tv_sec) + (double)(after.tv_nsec - before.tv_nsec) / 1e9;
	adamic_release(weak);
	adamic_release(value);
	printf("%.9f\n", seconds);
	return 0;
}
