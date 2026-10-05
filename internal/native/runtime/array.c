// array.c: arrays.

#include "adamic.h"

#include <stdlib.h>
#include <string.h>

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
		case adamic_join_strings: {
			const adamic_string *string = array->elements[index].reference;
			bytes = string->bytes;
			size = string->length;
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
	adamic_string piece = {{0, adamic_kind_string}, length, buffer};
	adamic_string *joined = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(buffer);
	return joined;
}
