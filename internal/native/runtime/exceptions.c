// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"
#include "view_representations.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_object *adamic_thrown;

static const char *const error_names[] = {"name", "message"};
static const bool error_references[] = {true, true};
static const adamic_shape error_shape = {2, error_names, error_references, NULL};
static const char *const absent_message_names[] = {"name", "#message"};
static const adamic_shape absent_message_shape = {2, absent_message_names, error_references, NULL};
static adamic_string empty_message = ADAMIC_STRING("");
static adamic_string error_name = ADAMIC_STRING("Error");

adamic_object *adamic_error_new(adamic_string *message) {
	// Undefined creates no own message property. Keep the inherited empty fallback
	// in private storage so ordinary and checked field reads share the same slot.
	adamic_object *error = adamic_object_new(message == NULL ? &absent_message_shape : &error_shape);
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message == NULL ? &empty_message : message);
	adamic_object_field_types(error)[0] = adamic_rep_string;
	adamic_object_field_types(error)[1] = adamic_rep_string;
	return error;
}

// Public lookup sees Error.prototype.message without creating an own property.
const char *adamic_error_field_name(const adamic_object *object, const char *name) {
	return object->shape == &absent_message_shape && strcmp(name, "message") == 0 ? "#message" : name;
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
