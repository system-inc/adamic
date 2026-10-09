// class_static.c: constructor objects inherit live data until an own write shadows it.
#include "adamic.h"

adamic_value *adamic_static_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
    adamic_value *slot = adamic_object_find(object, name, cache);
    size_t flag = object->class->static_flags[(size_t)(slot - object->slots)];
    if (flag != 0 && object->slots[flag - 1].number == 0 && object->class->static_parent != 0) {
        return adamic_static_field(object->slots[object->class->static_parent - 1].reference, name, cache);
    }
    return slot;
}

adamic_value *adamic_object_write_field(adamic_object *object, const char *name, adamic_slot_cache *cache) {
    if (object->class == NULL || !object->class->is_static) {
        adamic_value *slot = adamic_object_field(object, name, cache);
        size_t index = adamic_slot_index(object, slot);
        adamic_object_initialized(object)[index] = 1;
        adamic_object_present(object, index);
        return slot;
    }
    adamic_value *slot = adamic_object_find(object, name, cache);
    size_t index = adamic_slot_index(object, slot);
    adamic_object_initialized(object)[index] = 1;
    size_t flag = object->class->static_flags[index];
    if (flag != 0 && object->slots[flag - 1].number == 0) {
        double order = 0;
        for (size_t index = 0; index < object->shape->count; index++) {
            size_t previous = object->class->static_flags[index];
            if (previous != 0 && object->slots[previous - 1].number > order) { order = object->slots[previous - 1].number; }
        }
        object->slots[flag - 1].number = order + 1;
    }
    return slot;
}
