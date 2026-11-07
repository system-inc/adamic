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
    // Timers are not admitted under custom host loop hooks; they do not fire late.
    if (host_wait_hook != host_default_wait) {
        static const char message[] = "NotYet: timers with custom host loop hooks";
        adamic_panic(message, sizeof message - 1);
    }
    /* Node makes such a delay 1 ms with a warning on stderr; a run-time NotYet (Adamic's ruling on
     * #995spev), said in the JavaScript backend's words too. */
    if (!isfinite(delay_ms) || delay_ms < 0 || delay_ms > INT32_MAX) {
        static const char message[] = "a timer delay outside Node's range (NaN, negative or past 2^31 - 1 ms) isn't built yet";
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
    /* The slot holds the identity as a double, exact through 2^53. */
    if (handle.identity > 9007199254740991u) adamic_invariant("setTimeout: a timer identity past 2^53");
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
            static const char message[] = "clearTimeout of an object that isn't a timer but holds _onTimeout isn't built yet";
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
    if (timers == NULL) return -1;
    double deadline = timers->deadline;
    for (timer_entry *entry = timers->next; entry != NULL; entry = entry->next)
        if (entry->deadline < deadline) deadline = entry->deadline;
    /* Millisecond rounding and poll's integer timeout can wake up to 1 ms
     * early. A zero-timeout turn is harmless: dispatch rechecks the deadline. */
    double delay = deadline - timer_now();
    return delay <= 0 ? 0 : delay > INT_MAX ? INT_MAX : (int)delay;
}
#endif
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
