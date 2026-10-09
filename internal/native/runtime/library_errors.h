#ifndef ADAMIC_LIBRARY_ERRORS_H
#define ADAMIC_LIBRARY_ERRORS_H
#include "adamic.h"
// Raise an ordinary library Error through the existing counted unwind protocol.
void adamic_library_throw(const char *name, const char *message, size_t length);
bool adamic_library_concat_check(size_t count, adamic_string *const parts[]);
adamic_string *adamic_library_concat(size_t count, adamic_string *const parts[]);
bool adamic_library_error_is_type(const adamic_object *error, const adamic_string *name);
#endif
