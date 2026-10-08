#ifndef ADAMIC_VIEW_CALLABLES_CONTRACT_H
#define ADAMIC_VIEW_CALLABLES_CONTRACT_H
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* Recorded by the implementation producer, never supplied by an asserted view.
   Zero representation is unknown. Exact matching is deliberately conservative. */
typedef struct adamic_callable_signature {
    size_t arity;
    const unsigned char *parameters;
    unsigned char result;
    const char *name;
} adamic_callable_signature;

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
        } else if (recorded->result != expected->result) {
            found = "function with incompatible result representation";
        } else {
            bool compatible = recorded->arity == 0 || (recorded->parameters != NULL && expected->parameters != NULL);
            for (size_t index = 0; compatible && index < recorded->arity; index++) {
                compatible = recorded->parameters[index] != 0 && expected->parameters[index] != 0 && recorded->parameters[index] == expected->parameters[index];
            }
            if (compatible) { return value; }
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
#endif
