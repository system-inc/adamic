// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_heap *adamic_thrown;
bool adamic_exception_pending;

static const char *const error_names[] = {"name", "message"};
static const bool error_references[] = {true, true};
static const adamic_shape error_shape = {2, error_names, error_references, NULL};
static adamic_string error_name = ADAMIC_STRING("Error");

adamic_object *adamic_error_new(adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message);
	return error;
}

// The filesystem producer owns its distinct Error shape with a code field.
extern bool adamic_fs_file_is_error(const adamic_object *value);

bool adamic_is_error(const adamic_heap *value) {
	return value != NULL && value->kind == adamic_kind_object && (((const adamic_object *)value)->shape == &error_shape || adamic_fs_file_is_error((const adamic_object *)value));
}

_Noreturn void adamic_uncaught(void) {
	if (!adamic_is_error(adamic_thrown)) {
		if (adamic_thrown != NULL && adamic_thrown->kind == adamic_kind_object) {
			static const char message[] = "[object Object]";
			adamic_panic(message, sizeof message - 1);
		}
		adamic_string *text = adamic_union_to_string(adamic_thrown);
		adamic_panic(text->bytes, text->length);
	}
	// String(error), as Node's runner reports an error nothing caught: name, and ": " and the message
	// when there is one.
	static adamic_slot_cache name_cache, message_cache;
	const adamic_string *name = adamic_object_field((adamic_object *)adamic_thrown, "name", &name_cache)->reference;
	const adamic_string *message = adamic_object_field((adamic_object *)adamic_thrown, "message", &message_cache)->reference;
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
