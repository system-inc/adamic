#include "fixture.h"
static adamic_closure *winner;
static adamic_closure *reentrant_make(adamic_closure *underlying, const void *view_key) {
    // If make runs under the nonrecursive mutex, this nested lookup cannot finish.
    winner = adamic_view_adapter_intern(underlying, view_key, make_adapter);
    return make_adapter(underlying, view_key);
}
int main(void) {
    adamic_closure *root = adamic_closure_new(producer, 0);
    adamic_closure *result = adamic_view_adapter_intern(root, &view_a, reentrant_make);
    bool rechecked = result == winner && made == 2 && root->heap.references == 2 && adamic_counted.live == 2;
    adamic_release(result); adamic_release(winner); adamic_release(root);
    require(adamic_counted.live == 0, "losing factory result leaked");
    require(rechecked, "factory winner was not rechecked");
    puts("factory outside lock and winner rechecked");
    return 0;
}
