#ifndef ADAMIC_JSON_STRINGIFY_H
#define ADAMIC_JSON_STRINGIFY_H
#include "adamic.h"
enum adamic_json_kind {
	adamic_json_undefined, adamic_json_null, adamic_json_number, adamic_json_boolean,
	adamic_json_string, adamic_json_map, adamic_json_function, adamic_json_union,
	adamic_json_maybe_number, adamic_json_maybe_boolean, adamic_json_array,
	adamic_json_tuple, adamic_json_object, adamic_json_toJSON
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
// Encoder-facing interface. Runtime providers are implemented by the runtime owner.
typedef struct adamic_json_result {
    adamic_value value;
    const adamic_json_schema *schema;
} adamic_json_result;
typedef struct adamic_json_runtime {
    const adamic_json_schema *(*describe)(const adamic_heap *value);
    adamic_json_result (*array_element)(const adamic_array *array, size_t index);
    adamic_json_result (*to_json)(adamic_object *receiver, const adamic_string *key);
} adamic_json_runtime;
// describe must cover the actual allocation, including hidden fields and hook presence.
// Descriptors/element reads are borrowed and remain valid throughout encoding.
// The callback key is borrowed for the call; a hook runs once per visited position.
// Hook reference results transfer one owner.
// On throw the provider must return a described undefined value with no owner.
adamic_string *adamic_json_stringify_runtime(adamic_value value, const adamic_json_schema *schema,
    adamic_value replacer, const adamic_json_schema *replacer_schema,
    adamic_value space, const adamic_json_schema *space_schema, const adamic_json_runtime *runtime);
adamic_string *adamic_json_stringify(adamic_value value, const adamic_json_schema *schema,
	adamic_value replacer, const adamic_json_schema *replacer_schema,
	adamic_value space, const adamic_json_schema *space_schema);
#endif
