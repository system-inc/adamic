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
