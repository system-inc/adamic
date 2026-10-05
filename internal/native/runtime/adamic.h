// adamic.h: the Adamic runtime. Small, plain C11, compiled with every program.

#ifndef ADAMIC_H
#define ADAMIC_H

#include <stdbool.h>
#include <stddef.h>

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
};

typedef struct adamic_heap {
	size_t references;
	enum adamic_kind kind;
} adamic_heap;

// adamic_retain and adamic_release take any heap value. NULL (undefined) is left alone.
void *adamic_retain(void *value);
void adamic_release(void *value);

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

// adamic_string is an immutable string: UTF-8 bytes (string.c).
typedef struct adamic_string {
	adamic_heap heap;
	size_t length;
	const char *bytes;
} adamic_string;

// ADAMIC_STRING is a constant: ADAMIC_STRING("text") as a static adamic_string's initializer.
#define ADAMIC_STRING(text) {{0, adamic_kind_string}, sizeof text - 1, text}

// adamic_shape is an object's layout: its fields' names in order, and which fields hold references.
typedef struct adamic_shape {
	size_t count;
	const char *const *names;
	const bool *references;
} adamic_shape;

// adamic_object is a plain object (object.c). Its shape travels with it, so the same object can be
// seen through any type it satisfies, as JavaScript allows, without ever being copied.
typedef struct adamic_object {
	adamic_heap heap;
	const adamic_shape *shape;
	adamic_value slots[];
} adamic_object;

// adamic_slot_cache remembers, at one place in the program that reads a field, where the field was in
// the last shape seen there.
typedef struct adamic_slot_cache {
	const adamic_shape *shape;
	size_t index;
} adamic_slot_cache;

// adamic_object_new makes an object of a shape, its fields zeroed for the caller to fill; a reference
// stored in a field belongs to the object.
adamic_object *adamic_object_new(const adamic_shape *shape);

// adamic_object_copy is { ...source }: the same shape, its references retained.
adamic_object *adamic_object_copy(const adamic_object *source);

// adamic_object_field finds a field by name. The checker proved the field is there.
adamic_value *adamic_object_field(const adamic_object *object, const char *name, adamic_slot_cache *cache);

// adamic_array is an array (array.c). references says whether its elements are references.
typedef struct adamic_array {
	adamic_heap heap;
	size_t length;
	size_t capacity;
	bool references;
	adamic_value *elements;
} adamic_array;

adamic_array *adamic_array_new(size_t capacity, bool references);

// adamic_array_push appends; a reference pushed belongs to the array.
void adamic_array_push(adamic_array *array, adamic_value value);

// adamic_map is a Map with string or number keys (map.c).
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
	bool reference_values;
} adamic_map;

adamic_map *adamic_map_new(bool string_keys, bool reference_values);

// adamic_map_get is the value's slot, or NULL when the key isn't there.
adamic_value *adamic_map_get(const adamic_map *map, adamic_value key);

// adamic_map_set takes the key and value it's given: references passed in belong to the map.
void adamic_map_set(adamic_map *map, adamic_value key, adamic_value value);

bool adamic_map_delete(adamic_map *map, adamic_value key);

// adamic_map_free_children lets go of what a map holds, for the heap's freeing.
void adamic_map_free_children(adamic_map *map, void (*let_go)(void *));

// adamic_array_join is array.join(separator), each element written as String() would.
enum adamic_join {
	adamic_join_numbers = 1,
	adamic_join_booleans,
	adamic_join_strings,
};
struct adamic_string *adamic_array_join(const adamic_array *array, const struct adamic_string *separator, enum adamic_join kind);

// These return a reference the caller owns.
adamic_string *adamic_string_from_number(double value);
adamic_string *adamic_string_concat(size_t count, adamic_string *const parts[]);

// The strings every program has: "", and String(true) and String(false).
extern adamic_string adamic_string_empty;
extern adamic_string adamic_string_true;
extern adamic_string adamic_string_false;

// A string's UTF-16 view (string.c): length, charCodeAt and trim as JavaScript means them.
double adamic_string_length(const adamic_string *string);
double adamic_string_char_code_at(const adamic_string *string, double position);
adamic_string *adamic_string_trim(adamic_string *string);

// for...of over a string walks code points: adamic_string_next is the byte size of the one at offset,
// and adamic_string_slice_bytes makes it a string the caller owns.
size_t adamic_string_next(const adamic_string *string, size_t offset);
adamic_string *adamic_string_slice_bytes(const adamic_string *string, size_t offset, size_t size);

// adamic_string_equal is ===.
int adamic_string_equal(const adamic_string *left, const adamic_string *right);

// adamic_write_line writes a string and a newline, as console.log does with one string.
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

// adamic_number_to_fixed is value.toFixed(digits), and returns a string the caller owns.
adamic_string *adamic_number_to_fixed(double value, double digits);

// adamic_panic writes "adamic: panic: <message>" to stderr and exits 70 (EX_SOFTWARE).
_Noreturn void adamic_panic(const char *message, size_t length);

// adamic_unreachable ends a function the checker proved always returns. Reaching it is a compiler
// bug, and it says so rather than returning garbage.
_Noreturn void adamic_unreachable(void);

#endif
