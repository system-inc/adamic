#include "harness.h"
static adamic_map *map;
static adamic_string key = ADAMIC_STRING("Ké😀");
static void prepare(void) {
 map = adamic_map_new(true, false);
 adamic_map_set(map, (adamic_value){.reference = &key}, (adamic_value){.number = 17});
 adamic_share(map);
}
static void cleanup(void) { adamic_release(map); }
static double exercise(size_t index) {
 (void)index;
 adamic_value *value = adamic_map_get(map, (adamic_value){.reference = &key});
 adamic_map_iterator *iterator = adamic_map_iterate(map);
 adamic_value found_key, found_value;
 if (value == NULL || value->number != 17 || !adamic_map_iterator_next(iterator, &found_key, &found_value) || found_value.number != 17) { abort(); }
 adamic_release(iterator);
 return 1;
}
