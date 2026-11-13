// json_objects.c: JSON.stringify of an object reference (#d9ep7de), the production describe provider
// json_stringify.c asks when a value lowering couldn't describe statically is an object or an array.
// It answers from the allocation itself, never from a declared type, since a cast can make a type lie
// about the fields behind it (reland_refused/view_widened_held_json.ts): the actual shape's names, in
// Node's order, and each slot read by its own representation byte, as union.c's dynamic_slot reads
// one. Whatever it can't tell, it stops at, by name, rather than guess.
#include "json_stringify.h"
#include "json_error_stop.h"

#include <stdlib.h>
#include <string.h>

// json_described is one object's descriptor, with its fields' names, kept until the stringify that
// asked for it returns: the encoder borrows a descriptor for the whole walk (json_stringify.h).
typedef struct json_described {
	struct json_described *next;
	adamic_json_schema schema;
	adamic_json_field fields[];
} json_described;

// The descriptors the running stringify made, on this thread. Thread-local, since parallel tasks
// stringify too; a stringify runs no program code (toJSON and replacer functions are refused), so
// one never starts inside another on a thread.
static _Thread_local json_described *described;

static const adamic_json_schema json_number = {adamic_json_number, NULL, 0, NULL};
static const adamic_json_schema json_maybe_number = {adamic_json_maybe_number, NULL, 0, NULL};
static const adamic_json_schema json_boolean = {adamic_json_boolean, NULL, 0, NULL};
static const adamic_json_schema json_null = {adamic_json_null, NULL, 0, NULL};
static const adamic_json_schema json_undefined = {adamic_json_undefined, NULL, 0, NULL};
// A reference names its own kind by its heap header, and &adamic_null is null: the union reading.
static const adamic_json_schema json_reference = {adamic_json_union, NULL, 0, NULL};
static const adamic_json_schema json_number_array = {adamic_json_array, &json_number, 0, NULL};
static const adamic_json_schema json_boolean_array = {adamic_json_array, &json_boolean, 0, NULL};
static const adamic_json_schema json_maybe_number_array = {adamic_json_array, &json_maybe_number, 0, NULL};
static const adamic_json_schema json_reference_array = {adamic_json_array, &json_reference, 0, NULL};

_Noreturn static void stop(const char *message) {
	adamic_panic(message, strlen(message));
}

// The array's element byte says how its slots are held (adamic.h); every reference array's is Union,
// whose null and undefined elements both print null, as Node prints them in an array.
static const adamic_json_schema *describe_array(const adamic_array *array) {
	switch (array->element) {
	case adamic_ir_Number: return &json_number_array;
	case adamic_ir_Boolean: return &json_boolean_array;
	case adamic_ir_MaybeNumber: return &json_maybe_number_array;
	case adamic_ir_Union: return &json_reference_array;
	default: stop("NotYet: JSON.stringify an array whose element representation isn't recorded");
	}
}

// field_schema is how one present, initialized slot prints, by its storage kind and the byte its last
// write left (emit_objects.go's fieldRepresentation, written in every program that stringifies an
// object reference).
static const adamic_json_schema *field_schema(const adamic_object *object, size_t index) {
	unsigned char type = adamic_object_field_type(object, index);
	if (object->shape->references[index]) {
		if (object->slots[index].reference != NULL) return &json_reference;
		switch (type) {
		case adamic_ir_NullRepresentation: return &json_null;
		case adamic_ir_UndefinedRepresentation: return &json_undefined;
		// A string, a union, a weak link and a record are each boxed where they may be null
		// (lowering's representation), so a null pointer held as one is undefined.
		case adamic_ir_String: case adamic_ir_Union: case adamic_ir_Weak: case adamic_ir_Record: return &json_undefined;
		default:
			// Nullable bare-pointer writes record their semantic nullish tag at run time.
			// A cast can still hide the value's nullish contract from the writer, so an
			// allocation without that tag keeps this backstop, pinned by
			// testdata/json-objects-runtime/harness.c (TestJSONObjectsNullPointers).
			stop("NotYet: JSON.stringify a reference field holding null or undefined that the allocation doesn't record which");
		}
	}
	if (object->shape->kinds == NULL) stop("NotYet: JSON.stringify an object whose shape records no storage kinds");
	if (object->shape->kinds[index] == adamic_field_boolean) return &json_boolean;
	switch (type) {
	case adamic_ir_Number: return &json_number;
	case adamic_ir_MaybeNumber: return &json_maybe_number;
	// Zero is no evidence: an unowned implementation word is held as a number, and a number slot
	// nothing recorded can't be told from one.
	case 0: stop("NotYet: JSON.stringify an object holding an implementation word (a number slot with no representation byte)");
	default: stop("NotYet: JSON.stringify a field with a two-word representation");
	}
}

static bool named(const char *name, const char *const *names, size_t count) {
	for (size_t index = 0; index < count; index++) {
		if (strcmp(name, names[index]) == 0) return true;
	}
	return false;
}

// describe_object stops by name on what Node prints by other rules than its own fields, before it
// reads a name: a host object (a Date, a Stats, a collator), an unrecorded Error, a class itself, and anything
// with a toJSON, a class's method, inherited ones included (lowering copies a base's methods into
// each subclass's table), or an own field of that name however it got there.
static const adamic_json_schema *describe_object(const adamic_object *object) {
	if (adamic_record_is(object)) stop("NotYet: JSON.stringify a record held by an object");
	if (object->tuple) stop("NotYet: JSON.stringify a tuple held by an object");
	const adamic_shape *shape = object->shape;
	if (shape->methods == &adamic_host_methods) stop("NotYet: JSON.stringify a runtime library value (a Date, a Stats or another whose own fields aren't Node's)");
	const adamic_error_json *error = NULL;
	if (adamic_error_is(object)) {
		// An inherited empty message is held as #message by absent_message_shape; writing it
		// creates an enumerable own property on Node. A constructed message is named message
		// and stays own and non-enumerable after a write. Only name and #message need their
		// construction bytes intact; the ordinary message byte does not decide this stop.
		for (size_t index = 0; index < shape->count; index++) {
			const char *name = shape->names[index];
			if ((strcmp(name, "name") == 0 || strcmp(name, "#message") == 0) &&
				adamic_object_field_types(object)[index] != adamic_error_unwritten)
				stop(ADAMIC_JSON_ERROR_WRITE_STOP);
		}
		error = adamic_error_json_lookup(shape);
		if (error == NULL) stop("NotYet: JSON.stringify an Error whose own keys aren't recorded");
	}
	if (object->class != NULL && object->class->is_static) stop("NotYet: JSON.stringify a class itself");
	if (shape->methods != NULL && named("toJSON", shape->methods->names, shape->methods->count)) stop("NotYet: JSON.stringify an object whose class has a toJSON method");
	if (named("toJSON", shape->names, shape->count)) stop("NotYet: JSON.stringify an object with its own toJSON field");
	json_described *entry = malloc(sizeof *entry + (shape->count == 0 ? 1 : shape->count) * sizeof entry->fields[0]);
	if (entry == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
	entry->schema = (adamic_json_schema){adamic_json_object, NULL, 0, entry->fields};
	entry->next = described;
	described = entry;
	// Node's order: a class instance's public fields as Object.keys gives them (class_features.c),
	// anything else's index keys first, then its names in the order they were written.
	const adamic_shape *public_shape = object->class != NULL && object->write_order == NULL ? object->class->public_shape : NULL;
	size_t *order = error == NULL && public_shape == NULL ? adamic_object_ordered(object) : NULL;
	size_t count = error != NULL ? error->count : public_shape == NULL ? shape->count : public_shape->count;
	for (size_t position = 0; position < count; position++) {
		size_t index = shape->count;
		if (error != NULL) {
			for (size_t at = 0; at < shape->count; at++) {
				if (strcmp(error->keys[position], shape->names[at]) == 0) { index = at; break; }
			}
			if (index == shape->count) stop("NotYet: JSON.stringify an Error whose recorded key has no slot");
		} else if (public_shape == NULL) {
			index = order[position];
		} else {
			const char *name = public_shape->names[adamic_public_index(public_shape, position)];
			for (size_t at = 0; at < shape->count; at++) {
				if (strcmp(name, shape->names[at]) == 0) { index = at; break; }
			}
			if (index == shape->count) stop("NotYet: JSON.stringify a class instance missing a public field of its class");
		}
		const char *name = shape->names[index];
		if (name[0] == '#' || !adamic_object_present(object, index) || !adamic_object_initialized(object)[index]) continue;
		const adamic_json_schema *child = field_schema(object, index);
		size_t length = strlen(name);
		adamic_string *key = adamic_string_allocate(length);
		memcpy((char *)key->bytes, name, length);
		entry->fields[entry->schema.count++] = (adamic_json_field){key, index, child};
	}
	free(order);
	return &entry->schema;
}

static const adamic_json_schema *describe(const adamic_heap *value) {
	if (value->kind == adamic_kind_array) return describe_array((const adamic_array *)value);
	return describe_object((const adamic_object *)value);
}

static const adamic_json_runtime objects = {describe, NULL, NULL};

adamic_string *adamic_json_stringify_objects(adamic_value value, const adamic_json_schema *schema,
	adamic_value replacer, const adamic_json_schema *replacer_schema,
	adamic_value space, const adamic_json_schema *space_schema) {
	json_described *outer = described;
	described = NULL;
	adamic_string *result = adamic_json_stringify_runtime(value, schema, replacer, replacer_schema, space, space_schema, &objects);
	while (described != NULL) {
		json_described *next = described->next;
		for (size_t index = 0; index < described->schema.count; index++) {
			adamic_release(described->fields[index].name);
		}
		free(described);
		described = next;
	}
	described = outer;
	return result;
}
