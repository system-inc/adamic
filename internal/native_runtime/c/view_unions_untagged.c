#include "view_unions_untagged.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

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

// static_own_slot is a class constructor's own data static named field, or NULL: a public slot whose
// presence a write has set, the own keys Object.keys reads (class_features.c). A static the class only
// inherits has a slot here too, copied from its base's layout, but no presence, and reads through the
// parent (class_static.c), so it isn't the class's own. An accessor has no slot at all.
static adamic_value *static_own_slot(const adamic_object *object, const char *field) {
    adamic_slot_cache cache = {0};
    adamic_value *slot = adamic_object_optional_find(object, field, &cache);
    if (slot == NULL || object->class->static_flags == NULL) { return NULL; }
    size_t flag = object->class->static_flags[adamic_slot_index(object, slot)];
    if (flag == 0 || object->slots[flag - 1].number == 0) { return NULL; }
    return slot;
}

// plain_union_holds_callable is whether a callable could stand in union id: a callable or closure
// member, or one the table can't name (an unknown kind, an id out of range). A class constructor is a
// function there, so it keeps the stop such a union gives it. Id zero is no union.
static bool plain_union_holds_callable(const adamic_view_untagged_contract *contracts, size_t count, size_t id) {
    if (id == 0) { return false; }
    if (id > count) { return true; }
    const adamic_view_untagged_contract *union_contract = &contracts[id - 1];
    for (size_t i = 0; i < union_contract->member_count; i++) {
        size_t member = union_contract->members[i];
        if (member == 0 || member > count) { return true; }
        const adamic_view_untagged_contract *held = &contracts[member - 1];
        if (held->kind == 0 || held->kind == 5 || held->of == adamic_rep_closure) { return true; }
    }
    return false;
}

// plain_slot_read is the plain slot probe, and with statics, a class constructor's own data static
// too. Only a member its contract admitted (statics_admitted) reads one.
static bool plain_slot_read(const adamic_view_union_value *value, const char *field, bool statics, adamic_view_union_value *result) {
    if (value->kind != adamic_view_union_object || value->payload.reference == NULL) { return false; }
    const adamic_object *object = value->payload.reference;
    adamic_value *slot = NULL;
    if (object->class != NULL && object->class->is_static) {
        if (!statics) { return false; }
        slot = static_own_slot(object, field);
    } else {
        adamic_slot_cache cache = {0};
        slot = adamic_object_optional_field(object, field, &cache);
    }
    if (slot == NULL || !adamic_object_initialized(object)[adamic_slot_index(object, slot)]) { return false; }
    unsigned char actual = adamic_object_field_type(object, adamic_slot_index(object, slot));
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
        if (heap->kind == adamic_kind_null && actual == adamic_rep_union) { result->kind = adamic_view_union_null; return true; } /* The union's null, as view_unions_mixed.c reads it. */
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

bool adamic_view_untagged_plain_slot(void *context, const adamic_view_union_value *value, const char *field, adamic_view_union_value *result) {
    (void)context;
    return plain_slot_read(value, field, false, result);
}

// Each frame owns a snapshot and an index into the active path. Growing the
// work stack never leaves pointers into its old allocation in another frame.
// statics marks a class constructor its object contract admitted, whose own
// data statics the frame's slot reads take.
typedef struct plain_match_frame {
    size_t contract;
    adamic_view_union_value value;
    size_t next;
    unsigned int state;
    bool tagged;
    bool statics;
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
            // Array membership defers elements to their own checked loads.
            if (contract->kind == 3) { result = current->kind == adamic_view_union_array; goto complete; }
            if (contract->kind == 4) { frame->state = plain_union; continue; }
            if (contract->kind == 6) { result = current->kind == adamic_view_union_null; goto complete; }
            if (contract->kind == 7) { result = current->kind == adamic_view_union_undefined; goto complete; }
            result = false;
            if (contract->kind != 2 || current->kind != adamic_view_union_object || current->payload.reference == NULL) { goto complete; }
            const adamic_object *object = current->payload.reference;
            if (object->class != NULL && object->class->is_static) {
                // A class constructor stands for an object whose contract admits one (statics_admitted),
                // outside a union a callable could stand in, when every required field is its own data
                // static, as Node reads Klass.count: never an accessor, never one only a base declares.
                size_t parent = length >= 2 && frames[length - 2].state == plain_union ? frames[length - 2].contract : 0;
                if (!contract->statics_admitted || plain_union_holds_callable(contracts, count, parent)) { goto complete; }
                for (size_t i = 0; i < contract->field_count; i++) {
                    if (!contract->fields[i].optional && static_own_slot(object, contract->fields[i].name) == NULL) { goto complete; }
                }
                frame->statics = true;
            }
            if (contract->fixed_tuple && (!object->tuple || object->shape->count != contract->tuple_count)) { goto complete; }
            // Preserve the open kind selector without certifying unread payloads.
            for (size_t i = 0; i < contract->field_count; i++) {
                const adamic_view_untagged_field *field = &contract->fields[i];
                if (field->optional || strcmp(field->name, "kind") != 0 || field->contract == 0 || field->contract > count) { continue; }
                const adamic_view_untagged_contract *child = &contracts[field->contract - 1];
                if (child->kind != 1 || (child->of != adamic_rep_number && child->of != adamic_rep_boolean && child->of != adamic_rep_string)) { continue; }
                adamic_view_union_value slot;
                if (!plain_slot_read(current, field->name, frame->statics, &slot)) { goto complete; }
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
                if (!plain_slot_read(current, field->name, frame->statics, &slot)) { result = false; goto complete; }
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
                // A class's optional static it doesn't own is absent here, as Object.hasOwn says on Node.
                bool absent = frame->statics ? static_own_slot(object, field->name) == NULL : adamic_object_optional_field(object, field->name, &cache) == NULL;
                if (absent && field->optional) { continue; }
                adamic_view_union_value slot;
                if (!plain_slot_read(current, field->name, frame->statics, &slot)) { result = false; goto complete; }
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
        // The root union is this loop, not a frame plain_matches sees, so a class constructor's stop
        // in a union a callable could stand in is kept here.
        bool constructor = value.kind == adamic_view_union_object && object != NULL && object->class != NULL && object->class->is_static;
        if (!constructor || !plain_union_holds_callable(contracts, count, id)) {
            for (size_t i = 0; i < root->member_count; i++) {
                if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value)) { return root->members[i]; }
            }
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

void adamic_view_static_write(const adamic_object *object, const char *name) {
    if (object == NULL || object->class == NULL || !object->class->is_static) { return; }
    static const char prefix[] = "field write failed: a write to ", suffix[] = " through a checked view that holds a class constructor is not yet lowered";
    size_t capacity = sizeof prefix + strlen(name) + sizeof suffix;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "%s%s%s", prefix, name, suffix);
    adamic_panic(message, (size_t)length);
}

void adamic_view_static_read(const adamic_object *value, bool admitted, const char *const *required, size_t count, const char *expression, const char *type) {
    if (value == NULL || value->heap.kind != adamic_kind_object || value->class == NULL || !value->class->is_static) { return; }
    // Each required field the class's own data static, initialized, as plain_slot_read reads one.
    size_t owned = 0;
    while (admitted && owned < count) {
        adamic_value *slot = static_own_slot(value, required[owned]);
        if (slot == NULL || !adamic_object_initialized(value)[adamic_slot_index(value, slot)]) { break; }
        owned++;
    }
    if (admitted && owned == count) { return; }
    // object.c's view_failure words, as a class constructor read where no object contract admits one used to stop.
    size_t capacity = strlen(expression) + 2 * strlen(type) + 64;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s is not a %s; expected %s, found function", expression, type, type);
    adamic_panic(message, (size_t)length);
}

// Checked per read, never cached: an alias can push a lie between reads.
// Bounds precede storage validation; no bits are interpreted or retained before
// their storage flag has been checked. The returned slot owns its reference.
adamic_view_union_value adamic_view_element(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_array *array, double index, bool relative, unsigned int wanted, const char *where, const char *view_type) {
    adamic_view_union_value result = {adamic_view_union_undefined, {.reference = NULL}};
    adamic_value *slot = relative ? adamic_array_at_relative(array, index) : adamic_array_at(array, index);
    if (slot == NULL) { return result; }
    bool references = wanted == adamic_rep_string || wanted == adamic_rep_object || wanted == adamic_rep_array || wanted == adamic_rep_map || wanted == adamic_rep_union || wanted == adamic_rep_closure;
    const char *found = "unsupported representation";
    // Reference storage under a scalar contract would read a pointer as a number, so it stops here.
    // Scalar storage under a reference contract is read honestly by its own element byte below, boxed,
    // and judged by the contract (Runtime and Memory, #6wfvfhm): 7 in a number[] is a number to a
    // readonly (Node | number)[] view, and a Node to none.
    if (id != 0 && array->references && !references) {
        found = "reference storage";
        goto failure;
    }
    if (array->references) {
        result = adamic_view_union_heap(slot->reference);
    } else if (array->element == adamic_rep_number) {
        result = (adamic_view_union_value){adamic_view_union_number, *slot};
    } else if (array->element == adamic_rep_boolean) {
        result = (adamic_view_union_value){adamic_view_union_boolean, *slot};
    } else if (array->element == adamic_rep_maybe_number) {
        adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
        result.kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;
        result.payload.number = number.number;
    } else { result.kind = adamic_view_union_unknown; }
    found = adamic_view_union_kind_name(result.kind);
    bool matches = id == 0 || adamic_view_untagged_plain_matches(contracts, count, id, &result);
    // TypeScript's object admits every nonprimitive, including arrays/functions.
    if (id != 0 && id <= count && contracts[id - 1].kind == 2 && contracts[id - 1].of == adamic_rep_union) {
        matches = result.kind == adamic_view_union_object || result.kind == adamic_view_union_array || result.kind == adamic_view_union_map || result.kind == adamic_view_union_function;
    }
    if (!matches) { goto failure; }
    adamic_value normalized = result.payload;
    result.payload = *slot;
    if (references) {
        if (!array->references) {
            if (result.kind == adamic_view_union_number) { result.payload.reference = adamic_box_number(normalized.number); }
            else if (result.kind == adamic_view_union_boolean) { result.payload.reference = normalized.boolean ? &adamic_box_true : &adamic_box_false; }
            else { result.payload.reference = NULL; }
        } else { adamic_retain(result.payload.reference); }
    }
    return result;
failure:;
    size_t capacity = strlen(where) + 2 * strlen(view_type) + strlen(found) + 64;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "element read failed: %s is not a %s; expected %s, found %s", where, view_type, view_type, found);
    adamic_panic(message, (size_t)length);
    return result;
}
