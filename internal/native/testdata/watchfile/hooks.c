#if defined(__linux__)
#define _POSIX_C_SOURCE 200809L
#endif
#define ADAMIC_NODE_HOST 1
#include "adamic.h"
#include "async.h"
#include "timers.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
static void hook(void) {}
static adamic_value changed(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments; return (adamic_value){.number = 0};
}
static adamic_string missing = ADAMIC_STRING("missing");
static void no_children(adamic_async_frame *frame, void (*drop)(void *)) { (void)frame; (void)drop; }
static void resume_job(adamic_async_frame *frame, adamic_value value, bool rejected) {
    (void)frame; (void)value; assert(!rejected); puts("job");
}
static adamic_value first(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments; puts("first");
    adamic_async_promise *promise = adamic_async_new();
    adamic_async_frame *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);
    *frame = (adamic_async_frame){.heap = frame->heap, .resume = resume_job, .children = no_children};
    adamic_async_settle(promise, (adamic_value){.number = 0}, false, false);
    adamic_async_await(frame, promise);
    adamic_release(frame); adamic_release(promise);
    (void)adamic_fs_file_unwatch(&missing, NULL);
    return (adamic_value){.number = 0};
}
static adamic_value second(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments; puts("second"); return (adamic_value){.number = 0};
}
static unsigned notifications;
static adamic_value counted(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments; notifications++;
    FILE *file = fopen("missing", "a"); assert(file != NULL);
    assert(fputc('x', file) != EOF); assert(fclose(file) == 0);
    return (adamic_value){.number = 0};
}
static void stop_count(adamic_closure *self, size_t argc, const adamic_value *arguments) {
    (void)self; (void)argc; (void)arguments;
    (void)adamic_fs_file_unwatch(&missing, NULL);
    printf("interval %s\n", notifications <= 7 ? "true" : "false");
}
static adamic_value failing(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    static adamic_string message = ADAMIC_STRING("watch listener failed");
    adamic_thrown = adamic_error_new(&message);
    return (adamic_value){.number = 0};
}
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    if (argc == 2 && strcmp(argv[1], "throw") == 0) {
        adamic_closure *listener = adamic_closure_new(failing, 0);
        (void)adamic_fs_file_watch(&missing, true, 10, listener);
        adamic_release(listener); adamic_async_run(); puts("wrong survived"); return 0;
    }
    if (argc == 2 && (strcmp(argv[1], "batch") == 0 || strcmp(argv[1], "interval") == 0)) {
        char path[] = "/tmp/adamic-watch-batch-XXXXXX";
        char *directory = mkdtemp(path); assert(directory != NULL);
        assert(chdir(directory) == 0);
        if (strcmp(argv[1], "interval") == 0) {
            adamic_closure *listener = adamic_closure_new(counted, 0);
            (void)adamic_fs_file_watch(&missing, true, 100, listener);
            (void)adamic_fs_file_watch(&missing, true, 1, listener);
            // Two listeners receive each event, so remove one registration.
            (void)adamic_fs_file_unwatch(&missing, listener);
            adamic_release(listener);
            adamic_closure *token = adamic_closure_new((adamic_code)NULL, 0);
            (void)adamic_timer_start(adamic_timers_provider(), token, stop_count, 500, false, 0, NULL, NULL);
            adamic_release(token);
        } else {
            adamic_closure *a = adamic_closure_new(first, 0), *b = adamic_closure_new(second, 0);
            (void)adamic_fs_file_watch(&missing, true, 10, a);
            (void)adamic_fs_file_watch(&missing, true, 10, b);
            adamic_release(a); adamic_release(b);
        }
        adamic_async_run();
        if (strcmp(argv[1], "interval") == 0) assert(unlink("missing") == 0);
        assert(chdir("/") == 0); assert(rmdir(directory) == 0);
        return 0;
    }
    assert(adamic_host_set_loop_hooks(hook, hook));
    static adamic_string path = ADAMIC_STRING("missing");
    adamic_closure *listener = adamic_closure_new(changed, 0);
    (void)adamic_fs_file_watch(&path, true, 20, listener);
    adamic_release(listener);
    puts("admitted");
    adamic_async_teardown();
    return 0;
}
