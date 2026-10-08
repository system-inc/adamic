#ifndef ADAMIC_VIEW_ARRAY_TUPLES_H
#define ADAMIC_VIEW_ARRAY_TUPLES_H
#include "view_unions_object_primitive.h"
adamic_value adamic_array_tuple_view(const adamic_object *,const char *,adamic_slot_cache *,unsigned char,const char *,const char *);
adamic_value adamic_array_tuple_optional_view(const adamic_object *,const char *,adamic_slot_cache *,unsigned char,const char *,const char *,bool,bool,bool);
adamic_heap *adamic_array_tuple_primitive_view(const adamic_object *,const char *,adamic_slot_cache *,const adamic_object_primitive_member *,size_t,bool,bool,bool,const char *,const char *);
#endif
