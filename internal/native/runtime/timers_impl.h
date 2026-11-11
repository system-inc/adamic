/* Private to async.c: registrations and deadlines never cross threads. */
#include "timers.h"
#include "node_process.h"
#include <limits.h>
#include <math.h>
#include <time.h>

struct adamic_timer_provider { unsigned brand; };
static const adamic_timer_provider timer_provider = {1};
typedef struct timer_entry timer_entry;
typedef struct timer_list timer_list;
/* A timer is in timers while it's live, and, while it waits, in the list of its duration: start is
 * Node's _idleStart, the clock read its wait counts from. */
struct timer_entry {
    uintptr_t identity;
    double start;
    unsigned delay;
    bool repeat, referenced, firing, active;
    adamic_closure *callback;
    adamic_timer_invoke invoke;
    size_t argc;
    adamic_value *arguments;
    bool *references;
    timer_entry *next;
    timer_list *list;
    timer_entry *list_previous, *list_next;
};
/* Node's TimersList (lib/internal/timers.js, TimersList and insert, v24): one list per duration, its
 * timers in insertion order, an expiry and an id from one counter. Lists are taken by expiry, ties by
 * id (compareTimersLists), and every due timer of a list runs before the next list is chosen
 * (processTimers, listOnTimeout), so a later list's earlier deadline waits: a 5 ms list due first runs
 * both its timers before a 3 ms timer due between them (#a6yb7d1). Loop-thread only. */
struct timer_list {
    unsigned msecs;
    double expiry;
    uint64_t id;
    timer_entry *head, *tail;
    timer_list *next;
};
static timer_entry *timers;
static uintptr_t timer_next_identity = 1;
static timer_list *timer_lists;
static uint64_t timer_list_next_id;

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
static timer_list *timer_list_of(unsigned msecs) {
    for (timer_list *list = timer_lists; list != NULL; list = list->next)
        if (list->msecs == msecs) return list;
    return NULL;
}
static void timer_list_delete(timer_list *list) {
    timer_list **at = &timer_lists;
    while (*at != list) at = &(*at)->next;
    *at = list->next;
    free(list);
}
/* Node's insert: the timer's start is a clock read of its own (getLibuvNow reads the clock at every
 * setTimeout, which a timer armed after a busy wait shows by firing after one armed before it), its
 * list the one of its duration, made with expiry start + msecs and the next id when there's none. An
 * existing list keeps its expiry and id, an empty one too. */
static void timer_insert(timer_entry *entry, double start) {
    entry->start = start;
    timer_list *list = timer_list_of(entry->delay);
    if (list == NULL) {
        list = calloc(1, sizeof *list);
        if (list == NULL) abort();
        *list = (timer_list){.msecs = entry->delay, .expiry = start + entry->delay, .id = timer_list_next_id++, .next = timer_lists};
        timer_lists = list;
    }
    entry->list = list;
    entry->list_previous = list->tail;
    entry->list_next = NULL;
    if (list->tail != NULL) list->tail->list_next = entry;
    else list->head = entry;
    list->tail = entry;
}
/* Node's L.remove: a timer that isn't waiting in a list is left as it is. */
static void timer_list_remove(timer_entry *entry) {
    timer_list *list = entry->list;
    if (list == NULL) return;
    if (entry->list_previous != NULL) entry->list_previous->list_next = entry->list_next;
    else list->head = entry->list_next;
    if (entry->list_next != NULL) entry->list_next->list_previous = entry->list_previous;
    else list->tail = entry->list_previous;
    entry->list = NULL;
    entry->list_previous = entry->list_next = NULL;
}
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
/* Node's Timeout warns on stderr for a delay that's NaN, negative or past 2^31 - 1 ms, in Node's
 * words with this process's id (Adamic's ruling on #995spev, Nov 7, lifted by #zhgmbwz): NaN and a
 * negative delay once a process, an overflowing one every time, the first warning of any kind followed
 * by Node's hint line. process.emitWarning writes on a nextTick, so the text waits here until
 * timer_warnings_flush, called where Node's ticks would run: after a timer's callback, and after each
 * drain of the promise jobs (async.c). Loop-thread only, as the registrations are. */
static char *timer_warnings;
static size_t timer_warnings_length, timer_warnings_capacity;
static bool timer_nan_warned, timer_negative_warned, timer_hint_written;
static void timer_warnings_append(const char *text, size_t length) {
    if (length > SIZE_MAX - timer_warnings_length) abort();
    if (timer_warnings_length + length > timer_warnings_capacity) {
        size_t capacity = timer_warnings_capacity == 0 ? 256 : timer_warnings_capacity;
        while (capacity < timer_warnings_length + length) {
            if (capacity > SIZE_MAX / 2) abort();
            capacity *= 2;
        }
        char *grown = realloc(timer_warnings, capacity);
        if (grown == NULL) abort();
        timer_warnings = grown;
        timer_warnings_capacity = capacity;
    }
    memcpy(timer_warnings + timer_warnings_length, text, length);
    timer_warnings_length += length;
}
/* Node's process name in the hint is basename(process.argv0), which the oracle runs as node. */
static void timer_warn(const char *name, double delay_ms, const char *reason) {
    static const char duration[] = "\nTimeout duration was set to 1.\n";
    static const char hint[] = "(Use `node --trace-warnings ...` to show where the warning was created)\n";
    char number[ADAMIC_NUMBER_FORMAT_MAX];
    timer_warnings_append("(node:", 6);
    timer_warnings_append(number, adamic_number_format(adamic_node_pid(), number));
    timer_warnings_append(") ", 2);
    timer_warnings_append(name, strlen(name));
    timer_warnings_append(": ", 2);
    timer_warnings_append(number, adamic_number_format(delay_ms, number));
    timer_warnings_append(reason, strlen(reason));
    timer_warnings_append(duration, sizeof duration - 1);
    if (!timer_hint_written) {
        timer_hint_written = true;
        timer_warnings_append(hint, sizeof hint - 1);
    }
}
/* Writes what timer_warn queued through the output lock, after stdout's buffered text and in
 * stderr's order, as Node's tick would. */
static void timer_warnings_flush(void) {
    if (timer_warnings_length == 0) return;
    adamic_string text = ADAMIC_STRING_BYTES(timer_warnings, timer_warnings_length);
    (void)adamic_write_raw(adamic_stderr, &text);
    timer_warnings_length = 0;
}
adamic_timer_handle adamic_timer_start(const adamic_timer_provider *provider,
    adamic_closure *callback, adamic_timer_invoke invoke, double delay_ms,
    bool repeat, size_t argc, const adamic_value *arguments, const bool *references) {
    host_loop_thread();
    if (provider != &timer_provider || callback == NULL || invoke == NULL || host_closed ||
        (argc != 0 && (arguments == NULL || references == NULL)) ||
        argc > SIZE_MAX / sizeof(adamic_value) || timer_next_identity == UINTPTR_MAX) abort();
    // Timers are not admitted under custom host loop hooks; they do not fire late.
    if (host_wait_hook != host_default_wait) {
        static const char message[] = "NotYet: timers with custom host loop hooks";
        adamic_panic(message, sizeof message - 1);
    }
    /* Node's Timeout constructor's branch, its order of tests and its 1 ms. */
    if (!(delay_ms >= 1 && delay_ms <= INT32_MAX)) {
        if (delay_ms > INT32_MAX) {
            timer_warn("TimeoutOverflowWarning", delay_ms, " does not fit into a 32-bit signed integer.");
        } else if (delay_ms < 0 && !timer_negative_warned) {
            timer_negative_warned = true;
            timer_warn("TimeoutNegativeWarning", delay_ms, " is a negative number.");
        } else if (isnan(delay_ms) && !timer_nan_warned) {
            timer_nan_warned = true;
            timer_warn("TimeoutNaNWarning", delay_ms, " is not a number.");
        }
        delay_ms = 1;
    }
    unsigned delay = (unsigned)delay_ms;
    timer_entry *entry = calloc(1, sizeof *entry);
    if (entry == NULL) abort();
    *entry = (timer_entry){.identity = timer_next_identity++,
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
    timer_insert(entry, timer_now());
    host_started = true; /* Freeze loop hooks; timer operations bind the loop owner. */
    return (adamic_timer_handle){provider, entry->identity};
}
bool adamic_timer_cancel(adamic_timer_handle handle) {
    host_loop_thread();
    timer_entry *entry = timer_find(handle);
    if (entry == NULL) return false;
    timer_unlink(entry);
    /* Node's unenroll (lib/timers.js): L.remove, then, for a referenced timer alone, its duration's list
     * deleted when it's empty, even when the timer wasn't in it (a repeat cleared in its own callback);
     * an unreferenced timer's empty list stays, with its expiry and id, for the next insert to reuse. */
    timer_list_remove(entry);
    if (entry->referenced) {
        timer_list *list = timer_list_of(entry->delay);
        if (list != NULL && list->head == NULL) timer_list_delete(list);
    }
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
/* A Timeout's shape: only adamic_node_set_timeout makes an object of it, so its address alone tells
 * clearTimeout a handle from a plain object with a number in its one field. The field is private, as
 * a collator's #compare is, and fields.go's layout proof holds its slot. */
static const char *const timeout_names[] = {"#identity"};
static const bool timeout_references[] = {false};
static const adamic_field_kind timeout_kinds[] = {adamic_field_number};
static const adamic_shape timeout_shape = {1, timeout_names, timeout_references, &adamic_host_methods, timeout_kinds};
adamic_object *adamic_node_set_timeout(adamic_closure *callback, adamic_timer_invoke invoke,
    double delay_ms, size_t argc, const adamic_value *arguments, const bool *references) {
    adamic_timer_handle handle = adamic_timer_start(&timer_provider, callback, invoke, delay_ms, false, argc, arguments, references);
    /* The slot holds the identity as a double, exact through 2^53. A uintptr_t narrower than that
     * (wasm32's 32 bits) can't pass it, and the comparison would be one -Werror refuses as always false. */
#if UINTPTR_MAX > 9007199254740991u
    if (handle.identity > 9007199254740991u) adamic_invariant("setTimeout: a timer identity past 2^53");
#endif
    adamic_object *timeout = adamic_object_new(&timeout_shape);
    timeout->slots[0].number = (double)handle.identity;
    return timeout;
}
/* ToBoolean of a slot by how its field holds it. An absent number | undefined is a NaN and an absent
 * boolean | undefined is 2, both false, as undefined is. */
static bool timer_slot_truthy(bool reference, adamic_field_kind kind, adamic_value slot) {
    if (reference) return adamic_census_to_boolean(slot.reference);
    if (kind == adamic_field_boolean) return slot.maybe_boolean == 1;
    return slot.number != 0 && !isnan(slot.number);
}
/* Node's clearTimeout unenrolls any value whose _onTimeout is truthy, writing its fields, which a
 * no-op wouldn't (Adamic's ruling on #995spev, Nov 7), so an object that isn't a Timeout but holds a
 * truthy _onTimeout, found by its name as an object's own field or a record's key, stops. A runtime
 * host shape holds none. A method or accessor of the name isn't a field: lowering refuses the clear
 * where the program declares one (library_node_timers.go). */
static bool timer_foreign_handle(const adamic_object *object) {
    if (adamic_record_is(object)) {
        adamic_string key = ADAMIC_STRING("_onTimeout");
        const adamic_value *slot = adamic_record_get(object, &key);
        adamic_field_kind kind = adamic_record_values(object);
        return slot != NULL && timer_slot_truthy(kind == adamic_field_reference, kind, *slot);
    }
    if (object->shape->methods == &adamic_host_methods) return false;
    for (size_t index = 0; index < object->shape->count; index++) {
        if (strcmp(object->shape->names[index], "_onTimeout") != 0 || !adamic_object_present(object, index)) continue;
        adamic_field_kind kind = object->shape->kinds == NULL ? adamic_field_number : object->shape->kinds[index];
        return timer_slot_truthy(object->shape->references[index], kind, object->slots[index]);
    }
    return false;
}
double adamic_node_clear_timeout(const adamic_heap *value) {
    if (value == NULL || value->kind != adamic_kind_object) return 0;
    const adamic_object *object = (const adamic_object *)value;
    if (object->shape != &timeout_shape) {
        if (timer_foreign_handle(object)) {
            static const char message[] = "clearTimeout of an object that isn't a timer but holds _onTimeout isn't built yet; pass clearTimeout only what setTimeout or setInterval returned, or undefined";
            adamic_panic(message, sizeof message - 1);
        }
        return 0;
    }
    /* A fired or cleared Timeout's identity is stale, which cancel answers false. */
    (void)adamic_timer_cancel((adamic_timer_handle){&timer_provider, (uintptr_t)object->slots[0].number});
    return 0;
}
static bool timer_live(void) {
    for (timer_entry *entry = timers; entry != NULL; entry = entry->next)
        if (entry->referenced) return true;
    return false;
}
/* Only the native poll wait reads a deadline; WASI's wait has none, so the
 * strict WASI build would reject an unused definition. */
#ifndef ADAMIC_TARGET_WASI
/* Includes unref timers: while other work lives, their deadlines still run. */
static int timer_wait_ms(void) {
    if (timer_lists == NULL) return -1;
    /* Node schedules its loop's timer for the earliest list's expiry, an empty list's among them. */
    double expiry = timer_lists->expiry;
    for (timer_list *list = timer_lists->next; list != NULL; list = list->next)
        if (list->expiry < expiry) expiry = list->expiry;
    /* Millisecond rounding and poll's integer timeout can wake up to 1 ms
     * early. A zero-timeout turn is harmless: dispatch rechecks the expiry. */
    double delay = expiry - timer_now();
    return delay <= 0 ? 0 : delay > INT_MAX ? INT_MAX : (int)delay;
}
#endif
/* Node's processTimers(now) and listOnTimeout(list, now), one callback a call, so the loop drains the
 * jobs between callbacks as Node's runNextTicks does; now is the turn's one read, as processTimers'.
 * The list chosen stays the earliest while its timers run: a list made in a callback has a later
 * expiry, and an insert into an existing list leaves its expiry as it was. */
static bool timer_dispatch(double now) {
    for (;;) {
        timer_list *list = NULL;
        for (timer_list *each = timer_lists; each != NULL; each = each->next)
            if (list == NULL || each->expiry < list->expiry || (each->expiry == list->expiry && each->id < list->id)) list = each;
        if (list == NULL || list->expiry > now) return false;
        timer_entry *chosen = list->head;
        if (chosen == NULL) {
            /* Every due timer has run, or its timers were cleared unreferenced: the list goes. */
            timer_list_delete(list);
            continue;
        }
        if (now - chosen->start < list->msecs) {
            /* The rest of the list isn't due yet: it waits for its head, with a new id. */
            list->expiry = fmax(chosen->start + list->msecs, now + 1);
            list->id = timer_list_next_id++;
            continue;
        }
        timer_list_remove(chosen);
        /* A repeat's next wait counts from a clock read before its callback, as Node's start does. */
        double start = chosen->repeat ? timer_now() : 0;
        chosen->firing = true;
        if (!chosen->repeat) timer_unlink(chosen); /* A timeout is retired before its invocation. */
        chosen->invoke(chosen->callback, chosen->argc, chosen->arguments);
        chosen->firing = false;
        if (chosen->repeat && chosen->active) timer_insert(chosen, start);
        if (!chosen->active) timer_drop(chosen);
        if (adamic_thrown != NULL) adamic_uncaught();
        /* Node runs a callback's ticks before its promise jobs, so a warning its setTimeout gave is
         * written here, ahead of the drain that follows. */
        timer_warnings_flush();
        return true;
    }
}
static void timer_shutdown(void) {
    while (timers != NULL) {
        timer_entry *entry = timers;
        timer_unlink(entry);
        timer_drop(entry);
    }
    while (timer_lists != NULL) timer_list_delete(timer_lists);
    free(timer_warnings);
    timer_warnings = NULL;
    timer_warnings_length = timer_warnings_capacity = 0;
}
