#include "async.h"
#include <assert.h>
#include <stdlib.h>

/* Loop-local internal edges are audited here, never exposed as writable source fields. */
struct adamic_async_reaction {
    adamic_heap heap;
    adamic_async_frame *frame;
    adamic_async_promise *promise;
    adamic_async_reaction *next;
};
typedef struct pending pending;
struct pending { adamic_async_promise *promise; pending *next; };
static pending *subscriptions;
static adamic_async_reaction *jobs, *jobs_last;

/* Compiler-private abandonment cleanup; records borrow a frame, never own a protocol edge. */
typedef struct cleanup_record cleanup_record;
struct cleanup_record { adamic_async_frame *frame; void (*cleanup)(adamic_async_frame *); cleanup_record *next; };
static cleanup_record *cleanups;
static void (*take_cleanup(adamic_async_frame *frame))(adamic_async_frame *) {
    cleanup_record **at = &cleanups;
    while (*at != NULL) {
        cleanup_record *record = *at;
        if (record->frame == frame) {
            void (*cleanup)(adamic_async_frame *) = record->cleanup;
            *at = record->next; free(record); return cleanup;
        }
        at = &record->next;
    }
    return NULL;
}
void adamic_async_register_cleanup(adamic_async_frame *frame, void (*cleanup)(adamic_async_frame *)) {
    cleanup_record *record = malloc(sizeof *record);
    if (record == NULL) adamic_panic("out of memory", 13);
    *record = (cleanup_record){frame, cleanup, cleanups}; cleanups = record;
}
void adamic_async_forget_cleanup(adamic_async_frame *frame) { (void)take_cleanup(frame); }


adamic_async_promise *adamic_async_new(void) {
    adamic_async_promise *promise = adamic_allocate(sizeof *promise, adamic_kind_async_promise);
    promise->settled = promise->rejected = promise->references = false;
    promise->value = (adamic_value){.number = 0};
    promise->first = promise->last = NULL;
    return promise;
}
static void unregister(adamic_async_promise *promise) {
    pending **at = &subscriptions;
    while (*at != NULL) {
        pending *entry = *at;
        if (entry->promise == promise) {
            *at = entry->next;
            adamic_release(entry->promise);
            free(entry);
            return;
        }
        at = &entry->next;
    }
}
static void queue(adamic_async_reaction *reaction) {
    reaction->next = NULL;
    if (jobs_last != NULL) jobs_last->next = reaction; else jobs = reaction;
    jobs_last = reaction;
}
static void clear_wait(adamic_async_reaction *reaction) {
    adamic_async_promise *waiting = reaction->frame->waiting;
    reaction->frame->waiting = NULL;
    adamic_release(waiting);
}
void adamic_async_await(adamic_async_frame *frame, adamic_async_promise *promise) {
    assert(frame->waiting == NULL);
    adamic_async_reaction *reaction = adamic_allocate(sizeof *reaction, adamic_kind_async_reaction);
    reaction->frame = adamic_retain(frame);
    reaction->promise = NULL;
    reaction->next = NULL;
    frame->waiting = adamic_retain(promise);
    if (promise->settled) {
        reaction->promise = adamic_retain(promise);
        clear_wait(reaction);
#ifdef ADAMIC_ASYNC_INLINE_MUTANT
        frame->resume(frame, promise->value, promise->rejected);
        adamic_release(reaction);
#else
        queue(reaction);
#endif
        return;
    }
    if (promise->first == NULL) {
        pending *entry = malloc(sizeof *entry);
        if (entry == NULL) adamic_panic("out of memory", 13);
        *entry = (pending){adamic_retain(promise), subscriptions};
        subscriptions = entry;
    }
    if (promise->last != NULL) promise->last->next = reaction; else promise->first = reaction;
    promise->last = reaction;
}
void adamic_async_settle(adamic_async_promise *promise, adamic_value value, bool references, bool rejected) {
    if (promise->settled) return;
    adamic_retain(promise); /* Protect against clearing the last frame/registry edges. */
    promise->settled = true;
    promise->rejected = rejected;
    promise->references = references;
    promise->value = value;
    if (references) adamic_retain(value.reference);
    adamic_async_reaction *reaction = promise->first;
#ifndef ADAMIC_ASYNC_SETTLE_CYCLE_MUTANT
    promise->first = promise->last = NULL;
#endif
    while (reaction != NULL) {
        adamic_async_reaction *next = reaction->next;
#ifndef ADAMIC_ASYNC_SETTLE_CYCLE_MUTANT
        reaction->promise = adamic_retain(promise);
        clear_wait(reaction);
        queue(reaction); /* Transfer the list's owned count. */
#endif
        reaction = next;
    }
    unregister(promise);
    adamic_release(promise);
}
static void abandon(adamic_async_promise *promise, bool cancelling) {
    adamic_retain(promise);
    adamic_async_reaction *reaction = promise->first;
    bool break_cycle = true;
#ifdef ADAMIC_ASYNC_CANCEL_CYCLE_MUTANT
    if (cancelling) break_cycle = false;
#endif
#ifdef ADAMIC_ASYNC_EXIT_CYCLE_MUTANT
    if (!cancelling) break_cycle = false;
#endif
    (void)cancelling;
    for (adamic_async_reaction *entry = reaction; entry != NULL; entry = entry->next) {
        void (*cleanup)(adamic_async_frame *) = take_cleanup(entry->frame);
        if (break_cycle && cleanup != NULL) cleanup(entry->frame);
    }
    if (break_cycle) {
        promise->first = promise->last = NULL;
        while (reaction != NULL) {
            adamic_async_reaction *next = reaction->next;
#ifndef ADAMIC_ASYNC_NEXT_LINK_MUTANT
            reaction->next = NULL; /* Transfer next's owned count to the local before freeing this. */
#endif
            clear_wait(reaction);
            adamic_release(reaction);
            reaction = next;
        }
    }
    unregister(promise);
    adamic_release(promise);
}
/* Low-level subscription cancellation only. Source cancellation/finally is not implemented. */
void adamic_async_cancel(adamic_async_promise *promise) { abandon(promise, true); }
static void drain_jobs(void);
static int timer_wait_ms(void);
#include "async_host_impl.h"
#include "timers_impl.h"

void adamic_async_teardown(void) {
    timer_shutdown();
    host_shutdown();
    while (subscriptions != NULL) abandon(subscriptions->promise, false);
}
static void drain_jobs(void) {
    while (jobs != NULL) {
        adamic_async_reaction *reaction = jobs;
        jobs = reaction->next;
        if (jobs == NULL) jobs_last = NULL;
        reaction->next = NULL;
        reaction->frame->resume(reaction->frame, reaction->promise->value, reaction->promise->rejected);
        adamic_release(reaction);
    }
}
void adamic_async_run(void) {
    for (;;) {
        adamic_host_process_completions();
        if (!host_live() && !timer_live()) break;
        if (timers != NULL && timer_dispatch()) continue;
        host_wait_hook();
    }
    adamic_async_teardown();
}
void adamic_async_free_children(void *value, void (*drop)(void *)) {
    adamic_heap *heap = value;
    switch (heap->kind) {
        case adamic_kind_async_frame: {
            adamic_async_frame *frame = value;
            adamic_async_forget_cleanup(frame);
            drop(frame->waiting); drop(frame->output);
            frame->children(frame, drop);
            break;
        }
        case adamic_kind_async_promise: {
            adamic_async_promise *promise = value;
            if (promise->references) drop(promise->value.reference);
            drop(promise->first);
            break;
        }
        case adamic_kind_async_reaction: {
            adamic_async_reaction *reaction = value;
            drop(reaction->frame); drop(reaction->promise); drop(reaction->next);
            break;
        }
        default: abort();
    }
}
