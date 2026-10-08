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
// its names or ordering. Generated scalar layouts are handled by the emitter.
adamic_heap *adamic_union_runtime_field(const adamic_object *owner, const adamic_value *slot) {
    size_t index = 0;
    while (index < owner->shape->count && &owner->slots[index] != slot) index++;
    if (index == owner->shape->count) {
        static const char message[] = "union field slot is outside its layout";
        adamic_panic(message, sizeof message - 1);
    }
    // Scalar classification depends on runtime/shape-field-kinds 17b5a053.
    // Until that dependency lands on main, never guess a scalar's representation.
    if (owner->shape->references[index]) return adamic_retain(slot->reference);
    static const char message[] = "union scalar field needs runtime/shape-field-kinds 17b5a053";
    adamic_panic(message, sizeof message - 1);
    return NULL;
}
