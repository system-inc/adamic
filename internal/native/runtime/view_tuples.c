#include "view_tuples.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Shape size is the immutable fixed tuple length; ordinary records never acquire
// tuple identity through an assertion or a numeric-looking property name.
void adamic_view_tuple(const adamic_object *value, size_t length, const char *expected, const char *expression) {
 if (value != NULL && value->tuple && value->shape->count == length) return;
 const char *found = value == NULL ? "undefined" : value->tuple ? "array" : "object";
 size_t capacity = strlen(expression) + 2 * strlen(expected) + strlen(found) + 100;
 char *message = malloc(capacity);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int written = snprintf(message, capacity, "field read failed: %s is not a %s; expected %s, found %s", expression, expected, expected, found);
 adamic_panic(message, (size_t)written);
}
