// foreign.c: values another runtime made, held by Adamic (adamic.h, docs/apple.md).

#include "adamic.h"

// Every foreign value has this shape: no fields, no methods. A property read that reaches one by name
// panics as a compiler bug, out loud, rather than reading what isn't there.
static const adamic_shape foreign_shape = {0, NULL, NULL, NULL};

// A foreign value's two slots, past the shape's count where nothing looks: the pointer, and its kind.
adamic_object *adamic_foreign_new(void *pointer, const adamic_foreign_kind *kind) {
	adamic_object *object = adamic_allocate(sizeof *object + 2 * sizeof object->slots[0], adamic_kind_foreign);
	object->shape = &foreign_shape;
	object->class = NULL;
	object->frozen = true;
	object->slots[0].reference = pointer;
	object->slots[1].reference = (void *)kind;
	return object;
}

void *adamic_foreign_pointer(const adamic_object *object) {
	if (object == NULL) {
		return NULL;
	}
	if (object->heap.kind != adamic_kind_foreign) {
		static const char message[] = "compiler bug: a value handed to native code as foreign isn't one";
		adamic_panic(message, sizeof message - 1);
	}
	return object->slots[0].reference;
}

void adamic_foreign_dropped(adamic_object *object) {
	const adamic_foreign_kind *kind = object->slots[1].reference;
	kind->release(object->slots[0].reference);
}
