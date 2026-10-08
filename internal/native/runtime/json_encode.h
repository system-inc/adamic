#ifndef ADAMIC_JSON_ENCODE_H
#define ADAMIC_JSON_ENCODE_H
#include "adamic.h"

// An encoder view of the shared JSON schema. Decoder descriptors stay immutable;
// encoder fields hold the same last-shape cache as ordinary emitted field reads.
typedef enum adamic_encode_kind {
	adamic_encode_number,
	adamic_encode_string,
	adamic_encode_boolean,
	adamic_encode_literal,
	adamic_encode_union,
	adamic_encode_array,
	adamic_encode_tuple,
	adamic_encode_object
} adamic_encode_kind;

// The encoder's own memory of where its field was in the last shape it read: runtime's slot cache
// packs a shape and a slot into one word, which this loop's names-only lookup doesn't need.
typedef struct adamic_encode_cache {
	const adamic_shape *shape;
	size_t index;
} adamic_encode_cache;

typedef struct adamic_encode_field {
	const char *name;
	size_t node;
	bool optional;
	adamic_encode_cache cache;
} adamic_encode_field;

typedef struct adamic_encode_node {
	adamic_encode_kind kind;
	int representation;
	const char *expected;
	adamic_string *literal;
	double number;
	bool boolean;
	size_t child_count;
	const size_t *children;
	size_t field_count;
	adamic_encode_field *fields;
	const char *discriminant;
} adamic_encode_node;

typedef struct adamic_encode_schema {
	const adamic_encode_node *nodes;
	size_t root;
} adamic_encode_schema;

adamic_string *adamic_json_encode(adamic_value value, const adamic_encode_schema *schema);
#endif
