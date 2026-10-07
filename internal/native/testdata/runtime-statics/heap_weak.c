#include "harness.h"
static const adamic_shape shape = {0, NULL, NULL, NULL};
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_object *target = adamic_object_new(&shape);
 adamic_weak *weak = adamic_weak_of(target);
 if (adamic_weak_target(weak) != target || !adamic_weak_held(target)) { abort(); }
 adamic_release(target);
 if (adamic_weak_target(weak) != NULL) { abort(); }
 adamic_release(weak);
 return 1;
}
