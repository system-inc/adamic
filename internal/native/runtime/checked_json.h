#ifndef ADAMIC_CHECKED_JSON_H
#define ADAMIC_CHECKED_JSON_H
#include "adamic.h"
typedef struct adamic_json_contract adamic_json_contract;
typedef struct { const char *name; bool optional; const adamic_json_contract *contract; } adamic_json_contract_field;
struct adamic_json_contract {
 const char *kind;
 const char *name;
 const adamic_json_contract *element;
 size_t count;
 const adamic_json_contract_field *fields;
 const adamic_json_contract *const *alternatives;
 const char *literal_text;
 size_t literal_length;
 double literal_number;
 bool literal_boolean;
};
void adamic_check_json(adamic_heap *value, const adamic_json_contract *contract, const char *path);
adamic_heap *adamic_json_array_get(adamic_array *array, double index);
#endif
