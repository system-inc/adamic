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

// Program shape metadata is registered when an object of the shape is made: the emitter's and the
// runtime's for its plain results, static, and the JSON decoder's, one for each canonical shape it
// builds, which lives until exit (json_decode.c).
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

static bool array_key(const char *name, size_t *result) {
	if (*name == '\0' || (name[0] == '0' && name[1] != '\0')) return false;
	uint64_t index = 0;
	for (const unsigned char *at = (const unsigned char *)name; *at != '\0'; at++) {
		if (*at < '0' || *at > '9') return false;
		index = index * 10 + (*at - '0');
		if (index >= UINT32_MAX) return false;
	}
	*result = (size_t)index;
	return true;
}

bool adamic_has_property(const adamic_heap *value, const char *name) {
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
		if (array_key(name, &index)) return index < array->length;
		if (array->metadata != NULL && adamic_has_property(&array->metadata->heap, name)) return true;
		if (array->properties != NULL && adamic_has_property(&array->properties->heap, name)) return true;
		static const char *const names[] = {"length", "at", "concat", "copyWithin", "fill", "find", "findIndex", "findLast", "findLastIndex", "lastIndexOf", "pop", "push", "reverse", "shift", "unshift", "slice", "sort", "splice", "includes", "indexOf", "join", "keys", "entries", "values", "forEach", "filter", "flat", "flatMap", "map", "every", "some", "reduce", "reduceRight", "toReversed", "toSorted", "toSpliced", "with"};
		return object_prototype_name(name) || named(name, names, sizeof names / sizeof names[0]);
	}
	adamic_panic("in on an unsupported runtime value", sizeof "in on an unsupported runtime value" - 1);
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
		}
	}
	adamic_panic("dynamic read of a host scalar without type metadata", sizeof "dynamic read of a host scalar without type metadata" - 1);
}

adamic_heap *adamic_dynamic_property(adamic_heap *value, const char *name) {
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
		if (array->properties != NULL) return adamic_dynamic_property(&array->properties->heap, name);
		return NULL;
	}
	adamic_panic("dynamic property read on an unsupported runtime value", sizeof "dynamic property read on an unsupported runtime value" - 1);
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

static adamic_value prototype_method_called(adamic_closure *self, adamic_value *arguments) {
	(void)arguments;
	const char *name = "an unknown method";
	for (size_t index = 0; index < sizeof prototype_methods / sizeof prototype_methods[0]; index++) {
		if (prototype_methods[index] == self) name = adamic_dynamic_keyed_prototype_method_names[index];
	}
	static const char prefix[] = "Object.prototype.", suffix[] = " isn't called yet";
	size_t length = sizeof prefix - 1 + strlen(name) + sizeof suffix - 1;
	char *message = malloc(length);
	if (message == NULL) {
		static const char out_of_memory[] = "out of memory";
		adamic_panic(out_of_memory, sizeof out_of_memory - 1);
	}
	memcpy(message, prefix, sizeof prefix - 1);
	memcpy(message + sizeof prefix - 1, name, strlen(name));
	memcpy(message + sizeof prefix - 1 + strlen(name), suffix, sizeof suffix - 1);
	adamic_panic(message, length);
}

// program_shape reports whether the program made objects of this shape. In a program with a run-time
// key read the emitter registers each literal's shape as it makes one (emit_objects.go), and the
// runtime registers each shape whose object is a plain one on Node, Object's prototype and no other:
// each the JSON decoder builds (json_decode.c), and the results of decodeJson, readTextFile,
// writeTextFile, readDirectory and fileStatus, and a collection's or matchAll's next() (json_decode.c,
// input.c, directory.c, map_set.c, regexp.c). A host object's shape is never on the list.
static bool program_shape(const adamic_shape *shape) {
	for (const adamic_shape_types *entry = __atomic_load_n(&shape_types, __ATOMIC_ACQUIRE); entry != NULL; entry = entry->next) {
		if (entry->shape == shape) return true;
	}
	return false;
}

// host_object is the one predicate for a value whose prototype on Node isn't Object's, a program
// class's instance aside (#ey7aysy, Adamic's ruling on #0sps03r): an object with no class, not a
// record, whose shape has fields and isn't registered. A record, a class instance, a shape the
// program or the runtime registered and a zero-field shape are plain. The runtime's own objects whose
// prototype on Node is another's, a Stats, a Dirent, an iterator, a match's groups object (whose is
// null) or process.memoryUsage()'s, aren't registered, and so are host objects here. The JavaScript
// backend's adamicDynamicGet names the same values by their prototype (dynamic_keyed.go).
static bool host_object(const adamic_object *object) {
	return object->class == NULL && !adamic_record_is(object) && object->shape->count != 0 && !program_shape(object->shape);
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
		keyed_stop(adamic_dynamic_keyed_stop_array, key);
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
	if (object->tuple) keyed_stop(adamic_dynamic_keyed_stop_array, key);
	if (object->class != NULL && object->class->is_static) keyed_stop(adamic_dynamic_keyed_stop_static_class, key);
	if (key->length > 0 && key->bytes[0] == '#') keyed_stop(adamic_dynamic_keyed_stop_private, key);
	if (key_is(key, "__proto__")) keyed_stop(adamic_dynamic_keyed_stop_proto, key);
	bool error = adamic_error_is(object);
	if (error && key_named(key, adamic_dynamic_keyed_error_names, sizeof adamic_dynamic_keyed_error_names / sizeof adamic_dynamic_keyed_error_names[0])) {
		keyed_stop(adamic_dynamic_keyed_stop_error_property, key);
	}
	// A host object stops for every key, before its own slots are read: a slot may be internal (a
	// matchAll iterator's input and regex), its name may be Node's prototype member (a Stats' isFile),
	// or Node may hold a field native's shape doesn't (memoryUsage's rss). An Error is carved out: its
	// own fields answer as Node's do, and its prototype's names stop below.
	if (!error && host_object(object)) keyed_stop(adamic_dynamic_keyed_stop_runtime_value, key);
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
	// Only an Error reaches here as a host object. Its prototype is Error.prototype, whose own members
	// Object's closures would stand in for, so an Object.prototype name stops on it.
	if (inherited != NULL && host_object(object)) keyed_stop(adamic_dynamic_keyed_stop_runtime_value, key);
	return inherited;
}
