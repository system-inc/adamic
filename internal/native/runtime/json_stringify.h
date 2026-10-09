#ifndef ADAMIC_JSON_STRINGIFY_H
#define ADAMIC_JSON_STRINGIFY_H
#include "adamic.h"
enum adamic_json_kind {
	adamic_json_undefined, adamic_json_null, adamic_json_number, adamic_json_boolean,
	adamic_json_string, adamic_json_nullable_string, adamic_json_date, adamic_json_map, adamic_json_function, adamic_json_union,
	adamic_json_maybe_number, adamic_json_maybe_boolean, adamic_json_array,
	adamic_json_tuple, adamic_json_object
};
typedef struct adamic_json_schema adamic_json_schema;
typedef struct adamic_json_field {
	adamic_string *name;
	size_t slot;
	const adamic_json_schema *schema;
} adamic_json_field;
struct adamic_json_schema {
	enum adamic_json_kind kind;
	const adamic_json_schema *element;
	size_t count;
	const adamic_json_field *fields;
	bool null_reference;
};
adamic_string *adamic_json_stringify(adamic_value value, const adamic_json_schema *schema,
	adamic_value replacer, const adamic_json_schema *replacer_schema,
	adamic_value space, const adamic_json_schema *space_schema);
#endif
