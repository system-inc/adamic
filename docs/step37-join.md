# Step 37: exclusive AST handoff at join

Label: runtime-step37-join. Roadmap #xfpyaj8. Ruling by system_adamic,
October 8: each mutable parsed AST transfers exclusively to the parent; no
worker reference survives. A whole per-file graph region may move under that
proof. Intern pools and caches stay private and merge deterministically.
Parallel checking is NotYet. Program adoption waits for step 06.

## Status and named stop reason

**Whole-graph return inventory is missing.** This branch does not enable cyclic
AST handoff. The starting query, `runtime/ownership-query` at `e98fe2a9`, extracts
snapshots only at moved input boundaries. Its extractor explicitly sets
`Complete=false` for candidate graph allocations because it has no dynamic region
inventory. It does not inventory worker roots after callback cleanup or produce
return-boundary transfer plans. A complete hand-written snapshot can prove a ring;
that is not evidence that a particular parser return owns that ring exclusively.

The branch starts at the fetched `origin/area/runtime` head `9bc8201f` and merges
the named query and `runtime/scout-37-parallel` at `378971fb`. The query remains
off by default. `runtime/program-region-prototype` at `ff91354d` is not merged.

A reduced parser with strong child-to-file parent pointers built on this base,
but native exited 70 at `adamic_share` while source Node completed normally.
The new post-classification boundary guard refuses results whose reachable type
or captured environment contains a graph member. It uses the existing cycle
finder's reachability, structural views and closure edges, after `graphFlows`.
Unresolved generic result frontiers conservatively refuse in graph-bearing
programs; no task boundary means no additional source walk. It detects demand only; it supplies no ownership certificate. Refusal holds with
the query flag both off and on, before native or JavaScript emission. Existing
worker-local graphs returning scalar summaries remain admitted.

The positive parser control returns a **mutable acyclic** AST and the parent
changes declaration values after join. Its existing shared publication gives
atomic counts; it is not an exclusive cyclic graph transfer. The other control
constructs a fresh intern Map and cache per file invocation, then merges ordered
pool projections on the parent. No production tsc parser or checker was ported.

## Proof required before enabling handoff

Extend `internal/fresh`'s existing abstract heap, rather than add another source
heap analysis. Query each possible callback return after evaluating its result
and all normal-exit cleanup, including finally overrides. Source declarations
and compiler temporaries that own counts must have explicit ownership endings.
Local liveness alone does not end a count. A surviving alias that is never read
still refuses. Keep exception exits distinct from successful result publication.

The snapshot must inventory allocation identities, labelled strong and weak
edges, dynamic union membership and every outside root: locals, parameters,
globals, task environments and capture cells, parser reuse pools, caches, pending
callbacks, nested tasks and borrowed references. Resolve interior environment
cells to their whole owner. Summary fallback, unknown targets, escaped/older
allocation instances, uncertain region unions and budget exhaustion yield
Unknown. Proven is the only verdict that permits transfer; Refused and Unknown
both stop at the parallelMap boundary with their evidence.

Compute closure to a fixed point over strong mutable reachability and whole
region membership. Overwritten edges never undo a region union. Include the
region's disconnected members and their counted descendants. Internal aliases
and cycles are allowed. Any surviving worker root or Weak observer into that
closure refuses, including one reaching only a disconnected member. Regions
from different file tasks must be disjoint. A parser cache can die before the
boundary; a worker reuse cache retaining a node cannot. A returned graph rooted
in a borrowed reference has no outside count to transfer.

Immutable source strings may remain shared under the existing publication
protocol. Their owners and lazy caches stay on the atomic shared path. The
return plan must distinguish these leaves from transferred mutable storage;
sharing a readonly facade over mutable backing storage is not sufficient.
Do not reset any existing shared bit. Prove cleanup drops worker administrative
counts while preserving the return owner's outside count, without using a
runtime `count == 1` as a substitute for the static inventory.

Record the result decision explicitly in typed IR, keyed by callback function,
source call site and CFG return instruction. The current `ParallelMap.Moved`
means an input move and must not authorize an independently allocated result.
Use a separate result disposition and a cleanup plan. Evaluate all callback
and helper targets from CallTargets/ClosureTargets. An unresolved receiver or
callback refuses. Borrow, reuse and exception cleanup must consume the same
ownership decision. No source ownership annotation is introduced.

## Runtime handoff protocol

The existing scheduler mutex already orders a completed range before parent
join return. Keep that happens-before edge; do not publish completion until all
callback cleanup and result-slot writes finish.

| State | Owner and allowed operations |
| --- | --- |
| Building | One task owns the graph and may mutate/merge its plain regions. Other executors cannot inspect them. |
| Sealed result | Callback cleanup has ended every worker owner. The result slot holds the transferred outside count. Worker scratch retains no pointer into it. |
| Joined | Parent observed completion under the scheduler mutex and consumes the slot's ownership. Parent may bind/mutate the AST. |
| Abandoned | After all relevant tasks finish, the joining executor releases completed outputs and errors exactly once. |

For a Proven result, skip `adamic_share` on the transferred mutable closure.
Do not retain the graph on the parent while its task still runs. No extra count
is necessary to move the slot's ownership; internal graph edges remain uncounted.
Workers must not access a sealed result again, including during thread shutdown
or reuse-pool cleanup. The existing allocator's cross-thread free path handles
storage freed by its new owner; graph union-find/count operations remain plain
because access is sequential across the join, never concurrent.

On any task exception, preserve the existing lowest-input-index exception rule,
join all work and release every completed result. A task that throws cleans its
private graph; a successful task result stays owned until the parent releases
it. Nested joins need the same proof and ownership endings. No cancellation or
parallel checking contract is added.

## Deterministic intern and cache merge

Use a private pool/cache per parse invocation. Persistent OS-worker caches need
a separate lifecycle and escape proof; thread assignment is not a stable merge
key. The control returns only immutable ordered names and local-use IDs after
its Maps die. The parent traverses results in input file order, then names in
that file's insertion order. It assigns a canonical ID at first occurrence and
rewrites each file's local IDs through an explicit remap. Unicode strings and
`__proto__` remain ordinary value keys. Completion order, worker number, hash
bucket order and thread count cannot select canonical IDs.

Actual parser caches must supply a deterministic semantic merge or be dropped
at join. Maps keyed by node identity cannot merge as string intern tables do;
a surviving pointer into an AST must be included in the handed-off closure or
ended before transfer. Never silently keep such a cache on a worker.

## Program region seam, design only

`ff91354d` has an opt-in provisional membership census and a single static
Program member list. That list is not a worker allocator or a graph handoff API.
Its fresh-allocation adoption may change an address, so calling it on completed
AST nodes at join would invalidate established edges. Do not use it as a cast,
do not allocate concurrently into its global list, and do not build adoption in
this unit.

After step 06 lands, add a parent-only adoption operation accepting an owned
per-file descriptor and its compiler certificate. It must preserve all node
addresses, transfer ownership of complete region records and counted boundary
children, and consume the descriptor exactly once. Until then, a Program-side
outside anchor can retain a joined graph's existing storage; it is not a Program
region optimization. Keep per-file anchors alive through binding/checking and
release only after their last consumer, including escaped handles, ends.

The seam must account for later cross-file links: parent-only graph merges are
permanent; dropping one file cannot invalidate another file's symbol/backlink.
A Program descriptor needs membership/lifetime provenance, not type SCC IDs.
Program teardown invalidates Weak handles, releases external counted children
while all members remain alive, then frees regions once. Partial parse failure
and partial adoption must have one owner for every completed file. These are
requirements for integration with step 06, not implemented APIs.

## Executable evidence

`TestStep37JoinInventory` checks supplied complete snapshots for an exclusive
cyclic region, worker cache ingress to a child or disconnected member, a Weak
observer and incomplete inventory. `TestStep37JoinExtractionIsIncomplete` checks
the actual IR input extractor returns Unknown for graph allocation demand.
These are analysis controls, not a completed return extractor.

`TestStep37JoinRefusesUnprovenGraph` runs the reduced cyclic parser on Node, then
requires the new boundary refusal with the query both off and on. A source mutant
that stores a node into a worker-visible global is caught by the existing task
effect proof, despite identical successful Node output.

`TestStep37JoinControlsAndMutants` holds mutable acyclic parsing and ordered
intern merging to source Node and emitted JavaScript. Native controls and
semantic mutants run at 1, 4 and 16 threads under ASan/UBSan and internal/leakcheck.
Linux controls run TSan three times per size. The merge mutant reverses file
pools, representing a different completion order; it must compile, match its own
Node execution and finish leak-clean, then differ from the pristine Node output.
The binder mutant omits the parent's update under the same conditions.

The durable runner in `internal/oracle/testdata/step37/prove.py` uses isolated Go
overlays to remove the graph boundary guard, drop the second-owner check, or
ignore inventory completeness. Build errors do not count as killing a mutant.
Validation commands, actual outputs and limits are recorded alongside this file
in [step37-join/report.md](step37-join/report.md).
