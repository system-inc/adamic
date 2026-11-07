#if defined(__linux__)
#define _POSIX_C_SOURCE 200809L
#endif
#include "sources.h"
#include "timers.h"
#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static const char *mode;
static int first[2] = {-1, -1}, second[2] = {-1, -1};
static adamic_source_handle first_source, second_source;
static adamic_timer_handle interval;
static unsigned dispatches, ticks;

/* A ref of a closed handle must answer false. One that answers true ends the run
 * with a plain exit, not assert's abort: the box pipes every abort to its crash
 * reporter, which can take past the pin's 3 s deadline under load (#92h0vtk). */
static void ref_is_stale(adamic_source_handle handle) {
    if (!adamic_source_ref(handle)) return;
    fflush(stdout);
    fputs("harness: a ref after close answered true\n", stderr);
    _exit(3);
}

/* A pipe as the source, its ends made non-blocking unless the mode wants the
 * refusal, and one byte already readable when filled. */
static void open_pipe(int descriptors[2], bool blocking, bool filled) {
    assert(pipe(descriptors) == 0);
    for (unsigned index = 0; index < 2 && !blocking; index++) {
        int flags = fcntl(descriptors[index], F_GETFL);
        assert(flags != -1 && fcntl(descriptors[index], F_SETFL, flags | O_NONBLOCK) == 0);
    }
    unsigned char byte = 1;
    if (filled) assert(write(descriptors[1], &byte, 1) == 1);
}
static void drain(int descriptor) {
    unsigned char bytes[16];
    ssize_t result;
    while ((result = read(descriptor, bytes, sizeof bytes)) > 0) {}
    assert(result == -1 && (errno == EAGAIN || errno == EWOULDBLOCK));
}
static void empty_children(adamic_async_frame *frame, void (*drop)(void *)) { (void)frame; (void)drop; }
static void resume(adamic_async_frame *frame, adamic_value value, bool rejected) {
    (void)frame; (void)value; assert(!rejected);
    puts("job");
}
static void job(void) {
    adamic_async_promise *promise = adamic_async_new();
    adamic_async_settle(promise, (adamic_value){.number = 0}, false, false);
    adamic_async_frame *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);
    *frame = (adamic_async_frame){.heap = frame->heap, .children = empty_children, .resume = resume};
    adamic_async_await(frame, promise);
    adamic_release(frame); adamic_release(promise);
}
static void dispatch(void *data, unsigned ready) {
    int *descriptors = data;
    dispatches++;
    if (strcmp(mode, "fairness") == 0) return; /* Never drained: it stays readable every turn. */
    if (strcmp(mode, "error") == 0) {
        /* The write end is closed: POLLHUP arrives here, never at abort(). */
        puts((ready & adamic_source_error) != 0 ? "error" : "no error");
        assert(adamic_source_close(first_source));
        return;
    }
    assert(ready == adamic_source_readable);
    drain(descriptors[0]);
    if (strcmp(mode, "throw") == 0) {
        static const char message[] = "source dispatch failed";
        adamic_string *text = adamic_string_allocate(sizeof message - 1);
        memcpy((char *)text->bytes, message, sizeof message - 1);
        adamic_thrown = adamic_error_new(text);
        adamic_release(text);
    } else if (strcmp(mode, "self-close") == 0) {
        puts("source");
        assert(adamic_source_close(first_source));
        assert(!adamic_source_close(first_source)); /* Stale at once, though still dispatching. */
    } else if (strcmp(mode, "unref-live") == 0) {
        puts("source");
    } else if (strcmp(mode, "close-other") == 0) {
        /* Both were recorded ready in one wait; the second must not run now. */
        puts(descriptors == first ? "first" : "second");
        assert(adamic_source_close(second_source));
        assert(adamic_source_close(first_source));
    } else if (strcmp(mode, "checkpoint") == 0) {
        /* The job runs before the second source: each dispatch is a macrotask. */
        puts(descriptors == first ? "first" : "second");
        if (descriptors == first) job();
        assert(adamic_source_close(descriptors == first ? first_source : second_source));
    } else if (strcmp(mode, "pending-close") == 0 || strcmp(mode, "ref-flip") == 0) {
        /* The unreferenced sibling's byte, read in the one turn the close keeps. A
         * refness flip keeps no turn, so ref-flip never reaches it. */
        puts("sibling");
    } else if (strcmp(mode, "ref-stale") == 0) {
        /* A ref after close finds nothing, and the close's one turn ends the loop. */
        puts("closed");
        assert(adamic_source_close(first_source));
        ref_is_stale(first_source);
    } else if (strcmp(mode, "pending-close-once") == 0) {
        /* Readable again at once, through a third dispatch, as a watched file that
         * keeps changing: a second turn, or a loop kept until no source is ready,
         * dispatches it again. */
        printf("sibling %u\n", dispatches);
        unsigned char byte = 1;
        if (dispatches < 3) assert(write(second[1], &byte, 1) == 1);
    } else abort();
}
static adamic_value tick(adamic_closure *self, adamic_value *args) {
    (void)self; (void)args;
    if (strcmp(mode, "fairness") == 0) {
        printf("interval %u\n", ++ticks);
        if (ticks == 3) {
            assert(adamic_timer_cancel(interval));
            assert(adamic_source_close(first_source));
        }
    } else puts("timer");
    return (adamic_value){.number = 0};
}
/* The library supplies a signature-checked thunk, not an unchecked code cast. */
static void invoke(adamic_closure *closure, size_t argc, const adamic_value *args) {
    assert(argc == 0); (void)args;
    (void)closure->code(closure, NULL);
}
static adamic_timer_handle start(double delay, bool repeat) {
    adamic_closure *closure = adamic_closure_new(tick, 0);
    adamic_timer_handle handle = adamic_timer_start(adamic_timers_provider(), closure, invoke, delay, repeat, 0, NULL, NULL);
    adamic_release(closure);
    return handle;
}
static void custom_hook(void) {}
int main(int argc, char **argv) {
    assert(argc == 2); mode = argv[1];
    puts("prefix");
    assert(fflush(stdout) == 0);
    if (strcmp(mode, "custom-hooks") == 0) {
        assert(adamic_host_set_loop_hooks(custom_hook, custom_hook));
        open_pipe(first, false, true);
        (void)adamic_source_register(first[0], dispatch, first, true);
        return 0; /* Admission is the mutant: this must have panicked above. */
    } else if (strcmp(mode, "blocking") == 0) {
        open_pipe(first, true, true);
        (void)adamic_source_register(first[0], dispatch, first, true);
        return 0; /* The registration must have stopped above. */
    } else if (strcmp(mode, "self-close") == 0 || strcmp(mode, "throw") == 0) {
        open_pipe(first, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
    } else if (strcmp(mode, "unref") == 0) {
        /* Nothing to read: only a liveness that counted it could wait here. */
        open_pipe(first, false, false);
        first_source = adamic_source_register(first[0], dispatch, first, false);
    } else if (strcmp(mode, "unref-live") == 0) {
        open_pipe(first, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
        assert(adamic_source_unref(first_source));
        (void)start(20, false);
    } else if (strcmp(mode, "fairness") == 0) {
        open_pipe(first, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
        interval = start(1, true);
    } else if (strcmp(mode, "close-other") == 0 || strcmp(mode, "checkpoint") == 0) {
        open_pipe(first, false, true);
        open_pipe(second, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
        second_source = adamic_source_register(second[0], dispatch, second, true);
    } else if (strcmp(mode, "error") == 0) {
        open_pipe(first, false, false);
        assert(close(first[1]) == 0);
        first[1] = -1;
        first_source = adamic_source_register(first[0], dispatch, first, true);
    } else if (strcmp(mode, "pending-close") == 0 || strcmp(mode, "pending-close-once") == 0) {
        /* The only referenced source closes before the loop: its pending close is
         * all that keeps the loop, and the unreferenced sibling has a byte. */
        open_pipe(first, false, false);
        open_pipe(second, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
        second_source = adamic_source_register(second[0], dispatch, second, false);
        assert(adamic_source_close(first_source));
    } else if (strcmp(mode, "ref-flip") == 0) {
        /* Flipped both ways, unref then ref then unref, and never closed: nothing is
         * referenced, so the loop ends before any wait, as Node's does with only
         * unreferenced handles, and the sibling's byte is never read. */
        open_pipe(first, false, false);
        open_pipe(second, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
        second_source = adamic_source_register(second[0], dispatch, second, false);
        assert(adamic_source_unref(first_source));
        assert(adamic_source_ref(first_source));
        assert(adamic_source_unref(first_source));
    } else if (strcmp(mode, "ref-stale") == 0) {
        open_pipe(first, false, true);
        first_source = adamic_source_register(first[0], dispatch, first, true);
    } else if (strcmp(mode, "pending-close-unref-timer") == 0) {
        /* A pending close beside an unreferenced 10 s timer: the turn's wait must
         * not block on the timer, past the pin's 3 s deadline. */
        open_pipe(first, false, false);
        first_source = adamic_source_register(first[0], dispatch, first, true);
        assert(adamic_source_close(first_source));
        assert(adamic_timer_unref(start(10000, false)));
    } else abort();
    adamic_async_run();
    /* Turns run far faster than the 1 ms interval, and each dispatches the
     * undrained source once, so it ran again between the interval's fires. */
    if (strcmp(mode, "fairness") == 0) puts(dispatches >= 2 ? "source repeated" : "source starved");
    /* Closed by a dispatch or by teardown, every handle is stale now. */
    assert(!adamic_source_close(first_source));
    assert(!adamic_source_close(second_source));
    ref_is_stale(first_source);
    ref_is_stale(second_source);
    for (unsigned index = 0; index < 2; index++) {
        if (first[index] != -1) assert(close(first[index]) == 0);
        if (second[index] != -1) assert(close(second[index]) == 0);
    }
    return 0;
}
