#ifndef ADAMIC_VIEW_ARRAYS_H
#define ADAMIC_VIEW_ARRAYS_H

// This byte certifies physical element storage, not initialization or a view's
// asserted logical element type. Zero means no scalar storage certificate.
typedef struct adamic_array_view_metadata {
    unsigned char storage;
} adamic_array_view_metadata;
struct adamic_array;
void adamic_view_array_missing(const char *expression, const char *expected);
void adamic_array_view_storage(struct adamic_array *array, unsigned char storage);
adamic_value *adamic_view_array_at(const struct adamic_array *array, double index, bool relative, bool undefined_allowed, unsigned char wanted, const char *expected, const char *expression, adamic_value *snapshot);
void adamic_view_array_storage_check(const struct adamic_array *array, unsigned char storage, const char *expression);
#endif
