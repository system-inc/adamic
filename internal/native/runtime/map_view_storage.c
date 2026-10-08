// Read conversion preserves a Map's original storage and object identity.
#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>

static bool map_read_reference(unsigned char type) {
 return type == 3 || type == 4 || type == 5 || type == 6 || type == 8 || type == 10 || type == 11;
}

static _Noreturn void map_storage_failure(const char *part, unsigned char wanted, unsigned char source) {
 static const char *const names[] = {"uncertified", "number", "boolean", "string", "object", "array", "Map", "number | undefined", "function", "boolean | undefined", "boxed union", "Weak"};
 char message[256];
 int length = snprintf(message, sizeof message, "field read failed: Map %s storage cannot be converted; expected %s, found %s", part, wanted < 12 ? names[wanted] : "unknown", source < 12 ? names[source] : "unknown");
 adamic_panic(message, (size_t)length);
}

// References are returned owned, including freshly boxed scalar reads.
static adamic_value map_read_slot(unsigned char source, adamic_value value, unsigned char wanted, const char *part) {
 if (source == 0 || source == wanted || (wanted == 10 && map_read_reference(source))) {
  if (map_read_reference(wanted)) { value.reference = adamic_retain(value.reference); }
  return value;
 }
 if (wanted == 10 && source == 1) { return (adamic_value){.reference = adamic_box_number(value.number)}; }
 if (wanted == 10 && source == 2) { return (adamic_value){.reference = value.boolean ? &adamic_box_true : &adamic_box_false}; }
 if (wanted == 10 && source == 7) {
  adamic_maybe_number present = adamic_maybe_number_unpack(value.number);
  return (adamic_value){.reference = present.present ? adamic_box_number(present.number) : NULL};
 }
 if (wanted == 10 && source == 9) {
  adamic_maybe_boolean present = adamic_maybe_boolean_unpack(value.maybe_boolean);
  return (adamic_value){.reference = !present.present ? NULL : present.boolean ? (void *)&adamic_box_true : (void *)&adamic_box_false};
 }
 if (wanted == 7 && source == 1) { return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){true, value.number})}; }
 if (wanted == 9 && source == 2) { return (adamic_value){.maybe_boolean = adamic_maybe_boolean_pack((adamic_maybe_boolean){true, value.boolean})}; }
 map_storage_failure(part, wanted, source);
}

adamic_value adamic_map_read_value(const adamic_map *map, adamic_value value, unsigned char wanted) {
 return map_read_slot(map->value_type, value, wanted, "value");
}

adamic_array *adamic_map_values_as(const adamic_map *map, unsigned char wanted) {
 adamic_array *values = adamic_array_new(map->count, map_read_reference(wanted));
 adamic_array_view_storage(values, wanted);
 for (size_t index = 0; index < map->used; index++) {
  if (!map->entries[index].deleted) {
   adamic_array_push(values, adamic_map_read_value(map, map->entries[index].value, wanted));
  }
 }
 return values;
}

adamic_array *adamic_map_entries_as(const adamic_map *map, const adamic_shape *shape, unsigned char key, unsigned char wanted) {
 adamic_array *entries = adamic_array_new(map->count, true);
 adamic_array_view_storage(entries, 4);
 for (size_t index = 0; index < map->used; index++) {
  const adamic_map_entry *entry = &map->entries[index];
  if (entry->deleted) { continue; }
  adamic_object *pair = adamic_object_new(shape);
  pair->slots[0] = adamic_map_read_key(map, entry->key, key);
  pair->slots[1] = adamic_map_read_value(map, entry->value, wanted);
  adamic_object_field_types(pair)[0] = key;
  adamic_object_field_types(pair)[1] = wanted;
  if (adamic_graph_is(map)) { pair = adamic_graph_adopt_owned(pair, sizeof *pair + shape->count * sizeof(adamic_value)); }
  adamic_array_push(entries, (adamic_value){.reference = pair});
 }
 return entries;
}

// A converted key snapshot owns references just like a value snapshot.
adamic_value adamic_map_read_key(const adamic_map *map, adamic_value key, unsigned char wanted) {
 return map_read_slot(map->key_type, key, wanted, "key");
}

adamic_value *adamic_map_get_as(const adamic_map *map, adamic_value key, unsigned char supplied) {
 if (map->key_type == 0 || map->key_type == supplied) { return adamic_map_get(map, key); }
 if (map->key_type == 1 && supplied == 7) {
  adamic_maybe_number query = adamic_maybe_number_unpack(key.number);
  if (!query.present) { return NULL; }
  return adamic_map_get(map, (adamic_value){.number = query.number});
 }
 if (map->key_type == 2 && supplied == 9) {
  adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(key.maybe_boolean);
  if (!boolean.present) { return NULL; }
  return adamic_map_get(map,(adamic_value){.boolean=boolean.boolean});
 }
 if (supplied == 10) {
  const adamic_heap *reference = key.reference;
  if (reference == NULL) {
   if (map->key_type == 3) return adamic_map_get(map,key);
   if (map->key_type == 7) return adamic_map_get(map,(adamic_value){.number=adamic_maybe_number_pack((adamic_maybe_number){false,0})});
   if (map->key_type == 9) return adamic_map_get(map,(adamic_value){.maybe_boolean=2});
   return NULL;
  }
  if (reference->kind == adamic_kind_number && (map->key_type == 1 || map->key_type == 7)) {
   double number = ((const adamic_number_box *)reference)->number;
   return adamic_map_get(map,(adamic_value){.number=map->key_type == 7 ? adamic_maybe_number_pack((adamic_maybe_number){true,number}) : number});
  }
  if (reference->kind == adamic_kind_boolean && (map->key_type == 2 || map->key_type == 9)) {
   bool boolean = ((const adamic_boolean_box *)reference)->boolean;
   return adamic_map_get(map,map->key_type == 9 ? (adamic_value){.maybe_boolean=adamic_maybe_boolean_pack((adamic_maybe_boolean){true,boolean})} : (adamic_value){.boolean=boolean});
  }
  if (reference->kind == adamic_kind_string && map->key_type == 3) return adamic_map_get(map,key);
  if (map->key_type == 1 || map->key_type == 2 || map->key_type == 3 || map->key_type == 7 || map->key_type == 9) return NULL;
 }
 map_storage_failure("key lookup", map->key_type, supplied);
}

adamic_array *adamic_map_keys_as(const adamic_map *map, unsigned char wanted) {
 adamic_array *keys = adamic_array_new(map->count, map_read_reference(wanted));
 adamic_array_view_storage(keys, wanted);
 for (size_t index = 0; index < map->used; index++) {
  if (map->entries[index].deleted) { continue; }
  adamic_value key = adamic_map_read_key(map, map->entries[index].key, wanted);
  adamic_array_push(keys, key);
 }
 return keys;
}
