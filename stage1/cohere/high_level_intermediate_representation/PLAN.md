# Stage 1 HIR lane #2kb2kje

Base: origin/area/stage1-lint, 1f9e223d. Oracle: cohere gitlink 7945d102a6c18dd36adf9114a758ce646e8b2359. Scope is 59 production files, 25,071 Go lines and 25,666 test lines at this gitlink (the dispatch estimated 26,166). [INVENTORY.md](INVENTORY.md) records every production and test file's definitions, package references, external imports, test associations, and line count. The checked-in mapping script regenerates it from the pinned source. Package reference sets are conservative, as documented there; Go files share a namespace, so there is no truthful import DAG between individual Go files.

## Dependency order and seams

1. Define function-owned tables, identifiers/declarations, places/effects, instructions, patterns, terminals and visitors. IDs are meaningful only in their owning function; captures cross the boundary by positional pairing, not by ID equality.
2. AST lowering (`lower*`, `context_identifiers`, `optional_chains`) resolves symbols through the resident checker, sets captures, builds blocks, then finalizes. `spelling` supplies the cheap conservative file gate. Checker-less lowering is valid only for syntax that cannot reference bindings: it cannot stand in for rule construction.
3. `graph` implements the imported SSA graph adapter; finalize is reverse postorder → predecessors → evaluation order. `ssa` invokes imported construction and elimination, recursively per nested function; `ssa_verify` invokes the imported verifier. No implementation of these algorithms belongs in HIR.
4. Cached rule entry points expose constructed read-only functions. Clone is a deep owned-table copy. Memo-erased graphs get a distinct cache entry; erasure plus IIFE splicing can require reconstruction. Postdominators and control dominators operate on this HIR, not on cohere's unrelated AST control-flow graph.
5. Effect extraction and reactive facts feed imported mutation/alias range analysis, then disjoint sets, scopes, alignment and merging. Scope terminals lead to dependency collection and the reactive tree. Pruning/flattening/merging feeds preservation validation. This order is the actual `AnalyzePreservedManualMemoization` pipeline, not alphabetical file order.

There are cycles inside layers: visitors and variants depend on each other; the lowering builder's expression/statement/optional routines recurse; effects/reactivity/ranges call one another. Land each such group with stub-free public boundaries. Units 1 and 2 intentionally split core files by supported variants: the first slice does not claim all 43 instruction variants or any branch terminal.

## Entry points

| Entry | Required pieces and behavior |
| --- | --- |
| `ForFunction` | spelling gate at callers; checker + full lowering; finalize + imported SSA; per-file node-position cache. Construct exactly once per cache entry. |
| `ForFunctionWithoutManualMemoization` | above + separate cache key, manual-memo spelling recognition, drop pass and memo-inclusive IIFE splice; reconstruct if spliced. |
| `MayHoldComponentOrHook` | spelling.go's conservative syntax/file gate; must admit escaped names and unreadable spans. No graph needed. |
| `CloneFunction` | core owned tables, all variant/pattern/terminal copy visitors and nested functions; preserve intra-copy IDs and sever mutable storage from original. |
| `AnalyzePreservedManualMemoization` | constructed cloned function → outline → infer reactive → drop memo → inline/merge blocks → DCE → imported ranges → disjoint sets → assign/align/merge scopes → scope terminals → flatten loops/hooks → hoistable dependencies → reactive tree → nonescaping/nonreactive/unused pruning → invalidation merge → always-invalidating pruning → preservation/dependency comparison. |

`AsCompilationUnit` and `ForEachFunctionLike` are additional required caller seams: a nested compilation unit is re-lowered against its own enclosing symbol scope. Do not silently analyze test-local captures as component locals.

## Eight rules

These are HIR dependencies; validators/diagnostic emitters themselves remain separate rule ports.

| Rule | Entry and required graph features | First complete HIR unit |
| --- | --- | --- |
| static-components | MayHold + ForFunction + AsCompilationUnit; nested function tables/captures, calls, globals, properties, JSX tag places, phis and graph walking | 2 |
| purity | MayHold + ForFunction + AsCompilationUnit; globals, property/computed chains, calls, nested functions and SSA identity | 2 |
| set-state-in-render | MayHold + ForFunction + AsCompilationUnit; setter aliases, calls/memo callbacks, nested functions and UnconditionalBlocks (postdominators) | 2 |
| immutability | MayHold + ForFunction + AsCompilationUnit; mutation variants, place visitors, checker node handles, branches, phis and closure captures; the rule owns its own abstract environment | 2 |
| set-state-in-effect | MayHold + ForFunctionWithoutManualMemoization + AsCompilationUnit; erasure/inlining, calls, captured setters, ControlDominators for ref-controlled blocks | 3 |
| refs | MayHold + memo-erased entry + AsCompilationUnit; JSX, property/computed accesses, place and terminal visitors, checker handles, branch aliases and closures | 3 |
| no-deriving-state-in-effects | MayHold + memo-erased entry + AsCompilationUnit; phis, instruction/terminal visitors, effects and callback aliases (rule-local taint) | 3 |
| preserve-manual-memoization | MayHold + ForFunction, CloneFunction, AnalyzePreservedManualMemoization; complete analysis pipeline and preservation findings | 10 |

## Existing analysis modules: import only

SSA is present at `stage1/cohere/static_single_assignment`. Its `GraphInterface<F,B,P>` has the necessary entry/table/block/phi/edge/parameter/returns/declaration/mint/order methods. Go's pointer visitor differs: Adamic's visitor RETURNS the renamed place, and the HIR adapter must write it back at each instruction/terminal/phi position. Arrays replace Go's pointer/slice API; `block` returns undefined rather than `(block, bool)`. Phi operands are sorted `{predecessor, place}` arrays. These are adapter differences, not reasons to copy any analysis. The first slice imports graph passes and construct directly and shares the ID and Phi types.

**Base mismatch:** `stage1/cohere/mutation_aliasing` is absent on this requested base. `go.mod` references Go cohere/mutation_aliasing, but that is not an Adamic implementation. Its port's interface therefore cannot be certified here. Unit 5 must wait for or integrate that module by import, never copy it. The Go contract extends the SSA adapter with effect iteration/projection, instruction and terminal orders, call/apply effects, value kinds, mutable ranges and nested summaries. HIR `effects.go` aliases generic `mutation_aliasing.AliasingEffect[Place]`; `ranges.go` uses its graph adapter and analysis, and `disjoint.go` consumes its range/disjoint operations. Confirm returned-place mutation, identifier/declaration ID domains, effect discriminants, range half-open boundaries, context seeding and summary ownership against the landed port before starting unit 5. Do not guess export names. The exact Go graph extension is `InstructionOrder` (order, valid), `TerminalOrder`, `Effects`, `ParametersFrozen`, `Context`, `ReturnValue` (place, valid), `StoredContextValue` (id, valid) and `Closure` (captures, frozen; read-only kinds map), at cohere/mutation_aliasing/mutation_aliasing.go:214–248. `Options.ParametersDefinedOnEntry` must remain false for React; `Options.ContextKinds` seeds closure probes. HIR's adapter must implement these on top of its SSA adapter once the imported port exists.

## Oracle contract

The adapter lives here in `testdata/oracle_test.go`, overlaid beside the Go HIR package under `-tags=lintoracle`. It calls real Go Lower and Construct. It never edits cohere or its gitlink. Run with cohere's existing go.work: the package imports internal TypeScript modules. Each selected source is parsed as `/test.tsx` on both sides. The adapter fails on a variant outside its declared dump subset rather than omitting its payload.

`hir-v1` is LF-terminated UTF-8 records: function name/kind/entry/async/generator; ordered params/context/returns; the complete identifier table (id/declaration/name); RPO blocks (id/kind/predecessors); instructions (table id/evaluation order/lvalue/source span/variant/payload); terminal (order/variant/places); scopes; end. A place is `identifier:effect:reactive:pos:end`. Literal payloads distinguish nil, bool and string (Go deliberately stores numeric spellings as strings). The first slice has empty scope results, spelled `scopes -`; it never pretends that a scope analysis ran.

Expansion contract: add all variant fields in declaration order, tagged pattern trees, successor edge kind, phi operands sorted numerically by predecessor, ordered nested-function dumps keyed by function-table index, sorted outlined/context-declaration tables, projected per-reference effects, range sets, scope IDs/start/end/declarations/dependencies, and analysis checkpoint records. Strings outside the present numeric/name alphabet must use a specified JSON string escape routine on BOTH sides before admitting those cases. A checkpoint selects raw lowering, constructed, memo-erased, effects/ranges, scopes, or final preservation. Freeze new records before a unit lands and compare all prior checkpoints as regression coverage.

IDs are the original function-local allocation order (blocks start at 1; identifiers/instructions start at 0; fresh declaration = identifier+1). Keep them, including SSA-minted identifiers and unused table entries; do not renumber to hide differences. Array order is semantic except unordered Go maps, whose keys are explicitly sorted. No addresses, timestamps, absolute corpus paths, checker pointers or map iteration order enter the dump. Source spans retain parser trivia positions. Repeat each oracle generation and assert identical bytes when adding map-backed results.

Corpus tiers:

* Every Go HIR test's function sources, preserving checker setup when it needs symbol resolution. Associate cases with the test symbol and pinned line. Build real checker-backed compilation units for the full lowering tier.
* All vendored `cohere/internal/lint/rules/react/conformance/testdata/fixtures` upstream sources for these eight rules, including nested functions. Exclude Flow only with a recorded reason; include valid and error cases, not only those passing cohere today.
* Stage 1 parser/ESTree React and JSX fixtures and lint React fixture sources, including malformed cases as explicit declines. Do not count ordinary parser implementation functions as fixtures.

The first selector extracts whole no-parameter function declarations containing only numeric/bool/null expression statements, empty statements and returns, with path/line provenance in `testdata/manifest.json`. It also includes five labeled probes. It lifts the function as an independent source; there are no binding references in this subset. Unmatched corpus functions are out of coverage, not successes. The selector currently finds no eligible stage 1 React/JSX function; that tier is deferred to unit 2.

## Ordered landable units

Each unit has a byte comparison against the tagged Go adapter at its named checkpoint, all previous checkpoints, and at least the indicated successful-but-wrong semantic mutant. File sets and full upstream line budgets below partition the 59 production files exactly. Budgets for unit 1 include core files completed in unit 2; this landing implements only their smallest variant subset.

| Unit | Files (Go basenames) | Lines | First rule / support | Oracle and mutant | Risks |
| --- | --- | ---: | --- | --- | --- |
| 1. Core and versioned oracle | high_level_intermediate_representation instruction terminal pattern print | 1,960 | No complete rule; foundation for static-components | constructed literal-only functions; return-store-to-nil | 43-way tagged unions, owned tables, source spans; string escaping before expansion |
| 2. Complete construction and rule-facing graph | lower lower_expression lower_global lower_optional context_identifiers optional_chains spelling graph ssa ssa_eliminate ssa_verify visitor visitor_mutate cache clone postdominator | 5,471 | static-components first; also purity, set-state-in-render, immutability | checker-backed constructed IR and clone; swap branch successors / alias clone storage | checker symbol identity, captured closures, generic SSA callbacks, iterator lowering, exception/finally and optional-chain evaluation order |
| 3. Memo-erased view and graph rewrites | drop_manual_memoization inline_iife inline_remap invoked_functions outline_functions dead_code_elimination merge_consecutive_blocks | 2,427 | set-state-in-effect first; also refs, no-deriving-state-in-effects | memo-erased dump; keep useCallback wrapper or remap a capture to wrong ID | closure substitution, deep remapping, cache isolation, post-splice SSA reconstruction |
| 4. Effects and reactive facts | effects effects_custom_hooks primitive_property_constraints reactive | 2,879 | preserve-manual-memoization foundation; no additional complete rule | effect extraction/reactive checkpoint; treat capture as read | struct-key maps (primitiveConstraintKey), signature closures, checker type nodes, effect lists and recursive custom-hook summaries |
| 5. Imported ranges and disjoint adapter | ranges ranges_readonly_closures disjoint | 1,047 | preserve-manual-memoization foundation; no additional complete rule | ranges/representatives checkpoint; omit captured-value widening | BLOCKED until mutation_aliasing port is available; generic effects, nested summary identity, frozen joins and closure worklist convergence |
| 6. Scopes and alignment | scopes align_method_calls align_scopes merge_scopes scope_terminals memoization_graph memoization_inputs memoization_level | 3,173 | preserve-manual-memoization foundation; no additional complete rule | scope checkpoint; merge overlapping scopes that do not share a value | sets of ranges, nested numeric maps, scope identity remapping, loop/branch terminal intervals |
| 7. Dependencies and hoistability | dependencies hoistable always_invalidating | 3,213 | preserve-manual-memoization foundation; no additional complete rule | dependency path checkpoint; truncate property path to root | path identity registries, null/optional path propagation, nested callback invocation and stable property identity |
| 8. Reactive tree and visitors | reactive_function reactive_build reactive_transform reactive_visitor | 2,199 | preserve-manual-memoization foundation; no additional complete rule | reactive tree checkpoint; drop one scope exit | recursive discriminated trees, visitor closures, flattening structured terminals while preserving order |
| 9. Prune and invalidation pipeline | prune_always_invalidating prune_non_escaping_scopes prune_non_reactive_dependencies prune_unused_scopes merge_invalidating flatten_reactive_loops flatten_scopes_with_hooks | 1,949 | preserve-manual-memoization foundation; no additional complete rule | post-prune tree + pruned identities; retain an always-invalidating scope | iterator mutation during traversal, scope identity sets, checker-backed escaping callback distinctions |
| 10. Preservation orchestration and comparison | manual_memo_comparison preserve_manual_memoization | 753 | preserve-manual-memoization | all pipeline checkpoints + findings; accept a missing manual dependency | closures over scope stacks, dependency path identity, preservation scoring and precise diagnostic spans |


Dependencies: 1 → 2 → 3; 2 → 4 → 5 → 6 → 7 → 8 → 9 → 10, with unit 3 also feeding 10. Unit 8 can be implemented after 6 while 7 develops, but the landing order above is intentionally serial. Units 4–9 expose separately useful analysis checkpoints but do not by themselves enable another whole rule. Claiming they unblock preservation before unit 10 would be false. Unit 2 is the earliest complete rule seam; the finished first slice alone cannot unblock static-components.

For unsupported language constructs, submit the original Go path:line, minimal port reproducer and exact compiler diagnostic to the lane owner. Do not replace required identity semantics with stringified keys, flatten closure state, duplicate imported analysis, or silently skip iterator cases. Ordinary supported rewrites such as separate pushes for variadic append are porting. Inventory the confirmed struct-key maps before implementation (`primitive_property_constraints.go:19–21` and `155–157`, keyed by `primitiveConstraintKey`), generic adapters (`graph`, `ssa`, `effects`, `ranges`), closures (`lower`, `visitor*`, `reactive*`, `inline*`) and JavaScript iterator protocol lowering (`lower.go`, `lower_expression.go`). No Go `iter.Seq` import was found in this package; iterator risk here is protocol/effect fidelity and map iteration during mutation, not an observed range-over-function requirement. A risk is not a confirmed compiler gap.

## First landing acceptance

Implemented files: core.ts (IDs imported from SSA, function/block/instruction/identifier/place data); graph.ts (SSA adapter and imported passes); lower.ts (AST lowering of no-parameter literal-only declarations); dump.ts and main.ts; the tagged overlay adapter and Go harness; provenance corpus selector; package inventory and this plan. All other syntax is declined; no checker resolution, branches, nested functions, strings, JSX, mutation analysis or scopes are claimed.

Run `source /workspace/adamic-tools/env.sh; go test -v -count=1 -timeout 20m ./stage1/cohere/high_level_intermediate_representation`. The test compares Go with Node and the native Adamic build, then compiles and executes the semantic mutant on both. The mutant replaces the return's LoadLocal copy with Primitive nil, preserving successful execution; matching a changed pretty printer is not sufficient. The checked-in corpus is seven actual corpus functions (one HIR test, six upstream occurrences), plus five distinct probes; zero stage 1 React/JSX functions fit this initial subset. Every case must match byte for byte. The full corpus is deliberately not counted as covered.


Native compilation feedback during this slice was on port expressions, not an unavoidable Go language gap: an optional method call in classification was rejected (`core.ts:51:52: stage 0 can't lower a call through ?. (an optional call) yet`), and a constructor called a method before initializing its fields (`core.ts:52:24: Adamic 0.1 refuses this escaping a constructor before every field is set`). The final port uses Go lower.go:1366's direct ASCII classification and initializes every field before method calls; no required Go operation was bypassed. No confirmed blocking language gap remains in the slice.

## Unit 2 continuation

See [PROGRESS.md](PROGRESS.md) for the current clean seam and exact remaining construction/rule work. Its executed-test census supersedes the first landing's source-text selector for unit 2 coverage: 51/1,465 corpus functions, plus 9/9 path probes. Unit 2 and static-components remain unfinished. The owner confirms mutation_aliasing is scheduled for tonight's area pin bump and is not required before unit 4.
