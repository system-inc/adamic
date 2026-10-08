// Write only the declared schema, looking up slots by name in the runtime layout.
#include "json_encode.h"
#include <stdlib.h>
#include <string.h>

typedef struct encode_builder {
	char *bytes;
	size_t length, capacity, units;
} encode_builder;

static void append(encode_builder *w, const char *bytes, size_t length, size_t units) {
	adamic_string_check_length((double)w->units + (double)units);
	if (length > SIZE_MAX - w->length) {
		static const char message[] = "RangeError: Invalid string length";
		adamic_panic(message, sizeof message - 1);
	}
	size_t needed = w->length + length;
	if (needed > w->capacity) {
		size_t capacity = w->capacity;
		while (capacity < needed) {
			capacity = capacity > SIZE_MAX / 2 ? needed : capacity * 2;
		}
		char *grown = realloc(w->bytes, capacity);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		w->bytes = grown;
		w->capacity = capacity;
	}
	if (length != 0) {
		memcpy(w->bytes + w->length, bytes, length);
	}
	w->length += length;
	w->units += units;
}
static void ascii(encode_builder *w, const char *text) {
	size_t length = strlen(text);
	append(w, text, length, length);
}
// Strings and descriptor field names are canonical WTF-8. Copy ordinary runs
// once; encoded lone surrogates need the well-formed JSON.stringify escape.
static void quote_bytes(encode_builder *w, const char *bytes, size_t length) {
	static const char hex[] = "0123456789abcdef";
	ascii(w, "\"");
	size_t run = 0;
	size_t units = 0;
	for (size_t at = 0; at < length;) {
		const unsigned char *here = (const unsigned char *)bytes + at;
		unsigned code = here[0];
		size_t width = code < 0x80 ? 1 : code < 0xe0 ? 2 : code < 0xf0 ? 3 : 4;
		bool surrogate = width == 3 && code == 0xed && here[1] >= 0xa0;
		if (code != '"' && code != '\\' && code >= 0x20 && !surrogate) {
			units += width == 4 ? 2 : 1;
			at += width;
			continue;
		}
		append(w, bytes + run, at - run, units);
		if (surrogate) {
			code = ((code & 15) << 12) | ((here[1] & 63) << 6) | (here[2] & 63);
		}
		switch (code) {
		case '"':
			ascii(w, "\\\"");
			break;
		case '\\':
			ascii(w, "\\\\");
			break;
		case '\b':
			ascii(w, "\\b");
			break;
		case '\f':
			ascii(w, "\\f");
			break;
		case '\n':
			ascii(w, "\\n");
			break;
		case '\r':
			ascii(w, "\\r");
			break;
		case '\t':
			ascii(w, "\\t");
			break;
		default: {
			char escape[] = {'\\',
							 'u',
							 hex[(code >> 12) & 15],
							 hex[(code >> 8) & 15],
							 hex[(code >> 4) & 15],
							 hex[code & 15]};
			append(w, escape, sizeof escape, sizeof escape);
			break;
		}
		}
		at += width;
		run = at;
		units = 0;
	}
	append(w, bytes + run, length - run, units);
	ascii(w, "\"");
}
static void quote(encode_builder *w, const adamic_string *text) {
	quote_bytes(w, text->bytes, text->length);
}

// A slot's storage may be an optional scalar, even though its schema names the
// present type. Layout metadata distinguishes scalar slots from undefined NULL.
static bool encode_field(const adamic_object *object, adamic_encode_field *field,
						 const adamic_encode_node *type, adamic_value *value) {
	if (field->cache.shape != object->shape) {
		field->cache.shape = object->shape;
		field->cache.index = object->shape->count;
		for (size_t i = 0; i < object->shape->count; i++) {
			if (strcmp(object->shape->names[i], field->name) == 0) {
				field->cache.index = i;
				break;
			}
		}
	}
	size_t position = field->cache.index;
	if (position == object->shape->count) {
		return false;
	}
	*value = object->slots[position];
	if (!field->optional) {
		return true;
	}
	if (object->shape->references[position]) {
		if (value->reference == NULL) {
			return false;
		}
		if (type->representation == 1) {
			value->number = ((adamic_number_box *)value->reference)->number;
		} else if (type->representation == 2) {
			value->boolean = ((adamic_boolean_box *)value->reference)->boolean;
		}
	} else if (type->representation == 1) {
		adamic_maybe_number number = adamic_maybe_number_unpack(value->number);
		if (!number.present) {
			return false;
		}
		value->number = number.number;
	}
	return true;
}
static bool encode_literal(adamic_value value, int representation, const adamic_encode_node *type) {
	if (representation != type->representation) {
		return false;
	}
	switch (representation) {
	case 1:
		return value.number == type->number;
	case 2:
		return value.boolean == type->boolean;
	case 3:
		return adamic_string_equal(value.reference, type->literal);
	default:
		return false;
	}
}
static void encode_write(encode_builder *builder, adamic_value value,
						 const adamic_encode_schema *schema, size_t index, size_t depth) {
	ADAMIC_CHECK_STACK();
	const adamic_encode_node *type = &schema->nodes[index];
	if (type->kind == adamic_encode_union) {
		int representation = type->representation;
		if (representation == 10) {
			const adamic_heap *heap = value.reference;
			if (heap == NULL) {
				adamic_panic("encodeJson: undefined union",
							 sizeof "encodeJson: undefined union" - 1);
			}
			switch (heap->kind) {
			case adamic_kind_number:
				representation = 1;
				value.number = ((const adamic_number_box *)heap)->number;
				break;
			case adamic_kind_boolean:
				representation = 2;
				value.boolean = ((const adamic_boolean_box *)heap)->boolean;
				break;
			case adamic_kind_string:
				representation = 3;
				break;
			case adamic_kind_object:
				representation = 4;
				break;
			default:
				adamic_panic("encodeJson: invalid union representation",
							 sizeof "encodeJson: invalid union representation" - 1);
			}
		}
		for (size_t i = 0; i < type->child_count; i++) {
			size_t child = type->children[i];
			const adamic_encode_node *member = &schema->nodes[child];
			bool match = member->kind == adamic_encode_literal
							 ? encode_literal(value, representation, member)
							 : representation == member->representation;
			if (match && member->kind == adamic_encode_object && type->discriminant[0] != '\0') {
				match = false;
				for (size_t f = 0; f < member->field_count; f++) {
					adamic_encode_field *field = &member->fields[f];
					if (strcmp(field->name, type->discriminant) != 0) {
						continue;
					}
					const adamic_encode_node *literal = &schema->nodes[field->node];
					adamic_value tag = {0};
					match = encode_field(value.reference, field, literal, &tag) &&
							encode_literal(tag, literal->representation, literal);
					break;
				}
			}
			if (match) {
				encode_write(builder, value, schema, child, depth);
				return;
			}
		}
		adamic_panic("encodeJson: value does not match its declared union",
					 sizeof "encodeJson: value does not match its declared union" - 1);
	}
	if (type->kind == adamic_encode_array || type->kind == adamic_encode_tuple ||
		type->kind == adamic_encode_object) {
		if (type->kind == adamic_encode_array) {
			const adamic_array *array = value.reference;
			ascii(builder, "[");
			for (size_t i = 0; i < array->length; i++) {
				if (i != 0) {
					ascii(builder, ",");
				}
				encode_write(builder, array->elements[i], schema, type->children[0], depth + 1);
			}
			ascii(builder, "]");
			return;
		}
		bool tuple = type->kind == adamic_encode_tuple;
		const adamic_object *object = value.reference;
		ascii(builder, tuple ? "[" : "{");
		size_t written = 0;
		for (size_t i = 0; i < type->field_count; i++) {
			adamic_encode_field *field = &type->fields[i];
			adamic_value child = {0};
			if (!encode_field(object, field, &schema->nodes[field->node], &child)) {
				if (field->optional) {
					continue;
				}
				adamic_panic("encodeJson: required field missing",
							 sizeof "encodeJson: required field missing" - 1);
			}
			if (written++ != 0) {
				ascii(builder, ",");
			}
			if (!tuple) {
				quote_bytes(builder, field->name, strlen(field->name));
				ascii(builder, ":");
			}
			encode_write(builder, child, schema, field->node, depth + 1);
		}
		ascii(builder, tuple ? "]" : "}");
		return;
	}
	switch (type->representation) {
	case 1: {
		if (!isfinite(value.number)) {
			ascii(builder, "null");
			return;
		}
		char bytes[ADAMIC_NUMBER_FORMAT_MAX];
		size_t length = adamic_number_format(value.number, bytes);
		append(builder, bytes, length, length);
		return;
	}
	case 2:
		ascii(builder, value.boolean ? "true" : "false");
		return;
	case 3:
		quote(builder, value.reference);
		return;
	default:
		adamic_panic("encodeJson: invalid descriptor", sizeof "encodeJson: invalid descriptor" - 1);
	}
}
adamic_string *adamic_json_encode(adamic_value value, const adamic_encode_schema *schema) {
	encode_builder builder = {NULL, 0, 64, 0};
	builder.bytes = malloc(builder.capacity);
	if (builder.bytes == NULL) {
		adamic_panic("out of memory", sizeof "out of memory" - 1);
	}
	encode_write(&builder, value, schema, schema->root, 0);
	adamic_string *result = adamic_string_allocate(builder.length);
	if (builder.length != 0) {
		memcpy((char *)result->bytes, builder.bytes, builder.length);
	}
	// The units-plus-one cache also selects the ASCII fast path when units equal bytes.
	result->units = builder.units + 1;
	free(builder.bytes);
	return result;
}
