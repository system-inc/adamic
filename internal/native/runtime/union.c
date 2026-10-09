// union.c: a union whose members are held differently, as one counted reference (adamic.h).

#include "adamic.h"

#include <math.h>
#include <string.h>

adamic_heap adamic_null = {0, adamic_kind_object, 0};

adamic_boolean_box adamic_box_true = {{0, adamic_kind_boolean, 0}, true};
adamic_boolean_box adamic_box_false = {{0, adamic_kind_boolean, 0}, false};

adamic_string adamic_typeof_number = ADAMIC_STRING("number");
adamic_string adamic_typeof_string = ADAMIC_STRING("string");
adamic_string adamic_typeof_boolean = ADAMIC_STRING("boolean");
adamic_string adamic_typeof_undefined = ADAMIC_STRING("undefined");
adamic_string adamic_typeof_object = ADAMIC_STRING("object");
adamic_string adamic_typeof_function = ADAMIC_STRING("function");

adamic_heap *adamic_box_number(double number) {
	adamic_number_box *box = adamic_allocate(sizeof *box, adamic_kind_number);
	box->number = number;
	return &box->heap;
}

bool adamic_union_equal(const adamic_heap *left, const adamic_heap *right) {
	if (left == NULL || right == NULL) {
		return left == right;
	}
	if (left->kind != right->kind) {
		return false;
	}
	switch (left->kind) {
	case adamic_kind_number:
		// Compared as numbers, never as boxes: NaN is never ===, even to itself, and -0 === 0.
		return ((const adamic_number_box *)left)->number == ((const adamic_number_box *)right)->number;
	case adamic_kind_string:
		return adamic_string_equal((const adamic_string *)left, (const adamic_string *)right);
	default:
		// A boolean is one of two boxes, and an object, array, map or function is equal only to itself.
		return left == right;
	}
}

adamic_string *adamic_union_to_string(adamic_heap *value) {
	if (value == NULL) {
		return &adamic_string_undefined;
	}
	switch (value->kind) {
	case adamic_kind_number:
		return adamic_string_from_number(((adamic_number_box *)value)->number);
	case adamic_kind_boolean:
		return ((adamic_boolean_box *)value)->boolean ? &adamic_string_true : &adamic_string_false;
	case adamic_kind_string:
		return adamic_retain(value);
	default: {
		// Lowering lets only numbers, booleans, strings and undefined through; anything else here is a
		// compiler bug, said out loud.
		static const char message[] = "a union holding an object was written as a string";
		adamic_panic(message, sizeof message - 1);
	}
	}
}

adamic_string *adamic_union_typeof(const adamic_heap *value, bool null) {
	if (value == &adamic_null) { return &adamic_typeof_object; }
	if (value == NULL) {
		return null ? &adamic_typeof_object : &adamic_typeof_undefined;
	}
	switch (value->kind) {
	case adamic_kind_number:
		return &adamic_typeof_number;
	case adamic_kind_boolean:
		return &adamic_typeof_boolean;
	case adamic_kind_string:
		return &adamic_typeof_string;
	case adamic_kind_closure:
		return &adamic_typeof_function;
	case adamic_kind_object: {
		const adamic_object *object = (const adamic_object *)value;
		return object->class != NULL && object->class->is_static ? &adamic_typeof_function : &adamic_typeof_object;
	}
	default:
		return &adamic_typeof_object;
	}
}

// Program shape metadata is static, registered when an object of the shape is made.
static adamic_shape_types *shape_types;

void adamic_register_shape_types(adamic_shape_types *metadata) {
	for (adamic_shape_types *entry = shape_types; entry != NULL; entry = entry->next) {
		if (entry == metadata) return;
	}
	metadata->next = shape_types;
	shape_types = metadata;
}

static bool named(const char *name, const char *const *names, size_t count) {
	for (size_t index = 0; index < count; index++) {
		if (strcmp(name, names[index]) == 0) return true;
	}
	return false;
}

static bool object_prototype_name(const char *name) {
	static const char *const names[] = {"constructor", "__proto__", "toString", "toLocaleString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__"};
	return named(name, names, sizeof names / sizeof names[0]);
}

static bool array_key(const char *name, size_t *result) {
	if (*name == '\0' || (name[0] == '0' && name[1] != '\0')) return false;
	uint64_t index = 0;
	for (const unsigned char *at = (const unsigned char *)name; *at != '\0'; at++) {
		if (*at < '0' || *at > '9') return false;
		index = index * 10 + (*at - '0');
		if (index >= UINT32_MAX) return false;
	}
	*result = (size_t)index;
	return true;
}

bool adamic_has_property(const adamic_heap *value, const char *name) {
	if (value == NULL || value == &adamic_null) {
		adamic_panic("in on null or undefined", sizeof "in on null or undefined" - 1);
	}
	if (value->kind == adamic_kind_object) {
		const adamic_object *object = (const adamic_object *)value;
		if (adamic_record_is(object)) {
			adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
			return adamic_record_has(object, &key);
		}
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(name, object->shape->names[index]) != 0) continue;
			if (object->class == NULL || name[0] != '#') return true;
			const adamic_shape *public = object->class->public_shape;
			if (named(name, public->names, public->count)) return true;
		}
		const adamic_methods *methods = object->shape->methods;
		if (methods != NULL && named(name, methods->names, methods->count)) return true;
		for (const adamic_class *class = object->class; class != NULL; class = class->base) {
			for (size_t index = 0; index < class->accessor_count; index++) {
				if (strcmp(name, class->accessors[index].name) == 0) return true;
			}
		}
		return object_prototype_name(name);
	}
	if (value->kind == adamic_kind_array) {
		const adamic_array *array = (const adamic_array *)value;
		size_t index;
		if (array_key(name, &index)) return index < array->length;
		if (array->properties != NULL && adamic_has_property(&array->properties->heap, name)) return true;
		static const char *const names[] = {"length", "at", "concat", "copyWithin", "fill", "find", "findIndex", "findLast", "findLastIndex", "lastIndexOf", "pop", "push", "reverse", "shift", "unshift", "slice", "sort", "splice", "includes", "indexOf", "join", "keys", "entries", "values", "forEach", "filter", "flat", "flatMap", "map", "every", "some", "reduce", "reduceRight", "toReversed", "toSorted", "toSpliced", "with"};
		return object_prototype_name(name) || named(name, names, sizeof names / sizeof names[0]);
	}
	adamic_panic("in on an unsupported runtime value", sizeof "in on an unsupported runtime value" - 1);
}

static adamic_heap *dynamic_slot(const adamic_object *object, size_t index) {
	const adamic_value slot = object->slots[index];
	if (object->shape->references[index]) return adamic_retain(slot.reference);
	for (const adamic_shape_types *entry = shape_types; entry != NULL; entry = entry->next) {
		if (entry->shape != object->shape) continue;
		// These are ir.Type's scalar representations, written by the emitter.
		switch (entry->types[index]) {
		case 1: return adamic_box_number(slot.number);
		case 2: return slot.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
		case 7: {
			adamic_maybe_number number = adamic_maybe_number_unpack(slot.number);
			return number.present ? adamic_box_number(number.number) : NULL;
		}
		}
	}
	adamic_panic("dynamic read of a host scalar without type metadata", sizeof "dynamic read of a host scalar without type metadata" - 1);
}

adamic_heap *adamic_dynamic_property(adamic_heap *value, const char *name) {
	if (value != NULL && value->kind == adamic_kind_closure && strcmp(name, "length") == 0) {
		return adamic_box_number((double)((adamic_closure *)value)->source_length);
	}
	if (value == NULL || value == &adamic_null) {
		adamic_panic("dynamic property read on null or undefined", sizeof "dynamic property read on null or undefined" - 1);
	}
	if (value->kind == adamic_kind_object) {
		const adamic_object *object = (const adamic_object *)value;
		if (adamic_record_is(object)) {
			adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
			const adamic_value *slot = adamic_record_get(object, &key);
			return slot == NULL ? NULL : adamic_retain(slot->reference);
		}
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(name, object->shape->names[index]) == 0) return dynamic_slot(object, index);
		}
		return NULL;
	}
	if (value->kind == adamic_kind_array) {
		adamic_array *array = (adamic_array *)value;
		if (strcmp(name, "length") == 0) return adamic_box_number((double)array->length);
		if (array->properties != NULL) return adamic_dynamic_property(&array->properties->heap, name);
		return NULL;
	}
	adamic_panic("dynamic property read on an unsupported runtime value", sizeof "dynamic property read on an unsupported runtime value" - 1);
}
