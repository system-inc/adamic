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

`hir-v1` is LF-terminated UTF-8 records: function name/kind/entry/async/generator; ordered params/context/returns; the complete identifier table (id/declaration/name); RPO blocks (id/kind/predecessors) and phis; instructions (table id/evaluation order/lvalue/source span/variant/payload), followed by orphan instructions in table order; terminal (order/variant/places); scopes; end. A place is `identifier:effect:reactive:pos:end`. Literal payloads distinguish nil, bool and string (Go deliberately stores numeric spellings as strings). The first slice has empty scope results, spelled `scopes -`; it never pretends that a scope analysis ran.

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

See [PROGRESS.md](PROGRESS.md) and [REPORT.md](REPORT.md) for the current certificate: 1,442/1,465 originals on native and Node (all 1,442 non-Flow graphs), plus 72/72 path probes. The optional-field ruling is applied. Static-components' owned rule certificate is recorded there. Unit 2 is stopped on the isolated native CloneFunction static-constructor/spread gap, with its Node-only certificate kept separate. The parallel pass assignments below supersede implementation sequencing while preserving Go's integrated execution order.


## Parallel pass lanes after unit 2 (step 08 / step 28)

The arrows above describe **integrated execution order**, not an implementation
prerequisite. A worker can consume the Go state immediately before its pass and
compare its result with Go immediately after the pass. No unit 3–10 needs a
previous worker's Adamic algorithm to begin this isolated work. Every checkpoint
keeps the original 1,465 census keys: no new hash from a changed graph, no sample
and no silently dropped error path. Go-input replay can retain the 23 Flow graphs
without parsing them in Adamic; integrated construction still catalogs those
23 exclusions. Each adapter exports the original file kind, source/AST handles,
checker availability and query facts separately from analysis answers.

### Replay bootstrap and the limit of today's dump

The current `hir-v1` **construction** oracle is byte-identical on the admitted
corpus. It is a printer, not yet a decoder, and its `scopes -` line is a
placeholder. It does not yet print `AliasingEffects`, mutable ranges, disjoint
representatives, scope alignment/dependencies, reactive trees or preservation
findings. Calling those later lanes immediately oracle-certified would be false.
They are **ready to staff for Go-input development today**, with the shared replay
bootstrap below; none is waiting for a prior Adamic pass. Their full comparison
becomes runnable when their named before/after state codec exists.

Assign one replay/integration owner before the pass workers start. That owner
alone writes `replay/`, `core.ts`, `dump.ts`, `graph.ts`, `cache.ts`,
`symbol_query.ts`, `symbol_live.ts`, `testdata/oracle_test.go`, the census exporter,
and `../arena/arena_index.a`. Pass workers request missing record/arena types from
that owner; they never edit those shared files. Add all Go instruction variants
and the scope/reactive arena index classes centrally once, with private minting
and checked reads. Keep the construction dump unchanged where new state is absent.
SSA and mutation_aliasing remain imported algorithms, never copied.

The owner's codec exports and reads **checkpoint bundles**: the existing hir-v1
function/block/instruction/place rows plus canonical sidecar rows below, linked by
function-nesting path and existing identifier/block/instruction IDs. AST/type
facts use source kind/span handles; they are input facts, not expected analysis
results. Maps and sets print sorted numeric identities; dependency paths retain
segment order, computed/optional distinctions and their identity registry. Effects
retain their original sequence. Scope and reactive-node IDs use Go's IDs, and
fresh IDs follow Go's allocation order. Recursive graphs are typed index arenas;
no object-address strings or cycle-capable links. The codec roundtrip must itself
be byte-identical and have a corrupt-index mutant before any pass certificate.

Each worker exclusively owns `passes/unit-N/` (implementation, local types,
Go tagged adapter, input/output fixtures, comparison test and mutants). Its
sidecar encoder belongs in that directory, using the shared framing/identity
reader by import. This gives separate files to simultaneous workers. Shared-file
changes and pipeline wiring are serialized by the replay/integration owner; unit
10 owns the composed orchestration, requesting the final public entry/cache glue.
The pipeline order must follow Go `AnalyzePreservedManualMemoization`, not merely
unit number: outline → reactive facts → drop memo → inline/merge/DCE → ranges →
scopes/terminals → loop/hook flattening → dependencies → reactive tree → prune →
validate. In particular unit 9's flattening runs before units 7 and 8.

| Unit / exclusive write directory | Go entry point(s) and before state | Output in hir-v1 checkpoint terms | Shared files / modules read; writes reserved to integration owner | Can start from Go input today? |
| --- | --- | --- | --- | --- |
| **3 memo / rewrites**, `passes/unit-3/` | `drop_manual_memoization.go:198 DropManualMemoization`; `inline_iife.go:81 InlineImmediatelyInvokedFunctionExpressions` / `:116 ...IncludingMemoCallbacks`; `inline_remap.go:67 CopyNestedBodyInto`; `outline_functions.go:38 OutlineFunctions`; `dead_code_elimination.go:43 EliminateDeadCode`; `merge_consecutive_blocks.go:62 MergeConsecutiveBlocks`; cache entry `ForFunctionWithoutManualMemoization`. Supply constructed SSA graph, captures, callee origins and source dependency facts. Preservation's drop checkpoint is after outline + InferReactive; export each subpass separately. | Rewritten instruction/identifier/block tables, memo Start/Finish markers and source dependencies, nested/outlined identities, remap tables and rewrite/DCE counts. Rebuild SSA with imported module at Go's corresponding checkpoint. | Read core/dump/graph, shared arena, SSA and symbol facts. Only owner adds missing Memo/DeclareContext variants or the public cache hook; worker writes its own memo cache module. | **Ready for isolated Go-input work.** Needs replay graph + AST/memo sidecars, not unit 4's Adamic code. The pending native CloneFunction gap blocks integration through that clone, not fresh-arena replay. |
| **4 effects / reactive facts**, `passes/unit-4/` | `effects.go:493 InferAliasingEffects`, `:523 InferAliasingEffectsForNested`, `:1189 ProjectEffects`; `reactive.go:189 InferReactive`. Supply appropriate prepass graph, hook origins, imported/callee signatures, AST/type-checker facts and nil-checker context where originally nil. ProjectEffects receives Go-produced ranges at its own independent checkpoint. | Place effect/reactive flags; ordered alias effect records, captures and closure/custom-hook summaries, value kinds, primitive constraints and projected effect map. | Read frozen core, SSA and symbol-query bridge; mutation_aliasing types for projection. Unit 4 owns all effects/primitive/reactive-fact types. Owner alone extends generic replay fact selectors. | **Ready for isolated Go-input work.** Graph and checker-fact codec required; no prior Adamic pass. Full native import must use the area's mutation_aliasing module where projection requires it. |
| **5 ranges / disjoint**, `passes/unit-5/` | `ranges.go:31 InferMutableRangesWithEffects`, `:44 RangesForNested`, `:115 MutationSites`; `disjoint.go:331 FindDisjointMutableValuesWithRanges` (wrappers at `ranges.go:20`, `disjoint.go:315`). Input is graph + Go-produced alias effects, closure value kinds and nested summaries; disjoint's input additionally has Go ranges. | Mutable intervals keyed by function path/identifier, frozen/value-kind facts, closure widening and mutation sites; deterministic disjoint representative and union membership rows. Preserve separate nested tables. | Import `static_single_assignment` and `mutation_aliasing` **by reference**. Read core, unit-4 public fact types and replay codecs. Worker owns its adapters and disjoint state; does not edit either analysis module. | **Prepare Go-input adapters now; execution BLOCKED on the mutation_aliasing import in this checkout.** The promised lint-area pin/module bump is not present here (`stage1/cohere/mutation_aliasing/` is absent). No Adamic unit 4 algorithm is needed once that existing module lands; never copy the Go algorithm to bypass it. |
| **6 scopes / alignment**, `passes/unit-6/` | `scopes.go:264 AssignReactiveScopesWithSets`, `:409 ValidateScopes`; `align_method_calls.go:44 AlignMethodCallScopes`; `align_scopes.go:375 AlignThenMergeReactiveScopes`; `scope_terminals.go:335 BuildReactiveScopeTerminals`. Input: graph + Go ranges/disjoint representatives, then scopes/alignment at each subpass boundary. | Scope assignments and intervals, aligned/merged identity maps, declarations/outputs and memoization graph levels; inserted scope terminals, including structured edge IDs and validation results. | Read frozen core/arena, unit-5 range/disjoint schema and shared framing. Scope index class/core terminal variants are added once by owner. Unit 6 owns ScopeIdentity and all scope/alignment data types. | **Ready for isolated Go-input work.** Needs scope sidecars and decoder, not Adamic ranges/effects. Current `scopes -` alone is insufficient input or output. |
| **7 dependencies / hoistability**, `passes/unit-7/` | `hoistable.go:826 CollectHoistablePropertyLoads`; `dependencies.go:1138 CollectScopeDependenciesWithHoistable` (wrapper `:1077 CollectScopeDependencies`); `always_invalidating.go:50 IsAlwaysInvalidatingType`. Input: graph after scope terminals and loop/hook flattening, Go scope identities/ranges, invoked-function facts, null/optional paths and checker facts. | Per-scope ordered dependency paths/outputs, path identity and temporary registries, hoistable-load tree and stable/invalidating facts. Unchanged graph rows still compare. | Read frozen core, unit-3 invoked schema, unit-5 ranges and unit-6 ScopeIdentity schema. Worker owns path/hoistable/dependency types and encoder; new shared path arena handles only owner writes. | **Ready for isolated Go-input work.** Go-produced flattening input replaces waiting for units 6 or 9 Adamic implementations. Requires full path/identity and checker-fact codec. |
| **8 reactive tree**, `passes/unit-8/` | `reactive_build.go:281 BuildReactiveFunction`, `:293 BuildReactiveFunctionWithFlattenedScopes`; `reactive_visitor.go:81 VisitReactiveFunction`; `reactive_transform.go:103 TransformReactiveFunction`. Input: Go graph with scope terminals, Go scope/flattened identity sets and required dependency facts. Visitor/transform fixture states include the explicit operation to apply. | Typed reactive arena node/edge rows, structured blocks/loops/scopes, instruction order, scope outputs/dependencies and `ReactiveBuildResult` failures. Tree and source graph both dump; impossible builds are counted outcomes, not discarded cases. | Read frozen core/arena and unit-6/7 schema. Owner creates shared ReactiveIndex; worker exclusively owns reactive types, builder, visitors and transforms. | **Ready for isolated Go-input work.** No earlier Adamic pass required. Needs reactive-tree codec; construction hir-v1 cannot yet express this tree. |
| **9 flatten / prune / merge**, `passes/unit-9/` | `flatten_reactive_loops.go:51 FlattenReactiveLoops`; `flatten_scopes_with_hooks.go:38 FlattenScopesWithHooksOrUse`; `prune_non_escaping_scopes.go:95 PruneNonEscapingScopesWithScopes`; `PruneNonReactiveDependencies`; `PruneUnusedScopes`; `MergeReactiveScopesThatInvalidateTogether`; `PruneAlwaysInvalidatingScopes`. Export before/after each exact Go subpass, not one invented late flattening stage. | Flattened/pruned scope identity sets, changed graph terminals for flattening, changed reactive tree and dependency table, merge/last-use facts and per-pass counts. Preserve pruned identities even after the tree removes the scope. | Read core, unit-6 ScopeIdentity, unit-7 dependencies and unit-8 reactive schema; read checker facts for escaping/type distinctions. Worker owns prune/flatten code and result schema. | **Ready for isolated Go-input work.** Go trees/tables supply prerequisites; no need to wait for Adamic unit 8. Requires graph + reactive + dependency replay codecs. |
| **10 preservation / composition**, `passes/unit-10/` | `manual_memo_comparison.go:86 CompareManualMemoDependencies`; `preserve_manual_memoization.go:107 ValidatePreservedManualMemoizationWithPruned`; orchestration `:439 AnalyzePreservedManualMemoization`. Input: Go post-prune tree, source graph/memo markers, aligned scope identities, inferred/source dependency tables and pruned-by-chain set. | Ordered preservation findings (identifier, scope, evaluation order, kind), comparison results and canonical per-subpass checkpoints. Rule adapter separately compares exact Go diagnostics. | Read all public pass schemas, core and cache. Worker owns comparison, validation and composed pipeline module. Only integration owner edits root public exports/cache and rule registration remains in the rule's own directory. | **Ready for isolated validator/comparison work.** Go-produced final states suffice today. **Full composed Adamic AnalyzePreservedManualMemoization truly needs units 3–9 and the native clone gap resolved.** |

Units 3, 4, 6, 7, 8, 9 and 10 can be staffed now for isolated implementation,
with one shared replay/integration owner in parallel. Unit 5 can prepare its
adapters now, but executing them awaits the mutation_aliasing area import; unit
4's projection subpass has the same dependency. All eight are independent of
previous Adamic algorithms for their Go-input certificates.
No pass is marked ready-to-run against a decoder that has not been built. Each
lane must land its complete 1,465-record checkpoint comparison and a semantic
mutant before claiming completion; native and Node counts remain separate until
both execute. Shared schemas land before dependent lane imports, but schema
availability is distinct from waiting for the earlier algorithm.
