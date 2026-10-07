#include "adamic.h"
#include "view_callables.h"
#include <string.h>

adamic_closure *adamic_view_callable(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method, const char *expression) {
    if (object != NULL) {
        adamic_value *own = adamic_object_optional_field(object, name, cache);
        if (own == NULL && object->shape->methods != NULL) {
            const adamic_methods *methods = object->shape->methods;
            for (size_t index = 0; index < methods->count; index++) {
                if (strcmp(methods->names[index], name) == 0) { *method = methods->code[index]; return NULL; }
            }
        }
    }
    return adamic_object_view(object, name, cache, 8, "function", expression).reference;
}
