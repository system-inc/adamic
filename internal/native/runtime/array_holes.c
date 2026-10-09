// Length-form arrays use an ordered numeric map for present slots. Construction
// is constant space even at the largest valid length. Dense arrays keep their
// existing storage and lookup functions.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"

static const char *const range_names[] = {"name", "message", "code"};
static const bool range_references[] = {true, true, true};
static const adamic_shape range_shape = {3, range_names, range_references, NULL};
static const adamic_class range_class = {.count = 3};
static adamic_string range_name = ADAMIC_STRING("RangeError");
static adamic_string range_message = ADAMIC_STRING("Invalid array length");

bool adamic_array_is_range_error(const adamic_object *value) {
    return value != NULL && value->class == &range_class;
}

adamic_array *adamic_array_holes(double length, bool references) {
    if (!(length >= 0) || length > 4294967295.0 || length != trunc(length)) {
        adamic_thrown = adamic_object_new(&range_shape);
        adamic_thrown->class = &range_class;
        adamic_thrown->slots[0].reference = adamic_retain(&range_name);
        adamic_thrown->slots[1].reference = adamic_retain(&range_message);
        return NULL;
    }
    adamic_array *array = adamic_array_new(0, references);
    array->length = (size_t)length;
    array->sparse = adamic_map_new(false, references);
    return array;
}

adamic_value *adamic_array_holes_at(const adamic_array *array, double index) {
    if (array->sparse == NULL) return adamic_array_at(array, index);
    if (!(index >= 0) || index >= (double)array->length || index != trunc(index)) return NULL;
    return adamic_map_get(array->sparse, (adamic_value){.number = index});
}

void adamic_array_holes_set(adamic_array *array, double index, adamic_value value) {
    if (array->sparse == NULL) { adamic_array_set(array, index, value); return; }
    if (!(index >= 0) || index >= (double)array->length || index != trunc(index)) {
        static const char message[] = "holey array writes outside existing length are not implemented";
        adamic_panic(message, sizeof message - 1);
    }
    adamic_map_set(array->sparse, (adamic_value){.number = index}, value);
}
