#ifndef ADAMIC_VIEW_ARRAYS_H
#define ADAMIC_VIEW_ARRAYS_H

// Internal extraction selector for nullable class references; not an IR type.
enum { adamic_view_array_nominal_union = 16 };

// Physical storage lives in the shared adamic_array.element_kind byte.
struct adamic_array;
struct adamic_array *adamic_view_array_slice(const struct adamic_array *array, double start, double end, bool has_end);
void adamic_view_array_push(struct adamic_array *array, adamic_value value);
void adamic_view_array_pop_commit(struct adamic_array *array);
void adamic_view_array_missing(const char *expression, const char *expected);
void adamic_array_view_storage(struct adamic_array *array, unsigned char storage);
adamic_value *adamic_view_array_at(const struct adamic_array *array, double index, bool relative, bool undefined_allowed, unsigned char wanted, const char *expected, const char *expression, adamic_value *snapshot, struct adamic_array *owner);
void adamic_view_array_storage_check(const struct adamic_array *array, unsigned char storage, const char *expression);
adamic_string *adamic_view_array_string(const struct adamic_array *array, const adamic_string *separator, bool checked, unsigned char wanted, bool undefined_allowed, const char *expected, const char *expression, size_t allowed_count, const adamic_value *allowed);
int adamic_view_array_default_compare(adamic_value left, adamic_value right, void *context);
#include "view_array_writes.h"
#endif
