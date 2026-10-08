/* Private host bridge, included only by async.c. Counts and values stay on the
 * loop. The mutex protects identities, copied buffers, and queue publication. */
#include <pthread.h>
#include <stdio.h>
#include <string.h>
#ifndef ADAMIC_TARGET_WASI
#include <errno.h>
#include <fcntl.h>
#include <unistd.h>
#include <poll.h>
#endif

typedef struct host_entry host_entry;
struct host_entry {
    uintptr_t identity;
    adamic_promise *promise;
    unsigned char *bytes;
    size_t length;
    double status;
    bool queued, rejected, abandoned;
    host_entry *next, *completion;
};
static pthread_mutex_t host_mutex = PTHREAD_MUTEX_INITIALIZER;
static host_entry *host_requests, *host_completions, *host_completions_last;
static uintptr_t host_next_identity = 1;
static bool host_started, host_closed, host_hooks_set, host_owner_set;
static pthread_t host_owner;
#ifndef ADAMIC_TARGET_WASI
static int host_pipe[2] = {-1, -1};
#endif
static void host_default_wake(void);
static void host_default_wait(void);
static adamic_host_loop_hook host_wake_hook = host_default_wake;
static adamic_host_loop_hook host_wait_hook = host_default_wait;

static void host_lock(void) {
    (void)host_mutex;
#ifndef ADAMIC_HOST_QUEUE_LOCK_MUTANT
    if (pthread_mutex_lock(&host_mutex) != 0) abort();
#endif
}
static void host_unlock(void) {
#ifndef ADAMIC_HOST_QUEUE_LOCK_MUTANT
    if (pthread_mutex_unlock(&host_mutex) != 0) abort();
#endif
}
static void host_loop_thread(void) {
    if (!host_owner_set) {
        host_owner = pthread_self();
        host_owner_set = true;
    } else if (!pthread_equal(host_owner, pthread_self())) {
        fputs("adamic: host promise operation requires the loop thread\n", stderr);
        abort();
    }
}
#ifdef ADAMIC_TARGET_WASI
/* WASI preview 1 has no pipe and no worker pool. A host using promises must
 * provide both loop hooks; synchronous async programs never enter these. */
static void host_pipe_start(void) {
    static const char message[] = "WASI host promises require wake and wait loop hooks";
    adamic_panic(message, sizeof message - 1);
}
static void host_default_wake(void) { host_pipe_start(); }
static void host_default_wait(void) { host_pipe_start(); }
#else
static void host_pipe_start(void) {
    if (host_pipe[0] != -1) return;
    if (pipe(host_pipe) != 0) abort();
    for (unsigned index = 0; index < 2; index++) {
        int flags = fcntl(host_pipe[index], F_GETFL);
        if (flags == -1 || fcntl(host_pipe[index], F_SETFL, flags | O_NONBLOCK) == -1 ||
            fcntl(host_pipe[index], F_SETFD, FD_CLOEXEC) == -1) abort();
    }
}
static void host_default_wake(void) {
    /* Called under host_mutex: shutdown cannot close/reuse this descriptor. */
    unsigned char byte = 1;
    ssize_t result;
    do { result = write(host_pipe[1], &byte, 1); } while (result == -1 && errno == EINTR);
    if (result == -1 && errno != EAGAIN && errno != EWOULDBLOCK) abort();
}
static void host_default_wait(void) {
    /* The queue is the predicate. A byte published before poll stays readable. */
    host_lock();
    bool ready = host_completions != NULL || host_requests == NULL;
    host_unlock();
    if (ready) return;
    struct pollfd descriptor = {.fd = host_pipe[0], .events = POLLIN};
    int result;
    do { result = poll(&descriptor, 1, -1); } while (result == -1 && errno == EINTR);
    if (result < 0 || (descriptor.revents & (POLLERR | POLLNVAL))) abort();
}
#endif
bool adamic_host_set_loop_hooks(adamic_host_loop_hook wake, adamic_host_loop_hook wait) {
    host_loop_thread();
    if (host_hooks_set || host_started || host_closed) return false;
    host_wake_hook = wake != NULL ? wake : host_default_wake;
    host_wait_hook = wait != NULL ? wait : host_default_wait;
    host_hooks_set = true;
    return true;
}
adamic_host_request *adamic_host_promise_new(adamic_promise **out) {
    host_loop_thread();
    if (out == NULL) abort();
    host_lock();
    if (host_closed || host_next_identity == UINTPTR_MAX) abort();
    host_started = true;
    if (host_wake_hook == host_default_wake || host_wait_hook == host_default_wait) host_pipe_start();
    host_entry *entry = calloc(1, sizeof *entry);
    if (entry == NULL) abort();
    entry->identity = host_next_identity++;
    entry->promise = adamic_async_new();
    *out = adamic_retain(entry->promise);
    entry->next = host_requests;
    host_requests = entry;
    host_unlock();
    /* Supported clang targets provide uintptr_t pointer conversion. Identity
     * tokens are compared only, never dereferenced. Never reuse a retired token:
     * unlike allocation addresses this cannot alias a later live request. */
    return (adamic_host_request *)entry->identity;
}
static host_entry *host_find(adamic_host_request *request) {
    uintptr_t identity = (uintptr_t)request;
    for (host_entry *entry = host_requests; entry != NULL; entry = entry->next)
        if (entry->identity == identity) return entry;
    return NULL;
}
static void host_publish(host_entry *entry) {
    entry->queued = true;
    if (host_completions_last != NULL) host_completions_last->completion = entry;
    else host_completions = entry;
    host_completions_last = entry;
#ifndef ADAMIC_HOST_WAKE_MUTANT
    host_wake_hook();
#endif
}
static bool host_complete(adamic_host_request *request, const void *bytes,
                          size_t length, double status, bool rejected) {
    host_lock();
    host_entry *entry = host_find(request);
    if (entry == NULL || entry->queued) {
        host_unlock();
        fputs("adamic: host completion refused: request retired or already completed\n", stderr);
        return false;
    }
    if ((bytes == NULL && length != 0) || length > SIZE_MAX - sizeof(adamic_string)) {
        host_unlock();
        fputs("adamic: host completion refused: invalid buffer\n", stderr);
        return false;
    }
    entry->bytes = malloc(length == 0 ? 1 : length);
    if (entry->bytes == NULL) abort();
    if (length != 0) memcpy(entry->bytes, bytes, length);
    entry->length = length;
    entry->status = status;
    entry->rejected = rejected;
#ifdef ADAMIC_HOST_DIRECT_SETTLE_MUTANT
    /* Deliberately violate the boundary, without an affinity guard, so TSan
     * must observe the loop-local promise/count race rather than an assertion. */
    adamic_async_settle(entry->promise, (adamic_value){.number = status}, false, rejected);
#endif
    host_publish(entry);
    host_unlock();
    return true;
}
bool adamic_host_resolve(adamic_host_request *request, const void *bytes, size_t length, double status) {
    return host_complete(request, bytes, length, status, false);
}
bool adamic_host_reject(adamic_host_request *request, const char *message) {
    if (message == NULL) return false;
    return host_complete(request, message, strlen(message), 0, true);
}
void adamic_host_abandon(adamic_host_request *request) {
    host_lock();
    host_entry *entry = host_find(request);
    if (entry != NULL && !entry->queued) {
        entry->abandoned = true;
        host_publish(entry);
    }
    host_unlock();
}
static bool host_live(void) {
    host_lock();
    bool live = host_requests != NULL;
    host_unlock();
    return live;
}
static bool host_valid_utf8(const unsigned char *bytes, size_t length) {
    for (size_t at = 0; at < length;) {
        unsigned first = bytes[at++], remaining;
        uint32_t code, minimum;
        if (first < 0x80) continue;
        if (first >= 0xc2 && first <= 0xdf) { remaining = 1; code = first & 0x1f; minimum = 0x80; }
        else if (first >= 0xe0 && first <= 0xef) { remaining = 2; code = first & 0xf; minimum = 0x800; }
        else if (first >= 0xf0 && first <= 0xf4) { remaining = 3; code = first & 7; minimum = 0x10000; }
        else return false;
        if (remaining > length - at) return false;
        while (remaining-- != 0) {
            unsigned next = bytes[at++];
            if ((next & 0xc0) != 0x80) return false;
            code = (code << 6) | (next & 0x3f);
        }
        if (code < minimum || code > 0x10ffff || (code >= 0xd800 && code <= 0xdfff)) return false;
    }
    return true;
}
static void host_settle(host_entry *entry) {
    if (entry->abandoned) return; /* Exit owns subscriptions once liveness ends. */
    bool rejected = entry->rejected;
    const unsigned char *bytes = entry->bytes;
    size_t length = entry->length;
    if (!host_valid_utf8(bytes, length)) {
        static const unsigned char invalid[] = "host payload is not valid UTF-8";
        bytes = invalid; length = sizeof invalid - 1; rejected = true;
    }
    adamic_string *text = adamic_string_allocate(length);
    if (length != 0) memcpy((char *)text->bytes, bytes, length);
    adamic_string_check_length(adamic_string_length(text));
    adamic_object *value;
    if (rejected) {
        value = adamic_error_new(text);
        adamic_release(text);
    } else {
        static const char *const names[] = {"body", "status"};
        static const bool references[] = {true, false};
        static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_number};
        static const adamic_shape shape = {2, names, references, NULL, kinds, NULL, NULL};
        value = adamic_object_new(&shape);
        value->slots[0].reference = text; /* Move the owned string into the result. */
        value->slots[1].number = entry->status;
    }
    adamic_async_settle(entry->promise, (adamic_value){.reference = value}, true, rejected);
    adamic_release(value);
}
static void host_drop(host_entry *entry) {
    adamic_release(entry->promise);
    free(entry->bytes);
    free(entry);
}
void adamic_host_process_completions(void) {
    if (host_started) host_loop_thread();
#ifndef ADAMIC_TARGET_WASI
    /* Only the loop drains pipe bytes. Workers publish while holding the mutex;
     * a publication after this read either appears below or leaves a new byte. */
    if (host_pipe[0] != -1) {
        unsigned char bytes[256];
        ssize_t result;
        do { result = read(host_pipe[0], bytes, sizeof bytes); }
        while (result > 0 || (result == -1 && errno == EINTR));
        if (result == -1 && errno != EAGAIN && errno != EWOULDBLOCK) abort();
    }
#endif
    for (;;) {
        host_lock();
        host_entry *entry = host_completions;
        if (entry != NULL) {
            host_completions = entry->completion;
            if (host_completions == NULL) host_completions_last = NULL;
            host_entry **at = &host_requests;
            while (*at != entry) at = &(*at)->next;
            *at = entry->next;
        }
        host_unlock();
        if (entry == NULL) break;
        host_settle(entry);
        host_drop(entry);
#ifndef ADAMIC_HOST_CHECKPOINT_MUTANT
        drain_jobs(); /* Host callback checkpoint, before the next completion. */
#endif
    }
    drain_jobs();
}
static void host_shutdown(void) {
    if (host_started) host_loop_thread();
    host_lock();
    host_closed = true;
    host_entry *entry = host_requests;
    host_requests = host_completions = host_completions_last = NULL;
#ifndef ADAMIC_TARGET_WASI
    for (unsigned index = 0; index < 2; index++) {
        if (host_pipe[index] != -1) close(host_pipe[index]);
        host_pipe[index] = -1;
    }
#endif
    host_unlock();
    while (entry != NULL) {
        host_entry *next = entry->next;
#ifndef ADAMIC_HOST_EXIT_REGISTRY_MUTANT
        host_drop(entry);
#endif
        entry = next;
    }
}
