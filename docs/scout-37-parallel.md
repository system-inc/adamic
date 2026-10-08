# Scout for roadmap step 37: parallel parsing and checking

Issue: #xfpyaj8. The initial scout was source inspection and a scratch build
probe. The expansion below adds executable source-shaped reductions, independent
tsc/Node projections, and sanitizer-backed mutants. The full 77-file stage 1
class parser still does **not build with parallelMap**, so there are no
1–16-thread whole-parser measurements or speedup claims. The reduced worker-local
parser is built under the existing rules; no production compiler rule changed.

## Inputs and scope

- Adamic base: `origin/area/runtime`, fetched by name, at
  `cdfa22555589194e6f3133d986dd110061aafaa9`.
- Report branch: `runtime/scout-37-parallel`.
- typescript-go source: the existing `cohere/TypeScript` submodule, commit
  `d92d9bfee114c80be2c375d72edae966176e3a4f`, particularly its `tsc/internal`
  compiler, parser, checker and core packages. This is the repository's pinned
  fork, not an unpinned upstream checkout.
- Stage 1 corpus: TypeScript tag `v6.0.3`, commit
  `050880ce59e30b356b686bd3144efe24f875ebc8`; all 77 `.ts` files recursively
  under `src/compiler`, 9,400,075 source bytes. The checkout is in scratch.

[Stage 1's whole-parser report](../stage1/typescript/parser/WHOLE_REPORT.md)
previously measured native slower than Go by 4.994x on this corpus. Those are
historical measurements on another machine and compiler revision. They motivate
this conditional roadmap step; they do not establish today's single-core ratio.
The indexed-node stage 1 parser is also distinct from stage 3's port of the
original TypeScript compiler.

## How typescript-go does it

**Parsing is per file, with private parser state.**
[`filesParser.start`](../cohere/TypeScript/tsc/internal/compiler/filesparser.go)
queues parse tasks into a `core.WorkGroup`. Tasks can discover imports and queue
more tasks before the join. A synchronized map deduplicates canonical paths;
per-path mutexes protect casing variants, depth updates and loading the same
file. This is a dynamically discovered dependency graph, not merely a map over
the initial roots. The host's
[`GetSourceFile`](../cohere/TypeScript/tsc/internal/compiler/host.go)
reads a file and calls `parser.ParseSourceFile`.

[`ParseSourceFile`](../cohere/TypeScript/tsc/internal/parser/parser.go) acquires
an exclusive parser instance from `sync.Pool`, initializes its file state and
scanner, parses, then resets and returns the parser to the pool. Its mutable
scanner, diagnostics and current-token state belong to that invocation.
[`core.NewWorkGroup`](../cohere/TypeScript/tsc/internal/core/workgroup.go) uses
`sync.WaitGroup.Go` for parallel execution, with a queued sequential alternative
when `SingleThreaded` is selected. Goroutine execution is scheduled by Go; a
work group does not create one dedicated OS thread for every file.

**Binding precedes sharing the bound program for checking.**
[`Program.BindSourceFiles`](../cohere/TypeScript/tsc/internal/compiler/program.go)
queues one bind operation for each unbound source file and waits for completion.
Binding uses a separate pooled binder. A parallel checker design must preserve
this phase boundary and audit any lazy mutations that remain afterward.

**Semantic work uses multiple private checkers, not concurrent calls on one
checker.** The compiler
[`checkerPool`](../cohere/TypeScript/tsc/internal/compiler/checkerpool.go)
defaults to four checkers; `SingleThreaded` selects one and the `Checkers` option
can override the count, clamped to the file count and 1–256. Checkers are created
once, in parallel. Each has its own mutable symbol/type/instantiation caches and
node links, plus an exclusive acquisition lock
([`checker.NewChecker`](../cohere/TypeScript/tsc/internal/checker/checker.go)).

Files have stable checker associations. The inspected revision uses weighted
FENNEL partitioning of the resolved import graph: node count, text size and
syntactic imports estimate work; import adjacency rewards cache locality;
load penalties and a preferred load cap keep partitions balanced. Its source
ordering policy distinguishes source-dominated and declaration-heavy programs.
This reduces duplicate semantic cache construction without concentrating every
related file on one checker.

For whole-program diagnostics, `forEachCheckerGroupDo` queues one job per
checker, acquires its lock, and visits that checker's associated files in their
original order. Results occupy original file-index slots; Program subsequently
sorts and deduplicates diagnostics. Thus checker concurrency is bounded by the
pool, with serial work and reusable caches inside each partition. Query APIs
acquire the file's associated checker exclusively; nested operations pass that
checker onward rather than reacquiring it.

The language-server
[`project.checkerPool`](../cohere/TypeScript/tsc/internal/project/checkerpool.go)
is a different lifetime policy: a diagnostics checker, temporary query checkers,
and a persistent API checker for stable handle identity, with request/file
associations and exclusive acquisition. A batch fork-join map does not by itself
provide this persistent service pool.

## What Adamic can express at the inspected base

[`parallelMap`](concurrency.md) is synchronous structured fork-join with ordered
results and a sequential Node witness. `ADAMIC_THREADS=1` is the native baseline;
the override can select 1–16 threads, but does not change compilation proofs.
The native pool is lazy, uses work stealing and adaptive grain, and supports
nested joins. Defaults account for affinity and CPU quotas. WASI remains
sequential ([runtime](../internal/native/runtime/parallel.c)).

A fixed list of readonly file names and source strings can cross into workers.
Pre-read files on the parent, parse independently, then merge ordered results
on the parent. Source-level work may allocate and mutate proven fresh locals;
immutable captures must be structurally Shareable. This expresses the *shape*
of per-file parsing and pure summary computation. It does not currently admit
the actual stage 1 parser call. Parallel file I/O is explicitly refused, so
putting `readTextFile` inside the work callback is not an available shortcut.

A checker partition could likewise be one task: construct private checker state,
process a readonly list of files serially, return diagnostics, then merge them
in order. Today the effect proof does not establish the necessary class/method
and callback ownership. A mutable checker captured from the parent is refused;
a `readonly` field containing a mutable Map does not make its caches Shareable.
Raw locks and a persistent checker-acquisition API are not source-level features.

**The whole-graph ownership query is a design seam, not a landed general query
at this base.** [Concurrency moves](concurrency-moves.md) describes extending
`internal/flow` and `internal/fresh` to prove confinement, absence of surviving
aliases or Weak observers, and disjointness of item graphs at a transfer.
[`fresh.ProveWrites`](../internal/fresh/fresh.go) already interprets abstract
heaps, escapes and direct-call summaries, but its exported proof answers whether
writes can close cycles. Confinement does not establish exclusive ownership.
The current [`moves.go`](../internal/lower/moves.go) admits only a construction
proof for local arrays of distinct flat object literals with scalar fields and
restricted inline callbacks. It cannot transfer Parser/Scanner state, an AST
with nested arrays, or a checker's caches. No runtime root-count test substitutes
for the missing graph query.

## What the Program region of step 06 needs

There are two existing region mechanisms to distinguish:

- [`region.go`](../internal/native/region.go) and
  [`runtime/region.c`](../internal/native/runtime/region.c) provide object-only,
  statement-scoped bump arenas. One mutable block chain and bump cursor are not
  safe for simultaneous allocation. A parsed AST retained by Program outlives
  this scope; strings, arrays and Maps need broader allocation routing.
- [`graph_regions.c`](../internal/native/runtime/graph_regions.c) handles cyclic
  graphs through dynamically merged regions and counts at their boundaries.
  Program can anchor a graph by holding an outside reference. This is not a
  Program bump arena. Its union-find records, member lists and counts are plain,
  and [`share.c`](../internal/native/runtime/share.c) rejects graph publication
  to workers. Task-local graphs that die before returning a scalar are already
  represented by the accepted `graph_regions_per_task.a` fixture.

Parallel parsing needs an allocation/lifetime protocol as well as the scheduler.
Two possible designs, neither implemented by this scout:

| Choice | Allocation and join | Ownership obligations |
| --- | --- | --- |
| One Program region with private allocator lanes | Each worker/task allocates from its own chunks and bump cursor; the parent joins all work before collecting chunk lists. A common lifetime anchor keeps every lane alive. | Route AST objects and their container/string storage consistently. Synchronize any common metadata and cross-lane graph links; an atomic cursor alone does not protect graph union/count operations. Region end must wait for all tasks, including exceptional exits. |
| Per-file regions adopted by Program after join | Each parse owns its allocations privately. Transfer the completed file and its region to the parent; Program keeps them through bind/check. Adopt chunks without copying or invalidating node addresses. | Prove no surviving worker aliases, captures or Weak observers; preserve source-string owners and external counted children. Cross-file symbols/backlinks need validated region merging or counted boundaries. Dropping a per-file owner cannot free nodes still reachable from another file or checker. |

Per-file ownership is the cleaner first experiment because parsing has a natural
file boundary and no concurrently edited allocator chain. It still needs a
transfer proof and a runtime handoff: returning a graph currently goes through
shared publication, which panics. Existing graph union can guide a parent-only
merge after all tasks finish, but does not supply thread-safe publication or
concurrent merging. A task region cannot simply be freed when the callback ends
if its AST is the result.

The Program lifetime includes bound ASTs and checker caches, not just parsing.
Readonly syntax sharing is useful only after mutable binder/parser scratch is
isolated. Checker-local caches should stay private to partitions; common intern
pools and lazily written AST fields need a separate proof or synchronization
protocol. The ownership query must also distinguish transfer from subsequent
readonly publication. [The compiler profile](stage3-tsc-profile.md) explains why
per-file storage cannot be reclaimed while the checker still reaches it.

## Build probe and refusals

The scratch driver pre-reads the manifest and sources, builds a
`readonly Input[]` whose `path` and `text` fields are readonly strings, then calls
`parallelMap(files, parse)`. Each callback constructs a fresh stage 1 Parser,
parses one whole file and returns only its node count. No parser is captured,
no file read occurs inside the callback, and no AST escapes. Its callback is:

```ts
function parse(input: Input): number {
    const parser = new Parser(input.text, input.path);
    const root = parser.file();
    return countTree(parser.nodes, root);
}
```

Building `/workspace/scratch/scout-37/parallel.ts` stops at line 10, column 20,
on `new Parser(...)`, with exit 1:

```text
Adamic 0.1 refuses parallelMap work writes state another task can reach: parse calls a closure whose body isn't known; keep work and everything it calls pure; write only task locals and fresh objects
```

[`parallelProof.call`](../internal/lower/parallel.go) handles known functions
and selected builtins, but does not prove this class constructor. Later method
calls are refused when their complete dispatch hierarchy is not proven, and
`this` ownership is explicitly unproven. Creating a private parser is sensible
semantically, but that fact is not yet established by the task proof.

Other relevant barriers, distinct from the first observed build refusal:

| Barrier | Evidence and implication |
| --- | --- |
| Class methods and `this` | `parallel.go`; existing `dispatch.a` exact refusal. Prove constructor/receiver confinement and every possible method target. |
| Effectful worker file reads | Existing `file_read.a` exact refusal. Preload source text or design a separate admitted I/O contract. |
| Mutable captures/items and unknown callbacks | `shareable.go`, `parallel.go`, `moves.go`. Parser state and checker caches need private allocation or proven exclusive transfer. |
| Graphs crossing workers | Existing `graph_region.a` refusal and `share.c` runtime backstop. A private graph can be used within a task; returning its AST requires a new handoff. |
| Original TypeScript parser singleton | The corpus's `parser.ts` namespace `Parser` owns a module-level scanner, file name, source text and current token. Directly mapping its `createSourceFile` would share mutable parser state; worker-local instances are a prerequisite for the stage 3 version. |
| Dynamic dependency discovery | Go queues newly discovered imports during parsing. Adamic can use parent-controlled batches and joins, but cannot reproduce a shared mutable loader/dedup table inside work callbacks under the current contract. |
| Persistent checker pools | Structured tasks can own private checkers for one batch; request-affine reuse and stable API handle lifetime need a separate service design. |

## Initial scout validation and measurement limits

Setup ran `export GOPROXY='https://proxy.golang.org|direct'` and
`bash cloud/setup.sh`, then sourced `/workspace/adamic-tools/env.sh`.
Tools: Node v24.19.0, Go 1.27.1, clang 20.1.8. No runtime source was touched,
so no WASI SDK setup was needed. No AGENTS.md was found in the workspace.

Machine: Linux amd64, host `11f79d9f08c2`, Intel Xeon Platinum 8573C,
5 visible logical CPUs, `cpu.max=400000 100000` (four CPU quota),
`memory.max=17179869184` (16 GiB). Setup load before/after was
`0.47 / 0.19 / 0.30` and `5.62 / 1.84 / 0.87`; during probe validation the
observed load included `1.59 / 1.68 / 0.99`. These are environment observations,
not a parallel performance experiment. Sixteen threads would oversubscribe
this machine even after compilation is enabled.

The sequential Node witness of the actual parallel driver completed all 77 files
and reported **887,803 nodes**, matching the stage 1 documented whole-tree count.
A sequential native control, differing only by replacing
`parallelMap(files, parse)` with `files.map(parse)`, also built and reported
**77 files, 887,803 nodes**. This isolates the parallel boundary from unrelated
loader or parser compilation failures. These controls validate driver execution
and the aggregate count, not full AST bytes.
The native parallel build failed before producing an executable. Consequently
**every requested native thread count 1 through 16 is unmeasured**. Repeating the
build with `ADAMIC_THREADS` would not bypass its source-level refusal. There is
no throughput comparison with typescript-go from this scout.

Commands run from the repository root (scratch driver imports the repository's
Parser and countTree by absolute path):

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs /workspace/scratch/scout-37/parallel.ts /workspace/scratch/scout-37/compiler.manifest
go run ./cmd/adamic build /workspace/scratch/scout-37/parallel.ts -o /workspace/scratch/scout-37/parser-parallel
go run ./cmd/adamic build /workspace/scratch/scout-37/sequential.ts -o /workspace/scratch/scout-37/parser-sequential
/workspace/scratch/scout-37/parser-sequential /workspace/scratch/scout-37/compiler.manifest
go test ./internal/oracle -run '^TestConcurrencyRefusals$/(dispatch|file_read|graph_region)\.a$|^TestGraphParallelMapRefusesRegions$' -count=1 -v
go test ./internal/native -run '^TestGraphValueIsNeverShared$' -count=1 -v
```

All four selected existing refusal checks and the runtime graph-publication
backstop passed. Relative document links resolve, code fences are balanced, and
the whitespace check passed. Scratch sources and logs are
under `/workspace/scratch/scout-37`. The initial scout commit
`057d554cdd7eda55ee8fdd60227904eb3394125f` contained only this document. That
initial pass made no production, fixture or count-table changes. The additions
below supersede its documentation-only scope.

Step 37 has an expressible fork-join structure, but cannot yet run this parser.
The next concrete proof boundary is worker-local constructor/method effects,
followed by ownership transfer and graph publication for retained parse results.
The Program region work must explicitly cover allocator lanes or file-region
adoption and the lifetime of bind/check consumers before parallel parse can use it.

## Expanded research: exact source sites

The [AST census](scout-37-parallel/census.cjs) reads the same 77 source files with
TypeScript 6.0.3 on Node. [sites.json](scout-37-parallel/sites.json) records every
surveyed call's file, line and column, declarations separately, and SHA-256s for
all source files. It also lists direct variable bindings in namespace `Parser`
and function `createTypeChecker`. These are syntactic inventories, not resolved
call targets or escape-analysis roots. Property calls with the same spelling are
included; aliases and other entry names are not. Do not add these counts to the
stage 3 not-yet root count or treat them as independently necessary edits.

Source links below use the pinned TypeScript commit. Counts cover the entire
77-file corpus, not only the linked function body.

| Shape and definition | Call expressions and locations | Why parallelism bites |
| --- | --- | --- |
| [createSourceFile, parser.ts:1344](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/parser.ts#L1344) | 5 named calls: parser.ts:1716,1819,1990,8797; program.ts:410 | The name also denotes a parser-local helper and a factory member; only the program.ts call is the surveyed external host entry. Three definitions have this spelling. |
| [parseSourceFile, parser.ts:1603](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/parser.ts#L1603) | 4 calls: parser.ts:1356,1363,9959,10020 | Normal, JSON and incremental paths initialize/clear shared Parser state; incremental parsing can retain old tree nodes. |
| [parseList, parser.ts:3094](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/parser.ts#L3094) | 11 calls: parser.ts:1814,4381,4433,6191,6831,7018,7026,7037,8201,8236,8294 | Lists mutate a private array but also save/restore the Parser's shared parsingContext and invoke parse callbacks. SourceElements at :1814 is the first reduction. |
| [internIdentifier, parser.ts:2637](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/parser.ts#L2637) | 7 calls: parser.ts:2654,2717,2749,6440,8325,8547,9918 | A shared identifier Map would be mutated by different files. Fresh per-file Maps fit existing rules. |
| [createTypeChecker, checker.ts:1486](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L1486) | 1 call: program.ts:2685 | Program lazily caches one checker; parallel file checks cannot share its scratch state and mutable caches. |
| [getDiagnosticsHelper, program.ts:2778](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/program.ts#L2778) | 3 calls: program.ts:2795,2799,2828 | The whole-program path flatMaps files in order, checks cancellation and sorts/deduplicates. Syntactic and semantic paths have different mutation needs. |
| [checkSourceFileWithEagerDiagnostics, checker.ts:49710](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L49710) | 3 calls: checker.ts:1919,49732,49756 | Changes addLazyDiagnostic around checking; getDiagnosticsWorker can add newly deferred global diagnostics. Per-file results are not independent when this checker is shared. |
| [sortAndDeduplicateDiagnostics, utilitiesPublic.ts:303](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/utilitiesPublic.ts#L303) | 6 calls: program.ts:648,2784,2786,3268,3284; watch.ts:612 | Publication/merge must retain deterministic diagnostics; collection order alone is not its final ordering or deduplication rule. |

The eight spellings yield **40 call expressions**. Namespace Parser has **30**
direct variable bindings, **29** mutable bindings: for example scanner at
parser.ts:1444, fileName :1501, sourceText :1503, currentToken :1511,
identifiers :1513 and parsingContext :1518. `createTypeChecker` has **280**
direct local variable bindings, **279** mutable bindings, including deferred
callbacks :1490, lazy-diagnostic dispatch :1492, cancellationToken :1505 and
scanner :1507. Constness is a binding classification, not a proof that any
referenced container is immutable. These counts exclude nested helpers' locals.

For comparison, inspected typescript-go places the file scheduling boundary at
`tsc/internal/compiler/filesparser.go:273` (`start`), file reading/parsing at
`compiler/host.go:83`, and private parser acquisition at `parser/parser.go:134`.
Checker count and partition construction are at `compiler/checkerpool.go:309`,
exclusive acquisition at :331, and one job per checker group at :476. These are
references within the pinned fork, not claims about every upstream revision.

## Node and specification evidence

On Node, tsc's ordinary function calls and array traversals execute synchronously
on the calling JavaScript agent. Its Parser namespace and cached checker are
therefore reused serially. Adamic's Node witness is
[`oracle/adamic.mjs:137`](../oracle/adamic.mjs), which uses ordinary `items.map((item, index) => work(item, index))`.
No worker pool is hidden in that oracle.

[ECMAScript Array.prototype.map](https://tc39.es/ecma262/multipage/indexed-collections.html#sec-array.prototype.map)
captures length, traverses indices in ascending order, uses HasProperty and calls
the callback for present elements; abrupt completion propagates. Sparse arrays,
getters, proxies, mutation during traversal and observable side effects cannot be
silently treated as independent parse jobs. The first reduction uses dense,
readonly string inputs and effects confined to fresh worker allocations.
[Array.prototype.sort](https://tc39.es/ecma262/multipage/indexed-collections.html#sec-array.prototype.sort)
requires stable sorting when the comparator is consistent. TypeScript additionally
supplies its own diagnostic comparison and deduplication semantics; the independent
reference invokes that actual implementation, rather than inferring it from sort.

[ECMAScript Map.prototype.set](https://tc39.es/ecma262/multipage/keyed-collections.html#sec-map.prototype.set)
and [get](https://tc39.es/ecma262/multipage/keyed-collections.html#sec-map.prototype.get)
define key lookup and replacement semantics; ordinary identifier strings compare
by their value. `__proto__` has no object-prototype setter meaning as a Map key.
[The String type](https://tc39.es/ecma262/multipage/ecmascript-data-types-and-values.html#sec-ecmascript-language-types-string-type)
is a sequence of UTF-16 code units, including surrogate code units. Interning is
not a license to alter identifier spelling or Unicode contents. The reduced
parser's grammar is ASCII-only; the interning fixture separately includes Greek
text and a supplementary-plane identifier.

[Node v24.19.0 worker_threads](https://nodejs.org/download/release/v24.19.0/docs/api/worker_threads.html)
documents CPU work in independent workers, structured cloning of ordinary
workerData/messages, and explicit ArrayBuffer transfer or SharedArrayBuffer
sharing. It does not make one ordinary mutable JS object graph safely usable by
two workers. Cloning ASTs loses shared identity across messages and costs work;
Adamic's internal readonly publication uses its own counted-heap protocol.
Neither ECMAScript nor Node prescribes tsc's parallelization or mandates Adamic's
internal allocator. Observable results and supported object identity remain
requirements; Node's implementation of garbage collection is not a proposed
Adamic memory model.

## Design within the existing rules

The implemented first shape is a narrow SourceElements parser:
`("const " ASCIIIdentifier "=" Decimal ";")*`. It accepts dense readonly source
strings, creates a fresh declaration list and node records in each worker,
records positions, and returns an acyclic readonly SourceFile record. It is a
fixture-sized data-oriented parser, not a replacement for stage 1's whole parser.
Its supported test inputs are projected to the actual tsc AST on Node before
native comparison. There is no constructor, receiver dispatch, mutable capture,
shared intern pool or hidden I/O to adjudicate.

Use the existing counted heap for its objects, arrays, strings and views.
`parallelMap` keeps inputs alive through the join and marks complete result graphs
before publication. String views retain their byte owners; releasing callback
locals cannot release a returned AST's reachable storage early. The parent reads
results only after the join. Destruction after the last counted owner must free
all acyclic results; sanitizer and leak checks verify that obligation. This
first implementation adds no GC, arena adoption, reference-count bypass, unchecked
cast, ownership annotation or new language acceptance rule.

A larger Program-region implementation should keep the per-file allocation and
phase boundaries described above, with private parse scratch and parent-only
ordered assembly. Allocation lanes or per-file chunk adoption remain allocator
choices, not permission to publish mutable or cyclic state. Escaping values must
retain counted ownership or a validated common region owner. The worker's end
cannot be a free point for an AST returned to Program. No graph region crosses
threads until the ownership/publication questions below receive rulings and the
corresponding runtime protocol is proved.

Checking can first use already immutable diagnostic inputs and parent-side
assembly, as the reduction does. Actual semantic checks require private checker
partitions with private caches; calls that still touch shared parser/checker
state stay refused. No persistent checker acquisition, cancellation API, full
source-file parser admission or cyclic AST handoff was implemented by this scout.

## Hard cases and current blockers

- Parser speculation saves/restores token and parsingContext state; JSON, JSX,
  JSDoc and incremental reuse do not have the reduced grammar's simple ownership.
  Returning old nodes requires preserving their original owner as well as new
  allocations. Error recovery and cancellation can interrupt construction.
- AST parent pointers, symbols, captured closures and type caches can form cycles
  and cross-file edges. Per-file regions cannot be dropped independently once
  Program or another file still reaches them. The existing graph-region worker
  refusal remains in place.
- Full stage 1 `new Parser` still fails the task effect proof. Constructors,
  field initializers, base constructors, overridden methods, callbacks and `this`
  aliases need complete effect/ownership evidence; the reduced parser sidesteps
  these shapes without admitting them.
- Global intern pools and lazy checker state are mutable. Sharing a readonly
  facade does not make their backing maps immutable. Weak observers and surviving
  caller aliases prevent an exclusive transfer proof.
- Cancellation/deferred diagnostics can change global results as checking
  proceeds. API identity of nodes/types, deterministic ordering, duplicate
  diagnostics and structured-clone identity loss need separate coverage before a
  checker-pool port. The diagnostic reduction handles only flat string messages
  and fixed file/start/code projections.
- Native tasks may finish in another order or execute later pure work after a
  lower-index task throws. The existing runtime selects the lowest input-index
  exception (`parallel.c:204`); source purity prevents intermediate effects from
  becoming visible. A new cancellation or host-I/O contract is not inferred here.
- Three additional source-form gaps were observed while building reductions:
  string assignment as a call argument, inferred `never[]` in an empty fallback,
  and numeric `||` in a comparator. Splitting the assignment, using a typed empty
  array and explicit numeric comparisons preserve the reference output and need
  no compiler or language change. No claim is made that those forms were fixed.

## Questions for system_adamic

- May the task effect proof admit class constructors and methods on a receiver
  proved wholly worker-private, and what complete-target evidence is required
  for field initializers, base calls, virtual dispatch and escaped callbacks?
- Should a completed mutable AST be transferred exclusively to its parent for
  binding, or must the published AST and all reachable state be readonly at the
  boundary, and how should phase-specific mutable views be represented?
- May a cyclic per-file graph region be transferred at join under a whole-graph
  ownership proof, and does later readonly sharing require an atomic region
  count or another explicitly approved publication model?
- May Program-owned intern pools or checker caches ever be shared across workers,
  or must every parser/checker partition keep independent mutable caches?
- Should structured parsing/checking gain an explicit cancellation operation,
  and what observable exception, partial-result and cleanup behavior must it have
  when several files fail or cancellation races completion?
- Should parallel work ever admit host file reads or dependency discovery with
  observable host effects, and what Node witness would define their ordering?
- Must persistent checker services preserve node/type/symbol handle identity
  across acquisitions and Program versions, and what ownership must an escaped
  handle carry when an old Program or checker is reclaimed?

These questions are not answered by this implementation. The first reduction
uses only the already admitted fork-join, readonly-input and fresh-acyclic-result
shapes; it does not depend on a ruling for any question above.

## Executable work and mutants

[scout37_test.go](../internal/oracle/scout37_test.go) holds each fixture to its
pristine Node source and the JavaScript backend. Native controls and source
mutants run with ASan/UBSan and LeakSanitizer at 1 and 16 threads. Linux controls
also run with TSan, three fresh processes at each thread count. A mutant must
compile, exit successfully, stay leak-clean and match its own Node source before
its differing output is accepted as a rejection by the pristine Node witness.
A build failure, panic or sanitizer failure does not count as killing these
semantic mutants.

| Fixture | Upstream shape and exposed behavior | Mutant rejected by pristine Node |
| --- | --- | --- |
| [scout37_parse_list.a](../internal/oracle/testdata/concurrency/accepted/scout37_parse_list.a) | parser.ts:1814,2594,2600,3094; 128 file jobs, including empty files, exact const-node positions/names/values and retained source-string views in returned ASTs | Add one to every parsed initializer value |
| [scout37_intern.a](../internal/oracle/testdata/concurrency/accepted/scout37_intern.a) | parser.ts:2637; 128 independent identifier Maps, repeated names, `__proto__`, Greek and supplementary characters, temporary string-owner views surviving callback exit | Omit Map insertion; the interned spelling remains but pool size is wrong |
| [scout37_diagnostics.a](../internal/oracle/testdata/concurrency/accepted/scout37_diagnostics.a) | program.ts:2778, utilitiesPublic.ts:303; 96 file partitions, empty files, borrowed immutable diagnostic records, ordered partition assembly, sorting and deduplication | Drop the first diagnostic of each nonempty partition |

These are reductions of tsc's source shapes, not copied full parser/checker
implementations. [reference.cjs](scout-37-parallel/reference.cjs) independently
parses the literal const inputs with tsc, executes the actual upstream
`internIdentifier` body with fresh Maps, and invokes the actual
`sortAndDeduplicateDiagnostics` on projected diagnostic records.
[reference.json](scout-37-parallel/reference.json) records the pin and exact
Node output hashes: 2,610, 2,528 and 1,489 bytes respectively. It validates all
three reductions, including the syntax adaptations, against tsc behavior on Node.
The always-on Go test needs no downloaded TypeScript package; the independent
research replay uses the pinned reader and corpus specified below.

All three fixtures join the existing concurrency fixture discovery, so the main
oracle and count recorder will continue to exercise them. No production compiler
or runtime file changed. The complete stage 1 77-file parallel build remains
blocked at its constructor; neither these fixtures nor the 16-thread sanitizer
runs constitute a whole-parser speed measurement.

## Expansion validation and replay

Setup was rerun with the same GOPROXY and cloud script, and the printed tool env
was sourced. Node v24.19.0, Go 1.27.1 and clang 20.1.8; machine/quota are unchanged
from the initial scout. Setup load before/after was `0.26 / 0.39 / 0.62` and
`5.14 / 4.84 / 2.64`. No performance comparison was attempted during overlapping
validation; load observed during the final focused run included
`6.29 / 3.57 / 2.44`. There are no new thread-scaling measurements for the 77-file parser.

Replay the source inventory and independent references with TypeScript 6.0.3
installed outside the repository and the pinned clean source checkout:

```sh
node docs/scout-37-parallel/census.cjs /workspace/scratch/scout-37/reference/node_modules/typescript/lib/typescript.js /workspace/scratch/typescript-6.0.3
node docs/scout-37-parallel/reference.cjs /workspace/scratch/scout-37/reference/node_modules/typescript/lib/typescript.js /workspace/scratch/typescript-6.0.3
go test ./internal/oracle -run '^TestScout37ShapesAndMutants$' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
```

The three controls, their three semantic mutants, and the pinned Node reference
projections pass. [counts.md](../internal/oracle/counts.md) is regenerated and
committed separately; its only changes are the three new fixture rows. At one
thread, parse-list reports 1,799 allocations and 1,799 frees, interning 1,767 and
1,767, and diagnostics 691 and 691. All three report zero region allocations;
these are counted-heap results, not a Program-region optimization. Document
links, code fences, JavaScript syntax, Go formatting and whitespace are checked.
No whole-package test run was used; the focused test owns the new shapes.
