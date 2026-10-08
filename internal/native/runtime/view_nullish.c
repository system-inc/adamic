#include "view_nullish.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static unsigned char logical_kind(const adamic_heap *value) {
 if (value == NULL) return adamic_rep_undefined;
 switch (value->kind) {
 case adamic_kind_number: return adamic_rep_number;
 case adamic_kind_boolean: return adamic_rep_boolean;
 case adamic_kind_string: return adamic_rep_string;
 case adamic_kind_object: return adamic_rep_object;
 case adamic_kind_array: return adamic_rep_array;
 case adamic_kind_map: return adamic_rep_map;
 case adamic_kind_closure: return adamic_rep_closure;
 case adamic_kind_null: return adamic_rep_null;
 default: return 0;
 }
}
static _Noreturn void failure(const char *expression, const char *expected, unsigned char kind) {
 const char *found=kind==adamic_rep_number?"number":kind==adamic_rep_boolean?"boolean":kind==adamic_rep_string?"string":kind==adamic_rep_object?"object":kind==adamic_rep_array?"array":kind==adamic_rep_map?"Map":kind==adamic_rep_closure?"function":kind==adamic_rep_null?"null":kind==adamic_rep_undefined?"undefined":"unsupported representation";
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
 if(actual==adamic_rep_number)number=slot->number;
 else if(actual==adamic_rep_boolean)boolean=slot->boolean;
 else if(actual==adamic_rep_maybe_number){adamic_maybe_number value=adamic_maybe_number_unpack(slot->number);kind=value.present?adamic_rep_number:adamic_rep_undefined;number=value.number;}
 else if(actual==adamic_rep_maybe_boolean){adamic_maybe_boolean value=adamic_maybe_boolean_unpack(slot->maybe_boolean);kind=value.present?adamic_rep_boolean:adamic_rep_undefined;boolean=value.boolean;}
 else if(actual==adamic_rep_union){reference=slot->reference;kind=logical_kind(reference);if(kind==adamic_rep_number)number=((adamic_number_box *)reference)->number;else if(kind==adamic_rep_boolean)boolean=((adamic_boolean_box *)reference)->boolean;}
 else if((actual>=adamic_rep_string&&actual<=adamic_rep_map)||actual==adamic_rep_closure||actual==adamic_rep_record){reference=slot->reference;kind=reference==NULL?adamic_rep_undefined:logical_kind(reference);if(reference!=NULL&&kind!=(actual==adamic_rep_record?adamic_rep_object:actual))kind=0;}
 else if(actual!=adamic_rep_null&&actual!=adamic_rep_undefined)kind=0;
 if(kind==adamic_rep_null){if(null_allowed)return &adamic_null;failure(expression,expected,kind);}
 if(kind==adamic_rep_undefined){if(undefined_allowed)return NULL;failure(expression,expected,kind);}
 if(kind==0||(kinds&(1u<<kind))==0)failure(expression,expected,kind);
 if(kind==adamic_rep_number)return adamic_box_number(number);
 if(kind==adamic_rep_boolean)return boolean?&adamic_box_true.heap:&adamic_box_false.heap;
 return adamic_retain(reference);
}
