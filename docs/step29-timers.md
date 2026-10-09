# Step 29: native CLI timer provider

Ruling, system_adamic, October 8, points 8 and 9:

> An escaping host callback is retained by its registration until cancelled or
> closed. Its captures follow the cycle rules (Program region or graph regions).
> Arity and optional arguments follow the signature, checked.
>
> One event loop, single-threaded delivery, Node's ordering for what tsc uses.
> Timers and watchers beyond the command line's needs are NotYet.

Base: origin/area/runtime b0facff530f6c8bde18c4ba9dd9222f47271d5d9.
This unit supplies the runtime C seam. Node declarations, signature checking,
handle wrappers and source lowering belong to library/compiler. The fixtures
call this production primitive through C; their independent `.a` host oracles
run on Node v24.19.0. This does not claim setTimeout now lowers from source.

## Scope and observed Node behavior

The tsc compiler fixture tree is cohere/TypeScript/tsc/testdata/fixtures/compiler.
The actual scheduling calls include sys.ts:377 (polling queue and label), :490
(poll queue and label), :770 (child-watch update), watchPublic.ts:868/:890
(resolution invalidation and program update), and tsbuildPublic.ts:2068
(rebuild with label, state and change flag). Other matches bind, declare or test
availability of setTimeout. There is no setInterval or unref call in this tree.
These watch services remain NotYet; this primitive does not implement filesystem
watchers, polling policy or language-service lifetime.

Node observations on the same Linux box: clearTimeout on an already cleared
Timeout, `{}`, a nonexistent numeric identity, null and undefined returns
undefined without an exception or output. clearTimeout/clearInterval are
interchangeable for timeout/interval handles. Consequently the checked C cancel
returns false for foreign/stale handles; library discards that result at the Node
surface. “Checked cancellation” does not introduce an exception Node never
throws. Handle identities are never reused; callers never dereference them.

Node normalizes 0 and 0.5 to 1 ms, and truncates positive fractional durations.
Negative, NaN and overflowing durations also become 1 but emit Node-specific
warnings. This primitive accepts finite [0, INT32_MAX] durations and names the
warning-requiring cases NotYet. Coercion from other source types is library's
responsibility; an omitted duration must become 1 ms in the binding. Invalid
callback types likewise need the binding's checked signature or Node error
path. None is implemented as an unchecked C conversion.

The referenced library host fixture 03 was not present in fetched
origin/area/library, origin/codex/host-surface-area or
origin/library/area-on-next: their host/03 is the
UTF-16BE readFile fixture. The unref oracle here directly establishes the stated
exit behavior against Node; binding compatibility with that unnamed timer
fixture still needs library's review.

## Ownership and loop contract

A registration copies all actual values and the reference mask, and owns one
ordinary retain on its closure and every reference argument. Those operations
dispatch through the heap header, so graph closures/captures keep their region
alive without becoming counted internal cycle edges. A timeout retires before
invocation but keeps these references until invocation returns. An interval
retains them across fires. Cancellation between fires drops them immediately;
cancellation during its invocation unlinks it immediately and defers dropping
until return. Pending unref timers are dropped at loop teardown, without firing.
Timer registration storage and argument arrays are malloc-owned by the registry,
not Adamic heap values or graph edges, and leak checking includes them.
Registration storage is one timer_entry plus argc values and reference flags.
Lookup and deadline selection scan the live list, O(n); this CLI primitive
does not claim a high-volume watcher scheduler.

All timer operations require the existing host loop thread. There is no timer
worker, shared region, tracing collector or delivery pause. Deadlines use
CLOCK_MONOTONIC milliseconds. Eligible callbacks are dispatched by deadline,
then registration identity for equal deadlines. Intervals rearm relative to the
start of their callback, not its completion. Promise jobs drain to exhaustion
before timers and between callbacks. A due timer never synchronously invokes its
callback from registration, and cancellation before dispatch cannot leave a
queued invocation behind.

The existing native poll wait receives the next deadline (including unref timers
when other work lives), and still wakes on the existing host completion pipe.
Only referenced timers or host requests keep the loop alive. A program with no
timers retains the previous host wait behavior. Custom host wait hooks cannot
accept deadlines: registering a timer with one is named NotYet instead of
silently blocking forever. WASI native timers are likewise named NotYet.

This is a CLI timer/promise ordering subset, not a complete libuv phase model.
nextTick, setImmediate, refresh, ref/hasRef, timer disposal, Timeout as callback
`this`, arbitrary I/O/timer races and watchdog/watch-service policy are outside this unit. C callers must
not reenter async_run or teardown from an invocation; generated source does not
expose those control entry points. A thunk leaving adamic_thrown set causes the
ordinary uncaught-error path after normal invocation bookkeeping; there is no
new timer exception-catching policy.

## C entry points for library

Declarations and one contract comment per entry point live in
internal/native/runtime/timers.h. Include that header, not timers_impl.h.

```c
// Borrow the one native provider. Its identity is immutable for the process.
const adamic_timer_provider *adamic_timers_provider(void);

// Copy argc actual arguments; retain callback and mask-marked reference values.
// repeat=false is setTimeout; repeat=true is setInterval. Newly created timers
// are referenced. All actual values must already match the checked signature.
adamic_timer_handle adamic_timer_start(const adamic_timer_provider *provider,
    adamic_closure *callback, adamic_timer_invoke invoke, double delay_ms,
    bool repeat, size_t argc, const adamic_value *arguments,
    const bool *references);

// Return true only when this provider still owns a live registration. Both
// clearTimeout and clearInterval use this. Source result is undefined either way.
bool adamic_timer_cancel(adamic_timer_handle handle);

// Return true for a live handle after marking it unreferenced. Source Timeout
// wrapper returns itself; its wrapper does not own the registration's lifetime.
bool adamic_timer_unref(adamic_timer_handle handle);
```

The handle contains the provider identity and monotonically increasing uintptr_t
registration identity. Zero is invalid. A handle is borrowed token storage;
copying or dropping it does not retain/release a closure, and a provider mismatch
cannot cancel a matching identity from this provider. Exhaustion aborts rather
than reusing an identity. Invalid C registration inputs abort, while supported
source inputs are established by the binding's checked signature.

`adamic_timer_invoke` is `void (*)(adamic_closure *, size_t argc,
const adamic_value *arguments)`. Library must supply a signature-checked thunk
which materializes missing optional/default parameters, preserves rest values
and ignores allowed excess actual arguments according to the source signature.
The thunk must dispose of an owned ignored return value. The runtime neither
casts a C function pointer nor blindly calls closure->code, whose ABI has no
argc. Numeric/boolean/reference representation masks remain required; a union
argument needs a binding-owned box and its reference mask. The fixture thunk
checks its exact two-argument signature before entering closure code.

The generated entry invokes adamic_async_run after module-body work and disposes
of ordinary globals afterward, as today. async_run handles timer dispatch and
tears down registrations before abandoning pending subscriptions; library must
not add a second event loop. Host-only embedding hooks remain unchanged when no
timers are registered.

## Evidence

TestTimersAgainstNode covers promise ordering, three interval fires, cancellation
before fire, cancellation inside a callback with subsequent capture/argument
reads, provider/identity mismatch and stale cancellation, unref exit, unref
delivery while another timer keeps the loop alive, deadline ordering and a
self-linked graph capture whose registration is the last outside owner,
optional/default/rest arguments and a timer waking a live host request.
Every positive control is counted, ASan/UBSan checked, compared to the -O2
ThinLTO release build, held byte-for-byte to Node, and run through internal/leakcheck. No language lowering is inferred from C.

TestTimerMutants proves the following checks fail on a compiling changed runtime:

| Mutant | Fixture | Required observer |
|---|---|---|
| Release interval callback after its first fire | interval | ASan heap-use-after-free at a later fire |
| Ignore handle provider | foreign | C cancellation contract assertion |
| Ignore unref | unref | Node stdout comparison |
| Unlink cancel without releasing registration | cancel | Shared leak check |
| Free a running registration on cancel | cancel-self | ASan heap-use-after-free |
| Deliver timers before the initial promise checkpoint | order | Node stdout comparison |

Validation on Linux, Intel Xeon Platinum 8573C, nproc 5 (CPU quota 4),
Go 1.27.1, clang 20.1.8, Node v24.19.0:

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`
  succeeded. Timing lines: Node 0.044s, Go 0.064s, markdown dependencies 0.149s,
  clang 0.404s, submodules 198.116s, build cache warm 573.808s, done 573.856s.
  Environment: `source /workspace/adamic-tools/env.sh`.
- `go build ./...` and `go vet ./...`: exit 0. Final vet after adding release
  comparisons also exited 0. gofmt and git diff --check are clean.
- `go test ./internal/native -run 'TestTimersAgainstNode|TestTimerMutants|TestHostPromises|TestGraph|TestRuntimeStaticsAreListed' -count=1 -timeout 20m -v`:
  passed, 67.396s. Eleven Node comparisons with balanced counts, shared leak
  checks and matching -O2 ThinLTO release output; all six timer mutants caught.
  Existing host-promise ASan/UBSan and TSan controls and mutants passed, as did
  graph controls/mutants, including the million-member test. Log:
  /tmp/step29-timers-green-native.log.
- `go test ./internal/lower ./internal/fresh ./internal/flow -count=1 -timeout 20m`:
  passed, 126.986s / 168.678s / 257.738s. Log: /tmp/step29-timers-lower.log.
- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(async|closures|graph_regions|cycles)' -count=1 -timeout 20m`:
  passed, 197.414s, including the oracle's sanitizer and leak runs.
  Log: /tmp/step29-timers-oracle.log.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 20m`:
  passed, 140.963s. Every existing counts row is unchanged; no regeneration or
  table edit was needed. Log: /tmp/step29-timers-counts.log.
- `go test ./internal/native -run '^TestTimerMutants/cancel-leaks-registration$' -count=1 -v`:
  passed, 0.400s, after strengthening the assertion to require a specific leak
  diagnostic. Final `go vet ./...` also passed after this test-only change.
  Log: /tmp/step29-timers-leak-mutant.log.
- `go test ./internal/native -run '^TestWASIHostPromises$' -count=1 -v`:
  skipped: ADAMIC_TEST_WASI not enabled and WASI_SYSROOT unset. No WASI execution
  or macOS execution is claimed. Log: /tmp/step29-timers-wasi.log.

The first release comparison correctly failed on a fixture-only duplicate
adamic_heap_end call: the runtime already registers that cleanup with atexit.
The call was removed, and the command above re-ran green. No runtime allocator
change or suppressed assertion was needed. No full-repository test gate ran;
the commands above are the exact covered scope.

The self-cycle graph callback control reports 5 allocations / 5 frees,
2 retains / 6 releases, peak 5, one graph region and two merges. Its three
members all free when the last interval invocation returns. Timer entries and
argument buffers are outside the heap count table but covered by leakcheck.

