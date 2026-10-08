#ifndef ADAMIC_JSON_METADATA_H
#define ADAMIC_JSON_METADATA_H
#include "json_stringify.h"
extern const adamic_json_runtime adamic_json_runtime_providers;
const adamic_json_schema *adamic_json_describe(const adamic_heap *value);
adamic_json_result adamic_json_array_element(const adamic_array *array, size_t index);
adamic_json_result adamic_json_to_json(adamic_object *receiver, const adamic_string *key);
adamic_json_result adamic_json_resolve(adamic_value value, const adamic_json_schema *schema);
#endif
