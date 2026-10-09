#include "harness.h"
static adamic_string text = ADAMIC_STRING("a");
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_string *letter = adamic_string_at(&text, 0);
 if (letter == NULL || letter->length != 1 || letter->bytes[0] != 'a' || adamic_string_length(letter) != 1) abort();
 adamic_release(letter);
 return 1;
}
