#ifndef ADAMIC_VIEW_UNIONS_OBJECT_PRIMITIVE_H
#define ADAMIC_VIEW_UNIONS_OBJECT_PRIMITIVE_H
#include "adamic.h"
typedef struct adamic_object_primitive_member {
 unsigned char type;
 bool literal;
 adamic_value value;
} adamic_object_primitive_member;
// Owns the returned union reference. Readiness and storage validation use the
// existing object reader; member fields remain checked at their actual reads.
adamic_heap *adamic_object_primitive_view(const adamic_object *, const char *, adamic_slot_cache *, const adamic_object_primitive_member *, size_t, bool, bool, bool, const char *, const char *);
#endif
