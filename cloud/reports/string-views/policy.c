#include "adamic.h"
#include <stdio.h>
#include <string.h>
int main(void) {
 const size_t capacities[]={128,256,512,1024,8192,65536};
 const size_t sizes[]={1,4,8,16,32,64,128};
 size_t views=0,copies=0,copied=0,pinned=0;
 for(size_t i=0;i<sizeof capacities/sizeof *capacities;i++) {
  size_t capacity=capacities[i];
  adamic_string *parent=adamic_string_allocate(capacity);
  memset((char *)parent->bytes,'a',capacity);parent->units=capacity+1;
  for(size_t j=0;j<sizeof sizes/sizeof *sizes;j++) {
   size_t size=sizes[j];if(size>=capacity)continue;
   adamic_string *view=adamic_string_slice(parent,1,(double)(1+size),true);
   if(view->owner) {views++;pinned+=sizeof *parent+capacity;} else {copies++;copied+=size;}
   adamic_release(view);
  }
  adamic_release(parent);
 }
 printf("{\"fraction\":%d,\"views\":%zu,\"copies\":%zu,\"copied_bytes\":%zu,\"isolated_pinned_bytes\":%zu}\n",SHARE_FRACTION,views,copies,copied,pinned);
}
