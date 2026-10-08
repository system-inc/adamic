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
