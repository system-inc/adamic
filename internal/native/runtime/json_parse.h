#ifndef ADAMIC_JSON_PARSE_H
#define ADAMIC_JSON_PARSE_H
#include "json_stringify.h"
enum adamic_json_parse_mode {
 adamic_json_parse_discard, adamic_json_parse_number, adamic_json_parse_boolean,
 adamic_json_parse_string, adamic_json_parse_canonical
};
adamic_value adamic_json_parse(const adamic_string *text, adamic_closure *reviver,
 enum adamic_json_kind returned, size_t takes, enum adamic_json_parse_mode mode);
#endif
