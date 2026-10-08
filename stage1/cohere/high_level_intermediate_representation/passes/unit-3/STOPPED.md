# Unit 3 — seven backend passes; memo dropping stopped

Merged the owner's replay framing at `9addb0e8` without rebasing (merge
`198eebb43a7a2078f44b936c4a8a0b5ffb6c59e2`). All changes are lane-owned;
shared replay, graph records, caches and arena classes remain untouched.

## Named oracle failures

The five IncludingMemoCallbacks mismatches are parked individually under
**#6d8y0pf** in `testdata/parked_oracle_failures.json`. Every raw comparison still
runs. An unlisted mismatch fails the test; a repaired named case counts as a
match automatically. `HIR_UNIT3_REQUIRE_CERTIFICATE=1` still fails on any raw
mismatch, including a parked one. No IDs or output bytes are normalized.

Go's allocating map range at `inline_iife.go:379` is confirmed nondeterministic;
the upstream fix is an ordered walk in React Compiler's order plus an audit of
other allocating map ranges. The minimum source and two Go outputs remain in
`testdata/oracle_order_input.ts`, `oracle_order_before.dump`,
`oracle_order_after_a.dump` and `oracle_order_after_b.dump`.
`TestGoInlineOracleOrderGap` reproduces it from 128 identical prepass graphs.
The fixture archive must be regenerated after #6d8y0pf lands before expecting
all five old outputs to match the repaired oracle.

## Raw Go agreement

| Subpass | Node originals | Sanitized native originals | Emitted JavaScript originals | Probes per executed backend |
| --- | ---: | ---: | ---: | ---: |
| OutlineFunctions | 1465/1465 | 1465/1465 | 1465/1465 | 72/72 |
| DropManualMemoization | 1465/1465 | compile stopped | compile stopped | Node 72/72 |
| InlineImmediatelyInvokedFunctionExpressions | 1465/1465 | 1465/1465 | 1465/1465 | 72/72 |
| IncludingMemoCallbacks | 1460/1465 | 1460/1465 | 1460/1465 | 72/72 |
| CopyNestedBodyInto | 1591/1591 | 1591/1591 | 1591/1591 | 75/75 |
| CollectAssumedInvokedFunctions | 1465/1465 | 1465/1465 | 1465/1465 | 72/72 |
| EliminateDeadCode | 1465/1465 | 1465/1465 | 1465/1465 | 72/72 |
| MergeConsecutiveBlocks | 1465/1465 | 1465/1465 | 1465/1465 | 72/72 |

Node: **12420/12425** checkpoints. Sanitized native: **10883/10888**.
Emitted JavaScript: **10883/10888**. Each executed subpass retains all 1465
originals, all 23 Flow graphs and all 72 probes. Remapping additionally covers
126 original and 3 probe nested functions, including 3 extra Flow remaps.
Receipts are `testdata/node_receipt.json`, `native_receipt.json`,
`native_merge_receipt.json`, `native_inliners_receipt.json`,
`javascript_receipt.json` and `javascript_inliners_receipt.json`.
The archive contains actual Go input/output checkpoints, not execution logs.

## Rulings applied and remaining stop

The graph's block type is now the lane-local `BasicBlockInterface` contract.
The shared concrete `BasicBlock` supplies that contract; concrete arena index
classes remain branded and checked. No nominal brand became an interface.
`testdata/nominal_callback_gap.a` is retained as a correct nominal refusal,
not a compiler gap, under `TestNativeNominalContractRefusal`.

Recursive source traversal, syntax traversal and liveness use ordinary module
function declarations with explicit state. Candidate invalidation, dependency
collection and marker allocation likewise use plain declarations. The original
recursive initializer remains the compiler fixture
`testdata/recursive_initializer_gap.a`; no cycle support is requested for it.
SSA algorithms are imported by reference. mutation_aliasing is not needed here
and its API was not copied or invented.

**DropManualMemoization remains compile stopped**, independently of the other
seven passes. Shortest retained entry: `testdata/memo_cycle_refusal.a` (seven
lines). Both native and emitted JavaScript refuse the owner's
`core.ts:144` instruction array, identifying the write in
`replay/decode.ts:163:178`: `Adamic 0.1 refuses Instruction[], an array whose
elements can reach back to an array like it` (`adamic/cycle-capable`).
Its Go entry is `drop_manual_memoization.go:198`.
`TestMemoInstructionCycleRefusal` checks both refusals. The lane has not edited
owner files, weakened ownership, or added object-address/cycle links.
Request to hir-01/compiler: resolve that instruction-array refusal using the
retained entry. Cache integration through memo dropping stays uncertified.
No new record, instruction variant or arena index class is requested.

## Mutants and reproduction

Node catches **7/7** semantic mutants. Sanitized native catches **6/6** executed
mutants (outline, inline, remap, invocation, DCE and merge); memo dropping's native
mutant is compile stopped. Each detection changes an answer after a baseline
matches Go exactly; compiler refusals are not mutant detections.

Source Node, Go adapters, negative fixtures and oracle checks:

```sh
go test -v -count=1 -timeout=3h ./stage1/cohere/high_level_intermediate_representation/passes/unit-3
```

Build the seven available subpasses (ordinary alternative entry; same pass
implementations and replay imports):

```sh
go run ./cmd/adamic build stage1/cohere/high_level_intermediate_representation/passes/unit-3/independent_main.ts -o /tmp/unit3 --sanitize
go run ./cmd/adamic js stage1/cohere/high_level_intermediate_representation/passes/unit-3/independent_main.ts > /tmp/unit3.js
HIR_UNIT3_NATIVE=/tmp/unit3 HIR_UNIT3_PASS_GROUP=available go test -v -count=1 -timeout=3h ./stage1/cohere/high_level_intermediate_representation/passes/unit-3 -run 'TestNode(Subpasses|SemanticMutants)$'
HIR_UNIT3_JAVASCRIPT=/tmp/unit3.js HIR_UNIT3_PASS_GROUP=available go test -v -count=1 -timeout=3h ./stage1/cohere/high_level_intermediate_representation/passes/unit-3 -run '^TestNodeSubpasses$'
```

All new top-level Go tests call `t.Parallel`. No complete eight-pass certificate
or native boxing count is claimed.
