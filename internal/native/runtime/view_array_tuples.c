#include "view_array_tuples.h"
#include <stdlib.h>
#include <string.h>

/* A transient one-slot probe reuses the object reader's exact representation
 * and readiness rules. It is never an escaping tuple or a stored array copy. */
static const adamic_object *array_tuple_probe(const adamic_object *value,const char *field,adamic_object *probe,adamic_shape *shape,const char **names,bool *references){
 if(value==NULL || value->heap.kind!=adamic_kind_array)return value;
 const adamic_array *array=(const adamic_array *)value;
 char *end=NULL;
 unsigned long index=strtoul(field,&end,10);
 adamic_value *slot=field[0]=='\0'||end==NULL||*end!='\0'?NULL:adamic_array_holes_at(array,(double)index);
 names[0]=field;references[0]=array->references;
 *shape=(adamic_shape){.count=slot==NULL?0:1,.names=names,.references=references};
 memset(probe,0,adamic_object_size(1));
 probe->heap.kind=adamic_kind_object;
 probe->shape=shape;
 probe->real_type="array";
 if(slot!=NULL){probe->slots[0]=*slot;adamic_object_initialized(probe)[0]=1;adamic_object_field_types(probe)[0]=array->element_kind;}
 return probe;
}

#define ARRAY_TUPLE_PROBE \
 _Alignas(adamic_object) unsigned char memory[adamic_object_size(1)]; \
 adamic_shape shape; const char *names[1]; bool references[1]; \
 adamic_slot_cache local_cache={0}; \
 const adamic_object *selected=array_tuple_probe(object,field,(adamic_object *)(void *)memory,&shape,names,references); \
 if(selected!=object)cache=&local_cache

adamic_value adamic_array_tuple_view(const adamic_object *object,const char *field,adamic_slot_cache *cache,unsigned char wanted,const char *type,const char *expression){
 ARRAY_TUPLE_PROBE;
 return adamic_object_view(selected,field,cache,wanted,type,expression);
}
adamic_value adamic_array_tuple_optional_view(const adamic_object *object,const char *field,adamic_slot_cache *cache,unsigned char wanted,const char *type,const char *expression,bool absent,bool optional,bool undefined_member){
 ARRAY_TUPLE_PROBE;
 return adamic_object_optional_view_undefined(selected,field,cache,wanted,type,expression,absent,optional,undefined_member);
}
adamic_heap *adamic_array_tuple_primitive_view(const adamic_object *object,const char *field,adamic_slot_cache *cache,const adamic_object_primitive_member *members,size_t count,bool undefined,bool absent,bool optional,const char *type,const char *expression){
 ARRAY_TUPLE_PROBE;
 return adamic_object_primitive_view(selected,field,cache,members,count,undefined,absent,optional,type,expression);
}
