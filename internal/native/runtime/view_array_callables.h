#ifndef ADAMIC_VIEW_ARRAY_CALLABLES_H
#define ADAMIC_VIEW_ARRAY_CALLABLES_H
#include "view_callables_contract.h"

/* A void callback with fewer parameters ignores extra JavaScript arguments.
 * Both producer and caller use the existing closure argument-vector ABI. */
static inline const adamic_heap *adamic_array_callable_shape(const adamic_heap *value,const adamic_callable_signature *recorded,const adamic_callable_signature *expected,const char *expression,bool optional){
 if(recorded!=NULL && expected!=NULL && expected->result==254){
  adamic_callable_signature discarded=*expected;
  if(recorded->arity<expected->arity)discarded.arity=recorded->arity;
  discarded.result=255;
  return adamic_view_callable_shape(value,recorded,&discarded,expression,optional);
 }
 if(recorded!=NULL && expected!=NULL && expected->result!=255 && recorded->result!=0 && expected->result!=0 && recorded->result!=expected->result){
  size_t capacity=strlen(expression)+strlen(expected->name)+100;
  char *message=malloc(capacity);
  if(message==NULL){static const char oom[]="out of memory";adamic_panic(oom,sizeof oom-1);}
  int length=snprintf(message,capacity,"field read failed: %s expected %s, found function with incompatible result representation",expression,expected->name);
  adamic_panic(message,(size_t)length);
 }

 return adamic_view_callable_shape(value,recorded,expected,expression,optional);
}
#endif
