/* Native CLI timer provider, delivered only by async.c's single event loop. */
#ifndef ADAMIC_TIMERS_H
#define ADAMIC_TIMERS_H
#include "async.h"

typedef struct adamic_timer_provider adamic_timer_provider;
/* Opaque identity plus provider brand: neither pointer is dereferenced by callers. */
typedef struct adamic_timer_handle {
    const adamic_timer_provider *provider;
    uintptr_t identity;
} adamic_timer_handle;
/* A signature-checked library thunk supplies optional/rest arguments and discards
 * any owned return value. The runtime never calls closure->code without it. */
typedef void (*adamic_timer_invoke)(adamic_closure *, size_t, const adamic_value *);

/* Return the process's native CLI provider; it owns no caller reference. */
const adamic_timer_provider *adamic_timers_provider(void);
/* Register a timeout/interval, copying arguments and retaining each reference
 * and the callback. Fractions of delay_ms truncate, and one outside [1, INT32_MAX]
 * becomes 1 ms, NaN, negative or past INT32_MAX with Node's warning, written at the
 * loop's next tick. repeat selects interval. Library checks the signature and
 * supplies its invocation thunk. All operations require the loop thread. */
adamic_timer_handle adamic_timer_start(const adamic_timer_provider *provider,
    adamic_closure *callback, adamic_timer_invoke invoke, double delay_ms,
    bool repeat, size_t argc, const adamic_value *arguments, const bool *references);
/* Retire a live registration. False means foreign/stale; Node clearTimeout and
 * clearInterval adapters ignore that result and return undefined. Inside a
 * callback, retain its callback/arguments until invocation returns. */
bool adamic_timer_cancel(adamic_timer_handle handle);
/* Mark a live timer unreferenced: it may fire while other work keeps the loop
 * alive, but cannot keep the loop alive itself. False means foreign/stale. */
bool adamic_timer_unref(adamic_timer_handle handle);
/* setTimeout called from source (#995spev): a timeout on this provider, and its Timeout, an owned
 * object of a host shape (adamic_host_methods), so a run-time key read of it stops. The shape's
 * address is the provider's brand, its one number slot the registration's identity, and the object
 * owns nothing of the registration, which lives until it fires or is cancelled. The emitter checks
 * the signature and supplies the thunk; delay_ms is adamic_timer_start's. */
adamic_object *adamic_node_set_timeout(adamic_closure *callback, adamic_timer_invoke invoke,
    double delay_ms, size_t argc, const adamic_value *arguments, const bool *references);
/* clearTimeout called from source takes any value, as Node's does (Adamic's ruling on #b5w3ycg): a
 * Timeout, known by its shape's address and never by a field it holds, is cancelled, and anything
 * else, a primitive as NULL among them, is a no-op that doesn't throw, but an object holding a truthy
 * _onTimeout, which Node would unenroll, stops (#995spev). 0 is the discarded result. */
double adamic_node_clear_timeout(const adamic_heap *value);
#endif
