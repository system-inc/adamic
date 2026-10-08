# Unit 3 — implementation present; full certificate stopped

Merged `9addb0e8` from `origin/stage1-hir/wip` without rebasing (merge
`198eebb43a7a2078f44b936c4a8a0b5ffb6c59e2`). The earlier shared requests are
answered. All implementation, adapters, fixtures, tests and mutants written by
this lane remain in `passes/unit-3/`; no owner-owned file was edited.

## Raw byte agreement, with every input retained

| Go subpass | Node original checkpoints | Node probes | Sanitized native | Emitted JavaScript |
| --- | ---: | ---: | ---: | ---: |
| OutlineFunctions | 1,465 / 1,465 | 72 / 72 | 0 executed | 0 executed |
| DropManualMemoization | 1,465 / 1,465 | 72 / 72 | 0 executed | 0 executed |
| InlineImmediatelyInvokedFunctionExpressions | 1,465 / 1,465 | 72 / 72 | 0 executed | 0 executed |
| IncludingMemoCallbacks | 1,460 / 1,465 | 72 / 72 | 0 executed | 0 executed |
| CopyNestedBodyInto | 1,591 / 1,591 | 75 / 75 | 0 executed | 0 executed |
| CollectAssumedInvokedFunctions | 1,465 / 1,465 | 72 / 72 | 0 executed | 0 executed |
| EliminateDeadCode | 1,465 / 1,465 | 72 / 72 | 0 executed | 0 executed |
| MergeConsecutiveBlocks | 1,465 / 1,465 | 72 / 72 | 0 executed | 0 executed |

Every subpass includes all 1,465 originals, all 23 Flow graphs, and 72 separately
identified probes. Remapping additionally captures every remaining nested function:
126 original and 3 probe remap inputs beyond the primary record, including 3 more
Flow remaps. Absent nested functions and unsuccessful remaps remain outcomes.
Overall Node: **12,420 / 12,425** raw checkpoint comparisons match. The five
inclusive-inliner mismatches remain failures, never filtered or renumbered.
`testdata/node_receipt.json` records separate per-pass counts and first failures.
These observations are not a complete three-backend certificate.

`testdata/fixtures.json.gz` contains real Go before/after checkpoint bundles, not
compressed execution logs. The adapter uses the owner's encoder and construction
observation hook. Preservation's drop state is after Go outline and InferReactive;
DCE's input retains the actual conditional merge. Each nested remap uses a fresh
parent. AST cooked identifier text is a lane-owned syntax input; no hook answer or
expected analysis result is used as input. SSA maintenance is imported by reference.
This lane does not require mutation_aliasing, invent its API, or copy its algorithm.

## Exact stops and shortest inputs

1. **Constructed nominal return in a generic graph callback.** Go
   `graph.go:153-161` returns its newly allocated BasicBlock placeholder. The
   corresponding `graph_adapter.ts:15` is refused with:
   `Adamic 0.1 refuses a value without nominal ancestry seen as BasicBlock;
   construct that class or a subclass; use an interface for structural values
   (adamic/nominal-class)`.
   Shortest retained input: `testdata/nominal_callback_gap.a` (five lines).
   Node prints `1`; native refuses the actual `new Block` return. The full emitted
   JavaScript compiler is refused at the same callback. No brand was weakened.
2. **Local recursive closure initializer.** Go's lane-owned syntax encoder recurses
   at `testdata/census_adapter_test.go:35`; its Adamic counterpart is
   `facts.ts:24`. Native refuses it with:
   `stage 0 can't lower a function value that captures the variable its own
   initializer declares yet`.
   Shortest retained input: `testdata/recursive_initializer_gap.a` (five lines).
   Node prints `0`. No closure or graph representation workaround was applied.
3. **Go's inliner has no single byte output for the same state.** Production
   `cohere/internal/lint/ecmascript/high_level_intermediate_representation/inline_iife.go:379`
   ranges over `remap.Blocks` while allocating return-rewrite instructions and
   identifiers. `TestGoInlineOracleOrderGap` proves **two distinct byte outputs
   from 128 identical prepass graphs**. The minimum source is
   `testdata/oracle_order_input.ts`; the unchanged input and two outputs are
   `oracle_order_before.dump`, `oracle_order_after_a.dump`, and
   `oracle_order_after_b.dump`. No Go production file or oracle output was changed.

Requests to relay: compiler support for the two retained repros; a deterministic
Go traversal/allocation order at `inline_iife.go:379`. No additional shared record,
variant or arena class is requested from hir-01. Native cache integration remains
uncertified with the same program; its lane callbacks import the supplied public
cache hook without copying keying or construction.

## Local checks and semantic mutants

Seven semantic mutants change outlining, memo finish values, IIFE candidates,
capture remapping, invocation facts, DCE retention and merge selection. **Node
7 / 7 caught with changed answers and exact-byte baseline witnesses; native
0 executed, 0 caught** because lowering is stopped. Compiler-refusal tests are
not counted as native semantic mutant detections.

Run `go test -v -count=1 -timeout=3h
./stage1/cohere/high_level_intermediate_representation/passes/unit-3`.
All new top-level Go tests call `t.Parallel`. The local suite checks adapter
boundaries/failures, all-input Node accounting, semantic mutants, both native
refusals and Go nondeterminism. Its report retains the known inliner failures;
set `HIR_UNIT3_REQUIRE_CERTIFICATE=1` to make those five raw-byte mismatches fail
the certificate check. No whole-lane certificate or native boxing count is claimed.
