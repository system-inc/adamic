# Concurrency, part 3: async and await

Status: **core design approved by @system_adamic with Kirk briefed, October 6, 2026; cycle protocol and compiler landing follow-up below**. Task p286ycm. Branch `codex/concurrency-async`, based on `origin/codex/concurrency` at `ced5530e68357ca9880f4366870d1f596f453d1a`. Nothing in this document changes the approved language or part 1 contract. The first compiler landing is the closed primitive subset described below; broader runtime and library contracts remain staged.

## Recommendation for review

Generate stackless C state machines in the compiler. Run ordinary async bodies and all Promise reactions on one event loop thread. Use a separate bounded I/O service for blocking files, and readiness notifications for sockets. Keep part 1's work-stealing pool for explicitly requested, proven pure Shareable work. Do not treat `async` as permission to move source code between threads.

Represent a Promise as an ordinary counted heap object with a settlement tag, one owned result or error, and ordered reaction records. Give each suspended frame an owned slot for every reference that remains live across suspension, plus a compiler-generated cleanup path. Use explicit structured scopes for task lifetime and opt-in cooperative cancellation, without silently changing JavaScript Promise semantics. Initially accept only the subset whose lifecycle and ordering the compiler can prove, with specific NotYet diagnostics for missing implementation and Refused diagnostics for forbidden detachment or cycle-capable user edges.

This chooses portable generated C, visible ownership, and Node-compatible continuations over small compiler changes. It requires a suspension-aware control-flow pass and an actual Node-compatible loop adapter. The prototype proves a narrow mechanism; it does not prove that all of Node can be reproduced by a two-queue scheduler.

## Primary sources and what they establish

Sources were read October 6 and 7, 2026. Versioned sources below are preferred to rolling documentation. Rolling links and proposals describe their observed contents, not guarantees about future releases. The implications for Adamic in the last column are recommendations, not observations.

| Prior art, primary source | Observed mechanism | Implication for Adamic |
| --- | --- | --- |
| C#: Roslyn [AsyncRewriter](https://github.com/dotnet/roslyn/blob/main/src/Compilers/CSharp/Portable/Lowering/AsyncRewriter/AsyncRewriter.cs) and [AsyncMethodToStateMachineRewriter](https://github.com/dotnet/roslyn/blob/main/src/Compilers/CSharp/Portable/Lowering/AsyncRewriter/AsyncMethodToStateMachineRewriter.cs) | Generates a state-machine type, state and builder fields, parameter storage, and `MoveNext`. Hoisted locals and awaiters survive suspension. Builder completes with result or exception; exception exit clears hoisted fields. A completed awaiter can continue without scheduling. Release normally emits a struct; some modes emit a class. | Copy the state-machine shape, not the completed-await fast path. JavaScript `await` still yields for an already fulfilled native Promise. C#'s collector does not supply Adamic's releases. |
| Rust 1.90: [coroutine lowering](https://github.com/rust-lang/rust/blob/1.90.0/compiler/rustc_mir_transform/src/coroutine.rs), [`Future`](https://github.com/rust-lang/rust/blob/1.90.0/library/core/src/future/future.rs), [`Pin`](https://github.com/rust-lang/rust/blob/1.90.0/library/core/src/pin.rs) | Compiler computes a frame with captures, state, and locals live across suspension; generates resume/poll and state-specific drop glue. `poll(Pin<&mut Self>, Context)` returns Pending or Ready. The executor uses a Waker rather than repeatedly busy-polling. Async bodies begin when polled, unlike JavaScript's eager prefix. Dropping a suspended future drops its initialized contents; this is distinct from running an async cancellation handler. Pin protects address-sensitive/self-referential frames from moving until destruction, not from destruction itself. | Use state-specific initialization/ownership masks and drop glue. Adamic can allocate a stable-address frame before exposing it; no source-level Pin is needed if it never borrows C stack storage or moves the frame. Do not copy Rust's lazy start or its drop-as-cancellation semantics. |
| Kotlin: kotlinx.coroutines [structured lifetime](https://github.com/Kotlin/kotlinx.coroutines/blob/master/docs/topics/coroutines-basics.md), [cancellation](https://github.com/Kotlin/kotlinx.coroutines/blob/master/docs/topics/coroutines-cancellation.md), [composition](https://github.com/Kotlin/kotlinx.coroutines/blob/master/docs/topics/composing-suspending-functions.md) | `coroutineScope` waits for its children. `async` returns Deferred; parent cancellation propagates. Cancellation is cooperative, throws CancellationException at cooperating operations, and runs finally cleanup. GlobalScope leaves the hierarchy. Dispatchers select execution context separately from suspension. | Adopt explicit parent/child ownership and cancellation tokens for an Adamic scope API. Do not infer a worker dispatcher from `async`, and do not add GlobalScope. |
| Swift: [SE-0304 structured concurrency](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0304-structured-concurrency.md), [SE-0317 async let](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0317-async-let.md) | Task groups bound child lifetime; leaving a group waits for children, even after cancellation. Cancellation is cooperative. `async let` creates scoped children; unconsumed children are cancelled and awaited at scope exit. Group iteration can yield completion order. `Task` and `Task.detached` are separate unstructured APIs. | A useful ownership discipline, but implicit cancel-and-await at every scope exit would change ordinary TypeScript timing. Offer explicit scopes; keep Promise.all input ordering and fail-fast rejection rather than Swift group completion ordering. |
| Zig: [0.10.0 release notes](https://ziglang.org/download/0.10.0/release-notes.html), [0.15.1 release notes](https://ziglang.org/download/0.15.1/release-notes.html), [proposal #23446](https://github.com/ziglang/zig/issues/23446) | Self-hosted 0.10 lacked the old async implementation. 0.15.1 removed async/await keywords and frameSize, directing these operations toward the library Io interface. The open low-level stackless proposal discusses userland blocking, pools and fibers, and target limits for stackful coroutines. The proposed primitives are not a landed contract. | A warning about maintaining implicit suspension and frame ABIs across backends. Keep lowering and scheduling separate. Zig's library syntax redesign does not meet Adamic's requirement to accept TypeScript syntax. |
| Go 1.25.1: [runtime HACKING](https://github.com/golang/go/blob/go1.25.1/src/runtime/HACKING.md), [language go statements](https://go.dev/ref/spec#Go_statements) | G/M/P scheduler multiplexes goroutines; each live goroutine has a small growable stack. A go statement starts execution independently of the caller and does not join on lexical exit. Runtime stack growth and GC support are part of this machinery. | The counterexample: arbitrary thread migration and independently surviving goroutines would lose Node's observable order and part 1's structured ownership. Growable C stacks are also a major runtime project. |
| Node 24.19.0: [process microtask documentation](https://github.com/nodejs/node/blob/v24.19.0/doc/api/process.md), [task queue implementation](https://github.com/nodejs/node/blob/v24.19.0/lib/internal/process/task_queues.js), [event loop guide](https://nodejs.org/en/learn/asynchronous-work/event-loop-timers-and-nexttick) | Promise reactions and queueMicrotask use the same microtask queue. nextTick is separate; module context affects its relative order. Timer, poll, check, close and pending callback phases are distinct. Since libuv 1.45 timers run after poll, with an initial compatibility timer pass. | Node is the behavioral oracle. A generic FIFO of macrotasks is insufficient for all APIs. Pin Node/libuv version and module context for each supported API. |
| ECMAScript: [Await](https://tc39.es/ecma262/#await), [PerformPromiseThen](https://tc39.es/ecma262/#sec-performpromisethen), [Promise.all](https://tc39.es/ecma262/#sec-promise.all) | Await normalizes the value then installs reactions. Reaction jobs are enqueued; Promise.all preserves input positions and rejects on an observed rejection, without cancelling siblings. Resolving thenables involves further jobs and once-only resolution. | These algorithms define the Promise layer; Node defines the host scheduling around it. The sources above supply compiler and lifetime techniques, not a replacement oracle. |

## State machines versus stackful coroutines

A compiler frame contains a state number, owned reference slots, value slots, initialized-slot bits, an owned output Promise, handler/finally state, and scope/cancellation metadata. Entry executes synchronously through the first await and returns an owned Promise. Await splits control flow: snapshot operands in source order, spill live values, install a reaction, and return to the loop. Resumption is always a microtask, including await of a fulfilled Promise or a plain value. Completing the body settles the output; a throw rejects it rather than throwing synchronously to its caller.

Stackful fibers could preserve ordinary C call stacks and reduce explicit lowering, but every suspended stack still owns references. They need stack allocation/guards, sanitizer fiber-switch support, platform-specific context switching, an unwind/drop protocol for cancellation, and a safe boundary for synchronous calls. Arbitrary stack relocation is incompatible with C pointers. A fixed stack wastes memory per parked I/O task; a growing stack adds compiler/runtime machinery. Fiber scheduling would still need Node's microtasks, so it does not remove the central compatibility work. Stackless frames expose exactly what must be freed and work wherever clang can compile C. The cost is generated resume code, frame allocations and more compiler analysis.

No address into a caller's C stack may survive await. Parameters borrowed in synchronous functions must be retained or moved into the async frame before the eager entry can return. A borrowed array element or lent global read cannot cross a suspension on the current proof: another continuation may mutate its holder even on one thread. Retain its snapshot, or later prove an owner immutable and held for the entire suspension. Captured mutable cells remain loop-local and counted. A statement region may span await only if the region itself has a frame-owned lifetime; first implementation should disable that optimization for suspending statements and use heap storage. Suspend is a full effect boundary for fresh/escape, borrow, moves and reuse; source values visible from other continuations are not uniquely reusable merely because a local last use was found.

## Scheduling contract

One loop thread executes the synchronous prefix and every continuation, `.then`, `.catch`, `.finally` and queueMicrotask callback. No Promise/frame counting or settlement happens on I/O workers. Workers own request buffers and publish completions through a synchronized queue; the loop converts them into values and settles Promises. File service must be bounded separately from the CPU pool: blocking file calls cannot occupy all pure-work workers, and async joins cannot invoke a synchronous pool join that blocks the loop. Use a completion-producing pool submission adapter if CPU work is exposed asynchronously.

For the initial supported subset, drain microtasks after the module body and after every host callback before the next callback. Drain until empty, including newly enqueued reactions. Do not add fairness yields inside this drain: an endless microtask chain starves timers on Node too. Timer callbacks resolve their Promises synchronously; installed await reactions enter the queue before a subsequent queueMicrotask in that callback. Timer deadlines are monotonic; durations follow Node's normalization once that API is implemented.

For broader support, use libuv (prefer the Node version) for timer/file/socket phases, plus the Promise microtask layer. Libuv alone does not implement V8 Promise jobs, nextTick, unhandled rejection reporting or module checkpoints. Supporting setImmediate, nextTick and ESM dependency evaluation with top-level await each requires dedicated Node fixtures. Until then, those APIs are NotYet. Even libuv does not make races between independent external I/O completions reproducible across two separately executed programs. Differential fixtures must establish causal order, or compare a controlled completion trace replayed in both runtimes. Arbitrary timer-versus-network race traces cannot honestly be promised byte for byte across independent runs. Never use this limitation to reorder causally ordered callbacks.

The prototype starts the file read only after the timer has resolved, so its trace has no I/O/timer race. It uses a real monotonic timer, a blocking file reader on one helper pthread, a pipe completion and a join. Its scheduler handles one timer and one file request, not Node's complete phase model. Production libuv integration remains proposed and unmeasured.

## Promise representation, counts and cycles

Proposed ABI: monomorphized `Promise<T>` holds the usual heap header, pending/fulfilled/rejected tag, a typed owned payload (`T` or Error), and an ordered list of reaction tickets. Each ticket owns its callback/captures or frame, destination Promise where chaining needs one, and the payload snapshot after settlement. A Promise may be observed repeatedly; it is not a single-consumer future. Settling is once-only. It retains or takes the payload, detaches the reaction list, and queues jobs in registration order. Later observers enqueue new jobs; none run inline. Destruction releases the payload and any inert metadata. Scalar payloads need no retain; reference payloads do. Initial Promise state and reaction identity stay loop-local, so counts are plain; immutable results originating in part 1 already carry its shared tag and retain/release dispatch.

A producer/request owns the pending Promise until completion. The event loop registry additionally owns every pending subscription, including subscriptions with no live host handle. A frame owns its output Promise, its live locals, and the awaited Promise while suspended. The awaited Promise owns its reaction, and that reaction owns the frame. This can form a strong cycle; reference counts alone do not free it. The following explicit runtime protocol breaks it, with no cycle collector.

1. **Settlement, fulfilled or rejected:** the settling callback first holds a temporary Promise count. It marks settlement once and owns the result/error. It detaches the complete reaction list from the Promise before running any user code. For each await reaction, it clears the frame's matching awaited-Promise slot before releasing the count formerly in that slot. It transfers the detached reaction's frame count to a queued microtask, unregisters the pending subscription, and releases the registry's Promise count. Only then may it release the temporary settling count. The queued microtask owns the frame and payload through resume; its completion releases those counts. Finished-frame cleanup clears/releases output, locals and any initialized slots. The critical broken edge is frame -> awaited Promise, while Promise -> reaction is detached before queueing.
2. **Cancellation:** on the loop, hold a temporary Promise count and detach the Promise's reaction list, making those subscriptions unavailable to later settlement. Each detached reaction still owns its frame: clear/release frame -> awaited Promise, then release the reaction and its frame count. Frame destruction releases locals and output. Remove/release the registry entry and release the temporary Promise count last. This is the implemented low-level cancellation of pending subscriptions, before they become queued jobs. Source cancellation must instead transfer a reaction's frame count to a cancellation continuation that executes source cleanup/finally; it remains NotYet. Cancellation of already queued jobs also remains NotYet. Pending OS requests will retain only their request/Promise buffers until acknowledgement, never an obsolete reaction pointer. Cancelling pending I/O remains NotYet in the first compiler landing.
3. **Never settled at normal program exit:** Promises alone do not keep Node alive. After no host handles remain, stop accepting new work and drain already runnable microtasks. Walk the event loop's pending-subscription registry while holding each Promise temporarily; its detached reactions own their frames until released. Unlink each reaction, clear/release the matching frame await slot, release the reaction's frame count, then remove/release the registry's Promise count and temporary counts. Destruction releases initialized frame locals and output Promises. This exit-only abandonment does not resume the body or run source finally: Node does not execute finally in a never-resumed await at process exit. Clear all registry roots before the leak check. If outstanding I/O exists, join/acknowledge it before freeing request buffers and repeat until shutdown owns no tickets. Fatal panics retain the existing immediate-exit policy; this protocol is for normal exit.

A subscription has exactly one terminal owner: settle, cancellation, or exit teardown. All run on the event-loop thread. Removing the registry entry and clearing slots are idempotent; a consumed ticket cannot be resumed twice. These are runtime ownership transitions, not weak references supplied by user code. Each of the three terminal paths must have a LeakSanitizer fixture and a mutant omitting its detach/clear transitions while still discarding registry roots, exposing the retained cycle as a leak.

**Cycle finder boundary:** compiler-generated frame/Promise/reaction types have a compiler-only provenance identity, associated with the audited settle/cancel/teardown protocol. The check at `internal/lower/cycles.go` must exempt only that identity's internal protocol edges, never a name, structural shape, source annotation or a user-declared brand. Their user payloads/captures still go through ordinary cycle and fresh-write analysis. `internal/fresh` must treat suspension and user payload writes conservatively. Generated frame layouts are compiler IR/runtime metadata, not source declarations that can request an exemption. A test must accept the generated graph and refuse a user-created graph spelling the same names/fields; a mutant granting exemption by spelling must fail it. No public Weak requirement is added to async functions.

Breaking these runtime edges is necessary, not sufficient: user locals can hold Promises, a Promise can fulfill with an object pointing back to it, and resolver closures can capture the frame whose output they own. Extend memory.md's cycle-capable-slot finder and fresh-write proofs to Promise payloads, reaction captures and async frame edges. A mutable cell capturing its own task still needs Weak or refusal. Do not introduce a cycle collector. Internal ticket lists must support unlinking pending reactions on termination; draining reachable jobs and deleting scope registrations cannot depend on the Promise destructor breaking a strong cycle.

A dropped Promise handle does not cancel a producer. Likewise a Promise with no remaining producer is not a root that keeps a Node process alive by itself. Scope bookkeeping must recognize a permanently pending wait with no live host handles: report/diagnose the unsupported lifecycle, or explicitly drain/drop registrations under an approved exit rule. First implementation should expose compiler-generated and trusted runtime producers only; arbitrary `new Promise` executors, thenables, subclass/species, resolver escape and never-settling external APIs remain NotYet until their lifecycle can be represented and checked. This is a real compatibility cost for promise-shaped libraries, not a solved problem.

## Throws, cleanup and cancellation

Use memory.md's explicit cleanup paths, not longjmp over a suspended frame. Each state records which slots own a count. Slots moved out are cleared immediately; every exceptional or normal edge releases only still-initialized owned slots. Include partially evaluated expression temporaries, loop iterators, pending return values, caught errors and resources held by finally. Finally is control flow, not a destructor substitute: it may await, throw or override a return, and then the saved completion is itself frame-owned until superseded or delivered.

The pending-exception word used by synchronous callees cannot remain set on the loop while a task suspends. Resume must move an error into that frame's pending completion, clear the thread-local word, and select its catch/finally state before returning to the scheduler. On entry to a synchronous call, use the existing exception protocol; on exit, snapshot and clear it before another task runs. Root unhandled rejection needs an approved timing/message policy and a Node wrapper consistent with existing owned panic text; the prototype observes every rejection and does not implement this policy.

Cancellation is a request to an explicit task scope/token, not a Promise state or automatic effect of rejection. Check it at approved I/O cancellation points and explicit checks, then raise an Error through generated cleanup, including finally. Plain await of a fulfilled Promise must not acquire a new cancellation throw unless the source API specifies it. The JavaScript backend's scope/token library must implement the same checks. Suppressing cancellation is possible; a scope still waits for its children, so cancellation has no bounded-time guarantee for uncooperative computation.

To cancel a pending operation, remove the resume ticket or mark its generation cancelled so an already queued completion cannot resume a finished frame. If OS cancellation succeeds, still wait for its completion/acknowledgement before freeing buffers. If it cannot stop a file read, retain the request until completion, discard its result and release it on the loop. A scope cannot exit until all child cleanup and request acknowledgements finish. Test cancellation before registration, while pending, after settlement but before resume, and during an awaiting finally. The prototype only tests cooperative cancellation when the timer callback sets a flag and the queued continuation raises an error; it does not implement pending-request cancellation.

## Structured tasks, refusals, and Promise.all

Proposal: require every task-producing call to have a statically attributable owner: an await, a returned Promise transferring the obligation, a Promise.all member, or an explicitly registered child of an async scope. A local can temporarily hold a Promise before consuming it; aliases must not duplicate the obligation. `.then` attaches a reaction but returns another Promise whose obligation remains. Refuse a discarded Promise expression and `void task()`; the fix names an await, return, or explicit scope registration. Refuse detached task APIs and raw thread access because nobody then owns completion, errors and cleanup. The proof must be about task-producing effects, not just syntax: scalar Promise.resolve values and stored reusable Promise handles do not necessarily create new tasks. Conservative initial diagnostics may refuse more library patterns than necessary.

Do not insert hidden awaits at ordinary lexical block exit: that would change output order, and unlike Swift async let it is absent from TypeScript source. Proposed explicit scope API can wait for registered children at its own exit; its Node implementation is the sequential/async witness. Its exact spelling and treatment of escaped/never-settling children need Kirk and Ahra's approval. No such API is implemented here.

Promise.all preserves input result order, resolves an empty input as specified (its observers still run as jobs), and rejects on the first observed rejection. It does not cancel siblings. Root/scope ownership keeps siblings alive through cleanup even when the aggregate handle rejects early. Do not use part 1's lowest-input-index exception rule: it differs from first-observed Promise rejection. Do not block the loop while awaiting a group.

`Promise.all(items.map(async work))` cannot automatically run its bodies on part 1's pool. JavaScript evaluates each call's eager prefix in input order; it may log, mutate, throw, or start I/O there. Even pure async bodies create observable microtask turns. Prefer `parallelMap` for pure synchronous CPU work; inside an async function it is still a synchronous call. A later explicit async adapter may submit proven pure Shareable jobs to the pool and settle its output on the loop. Its Node witness must define completion publication order and error selection, independently of the worker schedule. Automatic Promise.all fusion would require a proof covering all these observations, not merely readonly captures; defer it.

## Feasibility artifact and observations

Files: [prototype.c](../prototypes/async/prototype.c), [oracle.a](../prototypes/async/oracle.a), [check.py](../prototypes/async/check.py), and [README.md](../prototypes/async/README.md). Handwritten C, not generated lowering. The oracle is a Node-only `.a` program copied verbatim into a scratch `.cjs` file; its timer/file executors and callbacks remain outside the compiler subset. Existing tsconfig does not include prototypes. No new `.ts` files. The original artifact remains standalone; the follow-up compiler and runtime evidence is recorded separately below.

The three awaits are fulfilled Promise, timer Promise, and real file read. `held` is dynamically allocated, owned across all three awaits. File data is dynamically allocated on completion, retained into the frame and released on success or throw. Throw mode raises `boom` after the third await; cancel mode raises `cancelled` on resumption from the timer. Both use the same rejection/cleanup mechanism. Root handlers are queued reactions, not direct print calls from the state machine's error exit.

Observed success stdout, identical native and Node bytes:

```text
start
end
after sync held
top microtask
timer
after timer held
timer microtask
after file payload
done held
```

Throw mode changes the last line to `caught boom`. Cancellation ends after `timer microtask` with `caught cancelled`, omitting `after timer`, file and done lines. Node stderr is empty and exit 0 in all modes. Native program stderr contains only the independently checked count line. The harness separates this instrumentation from language stderr; the remaining stderr is empty, including sanitizer diagnostics on controls.

| Mode | Counted allocations | Counted frees | Retains | Releases |
| --- | ---: | ---: | ---: | ---: |
| success | 7 | 7 | 12 | 19 |
| throw after file await | 8 | 8 | 12 | 20 |
| cooperative timer cancellation | 6 | 6 | 7 | 13 |

These are this prototype's Object-header counts, not Adamic ADAMIC_COUNT. Each allocation begins with one reference, so allocations + retains = releases; frees = allocations. Jobs, pthread resources and libc allocations are outside the counted objects, but remain subject to LeakSanitizer. The harness asserts the exact tuples as well as balance, without using the compiler/runtime to predict them.

All mutants compiled successfully under the same warning and sanitizer flags before running. These are executable preprocessor variants, never edits to the production runtime.

| Mutant run | Mode | What caught it |
| --- | --- | --- |
| MUTANT_INLINE: fulfilled observer runs inline | success | Node stdout comparison: eager resumption before end |
| MUTANT_CHECKPOINT: dispatch the later callback before the Promise checkpoint | success | Node stdout comparison: timer microtask prints before the await continuation |
| MUTANT_BORROW: frame does not retain file result | throw | ASan heap-use-after-free when frame/job ownership overlaps; real file buffer, no immortal literal |
| MUTANT_LEAK: frame omits release of held local | throw | LeakSanitizer reports the dynamically allocated held object |
| MUTANT_LEAK, separate cancellation run | cancel | LeakSanitizer reports the same held object |
| MUTANT_COUNTS: extra balanced retain/release of file result | throw | Exact count comparison (13 retains, 21 releases); Node and sanitizers pass |

Both scheduling mutants finish with exit 0 and unchanged counts, without sanitizer errors; only their Node stdout comparisons fail. The isolated leak mutant completes with correct stdout and balanced scheduling, so only the memory checks see the lost local. UBSan ran on all variants but no dedicated UB mutant was added, and no claim is made that this artifact independently validates UBSan instrumentation.

Reproduce on Linux with clang sanitizer runtimes, Python 3 and Node 24:

```sh
bash cloud/setup.sh > /tmp/p286ycm-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 prototypes/async/check.py > /tmp/p286ycm-check-final.log 2>&1
```

Observed setup: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 74s, total 74s. `nproc` 5; cgroup `cpu.max` `400000 100000`. Go 1.27.1, clang 20.1.8, Node 24.19.0. Harness exit 0, prints all three count tuples, all six mutant verdicts, then PASS. Each command's stdout/stderr goes directly to separate files under the printed scratch directory; the final run used `/tmp/adamic-gate/p286ycm-0j38voa8`. No test output is piped.

Existing base concurrency checks also passed, with all output logged:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestConcurrencyRefusals|TestParallel' > /tmp/p286ycm-oracle.log 2>&1
# ok github.com/system-inc/adamic/internal/oracle 14.617s
go test -count=1 -timeout 30m ./internal/native -run TestParallel > /tmp/p286ycm-native.log 2>&1
# ok github.com/system-inc/adamic/internal/native 21.461s
go vet ./... > /tmp/p286ycm-vet.log 2>&1
# exit 0, empty log
```

In the original design-only commit no production package changed. The standalone prototype checks and filtered existing concurrency gates were run instead of the complete `go test ./...` gate. These existing checks exercise part 1 on the base; they are not evidence of compiler async support. Repository gofmt and staged diff whitespace checks were clean.

## Units needed to land it

Estimates below are reviewable implementation units, not elapsed-time promises. One unit means a separately reviewed commit series with its own controls, refusal fixtures and targeted mutants. The core architecture is approved; each expanded semantic surface still needs its own review. Basic async I/O needs units 1 through 7 (roughly 10 worker-sized units after splitting the larger rows); networking, structured cancellation and pool adapters are additional work.

| Piece | Estimated units | Deliverable and exit evidence | Depends on |
| --- | ---: | --- | --- |
| 1. Compiler integration first, closed primitive subset | 1 | This landing: named straight-line async functions, separate suspension IR, counted generated frames/Promises/reactions, C state machines and native async JavaScript output. Plain/fulfilled awaits, nested calls, return await and throws match Node; identity-spoof, resume-value, retain and release mutants. Synchronous lifetime optimizations are bypassed. Remaining semantics have named NotYet diagnostics. | Approved architecture |
| 2. Shared lowering and flow-graph suspension cutting (DRY #xjdce2d) | 2 | Replace the separate async grammar: ordinary statement/expression lowering produces ordinary IR with await nodes and async function metadata. Normalize evaluation order, then cut the flow graph into states, one per region between awaits. Delete lowerAsync and its grammar helpers, remove the hasAsync early-return route, and share synchronous feature lowering. Node fixtures for mixed sync/async modules, control flow and nested-expression awaits; wrong-resume and missing-live-slot mutants. | 1 |
| 3. Host loop and timer adapter | 1 to 2 | libuv dependency/build policy, monotonic timers, microtask checkpoints, loop liveness and shutdown. Node fixture for timer resolution followed by microtask; omitted-checkpoint and wrong-phase mutants. | 1, 2 |
| 4. Promise payloads, reactions and suspension ownership | 2 | Extend unit 2's shared pipeline with typed payload destruction, multiple observers, once-only settlement, chaining and root registrations; prove live-across-await ownership, initialization masks and handler/finally completion storage. Disable unsafe borrowing/regions/reuse until proved. ASan/LSan/count and user-cycle mutants, plus suspension-aware Node traces. Thenable adoption is separately scoped. | 1, 2 |
| 5. C emission and JavaScript backend | 2 | Stable frames, eager entry, queued resume, returns/rejections, synchronous callee exception-word isolation. JS emits native async/await with identical inserted checks. Three-way oracle including loops, closures, nested awaits, partially evaluated expressions and awaiting finally; missing-retain/release mutants. | 2, 3, 4 |
| 6. Async file service | 1 | Bounded requests, complete reads/UTF-8/error mapping, descriptor close, request buffer lifetime, drain at exit. Node fixtures for empty, long, missing and permission-denied files and controlled completion order; leak/UAF mutants. | 2, 3, 5 |
| 7. Promise combinators and integration gate | 1 | all input ordering, empty input, first observed rejection, live siblings, chaining cleanup, full uncached oracle and counts. Wrong-index, wait-all-before-reject and sibling-drop mutants. | 2, 5, 6 |
| 8. Explicit async scopes and cancellation | 2 | Parent/child obligation analysis, refused detachment, cooperative token library, unlink/generation protocol, finally that awaits, late I/O acknowledgement. Node semantics supplied by the same scope library. Late-completion UAF, skipped-finally and dropped-child mutants. | 1 through 7 |
| 9. Network I/O | 2+ | Socket/DNS/connect/read/write/close ownership, backpressure, bounded buffers, cancellation and errors. Local deterministic server/trace oracle; sanitizer and fd-leak controls. TLS and full fetch/HTTP API are separately scoped units. | 3, 5, 8 |
| 10. Pure-work pool adapter | 1 to 2 | Nonblocking pool submission/completion, part 1 graph marking and error protocol, explicit Node witness, one-worker/nested completion controls and TSan. No Promise.all rewrite. | 7, 8 and part 1 ABI |
| 11. Broad promise-shaped compatibility | 2+ | User executors/resolver escape, thenables, non-Error rejection representation, finally assimilation, never-settling APIs and complete lifecycle proofs. Only remove NotYet after matching source behavior. nextTick/setImmediate/ESM and unhandled-rejection behavior each need their own scope and phase fixtures. | 1 through 8 |

A first implementation can stage straight-line async lowering before loops and awaiting finally, provided all excluded constructs diagnose NotYet rather than compiling with partial semantics. It cannot announce general async support while those edges silently lose ownership. Networking and the pure-work adapter are independent additions after the foundations, not requirements to validate this prototype.

## Second landing unit: ordinary lowering, then state machines

Unit 2 replaces the first unit's closed grammar; it does not extend it. Kirk's standing DRY rule (#xjdce2d) applies here. Async bodies go through the ordinary statement and expression lowerers into ordinary IR, with await as an IR node and async metadata on ordinary functions. Synchronous functions beside async functions use that same path. After lowering, expression normalization exposes suspension boundaries in source evaluation order, and the flow graph is cut into a state machine with a state per region between awaits. Branches, loops and exception edges remain graph edges, rather than requiring a second AST grammar. Every feature the synchronous lowerer has therefore uses its existing lowering in async bodies without a copy. The unit deletes `lowerAsync` and its statement/expression grammar helpers, removes `if hasAsync { return lowerAsync(modules) }`, and replaces the separate `AsyncProgram` source-lowering route. Subsequent units extend the shared IR, graph and runtime, never grow that deleted grammar.

Checked against first-unit commit `3a2dbd6`: there is no concrete blocker requiring source lowering inside a state machine. `asyncBody` currently chooses states while walking statements, but those choices depend on await boundaries and control flow that can be represented after ordinary lowering. `asyncValue` already calls the ordinary expression lowerer for its accepted subset. The current flow builder already models ordinary branches, loops and exception/finally paths; it needs suspension edges and await normalization, since an instruction can currently contain a whole expression. Normalization must preserve partially evaluated operands in temporaries and resume both fulfillment and rejection at the correct graph successor. Ordinary function/call typing must distinguish the returned Promise handle from its awaited payload; unwrapping the payload belongs to await, not to all calls of an async function. These are shared IR and flow extensions, not reasons to duplicate statement or expression syntax.

The cutter determines frame slots from values and owned environments live across suspension, including captured cells that the current flow builder excludes from its tracked locals. It records initialization and handler/finally completion state, preserves thrown values across suspension instead of relying on a thread-local exception word, and binds generated frames to the existing counted Promise/reaction lifecycle and identity-based cycle exemption. Frame layout, state numbering and resume registration happen after lowering. Borrowing, regions and reuse stay disabled across awaits until suspension-aware proofs permit them. Sharing syntax does not by itself prove these lifetime rules or implement missing host APIs: unsupported suspension semantics and networking, cancelling pending I/O and the full Promise surface retain named NotYet diagnostics; permanent refusals remain refusals.

Exit evidence includes the existing first-unit Node and ownership controls, plus mixed synchronous/async functions, branches/loops, nested awaits with observable operand order, objects/classes/closures and try/catch/finally wherever their suspension ownership is proved. The ordinary graph/trace checks must cover async bodies instead of skipping `Program.Async`; suspension tracing must represent resume and rejection edges. Each newly claimed semantic surface needs its own Node fixtures and failing mutants. A feature still lacking that proof stays explicitly NotYet through the shared pipeline, without an alternate lowerer.

## Unit 2 state-cutting checkpoint, October 7, 2026

Status: **design for compiler review, before implementation**. Branch
`codex/async-ordinary` starts at `55d2c89b5de849ac3789eb7810200184530eff05`
and merges `origin/codex/call-targets` at
`280fc491fa8cfb5fb67cd664444e24e7107ad3a0` first. This section describes the
replacement pipeline, not features already built. The October 9 landing must
supply the evidence below. Implementation stops at this checkpoint for review.

### Ordinary IR is the only semantic input

The ordinary function registry, signatures, statement lowerer and expression
lowerer handle declarations, instance/static methods, arrows and function
expressions. An ordinary `ir.Function` records whether it is async, its callable
result (a Promise handle), and the body's result type. `ir.Await` holds an ordinary
expression operand, its proven payload type, and source location. Await unwraps
that type; calling an async function does not unwrap it. Closures and virtual
methods use the existing call representation and receiver/capture rules.

Delete `lowerAsync`, its syntax/grammar helpers, the `hasAsync` early return,
and the `AsyncProgram` source route. State-machine metadata belongs to ordinary
functions and is derived from their completed bodies. It contains block/state
ids, suspension successors and slot layouts, never checker AST nodes. The cutter
must not import the parser or checker, call a source lowerer, or encode a second
statement grammar. The ordinary body remains available for the JavaScript
backend, with native async/await and the same inserted checks. Native emission
uses derived state metadata but emits ordinary IR operations through the existing
statement/expression emission. Deleting a route cannot mean copying its emitter
or copying synchronous feature lowering into an async visitor.

The backend hooks for ordinary function async metadata and `ir.Await` require
small JavaScript dispatch changes when implementation resumes. Those hooks are
an integration seam outside this unit's originally listed territory; this
checkpoint does not edit that backend or the protected native assembly files.
The compiler must assign that seam before the separate `AsyncProgram` route can
be removed end to end.

### Normalize evaluation before making suspension edges

One IR normalization pass exposes computations surrounding awaits. It visits
ordinary IR operands in their established evaluation order, materializes already
evaluated values as typed temporary locals, and represents the await result as
another local. Every normalized instruction evaluates without suspension; a
suspension is a distinct graph terminal. A verifier rejects any residual await
inside an executable instruction before native emission.

For `combine(first(), await second(), third())`, the normalized order is:
evaluate and own the first result; evaluate second's operand and subscribe;
return to the loop; on fulfillment bind the awaited result; evaluate third; call
combine. A rejection skips third and combine and selects the surrounding
handler. Method receivers, computed keys, callee values and earlier arguments
are snapshots too. A property value read before await is retained as a value;
it is not read again from its possibly changed holder after resume. For an
assignment, preserve the evaluated destination receiver/key until the store.
Object spread snapshots occur where ordinary IR requires them, before later
field initializers. No C operand evaluation order is used to infer source order.

Conditional, optional and short-circuit operations normalize into guarded
regions so the unselected operand is never evaluated or subscribed. A loop's
condition, body and update normalize in their original positions; each iteration
reevaluates its own await operand. For-loop continue still reaches the update,
while while-loop continue reaches the condition. Break reaches the existing exit.
Normalization does not move an effect or an inserted check across an await,
a branch or a potentially throwing operand. Await-free subtrees remain ordinary
operations. This is an IR transformation, with one shared operand-order visitor,
not a transcription of the source expression lowerer's feature switches.

### Graph boundary and the shared SSA dependency

Use Adamic's `flow.Function`, blocks, instructions and terminal successor
adapter. A proposed `Suspend` terminal has a fulfillment successor, a rejection
successor, and the await operand/result metadata. Its rejection successor is
present even when no target body has an explicit throw: Promise rejection is
an independent outcome. Fulfillment defines the result only on its own edge,
through a dedicated binding block; rejection binds the owned error on its edge.
The resume entry must not claim the result definition dominates a catch.

Graph maintenance and SSA use cohere's `static_single_assignment` module through
its `Graph` interface. The adapter supplies entry and block access, instruction
uses/definitions, successor enumeration, predecessor updates, phi access and
identifier creation. `Suspend` contributes both logical successors to this
interface, liveness and dominance. Runtime parking is not an exit in that logical
graph. All resumed entries are reachable from the original entry through those
edges. Captured mutable cell contents are not renamed as SSA locals; the stable
environment/slot identities and explicit snapshot locals are tracked instead.

Observed dependency availability: on October 7, `origin/codex/shared-ssa` was not
advertised by Adamic's remote. Fetching abbreviated cohere ref `d19d0023` failed
with `couldn't find remote ref`; that object is absent locally. Fetched cohere
main `e7cfe4d1` has no `static_single_assignment` directory. Consequently this
checkpoint assumes the compiler's stated `Graph` contract; it does not claim to
have verified its Go method signatures. Resolve the exact package path and type
arguments against `d19d0023` or later when shared-ssa is published, without
copying its algorithms or pinning cohere backward.

The required call migration is explicit:

| Current Adamic call | After shared-ssa lands |
| --- | --- |
| `flow.ReversePostorder(function)` | module `ReversePostorder` on the flow Graph adapter |
| `flow.MarkPredecessors(function)` | module predecessor maintenance on the adapter |
| `flow.MarkEvaluationOrder(function)` | module evaluation-order maintenance on the adapter |
| `flow.Finalize(function)` | the above module operations in that order; use shared-ssa's retained adapter wrapper if supplied |
| `flow.Construct(function)` | module SSA construction on the adapter |
| `flow.EliminateRedundantPhis(function)` | module redundant-phi elimination, if not internal to construction |
| `flow.VerifySSA(function)` | module SSA verifier on the adapter |

The last two names must follow the exported module API rather than inventing
new exports. No cutter code goes into `graph.go`, `ssa.go`, `ssa_eliminate.go`
or `ssa_verify.go`, and it reads none of their private bookkeeping. Rebuild
maintenance facts after block splitting, construct/verify SSA on the normalized
logical graph, then compute liveness including exceptional successors. Cutting
adds a state map over those blocks; it does not invalidate the SSA graph. Any
later graph mutation must rebuild its maintenance and analysis facts.

### Partition blocks without duplicating shared continuations

State zero starts at the ordinary entry. Each fulfillment/rejection binding
block following a suspension is a region entry. Starting from these entries,
walk synchronous successor edges without crossing `Suspend`. A join reachable
from different entries becomes its own region entry; repeat until every block
belongs to exactly one entry. Promote an entry into a loop when external entries
would otherwise give that loop multiple owners. This finite procedure only adds
entries from the finite block set. Number entries deterministically in the
module's reverse postorder, retaining zero for the eager entry. A state is one
single-entry region between suspension boundaries; it can contain ordinary
branches and cycles with no await. Shared joins are emitted once.

An edge inside a region is ordinary control flow. An edge to another region
without an await is an immediate jump in the same resume invocation, with edge
copies/cleanup performed first. It neither returns to the scheduler nor creates
a microtask. Only `Suspend` saves the selected resume state, installs a reaction
and returns. Completion settles the output and releases task-owned storage.
This distinction prevents an if/loop join from adding observable Promise turns.
Phi operands are selected on predecessor edges, before entering the destination;
loop-carried values and cross-region joins cannot be loaded from an arbitrary
predecessor or reread from mutable source storage.

The graph verifier checks unique region membership, entry reachability, all
successors mapped, no residual instruction await, and no bypass of a suspension.
Trace checks walk eager, park, fulfill/reject and immediate-jump events against
the same logical graph, rather than skipping ordinary async functions. State
numbers themselves are not semantic; the selected successor and payload are.

### One environment allocation for captures and suspension storage

Observed `origin/codex/nested-functions` at
`b15216dabf65ffaa7152f6e64709b7b062ea01a9` exposes
`ir.AllocateEnvironment{Cells []int}`, `Function.FrameEnvironment`,
`Local.EnvironmentCell`, and `Local.Preallocated`. Its runtime environment owns
interior `adamic_cell` slots; retaining a slot retains the whole environment.
That is the right allocation construct for values live across an await.

Use that single allocation site per enclosing function activation, extending
its layout metadata for async control storage. Take the union of lexical capture
slots, values live on either resume edge, operand snapshots, parameters/this
needed after the eager entry, loop iterator owners, and pending completion/error
storage. Synthetic private locals use the same slot mechanism as captured state.
Stable cell addresses survive suspension. Globals themselves are not copied into
the environment; a read evaluated before suspension has its own owned snapshot.
Inherited closure environments are retained, never borrowed from the caller's
stack. Each async invocation has its own activation environment.

The native allocation has one heap header and owner, with the existing
`adamic_async_frame` prefix followed by slot storage and initialization metadata.
It must not allocate a separate captured-value environment beside an async frame.
Synchronous environments use the same IR construct and slot access/destruction
rules, with no async control prefix. To support both owners, nested-functions'
currently typed `adamic_environment *` interior-cell owner needs a common heap
owner pointer (or an equivalent shared owner accessor), with counting routed
through that owner. This is an internal environment/cell integration change,
not a Promise ABI change. It touches closure/cell runtime code outside this
unit's original async-runtime territory and must be coordinated with its owner;
do not duplicate the representation to avoid that coordination.

A captured source binding and a private spill have different lifetime ends.
Finish clears output/waiting, private spills, and pending completion; a captured
binding still reachable by an escaped closure remains initialized until that
owner is destroyed. Do not clear a captured binding simply because the task
completed. Per-slot ready/owned bits distinguish not-yet-initialized storage,
undefined values and moved-out values. Reassignment acquires the new count before
clearing/releasing the old one. Every cleanup path clears a slot before release.
Loop-local lexical bindings retain ordinary per-iteration identity; moving a
binding into suspension storage must not collapse distinct captured iterations.

Async allocation is heap-placed, since the eager call returns before resume.
Future region analysis may stack-place ordinary nonescaping environments through
the same allocation construct, but cannot stack-place a suspended activation.
Initially disable borrowing, lent reads, consuming moves, reuse and statement
regions for async bodies and any value/statement spanning suspension. Retain
parameters, receivers, snapshots and inherited environments before eager entry
returns. This conservative boundary includes values used by catch/cleanup, not
only the happy-path live set. Unsupported cycle/freshness proofs remain specific
NotYet or ordinary permanent cycle refusals; protocol provenance exempts only
audited generated edges, never user captures or payloads.

### Throw edges, resume ownership and the unchanged Promise ABI

Call target lookup is exclusively `Program.CallTargets` for ordinary calls and
`Program.ClosureTargets` for function-value calls, including virtual methods.
Use their `CallMayThrow`/`ClosureMayThrow` helpers for synchronous throw effects;
Unknown closure targets are conservative, not empty. Neither normalization nor
cutting reads `Call.Function` to decide which body, result obligation or throw
edge can run. Method overrides and closures returning rejected Promises must be
represented by the whole target set. The merged `TestCallTargetReaders` guards
that seam and must remain enabled with no new async allowlist exemption.

Distinguish an error evaluating the await operand, an error from a synchronous
callee during that evaluation, and rejection delivered by the await reaction.
Each reaches the lexical handler through its corresponding graph edge. An async
callee's body error rejects its returned Promise; it does not escape its eager
call synchronously. A returned Promise can reject even when a body has no throw,
so suspension rejection edges are unconditional. MayThrow propagation must model
this call-entry/body distinction while still obtaining targets through the shared
lookup. No target-specific shortcut can erase a resumption's rejection edge.

On resume, retain/copy a delivered reference or Error into an initialized owned
slot before the reaction releases its payload owner. Move synchronous pending
exceptions into the frame's completion storage and clear the exception word
before returning to the loop. Synchronous callees still use ordinary cleanup
paths. A caught rejected await enters the same catch as a synchronous throw;
await in that catch remains NotYet. A non-suspending finally uses ordinary
completion override/cleanup control flow; await in finally remains NotYet.

Keep unit 1's `adamic_async_promise` fields and the signatures/ownership of
`adamic_async_new`, `adamic_async_settle`, `adamic_async_await`,
`adamic_async_cancel`, `adamic_async_run` and `adamic_async_teardown` unchanged.
Payloads still use `adamic_value` plus the `references` flag; reference payloads
are ordinary counted objects, class instances, strings or arrays. The frame
prefix's `resume` and `children` callbacks remain unchanged. Broader slot layout
is compiler-private tail storage. No change to the host-promises producer ABI
is proposed. Unit 1 settlement detaches the Promise's reaction list, clears
frame waiting edges and transfers ownership to queued jobs; cancel and normal
exit detach/clear without source resumption. Use that exact protocol for the
new environment. A completed task releases its output edge even when an escaped
closure still holds the activation, avoiding retention of obsolete protocol roots.

Source cancellation, queued-job cancellation and cancelling pending I/O remain
NotYet. A pending Promise without host handles is abandoned only at normal loop
exit, with no execution of source finally, as already specified. The three
terminal detach paths remain separately leak-tested with their own mutants.

### Evidence required after this review

All new source fixtures are `.a`. Each supported behavior runs as source on
Node, native under ASan/UBSan and leak checking, and the JavaScript backend.
Use dynamically built strings/objects for lifetime evidence. Add the following
independent controls rather than treating one broad fixture as all proofs:

| Surface | Required failing mutant and decisive check |
| --- | --- |
| Functions, instance/static methods, arrow/function closures called from sync and async code | wrong target/resume successor or payload; Node output and call-target guard |
| Sequential awaits and fulfilled/plain awaits | inline or reordered resume; Node eager-prefix/FIFO trace |
| If/else and while/for with break/continue | wrong branch/update/back edge; Node output and graph trace |
| Earlier operands, receiver/key/callee snapshots and guarded awaits | reevaluate/move an operand or evaluate an unselected operand; Node effect trace |
| Try/catch around await, rejected await caught, synchronous operand throws | drop rejection/throw edge or leave exception word set; Node output and graph trace |
| Objects, classes, arrays, strings and captures live across await | omit an owned retain, spill or slot drop; ASan or LeakSanitizer |
| Escaped closure observes a binding after task completion | clear captured binding on finish; Node output or ASan |
| Settle, cancel and never-settled exit cycles | one skip-detach mutant per terminal path, registry roots still removed; LeakSanitizer |
| Shared environment placement and iteration bindings | borrow caller storage or collapse iteration slots; ASan or Node output |
| Unsupported suspension/Promise surfaces | change each named NotYet to acceptance or Refused; diagnostic fixture |

Retain the existing provenance-spoof controls and their mutants. Await in catch,
await in finally, Promise.all, arbitrary Promise executors and thenables remain
explicit NotYet through the ordinary pipeline. No feature is accepted solely
because normalization can encounter its node.

After implementation, run the four-package gate, filtered async oracle including
new fixtures and mutants, exact counts update/check, then the full gate with
output files. Report actual counts deltas per async fixture, separately from
merge-only changes to unrelated rows. This design checkpoint adds no fixture
and predicts no numeric counts delta; record measurements only after generated
code exists. Full async support and the October 9 gate are not claimed here.

### Checkpoint validation, before implementation

Setup ran with `bash cloud/setup.sh > /tmp/async-ordinary-setup.log 2>&1`,
then shells sourced `/workspace/adamic-tools/env.sh`. Observed timing lines:
Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache
warm 106s, done 106s. `nproc` is 5; cpu.max is `400000 100000`.
Go 1.27.1, clang 20.1.8 and Node 24.19.0 were used. The initial attempt to
format the conflict resolution before setup found no `gofmt` on PATH; the
post-setup repository formatting check is clean.

Commands and observed results, each with output redirected directly to a log:

| Command | Result and log |
| --- | --- |
| `go test -count=1 -timeout 30m ./internal/lower ./internal/flow ./internal/ir ./internal/native` | exit 0; lower 30.165s, flow 75.313s, IR 1.458s, native 138.033s; `/tmp/async-ordinary-packages.log` |
| `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/async_|TestAsync'` | exit 0, 45.060s; `/tmp/async-ordinary-oracle.log` |
| `python3 prototypes/async/check_cycles.py` | PASS; `/tmp/async-ordinary-cycles.log`; artifacts `/tmp/adamic-gate/p286ycm-cycles-bwvgx7dx` |
| `go test -count=1 ./internal/ir -run '^TestCallTargetReaders$'` with temporary direct-read mutant | expected exit 1, unapproved `Call.Function` read; `/tmp/async-ordinary-target-guard-mutant.log` |
| Same guard command after removing that mutant | exit 0, 0.462s; `/tmp/async-ordinary-target-guard-control.log` |
| `go vet ./...` | exit 0, empty `/tmp/async-ordinary-vet.log` |
| `gofmt -l cmd internal` and `git diff --check` | clean |

Nine existing or temporary mutants were run: wrong resumed number (only Node
stdout), missing dynamic parameter retain (ASan use-after-free), missing
throw-frame local release (LeakSanitizer), skipped settle detach, skipped cancel
detach and skipped exit detach (each LeakSanitizer), lost next-link ownership
(ASan with two observers), inline fulfilled resume (only Node stdout), and the
compilable direct call-target reader (the merged guard). The temporary Go reader
was removed before the restored guard run. These establish the inherited
mechanism and guard, not the proposed cutter or expanded syntax.

The initial counts invocation filtered to the four async fixtures was unsuitable:
the harness compares the whole table and failed with no changed row values.
The complete uncached counts check also failed with no changed row values
(`/tmp/async-ordinary-counts-full.log`, 15.981s). The merge had placed eleven
inherited rows after concurrency rows instead of before them. Regeneration with
`go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$'
-args -update-counts` passed in 15.354s and changed only that ordering, with no
numeric row changes. The complete counts check after repair passed in 14.277s, exit 0
(`/tmp/async-ordinary-counts-final.log`).

All four async rows are unchanged from `55d2c89`: plain 21/21 allocations/frees,
three 20/20, nested 22/22 and throw 12/10; retains/releases/peak/regions are also
unchanged. No new fixture was added. The full gate is deferred until the
reviewed implementation; this checkpoint does not claim expanded async support.

## First compiler landing and cycle evidence

The lowering hook in `internal/lower/lower.go` dispatches to new `async.go` after the permanent refusal preflight. `internal/ir/async.go` carries straight-line states separately from synchronous tree bodies; `internal/native/async.go` emits eager starts and queued resume functions, and `internal/javascript/async.go` emits native JavaScript async/await. The native and JavaScript assembly hooks are small. Frames use Adamic's existing heap/header/destruction, not a second reference counter. Reference parameters and awaited strings are retained; all frame fields start zeroed and each local is initialized once, so drop can visit every reference field without a per-state mask. All locals are conservatively retained until frame destruction. Synchronous borrow, freshness, reuse and region optimizations do not run over suspension states. The synchronous SSA/Node path tracer explicitly excludes `Program.Async` rather than claiming an empty trace proves its graph. Suspension-state tracing is NotYet; async semantics are held by the dedicated three-way oracle and ownership mutants.

Covered source is one dependency-free module with named nongeneric async functions, required primitive parameters and results, straight-line initialized primitive locals, primitive expressions/templates, string methods, one-string console calls, explicit Error throws, and return/return await. Await accepts primitive values, primitive `Promise.resolve`, and direct calls to those async declarations. Async entry is eager; even fulfilled waits enqueue FIFO reactions. Root rejection uses the existing Node oracle wrapper's panic text/exit policy. This is not general Node unhandled-rejection timing. No new public cancellation API or implicit pool dispatch was added.

**Named NotYet:** networking; timer/file I/O; cancelling pending I/O; Promise executors, handles/aliasing, all/then/catch/finally/thenables and the full Promise surface; object/union payloads; control flow across awaits and try/catch/finally; nested-expression awaits; closures/classes; synchronous calls/functions in async modules; imported dependency/module evaluation; generic/default/destructured parameters; assignments; undefined/void-valued expressions and locals; nonstring console values; and throws other than new Error with one explicit message, including omitted messages and cause/options. Unknown syntax is diagnosed before C emission. A discarded named async call remains Refused, as do the existing void, ==, any and other permanent language refusals. `return` of an unawaited async call is not implemented; spell `return await`. Recursive async call graphs are outside this first slice as well.

Generated cycle provenance is an unexported canonical IR identity, checked by `fresh.RuntimeBreaksCycles` at the entrance to `lowering.findCycles`. The actual emitted frame spelling `adamic_generated_frame_0`, Promise spelling `adamic_async_promise`, and reaction spelling `adamic_async_reaction` are diagnostic metadata only. `TestAsyncGeneratedIdentityCannotBeClaimedBySource` accepts canonical identities, rejects fresh IR metadata with copied names, and refuses self-cycles in user classes with exactly each emitted spelling. The ordinary source `slotsOf` check has no name exemption. Primitive-only payloads exclude user cycles in the first slice; support for reference graphs will require ordinary cycle/fresh-write proofs, never the runtime exemption.

`prototypes/async/cycles.c` uses the production runtime to build the actual frame -> waiting Promise -> reaction -> frame cycle with a dynamic held string. Its two observing frames share the same output and waited-on Promise, exercising both owned frame edges and the reaction list. Abandonment transfers each next-link count to its traversal local before destroying that reaction; otherwise destruction would free the remaining list prematurely. `check_cycles.py` runs settlement, pending-subscription cancellation and never-settled normal-exit teardown under ASan/UBSan/default Linux LeakSanitizer. Each control exits 0 with `cycle clean` and no stderr. Each same-path mutant skips detach/clear, still removes the registry root, and LSan reports the remaining cycle. The fixture's sole loop thread is joined before process exit, eliminating stale stack/register pointers that would conservatively root leaked objects; controls and mutants use identical thread boundaries. No stack/register scanning is disabled.

The compiler fixtures are `async_plain.a`, `async_three.a`, `async_nested.a`, and `async_throw.a`; the main oracle compares source Node, emitted JavaScript, sanitized C, release C and sanitized slab C. Counts are recorded in `internal/oracle/counts.md`: plain 21 allocations/21 frees, three-await 20/20, nested 22/22. The uncaught-throw row stops at panic (12/10), as other panic fixtures do. `TestAsyncThrowReleasesFrame` consumes the root rejection in an internal C harness, releases it and exits normally; LSan is clean, and deleting a dynamic frame-local drop leaks. This separates observed frame cleanup from the existing fatal-panic policy.

Executable compiler mutants: resumed number +1 is caught only by Node stdout; missing dynamic parameter retain is caught by ASan use-after-free; missing throw-frame local drop is caught by LSan. Source mutants, restored after running: provenance granted by name fails the copied-identity check; skipping `slotsOf` for the exact generated frame name fails the user-class refusal check; a Promise-executor gap changed to Refused fails the NotYet check; unawaited-task refusal changed to NotYet fails the permanent-refusal check. The original six standalone prototype mutants remain independently reproducible.

Reproduce all new focused evidence with output files:

```sh
python3 prototypes/async/check_cycles.py > /tmp/p286-cycles.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/async_' > /tmp/p286-async-final-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestAsync|TestCountsAreRecorded' > /tmp/p286-async-final-mutants-counts.log 2>&1
go test -count=1 ./internal/lower -run TestAsync > /tmp/p286-async-lower.log 2>&1
```

Follow-up setup: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 16s, total 16s; `nproc` 5, quota 4. Tool versions are unchanged. The final cycle run also checks fulfilled-await FIFO order against the Node-only `cycles_oracle.a`: an inline-resume mutant finishes cleanly and fails only stdout comparison. A next-link mutant that omits this transfer is caught by ASan use-after-free with two observers. All child output is saved in the printed scratch directory.

Final observed checks: the four current-runtime Node fixtures passed in 31.244s; current compiler mutants and all count rows passed in 14.001s; final async lowering/diagnostic tests passed in 3.380s. The two-observer cycle run ended PASS in `/tmp/adamic-gate/p286ycm-cycles-649_uedy`. The original standalone prototype rerun also ended PASS, with unchanged exact counts and all six mutant kills.

The full uncached `go test -count=1 -timeout 30m ./...` was attempted and stopped after 13 minutes while unrelated stage-1 suites were still running. Its complete core oracle suite passed in 476.058s. It found the synchronous trace integration gap (fixed with the explicit suspension boundary above) and an existing `skip_items_share` pool mutant that escaped TSan once. The isolated retry caught that mutant as a TSan race in 1.115s. The post-fix touched-package gate `go test -count=1 -timeout 30m ./internal/lower ./internal/fresh ./internal/ir ./internal/javascript ./internal/native ./internal/flow > /tmp/p286-touched-gate.log 2>&1` passed: lower 59.685s, fresh 89.086s, native 198.719s, flow 159.018s; IR/JavaScript have no standalone tests. The later next-link cleanup correction was then rechecked by the two-observer ASan/LSan fixtures and final current-runtime async oracle/count/mutant runs above. Final `go vet ./...`, repository gofmt and diff whitespace checks were clean. This is a focused worker gate, not a claim that the entire uncached gate completed green.

## Limits of this evidence

The compiler has no host timer/file/socket service, complete Promise API, Node phase emulation, async scopes, source cancellation, pending-I/O cancellation, arbitrary reference payloads, await in loops/catch/finally, pool adapter or scalability/performance benchmark. The original standalone C timer/file runtime remains fixed to one timer and one small ASCII file request (127-byte payload); it is not the compiler's file API. No macOS/Windows run or TSan proof. The original helper thread joins before the loop reads its buffer. Broader async correctness remains work for the landing units above.


## Ordinary async implementation after checkpoint approval

Async declarations, instance and static methods, arrows and function expressions
now use the ordinary statement and expression lowerers. `ir.Function.Async` and
`AsyncReturns` describe an ordinary function; `ir.Await` is an expression.
`lowerAsync`, its grammar helpers, the early async module route and both backend
AsyncProgram routes are removed. Promise.resolve/reject are small shared lowering
leaves, not a source-language interpreter. Returning an unawaited Promise still
reports NotYet rather than claiming Promise adoption.

Normalization snapshots earlier operands and receiver reads before later awaits,
keeps conditional, boolean and coalescing operands guarded, and copies array and
object spread inputs at their evaluation point. Loop tests execute on each
iteration. Fulfillment binds its result and rejection follows the lexical throw
edge. MayThrow and closure effects use Program.CallTargets and ClosureTargets.
Async graph tracing parks an activation at await and restores that activation on
resume, including interleaved nested tasks; these functions now participate in
the ordinary SSA and Node path tests instead of being excluded.

`CutSuspensionGraph` works through EntryBlock, BlockOrder, Successors and
Suspension, promoting multi-entry joins to region roots until every block has
one owner. Native states name region entries; synchronous region edges are
immediate gotos, and only Suspend yields. The adapter currently obtains the
already-finalized reverse-postorder block list from flow.Function. When
codex/shared-ssa lands, replace this BlockOrder implementation with the shared
Graph ReversePostorder, and delegate entry and successor traversal to that
Graph adapter. Suspend remains an Adamic terminal. No implementation depends on
flow/graph.go, ssa.go, ssa_eliminate.go or ssa_verify.go. No shared module import
is claimed: after fetching depth 200 and then unshallowing cohere, the observed
origin/main was e7cfe4d1aceb524bc6f5936564284d54ca63d771 and d19d0023 remained
absent; origin/codex/shared-ssa was also absent. Reconciliation is pending its tip.

The merged nested-functions layout is used: one AllocateEnvironment at function
entry and Function.FrameEnvironment in ascending local-index order. The async
activation embeds the scheduler prefix, inherited closure and error edges, plus
the same adamic_cell layout and common initialize/drop helpers as a synchronous
environment. It is not a second captured-slot representation. All activation
locals are conservatively owned, rather than claiming minimal liveness. Private
reference slots clear before release at completion; captured slots remain for
escaped closures until environment destruction. Environment-cell owner is now
a heap-header pointer, with unchanged pointer size, so cells can name either
scheduler-prefixed or ordinary environments. Region placement stays disabled
for async activations. HasPromises conservatively disables borrowing, reuse and
region optimizations for the whole containing module; suspension also mutates
escaped values in flow effect inference.

The unit 1 Promise and reaction structs, frame prefix and existing protocol
function signatures are unchanged. Two additive compiler-private functions,
adamic_async_register_cleanup(frame, callback) and adamic_async_forget_cleanup,
register abandonment cleanup without owning the frame. They clear private
slots, inherited closure and output edges on cancel or normal-exit abandonment,
including a private closure that otherwise retains its own activation. Cleanup
records are removed even by skip-detach mutants so their raw pointers do not
hide leaks from LeakSanitizer. Source finally is not executed on abandonment.
The common cell helpers are adamic_environment_initialize_cells and
adamic_environment_drop_cells. The share walker normalizes interior cells to
owners and recognizes ordinary environments, fixing the merged runtime seam.

Named native hooks: asyncFunction (ordinary graph instructions and terminals),
asyncMain (root draining/rejection), promiseValue (settled values), asyncSlots in
cellReference/declareLocal, asyncHandler in jumpThrown, Promise in cType,
PromiseValue/Await expression dispatch, and appendsTo's refusal to use stack-local
string accumulation for frame slots. emit.go selects these hooks, includes the
async runtime and disables unproved optimization plans. Existing ordinary
statement, expression, store and exception emitters remain shared. JavaScript
uses Async metadata, Await/PromiseValue dispatch and loop Test; tracedAwait is
instrumentation-only, with Suspend/Resume options preserving activation traces.

Sixteen async source fixtures run through the Node/native/JavaScript oracle.
The twelve additions cover control flow, operand order, methods, closures, owned
reference values, caught rejection and synchronous throws, escaped captures,
awaited loop tests/updates, spread snapshots and receiver snapshots. Current
compiler mutants cover wrong resumed payload (Node stdout), missing dynamic
parameter retain (ASan use-after-free) and omitted throw-frame slot drop (LSan).
The production-runtime cycle fixtures separately pass settle, cancel and exit,
and each skip-detach mutant is caught by LSan; lost next-link ownership is caught
by ASan and inline fulfilled resumption by Node FIFO output. Additional expanded
callable/control mutants and final gates are recorded after their runs finish.

Observed setup: ready steps 0s, build cache 88s, total 88s; nproc 5. The complete
flow tests passed after enabling async path tracing. Lower, IR and native passed
the first package run; that run exposed an activation-trace parser error in flow,
which was fixed and the entire flow package rerun successfully. Tests are logged
under /tmp/async-ordinary-*.log. These observations are not a full-gate claim.

Counts (allocations/frees, retains/releases, peak; every async row has 0 regions):

| Fixture | Before | After |
| --- | --- | --- |
| plain | 21/21, 32/36, 13 | 21/21, 39/61, 15 |
| three | 20/20, 25/29, 11 | 20/20, 34/58, 12 |
| nested | 22/22, 38/41, 12 | 22/22, 45/70, 13 |
| throw | 12/10, 16/15, 10 | 12/10, 26/35, 10 |
| control | new | 39/39, 88/115, 13 |
| operand_order | new | 18/18, 27/51, 10 |
| methods | new | 21/21, 40/71, 13 |
| closures | new | 25/25, 51/86, 17 |
| live_values | new | 24/24, 35/61, 21 |
| catches | new | 23/23, 45/71, 16 |
| escaped_capture | new | 23/23, 53/80, 18 |
| conditions | new | 53/53, 129/168, 13 |
| snapshots | new | 48/48, 92/153, 28 |
| receiver | new | 30/30, 65/100, 17 |
| mutation | new | 15/15, 31/51, 14 |
| catch_return | new | 18/18, 33/59, 12 |

The throw fixture intentionally exits through fatal root rejection; its separate
consumed-rejection harness checks balanced cleanup. Unrelated counts changed
only row ordering inherited from the nested-functions merge, not numeric values.

Explicit gaps remain: awaits in catch/finally, all async finally completion
routing, async for-of/switch, captured per-iteration cells, structural-method
operands containing awaits, boolean-or-undefined slots, Promise.all, arbitrary
executors, thenables and Promise adoption. Host services and source cancellation
remain separate units. The broad ordinary async surface is implemented, but the
approved checkpoint's non-suspending finally routing and
shared-module reconciliation are not yet complete; this section does not claim
the entire October 9 bar is met.


### Further implementation evidence before the main reconciliation

Expanded wrong-fulfillment mutants for control, operand order, instance/static
methods, arrow/function closures and awaited loop tests are caught only by Node
stdout. Reordering the first operand snapshot after an eager async operand, and
routing rejection through fulfillment, are likewise caught only by Node stdout.
String, object, class and array private-slot drop mutants are each caught by
LSan. Clearing an escaped captured binding at finish is caught by sanitizers;
the first attempted metadata mutant failed to compile and is not counted as a
kill. The replacement edits only generated completion cleanup and compiles.
Generated shared-environment abandonment controls with a private async closure
retaining its parent pass on cancel and never-settled exit; each omission of the
cleanup callback leaks and is caught by LSan, after joining the loop thread.

The new queued-mutation fixture makes an async callee mutate its caller's object
after suspension. Omitting Suspend's escaped mutation effect fails the independent
Node range check: the mutation at order 5 lies outside [2, 4). The restored
control passes. Redirecting fulfillment to the rejection successor fails Node's
graph path check. Removing Promise payload traversal accepts a user back-reference
and fails the cycle refusal test. Turning Promise gaps into Refused fails named
NotYet controls; changing normalization failures into Refused independently
fails await-in-catch and await-in-finally controls. All temporary Go mutants were
restored. The numeric-only condition mutant initially survived because its
changes canceled out; the boolean-fulfillment mutant is the failing control.

A return inside a try with no finally previously traversed an empty return block,
which loses its payload for graph emission. Build now returns directly whenever
no open finally needs completion routing. async_catch_return.a checks fulfilled
and rejected return-await plus a catch return. Its three-way oracle passes.
Non-suspending finally remains NotYet: the existing flow Choose deliberately
overapproximates possible completions and has no exact completion discriminator
for native graph emission. It needs a shared completion representation rather
than a second async-only lowering of finally.

The first complete four-package implementation gate passes: lower 75.132s,
flow 210.100s, IR 7.171s, native 426.002s, in
/tmp/async-ordinary-packages-final.log. The return-routing changes trigger a fresh
gate rather than treating that earlier result as final. Counts regeneration with
both final new fixtures passes in 102.428s. The full uncached gate exposes the
freshness integration seam: internal/fresh recognizes neither Await nor
PromiseValue, so its known-write check fails. It also detects a counts-table
snapshot changed by adding fixtures during that run. A narrow conservative patch
for internal/fresh/fresh.go is prepared at /tmp/async-ordinary-fresh-seam.patch;
that territory extension is requested and not assumed. A green final full gate
is not claimed until that seam and main reconciliation are complete.


### Main reconciliation and final worker evidence

This branch incorporates origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965
through f0df6d1. The final fetch still reports that tip as an ancestor. No main
push or area-branch merge was performed. Implementation commits before the
reconciliation are 64a3eed and ac78ecb; nested environments were merged in 0fc1f05.
The merge keeps main's call-target freshness analysis and both fixture lists.
Async counts remain the measured values above after the merge.

Post-merge required packages pass in /tmp/async-ordinary-main-packages.log:
lower 69.720s, flow 178.109s, IR 35.281s, native 261.980s. The uncached async
oracle and existing async mutants pass in 92.476s. Complete counts regeneration
passes in 101.477s. Independent string-payload mutants for instance methods and
arrow closures pass in 3.209s, each killed only by Node stdout; these supplement
the numeric static-method and function-expression mutants, rather than treating
the numeric mutant as proof for a void-await instance method.

A focused pool probe compiled but panicked with "async values cannot cross worker
thread": its pure callback captured a cell owned by an async activation. The
shared IR pipeline now checks ParallelMap callback targets through ClosureTargets
and rejects captured async environments as NotYet before emission. Unknown
callback target sets are conservatively checked against possible functions.
Synchronous programs take a fast path. Omitting that check accepts the probe and
fails its diagnostic control. The restored control passes. Pool sharing of an
async environment is added to the explicit gaps; no sharing runtime or producer
ABI extension is claimed.

The final uncached command was:
`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run
'TestNativeAgreesWithNode/internal/oracle/testdata/async_|TestAsync|TestCountsAreRecorded'`.
It passes in 107.492s, with all sixteen source fixtures, every registered async
compiler mutant and the complete stable counts table, logged at
/tmp/async-ordinary-final-async-counts.log. After the pool guard, the changed
packages lower, flow and IR pass in 44.841s, 131.461s and 24.425s respectively
(/tmp/async-ordinary-pool-final-packages.log); native emission is unchanged from
the passing post-main package gate. Repository formatting, diff checks and
`go vet ./...` pass, with logs /tmp/async-ordinary-main-gofmt.log and
/tmp/async-ordinary-main-vet.log.

The proposed freshness seam is still outside granted territory. A temporary Go
overlay of the prepared patch passes the whole internal/fresh package in
103.653s (/tmp/async-ordinary-fresh-proposal-test.log), without modifying the
repository file. It recognizes Await and PromiseValue conservatively, disables
call freshness summaries in async-containing programs and escapes frame-held
values at Suspend. Applying /tmp/async-ordinary-fresh-seam.patch requires the
requested narrow territory extension. An overlay pass is not a repository gate
pass and is not reported as one.

The full uncached repository command was run, found the unknown Await/PromiseValue
freshness failure and a counts snapshot changed while new fixtures were being
added, then continued through the remaining stage-1 suites. Counts were repaired
and checked with a stable fixture list above. The already-failed full attempt
was stopped after more than 42 minutes under the explicit slow-worker-gate
exception; exit 143. Its full output is /tmp/async-ordinary-full-gate.log and the
termination manifest is /tmp/async-ordinary-full-stop.json. CSS, Unicode
properties, JSON and lint suites passed before stopping; remaining markdown and
type-aware suites were not completed. A final full green gate is not claimed.
The freshness seam, exact non-suspending finally completion routing and the
unavailable shared-SSA reconciliation remain blockers to reporting the entire
October 9 bar complete.

### Granted freshness seam and final completion scope

The compiler granted internal/fresh/fresh.go on October 7. The repository now
recognizes Await and PromiseValue as conservative calls, turns off call freshness
summaries whenever Program.HasAsync is true, and escapes every frame-held local
at Suspend. Ownership of those references is established; confinement across
queued work is not. The new async_fresh_holder.a fixture creates both a value and
a holder before an await and stores the value into the holder after resumption.
Its ordinary types are acyclic, so the program remains accepted while freshness
must decline the write proof. TestAsyncSuspensionEndsFreshConfinement checks that
specific field write. A mutant skipping the Suspend escape accepts that proof
and is killed by the test; the production implementation was restored.

Non-suspending finally is not a small emission hook. builder.route records
possible exits, and the finally ends in Choose without an exact runtime selector.
Returns routed through it also need a retained pending payload, rather than the
block-local return operand used by the current resume emitter. Correct lowering
must preserve normal, return, throw, break and continue completions through
nested cleanup, with replacement by a cleanup return or throw. This requires a
completion representation and its ownership proof. It remains explicit NotYet;
await in catch or finally remains NotYet as well. Shared-SSA reconciliation is
left pending the worker branch, as the compiler requested.

The current-main reconciliation is e9bfaf4, incorporating f8013f0. The freshness
implementation is 6d8afa4. Counts regeneration passes in 59.194s; the seventeenth
async fixture adds this measured row (allocations, frees, retains, releases, peak,
regions): async_fresh_holder.a = 12, 12, 23, 38, 11, 0. Existing async rows retain
their numbers; main's other fixture changes are incorporated by regeneration.

### Final granted-seam verification

The required lower, flow, IR and native packages plus fresh pass in
/tmp/async-fresh-packages.log (29.923s, 73.518s, 1.625s, 120.591s and 37.937s).
The seventeen-fixture async three-way oracle, registered async mutants and full
counts check pass in 30.896s (/tmp/async-fresh-oracle.log). The complete oracle
also passes on the latest-main reconciliation 8cbc05c, including main c01907a,
in 190.540s (/tmp/async-fresh-final-main-oracle.log). Main's only compiler-test
change since f8013f0 is the dormant Stage 3 oracle hook; the compiler implementation
and stage-one suites are unchanged. Neither main nor an area branch was pushed.

The complete command ran to termination:
`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...`.
It exits 1, logged in /tmp/async-fresh-full-gate.log. Every package except
stage1/cohere/markdownblocks passes, including the entire oracle, Unicode
properties, CSS, JSON, lint, markdowninline and typeaware. Markdownblocks first
reports a missing original Node width dependency at
/tmp/adamic-markdown-width/node_modules/emoji-regex/index.js, then reaches the
30-minute package timeout while layout/composition tests are still active.
The pinned dependencies from its documentation were installed outside the repo:
emoji-regex 10.6.0, get-east-asian-width 1.6.0, narrow-emojis 0.0.3, using npm
with scripts disabled. The focused retry
`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m
./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$'`
passes in 352.295s (/tmp/async-fresh-markdown-width.log). A complete green
markdownblocks package or full gate is not claimed; timed-out layout tests were
not rerun. There was no compiler-source edit to accommodate that dependency.

The added freshness controls independently prove both conservative switches:
TestAsyncSuspensionEndsFreshConfinement kills removal of the Suspend escape
(/tmp/async-fresh-mutant.log); TestAsyncProgramsDisableCallFreshnessSummaries
kills enabling summaries in an async program
(/tmp/async-fresh-summary-mutant.log). Both mutants were restored. The final
whole lower package passes in 11.961s (/tmp/async-fresh-final-lower.log).
Formatting, diff checks and vet passed before the documentation-only final
report. This seam adds no native emission hook and makes no Promise ABI change.
The previously listed hooks and cycle mutants remain in the earlier evidence.
Finally completion routing and shared-SSA reconciliation remain explicit gaps.

### Runtime area reconciliation counts

Base: area/runtime 915b9e05, merged in a64243ca. The regenerated Linux table has 514 rows, compared with 487 on the area. Fields below are allocations, frees, retains, releases, peak, regions. Only the following 13 existing area rows change. Every existing synchronous area row is unchanged.

| Fixture | Area counts | Merged counts | Cause |
|---|---|---|---|
| async_plain.a | 21 / 21 / 32 / 36 / 13 / 0 | 21 / 21 / 39 / 61 / 15 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_coverage_unions.a | 27 / 27 / 60 / 49 / 12 / 0 | 27 / 27 / 81 / 120 / 12 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_coverage_reject_empty.a | 5 / 4 / 9 / 7 / 4 / 0 | 6 / 4 / 15 / 18 / 5 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too; ordinary throwing constructs an Error object where unit 1 rejected with the message string; fatal exit leaves live values, as before |
| async_coverage_parameters.a | 17 / 17 / 18 / 22 / 12 / 0 | 17 / 17 / 24 / 46 / 12 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_coverage_typeof.a | 15 / 15 / 24 / 26 / 9 / 0 | 16 / 16 / 28 / 49 / 10 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too; ordinary function-value forwarder adds one closure allocation |
| async_coverage_values.a | 66 / 66 / 126 / 121 / 16 / 0 | 66 / 66 / 157 / 238 / 19 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_coverage_discard.a | 25 / 25 / 42 / 43 / 10 / 0 | 25 / 25 / 55 / 83 / 15 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_coverage_reject_eager.a | 12 / 9 / 18 / 16 / 10 / 0 | 12 / 9 / 27 / 36 / 10 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too; fatal exit leaves live values, as before |
| async_coverage_reject_nested.a | 18 / 15 / 31 / 26 / 17 / 0 | 18 / 15 / 45 / 64 / 17 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too; fatal exit leaves live values, as before |
| async_typeof.a | 15 / 15 / 24 / 26 / 9 / 0 | 16 / 16 / 29 / 50 / 10 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too; ordinary function-value forwarder adds one closure allocation |
| async_three.a | 20 / 20 / 25 / 29 / 11 / 0 | 20 / 20 / 34 / 58 / 12 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_nested.a | 22 / 22 / 38 / 41 / 12 / 0 | 22 / 22 / 45 / 70 / 13 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too |
| async_throw.a | 12 / 10 / 16 / 15 / 10 / 0 | 12 / 10 / 26 / 35 / 10 / 0 | Ordinary snapshots and owned frame slots replace the unit 1 straight-line grammar; extra NULL releases are counted too; fatal exit leaves live values, as before |

All 27 additions follow. They were already carried by async-ordinary, and are additions relative to the area, not changes to an existing area fixture.

| Fixture | Merged counts | Cause |
|---|---|---|
| async_control.a | 39 / 39 / 88 / 115 / 13 / 0 | Unit 2 ordinary async fixture |
| async_operand_order.a | 18 / 18 / 27 / 51 / 10 / 0 | Unit 2 ordinary async fixture |
| async_methods.a | 21 / 21 / 38 / 69 / 13 / 0 | Unit 2 ordinary async fixture |
| async_closures.a | 25 / 25 / 51 / 86 / 17 / 0 | Unit 2 ordinary async fixture |
| async_live_values.a | 24 / 24 / 35 / 61 / 21 / 0 | Unit 2 ordinary async fixture |
| async_catches.a | 23 / 23 / 45 / 71 / 16 / 0 | Unit 2 ordinary async fixture |
| async_escaped_capture.a | 23 / 23 / 53 / 80 / 18 / 0 | Unit 2 ordinary async fixture |
| async_conditions.a | 53 / 53 / 129 / 168 / 13 / 0 | Unit 2 ordinary async fixture |
| async_snapshots.a | 48 / 48 / 92 / 153 / 28 / 0 | Unit 2 ordinary async fixture |
| async_receiver.a | 30 / 30 / 65 / 100 / 17 / 0 | Unit 2 ordinary async fixture |
| async_mutation.a | 15 / 15 / 31 / 51 / 14 / 0 | Unit 2 ordinary async fixture |
| async_catch_return.a | 18 / 18 / 33 / 59 / 12 / 0 | Unit 2 ordinary async fixture |
| async_fresh_holder.a | 12 / 12 / 23 / 38 / 11 / 0 | Unit 2 ordinary async fixture |
| nested_minimal.a | 2 / 2 / 0 / 2 / 1 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_captures.a | 22 / 22 / 8 / 22 / 7 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_hoisting.a | 2 / 2 / 0 / 2 / 1 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_mutual.a | 10 / 10 / 46 / 52 / 5 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_returned.a | 21 / 21 / 10 / 29 / 11 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_array.a | 17 / 17 / 17 / 27 / 9 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_three_levels.a | 8 / 8 / 7 / 13 / 7 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_tdz.a | 2 / 0 / 1 / 0 / 2 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_tdz_write.a | 2 / 0 / 1 / 0 / 2 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_weak.a | 4 / 4 / 5 / 10 / 4 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_destructured.a | 13 / 13 / 7 / 17 / 6 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_destructured_tdz.a | 2 / 0 / 1 / 0 / 2 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_mixed.a | 7 / 7 / 9 / 12 / 5 / 0 | Shared nested-function environment fixture imported with unit 2 |
| nested_pattern_parameter.a | 6 / 6 / 5 / 10 / 5 / 0 | Shared nested-function environment fixture imported with unit 2 |

Relative to async-ordinary 0bb0fab, the following existing numeric rows change by importing the area. For each listed row the merged values equal the area's recorded values exactly. This is observed baseline inheritance of the area runtime/emission changes (including views, borrowing and inline counts); individual optimizations are not experimentally isolated here.

| Fixture | Previous async-ordinary counts | Merged counts | Cause |
|---|---|---|
| internal/oracle/testdata/async_methods.a | 21 / 21 / 40 / 71 / 13 / 0 | 21 / 21 / 38 / 69 / 13 / 0 | Ordinary async lowering delta above |
| internal/oracle/testdata/call_targets_element.a | 45 / 43 / 36 / 64 / 7 / 2 | 45 / 43 / 32 / 60 / 7 / 2 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_object_keys.a | 88 / 88 / 35 / 75 / 24 / 0 | 88 / 88 / 32 / 72 / 24 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_object_is.a | 163 / 163 / 29 / 198 / 7 / 0 | 163 / 163 / 21 / 190 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_object_order.a | 84 / 84 / 31 / 71 / 44 / 0 | 84 / 84 / 30 / 70 / 44 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_oct6_deep.a | 43 / 43 / 23 / 55 / 20 / 0 | 43 / 43 / 22 / 54 / 20 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_oct6_parameters.a | 43 / 43 / 16 / 44 / 10 / 0 | 43 / 43 / 9 / 37 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_oct6_release.a | 372 / 372 / 180 / 441 / 26 / 0 | 372 / 372 / 177 / 438 / 26 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_oct6_subclass_holder.a | 76 / 76 / 55 / 116 / 12 / 0 | 76 / 76 / 42 / 103 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_array_with.a | 68 / 68 / 57 / 123 / 10 / 0 | 68 / 68 / 43 / 109 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_array_flat_map.a | 50 / 50 / 66 / 115 / 17 / 0 | 50 / 50 / 60 / 109 / 17 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_array_flat.a | 33 / 33 / 84 / 100 / 19 / 0 | 33 / 33 / 41 / 57 / 19 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_array_spliced.a | 72 / 72 / 59 / 133 / 12 / 0 | 72 / 72 / 40 / 114 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_array_copy_within.a | 615 / 615 / 141 / 754 / 7 / 0 | 615 / 615 / 126 / 739 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_array_search.a | 87 / 87 / 11 / 98 / 11 / 0 | 87 / 87 / 10 / 97 / 11 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/json_stringify_scalars.a | 100 / 100 / 38 / 170 / 5 / 0 | 100 / 100 / 37 / 169 / 5 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/json_stringify_options.a | 74 / 74 / 30 / 83 / 6 / 0 | 69 / 69 / 34 / 82 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/json_stringify_numbers.a | 66 / 66 / 2 / 108 / 3 / 0 | 66 / 66 / 1 / 107 / 3 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_fnexpr_loops.a | 182 / 182 / 197 / 269 / 82 / 0 | 182 / 182 / 194 / 266 / 82 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_for_in.a | 79 / 79 / 135 / 123 / 65 / 0 | 79 / 79 / 130 / 118 / 65 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_for_in_keys.a | 38 / 38 / 110 / 79 / 34 / 0 | 38 / 38 / 108 / 77 / 34 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_for_in_live.a | 39 / 39 / 63 / 80 / 21 / 0 | 39 / 39 / 59 / 76 / 21 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/bitwise_sweep.a | 4152674 / 4152674 / 2076 / 4154757 / 9 / 0 | 4152674 / 4152674 / 1260 / 4153941 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/objects.a | 58 / 58 / 59 / 110 / 18 / 0 | 58 / 58 / 45 / 96 / 18 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/maps_and_text.a | 59 / 59 / 41 / 94 / 10 / 0 | 58 / 58 / 51 / 92 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/sorting.a | 61 / 61 / 57 / 75 / 38 / 0 | 59 / 59 / 57 / 73 / 36 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/classes.a | 16 / 16 / 24 / 42 / 10 / 0 | 16 / 16 / 16 / 34 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/indexing.a | 33 / 33 / 13 / 46 / 6 / 0 | 33 / 33 / 12 / 45 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/casts.a | 9 / 9 / 17 / 25 / 6 / 0 | 9 / 9 / 16 / 24 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/cast_fails.a | 2 / 0 / 5 / 3 / 2 / 0 | 2 / 0 / 4 / 2 / 2 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/updates.a | 27 / 27 / 27 / 57 / 9 / 0 | 27 / 27 / 23 / 53 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/string_index.a | 89 / 89 / 45 / 142 / 7 / 0 | 58 / 58 / 44 / 141 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/visits.a | 95 / 95 / 125 / 209 / 23 / 0 | 95 / 95 / 124 / 208 / 23 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/searches.a | 106 / 106 / 82 / 174 / 19 / 0 | 106 / 106 / 81 / 173 / 19 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/defaults.a | 38 / 36 / 17 / 52 / 6 / 2 | 38 / 36 / 15 / 50 / 6 / 2 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/search_halves.a | 40 / 40 / 10 / 44 / 20 / 0 | 38 / 38 / 12 / 44 / 18 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/strings_more.a | 128 / 128 / 30 / 164 / 11 / 0 | 112 / 112 / 38 / 163 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_math_number_math.a | 1280 / 1280 / 50 / 1331 / 8 / 0 | 1280 / 1280 / 25 / 1306 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_math_number_convert.a | 199 / 199 / 197 / 365 / 7 / 0 | 196 / 196 / 199 / 364 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_math_number_prototype.a | 71 / 71 / 2 / 74 / 8 / 0 | 71 / 71 / 1 / 73 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/navigation.a | 162 / 162 / 40 / 189 / 16 / 0 | 162 / 162 / 36 / 185 / 16 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/number_formats.a | 235 / 235 / 46 / 263 / 18 / 0 | 225 / 225 / 55 / 262 / 18 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/precision_range.a | 4 / 3 / 2 / 5 / 2 / 0 | 4 / 3 / 1 / 4 / 2 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/radixes.a | 696 / 696 / 97 / 332 / 25 / 0 | 695 / 695 / 96 / 330 / 25 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/timsort.a | 2769 / 2769 / 3779 / 6232 / 308 / 0 | 2769 / 2769 / 3778 / 6231 / 308 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/sort_top_level.a | 33 / 33 / 80 / 97 / 19 / 0 | 33 / 33 / 79 / 96 / 19 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/splices.a | 74 / 74 / 41 / 108 / 12 / 0 | 74 / 74 / 29 / 96 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/array_from.a | 94 / 94 / 57 / 132 / 25 / 0 | 94 / 94 / 56 / 131 / 25 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/weak_parent.a | 86 / 86 / 264 / 313 / 41 / 0 | 86 / 86 / 253 / 302 / 41 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/doubly_linked.a | 90 / 90 / 545 / 578 / 49 / 0 | 90 / 90 / 525 / 558 / 49 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/fresh_parser.a | 294 / 294 / 338 / 508 / 34 / 0 | 145 / 145 / 280 / 450 / 34 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/fresh_writes.a | 330 / 330 / 409 / 525 / 132 / 0 | 330 / 330 / 369 / 485 / 132 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/fresh_calls.a | 147 / 147 / 268 / 323 / 39 / 0 | 113 / 113 / 216 / 273 / 37 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/exceptions.a | 133 / 133 / 171 / 246 / 20 / 0 | 133 / 133 / 163 / 229 / 20 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/closures_throw.a | 418 / 418 / 1571 / 1806 / 142 / 0 | 418 / 418 / 1566 / 1801 / 142 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/finally_leaves.a | 157 / 157 / 61 / 176 / 10 / 0 | 157 / 157 / 59 / 174 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/reuse_foreach_global.a | 17 / 17 / 15 / 28 / 10 / 0 | 17 / 17 / 14 / 27 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/param_assigned_in_try.a | 46 / 46 / 32 / 67 / 12 / 0 | 46 / 46 / 29 / 64 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/named_function_values.a | 116 / 116 / 129 / 224 / 25 / 0 | 116 / 116 / 127 / 222 / 25 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/invariance_readonly.a | 30 / 30 / 39 / 52 / 16 / 0 | 30 / 30 / 38 / 51 / 16 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/tuples_kept.a | 31 / 31 / 54 / 74 / 13 / 0 | 31 / 31 / 46 / 66 / 13 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/undefined_strings.a | 15 / 15 / 21 / 43 / 6 / 0 | 15 / 15 / 18 / 40 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/maybe_booleans.a | 48 / 48 / 51 / 101 / 11 / 0 | 48 / 48 / 47 / 97 / 11 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/unions.a | 1449 / 1449 / 1111 / 2573 / 16 / 0 | 1449 / 1449 / 1106 / 2568 / 16 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/maybe_number_slots.a | 484 / 484 / 305 / 798 / 16 / 0 | 484 / 484 / 304 / 797 / 16 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/case_mapping.a | 203 / 203 / 20 / 217 / 16 / 0 | 203 / 203 / 19 / 216 / 16 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/undefined_elements.a | 20 / 20 / 46 / 43 / 9 / 0 | 19 / 19 / 46 / 42 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/map_zero_keys.a | 29 / 29 / 24 / 49 / 8 / 0 | 29 / 29 / 23 / 48 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/adversarial_order.a | 105 / 105 / 86 / 164 / 31 / 0 | 105 / 105 / 88 / 164 / 31 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/adversarial_exits.a | 120 / 120 / 80 / 175 / 25 / 0 | 120 / 120 / 60 / 155 / 25 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/adversarial_iteration.a | 59 / 59 / 120 / 152 / 21 / 0 | 59 / 59 / 118 / 150 / 21 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/normalize.a | 213 / 213 / 96 / 274 / 11 / 0 | 205 / 205 / 104 / 274 / 11 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/string_positions.a | 436 / 436 / 301 / 607 / 134 / 0 | 327 / 327 / 368 / 606 / 97 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/long_literals.a | 14518 / 14518 / 85 / 14610 / 8 / 0 | 6117 / 6117 / 85 / 14610 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_layouts.a | 70 / 70 / 112 / 159 / 27 / 0 | 70 / 70 / 103 / 150 / 27 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/ascii_scan.a | 48 / 48 / 14 / 54 / 18 / 0 | 47 / 47 / 15 / 54 / 18 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/size_class_churn.a | 675173 / 675173 / 270039 / 675213 / 8005 / 0 | 675173 / 675173 / 270038 / 675212 / 8005 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_as_interface.a | 372 / 372 / 349 / 514 / 60 / 0 | 372 / 372 / 303 / 468 / 60 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/optional_class_method.a | 40 / 40 / 47 / 74 / 9 / 0 | 40 / 40 / 44 / 71 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/spread_snapshot.a | 22 / 22 / 4 / 25 / 11 / 0 | 22 / 22 / 3 / 24 / 11 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/reuse.a | 76 / 76 / 57 / 122 / 16 / 0 | 76 / 76 / 55 / 120 / 16 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regions.a | 316 / 260 / 219 / 409 / 50 / 56 | 316 / 260 / 19 / 209 / 50 / 56 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regions_throw.a | 124 / 95 / 53 / 143 / 31 / 29 | 124 / 95 / 9 / 99 / 31 / 29 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/borrow_element.a | 113 / 113 / 68 / 129 / 15 / 0 | 113 / 113 / 67 / 128 / 15 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/borrow_element_throw.a | 38 / 38 / 34 / 54 / 8 / 0 | 38 / 38 / 30 / 50 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/borrow_element_virtual_store.a | 13 / 11 / 9 / 15 / 7 / 2 | 13 / 11 / 7 / 13 / 7 / 2 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/spread_undefined.a | 45 / 45 / 27 / 65 / 10 / 0 | 45 / 45 / 26 / 64 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/lent_reads.a | 36 / 36 / 12 / 47 / 9 / 0 | 36 / 36 / 13 / 47 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/trig_reduction.a | 133 / 133 / 14 / 148 / 7 / 0 | 133 / 133 / 1 / 135 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/sets.a | 224 / 224 / 13016 / 13113 / 65 / 0 | 224 / 224 / 13010 / 13107 / 65 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_map_set.a | 170 / 170 / 219 / 320 / 24 / 0 | 170 / 170 / 135 / 236 / 24 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_map_set_keys.a | 76 / 76 / 70 / 135 / 15 / 0 | 76 / 76 / 64 / 129 / 15 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_map_set_construct.a | 180 / 180 / 229 / 275 / 52 / 0 | 180 / 180 / 225 / 271 / 52 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/maybe_collections.a | 64 / 64 / 166 / 225 / 20 / 0 | 64 / 64 / 165 / 224 / 20 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/optional_strings.a | 52 / 52 / 28 / 78 / 8 / 0 | 43 / 43 / 37 / 78 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/replace_all_large.a | 200060 / 200060 / 200012 / 400075 / 200003 / 0 | 200057 / 200057 / 200015 / 400075 / 200003 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/collections.a | 363 / 363 / 395 / 565 / 101 / 0 | 362 / 362 / 392 / 550 / 103 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/gaps.a | 97 / 97 / 84 / 154 / 46 / 0 | 97 / 97 / 85 / 154 / 47 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/normalize_long_marks.a | 25 / 25 / 17 / 26 / 13 / 0 | 8693 / 8693 / 8521 / 8769 / 398 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/power_of_two_string.a | 19 / 19 / 2 / 22 / 4 / 0 | 19 / 19 / 1 / 21 / 4 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/declared_later.a | 56 / 56 / 76 / 128 / 15 / 0 | 56 / 56 / 75 / 127 / 15 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/return_panic.a | 13 / 13 / 14 / 27 / 9 / 0 | 12 / 12 / 13 / 25 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/return_panic_fires.a | 5 / 0 / 8 / 7 / 5 / 0 | 5 / 0 / 6 / 6 / 5 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/from_codes.a | 408 / 408 / 56 / 416 / 12 / 0 | 408 / 408 / 56 / 413 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/bitwise.a | 9304 / 9304 / 1406 / 9379 / 45 / 0 | 9304 / 9304 / 1369 / 9342 / 45 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/tuple_values.a | 119 / 119 / 126 / 207 / 37 / 0 | 119 / 119 / 112 / 193 / 37 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/generic_values.a | 22 / 22 / 39 / 52 / 10 / 0 | 22 / 22 / 33 / 46 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/undefined_references.a | 17 / 17 / 37 / 52 / 11 / 0 | 17 / 17 / 30 / 45 / 11 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/utf8_view.a | 59 / 59 / 34 / 80 / 7 / 0 | 59 / 59 / 33 / 79 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/search_from.a | 292 / 292 / 168 / 329 / 22 / 0 | 292 / 292 / 159 / 320 / 22 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/shared_slices.a | 12675 / 12675 / 5999 / 12774 / 87 / 0 | 12236 / 12236 / 10648 / 12767 / 86 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/string_append.a | 249 / 249 / 95 / 285 / 49 / 0 | 248 / 248 / 140 / 278 / 48 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/shared_slice_append.a | 200 / 200 / 20 / 206 / 8 / 0 | 200 / 200 / 37 / 206 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/search_from_sweep.a | 133 / 133 / 34 / 155 / 10 / 0 | 133 / 133 / 33 / 154 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/integer_format.a | 106826 / 106826 / 58737 / 146871 / 20 / 0 | 106826 / 106826 / 48056 / 136190 / 20 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/reuse_spread_method_alias.a | 8 / 8 / 8 / 15 / 6 / 0 | 8 / 8 / 7 / 14 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/try_assignments.a | 37 / 37 / 9 / 42 / 6 / 0 | 37 / 37 / 8 / 41 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_string_conversion.a | 15 / 15 / 4 / 21 / 3 / 0 | 15 / 15 / 3 / 20 / 3 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_string_prototype.a | 24 / 24 / 6 / 31 / 3 / 0 | 23 / 23 / 7 / 31 / 3 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_string_indices.a | 639 / 639 / 200 / 845 / 8 / 0 | 594 / 594 / 89 / 701 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_string_raw.a | 45 / 45 / 61 / 95 / 11 / 0 | 45 / 45 / 50 / 84 / 11 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_string_existing.a | 36 / 36 / 304 / 343 / 11 / 0 | 34 / 34 / 306 / 343 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_map_set_setops.a | 590 / 590 / 347 / 666 / 17 / 0 | 590 / 590 / 336 / 655 / 17 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp.a | 455 / 455 / 369 / 422 / 62 / 0 | 437 / 437 / 382 / 417 / 58 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/sweeps/regexp_methods.a | 448046 / 448046 / 131098 / 340026 / 68 / 0 | 345663 / 345663 / 228825 / 335370 / 54 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp_replace.a | 652 / 652 / 29 / 336 / 119 / 0 | 625 / 625 / 206 / 336 / 113 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp_split.a | 110 / 110 / 140 / 154 / 13 / 0 | 103 / 103 / 147 / 154 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp_exec.a | 179 / 179 / 82 / 160 / 18 / 0 | 163 / 163 / 93 / 155 / 17 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp_match.a | 83 / 83 / 85 / 103 / 18 / 0 | 83 / 83 / 84 / 102 / 18 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_static.a | 110 / 110 / 84 / 191 / 28 / 0 | 110 / 110 / 55 / 162 / 28 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_static_private.a | 42 / 42 / 71 / 112 / 12 / 0 | 42 / 42 / 53 / 94 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_private.a | 47 / 47 / 37 / 65 / 15 / 0 | 47 / 47 / 29 / 57 / 15 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_accessors.a | 67 / 67 / 53 / 110 / 21 / 0 | 67 / 67 / 47 / 104 / 21 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_twice.a | 26 / 26 / 6 / 35 / 6 / 0 | 26 / 26 / 4 / 33 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_retained.a | 48 / 48 / 57 / 88 / 18 / 0 | 48 / 48 / 42 / 73 / 18 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_features_distinct.a | 48 / 48 / 50 / 81 / 14 / 0 | 48 / 48 / 42 / 73 / 14 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_inheritance.a | 150 / 150 / 29 / 170 / 23 / 0 | 150 / 150 / 24 / 165 / 23 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_inheritance_exceptions.a | 29 / 29 / 22 / 42 / 8 / 0 | 29 / 29 / 19 / 39 / 8 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_inheritance_memory.a | 48 / 42 / 22 / 60 / 11 / 6 | 48 / 42 / 20 / 58 / 11 / 6 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_inheritance_generic.a | 89 / 89 / 104 / 167 / 39 / 0 | 89 / 89 / 85 / 148 / 39 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_inheritance_interface.a | 37 / 37 / 45 / 71 / 13 / 0 | 37 / 37 / 44 / 70 / 13 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/class_inheritance_conditional.a | 164 / 164 / 148 / 251 / 35 / 0 | 164 / 164 / 146 / 249 / 35 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/devirtualize.a | 43 / 39 / 26 / 66 / 10 / 4 | 43 / 39 / 22 / 62 / 10 / 4 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/user_iterators.a | 669 / 669 / 466 / 915 / 67 / 0 | 669 / 669 / 463 / 912 / 67 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/library_map_set_iterator_number_hash.a | 60 / 60 / 1040 / 1092 / 14 / 0 | 60 / 60 / 1039 / 1091 / 14 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/literal_optional_shapes.a | 36 / 36 / 40 / 68 / 9 / 0 | 36 / 36 / 32 / 65 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/e4eec87_u03_discriminated_undefined.a | 10 / 10 / 14 / 20 / 6 / 0 | 10 / 10 / 12 / 18 / 6 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/e4eec87_f1_field_narrowed.a | 1 / 0 / 2 / 3 / 1 / 0 | 1 / 0 / 1 / 2 / 1 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/e4eec87_f1_field_present.a | 6 / 6 / 2 / 9 / 3 / 0 | 6 / 6 / 1 / 8 / 3 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/047cb0d_n_param.a | 5 / 5 / 4 / 10 / 4 / 0 | 5 / 5 / 2 / 8 / 4 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/nested_array.a | 17 / 17 / 18 / 28 / 9 / 0 | 17 / 17 / 17 / 27 / 9 / 0 | Ordinary async lowering delta above |
| internal/oracle/testdata/nested_destructured.a | 13 / 13 / 8 / 18 / 6 / 0 | 13 / 13 / 7 / 17 / 6 / 0 | Ordinary async lowering delta above |
| internal/oracle/testdata/nested_mixed.a | 7 / 7 / 10 / 13 / 5 / 0 | 7 / 7 / 9 / 12 / 5 / 0 | Ordinary async lowering delta above |
| internal/oracle/testdata/override_same_representation.a | 15 / 15 / 8 / 22 / 9 / 0 | 15 / 15 / 7 / 21 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/concurrency/accepted/fresh.a | 20 / 20 / 11 / 26 / 12 / 0 | 20 / 20 / 10 / 25 / 12 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/concurrency/accepted/identity.a | 8 / 8 / 14 / 20 / 5 / 0 | 8 / 8 / 13 / 19 / 5 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/concurrency/accepted/large.a | 4105 / 4105 / 6155 / 6165 / 4102 / 0 | 4105 / 4105 / 6154 / 6164 / 4102 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/concurrency/accepted/recursive_tree.a | 10 / 10 / 11 / 17 / 10 / 0 | 10 / 10 / 9 / 15 / 10 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| bench/parallel_files.a | 3952651 / 3952651 / 8151046 / 8654931 / 8839 / 0 | 3952651 / 3952651 / 8151045 / 8654930 / 8839 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/object_prototype.a | 142 / 142 / 317 / 440 / 22 / 0 | 142 / 142 / 165 / 288 / 22 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp_cycle_fields.a | 61 / 61 / 134 / 132 / 44 / 0 | 61 / 61 / 132 / 130 / 44 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/regexp_cycle_collections.a | 31 / 31 / 86 / 90 / 23 / 0 | 31 / 31 / 85 / 89 / 23 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/read_files.a | 59 / 59 / 30 / 76 / 8 / 0 | 58 / 58 / 33 / 76 / 7 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/utf8_sweep.a | 38498 / 38498 / 7702 / 38502 / 7706 / 0 | 38498 / 38498 / 7701 / 38501 / 7706 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/arguments.a | 77 / 77 / 15 / 69 / 27 / 0 | 77 / 77 / 14 / 68 / 27 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/read_arguments.a | 15 / 15 / 12 / 19 / 9 / 0 | 15 / 15 / 11 / 18 / 9 / 0 | Imported area baseline; merged counts equal 915b9e05 |
| internal/oracle/testdata/walk.a | 222 / 222 / 123 / 294 / 36 / 0 | 222 / 222 / 117 / 288 / 36 / 0 | Imported area baseline; merged counts equal 915b9e05 |

### The parameter fixture's 17 to 18 retain

This is a short-slice owner retain, not a newly retained async parameter.
The original area compiler already retained its string parameter. In isolated
area 915b9e05 and merged compiler probes, changing only string_share.c to copy
proper slices shorter than 64 bytes produces these counts:

| Compiler | Real short views | Forced short copies |
|---|---|---|
| Area unit 1 | 17 / 17 / 18 / 22 / 12 / 0 | 17 / 17 / 17 / 22 / 12 / 0 |
| Merged unit 2 | 17 / 17 / 24 / 46 / 12 / 0 | 17 / 17 / 23 / 46 / 12 / 0 |

Both stdout and stderr excluding the count line agree byte for byte in each
pair. The fixture's held.slice(1) now reads its parent's bytes and retains that
owner; the previous minimum-slice policy copied them. Its destruction releases
the owner through the internal child-drop path, which is not a counted call to
adamic_release, explaining why the retain changes without a release-count
change. This corrects the earlier unverified parameter-ownership inference.
Unit 2 does not remove this owner retain. Probe log:
/tmp/async-area-parameter-causality.log; source and binaries:
/tmp/async-area-parameter-unit1 and /tmp/async-area-parameter-current.
The forced-copy runtime archives are isolated artifacts, not production edits.

The 153 rows added relative to async-ordinary 0bb0fab are imported area fixtures. Synchronous tuples equal the area baseline; the imported async fixtures have the ordinary-lowering deltas listed above:

| Fixture | Counts | Cause |
|---|---|---|
| internal/oracle/testdata/release_fma.a | 4 / 4 / 10 / 15 / 2 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_lifetime.a | 19 / 19 / 21 / 34 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_holders.a | 45 / 45 / 72 / 79 / 13 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_throw.a | 55 / 55 / 62 / 76 / 14 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_loops.a | 159 / 159 / 45 / 167 / 21 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_policy.a | 344 / 344 / 73 / 371 / 11 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_methods.a | 5253 / 5253 / 2022 / 6546 / 15 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_characters.a | 281 / 281 / 1542 / 1825 / 8 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_calls.a | 85 / 85 / 61 / 116 / 11 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_views_surrogates.a | 188 / 188 / 83 / 255 / 13 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_null.a | 4 / 4 / 1 / 9 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_callbacks.a | 35 / 35 / 43 / 65 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_virtual_fresh.a | 54 / 50 / 20 / 64 / 10 / 4 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_unknown.a | 36 / 36 / 18 / 49 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_bound_method.a | 37 / 37 / 27 / 52 / 9 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_recursive.a | 19 / 18 / 6 / 23 / 7 / 1 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_try_loop.a | 68 / 66 / 41 / 85 / 9 / 2 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_accessor.a | 14 / 14 / 13 / 22 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/route_targets_sort_values.a | 84 / 84 / 92 / 142 / 17 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_named_keys.a | 92 / 92 / 60 / 144 / 15 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_named_keys.a | 84 / 84 / 68 / 144 / 15 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_keys.a | 391 / 391 / 363 / 674 / 25 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_keys.a | 285 / 285 / 242 / 477 / 16 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_keep.a | 516 / 516 / 625 / 931 / 35 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_keep.a | 290 / 290 / 385 / 520 / 18 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_objects.a | 732 / 732 / 834 / 1076 / 63 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_objects.a | 368 / 368 / 515 / 586 / 30 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_closures.a | 816 / 816 / 1044 / 1277 / 87 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_closures.a | 411 / 411 / 634 / 705 / 39 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_numbers.a | 373 / 373 / 155 / 478 / 14 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_numbers.a | 214 / 214 / 140 / 317 / 11 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_arrays.a | 995 / 995 / 932 / 1355 / 95 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_arrays.a | 516 / 516 / 524 / 713 / 58 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_named_more.a | 395 / 395 / 569 / 819 / 36 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_named_more.a | 179 / 179 / 297 / 378 / 26 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_named_objects.a | 165 / 165 / 207 / 278 / 31 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_named_objects.a | 87 / 87 / 110 / 149 / 22 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_named_closures.a | 117 / 117 / 179 / 219 / 29 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_named_closures.a | 95 / 95 / 155 / 187 / 28 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_fnexpr.a | 44 / 44 / 64 / 100 / 15 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/set_foreach_fnexpr.a | 32 / 32 / 51 / 76 / 10 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/map_foreach_named_numbers.a | 117 / 117 / 73 / 177 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_guards.a | 24 / 24 / 9 / 21 / 11 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_callbacks.a | 28 / 28 / 14 / 28 / 11 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_exclusions.a | 21 / 21 / 12 / 23 / 10 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_spread.a | 12 / 12 / 3 / 10 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_scope.a | 10 / 10 / 2 / 8 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_depths.a | 7 / 7 / 0 / 4 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_values.a | 16 / 16 / 9 / 18 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_return.a | 16 / 16 / 7 / 13 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_method.a | 9 / 9 / 2 / 7 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_finally.a | 13 / 13 / 9 / 13 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_parents.a | 15 / 15 / 20 / 29 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_coverage_bounded.a | 11 / 11 / 3 / 10 / 9 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_override.a | 15 / 13 / 4 / 12 / 7 / 2 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_unknown.a | 8 / 8 / 3 / 8 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_argument.a | 6 / 6 / 0 / 4 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_walk.a | 16 / 16 / 7 / 16 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_listener.a | 7 / 7 / 3 / 8 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_write.a | 7 / 7 / 1 / 5 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_reassigned.a | 8 / 8 / 1 / 5 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_capture.a | 6 / 6 / 4 / 7 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_chain_store.a | 10 / 10 / 4 / 11 / 8 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/moves/accepted/objects.a | 2056 / 2056 / 6145 / 4104 / 2053 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/async_coverage_unions.a | 27 / 27 / 81 / 120 / 12 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_reject_empty.a | 6 / 4 / 15 / 18 / 5 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_parameters.a | 17 / 17 / 24 / 46 / 12 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_typeof.a | 16 / 16 / 28 / 49 / 10 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_values.a | 66 / 66 / 157 / 238 / 19 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_discard.a | 25 / 25 / 55 / 83 / 15 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_reject_eager.a | 12 / 9 / 27 / 36 / 10 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_coverage_reject_nested.a | 18 / 15 / 45 / 64 / 17 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/async_typeof.a | 16 / 16 / 29 / 50 / 10 / 0 | Imported async fixture executed through ordinary snapshots and owned frame slots; delta above |
| internal/oracle/testdata/memory_examples/list.a | 7 / 7 / 6 / 16 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/memory_examples/tree.a | 12 / 12 / 28 / 31 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/memory_examples/closures.a | 9 / 9 / 2 / 9 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/memory_examples/regions.a | 14 / 8 / 0 / 5 / 2 / 6 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/memory_examples/strings.a | 14 / 14 / 2 / 17 / 5 / 0 | Imported area fixture and measured area baseline |
| cmd/adamic/testdata/wasi/request.a | 4 / 4 / 5 / 7 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_caches.a | 26 / 26 / 2 / 26 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_padding.a | 759 / 759 / 126 / 877 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_join.a | 161 / 161 / 81 / 193 / 10 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_repeat.a | 240 / 240 / 32 / 264 / 8 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_cached_reads.a | 2684 / 2684 / 333 / 2974 / 9 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_boundaries.a | 2229 / 2229 / 342 / 2347 / 11 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/string_build_calls.a | 47 / 47 / 4 / 52 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/runtime_last_index_of.a | 679 / 679 / 120 / 730 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/lint_runtime_release_chain.a | 405034 / 405034 / 505776 / 303790 / 300003 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/lint_runtime_release_shared.a | 1008 / 1008 / 8478 / 6396 / 263 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/lint_runtime_search_boundaries.a | 568 / 568 / 161 / 647 / 18 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/lint_runtime_search_calls.a | 48 / 48 / 25 / 68 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/lint_runtime_equal_headers.a | 100 / 100 / 22 / 113 / 8 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/pad_infinity.a | 0 / 0 / 0 / 0 / 0 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/repeat_huge.a | 0 / 0 / 0 / 0 / 0 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_long.a | 10 / 10 / 0 / 11 / 10 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/field_access_paths.a | 47 / 47 / 14 / 65 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_field_values.a | 22 / 22 / 33 / 45 / 10 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_slot_kinds.a | 30 / 30 / 11 / 37 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_runtime_fields.a | 8 / 8 / 8 / 15 / 4 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_narrowed_receiver.a | 13 / 13 / 5 / 16 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_base_writes.a | 23 / 23 / 6 / 30 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_collection_fields.a | 39 / 39 / 35 / 61 / 17 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_write_read_order.a | 15 / 15 / 10 / 24 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_static_collision.a | 43 / 43 / 17 / 57 / 19 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_optional_references.a | 34 / 34 / 52 / 67 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/nbody_spread_fields.a | 23 / 23 / 9 / 32 / 13 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/field_write_paths.a | 46 / 46 / 11 / 60 / 8 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/inherited_static_field_read.a | 6 / 6 / 4 / 13 / 4 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/region_end.a | 116 / 19 / 99 / 22 / 65 / 97 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/reuse_comparator_spread.a | 20 / 20 / 13 / 32 / 9 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/reuse_move_before_ready.a | 0 / 0 / 0 / 0 / 0 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/regions_escapes.a | 78 / 78 / 27 / 75 / 38 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/regions_spread_fresh.a | 23 / 23 / 9 / 22 / 10 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/regions_big.a | 20 / 14 / 0 / 14 / 5 / 6 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/regions_paths.a | 453588 / 383864 / 122122 / 383714 / 278521 / 69724 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/reduce_undefined_initial.a | 8 / 8 / 12 / 19 / 7 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_loop.a | 128 / 126 / 110 / 181 / 14 / 2 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_loop_calls.a | 263 / 261 / 176 / 342 / 25 / 2 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/borrow_global_call.a | 18 / 16 / 7 / 20 / 6 / 2 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_coverage_quick.a | 80082 / 80082 / 9324 / 80505 / 22 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_coverage_repeat.a | 97540 / 97540 / 23496 / 99162 / 19 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_coverage_stream.a | 522657 / 522657 / 27071 / 523649 / 19 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_coverage_long.a | 17704541 / 17704541 / 180 / 17704697 / 16 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_coverage_cache.a | 171614 / 171614 / 3330 / 174233 / 14 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/normalize_coverage_limit.a | 3 / 2 / 0 / 3 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/long_field_name.a | 8 / 8 / 4 / 14 / 4 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/reuse_handover_throw.a | 44 / 44 / 21 / 57 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/aside_throw.a | 23 / 23 / 22 / 41 / 6 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/uncaught_names.a | 3 / 1 / 5 / 4 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/uncaught_names_empty.a | 1 / 0 / 5 / 2 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/coverage_plain_view.a | 6 / 6 / 10 / 10 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/coverage_uncaught_default_empty.a | 1 / 0 / 3 / 1 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/coverage_uncaught_custom_empty.a | 1 / 0 / 6 / 3 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/coverage_uncaught_custom_message.a | 5 / 2 / 4 / 5 / 4 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/coverage_uncaught_empty_changed.a | 3 / 1 / 5 / 4 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/coverage_uncaught_no_argument.a | 1 / 0 / 3 / 1 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/output_edges/fsize.a | 8 / 8 / 3 / 12 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/output_edges/fsize_out.a | 60000 / 60000 / 0 / 60000 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/output_edges/closed.a | 0 / 0 / 0 / 0 / 0 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/output_edges/panic_surrogate.a | 3 / 1 / 0 / 2 / 2 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/output_edges/usr1.a | 2 / 2 / 0 / 2 / 2 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/proven_guards.a | 42 / 42 / 40 / 69 / 12 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/proven_class_guards.a | 5 / 3 / 3 / 5 / 3 / 2 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/proven_assertions.a | 11 / 11 / 12 / 14 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/proven_satisfies.a | 28 / 28 / 30 / 57 / 13 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/proven_upcasts.a | 15 / 15 / 23 / 43 / 8 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_dispatch.a | 9 / 9 / 6 / 16 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_string_literal.a | 1 / 1 / 0 / 1 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_null_compare.a | 17 / 17 / 10 / 21 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_null_switch.a | 4 / 4 / 11 / 11 / 3 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_null_slots.a | 14 / 14 / 26 / 36 / 5 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/typeof_null_roll.a | 1 / 1 / 0 / 1 / 1 / 0 | Imported area fixture and measured area baseline |
| internal/oracle/testdata/empty-path.a | 4 / 4 / 4 / 8 / 4 / 0 | Imported area fixture and measured area baseline |
