#include "fixture.h"
// This read-only inspection is appended to closure.c in the test snapshot.
bool view_adapter_listed_for_test(adamic_closure *adapter);
static bool armed, order_failed;
void view_adapter_order_probe(adamic_closure *adapter) {
    if (!armed) { return; }
    if (view_adapter_listed_for_test(adapter) || adapter->view_underlying->heap.references != 1) {
        order_failed = true;
    }
}
int main(void) {
    adamic_closure *root = adamic_closure_new(producer, 0);
    adamic_closure *adapter = adamic_view_adapter_intern(root, &view_a, make_adapter);
    adamic_release(root);
    armed = true;
    adamic_release(adapter);
    require(adamic_counted.live == 0, "release order leaked values");
    require(!order_failed, "remove/release order violated");
    puts("removed before underlying release");
    return 0;
}
