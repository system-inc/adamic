#include "fixture.h"
static adamic_environment *frame;
static adamic_closure *canonical_make(adamic_closure *underlying, const void *key) {
    (void)underlying; (void)key;
    return adamic_closure_canonical(&frame->cells[0], invoke, 0, NULL);
}
int main(int count, char **arguments) {
    (void)arguments;
    frame = adamic_environment_new(1);
    adamic_closure *root = adamic_closure_canonical(&frame->cells[0], producer, 1, (adamic_cell *const[]){&frame->cells[0]});
    if (count > 1) {
        // A canonical factory result violates the API's fresh-result contract.
        adamic_closure *wrong = adamic_view_adapter_intern(root, &view_a, canonical_make);
        (void)wrong;
        puts("canonical result admitted");
        fflush(stdout);
        // Observe admission before an invalid canonical/adapter combination is destroyed.
        _Exit(0);
    }
    adamic_closure *adapter = adamic_view_adapter_intern(root, &view_a, make_adapter);
    require(adapter->canonical_owner == NULL && frame->functions == root && root->canonical_next == NULL, "adapter entered the canonical cache");
    adamic_release(adapter);
    require(frame->functions == root, "adapter release altered canonical cache");
    adamic_release(root);
    require(frame->functions == NULL && frame->heap.references == 1, "canonical underlying lifetime changed");
    adamic_release(frame);
    require(adamic_counted.live == 0, "canonical underlying leaked");
    puts("adapter remains noncanonical");
    return 0;
}
