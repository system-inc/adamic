#ifndef ADAMIC_VIEW_DICTIONARIES_H
#define ADAMIC_VIEW_DICTIONARIES_H
#include "view_unions_mixed.h"

// A producer-certified record with boxed reference values uses the existing
// record table and union boxes. Never infer boxed storage from a target view.
// The source adapter must establish record identity before passing its pointer.
typedef struct adamic_view_dictionary_result {
    adamic_view_union_value value;
    size_t contract;
} adamic_view_dictionary_result;

// Result and payload are borrowed. The caller preserves contract on aliases and
// applies the shared field/element checks at later reads. Admission scans no keys.
adamic_view_dictionary_result adamic_view_dictionary_read(const adamic_record *record, bool boxed_storage, const adamic_string *key, unsigned int kinds, size_t child_contract, const char *expression, const char *declared);
adamic_view_dictionary_result adamic_view_dictionary_source_read(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child_contract, const char *expression, const char *declared);
// Borrowed checked object reference. No keys or element values are inspected.
adamic_object *adamic_view_dictionary_source_conversion(const adamic_heap *value, const char *expression, const char *declared);
adamic_heap *adamic_view_dictionary_box(adamic_view_union_value value);
// Owned key array. Element values and their representation are never read.
adamic_array *adamic_view_dictionary_source_keys(const adamic_object *object);
#endif
