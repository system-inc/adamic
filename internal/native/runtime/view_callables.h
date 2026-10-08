#ifndef ADAMIC_VIEW_CALLABLES_H
#define ADAMIC_VIEW_CALLABLES_H
adamic_closure *adamic_view_callable(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method, const char *expression);
adamic_closure *adamic_view_callable_typed(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method, const char *expression, const char *expected, bool absent, bool optional);
#endif
