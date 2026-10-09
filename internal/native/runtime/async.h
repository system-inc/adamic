#ifndef ADAMIC_ASYNC_H
#define ADAMIC_ASYNC_H
#include "adamic.h"

typedef struct adamic_async_promise adamic_async_promise;
typedef struct adamic_async_frame adamic_async_frame;
typedef struct adamic_async_reaction adamic_async_reaction;
struct adamic_async_frame {
    adamic_heap heap;
    unsigned state;
    adamic_async_promise *output;
    adamic_async_promise *waiting;
    void (*resume)(adamic_async_frame *, adamic_value, bool);
    void (*children)(adamic_async_frame *, void (*)(void *));
    // Canonical closures over this frame's own locals, as adamic_environment's functions (closure.c).
    struct adamic_closure *functions;
};
struct adamic_async_promise {
    adamic_heap heap;
    bool settled, rejected, references;
    adamic_value value;
    adamic_async_reaction *first, *last;
};
adamic_async_promise *adamic_async_new(void);
void adamic_async_settle(adamic_async_promise *, adamic_value, bool, bool);
void adamic_async_await(adamic_async_frame *, adamic_async_promise *);
void adamic_async_cancel(adamic_async_promise *);
void adamic_async_run(void);
void adamic_async_teardown(void);
/* Compiler-private environment cleanup. Promise producer/settlement ABI is unchanged. */
void adamic_async_register_cleanup(adamic_async_frame *, void (*)(adamic_async_frame *));
void adamic_async_forget_cleanup(adamic_async_frame *);
void adamic_async_free_children(void *, void (*)(void *));

/* Host bridge. This alias preserves unit 1's promise layout exactly. */
typedef adamic_async_promise adamic_promise;
typedef struct adamic_host_request adamic_host_request;

/* Loop thread only. *out receives an owned pending promise. The live request
 * owns another count and keeps adamic_async_run alive until completion/abandon.
 * Requests are opaque identities, never dereferenceable or reusable addresses.
 * Retired identities remain safe to pass: no caller-side release is needed. */
adamic_host_request *adamic_host_promise_new(adamic_promise **out);

/* Any thread. Copy bytes/message before returning; never access Adamic values.
 * Resolve yields a counted object with fields body (owned validated UTF-8 string)
 * and status (double), in that order. Invalid UTF-8 rejects with an Error.
 * Reject takes a NUL-terminated UTF-8 message and yields an Error.
 * true means accepted; a second completion returns false with a stderr diagnostic.
 * Completion after abandonment returns false safely. NULL bytes is valid only
 * for length zero. The host retains ownership of all supplied input storage. */
bool adamic_host_resolve(adamic_host_request *request, const void *bytes,
                         size_t length, double status);
bool adamic_host_reject(adamic_host_request *request, const char *message);
void adamic_host_abandon(adamic_host_request *request);

/* Install once on the loop thread at startup, before creating requests or
 * publishing work. NULL selects the default for that hook. wake runs on the
 * publishing thread and must schedule adamic_host_process_completions on the
 * loop thread, never call it there inline. Default: signal a nonblocking pipe.
 * wait runs on the loop thread only when adamic_async_run has exhausted jobs
 * with live host handles. Default: wait for the pipe. Apple supplies a wake
 * using dispatch_async_f onto the main queue and a wait driving CFRunLoop.
 * AppKit owns the main loop: drain via wake, without calling the blocking run
 * entry while an app is running. Hook implementations belong to the host. */
typedef void (*adamic_host_loop_hook)(void);
bool adamic_host_set_loop_hooks(adamic_host_loop_hook wake,
                                adamic_host_loop_hook wait);

/* Loop thread only. Settle queued completions and drain their microtasks.
 * Safe when the queue is empty. Does not wait or tear down pending work.
 * Apple can call this through a dispatch_async_f adapter with ignored context. */
void adamic_host_process_completions(void);
#endif
