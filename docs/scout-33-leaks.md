# Step 33 scout: no leaks across a full native compile

Roadmap step **33 (#bwj5w0v)**. Scout on `runtime/scout-33-leaks`, from fetched
`origin/area/runtime` **cdfa22555589194e6f3133d986dd110061aafaa9**. This is ownership
coverage and readiness evidence; it does not complete step 33. No compiler,
loader, runtime, source adaptation, or upstream TypeScript production code changes.

## What “no leaks” means

When tsc builds natively, run the finished native compiler through
[`internal/leakcheck`](../internal/leakcheck/leakcheck.go). A clean full compile
of the pinned `src/compiler/tsconfig.json` project must finish, preserve the
reference diagnostics/output, and return an empty leak report. Check compilation
work performed by native tsc, not memory used by the Go process building tsc.
The build succeeding alone is not a leak test.

* **Linux:** `native.Options{Sanitize: true}` produces the ASan/UBSan binary;
  `leakcheck.Check` runs it again with `ASAN_OPTIONS=detect_leaks=1`. A nonzero
  exit or LeakSanitizer report fails. Sanitized allocation uses malloc rather
  than the production small-object slabs. LSan checks unreachable allocations;
  objects still reachable from module globals or exit-stack roots can escape
  its diagnosis. An empty report is evidence for the tested execution, not a
  proof about every ownership path.
* **macOS:** the helper builds with `native.Options{Count: true}`. A finished
  run must print its counts and satisfy **allocations = frees + regional values**
  via `leakcheck.Unbalanced`. It then runs `leaks --atExit -- COUNTED ...` to
  check malloc storage outside those value counts, including Map tables, array
  buffers and region blocks. `leaks` alone cannot detect small values left
  reachable in the runtime's size-class tables. Statement-region values are
  accounted separately; graph-region objects are included in frees.
* A checker diagnostic exit, refusal, crash, timeout, skipped test or failure
  to build the instrumented binary is not a leak-clean pass. The existing
  oracle checks programs that Node finishes with exit zero. Start the full
  compile gate with a clean project. Later negative-diagnostic/cancellation
  drivers must let their tested operation unwind and assert its expected
  diagnostics before returning normally to the shared check, rather than
  silently ignoring a nonzero sanitizer exit.

This scout ran on **Linux only**. The counted predicate was also exercised here;
there is no claimed macOS execution or `leaks --atExit` result.

## Ownership risks in tsc

Source inspection uses TypeScript **6.0.3**, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, and the existing stage 3 driver input
profile. Locations below refer to that source, before gathering rearranges lines.
These are prioritized shapes, not measured leak attribution from a native tsc.

| Risk | Concrete owners and why teardown matters | Evidence / next check |
| --- | --- | --- |
| **Program region teardown (step 06)** | AST parent/child backedges, bound symbols, attached metadata and outside string/array/Map payloads must be freed after the last external Program/checker keeper goes. Throwing during construction must unwind each owner exactly once. | Existing `graph_regions_parse`, `region_end`, `regions_throw`, `graph_regions_throw` fixtures; graph closure/container runtime controls. Verify step 06's Program lifetime when the real Program is executable; these small graphs do not prove its implementation. |
| **Module state** | `checker.ts` has top-level `typeofNEFacts` and `intrinsicTypeKinds` Maps (1313, 1443); compiler factories and namespaces initialize more shared values. Process lifetime is a legitimate lifetime, but all dynamically owned values must be released at normal exit. Keeping a Program in a global cache also hides a leak from LSan reachability. | Existing module initialization fixtures and the new cross-module checker cache. Distinguish intentionally shared intern data from accidental Program roots; check shutdown and in-process reuse. |
| **File/resolution caches** | `program.ts` host caches include `readFileCache`, `fileExistsCache`, `directoryExistsCache`, nested `sourceFileCache` (530–533), and `filesByName` (1709). Old-Program reuse and release callbacks (1834 onward) create overlapping lifetimes. | Existing `graph_regions_cache` plus new cache overwrite/delete/clear shape. Later run replacement of an old Program, host callbacks and cancellation. |
| **Maps of nodes / types / symbols** | `checker.ts`'s tuple, union, intersection, literal, indexed-access and template caches (2048 onward), `cachedTypes`, and `nodeLinks`/`symbolLinks` arrays (2357–2358) can retain whole graphs. Map keys own references too; table storage and overwritten values require distinct cleanup. | Existing `graph_regions_coverage_maps` includes node keys and values, copies, iteration, replacement, deletion and clear; runtime container-boundary control. New fixture combines node cache with a checker and AST cycle. |
| **Closures over the checker** | `program.ts` lazily retains `createTypeChecker(program)` (2684–2685); the returned checker object (checker.ts:1610) exposes methods closing over the checker scope and caches. Cancellation discards `typeChecker` (program.ts:2851). Retained callback environments can outlive a cache deletion or keep their own graph alive. | Existing `graph_regions_closure` and `graph_regions_coverage_closure_loop`; new checker↔closure cycle retained after module-cache deletion. Later test real checker methods, diagnostics callbacks and cancelled/replaced Programs. |

## Source-site census and cross-runtime evidence

The reproducible [AST inventory](scout-33-leaks/inventory.cjs) parses **77 tracked
`src/compiler/**/*.ts` files** from the pristine pinned Git tree with stock
TypeScript 6.0.3. Generated diagnostic source is excluded from this census.
[Every site, file:line:column, source hash and excerpt](scout-33-leaks/source-sites.json.gz)
is retained as compressed JSON (`gzip -dc` to inspect). Counts are syntactic exposure: `.set`, `.delete` and `.clear` can
have non-Map receivers, and nested functions are not necessarily closures over
owned values. No capture or leak is inferred from these totals alone.

| Exposure | Whole compiler sites | Important concentrations |
| --- | ---: | --- |
| `new Map` | 354 | checker.ts 90 (88 inside createTypeChecker), program.ts 30, resolutionCache.ts 12 |
| `new Set` | 109 | checker.ts 23, program.ts 7, resolutionCache.ts 17 |
| Map/Set construction outside a callable body | 30 Maps / 11 Sets | 2 Maps in checker.ts; this lexical category can include class initialization and is not proof of process-global reachability |
| Property calls `.set` / `.delete` / `.clear` | 399 / 93 / 55 | checker.ts 137 / 9 / 2; program.ts 46 / 6 / 0; resolutionCache.ts 15 / 19 / 17 |
| Direct `.parent = ...` assignments | 28 | checker.ts 18, program.ts 1; indirect setters and bulk copies are not counted |
| Calls to `getNodeLinks` / `getSymbolLinks` | 141 / 95 | checker.ts, associated stores declared at 2357–2358 |
| Callable bodies nested inside createTypeChecker | 3,286 | returned `checker` object at checker.ts:1610 has 174 properties, not 174 independently proven capturing closures |

Concrete sites from the pristine pin:
[core.ts:519–526](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/core.ts#L519)
implements `getOrUpdate` with `has/get`, a callback and a Map insertion;
[checker.ts:20305–20315](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L20305)
interns string/number literal types with `get` followed by `set` on a miss;
[checker.ts:2932–2941](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L2932)
creates node/symbol links lazily;
[program.ts:2684–2685](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/program.ts#L2684)
stores `createTypeChecker(program)` on first use;
[program.ts:2844–2854](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/program.ts#L2844)
discards that checker on cancellation. These sites connect cache retention,
callback environments and overlapping Program lifetimes to step 33.

### Node and ECMAScript

[Node probe](scout-33-leaks/node-probe.cjs), run on Node **v24.19.0**, V8
**13.6.233.17-node.51**, passes cache-hit identity, cached `undefined`, survival
of escaped values and checker closures after cache eviction, iterator behavior
across delete/clear/add, and permanent iterator exhaustion. An explicit V8 GC
request while escaped references remain live does not invalidate them. This
is observed behavior, not a bound or guarantee on reclamation of unreachable
cycles. The two compiler-shape fixtures are independently held to Node's exact
stdout in the native oracle and focused tests.

ECMAScript's [Map get](https://tc39.es/ecma262/#sec-map.prototype.get) and
[has](https://tc39.es/ecma262/#sec-map.prototype.has) algorithms distinguish a
missing entry from an entry containing `undefined`; `get` alone cannot make
that distinction. [Set](https://tc39.es/ecma262/#sec-map.prototype.set) updates
an existing entry without changing the value's independent aliases.
[Delete](https://tc39.es/ecma262/#sec-map.prototype.delete) and
[clear](https://tc39.es/ecma262/#sec-map.prototype.clear) remove entry visibility;
they do not revoke references returned earlier.
[CreateMapIterator](https://tc39.es/ecma262/#sec-createmapiterator) preserves
live iteration over deletion and subsequent appends; an exhausted iterator
stays exhausted. These algorithms specify observable associations, identity and
iteration, not deterministic free operations.

[Function environments](https://tc39.es/ecma262/#sec-function-environment-records)
and [OrdinaryFunctionCreate](https://tc39.es/ecma262/#sec-ordinaryfunctioncreate)
preserve a function's captured environment for later invocation. Freeing an
escaped checker method's environment when its original call ends violates that
behavior. [Managing memory](https://tc39.es/ecma262/multipage/managing-memory.html)
adds liveness requirements for WeakRef/FinalizationRegistry and kept-alive
values; it does not make finalization a deterministic ownership oracle.
[Node's memory documentation](https://nodejs.org/en/learn/diagnostics/memory/understanding-and-tuning-memory)
describes V8 garbage collection; Adamic must obtain the observable results
through ownership rather than adopting that collector.
[Node process shutdown](https://nodejs.org/docs/latest-v24.x/api/process.html#event-exit)
allows synchronous exit handlers to run while the process exits, so a future
native Node-compatible shutdown cannot release reachable handler state first.
The currently tested fixtures use the supported console/Map/function subset;
no process-hook compatibility is claimed.

Reference URLs, retrieval date, byte counts and response hashes are recorded in
[reference-manifest.json](scout-33-leaks/reference-manifest.json). The semantic
claims above are summaries; no runtime memory-cost claim is derived from them.

### typescript-go

The actual checked-in typescript-go source inspected here is `cohere/TypeScript`
at **d92d9bfee114c80be2c375d72edae966176e3a4f** (cohere itself is
**7945d102a6c18dd36adf9114a758ce646e8b2359**). This differs from older dashboard
pins; the scout records this checkout's source, not those historical results.

* `tsc/internal/compiler/program.go:97` retains its checker pool, source-file
  state and diagnostic caches. `:611`, `:628` and `:637` acquire checkers.
* `tsc/internal/compiler/checkerpool.go:17–26` describes a release callback.
  In the built-in pool, `:331–340` and `:351–359` return callbacks that unlock
  mutexes; `:362–364` returns a no-op release. These callbacks are not evidence
  of freeing a checker or its nodes. The pool still holds checker pointers.
* `tsc/internal/checker/checker.go:597` declares the Checker; `:929–1003`
  initializes it with its Program, file list and many strong maps, including
  literal/type/symbol caches. Bound method values such as `compareSymbols`
  (`:940–941`) retain a receiver just as an escaped JS function can retain
  checker state, although the representations differ.
* `tsc/internal/core/linkstore.go:7–22` owns a Map of node keys to link pointers
  and an arena. `tsc/internal/core/arena.go:13–23` replaces the arena's current
  backing slice when full, returning stable element pointers. Old chunks can
  survive through those pointers. It has no per-element free operation.

The project-system pool has a different policy: `tsc/internal/project/checkerpool.go:319–356`
releases a request/semaphore slot and disposes cancelled checkers;
`:361–368` removes completed request associations; `:405–455` drops idle checker
and association roots; `:493–505` stops cleanup on discard. Its release comment
at `:338–340` explicitly keeps discarded-pool checkers alive so API clients can
continue resolving type/symbol handles until the pool is collected. Clearing
roots and stopping timer callbacks are essential even with Go's collector;
none of these operations gives Adamic permission to free a live escaped handle.
These project-service paths were inspected, not executed as a native tsc test.

The [Go GC guide](https://go.dev/doc/gc-guide) describes tracing reachable
objects from roots. Dropping a Program/pool/cache reference permits Go's
collector to reclaim unreachable state, including cycles; a returned element
pointer still keeps its backing allocation reachable. Adamic cannot translate
that behavior into “free the old slice/Program now” or emulate a mutex unlock
as a destructor. The stage 1 whole-parser/scanner runs compare native outputs
with these Go ports, but do not measure Go's memory reclamation.

## Design inside Adamic's rules

This is an implementation proposal and its unresolved policy boundaries, not
new language rules. Keep **no garbage collector**, **never free early**, and
**Node as the semantic oracle**. The first delivered shape needs no policy
change: typed cache eviction removes the cache's ownership edge while escaped
values and closures continue to own what they use.

* Use ordinary counted ownership for independent strings, containers and
  acyclic values, and existing graph-region ownership for classified cyclic
  node/checker/closure shapes. A Program lexical scope is not by itself proof
  that every handle is dead. Release a region only at its established last
  outside keeper. Do not scan the heap to discover garbage.
* Treat a closure environment, a module cache and a returned checker method as
  real owners. Model their edges explicitly in lowering. Cross-Program aliases
  must use the existing supported region merge/keeper mechanism or retain
  their original owner; never rehome or invalidate a live node merely because
  a cache deletes it. Region merging can retain disconnected nodes until the
  last keeper; this is a reclamation-granularity limit to document and test.
* Preserve cache hit identity and distinguish absent from stored `undefined`.
  Lower Map overwrite/delete/clear with the container's own edge release, and
  retain/move each surviving alias correctly. Free backing storage only when
  no supported operation or iterator can still use it. Keep tombstones/live
  iterator order compatible with Node. The existing map runtime implements
  those ownership operations; this scout changes no production path.
* On normal module shutdown, release every initialized dynamic module owner
  once after its permitted users finish. Track partial initialization separately
  so failure paths do not release uninitialized cells. The full compile driver
  must drop its own roots and finish before invoking the leak verdict; hiding
  Program roots in a permanent table is not a substitute for ownership.
* Extend evidence incrementally: Node semantic comparisons, sanitized native
  execution, shared leak check, counted balance at lifetime boundaries, and
  intentional semantic/ownership mutants. LSan's reachability and macOS's
  allocator-table visibility are complementary limits, so neither detector's
  absence of a report replaces the other platform's acceptance run.

### Hard cases

An old Program and a new Program can share a SourceFile or interned symbol while
only one cache evicts it. A callback can outlive the compiler call that created
it, or a checker can retain that callback in its own cycle. Copying Maps,
iterating while deleting/adding, storing nodes as keys, and caches containing
`undefined` each change which edges are live. Factory exceptions, cancellation,
partially initialized modules and diagnostic exits exercise distinct teardown
paths. Weak observations and synchronous exit hooks impose additional semantic
constraints. Opaque host callbacks, workers and tasks can hold owners beyond a
statically visible compiler call. They require their supported ownership
contracts to be established before the larger shape graduates; the small
fixture is not evidence for every such boundary.

## Questions for system_adamic

* Does step 33 require exit-clean balance only, or also return to a specified
  live-allocation baseline after each Program in a long-running compiler host?
* What ownership boundary should unchanged tsc expose for Program teardown
  when a checker method or SourceFile escapes, and should step 06's region
  extend to that last keeper or use an explicit supported keeper API?
* What reclamation granularity is acceptable when merging two Program/checker
  graphs retains nodes that have become disconnected inside a still-live region?
* How should stage 3 admit tsc's generic `has/get!` helper under the current
  non-null-assertion refusal, especially when the Map value type includes
  `undefined`, without weakening the checked value rules?
* What supported contract should carry owners across opaque host callbacks,
  tasks or workers that may keep checker methods beyond a compile call?
* What observable WeakRef/FinalizationRegistry behavior, if any, belongs to
  Adamic's supported subset under the no-collector rule?
* Which Node-compatible module shutdown and synchronous process-exit observer
  behavior must be preserved before module-global owners may be released?
* Which expected diagnostic, cancellation and panic exits must count as finished
  leak-checkable workloads, and what normal-return driver contract should expose
  those paths without masking a sanitizer failure?

## Stage 1: executable ownership workloads

| Port | Selected test | Linux leak result |
| --- | --- | --- |
| `stage1/cohere/gitignore` | `TestThePortAnswersAsGoCohereAndGitDo` | PASS, shared helper |
| `stage1/cohere/graphql` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/mediaquery` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/values` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/suppression` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/selector` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/formatfiles` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/json` | `TestPortMatchesGoCohere` | PASS, LSan on every chunk |
| `stage1/cohere/css` | `TestThePortParsesAsGoCohereDoes` | PASS, shared helper |
| `stage1/cohere/cssnumbers` | `TestCSSNumbers` | PASS, shared helper |
| `stage1/cohere/cssstrings` | `TestCSSStrings` | PASS, shared helper |
| `stage1/cohere/markdowninline` | `TestMarkdownInline` | PASS, shared helper |
| `stage1/cohere/markdownblocks` | `TestNativeMdastConstruction` | PASS, shared helper |
| `stage1/typescript/scanner` | `TestScannerAgreesWithTypescriptGo` | PASS, explicit LSan; 77 compiler files, 476 stage 1 files, 18,237 generated inputs; 35,523,261 identical answer bytes |
| `stage1/typescript/parser` | `TestWholeCompilerAgrees` | PASS, explicit LSan; 78 pinned compiler files, 46,537,475 identical whole-tree bytes |

JSON's initial run failed its corpus pin because the 18 provisioned API JSON
files were missing; the locked npm setup restored their exact identity and the
final 2,790-case run passed all five native chunks and their leak checks.
The first scanner run's baseline passed but the test timed out during its
mutants under concurrent load; the final rerun passed all four mutants as well.
The whole-parser initially skipped without its corpus variable, then passed
with the pinned checkout supplied. Markdown inline's initial run timed out and
its next run refused the dirty tracked counts table; its final selected test
runs in a clean detached checkout of this exact base with the same pinned
submodules. None of those initial runs is counted as a final pass.


The table surveys all thirteen stage 1 cohere port directories that import the
shared leak helper on this base, plus the TypeScript scanner and whole-parser
ports. It is not a claim that every other stage 1 directory is implemented or
leak-checked. Lint/type-aware/SSA/printer-specific harnesses and audit-only YAML
are outside this selected port survey. Optional comparisons with external
GraphQL/postcss libraries, or inapplicable mutants, can skip without skipping
the mandatory native leak run; their exact skips remain in the logs.

The TypeScript ports currently use their own sanitizer execution rather than
`leakcheck.Check`; this scout explicitly supplied `ASAN_OPTIONS=detect_leaks=1`.
That exercises Linux's detector, but they need the shared helper wired into
their harnesses to get its macOS counted/build/storage check automatically.
JSON checks every Linux execution chunk directly with LSan and uses the shared
helper on macOS. The remaining surveyed cohere ports call `leakcheck.Report`.

## Stage 3: source slices and the full compiler

The input tree is made by the checked-in scanner runner: adaptations **00, 10,
50, 51**, with **20 excluded**. It includes upstream diagnostic generation.
The gathered scanner then applies existing slice adaptations **52–57, 59,
81–82, 85**. The parser input receives existing **60–64** full-tree repairs before gathering.
These are explicit source profiles, not a full adaptation-set integration or
an unmodified full compiler. The compiler executable is built from this scout's
base, with no feature-branch merges. Stock TypeScript and Node are separately
pinned. Imported declarations retain their bytes and the original ordered
module evaluation graph; no reference-only gathering is used.

| Entry / source profile | Native readiness on this base | Leak verdict |
| --- | --- | --- |
| `core.ts:returnFalse` + `returnTrue`; 2 code declarations, 79 copied spans, 78-module graph | Builds; exact `false true` stdout agrees with stock-source Node driver; byte audit passes | PASS through shared helper and counted balance: 1 allocation / 1 free; output mutant rejected |
| `core.ts:compareValues`; 5 code declarations, 82 spans, 78-module graph | Fails at core.ts:6:1, function declaration without a body (overload lowering); byte audit passes | BLOCKED before native execution |
| `scanner.ts:createScanner` + enum entries; 89 code declarations, 168 spans, existing scanner slice profile | Audited slice's API probe fails at corePublic.ts:9:5, index signature refused | BLOCKED before native execution |
| Parser driver's entries; 1,988 code declarations, 2,077 spans, 78-module graph; existing 60–64 repairs before gathering | Audited API probe fails strict checking, beginning at core.ts:110:34 (`T | undefined` not assignable to `T`), with additional recorded diagnostics | BLOCKED before native execution |
| Checked-in scanner driver against full 00/10/50/51 tree | Node runs; native build fails strict checking, starting at binder.ts:1109:17 (undefined optional field write) | BLOCKED; Node pass is not a native leak result |
| Real `src/tsc/tsc.ts` entry against that full tree | Native sanitized build fails strict checking, beginning at binder.ts:1109:17; no native tsc binary produced | Full `src/compiler` compile leak check unavailable |


The leaf control establishes that the shared check can run on a real gathered
stage 3 slice today. It has no AST/checker workload and cannot stand in for step
33. The scanner driver also ran on Node over 81 compiler-directory files:
1,369,441 tokens, 466 scanner error records, 108,020,030 output bytes, SHA256
`0f62f7c7dcc4799a49dcb9a1036f8aec932167976f0cf0877fd0976e55b22060`.
Its copied-output control passed and end-offset mutant was caught; its native
build was blocked. A Node result does not measure native ownership.

The unused driver Node declaration imports are also unsupported by this loader
base. Scratch scanner/parser API probes omit those unused imports so the table
can distinguish source/API blockers from driver declaration setup. No upstream
source is changed for those probes. Supplying the complete stock Node index
instead conflicts with the console prelude and is not treated as a repair.

An initial core/scanner gather overlapped source adaptation and failed its
byte audit. Those probes were discarded. The final core/leaf/scanner gathers
use the completed tree and pass their recorded audits; only final probes count.

## Risk fixtures and live leak controls

| Existing risk fixture / control | Fresh Linux result |
| --- | --- |
| `graph_regions_parse.a` — AST parent/child graph | PASS, Node parity + shared leak check |
| `graph_regions_cache.a` — Map keeps cyclic node graph | PASS, Node parity + shared leak check |
| `graph_regions_coverage_maps.a` — keys, values, copies, iteration, replacement/delete/clear | PASS, Node parity + shared leak check |
| `graph_regions_closure.a`, `graph_regions_coverage_closure_loop.a` — saved closure/environment cycles | PASS, Node parity + shared leak check |
| `region_end.a`, `regions_throw.a`, `graph_regions_throw.a` — normal/abrupt teardown | PASS, Node parity + shared leak check |
| `modules/main.a` and matched enum/devirtualization module controls | PASS, native initialization/output parity + shared leak check |
| `TestGraphRegionsRuntime`, `TestGraphClosureEnvironment`, `TestGraphContainerBoundary` | PASS, shared leak checks; anchor leak mutant detected |
| New `scout_33/main.a` | PASS, Node parity + shared leak check + 125/125 counted balance + publication mutant |
| New `scout_33/cache.a` | PASS, Node parity + shared leak check + 27/27 counted balance + identity/undefined/leak mutants |


The combined shape is now
[`internal/oracle/testdata/scout_33/main.a`](../internal/oracle/testdata/scout_33/main.a)
and `state.a`: four Program-like lifetimes create cyclic parent/child trees,
replace Map entries, create a checker↔closure cycle, publish it through a
module-global Map, delete its cache entry and clear the cache, then use the
escaped checker closure. Its direct checker↔closure backedge is a reduced
stress model for Program's cached-checker getter and checker methods that capture
host/Program state (`program.ts:2684`, `checker.ts:1610–1627`), not a claim that
tsc assigns `checker.read = read` verbatim. Parent links follow
`utilities.ts:10705`/`:10722`; literal caches follow `checker.ts:20305`.
All object payload labels and cache keys in this combined fixture are built dynamically. This checks both
premature frees and leaks across the interacting owners; it does not model all
real tsc caches or prove bounded peak memory during a long compile.

It is registered in the ordinary native oracle, so it gets source/Node and
backend comparisons, release/slab builds, sanitizers and the shared leak check.
`TestScout33CountedBalance` independently checks its release counted build on
Linux with the same predicate used by macOS. Result: **125 allocations, 125
frees, 84 retains, 198 releases, peak 59, 0 statement-region values; 4 graph
regions, 40 merges**. At final graph teardown the counted diagnostic records
**41 live objects / 3,984 bytes**, of which **1 object / 400 bytes** is reachable
from the remaining keeper and **40 objects / 3,584 bytes** are unreachable inside
the region; metadata is **912 bytes**. The module Map remains an outside keeper
even after its entries are removed, so this is actual observed coarse lifetime
retention, not a claim of eager reclamation after each compile. Everything is
freed at the final keeper release. No proof of bounded retained memory across
arbitrarily repeated Programs follows from 125/125 exit balance. The acceptable
region granularity and per-Program acceptance baseline remain questions above.

The regenerated counts table adds only the two new fixture rows; existing rows
are unchanged.

Existing runtime negative controls prove the detector is live: omitting the
last graph anchor release is caught as **383 bytes in 5 malloc allocations**
by the shared Linux check, and as **4 leaked values (5 allocations, 1 free)**
by the counted predicate. The restored anchor balances 5/5. Closure/environment
and container-boundary controls balance 2/2 and 6/6, respectively. No new
production-runtime mutant or suppression was added.

## First delivered shape and fixture mutants

The first no-ruling shape is the checked literal-cache port in
[`scout_33/cache.a`](../internal/oracle/testdata/scout_33/cache.a) and
[`getOrUpdate.a`](../internal/oracle/testdata/scout_33/getOrUpdate.a), green in
**`TestScout33CacheEvictionAndMutants/cache.a`**. It preserves tsc's memoization
and callback construction shapes while replacing the refused assertion with
explicit `get` narrowing for non-undefined literal objects. A separate checked
number-or-undefined specialization retains the source `has/get` distinction.
No generic cast, collector, early region release or production-runtime change
is used. Control output:

```text
2 true replacement3:2
true false
name3:1 name3:1 0
1 true true true
```

Native source, JavaScript backend, release/slab variants, ASan/UBSan and the
shared Linux leak check agree with Node. Its release counted build balances
**27 allocations / 27 frees**, with 23 retains, 42 releases, peak 12, no
statement or graph regions. The combined checker fixture balances **125/125**.

The [original generic helper excerpt](scout-33-leaks/core.getOrUpdate.ts.txt) was also
attempted and stops at the existing non-null-assertion refusal. It is retained
as research evidence, not a secretly rewritten full stage 3 implementation.
The question about admitting unchanged generic source remains above; the excerpt normalizes line endings only.

| Fixture / shape | Mutant | Required failure observed |
| --- | --- | --- |
| Combined module/checker/AST fixture | Do not publish the captured reader; publish a constant reader instead | Original Node stdout rejects `unpublished`; mutated Node/native agree, finish under sanitizers and are leak-clean |
| Literal-cache fixture | Return a fresh callback value on a cache hit | Original Node rejects changed creation count/identity; mutated Node/native agree and are leak-clean |
| Cached-undefined branch in cache fixture | Treat every `has` result as a miss | Original Node rejects callback count 2 and a present number in place of cached undefined; mutated Node/native agree and are leak-clean |
| Cache ownership | Add an extra value retain to generated C's Map insertion path | Identical original Node stdout, but shared LSan reports **245 bytes in 4 allocations**; counted predicate reports **4 leaked values (27 allocations / 23 frees)** |
| Gathered returnFalse/returnTrue leaf | Call returnFalse for the second result | Original Node comparison rejects `false false`; mutated Node/native agree; shared check and counted balance remain clean |

All mutants are confined to temporary copies or generated C; none is installed
in the production runtime. The source identity mutant originally removed its
narrowing and failed to type-check; it was replaced with the well-typed fresh
callback variant above, so final sensitivity is runtime/output evidence rather
than an incidental checker error. Tests require each mutation to change its
intended site and distinguish semantic failures from ownership failures.

## Staged acceptance plan

1. **Keep today's executable coverage live.** Run the selected stage 1 port
   oracles and ownership fixtures uncached. Route the TypeScript port harnesses
   through `internal/leakcheck` when that harness work is authorized. Keep both
   the unreleased-anchor detector control and semantic comparisons; sanitizer
   success with a skipped workload is insufficient.
2. **Graduate source slices as their native blockers land.** Start with the
   audited leaf control, then the scanner corpus, then parser/AST construction
   with JSDoc and parent links. Build each native slice sanitized and counted;
   compare the stock source driver's bytes, then invoke the same shared check
   with the same corpus and fresh output paths. Preserve manifests/profile pins
   so a changed adaptation or declaration closure is visible. Add binder,
   node/symbol Map and checker slices as they become executable; none is claimed
   leak-clean here.
3. **Verify the step 06 Program lifetime with real consumers.** Construct a
   Program, force its checker and resolution/type caches, collect diagnostics,
   emit where supported, and release all external handles. Repeat in one
   process with fresh Programs and old-Program reuse, retaining a checker method
   temporarily, then dropping it. Record counts at teardown boundaries and
   verify that live owners return to the documented baseline after caches and
   callbacks are released. A process-exit LSan pass alone cannot establish this.
4. **Gate the full native compile of `src/compiler`.** Use the pinned project
   and reference compiler options/libraries, run `tsc --noEmit` first and a full
   emitting build separately, and compare diagnostics and emitted artifacts
   before accepting an empty shared leak report. Give each execution fresh
   writable output paths through `leakcheck.Program.Arguments`. Require Linux
   LSan and actual macOS counted balance plus `leaks --atExit`, with recorded
   platform evidence. Exercise clean completion, expected diagnostics,
   cancellation, partial initialization, reused Program/host, and normal module
   shutdown through finished drivers. Missing native support remains a blocker,
   not a waived leak result.

Step 33 can close only when the full native compiler workload meets that gate.
The port and shape results locate today's executable frontiers; they do not
establish “no leaks across src/compiler” yet.

## Reproduction and evidence

Measured 2026-10-08 on host `cb1632b4fc8c`, Intel Xeon Platinum 8573C,
Linux 6.18.44, five affinity CPUs and a four-CPU cgroup quota. Go 1.27.1,
clang 20.1.8, Node **v24.19.0**, stock TypeScript **6.0.3**. Port checks ran
concurrently; these are correctness/leak results, not performance measurements.
Port tests' incidental throughput logs are retained as test evidence; they are not used to make a performance claim.

Setup: fetched `area/runtime` by explicit refspec, created the named branch,
exported `GOPROXY='https://proxy.golang.org|direct'`, ran
`bash cloud/setup.sh --wasi-sdk`, sourced `/workspace/adamic-tools/env.sh`.
Provisioned `stage3/api` with its existing npm lock to satisfy the JSON corpus
pin. No AGENTS.md was found. No whole-package test selection or full gate ran.

Selected execution commands (from the repository root with the setup environment
sourced; test deadlines are harness limits, not work estimates):

```sh
ASAN_OPTIONS=detect_leaks=1 ADAMIC_GATE_UNCACHED=1 go test \
  ./stage1/cohere/gitignore ./stage1/cohere/graphql ./stage1/cohere/mediaquery \
  ./stage1/cohere/values ./stage1/cohere/suppression ./stage1/cohere/selector \
  ./stage1/cohere/formatfiles ./stage1/cohere/json ./stage1/cohere/css \
  ./stage1/cohere/cssnumbers ./stage1/cohere/cssstrings \
  ./stage1/cohere/markdowninline ./stage1/cohere/markdownblocks \
  -run '^(TestThePortAnswersAsGoCohereAndGitDo|TestThePortParsesAsGoCohereDoes|TestPortMatchesGoCohere|TestCSSNumbers|TestCSSStrings|TestMarkdownInline|TestNativeMdastConstruction)$' \
  -count=1 -v
# Successful JSON rerun after npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund:
go test ./stage1/cohere/json -run '^TestPortMatchesGoCohere$' -count=1 -timeout 30m -v
ASAN_OPTIONS=detect_leaks=1 go test ./stage1/typescript/scanner \
  -run '^TestScannerAgreesWithTypescriptGo$' -count=1 -timeout 30m -v
ASAN_OPTIONS=detect_leaks=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/scout-33/stage3-scanner/adapted \
  go test ./stage1/typescript/parser -run '^TestWholeCompilerAgrees$' -count=1 -timeout 30m -v
# Markdown inline final run: clean detached cdfa2255 checkout, same pinned submodules.
ASAN_OPTIONS=detect_leaks=1 go test ./stage1/cohere/markdowninline \
  -run '^TestMarkdownInline$' -count=1 -timeout 30m -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^scout_33$/(cache|main)[.]a$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestScout33(CountedBalance|CacheEvictionAndMutants)$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^(graph_regions_parse|graph_regions_cache|graph_regions_coverage_maps|graph_regions_closure|graph_regions_coverage_closure_loop|graph_regions_throw|region_end|regions_throw)[.]a$' -count=1 -v
# Explicit module run, since Go splits -run patterns at slash boundaries:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(scout_33|modules)/main[.]a$' -count=1 -v
go test ./internal/native \
  -run '^(TestGraphRegionsRuntime|TestGraphUnreleasedAnchorCounted|TestGraphClosureEnvironment|TestGraphContainerBoundary)$' \
  -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
node --expose-gc docs/scout-33-leaks/node-probe.cjs
```

The first broad port invocation and the initial module fixture invocation retain
their precondition failures in the evidence. Final fixture/port logs named
`*-final` or `markdowninline-clean` establish the accepted result. The focused
mutant tests check both fixture controls, not only the mutants.

Recreate the source census and the executable leaf slice:

```sh
scout_dir=$(mktemp -d)
# Use the pinned Git object, not the adapted tree, for the source census.
git --git-dir=/home/agent/.cache/adamic-stage3/typescript.git \
  archive 050880ce59e30b356b686bd3144efe24f875ebc8 src/compiler | tar -x -C "$scout_dir"
SCOUT_TYPESCRIPT="$PWD/stage3/api/node_modules/typescript/lib/typescript.js" \
  node docs/scout-33-leaks/inventory.cjs "$scout_dir" "$scout_dir/source-sites.json"
bash stage3/drivers/scanner/run.sh "$scout_dir/scanner-proof"
# The scanner runner returns failure at its native gate on this base; keep that report.
SLICE_TYPESCRIPT="$PWD/stage3/api/node_modules/typescript/lib/typescript.js" \
  bash stage3/slice/run.sh "$scout_dir/scanner-proof/adapted" "$scout_dir/leaf" \
  src/compiler/core.ts:returnFalse src/compiler/core.ts:returnTrue
cp docs/scout-33-leaks/leaf-entry.a.txt "$scout_dir/leaf/probe.a"
node stage3/slice/verify.cjs "$scout_dir/leaf"
SCANNER_TYPESCRIPT="$PWD/stage3/api/node_modules/typescript/lib/typescript.js" \
SCANNER_RUNTIME="$PWD/oracle/adamic.mjs" \
  node --disable-warning=ExperimentalWarning stage3/drivers/scanner/node.mjs \
  "$scout_dir/leaf/probe.a"
go run ./docs/scout-33-leaks/check-slice.go "$scout_dir/leaf/probe.a"
```

Leaf Node/native stdout is compared byte for byte in the recorded run. Its
scratch mutant changes only the second call to `returnFalse`, compiles and runs
sanitized, agrees with its own Node source, is rejected against the original
Node stdout, and remains clean through the shared helper. Final scanner and
parser API entries, compressed JSON slice manifests and native diagnostics are retained beside
the leaf evidence. Parser 60–64 adaptations are applied before gathering; their
byte audit passes. Build blockers are expected readiness results, not ignored
failures of an executable leak test.


[Evidence directory](scout-33-leaks/) retains final logs, readiness diagnostics,
source profiles, byte-audit results and the leaf entry. The scouting helper
[`check-slice.go`](scout-33-leaks/check-slice.go) builds an entry and invokes
`internal/leakcheck.Check`, then also exercises counted balance on Linux. Use
it for a source slice only after a native semantic comparison establishes what
that slice should do. Passing its leak check does not itself establish semantics.
