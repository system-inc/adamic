#include "adamic.h"
#include "parallel.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>

static pthread_mutex_t first_lock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t first_ready = PTHREAD_COND_INITIALIZER;
static size_t first_arrived;
static _Thread_local bool arrived;

static adamic_value read_characters(adamic_closure *self, adamic_value *arguments) {
 (void)self;
 if(adamic_parallel_threads()>1 && !arrived) {
  arrived=true;
  pthread_mutex_lock(&first_lock);
  first_arrived++;pthread_cond_broadcast(&first_ready);
  while(first_arrived<adamic_parallel_threads()) pthread_cond_wait(&first_ready,&first_lock);
  pthread_mutex_unlock(&first_lock);
 }
 adamic_string *parent=arguments[0].reference;
 size_t hash=0;
 for(size_t repeat=0;repeat<8;repeat++) {
  hash=0;
  for(size_t i=0;i<128;i++) {
   adamic_string *character=adamic_string_at(parent,(double)i);
   if(character==NULL || character->length!=1 || character->units!=2 ||
      adamic_reference_count(&character->heap)!=0) abort();
   hash=(hash*33+(unsigned char)character->bytes[0])%1000003;
   adamic_release(character);
  }
 }
 return (adamic_value){.number=(double)hash};
}
int main(void) {
 adamic_string *parent=adamic_string_allocate(128);
 for(size_t i=0;i<128;i++) ((char *)parent->bytes)[i]=(char)i;
 parent->units=129;
 adamic_array *items=adamic_array_new(128,true);
 for(size_t i=0;i<128;i++) adamic_array_push(items,(adamic_value){.reference=adamic_retain(parent)});
 adamic_release(parent);
 adamic_closure *work=adamic_closure_new(read_characters,0);
 adamic_array *results=adamic_parallel_map(items,work,false);
 for(size_t i=0;i<results->length;i++) printf("%.0f\n",results->elements[i].number);
 adamic_release(results);adamic_release(work);adamic_release(items);
 return 0;
}
