# Wave 01 on checker facts

Base: `d845dccde` from `origin/lint-checker/facts`. Branch: `lint-rules/facts-wave-01`.
All twelve claims were reassessed. `symbol-description` is already ported on the requested base and was not ported again. The remaining eleven stop at the exact dependency below; no descriptor, private checker or shared helper copy was added. No shared file was edited.

| Rule | Exact Go call | Cohere location | Missing question or helper |
| --- | --- | --- | --- |
| nexus/correctness-no-implicit-return | `checker.Checker_functionHasImplicitReturn` | `nexus/correctness_no_implicit_return.go:96` | function-return-flow: function end-flow and implicit return |
| @typescript-eslint/no-deprecated | `checker.Checker_resolveAlias` | `typescript/no_deprecated.go:447` | alias-target identity/declarations; raw JSDoc records do not resolve alias chains or overloaded signature declarations |
| no-else-return | `ctx.TypeChecker.GetSymbolsInScope` | `core/no_else_return.go:372` | symbols-in-scope: ordered symbols, declarations and lexical scope ownership at the if |
| nexus/correctness-require-child-process-error-listener | `ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol` | `nexus/correctness_require_child_process_error_listener.go:378` | export value-binding identity; new shorthand declaration records do not supply export targets |
| nexus/correctness-require-response-status-check | `control_flow_graph.Build` | `nexus/correctness_require_response_status_check.go:521` | shared native CFG with the upstream event hooks |
| nexus/performance-no-independent-await-in-loop | `checker.SkipAlias` | `nexus/performance_no_independent_await_in_loop.go:479` | alias-skipped symbol identity |
| react-hooks/globals | `utilsreact.IsComponentOrHookLike; utilsreact.IsReachableRootPosition` | `react/globals.go:474; ecmascript/react/compiled.go:162,95` | shared native React compilation-root gate; neither helper is present under stage1/cohere/lint |
| react-hooks/immutability | `high_level_intermediate_representation.MayHoldComponentOrHook` | `react/immutability.go:166` | shared native React HIR/SSA/capture analysis |
| react-hooks/no-deriving-state-in-effects | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | `react/no_deriving_state_in_effects.go:142` | shared native memo-erased React HIR/SSA/capture analysis |
| require-atomic-updates | `control_flow_graph.Build` | `core/require_atomic_updates.go:402` | shared native CFG with read/write/suspend hooks |
| require-await | `checker.Checker_getResolvedSignature; resolved.Target()` | `core/require_await.go:988,992` | declared-call-signature: declared generic target, type parameters and substitutions |

Raw reproducers beside this report exercise the named dependency; they were not installed as witnesses or counted as certification. The export specimen observes an escaped child-process binding; the import-equals specimens force alias resolution without needing a second file.

Locations are relative to `cohere/internal/lint/rules/`, except `ecmascript/react/compiled.go`, relative to `cohere/internal/lint/`.
The new complete declaration and shorthand facts are acknowledged as available. Missing alias/export identities, CFG, scopes and React compilation gates are separate dependencies. For `globals`, native reference.WritesToBinding is available; it does not replace the shared React root gate. Its minimal positive input must include JSX or a hook call, unlike the older audit input.

```typescript
let g=0; function Component(){g=1; return <div/>;}
```

## Program read contract

No new rule descriptors were created and no inherited programReads declaration was changed. The new declarationAnswer wrapper requires ReadsOtherFiles. It cannot simply be copied into a rule whose cohere declaration does not authorize that read. Such local asks must use the shared checker question directly; arbitrary foreign reads cannot be smuggled through that local route. These notes do not add ReadsOtherFiles or ReadsDefaultLibrary to any rule.

## Certification scope

For each of the eleven blocked rules: newly certified upstream cases 0; new mutant executions 0. The package run certifies existing registered rules, including the inherited symbol-description port; it is not proof of an uninstalled claim. Full package results and caught mutant log lines are recorded beside this report.

## Full package result

Command: `go test -json -count=1 -timeout=3h ./stage1/cohere/lint`, after sourcing `/workspace/adamic-tools/env.sh`. `GOMAXPROCS=4`, `GOFLAGS=-buildvcs=false`; all optional corpus, benchmark and profile inputs were enabled as recorded in [evidence/result.json](evidence/result.json). The pinned compiler corpus was `/workspace/wave-01-typescript`, TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`.

Exit 0; top-level 34 pass, 0 fail, 1 skip. Go package elapsed 2000.020s; outer wall 2010.122s. nproc 5, quota 4 cores. One-minute load minimum/median/maximum 1.000/1.406/6.123. Setup completed in 42.897s.

Only skip: `TestCheckerBridgeRefusalPending`, reason `awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer`. This requested base has no TSGoError in its prelude. No optional-input check skipped, and the pending control was not edited or bypassed.

`TestRulesAgree` passed in 142.40s with 4051 unique source/rule/options cases. `symbol-description` accounted for 30 unique captured cases, deduplicated using the harness fields rule, file, options and source. It is inherited from the facts base, not newly ported here. Its captured cases are [evidence/symbol-description-capture.json](evidence/symbol-description-capture.json).

All 80 registered mutants caught; [evidence/mutants.txt](evidence/mutants.txt) retains each catch with its original JSON log line. The inherited Symbol mutant is [evidence/symbol-description-mutant.txt](evidence/symbol-description-mutant.txt). No new mutants belong to the eleven blocked claims.

Compiler/repository comparison: 881 files, 30424152 bytes identical across Go, sanitized native, Node and emitted JavaScript. Owned witnesses, profile artifacts, profile compilation and profile snapshot agreement passed. Existing explicitly marked parser recovery refusals remain visible in the full log and are not counted as skipped tests.

Only rule-directory notes, raw reproducer inputs and logs changed. No executable rule was added. No bridge, harness, shared helper, descriptor or programReads declaration changed. No full-repository gate was run. Obsolete prior-run scratch binaries and checker archives were deleted by binary-header checks to leave room for the gate; source files, logs and current gate artifacts were retained.
