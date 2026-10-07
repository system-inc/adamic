#include "harness.h"
static const adamic_regex_range letters[] = {{'k', 'k'}};
static const adamic_regex_instruction code[] = {
 {.op = 1, .direction = 1, .flags = 5, .ranges = letters, .range_count = 1}, {.op = 0}
};
static const adamic_regex_program program = {.code = code, .flags = 5};
static adamic_string source = ADAMIC_STRING("k");
static adamic_string flags = ADAMIC_STRING("iu");
static adamic_string input = ADAMIC_STRING("K");
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_regex_set_step_limit(UINT64_MAX);
 adamic_object *regex = adamic_regex_new(&program, &source, &flags);
 bool matched = adamic_regex_test(regex, &input);
 adamic_release(regex);
 if (!matched) { abort(); }
 return 1;
}
