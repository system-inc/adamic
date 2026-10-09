#ifndef ADAMIC_NAMESPACE_H
#define ADAMIC_NAMESPACE_H
#include "adamic.h"
bool adamic_namespace_is(const adamic_object *object);
adamic_object *adamic_namespace_new(void);
void adamic_namespace_install(adamic_object *, adamic_string *, adamic_value, bool);
void adamic_namespace_check_write(const adamic_object *, const adamic_string *);
adamic_value adamic_namespace_read(const adamic_object *, adamic_string *, unsigned char);
#endif
