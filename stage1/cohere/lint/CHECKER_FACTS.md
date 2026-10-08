# Checker facts and API migration

Baseline: `c8f6d74f5387df926ae12d8babd2780564b60097` on area/stage1-lint. Surveyed all 30 remote wave tips and the audits below. Counts are distinct claimed rules named by audits, not inherited branch occurrences. Supplying a fact removes that dependency; it does not certify a complete rule or its later analysis dependencies.

- `origin/lint-checker/audit-wave-18` at `e9956e769fd9d430056977022af0d1079b4abe60`, `stage1/cohere/lint/checker-audit/wave-18.md`.
- `origin/lint-checker/audit-wave-29` at `aaf52ab7ff55f88c06ff2d0e4e4caa3dffd7dddd`, `stage1/cohere/lint/checker-audit/wave-29.md`.
- `origin/lint-checker/audit-wave-30` at `bb092f8d30ad33e17bc35b6d4fac3790beec36b5`, `stage1/cohere/lint/checker-audit/wave-30.md`.

## One fact and API table

| Fact / old API | Available through context.checker? | Owner / migration | Audited distinct rule count and evidence |
| --- | --- | --- | --- |
| `ctx.TypeChecker.GetSymbolAtLocation` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 16: `id-denylist`, `id-match`, `nexus/concurrency-no-check-then-write`, `nexus/consistency-no-iso-string-date-cut`, `nexus/correctness-no-callback-in-parse-try`, `nexus/correctness-no-collection-misuse`, `nexus/correctness-no-discarded-pure-result`, `nexus/correctness-no-process-exit-after-output`, `nexus/correctness-no-uncleared-race-timeout`, `nexus/correctness-require-blocking-standard-streams`, `no-restricted-globals`, `no-setter-return`, `no-shadow-restricted-names`, `react/jsx-fragments`, `react/jsx-no-constructed-context-values`, `react/jsx-no-undef` |
| ctx.TypeChecker.GetSymbolAtLocation; ctx.TypeChecker.GetShorthandAssignmentValueSymbol; rule.DeclarationsIn | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 6: `no-class-assign`, `no-const-assign`, `no-constant-binary-expression`, `no-new-native-nonconstructor`, `prefer-promise-reject-errors`, `prefer-rest-params` |
| `high_level_intermediate_representation.ForFunction` | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 3: `react-hooks/purity`, `react-hooks/set-state-in-render`, `react-hooks/static-components` |
| `type_checking.IsSymbolFromDefaultLibrary` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 3: `nexus/consistency-no-iso-string-date-cut`, `nexus/correctness-no-collection-misuse`, `nexus/correctness-no-uncleared-race-timeout` |
| high_level_intermediate_representation.MayHoldComponentOrHook | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 3: `react-hooks/set-state-in-effect`, `react-hooks/set-state-in-render`, `react-hooks/static-components` |
| `ctx.TypeChecker.GetAliasedSymbol` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 2: `nexus/correctness-no-process-exit-after-output`, `nexus/correctness-require-blocking-standard-streams` |
| `ctx.TypeChecker.GetShorthandAssignmentValueSymbol` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 2: `nexus/correctness-no-callback-in-parse-try`, `nexus/correctness-no-uncleared-race-timeout` |
| `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 2: `react-hooks/refs`, `react-hooks/set-state-in-effect` |
| `reference.WritesToBinding` | Missing required payload/helper; related narrow questions can exist | Harness native reference analysis, not a checker verdict | 2: `nexus/concurrency-no-check-then-write`, `no-shadow-restricted-names` |
| `typeChecker.GetShorthandAssignmentValueSymbol`; `ctx.TypeChecker.GetShorthandAssignmentValueSymbol` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 2: `no-restricted-globals`, `no-shadow-restricted-names` |
| `type_checking.IsSourceFileDefaultLibrary` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 2: `nexus/correctness-no-discarded-pure-result`, `nexus/correctness-no-uncleared-race-timeout` |
| ctx.TypeChecker.GetSymbolAtLocation; ctx.TypeChecker.GetShorthandAssignmentValueSymbol | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 2: `react/static-property-placement`, `react/style-prop-object` |
| high_level_intermediate_representation.ForFunction | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 2: `react-hooks/set-state-in-render`, `react-hooks/static-components` |
| `checker.Checker_getAwaitedType` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `nexus/correctness-no-discarded-outcome` |
| `checker.SkipAlias` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `nexus/concurrency-no-check-then-write` |
| `checker.Type_symbol; symbol.Declarations; declaration.Parent` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `nexus/correctness-no-discarded-outcome` |
| `declarationAnchoredAt`; `resolvesToDeclaration`; `rule.DeclarationsIn` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `no-shadow-restricted-names` |
| `descriptor.IsFunctionUnder` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `no-setter-return` |
| `hir.ForFunction; hir.CloneFunction; hir.AnalyzePreservedManualMemoization` | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 1: `react-hooks/preserve-manual-memoization` |
| `reference.IsValueReference`; `reference.ReadSymbol` | Missing required payload/helper; related narrow questions can exist | Harness native reference analysis, not a checker verdict | 1: `no-restricted-globals` |
| `rule.IsDeclaredOnlyInDeclarationFiles` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `no-setter-return` |
| `typeChecker.GetExportSpecifierLocalTargetSymbol` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `no-restricted-globals` |
| `writers.ctx.TypeChecker.GetResolvedSignature; writers.ctx.TypeChecker.GetReturnTypeOfSignature` | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `nexus/correctness-no-process-exit-after-output` |
| ctx.TypeChecker.GetDeclaredTypeOfSymbol; ctx.TypeChecker.GetBaseTypes; ctx.TypeChecker.GetPropertyOfType | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `@typescript-eslint/class-literal-property-style` |
| ctx.TypeChecker.GetSymbolAtLocation | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `react/jsx-no-undef` |
| ctx.TypeChecker.IsArrayLikeType; ctx.TypeChecker.GetNumberIndexType; ctx.TypeChecker.GetTypeArguments; type_checking.GetWellKnownSymbolPropertyOfType; type_checking.NeedsToBeAwaited | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `@typescript-eslint/await-thenable` |
| high_level_intermediate_representation.ForFunctionWithoutManualMemoization | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 1: `react-hooks/set-state-in-effect` |
| jsxNoConstructedContextValuesCheckMemo | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 1: `react/jsx-no-constructed-context-values` |
| jsxNoConstructedContextValuesStaysHome | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 1: `react/jsx-no-constructed-context-values` |
| regexpattern.Walk; regexsyntax.ParseRegexFlags; regexsyntax.SkipPatternEscape; regexsyntax.ClassEnd | Missing required payload/helper; related narrow questions can exist | Bridge raw facts, exposed by ask / declared askFile | 1: `prefer-regex-literals` |
| walk.anyEscapes (receiver *jsxNoConstructedContextValuesStabilityWalk) | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 1: `react/jsx-no-constructed-context-values` |
| walk.functionEvaluation (receiver *jsxNoConstructedContextValuesStabilityWalk) | Missing required payload/helper; related narrow questions can exist | Native shared analysis; React compiler track, outside this unit | 1: `react/jsx-no-constructed-context-values` |
| `accessed-property` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `alias-declarations` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `annotated-return-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `annotation-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `apparent-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `assignable` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `assignable-types` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `ast-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `awaited-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/declaration_ancestry.go`, `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 2 wave tips |
| `awaited-type-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `base-constraint-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `base-member-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `base-shapes` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `base-type` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `binding-declarations` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `binding-origin` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 2 wave tips |
| `binding-state` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `binding-structure` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `call-count` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `call-declaration` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `call-declaration-chain` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `call-parameter-returns` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `call-parameters` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `call-returns` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `callable-type-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `callback-parameters` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `callback-symbol` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `character-properties` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `class-heritage-names` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `component-property-syntax-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `constraint-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `constructor-expression` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `container-bases` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `contextual-argument` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `contextual-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `declaration-chain` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `declaration-contract` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `declaration-details` | Missing question contract | Bridge; `bridge/tsgo/checker/declaration_facts.go`, `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `declaration-file-flags` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `declaration-lineage` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 2 wave tips |
| `declarations` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `declared-call-signature` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `enum-declarations` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `enum-types` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `execution-graph` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `export-module-properties` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `export-symbol-chain` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `function-signatures` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `global-binding` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `global-source` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `identical-types` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `index-signature-access` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `interface-bases` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `isolated-declarations` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `iteration-type-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `jsx-structure` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `jsx-syntax-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `literal-string` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `literal-value` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `member-parameters` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `module-links` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go`, `bridge/tsgo/checker/wave08_facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `module-records` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `name` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `namespace-binding` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `node-structure` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `node-symbol-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `node-symbol-details` | Missing question contract | Bridge; `bridge/tsgo/checker/declaration_facts.go`, `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `node-symbol-origin` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `nonnullable-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `number-index-type` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `numeric-literal` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `numeric-syntax-bindings` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `options` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `output-callee` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `output-flow` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `output-symbol` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `platform-symbol` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `preference-binding` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `preference-structure` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `process-node-fields` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `program-imports` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go`, `bridge/tsgo/checker/process_questions.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `program-module-resolution` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `program-modules` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 3 wave tips |
| `property-declarations` | Missing question contract | Bridge; `bridge/tsgo/checker/declaration_facts.go`, `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `property-exists` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `property-info` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `property-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `raw-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `raw-type` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `react-hir` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `react-syntax-details` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `react-type-names` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `read-symbol` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `reference-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `reference-node` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `reference-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `reference-symbol-origins` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `regex-pattern` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `regex-pattern-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `regexp-program` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `regular-expression-syntax` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `require-await` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `resolved-call` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `resolved-call-target` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `resolved-callee` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `resolved-declaration` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `resolved-name` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `resolved-signature-equal` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `return-type-sources` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `runtime-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `runtime-modules` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `scope-export-symbols` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `scope-locals` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `scope-symbols` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `signature` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/declaration_context.go`, `bridge/tsgo/checker/declaration_lineage.go`, `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `signature-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `signature-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `source-comments` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `source-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `source-parse-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `strict-this` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `stringification-type` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-ancestry` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-context` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go`, `bridge/tsgo/checker/wave_03_next_questions.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-declaration-paths` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-declaration-provenance` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-declaration-syntax` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-identities` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `symbol-lineage` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-locations` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-origin` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `symbol-provenance` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `symbol-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `symbols-in-scope` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `syntax-metadata` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go`, `bridge/tsgo/checker/wave08_facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `transformed-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/declaration_lineage.go`, `bridge/tsgo/checker/facts.go`, `bridge/tsgo/checker/require_await_facts.go`, `bridge/tsgo/checker/wave_27_checker_links.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `type-arguments` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type-declaration-ancestors` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type-declaration-ancestry` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 2 wave tips |
| `type-leaf-facts` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type-metadata` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `type-operations` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type-origin` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `type-projection` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type-properties` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `type-reference-graph` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `type-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `type-symbol` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `type-symbol-details` | Missing question contract | Bridge; `bridge/tsgo/checker/declaration_facts.go`, `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| `unicode-node-text` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `usage-shape` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `void-position-types` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `wave-27-call-declaration` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `wave-27-checker-links` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `wave-27-declaration-ancestry` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `wave-27-module-sources` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `wave-27-syntax-flow` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `wave08-symbol` | Missing question contract | Bridge; `bridge/tsgo/checker/facts.go`, `bridge/tsgo/checker/wave08_facts.go` | No audited per-question count asserted; inherited on 1 wave tips |
| `widened-shape` | Present via `ask`; file-wide reads use guarded `askFile` | Bridge; `bridge/tsgo/checker/facts.go` | No audited per-question count asserted; inherited on 30 wave tips |
| tsgoProgram / checker.Open | Harness opens one program per manifest run | Harness; never rule-owned | All typed rules |
| tsgoInspect / Program.Inspect | context.checker.ask(index, question) | Harness records/replays selectors and answers; bridge supplies facts | All typed rules |
| tsgoRelease | Harness releases after final row | Harness; borrowed type/symbol IDs last for the run | All typed rules |
| tsgoQuery / Program.Query / TypeParts | Low-level APIs exist; no facade, no direct audited rule demand | Bridge facts require explicit questions before adding a facade | Not counted as active demand |
| SourceFile.FileName() as string | Obsolete on old branches | Bridge wire boundary uses .AsString() | Old declaration serializers and worker questions |
| SourceFile.Path() | Obsolete | Bridge default-library check uses .PathKey() | Old declaration serializers and provenance questions |
| GetSourceFileForResolvedModule(path) | Obsolete | Bridge passes the resolved module object | Wave 30 module-resolution question |

## Priority and proof scope

First priority is symbol presence, stable identity and complete declaration records, including shorthand value bindings and alias targets. Keep default-library and declaration-file flags in the records. Next come native reference analysis and the worker raw syntax questions. React HIR, SSA, captures and memoization are routed to the React compiler track and will never be answered with Go lint verdicts here.

The old branch sources are implementation candidates, not passing evidence at this pin. New questions must be compared through unchanged Go rules on source Node, emitted JavaScript and sanitized native, with actual mutants. A fact needing an absent typescript-go accessor is named and left out rather than approximated.

No private program or fact replay driver is introduced. Existing program lifetimes, transcript selector/header hashes, missing-entry refusal, extra-entry guard, fix-pass exclusion and declared programReads remain.
