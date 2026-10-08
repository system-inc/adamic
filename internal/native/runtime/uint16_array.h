#ifndef ADAMIC_UINT16_ARRAY_H
#define ADAMIC_UINT16_ARRAY_H

#include "adamic.h"

// Uint16 values share the counted Int32 buffer until raw buffer views are supported.
adamic_typed_array *adamic_uint16_from_numbers(enum adamic_typed_array_kind kind, const adamic_array *numbers);

#endif
