#include "fixture.h"
int main(void) {
    adamic_closure *values[1000];
    for (size_t i=0; i<1000; i++) { values[i]=adamic_closure_new(producer, 0); require(values[i]->view == NULL, "ordinary closure has adapter metadata"); }
    size_t bytes=1000*sizeof *values[0];
    bool fits=bytes==40000 && adamic_counted.peak==1000;
    for (size_t i=0; i<1000; i++) { adamic_release(values[i]); }
    require(fits, "ordinary closure exceeds 40 bytes"); require(adamic_counted.live==0, "closure fixture leaked");
    printf("closure bytes %zu peak 1000\n",bytes); return 0;
}
