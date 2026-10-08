// adamic.h: the Adamic runtime. Small, plain C11, compiled with every program.

#ifndef ADAMIC_H
#define ADAMIC_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdatomic.h>
#include "count.h"
#include "tsan_test.h"
#include <string.h>

#include <math.h>

enum adamic_stream {
	adamic_stdout = 1,
	adamic_stderr = 2,
};

// adamic_heap begins every value on the heap: strings, objects and arrays (heap.c, docs/memory.md).
// references 0 is a constant the program spelled out, immortal; anything made at runtime starts at 1,
// and its last release frees it and releases what it holds.
enum adamic_kind {
	adamic_kind_string = 1,
	adamic_kind_object,
	adamic_kind_array,
	adamic_kind_map,
	adamic_kind_cell,
	adamic_kind_closure,
	adamic_kind_map_iterator,
	adamic_kind_number,
	adamic_kind_boolean,
	adamic_kind_weak,
	adamic_kind_async_frame,
	adamic_kind_async_promise,
	adamic_kind_async_reaction,
};

typedef struct adamic_heap {
	size_t references;
	enum adamic_kind kind;
	// slab is the number of the chunk the value's memory came from, plus one, or 0 for memory from
	// malloc and for a value never freed (heap.c). It fills what was padding.
	uint32_t slab;
} adamic_heap;

// The high count bit records permanent sharing; the remaining bits are the actual count.
// Unshared operations read and update the count plainly. Shared updates are atomic. A separate
// immutable slab bit dispatches before reading a count. Zero remains immortal, including regions.
#define ADAMIC_SHARED ((size_t)1 << (sizeof(size_t) * 8 - 1))
_Static_assert((intptr_t)ADAMIC_SHARED == INTPTR_MIN, "native count tags require two-complement intptr_t conversion");
#define ADAMIC_SHARED_HEADER UINT32_C(0x80000000)
#define ADAMIC_REGION_VALUE UINT32_C(0x40000000)
static inline bool adamic_is_shared(const adamic_heap *heap) {
	return (heap->slab & ADAMIC_SHARED_HEADER) != 0;
}
static inline size_t adamic_reference_count(const adamic_heap *heap) {
	return adamic_is_shared(heap) ? __atomic_load_n(&heap->references, __ATOMIC_ACQUIRE) & ~ADAMIC_SHARED : heap->references;
}
void adamic_share(void *value);
void adamic_heap_thread_end(void);
void adamic_heap_end(void);

// adamic_retain and adamic_release take any heap value. NULL (undefined) is left alone.
void *adamic_retain_slow(void *value);
void adamic_release_slow(void *value);

// Sharing is published before any other worker can reach a value and never cleared. Test that
// separate bit first: a plain read of a shared count would race with its atomic updates.
static inline void *adamic_retain(void *value) {
	ADAMIC_COUNT_RETAIN();
	adamic_heap *heap = value;
	if (heap != NULL && !adamic_is_shared(heap)) {
		size_t count = heap->references;
		if (count > 0) { ADAMIC_TSAN_PAUSE(adamic_tsan_plain_count); heap->references = count + 1; return value; }
	}
	return adamic_retain_slow(value);
}

static inline void adamic_release(void *value) {
	ADAMIC_COUNT_RELEASE();
	adamic_heap *heap = value;
	if (heap != NULL && !adamic_is_shared(heap)) {
		size_t count = heap->references;
		if (count > 1) { heap->references = count - 1; return; }
	}
	adamic_release_slow(value);
}

// adamic_allocate makes a heap value of size bytes, references 1, and panics when memory runs out.
void *adamic_allocate(size_t size, enum adamic_kind kind);

// adamic_value is a field or an element: which member is live is known statically, at every place
// that reads or writes one, and to the runtime through each object's shape and each array's flag.
typedef union adamic_value {
	double number;
	bool boolean;
	void *reference;
} adamic_value;

// adamic_maybe_number is number | undefined: present, and the number when it is.
typedef struct adamic_maybe_number {
	bool present;
	double number;
} adamic_maybe_number;

// adamic_maybe_boolean is boolean | undefined: present, and the boolean when it is.
typedef struct adamic_maybe_boolean {
	bool present;
	bool boolean;
} adamic_maybe_boolean;

// adamic_cell holds a variable a closure captured, so the function that declared it and every
// closure that captured it share one (closure.c). references says whether value is a reference.
typedef struct adamic_cell {
	adamic_heap heap;
	bool references;
	adamic_value value;
} adamic_cell;

adamic_cell *adamic_cell_new(adamic_value value, bool references);

// adamic_closure is a function value: its code, and the cells it captured. Every closure is called
// the same way, its arguments and its result as adamic_value, whatever its types.
typedef struct adamic_closure adamic_closure;
typedef adamic_value (*adamic_code)(adamic_closure *self, adamic_value *arguments);
struct adamic_closure {
	adamic_heap heap;
	adamic_code code;
	size_t count;
	adamic_cell *cells[];
};

// adamic_closure_new makes a closure of count cells, for the caller to fill with references it gives.
adamic_closure *adamic_closure_new(adamic_code code, size_t count);

// adamic_string is an immutable string: UTF-8 bytes (string.c).
typedef struct adamic_string {
	adamic_heap heap;
	size_t length;
	const char *bytes;
	// units is the length in UTF-16 units plus one, propagated when building, or 0 if unknown.
	// units == length + 1 is the ASCII flag: no unit counting or decoding is needed. index is
	// a long non-ASCII string's position index, once built (string_index.c). Both are caches, which
	// every initializer that leaves them out leaves empty.
	size_t units;
	struct adamic_string_index *index;
	// owner is the string whose bytes these are, when they aren't this string's own: a shared slice
	// holds a reference to it, let go when the slice is freed (string_share.c). It's NULL for a
	// string with bytes of its own, and for a slice of a constant, whose bytes never go away.
	struct adamic_string *owner;
	// capacity is how many bytes fit where this string's own bytes are, so an append can write in
	// place when nothing else holds the string (string_append.c). It's 0 for every string whose
	// bytes can't grow: a constant, a shared slice, one made on the stack.
	size_t capacity;
} adamic_string;

struct adamic_string_index {
	// Published with the view, never written after sharing.
	size_t units;
	// The last code point found: its first unit, and its byte offset.
	size_t cursor_unit;
	size_t cursor_offset;
	// checkpoints[k] is where unit k * STEP is: the byte offset of the code point holding it, shifted
	// left once, and 1 when that unit is the low half of a surrogate pair, whose code point starts a
	// unit earlier.
	size_t count;
	// Each UTF-16 unit, including both halves of supplementary points and lone surrogates.
	// The byte checkpoints still serve slices and searches.
	uint16_t *view;
	uint32_t checkpoints[];
};

// ADAMIC_LITERAL_INDEX marks a constant's index as not yet built: a constant lives as long as the
// program, so a long one can have an index that does too (string_index.c). Only a constant of static
// storage may carry it.
extern char adamic_literal_mark;
#define ADAMIC_LITERAL_INDEX ((struct adamic_string_index *)&adamic_literal_mark)

// ADAMIC_STRING is a constant: ADAMIC_STRING("text") as a static adamic_string's initializer.
#define ADAMIC_STRING(text) {{0, adamic_kind_string, 0}, sizeof text - 1, text, 0, ADAMIC_LITERAL_INDEX, NULL, 0}

// ADAMIC_STRING_BYTES is a constant too long for a C string literal: its bytes an array of size.
#define ADAMIC_STRING_BYTES(array, size) {{0, adamic_kind_string, 0}, size, array, 0, ADAMIC_LITERAL_INDEX, NULL, 0}

// adamic_shape is an object's layout: its fields' names in order, and which fields hold references.
//
// methods are, for the objects a class makes, the class's methods by name, and NULL for any other
// object: a call through an interface the class implements finds one there (adamic_object_callee).
typedef struct adamic_methods adamic_methods;
typedef struct adamic_shape {
	size_t count;
	const char *const *names;
	const bool *references;
	const adamic_methods *methods;
} adamic_shape;

typedef void (*adamic_virtual_method)(void);
typedef struct adamic_object adamic_object;
typedef struct adamic_accessor {
	const char *name;
	adamic_value (*get)(adamic_object *);
	void (*set)(adamic_object *, adamic_value, int);
	int type;
} adamic_accessor;
typedef struct adamic_class {
	const struct adamic_class *base;
	size_t own_start;
	size_t count;
	const adamic_virtual_method *methods;
	size_t definition;
	const adamic_shape *public_shape;
	const adamic_accessor *accessors;
	size_t accessor_count;
	bool is_static;
	size_t static_parent;
	const size_t *static_flags;
} adamic_class;

// adamic_object is a plain object (object.c). Its shape travels with it, so the same object can be
// seen through any type it satisfies, as JavaScript allows, without ever being copied.
typedef struct adamic_object {
	adamic_heap heap;
	const adamic_shape *shape;
	const adamic_class *class;
	bool frozen;
	bool sealed;
	bool nonextensible;
	// captureStackTrace owns one non-enumerable string outside the fixed shape.
	bool has_captured_stack;
	adamic_value captured_stack;
	adamic_value slots[];
} adamic_object;

void adamic_error_capture_stack(adamic_object *target);
adamic_string *adamic_error_read_stack(const adamic_object *target);

bool adamic_instanceof(const void *value, const adamic_class *wanted);
adamic_virtual_method adamic_virtual(const adamic_object *object, size_t slot);
void adamic_object_free_children(adamic_object *object, void (*release)(void *));

// adamic_slot_cache remembers, at one place in the program that reads a field, where the field was in
// the last shape seen there.
typedef struct adamic_slot_cache {
	uint64_t packed;
} adamic_slot_cache;
#define ADAMIC_SLOT_SHAPE_MASK UINT64_C(0x0000ffffffffffff)
_Static_assert(sizeof(uintptr_t) <= sizeof(uint64_t), "slot cache requires pointers of at most 64 bits");
_Static_assert(__atomic_always_lock_free(sizeof(uint64_t), 0), "slot cache requires lock-free 64-bit atomics");
// One relaxed word contains both the immutable shape identity and its slot. Large
// slot indices are simply not cached; shape addresses outside 48 bits are refused.
void adamic_slot_cache_store(adamic_slot_cache *cache, const adamic_shape *shape, size_t index);

// adamic_method is a class's method as a call through an interface calls it: the object as this, and
// the arguments and the result as adamic_value, as a closure's are (the result owned).
typedef adamic_value (*adamic_method)(adamic_object *self, adamic_value *arguments);
struct adamic_methods {
	size_t count;
	const char *const *names;
	const adamic_method *code;
};

// adamic_object_callee finds what object.name(...) calls, where the object is seen through an
// interface: its own field of that name, a function value, which it returns; or else its class's
// method of that name, which it puts in *method, returning NULL. The checker proved one is there.
adamic_closure *adamic_object_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_method *method);

// adamic_object_new makes an object of a shape, its fields zeroed for the caller to fill; a reference
// stored in a field belongs to the object.
adamic_object *adamic_object_new(const adamic_shape *shape);

// adamic_region is a region (region.c): the objects one statement makes that nothing reaches after
// it, let go of together when it ends. ADAMIC_REGION is an empty one. adamic_object_new_in makes an
// object in a region (on the heap, as adamic_object_new does, for a NULL region), and
// adamic_region_end lets go of the region's objects and what they hold.
typedef struct adamic_region_block adamic_region_block;
typedef struct adamic_region {
	adamic_region_block *blocks;
	size_t count;
} adamic_region;
#define ADAMIC_REGION {NULL, 0}
adamic_object *adamic_object_new_in(adamic_region *region, const adamic_shape *shape);
void adamic_region_end(adamic_region *region);

// adamic_object_copy is { ...source }: the same shape, its references retained.
adamic_object *adamic_object_copy(const adamic_object *source);

// adamic_object_has is object.hasOwnProperty(name).
bool adamic_object_has(const adamic_object *object, const adamic_string *name);

// Runtime scalar tags for fixed program shapes. Host reference slots already carry heap tags.
typedef struct adamic_shape_types {
	const adamic_shape *shape;
	const int *types;
	struct adamic_shape_types *next;
} adamic_shape_types;
extern adamic_heap adamic_null;
void adamic_register_shape_types(adamic_shape_types *metadata);
bool adamic_has_property(const adamic_heap *object, const char *name);
adamic_heap *adamic_dynamic_property(adamic_heap *object, const char *name);

// adamic_object_field finds a field by name. The checker proved the field is there. Where this place in
// the program last saw the same shape, the field is where it was then, which is inline, since it's
// what nearly every read is; anything else is adamic_object_find, which searches the shape's names.
adamic_value *adamic_object_find(const adamic_object *object, const char *name, adamic_slot_cache *cache);
adamic_value *adamic_static_field(const adamic_object *object, const char *name, adamic_slot_cache *cache);
adamic_value *adamic_object_write_field(adamic_object *object, const char *name, adamic_slot_cache *cache);
// A readonly numeric view may see a field made with the undefined-only reference representation.
adamic_maybe_number adamic_object_maybe_number(const adamic_object *object, const char *name, adamic_slot_cache *cache);
// Optional own fields may be absent; NULL then asks the reader to produce typed undefined.
adamic_value *adamic_object_optional_find(const adamic_object *object, const char *name, adamic_slot_cache *cache);
static inline adamic_value *adamic_object_optional_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	if (object->has_captured_stack && strcmp(name, "stack") == 0) { return &((adamic_object *)object)->captured_stack; }
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)object->shape) {
		return adamic_object_optional_find(object, name, cache);
	}
	size_t slot = packed >> 48;
	if (slot == object->shape->count) { return NULL; }
	return &((adamic_object *)object)->slots[slot];
}
// Whole-program data accesses exclude inherited constructor storage, so a cache
// hit needs only the shape comparison, without loading a class descriptor.
static inline adamic_value *adamic_object_data_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	if (object->has_captured_stack && strcmp(name, "stack") == 0) { return &((adamic_object *)object)->captured_stack; }
	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
	if ((packed & ADAMIC_SLOT_SHAPE_MASK) == (uintptr_t)object->shape) {
		return &((adamic_object *)object)->slots[packed >> 48];
	}
	return adamic_object_find(object, name, cache);
}

static inline adamic_value *adamic_object_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {
	if (object->has_captured_stack && strcmp(name, "stack") == 0) { return &((adamic_object *)object)->captured_stack; }
	if (object->class != NULL && object->class->is_static) { return adamic_static_field(object, name, cache); }
	return adamic_object_data_field(object, name, cache);
}
#include "object_integrity.h"

// Static Object methods (library_object.c). Returned collections and freeze own one reference.
bool adamic_object_is(const adamic_heap *left, const adamic_heap *right);
bool adamic_object_is_frozen(const adamic_heap *value);
adamic_object *adamic_object_freeze(adamic_object *object);
bool adamic_object_has_own(const adamic_object *object, const adamic_string *key);
struct adamic_array *adamic_object_keys(const adamic_object *object);
struct adamic_array *adamic_object_values(const adamic_object *object, bool references, bool entries);
void adamic_object_assign(adamic_object *target, const adamic_object *source);
void adamic_object_check_write(const adamic_object *object, const char *name);
// Keep frozen-object failures out of the ordinary write's call path.
static inline void adamic_object_check_data_write(const adamic_object *object, const char *name) {
	if (object->frozen) {
		adamic_object_check_write(object, name);
	}
}

// adamic_array is an array (array.c). references says whether its elements are references.
typedef struct adamic_array {
	adamic_heap heap;
	size_t length;
	size_t capacity;
	bool references;
	adamic_value *elements;
	// Extra fields of RegExp result arrays, owned and released with the array.
	adamic_object *properties;
} adamic_array;

adamic_array *adamic_array_new(size_t capacity, bool references);
size_t adamic_public_index(const adamic_shape *shape, size_t position);
adamic_array *adamic_class_object_keys(const adamic_object *object);
const adamic_accessor *adamic_accessor_find(const adamic_object *object, const char *name);
adamic_value adamic_accessor_get(adamic_object *object, const char *name);
void adamic_accessor_set(adamic_object *object, const char *name, adamic_value value, int type);

// Structured fork-join. Arguments are borrowed until the join; the result is owned.
adamic_array *adamic_parallel_map(adamic_array *items, adamic_closure *work, bool references);

// adamic_array_push appends; a reference pushed belongs to the array.
void adamic_array_push(adamic_array *array, adamic_value value);

// adamic_map stores up to four ordered entries inline, then uses a hash table (map.c).
typedef struct adamic_map_entry {
	adamic_value key;
	adamic_value value;
	bool deleted;
} adamic_map_entry;

typedef struct adamic_map {
	adamic_heap heap;
	size_t count;
	size_t used;
	size_t capacity;
	adamic_map_entry *entries;
	size_t bucket_count;
	size_t *buckets;
	bool string_keys;
	// reference_keys is keys that are counted references: strings, compared by their text, or, when
	// string_keys isn't set, objects, arrays, maps and functions, compared by identity, as === does.
	bool reference_keys;
	// boolean_keys is keys that are booleans, compared and hashed as booleans: a boolean set in an
	// adamic_value leaves the rest of its bytes unspecified, so it's never read as a number.
	bool boolean_keys;
	// maybe_number_keys are packed number | undefined keys. The reserved undefined NaN is distinct
	// from the canonical present NaN, while present numbers use SameValueZero (map.c).
	bool maybe_number_keys;
	bool reference_values;
	// iterating counts the iterations open over the map; while there are any, its entries keep their
	// places (map.c).
	_Atomic size_t iterating;
	// entries points here until a fifth historical slot is needed. An open iterator
	// can require a table even with fewer live keys; closing the last permits compaction.
	adamic_map_entry small[4];
} adamic_map;

// adamic_map_iterator is one for...of over a map, in insertion order: entries added before it gets
// to them are visited, and entries deleted before it gets to them aren't, as ECMA-262 requires. It
// holds the map; exhaustion or letting go ends the iteration exactly once.
typedef struct adamic_map_iterator {
	adamic_heap heap;
	adamic_map *map;
	size_t next;
	bool exhausted;
} adamic_map_iterator;

adamic_map_iterator *adamic_map_iterate(adamic_map *map);
// Ends an active iteration once and returns a small map to inline storage when safe.
void adamic_map_iterator_close(adamic_map_iterator *iterator);
adamic_object *adamic_collection_iterator(void *collection, int part, int key, int value, bool set);

// adamic_map_iterator_next gives the next live entry's key and value, borrowed, or false at the end.
bool adamic_map_iterator_next(adamic_map_iterator *iterator, union adamic_value *key, union adamic_value *value);

adamic_map *adamic_map_new(bool string_keys, bool reference_values);

// adamic_map_new_identity is a map whose keys are objects, arrays, maps or functions, each its own
// key, found by identity (map.c).
adamic_map *adamic_map_new_identity(bool reference_values);

// adamic_map_new_booleans is a map whose keys are booleans: true and false, two keys at most.
adamic_map *adamic_map_new_booleans(bool reference_values);
bool adamic_map_maybe_key_equal(double left, double right);
uint64_t adamic_map_maybe_key_hash(double key);
uint64_t adamic_map_number_hash(double number);

// adamic_map_new_maybe_numbers makes a map whose keys are packed number | undefined values.
adamic_map *adamic_map_new_maybe_numbers(bool reference_values);

// adamic_map_clear is map.clear() and set.clear(): every entry deleted, as if one by one, so an
// iteration open over it goes on with whatever is added after.
void adamic_map_clear(adamic_map *map);

// adamic_map_keys and adamic_map_values are [...map.keys()] and [...map.values()]: new arrays the
// caller owns, in insertion order.
adamic_array *adamic_map_keys(const adamic_map *map);

// adamic_map_add_pairs sets each [key, value] tuple of an array in a map, in order: new Map(pairs).
// The map takes its own references.
void adamic_map_add_pairs(adamic_map *map, const adamic_array *pairs);
adamic_array *adamic_map_values(const adamic_map *map);

// adamic_map_get is the value's slot, or NULL when the key isn't there.
adamic_value *adamic_map_get(const adamic_map *map, adamic_value key);

// adamic_map_set takes the key and value it's given: references passed in belong to the map.
void adamic_map_set(adamic_map *map, adamic_value key, adamic_value value);

bool adamic_map_delete(adamic_map *map, adamic_value key);

// Records own a string-keyed Map through a fixed-shape wrapper (record.c). These aliases
// use ordinary object cleanup; the compiler must use record operations for record views.
typedef adamic_object adamic_record;
typedef adamic_object adamic_record_iterator;

adamic_record *adamic_record_new(bool reference_values);
// Own lookup returns a borrowed slot, NULL for absence (including an inherited name).
adamic_value *adamic_record_get_own(const adamic_record *record, const adamic_string *key);
bool adamic_record_has_own(const adamic_record *record, const adamic_string *key);
// Dynamic get and in hold own keys only: on a miss naming an Object.prototype member,
// both panic with the member name and "records hold own keys only". Other misses return
// NULL/false. The member-name check runs only on a miss; an own value is always borrowed.
adamic_value *adamic_record_get(const adamic_record *record, const adamic_string *key);
bool adamic_record_has(const adamic_record *record, const adamic_string *key);
// Both writes consume key and value references, as Map.set does. define creates an own data
// property even for __proto__; set refuses __proto__ assignment with an explicit NotYet panic.
void adamic_record_define(adamic_record *record, adamic_string *key, adamic_value value);
void adamic_record_set(adamic_record *record, adamic_string *key, adamic_value value);
bool adamic_record_delete(adamic_record *record, const adamic_string *key);
size_t adamic_record_size(const adamic_record *record);
// keys returns an owned array in own-key order: array indices ascending, then insertion order.
adamic_array *adamic_record_keys(const adamic_record *record);
// Iteration snapshots keys, holds record and keys, skips deleted keys, and reads current values.
// New keys are not visited. next returns borrowed key/value pairs; release ends iteration.
adamic_record_iterator *adamic_record_iterate(adamic_record *record);
bool adamic_record_iterator_next(adamic_record_iterator *iterator, adamic_string **key, adamic_value *value);

// A Set is a map whose values aren't used (set.c). adamic_set_add_all adds an array's elements in
// order, new Set(array), each reference retained; adamic_set_values is [...set], a new array the caller
// owns.
void adamic_set_add_all(adamic_map *set, const adamic_array *values);
adamic_array *adamic_set_values(const adamic_map *set);

// adamic_map_free_children lets go of what a map holds, for the heap's freeing.
void adamic_map_free_children(adamic_map *map, void (*let_go)(void *));

// adamic_array_at is array[index]: the element's slot, or NULL (undefined) when there isn't one.
// It's inline, since a loop over an array reads through it every pass. An array index is an integer
// from 0 up to the length; anything else (negative, a fraction, NaN, past the end) is a property the
// array doesn't have, which reads as undefined.
//
// Past the bounds check, 0 <= index < length, so (size_t)index is defined and, back as a double, is
// trunc(index) exactly (below 2^53 the conversion is exact, and from there every double is whole), so
// a fraction is caught without calling trunc, which on x86-64 without SSE4.1 is a call into libm on
// every read. -0 is index 0, as JavaScript reads it.
// adamic_array_at_integer is array[index] for an index that's a whole number already, a loop counter
// kept in an integer (lower/counters.go): the same answer, with only the bounds to check.
static inline adamic_value *adamic_array_at_integer(const adamic_array *array, int64_t index) {
	if (index < 0 || (uint64_t)index >= array->length) {
		return NULL;
	}
	return &array->elements[index];
}

static inline adamic_value *adamic_array_at(const adamic_array *array, double index) {
	if (!(index >= 0) || index >= (double)array->length) {
		return NULL;
	}
	size_t whole = (size_t)index;
	if ((double)whole != index) {
		return NULL;
	}
	return &array->elements[whole];
}

// adamic_array_set is array[index] = value, which takes the value; it panics at an index the array
// doesn't have.
void adamic_array_set(adamic_array *array, double index, adamic_value value);

// adamic_array_at_relative is array.at(index): the element's slot, or NULL (undefined).
adamic_value *adamic_array_at_relative(const adamic_array *array, double index);

// adamic_array_index_of is array.indexOf(value), and array.includes(value) with same_value_zero:
// the first index whose element equals value, or -1. equality says how elements compare.
enum adamic_equality {
	adamic_equal_numbers = 1,
	adamic_equal_booleans,
	adamic_equal_strings,
	adamic_equal_identity,
	adamic_equal_maybe_numbers,
};
double adamic_array_index_of(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool same_value_zero);
double adamic_array_search_from(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool same_value_zero, double from, bool has_from, bool last);

// adamic_array_from_length is the length Array.from({ length }) makes (array_from.c): ToLength of
// the number, and a panic where JavaScript throws, past 2^32 - 1.
size_t adamic_array_from_length(double length);

// adamic_array_reverse reverses in place and is the array; adamic_array_concat makes a new array of
// every array's elements in order, references retained.
adamic_array *adamic_array_reverse(adamic_array *array);

// adamic_array_filled is new Array(length).fill(value), a new array the caller owns;
// adamic_array_fill is array.fill(value, start, end), in place, and is the array. Both hold the value
// once per element; the caller keeps its own.
adamic_array *adamic_array_filled(double length, adamic_value value, bool references);
adamic_array *adamic_array_fill(adamic_array *array, adamic_value value, double start, double end, bool has_start, bool has_end);

// adamic_array_splice is array.splice(start, count, ...items): what's removed, in a new array the
// caller owns. The items' references are the array's from then on.
adamic_array *adamic_array_splice(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items);
// adamic_array_remove is a splice whose result nothing uses: what it removes is let go of, and no
// array is made to hold it.
void adamic_array_remove(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items);

// adamic_array_append is a spread, [...source], into an array being made: each element pushed, a
// reference retained.
void adamic_array_append(adamic_array *array, const adamic_array *source);
adamic_array *adamic_array_concat(size_t count, adamic_array *const arrays[]);

// adamic_array_slice is array.slice(start, end); adamic_array_sort sorts in place, stably, by a
// comparator that gives -1, 0 or 1 and is passed context. adamic_compare_closure is that comparator for
// a function value, the context.
adamic_array *adamic_array_slice(const adamic_array *array, double start, double end, bool has_end);
void adamic_array_sort(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context);
int adamic_compare_closure(adamic_value left, adamic_value right, void *context);
// adamic_timsort sorts count values in place by V8's algorithm (sort.c). It says false, the values
// in some order of the sort's half done, when the comparator threw (adamic_thrown set).
bool adamic_timsort(adamic_value *work, size_t count, int (*compare)(adamic_value, adamic_value, void *), void *context);

// adamic_array_sort_undefined_last sorts an array of number | undefined as JavaScript does: every
// undefined goes to the end, never passed to the comparator (sort_undefined.c).
void adamic_array_sort_undefined_last(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context);

// adamic_map_entries is [...map]: [key, value] pairs, each an object of the shape given.
adamic_array *adamic_map_entries(const adamic_map *map, const adamic_shape *pair);

// adamic_array_join is array.join(separator), each element written as String() would.
enum adamic_join {
	adamic_join_numbers = 1,
	adamic_join_booleans,
	adamic_join_strings,
	adamic_join_maybe_numbers,
};
struct adamic_string *adamic_array_join(const adamic_array *array, const struct adamic_string *separator, enum adamic_join kind);

struct adamic_string *adamic_array_join_nested(const adamic_array *array, const struct adamic_string *separator, enum adamic_join kind, size_t depth);

// These return a reference the caller owns.
adamic_string *adamic_string_from_number(double value);

// Number.parseInt(text, radix) and Number.parseFloat(text), as Node answers them (parse.c). A radix
// of 0 is the one left out.
double adamic_number_parse_int(const adamic_string *text, double radix);
double adamic_number_parse_float(const adamic_string *text);
double adamic_number_from_string(const adamic_string *text);
double adamic_number_from_union(const adamic_heap *value);
bool adamic_number_has_own_property(const adamic_string *key, bool prototype);
double adamic_math_clz32(double value);
double adamic_math_imul(double left, double right);
double adamic_math_fround(double value);
adamic_string *adamic_string_concat(size_t count, adamic_string *const parts[]);

// The strings every program has: "", and String(true) and String(false).
extern adamic_string adamic_string_empty;
extern adamic_string adamic_string_true;
extern adamic_string adamic_string_false;

// String(undefined), and String(value) for number | undefined, a string the caller owns (maybe.c).
// adamic_maybe_number_equal and adamic_maybe_boolean_equal are === on two pairs: both missing, or
// both present and equal.
extern adamic_string adamic_string_undefined;
adamic_string *adamic_string_from_maybe_number(adamic_maybe_number value);
bool adamic_maybe_number_equal(adamic_maybe_number left, adamic_maybe_number right);

// adamic_maybe_number_pack makes number | undefined one double, for a field, an element, a cell or a
// function value's argument or result: undefined is ADAMIC_UNDEFINED_BITS, a NaN no arithmetic
// makes, and every other NaN is stored as the ordinary quiet NaN (JavaScript can't see a NaN's
// payload), so nothing present is ever mistaken for undefined. adamic_maybe_number_unpack reads one
// back.
#define ADAMIC_UNDEFINED_BITS 0x7ff8000000000001u
double adamic_maybe_number_pack(adamic_maybe_number value);
adamic_maybe_number adamic_maybe_number_unpack(double packed);
bool adamic_maybe_boolean_equal(adamic_maybe_boolean left, adamic_maybe_boolean right);


// A string's UTF-16 view (string.c): length, charCodeAt and trim as JavaScript means them.
//
// length reads the propagated or counted units inline. charCodeAt reads ASCII bytes or a built
// UTF-16 view inline. An unknown length, an unbuilt index, NaN, a negative or past the end goes
// to adamic_string_char_code. A position in range truncates to its index as (size_t) does.
size_t adamic_string_units(const adamic_string *string);
double adamic_string_char_code(const adamic_string *string, double position);
// A shared string publishes length and position metadata together in its index.
// Keep warm length and ASCII reads inline, without writing its plain units field.
static inline size_t adamic_string_known_units(const adamic_string *string) {
	if (string->units != 0) { return string->units; }
	if (adamic_is_shared(&string->heap)) {
		struct adamic_string_index *index = __atomic_load_n(&string->index, __ATOMIC_ACQUIRE);
		if (index != NULL && index != ADAMIC_LITERAL_INDEX) { return index->units + 1; }
	}
	return 0;
}
#include "string_wellformed.h"

static inline double adamic_string_length(const adamic_string *string) {
	size_t units = adamic_string_known_units(string);
	return units != 0 ? (double)(units - 1) : (double)adamic_string_units(string);
}
static inline double adamic_string_char_code_at(const adamic_string *string, double position) {
	if (string->units == string->length + 1 && position >= 0 && position < (double)string->length) {
		return (double)(unsigned char)string->bytes[(size_t)position];
	}
	// Shared views are published complete; an ordinary index is thread-local.
	struct adamic_string_index *index = adamic_is_shared(&string->heap) ?
		__atomic_load_n(&string->index, __ATOMIC_ACQUIRE) : string->index;
	if (index != NULL && index != ADAMIC_LITERAL_INDEX && position >= 0 && position < (double)index->units) {
		if (index->units == string->length) { return (double)(unsigned char)string->bytes[(size_t)position]; }
		if (index->view != NULL) { return (double)index->view[(size_t)position]; }
	}
	return adamic_string_char_code(string, position);
}

adamic_string *adamic_string_trim(adamic_string *string);

// for...of over a string walks code points: adamic_string_next is the byte size of the one at offset,
// and adamic_string_slice_bytes makes it a string the caller owns.
size_t adamic_string_next(const adamic_string *string, size_t offset);
adamic_string *adamic_string_slice_bytes(const adamic_string *string, size_t offset, size_t size);

// adamic_string_share is the size bytes at offset as a string the caller owns, reading the string's
// own bytes where that's worth it (string_share.c). offset and size must fall between code points.
adamic_string *adamic_string_share(const adamic_string *string, size_t offset, size_t size);

// The rest of a string's UTF-16 view (string.c), each as JavaScript means it. These that return a
// string return one the caller owns.
adamic_string *adamic_string_slice(const adamic_string *string, double start, double end, bool has_end);
adamic_maybe_number adamic_string_code_point_at(const adamic_string *string, double position);
int adamic_string_compare(const adamic_string *left, const adamic_string *right);
adamic_string *adamic_string_repeat(const adamic_string *string, double count);
adamic_string *adamic_string_pad(const adamic_string *string, double target, const adamic_string *fill, bool at_start);
double adamic_string_index_of(const adamic_string *string, const adamic_string *search);
// adamic_string_index_of_at is indexOf from the UTF-16 index from, at most the string's length, in
// place (string.c). adamic_string_index_of_from is indexOf with its position as JavaScript gives it
// (string_from.c).
double adamic_string_index_of_at(const adamic_string *string, const adamic_string *search, size_t from);
double adamic_string_index_of_from(const adamic_string *string, const adamic_string *search, double position);
bool adamic_string_starts_with(const adamic_string *string, const adamic_string *search);
bool adamic_string_ends_with(const adamic_string *string, const adamic_string *search);
struct adamic_array *adamic_string_code_points(const adamic_string *string);
struct adamic_array *adamic_string_split(const adamic_string *string, const adamic_string *separator);

// adamic_string_at is string[index]: the UTF-16 code unit there, as a string the caller owns, or NULL
// (undefined) where the string has no such index. adamic_string_at_relative is string.at(index).
adamic_string *adamic_string_at(const adamic_string *string, double index);
adamic_string *adamic_string_at_relative(const adamic_string *string, double index);

// trimStart and trimEnd; trim is both.
adamic_string *adamic_string_trim_sides(adamic_string *string, bool at_start, bool at_end);

// lastIndexOf, and replace and replaceAll with a string pattern and a replacement string (its $$, $&,
// $` and $' expanded, as JavaScript does).
double adamic_string_last_index_of(const adamic_string *string, const adamic_string *search);
adamic_string *adamic_string_replace(const adamic_string *string, const adamic_string *search, const adamic_string *replacement, bool all);

// ADAMIC_STRING_MAX_UNITS is V8's longest string, in UTF-16 units (String::kMaxLength on 64-bit).
#define ADAMIC_STRING_MAX_UNITS 536870888

// adamic_string_check_length panics, as V8 throws RangeError: Invalid string length, when a string
// would be longer than V8's longest, in UTF-16 units.
void adamic_string_check_length(double units);

// adamic_string_allocate makes a string of length bytes for the caller to fill, references 1.
adamic_string *adamic_string_allocate(size_t length);

// toUpperCase and toLowerCase (case.c): Unicode's full, locale-independent case mapping, final
// sigma included, as Node does it. They return a string the caller owns.
adamic_string *adamic_string_to_upper(const adamic_string *string);
adamic_string *adamic_string_to_lower(const adamic_string *string);

// normalize (normalize.c): NFC, NFD, NFKC or NFKD as form names it, and a panic, as JavaScript's
// RangeError, for any other form. It returns a string the caller owns.
adamic_string *adamic_string_normalize(const adamic_string *string, const adamic_string *form);

// adamic_string_units is a string's length in UTF-16 units, counted once. adamic_string_locate is
// where a unit below that length is: the byte offset of the code point holding it, and whether the unit
// is the low half of a surrogate pair there. Both take constant time amortized (string_index.c).
// adamic_string_join_halves makes each lone high surrogate followed by a lone low one in bytes, from
// offset from on, the character they are together, as JavaScript joins them, and returns the new
// length (string.c).
size_t adamic_string_join_halves(char *bytes, size_t from, size_t length);

// adamic_string_put writes part's bytes after the first written of bytes, joining halves of a pair
// where they meet, and returns how many bytes there are now (string.c). Every string's own halves are
// joined already, so only the meeting is looked at.
size_t adamic_string_put(char *bytes, size_t written, const adamic_string *part);

// adamic_string_append is string + parts, taking the caller's reference to string: written in place
// when the caller held the only reference and there's room, and otherwise a new string with room to
// grow, string let go (string_append.c).
adamic_string *adamic_string_append(adamic_string *string, size_t count, adamic_string *const parts[]);

size_t adamic_string_units(const adamic_string *string);
size_t adamic_string_locate(const adamic_string *string, size_t unit, bool *low);

// adamic_string_units_before is how many UTF-16 units come before a byte offset that starts a code
// point: indexOf's answer, found through the index rather than by counting from the start.
size_t adamic_string_units_before(const adamic_string *string, size_t offset);
void adamic_string_free_index(adamic_string *string);
void adamic_string_prepare_shared(adamic_string *string);
// Borrowed until the string is freed or appended to; NULL for ownerless stack pieces.
const uint16_t *adamic_string_utf16_view(adamic_string *string);

// adamic_string_equal is ===.
int adamic_string_equal(const adamic_string *left, const adamic_string *right);

// A union whose members are held differently (string | number) is one counted reference, its kind the
// member it is (union.c): a string, object, array, map or closure is itself, undefined is NULL, a
// number is boxed, and a boolean is one of two constant boxes.
typedef struct adamic_number_box {
	adamic_heap heap;
	double number;
} adamic_number_box;

typedef struct adamic_boolean_box {
	adamic_heap heap;
	bool boolean;
} adamic_boolean_box;

extern adamic_boolean_box adamic_box_true;
extern adamic_boolean_box adamic_box_false;

// adamic_box_number boxes a number, a reference the caller owns.
adamic_heap *adamic_box_number(double number);

// adamic_union_equal is === on two unions: the same member, equal as that member is compared.
bool adamic_union_equal(const adamic_heap *left, const adamic_heap *right);

// adamic_union_to_string is String(value) for a union of numbers, booleans, strings and undefined, a
// string the caller owns.
adamic_string *adamic_union_to_string(adamic_heap *value);

// adamic_union_typeof classifies every reference, including static constructors. null says what
// a missing pointer represents; null and undefined share NULL but have different typeof results.
adamic_string *adamic_union_typeof(const adamic_heap *value, bool null);
extern adamic_string adamic_typeof_number;
extern adamic_string adamic_typeof_string;
extern adamic_string adamic_typeof_boolean;
extern adamic_string adamic_typeof_undefined;
extern adamic_string adamic_typeof_object;
extern adamic_string adamic_typeof_function;

// adamic_weak is the handle a Weak<Target> slot holds (weak.c): counted itself, it doesn't count its
// target, and it says NULL once the target is freed. adamic_weak_of is the target's handle, retained
// for the caller (NULL for NULL); adamic_weak_target is what a handle points to, and
// adamic_weak_target_present the same where the checker proved it present, panicking if it was freed.
// The heap tells weak.c when a target is freed (adamic_weak_forget) and when a handle is
// (adamic_weak_dropped).
typedef struct adamic_weak adamic_weak;
adamic_weak *adamic_weak_of(void *target);
void *adamic_weak_target(const adamic_weak *handle);
void *adamic_weak_target_present(const adamic_weak *handle);
void adamic_weak_forget(void *target);
void adamic_weak_dropped(adamic_weak *handle);
// adamic_weak_held reports whether a Weak points at a value: reuse in place takes over only a value
// nothing else can reach, and a Weak reaches without counting.
bool adamic_weak_held(const void *target);

// adamic_thrown is the error being thrown, or NULL (exceptions.c): set by a throw, tested after
// every call that can throw, and taken by the catch that lands it. adamic_error_new is new
// Error(message), and adamic_uncaught the panic of an error nothing caught.
extern _Thread_local adamic_object *adamic_thrown;
adamic_object *adamic_error_new(adamic_string *message);
// Host errors have three owning slots (name, message, code) and built-in nominal identity.
void adamic_error_tag(adamic_object *error);
_Noreturn void adamic_uncaught(void);

// adamic_start begins every program: it keeps main's arguments, and writes to a closed pipe fail
// rather than kill, as on Node.
void adamic_start(int count, char **values);

// adamic_output_flush writes out what stdout's buffer holds, before a file is read or written: the
// file may be stdout itself, or stdin waiting on a prompt just printed (adamic.c).
void adamic_output_flush(void);

// adamic_write_line writes a string and a newline, as console.log does with one string. Stdout is
// buffered, and flushed wherever Node's writing it at once could be told apart (adamic.c).
void adamic_write_line(enum adamic_stream stream, const adamic_string *string);

// ADAMIC_NUMBER_FORMAT_MAX holds the longest number text, "-1.2345678901234567e-308", with room.
#define ADAMIC_NUMBER_FORMAT_MAX 32

// adamic_number_format writes value as JavaScript's String(value) does and returns the length. The
// buffer is not terminated.
size_t adamic_number_format(double value, char buffer[ADAMIC_NUMBER_FORMAT_MAX]);

// adamic_power is JavaScript's ** (number.c).
double adamic_power(double base, double exponent);

// JavaScript's Math functions where C's differ: round rounds half up and keeps -0, sign keeps -0 and
// NaN, and max and min see NaN and tell +0 from -0 (math.c).
double adamic_math_round(double value);
double adamic_math_sign(double value);
double adamic_math_max(double left, double right);
double adamic_math_min(double left, double right);

// The rest of JavaScript's Math, V8's port of fdlibm bit for bit (ieee754.c), and Math.hypot, V8's
// MathHypot builtin (hypot.c), over count values.
double adamic_math_acos(double x);
double adamic_math_acosh(double x);
double adamic_math_asin(double x);
double adamic_math_asinh(double x);
double adamic_math_atan(double x);
double adamic_math_atan2(double y, double x);
double adamic_math_atanh(double x);
double adamic_math_cbrt(double x);
double adamic_math_cos(double x);
double adamic_math_cosh(double x);
double adamic_math_exp(double x);
double adamic_math_expm1(double x);
double adamic_math_log(double x);
double adamic_math_log1p(double x);
double adamic_math_log2(double x);
double adamic_math_log10(double x);
double adamic_math_sin(double x);
double adamic_math_sinh(double x);
double adamic_math_tan(double x);
double adamic_math_tanh(double x);
double adamic_math_hypot(size_t count, const double *values);

// adamic_number_shortest_digits is V8's shortest digits for a positive, finite value (dtoa.c), as
// Number::toString writes them: value is 0.d1d2d3... times 10^point, and it returns how many.
int adamic_number_shortest_digits(double value, char digits[18], int *point);

// adamic_number_to_exponential is value.toExponential(digits), and the shortest digits when there
// are none; adamic_number_to_precision is value.toPrecision(digits), and String(value) when there
// are none. Both are V8's (dtoa.c), and return a string the caller owns.
adamic_string *adamic_number_to_exponential(double value, double digits, bool has_digits);
adamic_string *adamic_number_to_precision(double value, double digits, bool has_digits);

// Math.max, Math.min, Math.hypot, String.fromCharCode and String.fromCodePoint over arguments with a
// spread among them, already evaluated in order into an array of numbers (spread.c).
double adamic_math_max_of(const adamic_array *values);
double adamic_math_min_of(const adamic_array *values);
double adamic_math_hypot_of(const adamic_array *values);
adamic_string *adamic_string_from_char_codes_of(const adamic_array *values);
adamic_string *adamic_string_from_code_points_of(const adamic_array *values);

// utf8Length(text) and utf8At(text, index) from 'adamic': the text's UTF-8, read in place, a lone
// surrogate as U+FFFD's bytes as the WHATWG encoder writes it (utf8.c). utf8At panics where the index
// isn't one of the text's bytes.
double adamic_utf8_length(const adamic_string *text);
double adamic_utf8_at(const adamic_string *text, double index);

// JavaScript's bitwise operators on numbers, each through ToInt32 or ToUint32 exactly (bitwise.c).
double adamic_bitwise_and(double left, double right);
double adamic_bitwise_or(double left, double right);
double adamic_bitwise_xor(double left, double right);
double adamic_bitwise_not(double value);
double adamic_shift_left(double left, double right);
double adamic_shift_right(double left, double right);
double adamic_shift_right_unsigned(double left, double right);

// String.fromCharCode and String.fromCodePoint over their arguments, already evaluated in order
// (from_codes.c): each a string the caller owns. fromCodePoint panics with V8's RangeError where a
// value isn't a code point.
adamic_string *adamic_string_from_char_codes(size_t count, const double values[]);
adamic_string *adamic_string_from_code_points(size_t count, const double values[]);

// adamic_number_to_radix is value.toString(radix), V8's (radix.c), and returns a string the caller
// owns.
adamic_string *adamic_number_to_radix(double value, double radix);

// adamic_number_to_fixed is value.toFixed(digits), and returns a string the caller owns.
adamic_string *adamic_number_to_fixed(double value, double digits);

// Input, opened in 0.2 (input.c). adamic_arguments_save keeps main's argc and argv;
// adamic_program_arguments is programArguments(), a new array of the arguments after the program, and
// adamic_read_text_file is readTextFile(path), { kind: 'Ok', text } or { kind: 'Error', message }. Both
// return a reference the caller owns.
void adamic_arguments_save(int count, char **values);
adamic_array *adamic_program_arguments(void);
adamic_object *adamic_read_text_file(const adamic_string *path);

// adamic_decode_utf8 makes a string of bytes decoded as Node decodes them, WHATWG UTF-8 with U+FFFD
// for each invalid sequence, and adamic_path_bytes is a path as the terminated bytes Node names a file
// by, a lone surrogate as U+FFFD, or NULL when it holds a NUL, which names no file; the caller frees it
// (input.c).
adamic_string *adamic_decode_utf8(const unsigned char *bytes, size_t length);
char *adamic_path_bytes(const adamic_string *path);

// adamic_read_directory is readDirectory(path), { kind: 'Ok', names } with the names in Node's order,
// and adamic_file_status is fileStatus(path), { kind: 'Ok', type, size, symbolicLink }, each or
// { kind: 'Error', message } (directory.c). Both return a reference the caller owns.
adamic_object *adamic_read_directory(const adamic_string *path);
adamic_object *adamic_file_status(const adamic_string *path);
adamic_object *adamic_real_path(const adamic_string *path);

// adamic_write_text_file is writeTextFile(path, text) (input.c): { kind: 'Ok' } or { kind: 'Error',
// message }, a reference the caller owns.
adamic_object *adamic_write_text_file(const adamic_string *path, const adamic_string *text);

// adamic_panic writes "adamic: panic: <message>" to stderr and exits 70 (EX_SOFTWARE).
_Noreturn void adamic_panic(const char *message, size_t length);

// ADAMIC_CHECK_STACK starts every function the compiler emits: past adamic_stack_limit, the stack is
// nearly gone, and that's a panic, as Node's RangeError is, rather than a segfault (stack.c). The
// stack grows down on every processor Adamic targets. __builtin_frame_address is the real frame even
// when the address sanitizer keeps locals elsewhere.
extern _Thread_local uintptr_t adamic_stack_limit;
void adamic_stack_thread_start(void);
_Noreturn void adamic_stack_overflow(void);
#ifdef ADAMIC_TARGET_WASI
// Even a function using only Wasm locals must advance the linear stack. Otherwise
// its engine call stack can trap before this check sees any movement. The volatile
// endpoints preserve a 64-byte frame, including in optimized recursive functions.
#define ADAMIC_CHECK_STACK() \
	do { \
		volatile unsigned char adamic_stack_frame[64]; \
		adamic_stack_frame[0] = 0; \
		adamic_stack_frame[63] = 0; \
		if ((uintptr_t)adamic_stack_frame < adamic_stack_limit) { \
			adamic_stack_overflow(); \
		} \
	} while (0)
#else
#define ADAMIC_CHECK_STACK() \
	do { \
		if ((uintptr_t)__builtin_frame_address(0) < adamic_stack_limit) { \
			adamic_stack_overflow(); \
		} \
	} while (0)
#endif

// adamic_unreachable ends a function the checker proved always returns. Reaching it is a compiler
// bug, and it says so rather than returning garbage.
_Noreturn void adamic_unreachable(void);

#include "regexp.h"
#include "node_fs_file.h"
#include "node_buffer.h"
#include "node_crypto.h"
// Fixed plain literals can have public # keys; Object reflection refuses those shapes.
// Keep their enumeration distinct from the Object slice, which skips private class slots.
adamic_array *adamic_plain_object_keys(const adamic_object *object);
void *adamic_library_identity(size_t index);

adamic_maybe_number adamic_process_exit_code(void);
void adamic_process_set_exit_code(adamic_maybe_number code);
void adamic_process_exit(adamic_maybe_number code);
_Noreturn void adamic_process_exit_now(int code);
int adamic_process_status(void);
adamic_maybe_boolean adamic_process_is_tty(enum adamic_stream stream);
adamic_string *adamic_process_environment(const adamic_string *name);

#endif

#include "node_path.h"
#include "node_fs_directory.h"
#include "node_process.h"
