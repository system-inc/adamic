#ifndef ADAMIC_LIBRARY_PROTOTYPES_H
#define ADAMIC_LIBRARY_PROTOTYPES_H
#include "adamic.h"
adamic_object *adamic_object_intrinsic_prototype(void);
adamic_object *adamic_object_create_prototype(adamic_object *prototype);
adamic_object *adamic_object_set_prototype(adamic_object *object, adamic_object *prototype);
adamic_object *adamic_object_get_prototype(adamic_object *object);
#endif
