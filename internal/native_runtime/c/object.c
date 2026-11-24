// object.c: plain objects, each carrying its shape.

#include "adamic.h"
#include "graph_regions.h"
#include "view_unions_mixed.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

// empty makes new storage an object of shape with every field zeroed, wherever the storage came from.
static adamic_object *empty(adamic_object *object, const adamic_shape *shape) {
	object->shape = shape;
	object->class = NULL;
	object->frozen = false;
	object->tuple = false;
	object->write_order = NULL;
	memset(object->slots, 0, shape->count * sizeof object->slots[0]);
	memset(adamic_object_initialized(object), 1, shape->count);
	memset(adamic_object_field_types(object), 0, shape->count);
	adamic_error_initialize(object);
	return object;
}

adamic_object *adamic_object_new(const adamic_shape *shape) {
	adamic_shape_check(shape);
	return empty(adamic_allocate(adamic_object_size(shape->count), adamic_kind_object), shape);
}

// An object born in graph storage, as a graph record's wrapper is (record.c): no region's until
// adamic_graph_born, and never moved.
adamic_object *adamic_object_new_graph(const adamic_shape *shape) {
	adamic_shape_check(shape);
	return empty(adamic_heap_graph_allocate(adamic_object_size(shape->count), adamic_kind_object), shape);
}

// field_owned is adamic_object_field that also says which object holds the slot: a constructor's
// static field can live in a parent constructor until an own write shadows it (class_static.c).
// Per-slot metadata is indexed from that object's slots, never from the shared packed cache.
adamic_value *adamic_object_field_owned(const adamic_object *object, const char *name, adamic_slot_cache *cache, const adamic_object **owner) {
	while (object->class != NULL && object->class->is_static) {
		adamic_value *slot = adamic_object_find(object, name, cache);
		size_t flag = object->class->static_flags[adamic_slot_index(object, slot)];
		if (flag == 0 || object->slots[flag - 1].number != 0 || object->class->static_parent == 0) {
			*owner = object;
			return slot;
		}
		object = object->slots[object->class->static_parent - 1].reference;
	}
	*owner = object;
	return adamic_object_data_field(object, name, cache);
}

static adamic_value *adamic_object_read_mode(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner);

adamic_object *adamic_object_copy_checked(const adamic_object *source, const char *expression) {
	if (adamic_record_is(source)) adamic_panic("NotYet: spread of record storage", sizeof "NotYet: spread of record storage" - 1);
	const adamic_shape *shape = source->class == NULL ? source->shape : source->class->public_shape;
	adamic_object *object = source->write_order == NULL ? adamic_object_new(shape) : adamic_object_construct(shape);
	if (source->write_order != NULL) memcpy(object->write_order, source->write_order, shape->count * sizeof(size_t));
	for (size_t position = 0; position < shape->count; position++) {
		size_t index = adamic_public_index(shape, position);
		if (!adamic_object_present(source, index)) continue;
		adamic_slot_cache cache = {0};
		const adamic_accessor *accessor = adamic_accessor_find(source, shape->names[index]);

		if (accessor == NULL) {
			// The slot found says where the field is, not the cache: another thread may store a
			// different shape's slot in a shared cache (the packed cache, adamic.h).
			const adamic_object *owner = source;
			adamic_value *found = expression == NULL ? adamic_object_field_owned(source, shape->names[index], &cache, &owner) : adamic_object_read_mode(source, shape->names[index], &cache, expression, NULL, &owner);
			object->slots[index] = *found;
			adamic_object_initialized(object)[index] = adamic_object_initialized(owner)[adamic_slot_index(owner, found)];
			adamic_object_field_types(object)[index] = adamic_object_field_type(owner, adamic_slot_index(owner, found));
		} else {
			object->slots[index] = adamic_accessor_get((adamic_object *)source, shape->names[index]);
			if (adamic_thrown != NULL) {
				adamic_release(object);
				return NULL;
			}
		}
		if (shape->references[index] && accessor == NULL) {
			adamic_retain(object->slots[index].reference);
		}
	}
	return object;
}

adamic_object *adamic_object_copy(const adamic_object *source) { return adamic_object_copy_checked(source, NULL); }

// adamic_object_has is object.hasOwnProperty(name): one of the shape's own names, not a method on a
// prototype. A shape's names are C strings, so the lengths have to agree before the bytes do.
bool adamic_object_has(const adamic_object *object, const adamic_string *name) {
	// hasOwnProperty and ToPrimitive's own toString test: a host object's names aren't Node's own.
	adamic_host_stop(&object->heap, name);
	if (adamic_record_is(object)) return adamic_record_has_own(object, name);
	for (size_t index = 0; index < object->shape->count; index++) {
		const char *field = object->shape->names[index];
		size_t length = strlen(field);
		if (length == name->length && memcmp(field, name->bytes, length) == 0) {
			return adamic_object_present(object, index);
		}
	}
	return false;
}

void adamic_slot_cache_store(adamic_slot_cache *cache, const adamic_shape *shape, size_t index) {
	uintptr_t pointer = (uintptr_t)shape;
	if ((pointer & ~ADAMIC_SLOT_SHAPE_MASK) != 0) {
		adamic_panic("shape address exceeds 48 bits", sizeof "shape address exceeds 48 bits" - 1);
	}
	uint64_t packed = index <= UINT16_MAX ? pointer | ((uint64_t)index << 48) : 0;
	__atomic_store_n(&cache->packed, packed, __ATOMIC_RELAXED);
	ADAMIC_TSAN_PAUSE(adamic_tsan_cache_publication);
}

// A record keeps its keys in a table behind one slot, so a fixed-object read of it would find no
// field and say undefined, or read the table as a value. Lowering admits a plain read only where no
// record can reach it; this makes a path it missed stop loudly. Both lookups below are the miss
// paths of their inline caches, and a record's shape is never cached, so hits pay nothing.
void adamic_refuse_record_fixed_read(const adamic_object *object, const char *name) {
	if (!adamic_record_is(object)) { return; }
	size_t capacity = strlen(name) + 64;
	char *message = malloc(capacity);
	if (message == NULL) {
		static const char oom[] = "out of memory";
		adamic_panic(oom, sizeof oom - 1);
	}
	int length = snprintf(message, capacity, "compiler bug: a record reached a fixed-object read of %s", name);
	adamic_panic(message, (size_t)length);
}

adamic_value *adamic_object_find(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_refuse_record_fixed_read(object, name);
	name = adamic_error_field_name(object, name);
	adamic_object *mutable = (adamic_object *)object;
	for (size_t index = 0; index < object->shape->count; index++) {
		if (strcmp(object->shape->names[index], name) == 0) {
			adamic_slot_cache_store(cache, object->shape, index);
			return &mutable->slots[index];
		}
	}
	static const char message[] = "compiler bug: a field the checker proved is there is missing";
	adamic_panic(message, sizeof message - 1);
}

// intl_collator_shape is the Intl.Collator shape library_intl.c published (adamic_intl_collator_shape_publish),
// or null before any collator is made, held as union.c holds the Stats shape: a release store, an
// acquire load, and only an address to compare with.
static const void *_Atomic intl_collator_shape;

void adamic_intl_collator_shape_publish(const adamic_shape *shape) {
	atomic_store_explicit(&intl_collator_shape, shape, memory_order_release);
}

// Fixed methods and their truthiness use the same lookup, including host registrations.
static size_t object_callee_slot(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	const adamic_shape *shape = object->shape;
	// The cache's index counts the fields, then the methods after them.
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	size_t slot = packed >> 48;
	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)shape) {
		bool found = false;
		for (size_t index = 0; index < shape->count && !found; index++) {
			if (strcmp(shape->names[index], name) == 0) {
				slot = index;
				found = true;
			}
		}
		for (size_t index = 0; shape->methods != NULL && index < shape->methods->count && !found; index++) {
			if (strcmp(shape->methods->names[index], name) == 0) {
				slot = shape->count + index;
				found = true;
			}
		}
		// Optional binding observes absence without caching a made-up slot.
		if (!found) { return SIZE_MAX; }
		adamic_slot_cache_store(cache, shape, slot);
	}
	return slot;
}
adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method_entry *method) {
	if (adamic_record_is(object)) {
		adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
		const adamic_value *slot = adamic_record_get_own(object, &key);
		if (slot == NULL) { return NULL; }
		// A record of numbers or booleans keeps them unboxed in the slot, which is never a function, so its
		// key answers not callable before anything reads the slot as a pointer (#3tjjcjs).
		bool references = adamic_record_values(object) == adamic_field_reference;
		if (references && slot->reference == NULL) { return NULL; }
		const adamic_heap *reference = references ? slot->reference : NULL;
		if (reference != NULL && reference->kind == adamic_kind_closure) { return slot->reference; }
		size_t capacity = strlen(name) + 64;
		char *message = malloc(capacity);
		if (message == NULL) {
			static const char oom[] = "out of memory";
			adamic_panic(oom, sizeof oom - 1);
		}
		int length = snprintf(message, capacity, "TypeError: record key '%s' is not a function", name);
		adamic_panic(message, (size_t)length);
	}
	size_t slot = object_callee_slot(object, name, cache);
	const adamic_shape *shape = object->shape;
	if (slot == SIZE_MAX) {
		// A direct Collator call uses its private placeholder. The host view walk
		// forbids this member through a structural view.
		if (shape == atomic_load_explicit(&intl_collator_shape, memory_order_acquire) && strcmp(name, "compare") == 0) {
			static adamic_slot_cache compare_cache;
			return adamic_object_field(object, "#compare", &compare_cache)->reference;
		}
		adamic_regex_method(object, name, method);
		return NULL;
	}
	if (slot < shape->count) {
		if (!shape->references[slot] || !adamic_object_present(object, slot) ||
		    !adamic_object_initialized(object)[slot]) return NULL;
		adamic_heap *value = object->slots[slot].reference;
		return value != NULL && value->kind == adamic_kind_closure ? (adamic_closure *)value : NULL;
	}
	*method = shape->methods->code[slot - shape->count];
	return NULL;
}

bool adamic_object_method_truthy(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	if (object == NULL || (const adamic_heap *)object == &adamic_null) {
		static const char message[] = "method truthiness read on null or undefined";
		adamic_panic(message, sizeof message - 1);
	}
	if (adamic_record_is(object)) {
		adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
		const adamic_value *slot = adamic_record_get_own(object, &key);
		if (slot == NULL) return false;
		switch (adamic_record_values(object)) {
		case adamic_field_reference: return adamic_census_to_boolean(slot->reference);
		case adamic_field_boolean: return slot->boolean;
		default: return slot->number != 0 && !isnan(slot->number);
		}
	}
	adamic_method_entry method = {0};
	size_t slot = object_callee_slot(object, name, cache);
	// The host view walk holds Collator.compare out of structural truthiness reads.
	// Admitting that view would require handling its prototype getter here.
	if (slot == SIZE_MAX) return adamic_regex_method(object, name, &method);
	if (slot >= object->shape->count) return true;
	return adamic_object_slot_truthy(object, slot);
}

// adamic_object_slot_defined is adamic_object_method_truthy's lookup asked whether the slot holds anything but
// undefined, absence answering false: a presence test's (value as T).f !== undefined (lower's
// viewFieldPresence). A record's own key is read as dynamic_property reads it, as a union.
bool adamic_object_slot_defined(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	if (object == NULL || (const adamic_heap *)object == &adamic_null) {
		static const char message[] = "slot presence read on null or undefined";
		adamic_panic(message, sizeof message - 1);
	}
	if (adamic_record_is(object)) {
		adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
		const adamic_value *slot = adamic_record_get_own(object, &key);
		if (slot == NULL) return false;
		adamic_heap *value = adamic_record_value_union(object, slot);
		bool defined = value != NULL;
		adamic_release(value);
		return defined;
	}
	adamic_method_entry method = {0};
	size_t slot = object_callee_slot(object, name, cache);
	if (slot == SIZE_MAX) return adamic_regex_method(object, name, &method);
	if (slot >= object->shape->count) return true;
	return adamic_object_slot_defined_at(object, slot);
}

// A field made as undefined alone holds NULL, whereas number | undefined holds a packed number.
// The shape decides which union member is live; reading NULL's bits as a double would produce 0.
adamic_maybe_number adamic_object_maybe_number(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	adamic_value *slot = adamic_object_field(object, name, cache);
	adamic_value *own = object->class != NULL && object->class->is_static ? adamic_object_find(object, name, cache) : slot;
	if (object->shape->references[(size_t)(own - object->slots)]) {
		if (slot->reference != NULL) {
			static const char message[] = "compiler bug: a numeric field holds a reference";
			adamic_panic(message, sizeof message - 1);
		}
		return (adamic_maybe_number){false, 0.0};
	}
	return adamic_maybe_number_unpack(slot->number);
}

// Cache absence too, with count as the index, without adding a field to the object's shape.
adamic_value *adamic_object_optional_find(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	name = adamic_error_field_name(object, name);
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	size_t slot = packed >> 48;
	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)object->shape) {
		adamic_refuse_record_fixed_read(object, name);
		slot = object->shape->count;
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(object->shape->names[index], name) == 0) {
				slot = index;
				break;
			}
		}
		adamic_slot_cache_store(cache, object->shape, slot);
	}
	if (slot == object->shape->count || !adamic_object_present(object, slot)) { return NULL; }
	return &((adamic_object *)object)->slots[slot];
}

// Presence can observe a served prototype method without handing out its function value.
bool adamic_object_optional_function_present(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression) {
	if (object == NULL) return false;
	adamic_method_entry method = {0};
	if (adamic_regex_method(object, name, &method)) return true;
	adamic_value *slot = adamic_object_optional_field(object, name, cache);
	if (slot == NULL) return false;
	if (expression != NULL) slot = adamic_object_read(object, name, cache, expression);
	return slot->reference != NULL;
}

// Both public reads use the worker's readiness check; only their diagnostics differ.
static adamic_value *adamic_object_read_mode(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner) {
	adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
	if (slot != NULL && object->class != NULL && object->class->is_static) {
		size_t flag = object->class->static_flags[adamic_slot_index(object, slot)];
		if (flag != 0 && object->slots[flag - 1].number == 0 && object->class->static_parent != 0) {
			return adamic_object_read_mode(object->slots[object->class->static_parent - 1].reference, name, cache, expression, expected, owner);
		}
	}
	if (slot == NULL || !adamic_object_initialized(object)[adamic_slot_index(object, slot)]) {
		size_t capacity = strlen(name) + strlen(expression) + (expected == NULL ? 0 : strlen(expected)) + 100;
		char *message = malloc(capacity);
		if (message == NULL) {
			static const char failure[] = "out of memory";
			adamic_panic(failure, sizeof failure - 1);
		}
		int length = expected != NULL ? snprintf(message, capacity, "field read failed: %s is not initialized; expected %s, found %s", expression, expected, slot == NULL ? "missing" : "uninitialized") : snprintf(message, capacity, "read before assignment: field '%s' in %s", name, expression);
		adamic_panic(message, (size_t)length);
	}
	if (owner != NULL) { *owner = object; }
	return slot;
}

adamic_view_union_value adamic_object_view_union_snapshot(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *declared, bool absent) {
    adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
    // Current callers pass false; a future optional reader can request undefined
    // for a missing own slot. Required reads use the readiness failure below.
    if (object != NULL && slot == NULL && absent) {
        return (adamic_view_union_value){adamic_view_union_undefined, {.reference = NULL}};
    }
    const adamic_object *owner = NULL;
    slot = adamic_object_read_mode(object, name, cache, expression, declared, &owner);
    size_t index = adamic_slot_index(owner, slot);
    unsigned char storage = adamic_object_field_type(owner, index);
    adamic_view_union_value value = {adamic_view_union_unknown, *slot};
    if (storage == adamic_rep_number) { value.kind = adamic_view_union_number; }
    else if (storage == adamic_rep_boolean) { value.kind = adamic_view_union_boolean; }
    else if (storage == adamic_rep_maybe_number) {
        adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
        value.kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;
        value.payload.number = number.number;
    } else if (storage == adamic_rep_maybe_boolean) {
        adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot->maybe_boolean);
        value.kind = boolean.present ? adamic_view_union_boolean : adamic_view_union_undefined;
        value.payload = (adamic_value){0};
        if (boolean.present) { value.payload.boolean = boolean.boolean; }
    } else if (storage == adamic_rep_null) { value.kind = adamic_view_union_null; }
    else if (storage == adamic_rep_undefined) { value.kind = adamic_view_union_undefined; }
    else if ((storage >= adamic_rep_string && storage <= adamic_rep_map) || storage == adamic_rep_closure || storage == adamic_rep_union || storage == adamic_rep_weak) {
        value = adamic_view_union_heap(slot->reference);
        adamic_view_union_kind expected = storage == adamic_rep_string ? adamic_view_union_string : storage == adamic_rep_object || storage == adamic_rep_weak ? adamic_view_union_object : storage == adamic_rep_array ? adamic_view_union_array : storage == adamic_rep_map ? adamic_view_union_map : storage == adamic_rep_closure ? adamic_view_union_function : value.kind;
        if (value.kind != adamic_view_union_undefined && value.kind != expected) { value.kind = adamic_view_union_unknown; }
    }
    return value;
}

adamic_value *adamic_object_read(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression) {
	return adamic_object_read_mode(object, name, cache, expression, NULL, NULL);
}

void adamic_object_set_initialized(adamic_object *object, const char *name, bool initialized) {
	adamic_slot_cache cache = {0};
	// The object's own field: readiness belongs to the object being initialized, never a parent's.
	adamic_value *slot = adamic_object_find(object, name, &cache);
	adamic_object_initialized(object)[adamic_slot_index(object, slot)] = initialized;
	adamic_object_publish(object, adamic_slot_index(object, slot));
}

// heap_kind_name names a heap value's kind the way a checked field read's failure says what it found.
static const char *heap_kind_name(enum adamic_kind kind) {
	switch (kind) {
	case adamic_kind_string: return "string";
	case adamic_kind_object: return "object";
	case adamic_kind_array: return "array";
	case adamic_kind_map: return "Map";
	case adamic_kind_cell: return "cell";
	case adamic_kind_closure: return "function";
	case adamic_kind_map_iterator: return "Map iterator";
	case adamic_kind_number: return "number";
	case adamic_kind_boolean: return "boolean";
	case adamic_kind_weak: return "Weak";
	case adamic_kind_typed_array: return "typed array";
	case adamic_kind_typed_array_iterator: return "typed array iterator";
	case adamic_kind_async_frame: return "async frame";
	case adamic_kind_async_promise: return "promise";
	case adamic_kind_async_reaction: return "async reaction";
	case adamic_kind_foreign: return "foreign value";
	case adamic_kind_weak_map: return "WeakMap or WeakSet";
#ifdef ADAMIC_CANONICAL_CLOSURES
	case adamic_kind_environment: return "environment";
	case adamic_kind_generator: return "generator";
#endif
	case adamic_kind_null: return "null";
	}
	return "unknown heap kind";
}

// view_failure stops a checked field read that found found where it wanted type.
_Noreturn static void view_failure(const char *type, const char *expression, const char *found) {
	size_t capacity = strlen(expression) + strlen(type) + strlen(found) + 100;
	char *message = malloc(capacity);
	if (message == NULL) {
		static const char oom[] = "out of memory";
		adamic_panic(oom, sizeof oom - 1);
	}
	int length = snprintf(message, capacity, "field read failed: %s is not a %s; expected %s, found %s", expression, type, type, found);
	adamic_panic(message, (size_t)length);
}

// Required-field contract checks use the shared readiness bitmap. Representation evidence is
// checked before reading any union member, including before following a possible reference.
adamic_value adamic_object_view(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
	// The receiver's own kind comes first: another heap kind (a string, a generator's frame) has no
	// shape to read, not even the one adamic_record_is compares. A record is an object, so it passes.
	if (object != NULL && object->heap.kind != adamic_kind_object) { view_failure(type, expression, heap_kind_name(object->heap.kind)); }
	if (object != NULL && adamic_record_is(object)) return adamic_record_view(object, name, wanted, type, expression);
	const adamic_object *owner = NULL;
	adamic_value *slot = adamic_object_read_mode(object, name, cache, expression, type, &owner);
	unsigned char actual = adamic_object_field_type(owner, adamic_slot_index(owner, slot));
	// Boxed unions and packed maybe-numbers have a real runtime tag. Convert only
	// after that tag proves which payload is live; never interpret a pointer as a number.
	if (actual == 10 && slot->reference != NULL) {
		const adamic_heap *boxed = slot->reference;
		if (wanted == 1 && boxed->kind == adamic_kind_number) { return (adamic_value){.number = ((const adamic_number_box *)boxed)->number}; }
		if (wanted == 2 && boxed->kind == adamic_kind_boolean) { return (adamic_value){.boolean = ((const adamic_boolean_box *)boxed)->boolean}; }
		if (wanted >= 3 && wanted <= 6) {
			enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : adamic_kind_map;
			if (boxed->kind == kind) { return *slot; }
		}
	}
	if (actual == 7 && wanted == 1) {
		adamic_maybe_number unpacked = adamic_maybe_number_unpack(slot->number);
		if (unpacked.present) { return (adamic_value){.number = unpacked.number}; }
	}
	if (actual == wanted && wanted >= 1 && wanted <= 6) {
		if (wanted <= 2) { return *slot; }
		const adamic_heap *reference = slot->reference;
		enum adamic_kind kind = wanted == 3 ? adamic_kind_string : wanted == 4 ? adamic_kind_object : wanted == 5 ? adamic_kind_array : adamic_kind_map;
		if (reference != NULL && reference->kind == kind) { return *slot; }
	}
	const char *found = actual == 1 ? "number" : actual == 2 ? "boolean" : actual == 3 ? "string" : actual == 4 ? "object" : actual == 5 ? "array" : actual == 6 ? "Map" : actual == 7 ? "number" : actual == 8 ? "function" : actual == 11 ? "object" : "unsupported representation";
	if (actual >= 3 && actual <= 6 && slot->reference == NULL) { found = "nullish"; }
	if (actual == 7 && !adamic_maybe_number_unpack(slot->number).present) { found = "nullish"; }
	if (actual == 10) {
		const adamic_heap *boxed = slot->reference;
		// A union's null is nullish, as the JavaScript backend's adamicViewField says.
		found = boxed == NULL || boxed->kind == adamic_kind_null ? "nullish" : boxed->kind == adamic_kind_number ? "number" : boxed->kind == adamic_kind_boolean ? "boolean" : boxed->kind == adamic_kind_string ? "string" : boxed->kind == adamic_kind_object ? "object" : boxed->kind == adamic_kind_array ? "array" : boxed->kind == adamic_kind_map ? "Map" : "function";
	}
	if ((actual == adamic_rep_object || actual == adamic_rep_union || actual == adamic_rep_weak) && slot->reference != NULL && ((const adamic_heap *)slot->reference)->kind == adamic_kind_object) {
		const adamic_object *value = slot->reference;
		if (value->tuple) { found = "array"; }
		else if (value->class != NULL && value->class->is_static) { found = "function"; }
	}
	view_failure(type, expression, found);
}

void adamic_view_literal_failure(const char *expression, const char *expected, unsigned char type, adamic_value value) {
	const adamic_string *text = type == 3 ? value.reference : type == 1 ? adamic_string_from_number(value.number) : value.boolean ? &adamic_string_true : &adamic_string_false;
	const char *kind = type == 3 ? "string" : type == 1 ? "number" : "boolean";
	size_t capacity = strlen(expression) + strlen(expected) + text->length + 100;
	char *message = malloc(capacity);
	if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
	int prefix = snprintf(message, capacity, "field read failed: %s expected %s, found %s ", expression, expected, kind);
	memcpy(message + prefix, text->bytes, text->length);
	adamic_panic(message, (size_t)prefix + text->length);
}

// Writes through an asserted view must not reinterpret or release a differently typed slot.
void adamic_object_view_write(adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
	adamic_value *slot = adamic_object_optional_find(object, name, cache);
	// An absent reserved field still has writable storage. Find it without rereading a shared cache.
	if (slot == NULL) {
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(object->shape->names[index], name) == 0) { slot = &object->slots[index]; break; }
		}
	}
	if (slot != NULL) {
		size_t index = adamic_slot_index(object, slot);
		unsigned char actual = adamic_object_field_type(object, index);
		bool reference_write = (wanted >= 3 && wanted <= 6) || wanted == 8 || wanted == 10;
		if (actual == wanted || (actual == 13 && reference_write && object->shape->references[index]) || (actual == 10 && wanted <= 2) || (actual == 7 && wanted == 1)) { return; }
	}
	(void)adamic_object_view(object, name, cache, wanted, type, expression);
}

adamic_maybe_boolean adamic_object_maybe_boolean(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	const adamic_object *owner = object;
	adamic_value *slot = adamic_object_field_owned(object, name, cache, &owner);
	if (owner->shape->references[adamic_slot_index(owner, slot)]) {
		if (slot->reference != NULL) {
			static const char message[] = "compiler bug: a boolean field holds a reference";
			adamic_panic(message, sizeof message - 1);
		}
		return (adamic_maybe_boolean){false, false};
	}
	return adamic_maybe_boolean_unpack(slot->maybe_boolean);
}
