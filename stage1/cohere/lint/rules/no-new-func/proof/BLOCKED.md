# Wave 06 remaining dependencies

Baseline: `d845dccde413c89643293e808626344d12e3f023`. The three registered ports use the shared same-node checker.ask question grammar and shared SymbolDetails decoder. Their upstream ProgramReads is zero, so every descriptor has an empty programReads array. No ctx.Program reads, foreign-file opens, askFile calls, private checker or copied shared analysis helper were added. Go documents ProgramReads as reads through ctx.Program at cohere/internal/lint/rule/rule.go:323-337.

| Rule | Outcome / exact remaining call and needed shared contract |
| --- | --- |
| `adamic/nominal-class` | `flow.WalkerFor(ctx)` at `cohere/internal/lint/rules/adamic/nominal_class.go:68`; need shared source flow Walker/Pair/Site traversal and nominal type metadata. |
| `no-useless-return` | `control_flow_graph.Build` at `cohere/internal/lint/rules/core/no_useless_return.go:125`; need shared CFG builder/hooks and promised-return type question (no_useless_return.go:381). |
| `nexus/correctness-require-child-process-error-listener` | `ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol(parent)` at `cohere/internal/lint/rules/nexus/correctness_require_child_process_error_listener.go:378`; need exact export-specifier local target identity. GetSymbolAtLocation and shorthand declarations now exist. |
| `nexus/correctness-require-response-status-check` | `control_flow_graph.Build` at `cohere/internal/lint/rules/nexus/correctness_require_response_status_check.go:521`; need shared CFG Builder and hooks. Symbol/provenance facts alone do not supply this analysis. |
| `nexus/performance-no-independent-await-in-loop` | `checker.SkipAlias(symbol, analysis.ctx.TypeChecker)` at `cohere/internal/lint/rules/nexus/performance_no_independent_await_in_loop.go:479`; need resolved alias target identity and declaration body syntax. |
| `no-new-func` | Ported on shared facts; certification is recorded in REPORT.md. |
| `no-new-native-nonconstructor` | Already completed elsewhere per assignment; not ported again. |
| `no-new-wrappers` | Already completed elsewhere per assignment; not ported again. |
| `no-throw-literal` | Symbol facts now fit. Parked after the unchanged typed oracle rejects upstream `function f() { throw; }` before comparison: lint_test.go:377-382 lacks the recovery classification used by its syntax branch. See parked-no-throw-literal/BLOCKED.md and typed-recovery-failure.log. |
| `no-useless-backreference` | `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)` at `cohere/internal/lint/rules/core/no_useless_backreference.go:100`; need shared reference tracker, GlobalReferences and ConstantStringIn. |
| `prefer-arrow-callback` | Facts now fit. Parked on shared Linter.fixed edit-endpoint overlap refusal; see parked-prefer-arrow-callback/BLOCKED.md. |
| `react-hooks/set-state-in-effect` | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_effect.go:267`; need shared source HIR, SSA and capture analysis. |
| `react-hooks/set-state-in-render` | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_render.go:183`; need shared source HIR, SSA and capture analysis. |
| `react-hooks/static-components` | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/static_components.go:120`; need shared source HIR, SSA and capture analysis. |
| `react/jsx-fragments` | `jsx.ElementParts(opening)` at `cohere/internal/lint/rules/react/jsx_fragments.go:207` has no shared native equivalent. `jsxFragmentsImportModuleName` at line 393 also needs imported-name/module-specifier syntax not serialized by node-symbol-details. Need shared JSX element parts and declaration syntax facts; do not copy shared helper. |
| `react/jsx-no-undef` | Already completed elsewhere per assignment; not ported again. |
| `react/no-adjacent-inline-elements` | `isPragmaCreateElementCall(ctx, node)` at `cohere/internal/lint/rules/react/no_adjacent_inline_elements.go:236` is a shared helper absent on this baseline. Need shared pragma/import-binding classifier with initializer syntax; declaration presence alone is insufficient. |
| `@typescript-eslint/no-duplicate-type-constituents` | Historical shared fixer/parser refusal remains outside this unit: Linter.fixed reparses `type A = number & string & (  );` at lint.ts:198; parser.ts:499 refuses CloseParenToken. Original upstream input at no_duplicate_type_constituents_test.go:383 is `type A = number & string & (number & string);`. Checker raw-shape already exists. No fresh certification claimed. |

Original minimal reproducers and assessment are pinned in [the prior status file](https://github.com/system-inc/adamic/blob/c2deffd4260641f0c95c5de779f026732cfab607/stage1/cohere/lint/rules/no-iterator/wave-06-checker/claimed-rule-status.json). Those historical observations are not new parity results.
