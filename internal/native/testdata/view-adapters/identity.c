#include "fixture.h"
int main(void) {
    adamic_closure *array = adamic_library_identity(0);
    require(adamic_view_adapter_underlying(array) == array, "Array identity changed");
    require(adamic_view_adapter_intern(array, &view_a, make_adapter) == array && made == 0, "immortal identity adapted");
    require(adamic_view_adapter_underlying(array) == adamic_view_adapter_underlying(array), "Array strict identity changed");
    // Last constructor has no following header to conceal the old overread.
    adamic_closure *last = adamic_library_identity(6);
    require(adamic_view_adapter_underlying(last) == last, "last constructor identity changed");
    puts("immortal identity untouched"); return 0;
}
