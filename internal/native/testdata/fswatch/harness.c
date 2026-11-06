#if defined(__linux__)
#define _POSIX_C_SOURCE 200809L
#endif
#define ADAMIC_NODE_HOST 1
#include "adamic.h"
#include "async.h"
#include <assert.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

/* What no program can reach yet: a rename, and the error path, which Node takes only when libuv
 * fails reading its descriptor. The error modes put a pipe whose writer is gone where the watcher's
 * inotify descriptor was, so the next wait records POLLHUP on its source. */
static const char *mode;
static char directory[4096];
static adamic_object *watcher;
static void hook(void) {}
static void path(char *out, size_t size, const char *name) {
    assert((size_t)snprintf(out, size, "%s/%s", directory, name) < size);
}
static void print(const adamic_string *text) { printf("%.*s", (int)text->length, text->bytes); }
static adamic_value changed(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    adamic_string *type = arguments[0].reference, *name = arguments[1].reference;
    print(type); putchar(' '); print(name); putchar('\n');
    if (name->length == 3 && memcmp(name->bytes, "end", 3) == 0) {
        /* As a compiled call does, hold the method across the call that frees its watcher. */
        adamic_closure *close = adamic_retain(watcher->slots[0].reference);
        (void)adamic_closure_call(close, NULL, 0);
        adamic_release(close);
    }
    return (adamic_value){.reference = NULL};
}
static adamic_value failed(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    adamic_object *error = arguments[0].reference;
    print(error->slots[0].reference); fputs(": ", stdout); print(error->slots[1].reference); putchar('\n');
    fflush(stdout);
    return (adamic_value){.reference = NULL};
}
static void touch(const char *name) {
    char file[4200];
    path(file, sizeof file, name);
    FILE *made = fopen(file, "w");
    assert(made != NULL && fclose(made) == 0);
}
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    if (argc != 2) return 2;
    mode = argv[1];
    const char *temporary = getenv("TMPDIR");
    assert((size_t)snprintf(directory, sizeof directory, "%s/adamic-fswatch-XXXXXX", temporary != NULL && *temporary != 0 ? temporary : "/tmp") < sizeof directory);
    assert(mkdtemp(directory) != NULL);
    if (strcmp(mode, "rename") == 0) touch("b.txt");
    adamic_string *watched = adamic_decode_utf8((const unsigned char *)directory, strlen(directory));
    adamic_closure *listener = adamic_closure_new(changed, 0), *error = adamic_closure_new(failed, 0);
    if (strcmp(mode, "custom-hooks") == 0) assert(adamic_host_set_loop_hooks(hook, hook));
    /* The watcher's descriptor is the lowest one free, as inotify_init1 takes it. */
    int predicted = dup(STDIN_FILENO);
    assert(predicted >= 0 && close(predicted) == 0);
    watcher = adamic_fs_file_host_watch(watched, true, false, listener);
    assert(watcher != NULL && adamic_thrown == NULL);
    adamic_release(watched); adamic_release(listener);
    if (strcmp(mode, "error") == 0) {
        adamic_object *same = adamic_fs_file_host_watch_on(watcher, error);
        assert(same == watcher);
        adamic_release(same);
    }
    adamic_release(error);
    if (strcmp(mode, "error") == 0 || strcmp(mode, "error-uncaught") == 0) {
        char link[64] = {0};
        char proc[64];
        snprintf(proc, sizeof proc, "/proc/self/fd/%d", predicted);
        assert(readlink(proc, link, sizeof link - 1) > 0 && strcmp(link, "anon_inode:inotify") == 0);
        int pipes[2];
        assert(pipe(pipes) == 0 && close(pipes[1]) == 0);
        int flags = fcntl(pipes[0], F_GETFL);
        assert(flags != -1 && fcntl(pipes[0], F_SETFL, flags | O_NONBLOCK) == 0);
        assert(dup2(pipes[0], predicted) == predicted && close(pipes[0]) == 0);
    } else {
        char from[4200], to[4200];
        path(from, sizeof from, "b.txt"); path(to, sizeof to, "c.txt");
        assert(rename(from, to) == 0);
        touch("end");
    }
    adamic_release(watcher); /* The open watcher holds its own. */
    adamic_async_run();
    puts("done");
    char file[4200];
    path(file, sizeof file, "c.txt"); (void)unlink(file);
    path(file, sizeof file, "end"); (void)unlink(file);
    assert(rmdir(directory) == 0);
    return 0;
}
