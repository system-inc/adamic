# Canonical closures in graph regions

Branch: runtime/slice6b-canonical-graph. Base: 66fccddf6eac48fcc3a6a40e4b9c19558fc14f1b.
Date: October 9, 2026. Decision: option (a), adopt first and cache second.

## Behavior and ownership

The exact reported source is canonical_graph.a. It now prints true, matching
Node. canonical_graph_counted.a exercises the counted constructor with optional
arguments and arguments.length, across twenty separate frame lifetimes.

The runtime constructs a fresh closure and retains its captured cells, then
adamic_graph_adopt_owned moves it into prefixed storage and joins its owned
edges. Only afterward does canonical_insert publish its final address in the
owner's weak functions list. Both ordinary and counted graph variants share
this insertion path. Existing non-graph APIs remain unchanged. The emitter
chooses the graph variant from the function's existing GraphClosure flag.

The owner was allocated/adopted before its cells were exposed. Closure adoption
does not move that owner. Captured cell references keep a counted owner alive,
or join the closure to the owner's graph. Cache lookup returns an owned reference
without another adoption. The cache remains weak: it contributes no graph edge
or count. During graph destruction all member storage remains alive while
closures unlink themselves and counted children are released. No union-find,
weak-reference, or destruction ordering change is needed.

The named emission change is emitter.evaluate, ir.MakeClosure, in
internal/native/emit_expressions.go. It replaces the source-reachable panic
with constructor selection. The C contract adds adamic_closure_canonical_graph
and adamic_counted_closure_canonical_graph; both retain the existing typed
_Generic constructor checks. Adoption size is computed from the actual runtime
closure layout and its capture count.

## Prerequisite base repairs

The supplied merge was explicitly not built on the Mac. Two independent
build failures were reproduced before being repaired:

* internal/oracle/counts_test.go declared countsLine twice. The obsolete private
  regexp was removed; the shared leakcheck.CountsLine remains. The baseline
  oracle compile log identifies both declarations at lines 50 and 60.
* Graph adoption already calls adamic_object_size, but the supplied adamic.h
  contained no definition. Compiling map.c archived directly from 66fccddf
  reproduces the undeclared-function failure. The restored helper matches
  object.c allocation: object header, every slot, and two metadata bytes per slot.

These repairs enable the requested proofs; they are not closure speed claims.

## Mutant and positive controls

TestCanonicalGraphCacheOrderMutant snapshots the runtime and moves adoption
below '*functions = closure'. Each mutated binary must compile successfully,
then fail with an explicit ASan heap-use-after-free. Both ordinary and counted
fixtures caught it. The unchanged runtime passed source Node, emitted
JavaScript, sanitized native, release native, slab and leak comparisons.
An initial incorrectly filtered oracle command ran no fixtures; it supplies
no proof and was replaced with the correctly selected uncached run.

## Toolchain and commands

Setup passed: Go 0.025s, Node 0.026s, submodules 0.059s, markdown dependencies
0.081s, clang 0.249s, build 44.910s, cache warm 45.127s, total 45.154s.
Visible CPUs: 5; cgroup quota: 400000/100000. Linux, Go 1.27.1, Node 24.19.0,
clang 20.1.8. Every test shell sources /workspace/adamic-tools/env.sh.
The required Node types were separately installed with npm ci in stage3/api:
@types/node 25.3.3 and TypeScript 6.0.3, from the checked-in lock file.

Test outputs are log files in /workspace/scratch/slice6b-canonical-graph.

```text
go build ./...
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/canonical_graph' -count=1 -parallel=4 -v -timeout=30m
go test ./internal/native -run '^TestCanonicalGraphCacheOrderMutant$' -count=1 -parallel=4 -v -timeout=30m
go vet ./...
gofmt -l cmd internal
go test ./internal/lower ./internal/native -count=1 -parallel=4 -timeout=30m
go test ./internal/lower -count=1 -parallel=4 -timeout=30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/.*(closure|canonical|graph_regions|nested|omitted)' -count=1 -parallel=4 -timeout=30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^Test.*(Closure|Canonical|Graph|Nested).*' -count=1 -parallel=4 -v -timeout=30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=4 -timeout=30m -args -update-counts
```

Observed witness gate: PASS, 40.652s, with both fixtures executed.
Observed isolated mutant gate: PASS, 14.052s, both mutants caught by ASan.
Build and vet succeeded; formatting and diff checks were empty.
Count regeneration passed in 142.364s after dependencies were installed.
The broader source-fixture oracle passed uncached in 67.673s.
The explicit closure/canonical/graph/nested tests ran in 15.863s: all passed
except the pre-existing million-node byte expectation described below.
The complete internal/native package passed in 662.720s, including the cache
order mutants, runtime ownership, concurrency and signal/exit checks.
The regenerated ledger passed its subsequent verification in 168.328s.

Compile flags common to these native lanes: -std=c11 -Wall -Wextra -Werror
-Wcast-function-type-strict -pedantic -Wno-unused-variable
-Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter
-Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -pthread.
Sanitizers add -O1 -g -fsanitize=address,undefined
-fno-sanitize-recover=all. Count measurements use -O2 -DADAMIC_COUNT;
slab checks add -DADAMIC_SLABS. Shipped release uses -O2 -flto=thin.

## Failures and coverage limits

The first package/count attempts exposed missing pinned Node type dependencies.
Those attempts are preserved in packages.log and counts.log; the failed count
attempt returned before writing the table. After npm ci the complete lower
rerun had exactly four failures, each reproduced with baseline source overlays:

* TestNestedCallbackCycleIsRefused: callback cycles now lower using graph regions.
* TestAsyncGapsNameTheMissingPiece: awaited void now lowers.
* TestNamespaceAmbientHostInitialization/cwd: host cwd now lowers.
* TestUncheckableCastsStayRefused/callback_signature: diagnostic classification
  differs from its expected refusal.

Those lowering files and tests are unchanged by this fix. Logs:
base-lower-failures.log, base-namespace-failure.log and lower-final.log.
No whole-repository test gate, WASI lane, or performance improvement is claimed.

## Count ledger

Columns below are allocations/frees/retains/releases/peak/regions, followed by
graph records/merges. Every normally finishing witness is balanced:

| Fixture | Counts | Graph records/merges |
| --- | --- | --- |
| canonical_graph.a | 2/2/5/7/2/0 | 1/1 |
| canonical_graph_counted.a | 40/40/100/140/2/0 | 20/20 |

The exact witness has one live two-member graph: its activation environment and
its canonical closure. The counted variant destroys twenty such graphs.

The ledger on 66fccddf was not regenerated after its merge. Regeneration adds
missing existing-source rows alongside these two new fixtures. The only two
previously recorded fixture rows that move are library_string_raw.a
(retains/releases 50/84 to 55/89) and object_prototype.a (165/288 to 171/294).
Both exact new rows were independently reproduced using baseline emitter and
closure sources, with just the build prerequisites repaired. These movements
are inherited ledger drift, not costs of adopting canonical closures.
base-counts-rows.log records this counterfactual. No existing fixture's
allocation/free/region columns changed in that comparison.

## Pre-existing million-node expectation

TestGraphRegionsCompiledMillion expects payload bytes 64,000,000, reachable
60,800,064 and unreachable 3,199,936. The allocation layout is 70 bytes per node,
including its six per-slot metadata bytes, so the observed totals are
70,000,000, 66,500,070 and 3,499,930. Both the fixed sources and baseline
emitter/closure sources reproduce this same failure after prerequisite build
repairs. The actual allocations/frees are balanced at 1,000,003/1,000,003.
base-million-failure.log preserves that baseline reproduction. The stale
expectation remains unchanged; the complete native runtime million-node test
already asserts the 70-byte layout.
