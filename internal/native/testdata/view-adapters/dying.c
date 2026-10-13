#include "fixture.h"
static bool armed;
static adamic_closure *rescued;
// The test snapshot calls this at the actual free path's count-zero boundary,
// before removal. It deterministically exercises the dying-but-listed window.
void view_adapter_dying_probe(adamic_closure *dying) {
    if (!armed) { return; }
    armed = false;
    require(dying->heap.references == 0, "probe must observe a dying adapter");
    rescued = adamic_view_adapter_intern(dying->view->underlying, dying->view->key, make_adapter);
}
int main(void) {
    adamic_closure *root = adamic_closure_new(producer, 0);
    for (size_t i = 0; i < 10000; i++) {
        adamic_closure *adapter = adamic_view_adapter_intern(root, &view_a, make_adapter);
        armed = true;
        adamic_release(adapter);
        adamic_value result = adamic_closure_call(rescued, (adamic_value[]){{.number = (double)i}}, 1);
        require(result.number == (double)i, "re-viewed dying callable corrupted");
        adamic_release(rescued);
        rescued = NULL;
    }
    require(made == 20000 && root->heap.references == 1, "dying lookup must miss and balance underlying references");
    adamic_release(root);
    require(adamic_counted.live == 0, "dying loop leaked values");
    puts("dying adapters not resurrected");
    return 0;
}
