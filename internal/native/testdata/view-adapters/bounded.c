#include "fixture.h"
int main(void) {
    adamic_closure *root = adamic_closure_new(producer, 0);
    adamic_closure *held = adamic_view_adapter_intern(root, &view_a, make_adapter);
    size_t allocations = adamic_counted.allocations;
    size_t retains = adamic_counted.retains;
    size_t releases = adamic_counted.releases;
    for (size_t i = 0; i < 100000; i++) {
        adamic_closure *next = adamic_view_adapter_intern(held, &view_a, make_adapter);
        adamic_release(next);
    }
    bool bounded = made == 1 && adamic_counted.allocations == allocations && root->heap.references == 2 && held->heap.references == 1;
    bool work = adamic_counted.retains - retains == 100000 && adamic_counted.releases - releases == 100000;
    adamic_release(held); adamic_release(root);
    require(adamic_counted.live == 0, "loop leaked values");
    require(bounded, "bounded allocation exceeded: repeated reads rebuilt adapters");
    require(work, "each repeated read must retain and release exactly one adapter");
    puts("bounded allocations and retains");
    return 0;
}
