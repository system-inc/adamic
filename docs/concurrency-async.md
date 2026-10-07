# Concurrency, part 3: async and await

Status: **proposal and feasibility evidence only, awaiting Kirk and Ahra's approval**. Task p286ycm. Branch `codex/concurrency-async`, based on `origin/codex/concurrency` at `ced5530e68357ca9880f4366870d1f596f453d1a`. Nothing in this document changes the approved language or part 1 contract. No compiler or production runtime file changes here.

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

For broader support, use libuv (prefer the Node version) for timer/file/socket phases, plus the Promise microtask layer. Libuv alone does not implement V8 Promise jobs, nextTick, unhandled rejection reporting or module checkpoints. Supporting setImmediate, nextTick and ESM top-level await each requires dedicated Node fixtures. Until then, those APIs are NotYet. Even libuv does not make races between independent external I/O completions reproducible across two separately executed programs. Differential fixtures must establish causal order, or compare a controlled completion trace replayed in both runtimes. Arbitrary timer-versus-network race traces cannot honestly be promised byte for byte across independent runs. Never use this limitation to reorder causally ordered callbacks.

The prototype starts the file read only after the timer has resolved, so its trace has no I/O/timer race. It uses a real monotonic timer, a blocking file reader on one helper pthread, a pipe completion and a join. Its scheduler handles one timer and one file request, not Node's complete phase model. Production libuv integration remains proposed and unmeasured.

## Promise representation, counts and cycles

Proposed ABI: monomorphized `Promise<T>` holds the usual heap header, pending/fulfilled/rejected tag, a typed owned payload (`T` or Error), and an ordered list of reaction tickets. Each ticket owns its callback/captures or frame, destination Promise where chaining needs one, and the payload snapshot after settlement. A Promise may be observed repeatedly; it is not a single-consumer future. Settling is once-only. It retains or takes the payload, detaches the reaction list, and queues jobs in registration order. Later observers enqueue new jobs; none run inline. Destruction releases the payload and any inert metadata. Scalar payloads need no retain; reference payloads do. Initial Promise state and reaction identity stay loop-local, so counts are plain; immutable results originating in part 1 already carry its shared tag and retain/release dispatch.

A producer/request owns the pending Promise until completion. A registered scope owns producer and subscription tickets. A ticket owns the suspended frame; the frame owns its locals and output Promise. The await edge is a ticket/ID, not an additional strong frame-to-awaited-Promise back edge. On resumption the ticket transfers its frame count to the queued job, then unregisters; the job keeps frame and payload alive until resume returns. The frame may retain the payload into a local that outlives that job. Once the last job/scope reference leaves, frame destruction releases its remaining slots. This is the ownership pattern exercised by the prototype.

This acyclic machinery is necessary, not sufficient: user locals can hold Promises, a Promise can fulfill with an object pointing back to it, and resolver closures can capture the frame whose output they own. Extend memory.md's cycle-capable-slot finder and fresh-write proofs to Promise payloads, reaction captures and async frame edges. A mutable cell capturing its own task still needs Weak or refusal. Do not introduce a cycle collector. Internal ticket lists must support unlinking pending reactions on termination; draining reachable jobs and deleting scope registrations cannot depend on the Promise destructor breaking a strong cycle.

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

Files: [prototype.c](../prototypes/async/prototype.c), [oracle.a](../prototypes/async/oracle.a), [check.py](../prototypes/async/check.py), and [README.md](../prototypes/async/README.md). Handwritten C, not generated lowering. The oracle is a Node-only `.a` program copied verbatim into a scratch `.cjs` file; Adamic intentionally still refuses its async syntax. Existing tsconfig does not include prototypes. No new `.ts` files, compiler registrations or production APIs.

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

No production package changed. The standalone prototype checks and filtered existing concurrency gates were run instead of the complete `go test ./...` gate. These existing checks exercise part 1 on the base; they are not evidence of compiler async support. Repository gofmt and staged diff whitespace checks were clean.

## Units needed to land it

Estimates below are reviewable implementation units, not elapsed-time promises. One unit means a separately reviewed commit series with its own controls, refusal fixtures and targeted mutants. All units remain conditional on approval. Basic async I/O needs units 1 through 7 (roughly 10 worker-sized units after splitting the larger rows); networking, structured cancellation and pool adapters are additional work.

| Piece | Estimated units | Deliverable and exit evidence | Depends on |
| --- | ---: | --- | --- |
| 1. Approved semantic subset and library signatures | 1 | Exact supported Promise types/APIs, root ownership and exit rules, cancellation spelling, error policy, module context and explicit NotYet/Refused split. Node probes establish eager prefix, plain/fulfilled await and job order. | Kirk and Ahra |
| 2. Promise heap objects and reaction tickets | 1 to 2 | Typed payload destruction, multiple observers, once-only settlement, chaining and root registrations. ASan/LSan/count mutants; user-cycle refusal extension. Thenable adoption is a separate unit if required. | 1 |
| 3. Host loop and timer adapter | 1 to 2 | libuv dependency/build policy, monotonic timers, microtask checkpoints, loop liveness and shutdown. Node fixture for timer resolution followed by microtask; omitted-checkpoint and wrong-phase mutants. | 1, 2 |
| 4. Suspension-aware IR/control flow | 2 | Await nodes and effect summaries, expression normalization, live-across-await spills, loop/back-edge states, state-specific ownership/initialization masks, handlers/finally completions. Trace checks on Node; wrong-resume-state and missing-live-slot mutants. Disable unsafe borrowing/regions/reuse first. | 1 |
| 5. C emission and JavaScript backend | 2 | Stable frames, eager entry, queued resume, returns/rejections, synchronous callee exception-word isolation. JS emits native async/await with identical inserted checks. Three-way oracle including loops, closures, nested awaits, partially evaluated expressions and awaiting finally; missing-retain/release mutants. | 2, 3, 4 |
| 6. Async file service | 1 | Bounded requests, complete reads/UTF-8/error mapping, descriptor close, request buffer lifetime, drain at exit. Node fixtures for empty, long, missing and permission-denied files and controlled completion order; leak/UAF mutants. | 2, 3, 5 |
| 7. Promise combinators and integration gate | 1 | all input ordering, empty input, first observed rejection, live siblings, chaining cleanup, full uncached oracle and counts. Wrong-index, wait-all-before-reject and sibling-drop mutants. | 2, 5, 6 |
| 8. Explicit async scopes and cancellation | 2 | Parent/child obligation analysis, refused detachment, cooperative token library, unlink/generation protocol, finally that awaits, late I/O acknowledgement. Node semantics supplied by the same scope library. Late-completion UAF, skipped-finally and dropped-child mutants. | 1 through 7 |
| 9. Network I/O | 2+ | Socket/DNS/connect/read/write/close ownership, backpressure, bounded buffers, cancellation and errors. Local deterministic server/trace oracle; sanitizer and fd-leak controls. TLS and full fetch/HTTP API are separately scoped units. | 3, 5, 8 |
| 10. Pure-work pool adapter | 1 to 2 | Nonblocking pool submission/completion, part 1 graph marking and error protocol, explicit Node witness, one-worker/nested completion controls and TSan. No Promise.all rewrite. | 7, 8 and part 1 ABI |
| 11. Broad promise-shaped compatibility | 2+ | User executors/resolver escape, thenables, non-Error rejection representation, finally assimilation, never-settling APIs and complete lifecycle proofs. Only remove NotYet after matching source behavior. nextTick/setImmediate/ESM and unhandled-rejection behavior each need their own scope and phase fixtures. | 1 through 8 |

A first implementation can stage straight-line async lowering before loops and awaiting finally, provided all excluded constructs diagnose NotYet rather than compiling with partial semantics. It cannot announce general async support while those edges silently lose ownership. Networking and the pure-work adapter are independent additions after the foundations, not requirements to validate this prototype.

## Limits of this evidence

No compiler lowering, full production Promise implementation, Node event-loop phase emulation, async scopes, pending I/O cancellation, thenable adoption, Promise.all implementation, sockets, fd-leak tests, scalability or performance benchmark. The C runtime is fixed to one timer and one small file request (ASCII fixture, 127-byte buffer); it is not a general file API. No source-level Pin, stackful comparison benchmark, macOS/Windows run or TSan proof. The helper thread joins before the loop reads its buffer; this is an ownership design, not measured pool integration. Repeated full async correctness remains work for the landing units above.
