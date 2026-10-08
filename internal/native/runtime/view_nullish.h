#ifndef ADAMIC_VIEW_NULLISH_H
#define ADAMIC_VIEW_NULLISH_H
#include "adamic.h"
adamic_value *adamic_object_read_contract(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner);
adamic_heap *adamic_object_nullish_view(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned int kinds, bool null_allowed, bool undefined_allowed, bool absent, bool optional, const char *expression, const char *expected);
_Noreturn void adamic_nullish_failure(const char *expression, const char *expected, const adamic_heap *value);
#endif
