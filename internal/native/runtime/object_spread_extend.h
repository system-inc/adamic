#ifndef ADAMIC_OBJECT_SPREAD_EXTEND_H
#define ADAMIC_OBJECT_SPREAD_EXTEND_H
#include "adamic.h"
adamic_object *adamic_object_spread_extend(const adamic_object *snapshot, const adamic_shape *overrides);
const adamic_shape *adamic_object_spread_field_shape(const adamic_shape *shape, const char *name);
#endif
