#include "harness.h"
#include "tsgo_runtime.h"
static adamic_string path = ADAMIC_STRING("fixture.a");
static void prepare(void) {}
static void cleanup(void) { adamic_tsgo_release(1); }
static double exercise(size_t index) {
 (void)index;
 adamic_array *files = adamic_array_new(0, true);
 double handle = adamic_tsgo_program(&path, files);
 adamic_object *query = adamic_tsgo_query(handle, &path, 0);
 adamic_string *parts = adamic_tsgo_type_parts(handle, &path, 0, 1, &path);
 adamic_string *facts = adamic_tsgo_inspect(handle, &path, 0, 1, &path, &path);
 adamic_release(facts); adamic_release(parts); adamic_release(query); adamic_release(files);
 return 1;
}
