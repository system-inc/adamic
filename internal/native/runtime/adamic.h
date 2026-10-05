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
	adamic_kind_cell,
	adamic_kind_closure,
	adamic_kind_map_iterator,
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
	// iterating counts the iterations open over the map; while there are any, its entries keep their
	// places (map.c).
	size_t iterating;
} adamic_map;

// adamic_map_iterator is one for...of over a map, in insertion order: entries added before it gets
// to them are visited, and entries deleted before it gets to them aren't, as ECMA-262 requires. It
// holds the map, and letting go of it ends the iteration.
typedef struct adamic_map_iterator {
	adamic_heap heap;
	adamic_map *map;
	size_t next;
} adamic_map_iterator;

adamic_map_iterator *adamic_map_iterate(adamic_map *map);

// adamic_map_iterator_next gives the next live entry's key and value, borrowed, or false at the end.
bool adamic_map_iterator_next(adamic_map_iterator *iterator, union adamic_value *key, union adamic_value *value);

adamic_map *adamic_map_new(bool string_keys, bool reference_values);

// adamic_map_get is the value's slot, or NULL when the key isn't there.
adamic_value *adamic_map_get(const adamic_map *map, adamic_value key);

// adamic_map_set takes the key and value it's given: references passed in belong to the map.
void adamic_map_set(adamic_map *map, adamic_value key, adamic_value value);

bool adamic_map_delete(adamic_map *map, adamic_value key);

// adamic_map_free_children lets go of what a map holds, for the heap's freeing.
void adamic_map_free_children(adamic_map *map, void (*let_go)(void *));

// adamic_array_at is array[index]: the element's slot, or NULL (undefined) when there isn't one.
adamic_value *adamic_array_at(const adamic_array *array, double index);

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
};
double adamic_array_index_of(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool same_value_zero);

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

// adamic_map_entries is [...map]: [key, value] pairs, each an object of the shape given.
adamic_array *adamic_map_entries(const adamic_map *map, const adamic_shape *pair);

// adamic_array_join is array.join(separator), each element written as String() would.
enum adamic_join {
	adamic_join_numbers = 1,
	adamic_join_booleans,
	adamic_join_strings,
};
struct adamic_string *adamic_array_join(const adamic_array *array, const struct adamic_string *separator, enum adamic_join kind);

// These return a reference the caller owns.
adamic_string *adamic_string_from_number(double value);

// Number.parseInt(text, radix) and Number.parseFloat(text), as Node answers them (parse.c). A radix
// of 0 is the one left out.
double adamic_number_parse_int(const adamic_string *text, double radix);
double adamic_number_parse_float(const adamic_string *text);
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

// The rest of a string's UTF-16 view (string.c), each as JavaScript means it. These that return a
// string return one the caller owns.
adamic_string *adamic_string_slice(const adamic_string *string, double start, double end, bool has_end);
adamic_maybe_number adamic_string_code_point_at(const adamic_string *string, double position);
int adamic_string_compare(const adamic_string *left, const adamic_string *right);
adamic_string *adamic_string_repeat(const adamic_string *string, double count);
adamic_string *adamic_string_pad(const adamic_string *string, double target, const adamic_string *fill, bool at_start);
double adamic_string_index_of(const adamic_string *string, const adamic_string *search);
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

// adamic_number_to_fixed is value.toFixed(digits), and returns a string the caller owns.
adamic_string *adamic_number_to_fixed(double value, double digits);

// adamic_panic writes "adamic: panic: <message>" to stderr and exits 70 (EX_SOFTWARE).
_Noreturn void adamic_panic(const char *message, size_t length);

// adamic_unreachable ends a function the checker proved always returns. Reaching it is a compiler
// bug, and it says so rather than returning garbage.
_Noreturn void adamic_unreachable(void);

#endif
