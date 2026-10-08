#ifndef ADAMIC_ARRAY_UNION_H
#define ADAMIC_ARRAY_UNION_H
#include "adamic.h"
void adamic_array_sort_union(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context);
#endif
