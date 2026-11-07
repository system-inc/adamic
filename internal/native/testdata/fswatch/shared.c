#if defined(__linux__)
#define _POSIX_C_SOURCE 200809L
#endif
#define ADAMIC_NODE_HOST 1
#include "adamic.h"
#include "async.h"
#include <assert.h>
#include <dirent.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <unistd.h>

void adamic_fs_watch_test_fail_read(void);
uintptr_t adamic_fs_watch_test_source_identity(void);
size_t adamic_source_test_pending_close_count(void);
static adamic_object *handles[100];
static char directory[4096], file[4200];
// A verified watch failure is a chosen fixture result, not a harness status.
enum { fs_watch_cap_failure_status = 17, fs_watch_refness_failure_status = 18 };
static int phase;
static bool cap_opening;
static void finish(adamic_object *handle) {
    adamic_closure *method = adamic_retain(handle->slots[0].reference);
    (void)adamic_closure_call(method, NULL, 0);
    adamic_release(method);
}
static void write_file(void) {
    FILE *output = fopen(file, "w");
    assert(output != NULL);
    assert(fputs("changed", output) >= 0 && fclose(output) == 0);
}
static adamic_object *open_watch(adamic_code code, bool persistent) {
    adamic_string *name = adamic_decode_utf8((const unsigned char *)file, strlen(file));
    adamic_closure *listener = adamic_closure_new(code, 0);
    adamic_object *handle = adamic_fs_file_host_watch(name, persistent, false, listener);
    adamic_release(name); adamic_release(listener);
    if (handle == NULL && cap_opening) return NULL;
    if (handle == NULL && adamic_thrown != NULL) {
        const adamic_string *message = adamic_thrown->slots[1].reference;
        fprintf(stderr, "%.*s\n", (int)message->length, message->bytes);
    }
    assert(handle != NULL && adamic_thrown == NULL);
    return handle;
}
static adamic_value third_changed(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    puts("third change"); finish(handles[2]);
    return (adamic_value){.reference = NULL};
}
static adamic_value unexpected(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    puts("unexpected callback");
    return (adamic_value){.reference = NULL};
}
static adamic_value failed_first(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    puts("first error");
    finish(handles[1]);
    handles[2] = open_watch(third_changed, true);
    write_file();
    return (adamic_value){.reference = NULL};
}
static void on_error(adamic_object *handle, adamic_code code) {
    adamic_closure *listener = adamic_closure_new(code, 0);
    adamic_release(adamic_fs_file_host_watch_on(handle, listener));
    adamic_release(listener);
}
static adamic_value first_changed(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    if (phase == 0) { puts("first change"); finish(handles[0]); }
    return (adamic_value){.reference = NULL};
}
static adamic_value second_changed(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    if (phase++ == 0) { puts("second change"); write_file(); }
    else { puts("second survives"); finish(handles[1]); }
    return (adamic_value){.reference = NULL};
}
static adamic_value cap_changed(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    const adamic_string *type = arguments[0].reference;
    assert(type->length == 6 && memcmp(type->bytes, "change", 6) == 0);
    puts("cap change");
    for (size_t index = 0; index < 100; index++) finish(handles[index]);
    return (adamic_value){.reference = NULL};
}
static void assert_text(const adamic_string *value, const char *expected) {
    assert(value != NULL && value->length == strlen(expected) && memcmp(value->bytes, expected, value->length) == 0);
}
static size_t inotify_count(void) {
    DIR *entries = opendir("/proc/self/fd");
    assert(entries != NULL);
    size_t count = 0;
    for (struct dirent *entry; (entry = readdir(entries)) != NULL;) {
        char path[128], target[128];
        if (entry->d_name[0] == '.') continue;
        int descriptor = atoi(entry->d_name);
        assert(snprintf(path, sizeof path, "/proc/self/fd/%d", descriptor) > 0);
        ssize_t length = readlink(path, target, sizeof target - 1);
        if (length < 0) continue;
        target[length] = 0;
        if (strcmp(target, "anon_inode:inotify") == 0) count++;
    }
    assert(closedir(entries) == 0);
    return count;
}
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    int status = 0;
    assert(argc == 2);
    const char *temporary = getenv("TMPDIR");
    assert(snprintf(directory, sizeof directory, "%s/adamic-shared-XXXXXX", temporary == NULL ? "/tmp" : temporary) > 0);
    assert(mkdtemp(directory) != NULL);
    assert(snprintf(file, sizeof file, "%s/file", directory) > 0);
    write_file();
    if (strcmp(argv[1], "cap") == 0) {
        struct rlimit limit;
        assert(getrlimit(RLIMIT_NOFILE, &limit) == 0);
        limit.rlim_cur = 64;
        assert(setrlimit(RLIMIT_NOFILE, &limit) == 0);
        for (size_t index = 0; index < 100; index++) {
            assert(snprintf(file, sizeof file, "%s/file%zu", directory, index) > 0);
            write_file();
        }
        cap_opening = true;
        for (size_t index = 0; index < 100; index++) {
            assert(snprintf(file, sizeof file, "%s/file%zu", directory, index) > 0);
            handles[index] = open_watch(cap_changed, true);
            if (handles[index] == NULL) {
                assert(adamic_thrown != NULL);
                static adamic_slot_cache code_cache, message_cache;
                assert_text(adamic_object_field(adamic_thrown, "code", &code_cache)->reference, "EMFILE");
                char expected_message[sizeof file + 80];
                assert(snprintf(expected_message, sizeof expected_message, "EMFILE: too many open files, watch '%s'", file) > 0);
                assert_text(adamic_object_field(adamic_thrown, "message", &message_cache)->reference, expected_message);
                adamic_release(adamic_thrown); adamic_thrown = NULL;
                for (size_t opened = 0; opened < index; opened++) finish(handles[opened]);
                // The exact message above includes this run's temporary path.
                puts("watch failure code=EMFILE message=EMFILE: too many open files, watch '<fixture-path>'");
                status = fs_watch_cap_failure_status;
                break;
            }
        }
        cap_opening = false;
        if (status == 0) {
            assert(inotify_count() == 1); puts("100 watchers, 1 inotify");
            write_file(); adamic_async_run();
            assert(inotify_count() == 0);
        }
    } else if (strcmp(argv[1], "refness") == 0) {
        handles[0] = open_watch(unexpected, false);
        uintptr_t identity = adamic_fs_watch_test_source_identity();
        assert(identity != 0 && adamic_source_test_pending_close_count() == 0);
        handles[1] = open_watch(unexpected, true);
        bool changed = adamic_fs_watch_test_source_identity() != identity;
        size_t pending = adamic_source_test_pending_close_count();
        printf("refness source %s pending=%zu\n", changed ? "changed" : "unchanged", pending);
        if (changed || pending != 0) status = fs_watch_refness_failure_status;
        // All observations precede every close; cleanup cannot create a false pin.
        finish(handles[1]); finish(handles[0]); adamic_async_run();
    } else if (strcmp(argv[1], "failure") == 0) {
        handles[0] = open_watch(unexpected, true);
        handles[1] = open_watch(unexpected, true);
        on_error(handles[0], failed_first); on_error(handles[1], unexpected);
        adamic_fs_watch_test_fail_read(); write_file(); adamic_async_run();
    } else if (strcmp(argv[1], "same") == 0) {
        handles[0] = open_watch(first_changed, true);
        handles[1] = open_watch(second_changed, true);
        write_file(); adamic_async_run();
    } else if (strcmp(argv[1], "unref") == 0) {
        handles[0] = open_watch(unexpected, false);
        handles[1] = open_watch(unexpected, true);
        finish(handles[1]);
        // No queued events: the real watcher close keeps one nonblocking turn.
        // The surviving shared source is unreferenced throughout it.
        adamic_async_run(); puts("unref exits"); finish(handles[0]);
    } else abort();
    for (size_t index = 0; index < 100; index++) adamic_release(handles[index]);
    if (strcmp(argv[1], "cap") == 0) {
        for (size_t index = 0; index < 100; index++) {
            assert(snprintf(file, sizeof file, "%s/file%zu", directory, index) > 0);
            assert(unlink(file) == 0);
        }
    }
    assert(snprintf(file, sizeof file, "%s/file", directory) > 0);
    assert(unlink(file) == 0 && rmdir(directory) == 0);
    return status;
}
