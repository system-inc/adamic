// Default string sort uses the same stable sorter as comparator sort. Undefined
// is never passed to the comparator; absent slots follow explicit undefined.
#include "adamic.h"

static bool string_sort_index(double index, size_t length) {
    return index >= 0 && index < (double)length && index < 4294967295.0 && index == trunc(index);
}

static int string_sort_key(adamic_value left, adamic_value right, void *context) {
    (void)context;
    return left.number < right.number ? -1 : left.number > right.number ? 1 : 0;
}

void adamic_array_sort_strings(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
    if (array->length < 2) return;
    if (array->sparse == NULL) {
        size_t index = 0;
        while (index < array->length && array->elements[index].reference != NULL) index++;
        if (index == array->length) {
            adamic_array_sort(array, compare, context);
            return;
        }
    }

    adamic_array *work = adamic_array_new(0, true);
    adamic_array *keys = NULL;
    size_t undefined = 0;
    if (array->sparse != NULL) {
        // Snapshot numeric keys, not the potentially 2^32-1-long hole range.
        // Gathering in index order preserves stability even after out-of-order writes.
        keys = adamic_map_keys(array->sparse);
        adamic_array_sort(keys, string_sort_key, NULL);
        for (size_t index = 0; index < keys->length; index++) {
            adamic_value key = keys->elements[index];
            if (!string_sort_index(key.number, array->length)) continue;
            adamic_value value = *adamic_map_get(array->sparse, key);
            if (value.reference == NULL) undefined++;
            else adamic_array_push(work, (adamic_value){.reference = adamic_retain(value.reference)});
        }
    } else {
        for (size_t index = 0; index < array->length; index++) {
            void *value = array->elements[index].reference;
            if (value == NULL) undefined++;
            else adamic_array_push(work, (adamic_value){.reference = adamic_retain(value)});
        }
    }
    adamic_array_sort(work, compare, context);
    if (adamic_thrown == NULL) {
        if (keys != NULL) {
            for (size_t index = 0; index < keys->length; index++) {
                adamic_value key = keys->elements[index];
                if (string_sort_index(key.number, array->length)) adamic_map_delete(array->sparse, key);
            }
            // Sparse capacity is the number of absent indexed slots. Non-index
            // properties survive, and each following set consumes one absent slot.
            array->capacity = array->length;
            for (size_t index = 0; index < work->length; index++) {
                adamic_array_holes_set(array, (double)index, (adamic_value){.reference = adamic_retain(work->elements[index].reference)});
            }
            for (size_t index = 0; index < undefined; index++) {
                adamic_array_holes_set(array, (double)(work->length + index), (adamic_value){.reference = NULL});
            }
        } else {
            for (size_t index = 0; index < array->length; index++) {
                void *old = array->elements[index].reference;
                array->elements[index].reference = index < work->length ? adamic_retain(work->elements[index].reference) : NULL;
                adamic_release(old);
            }
        }
    }
    adamic_release(keys);
    adamic_release(work);
}
