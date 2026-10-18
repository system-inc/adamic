#include "fixture.h"
__attribute__((noinline)) static void abandon_adapter(void) {
    adamic_closure *root = adamic_closure_new(producer, 0);
    (void)adamic_view_adapter_intern(root, &view_a, make_adapter);
    adamic_release(root);
    // Intentionally lose the adapter's owned reference. A weak cache must not
    // make this leak appear reachable to LeakSanitizer.
}
int main(void) {
    abandon_adapter();
    return 0;
}
