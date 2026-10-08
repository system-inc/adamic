// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

_Thread_local adamic_object *adamic_thrown;

static const char *const error_names[] = {"name", "message", "code"};
static const bool error_references[] = {true, true, true};
static const adamic_shape error_shape = {3, error_names, error_references, NULL};
static adamic_string error_name = ADAMIC_STRING("Error");

// Reserve the built-in Error identities used by the compiler. Keep the host's
// own code-bearing shape and ownership layout while sharing nominal identity.
void adamic_error_tag(adamic_object *error) {
    static const adamic_class error_class = {NULL, 0, 3, NULL, 1u << 30, &error_shape, NULL, 0, false, 0, NULL};
    static const adamic_class type_error_class = {&error_class, 3, 3, NULL, (1u << 30) + 1, &error_shape, NULL, 0, false, 0, NULL};
    static const adamic_class range_error_class = {&error_class, 3, 3, NULL, (1u << 30) + 2, &error_shape, NULL, 0, false, 0, NULL};
    const adamic_string *name = error->slots[0].reference;
    error->class = name->length == 9 && memcmp(name->bytes, "TypeError", 9) == 0 ? &type_error_class :
                   name->length == 10 && memcmp(name->bytes, "RangeError", 10) == 0 ? &range_error_class : &error_class;
}

adamic_object *adamic_error_new(adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message);
	adamic_error_tag(error);
	return error;
}

_Noreturn void adamic_uncaught(void) {
    adamic_string *text = adamic_error_to_string(adamic_thrown);
    adamic_panic(text->bytes, text->length);
}

// Built-in Error objects. Constructor/InstallErrorCause ordering and toString
// follow V8 13.6.233.17 builtins-error.cc and ErrorUtils::ToString in messages.cc.
// Copyright the V8 project authors. BSD-3-Clause (THIRD_PARTY_NOTICES.md).
// Stack frames below are Adamic source frames, not V8 frames.
static adamic_string builtin_error_names[] = {
    ADAMIC_STRING("Object"), ADAMIC_STRING("Error"), ADAMIC_STRING("TypeError"),
    ADAMIC_STRING("RangeError"), ADAMIC_STRING("SyntaxError"),
    ADAMIC_STRING("ReferenceError"), ADAMIC_STRING("EvalError"),
    ADAMIC_STRING("URIError"), ADAMIC_STRING("AggregateError")
};
// Static prototype objects are immortal identities, not counted allocations.
typedef union {
    max_align_t alignment;
    unsigned char bytes[sizeof(adamic_object) + 3 * sizeof(adamic_value)];
} adamic_error_proto;
static adamic_error_proto error_prototypes[9];
adamic_object *adamic_error_prototype(int kind) { return (adamic_object *)error_prototypes[kind].bytes; }
void adamic_error_init_prototypes(void) {
    for (int kind = 0; kind <= 8; kind++) {
        adamic_object *prototype = adamic_error_prototype(kind);
        prototype->heap.kind = adamic_kind_object;
        prototype->shape = &error_shape;
        prototype->error_kind = -(kind + 1);
        prototype->error_own = kind == 0 ? 0 : 3;
        prototype->slots[0].reference = &builtin_error_names[kind];
        prototype->slots[1].reference = (void *)&adamic_string_empty;
    }
}
void adamic_error_reset_prototypes(void) {
    for (int kind = 0; kind <= 8; kind++) {
        for (int member = 0; member <= 1; member++) {
            adamic_heap *value = adamic_error_prototype(kind)->slots[member].reference;
            if (value != NULL && value->references != 0) adamic_release(value);
        }
        adamic_error_prototype(kind)->slots[0].reference = &builtin_error_names[kind];
        adamic_error_prototype(kind)->slots[1].reference = &adamic_string_empty;
    }
}
adamic_heap *adamic_error_get_prototype(const adamic_heap *value) {
    const adamic_object *object = (const adamic_object *)value;
    int kind = object->error_kind;
    if (kind > 0) return (adamic_heap *)adamic_error_prototype(kind);
    if (kind == -1) return &adamic_null;
    if (kind == -2) return (adamic_heap *)adamic_error_prototype(0);
    if (kind < -2) return (adamic_heap *)adamic_error_prototype(1);
    // A legacy host Error still has the real nominal Error ancestry.
    if (object->class != NULL && object->class->definition >= (1u << 30)) {
        return (adamic_heap *)adamic_error_prototype((int)(object->class->definition - (1u << 30)) + 1);
    }
    adamic_panic("compiler bug: unsupported Error prototype receiver", sizeof("compiler bug: unsupported Error prototype receiver") - 1);
}
bool adamic_error_is_prototype_of(const adamic_object *prototype, const adamic_heap *value) {
    while (value != NULL && value != &adamic_null && value->kind == adamic_kind_object) {
        value = adamic_error_get_prototype(value);
        if (value == (const adamic_heap *)prototype) return true;
    }
    return false;
}
bool adamic_error_instanceof(const adamic_heap *value, int wanted) {
    if (value == NULL || value == &adamic_null || value->kind != adamic_kind_object) return false;
    const adamic_object *object = (const adamic_object *)value;
    int kind = object->error_kind;
    if (kind > 0) return wanted == 0 || wanted == 1 || wanted == kind;
    if (kind < 0) return wanted == 0 ? kind != -1 : wanted == 1 && kind < -2;
    for (const adamic_class *c = object->class; c != NULL; c = c->base) {
        if (c->definition == (1u << 30) + (unsigned)(wanted - 1)) return true;
    }
    return false;
}
adamic_string *adamic_error_member(const adamic_object *error, int member) {
    if (error->error_kind > 0 && !(error->error_own & (1u << member))) {
        return adamic_retain(adamic_error_prototype(error->error_kind)->slots[member].reference);
    }
    return adamic_retain(error->slots[member].reference);
}
static void error_append_order(adamic_object *error, unsigned id) {
    unsigned shift = 0;
    while (((error->error_order >> shift) & 15) != 0) shift += 4;
    error->error_order |= (id + 1) << shift;
}
adamic_string *adamic_error_set_member(adamic_object *error, int member, adamic_string *value) {
    adamic_object_check_data_write(error, member == 0 ? "name" : "message");
    if (error->error_kind < 0) adamic_share(value);
    adamic_retain(value);
    adamic_release(error->slots[member].reference);
    error->slots[member].reference = value;
    if (error->error_kind > 0 && !(error->error_own & (1u << member))) {
        error->error_enumerable |= 1u << member;
        error_append_order(error, member == 0 ? 4 : 1);
    }
    error->error_own |= 1u << member;
    return adamic_retain(value);
}
adamic_string *adamic_error_to_string(const adamic_object *error) {
    adamic_string *name, *message;
    if (error->error_kind != 0 || (error->class != NULL && error->class->definition >= (1u << 30))) {
        name = adamic_error_member(error, 0);
        message = adamic_error_member(error, 1);
    } else {
        adamic_slot_cache a = {0}, b = {0};
        adamic_value *n = adamic_object_optional_field(error, "name", &a);
        adamic_value *m = adamic_object_optional_field(error, "message", &b);
        name = adamic_retain(n == NULL ? &builtin_error_names[1] : n->reference);
        message = adamic_retain(m == NULL ? &adamic_string_empty : m->reference);
    }
    adamic_string *result;
    if (name->length == 0) result = adamic_retain(message);
    else if (message->length == 0) result = adamic_retain(name);
    else {
        static adamic_string separator = ADAMIC_STRING(": ");
        adamic_string *parts[] = {name, &separator, message};
        result = adamic_string_concat(3, parts);
    }
    adamic_release(name); adamic_release(message);
    return result;
}
void adamic_error_capture_at(adamic_object *target, const adamic_string *frames, double limit) {
    adamic_object_check_data_write(target, "stack");
    if (target->has_captured_stack) adamic_release(target->captured_stack.reference);
    if (target->error_frames != NULL && target->error_frames->heap.references != 0) adamic_release(target->error_frames);
    target->error_frames = (adamic_string *)(limit >= 1 ? frames : &adamic_string_empty);
    if (target->error_frames->heap.references != 0) adamic_retain(target->error_frames);
    target->captured_stack.reference = NULL;
    target->has_captured_stack = true;
}
adamic_object *adamic_error_builtin_new(int kind, adamic_string *message, bool message_own,
    adamic_heap *cause, bool cause_own, adamic_array *errors, int element,
    const adamic_string *frames, double limit) {
    adamic_object *error = adamic_error_new(message);
    if (kind != 1) {
        adamic_release(error->slots[0].reference);
        error->slots[0].reference = adamic_retain(&builtin_error_names[kind]);
    }
    error->error_kind = kind;
    adamic_error_tag(error);
    error_append_order(error, 0);
    if (message_own) error_append_order(error, 1);
    if (cause_own) error_append_order(error, 2);
    if (errors != NULL) error_append_order(error, 3);
    error->error_own = (message_own ? 2 : 0) | (cause_own ? 4 : 0) | (errors != NULL ? 8 : 0);
    error->error_cause = cause_own ? adamic_retain(cause) : NULL;
    if (errors != NULL) {
        error->error_errors = adamic_array_new(errors->length, true);
        for (size_t index = 0; index < errors->length; index++) {
            adamic_heap *value;
            if (element == 1) value = adamic_box_number(errors->elements[index].number);
            else if (element == 2) value = adamic_retain(errors->elements[index].boolean ? &adamic_box_true : &adamic_box_false);
            else value = adamic_retain(errors->elements[index].reference);
            adamic_array_push(error->error_errors, (adamic_value){.reference = value});
        }
    }
    adamic_error_capture_at(error, frames, limit);
    return error;
}
adamic_heap *adamic_error_cause(const adamic_object *error) { return adamic_retain(error->error_cause); }
adamic_array *adamic_error_errors(const adamic_object *error) { return adamic_retain(error->error_errors); }
bool adamic_error_has_own(const adamic_object *error, const adamic_string *key) {
    const char *names[] = {"name", "message", "cause", "errors"};
    if (key->length == 5 && memcmp(key->bytes, "stack", 5) == 0) return error->has_captured_stack;
    for (unsigned i = 0; i < 4; i++) {
        if (key->length == strlen(names[i]) && memcmp(key->bytes, names[i], key->length) == 0) return (error->error_own & (1u << i)) != 0;
    }
    if (key->length == 11 && memcmp(key->bytes, "constructor", 11) == 0) return error->error_kind < -1;
    if (key->length == 8 && memcmp(key->bytes, "toString", 8) == 0) return error->error_kind == -2;
    return false;
}
adamic_array *adamic_error_names(const adamic_object *error, bool all) {
    adamic_array *result = adamic_array_new(0, true);
    if (!all && error->error_enumerable == 0) return result;
    const char *instance_names[] = {"stack", "message", "cause", "errors", "name", "constructor", "toString"};
    const char *prototype_names[] = {"constructor", "name", "message", "toString", "stack", "cause", "errors"};
    const char **names = error->error_kind < 0 ? prototype_names : instance_names;
    for (size_t i = 0; i < 7; i++) {
        unsigned id = (unsigned)i;
        if (error->error_kind > 0) {
            id = (error->error_order >> (i * 4)) & 15;
            if (id == 0) break;
            id--;
        }
        adamic_string *name = adamic_string_allocate(strlen(names[id]));
        memcpy((char *)name->bytes, names[id], name->length);
        if (all ? adamic_error_has_own(error, name) : adamic_error_enumerable(error, name)) adamic_array_push(result, (adamic_value){.reference = name});
        else adamic_release(name);
    }
    return result;
}

bool adamic_error_enumerable(const adamic_object *error, const adamic_string *key) {
    return (error->error_enumerable & 1) && key->length == 4 && memcmp(key->bytes, "name", 4) == 0
        ? true : (error->error_enumerable & 2) && key->length == 7 && memcmp(key->bytes, "message", 7) == 0;
}
