// nullable.c: one immortal null sentinel per migrated reference kind. NULL is undefined.
#include "adamic.h"

// This address is never returned by a string-producing operation. Its bytes are not its identity.
static adamic_string adamic_null_string = ADAMIC_STRING("");
static adamic_string adamic_string_null_text = ADAMIC_STRING("null");

void *adamic_reference_null(enum adamic_kind kind) {
 switch (kind) {
 case adamic_kind_string: return &adamic_null_string;
 default: return NULL; // Other kinds have not migrated from their existing representation yet.
 }
}

bool adamic_reference_is_null(const void *value, enum adamic_kind kind, bool undefined) {
 return (undefined && value == NULL) || value == adamic_reference_null(kind);
}

bool adamic_reference_is_sentinel(const void *value) {
 return value == &adamic_null_string;
}

adamic_string *adamic_reference_null_text(void) {
 return &adamic_string_null_text;
}
