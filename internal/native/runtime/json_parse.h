#ifndef ADAMIC_JSON_PARSE_H
#define ADAMIC_JSON_PARSE_H
#include "adamic.h"
enum json_parse_kind {
  json_raw,
  json_number,
  json_boolean,
  json_string,
  json_null,
  json_undefined,
  json_array,
  json_object,
  json_union
};
typedef struct adamic_json_parse_schema adamic_json_parse_schema;
typedef struct adamic_json_parse_field {
  adamic_string *name;
  const adamic_json_parse_schema *schema;
} adamic_json_parse_field;
struct adamic_json_parse_schema {
  enum json_parse_kind kind;
  int of;
  adamic_string *name;
  bool optional;
  adamic_string *literal;
  const adamic_json_parse_schema *element;
  size_t count;
  const adamic_json_parse_field *fields;
  const adamic_json_parse_schema *const *members;
  const adamic_shape *shape;
};
adamic_value adamic_json_parse(const adamic_string *text,
                               const adamic_json_parse_schema *check,
                               const adamic_json_parse_schema *layout);
#endif
