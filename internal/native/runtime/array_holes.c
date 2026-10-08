// Length-form arrays use an ordered numeric map for present slots. Construction
// is constant space even at the largest valid length. Dense arrays keep their
// existing storage and lookup functions.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <stdlib.h>

static const char *const range_names[] = {"name", "message", "code"};
static const bool range_references[] = {true, true, true};
static const adamic_shape range_shape = {3, range_names, range_references, NULL};
static const adamic_class range_class = {.count = 3};
static adamic_string range_name = ADAMIC_STRING("RangeError");
static adamic_string range_message = ADAMIC_STRING("Invalid array length");

bool adamic_array_is_range_error(const adamic_object *value) {
    return value != NULL && value->class == &range_class;
}

static bool array_length_valid(double length) {
    if (!(length >= 0) || length > 4294967295.0 || length != trunc(length)) {
        adamic_thrown = adamic_object_new(&range_shape);
        adamic_thrown->class = &range_class;
        adamic_thrown->slots[0].reference = adamic_retain(&range_name);
        adamic_thrown->slots[1].reference = adamic_retain(&range_message);
        return false;
    }
    return true;
}

adamic_array *adamic_array_holes(double length, bool references) {
    if (!array_length_valid(length)) return NULL;
    adamic_array *array = adamic_array_new(0, references);
    array->length = (size_t)length;
    array->capacity = (size_t)length; // Sparse capacity counts absent indexed slots.
    array->sparse = adamic_map_new(false, references);
    return array;
}


static bool array_index(double index) {
    return index >= 0 && index < 4294967295.0 && index == trunc(index);
}

adamic_value *adamic_array_holes_at(const adamic_array *array, double index) {
    if (array->sparse == NULL) return adamic_array_at(array, index);
    if (array_index(index) && (index >= (double)array->length || array->capacity == array->length)) return NULL;
    return adamic_map_get(array->sparse, (adamic_value){.number = index});
}

static void array_sparse(adamic_array *array) {
    if (array->sparse != NULL) return;
    array->sparse = adamic_map_new(false, array->references);
    for (size_t index = 0; index < array->length; index++) {
        // Ownership moves from the dense slots to the map.
        adamic_map_set(array->sparse, (adamic_value){.number = (double)index}, array->elements[index]);
    }
    free(array->elements);
    array->elements = NULL;
    array->capacity = 0;
}

void adamic_array_holes_set(adamic_array *array, double index, adamic_value value) {
    if (array->sparse == NULL && array_index(index) && index < (double)array->length) {
        adamic_array_set(array, index, value);
        return;
    }
    array_sparse(array);
    if (array_index(index)) {
        if (index >= (double)array->length) {
            size_t length = (size_t)index + 1;
            array->capacity += length - array->length;
            array->length = length;
        }
        if (adamic_map_get(array->sparse, (adamic_value){.number = index}) == NULL) {
            array->capacity--;
        }
    }
    adamic_map_set(array->sparse, (adamic_value){.number = index}, value);
}

adamic_string *adamic_array_holes_join(const adamic_array *array, const adamic_string *separator, enum adamic_join kind) {
    if (array->sparse == NULL) return adamic_array_join(array, separator, kind);
    static adamic_string hole_text = ADAMIC_STRING("");
    static adamic_string undefined_text = ADAMIC_STRING("");
    static adamic_string true_text = ADAMIC_STRING("true");
    static adamic_string false_text = ADAMIC_STRING("false");
    adamic_array *strings = adamic_array_new(0, true);
    for (size_t index = 0; index < array->length; index++) {
        adamic_value *slot = adamic_array_holes_at(array, (double)index);
        adamic_string *text = &hole_text;
        if (slot != NULL) {
            switch (kind) {
            case adamic_join_numbers: text = adamic_string_from_number(slot->number); break;
            case adamic_join_maybe_numbers: {
                adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
                text = number.present ? adamic_string_from_number(number.number) : &undefined_text;
                break;
            }
            case adamic_join_booleans: text = slot->boolean ? &true_text : &false_text; break;
            case adamic_join_strings: text = slot->reference == NULL ? &undefined_text : adamic_retain(slot->reference); break;
            }
        }
        adamic_array_push(strings, (adamic_value){.reference = text});
    }
    adamic_string *result = adamic_array_join(strings, separator, adamic_join_strings);
    adamic_release(strings);
    return result;
}

static bool sparse_equal(const adamic_array *array, double index, adamic_value value, enum adamic_equality equality, bool includes) {
    adamic_value *slot = adamic_array_holes_at(array, index);
    if (slot == NULL) {
        if (!includes) return false;
        if (equality == adamic_equal_maybe_numbers) return !adamic_maybe_number_unpack(value.number).present;
        return (equality == adamic_equal_strings || equality == adamic_equal_identity) && value.reference == NULL;
    }
    adamic_array view = *array;
    view.sparse = NULL;
    view.elements = slot;
    view.length = 1;
    return adamic_array_index_of(&view, value, equality, includes) == 0;
}

double adamic_array_holes_search_from(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool includes, double from, bool has_from, bool last) {
    if (array->sparse == NULL) return adamic_array_search_from(array, value, equality, includes, from, has_from, last);
    if (array->length == 0) return -1;
    double length = (double)array->length;
    from = has_from ? (isnan(from) ? 0 : trunc(from)) : (last ? length - 1 : 0);
    if (from < 0) from += length;
    if (last) {
        if (from < 0) return -1;
        if (from >= length) from = length - 1;
        for (size_t index = (size_t)from + 1; index-- > 0;) {
            if (sparse_equal(array, (double)index, value, equality, false)) return (double)index;
        }
    } else {
        if (from >= length) return -1;
        if (from < 0) from = 0;
        for (size_t index = (size_t)from; index < array->length; index++) {
            if (sparse_equal(array, (double)index, value, equality, includes)) return (double)index;
        }
    }
    return -1;
}

void adamic_array_holes_set_length(adamic_array *array, double length) {
    if (!array_length_valid(length)) return;
    array_sparse(array);
    adamic_map *slots = array->sparse;
    size_t present = 0;
    slots->iterating++;
    for (size_t index = 0; index < slots->used; index++) {
        adamic_map_entry *entry = &slots->entries[index];
        if (entry->deleted || !array_index(entry->key.number)) continue;
        if (entry->key.number >= length) {
            adamic_map_delete(slots, entry->key);
        } else {
            present++;
        }
    }
    slots->iterating--;
    array->length = (size_t)length;
    array->capacity = (size_t)length - present;
}
