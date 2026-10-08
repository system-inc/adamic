#include "harness.h"
static adamic_string decomposed = ADAMIC_STRING("é");
static adamic_string composed = ADAMIC_STRING("é");
// U+0065 and U+00A5 with U+0301 collide in the composition cache. Keep it
// changing after the first lookup so a global-cache mutant must keep racing.
static adamic_string collision = ADAMIC_STRING("¥́");
static adamic_string nfc = ADAMIC_STRING("NFC");
static adamic_string nfd = ADAMIC_STRING("NFD");
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_string *first = adamic_string_normalize(&decomposed, &nfc);
 adamic_string *second = adamic_string_normalize(&composed, &nfd);
 adamic_string *third = adamic_string_normalize(&collision, &nfc);
 if (!adamic_string_equal(first, &composed) || !adamic_string_equal(second, &decomposed) ||
     !adamic_string_equal(third, &collision)) { abort(); }
 adamic_release(first); adamic_release(second); adamic_release(third);
 return 1;
}
