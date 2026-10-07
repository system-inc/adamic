#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <stdio.h>
#include <time.h>

int main(void) {
	adamic_heap *value = adamic_box_number(7);
	struct timespec before, after;
	clock_gettime(CLOCK_MONOTONIC, &before);
	for (size_t index = 0; index < 50000000; index++) { adamic_retain(value); adamic_release(value); }
	clock_gettime(CLOCK_MONOTONIC, &after);
	double seconds = (double)(after.tv_sec - before.tv_sec) + (double)(after.tv_nsec - before.tv_nsec) / 1e9;
	adamic_release(value);
	printf("%.9f\n", seconds);
	return 0;
}
