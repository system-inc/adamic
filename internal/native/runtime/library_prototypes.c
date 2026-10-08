// Restricted V8 JSObject::SetPrototype, src/objects/js-objects.cc.
// The initial proof admits only empty prototypes. Every link is counted.
#include "library_prototypes.h"
#include "library_errors.h"
static const adamic_shape empty_shape = {0, NULL, NULL, NULL};
// No source program may mutate this immortal intrinsic identity.
static adamic_object object_prototype = {.heap = {.kind = adamic_kind_object}, .shape = &empty_shape};
adamic_object *adamic_object_intrinsic_prototype(void) { return &object_prototype; }
adamic_object *adamic_object_get_prototype(adamic_object *object) {
 return adamic_retain(object->prototype != NULL ? object->prototype : &object_prototype);
}
adamic_object *adamic_object_create_prototype(adamic_object *prototype) {
 adamic_object *object = adamic_object_new(&empty_shape);
 object->prototype = adamic_retain(prototype);
 return object;
}
adamic_object *adamic_object_set_prototype(adamic_object *object, adamic_object *prototype) {
 adamic_object *previous = object->prototype != NULL ? object->prototype : &object_prototype;
 if (previous == prototype) { return adamic_retain(object); }
 if (object->nonextensible) {
  adamic_library_throw("TypeError", "#<Object> is not extensible", sizeof "#<Object> is not extensible"-1);
  return NULL;
 }
 for (adamic_object *cursor=prototype; cursor!=NULL; cursor=cursor->prototype) {
  if (cursor==object) {
   adamic_library_throw("TypeError", "Cyclic __proto__ value", sizeof "Cyclic __proto__ value"-1);
   return NULL;
  }
 }
 adamic_retain(prototype);
 adamic_release(object->prototype);
 object->prototype=prototype;
 return adamic_retain(object);
}
