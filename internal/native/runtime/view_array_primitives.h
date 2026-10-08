#ifndef ADAMIC_VIEW_ARRAY_PRIMITIVES_H
#define ADAMIC_VIEW_ARRAY_PRIMITIVES_H
#include "view_unions_mixed.h"
// Borrow exactly one producer-certified slot; ownership remains with the array.
adamic_view_union_value adamic_view_primitive_array_snapshot(const adamic_array *array, double index, bool relative, const char *expression, const char *declared);
adamic_heap *adamic_view_primitive_array_box(adamic_view_union_value value);
#endif
