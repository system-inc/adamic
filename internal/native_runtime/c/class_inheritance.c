// class_inheritance.c: nominal identity, virtual dispatch and one derived-to-base release chain.
#include "adamic.h"

bool adamic_instanceof(const void *value, const adamic_class *wanted) {
	const adamic_heap *heap = value;
	if (heap == NULL || heap->kind != adamic_kind_object) { return false; }
	const adamic_object *object = value;
	for (const adamic_class *class = object->class; class != NULL; class = class->base) {
		if (class == wanted || (wanted->definition != 0 && class->definition == wanted->definition)) { return true; }
	}
	return false;
}

// A direct call names its one target by the receiver's static class, so nothing at the call reads the
// receiver's identity: a plain object that reaches a class's method, as a generic spread's copy of an
// instance did (#y4gjeb8), would run the method's code on a layout it doesn't have, reading inside the
// copy's allocation where ASan sees nothing wrong. Every such object is the compiler's bug, so it stops here
// rather than being called.
adamic_object *adamic_method_this(adamic_object *object, const adamic_class *owner) {
	if (!adamic_instanceof(object, owner)) { adamic_invariant("a class method entered with a this that isn't an instance of its class"); }
	return object;
}

size_t adamic_virtual(const adamic_object *object, size_t slot) {
	if (object == NULL || object->class == NULL) {
		static const char message[] = "compiler bug: virtual call on an object without class identity";
		adamic_panic(message, sizeof message - 1);
	}
	return object->class->methods[slot];
}

// Arena children and immortal values have no count to drop. Keeping them out of the callback
// preserves the region teardown's counting and avoids treating arena storage as heap-owned.
static void release_field(adamic_object *object, size_t index, void (*release)(void *)) {
	if (!object->shape->references[index]) { return; }
	adamic_heap *child = object->slots[index].reference;
	if (child != NULL && child->references != 0) { release(child); }
}

void adamic_object_free_children(adamic_object *object, void (*release)(void *)) {
	if (object->class == NULL) {
		for (size_t index = 0; index < object->shape->count; index++) {
			release_field(object, index, release);
		}
		return;
	}
	// Listing children through the callback keeps heap destruction iterative, even for deep trees.
	for (const adamic_class *class = object->class; class != NULL; class = class->base) {
		for (size_t index = class->count; index > class->own_start; ) {
			index--;
			release_field(object, index, release);
		}
	}
}
