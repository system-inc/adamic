// union.c: a union whose members are held differently, as one counted reference (adamic.h).

#include "adamic.h"

#include <math.h>

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
	if (adamic_reference_is_sentinel(value)) {
		return adamic_reference_null_text();
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
	if (adamic_reference_is_sentinel(value)) {
		return &adamic_typeof_object;
	}
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
