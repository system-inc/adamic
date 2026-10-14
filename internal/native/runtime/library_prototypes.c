// Restricted V8 JSObject::SetPrototype, src/objects/js-objects.cc.
// The initial proof admits only empty prototypes. Every link is counted.
#include "library_prototypes.h"
#include <string.h>
static void prototype_throw(const char *message, size_t length) {
 adamic_string *text = adamic_string_allocate(length);
 memcpy((char *)text->bytes, message, length);
 adamic_object *error = adamic_error_new(text);
 adamic_release(text);
 adamic_release(error->slots[0].reference);
 error->slots[0].reference = adamic_string_allocate(9);
 memcpy((char *)((adamic_string *)error->slots[0].reference)->bytes, "TypeError", 9);
 adamic_release(adamic_thrown);
 adamic_thrown = error;
}
static const adamic_shape empty_shape = {0, NULL, NULL, NULL};
// No source program may mutate this immortal intrinsic identity.
static adamic_object object_prototype = {.heap = {.kind = adamic_kind_object}, .shape = &empty_shape};
adamic_object *adamic_object_intrinsic_prototype(void) { return &object_prototype; }
adamic_object *adamic_object_get_prototype(adamic_object *object) {
 return adamic_retain(object->prototype != NULL ? object->prototype : (object->null_prototype ? NULL : &object_prototype));
}
adamic_object *adamic_object_create_prototype(adamic_object *prototype) {
 adamic_object *object = adamic_object_new(&empty_shape);
 object->prototype = adamic_retain(prototype);
 return object;
}
adamic_object *adamic_object_set_prototype(adamic_object *object, adamic_object *prototype) {
 adamic_object *previous = object->prototype != NULL ? object->prototype : (object->null_prototype ? NULL : &object_prototype);
 if (previous == prototype) { return adamic_retain(object); }
 #ifdef ADAMIC_OBJECT_INTEGRITY_H
 if (!adamic_object_is_extensible(&object->heap)) {
#else
 if (object->frozen) {
#endif
  prototype_throw("#<Object> is not extensible", sizeof "#<Object> is not extensible"-1);
  return NULL;
 }
 for (adamic_object *cursor=prototype; cursor!=NULL; cursor=cursor->prototype) {
  if (cursor==object) {
   prototype_throw("Cyclic __proto__ value", sizeof "Cyclic __proto__ value"-1);
   return NULL;
  }
 }
 adamic_retain(prototype);
 adamic_release(object->prototype);
 object->prototype=prototype;
 object->null_prototype=prototype==NULL;
 return adamic_retain(object);
}
