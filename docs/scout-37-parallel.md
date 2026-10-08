# Scout for roadmap step 37: parallel parsing and checking

Issue: #xfpyaj8. This is source inspection and a scratch build probe, not a
production implementation. The native parallel parser does **not build** at this
base, so there are no 1–16-thread parse measurements or speedup claims.

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

## Validation and measurement limits

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
under `/workspace/scratch/scout-37`; only this document is committed. No production
code, fixture or count table changes were made, and no new mutant was needed for
this scouting-only result.

Step 37 has an expressible fork-join structure, but cannot yet run this parser.
The next concrete proof boundary is worker-local constructor/method effects,
followed by ownership transfer and graph publication for retained parse results.
The Program region work must explicitly cover allocator lanes or file-region
adoption and the lifetime of bind/check consumers before parallel parse can use it.
