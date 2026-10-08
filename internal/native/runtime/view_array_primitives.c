#include "view_array_primitives.h"
#include <math.h>

static void primitive_array_failure(const char *expression, const char *declared) {
    adamic_view_union_value unknown = {adamic_view_union_unknown, {.reference = NULL}};
    (void)adamic_view_mixed_union_select(&unknown, NULL, 0, NULL, NULL, expression, declared);
}

adamic_view_union_value adamic_view_primitive_array_snapshot(const adamic_array *array, double index, bool relative, const char *expression, const char *declared) {
    if (array == NULL || array->heap.kind != adamic_kind_array) primitive_array_failure(expression, declared);
    unsigned char storage = array->element_kind;
    // Physical producer metadata, never the target union, authorizes decoding.
    if (array->references ? storage != 10 : storage != 1 && storage != 2 && storage != 7) primitive_array_failure(expression, declared);
    if (relative) { index = isnan(index) ? 0 : trunc(index); if (index < 0) index += (double)array->length; }
    const adamic_value *slot = adamic_array_holes_at(array, index);
    adamic_view_union_value value = {adamic_view_union_undefined, {.reference = NULL}};
    if (slot == NULL) return value;
    if (array->references) {
        if (slot->reference == &adamic_null) return (adamic_view_union_value){adamic_view_union_null, *slot};
        return adamic_view_union_heap(slot->reference);
    }
    if (storage == 1) return (adamic_view_union_value){adamic_view_union_number, *slot};
    if (storage == 2) return (adamic_view_union_value){adamic_view_union_boolean, *slot};
    adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
    if (number.present) return (adamic_view_union_value){adamic_view_union_number, {.number = number.number}};
    return value;
}

adamic_heap *adamic_view_primitive_array_box(adamic_view_union_value value) {
    if (value.kind == adamic_view_union_number) return adamic_box_number(value.payload.number);
    if (value.kind == adamic_view_union_boolean) return value.payload.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
    return adamic_retain(value.payload.reference);
}
