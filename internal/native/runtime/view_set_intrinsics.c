#include "view_set_intrinsics.h"

// Only constructors establish this identity and physical element representation.
adamic_map *adamic_view_set_producer(adamic_map *set, unsigned char element) {
    set->intrinsic_set = true;
    set->key_type = element;
    return set;
}

static _Noreturn void failure(const char *expression, const char *expected, const char *found) {
    size_t capacity = strlen(expression) + strlen(expected) + strlen(found) + 80;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s expected %s, found %s", expression, expected, found);
    adamic_panic(message, (size_t)length);
}

static adamic_value add(adamic_object *receiver, adamic_value *arguments, size_t count) {
    adamic_map *set = (adamic_map *)receiver;
    if (count != 1) { failure("Set.add", "one argument", "wrong arity"); }
    adamic_value key = arguments[0];
    if (set->reference_keys) { key.reference = adamic_graph_hold(set, key.reference); }
    adamic_map_set(set, key, (adamic_value){.number = 0});
    // A method result is owned by its caller, including this same receiver.
    return (adamic_value){.reference = adamic_retain(set)};
}

static adamic_value has(adamic_object *receiver, adamic_value *arguments, size_t count) {
    if (count != 1) { failure("Set.has", "one argument", "wrong arity"); }
    return (adamic_value){.boolean = adamic_map_get((adamic_map *)receiver, arguments[0]) != NULL};
}

// Distinct code identities select immutable signatures, independent of the view.
#ifdef ADAMIC_CLOSURE_CONVENTION
#define METHODS(label) \
static adamic_value add_##label(adamic_object *receiver, adamic_value *arguments, size_t count) { return add(receiver, arguments, count); } \
static adamic_value has_##label(adamic_object *receiver, adamic_value *arguments, size_t count) { return has(receiver, arguments, count); }
#else
#define METHODS(label) \
static adamic_value add_##label(adamic_object *receiver, adamic_value *arguments) { return add(receiver, arguments, 1); } \
static adamic_value has_##label(adamic_object *receiver, adamic_value *arguments) { return has(receiver, arguments, 1); }
#endif
METHODS(number)
METHODS(boolean)
METHODS(string)
#undef METHODS

static const unsigned char parameters[] = {1, 2, 3};
static const adamic_callable_signature signatures[] = {
    {1, parameters, 4, "Set<number>.add", NULL, 0},
    {1, parameters, 2, "Set<number>.has", NULL, 0},
    {1, parameters + 1, 4, "Set<boolean>.add", NULL, 0},
    {1, parameters + 1, 2, "Set<boolean>.has", NULL, 0},
    {1, parameters + 2, 4, "Set<string>.add", NULL, 0},
    {1, parameters + 2, 2, "Set<string>.has", NULL, 0},
};
#ifdef ADAMIC_CLOSURE_CONVENTION
#define ENTRY(code_pointer) {.counted = true, .counted_code = code_pointer}
#else
#define ENTRY(code_pointer) code_pointer
#endif
static const adamic_method_entry methods[] = {ENTRY(add_number), ENTRY(has_number), ENTRY(add_boolean), ENTRY(has_boolean), ENTRY(add_string), ENTRY(has_string)};
#undef ENTRY

adamic_method_entry adamic_view_set_method(const adamic_map *set, const char *name, const char *expression, const char *expected) {
    if (!set->intrinsic_set) { failure(expression, expected, "Map with unknown intrinsic signature"); }
    if (set->key_type < 1 || set->key_type > 3) { failure(expression, expected, "function with unknown signature"); }
    size_t offset = (size_t)(set->key_type - 1) * 2;
    if (strcmp(name, "add") == 0) { return methods[offset]; }
    if (strcmp(name, "has") == 0) { return methods[offset + 1]; }
    failure(expression, expected, "unsupported Set intrinsic");
}

const adamic_callable_signature *adamic_view_set_signature(adamic_method_entry method) {
    for (size_t i = 0; i < sizeof methods / sizeof methods[0]; i++) {
#ifdef ADAMIC_CLOSURE_CONVENTION
        if (method.counted && method.counted_code == methods[i].counted_code) { return &signatures[i]; }
#else
        if (method == methods[i]) { return &signatures[i]; }
#endif
    }
    return NULL;
}

adamic_value adamic_view_set_field(const adamic_map *set, const char *name, unsigned char wanted, const char *expected, const char *expression, bool absent) {
    if (set->intrinsic_set && strcmp(name, "size") == 0) {
        if (wanted == 1) { return (adamic_value){.number = (double)set->count}; }
        failure(expression, expected, "number");
    }
    if (absent) { return (adamic_value){.reference = NULL}; }
    failure(expression, expected, "missing");
}
