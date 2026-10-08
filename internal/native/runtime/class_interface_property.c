// Fixed public property presence, including prototypes, without invoking accessors.
#include "adamic.h"
#include <string.h>

static bool adamic_property_name_equal(const adamic_string *key, const char *name) {
 size_t length = strlen(name);
 return length == key->length && memcmp(name, key->bytes, length) == 0;
}

bool adamic_object_has_property(const adamic_object *object, const adamic_string *key) {
 if (adamic_object_has(object, key)) return true;
 const adamic_methods *methods = object->shape->methods;
 if (methods != NULL) {
  for (size_t index = 0; index < methods->count; index++) {
   if (adamic_property_name_equal(key, methods->names[index])) return true;
  }
 }
 for (const adamic_class *class = object->class; class != NULL; class = class->base) {
  for (size_t index = 0; index < class->accessor_count; index++) {
   if (adamic_property_name_equal(key, class->accessors[index].name)) return true;
  }
 }
 return false;
}
