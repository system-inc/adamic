#ifndef ADAMIC_OBJECT_REFLECTION_H
#define ADAMIC_OBJECT_REFLECTION_H
#include "adamic.h"

typedef struct adamic_reflection_member {
 int kind;
 bool literal;
 const char *text;
 size_t length;
 double number;
 bool boolean;
} adamic_reflection_member;
typedef struct adamic_reflection_layout { const int *kinds; const size_t *lengths; } adamic_reflection_layout;
typedef const adamic_reflection_layout *(*adamic_reflection_types)(const adamic_shape *);
adamic_array *adamic_checked_entries(const adamic_object *, int, size_t, const adamic_reflection_member *, adamic_reflection_types, bool, const char *);
void adamic_reflection_assign(adamic_object *, const adamic_object *, adamic_reflection_types, bool);
#endif
