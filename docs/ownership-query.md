# Ownership query for inferred moves

Status: default-off query implementation behind the existing move check. No new move is enabled. The production
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
ownership roots. The read-only snapshot/query slice below begins this facility; complete dynamic
region export remains future work. Retain all roots and storage ownership endings,
including compiler temporaries. Confinement
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

## Implementation and gate

`internal/fresh/ownership.go` exports `QueryOwnership`, a read-only query over an
explicit complete snapshot, and `OwnershipTransfers`, which extracts call-site
snapshots from the existing fresh interpreter. Answers are **Proven**, **Refused**
or **Unknown**, with the retained finite may-reach graph, closure members,
deterministic shortest ingress witnesses and evidence. Root strength and edge
strength distinguish strong ownership from weak observation. A second surviving
owner refuses even if never read again. A missing inventory, opaque frontier,
escaped or older allocation, summary fallback or exhausted budget cannot prove
exclusivity. Proven ownership endings are represented by reaching-state assignment,
not declaration liveness. Continuation witnesses follow CFG edges and stop at
successful redefinitions; throwing definitions retain the handler's old value.

`lower.Options.OwnershipQuery` or `ADAMIC_OWNERSHIP_QUERY=1` enables the additional
check. It runs after typed IR and graph classification and before borrow/reuse.
Production still runs `moves.go` first, including its callback and result checks;
no query answer bypasses those or enables a graph-region transfer. Preparation is
demanded only when an actual admitted move exists. Every query boundary retains
its CFG, IR instruction, statement and source site; known allocations retain their
producer instructions and statements.

The extraction slice is intentionally conservative: fresh does not yet export a
complete dynamic region inventory, so candidate-reachable graph allocations make
the extracted answer Unknown. The general snapshot query does close over complete
region inventories supplied to it, including disconnected members and their counted
descendants. Type SCCs are never used as dynamic region IDs. The existing flat
construction proof still supplies task disjointness and the administrative input
shell accounting; this patch adds no general per-item partition or transfer protocol.
Nested and cyclic inputs retain their existing production refusals. The normalized
ring proofs are evidence for the future protocol, not authorization to move rings.

`TestOwnershipQueryMoveAgreement` runs every existing concurrency and moves fixture
in accepted/refused directories with the old proof and an independent query path.
Its private harness bypasses only the old source-reference scan; all construction,
callback and runtime admission restrictions remain prerequisites. Any acceptance
disagreement fails. Production never uses that bypass. The test also independently
queries every admitted IR transfer and requires Proven. There are 74 fixtures,
one existing applicable admitted transfer and four independently queried refusals. A real disagreement found during this
work was fixed in extraction: fieldless descendants must be inventoried from heap
edge targets, even though they have no entries of their own in fresh's heap map.
There is no tie-break between proofs.

| Shape | Complete snapshot query | Node witness | Runtime move admission |
| --- | --- | --- | --- |
| Flat literal | Proven | `2` | Admitted; TSan checked |
| Array of distinct objects | Proven | `2`, `3`, `4` | Admitted; TSan checked |
| Nested object | Proven | `2` | Existing refusal retained |
| Local cyclic ring | Proven | `2 2` | Existing refusal retained |
| Ring with a second outside reference | Refused, via `other.saved`, even without a later read | `2 2` | Existing refusal retained |

The source fixtures live in `internal/oracle/testdata/ownership-query`. All five
are checked against Node v24.19.0. Every admitted move, including the existing
2,048-object fixture, runs the existing ASan, release, slab and ThreadSanitizer
variants with one thread and the default pool. TSan variants run three times per
pool choice. The snapshot shape tests additionally check closure membership and
witnesses. Extracted-state tests cover fieldless descendants, dead-but-owned aliases,
proven earlier ownership endings, Weak and continuation access. Additional tests
cover missing/opaque inventories, borrowed roots and disconnected region members.

`prove-mutant.py` deletes the second-owner check temporarily, requires the ownership
regression to fail for the intended reason, and restores the exact original source
in a finally block. It kills the mutant that incorrectly returns Proven for an
unused surviving alias.

Reproduce with the setup environment sourced and Node v24.19.0:

```sh
go test ./internal/fresh -run '^TestOwnershipQuery' -v -count=1
go test ./internal/lower -run '^TestOwnershipQueryMoveAgreement$' -v -count=1
go test ./internal/oracle -run '^TestOwnershipQueryNodeShapes$' -v -count=1 -timeout 20m
python3 internal/oracle/testdata/ownership-query/prove-mutant.py
```

## Preparation cost

`BenchmarkOwnershipQueryLowering` measures actual lowering with and without the
flag on the three largest available stage-1 ports. Loading/checking through
`load.Load`, registry generation and selecting TSGo are outside the timer; each
sample uses a freshly loaded input. Type queries performed by lowering, typed IR,
cycle classification, the optional demand scan and borrow/counters are inside.
Best of three single-lowering samples, milliseconds:

| Port | Functions / locals | Flag off samples (ms) | Flag on samples (ms) | Best off (ms) | Best on (ms) | Difference (ms) |
| --- | --- | --- | --- | ---: | ---: | ---: |
| parser/main.ts | 183 / 856 | 423.719, 414.416, 416.132 | 422.047, 425.699, 425.990 | 414.416 | 422.047 | +7.630 |
| lint/main.ts | 760 / 3,215 | 2520.704, 2462.877, 3097.172 | 2923.960, 3029.188, 2823.242 | 2462.877 | 2823.242 | +360.365 |
| typeaware/volume_suite.ts | 387 / 1,931 | 1339.138, 1102.131, 1079.171 | 1076.490, 978.388, 972.379 | 1079.171 | 972.379 | -106.792 |

Machine: Linux/amd64, Intel Xeon Platinum 8573C, five visible CPUs,
four-CPU cgroup quota, 16 GiB memory limit, Go 1.27.1, GOMAXPROCS 5,
Node v24.19.0, clang 20.1.8. No other task test/build was intentionally running
during measurements. Shared-machine 1/5/15-minute load before/after: parser
`0.27/1.58/2.31` → `0.38/1.56/2.29`; lint `0.38/1.56/2.29` →
`0.64/1.54/2.27`; checker `0.64/1.54/2.27` → `0.70/1.52/2.26`.

These ports have no admitted move boundaries, so the flag performs demand detection
and avoids heap preparation entirely. The table measures that actual lowering path,
not the cost of region extraction or of a demanded whole-program heap analysis.
Negative differences are sample variability, not a claimed speedup. Full demanded
preparation inherits fresh's CFG and interprocedural fixed-point costs; no linear
bound is claimed for it. The snapshot query indexes regions and reverse heap edges,
expands each region once and tracks two reverse-reachability states per object.
Witnesses use predecessor links. Sorting ensures deterministic output, and materializing
witness strings adds their output length to traversal cost. No cyclic walks are
enumerated. The snapshot budget returns Unknown rather than dropping roots.

Raw samples, machine/load data and gate logs are in
[the ownership-query evidence](../cloud/reports/ownership-query/).

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
for input in ../../stage1/typescript/parser/main.ts ../../stage1/cohere/lint/main.ts ../../stage1/cohere/typeaware/volume_suite.ts; do
  ADAMIC_OWNERSHIP_BENCH_SOURCE="$input" go test ./internal/lower -run '^$' \
    -bench '^BenchmarkOwnershipQueryLowering$' -benchtime=1x -count=3
done
```

Validation: all tests in `internal/ir`, `internal/flow`, `internal/fresh` and
`internal/lower` passed; the independent 74-fixture agreement gate passed; all
five Node witnesses and all admitted move sanitizer variants passed; the
second-owner mutant was killed in both normalized and extracted-state tests.
`go vet` on the affected analysis/lowering/oracle packages, gofmt and
`git diff --check` passed. The full repository gate and Darwin were not run.
