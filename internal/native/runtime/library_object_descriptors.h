#ifndef ADAMIC_LIBRARY_OBJECT_DESCRIPTORS_H
#define ADAMIC_LIBRARY_OBJECT_DESCRIPTORS_H
#include "adamic.h"
// Kinds are ir.Type values emitted only after the complete plain shape is proved.
typedef struct adamic_descriptor_field { const char *name; int kind; } adamic_descriptor_field;
adamic_object *adamic_object_descriptor(const adamic_object *, const adamic_string *, const adamic_descriptor_field *, size_t);
adamic_object *adamic_object_descriptors(const adamic_object *, const adamic_shape *, const adamic_descriptor_field *, size_t);
bool adamic_object_property_enumerable(const adamic_object *, const adamic_string *);
adamic_maybe_boolean adamic_descriptor_flag(const adamic_object *, const adamic_string *);
adamic_heap *adamic_descriptor_value(const adamic_object *);
adamic_maybe_number adamic_descriptor_number(const adamic_object *);
adamic_maybe_boolean adamic_descriptor_boolean(const adamic_object *);
adamic_object *adamic_object_define_property(adamic_object *, const adamic_string *, const adamic_object *, const adamic_descriptor_field *, size_t);
adamic_object *adamic_object_define_properties(adamic_object *, const adamic_object *, const adamic_descriptor_field *, size_t);
#endif
