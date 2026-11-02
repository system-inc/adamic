/* Darwin is unchecked until #d7jggtt. POSIX stat/time/fcntl/unistd headers
 * are shared; Linux alone uses _DEFAULT_SOURCE, linux/stat.h, sys/syscall.h and
 * sys/sysmacros.h for statx birthtimes (with the same stat fallback).
 * Private to node_fs_file.c. Pollers and listeners belong to the loop thread.
 * A normalized path shares its first interval/persistence, as Node does. */
#include "timers.h"
#if defined(ADAMIC_TARGET_WASI) || !defined(ADAMIC_NODE_HOST)
static void watch_unavailable(void) {
#ifdef ADAMIC_TARGET_WASI
    static const char message[] = "NotYet: wasi preview1 has no stat-polling watcher provider";
#else
    static const char message[] = "NotYet: stat-polling watchers require a Node host program";
#endif
    adamic_panic(message, sizeof message - 1);
}
double adamic_fs_file_watch(const adamic_string *path, bool persistent, double interval, adamic_closure *listener) {
    (void)path; (void)persistent; (void)interval; (void)listener; watch_unavailable(); return 0;
}
double adamic_fs_file_unwatch(const adamic_string *path, adamic_closure *listener) {
    (void)path; (void)listener; watch_unavailable(); return 0;
}
void adamic_fs_file_watch_shutdown(void) {}
#else
#if defined(__linux__)
#include <linux/stat.h>
#include <sys/syscall.h>
#include <sys/sysmacros.h>
#endif

typedef struct watch_snapshot {
    struct stat value;
    struct timespec birth;
    uint64_t flags, generation;
} watch_snapshot;
typedef struct watch_listener watch_listener;
struct watch_listener { adamic_closure *callback; watch_listener *next; };
typedef struct watch_poller watch_poller;
struct watch_poller {
    uintptr_t identity;
    char *path;
    unsigned interval;
    bool persistent, active, firing;
    int status;
    watch_snapshot previous;
    watch_listener *first, *last;
    adamic_timer_handle timer;
    watch_poller *next;
};
static watch_poller *watch_pollers;
static uintptr_t watch_next_identity = 1;

static double watch_now(void) {
    struct timespec now;
    if (clock_gettime(CLOCK_MONOTONIC, &now) != 0) abort();
    return floor((double)now.tv_sec * 1000 + (double)now.tv_nsec / 1000000);
}
static int watch_stat(const char *path, watch_snapshot *snapshot) {
    memset(snapshot, 0, sizeof *snapshot);
#if defined(__linux__) && defined(SYS_statx) && defined(STATX_BTIME)
    struct statx extended;
    if (syscall(SYS_statx, AT_FDCWD, path, 0, STATX_BASIC_STATS | STATX_BTIME, &extended) == 0) {
        struct stat *value = &snapshot->value;
        value->st_dev = makedev(extended.stx_dev_major, extended.stx_dev_minor);
        value->st_ino = extended.stx_ino; value->st_mode = extended.stx_mode;
        value->st_uid = extended.stx_uid; value->st_gid = extended.stx_gid;
        value->st_size = extended.stx_size; value->st_nlink = extended.stx_nlink;
        value->st_atim = (struct timespec){extended.stx_atime.tv_sec, extended.stx_atime.tv_nsec};
        value->st_mtim = (struct timespec){extended.stx_mtime.tv_sec, extended.stx_mtime.tv_nsec};
        value->st_ctim = (struct timespec){extended.stx_ctime.tv_sec, extended.stx_ctime.tv_nsec};
        snapshot->birth = (struct timespec){extended.stx_btime.tv_sec, extended.stx_btime.tv_nsec};
        return 0;
    }
    /* Match libuv: old seccomp filters and DVS filesystems may reject statx. */
    if (errno != ENOSYS && errno != EINVAL && errno != EPERM && errno != EOPNOTSUPP) return -errno;
#endif
    if (stat(path, &snapshot->value) != 0) return -errno;
#ifdef __APPLE__
    snapshot->birth = snapshot->value.st_birthtimespec;
    snapshot->flags = snapshot->value.st_flags;
    snapshot->generation = snapshot->value.st_gen;
#endif
    return 0;
}
static bool watch_times_differ(struct timespec a, struct timespec b) {
    return a.tv_sec != b.tv_sec || a.tv_nsec != b.tv_nsec;
}
static bool watch_changed(const watch_snapshot *old, const watch_snapshot *current) {
    if (old->value.st_size != current->value.st_size) return true;
#ifdef __APPLE__
    struct timespec old_mtime = old->value.st_mtimespec, mtime = current->value.st_mtimespec;
    struct timespec old_ctime = old->value.st_ctimespec, ctime = current->value.st_ctimespec;
#else
    struct timespec old_mtime = old->value.st_mtim, mtime = current->value.st_mtim;
    struct timespec old_ctime = old->value.st_ctim, ctime = current->value.st_ctim;
#endif
    // libuv fs-poll compares these fields, not atime or nlink.
    return watch_times_differ(old_mtime, mtime) || watch_times_differ(old_ctime, ctime) ||
        watch_times_differ(old->birth, current->birth) ||
        old->value.st_mode != current->value.st_mode || old->value.st_uid != current->value.st_uid ||
        old->value.st_gid != current->value.st_gid || old->value.st_ino != current->value.st_ino ||
        old->value.st_dev != current->value.st_dev || old->flags != current->flags ||
        old->generation != current->generation;
}
static char *watch_path(const adamic_string *path) {
    char *checked = path_bytes(path, true);
    if (checked == NULL) return NULL;
    bool absolute = checked[0] == '/';
    free(checked);
    adamic_string *cwd = absolute ? NULL : adamic_fs_file_host_cwd();
    if (adamic_thrown != NULL) return NULL;
    adamic_string *parts[2] = {cwd, (adamic_string *)path};
    adamic_string *normal = adamic_fs_file_host_join(absolute ? 1 : 2, absolute ? parts + 1 : parts);
    adamic_release(cwd);
    if (adamic_thrown != NULL) return NULL;
    char *result = bytes(normal);
    adamic_release(normal);
    return result;
}
static void watch_drop(watch_poller *poller) {
    watch_listener *listener = poller->first;
    while (listener != NULL) {
        watch_listener *next = listener->next;
        adamic_release(listener->callback); free(listener); listener = next;
    }
    free(poller->path); free(poller);
}
static void watch_retire(watch_poller *poller) {
    watch_poller **at = &watch_pollers;
    while (*at != poller) at = &(*at)->next;
    *at = poller->next; poller->active = false;
    (void)adamic_timer_cancel(poller->timer);
    if (!poller->firing) watch_drop(poller);
}
static void watch_notify(watch_poller *poller, const watch_snapshot *current) {
    size_t count = 0;
    for (watch_listener *listener = poller->first; listener != NULL; listener = listener->next) count++;
    adamic_closure **batch = malloc(count * sizeof *batch);
    if (batch == NULL && count != 0) abort();
    size_t index = 0;
    for (watch_listener *listener = poller->first; listener != NULL; listener = listener->next)
        batch[index++] = adamic_retain(listener->callback);
    adamic_object *curr = file_stat_value(&current->value);
    adamic_object *prev = file_stat_value(&poller->previous.value);
    adamic_value arguments[2] = {{.reference = curr}, {.reference = prev}};
    for (index = 0; index < count; index++) {
        (void)adamic_closure_call(batch[index], arguments, 2);
        adamic_release(batch[index]);
        if (adamic_thrown != NULL) {
            while (++index < count) adamic_release(batch[index]);
            break;
        }
    }
    adamic_release(curr); adamic_release(prev); free(batch);
}
// Poll tokens use watch_poll's invoke thunk, never the generic closure call path.
static adamic_value watch_token_not_callable(adamic_closure *token, adamic_value *arguments) {
    (void)token; (void)arguments;
    static const char message[] = "a watch poll token isn't callable";
    adamic_panic(message, sizeof message - 1);
    return (adamic_value){.number = 0};
}
static void watch_poll(adamic_closure *token, size_t argc, const adamic_value *arguments);
static void watch_schedule(watch_poller *poller, double delay, bool initial) {
    adamic_closure *token = adamic_closure_new(watch_token_not_callable, 0);
    adamic_value argument = {.number = (double)poller->identity};
    const bool reference = false;
    // Each fire explicitly schedules its next poll, so every timer is one-shot.
    poller->timer = adamic_timer_start(adamic_timers_provider(), token, watch_poll, delay, false, 1, &argument, &reference);
    adamic_release(token);
    // Node's initial asynchronous stat keeps the loop live even for an unref'd watcher.
    if (!initial && !poller->persistent) (void)adamic_timer_unref(poller->timer);
}
static void watch_poll(adamic_closure *token, size_t argc, const adamic_value *arguments) {
    (void)token; (void)argc;
    watch_poller *poller = watch_pollers;
    while (poller != NULL && poller->identity != (uintptr_t)arguments[0].number) poller = poller->next;
    if (poller == NULL) return;
    poller->firing = true;
    double began = watch_now();
    watch_snapshot current;
    int status = watch_stat(poller->path, &current);
    bool notify = status != 0 ? poller->status != status :
        poller->status != 0 && (poller->status < 0 || watch_changed(&poller->previous, &current));
    // StatWatcher's legacy initial -1 status suppresses an initial EPERM
    // whose zero nlink equals the previous zero snapshot.
    if (poller->status == 0 && status == -1 && poller->previous.value.st_nlink == 0) notify = false;
    if (status != 0) memset(&current, 0, sizeof current);
    if (notify) watch_notify(poller, &current);
    if (status == 0) poller->previous = current;
    poller->status = status == 0 ? 1 : status;
    poller->firing = false;
    if (!poller->active) { watch_drop(poller); return; }
    // libuv subtracts stat/listener time and schedules the next interval boundary.
    (void)adamic_timer_cancel(poller->timer);
    double remaining = poller->interval - fmod(watch_now() - began, poller->interval);
    if (adamic_thrown == NULL) watch_schedule(poller, remaining, false);
}
double adamic_fs_file_watch(const adamic_string *path, bool persistent, double interval, adamic_closure *callback) {
    if (!isfinite(interval) || interval < 0 || interval > INT32_MAX || floor(interval) != interval) {
        static const char message[] = "NotYet: watchFile interval outside the native timer provider's integer range";
        adamic_panic(message, sizeof message - 1);
    }
    char *normal = watch_path(path);
    if (normal == NULL) return 0;
    watch_poller *poller = watch_pollers;
    while (poller != NULL && strcmp(poller->path, normal) != 0) poller = poller->next;
    if (poller == NULL) {
        if (watch_next_identity > 9007199254740991ULL) abort();
        poller = calloc(1, sizeof *poller); if (poller == NULL) abort();
        *poller = (watch_poller){.identity = watch_next_identity++, .path = normal,
            .interval = interval < 1 ? 1 : (unsigned)interval, .persistent = persistent,
            .active = true, .next = watch_pollers};
        watch_pollers = poller;
        watch_schedule(poller, 1, true);
    } else free(normal);
    watch_listener *listener = malloc(sizeof *listener); if (listener == NULL) abort();
    *listener = (watch_listener){.callback = adamic_retain(callback)};
    if (poller->last == NULL) poller->first = listener; else poller->last->next = listener;
    poller->last = listener;
    return 0;
}
double adamic_fs_file_unwatch(const adamic_string *path, adamic_closure *callback) {
    char *normal = watch_path(path);
    if (normal == NULL) return 0;
    watch_poller *poller = watch_pollers;
    while (poller != NULL && strcmp(poller->path, normal) != 0) poller = poller->next;
    free(normal);
    if (poller == NULL) return 0;
    if (callback == NULL) { watch_retire(poller); return 0; }
    // EventEmitter removes the last matching registration, not every duplicate.
    watch_listener **match = NULL;
    for (watch_listener **at = &poller->first; *at != NULL; at = &(*at)->next)
        if ((*at)->callback == callback) match = at;
    if (match != NULL) {
        watch_listener *removed = *match; *match = removed->next;
        adamic_release(removed->callback); free(removed);
        poller->last = NULL;
        for (watch_listener *at = poller->first; at != NULL; at = at->next) poller->last = at;
        if (poller->first == NULL) watch_retire(poller);
    }
    return 0;
}
void adamic_fs_file_watch_shutdown(void) {
    while (watch_pollers != NULL) watch_retire(watch_pollers);
}
#endif
