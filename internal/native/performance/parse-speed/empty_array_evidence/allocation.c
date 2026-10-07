#include "adamic.h"
#include "count.h"
#include <stdio.h>
#include <stdlib.h>
static bool observing;
static size_t malloc_calls, realloc_calls;
void *__real_malloc(size_t size);
void *__real_realloc(void *pointer, size_t size);
void *__wrap_malloc(size_t size) { if (observing) malloc_calls++; return __real_malloc(size); }
void *__wrap_realloc(void *pointer, size_t size) { if (observing) realloc_calls++; return __real_realloc(pointer, size); }
int main(void) {
 observing = true;
 adamic_array *first = adamic_array_new(0, false);
 observing = false;
 if (malloc_calls != 1 || realloc_calls != 0 || adamic_counted.allocations != 1 || first->elements != NULL || first->capacity != 0) { fprintf(stderr,"empty array allocation check failed\n"); adamic_release(first); return 1; }
 adamic_array *second = adamic_array_new(0, false);
 if (first == second) { fprintf(stderr,"empty headers shared\n"); return 1; }
 observing = true;
 adamic_array_push(first, (adamic_value){.number=7});
 observing = false;
 if (realloc_calls != 1 || first->elements == NULL || first->length != 1 || first->elements[0].number != 7) { fprintf(stderr,"first push check failed\n"); return 1; }
 printf("empty: one header allocation, one malloc, zero buffer allocations; distinct headers\nfirst push: one realloc; header size %zu bytes\n",sizeof *first);
 adamic_release(first); adamic_release(second);
 if (adamic_counted.live != 0) { fprintf(stderr,"live array leak\n"); return 1; }
 return 0;
}
