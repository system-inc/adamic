#ifndef ADAMIC_VIEW_CALLABLES_CONTRACT_H
#define ADAMIC_VIEW_CALLABLES_CONTRACT_H
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* Recorded by the implementation producer, never supplied by an asserted view.
   Zero representation is unknown. Directional member checks remain conservative. */
typedef struct adamic_callable_signature {
    size_t arity;
    const unsigned char *parameters;
    unsigned char result;
    const char *name;
    const uint16_t *parameter_masks;
    uint16_t result_mask;
} adamic_callable_signature;

static inline uint16_t adamic_callable_members(unsigned char representation, uint16_t members) {
    if (representation == 0 || representation == 254 || representation == 255) return 0;
    if (members != 0 || representation == 10) return members;
    return (uint16_t)(1u << representation);
}

static inline bool adamic_callable_representation_compatible(unsigned char from, unsigned char to, uint16_t from_members, uint16_t to_members) {
    if (from == 0 || to == 0) return false;
    if (from != 10 && to != 10 && from != to) return false;
    if (from != 10 && to != 10 && (from_members == 0 || to_members == 0)) return from == to;
    uint16_t given = adamic_callable_members(from, from_members);
    uint16_t wanted = adamic_callable_members(to, to_members);
    return given != 0 && wanted != 0 && (given & wanted) == given;
}

/* Shared immutable producer predicate for field reads and union selectors. */
static inline bool adamic_view_callable_signatures_match(const adamic_callable_signature *recorded, const adamic_callable_signature *expected) {
    if (recorded == NULL || expected == NULL || recorded->arity != expected->arity || recorded->result == 0 || expected->result == 0) return false;
    if (expected->result != 255 && !adamic_callable_representation_compatible(recorded->result, expected->result, recorded->result_mask, expected->result_mask)) return false;
    if (recorded->arity != 0 && (recorded->parameters == NULL || expected->parameters == NULL)) return false;
    for (size_t i = 0; i < recorded->arity; i++) {
        if (!adamic_callable_representation_compatible(expected->parameters[i], recorded->parameters[i], expected->parameter_masks == NULL ? 0 : expected->parameter_masks[i], recorded->parameter_masks == NULL ? 0 : recorded->parameter_masks[i])) return false;
    }
    return true;
}

static inline const adamic_heap *adamic_view_callable_shape(
    const adamic_heap *value, const adamic_callable_signature *recorded,
    const adamic_callable_signature *expected, const char *expression,
    bool optional) {
    if (optional && value == NULL) { return NULL; }
    const char *found = value == NULL ? "undefined" :
        value->kind == adamic_kind_number ? "number" :
        value->kind == adamic_kind_boolean ? "boolean" :
        value->kind == adamic_kind_string ? "string" :
        value->kind == adamic_kind_closure ? "function with unknown signature" : "object";
    char arity[80];
    if (value != NULL && value->kind == adamic_kind_closure && recorded != NULL && expected != NULL) {
        if (recorded->arity != expected->arity) {
            (void)snprintf(arity, sizeof arity, "function with arity %zu", recorded->arity);
            found = arity;
        } else if (recorded->result == 0 || expected->result == 0) {
            found = "function with unknown signature";
        } else if (expected->result != 255 && !adamic_callable_representation_compatible(recorded->result, expected->result, recorded->result_mask, expected->result_mask)) {
            found = "function with incompatible result representation";
        } else {
            bool compatible = recorded->arity == 0 || (recorded->parameters != NULL && expected->parameters != NULL);
            for (size_t index = 0; compatible && index < recorded->arity; index++) {
                compatible = adamic_callable_representation_compatible(expected->parameters[index], recorded->parameters[index], expected->parameter_masks == NULL ? 0 : expected->parameter_masks[index], recorded->parameter_masks == NULL ? 0 : recorded->parameter_masks[index]);
            }
            if (compatible && adamic_view_callable_signatures_match(recorded, expected)) { return value; }
            found = "function with incompatible parameter representations";
        }
    }
    const char *wanted = expected == NULL || expected->name == NULL ? "function with known signature" : expected->name;
    size_t capacity = strlen(expression) + strlen(wanted) + strlen(found) + 80;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s expected %s, found %s", expression, wanted, found);
    adamic_panic(message, (size_t)length);
}

/* A known producer must serve every advertised overload. Extra arguments are
   evaluated and ignored; omitted producer parameters must admit undefined. */
static inline bool adamic_view_callable_overload_matches(const adamic_callable_signature *recorded, const adamic_callable_signature *expected) {
    if (recorded == NULL || expected == NULL || recorded->result == 0 || expected->result == 0) return false;
    if (!adamic_callable_representation_compatible(recorded->result, expected->result, recorded->result_mask, expected->result_mask)) return false;
    for (size_t index = 0; index < recorded->arity; index++) {
        if (recorded->parameters == NULL) return false;
        uint16_t takes = recorded->parameter_masks == NULL ? 0 : recorded->parameter_masks[index];
        if (index >= expected->arity) {
            if ((adamic_callable_members(recorded->parameters[index], takes) & 1u) == 0) return false;
        } else {
            if (expected->parameters == NULL || !adamic_callable_representation_compatible(expected->parameters[index], recorded->parameters[index], expected->parameter_masks == NULL ? 0 : expected->parameter_masks[index], takes)) return false;
        }
    }
    return true;
}

static inline const adamic_heap *adamic_view_callable_overloads(const adamic_heap *value, const adamic_callable_signature *recorded, const adamic_callable_signature *const *expected, size_t count, const char *name, const char *expression, bool optional) {
    if (optional && value == NULL) return NULL;
    if (value == NULL || value->kind != adamic_kind_closure || recorded == NULL || count == 0) return adamic_view_callable_shape(value, recorded, count == 0 ? NULL : expected[0], expression, optional);
    size_t index = 0;
    while (index < count && adamic_view_callable_overload_matches(recorded, expected[index])) index++;
    if (index == count) return value;
    size_t capacity = strlen(expression) + strlen(name) + 120;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s expected %s, found function with incompatible overload signature %zu", expression, name, index + 1);
    adamic_panic(message, (size_t)length);
}
#endif
