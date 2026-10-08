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
const adamic_class adamic_error_class = {NULL, 0, 2, NULL, 0, &error_shape, NULL, 0, false, 0, NULL};
const adamic_class adamic_host_error_class = {&adamic_error_class, 2, 3, NULL, 0, NULL, NULL, 0, false, 0, NULL};
static adamic_string error_name = ADAMIC_STRING("Error");

adamic_object *adamic_error_new(adamic_string *message) {
	adamic_object *error = adamic_object_new(&error_shape);
	error->class = &adamic_error_class;
	error->slots[0].reference = adamic_retain(&error_name);
	error->slots[1].reference = adamic_retain(message);
	return error;
}

_Noreturn void adamic_uncaught(void) {
	// No implicit conversion or stack rendering: the external contract is stdout and exit 1.
	adamic_output_flush();
	adamic_release(adamic_thrown);
	adamic_thrown = NULL;
	adamic_exception_pending = false;
	exit(1);
}
