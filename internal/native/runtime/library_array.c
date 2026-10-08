// Array searches with an optional starting index. Stage 0 arrays are dense.
#include "adamic.h"

// Reuse the established equality, with a borrowed one-element view and no allocation.
static bool array_search_equal(const adamic_array *array, size_t index, adamic_value value, enum adamic_equality equality, bool same_value_zero) {
	adamic_array view = *array;
	view.elements = &array->elements[index];
	view.length = 1;
	return adamic_array_index_of(&view, value, equality, same_value_zero) == 0;
}

double adamic_array_search_from(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool same_value_zero, double from, bool has_from, bool last) {
	if (array->length == 0) return -1;
	double length = (double)array->length;
	from = has_from ? (isnan(from) ? 0 : trunc(from)) : (last ? length - 1 : 0);
	if (from < 0) from += length;
	if (last) {
		if (from < 0) return -1;
		if (from >= length) from = length - 1;
		for (size_t index = (size_t)from + 1; index-- > 0;) {
			if (array_search_equal(array, index, value, equality, false)) return (double)index;
		}
	} else {
		if (from >= length) return -1;
		if (from < 0) from = 0;
		for (size_t index = (size_t)from; index < array->length; index++) {
			if (array_search_equal(array, index, value, equality, same_value_zero)) return (double)index;
		}
	}
	return -1;
}

// Only statically known homogeneous layers reach this helper. Inner joins always use comma.
adamic_string *adamic_array_join_nested(const adamic_array *array, const adamic_string *separator, enum adamic_join kind, size_t depth) {
    if (array == NULL) {
        static const adamic_array empty = {0};
        return adamic_array_join(&empty, separator, kind);
    }
    if (depth == 0) return adamic_array_join(array, separator, kind);
    static const adamic_string comma = ADAMIC_STRING(",");
    adamic_array *strings = adamic_array_new_typed(array->length, true, &adamic_json_string_schema);
    for (size_t index = 0; index < array->length; index++) {
        adamic_string *inner = adamic_array_join_nested(array->elements[index].reference, &comma, kind, depth - 1);
        adamic_array_push(strings, (adamic_value){.reference = inner});
    }
    adamic_string *result = adamic_array_join(strings, separator, adamic_join_strings);
    adamic_release(strings);
    return result;
}
