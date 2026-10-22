#include "view_unions_untagged.h"
#include <string.h>
#include <stdlib.h>

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
    if (slot == NULL || !adamic_object_initialized(object)[adamic_slot_index(object, slot)]) { return false; }
    unsigned char actual = adamic_object_field_types(object)[adamic_slot_index(object, slot)];
    *result = (adamic_view_union_value){adamic_view_union_unknown, *slot};
    if (actual == adamic_rep_number) { result->kind = adamic_view_union_number; return true; }
    if (actual == adamic_rep_boolean) { result->kind = adamic_view_union_boolean; return true; }
    if (actual == adamic_rep_undefined) { result->kind = adamic_view_union_undefined; return true; }
    if (actual == adamic_rep_maybe_boolean) {
        adamic_maybe_boolean b = adamic_maybe_boolean_unpack(slot->maybe_boolean);
        result->kind = b.present ? adamic_view_union_boolean : adamic_view_union_undefined;
        result->payload.boolean = b.boolean; return true;
    }
    if (actual == adamic_rep_null) { result->kind = adamic_view_union_null; return true; }
    if (actual == adamic_rep_maybe_number) {
        adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
        result->kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;
        result->payload.number = number.number;
        return true;
    }
    if ((actual >= adamic_rep_string && actual <= adamic_rep_map) || actual == adamic_rep_closure || actual == adamic_rep_union) {
        const adamic_heap *heap = slot->reference;
        if (heap == NULL) { return true; } /* Ambiguous nullish storage is Unknown. */
        if (heap->kind == adamic_kind_number && actual == adamic_rep_union) { result->kind = adamic_view_union_number; result->payload.number = ((const adamic_number_box *)heap)->number; }
        if (heap->kind == adamic_kind_boolean && actual == adamic_rep_union) { result->kind = adamic_view_union_boolean; result->payload.boolean = ((const adamic_boolean_box *)heap)->boolean; }
        if (heap->kind == adamic_kind_string && (actual == adamic_rep_string || actual == adamic_rep_union)) { result->kind = adamic_view_union_string; }
        if (heap->kind == adamic_kind_closure && (actual == adamic_rep_closure || actual == adamic_rep_union)) { result->kind = adamic_view_union_function; }
        if (heap->kind == adamic_kind_array && (actual == adamic_rep_array || actual == adamic_rep_union)) { result->kind = adamic_view_union_array; }
        if (heap->kind == adamic_kind_object && (actual == adamic_rep_object || actual == adamic_rep_union)) { result->kind = adamic_view_union_object; }
        return true;
    }
    return true;
}

// Each frame owns a snapshot and an index into the active path. Growing the
// work stack never leaves pointers into its old allocation in another frame.
typedef struct plain_match_frame {
    size_t contract;
    adamic_view_union_value value;
    size_t next;
    unsigned int state;
    bool tagged;
} plain_match_frame;

enum plain_match_state { plain_enter, plain_union, plain_tags, plain_fields, plain_return };

static void plain_match_push(plain_match_frame **frames, size_t *length, size_t *capacity, size_t id, adamic_view_union_value value) {
    if (*length == *capacity) {
        if (*capacity > SIZE_MAX / 2 / sizeof **frames) {
            static const char message[] = "out of memory";
            adamic_panic(message, sizeof message - 1);
        }
        size_t grown = *capacity == 0 ? 16 : *capacity * 2;
        plain_match_frame *next = realloc(*frames, grown * sizeof **frames);
        if (next == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
        *frames = next;
        *capacity = grown;
    }
    (*frames)[(*length)++] = (plain_match_frame){.contract = id, .value = value};
}

bool adamic_view_untagged_plain_matches(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_view_union_value *value) {
    ADAMIC_CHECK_STACK();
    plain_match_frame *frames = NULL;
    size_t length = 0, capacity = 0;
    bool returned = false, result = false;
    plain_match_push(&frames, &length, &capacity, id, *value);
    while (length != 0) {
        plain_match_frame *frame = &frames[length - 1];
        const adamic_view_union_value *current = &frame->value;
        if (returned) {
            returned = false;
            if (frame->state == plain_return || (frame->state == plain_union && result) || (frame->state != plain_union && !result)) {
                goto complete;
            }
        }
        if (frame->contract == 0 || frame->contract > count) { result = false; goto complete; }
        const adamic_view_untagged_contract *contract = &contracts[frame->contract - 1];
        if (frame->state == plain_enter) {
            bool reference = current->kind == adamic_view_union_object || current->kind == adamic_view_union_array;
            if (reference && current->payload.reference != NULL) {
                for (size_t i = 0; i + 1 < length; i++) {
                    const plain_match_frame *seen = &frames[i];
                    if (seen->contract == frame->contract && (seen->value.kind == adamic_view_union_object || seen->value.kind == adamic_view_union_array) && seen->value.payload.reference == current->payload.reference) {
                        result = true;
                        goto complete;
                    }
                }
            }
            if (current->kind == adamic_view_union_undefined && contract->undefined) { result = true; goto complete; }
            if (contract->kind == 1) {
                adamic_view_union_kind wanted = (contract->of == adamic_rep_number || contract->of == adamic_rep_maybe_number) ? adamic_view_union_number : (contract->of == adamic_rep_boolean || contract->of == adamic_rep_maybe_boolean) ? adamic_view_union_boolean : contract->of == adamic_rep_string ? adamic_view_union_string : adamic_view_union_unknown;
                result = wanted != adamic_view_union_unknown && current->kind == wanted;
                if (!result || contract->allowed_count == 0) { goto complete; }
                result = false;
                for (size_t i = 0; i < contract->allowed_count; i++) {
                    if (tag_matches(current, &contract->allowed[i])) { result = true; break; }
                }
                goto complete;
            }
            if (contract->kind == 5) {
                result = false;
                if (current->kind != adamic_view_union_function || current->payload.reference == NULL) { goto complete; }
                const adamic_closure *closure = current->payload.reference;
                for (size_t i = 0; i < contract->producer_count; i++) {
                    const adamic_view_untagged_callable_producer *producer = &contract->producers[i];
                    bool identity = closure->code == producer->code;
#ifdef ADAMIC_CLOSURE_CONVENTION
                    identity = closure->counted == producer->counted && (closure->counted ? closure->counted_code == producer->counted_code : closure->code == producer->code);
#endif
                    if (identity && adamic_view_callable_signatures_match(producer->signature, contract->callable)) { result = true; break; }
                }
                goto complete;
            }
            // Array storage still belongs to V3; an injected descriptor fails closed.
            if (contract->kind == 3) { result = false; goto complete; }
            if (contract->kind == 4) { frame->state = plain_union; continue; }
            if (contract->kind == 6) { result = current->kind == adamic_view_union_null; goto complete; }
            if (contract->kind == 7) { result = current->kind == adamic_view_union_undefined; goto complete; }
            result = false;
            if (contract->kind != 2 || current->kind != adamic_view_union_object || current->payload.reference == NULL) { goto complete; }
            const adamic_object *object = current->payload.reference;
            if (object->class != NULL && object->class->is_static) { goto complete; }
            if (contract->fixed_tuple && (!object->tuple || object->shape->count != contract->tuple_count)) { goto complete; }
            // Preserve the open kind selector without certifying unread payloads.
            for (size_t i = 0; i < contract->field_count; i++) {
                const adamic_view_untagged_field *field = &contract->fields[i];
                if (field->optional || strcmp(field->name, "kind") != 0 || field->contract == 0 || field->contract > count) { continue; }
                const adamic_view_untagged_contract *child = &contracts[field->contract - 1];
                if (child->kind != 1 || (child->of != adamic_rep_number && child->of != adamic_rep_boolean && child->of != adamic_rep_string)) { continue; }
                adamic_view_union_value slot;
                if (!adamic_view_untagged_plain_slot(NULL, current, field->name, &slot)) { goto complete; }
                frame->state = plain_return;
                plain_match_push(&frames, &length, &capacity, field->contract, slot);
                goto next_frame;
            }
            frame->state = plain_tags;
        }
        if (frame->state == plain_union) {
            if (frame->next == contract->member_count) { result = false; goto complete; }
            size_t child = contract->members[frame->next++];
            adamic_view_union_value slot = *current;
            plain_match_push(&frames, &length, &capacity, child, slot);
            continue;
        }
        if (frame->state == plain_tags) {
            // Finite own tags select a member; payload reads retain their checks.
            while (frame->next < contract->field_count) {
                const adamic_view_untagged_field *field = &contract->fields[frame->next++];
                if (field->optional || field->contract == 0 || field->contract > count) { continue; }
                const adamic_view_untagged_contract *child = &contracts[field->contract - 1];
                if (child->kind != 1 || child->allowed_count == 0) { continue; }
                frame->tagged = true;
                adamic_view_union_value slot;
                if (!adamic_view_untagged_plain_slot(NULL, current, field->name, &slot)) { result = false; goto complete; }
                plain_match_push(&frames, &length, &capacity, field->contract, slot);
                goto next_frame;
            }
            if (frame->tagged) { result = true; goto complete; }
            frame->state = plain_fields;
            frame->next = 0;
        }
        if (frame->state == plain_fields) {
            const adamic_object *object = current->payload.reference;
            while (frame->next < contract->field_count) {
                const adamic_view_untagged_field *field = &contract->fields[frame->next++];
                adamic_slot_cache cache = {0};
                if (adamic_object_optional_field(object, field->name, &cache) == NULL && field->optional) { continue; }
                adamic_view_union_value slot;
                if (!adamic_view_untagged_plain_slot(NULL, current, field->name, &slot)) { result = false; goto complete; }
                plain_match_push(&frames, &length, &capacity, field->contract, slot);
                goto next_frame;
            }
            result = true;
        }
complete:
        length--;
        returned = true;
next_frame:;
    }
    free(frames);
    return result;
}

size_t adamic_view_untagged_plain_select(const adamic_object *object, const adamic_view_untagged_contract *contracts, size_t count, size_t id, const char *expression, const char *declared) {
    adamic_view_union_value value = adamic_view_union_heap((const adamic_heap *)object);
    if (object != NULL && object->heap.kind == adamic_kind_array) { value.kind = adamic_view_union_array; }
    if (id != 0 && id <= count) {
        const adamic_view_untagged_contract *root = &contracts[id - 1];
        if (root->fixed_tuple && adamic_view_untagged_plain_matches(contracts, count, id, &value)) { return id; }
        for (size_t i = 0; i < root->member_count; i++) {
            if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value)) { return root->members[i]; }
        }
    }
    // Only this own-data adapter has proved the heap layout. Generic mixed
    // selection must reject unavailable adapters without inspecting a payload.
    if (value.kind == adamic_view_union_object && object != NULL) {
        if (object->tuple) { value.kind = adamic_view_union_array; }
        else if (object->class != NULL && object->class->is_static) { value.kind = adamic_view_union_function; }
    }
    return adamic_view_mixed_union_select(&value, NULL, 0, NULL, NULL, expression, declared);
}
