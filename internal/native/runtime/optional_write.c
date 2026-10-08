// A checked optional write must create an own property, including for undefined.
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

bool adamic_optional_write_presence(const adamic_object *object, const adamic_string *name, const char *site) {
 if (object != NULL && adamic_object_has(object, name)) return true;
 size_t capacity = name->length + strlen(site) + sizeof "optional write lost own presence: '' at ";
 char *message = malloc(capacity);
 if (message == NULL) adamic_panic("out of memory", 13);
 int length = snprintf(message, capacity, "optional write lost own presence: '%.*s' at %s", (int)name->length, name->bytes, site);
 adamic_panic(message, (size_t)length);
}

// Capture before later member initializers can mutate the source. Nullish spread
// contributes no properties, as Node does.
adamic_array *adamic_optional_spread_keys(const adamic_object *source) {
 return source == NULL ? adamic_array_new(0, true) : adamic_object_keys(source);
}

bool adamic_optional_spread_presence(const adamic_object *target, const adamic_array *keys, const char *site) {
 for (size_t index = 0; index < keys->length; index++) {
  const adamic_string *name = adamic_array_at_integer(keys, (int64_t)index)->reference;
  (void)adamic_optional_write_presence(target, name, site);
 }
 return true;
}

// The object ABI reserves insertion ranks separately from value slots. Optional
// interface views must keep this ABI, not reinterpret a value-only structure.
bool adamic_optional_view_storage(const adamic_object *object, bool nullable, const char *site) {
 if (object == NULL ? nullable : object->heap.kind == adamic_kind_object) return true;
 size_t capacity = strlen(site) + sizeof "optional view lacks own-presence storage at ";
 char *message = malloc(capacity);
 if (message == NULL) adamic_panic("out of memory", 13);
 int length = snprintf(message, capacity, "optional view lacks own-presence storage at %s", site);
 adamic_panic(message, (size_t)length);
}
