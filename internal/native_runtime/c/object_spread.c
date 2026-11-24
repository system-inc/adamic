// object_spread.c: { ...source, name: value, ... } for a literal whose result can't take its source's own
// shape, an overwrite of another kind than the slot the source holds, or a name the actual source lacks
// (#5c0z219, the runtime half of #7nhzq7h).
//
// The result is a fresh plain object whose shape is made here, at run time, from the source's own layout:
// every name the source holds keeps its index, an overwrite of one of them takes that slot, and a name the
// source lacks is appended after, in the literal's order. Each slot's kind is that of the value filling it,
// never the slot's in the source. Key order is Node's, through the write order and adamic_object_publish:
// an overwrite of a key the source has keeps its place, and one the source lacks comes last.
//
// Shapes made here are interned by their content and live until the program ends. A loop over one layout
// makes one shape, and since a slot cache keys a shape by its address (adamic_slot_cache), no address it
// has seen is ever freed and reused for another layout.

#include "adamic.h"

#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// spread_shape is one interned shape. Its names, references and kinds follow it in the same allocation,
// then the names' bytes, copied there, so a shape never names storage its source's shape owned: a
// decoded object's shape lives in the object's own allocation (json_decode.c) and goes with it.
typedef struct spread_shape {
	struct spread_shape *next;
	adamic_shape shape;
} spread_shape;

// spread_shapes is the table, newest first. Parallel work can spread, so an entry is pushed once under
// spread_shapes_lock and published with release order, as union.c's shape_types is: a reader loads the
// head with acquire, and an entry never changes or goes away once it's on the list.
static spread_shape *spread_shapes;
static pthread_mutex_t spread_shapes_lock = PTHREAD_MUTEX_INITIALIZER;

// interned_shape finds, among the entries from first on, the shape of exactly these names and kinds, or
// NULL. A slot's reference flag is its kind's (adamic_shape_check), so the kinds say it too, and the
// methods are always NULL, a plain object's.
static const adamic_shape *interned_shape(const spread_shape *first, size_t count, const char *const *names, const adamic_field_kind *kinds) {
	for (const spread_shape *entry = first; entry != NULL; entry = entry->next) {
		if (entry->shape.count != count) continue;
		bool same = true;
		for (size_t index = 0; index < count && same; index++) {
			same = entry->shape.kinds[index] == kinds[index] && strcmp(entry->shape.names[index], names[index]) == 0;
		}
		if (same) return &entry->shape;
	}
	return NULL;
}

// intern_shape is the table's shape of these names and kinds, made the first time it's asked for.
static const adamic_shape *intern_shape(size_t count, const char *const *names, const adamic_field_kind *kinds) {
	const adamic_shape *found = interned_shape(__atomic_load_n(&spread_shapes, __ATOMIC_ACQUIRE), count, names, kinds);
	if (found != NULL) return found;
	pthread_mutex_lock(&spread_shapes_lock);
	// Another thread may have made it since the read above, so the whole list is read again under the lock.
	spread_shape *head = __atomic_load_n(&spread_shapes, __ATOMIC_RELAXED);
	found = interned_shape(head, count, names, kinds);
	if (found == NULL) {
		size_t bytes = sizeof(spread_shape);
		for (size_t index = 0; index < count; index++) {
			size_t length = strlen(names[index]) + 1;
			size_t slot = sizeof(char *) + sizeof(bool) + sizeof(adamic_field_kind);
			if (length > SIZE_MAX - slot || bytes > SIZE_MAX - slot - length) adamic_panic("out of memory", sizeof "out of memory" - 1);
			bytes += slot + length;
		}
		spread_shape *entry = malloc(bytes);
		if (entry == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
		const char **copied = (const char **)(void *)(entry + 1);
		bool *references = (bool *)(void *)(copied + count);
		adamic_field_kind *copied_kinds = (adamic_field_kind *)(void *)(references + count);
		char *text = (char *)(void *)(copied_kinds + count);
		for (size_t index = 0; index < count; index++) {
			size_t length = strlen(names[index]) + 1;
			memcpy(text, names[index], length);
			copied[index] = text;
			text += length;
			references[index] = kinds[index] == adamic_field_reference;
			copied_kinds[index] = kinds[index];
		}
		entry->shape = (adamic_shape){count, copied, references, NULL, copied_kinds};
		entry->next = head;
		__atomic_store_n(&spread_shapes, entry, __ATOMIC_RELEASE);
#ifdef ADAMIC_COUNT
		adamic_counted.spread_shapes++;
#endif
		found = &entry->shape;
	}
	pthread_mutex_unlock(&spread_shapes_lock);
	return found;
}

// refuse stops on a source the first cut doesn't copy, in the NotYet form: each would need its own
// names dropped or its getters run, and Emit's lowering stops first wherever the type admits one, so
// this is only the backstop.
_Noreturn static void refuse(const char *what) {
	char message[160];
	int length = snprintf(message, sizeof message, "NotYet: a spread with overwrites of %s", what);
	adamic_panic(message, (size_t)length);
}

// agrees says whether an overwrite's kind and representation byte describe the same slot, and its value
// is one they describe (Runtime's invariant on #5c0z219): a value and its stated representation can't
// disagree silently past this call, where a reader would meet the disagreement later as a wrong value.
// A reference of a heap representation may be a null pointer, which a cast can leave where the byte
// names a kind (json_objects.c reads those by the byte); every value that is there must be of that kind.
static bool agrees(const adamic_spread_overwrite *overwrite) {
	unsigned char representation = overwrite->representation;
	switch (overwrite->kind) {
	case adamic_field_number:
		if (representation == adamic_ir_MaybeBoolean) return overwrite->value.maybe_boolean <= 2;
		return representation == adamic_ir_Number || representation == adamic_ir_MaybeNumber;
	case adamic_field_boolean:
		// Read as its byte, since a bool of any other bits isn't a value to load.
		return representation == adamic_ir_Boolean && overwrite->value.maybe_boolean <= 1;
	case adamic_field_reference: {
		const adamic_heap *value = overwrite->value.reference;
		if (representation == adamic_ir_UndefinedRepresentation) return value == NULL;
		if (representation == adamic_ir_NullRepresentation) return value == NULL || value == &adamic_null;
		if (!adamic_ir_type_reference((enum adamic_ir_type)representation)) return false;
		if (value == NULL) return true;
		switch (representation) {
		case adamic_ir_String: return value->kind == adamic_kind_string;
		case adamic_ir_Object: return value->kind == adamic_kind_object;
		case adamic_ir_Array: return value->kind == adamic_kind_array || (value->kind == adamic_kind_object && ((const adamic_object *)value)->tuple);
		case adamic_ir_Map: return value->kind == adamic_kind_map;
		case adamic_ir_Closure: return value->kind == adamic_kind_closure;
		default: return true;
		}
	}
	}
	return false;
}

adamic_object *adamic_object_spread_with(const adamic_object *source, const adamic_spread_overwrite *overwrites, size_t count) {
	for (size_t at = 0; at < count; at++) {
		if (agrees(&overwrites[at])) continue;
		const char *name = overwrites[at].name;
		size_t capacity = strlen(name) + 160;
		char *message = malloc(capacity);
		if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
		int length = snprintf(message, capacity, "compiler bug: a spread's overwrite of %s states kind %u and representation %u, which don't describe its value", name, (unsigned)overwrites[at].kind, (unsigned)overwrites[at].representation);
		adamic_panic(message, (size_t)length);
	}
	// A spread of undefined or of null copies nothing.
	if ((const adamic_heap *)source == &adamic_null) source = NULL;
	if (source != NULL) {
		if (source->heap.kind != adamic_kind_object) refuse("a value that isn't an object");
		if (adamic_record_is(source)) adamic_panic("NotYet: spread of record storage", sizeof "NotYet: spread of record storage" - 1);
		if (source->tuple) refuse("a tuple");
		if (adamic_error_is(source)) refuse("an Error");
		// A host's and a match's groups object alike: their names aren't all Node's own enumerable keys.
		if (adamic_host_shape(source->shape)) refuse("a runtime library value (a host or groups object)");
		// A getter would run at the copy, and its key's order is the public shape's.
		for (const adamic_class *class = source->class; class != NULL; class = class->base) {
			if (class->accessor_count != 0) refuse("a class instance with accessors");
		}
	}
	// The source's own layout: a class instance's public shape, as adamic_object_copy_checked copies it.
	const adamic_shape *base = source == NULL ? NULL : source->class == NULL ? source->shape : source->class->public_shape;
	size_t held = base == NULL ? 0 : base->count;
	// One allocation holds the layout as it's worked out: the result's names and kinds, each overwrite's
	// slot, and which of the source's slots an overwrite takes.
	size_t per_slot = sizeof(const char *) + sizeof(size_t) + sizeof(adamic_field_kind) + sizeof(bool);
	if (count > SIZE_MAX - held || held + count > SIZE_MAX / per_slot) adamic_panic("out of memory", sizeof "out of memory" - 1);
	size_t capacity = held + count;
	unsigned char *scratch = malloc(capacity == 0 ? 1 : capacity * per_slot);
	if (scratch == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
	const char **names = (const char **)(void *)scratch;
	size_t *slots = (size_t *)(void *)(names + capacity);
	adamic_field_kind *kinds = (adamic_field_kind *)(void *)(slots + capacity);
	bool *taken = (bool *)(void *)(kinds + capacity);
	for (size_t index = 0; index < held; index++) {
		names[index] = base->names[index];
		kinds[index] = base->kinds[index];
		taken[index] = false;
	}
	size_t length = held;
	for (size_t at = 0; at < count; at++) {
		size_t index = 0;
		while (index < length && strcmp(names[index], overwrites[at].name) != 0) index++;
		if (index == length) {
			names[length] = overwrites[at].name;
			length++;
		} else if (index >= held || taken[index]) {
			// TypeScript refuses a literal that names a property twice, so the emitter never writes one.
			size_t size = strlen(overwrites[at].name) + 64;
			char *message = malloc(size);
			if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
			int written = snprintf(message, size, "compiler bug: a spread overwrites %s twice", overwrites[at].name);
			adamic_panic(message, (size_t)written);
		} else {
			taken[index] = true;
		}
		// The slot's kind follows the value that fills it, never the source's slot.
		kinds[index] = overwrites[at].kind;
		slots[at] = index;
	}
	const adamic_shape *shape = intern_shape(length, names, kinds);
	// A source with a write order has slots it reserved and never wrote, which stay absent; the result keeps
	// its order, and each overwrite of an absent slot or a new name is published after it.
	bool ordered = source != NULL && source->write_order != NULL;
	adamic_object *result = ordered ? adamic_object_construct(shape) : adamic_object_new(shape);
	if (ordered) memcpy(result->write_order, source->write_order, held * sizeof(size_t));
	for (size_t index = 0; index < held; index++) {
		if (taken[index] || !adamic_object_present(source, index)) continue;
		// A constructor's static field can live in a parent constructor until an own write shadows it, so a
		// class's slot is found by name, with its bytes read from the object that holds it (class_static.c).
		const adamic_object *owner = source;
		const adamic_value *found = &source->slots[index];
		if (source->class != NULL) {
			adamic_slot_cache cache = {0};
			found = adamic_object_field_owned(source, base->names[index], &cache, &owner);
		}
		size_t at = adamic_slot_index(owner, found);
		result->slots[index] = *found;
		adamic_object_initialized(result)[index] = adamic_object_initialized(owner)[at];
		adamic_object_field_types(result)[index] = adamic_object_field_type(owner, at);
		// Retained by the result's bit, which is the source's here: a slot an overwrite takes is skipped above,
		// so a reference it replaces is never retained, and a number it replaces never read as one.
		if (shape->references[index]) adamic_retain(found->reference);
	}
	for (size_t at = 0; at < count; at++) {
		size_t index = slots[at];
		result->slots[index] = overwrites[at].value;
		adamic_object_field_types(result)[index] = overwrites[at].representation;
		adamic_object_initialized(result)[index] = 1;
		adamic_object_publish(result, index);
	}
	free(scratch);
	return result;
}
