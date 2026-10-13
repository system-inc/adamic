#include "harness.h"
static adamic_string decomposed = ADAMIC_STRING("é");
static adamic_string composed = ADAMIC_STRING("é");
// (e, acute) and (yen, acute) collide in the 64-slot pair cache. Alternating
// them keeps a shared-cache mutant writing after its first fill, even under load.
static adamic_string uncomposed = ADAMIC_STRING("¥́");
// U+0341 decomposes to U+0301; both collide in the class and mapping caches.
// Exercise those misses repeatedly too, rather than relying on cold initialization.
static adamic_string tone_mark = ADAMIC_STRING("é");
static adamic_string nfc = ADAMIC_STRING("NFC");
static adamic_string nfd = ADAMIC_STRING("NFD");
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_string *first = adamic_string_normalize(&decomposed, &nfc);
 adamic_string *second = adamic_string_normalize(&composed, &nfd);
 adamic_string *third = adamic_string_normalize(&uncomposed, &nfc);
 adamic_string *fourth = adamic_string_normalize(&tone_mark, &nfc);
 if (!adamic_string_equal(first, &composed) || !adamic_string_equal(second, &decomposed) || !adamic_string_equal(third, &uncomposed) || !adamic_string_equal(fourth, &composed)) { abort(); }
 adamic_release(first); adamic_release(second); adamic_release(third); adamic_release(fourth);
 return 1;
}
