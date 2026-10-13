// object_integrity.h: Object integrity methods over proven own-property shapes.
#ifndef ADAMIC_OBJECT_INTEGRITY_H
#define ADAMIC_OBJECT_INTEGRITY_H
bool adamic_object_is_extensible(const adamic_heap *value);
bool adamic_object_test_integrity(const adamic_heap *value, bool frozen);
adamic_object *adamic_object_set_integrity(adamic_object *object, bool sealed);
void adamic_object_throw_type_error(const char *message, size_t length);
void *adamic_receiver_set_integrity(adamic_heap *value, bool sealed, bool frozen, int shape);
#endif
