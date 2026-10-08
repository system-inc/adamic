// Optional boolean element storage, separate from shared runtime implementations.
#include "adamic.h"

uint8_t adamic_maybe_boolean_pack(adamic_maybe_boolean value) {
	if (!value.present) return 0;
	return value.boolean ? 2 : 1;
}

adamic_maybe_boolean adamic_maybe_boolean_unpack(uint8_t value) {
	return (adamic_maybe_boolean){value != 0, value == 2};
}
