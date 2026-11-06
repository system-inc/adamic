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
static adamic_object *shared_third;
static adamic_object *alias_first, *alias_second;
static int alias_seen;
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
/* Honest failure injection: replace the shared inotify fd with either a hung-up
 * pipe (POLLHUP), or an open directory (POLLIN followed by read's EISDIR). */
static adamic_value shared_failed(adamic_closure *self, adamic_value *arguments) {
    printf("%.0f ", self->cells[0]->value.number);
    (void)failed(self, arguments);
    if (self->cells[0]->value.number == 1) {
        adamic_closure *close = adamic_retain(shared_third->slots[0].reference);
        (void)adamic_closure_call(close, NULL, 0);
        adamic_release(close); adamic_release(shared_third); shared_third = NULL;
        adamic_string *name = adamic_decode_utf8((const unsigned char *)directory, strlen(directory));
        adamic_closure *listener = adamic_closure_new(changed, 0);
        watcher = adamic_fs_file_host_watch(name, true, false, listener);
        assert(watcher != NULL && adamic_thrown == NULL);
        adamic_release(name); adamic_release(listener);
        adamic_release(watcher); /* The open fresh watcher holds its object. */
        touch("end");
    }
    return (adamic_value){.reference = NULL};
}
static void shared_failure(const adamic_string *name, adamic_closure *listener) {
    int predicted = dup(STDIN_FILENO);
    assert(predicted >= 0 && close(predicted) == 0);
    for (int i = 1; i <= 3; i++) {
        adamic_object *object = adamic_fs_file_host_watch(name, true, false, listener);
        assert(object != NULL && adamic_thrown == NULL);
        adamic_closure *error = adamic_closure_new(shared_failed, 1);
        error->cells[0] = adamic_cell_new((adamic_value){.number = i}, false);
        adamic_object *same = adamic_fs_file_host_watch_on(object, error);
        adamic_release(same); adamic_release(error);
        if (i == 3) shared_third = object; else adamic_release(object);
    }
    int replacement;
    if (strcmp(mode, "shared-read-error") == 0) {
        replacement = open(directory, O_RDONLY | O_NONBLOCK | O_DIRECTORY);
        assert(replacement != -1);
    } else {
        int pipes[2];
        assert(pipe(pipes) == 0 && close(pipes[1]) == 0);
        replacement = pipes[0];
        int flags = fcntl(replacement, F_GETFL);
        assert(flags != -1 && fcntl(replacement, F_SETFL, flags | O_NONBLOCK) == 0);
    }
    assert(dup2(replacement, predicted) == predicted && close(replacement) == 0);
}
static adamic_value alias_changed(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    print(arguments[0].reference); putchar(' '); print(arguments[1].reference); putchar('\n');
    if (++alias_seen == 2) {
        adamic_closure *first = adamic_retain(alias_first->slots[0].reference);
        adamic_closure *second = adamic_retain(alias_second->slots[0].reference);
        (void)adamic_closure_call(first, NULL, 0);
        (void)adamic_closure_call(second, NULL, 0);
        adamic_release(first); adamic_release(second);
    }
    return (adamic_value){.reference = NULL};
}
static void aliases(void) {
    char first[4200], second[4200];
    path(first, sizeof first, "a"); path(second, sizeof second, "b");
    touch("a"); assert(link(first, second) == 0);
    adamic_string *a = adamic_decode_utf8((const unsigned char *)first, strlen(first));
    adamic_string *b = adamic_decode_utf8((const unsigned char *)second, strlen(second));
    adamic_closure *listener = adamic_closure_new(alias_changed, 0);
    alias_first = adamic_fs_file_host_watch(a, true, false, listener);
    alias_second = adamic_fs_file_host_watch(b, true, false, listener);
    assert(alias_first != NULL && alias_second != NULL && adamic_thrown == NULL);
    adamic_release(alias_first); adamic_release(alias_second);
    adamic_release(a); adamic_release(b); adamic_release(listener);
    FILE *file = fopen(second, "a");
    assert(file != NULL && fputs("x", file) >= 0 && fclose(file) == 0);
    adamic_async_run();
    assert(unlink(first) == 0 && unlink(second) == 0 && rmdir(directory) == 0);
}
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    if (argc != 2) return 2;
    mode = argv[1];
    const char *temporary = getenv("TMPDIR");
    assert((size_t)snprintf(directory, sizeof directory, "%s/adamic-fswatch-XXXXXX", temporary != NULL && *temporary != 0 ? temporary : "/tmp") < sizeof directory);
    assert(mkdtemp(directory) != NULL);
    if (strcmp(mode, "aliases") == 0) { aliases(); return 0; }
    if (strcmp(mode, "rename") == 0) touch("b.txt");
    adamic_string *watched = adamic_decode_utf8((const unsigned char *)directory, strlen(directory));
    adamic_closure *listener = adamic_closure_new(changed, 0), *error = adamic_closure_new(failed, 0);
    if (strcmp(mode, "custom-hooks") == 0) assert(adamic_host_set_loop_hooks(hook, hook));
    if (strcmp(mode, "shared-error") == 0 || strcmp(mode, "shared-read-error") == 0) {
        shared_failure(watched, listener);
        adamic_release(watched); adamic_release(listener); adamic_release(error);
        adamic_async_run();
        puts("done");
        char file[4200]; path(file, sizeof file, "end");
        assert(unlink(file) == 0 && rmdir(directory) == 0);
        return 0;
    }
    /* The shared descriptor is the lowest one free, as inotify_init1 takes it. */
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
