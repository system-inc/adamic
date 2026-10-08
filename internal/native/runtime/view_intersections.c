#include "view_intersections.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

bool adamic_view_intersection_matches(const adamic_view_union_value *value, const size_t *members, size_t count, adamic_view_intersection_match match, void *context) {
    if (value == NULL || value->kind == adamic_view_union_unknown || members == NULL || count == 0 || match == NULL) { return false; }
    for (size_t index = 0; index < count; index++) {
        if (members[index] == 0 || !match(context, members[index], value)) { return false; }
    }
    return true;
}

void adamic_view_intersection_require(const adamic_view_union_value *value, const size_t *members, size_t count, adamic_view_intersection_match match, void *context, const char *expression, const char *declared) {
    if (adamic_view_intersection_matches(value, members, count, match, context)) { return; }
    const char *found = adamic_view_union_kind_name(value == NULL ? adamic_view_union_unknown : value->kind);
    size_t capacity = strlen(expression) + strlen(declared) + strlen(found) + 96;
    char *message = malloc(capacity);
    if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
    int length = snprintf(message, capacity, "field read failed: %s does not satisfy every member; expected %s, found %s", expression, declared, found);
    adamic_panic(message, (size_t)length);
}
