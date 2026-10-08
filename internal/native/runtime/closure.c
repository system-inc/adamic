// closure.c: closures and the cells they capture.

#include "adamic.h"

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

adamic_closure *adamic_object_method_value(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_method method = NULL;
	adamic_closure *original = adamic_object_callee(object, name, cache, &method);
	if (original == NULL) {
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
