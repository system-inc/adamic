// Checked reflection over actual own data slots. Metadata is immutable generated storage.
#include "object_reflection.h"
#include <stdlib.h>
#include <string.h>

typedef struct reflected_value { int kind; adamic_value value; } reflected_value;

static const char *const pair_names[] = {"0","1"};
static const bool pair_scalar_references[]={true,false}, pair_reference_references[]={true,true};
static const adamic_shape pair_shapes[] = {
 {2,pair_names,pair_scalar_references,NULL}, {2,pair_names,pair_scalar_references,NULL},
 {2,pair_names,pair_reference_references,NULL}, {2,pair_names,pair_reference_references,NULL}
};
static const int pair_kinds[][2]={{3,1},{3,2},{3,3},{3,10}};
static const size_t pair_lengths[]={1,1};
static const adamic_reflection_layout pair_layouts[]={
 {pair_kinds[0],pair_lengths}, {pair_kinds[1],pair_lengths},
 {pair_kinds[2],pair_lengths}, {pair_kinds[3],pair_lengths}
};
static const adamic_reflection_layout *layout_of(const adamic_shape *shape, adamic_reflection_types types) {
 const adamic_reflection_layout *layout=types(shape);
 if (layout!=NULL) return layout;
 for (size_t index=0;index<4;index++) if (shape==&pair_shapes[index]) return &pair_layouts[index];
 return NULL;
}


static reflected_value classify(const adamic_object *object, size_t index, adamic_reflection_types types) {
 const adamic_reflection_layout *layout = layout_of(object->shape,types);
 const int *kinds = layout == NULL ? NULL : layout->kinds;
 reflected_value result = {0, object->slots[index]};
 if (kinds == NULL) return result;
 result.kind = kinds[index];
 if (result.kind == 7) {
  adamic_maybe_number number = adamic_maybe_number_unpack(result.value.number);
  result.kind = number.present ? 1 : 10;
  result.value.number = number.number;
 } else if (object->shape->references[index]) {
  const adamic_heap *reference = result.value.reference;
  if (reference == NULL) { result.kind = 10; return result; }
  if (reference->kind == adamic_kind_number) { result.kind = 1; result.value.number = ((const adamic_number_box *)reference)->number; }
  else if (reference->kind == adamic_kind_boolean) { result.kind = 2; result.value.boolean = ((const adamic_boolean_box *)reference)->boolean; }
  else if (reference->kind == adamic_kind_string) result.kind = 3;
  else result.kind = 0;
 }
 return result;
}

static bool member_matches(reflected_value value, const adamic_reflection_member *member) {
 if (value.kind != member->kind) return false;
 if (!member->literal) return true;
 if (value.kind == 1) return value.value.number == member->number;
 if (value.kind == 2) return value.value.boolean == member->boolean;
 if (value.kind == 3) {
  const adamic_string *text = value.value.reference;
  return text->length == member->length && memcmp(text->bytes, member->text, member->length) == 0;
 }
 return false;
}

static bool matches(reflected_value value, size_t count, const adamic_reflection_member *members) {
 for (size_t index=0; index<count; index++) if (member_matches(value, &members[index])) return true;
 return false;
}

static bool array_index(const char *name, uint32_t *value) {
 if (*name == '\0' || (name[0] == '0' && name[1] != '\0')) return false;
 uint64_t number=0;
 for (const char *at=name; *at!='\0'; at++) {
  if (*at<'0' || *at>'9') return false;
  number=number*10+(unsigned)(*at-'0');
  if (number>=UINT32_MAX) return false;
 }
 *value=(uint32_t)number;
 return true;
}

static int compare(const char *left, const char *right) {
 uint32_t a=0,b=0;
 bool ai=array_index(left,&a), bi=array_index(right,&b);
 if (ai!=bi) return ai ? -1 : 1;
 return ai ? (a<b ? -1 : a>b ? 1 : 0) : 0;
}

static size_t *indices(const adamic_object *object, adamic_reflection_types types, bool own_order) {
 size_t count=object->shape->count;
 const adamic_reflection_layout *layout=layout_of(object->shape,types);
 size_t *result=malloc((count==0 ? 1 : count)*sizeof *result);
 if (result==NULL) adamic_panic("out of memory", sizeof "out of memory"-1);
 for (size_t index=0;index<count;index++) {
  size_t place=index;
  while (own_order && place>0 && compare(layout != NULL && layout->lengths[index] != strlen(object->shape->names[index]) ? "x" : object->shape->names[index],layout != NULL && layout->lengths[result[place-1]] != strlen(object->shape->names[result[place-1]]) ? "x" : object->shape->names[result[place-1]])<0) {result[place]=result[place-1];place--;}
  result[place]=index;
 }
 return result;
}

static adamic_value converted(reflected_value value, int representation) {
 if (representation==1 || representation==2) return value.value;
 if (representation==7) return (adamic_value){.number=adamic_maybe_number_pack((adamic_maybe_number){value.kind==1,value.kind==1 ? value.value.number : 0})};
 if (representation==3) {adamic_retain(value.value.reference);return value.value;}
 adamic_heap *box=NULL;
 if (value.kind==1) box=adamic_box_number(value.value.number);
 else if (value.kind==2) box=(adamic_heap *)(value.value.boolean ? &adamic_box_true : &adamic_box_false);
 else if (value.kind==3) box=adamic_retain(value.value.reference);
 return (adamic_value){.reference=box};
}

adamic_array *adamic_checked_entries(const adamic_object *object, int element, size_t count, const adamic_reflection_member *members, adamic_reflection_types types, bool own_order, const char *message) {
 size_t *order=indices(object,types,own_order);
 adamic_array *result=adamic_array_new(object->shape->count,true);
 for (size_t at=0;at<object->shape->count;at++) {
  size_t index=order[at];
  const char *name=object->shape->names[index];
  if (object->class!=NULL && name[0]=='#') continue;
  reflected_value value=classify(object,index,types);
  // Never infer that this cannot fail from the apparent source type.
  if (!matches(value,count,members)) adamic_panic(message,strlen(message));
  const adamic_reflection_layout *layout=layout_of(object->shape,types);
  if (layout==NULL) adamic_panic(message,strlen(message));
  adamic_string *key=adamic_string_allocate(layout->lengths[index]);
  memcpy((char *)key->bytes,name,key->length);
  adamic_object *pair=adamic_object_new(&pair_shapes[element==1 ? 0 : element==2 ? 1 : element==3 ? 2 : 3]);
  pair->slots[0].reference=key;
  pair->slots[1]=converted(value,element);
  adamic_array_push(result,(adamic_value){.reference=pair});
 }
 free(order);
 return result;
}

static const adamic_reflection_field *field(size_t count, const adamic_reflection_field *fields, const char *name, size_t length) {
 for (size_t index=0;index<count;index++) if (!fields[index].index && fields[index].length==length && memcmp(fields[index].name,name,length)==0) return &fields[index];
 for (size_t index=0;index<count;index++) if (fields[index].index) return &fields[index];
 return NULL;
}

void adamic_checked_assign(adamic_object *target, const adamic_object *source, size_t source_count, const adamic_reflection_field *source_fields, size_t target_count, const adamic_reflection_field *target_fields, adamic_reflection_types types, bool own_order, const char *message) {
 size_t *order=indices(source,types,own_order);
 for (size_t at=0;at<source->shape->count;at++) {
  size_t index=order[at];const char *name=source->shape->names[index];
  if (source->class!=NULL && name[0]=='#') continue;
  const adamic_reflection_layout *source_layout=layout_of(source->shape,types);
  if (source_layout==NULL) adamic_panic(message,strlen(message));
  size_t length=source_layout->lengths[index];
  reflected_value value=classify(source,index,types);
  const adamic_reflection_field *from=field(source_count,source_fields,name,length), *into=field(target_count,target_fields,name,length);
  if (from==NULL || into==NULL || !matches(value,from->count,from->members) || !matches(value,into->count,into->members)) adamic_panic(message,strlen(message));
  const adamic_reflection_layout *target_layout=layout_of(target->shape,types);
  if (target_layout==NULL) adamic_panic(message,strlen(message));
  size_t position=target->shape->count;
  for (size_t next=0;next<target->shape->count;next++) if (target_layout->lengths[next]==length && memcmp(target->shape->names[next],name,length)==0) {position=next;break;}
  if (position==target->shape->count) adamic_panic(message,strlen(message));
  const int *kinds=target_layout->kinds;
  if (kinds==NULL) adamic_panic(message,strlen(message));
  if (kinds[position]!=value.kind && kinds[position]!=10 && !(kinds[position]==7 && value.kind==1)) adamic_panic(message,strlen(message));
  adamic_value next=converted(value,kinds[position]);
  adamic_object_check_write(target,name);
  if (target->shape->references[position]) adamic_release(target->slots[position].reference);
  target->slots[position]=next;
 }
 free(order);
}
