// array.c: arrays.

#include "adamic.h"
#include "graph_regions.h"

#include <math.h>
#include <stdio.h>

#include <stdlib.h>
#include <string.h>

adamic_array *adamic_array_new(size_t capacity, bool references) {
	return adamic_array_new_typed(capacity, references, NULL);
}

adamic_array *adamic_array_new_typed(size_t capacity, bool references, const adamic_json_schema *element) {
#ifdef ADAMIC_COUNT
	adamic_json_array_check(references, element);
#endif
	adamic_array *array = adamic_allocate(sizeof *array, adamic_kind_array);
	array->length = 0;
	array->capacity = capacity;
	array->references = references;
	array->elements = NULL;
	array->properties = NULL;
	array->json_element = element;
	array->json_elements = NULL;
	if (capacity > 0) {
		array->elements = malloc(capacity * sizeof *array->elements);
		if (array->elements == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
	}
	return array;
}

void adamic_array_push(adamic_array *array, adamic_value value) {
	if (array->length == array->capacity) {
		size_t capacity = array->capacity == 0 ? 4 : array->capacity * 2;
		adamic_value *grown = realloc(array->elements, capacity * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		array->elements = grown;
		if (array->json_elements != NULL) {
			const adamic_json_schema **tags = realloc(array->json_elements, capacity * sizeof *tags);
			if (tags == NULL)
				adamic_panic("out of memory", 13);
			array->json_elements = tags;
		}
		array->capacity = capacity;
	}
	if (array->json_elements != NULL)
		array->json_elements[array->length] = array->json_element;
	array->elements[array->length++] = value;
}

void adamic_array_tag(adamic_array *array, size_t index, const adamic_json_schema *schema) {
#ifdef ADAMIC_COUNT
	adamic_json_array_check(array->references, schema);
#endif
	if (array->json_elements == NULL && schema != array->json_element) {
		size_t capacity = array->capacity == 0 ? 4 : array->capacity;
		array->json_elements = malloc(capacity * sizeof *array->json_elements);
		if (array->json_elements == NULL)
			adamic_panic("out of memory", 13);
		for (size_t i = 0; i < array->length; i++)
			array->json_elements[i] = array->json_element;
	}
	if (array->json_elements != NULL)
		array->json_elements[index] = schema;
}

void adamic_array_push_typed(adamic_array *array, adamic_value value, const adamic_json_schema *schema) {
	adamic_array_push(array, value);
	adamic_array_tag(array, array->length - 1, schema);
}

void adamic_array_json_reset(adamic_array *array, const adamic_json_schema *schema) {
	free(array->json_elements);
	array->json_elements = NULL;
	array->json_element = schema;
}

static const adamic_json_schema *element_schema(const adamic_array *array, size_t index) {
	return array->json_elements == NULL ? array->json_element : array->json_elements[index];
}

// adamic_array_join builds the string in one buffer, growing it as it goes.
adamic_string *adamic_array_join(const adamic_array *array, const adamic_string *separator, enum adamic_join kind) {
	size_t length = 0, capacity = 64;
	char *buffer = malloc(capacity);
	if (buffer == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t index = 0; index < array->length; index++) {
		char number[ADAMIC_NUMBER_FORMAT_MAX];
		const char *bytes = NULL;
		size_t size = 0;
		switch (kind) {
		case adamic_join_numbers:
			size = adamic_number_format(array->elements[index].number, number);
			bytes = number;
			break;
		case adamic_join_booleans:
			bytes = array->elements[index].boolean ? "true" : "false";
			size = strlen(bytes);
			break;
		case adamic_join_maybe_numbers: {
			// join writes undefined as nothing at all.
			adamic_maybe_number element = adamic_maybe_number_unpack(array->elements[index].number);
			if (element.present) {
				size = adamic_number_format(element.number, number);
				bytes = number;
			}
			break;
		}
		case adamic_join_strings: {
			// An undefined element (a string | undefined array's NULL) joins as nothing, as JavaScript's does.
			const adamic_string *string = array->elements[index].reference;
			if (string != NULL) {
				bytes = string->bytes;
				size = string->length;
			}
			break;
		}
		}
		size_t needed = length + size + (index > 0 ? separator->length : 0);
		if (needed > capacity) {
			while (capacity < needed) {
				capacity *= 2;
			}
			char *grown = realloc(buffer, capacity);
			if (grown == NULL) {
				static const char message[] = "out of memory";
				adamic_panic(message, sizeof message - 1);
			}
			buffer = grown;
		}
		if (index > 0 && separator->length > 0) {
			memcpy(buffer + length, separator->bytes, separator->length);
			length += separator->length;
		}
		if (size > 0) {
			memcpy(buffer + length, bytes, size);
			length += size;
		}
	}
	adamic_string piece = {{0, adamic_kind_string, 0}, length, buffer, 0, NULL, NULL, 0};
	adamic_string *joined = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(buffer);
	return joined;
}

adamic_array *adamic_array_slice(const adamic_array *array, double start, double end, bool has_end) {
	// ECMAScript's relative indexes: ToIntegerOrInfinity, negative from the end, clamped.
	double length = (double)array->length;
	start = isnan(start) ? 0 : trunc(start);
	start = start < 0 ? (length + start < 0 ? 0 : length + start) : (start > length ? length : start);
	if (has_end) {
		end = isnan(end) ? 0 : trunc(end);
		end = end < 0 ? (length + end < 0 ? 0 : length + end) : (end > length ? length : end);
	} else {
		end = length;
	}
	size_t from = (size_t)start, to = end > start ? (size_t)end : from;
	adamic_array *sliced = adamic_array_new_typed(to - from, array->references, array->json_element);
	for (size_t index = from; index < to; index++) {
		adamic_value value = array->elements[index];
		if (array->references) {
			adamic_retain(value.reference);
		}
		adamic_array_push_typed(sliced, value, element_schema(array, index));
	}
	return sliced;
}

// adamic_array_sort sorts stably (ECMA-262 requires it since 2019), by V8's TimSort (sort.c), so a
// comparator that isn't consistent gives the order, and is called in the order, Node's is. compare is
// the program's comparator through an adapter that gives -1, 0 or 1, NaN read as 0, with context
// passed through. Like V8, it sorts a copy of the elements (each reference held) and writes the
// result back by index, so a comparator that changes the array can't pull memory out from under the
// sort.
typedef struct json_sort_entry {
	adamic_value value;
	const adamic_json_schema *schema;
} json_sort_entry;
typedef struct json_sort_context {
	int (*compare)(adamic_value, adamic_value, void *);
	void *context;
} json_sort_context;
static int compare_json_entries(adamic_value left, adamic_value right, void *opaque) {
	json_sort_context *context = opaque;
	return context->compare(((json_sort_entry *)left.reference)->value, ((json_sort_entry *)right.reference)->value, context->context);
}

void adamic_array_sort(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
	size_t length = array->length;
	if (length < 2) {
		return;
	}
	adamic_value *work = malloc(length * sizeof *work);
	if (work == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	bool references = array->references;
	for (size_t index = 0; index < length; index++) {
		work[index] = array->elements[index];
		if (references) {
			adamic_retain(work[index].reference);
		}
	}
	// The references just taken, as they were taken: a comparator that throws leaves the work copy
	// halfway through a merge, perhaps holding one twice and another not at all, so it's these that
	// are let go then.
	adamic_value *taken = NULL;
	if (references) {
		taken = malloc(length * sizeof *taken);
		if (taken == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		memcpy(taken, work, length * sizeof *taken);
	}
	json_sort_entry *entries = NULL;
	json_sort_context tagged_context = {compare, context};
	if (array->json_elements != NULL) {
		entries = malloc(length * sizeof *entries);
		if (entries == NULL)
			adamic_panic("out of memory", 13);
		for (size_t i = 0; i < length; i++) {
			entries[i] = (json_sort_entry){work[i], element_schema(array, i)};
			work[i].reference = &entries[i];
		}
	}
	if (!adamic_timsort(work, length, entries == NULL ? compare : compare_json_entries, entries == NULL ? context : &tagged_context)) {
		// A comparator threw: the array stays as it was, as V8 leaves it, and the throw goes on.
		if (references) {
			for (size_t index = 0; index < length; index++) {
				adamic_release(taken[index].reference);
			}
		}
		free(taken);
		free(entries);
		free(work);
		return;
	}
	free(taken);
	// Written back as V8 does, index by index: each held reference becomes the array's, and what it
	// held there is let go. Past a length the comparator shrank, the array grows again.
	for (size_t index = 0; index < length; index++) {
		const adamic_json_schema *schema = array->json_element;
		if (entries != NULL) {
			json_sort_entry *entry = work[index].reference;
			work[index] = entry->value;
			schema = entry->schema;
		}
		if (index < array->length) {
			adamic_value old = array->elements[index];
			array->elements[index] = work[index];
			if (references) {
				adamic_graph_take(array, work[index].reference);
			}
			if (references) {
				if (adamic_graph_is(array)) {
					adamic_graph_drop(array, old.reference);
				} else {
					adamic_release(old.reference);
				}
			}
		} else {
			if (references) {
				adamic_graph_take(array, work[index].reference);
			}
			adamic_array_push(array, work[index]);
		}
		adamic_array_tag(array, index, schema);
	}
	free(entries);
	free(work);
}

int adamic_compare_closure(adamic_value left, adamic_value right, void *context) {
	// JavaScript reads the comparator's result by its sign, and NaN as 0.
	adamic_closure *compare = context;
	double result = compare->code(compare, (adamic_value[]){left, right}).number;
	return result < 0 ? -1 : result > 0 ? 1
										: 0;
}

adamic_value *adamic_array_at_relative(const adamic_array *array, double index) {
	// array.at(index): ToIntegerOrInfinity, so NaN is 0 and a fraction truncates, then a negative
	// index counts from the end. Anywhere outside the array is undefined.
	index = isnan(index) ? 0 : trunc(index);
	if (index < 0) {
		index += (double)array->length;
	}
	if (!(index >= 0) || index >= (double)array->length) {
		return NULL;
	}
	return &array->elements[(size_t)index];
}

double adamic_array_index_of(const adamic_array *array, adamic_value value, enum adamic_equality equality, bool same_value_zero) {
	for (size_t index = 0; index < array->length; index++) {
		adamic_value element = array->elements[index];
		bool equal = false;
		switch (equality) {
		case adamic_equal_numbers:
			// === never finds NaN; SameValueZero (includes) does. Both take 0 and -0 as equal.
			equal = element.number == value.number || (same_value_zero && isnan(element.number) && isnan(value.number));
			break;
		case adamic_equal_booleans:
			equal = element.boolean == value.boolean;
			break;
		case adamic_equal_strings:
			equal = adamic_string_equal(element.reference, value.reference);
			break;
		case adamic_equal_identity:
			equal = element.reference == value.reference;
			break;
		case adamic_equal_maybe_numbers: {
			// undefined finds undefined; a number is compared as numbers are, NaN found only by includes.
			adamic_maybe_number left = adamic_maybe_number_unpack(element.number);
			adamic_maybe_number right = adamic_maybe_number_unpack(value.number);
			equal = left.present == right.present && (!left.present || left.number == right.number || (same_value_zero && isnan(left.number) && isnan(right.number)));
			break;
		}
		}
		if (equal) {
			return (double)index;
		}
	}
	return -1;
}

adamic_array *adamic_array_reverse(adamic_array *array) {
	for (size_t left = 0, right = array->length; left + 1 < right; left++, right--) {
		adamic_value swapped = array->elements[left];
		array->elements[left] = array->elements[right - 1];
		array->elements[right - 1] = swapped;
		if (array->json_elements != NULL) {
			const adamic_json_schema *tag = array->json_elements[left];
			array->json_elements[left] = array->json_elements[right - 1];
			array->json_elements[right - 1] = tag;
		}
	}
	return array;
}

adamic_array *adamic_array_filled(double length, adamic_value value, bool references) {
	return adamic_array_filled_typed(length, value, references, NULL);
}
adamic_array *adamic_array_filled_typed(double length, adamic_value value, bool references, const adamic_json_schema *schema) {
	// new Array(length): an integer from 0 to 2^32 - 1, or JavaScript throws.
	if (!(length >= 0) || length > 4294967295.0 || length != trunc(length)) {
		static const char message[] = "RangeError: Invalid array length";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_array *array = adamic_array_new_typed((size_t)length, references, schema);
	for (size_t index = 0; index < (size_t)length; index++) {
		if (references) {
			adamic_retain(value.reference);
		}
		adamic_array_push(array, value);
	}
	return array;
}

adamic_array *adamic_array_fill(adamic_array *array, adamic_value value, double start, double end, bool has_start, bool has_end) {
	return adamic_array_fill_typed(array, value, start, end, has_start, has_end, array->json_elements == NULL ? array->json_element : NULL);
}
adamic_array *adamic_array_fill_typed(adamic_array *array, adamic_value value, double start, double end, bool has_start, bool has_end, const adamic_json_schema *schema) {
	// ECMAScript's relative indexes, as slice reads them.
	double length = (double)array->length;
	start = has_start ? (isnan(start) ? 0 : trunc(start)) : 0;
	start = start < 0 ? (length + start < 0 ? 0 : length + start) : (start > length ? length : start);
	end = has_end ? (isnan(end) ? 0 : trunc(end)) : length;
	end = end < 0 ? (length + end < 0 ? 0 : length + end) : (end > length ? length : end);
	for (size_t index = (size_t)start; (double)index < end; index++) {
		if (array->references) {
			// The new reference first: the value may be the one already there.
			if (adamic_graph_is(array)) {
				adamic_graph_hold(array, value.reference);
				adamic_graph_drop(array, array->elements[index].reference);
			} else {
				adamic_retain(value.reference);
				adamic_release(array->elements[index].reference);
			}
		}
		array->elements[index] = value;
		adamic_array_tag(array, index, schema);
	}
	return array;
}

// splice_into is splice, the removed elements moving to removed, or let go when removed is NULL:
// a splice whose result nothing uses (adamic_array_remove) allocates no array to hold them.
static void splice_into(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items, adamic_array **removed, const adamic_json_schema *schema) {
	// ECMAScript's relative start, clamped to the array; a count left out is everything after it, and
	// a count given is clamped to what's there.
	double length = (double)array->length;
	start = isnan(start) ? 0 : trunc(start);
	start = start < 0 ? (length + start < 0 ? 0 : length + start) : (start > length ? length : start);
	double removing = length - start;
	if (has_count) {
		count = isnan(count) ? 0 : trunc(count);
		removing = count < 0 ? 0 : (count > removing ? removing : count);
	}
	size_t from = (size_t)start, removed_count = (size_t)removing;
	// The removed elements move to the result, their references with them, or are let go.
	if (removed != NULL) {
		*removed = adamic_array_new_typed(removed_count, array->references, array->json_element);
		for (size_t index = 0; index < removed_count; index++) {
			adamic_value value = array->elements[from + index];
			if (array->references) {
				adamic_graph_escape(array, value.reference);
			}
			adamic_array_push_typed(*removed, value, element_schema(array, from + index));
		}
	}
	// Let go of after the array is whole again, below: releasing one may free what releases another.
	size_t after = array->length - from - removed_count;
	size_t new_length = array->length - removed_count + item_count;
	adamic_value *dropped = NULL;
	if (removed == NULL && array->references && removed_count > 0) {
		dropped = malloc(removed_count * sizeof *dropped);
		if (dropped == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		memcpy(dropped, array->elements + from, removed_count * sizeof *dropped);
	}
	if (new_length > array->capacity) {
		adamic_value *grown = realloc(array->elements, new_length * sizeof *grown);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		array->elements = grown;
		array->capacity = new_length;
	}
	if (array->json_elements != NULL) {
		const adamic_json_schema **tags = realloc(array->json_elements, array->capacity * sizeof *tags);
		if (tags == NULL && array->capacity != 0)
			adamic_panic("out of memory", 13);
		array->json_elements = tags;
		if (after > 0)
			memmove(tags + from + item_count, tags + from + removed_count, after * sizeof *tags);
	}
	// The tail moves to make room (or close the gap), then the items, which the array takes, go in.
	// An array that never held anything has no storage, and arithmetic on its null pointer is undefined
	// even to move nothing, so nothing is moved unless there's a tail.
	if (after > 0) {
		memmove(array->elements + from + item_count, array->elements + from + removed_count, after * sizeof *array->elements);
	}
	if (item_count > 0) {
		memcpy(array->elements + from, items, item_count * sizeof *items);
	}
	array->length = new_length;
	for (size_t i = 0; i < item_count; i++)
		adamic_array_tag(array, from + i, schema);
	if (dropped != NULL) {
		for (size_t index = 0; index < removed_count; index++) {
			if (adamic_graph_is(array)) {
				adamic_graph_drop(array, dropped[index].reference);
			} else {
				adamic_release(dropped[index].reference);
			}
		}
		free(dropped);
	}
}

adamic_array *adamic_array_splice(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items) {
	adamic_array *removed;
	splice_into(array, start, count, has_count, item_count, items, &removed, array->json_elements == NULL ? array->json_element : NULL);
	return removed;
}

void adamic_array_remove(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items) {
	splice_into(array, start, count, has_count, item_count, items, NULL, array->json_elements == NULL ? array->json_element : NULL);
}

void adamic_array_append(adamic_array *array, const adamic_array *source) {
	// The length is read once: [...items, ...items] spreads items twice, never into itself.
	size_t length = source->length;
	for (size_t index = 0; index < length; index++) {
		adamic_value value = source->elements[index];
		if (array->references) {
			adamic_graph_hold(array, value.reference);
		}
		adamic_array_push_typed(array, value, element_schema(source, index));
	}
}

adamic_array *adamic_array_concat(size_t count, adamic_array *const arrays[]) {
	size_t length = 0;
	for (size_t which = 0; which < count; which++) {
		length += arrays[which]->length;
	}
	adamic_array *joined = adamic_array_new_typed(length, arrays[0]->references, arrays[0]->json_element);
	for (size_t which = 0; which < count; which++) {
		for (size_t index = 0; index < arrays[which]->length; index++) {
			adamic_value value = arrays[which]->elements[index];
			if (joined->references) {
				adamic_retain(value.reference);
			}
			adamic_array_push_typed(joined, value, element_schema(arrays[which], index));
		}
	}
	return joined;
}

void adamic_array_set(adamic_array *array, double index, adamic_value value) {
	// 0.1 writes only at an index the array has: JavaScript would grow the array, or leave a hole, and
	// a hole is something 0.1 can't hold. push is how to append.
	adamic_value *slot = adamic_array_at(array, index);
	if (slot == NULL) {
		char number[ADAMIC_NUMBER_FORMAT_MAX], length[ADAMIC_NUMBER_FORMAT_MAX], message[128];
		size_t number_size = adamic_number_format(index, number);
		size_t length_size = adamic_number_format((double)array->length, length);
		int written = snprintf(message, sizeof message, "index %.*s is outside an array of length %.*s", (int)number_size, number, (int)length_size, length);
		adamic_panic(message, (size_t)written);
	}
	if (array->references) {
		void *old = slot->reference;
		slot->reference = value.reference;
		if (adamic_graph_is(array)) {
			adamic_graph_drop(array, old);
		} else {
			adamic_release(old);
		}
		return;
	}
	*slot = value;
}

adamic_array *adamic_array_splice_typed(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items, const adamic_json_schema *schema) {
	adamic_array *removed;
	splice_into(array, start, count, has_count, item_count, items, &removed, schema);
	return removed;
}
void adamic_array_remove_typed(adamic_array *array, double start, double count, bool has_count, size_t item_count, const adamic_value *items, const adamic_json_schema *schema) {
	splice_into(array, start, count, has_count, item_count, items, NULL, schema);
}
void adamic_array_set_typed(adamic_array *array, double index, adamic_value value, const adamic_json_schema *schema) {
	adamic_array_set(array, index, value);
	adamic_array_tag(array, (size_t)index, schema);
}
