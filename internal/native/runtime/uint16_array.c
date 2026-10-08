// Uint16 construction uses existing counted storage and exact low-bit conversion.
#include "uint16_array.h"

adamic_typed_array *adamic_uint16_from_numbers(enum adamic_typed_array_kind kind, const adamic_array *numbers) {
	adamic_typed_array *array = adamic_typed_array_new(kind, (double)numbers->length);
	for (size_t index = 0; index < numbers->length; index++) {
		adamic_typed_array_set(array, (double)index, adamic_bitwise_and(numbers->elements[index].number, 65535.0));
	}
	return array;
}
