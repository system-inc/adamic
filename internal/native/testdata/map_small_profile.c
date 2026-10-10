// Synthetic map-only workload shaped by the tsc 6.0.3 census, not a replay of tsc.
#define _POSIX_C_SOURCE 200809L
// macOS hides struct rusage's ru_maxrss under strict POSIX unless asked; Linux ignores this.
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <time.h>

#define MAPS 271022
#define LIVE 143047
#define MAX_KEYS 2048
static const size_t populations[] = {103574, 37357, 107172, 10357, 11965, 451, 111, 35};
static const size_t sizes[] = {0, 1, 3, 8, 32, 128, 512, 2048};

static adamic_value key_for(size_t key, bool strings, adamic_string **names, bool owned) {
	if (strings) {
		return (adamic_value){.reference = owned ? adamic_retain(names[key]) : names[key]};
	}
	return (adamic_value){.number = (double)key};
}

int main(void) {
	adamic_map **maps = calloc(LIVE, sizeof *maps);
	adamic_string *names[MAX_KEYS];
	if (maps == NULL) { return 2; }
	for (size_t key = 0; key < MAX_KEYS; key++) {
		char text[32];
		int length = snprintf(text, sizeof text, "symbol_%zu", key);
		names[key] = adamic_string_allocate((size_t)length);
		memcpy((char *)names[key]->bytes, text, (size_t)length);
	}
	clock_t cpu = clock();
	struct timespec start, finish;
	clock_gettime(CLOCK_MONOTONIC, &start);
	size_t built = 0, kept = 0, sets = 0, gets = 0, has = 0;
	double checksum = 0;
	for (size_t group = 0; group < sizeof populations / sizeof populations[0]; group++) {
		for (size_t each = 0; each < populations[group]; each++) {
			// Spread sizes within the 2..4 bin; other bins use a recorded representative.
			size_t size = group == 2 ? 2 + each % 3 : sizes[group];
			bool strings = built % 2 != 0;
			adamic_map *map = adamic_map_new(strings, false);
			for (size_t key = 0; key < size; key++) {
				adamic_map_set(map, key_for(key, strings, names, true), (adamic_value){.number = (double)key});
				sets++;
			}
			built++;
			// Keep the census's checked live population, distributed through every size bin.
			if (built * LIVE / MAPS > kept) { maps[kept++] = map; }
			else { adamic_release(map); }
		}
	}
	if (built != MAPS || kept != LIVE || sets > 1201980) { return 3; }
	for (size_t step = 0; sets < 1201980; step++) {
		adamic_map *map = maps[step % LIVE];
		if (map->count == 0) { continue; }
		size_t key = step % map->count;
		adamic_map_set(map, key_for(key, map->string_keys, names, true), (adamic_value){.number = (double)key});
		sets++;
	}
	for (size_t step = 0; gets < 4429860 || has < 318504; step++) {
		adamic_map *map = maps[step % LIVE];
		size_t key = map->count == 0 ? 0 : step % map->count;
		adamic_value *found = adamic_map_get(map, key_for(key, map->string_keys, names, false));
		if (gets < 4429860) { checksum += found == NULL ? -1 : found->number; gets++; }
		else if (has < 318504) { checksum += found != NULL; has++; }
	}
	// Include releasing the complete retained graph in measured time.
	for (size_t index = 0; index < kept; index++) { adamic_release(maps[index]); }
	for (size_t key = 0; key < MAX_KEYS; key++) { adamic_release(names[key]); }
	free(maps);
	clock_gettime(CLOCK_MONOTONIC, &finish);
	struct rusage usage;
	getrusage(RUSAGE_SELF, &usage);
	printf("maps=%zu live=%zu set=%zu get=%zu has=%zu checksum=%.0f cpu=%.9f wall=%.9f rss=%ld\n",
		built, kept, sets, gets, has, checksum, (double)(clock() - cpu) / CLOCKS_PER_SEC,
		(double)(finish.tv_sec - start.tv_sec) + (double)(finish.tv_nsec - start.tv_nsec) / 1e9, usage.ru_maxrss);
	return 0;
}
