#include <pthread.h>
#include <stdio.h>
static int entered;
static void *worker(void *given) { (void)given; entered = 1; return NULL; }
int main(void) { pthread_t thread; int result = pthread_create(&thread, NULL, worker, NULL); printf("pthread_create=%d entered=%d\n", result, entered); return 0; }
