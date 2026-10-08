#ifndef ADAMIC_VIEW_TUPLES_H
#define ADAMIC_VIEW_TUPLES_H
#include "adamic.h"
void adamic_view_tuple(const adamic_object *, size_t, const char *, const char *);
bool adamic_tuple_matches(const adamic_object *, size_t);
// Probe-only until the tuple union read adapter is certified. Owns a present
// normalized reference; false preserves an absent array position.
bool adamic_tuple_array_union_at(const adamic_array *, double, bool, bool, const char *, const char *, adamic_value *);
#endif
