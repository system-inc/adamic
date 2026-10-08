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
typedef struct adamic_reflection_field {
 const char *name;
 bool index;
 size_t length;
 size_t count;
 const adamic_reflection_member *members;
} adamic_reflection_field;
typedef struct adamic_reflection_layout { const int *kinds; const size_t *lengths; } adamic_reflection_layout;
typedef const adamic_reflection_layout *(*adamic_reflection_types)(const adamic_shape *);
adamic_array *adamic_checked_entries(const adamic_object *, int, size_t, const adamic_reflection_member *, adamic_reflection_types, bool, const char *);
void adamic_checked_assign(adamic_object *, const adamic_object *, size_t, const adamic_reflection_field *, size_t, const adamic_reflection_field *, adamic_reflection_types, bool, const char *);
#endif
