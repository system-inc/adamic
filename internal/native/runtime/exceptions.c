// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_object *adamic_thrown;

static const char *const error_names[] = {"name", "message", "code"};
static const bool error_references[] = {true, true, true};
// Distinct immutable shapes carry constructor identity; name remains writable.
// The last two are host SystemError and DOMException, both inheriting Error.
static const adamic_shape error_shapes[] = {
	{3, error_names, error_references, NULL}, {3, error_names, error_references, NULL},
	{3, error_names, error_references, NULL}, {3, error_names, error_references, NULL},
	{3, error_names, error_references, NULL}, {3, error_names, error_references, NULL},
	{3, error_names, error_references, NULL}, {3, error_names, error_references, NULL}, {3, error_names, error_references, NULL}
};
static adamic_string error_labels[] = {
	ADAMIC_STRING("Error"), ADAMIC_STRING("TypeError"), ADAMIC_STRING("SyntaxError"),
	ADAMIC_STRING("RangeError"), ADAMIC_STRING("ReferenceError"), ADAMIC_STRING("EvalError"), ADAMIC_STRING("URIError"), ADAMIC_STRING("SystemError"), ADAMIC_STRING("SyntaxError")
};

bool adamic_error_is(const adamic_object *error, int kind, bool exact) {
	if (error == NULL) return false;
	for (int actual = 0; actual < 9; actual++) {
		if (error->shape == &error_shapes[actual]) return actual == kind || (!exact && kind == 0);
	}
	return false;
}

adamic_object *adamic_builtin_error_new(int kind, adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shapes[kind]);
	error->slots[0].reference = adamic_retain(&error_labels[kind]);
	error->slots[1].reference = adamic_retain(message);
	return error;
}

adamic_object *adamic_error_new(adamic_string *message) {
	return adamic_builtin_error_new(0, message);
}

_Noreturn void adamic_uncaught(void) {
	// String(error), as Node's runner reports an error nothing caught: name, and ": " and the message
	// when there is one.
	static adamic_slot_cache name_cache, message_cache;
	const adamic_string *name = adamic_object_field(adamic_thrown, "name", &name_cache)->reference;
	const adamic_string *message = adamic_object_field(adamic_thrown, "message", &message_cache)->reference;
	size_t length = name->length + (message->length > 0 ? 2 + message->length : 0);
	char *text = malloc(length + 1);
	if (text == NULL) {
		static const char out_of_memory[] = "out of memory";
		adamic_panic(out_of_memory, sizeof out_of_memory - 1);
	}
	memcpy(text, name->bytes, name->length);
	if (message->length > 0) {
		memcpy(text + name->length, ": ", 2);
		memcpy(text + name->length + 2, message->bytes, message->length);
	}
	adamic_panic(text, length);
}
