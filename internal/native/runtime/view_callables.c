#include "adamic.h"
#include "view_callables.h"
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "view_unions_mixed.h"

adamic_closure *adamic_view_callable_typed(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method, const char *expression, const char *expected, bool absent, bool optional) {
    if (object != NULL) {
        adamic_value *own = adamic_object_optional_field(object, name, cache);
        if (own == NULL && object->shape->methods != NULL) {
            const adamic_methods *methods = object->shape->methods;
            for (size_t index = 0; index < methods->count; index++) {
                if (strcmp(methods->names[index], name) == 0) { *method = methods->code[index]; return NULL; }
            }
        }
    }
    if (object == NULL && optional) return NULL;
    adamic_view_union_value value = adamic_object_view_union_snapshot(object, name, cache, expression, expected, absent);
    if (value.kind == adamic_view_union_function) return value.payload.reference;
    if (value.kind == adamic_view_union_undefined && absent) return NULL;
    const char *found = adamic_view_union_kind_name(value.kind);
    size_t capacity = strlen(expression) + 2 * strlen(expected) + strlen(found) + 100;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s is not a %s; expected %s, found %s", expression, expected, expected, found);
    adamic_panic(message, (size_t)length);
}

adamic_closure *adamic_view_callable(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method, const char *expression) {
 return adamic_view_callable_typed(object,name,cache,method,expression,"function",false,false);
}
