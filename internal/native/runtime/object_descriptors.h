// object_descriptors.h: scalar data descriptors with an explicit lifetime.
#ifndef ADAMIC_OBJECT_DESCRIPTORS_H
#define ADAMIC_OBJECT_DESCRIPTORS_H
adamic_string *adamic_object_define_data(adamic_object *object, const adamic_string *key,
 const adamic_heap *value, const adamic_heap *writable, const adamic_heap *enumerable,
 const adamic_heap *configurable, unsigned mask, int representation);
bool adamic_object_descriptor_has(const adamic_object *object, const adamic_string *key);
bool adamic_object_enumerable(const adamic_object *object, const adamic_string *key);
bool adamic_object_descriptor_writable(const adamic_object *object, const char *key);
bool adamic_object_descriptor_integrity(const adamic_object *object, bool frozen);
struct adamic_array *adamic_object_descriptor_names(const adamic_object *object, bool all);
void adamic_object_descriptors_free(adamic_object *object, void (*release)(void *));
adamic_object *adamic_object_type_error(adamic_string *message);
#endif
