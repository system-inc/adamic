// Copyright 2011 the V8 project authors. All rights reserved.
// BSD-3-Clause, reproduced in THIRD_PARTY_NOTICES.md.
// Restricted port of PropertyDescriptor::ToObject and ValidateAndApplyPropertyDescriptor
// from V8 13.6.233 src/objects/property-descriptor.cc and src/objects/js-objects.cc.
// Plain data-property descriptors over a complete shape proven by lowering.
// Individual attribute changes and accessor definitions never reach this runtime.
#include "library_object_descriptors.h"
#include <string.h>

static const char *const descriptor_names[] = {"value", "writable", "enumerable", "configurable"};
static const bool descriptor_references[] = {true, true, true, true};
static const adamic_shape descriptor_shape = {4, descriptor_names, descriptor_references, NULL};

static bool key_is(const adamic_string *key, const char *name) {
 return key->length == strlen(name) && memcmp(key->bytes,name,key->length)==0;
}
static size_t find(const adamic_object *object,const adamic_string *key) {
 for(size_t i=0;i<object->shape->count;i++) if(object->shape->names[i][0]!='#' && key_is(key,object->shape->names[i])) return i;
 return object->shape->count;
}
static int field_kind(const char *name,const adamic_descriptor_field *fields,size_t count) {
 for(size_t i=0;i<count;i++) if(strcmp(name,fields[i].name)==0) return fields[i].kind;
 adamic_panic("property descriptor lacks proven slot metadata",sizeof "property descriptor lacks proven slot metadata"-1);
}
static adamic_heap *boxed(adamic_value value,int kind) {
 switch(kind) {
 case 1:return adamic_box_number(value.number);
 case 2:return value.boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
 case 7:{adamic_maybe_number n=adamic_maybe_number_unpack(value.number);return n.present ? adamic_box_number(n.number) : NULL;}
 default:return adamic_retain(value.reference);
 }
}
adamic_object *adamic_object_descriptor(const adamic_object *object,const adamic_string *key,const adamic_descriptor_field *fields,size_t count) {
 size_t index=find(object,key);
 bool captured=object->has_captured_stack && key_is(key,"stack");
 if(captured) adamic_panic("captured stack accessor descriptor is not represented",sizeof "captured stack accessor descriptor is not represented"-1);
 if(index==object->shape->count) return NULL;
 adamic_object *result=adamic_object_new(&descriptor_shape);
 result->slots[0].reference=captured ? adamic_retain(object->captured_stack.reference) : boxed(object->slots[index],field_kind(object->shape->names[index],fields,count));
 result->slots[1].reference=object->frozen && !captured ? &adamic_box_false : &adamic_box_true;
 result->slots[2].reference=captured ? &adamic_box_false : &adamic_box_true;
 result->slots[3].reference=(object->sealed || object->frozen) && !captured ? &adamic_box_false : &adamic_box_true;
 return result;
}
adamic_object *adamic_object_descriptors(const adamic_object *object,const adamic_shape *shape,const adamic_descriptor_field *fields,size_t count) {
 adamic_object *result=adamic_object_new(shape);
 for(size_t i=0;i<shape->count;i++) {
  const char *name=shape->names[i];
  adamic_string key={{0,adamic_kind_string,0},strlen(name),name,0,NULL,NULL,0};
  result->slots[i].reference=adamic_object_descriptor(object,&key,fields,count);
 }
 return result;
}
bool adamic_object_property_enumerable(const adamic_object *object,const adamic_string *key) {
 if(object->has_captured_stack && key_is(key,"stack")) return false;
 return find(object,key)!=object->shape->count;
}
adamic_maybe_boolean adamic_descriptor_flag(const adamic_object *descriptor,const adamic_string *key) {
 if(descriptor==NULL) return (adamic_maybe_boolean){false,false};
 size_t index=find(descriptor,key);
 if(index==descriptor->shape->count) return (adamic_maybe_boolean){false,false};
 const adamic_boolean_box *box=descriptor->slots[index].reference;
 return (adamic_maybe_boolean){true,box->boolean};
}
adamic_heap *adamic_descriptor_value(const adamic_object *descriptor) {
 return descriptor==NULL ? NULL : adamic_retain(descriptor->slots[0].reference);
}
adamic_maybe_number adamic_descriptor_number(const adamic_object *descriptor) {
 const adamic_heap *value=descriptor==NULL ? NULL : descriptor->slots[0].reference;
 if(value==NULL) return (adamic_maybe_number){false,0};
 if(value->kind!=adamic_kind_number) adamic_panic("descriptor value is not a proven number",sizeof "descriptor value is not a proven number"-1);
 return (adamic_maybe_number){true,((const adamic_number_box *)value)->number};
}
adamic_maybe_boolean adamic_descriptor_boolean(const adamic_object *descriptor) {
 const adamic_heap *value=descriptor==NULL ? NULL : descriptor->slots[0].reference;
 if(value==NULL) return (adamic_maybe_boolean){false,false};
 if(value->kind!=adamic_kind_boolean) adamic_panic("descriptor value is not a proven boolean",sizeof "descriptor value is not a proven boolean"-1);
 return (adamic_maybe_boolean){true,((const adamic_boolean_box *)value)->boolean};
}
static void cannot_redefine(const adamic_string *key) {
 static const char prefix[]="Cannot redefine property: ";
 static adamic_string type_error=ADAMIC_STRING("TypeError");
 adamic_string *message=adamic_string_allocate(sizeof prefix-1+key->length);
 memcpy((char *)message->bytes,prefix,sizeof prefix-1);
 memcpy((char *)message->bytes+sizeof prefix-1,key->bytes,key->length);
 adamic_object *error=adamic_error_new(message);adamic_release(message);
 adamic_release(error->slots[0].reference);error->slots[0].reference=adamic_retain(&type_error);adamic_error_tag(error);
 adamic_thrown=error;
}
static bool same_value(adamic_value left,adamic_value right,int kind) {
 if(kind==1) return (isnan(left.number) && isnan(right.number)) || (left.number==right.number && (left.number!=0 || signbit(left.number)==signbit(right.number)));
 if(kind==2) return left.boolean==right.boolean;
 return adamic_string_equal(left.reference,right.reference);
}
adamic_object *adamic_object_define_property(adamic_object *object,const adamic_string *key,const adamic_object *descriptor,const adamic_descriptor_field *fields,size_t count) {
 size_t index=find(object,key);
 if(index==object->shape->count) adamic_panic("definition would change the proven shape",sizeof "definition would change the proven shape"-1);
 const adamic_value *next=NULL;
 bool writable=false,configurable=false;
 for(size_t i=0;i<descriptor->shape->count;i++) {
  const char *name=descriptor->shape->names[i];
  if(strcmp(name,"value")==0) next=&descriptor->slots[i];
  else if(strcmp(name,"writable")==0) writable=true;
  else if(strcmp(name,"configurable")==0) configurable=true;
 }
 int kind=field_kind(object->shape->names[index],fields,count);
 if(((object->sealed || object->frozen) && configurable) || (object->frozen && writable) || (object->frozen && next!=NULL && !same_value(object->slots[index],*next,kind))) {cannot_redefine(key);return adamic_retain(object);}
 // V8 returns success without writing a frozen property with the same value.
 if(object->frozen) return adamic_retain(object);
 if(next!=NULL) {
  if(object->shape->references[index]) {adamic_retain(next->reference);adamic_release(object->slots[index].reference);}
  object->slots[index]=*next;
 }
 return adamic_retain(object);
}
adamic_object *adamic_object_define_properties(adamic_object *object,const adamic_object *descriptors,const adamic_descriptor_field *fields,size_t count) {
 for(size_t at=0;at<descriptors->shape->count;at++) {
  size_t i=adamic_public_index(descriptors->shape,at);
  const char *name=descriptors->shape->names[i];adamic_string key={{0,adamic_kind_string,0},strlen(name),name,0,NULL,NULL,0};
  adamic_release(adamic_object_define_property(object,&key,descriptors->slots[i].reference,fields,count));
  if(adamic_thrown!=NULL) break;
 }
 return adamic_retain(object);
}
