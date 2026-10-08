// closure.c: closures and the cells they capture.

#include "adamic.h"
#include <string.h>

adamic_cell *adamic_cell_new(adamic_value value, bool references) {
	adamic_cell *cell = adamic_allocate(sizeof *cell, adamic_kind_cell);
	cell->references = references;
	cell->value = value;
	return cell;
}

adamic_closure *adamic_closure_new(adamic_code code, size_t count) {
	adamic_closure *closure = adamic_allocate(sizeof *closure + count * sizeof closure->cells[0], adamic_kind_closure);
	closure->code = code;
	closure->receiver = false;
	closure->method = NULL;
	closure->original = NULL;
	closure->unbound = NULL;
	closure->bound = NULL;
	closure->count = count;
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = NULL;
	}
	return closure;
}

// Wrappers borrow arguments and own their saved callable and receiver.
adamic_value adamic_method_value_call(adamic_closure *self, adamic_value *arguments, size_t argument_count) {
	if (self->method != NULL) {
		return self->method(self->bound, arguments, argument_count);
	}
	adamic_closure *original = self->original;
	if (!original->receiver) {
		return original->code(original, arguments, argument_count);
	}
	adamic_value received[argument_count + 1];
	received[0].reference = self->bound;
	for (size_t index = 0; index < argument_count; index++) {
		received[index + 1] = arguments[index];
	}
	return original->code(original, received, argument_count + 1);
}

adamic_closure *adamic_method_bind(adamic_closure *original, adamic_object *receiver) {
	adamic_closure *bound = adamic_closure_new(adamic_method_value_call, 0);
	bound->original = adamic_retain(original->original == NULL ? original : original->original);
	bound->method = original->method;
	bound->bound = adamic_retain(receiver);
	return bound;
}

adamic_closure *adamic_object_method_value(const adamic_object *object, const char *name, adamic_slot_cache *cache, bool optional) {
	if (optional && cache->shape != object->shape) {
		bool found = false;
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(object->shape->names[index], name) == 0) { found = true; break; }
		}
		const adamic_methods *methods = object->shape->methods;
		for (size_t index = 0; !found && methods != NULL && index < methods->count; index++) {
			if (strcmp(methods->names[index], name) == 0) { found = true; }
		}
		if (!found) { return NULL; }
	}
	adamic_method method = NULL;
	adamic_closure *original = adamic_object_callee(object, name, cache, &method);
	if (original == NULL) {
		if (method == NULL) { return NULL; }
		return adamic_retain(object->shape->methods->values[cache->index - object->shape->count]);
	}
	if (!original->receiver) {
		return adamic_retain(original);
	}
	if (original->unbound != NULL) {
		return adamic_retain(original->unbound);
	}
	adamic_closure *unbound = adamic_closure_new(adamic_method_value_call, 0);
	unbound->original = adamic_retain(original);
	original->unbound = unbound;
	return unbound;
}

// Each call owns a fresh rest array; its references survive the caller's temporaries.
adamic_array *adamic_rest_array(adamic_value *arguments, size_t count, size_t start, bool references) {
	adamic_array *rest = adamic_array_new(count > start ? count - start : 0, references);
	for (size_t index = start; index < count; index++) {
		adamic_value value = arguments[index];
		if (references) { adamic_retain(value.reference); }
		adamic_array_push(rest, value);
	}
	return rest;
}
