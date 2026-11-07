/* fs.watch on Linux inotify, dispatched by async.c's readiness sources (#j71fn59).
 * Private to node_fs_file.c. Watchers and listeners belong to the loop thread.
 * darwin waits on kqueue (#d7jggtt) and wasi preview1 has no file system events,
 * so both stop by name at the call.
 *
 * One inotify descriptor per loop, with one watch per watched inode. Each event maps as libuv's
 * linux.c maps it: IN_ATTRIB or IN_MODIFY alone is 'change', any other bit is
 * 'rename', and an event on the watched entry itself, which inotify sends without
 * a name, carries the watched path's basename. An open watcher holds its object
 * and its listeners; close(), the error path and teardown release them. The
 * object's close and on capture the watcher's identity, never the object or the
 * record, so nothing a program holds points back at the runtime. */
#if defined(ADAMIC_TARGET_WASI) || !defined(ADAMIC_NODE_HOST) || !defined(__linux__)
static void fs_watcher_unavailable(void) {
#if defined(ADAMIC_TARGET_WASI)
    static const char message[] = "NotYet: wasi preview1 has no file system event provider";
#elif !defined(ADAMIC_NODE_HOST)
    static const char message[] = "NotYet: fs.watch requires a Node host program";
#else
    static const char message[] = "NotYet: fs.watch on this host waits on kqueue (#d7jggtt)";
#endif
    adamic_panic(message, sizeof message - 1);
}
adamic_object *adamic_fs_file_host_watch(const adamic_string *path, bool persistent, bool recursive, adamic_closure *listener) {
    (void)path; (void)persistent; (void)recursive; (void)listener; fs_watcher_unavailable(); return NULL;
}
adamic_object *adamic_fs_file_host_watch_on(adamic_object *watcher, adamic_closure *listener) {
    (void)watcher; (void)listener; fs_watcher_unavailable(); return NULL;
}
void adamic_fs_file_host_watch_shutdown(void) {}
#else
#include "sources.h"
#include <sys/inotify.h>

typedef struct fs_watcher_listener fs_watcher_listener;
struct fs_watcher_listener { adamic_closure *callback; fs_watcher_listener *next; };
typedef struct fs_watcher fs_watcher;
struct fs_watcher {
    uintptr_t identity;
    int watch;
    bool persistent;
    char *base;
    adamic_object *object;
    adamic_closure *listener;
    fs_watcher_listener *errors, *errors_last;
    fs_watcher *next, *watch_next;
};
static fs_watcher *fs_watchers;
static uintptr_t fs_watcher_next_identity = 1;
typedef struct fs_watch_entry fs_watch_entry;
struct fs_watch_entry { int watch; fs_watcher *watchers; fs_watch_entry *next; };
static fs_watch_entry *fs_watch_entries;
static int fs_watch_descriptor = -1;
static adamic_source_handle fs_watch_source;
static uintptr_t fs_watch_generation;
static void fs_watcher_dispatch(void *data, unsigned ready);
#ifdef ADAMIC_FS_WATCH_TEST
static bool fs_watch_test_read_error;
void adamic_fs_watch_test_fail_read(void);
void adamic_fs_watch_test_fail_read(void) { fs_watch_test_read_error = true; }
uintptr_t adamic_fs_watch_test_source_identity(void);
uintptr_t adamic_fs_watch_test_source_identity(void) { return fs_watch_source.identity; }
#endif

static void fs_watch_stop(void) {
    (void)adamic_source_close(fs_watch_source);
    fs_watch_source = (adamic_source_handle){0};
    if (fs_watch_descriptor != -1) close(fs_watch_descriptor);
    fs_watch_descriptor = -1;
    while (fs_watch_entries != NULL) {
        fs_watch_entry *entry = fs_watch_entries;
        fs_watch_entries = entry->next;
        free(entry);
    }
}
static void fs_watch_reference(void) {
    if (fs_watch_descriptor == -1) return;
    bool referenced = false;
    for (fs_watcher *watcher = fs_watchers; watcher != NULL; watcher = watcher->next)
        if (watcher->persistent) referenced = true;
    if (fs_watch_source.identity == 0) {
        fs_watch_source = adamic_source_register(fs_watch_descriptor, fs_watcher_dispatch,
            (void *)fs_watch_generation, referenced);
    } else if (referenced) {
        (void)adamic_source_ref(fs_watch_source);
    } else {
        (void)adamic_source_unref(fs_watch_source);
    }
}
static void fs_watch_drop(int watch) {
    fs_watch_entry **at = &fs_watch_entries;
    while (*at != NULL && (*at)->watch != watch) at = &(*at)->next;
    if (*at != NULL) {
        fs_watch_entry *entry = *at;
        *at = entry->next;
        free(entry);
    }
    for (fs_watcher *watcher = fs_watchers; watcher != NULL; watcher = watcher->next)
        if (watcher->watch == watch) watcher->watch = -1;
}

static const char *const fs_watcher_names[] = {"close", "on"};
static const bool fs_watcher_references[] = {true, true};
static const adamic_field_kind fs_watcher_kinds[] = {adamic_field_reference, adamic_field_reference};
static const adamic_shape fs_watcher_shape = {2, fs_watcher_names, fs_watcher_references, &adamic_host_methods, fs_watcher_kinds};

/* Only an open watcher is found: a closed one's identity is never reused. */
static fs_watcher *fs_watcher_open(uintptr_t identity) {
    for (fs_watcher *watcher = fs_watchers; watcher != NULL; watcher = watcher->next)
        if (watcher->identity == identity) return watcher;
    return NULL;
}
static fs_watcher *fs_watcher_find(const adamic_closure *method) {
    return fs_watcher_open((uintptr_t)method->cells[0]->value.number);
}
/* Closing one of several watchers on an inode must leave its kernel watch live. */
static void fs_watcher_stop(fs_watcher *watcher) {
    fs_watch_entry *entry = fs_watch_entries;
    while (entry != NULL && entry->watch != watcher->watch) entry = entry->next;
    if (entry != NULL) {
        fs_watcher **at = &entry->watchers;
        while (*at != watcher) at = &(*at)->watch_next;
        *at = watcher->watch_next;
    }
    if (entry != NULL && entry->watchers == NULL) {
        if (fs_watch_descriptor != -1) (void)inotify_rm_watch(fs_watch_descriptor, watcher->watch);
        fs_watch_drop(watcher->watch);
    }
    /* Each FSWatcher closes a uv_fs_event handle on Node, even when sibling
     * watchers keep this descriptor alive. Preserve that real pending close;
     * ref/unref below changes only the shared source's liveness. The last
     * watcher's close pends through fs_watch_stop instead, never twice. */
    if (fs_watchers != NULL) adamic_source_pending_close();
    if (fs_watchers == NULL) fs_watch_stop(); else fs_watch_reference();
}
static void fs_watcher_close(fs_watcher *watcher) {
    fs_watcher **at = &fs_watchers;
    while (*at != watcher) at = &(*at)->next;
    *at = watcher->next;
    fs_watcher_stop(watcher);
    for (fs_watcher_listener *listener = watcher->errors; listener != NULL;) {
        fs_watcher_listener *next = listener->next;
        adamic_release(listener->callback); free(listener); listener = next;
    }
    adamic_release(watcher->listener);
    adamic_release(watcher->object);
    free(watcher->base); free(watcher);
}
/* Node calls a listener with the watcher as this and its own argument count. The
 * slots past that count hold undefined, which every parameter lowering admits holds
 * as NULL. */
static void fs_watcher_call(adamic_closure *listener, adamic_object *object, adamic_value *arguments, size_t count) {
#ifdef ADAMIC_CLOSURE_RECEIVERS
    adamic_value result = adamic_closure_receiver_call(listener, object, arguments, count, 3);
#else
    (void)object;
    adamic_value result = adamic_closure_call(listener, arguments, count);
#endif
#ifdef ADAMIC_CLOSURE_DISCARDS
    if (adamic_thrown == NULL) adamic_closure_discard(listener, result);
#else
    (void)result;
#endif
}
static void fs_watcher_change(fs_watcher *watcher, const char *type, const char *name) {
    adamic_closure *listener = adamic_retain(watcher->listener);
    adamic_object *object = adamic_retain(watcher->object);
    adamic_string *event = text(type), *file = text(name);
    adamic_value arguments[3] = {{.reference = event}, {.reference = file}, {.reference = NULL}};
    fs_watcher_call(listener, object, arguments, 2);
    adamic_release(event); adamic_release(file); adamic_release(object); adamic_release(listener);
}
/* Node closes the handle, then emits 'error' to the listeners it has at that
 * moment, and an emit nobody listens for throws. The error bit comes back every
 * turn until the source closes, so the close comes first. */
static void fs_watcher_fail(fs_watcher *watcher, const char *reason) {
    size_t count = 0;
    for (fs_watcher_listener *listener = watcher->errors; listener != NULL; listener = listener->next) count++;
    adamic_closure **batch = malloc((count == 0 ? 1 : count) * sizeof *batch);
    if (batch == NULL) abort();
    size_t index = 0;
    for (fs_watcher_listener *listener = watcher->errors; listener != NULL; listener = listener->next)
        batch[index++] = adamic_retain(listener->callback);
    adamic_object *object = adamic_retain(watcher->object);
    fs_watcher_close(watcher);
    adamic_string *message = text(reason);
    adamic_object *error = adamic_error_new_kind("Error", 5, message);
    adamic_release(message);
    if (count == 0) {
        adamic_thrown = error;
    } else {
        adamic_value arguments[3] = {{.reference = error}, {.reference = NULL}, {.reference = NULL}};
        for (index = 0; index < count; index++) {
            if (adamic_thrown == NULL) fs_watcher_call(batch[index], object, arguments, 1);
            adamic_release(batch[index]);
        }
        adamic_release(error);
    }
    adamic_release(object); free(batch);
}
/* Close the failed descriptor before callbacks can open a replacement. Snapshot only
 * identities: an error listener can close any later watcher and free its record. */
static void fs_watch_fail(const char *reason) {
    fs_watch_stop();
    size_t count = 0;
    for (fs_watcher *watcher = fs_watchers; watcher != NULL; watcher = watcher->next) count++;
    uintptr_t *batch = malloc((count == 0 ? 1 : count) * sizeof *batch);
    if (batch == NULL) abort();
    size_t index = count;
    for (fs_watcher *watcher = fs_watchers; watcher != NULL; watcher = watcher->next) {
        batch[--index] = watcher->identity;
        watcher->watch = -1;
    }
    for (index = 0; index < count; index++) {
        fs_watcher *watcher = fs_watcher_open(batch[index]);
        if (watcher != NULL) {
            fs_watcher_fail(watcher, reason);
            if (adamic_thrown == NULL) adamic_source_drain_jobs();
        }
        if (adamic_thrown != NULL) {
            /* A throw ends Node's emit. Later snapshot watchers are already detached
             * from the failed descriptor; close them without emitting another error. */
            for (index++; index < count; index++) {
                watcher = fs_watcher_open(batch[index]);
                if (watcher != NULL) fs_watcher_close(watcher);
            }
            break;
        }
    }
    free(batch);
}
/* Reads until EAGAIN. Both events and listener batches hold identities only across
 * calls and job checkpoints. A replacement descriptor belongs to a later dispatch. */
static void fs_watcher_dispatch(void *data, unsigned ready) {
    uintptr_t generation = (uintptr_t)data;
    if (fs_watch_descriptor == -1 || generation != fs_watch_generation) return;
    if ((ready & adamic_source_error) != 0) {
        fs_watch_fail("watch: the inotify descriptor hung up or became invalid");
        return;
    }
    _Alignas(struct inotify_event) char buffer[4096];
    for (;;) {
        ssize_t size;
#ifdef ADAMIC_FS_WATCH_TEST
        if (fs_watch_test_read_error) {
            fs_watch_test_read_error = false;
            errno = EIO; size = -1;
        } else
#endif
        do { size = read(fs_watch_descriptor, buffer, sizeof buffer); } while (size == -1 && errno == EINTR);
        if (size == -1 && (errno == EAGAIN || errno == EWOULDBLOCK)) return;
        if (size <= 0) {
            fs_watch_fail("watch: reading the inotify descriptor failed");
            return;
        }
        for (const char *at = buffer; at < buffer + size;) {
            const struct inotify_event *event = (const struct inotify_event *)(const void *)at;
            at += sizeof *event + event->len;
            // IN_Q_OVERFLOW carries watch -1, which libuv finds no watcher for and drops.
            if (event->wd == -1) continue;
            uintptr_t last = fs_watcher_next_identity - 1, identity = 0;
            for (;;) {
                fs_watcher *watcher = NULL;
                fs_watch_entry *entry = fs_watch_entries;
                while (entry != NULL && entry->watch != event->wd) entry = entry->next;
                for (fs_watcher *candidate = entry == NULL ? NULL : entry->watchers; candidate != NULL; candidate = candidate->watch_next)
                    if (candidate->identity > identity && candidate->identity <= last &&
                        (watcher == NULL || candidate->identity < watcher->identity)) watcher = candidate;
                if (watcher == NULL) break;
                identity = watcher->identity;
                const char *type = (event->mask & ~(uint32_t)(IN_ATTRIB | IN_MODIFY)) != 0 ? "rename" : "change";
                fs_watcher_change(watcher, type, event->len != 0 ? event->name : watcher->base);
                if (adamic_thrown != NULL) return;
                // Node takes a microtask checkpoint after each event's callback.
                adamic_source_drain_jobs();
                if (adamic_thrown != NULL || fs_watch_descriptor == -1 || generation != fs_watch_generation) return;
                // The next selection looks every watcher up again, including this identity.
            }
            // libuv delivers IN_IGNORED as the final rename before retiring the wd.
            // Listener calls and checkpoints above may have closed any of its watchers.
            if ((event->mask & IN_IGNORED) != 0) fs_watch_drop(event->wd);
        }
    }
}
static void fs_watcher_listen(fs_watcher *watcher, adamic_closure *callback) {
    fs_watcher_listener *listener = malloc(sizeof *listener);
    if (listener == NULL) abort();
    *listener = (fs_watcher_listener){.callback = adamic_retain(callback)};
    if (watcher->errors_last == NULL) watcher->errors = listener; else watcher->errors_last->next = listener;
    watcher->errors_last = listener;
}
static adamic_value fs_watcher_close_method(adamic_closure *self, adamic_value *arguments) {
    (void)arguments;
    fs_watcher *watcher = fs_watcher_find(self);
    // A second close finds nothing, and does nothing, as Node's does.
    if (watcher != NULL) fs_watcher_close(watcher);
    return (adamic_value){.reference = NULL};
}
/* on through a structural view: Node returns the watcher, which only an open
 * watcher still holds here. */
static adamic_value fs_watcher_on_method(adamic_closure *self, adamic_value *arguments) {
    const adamic_string *event = arguments[0].reference;
    if (event == NULL || event->heap.kind != adamic_kind_string || event->length != 5 || memcmp(event->bytes, "error", 5) != 0) {
        static const char message[] = "NotYet: FSWatcher.on with an event other than 'error'";
        adamic_panic(message, sizeof message - 1);
    }
    fs_watcher *watcher = fs_watcher_find(self);
    if (watcher == NULL) {
        static const char message[] = "NotYet: FSWatcher.on through a structural view after close";
        adamic_panic(message, sizeof message - 1);
    }
    fs_watcher_listen(watcher, arguments[1].reference);
    return (adamic_value){.reference = adamic_retain(watcher->object)};
}
static adamic_closure *fs_watcher_method(adamic_code code, uintptr_t identity, size_t length, bool reference) {
    adamic_closure *method = adamic_closure_new(code, 1);
#ifdef ADAMIC_CLOSURE_DISCARDS
    // on gives the watcher back owned.
    if (reference) method->discard = adamic_discard_reference;
#else
    (void)reference;
#endif
    method->source_length = length;
    method->cells[0] = adamic_cell_new((adamic_value){.number = (double)identity}, false);
    return method;
}

adamic_object *adamic_fs_file_host_watch(const adamic_string *path, bool persistent, bool recursive, adamic_closure *listener) {
    adamic_output_flush();
    if (recursive) {
        static const char message[] = "NotYet: fs.watch with recursive: true";
        adamic_panic(message, sizeof message - 1);
    }
    if (memchr(path->bytes, 0, path->length) != NULL) {
        invalid_value("The argument 'filename' must be a string, Uint8Array, or URL without null bytes. Received ", path);
        return NULL;
    }
    char *name = bytes(path);
    if (fs_watch_descriptor == -1) {
        fs_watch_descriptor = inotify_init1(IN_NONBLOCK | IN_CLOEXEC);
        if (fs_watch_descriptor == -1) { system_error(errno, "watch", name); free(name); return NULL; }
        if (++fs_watch_generation == 0) abort();
    }
    int watch = inotify_add_watch(fs_watch_descriptor, name, IN_ATTRIB | IN_CREATE | IN_MODIFY | IN_DELETE | IN_DELETE_SELF | IN_MOVE_SELF | IN_MOVED_FROM | IN_MOVED_TO);
    if (watch == -1) {
        int error = errno;
        if (fs_watch_source.identity == 0) fs_watch_stop();
        if (error == ENOSPC) {
            // Node words its own message for the watch limit (lib/internal/fs/watchers.js).
            size_t length = strlen(name) + 80;
            char *message = malloc(length);
            if (message == NULL) { adamic_panic("out of memory", 13); }
            snprintf(message, length, "ENOSPC: System limit for number of file watchers reached, watch '%s'", name);
            raise_error(&error_shape, "Error", "ENOSPC", message);
            free(message);
        } else {
            system_error(error, "watch", name);
        }
        free(name);
        return NULL;
    }
    if (fs_watcher_next_identity > 9007199254740991ULL) abort();
    const char *slash = strrchr(name, '/');
    char *base = strdup(slash == NULL ? name : slash + 1);
    free(name);
    fs_watcher *watcher = malloc(sizeof *watcher);
    if (base == NULL || watcher == NULL) abort();
    *watcher = (fs_watcher){.identity = fs_watcher_next_identity++, .watch = watch, .persistent = persistent,
        .base = base, .listener = adamic_retain(listener)};
    adamic_object *object = adamic_object_new(&fs_watcher_shape);
    object->slots[0].reference = fs_watcher_method(fs_watcher_close_method, watcher->identity, 0, false);
    object->slots[1].reference = fs_watcher_method(fs_watcher_on_method, watcher->identity, 2, true);
    watcher->object = adamic_retain(object);
    // Refused under custom host loop hooks in the source's own words.
    watcher->next = fs_watchers;
    fs_watchers = watcher;
    fs_watch_entry *entry = fs_watch_entries;
    while (entry != NULL && entry->watch != watch) entry = entry->next;
    if (entry == NULL) {
        entry = malloc(sizeof *entry);
        if (entry == NULL) abort();
        *entry = (fs_watch_entry){.watch = watch, .next = fs_watch_entries};
        fs_watch_entries = entry;
    }
    watcher->watch_next = entry->watchers;
    entry->watchers = watcher;
    fs_watch_reference();
    return object;
}
/* The direct call, recognized by lowering on an FSWatcher with the literal 'error'.
 * After close Node keeps the listener and never calls it; dropping it is the same. */
adamic_object *adamic_fs_file_host_watch_on(adamic_object *object, adamic_closure *listener) {
    adamic_closure *method = object->shape == &fs_watcher_shape ? object->slots[1].reference : NULL;
    if (method == NULL || method->code != fs_watcher_on_method) {
        static const char message[] = "NotYet: FSWatcher.on on a value fs.watch didn't make, or whose on was replaced";
        adamic_panic(message, sizeof message - 1);
    }
    fs_watcher *watcher = fs_watcher_find(method);
    if (watcher != NULL) fs_watcher_listen(watcher, listener);
    return adamic_retain(object);
}
void adamic_fs_file_host_watch_shutdown(void) {
    while (fs_watchers != NULL) fs_watcher_close(fs_watchers);
    fs_watch_stop();
}
#endif
