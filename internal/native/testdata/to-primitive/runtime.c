#include "adamic.h"
#include "count.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char trace[512];
static size_t used;
static int mode;
static bool changed;
static const adamic_shape shape = {0, NULL, NULL, NULL};
static adamic_string *text(const char *bytes) {
    size_t length = strlen(bytes);
    adamic_string *string = adamic_string_allocate(length);
    memcpy((char *)string->bytes, bytes, length);
    return string;
}
static void note(const char *operation, const char *name) {
    used += (size_t)snprintf(trace + used, sizeof trace - used, "%s%s %s", used ? "|" : "", operation, name);
}
static adamic_primitive primitive(enum adamic_primitive_kind kind, adamic_value value) {
    return (adamic_primitive){kind, value};
}
static adamic_primitive nothing(void) {
    return primitive(adamic_primitive_undefined, (adamic_value){.reference = NULL});
}
static void fail(const char *message) {
    adamic_string *string = text(message);
    adamic_thrown = adamic_error_new(string);
    adamic_release(string);
}
static adamic_value unused_closure(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments; return (adamic_value){.reference = NULL};
}
static adamic_primitive value_of(adamic_heap *self, void *owner) {
    note("call", "valueOf");
    if (owner != NULL && ((adamic_string *)owner)->length != 7) { abort(); }
    switch (mode) {
    case 17: return primitive(adamic_primitive_object, (adamic_value){.reference = adamic_closure_new(unused_closure, 0)});
    case 3: case 4: return primitive(adamic_primitive_number, (adamic_value){.number = 7});
    case 5: return primitive(adamic_primitive_object, (adamic_value){.reference = adamic_array_new(0, false)});
    case 8: case 16: fail("call failed"); return nothing();
    case 10: changed = true; return primitive(adamic_primitive_object, (adamic_value){.reference = adamic_object_new(&shape)});
    case 11: return nothing();
    case 12: return primitive(adamic_primitive_null, (adamic_value){.reference = NULL});
    case 13: return primitive(adamic_primitive_boolean, (adamic_value){.boolean = false});
    case 14: return primitive(adamic_primitive_number, (adamic_value){.number = -0.0});
    case 15: return primitive(adamic_primitive_string, (adamic_value){.reference = text("text")});
    default: return primitive(adamic_primitive_object, (adamic_value){.reference = adamic_retain(self)});
    }
}
static adamic_primitive to_string(adamic_heap *self, void *owner) {
    note("call", "toString");
    if (owner != NULL && ((adamic_string *)owner)->length != 8) { abort(); }
    if (mode == 17) { return primitive(adamic_primitive_object, (adamic_value){.reference = adamic_closure_new(unused_closure, 0)}); }
    if (mode == 4 || mode == 5) {
        return primitive(adamic_primitive_object, (adamic_value){.reference = adamic_object_new(&shape)});
    }
    if (mode == 2) {
        adamic_string *comma = text(",");
        adamic_string *result = adamic_array_join((adamic_array *)self, comma, adamic_join_strings);
        adamic_release(comma);
        return primitive(adamic_primitive_string, (adamic_value){.reference = result});
    }
    // Three numeric fields mirror the Version shape; a mutation must be read now.
    char bytes[64];
    adamic_object *version = (adamic_object *)self;
    snprintf(bytes, sizeof bytes, "%.0f.%.0f.%.0f", version->slots[0].number, version->slots[1].number, version->slots[2].number);
    return primitive(adamic_primitive_string, (adamic_value){.reference = text(changed ? "changed" : bytes)});
}
static adamic_primitive_method get(adamic_heap *self, const char *name) {
    (void)self;
    note("get", name);
    bool value = strcmp(name, "valueOf") == 0;
    adamic_primitive_method method = {NULL, text(name)};
    if (mode == 18 && value) { changed = true; return method; }
    if (mode == 19 && !value) { fail("second get failed"); return method; }
    if (mode == 9 && value) { fail("get failed"); return method; }
    if (mode == 7 || (mode == 6 && value)) { return method; }
    method.call = value ? value_of : to_string;
    return method;
}
// Leave the workload frame before leak checking, so stale object pointers cannot root it.
static void run(void) {
    static const char *const names[] = {"major", "minor", "patch"};
    static const bool references[] = {false, false, false};
    static const adamic_shape version_shape = {3, names, references, NULL};
    adamic_heap *receiver;
    if (mode == 2) {
        adamic_array *array = adamic_array_new(2, true);
        adamic_array_push(array, (adamic_value){.reference = text("Identifier")});
        adamic_array_push(array, (adamic_value){.reference = text("FunctionDeclaration")});
        receiver = &array->heap;
    } else {
        adamic_object *version = adamic_object_new(&version_shape);
        for (size_t i = 0; i < 3; i++) { version->slots[i].number = (double)(i + 1); }
        receiver = &version->heap;
    }
    if (mode == 21) { fail("kept"); }
    enum adamic_primitive_hint hint = mode == 0 || mode == 2 || mode == 4 ? adamic_hint_string : mode == 16 ? adamic_hint_number : adamic_hint_default;
    if (mode == 20) { hint = (enum adamic_primitive_hint)99; }
    adamic_primitive result = adamic_ordinary_to_primitive(receiver, hint, get);
    adamic_release(receiver);
    if (adamic_thrown != NULL) {
        adamic_object *error = adamic_thrown;
        adamic_thrown = NULL;
        adamic_slot_cache name = {0}, message = {0};
        adamic_string *n = adamic_object_field(error, "name", &name)->reference;
        adamic_string *m = adamic_object_field(error, "message", &message)->reference;
        printf("%.*s: %.*s\n", (int)n->length, n->bytes, (int)m->length, m->bytes);
        adamic_release(error);
    } else {
        switch (result.kind) {
        case adamic_primitive_undefined: puts("undefined"); break;
        case adamic_primitive_null: puts("null"); break;
        case adamic_primitive_boolean: puts(result.value.boolean ? "true" : "false"); break;
        case adamic_primitive_number: {
            char bytes[ADAMIC_NUMBER_FORMAT_MAX];
            size_t length = adamic_number_format(result.value.number, bytes);
            printf("%.*s\n", (int)length, bytes); break;
        }
        case adamic_primitive_string: {
            adamic_string *string = result.value.reference;
            printf("%.*s\n", (int)string->length, string->bytes);
            adamic_release(string); break;
        }
        case adamic_primitive_object: abort();
        }
    }
    puts(trace);
    puts("after catch");
}
int main(int argc, char **argv) {
    if (argc != 2) { return 2; }
    mode = atoi(argv[1]);
    run();
    ADAMIC_COUNT_REPORT();
    return 0;
}
