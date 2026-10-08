// New separate helper: union reads keep the actual slot owner's layout, including
// inherited constructor storage. No existing runtime operation is rewritten.
#include "adamic.h"

const adamic_object *adamic_union_slot_owner(const adamic_object *object, const adamic_value *slot) {
    while (object != NULL) {
        for (size_t index = 0; index < object->shape->count; index++) {
            if (&object->slots[index] == slot) return object;
        }
        if (object->class == NULL || !object->class->is_static || object->class->static_parent == 0) break;
        object = object->slots[object->class->static_parent - 1].reference;
    }
    static const char message[] = "union field slot has no owning layout";
    adamic_panic(message, sizeof message - 1);
    return NULL;
}

// Runtime field representations come from the owning shape, independently of
// its names or ordering. Generated packed scalar layouts are handled by the emitter.
adamic_heap *adamic_union_runtime_field(const adamic_object *owner, const adamic_value *slot) {
    size_t index = 0;
    while (index < owner->shape->count && &owner->slots[index] != slot) index++;
    if (index == owner->shape->count) {
        static const char message[] = "union field slot is outside its layout";
        adamic_panic(message, sizeof message - 1);
    }
    if (owner->shape->kinds != NULL) {
        switch (owner->shape->kinds[index]) {
        case adamic_field_reference: return adamic_retain(slot->reference);
        case adamic_field_number: {
            adamic_maybe_number value = adamic_maybe_number_unpack(slot->number);
            return value.present ? adamic_box_number(value.number) : NULL;
        }
        case adamic_field_boolean: {
            // Exhausted collection iterators use the reserved undefined word for
            // either scalar kind. Ordinary booleans and packed pairs use 0/1/2.
            if (!adamic_maybe_number_unpack(slot->number).present) return NULL;
            adamic_maybe_boolean value = adamic_maybe_boolean_unpack(slot->maybe_boolean);
            return !value.present ? NULL : value.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
        }
        }
    }
    static const char message[] = "union field has an invalid storage kind";
    adamic_panic(message, sizeof message - 1);
    return NULL;
}
