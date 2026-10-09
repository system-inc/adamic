#include "fixture.h"
int main(void) {
    adamic_closure *roots[4096], *adapters[4096];
    for (size_t i=0;i<4096;i++) { roots[i]=adamic_closure_new(producer,0); adapters[i]=adamic_view_adapter_intern(roots[i], &view_a, make_adapter); }
    for (size_t i=0;i<4096;i+=2) { adamic_release(adapters[i]); adapters[i]=NULL; }
    for (size_t i=0;i<4096;i++) {
        adamic_closure *next=adamic_view_adapter_intern(roots[i], &view_a, make_adapter);
        if (adapters[i]!=NULL) { require(next==adapters[i], "resize/tombstone lost a live key"); adamic_release(next); }
        else { adapters[i]=next; }
    }
    for (size_t i=0;i<4096;i++) { adamic_release(adapters[i]); adamic_release(roots[i]); }
    require(adamic_counted.live==0, "distinct adapter entries leaked"); puts("distinct keys and tombstones balanced"); return 0;
}
