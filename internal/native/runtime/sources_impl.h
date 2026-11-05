/* Private to async.c: registrations never cross threads. The default wait polls
 * every source after the two pipes and records what it saw; the next turn
 * dispatches from that record, never from a fresh poll. */
#include "sources.h"
#ifndef ADAMIC_TARGET_WASI
#include <fcntl.h>
#include <poll.h>

typedef struct source_entry source_entry;
struct source_entry {
    uintptr_t identity;
    int descriptor;
    unsigned ready;
    bool referenced, firing, active;
    adamic_source_dispatch dispatch;
    void *data;
    source_entry *next;
};
static source_entry *sources;
static uintptr_t source_next_identity = 1;

static source_entry *source_find(adamic_source_handle handle) {
    for (source_entry *entry = sources; entry != NULL; entry = entry->next)
        if (entry->identity == handle.identity) return entry;
    return NULL;
}
static void source_unlink(source_entry *entry) {
    source_entry **at = &sources;
    while (*at != entry) at = &(*at)->next;
    *at = entry->next;
    entry->next = NULL;
    entry->active = false;
}
adamic_source_handle adamic_source_register(int descriptor,
    adamic_source_dispatch dispatch, void *data, bool referenced) {
    host_loop_thread();
    if (descriptor < 0 || dispatch == NULL || host_closed || source_next_identity == UINTPTR_MAX) abort();
    // Sources are not admitted under custom host loop hooks; nothing would poll them.
    if (host_wait_hook != host_default_wait) {
        static const char message[] = "NotYet: readiness sources with custom host loop hooks";
        adamic_panic(message, sizeof message - 1);
    }
    /* A dispatch that blocked on its read would stall every other source and timer. */
    int flags = fcntl(descriptor, F_GETFL);
    if (flags == -1 || (flags & O_NONBLOCK) == 0) {
        fputs("adamic: readiness source descriptor must be non-blocking\n", stderr);
        abort();
    }
    source_entry *entry = malloc(sizeof *entry);
    if (entry == NULL) abort();
    *entry = (source_entry){.identity = source_next_identity++, .descriptor = descriptor,
        .referenced = referenced, .active = true, .dispatch = dispatch, .data = data, .next = sources};
    sources = entry;
    host_started = true; /* Freeze loop hooks; source operations bind the loop owner. */
    return (adamic_source_handle){entry->identity};
}
bool adamic_source_unref(adamic_source_handle handle) {
    host_loop_thread();
    source_entry *entry = source_find(handle);
    if (entry == NULL) return false;
    entry->referenced = false;
    return true;
}
bool adamic_source_close(adamic_source_handle handle) {
    host_loop_thread();
    source_entry *entry = source_find(handle);
    if (entry == NULL) return false;
    source_unlink(entry);
    if (!entry->firing) free(entry);
    return true;
}
static bool source_live(void) {
    for (source_entry *entry = sources; entry != NULL; entry = entry->next)
        if (entry->referenced) return true;
    return false;
}
/* Includes unreferenced sources: while other work lives, they are still polled. */
static size_t source_count(void) {
    size_t count = 0;
    for (source_entry *entry = sources; entry != NULL; entry = entry->next) count++;
    return count;
}
/* The wait fills its poll set and reads it back in list order. Nothing runs
 * between the two, so the list cannot change under either walk. */
static void source_poll_set(struct pollfd *descriptors) {
    for (source_entry *entry = sources; entry != NULL; entry = entry->next)
        *descriptors++ = (struct pollfd){.fd = entry->descriptor, .events = POLLIN};
}
static void source_record(const struct pollfd *descriptors) {
    for (source_entry *entry = sources; entry != NULL; entry = entry->next, descriptors++)
        entry->ready = ((descriptors->revents & POLLIN) != 0 ? adamic_source_readable : 0) |
            ((descriptors->revents & (POLLERR | POLLNVAL | POLLHUP)) != 0 ? adamic_source_error : 0);
}
/* Each call takes a fresh walk, so a dispatch may close any source, itself
 * included: one it closes is gone from the list before the next choice. */
static bool source_dispatch(void) {
    source_entry *chosen = NULL;
    for (source_entry *entry = sources; entry != NULL; entry = entry->next)
        if (entry->ready != 0 && (chosen == NULL || entry->identity < chosen->identity)) chosen = entry;
    if (chosen == NULL) return false;
    unsigned ready = chosen->ready;
    chosen->ready = 0; /* At most once a turn: only the next wait records it again. */
    chosen->firing = true;
    chosen->dispatch(chosen->data, ready);
    chosen->firing = false;
    if (!chosen->active) free(chosen);
    if (adamic_thrown != NULL) adamic_uncaught();
    return true;
}
static void source_shutdown(void) {
    while (sources != NULL) {
        source_entry *entry = sources;
        source_unlink(entry);
        free(entry);
    }
}
#else
/* Nothing registers a source under WASI, so none is ever live or ready. */
static bool source_live(void) { return false; }
static bool source_dispatch(void) { return false; }
static void source_shutdown(void) {}
#endif
