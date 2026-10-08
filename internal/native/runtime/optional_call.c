// Optional callable lookup, kept separate from required method lookup.
#include "adamic.h"
#include <string.h>

adamic_closure *adamic_object_optional_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method) {
    const adamic_shape *shape = object->shape;
    const size_t methods = shape->methods == NULL ? 0 : shape->methods->count;
    *method = NULL;
    if (cache->shape != shape) {
        cache->index = shape->count + methods;
        for (size_t index = 0; index < shape->count; index++) {
            if (strcmp(shape->names[index], name) == 0) { cache->index = index; break; }
        }
        if (cache->index == shape->count + methods) {
            for (size_t index = 0; index < methods; index++) {
                if (strcmp(shape->methods->names[index], name) == 0) { cache->index = shape->count + index; break; }
            }
        }
        cache->shape = shape;
    }
    if (cache->index < shape->count) { return object->slots[cache->index].reference; }
    if (cache->index < shape->count + methods) { *method = shape->methods->code[cache->index - shape->count]; }
    return NULL;
}
