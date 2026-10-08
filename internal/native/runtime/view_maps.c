#include "view_maps.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
static _Noreturn void failure(const adamic_map *map,const char *expression,const char *expected){
 const char *found=map->contract_name;size_t size=strlen(expression)+strlen(expected)+strlen(found)+100;char *message=malloc(size);
 if(message==NULL){static const char oom[]="out of memory";adamic_panic(oom,sizeof oom-1);}
 int length=snprintf(message,size,"Map contract failed: %s; expected %s, found %s",expression,expected,found);adamic_panic(message,(size_t)length);
}
void adamic_map_view_certificate(const adamic_map *map,const unsigned int *pairs,size_t count,const char *expression,const char *expected){
 for(size_t i=0;i<count;i++)if(map->key_contract!=0&&map->value_contract!=0&&map->key_contract==pairs[2*i]&&map->value_contract==pairs[2*i+1])return;
 failure(map,expression,expected);
}
