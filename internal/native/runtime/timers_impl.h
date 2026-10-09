/* Private to async.c: registrations and deadlines never cross threads. */
#include "timers.h"
#include <limits.h>
#include <math.h>
#include <time.h>

struct adamic_timer_provider { unsigned brand; };
static const adamic_timer_provider timer_provider = {1};
typedef struct timer_entry timer_entry;
struct timer_entry {
    uintptr_t identity;
    double deadline;
    unsigned delay;
    bool repeat, referenced, firing, active;
    adamic_closure *callback;
    adamic_timer_invoke invoke;
    size_t argc;
    adamic_value *arguments;
    bool *references;
    timer_entry *next;
};
static timer_entry *timers;
static uintptr_t timer_next_identity = 1;

static double timer_now(void) {
#ifdef ADAMIC_TARGET_WASI
    static const char message[] = "NotYet: native timers require a monotonic CLI event loop";
    adamic_panic(message, sizeof message - 1);
#else
    struct timespec now;
    if (clock_gettime(CLOCK_MONOTONIC, &now) != 0) abort();
    return floor((double)now.tv_sec * 1000 + (double)now.tv_nsec / 1000000);
#endif
}
const adamic_timer_provider *adamic_timers_provider(void) { return &timer_provider; }
static timer_entry *timer_find(adamic_timer_handle handle) {
    if (handle.provider != &timer_provider) return NULL;
    for (timer_entry *entry = timers; entry != NULL; entry = entry->next)
        if (entry->identity == handle.identity) return entry;
    return NULL;
}
static void timer_unlink(timer_entry *entry) {
    timer_entry **at = &timers;
    while (*at != entry) at = &(*at)->next;
    *at = entry->next;
    entry->next = NULL;
    entry->active = false;
}
static void timer_drop(timer_entry *entry) {
    adamic_release(entry->callback);
    for (size_t index = 0; index < entry->argc; index++)
        if (entry->references[index]) adamic_release(entry->arguments[index].reference);
    free(entry->arguments);
    free(entry->references);
    free(entry);
}
adamic_timer_handle adamic_timer_start(const adamic_timer_provider *provider,
    adamic_closure *callback, adamic_timer_invoke invoke, double delay_ms,
    bool repeat, size_t argc, const adamic_value *arguments, const bool *references) {
    host_loop_thread();
    if (provider != &timer_provider || callback == NULL || invoke == NULL || host_closed ||
        (argc != 0 && (arguments == NULL || references == NULL)) ||
        argc > SIZE_MAX / sizeof(adamic_value) || timer_next_identity == UINTPTR_MAX) abort();
    if (host_wait_hook != host_default_wait) {
        static const char message[] = "NotYet: timers with custom host loop hooks";
        adamic_panic(message, sizeof message - 1);
    }
    if (!isfinite(delay_ms) || delay_ms < 0 || delay_ms > INT32_MAX) {
        static const char message[] = "NotYet: timer delay requiring Node warnings";
        adamic_panic(message, sizeof message - 1);
    }
    unsigned delay = delay_ms < 1 ? 1 : (unsigned)delay_ms;
    double deadline = timer_now() + delay;
    timer_entry *entry = calloc(1, sizeof *entry);
    if (entry == NULL) abort();
    *entry = (timer_entry){.identity = timer_next_identity++, .deadline = deadline,
        .delay = delay, .repeat = repeat, .referenced = true, .active = true,
        .callback = adamic_retain(callback), .invoke = invoke, .argc = argc, .next = timers};
    if (argc != 0) {
        entry->arguments = malloc(argc * sizeof *entry->arguments);
        entry->references = malloc(argc * sizeof *entry->references);
        if (entry->arguments == NULL || entry->references == NULL) abort();
        memcpy(entry->arguments, arguments, argc * sizeof *arguments);
        memcpy(entry->references, references, argc * sizeof *references);
        for (size_t index = 0; index < argc; index++)
            if (references[index]) adamic_retain(arguments[index].reference);
    }
    timers = entry;
    host_started = true; /* Freeze loop hooks; timer operations bind the loop owner. */
    return (adamic_timer_handle){provider, entry->identity};
}
bool adamic_timer_cancel(adamic_timer_handle handle) {
    host_loop_thread();
    timer_entry *entry = timer_find(handle);
    if (entry == NULL) return false;
    timer_unlink(entry);
    if (!entry->firing) timer_drop(entry);
    return true;
}
bool adamic_timer_unref(adamic_timer_handle handle) {
    host_loop_thread();
    timer_entry *entry = timer_find(handle);
    if (entry == NULL) return false;
    entry->referenced = false;
    return true;
}
static bool timer_live(void) {
    for (timer_entry *entry = timers; entry != NULL; entry = entry->next)
        if (entry->referenced) return true;
    return false;
}
/* Includes unref timers: while other work lives, their deadlines still run. */
static int timer_wait_ms(void) {
    if (timers == NULL) return -1;
    double deadline = timers->deadline;
    for (timer_entry *entry = timers->next; entry != NULL; entry = entry->next)
        if (entry->deadline < deadline) deadline = entry->deadline;
    double delay = deadline - timer_now();
    return delay <= 0 ? 0 : delay > INT_MAX ? INT_MAX : (int)delay;
}
static bool timer_dispatch(void) {
    timer_entry *chosen = NULL;
    double now = timer_now();
    for (timer_entry *entry = timers; entry != NULL; entry = entry->next)
        if (entry->deadline <= now && (chosen == NULL || entry->deadline < chosen->deadline ||
            (entry->deadline == chosen->deadline && entry->identity < chosen->identity))) chosen = entry;
    if (chosen == NULL) return false;
    chosen->firing = true;
    if (chosen->repeat) chosen->deadline = now + chosen->delay;
    else timer_unlink(chosen); /* A timeout is retired before its invocation. */
    chosen->invoke(chosen->callback, chosen->argc, chosen->arguments);
    chosen->firing = false;
    if (!chosen->active) timer_drop(chosen);
    if (adamic_thrown != NULL) adamic_uncaught();
    return true;
}
static void timer_shutdown(void) {
    while (timers != NULL) {
        timer_entry *entry = timers;
        timer_unlink(entry);
        timer_drop(entry);
    }
}
