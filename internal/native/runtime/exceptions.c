// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_object *adamic_thrown;

static const char *const error_names[] = {"name", "message"};
static const bool error_references[] = {true, true};
static const adamic_shape error_shape = {2, error_names, error_references};
static adamic_string error_name = ADAMIC_STRING("Error");

adamic_object *adamic_error_new(adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message);
	return error;
}

static adamic_string type_error_name = ADAMIC_STRING("TypeError");

adamic_object *adamic_type_error_new(const char *message, size_t length) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->slots[0].reference = adamic_retain(&type_error_name);
	adamic_string *text = adamic_string_allocate(length);
	memcpy((char *)text->bytes, message, length);
	error->slots[1].reference = text;
	return error;
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
