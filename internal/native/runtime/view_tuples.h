#ifndef ADAMIC_VIEW_TUPLES_H
#define ADAMIC_VIEW_TUPLES_H
#include "adamic.h"
bool adamic_tuple_matches_range(const adamic_object *, size_t, size_t, bool);
void adamic_view_tuple_range(const adamic_object *, size_t, size_t, bool, const char *, const char *);
void adamic_view_tuple(const adamic_object *, size_t, const char *, const char *);
bool adamic_tuple_matches(const adamic_object *, size_t);
// Owns a present
// normalized reference; false preserves an absent array position.
bool adamic_tuple_array_union_at(const adamic_array *, double, bool, bool, const char *, const char *, adamic_value *);
#endif
