Built: optional captured callbacks use presence checks; synchronous cell/closure cycles use runtime graph regions, and async capture remains refused.
Commits: runtime base f09f8dbeccf9d9e7b2afdc4c56f5a534dc3952e6; merged capture adab0fe54fb8a53940b0c1481e2409b978c6050a in 9c62ffc5072b5b70621ef657f101383b17da1c6a; current main merged in 1571a852.
Commands and outputs: full lower, native graph tests, expanded Node oracle, recorded counts, sequential signals, vet and formatting pass; commands and timings are below.
Mutants: restoring the synchronous seed refusal fails adapted fixture a; omitting environment adoption preserves stdout but LeakSanitizer finds 243 bytes in three allocations.
Limits: Linux only, filtered repository gate; host fixture 14's cache/context shape is covered, not its filesystem imports; the original undefined! form still stops at the pinned assertion refusal.

## Exact base and the edge

I built the exact requested runtime tree separately. Its self-capturing closure compiles already. Its async equivalent is refused with the pinned `async frame capture cycle` message. Both byte-exact adapted memoize probes stop earlier: `a.a:4:13` and `b.a:5:13` say `Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0`. The host-shaped optional callback stops for the same reason. The baseline diagnostics are saved alongside the final traces.

`graphTypes` already seeds a synchronous captured cell, marks its closure, and propagates membership to its frame environment. I did not replace a valid cycle edge or change its graph-seed decision. The missing piece was `condition` accepting an optional object but not an optional closure. Both represent absence as NULL, so an optional closure now uses the same single-evaluation `IsUndefined` presence predicate. Ordinary nonoptional callable truthiness and numeric/string truthiness are unchanged.

The trace for adapted probe a is:

```text
cycle reach: cell callback -- slot contents --> (() => number) | undefined
  (() => number) | undefined -- union or intersection member --> () => number
  () => number -- closure capture at internal/oracle/testdata/memoize_regions/a.a:3:12 --> cell callback
  cell callback -- same captured cell --> cell callback
```

The returned closure has a type permitted in the callback cell and captures that cell. This is a real type-capability path. The supplied memoize invocation stores a leaf callback and then clears it, so it need not have an actual cyclic heap. The self fixture explicitly stores its returned closure into callback: environment -> cell -> closure -> environment is an actual cycle. Runtime's region owns that environment and closure together. Its counted run frees exactly one region, with six allocations and six frees.

The async trace is the same path with the closure at `async.a:2:20`, followed by:

```text
async frame capture cycle: 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free
```

This exact `Refused.What` is pinned. Tracing now also records the async-frame captured-slot edge instead of bypassing the parent recorder; this does not alter graph reach or decisions.

## Observations against Node

Adapted a.a and b.a are byte-exact inputs from `codex/stage3-adapt-memoize` d0e9a37362d02d1401fa5681777bebbd6c2cec0b. Node, the JavaScript backend, sanitized native and release native all print `42 42 1`, proving the callback runs once. The host14 shape prints `cleared`, `true`, `true 1`: immediately after assigning ordinary undefined to the honestly optional callback, `if (callback)` is false, and changing the source directory value does not rerun its cached computation. The real self-cycle prints `cycle:ready` and `freed`, has a dynamically allocated label, and frees as one region. All four successful fixtures are leak-clean under LeakSanitizer. Node prints `done` for the async fixture, which Adamic deliberately refuses.

The original library source uses `undefined!` in a nonoptional callback slot. That assertion remains refused on the pinned compiler lineage; `TestMemoizeCaptureStopsAtAssertion` pins it. This unit uses the supplied honest optional adaptations, so it does not claim to fix the earlier host-proof assertion/readiness behavior. Native ordinary undefined stores NULL and releases the old closure; the false-after-clear observation is executable coverage, not just a generated-C inference.

The old Promise-payload test still expected the pre-region cycle refusal even on the runtime base. Its updated assertion permits refusal or requires graph ownership on acceptance; a generated runtime type name still cannot exempt a user cycle. The dedicated async-frame refusal remains intact.

## Mutants

The graph-seed mutant replaces the synchronous captured-cell seed with `Refused` while preserving the async branch. Running `TestMemoizeRegions/a` with its Go overlay exits 1: `captured callback cycle seed refused again`. The unmodified fixture passes. The environment mutant deletes exactly one generated `adamic_graph_adopt_owned` call for the frame environment. It compiles, and with leak detection disabled prints Node's identical answer. LeakSanitizer alone kills it: 243 leaked bytes in three allocations. This second mutant remains an executable test; the seed overlay and its failure log are recorded evidence.

## Toolchain and verification

`GOPROXY='https://proxy.golang.org|direct'`, `ADAMIC_TOOLS=/workspace/adamic-tools`, `GOFLAGS=-buildvcs=false bash cloud/setup.sh`: Node ready 0.097 s; Go 0.174 s; submodules 0.467 s; clang 0.489 s; markdown 0.753 s; Go build 135.713 s; test binaries deferred 136.677 s; cache warm 136.681 s; done 136.819 s. `nproc` is 5; quota is four CPUs. The first setup attempt rejected a cohere symlink; proper shared Git worktrees fixed that, and `-buildvcs=false` avoids Go's VCS stamping failure for the linked checker workspace. No cohere source was copied. Disk pressure interrupted early builds; those artifacts were discarded and the separate exact runtime baseline was rebuilt successfully.

The requested merge had overlapping compiler changes. I retained runtime's typed-array dispatch and preallocated captured-variable guard alongside current main's enum dispatch, cast-proof implementation and cyclic-module guard. Landing current main ce0750f28ef3943057f1f852b3ae5d93e6c5d644 added its inherited ignored-signal reset; the conflict resolution retains runtime's concurrent signal loop. A queued oracle run embedded temporary conflict markers during this merge; it is superseded by the merged-tree rerun recorded below. No push is based on that failed run.

Merged-tree verification (all test output first written to files):

- `go test -buildvcs=false ./internal/lower -count=1 -timeout=30m`: PASS, 170.092 s.
- `ADAMIC_GATE_UNCACHED=1 go test -buildvcs=false ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(memoize_regions|graph_regions|nested|proven_|cast_|typed_arrays)|TestAsyncReaderProbes|TestMemoize' -count=1 -v -timeout=30m`: PASS, 226.871 s; 230 uncached Node observations, 426 native observations.
- `go test -buildvcs=false ./internal/native ./internal/oracle -run 'TestGraph|TestMemoize|TestASignalLeavesWhatWasPrinted|TestClosedStdoutEndsAsOnNode' -count=1 -v -timeout=30m`: native graph tests PASS, 170.189 s. Oracle initially found the stale counts table, and three timed Node startups missed their first line under concurrent load. All actual graph frees balanced; the count and signal reruns below replace those failures.
- `ADAMIC_GATE_UNCACHED=1 go test -buildvcs=false ./internal/oracle -run '^TestASignalLeavesWhatWasPrinted$' -parallel=1 -count=1 -v -timeout=30m`: PASS, 41.228 s, all six default/inherited-ignored cases.
- `go test -buildvcs=false ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=30m -args -update-counts`: PASS, 471.978 s. The table grows from 595 to 707 rows and records the two graph columns. Every existing first-six-column heap count is unchanged; no row was removed. The full audit is counts-diff.json.
- `GOFLAGS=-buildvcs=false go vet ./...`: PASS. `gofmt -l cmd internal`: empty.

The full repository gate was not run. A preliminary lower/native/fresh package run had the stale Promise test failure, later fixed and fully rechecked in lower; its unrelated native decode corpus is slow. The worker gate is full lower, native graph tests, the expanded oracle above and the recorded counts. macOS leak checking was not executed; the permanent environment-omission test accepts its counted leak report on macOS, while this Linux run proves LeakSanitizer catches it.

Final graph/count gate: `go test -buildvcs=false ./internal/oracle -run '^TestGraphRegionsCountsAndFree$|^TestMemoize' -count=1 -v -timeout=30m`: PASS, 17.781 s.
