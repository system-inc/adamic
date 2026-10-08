# Wave 30 port on checker facts

Base: d845dccde413c89643293e808626344d12e3f023, lint-checker/facts. This branch is lint-rules/facts-wave-30. No legacy bridge files, private checker, shared helper, harness or registration lists are changed.

## Completed implementation

nexus/consistency-no-iso-string-date-cut is implemented with supplied-node listeners for CallExpression, ElementAccessExpression and VariableDeclaration. Its single module uses the shared SymbolDetails decoder and the harness checker. Local declaration spans correlate against this file only. Date member provenance is queried through askFile(ReadsDefaultLibrary). The descriptor declares exactly upstream ReadsCompilerOptions and ReadsDefaultLibrary; it does not add ReadsOtherFiles. Findings have the exact upstream message and no fixes or suggestions.

## Claims stopped or already completed

| Rule | Exact upstream call/location | Missing question/helper |
| --- | --- | --- |
| `nexus/correctness-no-callback-in-parse-try` | `checker.SkipAlias(symbol, ctx.TypeChecker)` at `cohere/internal/lint/rules/nexus/correctness_no_callback_in_parse_try.go:247` | alias-declarations / alias-followed-symbol-details |
| `nexus/correctness-no-collection-misuse` | `part.AsLiteralType().Value()` at `cohere/internal/lint/rules/nexus/correctness_no_collection_misuse.go:258` | literal-value for a type identity, preserving string contents |
| `nexus/correctness-no-discarded-outcome` | `checker.Checker_getAwaitedType(ctx.TypeChecker, valueType)` at `cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome.go:135` | awaited-shape and type-declaration ancestry/union arm identity |
| `nexus/correctness-no-process-exit-after-output` | `ctx.TypeChecker.GetAliasedSymbol(symbol)` at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:413` | alias-declarations; also resolved signature declaration/body and selected return type at :314/:342 |
| `nexus/correctness-require-blocking-standard-streams` | `analysis.ctx.TypeChecker.GetAliasedSymbol(symbol)` at `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:586` | alias-declarations; also program runtime module graph and foreign function bodies |
| `react/jsx-fragments` | `declaration.AsVariableDeclaration().Initializer` at `cohere/internal/lint/rules/react/jsx_fragments.go:337` | foreign declaration initializer syntax; no ReadsOtherFiles is declared by this upstream rule |
| `react/jsx-no-constructed-context-values` | `walk.ctx.TypeChecker.GetResolvedSignature(call)` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:488` | resolved signature declaration/body; native shared construction/stability/escape analysis remains absent |
| `react-hooks/purity` | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/purity.go:250` | shared native React HIR, SSA and render/capture analysis |
| `react-hooks/refs` | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` at `cohere/internal/lint/rules/react/refs.go:184` | shared native React HIR, SSA and capture analysis |
| `react-hooks/preserve-manual-memoization` | `hir.CloneFunction(hir.ForFunction(ctx, functionNode))` at `cohere/internal/lint/rules/react/preserve_manual_memoization.go:195` | shared native React HIR/SSA/reactive scopes; hir.AnalyzePreservedManualMemoization at :199 |

Already completed by wave 1 according to Ahra, so not ported again: nexus/correctness-no-uncleared-race-timeout, nexus/correctness-no-discarded-pure-result, react/jsx-no-undef. Their upstream cases and mutants were not rerun as owned work in this unit.

The remaining JSX fragment blocker concerns a symbol whose variable declaration is in another file. Complete symbol records expose its span, but not its initializer AST. Same-file initializer lookup does not prove full rule behavior for that case. Opening/parsing the foreign file, adding a private question, or declaring ReadsOtherFiles where upstream declares none would violate the unit requirements. No partial registered rule is installed.

## Validation

Driver type-check and lint-registry generation pass. The full lint package is running once with all corpus, benchmark and profile inputs supplied and -timeout=3h. Results, counts, mutant comparison lines and setup measurements will be recorded after it finishes.
