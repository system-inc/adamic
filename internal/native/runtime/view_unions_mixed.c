#include "view_unions_mixed.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_view_union_value adamic_view_union_heap(const adamic_heap *value) {
    adamic_view_union_value snapshot = {adamic_view_union_unknown, {.reference = (void *)value}};
    if (value == NULL) { snapshot.kind = adamic_view_union_undefined; return snapshot; }
    switch (value->kind) {
    case adamic_kind_number:
        snapshot.kind = adamic_view_union_number;
        snapshot.payload.number = ((const adamic_number_box *)value)->number;
        break;
    case adamic_kind_boolean:
        snapshot.kind = adamic_view_union_boolean;
        snapshot.payload.boolean = ((const adamic_boolean_box *)value)->boolean;
        break;
    case adamic_kind_string: snapshot.kind = adamic_view_union_string; break;
    case adamic_kind_object: snapshot.kind = adamic_view_union_object; break;
    case adamic_kind_array: snapshot.kind = adamic_view_union_array; break;
    case adamic_kind_map: snapshot.kind = adamic_view_union_map; break;
    case adamic_kind_closure: snapshot.kind = adamic_view_union_function; break;
    default: break;
    }
    return snapshot;
}

const char *adamic_view_union_kind_name(adamic_view_union_kind kind) {
    switch (kind) {
    case adamic_view_union_number: return "number";
    case adamic_view_union_boolean: return "boolean";
    case adamic_view_union_string: return "string";
    case adamic_view_union_object: return "object";
    case adamic_view_union_array: return "array";
    case adamic_view_union_map: return "Map";
    case adamic_view_union_function: return "function";
    case adamic_view_union_null: return "null";
    case adamic_view_union_undefined: return "undefined";
    default: return "unsupported representation";
    }
}

static bool adamic_view_union_literal_matches(const adamic_view_union_member *member, const adamic_view_union_value *value) {
    if (!member->literal) { return true; }
    if (member->kind != value->kind) { return false; }
    switch (member->kind) {
    case adamic_view_union_number: return value->payload.number == member->value.number;
    case adamic_view_union_boolean: return value->payload.boolean == member->value.boolean;
    case adamic_view_union_string:
        return value->payload.reference != NULL && member->value.reference != NULL && adamic_string_equal(value->payload.reference, member->value.reference);
    default: return false;
    }
}

size_t adamic_view_mixed_union_select(const adamic_view_union_value *value, const adamic_view_union_member *members, size_t count, adamic_view_union_match match, void *context, const char *expression, const char *declared) {
    for (size_t index = 0; index < count; index++) {
        const adamic_view_union_member *member = &members[index];
        if (value->kind == adamic_view_union_unknown || member->kind != value->kind) { continue; }
        if ((value->kind == adamic_view_union_string || value->kind == adamic_view_union_object || value->kind == adamic_view_union_array || value->kind == adamic_view_union_map || value->kind == adamic_view_union_function) && value->payload.reference == NULL) { continue; }
        if (!adamic_view_union_literal_matches(member, value)) { continue; }
        if (member->kind == adamic_view_union_object || member->kind == adamic_view_union_array || member->kind == adamic_view_union_map || member->kind == adamic_view_union_function) {
            if (member->contract == 0 || match == NULL || !match(context, member, value)) { continue; }
        }
        return index;
    }
    const char *found = adamic_view_union_kind_name(value->kind);
    size_t capacity = strlen(expression) + 2 * strlen(declared) + strlen(found) + 96;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "cast failed: field read failed: %s matches no member of %s; expected %s, found %s", expression, declared, declared, found);
    adamic_panic(message, (size_t)length);
    return 0;
}
