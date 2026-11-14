/* Native readiness sources, delivered only by async.c's single event loop. */
#ifndef ADAMIC_SOURCES_H
#define ADAMIC_SOURCES_H
#include "async.h"

/* WASI preview 1 has no poll wait, so nothing there can register a source. */
#ifndef ADAMIC_TARGET_WASI
/* Ready bits a dispatch receives. error stands for POLLERR, POLLNVAL or POLLHUP
 * on the descriptor; it comes back every turn until the owner closes the source. */
enum { adamic_source_readable = 1, adamic_source_error = 2 };
/* Opaque identity: callers compare it only, and a closed one is never reused. */
typedef struct adamic_source_handle { uintptr_t identity; } adamic_source_handle;
/* Runs as a macrotask with the ready bits the last wait recorded. data is the
 * registration's, borrowed. An adamic_thrown left by it is uncaught. */
typedef void (*adamic_source_dispatch)(void *data, unsigned ready);

/* Register a descriptor polled for POLLIN. It must already be non-blocking. The
 * descriptor and data stay the caller's: neither closing a source nor teardown
 * closes or frees them. A referenced source keeps the loop alive; an unreferenced
 * one is still polled while other work lives. Refused under custom host loop
 * hooks. All operations require the loop thread. */
adamic_source_handle adamic_source_register(int descriptor,
    adamic_source_dispatch dispatch, void *data, bool referenced);
/* Mark a live source unreferenced. False means stale. */
bool adamic_source_unref(adamic_source_handle handle);
/* Mark a live source referenced again, unref's mirror, as uv_ref is uv_unref's.
 * A flip either way is no close: it leaves no pending close and costs no turn.
 * False means stale. */
bool adamic_source_ref(adamic_source_handle handle);
/* Retire a live source, leaving its descriptor open. It is never dispatched
 * again, not even later in the turn that closes it, and may close itself from
 * its own dispatch. False means stale. As a handle closed through uv_close does,
 * it leaves a pending close: the loop stays alive one more turn whose wait never
 * blocks, so an unreferenced source's queued readiness is still dispatched. */
bool adamic_source_close(adamic_source_handle handle);
/* The same pending close for a handle Node closes through uv_close that isn't a
 * source here: fs.watchFile's poller, a uv_fs_poll handle in Node, lives on the
 * timer provider, and its retirement calls this (node_watch_file_impl.h). */
void adamic_source_pending_close(void);
/* Run the microtasks queued so far: the checkpoint Node takes after each callback
 * a dispatch makes, where one dispatch makes several (one per inotify event). Only
 * from inside a source's dispatch, on the loop thread; anywhere else it aborts. */
void adamic_source_drain_jobs(void);
#endif
#endif
