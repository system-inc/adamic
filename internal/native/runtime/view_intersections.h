#ifndef ADAMIC_VIEW_INTERSECTIONS_H
#define ADAMIC_VIEW_INTERSECTIONS_H
#include "view_unions_mixed.h"

// The owner supplies a pure matcher over the shared normalized borrowed snapshot.
// It must preserve child obligations on later reads and never invoke user code.
typedef bool (*adamic_view_intersection_match)(void *context, size_t contract, const adamic_view_union_value *value);
bool adamic_view_intersection_matches(const adamic_view_union_value *value, const size_t *members, size_t count, adamic_view_intersection_match match, void *context);
void adamic_view_intersection_require(const adamic_view_union_value *value, const size_t *members, size_t count, adamic_view_intersection_match match, void *context, const char *expression, const char *declared);
#endif
