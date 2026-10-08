// Only tests supply the runtime side of the encoder contract.
#include "json_stringify.h"
#include <stdio.h>
#include <string.h>
static const adamic_json_schema number = {.kind=adamic_json_number};
static const adamic_json_schema boolean = {.kind=adamic_json_boolean};
static const adamic_json_schema string = {.kind=adamic_json_string};
static const adamic_json_schema null_value = {.kind=adamic_json_null};
static const adamic_json_schema absent = {.kind=adamic_json_undefined};
static const adamic_json_schema union_value = {.kind=adamic_json_union};
static const adamic_json_schema array_number = {.kind=adamic_json_array, .element=&number};
static const adamic_json_schema array_dynamic = {.kind=adamic_json_array};
static const adamic_json_schema hook = {.kind=adamic_json_toJSON};
static adamic_string field_name = ADAMIC_STRING("field");
static adamic_string x_name = ADAMIC_STRING("x");
static const adamic_json_field x_fields[] = {{&x_name,0,&boolean}};
static const adamic_json_schema returned_hook = {.kind=adamic_json_toJSON,.count=1,.fields=x_fields};
static const adamic_json_schema object = {.kind=adamic_json_object,.count=1,.fields=x_fields};
static const char *const x_names[] = {"x"};
static const bool x_references[] = {false};
static const adamic_shape x_shape = {1,x_names,x_references,NULL};
static const char *const field_names[] = {"field"};
static const bool field_references[] = {true};
static const adamic_shape field_shape = {1,field_names,field_references,NULL};
static const adamic_json_field hook_fields[] = {{&field_name,0,&hook}};
static const adamic_json_schema wrapper = {.kind=adamic_json_object,.count=1,.fields=hook_fields};
static adamic_json_result elements[8];
static adamic_json_result returned;
static adamic_object *hook_receiver;
static const char *expected_key;
static bool throwing_hook;
static size_t hook_calls;
static const adamic_json_schema *describe(const adamic_heap *value) {
 if (value->kind==adamic_kind_array) return &array_number;
 if (value->kind==adamic_kind_object) return &object;
 return NULL;
}
static adamic_json_result element(const adamic_array *array, size_t index) {
 (void)array;
 return elements[index];
}
static adamic_json_result to_json(adamic_object *receiver,const adamic_string *key) {
 if (receiver!=hook_receiver || key->length!=strlen(expected_key) || memcmp(key->bytes,expected_key,key->length)!=0) { adamic_panic("wrong hook receiver or key",26); }
 hook_calls++;
 if (throwing_hook) {adamic_thrown=adamic_retain(receiver);return (adamic_json_result){{.reference=NULL},&absent};}
 if (returned.schema != NULL && (returned.schema->kind==adamic_json_string || returned.schema->kind==adamic_json_array || returned.schema->kind==adamic_json_object || returned.schema->kind==adamic_json_union || returned.schema->kind==adamic_json_toJSON)) adamic_retain(returned.value.reference);
 return returned;
}
static const adamic_json_runtime runtime = {describe,element,to_json};
static void print_json(adamic_json_result value,const adamic_json_runtime *provider) {
 adamic_value empty = {.reference=NULL};
 adamic_string *text=adamic_json_stringify_runtime(value.value,value.schema,empty,NULL,empty,NULL,provider);
 if(adamic_thrown!=NULL) {adamic_release(adamic_thrown);adamic_thrown=NULL;puts("thrown");return;}
 if (text==NULL) puts("undefined"); else { fwrite(text->bytes,1,text->length,stdout); puts(""); adamic_release(text); }
}
int main(int argc,char **argv) {
 adamic_array *nested=adamic_array_new(1,false);
 adamic_array_push(nested,(adamic_value){.number=2});
 adamic_object *child=adamic_object_new(&x_shape);
 child->slots[0].boolean=true;
 static adamic_string built=ADAMIC_STRING("built");
 adamic_string *word=adamic_string_repeat(&built,2);
 adamic_array *mixed=adamic_array_new(8,false);
 elements[0]=(adamic_json_result){{.number=7},&number};
 elements[1]=(adamic_json_result){{.reference=word},&string};
 elements[2]=(adamic_json_result){{.boolean=false},&boolean};
 elements[3]=(adamic_json_result){{.reference=NULL},&null_value};
 elements[4]=(adamic_json_result){{.reference=nested},&array_number};
 elements[5]=(adamic_json_result){{.reference=child},&object};
 elements[6]=(adamic_json_result){{.reference=NULL},&absent};
 elements[7]=(adamic_json_result){{.number=NAN},&number};
 for(size_t i=0;i<8;i++) adamic_array_push(mixed,elements[i].value);
 if(argc>1) {
  if(strcmp(argv[1],"empty_array")==0) {mixed->length=0;print_json((adamic_json_result){{.reference=mixed},&array_dynamic},NULL);}
  if(strcmp(argv[1],"cycle")==0) {mixed->length=1;elements[0]=(adamic_json_result){{.reference=mixed},&array_dynamic};print_json((adamic_json_result){{.reference=mixed},&array_dynamic},&runtime);}
  if(strcmp(argv[1],"array")==0) print_json((adamic_json_result){{.reference=mixed},&array_dynamic},NULL);
  if(strcmp(argv[1],"union")==0) print_json((adamic_json_result){{.reference=child},&union_value},NULL);
  if(strcmp(argv[1],"element")==0) {elements[0].schema=NULL;print_json((adamic_json_result){{.reference=mixed},&array_dynamic},&runtime);}
  if(strcmp(argv[1],"hook")==0) {returned=(adamic_json_result){{.reference=NULL},NULL};hook_receiver=child;expected_key="";print_json((adamic_json_result){{.reference=child},&hook},&runtime);}
  return 0;
 }
 print_json((adamic_json_result){{.reference=mixed},&array_dynamic},&runtime);
 adamic_json_result returns[]={elements[3],elements[6],elements[2],{{.number=3},&number},elements[1],elements[4],elements[5]};
 hook_receiver=child;
 for(size_t i=0;i<7;i++) {
  returned=returns[i];expected_key="";print_json((adamic_json_result){{.reference=child},&hook},&runtime);
  adamic_object *holder=adamic_object_new(&field_shape);holder->slots[0].reference=adamic_retain(child);
  expected_key="field";print_json((adamic_json_result){{.reference=holder},&wrapper},&runtime);adamic_release(holder);
  elements[0]=(adamic_json_result){{.reference=child},&hook};mixed->length=1;
  expected_key="0";print_json((adamic_json_result){{.reference=mixed},&array_dynamic},&runtime);
 }
 adamic_heap *boxed=adamic_box_number(7);
 adamic_json_result unions[]={{{.reference=boxed},&union_value},{{.reference=&adamic_box_false},&union_value},{{.reference=word},&union_value},{{.reference=&adamic_null},&union_value},{{.reference=NULL},&union_value},{{.reference=nested},&union_value},{{.reference=child},&union_value}};
 for(size_t i=0;i<7;i++) print_json(unions[i],&runtime);
 for(size_t i=0;i<2;i++) {
  throwing_hook=i!=0;returned=(adamic_json_result){{.reference=child},&returned_hook};
  expected_key="";print_json((adamic_json_result){{.reference=child},&hook},&runtime);
  adamic_object *holder=adamic_object_new(&field_shape);holder->slots[0].reference=adamic_retain(child);
  expected_key="field";print_json((adamic_json_result){{.reference=holder},&wrapper},&runtime);adamic_release(holder);
  elements[0]=(adamic_json_result){{.reference=child},&hook};mixed->length=1;
  expected_key="0";print_json((adamic_json_result){{.reference=mixed},&array_dynamic},&runtime);
 }
 if(hook_calls!=27) {adamic_panic("hook invoked more than once",27);}
 adamic_release(boxed);adamic_release(mixed);adamic_release(word);adamic_release(nested);adamic_release(child);
 return 0;
}
