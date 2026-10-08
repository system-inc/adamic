#ifndef ADAMIC_ARRAY_MAYBE_BOOLEAN_H
#define ADAMIC_ARRAY_MAYBE_BOOLEAN_H
#include "adamic.h"
void adamic_array_sort_maybe_boolean(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context);
#endif
