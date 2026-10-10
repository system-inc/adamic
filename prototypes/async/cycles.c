#include "async.h"
#include <stdio.h>
#include <string.h>
#include <pthread.h>

typedef struct { adamic_async_frame base; adamic_string *held; } Frame;
static void children(adamic_async_frame *base, void (*drop)(void *)) { drop(((Frame *)base)->held); }
static void resume(adamic_async_frame *frame, adamic_value value, bool rejected) {
    (void)value; (void)rejected;
    if (frame->state == 1) puts("resume");
}
static void *fixture(void *argument) {
    const char *mode = argument;
    adamic_async_promise *promise = adamic_async_new();
    bool fulfilled = strcmp(mode, "fulfilled") == 0;
    for (int observer = 0; observer < 2; observer++) {
        Frame *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);
        frame->base.state = fulfilled ? 1 : 0; frame->base.output = adamic_retain(promise); frame->base.waiting = NULL;
        frame->base.resume = resume; frame->base.children = children;
        frame->held = adamic_string_from_number(12345);
        /* The same Promise is output and waiting: both frame-owned edges must be released. */
        if (fulfilled) adamic_async_settle(promise, (adamic_value){.number = 1}, false, false);
        adamic_async_await(&frame->base, promise);
        adamic_release(frame); /* The promise's reaction now owns the suspended frame. */
    }
    if (fulfilled) puts("prefix");
    else if (strcmp(mode, "settle") == 0) adamic_async_settle(promise, (adamic_value){.number = 1}, false, false);
    else if (strcmp(mode, "cancel") == 0) adamic_async_cancel(promise);
    else if (strcmp(mode, "exit") != 0) adamic_unreachable();
    adamic_release(promise); /* Registry is the only external root. */
    adamic_async_run(); /* Microtasks first, then abandon never-settled subscriptions. */
    adamic_heap_thread_end();
    return NULL;
}
int main(int argc, char **argv) {
    if (argc != 2) return 2;
    /* Run the one loop thread to completion; joined stacks cannot conservatively root a leak. */
    pthread_t thread;
    if (pthread_create(&thread, NULL, fixture, argv[1]) != 0) return 2;
    if (pthread_join(thread, NULL) != 0) return 2;
    puts("cycle clean");
    return 0;
}
