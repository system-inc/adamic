# Timer slot census — 995spev, condition 1

Base: adamic main `e3f59bf5509b1feef85b405807a732e11dc383b6`. TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

**17 property calls, 9 with trailing arguments, in each tree.** Of these, **9 calls / 4 tails** have receivers assignable to System (including ServerHost and TestServerHost); **8 calls / 5 tails** use partial System timer-host interfaces. Another **2 calls / 1 tail** invoke destructured System slots. Thus the inclusive slot census is **19 calls / 10 tails per tree**, not 38 distinct sites across the two snapshots.

**13 supplier/implementer sites: 5 concrete host supply sites (4 object/class definitions plus 1 overwrite of an existing host) and 8 forwarding object sites; 24 individual slot supplies. Stop condition: NO.** None supplies its own replacement callback arguments. Fakes never invoke callbacks. Counts refer to source sites, not runtime instances.

## Method and scope

TypeScript compiler API 6.0.3; one Program over every source file under src, using upstream tsconfig-base options, Node/Mocha/Chai type roots, and no emit. Upstream has 710 roots; adapted has 712. All loaded src files are visited. The scanner records receiver types, assignability, inheritance, slot-symbol declarations, callback bodies, every matching property read/write, destructured bindings, bare timer calls, suppliers, spreads, Object.assign and typed assign/extend/clone/copyProperties/Object.create candidates. The saved raw scans retain excluded Mocha/global calls and diagnostics. There are no timer-bearing spreads or Object.assign sites; the typed bulk-copy scan finds exactly two copyProperties calls and no timer-bearing assign/extend/clone/Object.create calls. Classification and pairing assertions fail rather than silently drop new candidates.

WatchHost does not extend System, and the internal polling hosts contain only picked System slots. They are included because createWatchHost/createSystemWatchFunctions forward System timers into them. The strict System-compatible subtotal is separately reported. TestServerHost explicitly implements ServerHost; structural assignability also catches it despite getBaseTypes not representing implements clauses.

The unified programs report 24 upstream and 23 adapted diagnostics (including cross-project/global-library merging and existing inference issues); these are retained, not suppressed. No missing module/type-definition diagnostics remain; no source file containing a timer-slot invocation has a diagnostic. This is a census, not a claim that a unified all-src project is an upstream build configuration.

## Every call

Locations below are upstream → adapted. The receiver column is identical in both independently checked trees (asserted); both raw checker types and both source locations are also in count.json. A `setTimeout` tail begins after callback and ms. clearTimeout has no callback.

| Upstream → adapted | Slot | Receiver | Args | Callback |
|---|---|---|---:|---|
| src/compiler/sys.ts:377 → src/compiler/sys.ts:384 | setTimeout | `{ getModifiedTime: NonNullable<System["getModifiedTime"]>; setTimeout: NonNullable<System["setTimeout"]>; }` | 4 | Conditional choice of two local functions; neither is a method or reads dynamic this. |
| src/compiler/sys.ts:490 → src/compiler/sys.ts:497 | setTimeout | `{ getModifiedTime: NonNullable<System["getModifiedTime"]>; setTimeout: NonNullable<System["setTimeout"]>; }` | 3 | Local pollQueue function; no method/dynamic this. |
| src/compiler/watchPublic.ts:857 → src/compiler/watchPublic.ts:856 | clearTimeout | `WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>` | 1 | — |
| src/compiler/watchPublic.ts:868 → src/compiler/watchPublic.ts:867 | setTimeout | `WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>` | 3 | Local invalidateResolutionsOfFailedLookup function; no method/dynamic this. |
| src/compiler/watchPublic.ts:887 → src/compiler/watchPublic.ts:886 | clearTimeout | `WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>` | 1 | — |
| src/compiler/watchPublic.ts:890 → src/compiler/watchPublic.ts:889 | setTimeout | `WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>` | 3 | Local updateProgramWithWatchStatus function; no method/dynamic this. |
| src/compiler/tsbuildPublic.ts:2066 → src/compiler/tsbuildPublic.ts:2065 | clearTimeout | `SolutionBuilderWithWatchHost<T>` | 1 | — |
| src/compiler/tsbuildPublic.ts:2068 → src/compiler/tsbuildPublic.ts:2067 | setTimeout | `SolutionBuilderWithWatchHost<T>` | 5 | Local buildNextInvalidatedProject function; no method/dynamic this. |
| src/server/utilities.ts:27 → src/server/utilities.ts:27 | clearTimeout | `ServerHost` | 1 | — |
| src/server/utilities.ts:30 → src/server/utilities.ts:30 | setTimeout | `ServerHost` | 5 | ThrottledOperations.run is a static method; reads self, not this. No dynamic this read. |
| src/server/utilities.ts:39 → src/server/utilities.ts:39 | clearTimeout | `ServerHost` | 1 | — |
| src/server/utilities.ts:63 → src/server/utilities.ts:63 | setTimeout | `ServerHost` | 3 | GcTimer.run is a static method; reads self, not this. No dynamic this read. |
| src/server/session.ts:404 → src/server/session.ts:406 | setTimeout | `ServerHost` | 3 | Arrow function reads lexical this; it is not a method and has no dynamic this dependency. |
| src/server/session.ts:450 → src/server/session.ts:452 | clearTimeout | `ServerHost` | 1 | — |
| src/server/typingInstallerAdapter.ts:242 → src/server/typingInstallerAdapter.ts:242 | setTimeout | `server.ServerHost` | 3 | Arrow function reads lexical this; it is not a method and has no dynamic this dependency. |
| src/testRunner/unittests/sys/symlinkWatching.ts:20 → src/testRunner/unittests/sys/symlinkWatching.ts:20 | setTimeout | `ts.System` | 2 | op: callback parameter of Promise wrapper; no method extraction at this site; arbitrary supplied function may itself depend on this. |
| src/testRunner/unittests/tsserver/reloadProjects.ts:86 → src/testRunner/unittests/tsserver/reloadProjects.ts:86 | clearTimeout | `TestServerHost` | 1 | — |
| src/compiler/sys.ts:767 → src/compiler/sys.ts:774 | clearTimeout | `RecursiveDirectoryWatcherHost (destructured parameter)` | 1 | — |
| src/compiler/sys.ts:770 → src/compiler/sys.ts:777 | setTimeout | `RecursiveDirectoryWatcherHost (destructured parameter)` | 3 | Local onTimerToUpdateChildWatches function; no method/dynamic this. |

## Every trailing argument

Value stability distinguishes a stable object reference from mutation of the referenced object. Function calls are inspected rather than automatically declared effectful. The classification is the same in both trees.

### src/compiler/sys.ts:377 → src/compiler/sys.ts:384

Conditional choice of two local functions; neither is a method or reads dynamic this.

- `pollingInterval === PollingInterval.Low ? "pollLowPollingIntervalQueue" : "pollPollingIntervalQueue"` — Evaluation: No writes/calls; reads the mutable PollingInterval.Low enum property. Stability: Yes in principle if PollingInterval.Low changes (setCustomPollingValues mutates the enum); pollingInterval itself is not reassigned.

- `pollingIntervalQueue(pollingInterval)` — Evaluation: Call, but inspected body is a pure switch returning an existing queue; no writes. Stability: Yes in principle: switch reads mutable PollingInterval.Low/Medium/High; same interval normally returns the same queue identity, whose contents are mutable.

### src/compiler/sys.ts:490 → src/compiler/sys.ts:497

Local pollQueue function; no method/dynamic this.

- `"pollQueue"` — Evaluation: None. Stability: No; string literal.

### src/compiler/watchPublic.ts:868 → src/compiler/watchPublic.ts:867

Local invalidateResolutionsOfFailedLookup function; no method/dynamic this.

- `"timerToInvalidateFailedLookupResolutions"` — Evaluation: None. Stability: No; string literal.

### src/compiler/watchPublic.ts:890 → src/compiler/watchPublic.ts:889

Local updateProgramWithWatchStatus function; no method/dynamic this.

- `"timerToUpdateProgram"` — Evaluation: None. Stability: No; string literal.

### src/compiler/tsbuildPublic.ts:2068 → src/compiler/tsbuildPublic.ts:2067

Local buildNextInvalidatedProject function; no method/dynamic this.

- `"timerToBuildInvalidatedProject"` — Evaluation: None. Stability: No; string literal.

- `state` — Evaluation: None; parameter read. Stability: Reference is not reassigned; referenced SolutionBuilderState is mutable and changes between schedule and fire.

- `changeDetected` — Evaluation: None; boolean parameter read. Stability: No; parameter is not reassigned.

### src/server/utilities.ts:30 → src/server/utilities.ts:30

ThrottledOperations.run is a static method; reads self, not this. No dynamic this read.

- `operationId` — Evaluation: None; string parameter read. Stability: No; parameter is not reassigned.

- `this` — Evaluation: None. Stability: Object identity is stable; the ThrottledOperations instance is mutable.

- `cb` — Evaluation: None; callback parameter read. Stability: No; function reference parameter is not reassigned (its closed-over state may change).

### src/server/utilities.ts:63 → src/server/utilities.ts:63

GcTimer.run is a static method; reads self, not this. No dynamic this read.

- `this` — Evaluation: None. Stability: Object identity is stable; the GcTimer instance is mutable.

### src/server/session.ts:404 → src/server/session.ts:406

Arrow function reads lexical this; it is not a method and has no dynamic this dependency.

- `actionType` — Evaluation: None; string parameter read. Stability: No; parameter is not reassigned.

### src/server/typingInstallerAdapter.ts:242 → src/server/typingInstallerAdapter.ts:242

Arrow function reads lexical this; it is not a method and has no dynamic this dependency.

- `` `${request.projectName}::${request.kind}` `` — Evaluation: Reads two properties and interpolates them. With the ordinary typed data object these are plain reads; getters/proxies/coercion could have effects at runtime. Stability: Yes; reads properties of request, a mutable object. Preserve the computed string at schedule time.

### src/compiler/sys.ts:770 → src/compiler/sys.ts:777

Local onTimerToUpdateChildWatches function; no method/dynamic this.

- `"timerToUpdateChildWatches"` — Evaluation: None. Stability: No; string literal.

The PollingInterval enum is writable: sys.ts setCustomPollingValues/setCustomLevels updates its properties. Even where ordinary runs leave it stable, moving those reads into a firing closure is not a generally valid equivalence. The two static method callbacks read their explicit self arguments; neither reads dynamic this. The two arrow callbacks retain lexical this.

## Every implementer and forwarding supplier

Each row counts one object/class/host-initialization source site; set and clear locations are listed individually. The two polling input objects supply setTimeout only. clearTimeout never invokes a callback.

- **ensureDynamicPollingWatchFile input object** (forwarder): setTimeout src/compiler/sys.ts:1054 → src/compiler/sys.ts:1061. Forwards the supplied setTimeout unchanged; eventual provider receives original rest arguments.

- **ensureFixedChunkSizePollingWatchFile input object** (forwarder): setTimeout src/compiler/sys.ts:1058 → src/compiler/sys.ts:1065. Forwards the supplied setTimeout unchanged; eventual provider receives original rest arguments.

- **createRecursiveDirectoryWatcher input object** (forwarder): setTimeout src/compiler/sys.ts:1122 → src/compiler/sys.ts:1129; clearTimeout src/compiler/sys.ts:1123 → src/compiler/sys.ts:1130. Forwards supplied timers unchanged; callback gets original rest arguments from eventual provider.

- **getNodeSystem createSystemWatchFunctions input object** (forwarder): setTimeout src/compiler/sys.ts:1505 → src/compiler/sys.ts:1501; clearTimeout src/compiler/sys.ts:1506 → src/compiler/sys.ts:1502. Passes Node global setTimeout/clearTimeout unchanged; Node invokes callback with original rest arguments.

- **getNodeSystem nodeSystem** (concrete host): setTimeout src/compiler/sys.ts:1603 → src/compiler/sys.ts:1599; clearTimeout src/compiler/sys.ts:1604 → src/compiler/sys.ts:1600. Node globals handed through; callback receives original rest arguments. clearTimeout is Node cancellation.

- **createWatchHost return object** (forwarder): setTimeout src/compiler/watch.ts:679 → src/compiler/watch.ts:678; clearTimeout src/compiler/watch.ts:680 → src/compiler/watch.ts:679. maybeBind binds only this to system (no argument prefix), preserving rest arguments; absent timer uses noop and never invokes callback.

- **SessionServerHost** (concrete host): setTimeout src/harness/harnessLanguageService.ts:593 → src/harness/harnessLanguageService.ts:593; clearTimeout src/harness/harnessLanguageService.ts:597 → src/harness/harnessLanguageService.ts:597. setTimeout empty body: never invokes callback; clearTimeout empty body.

- **TestServerHost createSystemWatchFunctions input object** (forwarder): setTimeout src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468 → src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468; clearTimeout src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469 → src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469. bind(this) binds only receiver; TestServerHost later invokes callback with original rest arguments.

- **TestServerHost** (concrete host): setTimeout src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103 → src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103; clearTimeout src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111 → src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111. Registers callback and args with Callbacks.register; invokeCallback destructures stored args and calls cb(...args). clearTimeout unregisters id.

- **mockHost** (concrete host): setTimeout src/testRunner/unittests/tsserver/session.ts:42 → src/testRunner/unittests/tsserver/session.ts:42; clearTimeout src/testRunner/unittests/tsserver/session.ts:45 → src/testRunner/unittests/tsserver/session.ts:45. setTimeout returns 0 and never invokes callback; clearTimeout is ts.noop.

- **initializeNodeSystem existing sys host** (concrete host): setTimeout src/tsserver/nodeServer.ts:268 → src/tsserver/nodeServer.ts:268; clearTimeout src/tsserver/nodeServer.ts:269 → src/tsserver/nodeServer.ts:269. Assigns Node globals directly; Node invokes callback with original rest arguments; Node clearTimeout cancels.

- **createWatchCompilerHost result** (forwarder): setTimeout src/compiler/watch.ts:874 → src/compiler/watch.ts:873; clearTimeout src/compiler/watch.ts:874 → src/compiler/watch.ts:873. copyProperties copies both timer values unchanged from WatchHost; original callback arguments pass through the previously bound System timers, or noop never invokes callback.

- **createSolutionBuilderWithWatchHost result** (forwarder): setTimeout src/compiler/tsbuildPublic.ts:327 → src/compiler/tsbuildPublic.ts:326; clearTimeout src/compiler/tsbuildPublic.ts:327 → src/compiler/tsbuildPublic.ts:326. copyProperties copies both timer values unchanged from WatchHost; original callback arguments pass through the previously bound System timers, or noop never invokes callback.

TestServerHost execution path (identical in both snapshots):

```ts
setTimeout(callback: TimeOutCallback, ms: number, ...args: any[]): number {
    return this.timeoutCallbacks.register(callback, args, ms);
}
// Callbacks.register stores { cb, args, ms, time }. invokeCallback later does:
const { cb, args, ms, time } = data;
// ... removes the timer and advances simulated time ...
cb(...args);
```

No stop-condition implementer exists, so there is no offending body to quote. All supplying member bodies/expressions are retained verbatim in count.json. The typings installer uses the existing sys/installTypingHost, not a distinct timer-slot implementation. watchApi.ts:190–191 (same adapted lines) removes both slots by assigning undefined; those are writes, not reads or implementations. Mocha Timeout.clearTimeout is unrelated and excluded.

## Slot reads without invocation

| Upstream → adapted | Expression | Use |
|---|---|---|
| src/compiler/watch.ts:679 → src/compiler/watch.ts:678 | `system.setTimeout` | forward/bind |
| src/compiler/watch.ts:680 → src/compiler/watch.ts:679 | `system.clearTimeout` | forward/bind |
| src/compiler/watchPublic.ts:863 → src/compiler/watchPublic.ts:862 | `host.setTimeout` | guard |
| src/compiler/watchPublic.ts:863 → src/compiler/watchPublic.ts:862 | `host.clearTimeout` | guard |
| src/compiler/watchPublic.ts:882 → src/compiler/watchPublic.ts:881 | `host.setTimeout` | guard |
| src/compiler/watchPublic.ts:882 → src/compiler/watchPublic.ts:881 | `host.clearTimeout` | guard |
| src/compiler/tsbuildPublic.ts:2062 → src/compiler/tsbuildPublic.ts:2061 | `hostWithWatch.setTimeout` | guard |
| src/compiler/tsbuildPublic.ts:2062 → src/compiler/tsbuildPublic.ts:2061 | `hostWithWatch.clearTimeout` | guard |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468 → src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468 | `this.setTimeout` | forward/bind |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469 → src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469 | `this.clearTimeout` | forward/bind |
| src/compiler/sys.ts:599 → src/compiler/sys.ts:606 | `setTimeout` | destructure slot from RecursiveDirectoryWatcherHost or CreateSystemWatchFunctions |
| src/compiler/sys.ts:600 → src/compiler/sys.ts:607 | `clearTimeout` | destructure slot from RecursiveDirectoryWatcherHost or CreateSystemWatchFunctions |
| src/compiler/sys.ts:993 → src/compiler/sys.ts:1000 | `setTimeout` | destructure slot from RecursiveDirectoryWatcherHost or CreateSystemWatchFunctions |
| src/compiler/sys.ts:994 → src/compiler/sys.ts:1001 | `clearTimeout` | destructure slot from RecursiveDirectoryWatcherHost or CreateSystemWatchFunctions |
| src/compiler/sys.ts:1054 → src/compiler/sys.ts:1061 | `setTimeout` | forward destructured slot into object |
| src/compiler/sys.ts:1058 → src/compiler/sys.ts:1065 | `setTimeout` | forward destructured slot into object |
| src/compiler/sys.ts:1122 → src/compiler/sys.ts:1129 | `setTimeout` | forward destructured slot into object |
| src/compiler/sys.ts:1123 → src/compiler/sys.ts:1130 | `clearTimeout` | forward destructured slot into object |
| src/compiler/core.ts:1516 → src/compiler/core.ts:1513 | `second[id]` | generic bulk-copy read, instantiated with WatchHost by both listed copyProperties calls |

19 source read sites per tree: 10 named property reads (6 presence guards, 4 passed/bound), 4 destructuring reads, 4 forwarding shorthand reads of those destructured aliases, and the generic second[id] read in copyProperties. Node-global identifier reads in getNodeSystem/nodeServer are provider inputs, not reads of an existing System slot; they are retained under implementers.

The generic bulk-copy read is instantiated for timers at: src/compiler/watch.ts:874 → src/compiler/watch.ts:873 (`copyProperties(result, createWatchHost(system, reportWatchStatus))`); src/compiler/tsbuildPublic.ts:327 → src/compiler/tsbuildPublic.ts:326 (`copyProperties(host, watchHost)`). Each copies both slots without calling either.

## Runtime cross-check

The lane's stage3/oracle/run.sh ran all runners with eight workers on Node v24.19.0, on a disposable instrumented adapted tree. Its install/build/test child exits and durations are retained in count.json. The outer bounded runner also prints its child's actual return code. Measurement instrumentation replaced only the 19 census callees; receiver and callee are obtained before argument evaluation, arguments are evaluated once, and Reflect.apply preserves the original receiver and argument list. No callback is wrapped. Counts are taken at the call site, before forwarding, so internal forwarding is not double-counted. This cross-check is instrumented, not an unmodified lane verdict.

Observed **8,251 total calls**, **5,948 carrying a tail**, and **3,628 distinct tails**. count.json runtime.tails contains every distinct tail, occurrence count and adapted source sites. Primitive values use value equality; objects/functions use process-scoped identity, without reading mutable object contents.

Oracle counts: {'passing': 106366, 'failing': 1, 'pending': 0}; status fail. Baseline differences: ['api/typescript.d.ts']. This exactly matches the lane's expected counts, phase exits and sole API-baseline failure title; the oracle status fail is expected. The cross-check completed in 323.443 seconds with no timeout. See commands.md for reproduction.

| Adapted site | Runtime calls |
|---|---:|
| src/compiler/sys.ts:384 | 267 |
| src/compiler/sys.ts:497 | 8 |
| src/compiler/sys.ts:774 | 54 |
| src/compiler/sys.ts:777 | 83 |
| src/compiler/tsbuildPublic.ts:2065 | 26 |
| src/compiler/tsbuildPublic.ts:2067 | 407 |
| src/compiler/watchPublic.ts:856 | 224 |
| src/compiler/watchPublic.ts:867 | 329 |
| src/compiler/watchPublic.ts:886 | 137 |
| src/compiler/watchPublic.ts:889 | 901 |
| src/server/session.ts:406 | 433 |
| src/server/session.ts:452 | 4 |
| src/server/typingInstallerAdapter.ts:242 | 9 |
| src/server/utilities.ts:27 | 1587 |
| src/server/utilities.ts:30 | 3511 |
| src/server/utilities.ts:39 | 176 |
| src/testRunner/unittests/sys/symlinkWatching.ts:20 | 94 |
| src/testRunner/unittests/tsserver/reloadProjects.ts:86 | 1 |

## Coverage limits

Covers source sites in src (compiler, services, server, harness, testRunner, typingsInstaller, tsserver, jsTyping and all other loaded src files), including type declarations for validation. Does not count external user hosts, dependencies, fixtures under tests/, string-embedded evaluation programs, generated JS/built output as additional source sites, or hypothetical runtime-created/computed unknown property names. The two typed bulk-copy contexts and their generic property read are explicitly covered. Global Node/DOM timer calls and Mocha timeout methods are not System-slot calls. Aliased slot invocations found here are included separately from property-call syntax. No adaptation is changed. Runtime observations cover only executed oracle paths and are not evidence that unexecuted sites cannot receive other values.
