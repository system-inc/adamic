#ifndef ADAMIC_JSON_DECODE_H
#define ADAMIC_JSON_DECODE_H
#include "adamic.h"
#define ADAMIC_JSON_DEPTH 128

typedef struct adamic_decode_field {
 const char *name;
 size_t node;
 bool optional;
} adamic_decode_field;
typedef struct adamic_decode_node {
 const char *kind;
 int representation;
 const char *expected;
 adamic_string *literal;
 double number;
 bool boolean;
 size_t child_count;
 const size_t *children;
 size_t field_count;
 const adamic_decode_field *fields;
 const char *discriminant;
} adamic_decode_node;
typedef struct adamic_decode_schema {
 const adamic_decode_node *nodes;
 size_t root;
} adamic_decode_schema;
adamic_object *adamic_json_decode(const adamic_string *text, const adamic_decode_schema *schema);
#endif
