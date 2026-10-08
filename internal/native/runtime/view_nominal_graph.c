#include "view_nominal_graph.h"
#include "view_nullish.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

typedef struct {void *value;unsigned int schema;} graph_item;
static void graph_grow(graph_item **items,size_t *capacity) {
 if(*capacity>SIZE_MAX/(2*sizeof **items)){static const char message[]="nominal producer witness exhausted";adamic_panic(message,sizeof message-1);}
 size_t next=*capacity==0?16:2*(*capacity);
 graph_item *grown=realloc(*items,next*sizeof *grown);
 if(grown==NULL){static const char message[]="out of memory";adamic_panic(message,sizeof message-1);}
 *items=grown;*capacity=next;
}
static _Noreturn void graph_failure(const char *where,const char *expected) {
 size_t capacity=strlen(where)+strlen(expected)+96;
 char *message=malloc(capacity);
 if(message==NULL){static const char oom[]="out of memory";adamic_panic(oom,sizeof oom-1);}
 int length=snprintf(message,capacity,"Map nominal producer failed: %s expected %s, found value without its class identity",where,expected);
 adamic_panic(message,(size_t)length);
}
// A work item owns its snapshot. Visited keys include the schema as well as the
// address: the same object reached with another obligation must be checked again.
void adamic_nominal_graph_check(const void *value,unsigned int root,const adamic_nominal_graph_schema *schemas,const char *where) {
 graph_item *work=NULL,*seen=NULL;
 size_t work_count=0,work_capacity=0,seen_count=0,seen_capacity=0;
 graph_grow(&work,&work_capacity);
 work[work_count++]=(graph_item){adamic_retain((void *)value),root};
 while(work_count!=0) {
  graph_item item=work[--work_count];
  const adamic_nominal_graph_schema *schema=&schemas[item.schema];
  bool null=item.value==&adamic_null,undefined=item.value==NULL;
  if(null||undefined){if((null&&!schema->null_allowed)||(undefined&&!schema->undefined_allowed))graph_failure(where,schema->name);adamic_release(item.value);continue;}
  const adamic_heap *heap=item.value;
  if(heap->kind!=adamic_kind_object||(schema->nominal!=NULL&&!adamic_instanceof(item.value,schema->nominal)))graph_failure(where,schema->name);
  bool visited=false;
  for(size_t i=0;i<seen_count;i++){if(seen[i].value==item.value&&seen[i].schema==item.schema){visited=true;break;}}
  if(visited){adamic_release(item.value);continue;}
  if(seen_count==seen_capacity)graph_grow(&seen,&seen_capacity);
  seen[seen_count++]=item;
  for(size_t i=0;i<schema->field_count;i++) {
   const adamic_nominal_graph_field *field=&schema->fields[i];
   const adamic_nominal_graph_schema *child=&schemas[field->schema];
   adamic_slot_cache cache={0};
   adamic_heap *snapshot=adamic_object_nullish_view((const adamic_object *)item.value,field->name,&cache,1u<<4,child->null_allowed,child->undefined_allowed||field->optional,field->optional,false,field->name,child->name);
   if(work_count==work_capacity)graph_grow(&work,&work_capacity);
   work[work_count++]=(graph_item){snapshot,field->schema};
  }
  adamic_release(item.value);
 }
 free(work);free(seen);
}
