// Appended to the runtime's map.c by map_hash_test.go: hash_key stays private in production.
#include <stdio.h>

static size_t failures = 0;

static void check(bool condition, const char *message) {
	if (!condition) {
		fprintf(stderr, "%s\n", message);
		failures++;
	}
}

static double pattern_number(size_t pattern, size_t index) {
	double i = (double)index;
	switch (pattern) {
		case 0: return i;
		case 1: return i * 1024;
		case 2: return -i;
		case 3: return i / 8;
		case 4: return i + 0.5;
		case 5: return i / 10;
		case 6: return 9007199254740992.0 - i;
		case 7: return 9007199254740992.0 + i * 2;
		default: return index % 2 == 0 ? i : -i;
	}
}

static void distribution(size_t pattern, size_t count, bool optional) {
	const char *names[] = {"integers", "stride1024", "negative", "eighths", "halves", "tenths", "below2^53", "above2^53", "signed", "strings", "objects"};
	adamic_map *map = pattern == 9 ? adamic_map_new(true, false) :
		pattern == 10 ? adamic_map_new_identity(false) :
		optional ? adamic_map_new_maybe_numbers(false) : adamic_map_new(false, false);
	static const adamic_shape shape = {0, NULL, NULL, NULL};
	for (size_t index = 0; index < count; index++) {
		adamic_value key;
		if (pattern == 9) {
			char text[64];
			int length = snprintf(text, sizeof text, "tenant_%zu/item_%05zu", index % 17, index);
			adamic_string *string = adamic_string_allocate((size_t)length);
			memcpy((char *)string->bytes, text, (size_t)length);
			key.reference = string;
		} else if (pattern == 10) {
			key.reference = adamic_object_new(&shape);
		} else {
			key.number = pattern_number(pattern, index);
			if (optional) {
				key.number = adamic_maybe_number_pack((adamic_maybe_number){true, key.number});
			}
		}
		adamic_map_set(map, key, (adamic_value){.number = (double)index});
	}
	check(map->count == count, "distinct input keys merged");
	size_t *homes = calloc(map->bucket_count, sizeof *homes);
	check(homes != NULL, "home histogram allocation failed");
	if (homes == NULL) { adamic_release(map); return; }
	size_t mask = map->bucket_count - 1;
	size_t maximum_probe = 0, maximum_home = 0, distinct_homes = 0, maximum_miss = 1;
	for (size_t bucket = 0; bucket < map->bucket_count; bucket++) {
		if (map->buckets[bucket] == 0) { continue; }
		const adamic_map_entry *entry = &map->entries[map->buckets[bucket] - 1];
		size_t home = hash_key(map, entry->key) & mask;
		homes[home]++;
		size_t probe = ((bucket - home) & mask) + 1;
		if (probe > maximum_probe) { maximum_probe = probe; }
		check(adamic_map_get(map, entry->key) == &entry->value, "inserted key cannot be found");
	}
	// Start after an empty bucket, then walk once around: includes a run wrapping past bucket zero.
	size_t empty = 0;
	while (map->buckets[empty] != 0) { empty++; }
	size_t run = 0;
	for (size_t offset = 1; offset <= map->bucket_count; offset++) {
		size_t bucket = (empty + offset) & mask;
		if (map->buckets[bucket] == 0) { run = 0; } else { run++; }
		if (run + 1 > maximum_miss) { maximum_miss = run + 1; }
		if (homes[bucket] > 0) { distinct_homes++; }
		if (homes[bucket] > maximum_home) { maximum_home = homes[bucket]; }
	}
	printf("%s%s entries=%zu buckets=%zu homes=%zu max_home=%zu hit=%zu miss=%zu\n",
		names[pattern], optional ? "?" : "", count, map->bucket_count, distinct_homes, maximum_home, maximum_probe, maximum_miss);
	// At the runtime's maximum 50% load, 64 is generous for these fixed patterns and still catches
	// severe clustering. This is a regression bound, not a worst-case guarantee for arbitrary keys.
	if (maximum_probe > 64 || maximum_miss > 64) {
		fprintf(stderr, "probe bound 64 exceeded: %s%s entries=%zu hit=%zu miss=%zu\n",
			names[pattern], optional ? "?" : "", count, maximum_probe, maximum_miss);
		failures++;
	}
	free(homes);
	adamic_release(map);
}

static void same_value_zero(bool optional) {
	adamic_map *map = optional ? adamic_map_new_maybe_numbers(false) : adamic_map_new(false, false);
	adamic_value zero = {.number = 0}, negative_zero = {.number = -0.0};
	check(hash_key(map, zero) == hash_key(map, negative_zero), "zero hashes differ");
	adamic_map_set(map, negative_zero, (adamic_value){.number = 1});
	adamic_map_set(map, zero, (adamic_value){.number = 2});
	check(map->count == 1 && !signbit(map->entries[0].key.number), "zero storage or equality changed");
	uint64_t representations[] = {0x7ff8000000000000ull, 0x7ff8000000000042ull, 0xfff8000000000042ull, 0x7ff0000000000001ull};
	uint64_t expected = 0;
	for (size_t index = 0; index < sizeof representations / sizeof *representations; index++) {
		adamic_value nan;
		memcpy(&nan.number, &representations[index], sizeof nan.number);
		if (optional) { nan.number = adamic_maybe_number_pack((adamic_maybe_number){true, nan.number}); }
		uint64_t hash = hash_key(map, nan);
		if (index == 0) { expected = hash; }
		check(hash == expected, "NaN hashes differ");
		adamic_map_set(map, nan, (adamic_value){.number = (double)index});
		check(map->count == 2 && adamic_map_get(map, nan)->number == (double)index, "NaN equality changed");
	}
	if (optional) {
		adamic_value missing = {.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0})};
		adamic_map_set(map, missing, (adamic_value){.number = 7});
		check(map->count == 3 && adamic_map_get(map, missing)->number == 7, "undefined merged with NaN");
	}
	adamic_release(map);
}

int main(void) {
	same_value_zero(false);
	same_value_zero(true);
	for (size_t count = 8; count <= 4096; count *= 2) {
		for (size_t pattern = 0; pattern <= 10; pattern++) {
			distribution(pattern, count, false);
			if (pattern < 9) { distribution(pattern, count, true); }
		}
	}
	return failures > 0 ? 1 : 0;
}
