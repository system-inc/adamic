# Host-resolved promises

The public C interface is `../../runtime/async.h`. `adamic_promise` is an alias
for unit 1's `adamic_async_promise`; its fields, layout, and heap kind are
unchanged. The implementation is private to `async.c` through
`async_host_impl.h`. No compiler or Apple implementation is added here.

Create requests on the loop thread with `adamic_host_promise_new(&promise)`.
The caller owns the returned promise count. Independently, the host registry
owns a producer count and keeps `adamic_async_run()` alive. Releasing the caller's
promise does not cancel the request. Low-level `adamic_async_cancel()` unlinks
subscriptions on the loop; the producer remains live until completion or host
abandonment. Source cancellation and OS cancellation are still NotYet.

`adamic_host_resolve(request, bytes, length, status)` copies the input before
returning. A successful result is an ordinary counted object, with `body` as an
owned UTF-8 string in slot 0 and `status` as a double in slot 1. Embedded NULs
are preserved by length. Strict UTF-8 validation rejects malformed sequences,
overlong encodings, surrogate code points, and code points above U+10FFFF; invalid
input rejects with `Error("host payload is not valid UTF-8")`. A valid
`adamic_host_reject(request, message)` creates an Error from the copied message.
Nothing parsed and no Adamic value crosses the worker boundary.

Resolve, reject, and abandon compete under one mutex; the first terminal
publication wins. Accepted completion returns true. Repeated or retired
completion returns false and prints `adamic: host completion refused: request
retired or already completed` to stderr. Abandonment is idempotent; it queues a
loop-owned retirement, without running or cancelling source continuations. Once
all handles are retired, normal exit tears down any remaining subscriptions.
Calling completion after abandonment or teardown is safe, including while a new
request is live. There is no caller-side request free.

Handles are nonzero, monotonically allocated identity tokens represented through
`uintptr_t` and the opaque pointer type. They are never dereferenced, never reused,
and never passed to a counted allocator. Exhausting the identity space aborts
rather than reusing an identity. This uses the integer/pointer round-trip of the
supported clang native targets; Linux is tested here, Darwin is not. This choice
avoids either permanent allocated tombstones or an unsafe callback into freed
request storage. Actual requests and copied buffers are freed on the loop.

# Loop integration

Install `adamic_host_set_loop_hooks(wake, wait)` once on the loop thread, before
creating any requests. NULL selects a default for that hook. A second installation
or installation after requests have started returns false. The wake callback
runs on the publishing thread, under the bridge mutex. It must schedule work and
return; it must not call bridge functions inline. Hooks have process lifetime.

The default wake writes one byte to a nonblocking, close-on-exec pipe. The default
wait checks the queue predicate and polls that pipe. A publication before polling
remains readable; a full pipe already carries the notification. Teardown closes
the descriptors under the publication mutex, so late callbacks cannot signal a
closed or reused descriptor. A live host request with no completion deliberately
keeps the standalone loop waiting, like a referenced timer in Node. The harness
puts deadlines around every subprocess, including the missing-wake mutant.

`adamic_host_process_completions()` is a plain no-argument loop-thread function.
It is safe on an empty queue, never waits, and never tears down pending
subscriptions. It removes each request under the mutex, then converts its copied
payload, settles its promise, releases its producer count and buffer, and drains
microtasks before processing the next host completion. Duplicate callers only
look up an identity; they never read a promise or frame.

The Apple worker can supply a wake that uses `dispatch_async_f` onto the main
queue, with an adapter that ignores its context and calls
`adamic_host_process_completions()`. AppKit owns the loop thread, so while the app
runs it should use those scheduled drain callbacks rather than enter the
standalone blocking `adamic_async_run()`. For standalone top-level completion,
Apple's wait can drive CFRunLoop until scheduled drain callbacks settle the
requests. That wait is entered only after runnable work is drained and live
handles remain. No Apple implementation is provided or tested here.

`adamic_async_run()` waits until every host handle completes or is abandoned,
then performs subscription teardown. Explicit `adamic_async_teardown()` is the
shutdown primitive: it disarms outstanding identities, discards copied queued
completions, releases the host registry, and breaks pending subscription cycles.
It does not wait for native host work; host-owned external buffers and threads
remain the host's responsibility. Late callbacks are still safe identity lookups.

# Evidence

After sourcing the path printed by `bash cloud/setup.sh`, run:

```sh
python3 internal/native/testdata/async_host/check.py > /tmp/host-promises-check.log 2>&1
go test -count=1 -timeout 30m ./internal/native -run TestHostPromises > /tmp/host-promises-native.log 2>&1
```

Every child build and execution writes separate `.out` and `.err` files in the
printed artifact directory, with a 30-second deadline. The missing-wake run has
a three-second deadline. The single loop runs on a joined pthread; its completed
stack cannot conservatively hide leaks from LSan. Each observer holds a dynamic
string, an output promise and the awaited host promise. Two observers exercise
the complete reaction list and both frame-owned promise edges.

Each scenario runs under TSan and separately ASan/UBSan with Linux LSan enabled:
delayed cross-thread resolve, reject, pre-wait resolve, two completions (before
and after retirement), delayed abandon and late resolve, cancellation followed
by completion, cancellation followed by abandonment, explicit teardown with a
host promise pending, invalid UTF-8, custom wake/wait, simulated AppKit draining
without waiting, concurrent producer stress, per-completion microtask checkpoints,
and stale-token isolation while a new request is live. Callback inputs are stack
buffers overwritten immediately after publication. Resolve, reject and early
resolution traces also match the independent Node program `oracle.a`.

Mutants are compiled with the same warning gate before execution:

| Mutant | Intended detector |
| --- | --- |
| Settle directly on the native worker | TSan data race against real loop promise-state reads |
| Remove the completion/registry mutex | TSan data race with concurrent producers and loop draining |
| Skip wake | Three-second subprocess timeout |
| Discard the host registry without releasing entries | LSan leak on pending-host teardown |
| Skip settlement detach/clear | LSan retained frame/promise/reaction cycle |
| Skip cancellation detach/clear | LSan cycle after cancel and host abandonment |
| Skip exit detach/clear | LSan cycle on pending-host teardown |
| Drain microtasks after all completions | Assertion: the next completion settled before the current reaction |
| Replace the buffer copy with unrelated bytes | Payload-byte assertion |
| Alter the HTTP status | Number assertion |
| Skip UTF-8 validation | Rejection assertion |
| Reuse a retired token | Stale-token assertion |
| Permit a second hook installation | Startup-hook assertion |

The cancellation leak probe abandons rather than resolves its host request:
later settlement must not repair and mask the deliberately broken cancellation.
The settlement leak mutant disables only the missing-resume assertion, allowing
normal process exit so LSan is the detector.

Unit 1's `internal/lower/async.go:asyncWait` only lowers primitive Promise.resolve
and named async calls. Host/network calls explicitly produce NotYet; its payload
subset also excludes this object result. There is consequently no compiled Adamic
host-await fixture in this unit. `oracle.a` is an external Node witness, not a
claim of new compiler support. No HTTP implementation, Apple build, OS cancellation,
Windows port, general Node phase adapter, or performance claim is included.
