#include "adamic.h"
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
    return storage == 1 || storage == 7 ? "number" : storage == 2 ? "boolean" : storage == 3 ? "string" : storage == 4 ? "object" : storage == 5 ? "array" : storage == 6 ? "Map" : storage == 8 ? "function" : "uncertified storage";
}

void adamic_array_view_storage(adamic_array *array, unsigned char storage) {
    if (array->view.storage != 0 && array->view.storage != storage) { array_view_failure("<array allocation>", array_storage_name(storage), array_storage_name(array->view.storage)); }
    array->view.storage = storage;
}

void adamic_view_array_storage_check(const adamic_array *array, unsigned char storage, const char *expression) {
    if (array->view.storage != storage) { array_view_failure(expression, array_storage_name(storage), array_storage_name(array->view.storage)); }
}

adamic_value *adamic_view_array_at(const adamic_array *array, double index, bool relative, bool undefined_allowed, unsigned char wanted, const char *expected, const char *expression, adamic_value *snapshot) {
    adamic_value *slot = relative ? adamic_array_at_relative(array, index) : adamic_array_at(array, index);
    if (slot == NULL) { return NULL; }
    unsigned char actual = array->view.storage;
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
    if (actual != wanted && !(wanted == 7 && actual == 1) && !(wanted == 10 && array->view.storage == 10)) { array_view_failure(expression, expected, array_storage_name(actual)); }
    if (wanted == 7 && actual == 1) { snapshot->number = adamic_maybe_number_pack((adamic_maybe_number){true, snapshot->number}); }
    if (wanted == 10) { *snapshot = *slot; }
    return snapshot;
}

void adamic_view_array_missing(const char *expression, const char *expected) { array_view_failure(expression, expected, "undefined"); }
