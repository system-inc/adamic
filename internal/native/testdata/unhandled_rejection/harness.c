#if defined(__linux__)
#define _POSIX_C_SOURCE 200809L
#endif
#include "timers.h"
#include <assert.h>
#include <errno.h>
#include <string.h>
#include <time.h>

static void children(adamic_async_frame *frame, void (*drop)(void *)) {
    (void)frame; (void)drop;
}
static adamic_async_frame *frame_new(void (*resume)(adamic_async_frame *, adamic_value, bool), adamic_async_promise *output) {
    adamic_async_frame *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);
    *frame = (adamic_async_frame){.heap = frame->heap, .resume = resume, .children = children,
        .output = output == NULL ? NULL : adamic_retain(output)};
    return frame;
}
static adamic_string *text(const char *bytes) {
    size_t length = strlen(bytes);
    adamic_string *string = adamic_string_allocate(length);
    memcpy((char *)string->bytes, bytes, length);
    return string;
}
static void line(const char *bytes) {
    adamic_string *string = text(bytes);
    adamic_write_line(adamic_stdout, string);
    adamic_release(string);
}
static void reject(adamic_async_promise *promise, const char *message) {
    adamic_string *string = text(message);
    adamic_object *error = adamic_error_new(string);
    adamic_release(string);
    adamic_async_settle(promise, (adamic_value){.reference = error}, true, true);
    adamic_release(error);
}
static void queue_frame(adamic_async_frame *frame) {
    adamic_async_promise *ready = adamic_async_new();
    adamic_async_settle(ready, (adamic_value){.number = 0}, false, false);
    adamic_async_await(frame, ready);
    adamic_release(ready);
    adamic_release(frame);
}
static void produce(adamic_async_frame *frame, adamic_value value, bool rejected) {
    (void)value; assert(!rejected);
    const char *messages[] = {"rejected", "chained", "timer rejection", "first", "second"};
    assert(frame->state < sizeof messages / sizeof messages[0]);
    reject(frame->output, messages[frame->state]);
}
static adamic_async_promise *fails(unsigned message) {
    adamic_async_promise *promise = adamic_async_new();
    adamic_async_frame *frame = frame_new(produce, promise);
    frame->state = message;
    queue_frame(frame);
    return promise;
}
static void drop_failure(unsigned message) {
    /* The frame owns the result until settlement; only the rejection queue owns it afterwards. */
    adamic_async_promise *promise = fails(message);
    adamic_release(promise);
}
static void propagate(adamic_async_frame *frame, adamic_value value, bool rejected) {
    assert(rejected);
    adamic_async_settle(frame->output, value, true, true);
}
static void chained(void) {
    adamic_async_promise *inner = fails(1);
    adamic_async_promise *outer = adamic_async_new();
    adamic_async_frame *frame = frame_new(propagate, outer);
    adamic_async_await(frame, inner);
    assert(inner->handled);
    adamic_release(frame);
    adamic_release(inner);
    adamic_release(outer);
}
static void handle_later(adamic_async_frame *frame, adamic_value value, bool rejected) {
    (void)value;
    if (frame->state == 0) {
        assert(!rejected);
        frame->state = 1;
        adamic_async_await(frame, frame->output);
    } else {
        assert(rejected);
        line("handled");
    }
}
static void handled_later(void) {
    adamic_async_promise *promise = adamic_async_new();
    reject(promise, "handled");
    adamic_async_frame *frame = frame_new(handle_later, promise);
    adamic_release(promise);
    queue_frame(frame);
}
static adamic_value timer_callback(adamic_closure *closure, adamic_value *arguments) {
    (void)closure;
    if (arguments[0].number == 1) {
        line("timer");
        drop_failure(2);
    } else line("second");
    return (adamic_value){.number = 0};
}
static void invoke(adamic_closure *closure, size_t argc, const adamic_value *arguments) {
    assert(argc == 1);
    adamic_value argument = arguments[0];
    (void)closure->code(closure, &argument);
}
static void timers(void) {
    adamic_closure *callback = adamic_closure_new(timer_callback, 0);
    bool references[] = {false};
    for (unsigned index = 1; index <= 2; index++) {
        adamic_value argument = {.number = index};
        (void)adamic_timer_start(adamic_timers_provider(), callback, invoke, 1, false, 1, &argument, references);
    }
    adamic_release(callback);
    /* Both deadlines have passed before the first dispatch, so the second timer is ready too. */
    struct timespec pause = {.tv_nsec = 5000000};
    while (nanosleep(&pause, &pause) != 0) assert(errno == EINTR);
}
int main(int argc, char **argv) {
    assert(argc == 2);
    adamic_start(argc, argv);
    if (strcmp(argv[1], "unhandled") == 0) drop_failure(0);
    else if (strcmp(argv[1], "chained") == 0) chained();
    else if (strcmp(argv[1], "handled-later") == 0) handled_later();
    else if (strcmp(argv[1], "timer-checkpoint") == 0) timers();
    else { assert(strcmp(argv[1], "first-only") == 0); drop_failure(3); drop_failure(4); }
    adamic_async_run();
    return 0;
}
