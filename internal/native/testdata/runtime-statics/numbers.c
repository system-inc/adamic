#include "harness.h"
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 char bytes[ADAMIC_NUMBER_FORMAT_MAX];
 size_t count = adamic_number_format(1.25, bytes);
 if (count != 4 || memcmp(bytes, "1.25", 4) != 0) { abort(); }
 adamic_string *fixed = adamic_number_to_fixed(1.25, 2);
 if (fixed->length != 4 || memcmp(fixed->bytes, "1.25", 4) != 0) { abort(); }
 adamic_release(fixed);
 return 1;
}
