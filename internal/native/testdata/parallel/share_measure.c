#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static double now(void) {
 struct timespec time;
 clock_gettime(CLOCK_MONOTONIC, &time);
 return (double)time.tv_sec + (double)time.tv_nsec / 1e9;
}
static adamic_value work(adamic_closure *self, adamic_value *arguments) {
 (void)self;
 adamic_string *text = arguments[0].reference;
 double sum = 0;
 for (size_t i = 0; i < (size_t)adamic_string_length(text); i++) {
  sum += adamic_string_char_code_at(text, (double)i);
 }
 return (adamic_value){.number = sum};
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 adamic_array *items = adamic_array_new(4096, true);
 for (size_t i = 0; i < 4096; i++) {
  adamic_string *text = adamic_string_allocate(8192);
  for (size_t j = 0; j < text->length; j += 2) { memcpy((char *)text->bytes + j, "é", 2); }
  adamic_array_push(items, (adamic_value){.reference = text});
 }
 adamic_closure *callback = adamic_closure_new(work, 0);
 double start = now();
 adamic_share(items); adamic_share(callback);
 double shared = now();
 adamic_array *results = adamic_parallel_map(items, callback, false);
 double end = now();
 double sum = 0;
 for (size_t i = 0; i < results->length; i++) { sum += results->elements[i].number; }
 if (sum != 3909091328.0) { abort(); }
 printf("{\"share\":%.9f,\"map\":%.9f,\"total\":%.9f}\n", shared - start, end - shared, end - start);
 adamic_release(results); adamic_release(callback); adamic_release(items);
 return 0;
}
