#include "view_nullish.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static unsigned char logical_kind(const adamic_heap *value) {
 if (value == NULL) return 13;
 switch (value->kind) {
 case adamic_kind_number: return 1;
 case adamic_kind_boolean: return 2;
 case adamic_kind_string: return 3;
 case adamic_kind_object: return 4;
 case adamic_kind_array: return 5;
 case adamic_kind_map: return 6;
 case adamic_kind_closure: return 8;
 case adamic_kind_null: return 12;
 default: return 0;
 }
}
static _Noreturn void failure(const char *expression, const char *expected, unsigned char kind) {
 const char *found=kind==1?"number":kind==2?"boolean":kind==3?"string":kind==4?"object":kind==5?"array":kind==6?"Map":kind==8?"function":kind==12?"null":kind==13?"undefined":"unsupported representation";
 size_t capacity=strlen(expression)+2*strlen(expected)+strlen(found)+96;
 char *message=malloc(capacity);
 if(message==NULL){static const char oom[]="out of memory";adamic_panic(oom,sizeof oom-1);}
 int length=snprintf(message,capacity,"field read failed: %s matches no member of %s; expected %s, found %s",expression,expected,expected,found);
 adamic_panic(message,(size_t)length);
}
_Noreturn void adamic_nullish_failure(const char *expression,const char *expected,const adamic_heap *value){failure(expression,expected,logical_kind(value));}

// Check readiness and producer storage before touching the active union member.
// Return an owned snapshot: scalar boxes are made only after the logical check.
adamic_heap *adamic_object_nullish_view(const adamic_object *object,const char *name,adamic_slot_cache *cache,unsigned int kinds,bool null_allowed,bool undefined_allowed,bool absent,bool optional,const char *expression,const char *expected){
 if(optional&&object==NULL)return NULL;
 adamic_value *slot=object==NULL?NULL:adamic_object_optional_field(object,name,cache);
 if(slot==NULL&&absent)return NULL;
 const adamic_object *owner=NULL;
 slot=adamic_object_read_contract(object,name,cache,expression,expected,&owner);
 unsigned char actual=adamic_object_field_types(owner)[cache->index],kind=actual;
 adamic_heap *reference=NULL;
 double number=0;bool boolean=false;
 if(actual==1)number=slot->number;
 else if(actual==2)boolean=slot->boolean;
 else if(actual==7){adamic_maybe_number value=adamic_maybe_number_unpack(slot->number);kind=value.present?1:13;number=value.number;}
 else if(actual==9){adamic_maybe_boolean value=adamic_maybe_boolean_unpack(slot->maybe_boolean);kind=value.present?2:13;boolean=value.boolean;}
 else if(actual==10){reference=slot->reference;kind=logical_kind(reference);if(kind==1)number=((adamic_number_box *)reference)->number;else if(kind==2)boolean=((adamic_boolean_box *)reference)->boolean;}
 else if((actual>=3&&actual<=6)||actual==8||actual==14){reference=slot->reference;kind=reference==NULL?13:logical_kind(reference);if(reference!=NULL&&kind!=(actual==14?4:actual))kind=0;}
 else if(actual!=12&&actual!=13)kind=0;
 if(kind==12){if(null_allowed)return &adamic_null;failure(expression,expected,kind);}
 if(kind==13){if(undefined_allowed)return NULL;failure(expression,expected,kind);}
 if(kind==0||(kinds&(1u<<kind))==0)failure(expression,expected,kind);
 if(kind==1)return adamic_box_number(number);
 if(kind==2)return boolean?&adamic_box_true.heap:&adamic_box_false.heap;
 return adamic_retain(reference);
}
