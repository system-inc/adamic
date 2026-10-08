#include "view_dictionaries.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void dictionary_failure(const char *expression, const char *declared, adamic_view_union_kind kind) {
    const char *found = adamic_view_union_kind_name(kind);
    size_t capacity = strlen(expression) + strlen(declared) + strlen(found) + 96;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s; expected %s, found %s", expression, declared, found);
    adamic_panic(message, (size_t)length);
}

adamic_view_dictionary_result adamic_view_dictionary_read(const adamic_record *record, bool boxed_storage, const adamic_string *key, unsigned int kinds, size_t child_contract, const char *expression, const char *declared) {
    // Unknown source representation refuses before any record slot is followed.
    if (!boxed_storage || record == NULL || key == NULL) {
        dictionary_failure(expression, declared, adamic_view_union_unknown);
    }
    // Preserve record.c's own-key prototype refusal. No second hash table or
    // readiness bitmap is introduced. Each invocation observes the current slot.
    const adamic_value *slot = adamic_record_get(record, key);
    const adamic_heap *boxed = slot == NULL ? NULL : slot->reference;
    adamic_view_union_value value = {adamic_view_union_undefined, {.reference = NULL}};
    if (boxed != NULL) {
        value.payload.reference = (void *)boxed;
        switch (boxed->kind) {
        case adamic_kind_number:
            value.kind = adamic_view_union_number;
            value.payload.number = ((const adamic_number_box *)boxed)->number;
            break;
        case adamic_kind_boolean:
            value.kind = adamic_view_union_boolean;
            value.payload.boolean = ((const adamic_boolean_box *)boxed)->boolean;
            break;
        case adamic_kind_string: value.kind = adamic_view_union_string; break;
        case adamic_kind_object: value.kind = adamic_view_union_object; break;
        case adamic_kind_array: value.kind = adamic_view_union_array; break;
        case adamic_kind_map: value.kind = adamic_view_union_map; break;
        case adamic_kind_closure: value.kind = adamic_view_union_function; break;
        default: value.kind = adamic_view_union_unknown; break;
        }
    }
    if (value.kind == adamic_view_union_unknown || (kinds & (1u << value.kind)) == 0) {
        dictionary_failure(expression, declared, value.kind);
    }
    bool reference = value.kind == adamic_view_union_object || value.kind == adamic_view_union_array || value.kind == adamic_view_union_map || value.kind == adamic_view_union_function;
    if (reference && child_contract == 0) {
        dictionary_failure(expression, declared, adamic_view_union_unknown);
    }
    return (adamic_view_dictionary_result){value, reference ? child_contract : 0};
}

static adamic_view_union_value dictionary_boxed_value(const adamic_heap *boxed) {
    adamic_view_union_value value = {adamic_view_union_undefined, {.reference = NULL}};
    if (boxed == NULL) return value;
    value.payload.reference = (void *)boxed;
    switch (boxed->kind) {
    case adamic_kind_number: value.kind = adamic_view_union_number; value.payload.number = ((const adamic_number_box *)boxed)->number; break;
    case adamic_kind_boolean: value.kind = adamic_view_union_boolean; value.payload.boolean = ((const adamic_boolean_box *)boxed)->boolean; break;
    case adamic_kind_string: value.kind = adamic_view_union_string; break;
    case adamic_kind_object: value.kind = adamic_view_union_object; break;
    case adamic_kind_array: value.kind = adamic_view_union_array; break;
    case adamic_kind_map: value.kind = adamic_view_union_map; break;
    case adamic_kind_closure: value.kind = adamic_view_union_function; break;
    default: value.kind = adamic_view_union_unknown; break;
    }
    return value;
}

// Fixed-shape source objects keep their identity and existing storage. A target
// dictionary never reinterprets their scalar slots as a record table or boxes.
adamic_view_dictionary_result adamic_view_dictionary_source_read(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child_contract, const char *expression, const char *declared) {
    if (object == NULL || key == NULL) dictionary_failure(expression, declared, adamic_view_union_undefined);
    if (object->heap.kind != adamic_kind_object) dictionary_failure(expression, declared, dictionary_boxed_value(&object->heap).kind);
    if (adamic_record_is(object)) {
        const adamic_map *storage = object->slots[0].reference;
        unsigned char element = (unsigned char)object->slots[1].number;
        if (storage->reference_values) return adamic_view_dictionary_read(object, true, key, kinds, child_contract, expression, declared);
        const adamic_value *slot = adamic_record_get(object, key);
        adamic_view_union_value value = {adamic_view_union_undefined, {.reference = NULL}};
        if (slot != NULL) {
            if (element == 1) value = (adamic_view_union_value){adamic_view_union_number, *slot};
            else if (element == 2) value = (adamic_view_union_value){adamic_view_union_boolean, *slot};
            else if (element == 7) {
                adamic_maybe_number maybe = adamic_maybe_number_unpack(slot->number);
                if (maybe.present) value = (adamic_view_union_value){adamic_view_union_number, {.number = maybe.number}};
            } else value.kind = adamic_view_union_unknown;
        }
        if (value.kind == adamic_view_union_unknown || (kinds & (1u << value.kind)) == 0) dictionary_failure(expression, declared, value.kind);
        return (adamic_view_dictionary_result){value, 0};
    }
    // Class getters and live inherited static storage need their owning adapter.
    if (object->class != NULL) dictionary_failure(expression, declared, adamic_view_union_unknown);
    adamic_view_union_value value = {adamic_view_union_undefined, {.reference = NULL}};
    bool found = false;
    for (size_t index = 0; index < object->shape->count; index++) {
        const char *name = object->shape->names[index];
        if (strlen(name) != key->length || memcmp(name, key->bytes, key->length) != 0) continue;
        found = true;
        if (!adamic_object_initialized(object)[index]) {
            size_t capacity = strlen(expression) + strlen(declared) + 96;
            char *message = malloc(capacity);
            if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
            int length = snprintf(message, capacity, "field read failed: %s; expected %s, found uninitialized", expression, declared);
            adamic_panic(message, (size_t)length);
        }
        unsigned char actual = adamic_object_field_types(object)[index];
        const adamic_value *slot = &object->slots[index];
        if (actual == 1) value = (adamic_view_union_value){adamic_view_union_number, *slot};
        else if (actual == 2) value = (adamic_view_union_value){adamic_view_union_boolean, *slot};
        else if (actual == 7) {
            adamic_maybe_number maybe = adamic_maybe_number_unpack(slot->number);
            if (maybe.present) value = (adamic_view_union_value){adamic_view_union_number, {.number = maybe.number}};
        } else if (actual == 9) {
            adamic_maybe_boolean maybe = adamic_maybe_boolean_unpack(slot->maybe_boolean);
            if (maybe.present) value = (adamic_view_union_value){adamic_view_union_boolean, {.boolean = maybe.boolean}};
        } else if (actual == 12) value.kind = adamic_view_union_null;
        else if (actual == 13) value.kind = adamic_view_union_undefined;
        else if ((actual >= 3 && actual <= 6) || actual == 8 || actual == 10 || actual == 14) value = dictionary_boxed_value(slot->reference);
        else value.kind = adamic_view_union_unknown;
        break;
    }
    if (!found) adamic_record_check_missing_member(key);
    if (value.kind == adamic_view_union_unknown || (kinds & (1u << value.kind)) == 0) dictionary_failure(expression, declared, value.kind);
    bool reference = value.kind == adamic_view_union_object || value.kind == adamic_view_union_array || value.kind == adamic_view_union_map || value.kind == adamic_view_union_function;
    if (reference && child_contract == 0) dictionary_failure(expression, declared, adamic_view_union_unknown);
    return (adamic_view_dictionary_result){value, reference ? child_contract : 0};
}

adamic_heap *adamic_view_dictionary_box(adamic_view_union_value value) {
    if (value.kind == adamic_view_union_number) return adamic_box_number(value.payload.number);
    if (value.kind == adamic_view_union_boolean) return value.payload.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
    return adamic_retain(value.payload.reference);
}
