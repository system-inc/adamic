#include "harness.h"
static adamic_string decomposed = ADAMIC_STRING("é");
static adamic_string composed = ADAMIC_STRING("é");
static adamic_string nfc = ADAMIC_STRING("NFC");
static adamic_string nfd = ADAMIC_STRING("NFD");
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_string *first = adamic_string_normalize(&decomposed, &nfc);
 adamic_string *second = adamic_string_normalize(&composed, &nfd);
 if (!adamic_string_equal(first, &composed) || !adamic_string_equal(second, &decomposed)) { abort(); }
 adamic_release(first); adamic_release(second);
 return 1;
}
