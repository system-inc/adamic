# PollingInterval measurement (baobab)

Base: `464e5daa23a60588f5aaa656e95d93af05d582c3`. Upstream pin: TypeScript 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. No adaptation changed.

The compiler-API census uses the locked 6.0.3 checker over every .ts/.tsx file under src,
resolves imported symbols, distinguishes enum objects from enum-valued numbers, and
propagates object identity through casts, variable bindings, assignments and resolved
call parameters to a fixed point. The non-member object-use inventory is retained for
manual escape review. This is a closed-source census, not a proof about external clients
or arbitrary reflective JavaScript. Enum declaration initialization is not counted as
an application write.

| Tree | Source files | Value sites | (a) | (b) | (c) | Enum-write sites |
|---|---:|---:|---:|---:|---:|---:|
| Pin | 709 | 51 | 36 | 15 | 0 | 1 |
| Adapted | 712 | 51 | 36 | 15 | 0 | 1 |

## Writes and the runtime path

The single enum-writing expression is `levels[level] = customLevels![level] || levels[level]`
in nested `setLevel`. `setCustomPollingValues` passes the actual `PollingInterval` object
as the second argument of `setCustomLevels(baseVariable, levels: Levels)`; `setLevel`
captures that writable view. Its Low/Medium/High calls write all three properties
whenever any nonempty `TSC_WATCH_POLLINGINTERVAL_{LOW,MEDIUM,HIGH}` variable is supplied.
The right side preserves an old value for zero or NaN; other truthy numbers are accepted.

`pollingChunkSize` and `unchangedPollThresholds` are plain numeric-keyed objects,
not enums. `defaultChunkLevels` is a plain `Levels` object `{Low:32,Medium:64,High:256}`;
it is read and spread, never mutated by this path. `setCustomLevels` has just one call.
`getCustomLevels` writes a fresh temporary `Partial<Levels>` for each of the three
variable families. `getCustomPollingBasedLevels` builds a new table if intervals
changed or its own custom levels exist. The two outer table bindings are assigned on
every configuration call with an environment getter, even when their identity and
contents do not change. They are not additional enum-object writes.

- Pin: `src/compiler/sys.ts:171`: `levels[level] = customLevels![level] || levels[level]`; enum `PollingInterval`.
- Adapted: `src/compiler/sys.ts:176`: `levels[level] = customLevels![level] || levels[level]`; enum `PollingInterval`.

No other enum write was found. In particular the `FileIncludeKind as any` escape is
a reverse read in program.ts, and `DiagnosticCategory` is passed to `Object.keys`
in harness/client.ts. Qualified enum names in the retained object inventory are
namespace access components; the CompletionSource import-equals is an alias.

## Reader classification and scheduling trace

Class (a) means the source-level intent requires a configured interval. Queue
construction stores its interval in `queue.pollingInterval`. Queue insertion and
transitions call `scheduleNextPollIfNotAlreadyScheduled`, then `scheduleNextPoll`,
then `host.setTimeout(..., pollingInterval, ...)`. Those transition arguments therefore
belong to (a), even though they also choose queues. Watch file/directory arguments
flow through `createSystemWatchFunctions`, `pollingWatchFile` and ultimately the host
polling worker or dynamic queue timer. Class (b) means a symbolic level: table keys,
threshold lookups, comparisons and switch labels. These stay symbolic, but a future
adaptation must make the boundary between symbolic queue identity and timer delay
consistent. Merely changing a few timer constants cannot establish that invariant.

The 3 keys in `createPollingIntervalBasedLevels` label the numeric-keyed chunk/threshold
tables. `pollPollingIntervalQueue` reads `pollingChunkSize[queue.pollingInterval]`;
`pollQueue` reads `unchangedPollThresholds[pollingInterval]`. Fixed-chunk polling reads
the low chunk size but schedules at High. Chunk and threshold customization is therefore
read indirectly as table contents, separately from the enum member sites below.

No direct PollingInterval member reads occur in watch.ts, watchUtilities.ts, or
server/types.ts. watchUtilities.ts's watch factory types and callbacks forward the
interval supplied by their callers. Every observed production reader serves watching; no unrelated compiler/checker consumer was found.
Outside compiler watch internals, (a) has four
sites in tsbuildPublic.ts and ten in server/project.ts and server/editorServices.ts.

All sites, including every (a) reader and its enclosing function:

| Class | Pin location | Adapted location | Function | Expression |
|---|---|---|---|---|
| b | src/compiler/sys.ts:122 | src/compiler/sys.ts:127 | `createPollingIntervalBasedLevels` | `PollingInterval.Low` |
| b | src/compiler/sys.ts:123 | src/compiler/sys.ts:128 | `createPollingIntervalBasedLevels` | `PollingInterval.Medium` |
| b | src/compiler/sys.ts:124 | src/compiler/sys.ts:129 | `createPollingIntervalBasedLevels` | `PollingInterval.High` |
| a | src/compiler/sys.ts:256 | src/compiler/sys.ts:261 | `createDynamicPriorityPollingWatchFile` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:257 | src/compiler/sys.ts:262 | `createDynamicPriorityPollingWatchFile` | `PollingInterval.Medium` |
| a | src/compiler/sys.ts:258 | src/compiler/sys.ts:263 | `createDynamicPriorityPollingWatchFile` | `PollingInterval.High` |
| b | src/compiler/sys.ts:303 | src/compiler/sys.ts:308 | `pollLowPollingIntervalQueue` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:310 | src/compiler/sys.ts:315 | `pollLowPollingIntervalQueue` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:339 | src/compiler/sys.ts:344 | `onWatchFileStat` | `PollingInterval.Low` |
| b | src/compiler/sys.ts:341 | src/compiler/sys.ts:346 | `onWatchFileStat` | `PollingInterval.High` |
| b | src/compiler/sys.ts:344 | src/compiler/sys.ts:349 | `onWatchFileStat` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:344 | src/compiler/sys.ts:349 | `onWatchFileStat` | `PollingInterval.Medium` |
| a | src/compiler/sys.ts:344 | src/compiler/sys.ts:349 | `onWatchFileStat` | `PollingInterval.High` |
| b | src/compiler/sys.ts:351 | src/compiler/sys.ts:356 | `pollingIntervalQueue` | `PollingInterval.Low` |
| b | src/compiler/sys.ts:353 | src/compiler/sys.ts:358 | `pollingIntervalQueue` | `PollingInterval.Medium` |
| b | src/compiler/sys.ts:355 | src/compiler/sys.ts:360 | `pollingIntervalQueue` | `PollingInterval.High` |
| a | src/compiler/sys.ts:367 | src/compiler/sys.ts:372 | `addChangedFileToLowPollingIntervalQueue` | `PollingInterval.Low` |
| b | src/compiler/sys.ts:377 | src/compiler/sys.ts:385 | `<anonymous>` | `PollingInterval.Low` |
| b | src/compiler/sys.ts:377 | src/compiler/sys.ts:387 | `<anonymous>` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:448 | src/compiler/sys.ts:461 | `createDirectoryWatcher` | `PollingInterval.Medium` |
| b | src/compiler/sys.ts:484 | src/compiler/sys.ts:497 | `pollQueue` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:490 | src/compiler/sys.ts:507 | `<anonymous>` | `PollingInterval.High` |
| a | src/compiler/sys.ts:1027 | src/compiler/sys.ts:1052 | `watchFile` | `PollingInterval.Low` |
| a | src/compiler/sys.ts:1109 | src/compiler/sys.ts:1134 | `watchDirectory` | `PollingInterval.Medium` |
| a | src/compiler/sys.ts:1138 | src/compiler/sys.ts:1163 | `nonRecursiveWatchDirectory` | `PollingInterval.Medium` |
| a | src/compiler/sys.ts:1145 | src/compiler/sys.ts:1170 | `nonRecursiveWatchDirectory` | `PollingInterval.Medium` |
| a | src/compiler/sys.ts:1161 | src/compiler/sys.ts:1186 | `nonRecursiveWatchDirectory` | `PollingInterval.Medium` |
| a | src/compiler/watchPublic.ts:490 | src/compiler/watchPublic.ts:489 | `createWatchProgram` | `PollingInterval.High` |
| a | src/compiler/watchPublic.ts:509 | src/compiler/watchPublic.ts:508 | `<anonymous>` | `PollingInterval.High` |
| a | src/compiler/watchPublic.ts:775 | src/compiler/watchPublic.ts:774 | `getVersionedSourceFileByPath` | `PollingInterval.Low` |
| a | src/compiler/watchPublic.ts:788 | src/compiler/watchPublic.ts:787 | `getVersionedSourceFileByPath` | `PollingInterval.Low` |
| a | src/compiler/watchPublic.ts:1083 | src/compiler/watchPublic.ts:1096 | `watchMissingFilePath` | `PollingInterval.Medium` |
| a | src/compiler/watchPublic.ts:1187 | src/compiler/watchPublic.ts:1200 | `<anonymous>` | `PollingInterval.High` |
| a | src/compiler/watchPublic.ts:1206 | src/compiler/watchPublic.ts:1219 | `watchReferencedProject` | `PollingInterval.High` |
| a | src/compiler/tsbuildPublic.ts:2120 | src/compiler/tsbuildPublic.ts:2128 | `watchConfigFile` | `PollingInterval.High` |
| a | src/compiler/tsbuildPublic.ts:2138 | src/compiler/tsbuildPublic.ts:2146 | `<anonymous>` | `PollingInterval.High` |
| a | src/compiler/tsbuildPublic.ts:2191 | src/compiler/tsbuildPublic.ts:2199 | `<anonymous>` | `PollingInterval.Low` |
| a | src/compiler/tsbuildPublic.ts:2212 | src/compiler/tsbuildPublic.ts:2220 | `<anonymous>` | `PollingInterval.High` |
| a | src/server/project.ts:855 | src/server/project.ts:856 | `watchAffectingFileLocation` | `PollingInterval.High` |
| a | src/server/project.ts:1575 | src/server/project.ts:1576 | `createProjectWatcher` | `PollingInterval.High` |
| a | src/server/project.ts:1854 | src/server/project.ts:1855 | `addMissingFileWatcher` | `PollingInterval.Medium` |
| a | src/server/project.ts:1900 | src/server/project.ts:1901 | `createGeneratedFileWatcher` | `PollingInterval.High` |
| a | src/server/editorServices.ts:2433 | src/server/editorServices.ts:2436 | `createConfigFileWatcherForParsedConfig` | `PollingInterval.High` |
| a | src/server/editorServices.ts:2568 | src/server/editorServices.ts:2571 | `<anonymous>` | `PollingInterval.High` |
| a | src/server/editorServices.ts:3054 | src/server/editorServices.ts:3059 | `<anonymous>` | `PollingInterval.High` |
| a | src/server/editorServices.ts:3528 | src/server/editorServices.ts:3533 | `watchClosedScriptInfo` | `PollingInterval.Medium` |
| a | src/server/editorServices.ts:3889 | src/server/editorServices.ts:3894 | `addMissingSourceMapFile` | `PollingInterval.High` |
| a | src/server/editorServices.ts:5661 | src/server/editorServices.ts:5666 | `watchPackageJsonFile` | `PollingInterval.Low` |
| b | src/testRunner/unittests/tscWatch/watchEnvironment.ts:44 | src/testRunner/unittests/tscWatch/watchEnvironment.ts:44 | `<anonymous>` | `ts.PollingInterval.Medium` |
| b | src/testRunner/unittests/tscWatch/watchEnvironment.ts:72 | src/testRunner/unittests/tscWatch/watchEnvironment.ts:72 | `<anonymous>` | `ts.PollingInterval.Medium` |
| b | src/testRunner/unittests/tscWatch/watchEnvironment.ts:73 | src/testRunner/unittests/tscWatch/watchEnvironment.ts:73 | `<anonymous>` | `ts.PollingInterval.Low` |

Class files:

- (a): src/compiler/sys.ts, src/compiler/tsbuildPublic.ts, src/compiler/watchPublic.ts, src/server/editorServices.ts, src/server/project.ts.
- (b): src/compiler/sys.ts, src/testRunner/unittests/tscWatch/watchEnvironment.ts.
- (c): none.

## Runtime measurement

See `oracle-run.json`, `oracle/report.json`, `count.json`, and per-process counter files.
The disposable counted tree wraps source value expressions without changing their
values, and records the three enum property writes plus both table-binding assignments.
A preload supplies the observer to build/test processes and their forked workers.
Counts are source-expression evaluations: esbuild may inline enum members into constants.
`afterAnyWrite` is process-local after either table-binding assignment or enum mutation;
`afterEnumWrite` is after the first actual enum property assignment. Local temporary
bindings and fresh customLevels construction are not configuration-target writes.
These definitions avoid treating the two unconditional binding assignments as evidence
that the lane exercised custom configuration.

Observed configuration calls: 9.

Oracle status: fail; test counts: `{"failing": 1, "passing": 106366, "pending": 0}`.

All observed numeric customization variables were empty. The lane did not exercise custom polling intervals, chunk sizes, or thresholds. This is a coverage gap: the adaptation witness must supply those variables. Tests do set watch-mode variables such as TSC_WATCHFILE and TSC_WATCHDIRECTORY; that does not cover numeric customization.

The first instrumented attempt failed during build because an any-valued observer lost literal types. Its exit and logs are retained under oracle-attempt1. The final observer preserves its input type through a generic identity signature; the final run is the only input to count.json.

Enum property writes: 0; table-binding assignments: 18. Class (a): 21163 evaluations, 21163 after table/enum writes, 0 after enum mutation. 35 of 36 sites executed; sys.ts:1052 did not execute.

| Target write | Count | Assigned values and frequencies |
|---|---:|---|
| write:enum:Low | 0 | none |
| write:enum:Medium | 0 | none |
| write:enum:High | 0 | none |
| write:pollingChunkSize | 9 | `{"{\"250\":32,\"500\":64,\"2000\":256}": 9}` |
| write:unchangedPollThresholds | 9 | `{"{\"250\":32,\"500\":64,\"2000\":256}": 9}` |

| (a) reader (adapted) | Reads | After any target write | After enum write |
|---|---:|---:|---:|
| src/compiler/sys.ts:261 PollingInterval.Low | 3 | 3 | 0 |
| src/compiler/sys.ts:262 PollingInterval.Medium | 3 | 3 | 0 |
| src/compiler/sys.ts:263 PollingInterval.High | 3 | 3 | 0 |
| src/compiler/sys.ts:315 PollingInterval.Low | 33 | 33 | 0 |
| src/compiler/sys.ts:344 PollingInterval.Low | 1 | 1 | 0 |
| src/compiler/sys.ts:349 PollingInterval.Medium | 3 | 3 | 0 |
| src/compiler/sys.ts:349 PollingInterval.High | 2 | 2 | 0 |
| src/compiler/sys.ts:372 PollingInterval.Low | 3 | 3 | 0 |
| src/compiler/sys.ts:461 PollingInterval.Medium | 4 | 4 | 0 |
| src/compiler/sys.ts:507 PollingInterval.High | 8 | 8 | 0 |
| src/compiler/sys.ts:1052 PollingInterval.Low | 0 | 0 | 0 |
| src/compiler/sys.ts:1134 PollingInterval.Medium | 4555 | 4555 | 0 |
| src/compiler/sys.ts:1163 PollingInterval.Medium | 4 | 4 | 0 |
| src/compiler/sys.ts:1170 PollingInterval.Medium | 4 | 4 | 0 |
| src/compiler/sys.ts:1186 PollingInterval.Medium | 230 | 230 | 0 |
| src/compiler/watchPublic.ts:489 PollingInterval.High | 365 | 365 | 0 |
| src/compiler/watchPublic.ts:508 PollingInterval.High | 290 | 290 | 0 |
| src/compiler/watchPublic.ts:774 PollingInterval.Low | 6 | 6 | 0 |
| src/compiler/watchPublic.ts:787 PollingInterval.Low | 1600 | 1600 | 0 |
| src/compiler/watchPublic.ts:1096 PollingInterval.Medium | 15 | 15 | 0 |
| src/compiler/watchPublic.ts:1200 PollingInterval.High | 10 | 10 | 0 |
| src/compiler/watchPublic.ts:1219 PollingInterval.High | 43 | 43 | 0 |
| src/compiler/tsbuildPublic.ts:2128 PollingInterval.High | 182 | 182 | 0 |
| src/compiler/tsbuildPublic.ts:2146 PollingInterval.High | 13 | 13 | 0 |
| src/compiler/tsbuildPublic.ts:2199 PollingInterval.Low | 288 | 288 | 0 |
| src/compiler/tsbuildPublic.ts:2220 PollingInterval.High | 98 | 98 | 0 |
| src/server/project.ts:856 PollingInterval.High | 1204 | 1204 | 0 |
| src/server/project.ts:1576 PollingInterval.High | 20 | 20 | 0 |
| src/server/project.ts:1855 PollingInterval.Medium | 47 | 47 | 0 |
| src/server/project.ts:1901 PollingInterval.High | 45 | 45 | 0 |
| src/server/editorServices.ts:2436 PollingInterval.High | 1982 | 1982 | 0 |
| src/server/editorServices.ts:2571 PollingInterval.High | 3449 | 3449 | 0 |
| src/server/editorServices.ts:3059 PollingInterval.High | 26 | 26 | 0 |
| src/server/editorServices.ts:3533 PollingInterval.Medium | 6448 | 6448 | 0 |
| src/server/editorServices.ts:3894 PollingInterval.High | 28 | 28 | 0 |
| src/server/editorServices.ts:5666 PollingInterval.Low | 148 | 148 | 0 |

## Node reference and reproduction

`node-reference.cjs` uses the locked, published TypeScript 6.0.3 bundle, whose `sys`
is initialized by getNodeSystem. It intercepts timer registration and selects dynamic
priority polling with Low=17, Medium=41, High=89. Actual output:

```
mode published bundle
TypeScript 6.0.3
PollingInterval {"250":"Low","500":"Medium","2000":"High","High":89,"Medium":41,"Low":17}
watchFile 250 ok
watchFile 500 ok
watchFile 2000 ok
watchFile 17 TypeError: Cannot read properties of undefined (reading 'push')
watchFile 41 TypeError: Cannot read properties of undefined (reading 'push')
watchFile 89 TypeError: Cannot read properties of undefined (reading 'push')
scheduled [{"callback":"pollLowPollingIntervalQueue","interval":250},{"callback":"pollPollingIntervalQueue","interval":500},{"callback":"pollPollingIntervalQueue","interval":2000}]
```

Direct adapted-source CommonJS output (same getNodeSystem, with compiler-API
transpilation and no bundle enum inlining):

```
mode source CommonJS (transpileModule; no enum inlining)
TypeScript 6.0.3
PollingInterval {"250":"Low","500":"Medium","2000":"High","High":89,"Medium":41,"Low":17}
watchFile 250 TypeError: Cannot read properties of undefined (reading 'push')
watchFile 500 TypeError: Cannot read properties of undefined (reading 'push')
watchFile 2000 TypeError: Cannot read properties of undefined (reading 'push')
watchFile 17 ok
watchFile 41 ok
watchFile 89 ok
scheduled [{"callback":"","interval":17},{"callback":"","interval":41},{"callback":"","interval":89}]
```

The published bundle inlines watch-code enum member uses to 250/500/2000 while retaining
the mutable enum object. Therefore configured object values and internal queue identities
diverge: default numeric inputs schedule the defaults, and configured numeric inputs
fail queue lookup. This reference must not be reported as successful custom intervals.
A future witness needs an explicit custom-environment case and a decision about this
observable difference between source intent and the published bundle.

Reproduce from this base with stage3/api's locked npm install, the pinned upstream tree,
and `ADAMIC_TYPESCRIPT_SOURCE` pointing to it. Build the adapted tree using stage3/apply.sh.
Run `census.cjs TREE` for each tree. Copy the adapted tree, then run
`instrument.cjs COUNTED_TREE adapted-census.json`. Run stage3/oracle/run.sh on that copy
with `NODE_OPTIONS=--require=ABSOLUTE/preload.cjs` and `POLLING_MEASURE_DIR` set to an
absolute output directory. Use `run.py OUTPUT COMMAND...` to bound the whole invocation
at 1700 seconds, retain its log and report the child's exit code. Never instrument the
production adaptation. `summarize.py` aggregates the retained process files.
