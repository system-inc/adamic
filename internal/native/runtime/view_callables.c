#include "adamic.h"
#include "view_callables.h"
#include "view_set_intrinsics.h"
#include <string.h>

adamic_closure *adamic_view_callable_typed(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method, const char *expression, const char *expected, bool absent, bool optional) {
    if (object != NULL && object->heap.kind == adamic_kind_map) {
        *method = adamic_view_set_method((const adamic_map *)object, name, expression, expected);
        return NULL;
    }
    if (object != NULL) {
        adamic_value *own = adamic_object_optional_field(object, name, cache);
        if (own == NULL && object->shape->methods != NULL) {
            const adamic_methods *methods = object->shape->methods;
            for (size_t index = 0; index < methods->count; index++) {
                if (strcmp(methods->names[index], name) == 0) { *method = methods->code[index]; return NULL; }
            }
        }
    }
    return adamic_object_optional_view(object, name, cache, 8, expected, expression, absent, optional).reference;
}

adamic_closure *adamic_view_callable(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method, const char *expression) {
 return adamic_view_callable_typed(object,name,cache,method,expression,"function",false,false);
}
