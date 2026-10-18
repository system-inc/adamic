// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

_Thread_local adamic_object *adamic_thrown;

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

// Get must stay inside the loop: the first method or getter can change the second.
adamic_primitive adamic_ordinary_to_primitive(adamic_heap *receiver, enum adamic_primitive_hint hint, adamic_primitive_get get) {
	adamic_primitive empty = {adamic_primitive_undefined, {.reference = NULL}};
	if (adamic_thrown != NULL) { return empty; }
	if (hint != adamic_hint_default && hint != adamic_hint_number && hint != adamic_hint_string) {
		static const char message[] = "compiler bug: invalid primitive hint";
		adamic_panic(message, sizeof message - 1);
	}
	const char *names[2] = {"valueOf", "toString"};
	if (hint == adamic_hint_string) { names[0] = "toString"; names[1] = "valueOf"; }
	for (size_t i = 0; i < 2; i++) {
		adamic_primitive_method method = get(receiver, names[i]);
		if (adamic_thrown != NULL) { adamic_release(method.owner); return empty; }
		if (method.call == NULL) { adamic_release(method.owner); continue; }
		adamic_primitive result = method.call(receiver, method.owner);
		adamic_release(method.owner);
		if (adamic_thrown != NULL) { return empty; }
		if (result.kind != adamic_primitive_object) { return result; }
		adamic_release(result.value.reference);
	}
	// A JavaScript failure owns an Error on the pending word, never a terminal panic.
	// Build these rare strings rather than adding mutable global literal caches.
	static const char name[] = "TypeError";
	static const char message[] = "Cannot convert object to primitive value";
	adamic_string *text = adamic_string_allocate(sizeof message - 1);
	memcpy((char *)text->bytes, message, sizeof message - 1);
	adamic_object *error = adamic_error_new(text);
	adamic_release(text);
	adamic_string *type_name = adamic_string_allocate(sizeof name - 1);
	memcpy((char *)type_name->bytes, name, sizeof name - 1);
	adamic_release(error->slots[0].reference);
	error->slots[0].reference = type_name;
	adamic_thrown = error;
	return empty;
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
