// array.c: arrays.

#include "adamic.h"

#include <stdlib.h>

adamic_array *adamic_array_new(size_t capacity, bool references) {
	adamic_array *array = adamic_allocate(sizeof *array, adamic_kind_array);
	array->length = 0;
	array->capacity = capacity;
	array->references = references;
	array->elements = NULL;
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
		array->capacity = capacity;
	}
	array->elements[array->length++] = value;
}
