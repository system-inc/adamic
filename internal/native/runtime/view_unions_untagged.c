#include "view_unions_untagged.h"
#include <string.h>

static bool tag_matches(const adamic_view_union_value *slot, const adamic_view_union_member *allowed) {
    if (!allowed->literal || slot->kind != allowed->kind) { return false; }
    switch (slot->kind) {
    case adamic_view_union_number: return slot->payload.number == allowed->value.number;
    case adamic_view_union_boolean: return slot->payload.boolean == allowed->value.boolean;
    case adamic_view_union_string:
        return slot->payload.reference != NULL && allowed->value.reference != NULL && adamic_string_equal(slot->payload.reference, allowed->value.reference);
    default: return false;
    }
}

size_t adamic_view_untagged_union_select(const adamic_view_union_value *value, const adamic_view_untagged_member *members, size_t count, adamic_view_untagged_probe probe, adamic_view_union_match match, void *context, const char *expression, const char *declared) {
    if (value->kind == adamic_view_union_object && value->payload.reference != NULL && match != NULL) {
        for (size_t index = 0; index < count; index++) {
            const adamic_view_untagged_member *member = &members[index];
            if (member->contract == 0) { continue; }
            bool candidate = true;
            for (size_t tag = 0; tag < member->count; tag++) {
                const adamic_view_untagged_tag *constraint = &member->tags[tag];
                adamic_view_union_value slot = {0};
                if (probe == NULL || !probe(context, value, constraint->field, &slot)) { candidate = false; break; }
                bool accepted = false;
                for (size_t literal = 0; literal < constraint->count; literal++) {
                    if (tag_matches(&slot, &constraint->allowed[literal])) { accepted = true; break; }
                }
                if (!accepted) { candidate = false; break; }
            }
            if (!candidate) { continue; }
            adamic_view_union_member descriptor = { .kind = adamic_view_union_object, .contract = member->contract };
            if (match(context, &descriptor, value)) { return member->contract; }
        }
    }
    // Reuse the common field/expected/found failure text and exit-70 path.
    return adamic_view_mixed_union_select(value, NULL, 0, NULL, NULL, expression, declared);
}

bool adamic_view_untagged_plain_slot(void *context, const adamic_view_union_value *value, const char *field, adamic_view_union_value *result) {
    (void)context;
    if (value->kind != adamic_view_union_object || value->payload.reference == NULL) { return false; }
    const adamic_object *object = value->payload.reference;
    if (object->class != NULL && object->class->is_static) { return false; }
    adamic_slot_cache cache = {0};
    adamic_value *slot = adamic_object_optional_field(object, field, &cache);
    if (slot == NULL || !adamic_object_initialized(object)[cache.index]) { return false; }
    unsigned char actual = adamic_object_field_types(object)[cache.index];
    *result = (adamic_view_union_value){adamic_view_union_unknown, *slot};
    if (actual == 1) { result->kind = adamic_view_union_number; return true; }
    if (actual == 2) { result->kind = adamic_view_union_boolean; return true; }
    if (actual == 13) { result->kind = adamic_view_union_undefined; return true; }
    if (actual == 12) { result->kind = adamic_view_union_null; return true; }
    if (actual == 7) {
        adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
        result->kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;
        result->payload.number = number.number;
        return true;
    }
    if ((actual >= 3 && actual <= 6) || actual == 8 || actual == 10) {
        const adamic_heap *heap = slot->reference;
        if (heap == NULL) { return true; } /* Ambiguous nullish storage is Unknown. */
        if (heap->kind == adamic_kind_number && actual == 10) { result->kind = adamic_view_union_number; result->payload.number = ((const adamic_number_box *)heap)->number; }
        if (heap->kind == adamic_kind_boolean && actual == 10) { result->kind = adamic_view_union_boolean; result->payload.boolean = ((const adamic_boolean_box *)heap)->boolean; }
        if (heap->kind == adamic_kind_string && (actual == 3 || actual == 10)) { result->kind = adamic_view_union_string; }
        if (heap->kind == adamic_kind_array && (actual == 5 || actual == 10)) { result->kind = adamic_view_union_array; }
        if (heap->kind == adamic_kind_object && (actual == 4 || actual == 10)) { result->kind = adamic_view_union_object; }
        return true;
    }
    return true;
}

typedef struct plain_match_path {
    size_t contract;
    const void *reference;
    const struct plain_match_path *previous;
} plain_match_path;

static bool plain_matches_depth(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_view_union_value *value, size_t depth, const plain_match_path *path) {
    bool reference = value->kind == adamic_view_union_object || value->kind == adamic_view_union_array;
    if (reference && value->payload.reference != NULL) {
        for (const plain_match_path *seen = path; seen != NULL; seen = seen->previous) {
            if (seen->contract == id && seen->reference == value->payload.reference) { return true; }
        }
    }
    plain_match_path current = {id, reference ? value->payload.reference : NULL, path};
    path = &current;
    if (depth > 128) { return false; }
    if (id == 0 || id > count) { return false; }
    const adamic_view_untagged_contract *contract = &contracts[id - 1];
    if (value->kind == adamic_view_union_undefined && contract->undefined) { return true; }
    if (contract->kind == 1) {
        adamic_view_union_kind wanted = (contract->of == 1 || contract->of == 7) ? adamic_view_union_number : (contract->of == 2 || contract->of == 9) ? adamic_view_union_boolean : contract->of == 3 ? adamic_view_union_string : adamic_view_union_unknown;
        if (wanted == adamic_view_union_unknown || value->kind != wanted) { return false; }
        if (contract->allowed_count == 0) { return true; }
        for (size_t i = 0; i < contract->allowed_count; i++) {
            if (tag_matches(value, &contract->allowed[i])) { return true; }
        }
        return false;
    }
    if (contract->kind == 3) {
        if (value->kind != adamic_view_union_array || value->payload.reference == NULL || contract->element == 0) { return false; }
        const adamic_array *array = value->payload.reference;
        for (size_t i = 0; i < array->length; i++) {
            adamic_value *slot = adamic_array_holes_at(array, (double)i);
            adamic_view_union_value item = {adamic_view_union_undefined, {0}};
            if (slot != NULL) {
                item.payload = *slot;
                item.kind = adamic_view_union_unknown;
                if (array->element_kind == 1) { item.kind = adamic_view_union_number; }
                if (array->element_kind == 2) { item.kind = adamic_view_union_boolean; }
                if (array->element_kind == 7) {
                    adamic_maybe_number n = adamic_maybe_number_unpack(slot->number);
                    item.kind = n.present ? adamic_view_union_number : adamic_view_union_undefined;
                    item.payload.number = n.number;
                }
                if (array->element_kind == 10 && slot->reference != NULL) {
                    const adamic_heap *heap = slot->reference;
                    if (heap->kind == adamic_kind_object) { item.kind = adamic_view_union_object; }
                    if (heap->kind == adamic_kind_array) { item.kind = adamic_view_union_array; }
                    if (heap->kind == adamic_kind_string) { item.kind = adamic_view_union_string; }
                    if (heap->kind == adamic_kind_number) {item.kind = adamic_view_union_number;item.payload.number = ((const adamic_number_box *)heap)->number;}
                    if (heap->kind == adamic_kind_boolean) {item.kind = adamic_view_union_boolean;item.payload.boolean = ((const adamic_boolean_box *)heap)->boolean;}
                }
            }
            if (!plain_matches_depth(contracts, count, contract->element, &item, depth + 1, path)) { return false; }
        }
        return true;
    }
    if (contract->kind == 4) {
        for (size_t i = 0; i < contract->member_count; i++) {
            if (plain_matches_depth(contracts, count, contract->members[i], value, depth + 1, path)) { return true; }
        }
        return false;
    }
    if (contract->kind == 7) { return value->kind == adamic_view_union_undefined; }
    if (contract->kind != 2 || value->kind != adamic_view_union_object || value->payload.reference == NULL) { return false; }
    const adamic_object *object = value->payload.reference;
    if (object->class != NULL && object->class->is_static) { return false; }
    /* An open kind contract remains a checked selector, not a payload
     * certificate. Numeric enums intentionally admit unnamed numbers. */
    for (size_t i = 0; i < contract->field_count; i++) {
        const adamic_view_untagged_field *field = &contract->fields[i];
        if (field->optional || strcmp(field->name, "kind") != 0 || field->contract == 0 || field->contract > count) { continue; }
        const adamic_view_untagged_contract *child = &contracts[field->contract - 1];
        if (child->kind != 1 || (child->of != 1 && child->of != 2 && child->of != 3)) { continue; }
        adamic_view_union_value slot;
        return adamic_view_untagged_plain_slot(NULL, value, field->name, &slot) && plain_matches_depth(contracts, count, field->contract, &slot, depth + 1, path);
    }
    /* A member's own finite tags select it without inspecting unread payloads.
     * Subsequent reads still use the existing per-field checks. */
    bool tagged = false;
    for (size_t i = 0; i < contract->field_count; i++) {
        const adamic_view_untagged_field *field = &contract->fields[i];
        if (field->optional || field->contract == 0 || field->contract > count) { continue; }
        const adamic_view_untagged_contract *child = &contracts[field->contract - 1];
        if (child->kind != 1 || child->allowed_count == 0) { continue; }
        tagged = true;
        adamic_view_union_value slot;
        if (!adamic_view_untagged_plain_slot(NULL, value, field->name, &slot) || !plain_matches_depth(contracts, count, field->contract, &slot, depth + 1, path)) { return false; }
    }
    if (tagged) { return true; }
    for (size_t i = 0; i < contract->field_count; i++) {
        const adamic_view_untagged_field *field = &contract->fields[i];
        adamic_slot_cache cache = {0};
        adamic_value *present = adamic_object_optional_field(object, field->name, &cache);
        if (present == NULL && field->optional) { continue; }
        adamic_view_union_value slot;
        if (!adamic_view_untagged_plain_slot(NULL, value, field->name, &slot) || !plain_matches_depth(contracts, count, field->contract, &slot, depth + 1, path)) { return false; }
    }
    return true;
}

bool adamic_view_untagged_plain_matches(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_view_union_value *value) {
    return plain_matches_depth(contracts, count, id, value, 0, NULL);
}

size_t adamic_view_untagged_plain_select(const adamic_object *object, const adamic_view_untagged_contract *contracts, size_t count, size_t id, const char *expression, const char *declared) {
    adamic_view_union_value value = {adamic_view_union_object, {.reference = (void *)object}};
    if (object != NULL && object->heap.kind == adamic_kind_array) { value.kind = adamic_view_union_array; }
    if (id != 0 && id <= count) {
        const adamic_view_untagged_contract *root = &contracts[id - 1];
        for (size_t i = 0; i < root->member_count; i++) {
            if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value)) { return root->members[i]; }
        }
    }
    return adamic_view_mixed_union_select(&value, NULL, 0, NULL, NULL, expression, declared);
}
