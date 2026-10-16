#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static adamic_string *text;
static pthread_mutex_t gate = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t ready = PTHREAD_COND_INITIALIZER;
static size_t built;

// Only snapshot builds call this hook, after each candidate is fully built and
// before any candidate can be published. Four builders guarantee three losers.
void adamic_test_cache_built(void) {
 pthread_mutex_lock(&gate);
 built++;
 pthread_cond_broadcast(&ready);
 while (built != 4) { pthread_cond_wait(&ready, &gate); }
 pthread_mutex_unlock(&gate);
}
static void *read_text(void *unused) {
 (void)unused;
 if (adamic_string_length(text) != 4194304 ||
     adamic_string_char_code_at(text, 1000000) != 233 ||
     adamic_string_code_point_at(text, 1000000).number != 233) { abort(); }
 return NULL;
}
int main(void) {
 text = adamic_string_allocate((size_t)8 << 20);
 for (size_t i = 0; i < text->length; i += 2) { memcpy((char *)text->bytes + i, "é", 2); }
 adamic_share(text);
 if (text->index != NULL) { abort(); }
 pthread_t readers[4];
 for (size_t i = 0; i < 4; i++) { if (pthread_create(&readers[i], NULL, read_text, NULL) != 0) { abort(); } }
 for (size_t i = 0; i < 4; i++) { pthread_join(readers[i], NULL); }
 if (built != 4 || text->index->cursor_unit != 0) { abort(); }
 adamic_release(text);
 puts("cache publication clean");
 return 0;
}
