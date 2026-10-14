#include "fixture.h"
int main(void) {
    adamic_closure *root = adamic_closure_new(producer, 0);
    size_t before = adamic_counted.allocations;
    adamic_closure *first = adamic_view_adapter_intern(root, &view_a, make_adapter);
    adamic_closure *second = adamic_view_adapter_intern(root, &view_a, make_adapter);
    require(first == second && made == 1 && adamic_counted.allocations == before + 1, "interned twice must allocate once");
    require(root->heap.references == 2 && first->heap.references == 2, "cache must own neither entry nor key");
    require(adamic_view_adapter_underlying(first) == root && adamic_view_adapter_underlying(root) == root && adamic_view_adapter_underlying(NULL) == NULL, "underlying identity");
#ifdef ADAMIC_CANONICAL_CLOSURES
    require(first->canonical_owner == NULL, "adapter entered canonical cache");
#endif
    adamic_closure *again = adamic_view_adapter_intern(first, &view_a, make_adapter);
    require(again == first && made == 1, "same view stacked an adapter");
    adamic_closure *different = adamic_view_adapter_intern(first, &view_b, make_adapter);
    require(different != first && adamic_view_adapter_underlying(different) == root, "different view must also unwrap");
    adamic_value result = adamic_closure_call(different, (adamic_value[]){{.number = 7}}, 1);
    require(result.number == 7, "adapter must retain a callable underlying");
    adamic_release(first); adamic_release(second); adamic_release(again); adamic_release(different);
    require(root->heap.references == 1, "adapter destruction did not drop underlying");
    before = adamic_counted.allocations;
    first = adamic_view_adapter_intern(root, &view_a, make_adapter);
    require(made == 3 && adamic_counted.allocations == before + 1, "freed adapter left an entry");
    adamic_release(first); adamic_release(root);
    require(adamic_counted.live == 0, "cache retained dead values");
    puts("interned and removed");
    return 0;
}
