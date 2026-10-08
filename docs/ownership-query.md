# Ownership query for inferred moves

Status: design and test-only prototype. No new move is enabled. The production
flat-literal proof in `internal/lower/moves.go`, task/result restrictions, native
ABI and graph-region refusals stay as they are. This extends the future-work
question in [concurrency-moves.md](concurrency-moves.md), not its approved slice.

## Contract at the call

For an evaluated value and a particular `parallelMap` transfer point, answer:

1. Which mutable allocations are strongly reachable from the value? Which whole
   graph regions must accompany those allocations, including disconnected members
   retained by past merges, and what counted descendants do those members own?
2. Every way another root can reach any of those allocations: locals, parameters,
   globals, fields/container slots, closure environments/capture cells, borrowed
   references and Weak observers. Include aliases to descendants, not just to the
   root. Record the candidate binding's own path too.
3. Whether each such path can be used after the transfer, on normal, exceptional,
   loop-back or callback paths. Record a source-bearing witness for a possible
   later use. Results are new owners; their accesses are allowed only after join.
4. Is the candidate an owned, private graph with no other surviving root or Weak
   observer, no later source access, and no overlap with another task's graph?
   Return Proven, Refused or Unknown with evidence, never a guessed alias.

“Every path” cannot mean a list of strings: a ring admits infinitely many walks.
Return a finite labelled may-reach graph, root ingress edges and continuation-use
facts; its transitive closure denotes all paths. Provide a deterministic shortest
witness per failing root for diagnostics, and retain the full graph for auditing.
A may-path suffices to refuse. Unresolved information produces Unknown, not an
empty list of other owners. Track strong ownership separately from weak observation.
Immortal scalar leaves require no ownership. Mixing other Shareable descendants
with moved mutable objects still needs the publication decision deferred in part 2.

Uniqueness and liveness are separate. A confined object can have two locals; an
unused alias can still own a count or reside in an uncalled closure. The approved
rule refuses any second surviving owner, even one with no later read. Only a
proven earlier ownership end removes a root. Do not silently consume dead aliases
as a group. A borrowed root has no count to transfer. The private input array shell
is an explicitly accounted administrative owner through join, not a second task.

## Derivation from lowering's existing information

Run a future query after typed IR and cycle/graph classification are complete;
`parallelMoves` currently runs earlier and cannot by itself inspect that final
state. Identify the boundary by function/instruction and source site, rather than
AST position. Evaluate items and work first: their evaluation may throw, capture
or alias an item. Query the state just before successful dispatch. A future explicit
IR ownership decision must coordinate borrow, reuse, exception cleanup and source
consumption; this patch adds no such decision.

Reuse `flow.Build`'s exception-aware control flow, reaching values/SSA and alias
inference. Assign/Alias propagate identities; CreateFrom requires the operation's
actual heap semantics rather than assuming its descendants are new; Capture
expresses information flow and must be interpreted using the actual captured
storage, not treated as an unconditional identity alias. MaybeAlias introduces
an uncertain relation. Fields, array elements, Map keys/values and Set elements
need labelled heap edges, not just variable ranges. Preserve source sites along
edges. Reassigning a binding kills its old value on that path; joins union possible
values. A branch that aliases on only one path still prevents unconditional proof.

Extend `internal/fresh`'s abstract interpretation, rather than creating a second
heap model. It already has locals, strong/weak heap edges, escaped/leaked sets,
newest/older allocation recency, joins and call summaries over flow graphs.
`fresh.ProveWrites` currently exports write verdicts, not call-site heaps or
ownership roots. Add a read-only snapshot/query facility in future work; retain
all roots and storage ownership endings, including compiler temporaries. Confinement
and a Proven write are supporting facts, never certificates of uniqueness.
A locally constructed ring's closing write is not acyclic, yet its external
ownership may be unique; rejecting that write is not the ownership answer.

Use `Program.CallTargets` for all direct/virtual targets and `ClosureTargets` for
function-value calls. Join summaries for every possible target. Do not choose a
compatible signature or just Call.Function: graph_flow.go's override repair
shows why that misses allocations. Existing fresh summaries conservatively let
arguments escape and summarize fresh results. General borrowed/non-retaining helper
acceptance would require new argument/escape/alias/effect summaries; until then
such arguments remain escaped. Unknown closure targets, summary fallback after
`summaryRounds` (32), missing operations and opaque effects yield Unknown.

`flow.LiveOut` provides declaration liveness for tracked locals. Its own contract
excludes globals and captured variables: treat those as always observable unless
separate storage proofs establish otherwise. Combine reaching values with liveness
so an old value is not confused with a new assignment to the same declaration.
Follow continuation CFG edges for source-use witnesses; do not compare source
positions or EvaluationOrder alone. Respect MayThrow's pre-definition handler
state and distinguish failure before dispatch (no move) from task failure after
transfer (consumed source, outputs/input cleanup). Hidden field/global/closure
uses require heap and call summaries, not local liveness alone.

The cycle finder's type graph supplies the complete possible slot vocabulary,
structural views, generic instantiations, closure cells and strong paths connecting
cycle-capable edges. Its SCC classification and Program.GraphTypes are a conservative
representation classification, not an object identity or region membership proof.
Graph allocation IDs from graph_flow.go identify producers. In the heap interpreter,
track possible dynamic region unions for graph-to-graph stores and initializers.
Do not undo a union on overwrite. At joins preserve possible co-membership;
uncertain or older/merged allocation instances refuse rather than treating one
allocation site as one runtime object. Weak edges do not merge regions, but an
external Weak observer still blocks moving the object it can read.

Compute the transfer closure by alternating strong reachability and whole-region
membership to a fixed point. Build reverse heap edges and root ingress indices;
walk backwards from the closure to find all external roots. This catches a counted
wrapper retaining a child, an environment interior cell (resolve its owner), and
an alias to a now-disconnected member of a merged region. Compare each item's
closure for overlap: `[value, value]`, distinct outer objects sharing a child and
items occupying the same indivisible region all refuse. Internal aliases and
cycles within one task's complete region are harmless to exclusivity.

## Region outside count

`internal/native/runtime/graph_regions.c` resolves interior cells to environment
owners, unions roots and sums their outside references before publishing stores.
Internal graph links own no outside count. Owned locals, kept parameters and
counted fields/containers/captures do. A standalone member uses its heap header
until a region record is required; `find` resolves merged records.

A private ring with one external owner can therefore have outside count one
regardless of its internal edges. Another local/field into any member adds another
outside count. One is a necessary boundary-accounting check for a single incoming
owner, not a static proof: it cannot detect Weak, tell which root owns the count,
prove task disjointness or account for outside counted descendants. Evaluation
and shell temporaries can legitimately add administrative counts. A future runtime
check would need an explicit expected count/transfer plan, not blindly `== 1`.
This design uses static ownership and never speculates by reading plain counts
while other threads may access them. The runtime does not expose a transfer query
or safe graph task protocol today. A unique query result for a ring is evidence
for future design, not permission to bypass today's graph-region refusal.

## Refusal messages

Emit at the task boundary with source location, a proved path or unknown frontier,
and exactly one approved fix. Proposed general-query messages below preserve the
part-2 vocabulary; current narrower diagnostics remain unchanged.

| Condition | Message | Fix |
| --- | --- | --- |
| Later source or alias access | `use after move: alias.next; items was moved into parallelMap` | `don't use it after the parallelMap` |
| Second surviving owner, including dead-but-owned alias | `cannot move items[0].child: another variable, field or closure may reach the graph via other.saved` | `return it through the results` |
| Weak observer | `cannot move items[0]: weak observer weak may reach the graph` | `return it through the results` |
| Borrowed input | `cannot move items: borrowed owner has no ownership to transfer` | `return it through the results` |
| Overlapping task graphs or merged region | `cannot move items[1]: task graph overlaps items[0] via items[1].child` | `return it through the results` |
| Unknown/escaped heap, recency merge, uncertain region, opaque call, summary fallback or budget exhausted | `cannot move items[0].child: whole reachable ownership is not proven at helper.result` | `return it through the results` |
| Incomplete region inventory | `cannot move items[0]: whole region ownership is not proven` | `return it through the results` |
| Callback/result proof fails | `cannot move work.result: return the owned item` | `return it through the results` |

A precise alias message requires an actual may-reach witness. Unknown frontiers
must not invent an alias or imply a proved use. Even a positive input query cannot
waive callback capture/effect/result proofs or graph runtime restrictions.

## Test-only prototype and answers

[ownership_query_test.go](../internal/lower/ownership_query_test.go) consumes an
explicit normalized snapshot of objects, labelled strong edges, region IDs and
surviving roots with after-call flags. It closes over whole regions and descendants,
then searches every root for an ingress witness. Snapshots are hand-built proof
inputs, not extracted from source or production fresh states; flags assume the
future continuation analysis. Missing candidate objects and opaque outside roots yield Unknown. Complete
root/edge inventories and exact region IDs are preconditions, not proved here.

| Shape | Closure | External paths (used after?) | Unique |
| --- | --- | --- | --- |
| Flat literal | 1 | item (no) | yes |
| Nested object | 1, 2 | item (no) | yes |
| Array of distinct objects | 1, 2, 3 | items (no) | yes |
| Locally built cyclic ring | 1, 2, 3 | ring (no) | yes |
| Ring with second outside reference | 1, 2 | ring (no), other.saved (yes) | no |

The flat row isolates one of today's flat items; the array row models the current
whole-input construction proof. Ring-building locals other than ring must already
have ended ownership in that snapshot. Leaving any one alive refuses. The fifth
row reports `cannot move ring: another variable, field or closure may reach the
graph via other.saved`. Additional tests refuse dead aliases, Weak, borrowed owners,
later source use, unknown objects, opaque outside roots and aliases to disconnected region members.
Prototype refusal text is deliberately minimal; the table above specifies future
boundary diagnostics. It does not implement per-item partitions or full path graph
export: snapshot edges already supply the finite representation, while Paths are
shortest ingress witnesses only.

Reproduce answers with the setup env sourced and Node v24.19.0:

```sh
go test ./internal/lower -run '^TestOwnershipQuery' -v -count=1
```

## Cost and limits

Reuse flow graphs, fresh fixed-point states, call summaries and type classification
once per compilation rather than re-running ProveWrites for each task. For a
snapshot with V objects, E heap edges, R roots and M region membership entries,
indexed forward/reverse traversal and region expansion cost O(V+E+R+M) time and
space per boundary; disjointness can use a member-to-task index. Shortest witnesses
need predecessor links, not copied full strings. Do not enumerate all path strings.
The small prototype instead scans region membership repeatedly and does BFS per
root: worst-case O(V² + R(V+E)), with path-string copying overhead. It is a semantic
sketch, not the intended production data structure.

Whole-program preparation has additional CFG iteration, heap-state joins and
interprocedural summary convergence costs; no linear bound is claimed for it.
Use recency/site bounds, sparse query demand, bounded summaries and explicit
Unknown on work-budget exhaustion. Do not cap work by silently omitting roots.

The prior graph-walk report selected the 9-file parser (155 functions/784
locals), 17-file lint entry (257/1,358) and 31-file checker volume suite (359/1,859)
as the largest available stage-1 ports. The current generated lint registry
can select more rule implementations; benchmark logs record current IR sizes.
The original 78-file tsc corpus is absent. Their manifests and prior site-walk
costs are in [the graph walk report](../internal/lower/performance/graph-walk/REPORT.md).
That report's 9.29/15.08/16.54ms walks are not ownership-query timings.

`BenchmarkOwnershipQueryPreparation` measures actual flow.Build plus LiveOut for
all functions and main in those lowered ports. Loading/checking/lowering are outside
the timer. It rebuilds rather than reuses graphs, so it describes preparation when
uncached. It excludes snapshot extraction, type graph work, fresh summary/state
retention and the final ownership query. Those costs remain unmeasured until the
facility exists; no full-query overhead claim follows from these measurements.

Measured on base `ae57b2d84bc8ee572830a49f4e9a7f78737f7205` with this test-only
benchmark: Go 1.27.1 linux/amd64, Node v24.19.0, AMD EPYC 9V74, GOMAXPROCS 5,
four-CPU quota. Three samples of ten preparations, milliseconds per preparation:

| Entry | Current functions / locals | All ms samples | Median ms | Go bytes/op range | Go allocs/op |
| --- | --- | --- | ---: | --- | --- |
| parser/main.ts | 183 / 856 | 16.351, 45.254, 12.706 | 16.351 | 3,266,025–3,266,056 | 50,981 |
| lint/main.ts | 760 / 3,215 | 64.671, 37.872, 30.056 | 37.872 | 10,244,792–10,244,819 | 155,602 |
| typeaware/volume_suite.ts | 387 / 1,931 | 18.169, 21.371, 20.397 | 20.397 | 5,838,504–5,838,561 | 92,048–92,049 |

These are shared-machine observations with wide ranges, not speedup claims or
peak RSS. Setup had finished; a final focused test/vet invocation overlapped the
start of the parser run. The older report's type/file totals describe its own
revision, not this expanded registry or today's IR. Initial benchmark attempts
without the generated registry and TSGo option failed at load/lower and supplied
no lint/checker measurements. Earlier parser samples during cold setup are excluded.

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry > /tmp/ownership-query-registry.log
for input in ../../stage1/typescript/parser/main.ts ../../stage1/cohere/lint/main.ts ../../stage1/cohere/typeaware/volume_suite.ts; do
  ADAMIC_OWNERSHIP_BENCH_SOURCE="$input" ADAMIC_OWNERSHIP_BENCH_TSGO=1 \
    go test ./internal/lower -run '^$' -bench '^BenchmarkOwnershipQueryPreparation$' -benchtime=10x -count=3
done
```

Validation: all `internal/lower` tests passed (49.492s); final focused ownership
checks passed, including the opaque-root refusal and exact second-owner message;
`go vet ./internal/lower`, gofmt and `git diff --check` passed. Setup's repository
build passed. The whole repository gate, native transfer experiments and Darwin
were not run for this design-only patch.
