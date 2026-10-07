#include "harness.h"
static adamic_string line = ADAMIC_STRING("worker line");
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_write_line(adamic_stdout, &line);
 return 1;
}
