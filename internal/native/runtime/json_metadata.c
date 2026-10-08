// json_metadata.c supplies borrowed storage descriptors; JSON encoding belongs to library.
#include "json_metadata.h"
#include <stdlib.h>

const adamic_json_schema adamic_json_number_schema = {.kind = adamic_json_number};
const adamic_json_schema adamic_json_boolean_schema = {.kind = adamic_json_boolean};
const adamic_json_schema adamic_json_string_schema = {.kind = adamic_json_string};
const adamic_json_schema adamic_json_union_schema = {.kind = adamic_json_union};
const adamic_json_schema adamic_json_maybe_number_schema = {.kind = adamic_json_maybe_number};
const adamic_json_schema adamic_json_undefined_schema = {.kind = adamic_json_undefined};
const adamic_json_schema adamic_json_null_schema = {.kind = adamic_json_null};
static const adamic_json_schema array_schema = {.kind = adamic_json_array};
static const adamic_json_schema map_schema = {.kind = adamic_json_map};
static const adamic_json_schema function_schema = {.kind = adamic_json_function};

const adamic_json_schema *adamic_json_describe(const adamic_heap *value) {
	if (value == &adamic_null)
		return &adamic_json_null_schema;
	if (value == NULL)
		return &adamic_json_undefined_schema;
	switch (value->kind) {
	case adamic_kind_number:
		return &adamic_json_number_schema;
	case adamic_kind_boolean:
		return &adamic_json_boolean_schema;
	case adamic_kind_string:
		return &adamic_json_string_schema;
	case adamic_kind_closure:
		return &function_schema;
	case adamic_kind_map:
		return &map_schema;
	case adamic_kind_array: {
		const adamic_array *array = (const adamic_array *)value;
		// Reject even an empty untyped array. Ownership is never a type descriptor.
		if (array->json_element == NULL)
			return NULL;
		return &array_schema;
	}
	case adamic_kind_object: {
		const adamic_object *object = (const adamic_object *)value;
		// Static own-presence and accessors need their own proof, not a structural view.
		if (object->class != NULL && (object->class->is_static || object->class->accessor_count != 0))
			return NULL;
		return object->shape->json;
	}
	default:
		return NULL;
	}
}

adamic_json_result adamic_json_resolve(adamic_value value, const adamic_json_schema *schema) {
	if (schema == NULL)
		return (adamic_json_result){value, NULL};
	if (schema->null_reference && value.reference == NULL)
		return (adamic_json_result){value, &adamic_json_null_schema};
#ifdef ADAMIC_COUNT
	if (value.reference != NULL) {
		enum adamic_kind expected = 0;
		switch (schema->kind) {
		case adamic_json_string:
			expected = adamic_kind_string;
			break;
		case adamic_json_map:
			expected = adamic_kind_map;
			break;
		case adamic_json_function:
			expected = adamic_kind_closure;
			break;
		case adamic_json_array:
			expected = adamic_kind_array;
			break;
		case adamic_json_object:
		case adamic_json_tuple:
		case adamic_json_toJSON:
			expected = adamic_kind_object;
			break;
		default:
			break;
		}
		if (expected != 0 && ((const adamic_heap *)value.reference)->kind != expected) {
			fprintf(stderr, "adamic: inconsistent JSON descriptor heap kind\n");
			abort();
		}
	}
#endif
	if (schema->kind == adamic_json_maybe_number) {
		adamic_maybe_number number = adamic_maybe_number_unpack(value.number);
		return number.present ? (adamic_json_result){{.number = number.number}, &adamic_json_number_schema}
							  : (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema};
	}
	if (schema->kind == adamic_json_union || schema->kind == adamic_json_maybe_boolean) {
		const adamic_heap *heap = value.reference;
		if (heap == &adamic_null)
			return (adamic_json_result){value, &adamic_json_null_schema};
		if (heap == NULL)
			return (adamic_json_result){value, schema->null_reference ? &adamic_json_null_schema : &adamic_json_undefined_schema};
		if (heap->kind == adamic_kind_number)
			return (adamic_json_result){{.number = ((const adamic_number_box *)heap)->number}, &adamic_json_number_schema};
		if (heap->kind == adamic_kind_boolean)
			return (adamic_json_result){{.boolean = ((const adamic_boolean_box *)heap)->boolean}, &adamic_json_boolean_schema};
		return (adamic_json_result){value, adamic_json_describe(heap)};
	}
	if (schema->kind == adamic_json_string && value.reference == NULL)
		return (adamic_json_result){value, &adamic_json_undefined_schema};
	return (adamic_json_result){value, schema};
}

adamic_json_result adamic_json_array_element(const adamic_array *array, size_t index) {
	// Do not read a slot before its descriptor is established.
	if (adamic_json_describe(&array->heap) == NULL)
		adamic_panic("JSON.stringify runtime array without complete element descriptors", sizeof "JSON.stringify runtime array without complete element descriptors" - 1);
	if (index >= array->length)
		return (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema};
	const adamic_json_schema *schema = array->json_elements == NULL ? array->json_element : array->json_elements[index];
	if (schema == NULL)
		adamic_panic("JSON.stringify runtime array without complete element descriptors", sizeof "JSON.stringify runtime array without complete element descriptors" - 1);
	return adamic_json_resolve(array->elements[index], schema);
}

adamic_json_result adamic_json_to_json(adamic_object *receiver, const adamic_string *key) {
	if (receiver->shape->to_json == NULL)
		adamic_panic("JSON union lacks proven container metadata", sizeof "JSON union lacks proven container metadata" - 1);
	// An element read is borrowed. A reentrant hook may remove its receiver
	// from that container; pin the actual receiver until the call finishes.
	adamic_retain(receiver);
	adamic_json_result result = receiver->shape->to_json(receiver, key);
	adamic_release(receiver);
	return result;
}

const adamic_json_runtime adamic_json_runtime_providers = {
	adamic_json_describe, adamic_json_array_element, adamic_json_to_json};

void adamic_json_array_check(bool references, const adamic_json_schema *schema) {
#ifdef ADAMIC_COUNT
	if (schema == NULL || schema->kind == adamic_json_undefined || schema->kind == adamic_json_null)
		return;
	bool scalar = schema->kind == adamic_json_number || schema->kind == adamic_json_boolean || schema->kind == adamic_json_maybe_number;
	if (references == scalar) {
		fprintf(stderr, "adamic: inconsistent JSON element ownership\n");
		abort();
	}
#else
	(void)references;
	(void)schema;
#endif
}

void adamic_json_shape_check(const adamic_shape *shape) {
#ifdef ADAMIC_COUNT
	if (shape->json == NULL)
		return;
	const adamic_json_schema *schema = shape->json;
	if (schema->count != 0 && schema->fields == NULL)
		abort();
	for (size_t i = 0; i < schema->count; i++) {
		const adamic_json_field *field = &schema->fields[i];
		if (field->slot >= shape->count || field->schema == NULL || shape->kinds == NULL)
			abort();
		enum adamic_json_kind kind = field->schema->kind;
		adamic_field_kind storage = shape->kinds[field->slot];
		if ((kind == adamic_json_boolean && storage != adamic_field_boolean) ||
			((kind == adamic_json_number || kind == adamic_json_maybe_number) && storage != adamic_field_number)) {
			fprintf(stderr, "adamic: inconsistent JSON shape kind for field %s\n", shape->names[field->slot]);
			abort();
		}
		adamic_json_array_check(shape->references[field->slot], field->schema);
	}
#else
	(void)shape;
#endif
}
