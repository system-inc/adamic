// JSON.stringify for the complete representations lowering has proved (library_json_stringify.go).
// Descriptors for literals describe the actual object, never just a structural view of it.
#include "json_stringify.h"

#include <stdlib.h>
#include <string.h>

typedef struct json_writer {
	char *bytes;
	size_t length;
	size_t capacity;
	size_t units;
	adamic_string *gap;
	adamic_string **keys;
	size_t key_count;
	bool key_list;
	bool pretty;
} json_writer;

static void *json_memory(size_t size) {
	void *value = malloc(size);
	if (value == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return value;
}
static void append(json_writer *w, const char *bytes, size_t length, size_t units) {
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
	if (length != 0) { memcpy(w->bytes + w->length, bytes, length); }
	w->length += length;
	w->units += units;
}
static void ascii(json_writer *w, const char *text) {
	size_t length = strlen(text);
	append(w, text, length, length);
}
static void code_point(json_writer *w, unsigned code) {
	char bytes[4];
	size_t count;
	if (code < 0x80) { bytes[0] = (char)code; count = 1; }
	else if (code < 0x800) {
		bytes[0] = (char)(0xc0 | (code >> 6)); bytes[1] = (char)(0x80 | (code & 63)); count = 2;
	} else if (code < 0x10000) {
		bytes[0] = (char)(0xe0 | (code >> 12)); bytes[1] = (char)(0x80 | ((code >> 6) & 63)); bytes[2] = (char)(0x80 | (code & 63)); count = 3;
	} else {
		bytes[0] = (char)(0xf0 | (code >> 18)); bytes[1] = (char)(0x80 | ((code >> 12) & 63));
		bytes[2] = (char)(0x80 | ((code >> 6) & 63)); bytes[3] = (char)(0x80 | (code & 63)); count = 4;
	}
	append(w, bytes, count, code >= 0x10000 ? 2 : 1);
}
static void quote(json_writer *w, const adamic_string *text) {
	static const char hex[] = "0123456789abcdef";
	ascii(w, "\"");
	size_t length = (size_t)adamic_string_length(text);
	for (size_t index = 0; index < length; index++) {
		unsigned code = (unsigned)adamic_string_char_code_at(text, (double)index);
		switch (code) {
		case '"': ascii(w, "\\\""); continue;
		case '\\': ascii(w, "\\\\"); continue;
		case '\b': ascii(w, "\\b"); continue;
		case '\f': ascii(w, "\\f"); continue;
		case '\n': ascii(w, "\\n"); continue;
		case '\r': ascii(w, "\\r"); continue;
		case '\t': ascii(w, "\\t"); continue;
		default: break;
		}
		if (code >= 0xd800 && code <= 0xdbff && index + 1 < length) {
			unsigned low = (unsigned)adamic_string_char_code_at(text, (double)(index + 1));
			if (low >= 0xdc00 && low <= 0xdfff) {
				code_point(w, 0x10000 + ((code - 0xd800) << 10) + low - 0xdc00);
				index++;
				continue;
			}
		}
		if (code < 0x20 || (code >= 0xd800 && code <= 0xdfff)) {
			char escape[] = {'\\', 'u', hex[(code >> 12) & 15], hex[(code >> 8) & 15], hex[(code >> 4) & 15], hex[code & 15]};
			append(w, escape, sizeof escape, sizeof escape);
		} else { code_point(w, code); }
	}
	ascii(w, "\"");
}

typedef struct json_scalar { enum adamic_json_kind kind; adamic_value value; } json_scalar;
static json_scalar scalar(adamic_value value, const adamic_json_schema *schema) {
	enum adamic_json_kind kind = schema->kind;
	if (schema->null_reference && value.reference == NULL) {
		return (json_scalar){adamic_json_null, value};
	}
	if (kind == adamic_json_maybe_number) {
		adamic_maybe_number number = adamic_maybe_number_unpack(value.number);
		kind = number.present ? adamic_json_number : adamic_json_undefined;
		value.number = number.number;
	} else if (kind == adamic_json_union || kind == adamic_json_maybe_boolean) {
		adamic_heap *reference = value.reference;
		if (reference == NULL) { kind = adamic_json_undefined; }
		else {
			switch (reference->kind) {
			case adamic_kind_number: kind = adamic_json_number; value.number = ((adamic_number_box *)reference)->number; break;
			case adamic_kind_boolean: kind = adamic_json_boolean; value.boolean = ((adamic_boolean_box *)reference)->boolean; break;
			case adamic_kind_string: kind = adamic_json_string; break;
			case adamic_kind_map: kind = adamic_json_map; break;
			case adamic_kind_closure: kind = adamic_json_function; break;
			default: {
				static const char message[] = "JSON union lacks proven container metadata";
				adamic_panic(message, sizeof message - 1);
			}
			}
		}
	} else if ((kind == adamic_json_string || kind == adamic_json_map || kind == adamic_json_function || kind == adamic_json_array) && value.reference == NULL) {
		kind = adamic_json_undefined;
	}
	return (json_scalar){kind, value};
}
static void indent(json_writer *w, size_t depth) {
	if (!w->pretty) { return; }
	ascii(w, "\n");
	for (size_t index = 0; index < depth; index++) {
		append(w, w->gap->bytes, w->gap->length, (size_t)adamic_string_length(w->gap));
	}
}
static bool write_value(json_writer *w, adamic_value value, const adamic_json_schema *schema, size_t depth);
static void write_field(json_writer *w, const adamic_object *object, const adamic_json_field *field, size_t depth, size_t *written) {
	// Undefined and function-valued object fields are omitted, rather than becoming null.
	json_scalar s = scalar(object->slots[field->slot], field->schema);
	if (s.kind == adamic_json_undefined || s.kind == adamic_json_function) { return; }
	if ((*written)++ != 0) { ascii(w, ","); }
	indent(w, depth + 1);
	quote(w, field->name);
	ascii(w, !w->pretty ? ":" : ": ");
	(void)write_value(w, object->slots[field->slot], field->schema, depth + 1);
}
static bool write_value(json_writer *w, adamic_value value, const adamic_json_schema *schema, size_t depth) {
	ADAMIC_CHECK_STACK();
	json_scalar s = scalar(value, schema);
	switch (s.kind) {
	case adamic_json_undefined: case adamic_json_function: return false;
	case adamic_json_null: ascii(w, "null"); return true;
	case adamic_json_number: {
		if (!isfinite(s.value.number)) { ascii(w, "null"); return true; }
		char bytes[ADAMIC_NUMBER_FORMAT_MAX];
		size_t length = adamic_number_format(s.value.number, bytes);
		append(w, bytes, length, length);
		return true;
	}
	case adamic_json_boolean: ascii(w, s.value.boolean ? "true" : "false"); return true;
	case adamic_json_string: quote(w, s.value.reference); return true;
	case adamic_json_date: {
		adamic_string *text = adamic_date_json(s.value.reference);
		if (text == NULL) { ascii(w, "null"); } else { quote(w, text); adamic_release(text); }
		return true;
	}
	case adamic_json_map: ascii(w, "{}"); return true;
	case adamic_json_array: case adamic_json_tuple: {
		bool tuple = s.kind == adamic_json_tuple;
		const adamic_array *array = value.reference;
		const adamic_object *object = value.reference;
		size_t count = tuple ? schema->count : array->length;
		ascii(w, "[");
		for (size_t index = 0; index < count; index++) {
			if (index != 0) { ascii(w, ","); }
			indent(w, depth + 1);
			const adamic_json_schema *element = tuple ? schema->fields[index].schema : schema->element;
			adamic_value item = tuple ? object->slots[schema->fields[index].slot] : array->elements[index];
			if (!write_value(w, item, element, depth + 1)) { ascii(w, "null"); }
		}
		if (count != 0) { indent(w, depth); }
		ascii(w, "]");
		return true;
	}
	case adamic_json_object: {
		const adamic_object *object = value.reference;
		ascii(w, "{");
		size_t written = 0;
		if (w->key_list) {
			for (size_t key = 0; key < w->key_count; key++) {
				for (size_t index = 0; index < schema->count; index++) {
					const adamic_json_field *field = &schema->fields[index];
					if (adamic_string_equal(w->keys[key], field->name)) { write_field(w, object, field, depth, &written); break; }
				}
			}
		} else {
			for (size_t index = 0; index < schema->count; index++) { write_field(w, object, &schema->fields[index], depth, &written); }
		}
		if (written != 0) { indent(w, depth); }
		ascii(w, "}");
		return true;
	}
	default: {
		static const char message[] = "JSON descriptor has an unknown representation";
		adamic_panic(message, sizeof message - 1);
	}
	}
}
static void keys(json_writer *w, adamic_value value, const adamic_json_schema *schema) {
	if (schema == NULL || (schema->kind != adamic_json_array && schema->kind != adamic_json_tuple) || value.reference == NULL) { return; }
	w->key_list = true;
	bool tuple = schema->kind == adamic_json_tuple;
	const adamic_array *array = value.reference;
	const adamic_object *object = value.reference;
	size_t count = tuple ? schema->count : array->length;
	if (count != 0) { w->keys = json_memory(count * sizeof *w->keys); }
	for (size_t index = 0; index < count; index++) {
		const adamic_json_schema *element = tuple ? schema->fields[index].schema : schema->element;
		adamic_value item = tuple ? object->slots[schema->fields[index].slot] : array->elements[index];
		json_scalar s = scalar(item, element);
		adamic_string *key;
		if (s.kind == adamic_json_string) { key = adamic_retain(s.value.reference); }
		else if (s.kind == adamic_json_number) { key = adamic_string_from_number(s.value.number); }
		else { continue; }
		bool duplicate = false;
		for (size_t previous = 0; previous < w->key_count; previous++) {
			if (adamic_string_equal(key, w->keys[previous])) { duplicate = true; break; }
		}
		if (duplicate) { adamic_release(key); } else { w->keys[w->key_count++] = key; }
	}
}
adamic_string *adamic_json_stringify(adamic_value value, const adamic_json_schema *schema,
	adamic_value replacer, const adamic_json_schema *replacer_schema,
	adamic_value space, const adamic_json_schema *space_schema) {
	static adamic_string empty = ADAMIC_STRING("");
	static adamic_string blanks = ADAMIC_STRING("          ");
	adamic_string *gap = &empty;
	bool pretty = false;
	if (space_schema != NULL) {
		json_scalar s = scalar(space, space_schema);
		if (s.kind == adamic_json_string) {
			const adamic_string *text = s.value.reference;
			double width = fmin(10, adamic_string_length(text));
			pretty = width != 0;
			// ECMA-262 25.5.2 sets gap to the first ten code units, including NUL. V8's
			// SerializeJSONProperty indentation buffer instead ends at its first NUL.
			// Match Node, retaining the multiline path even when NUL is the first unit.
			for (double index = 0; index < width; index++) {
				if (adamic_string_char_code_at(text, index) == 0) { width = index; break; }
			}
			gap = adamic_string_slice((adamic_string *)text, 0, width, true);
		}
		else if (s.kind == adamic_json_number) {
			// Node 24.19 (V8) enables its multiline path before truncating a positive fraction.
			// Thus 0 < space < 1 writes newlines with an empty gap, as the Node oracle observes.
			pretty = s.value.number > 0;
			double width = isnan(s.value.number) || s.value.number < 0 ? 0 : s.value.number > 10 ? 10 : trunc(s.value.number);
			gap = adamic_string_slice(&blanks, 0, width, true);
		}
	}
	json_writer writer = {json_memory(64), 0, 64, 0, gap, NULL, 0, false, pretty};
	keys(&writer, replacer, replacer_schema);
	bool present = write_value(&writer, value, schema, 0);
	adamic_string *result = NULL;
	if (present) {
		result = adamic_string_allocate(writer.length);
		if (writer.length != 0) { memcpy((char *)result->bytes, writer.bytes, writer.length); }
		// Repeated gaps may join a high surrogate at one gap's end to the next gap's low.
		result->length = adamic_string_join_halves((char *)result->bytes, 0, writer.length);
	}
	for (size_t index = 0; index < writer.key_count; index++) { adamic_release(writer.keys[index]); }
	free(writer.keys);
	free(writer.bytes);
	adamic_release(gap);
	return result;
}
