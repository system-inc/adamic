#define _POSIX_C_SOURCE 200809L
#include "async.h"
#include <assert.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static const char body[] = "fetch \xe4\xb8\x96\xe7\x95\x8c\0tail";
static pthread_t loop_thread;
static unsigned resumed;
static bool expect_rejected;
static const char *expect_message;
static bool checkpoint;
static void delay(void) {
    struct timespec duration = {.tv_nsec = 50000000};
    nanosleep(&duration, NULL);
}
typedef struct { adamic_async_frame base; adamic_string *held; } Frame;
static void children(adamic_async_frame *base, void (*drop)(void *)) {
    drop(((Frame *)base)->held);
}
static void resume(adamic_async_frame *base, adamic_value value, bool rejected) {
    assert(pthread_equal(loop_thread, pthread_self()));
    assert(base->waiting == NULL);
    assert(rejected == expect_rejected);
    Frame *frame = (Frame *)base;
    assert(frame->held->length != 0);
    adamic_object *object = value.reference;
    if (rejected) {
        adamic_string *message = object->slots[1].reference;
        assert(message->length == strlen(expect_message));
        assert(memcmp(message->bytes, expect_message, message->length) == 0);
    } else {
        adamic_string *text = object->slots[0].reference;
        assert(text->length == sizeof body - 1);
        assert(memcmp(text->bytes, body, text->length) == 0);
        assert(object->slots[1].number == 201.5);
    }
    puts(rejected ? "rejected" : "fulfilled");
    resumed++;
    if (checkpoint) {
        /* The next completion must still be pending until this checkpoint. */
        assert(!frame->base.output->settled);
        checkpoint = false;
    }
}
static void observe(adamic_promise *promise, adamic_promise *output) {
    Frame *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);
    frame->base.state = 0;
    frame->base.output = adamic_retain(output);
    frame->base.waiting = NULL;
    frame->base.children = children;
    frame->base.resume = resume;
    frame->held = adamic_string_from_number(12345);
    adamic_async_await(&frame->base, promise);
    adamic_release(frame);
}
typedef struct {
    adamic_host_request *request;
    const char *mode;
    _Atomic bool done;
} Work;
static void *complete(void *argument) {
    Work *work = argument;
    delay();
    if ((strcmp(work->mode, "abandon") == 0 || strcmp(work->mode, "cancel-abandon") == 0)) adamic_host_abandon(work->request);
    else if (strcmp(work->mode, "reject") == 0) {
        char message[] = "network failed";
        assert(adamic_host_reject(work->request, message));
        memset(message, 'x', sizeof message - 1);
    } else {
        char bytes[sizeof body];
        memcpy(bytes, body, sizeof bytes);
        assert(adamic_host_resolve(work->request, bytes, sizeof bytes - 1, 201.5));
        memset(bytes, 'x', sizeof bytes);
    }
    atomic_store(&work->done, true);
    return NULL;
}
static pthread_mutex_t hook_mutex = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t hook_condition = PTHREAD_COND_INITIALIZER;
static unsigned wakes, waits;
static void wake(void) {
    assert(pthread_mutex_lock(&hook_mutex) == 0);
    wakes++;
    assert(pthread_cond_signal(&hook_condition) == 0);
    assert(pthread_mutex_unlock(&hook_mutex) == 0);
}
static void wait(void) {
    assert(pthread_equal(loop_thread, pthread_self()));
    assert(pthread_mutex_lock(&hook_mutex) == 0);
    while (wakes == 0) assert(pthread_cond_wait(&hook_condition, &hook_mutex) == 0);
    wakes--;
    waits++;
    assert(pthread_mutex_unlock(&hook_mutex) == 0);
}
#define WORKERS 16
static void stress(void) {
    Work work[WORKERS];
    pthread_t workers[WORKERS];
    for (unsigned index = 0; index < WORKERS; index++) {
        adamic_promise *promise;
        work[index] = (Work){.request = adamic_host_promise_new(&promise), .mode = "resolve"};
        observe(promise, promise);
        adamic_release(promise);
        assert(pthread_create(&workers[index], NULL, complete, &work[index]) == 0);
    }
    /* Drain concurrently with producers, so removing the queue lock races. */
    for (;;) {
        adamic_host_process_completions();
        bool done = true;
        for (unsigned index = 0; index < WORKERS; index++) done &= atomic_load(&work[index].done);
        if (done) break;
    }
    adamic_async_run();
    for (unsigned index = 0; index < WORKERS; index++) assert(pthread_join(workers[index], NULL) == 0);
    assert(resumed == WORKERS);
}
static void *fixture(void *argument) {
    const char *mode = argument;
    loop_thread = pthread_self();
    puts("prefix");
    if (strcmp(mode, "stress") == 0) { stress(); adamic_heap_thread_end(); return NULL; }
    if (strcmp(mode, "checkpoint") == 0) {
        adamic_promise *first, *second;
        adamic_host_request *one = adamic_host_promise_new(&first);
        adamic_host_request *two = adamic_host_promise_new(&second);
        observe(first, second);
        observe(second, second);
        checkpoint = true;
        assert(adamic_host_resolve(one, body, sizeof body - 1, 201.5));
        assert(adamic_host_resolve(two, body, sizeof body - 1, 201.5));
        adamic_release(first); adamic_release(second);
        adamic_async_run();
        assert(resumed == 2 && !checkpoint);
        adamic_heap_thread_end();
        return NULL;
    }
    if (strcmp(mode, "stale") == 0) {
        adamic_promise *retired, *live;
        adamic_host_request *old = adamic_host_promise_new(&retired);
        adamic_release(retired);
        adamic_host_abandon(old);
        adamic_host_process_completions();
        adamic_host_request *current = adamic_host_promise_new(&live);
        assert(current != old);
        assert(!adamic_host_resolve(old, body, sizeof body - 1, 201.5));
        assert(!live->settled);
        observe(live, live); observe(live, live);
        assert(adamic_host_resolve(current, body, sizeof body - 1, 201.5));
        adamic_release(live);
        adamic_async_run();
        assert(resumed == 2);
        adamic_host_process_completions(); /* Empty, including after teardown. */
        adamic_heap_thread_end();
        return NULL;
    }
    bool hooks = strcmp(mode, "hooks") == 0 || strcmp(mode, "app") == 0;
    if (hooks) {
        assert(adamic_host_set_loop_hooks(wake, wait));
        assert(!adamic_host_set_loop_hooks(wake, wait));
    }
    adamic_promise *promise;
    adamic_host_request *request = adamic_host_promise_new(&promise);
    assert(!adamic_host_set_loop_hooks(NULL, NULL));
    bool cancel = strcmp(mode, "cancel") == 0 || strcmp(mode, "cancel-abandon") == 0;
    bool exit_pending = strcmp(mode, "exit") == 0;
    bool abandoned = strcmp(mode, "abandon") == 0 || strcmp(mode, "cancel-abandon") == 0;
    expect_rejected = strcmp(mode, "reject") == 0 || strcmp(mode, "invalid") == 0;
    expect_message = strcmp(mode, "invalid") == 0 ? "host payload is not valid UTF-8" : "network failed";
    /* Two observers; output and waiting both point to the host promise. */
    observe(promise, promise); observe(promise, promise);
    if (cancel) adamic_async_cancel(promise);
    Work work = {.request = request, .mode = mode};
    pthread_t worker;
    bool threaded = strcmp(mode, "early") != 0 && strcmp(mode, "double") != 0 &&
                    strcmp(mode, "invalid") != 0 && !exit_pending;
    if (threaded) assert(pthread_create(&worker, NULL, complete, &work) == 0);
    if (strcmp(mode, "early") == 0 || strcmp(mode, "double") == 0)
        assert(adamic_host_resolve(request, body, sizeof body - 1, 201.5));
    if (strcmp(mode, "double") == 0) assert(!adamic_host_reject(request, "second"));
    if (strcmp(mode, "invalid") == 0) {
        const unsigned char bytes[] = {0xed, 0xa0, 0x80}; /* A surrogate is not UTF-8. */
        assert(adamic_host_resolve(request, bytes, sizeof bytes, 201.5));
    }
    if (strcmp(mode, "direct") == 0) {
        /* Read the promise and its plain count while forbidden worker-side
         * settlement retains it. Volatile source loads prevent hoisting; they
         * add no synchronization. The control worker never touches the value. */
        while (!atomic_load(&work.done)) {
            bool settled = *(volatile bool *)&promise->settled;
            size_t references = *(volatile size_t *)&promise->heap.references;
            (void)settled; (void)references;
        }
    }
    adamic_release(promise);
    if (exit_pending) adamic_async_teardown();
    else if (strcmp(mode, "app") == 0) {
        /* Simulated AppKit: wait hook is not entered. Scheduled callbacks drain. */
        while (!atomic_load(&work.done)) { adamic_host_process_completions(); }
        adamic_host_process_completions();
        assert(waits == 0);
        adamic_async_teardown();
    } else adamic_async_run();
    if (threaded) assert(pthread_join(worker, NULL) == 0);
    if (hooks && strcmp(mode, "hooks") == 0) assert(waits > 0);
#ifndef ADAMIC_ASYNC_SETTLE_CYCLE_MUTANT
    assert(resumed == (cancel || abandoned || exit_pending ? 0u : 2u));
#endif
    if (strcmp(mode, "double") == 0 || abandoned || exit_pending)
        assert(!adamic_host_resolve(request, body, sizeof body - 1, 201.5));
    adamic_heap_thread_end();
    return NULL;
}
int main(int argc, char **argv) {
    if (argc != 2) return 2;
    /* A joined loop stack cannot conservatively hide leaked cycle roots. */
    pthread_t loop;
    assert(pthread_create(&loop, NULL, fixture, argv[1]) == 0);
    assert(pthread_join(loop, NULL) == 0);
    printf("%s clean\n", argv[1]);
    return 0;
}
