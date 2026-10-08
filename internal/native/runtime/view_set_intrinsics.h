#ifndef ADAMIC_VIEW_SET_INTRINSICS_H
#define ADAMIC_VIEW_SET_INTRINSICS_H
#include "view_callables_contract.h"
adamic_map *adamic_view_set_producer(adamic_map *set, unsigned char element, unsigned char callable_element);
adamic_method adamic_view_set_method(const adamic_map *set, const char *name, const char *expression, const char *expected);
const adamic_callable_signature *adamic_view_set_signature(adamic_method method);
adamic_value adamic_view_set_field(const adamic_map *set, const char *name, unsigned char wanted, const char *expected, const char *expression, bool absent);
#endif
