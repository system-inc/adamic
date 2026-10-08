#ifndef ADAMIC_VIEW_INTERSECTIONS_RECURSIVE_H
#define ADAMIC_VIEW_INTERSECTIONS_RECURSIVE_H
typedef struct {
 unsigned char type;
 double number;
 bool boolean;
 const char *string;
 size_t length;
} adamic_intersection_literal;
typedef struct {
 const char *name;
 const char *expected;
 unsigned char type;
 bool optional;
 size_t child;
 const adamic_intersection_literal *allowed;
 size_t allowed_count;
} adamic_intersection_field;
typedef struct {
 const adamic_intersection_literal *allowed;
 size_t allowed_count;
 size_t contract;
} adamic_intersection_arm;
// A contract with a tag selects one arm by that literal field; otherwise it
// checks its fields. An arm without a contract checks only the tag.
typedef struct {
 const adamic_intersection_field *fields;
 size_t count;
 const char *tag;
 const char *tag_expected;
 unsigned char tag_type;
 const adamic_intersection_arm *arms;
 size_t arm_count;
} adamic_intersection_contract;
void adamic_intersection_recursive_require(const adamic_object *object, const adamic_intersection_contract *contracts, size_t id, const char *path);
void adamic_intersection_bounded_require(const adamic_object *object, const adamic_intersection_contract *contracts, size_t id, const char *path);
#endif
