// New separate helper: union reads keep the actual slot owner's layout, including
// inherited constructor storage. No existing runtime operation is rewritten.
#include "adamic.h"
#include <string.h>

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

// Generated scalar layouts are handled by the emitter. These are runtime-only
// scalar layouts with exact field representations. Unknown scalars fail closed.
adamic_heap *adamic_union_runtime_field(const adamic_object *owner, const adamic_value *slot) {
    size_t index = 0;
    while (index < owner->shape->count && &owner->slots[index] != slot) index++;
    if (index == owner->shape->count) {
        static const char message[] = "union field slot is outside its layout";
        adamic_panic(message, sizeof message - 1);
    }
    if (owner->shape->references[index]) return adamic_retain(slot->reference);
    const char *const *names = owner->shape->names;
    size_t count = owner->shape->count;
    bool number = false, boolean = false;
    if (count == 12 && strcmp(names[0], "__program") == 0 && strcmp(names[1], "lastIndex") == 0 && strcmp(names[11], "dotAll") == 0) {
        number = index == 1;
        boolean = index >= 4;
    } else if (count == 4 && strcmp(names[0], "kind") == 0 && strcmp(names[1], "type") == 0 && strcmp(names[2], "size") == 0 && strcmp(names[3], "symbolicLink") == 0) {
        number = index == 2;
        boolean = index == 3;
    } else if (count == 2 && strcmp(names[0], "done") == 0 && strcmp(names[1], "value") == 0) {
        boolean = index == 0;
    } else if (count == 3 && strcmp(names[0], "nodeKind") == 0 && strcmp(names[1], "symbolName") == 0 && strcmp(names[2], "type") == 0) {
        number = index == 0;
    }
    if (number) return adamic_box_number(slot->number);
    if (boolean) return slot->boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
    static const char message[] = "union field view of a runtime scalar without representation metadata";
    adamic_panic(message, sizeof message - 1);
    return NULL;
}
