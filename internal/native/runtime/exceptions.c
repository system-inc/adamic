// exceptions.c: throw, try, catch and finally, by cleanup paths (docs/memory.md, "Exceptions,
// designed into counting"). The emitter does the unwinding; the runtime holds what's thrown.

#include "adamic.h"
#include "library_errors.h"
#include "count.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

_Thread_local adamic_object *adamic_thrown;

static const char *const error_names[] = {"name", "message", "code"};
static const bool error_references[] = {true, true, true};
static const adamic_shape error_shape = {3, error_names, error_references, NULL};
static adamic_string error_name = ADAMIC_STRING("Error");

adamic_object *adamic_error_new(adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message);
	return error;
}

_Noreturn void adamic_uncaught(void) {
	// Error.prototype.toString, omitting source locations and frames because stack is refused.
	static adamic_slot_cache name_cache, message_cache;
	const adamic_string *name = adamic_object_field(adamic_thrown, "name", &name_cache)->reference;
	const adamic_string *message = adamic_object_field(adamic_thrown, "message", &message_cache)->reference;
	adamic_string *text;
	if (name->length == 0) {
		text = adamic_retain((void *)message);
	} else if (message->length == 0) {
		text = adamic_retain((void *)name);
	} else {
		static adamic_string separator = ADAMIC_STRING(": ");
		text = adamic_string_concat(3, (adamic_string *const[]){(adamic_string *)name, &separator, (adamic_string *)message});
	}
	// write_line applies Node's UTF-8 replacement for lone surrogates.
	static adamic_string prefix = ADAMIC_STRING("adamic: panic: ");
	adamic_string *line = adamic_string_concat(2, (adamic_string *const[]){&prefix, text});
	adamic_write_line(adamic_stderr, line);
	adamic_release(line);
	adamic_release(text);
	adamic_release(adamic_thrown);
	adamic_thrown = NULL;
	adamic_output_flush();
	fflush(NULL);
	ADAMIC_COUNT_REPORT();
	_Exit(70);
}

_Noreturn void adamic_uncaught_library_error(const char *message, size_t length) {
	// Runtime library messages contain only ASCII, including formatted numbers.
	adamic_output_flush();
	fputs("adamic: panic: ", stderr);
	fwrite(message, 1, length, stderr);
	fputc('\n', stderr);
	adamic_output_flush();
	fflush(NULL);
	ADAMIC_COUNT_REPORT();
	_Exit(70);
}
