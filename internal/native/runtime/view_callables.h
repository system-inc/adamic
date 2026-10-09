#ifndef ADAMIC_VIEW_CALLABLES_H
#define ADAMIC_VIEW_CALLABLES_H
#include "view_unions_mixed.h"
adamic_view_union_value adamic_view_callable_candidate(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method, const char *expression, const char *expected);
adamic_closure *adamic_view_callable(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method, const char *expression);
adamic_closure *adamic_view_callable_typed(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method, const char *expression, const char *expected, bool absent, bool optional);
#endif
