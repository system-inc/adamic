#include "view_tuples.h"
#include "view_arrays.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Shape size is the immutable fixed tuple length; ordinary records never acquire
// tuple identity through an assertion or a numeric-looking property name.
bool adamic_tuple_matches(const adamic_object *value, size_t length) {
 return value != NULL && (value->heap.kind == adamic_kind_array ? ((const adamic_array *)value)->length == length : value->heap.kind == adamic_kind_object && value->tuple && value->shape->count == length);
}
void adamic_view_tuple(const adamic_object *value, size_t length, const char *expected, const char *expression) {
 if (adamic_tuple_matches(value, length)) return;
 const char *found = value == NULL ? "undefined" : value->heap.kind == adamic_kind_array || (value->heap.kind == adamic_kind_object && value->tuple) ? "array" : "object";
 size_t capacity = strlen(expression) + 2 * strlen(expected) + strlen(found) + 100;
 char *message = malloc(capacity);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int written = snprintf(message, capacity, "field read failed: %s is not a %s; expected %s, found %s", expression, expected, expected, found);
 adamic_panic(message, (size_t)written);
}

// Variable tuple plans use actual arity in the existing immutable shape.
bool adamic_tuple_matches_range(const adamic_object *value, size_t minimum, size_t maximum, bool rest) {
 if (value == NULL) return false;
 if (value->heap.kind == adamic_kind_array) { size_t length=((const adamic_array *)value)->length; return length>=minimum && (rest || length<=maximum); }
 return value->heap.kind == adamic_kind_object && value->tuple && value->shape->count >= minimum && (rest || value->shape->count <= maximum);
}
void adamic_view_tuple_range(const adamic_object *value, size_t minimum, size_t maximum, bool rest, const char *expected, const char *expression) {
 if (adamic_tuple_matches_range(value, minimum, maximum, rest)) return;
 adamic_view_tuple(value, SIZE_MAX, expected, expression);
}

// Normalize only the selected source-certified slot, retaining or boxing it.
// The existing reader owns bounds, holes, storage and undefined checks.
bool adamic_tuple_array_union_at(const adamic_array *array, double index, bool relative, bool undefined_allowed, const char *expected, const char *expression, adamic_value *result) {
 adamic_value snapshot;
 unsigned char storage = array == NULL ? 0 : array->references ? 10 : array->element_kind;
 adamic_value *slot = adamic_view_array_at(array, index, relative, undefined_allowed, storage == 10 ? adamic_view_array_tuple_union : storage, expected, expression, &snapshot, NULL);
 if (slot == NULL) { result->reference = NULL; return false; }
 if (storage == 10) result->reference = adamic_retain(slot->reference);
 else if (storage == 1) result->reference = adamic_box_number(slot->number);
 else if (storage == 7) result->reference = adamic_box_number(adamic_maybe_number_unpack(slot->number).number);
 else if (storage == 2) result->reference = slot->boolean ? (void *)&adamic_box_true : (void *)&adamic_box_false;
 else { adamic_view_array_missing(expression, expected); }
 return true;
}
