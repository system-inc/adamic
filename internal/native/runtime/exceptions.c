// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

_Thread_local adamic_object *adamic_thrown;

static const char *const error_names[] = {"name", "message"};
static const bool error_references[] = {true, true};
static const adamic_field_kind error_kinds[] = {adamic_field_reference, adamic_field_reference};
static const adamic_shape error_shape = {2, error_names, error_references, NULL, error_kinds, NULL, NULL};
static adamic_string error_name = ADAMIC_STRING("Error");

adamic_object *adamic_error_new(adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message);
	return error;
}

_Noreturn void adamic_uncaught(void) {
	// String(error), as Node's runner reports an error nothing caught (Error.prototype.toString): the
	// name and the message joined by ": " when both are there, and whichever one is when the other is
	// empty.
	static adamic_slot_cache name_cache, message_cache;
	const adamic_string *name = adamic_object_field(adamic_thrown, "name", &name_cache)->reference;
	const adamic_string *message = adamic_object_field(adamic_thrown, "message", &message_cache)->reference;
	size_t separator = name->length > 0 && message->length > 0 ? 2 : 0;
	size_t length = name->length + separator + message->length;
	char *text = malloc(length + 1);
	if (text == NULL) {
		static const char out_of_memory[] = "out of memory";
		adamic_panic(out_of_memory, sizeof out_of_memory - 1);
	}
	memcpy(text, name->bytes, name->length);
	memcpy(text + name->length, ": ", separator);
	memcpy(text + name->length + separator, message->bytes, message->length);
	adamic_panic(text, length);
}
