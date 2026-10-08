#include "adamic.h"

bool adamic_census_to_boolean(const adamic_heap *value) {
    if (value == NULL || value == &adamic_null) return false;
    switch (value->kind) {
    case adamic_kind_null: return false;
    case adamic_kind_number: {
        double number = ((const adamic_number_box *)value)->number;
        return number != 0.0 && !isnan(number);
    }
    case adamic_kind_boolean:
        return ((const adamic_boolean_box *)value)->boolean;
    case adamic_kind_string:
        return ((const adamic_string *)value)->length != 0;
    default:
        return true;
    }
}
