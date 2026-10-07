#include "adamic.h"
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void array_view_failure(const char *expression, const char *expected, const char *found) {
    size_t capacity = strlen(expression) + strlen(expected) + strlen(found) + 80;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "element read failed: %s expected %s, found %s", expression, expected, found);
    adamic_panic(message, (size_t)length);
}

static const char *array_storage_name(unsigned char storage) {
    return storage == 1 || storage == 7 ? "number" : storage == 2 ? "boolean" : storage == 3 ? "string" : storage == 4 ? "object" : storage == 5 ? "array" : storage == 6 ? "Map" : storage == 8 ? "function" : storage == 10 ? "heap pointers" : "uncertified storage";
}

void adamic_array_view_storage(adamic_array *array, unsigned char storage) {
    if (array->references) { storage = 10; }
    if (array->element_kind != 0 && array->element_kind != storage) { array_view_failure("<array allocation>", array_storage_name(storage), array_storage_name(array->element_kind)); }
    array->element_kind = storage;
}

void adamic_view_array_storage_check(const adamic_array *array, unsigned char storage, const char *expression) {
    unsigned char physical = storage == 3 || storage == 4 || storage == 5 || storage == 6 || storage == 8 || storage == 10 ? 10 : storage;
    if (array->element_kind != physical) { array_view_failure(expression, array_storage_name(storage), array_storage_name(array->element_kind)); }
    // A pointer byte cannot certify the original object/class/map signature.
    // Keep reference writes closed until the source element contract is known.
    if (storage == 4 || storage == 5 || storage == 6 || storage == 8 || storage == 9 || storage == 10) { array_view_failure(expression, array_storage_name(storage), "uncertified source element contract"); }
}

adamic_value *adamic_view_array_at(const adamic_array *array, double index, bool relative, bool undefined_allowed, unsigned char wanted, const char *expected, const char *expression, adamic_value *snapshot) {
    if (array == NULL) array_view_failure(expression, expected, "undefined");
    if (relative) {
        index = isnan(index) ? 0 : trunc(index);
        if (index < 0) index += (double)array->length;
    }

    adamic_value *slot = adamic_array_holes_at(array, index);
    if (slot == NULL) { return NULL; }
    unsigned char actual = array->element_kind;
    *snapshot = *slot;
    if (actual == 7) {
        adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
        if (!number.present) { if (!undefined_allowed) { array_view_failure(expression, expected, "undefined"); } return NULL; }
        actual = 1;
        snapshot->number = number.number;
    }
    if (array->references) {
        const adamic_heap *reference = slot->reference;
        if (reference == NULL) { if (!undefined_allowed) { array_view_failure(expression, expected, "nullish"); } return NULL; }
        if (reference->kind == adamic_kind_number) { actual = 1; snapshot->number = ((const adamic_number_box *)reference)->number; }
        else if (reference->kind == adamic_kind_boolean) { actual = 2; snapshot->boolean = ((const adamic_boolean_box *)reference)->boolean; }
        else { actual = reference->kind == adamic_kind_string ? 3 : reference->kind == adamic_kind_object ? 4 : reference->kind == adamic_kind_array ? 5 : reference->kind == adamic_kind_map ? 6 : reference->kind == adamic_kind_closure ? 8 : 0; }
    }
    if (actual != wanted && !(wanted == 7 && actual == 1) && !(wanted == 10 && array->element_kind == 10)) { array_view_failure(expression, expected, array_storage_name(actual)); }
    if (wanted == 7 && actual == 1) { snapshot->number = adamic_maybe_number_pack((adamic_maybe_number){true, snapshot->number}); }
    if (wanted == 10) { *snapshot = *slot; }
    return snapshot;
}

void adamic_view_array_missing(const char *expression, const char *expected) { array_view_failure(expression, expected, "undefined"); }

// Array stringification consumes every present element, checking each at its
// read. The null array representation spells undefined without dereferencing it.
adamic_string *adamic_view_array_string(const adamic_array *array, const adamic_string *separator, bool checked, unsigned char wanted, bool undefined_allowed, const char *expected, const char *expression, size_t allowed_count, const adamic_value *allowed) {
    if (array == NULL) {
        adamic_string *result = adamic_string_allocate(9);
        memcpy((char *)result->bytes, "undefined", 9);
        return result;
    }
    adamic_array *parts = adamic_array_new(array->length, true);
    parts->length = array->length;
    for (size_t index = 0; index < array->length; index++) {
        adamic_value snapshot;
        adamic_value *slot = checked ? adamic_view_array_at(array, (double)index, false, undefined_allowed, wanted, expected, expression, &snapshot) : adamic_array_holes_at(array, (double)index);
        if (slot != NULL && checked && allowed_count != 0) {
            bool accepted = false;
            for (size_t literal = 0; literal < allowed_count; literal++) {
                accepted = accepted || (wanted == 1 ? slot->number == allowed[literal].number : wanted == 2 ? slot->boolean == allowed[literal].boolean : wanted == 3 && adamic_string_equal(slot->reference, allowed[literal].reference));
            }
            if (!accepted) adamic_view_literal_failure(expression, expected, wanted, *slot);
        }
        adamic_string *text = NULL;
        if (slot != NULL && wanted == 7) {
            adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
            if (number.present) text = adamic_string_from_number(number.number);
        } else if (slot != NULL && wanted == 1) text = adamic_string_from_number(slot->number);
        else if (slot != NULL && wanted == 3) text = adamic_retain(slot->reference);
        else if (slot != NULL && wanted == 2) {
            const char *word = slot->boolean ? "true" : "false";
            size_t size = strlen(word);
            text = adamic_string_allocate(size);
            memcpy((char *)text->bytes, word, size);
        }
        parts->elements[index].reference = text;
    }
    adamic_string *result = adamic_array_join(parts, separator, adamic_join_strings);
    adamic_release(parts);
    return result;
}

static adamic_string *array_view_sort_string(adamic_value value, unsigned char element) {
    if (element == 1) return adamic_string_from_number(value.number);
    if (element == 7) return adamic_string_from_number(adamic_maybe_number_unpack(value.number).number);
    if (element == 3) return adamic_retain(value.reference);
    const char *word = value.boolean ? "true" : "false";
    size_t length = strlen(word);
    adamic_string *result = adamic_string_allocate(length);
    memcpy((char *)result->bytes, word, length);
    return result;
}

// Explicitly undefined comparers use ECMAScript's UTF-16 string ordering.
// The existing undefined-last sort handles missing elements before this hook.
int adamic_view_array_default_compare(adamic_value left, adamic_value right, void *context) {
    unsigned char element = (unsigned char)(uintptr_t)context;
    adamic_string *a = array_view_sort_string(left, element);
    adamic_string *b = array_view_sort_string(right, element);
    int result = adamic_string_compare(a, b);
    adamic_release(a);
    adamic_release(b);
    return result;
}

// The selected value has already been checked and retained. Commit removal only
// afterward, preserving sparse ownership and missing-slot accounting.
void adamic_view_array_pop_commit(adamic_array *array) {
    if (array->length == 0) return;
    if (array->sparse != NULL) {
        double index = (double)array->length - 1;
        if (adamic_array_holes_at(array, index) == NULL) { array->capacity--; }
        else { adamic_map_delete(array->sparse, (adamic_value){.number = index}); }
        array->length--;
        return;
    }
    array->length--;
    if (array->references) { adamic_release(array->elements[array->length].reference); }
}

void adamic_view_array_push(adamic_array *array, adamic_value value) {
    if (array->sparse != NULL) { adamic_array_holes_set(array, (double)array->length, value); }
    else { adamic_array_push(array, value); }
}

adamic_array *adamic_view_array_slice(const adamic_array *array, double start, double end, bool has_end) {
    if (array->sparse == NULL) return adamic_array_slice(array, start, end, has_end);
    double length = (double)array->length;
    start = isnan(start) ? 0 : trunc(start);
    start = start < 0 ? fmax(length + start, 0) : fmin(start, length);
    end = !has_end ? length : isnan(end) ? 0 : trunc(end);
    end = end < 0 ? fmax(length + end, 0) : fmin(end, length);
    if (end < start) end = start;
    adamic_array *result = adamic_array_holes(end - start, array->references);
    if (result == NULL) return NULL;
    result->element_kind = array->element_kind;
    for (size_t at = 0; at < array->sparse->used; at++) {
        const adamic_map_entry *entry = &array->sparse->entries[at];
        double index = entry->key.number;
        if (entry->deleted || !(index >= start && index < end) || index != trunc(index)) continue;
        adamic_value value = entry->value;
        if (array->references) adamic_retain(value.reference);
        adamic_array_holes_set(result, index - start, value);
    }
    return result;

}
