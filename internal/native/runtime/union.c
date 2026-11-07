// union.c: a union whose members are held differently, as one counted reference (adamic.h).

#include "adamic.h"
#include "dynamic_keyed_names.h"

#include <math.h>
#include <pthread.h>
#include <stdlib.h>
#include <string.h>

adamic_heap adamic_null = {0, adamic_kind_object, 0};

adamic_boolean_box adamic_box_true = {{0, adamic_kind_boolean, 0}, true};
adamic_boolean_box adamic_box_false = {{0, adamic_kind_boolean, 0}, false};

adamic_string adamic_typeof_number = ADAMIC_STRING("number");
adamic_string adamic_typeof_string = ADAMIC_STRING("string");
adamic_string adamic_typeof_boolean = ADAMIC_STRING("boolean");
adamic_string adamic_typeof_undefined = ADAMIC_STRING("undefined");
adamic_string adamic_typeof_object = ADAMIC_STRING("object");
adamic_string adamic_typeof_function = ADAMIC_STRING("function");

adamic_heap *adamic_box_number(double number) {
	adamic_number_box *box = adamic_allocate(sizeof *box, adamic_kind_number);
	box->number = number;
	return &box->heap;
}

bool adamic_union_equal(const adamic_heap *left, const adamic_heap *right) {
	if (left == NULL || right == NULL) {
		return left == right;
	}
	if (left->kind != right->kind) {
		return false;
	}
	switch (left->kind) {
	case adamic_kind_number:
		// Compared as numbers, never as boxes: NaN is never ===, even to itself, and -0 === 0.
		return ((const adamic_number_box *)left)->number == ((const adamic_number_box *)right)->number;
	case adamic_kind_string:
		return adamic_string_equal((const adamic_string *)left, (const adamic_string *)right);
	default:
		// A boolean is one of two boxes, and an object, array, map or function is equal only to itself.
		return left == right;
	}
}

adamic_string *adamic_union_to_string(adamic_heap *value) {
	if (value == NULL) {
		return &adamic_string_undefined;
	}
	switch (value->kind) {
	case adamic_kind_number:
		return adamic_string_from_number(((adamic_number_box *)value)->number);
	case adamic_kind_boolean:
		return ((adamic_boolean_box *)value)->boolean ? &adamic_string_true : &adamic_string_false;
	case adamic_kind_string:
		return adamic_retain(value);
	default: {
		// Lowering lets only numbers, booleans, strings and undefined through; anything else here is a
		// compiler bug, said out loud.
		static const char message[] = "a union holding an object was written as a string";
		adamic_panic(message, sizeof message - 1);
	}
	}
}

adamic_string *adamic_union_typeof(const adamic_heap *value, bool null) {
	if (value == &adamic_null) { return &adamic_typeof_object; }
	if (value == NULL) {
		return null ? &adamic_typeof_object : &adamic_typeof_undefined;
	}
	switch (value->kind) {
	case adamic_kind_number:
		return &adamic_typeof_number;
	case adamic_kind_boolean:
		return &adamic_typeof_boolean;
	case adamic_kind_string:
		return &adamic_typeof_string;
	case adamic_kind_closure:
		return &adamic_typeof_function;
	case adamic_kind_object: {
		const adamic_object *object = (const adamic_object *)value;
		return object->class != NULL && object->class->is_static ? &adamic_typeof_function : &adamic_typeof_object;
	}
	default:
		return &adamic_typeof_object;
	}
}

// The numbers are internal/ir's UnknownKind numbers, which the JavaScript backend's adamicUnknownKind
// answers too, in this order: an Error before a class, since an Error's shape has no class. A number
// rather than a name, so the runtime keeps no storage for it.
double adamic_union_kind(const adamic_heap *value) {
	enum { undefined, null, number, boolean, string, error, instance, static_class, object, array, function, collection, typed_array, other };
	if (value == NULL) return undefined;
	if (value == &adamic_null) return null;
	switch (value->kind) {
	case adamic_kind_number:
		return number;
	case adamic_kind_boolean:
		return boolean;
	case adamic_kind_string:
		return string;
	case adamic_kind_closure:
		return function;
	case adamic_kind_array:
		return array;
	case adamic_kind_map:
		return collection;
	case adamic_kind_typed_array:
		return typed_array;
	case adamic_kind_object: {
		const adamic_object *held = (const adamic_object *)value;
		// A tuple is an array on Node, and on the JavaScript backend.
		if (held->tuple) return array;
		if (adamic_error_is(held)) return error;
		if (held->class != NULL) return held->class->is_static ? static_class : instance;
		return object;
	}
	default:
		return other;
	}
}

// Only a reference slot can hold a function, so a scalar one, a record of numbers or booleans among
// them, never answers yes, and its bits are never read as a pointer.
bool adamic_object_own_function(const adamic_heap *value, const char *name) {
	// ToPrimitive's own-member lookup, which on a host object would answer from its slots, not
	// Node's prototype (a matchAll iterator's String is [object RegExp String Iterator]).
	adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
	adamic_host_stop(value, &key);
	const adamic_object *object = (const adamic_object *)value;
	if (adamic_record_is(object)) {
		if (!((const adamic_map *)object->slots[0].reference)->reference_values) return false;
		adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
		const adamic_value *own = adamic_record_get_own(object, &key);
		const adamic_heap *held = own == NULL ? NULL : own->reference;
		return held != NULL && held->kind == adamic_kind_closure;
	}
	for (size_t index = 0; index < object->shape->count; index++) {
		if (strcmp(name, object->shape->names[index]) != 0) continue;
		if (!adamic_object_present(object, index) || !adamic_object_initialized(object)[index] || !object->shape->references[index]) return false;
		const adamic_heap *held = object->slots[index].reference;
		return held != NULL && held->kind == adamic_kind_closure;
	}
	return false;
}

// Program shape metadata is registered when an object of the shape is made: the emitter's and the
// runtime's for its plain results, static, and the JSON decoder's, one for each canonical shape it
// builds, which lives until exit (json_decode.c). The list answers a scalar slot's representation
// (dynamic_slot); it no longer says whose prototype an object has, which its shape's methods do.
// Objects are made in parallel tasks too, so an entry is pushed once under the lock and published with
// release order; readers load the list head with acquire, and an entry's fields never change once
// it's on the list.
static adamic_shape_types *shape_types;
static pthread_mutex_t shape_types_lock = PTHREAD_MUTEX_INITIALIZER;

void adamic_register_shape_types(adamic_shape_types *metadata) {
	if (__atomic_load_n(&metadata->registered, __ATOMIC_ACQUIRE)) return;
	pthread_mutex_lock(&shape_types_lock);
	if (!metadata->registered) {
		metadata->next = shape_types;
		__atomic_store_n(&shape_types, metadata, __ATOMIC_RELEASE);
		__atomic_store_n(&metadata->registered, true, __ATOMIC_RELEASE);
	}
	pthread_mutex_unlock(&shape_types_lock);
}

static bool named(const char *name, const char *const *names, size_t count) {
	for (size_t index = 0; index < count; index++) {
		if (strcmp(name, names[index]) == 0) return true;
	}
	return false;
}

static bool object_prototype_name(const char *name) {
	static const char *const names[] = {"constructor", "__proto__", "toString", "toLocaleString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__"};
	return named(name, names, sizeof names / sizeof names[0]);
}

// array_key says whether a name of length bytes is a canonical array index, digits with no leading
// zero below 2^32-1, and which. A run-time key's bytes aren't terminated, so it's given by length.
static bool array_key(const char *name, size_t length, size_t *result) {
	if (length == 0 || (name[0] == '0' && length > 1)) return false;
	uint64_t index = 0;
	for (size_t at = 0; at < length; at++) {
		unsigned char digit = (unsigned char)name[at];
		if (digit < '0' || digit > '9') return false;
		index = index * 10 + (digit - '0');
		if (index >= UINT32_MAX) return false;
	}
	*result = (size_t)index;
	return true;
}

static bool host_object(const adamic_object *object);
_Noreturn static void keyed_stop(enum adamic_dynamic_keyed_stop stop, const adamic_string *key);

// host_stop is the guard at each entrance by which a name or a key reaches a value viewed as
// unknown: an in test (adamic_has_property), a read narrowed by one (adamic_dynamic_property) and a
// run-time key read (adamic_dynamic_property_keyed, which tests host_object itself, after its own
// earlier stops). On a host object each stops in the runtime value form before a slot or a slot's
// name is consulted, since the object's own names aren't Node's (Adamic's ruling on #ey7aysy at
// 23:03Z): a Stats holds _fsFileMode and lacks isFile. An Error isn't a host object, so it answers
// its own names on all three as Node does. The runtime's other readers of a value's names by a key
// call adamic_host_stop, and host_shapes_test.go holds every function that reads a shape's names to
// one or the other, or to an allowlist that says why it needn't.
void adamic_host_stop(const adamic_heap *value, const adamic_string *key) {
	if (value == NULL || value == &adamic_null || value->kind != adamic_kind_object || !host_object((const adamic_object *)value)) return;
	keyed_stop(adamic_dynamic_keyed_stop_runtime_value, key);
}

// adamic_host_stop_keys stops Object.keys of a host object in a message of its own, with no key, since
// an enumeration reads none (Adamic's ruling on #ey7aysy at 08:28Z; ir.DynamicKeysStop's words).
void adamic_host_stop_keys(const adamic_heap *value) {
	if (value == NULL || value == &adamic_null || value->kind != adamic_kind_object || !host_object((const adamic_object *)value)) return;
	static const char message[] = "this runtime value's own keys aren't enumerated yet";
	adamic_panic(message, sizeof message - 1);
}

static void host_stop(const adamic_heap *value, const char *name) {
	adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
	adamic_host_stop(value, &key);
}

// has_property answers an in test past the guard: an array's own properties are read through it, so
// a match's index answers though the object holding it is marked host.
static bool has_property(const adamic_heap *value, const char *name) {
	if (value == NULL || value == &adamic_null) {
		adamic_panic("in on null or undefined", sizeof "in on null or undefined" - 1);
	}
	if (value->kind == adamic_kind_object) {
		const adamic_object *object = (const adamic_object *)value;
		if (adamic_record_is(object)) {
			adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
			return adamic_record_has(object, &key);
		}
		// Error.prototype supplies message, so on Node an Error made without one still has it, as "". Its
		// absent-message shape keeps that "" in #message (exceptions.c), so look for it there.
		const char *key = name;
		name = adamic_error_field_name(object, name);
		// Only the plain key "message", mapped above, reaches that slot. On Node a string key never names
		// a private and an Error has no own or inherited "#message", so a # key on that shape is false.
		// The shape is the one whose "message" is mapped; a plain object's '#message' key isn't touched.
		if (key[0] == '#' && adamic_error_field_name(object, "message")[0] == '#') return false;
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(name, object->shape->names[index]) != 0) continue;
			if (!adamic_object_present(object, index)) continue;
			if (object->class == NULL || key[0] != '#') return true;
			const adamic_shape *public = object->class->public_shape;
			if (named(name, public->names, public->count)) return true;
		}
		const adamic_methods *methods = object->shape->methods;
		if (methods != NULL && named(name, methods->names, methods->count)) return true;
		for (const adamic_class *class = object->class; class != NULL; class = class->base) {
			for (size_t index = 0; index < class->accessor_count; index++) {
				if (strcmp(name, class->accessors[index].name) == 0) return true;
			}
		}
		return object_prototype_name(name);
	}
	if (value->kind == adamic_kind_array) {
		const adamic_array *array = (const adamic_array *)value;
		size_t index;
		if (array_key(name, strlen(name), &index)) return index < array->length;
		if (array->metadata != NULL && has_property(&array->metadata->heap, name)) return true;
		if (array->properties != NULL && has_property(&array->properties->heap, name)) return true;
		static const char *const names[] = {"length", "at", "concat", "copyWithin", "fill", "find", "findIndex", "findLast", "findLastIndex", "lastIndexOf", "pop", "push", "reverse", "shift", "unshift", "slice", "sort", "splice", "includes", "indexOf", "join", "keys", "entries", "values", "forEach", "filter", "flat", "flatMap", "map", "every", "some", "reduce", "reduceRight", "toReversed", "toSorted", "toSpliced", "with"};
		return object_prototype_name(name) || named(name, names, sizeof names / sizeof names[0]);
	}
	adamic_panic("in on an unsupported runtime value", sizeof "in on an unsupported runtime value" - 1);
}

bool adamic_has_property(const adamic_heap *value, const char *name) {
	host_stop(value, name);
	return has_property(value, name);
}

static adamic_heap *dynamic_slot(const adamic_object *object, size_t index) {
	if (!adamic_object_present(object, index) || !adamic_object_initialized(object)[index]) return NULL;
	const adamic_value slot = object->slots[index];
	if (object->shape->references[index]) return adamic_retain(slot.reference);
	// The slot's own representation byte is ir.Type's, as its last write left it, whatever made the
	// object: every allocator zeroes the bytes, the emitter writes one with each construction and write
	// in a program that reads by a run-time key (fieldTypesNeeded), and the runtime writes those of
	// the plain objects it hands a program (json_decode.c, map_set.c, directory.c, regexp.c). Zero is
	// no evidence, so the shape's registered tags decide then.
	switch (adamic_object_field_types(object)[index]) {
	case 1: return adamic_box_number(slot.number);
	case 2: return slot.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
	case 7: {
		adamic_maybe_number number = adamic_maybe_number_unpack(slot.number);
		return number.present ? adamic_box_number(number.number) : NULL;
	}
	// A boolean iterator's result at done (map_set.c's adamic_iterator_result).
	case 9: {
		adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot.maybe_boolean);
		if (!boolean.present) return NULL;
		return boolean.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
	}
	}
	for (const adamic_shape_types *entry = __atomic_load_n(&shape_types, __ATOMIC_ACQUIRE); entry != NULL; entry = entry->next) {
		if (entry->shape != object->shape) continue;
		// These are ir.Type's scalar representations, written by the emitter.
		switch (entry->types[index]) {
		case 1: return adamic_box_number(slot.number);
		case 2: return slot.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
		case 7: {
			adamic_maybe_number number = adamic_maybe_number_unpack(slot.number);
			return number.present ? adamic_box_number(number.number) : NULL;
		}
		case 9: {
			adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot.maybe_boolean);
			if (!boolean.present) return NULL;
			return boolean.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
		}
		}
	}
	adamic_panic("dynamic read of a host scalar without type metadata", sizeof "dynamic read of a host scalar without type metadata" - 1);
}

// dynamic_property is a narrowed read past the guard, as has_property is an in test.
static adamic_heap *dynamic_property(adamic_heap *value, const char *name) {
	if (value != NULL && value->kind == adamic_kind_closure && strcmp(name, "length") == 0) {
		return adamic_box_number((double)((adamic_closure *)value)->source_length);
	}
	if (value == NULL || value == &adamic_null) {
		adamic_panic("dynamic property read on null or undefined", sizeof "dynamic property read on null or undefined" - 1);
	}
	if (value->kind == adamic_kind_object) {
		const adamic_object *object = (const adamic_object *)value;
		if (adamic_record_is(object)) {
			adamic_string key = {{0, adamic_kind_string, 0}, strlen(name), name, 0, NULL, NULL, 0};
			const adamic_value *slot = adamic_record_get(object, &key);
			return slot == NULL ? NULL : adamic_retain(slot->reference);
		}
		// The inherited message, as in adamic_has_property.
		name = adamic_error_field_name(object, name);
		for (size_t index = 0; index < object->shape->count; index++) {
			if (strcmp(name, object->shape->names[index]) == 0) return dynamic_slot(object, index);
		}
		return NULL;
	}
	if (value->kind == adamic_kind_array) {
		adamic_array *array = (adamic_array *)value;
		if (strcmp(name, "length") == 0) return adamic_box_number((double)array->length);
		if (array->metadata != NULL) {
			for (size_t index = 0; index < array->metadata->shape->count; index++) {
				if (strcmp(name, array->metadata->shape->names[index]) == 0) return dynamic_slot(array->metadata, index);
			}
		}
		if (array->properties != NULL) return dynamic_property(&array->properties->heap, name);
		return NULL;
	}
	adamic_panic("dynamic property read on an unsupported runtime value", sizeof "dynamic property read on an unsupported runtime value" - 1);
}

adamic_heap *adamic_dynamic_property(adamic_heap *value, const char *name) {
	host_stop(value, name);
	return dynamic_property(value, name);
}

// A run-time key is any string, a slice, a view or one with bytes of its own, so it's compared with a
// name by its content, never by pointer or a cached hash.
static bool key_is(const adamic_string *key, const char *name) {
	size_t length = strlen(name);
	return key->length == length && memcmp(key->bytes, name, length) == 0;
}

static bool key_named(const adamic_string *key, const char *const *names, size_t count) {
	for (size_t index = 0; index < count; index++) {
		if (key_is(key, names[index])) return true;
	}
	return false;
}

// keyed_stop says why a run-time key read stops, in the words the JavaScript backend prints too
// (dynamic_keyed_names.h), and then which key it was reading.
_Noreturn static void keyed_stop(enum adamic_dynamic_keyed_stop stop, const adamic_string *key) {
	const char *reason = adamic_dynamic_keyed_stop_messages[stop];
	static const char reading[] = " (reading '";
	// No overflow check: a key is a string the program made, and no string comes near SIZE_MAX bytes
	// (V8's own limit is about 2^29 characters), so the sum can't wrap.
	size_t length = strlen(reason) + sizeof reading - 1 + key->length + 2;
	char *message = malloc(length);
	if (message == NULL) {
		static const char out_of_memory[] = "out of memory";
		adamic_panic(out_of_memory, sizeof out_of_memory - 1);
	}
	char *at = message;
	memcpy(at, reason, strlen(reason));
	at += strlen(reason);
	memcpy(at, reading, sizeof reading - 1);
	at += sizeof reading - 1;
	memcpy(at, key->bytes, key->length);
	at += key->length;
	memcpy(at, "')", 2);
	adamic_panic(message, length);
}

// keyed_accessor is the accessor the key names on the object's class chain, as adamic_accessor_find
// finds one by a name of the program's, or NULL.
static const adamic_accessor *keyed_accessor(const adamic_object *object, const adamic_string *key) {
	for (const adamic_class *class = object->class; class != NULL; class = class->base) {
		for (size_t index = 0; index < class->accessor_count; index++) {
			if (key_is(key, class->accessors[index].name)) return &class->accessors[index];
		}
	}
	return NULL;
}

// keyed_getter calls a class getter with the object as this, as Reflect.get with the object as the
// receiver does, and boxes what it gives by the getter's own representation (ir.Type's, written by
// the emitter into the accessor), owned. Lowering admits a run-time key read only where every class
// getter is one return of its own fields, literals and operators, which can't throw, of a
// representation boxed here, and null only as a union (reflect_get.go).
static adamic_heap *keyed_getter(adamic_object *object, const adamic_accessor *accessor) {
	adamic_value value = accessor->get(object);
	switch (accessor->type) {
	case 1: return adamic_box_number(value.number);
	case 2: return value.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
	case 7: {
		adamic_maybe_number number = adamic_maybe_number_unpack(value.number);
		return number.present ? adamic_box_number(number.number) : NULL;
	}
	case 9: {
		adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(value.maybe_boolean);
		if (!boolean.present) return NULL;
		return boolean.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
	}
	// A string, an object, an array, a Map, a closure, a union or a record is one counted reference,
	// which a union holds as it is, undefined as NULL.
	case 3: case 4: case 5: case 6: case 8: case 10: case 14:
		return value.reference;
	default: {
		static const char message[] = "compiler bug: a getter read by a run-time key has no union representation";
		adamic_panic(message, sizeof message - 1);
	}
	}
}

// keyed_record_value is a record's own value as a union, owned, boxed by the kind its shape names
// (record.c): a number or a boolean may be packed with undefined, and reads as undefined then. A
// number's undefined is a NaN payload arithmetic never makes, and no program can store raw bytes as a
// number today: ArrayBuffer and DataView are NotYet (lower's typed_arrays.go), so no two views share a
// buffer. Whatever admits them canonicalizes a NaN read from bytes before it's stored, as function
// arguments are.
static adamic_heap *keyed_record_value(const adamic_record *record, const adamic_value *own) {
	switch (adamic_record_values(record)) {
	case adamic_field_reference:
		return adamic_retain(own->reference);
	case adamic_field_number: {
		adamic_maybe_number number = adamic_maybe_number_unpack(own->number);
		return number.present ? adamic_box_number(number.number) : NULL;
	}
	case adamic_field_boolean: {
		adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(own->maybe_boolean);
		if (!boolean.present) return NULL;
		return boolean.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
	}
	}
	static const char message[] = "compiler bug: a record of unknown value kind";
	adamic_panic(message, sizeof message - 1);
}

// keyed_class is the emitter's row for the object's class, found by the class itself, or NULL.
static const adamic_keyed_class *keyed_class(const adamic_object *object, const adamic_keyed_class *classes, size_t count) {
	for (size_t index = 0; index < count; index++) {
		if (classes[index].class == object->class) return &classes[index];
	}
	return NULL;
}

// keyed_slot is where an object's own shape holds the name, never a slot cache's packed word (another
// shape's, as another thread may have left it), or the shape's count.
static size_t keyed_slot(const adamic_object *object, const adamic_string *key, const char *remapped) {
	for (size_t index = 0; index < object->shape->count; index++) {
		const char *name = object->shape->names[index];
		if (remapped != NULL ? strcmp(name, remapped) == 0 : key_is(key, name)) return index;
	}
	return object->shape->count;
}

// Object.prototype's methods, and Object as their constructor, as a run-time key read hands them out:
// one immortal closure for each, the same from every receiver and every read, as Node's are one
// function each, with Node's length. No call reaches one yet: lowering refuses every call through an
// unknown narrowed to a function ("calling through the unchecked Function type", function_rulings.go),
// so each one's code is a guard that stops loudly.
static adamic_value prototype_method_called(adamic_closure *self, adamic_value *arguments);
static adamic_value object_called(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	(void)arguments;
	static const char message[] = "Object isn't called yet";
	adamic_panic(message, sizeof message - 1);
}
static adamic_closure object_constructor = {.heap = {0, adamic_kind_closure, 0}, .code = object_called, .count = 0, .source_length = 1};
static adamic_closure prototype_to_string = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 0};
static adamic_closure prototype_to_locale_string = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 0};
static adamic_closure prototype_value_of = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 0};
static adamic_closure prototype_has_own_property = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 1};
static adamic_closure prototype_is_prototype_of = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 1};
static adamic_closure prototype_property_is_enumerable = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 1};
static adamic_closure prototype_define_getter = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 2};
static adamic_closure prototype_define_setter = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 2};
static adamic_closure prototype_lookup_getter = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 1};
static adamic_closure prototype_lookup_setter = {.heap = {0, adamic_kind_closure, 0}, .code = prototype_method_called, .count = 0, .source_length = 1};
// In adamic_dynamic_keyed_prototype_method_names' order.
static adamic_closure *const prototype_methods[] = {
	&prototype_to_string, &prototype_to_locale_string, &prototype_value_of, &prototype_has_own_property, &prototype_is_prototype_of,
	&prototype_property_is_enumerable, &prototype_define_getter, &prototype_define_setter, &prototype_lookup_getter, &prototype_lookup_setter,
};
_Static_assert(sizeof prototype_methods / sizeof prototype_methods[0] == sizeof adamic_dynamic_keyed_prototype_method_names / sizeof adamic_dynamic_keyed_prototype_method_names[0], "one closure for each of Object.prototype's methods");

// not_called stops a call of a prototype's closure, naming it: the prefix and then the name.
_Noreturn static void not_called(const char *prefix, const char *name) {
	static const char suffix[] = " isn't called yet";
	size_t length = strlen(prefix) + strlen(name) + sizeof suffix - 1;
	char *message = malloc(length);
	if (message == NULL) {
		static const char out_of_memory[] = "out of memory";
		adamic_panic(out_of_memory, sizeof out_of_memory - 1);
	}
	memcpy(message, prefix, strlen(prefix));
	memcpy(message + strlen(prefix), name, strlen(name));
	memcpy(message + strlen(prefix) + strlen(name), suffix, sizeof suffix - 1);
	adamic_panic(message, length);
}

static adamic_value prototype_method_called(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	const char *name = "an unknown method";
	for (size_t index = 0; index < sizeof prototype_methods / sizeof prototype_methods[0]; index++) {
		if (prototype_methods[index] == self) name = adamic_dynamic_keyed_prototype_method_names[index];
	}
	not_called("Object.prototype.", name);
}

// Array.prototype's own functions, which an array and a tuple inherit, as a run-time key read hands
// them out (#g187f1e): one immortal closure for each name in
// adamic_dynamic_keyed_array_prototype_names, the same from every array and every read, as Node's are
// one function each, with Node's length (dynamic_keyed_names.h, pinned by
// TestDynamicKeyedArrayPrototypeLengthsAreNodes). An array's constructor, toString and toLocaleString
// are these, never Object.prototype's, as on Node. No call reaches one yet, as none reaches
// Object.prototype's, so each one's code is a guard that stops loudly.
static adamic_value array_prototype_called(adamic_closure *self, adamic_value *arguments);
#define ARRAY_PROTOTYPE_CLOSURE(index, length) static adamic_closure array_prototype_##index = {.heap = {0, adamic_kind_closure, 0}, .code = array_prototype_called, .count = 0, .source_length = length};
ADAMIC_DYNAMIC_KEYED_ARRAY_PROTOTYPE_FUNCTIONS(ARRAY_PROTOTYPE_CLOSURE)
#undef ARRAY_PROTOTYPE_CLOSURE
#define ARRAY_PROTOTYPE_CLOSURE(index, length) &array_prototype_##index,
// In adamic_dynamic_keyed_array_prototype_names' order.
static adamic_closure *const array_prototype_functions[] = {ADAMIC_DYNAMIC_KEYED_ARRAY_PROTOTYPE_FUNCTIONS(ARRAY_PROTOTYPE_CLOSURE)};
#undef ARRAY_PROTOTYPE_CLOSURE
_Static_assert(sizeof array_prototype_functions / sizeof array_prototype_functions[0] == sizeof adamic_dynamic_keyed_array_prototype_names / sizeof adamic_dynamic_keyed_array_prototype_names[0], "one closure for each of Array.prototype's functions");

static adamic_value array_prototype_called(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	const char *name = "an unknown method";
	for (size_t index = 0; index < sizeof array_prototype_functions / sizeof array_prototype_functions[0]; index++) {
		if (array_prototype_functions[index] == self) name = adamic_dynamic_keyed_array_prototype_names[index];
	}
	if (strcmp(name, "constructor") == 0) not_called("", "Array");
	not_called("Array.prototype.", name);
}

// The functions a matchAll iterator inherits, as a run-time key read hands them out (#zhrjf2t): one
// immortal closure for each name in adamic_dynamic_keyed_regexp_string_iterator_names,
// %RegExpStringIteratorPrototype%'s next and then Iterator.prototype's own, its constructor (Iterator)
// among them, the same from every iterator and every read, as Node's are one function each, with
// Node's length (dynamic_keyed_names.h, pinned by TestDynamicKeyedRegExpStringIteratorLengthsAreNodes).
// No call reaches one yet, as none reaches Object.prototype's, so each one's code is a guard that
// stops loudly.
static adamic_value regexp_string_iterator_called(adamic_closure *self, adamic_value *arguments);
#define REGEXP_STRING_ITERATOR_CLOSURE(index, length) static adamic_closure regexp_string_iterator_##index = {.heap = {0, adamic_kind_closure, 0}, .code = regexp_string_iterator_called, .count = 0, .source_length = length};
ADAMIC_DYNAMIC_KEYED_REGEXP_STRING_ITERATOR_FUNCTIONS(REGEXP_STRING_ITERATOR_CLOSURE)
#undef REGEXP_STRING_ITERATOR_CLOSURE
#define REGEXP_STRING_ITERATOR_CLOSURE(index, length) &regexp_string_iterator_##index,
// In adamic_dynamic_keyed_regexp_string_iterator_names' order.
static adamic_closure *const regexp_string_iterator_functions[] = {ADAMIC_DYNAMIC_KEYED_REGEXP_STRING_ITERATOR_FUNCTIONS(REGEXP_STRING_ITERATOR_CLOSURE)};
#undef REGEXP_STRING_ITERATOR_CLOSURE
_Static_assert(sizeof regexp_string_iterator_functions / sizeof regexp_string_iterator_functions[0] == sizeof adamic_dynamic_keyed_regexp_string_iterator_names / sizeof adamic_dynamic_keyed_regexp_string_iterator_names[0], "one closure for each function a matchAll iterator inherits");

static adamic_value regexp_string_iterator_called(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	const char *name = "an unknown method";
	for (size_t index = 0; index < sizeof regexp_string_iterator_functions / sizeof regexp_string_iterator_functions[0]; index++) {
		if (regexp_string_iterator_functions[index] == self) name = adamic_dynamic_keyed_regexp_string_iterator_names[index];
	}
	if (strcmp(name, "next") == 0) not_called("RegExp String Iterator's ", name);
	if (strcmp(name, "constructor") == 0) not_called("", "Iterator");
	not_called("Iterator.prototype.", name);
}

// The functions an fs.Stats inherits, as a run-time key read hands them out (#zhrjf2t): one immortal
// closure for each name in adamic_dynamic_keyed_stats_names, Stats.prototype's constructor (Stats) and
// _checkModeProperty and then StatsBase.prototype's is methods, the same from every Stats and every
// read, as Node's are one function each, with Node's length (dynamic_keyed_names.h, pinned by
// TestDynamicKeyedStatsLengthsAreNodes). No call reaches one yet, as none reaches Object.prototype's,
// so each one's code is a guard that stops loudly, naming the prototype Node's function lives on:
// "Stats isn't called yet", "Stats.prototype._checkModeProperty isn't called yet",
// "StatsBase.prototype.isFile isn't called yet".
static adamic_value stats_called(adamic_closure *self, adamic_value *arguments);
#define STATS_CLOSURE(index, length) static adamic_closure stats_##index = {.heap = {0, adamic_kind_closure, 0}, .code = stats_called, .count = 0, .source_length = length};
ADAMIC_DYNAMIC_KEYED_STATS_FUNCTIONS(STATS_CLOSURE)
#undef STATS_CLOSURE
#define STATS_CLOSURE(index, length) &stats_##index,
// In adamic_dynamic_keyed_stats_names' order.
static adamic_closure *const stats_functions[] = {ADAMIC_DYNAMIC_KEYED_STATS_FUNCTIONS(STATS_CLOSURE)};
#undef STATS_CLOSURE
_Static_assert(sizeof stats_functions / sizeof stats_functions[0] == sizeof adamic_dynamic_keyed_stats_names / sizeof adamic_dynamic_keyed_stats_names[0], "one closure for each function a Stats inherits");

static adamic_value stats_called(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	const char *name = "an unknown method";
	for (size_t index = 0; index < sizeof stats_functions / sizeof stats_functions[0]; index++) {
		if (stats_functions[index] == self) name = adamic_dynamic_keyed_stats_names[index];
	}
	if (strcmp(name, "constructor") == 0) not_called("", "Stats");
	if (strcmp(name, "_checkModeProperty") == 0) not_called("Stats.prototype.", name);
	not_called("StatsBase.prototype.", name);
}

// host_object is the one predicate for a value whose prototype on Node isn't Object's, a program
// class's instance aside (#ey7aysy, Adamic's ruling on #0sps03r): an object with no class, not a
// record, whose shape's methods are adamic_host_methods (adamic.h). The runtime marks each of its own
// shapes whose object's prototype on Node is another's, a Stats, a Dirent, an iterator, a match's
// groups object (whose is null) or process.memoryUsage()'s, and host_shapes_test.go holds every
// runtime shape to that marking. The emitter and the JSON decoder never write the sentinel, so an
// object the program makes is plain, as on Node, whichever path made it. An Error keeps its own
// sentinel and is carved out below. The JavaScript backend's adamicDynamicGet names the same values
// by their prototype (dynamic_keyed.go).
static bool host_object(const adamic_object *object) {
	return object->class == NULL && !adamic_record_is(object) && object->shape->methods == &adamic_host_methods;
}

// prototype_value is what an object inherits from Object.prototype under the key, as an immortal
// closure, or NULL.
static adamic_heap *prototype_value(const adamic_string *key) {
	if (key_is(key, "constructor")) return &object_constructor.heap;
	for (size_t index = 0; index < sizeof prototype_methods / sizeof prototype_methods[0]; index++) {
		if (key_is(key, adamic_dynamic_keyed_prototype_method_names[index])) return &prototype_methods[index]->heap;
	}
	return NULL;
}

// keyed_array_element is an array's element as a union, owned, boxed by how the array holds its
// slots (adamic_array's element byte), as dynamic_slot boxes a slot. A kind not known (0) is a compiler
// bug said out loud: a guess could read a number's bits as a pointer.
static adamic_heap *keyed_array_element(const adamic_array *array, size_t index) {
	const adamic_value element = array->elements[index];
	switch (array->element) {
	case adamic_ir_Number: return adamic_box_number(element.number);
	case adamic_ir_Boolean: return element.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
	case adamic_ir_MaybeNumber: {
		adamic_maybe_number number = adamic_maybe_number_unpack(element.number);
		return number.present ? adamic_box_number(number.number) : NULL;
	}
	case adamic_ir_Union: return adamic_retain(element.reference);
	}
	static const char message[] = "compiler bug: a run-time key read reached an array of unknown element kind";
	adamic_panic(message, sizeof message - 1);
}

// array_inherited is what an array or a tuple inherits under the key, as an immortal closure:
// Array.prototype's own function first, then Object.prototype's one (prototype_value), so an array's
// toString is Array.prototype's and its valueOf Object.prototype's, as on Node. __proto__ stops, as
// it does on an object, and any other key reads undefined (NULL).
static adamic_heap *array_inherited(const adamic_string *key) {
	if (key_is(key, "__proto__")) keyed_stop(adamic_dynamic_keyed_stop_proto, key);
	for (size_t index = 0; index < sizeof array_prototype_functions / sizeof array_prototype_functions[0]; index++) {
		if (key_is(key, adamic_dynamic_keyed_array_prototype_names[index])) return &array_prototype_functions[index]->heap;
	}
	return prototype_value(key);
}

// keyed_array is a run-time key read of an array as Node's answers it (#ys4naww): an index below
// length its element and one past it undefined, length a number, and an own field its value, one the
// program set (metadata) or a match's (properties). A name the array inherits answers its immortal
// closure (array_inherited, #g187f1e), and any other key reads undefined. Only array_key's canonical
// indexes are indexes: "01", "-1" and "1.5" are names, as on Node.
static adamic_heap *keyed_array(const adamic_array *array, const adamic_string *key) {
	size_t index;
	if (array_key(key->bytes, key->length, &index)) return index < array->length ? keyed_array_element(array, index) : NULL;
	if (key_is(key, "length")) return adamic_box_number((double)array->length);
	if (array->metadata != NULL) {
		size_t slot = keyed_slot(array->metadata, key, NULL);
		if (slot < array->metadata->shape->count && adamic_object_present(array->metadata, slot)) return dynamic_slot(array->metadata, slot);
	}
	// A match's fields are regexp.c's match shape, marked host, so they're read here, by the array, and
	// by the shape's kinds: the shape registers no types for dynamic_slot. An indices array (the d
	// flag's) holds the same shape with groups alone set, as Node's holds groups alone, while a result's
	// input is always its string.
	const adamic_object *match = array->properties;
	if (match != NULL) {
		size_t slot = keyed_slot(match, key, NULL);
		size_t input = keyed_slot(match, key, "input");
		bool indices = input < match->shape->count && match->slots[input].reference == NULL;
		if (slot < match->shape->count && (!indices || key_is(key, "groups"))) {
			switch (match->shape->kinds[slot]) {
			case adamic_field_number: return adamic_box_number(match->slots[slot].number);
			case adamic_field_boolean: return match->slots[slot].boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
			case adamic_field_reference: return adamic_retain(match->slots[slot].reference);
			}
		}
	}
	return array_inherited(key);
}

// boxable_type says whether dynamic_slot boxes a scalar of this ir.Type representation: a number, a
// boolean, or a number that may be undefined.
static bool boxable_type(int type) {
	return type == 1 || type == 2 || type == 7;
}

// keyed_tuple_slot is a tuple's slot as a union, owned. dynamic_slot boxes it where it can know the
// slot: a reference, or a scalar whose representation byte or whose shape's registered types name it,
// which a program's own tuple, a Map's spread pair and map_set.c's iterator pair carry. A pair the
// runtime makes with neither, library_object.c's Object.entries pair or regexp.c's indices pair, is
// marked host, and its shape's kinds say how it holds each slot. Those kinds are exact: an entries
// pair's value is a plain number, string or boolean (lowering refuses an optional one), and an indices
// pair holds two numbers. A slot neither path can box is a compiler bug said out loud.
static adamic_heap *keyed_tuple_slot(const adamic_object *tuple, size_t slot) {
	unsigned char type = adamic_object_field_types(tuple)[slot];
	bool registered = false;
	for (const adamic_shape_types *entry = __atomic_load_n(&shape_types, __ATOMIC_ACQUIRE); entry != NULL; entry = entry->next) {
		if (entry->shape == tuple->shape && boxable_type(entry->types[slot])) registered = true;
	}
	if (tuple->shape->references[slot] || boxable_type(type) || registered) return dynamic_slot(tuple, slot);
	if (type == 0 && host_object(tuple)) {
		switch (tuple->shape->kinds[slot]) {
		case adamic_field_number: return adamic_box_number(tuple->slots[slot].number);
		case adamic_field_boolean: return tuple->slots[slot].boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
		}
	}
	static const char message[] = "compiler bug: a run-time key read reached a tuple slot of unknown representation";
	adamic_panic(message, sizeof message - 1);
}

// keyed_tuple is a run-time key read of a tuple by the array rule (#g187f1e), since a tuple is an array
// on Node and on the JavaScript backend: an index below its length answers its slot, one past it
// undefined, length a number, a name it inherits its closure (array_inherited), and any other key
// undefined. Its length is its shape's count: lowering refuses, where an unknown can reach it, a tuple
// with an optional or rest element, whose length on Node the shape can't say (a left-out optional
// element is a slot holding undefined, as one written undefined is), so every tuple that reaches here
// holds all its elements.
static adamic_heap *keyed_tuple(const adamic_object *tuple, const adamic_string *key) {
	size_t index;
	if (array_key(key->bytes, key->length, &index)) {
		if (index >= tuple->shape->count) return NULL;
		size_t slot = keyed_slot(tuple, key, NULL);
		if (slot == tuple->shape->count) {
			static const char message[] = "compiler bug: a tuple's slots aren't named by their indexes";
			adamic_panic(message, sizeof message - 1);
		}
		return keyed_tuple_slot(tuple, slot);
	}
	if (key_is(key, "length")) return adamic_box_number((double)tuple->shape->count);
	return array_inherited(key);
}

// keyed_regexp_string_iterator is a run-time key read of a matchAll iterator as Node's answers it
// (#zhrjf2t, the first host shape lifted): the iterator has no own property on Node, so a name it
// inherits answers its immortal closure, its prototype's or Iterator.prototype's first and then
// Object.prototype's (prototype_value), so its toString is an object's and its constructor Iterator,
// and any other key reads undefined (NULL). Native's iterator holds its regex, input and done as
// slots, which are private: this is handed the key alone, so no slot is ever read.
static adamic_heap *keyed_regexp_string_iterator(const adamic_string *key) {
	for (size_t index = 0; index < sizeof regexp_string_iterator_functions / sizeof regexp_string_iterator_functions[0]; index++) {
		if (key_is(key, adamic_dynamic_keyed_regexp_string_iterator_names[index])) return &regexp_string_iterator_functions[index]->heap;
	}
	return prototype_value(key);
}

// keyed_stats_shape is the fs.Stats shape node_fs_file.c published (adamic_keyed_stats_shape_publish),
// or null before any Stats is made. A release store publishes it and an acquire load reads it, so a
// thread that reads a Stats made on another sees the shape it was made with. It is spelled with const
// after adamic_shape so TestRuntimeFieldLayoutsAreIncluded, which counts every `static const adamic_shape `
// as a shape the runtime makes, doesn't take a pointer to one for a declaration of one.
static adamic_shape const *_Atomic keyed_stats_shape;

void adamic_keyed_stats_shape_publish(const adamic_shape *shape) {
	atomic_store_explicit(&keyed_stats_shape, shape, memory_order_release);
}

// The fields node_fs_file.c's stat shape holds that Node's Stats has too, which a run-time key read
// answers from the shape's slots: its stat_fields, less the private _fsFileMode.
static const char *const stats_held_names[] = {"size", "mtimeMs", "mtime", "atime"};

// keyed_stats is a run-time key read of an fs.Stats as Node's answers it, where native holds the
// answer (#zhrjf2t, the second host shape lifted). Size and mtimeMs are the slot's number and mtime
// and atime its Date, retained, the same object on every read, as Node's getter caches its own. Each
// is read by the shape's kinds, as keyed_array reads a match's fields: the shape registers no types
// for dynamic_slot. A name the Stats inherits answers its immortal closure, Stats.prototype's or
// StatsBase.prototype's first and then Object.prototype's (prototype_value), so its constructor is
// Stats and its toString an object's. One of Node's fields native doesn't hold
// (adamic_dynamic_keyed_stats_field_names) stops in its own form, since native would have to make up
// its value, and
// any other key reads undefined (NULL), the private _fsFileMode among them: no slot is read for any
// key but the four held.
static adamic_heap *keyed_stats(const adamic_object *stats, const adamic_string *key) {
	if (key_named(key, stats_held_names, sizeof stats_held_names / sizeof stats_held_names[0])) {
		size_t slot = keyed_slot(stats, key, NULL);
		if (slot < stats->shape->count) {
			switch (stats->shape->kinds[slot]) {
			case adamic_field_number: return adamic_box_number(stats->slots[slot].number);
			case adamic_field_reference: return adamic_retain(stats->slots[slot].reference);
			}
		}
		static const char message[] = "compiler bug: a Stats' shape doesn't hold a field a run-time key read answers from it";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t index = 0; index < sizeof stats_functions / sizeof stats_functions[0]; index++) {
		if (key_is(key, adamic_dynamic_keyed_stats_names[index])) return &stats_functions[index]->heap;
	}
	if (key_named(key, adamic_dynamic_keyed_stats_field_names, sizeof adamic_dynamic_keyed_stats_field_names / sizeof adamic_dynamic_keyed_stats_field_names[0])) {
		keyed_stop(adamic_dynamic_keyed_stop_stats_field, key);
	}
	return prototype_value(key);
}

adamic_heap *adamic_dynamic_property_keyed(adamic_heap *value, const adamic_string *key, const adamic_keyed_class *classes, size_t count) {
	if (value == NULL || value == &adamic_null) keyed_stop(adamic_dynamic_keyed_stop_nullish, key);
	switch (value->kind) {
	case adamic_kind_number:
	case adamic_kind_boolean:
	case adamic_kind_string:
		keyed_stop(adamic_dynamic_keyed_stop_primitive, key);
	case adamic_kind_closure:
		keyed_stop(adamic_dynamic_keyed_stop_function, key);
	case adamic_kind_array:
		return keyed_array((const adamic_array *)value, key);
	case adamic_kind_map:
		keyed_stop(adamic_dynamic_keyed_stop_collection, key);
	case adamic_kind_typed_array:
		keyed_stop(adamic_dynamic_keyed_stop_typed_array, key);
	case adamic_kind_object:
		break;
	default:
		keyed_stop(adamic_dynamic_keyed_stop_runtime_value, key);
	}
	adamic_object *object = (adamic_object *)value;
	// A tuple is read as an array, before the host test: the pairs the runtime makes are marked host,
	// and their prototype on Node is Array.prototype.
	if (object->tuple) return keyed_tuple(object, key);
	if (object->class != NULL && object->class->is_static) keyed_stop(adamic_dynamic_keyed_stop_static_class, key);
	if (key->length > 0 && key->bytes[0] == '#') keyed_stop(adamic_dynamic_keyed_stop_private, key);
	if (key_is(key, "__proto__")) keyed_stop(adamic_dynamic_keyed_stop_proto, key);
	bool error = adamic_error_is(object);
	if (error && key_named(key, adamic_dynamic_keyed_error_names, sizeof adamic_dynamic_keyed_error_names / sizeof adamic_dynamic_keyed_error_names[0])) {
		keyed_stop(adamic_dynamic_keyed_stop_error_property, key);
	}
	// A matchAll iterator answers as Node's does (#zhrjf2t), known by its shape, before the host test and
	// before any slot is read, since its slots are private. __proto__ has stopped above, as on an object.
	if (adamic_regex_iterator_is(object)) return keyed_regexp_string_iterator(key);
	// So does a Stats, known by the shape node_fs_file.c published (keyed_stats_shape), reading only the
	// slots that hold what Node's own fields hold. union.c never calls into node_fs_file.c, which a
	// runtime without the fs doesn't link.
	const adamic_shape *stats_shape = atomic_load_explicit(&keyed_stats_shape, memory_order_acquire);
	if (stats_shape != NULL && object->shape == stats_shape) return keyed_stats(object, key);
	// Any other host object stops for every key, before its own slots are read: a slot may be internal,
	// its name may be Node's prototype member (a Dirent's isFile), or Node may hold a field native's
	// shape doesn't (memoryUsage's rss). An Error isn't one: its own fields answer as Node's do, and its
	// prototype's names stop below.
	if (host_object(object)) keyed_stop(adamic_dynamic_keyed_stop_runtime_value, key);
	if (object->class != NULL) {
		const adamic_keyed_class *row = keyed_class(object, classes, count);
		if (row == NULL) keyed_stop(adamic_dynamic_keyed_stop_unknown_class, key);
		const adamic_accessor *accessor = keyed_accessor(object, key);
		if (accessor != NULL) {
			if (row->literal) keyed_stop(adamic_dynamic_keyed_stop_object_literal_getter, key);
			if (accessor->get == NULL) keyed_stop(adamic_dynamic_keyed_stop_setter_only, key);
			return keyed_getter(object, accessor);
		}
		// By content, as every name the key meets is.
		if (key_named(key, row->method_names, row->method_count)) keyed_stop(adamic_dynamic_keyed_stop_method, key);
		// An object literal's constructor is Object, as a plain object's is.
		if (!row->literal && key_is(key, "constructor")) keyed_stop(adamic_dynamic_keyed_stop_class_constructor, key);
	}
	if (adamic_record_is(object)) {
		const adamic_value *own = adamic_record_get_own(object, key);
		if (own != NULL) return keyed_record_value(object, own);
	} else {
		// An Error made without a message keeps Error.prototype's empty one in private storage.
		const char *remapped = error && key_is(key, "message") ? adamic_error_field_name(object, "message") : NULL;
		size_t index = keyed_slot(object, key, remapped);
		// An absent optional slot reads undefined, as does a present one holding undefined, but only
		// after the stops above, so neither hides a getter, a method or a private name.
		if (index < object->shape->count && adamic_object_present(object, index)) return dynamic_slot(object, index);
	}
	adamic_heap *inherited = prototype_value(key);
	// An Error's prototype is Error.prototype, whose own members Object's closures would stand in for,
	// so an Object.prototype name stops on it, as on a host object.
	if (inherited != NULL && error) keyed_stop(adamic_dynamic_keyed_stop_runtime_value, key);
	return inherited;
}
