# Wave 12 checker dependency audit

Audit only; no ports, shared files or checker implementations changed.

Area snapshot: `fb6cb5dbf79e247d74dccf16d8e3aefc5b31c039`.
Upstream cohere source inspected: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Legacy wave branch inspected: local `e1ca3a2bfc6e5bbe2a1566cc7175de850529205b`, whose pushed source tip was `a2ea9f82a7409feda6c8b7bdd9cf439e62c6790a`.
Claims: `stage1/cohere/typeaware/claims/wave-12.md` on that legacy branch.
Earlier reproducer and full-package evidence commit: `8f290f3cbe37becb196319b193e852f396d0800e` on `codex/lint-checker-wave-12`, under `stage1/cohere/lint/rules/no-multi-assign/checker-audit/`.

This records the first blocking dependency for every one of the 18 claims, with overlapping dependencies grouped for scheduling. It is not a claim that implementing these calls alone completes every upstream algorithm. Rule slugs below retain their namespace in the per-rule list. Paths in the call-site column are relative to `cohere/internal/lint/rules/`.

| Missing Go call or shared helper, as called upstream | Blocked rules | Call sites and required answer |
| --- | --- | --- |
| `typeChecker.GetGlobalSymbol(name, ast.SymbolFlagsAll, nil)` | no-redeclare | `typescript/builtin_globals.go:59,86`; global symbol, ordered declarations and default-library provenance. |
| `ctx.TypeChecker.GetSymbolAtLocation(...)` | correctness-no-test-on-global-regex; correctness-no-write-only-collection; correctness-no-process-exit-after-output; correctness-no-uncleared-race-timeout; correctness-require-blocking-standard-streams; no-new-func; no-new-native-nonconstructor; no-new-wrappers; prefer-rest-params; exhaustive-deps; jsx-fragments; jsx-no-constructed-context-values; jsx-no-undef | Resolved symbol presence, stable identity and full ordered declarations. Representative sites: `nexus/correctness_no_test_on_global_regex.go:118,322`, `core/no_new_native_nonconstructor.go:103`, `core/prefer_rest_params.go:114`, `react/exhaustive_deps.go:1495,1510,1892`, `react/jsx_fragments.go:290`, `react/jsx_no_constructed_context_values.go:320,570`, `react/jsx_no_undef.go:101`. The blocking-standard-streams receiver is `analysis.ctx.TypeChecker`. |
| `ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent)` | correctness-no-write-only-collection; correctness-no-uncleared-race-timeout; exhaustive-deps | `nexus/correctness_no_write_only_collection.go:124`, `nexus/correctness_no_uncleared_race_timeout.go:337`, `react/exhaustive_deps.go:1888`; resolve the shorthand value binding rather than its property symbol. |
| `ctx.TypeChecker.GetAliasedSymbol(symbol)`; `analysis.ctx.TypeChecker.GetAliasedSymbol(symbol)` | correctness-no-process-exit-after-output; correctness-require-blocking-standard-streams | `nexus/correctness_no_process_exit_after_output.go:413`, `nexus/correctness_require_blocking_standard_streams.go:586,600`; alias target and declaration provenance. |
| `resolvesToAGlobal(ctx, identifier)` | no-new-func; no-new-native-nonconstructor; no-new-wrappers; prefer-regex-literals | Definition: `core/no_new_native_nonconstructor.go:99`; use: `core/prefer_regex_literals.go:287`. Requires a nonnil symbol with at least one declaration and the first declaration's source file `IsDeclarationFile`. |
| `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)` | prefer-regex-literals | `core/prefer_regex_literals.go:196`; shared reference tracking, including aliases and global constructor references. |
| `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` | set-state-in-effect | `react/set_state_in_effect.go:267`; shared native high-level IR and its analysis data. |
| `high_level_intermediate_representation.ForFunction(ctx, functionNode)` | set-state-in-render; static-components | `react/set_state_in_render.go:183`, `react/static_components.go:120`; shared native high-level IR, single assignment and capture analysis. |

## Why existing answers do not substitute

The area has `context.checker`; the checker itself is not the blocker. Its `symbol-origin` answer emits the filename of `symbol.ValueDeclaration`, and emits an empty answer for either a missing symbol or a symbol without a value declaration. It cannot distinguish implicit `arguments`, type-only symbols, or declarationless symbols from unresolved references. It provides neither symbol identity nor all declarations. `type-origin` describes the type's symbol rather than the identifier's binding. The `declarations` selector accepts class/interface declarations, not arbitrary identifier bindings. `scope-locals` is a binder table, not reference resolution, alias following or global lookup. The shared high-level IR and reference-tracker helpers above have inventory entries, but no callable Adamic implementation in the inspected lint sources.

## Per-rule reproducer locations

All paths in this table are relative to the evidence directory at commit `8f290f3cbe37becb196319b193e852f396d0800e`, named above. Each adjacent `BLOCKED.md` records the exact dependency and reproducer purpose. These inputs document missing questions, not successful new-rule parity or caught new mutants.

| Claimed rule | Reproducer |
| --- | --- |
| @typescript-eslint/no-redeclare | `no-redeclare/reproducer.ts.txt` |
| nexus/correctness-no-test-on-global-regex | `correctness-no-test-on-global-regex/reproducer.ts.txt` |
| nexus/correctness-no-write-only-collection | `correctness-no-write-only-collection/reproducer.ts.txt` |
| nexus/correctness-no-process-exit-after-output | `correctness-no-process-exit-after-output/reproducer.ts.txt` |
| nexus/correctness-no-uncleared-race-timeout | `correctness-no-uncleared-race-timeout/reproducer.ts.txt` |
| nexus/correctness-require-blocking-standard-streams | `correctness-require-blocking-standard-streams/reproducer.ts.txt` |
| no-new-func | `no-new-func/reproducer.ts.txt` |
| no-new-native-nonconstructor | `no-new-native-nonconstructor/reproducer.ts.txt` |
| no-new-wrappers | `no-new-wrappers/reproducer.ts.txt` |
| prefer-regex-literals | `prefer-regex-literals/reproducer.ts.txt` |
| prefer-rest-params | `prefer-rest-params/reproducer.ts.txt` |
| react-hooks/exhaustive-deps | `exhaustive-deps/reproducer.ts.txt` |
| react-hooks/set-state-in-effect | `set-state-in-effect/reproducer.ts.txt` |
| react-hooks/set-state-in-render | `set-state-in-render/reproducer.ts.txt` |
| react-hooks/static-components | `static-components/reproducer.tsx.txt` |
| react/jsx-fragments | `jsx-fragments/reproducer.tsx.txt` |
| react/jsx-no-constructed-context-values | `jsx-no-constructed-context-values/reproducer.tsx.txt` |
| react/jsx-no-undef | `jsx-no-undef/reproducer.tsx.txt` |

## Legacy bridge wire APIs absent from area

These questions exist and have callers on the legacy branch. They were not landed in the inspected area checker; this is not evidence that integration removed them. All Go paths below are relative to `bridge/tsgo/checker/`, and caller paths are relative to `stage1/cohere/typeaware/` on the legacy branch. Moving them requires adapting the shared harness recording/replay contract and the current TypeScript pin, not installing a private checker.

| Legacy question | Legacy implementation | Legacy caller |
| --- | --- | --- |
| `global-symbol-details` | `global_symbol_details.go` | `global_symbol_details.a` |
| `scope-metadata` | `scope_metadata.go` | `scope_metadata.a`, `wave12_next/source_files.a` |
| `node-symbol-details` | `declaration_facts.go`, dispatch in `facts.go` | `wave_12_symbols.a`, `caller.ts` |
| `symbol-identities` | `symbol_identity.go`, dispatch in `facts.go` | `wave_12_symbols.a`, `unused.ts` |
| `binding-declarations`, `alias-declarations` | `facts.go` | `bindings.ts`, `flow.ts`, `property_alias.ts`, `reassign.ts` |
| `identifier-binding` | `identifier_binding.go` | `wave12_fourth/identifier_binding.a` |
| `symbol-declaration-files` | `symbol_declaration_files.go` | `wave12_third/symbol_declaration_files.a` |
| `call-declaration` | `call_declaration.go` | `wave12_next/process_facts.a` |
| `program-imports` | `program_imports.go` | `wave12_next/program_imports.a` |
| `symbol-ancestry` | `symbol_ancestry.go` | `wave12_next/symbol_ancestry.a` |
| `source-has-jsx` | `source_has_jsx.go` | `wave12_fourth/source_has_jsx.a` |
| `resolved-name` | `facts.go` | `coercion.ts` |
| `type-metadata` | `facts.go` | `checker_facts.ts` |

Additional used legacy questions from the inherited type-aware helper layer, also absent from area:

| Legacy questions | Legacy implementation | Legacy callers |
| --- | --- | --- |
| `declaration-details`, `type-symbol-details`, `property-declarations` | `declaration_facts.go`, dispatch in `facts.go` | `caller.ts` |
| `symbol-shape`, `container-bases` | `facts.go` | `caller.ts`, `property_alias.ts` |
| `annotation-shape` | `facts.go` | `casts.ts`, `flow.ts`, `property_alias.ts` |
| `type-properties`, `identical-types` | `facts.go` | `checker_facts.ts` |
| `literal-value` | `facts.go` | `coercion.ts` |
| `function-signatures`, `property-exists`, `reference-shape`, `annotated-return-shape`, `contextual-argument` | `facts.go` | `flow.ts` |

Legacy extension registration uses `questionExtensions` and `(*Program).inspectExtension` in `question_extensions.go`; that dispatcher is also absent from area. Other legacy-supported question names are not listed as used APIs without a source caller. The raw program/inspect/release handle mechanism is not reported missing: the shared checker already owns that lifecycle.

Validation for this documentation-only branch: upstream call sites and both bridge dispatch tables inspected; exactly one new Markdown file in the commit. No porting or tests rerun.
