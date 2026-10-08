Whole program: 2/79
Own file: 57/79

Stage 3 meter 20261008T012240Z

| File | Whole program | Own file | Lowering |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.ts | fail | pass | blocked |
| src/compiler/binder.ts | fail | pass | blocked |
| src/compiler/builder.ts | fail | fail | blocked |
| src/compiler/builderPublic.ts | fail | pass | blocked |
| src/compiler/builderState.ts | fail | pass | blocked |
| src/compiler/builderStatePublic.ts | fail | pass | blocked |
| src/compiler/checker.ts | fail | fail | blocked |
| src/compiler/commandLineParser.ts | fail | fail | blocked |
| src/compiler/core.ts | fail | fail | blocked |
| src/compiler/corePublic.ts | pass | pass | fail |
| src/compiler/debug.ts | fail | pass | blocked |
| src/compiler/diagnosticInformationMap.generated.ts | fail | pass | blocked |
| src/compiler/diagnosticMessages.generated.json | fail | fail | blocked |
| src/compiler/diagnosticMessages.json | fail | fail | blocked |
| src/compiler/emitter.ts | fail | pass | blocked |
| src/compiler/executeCommandLine.ts | fail | pass | blocked |
| src/compiler/expressionToTypeNode.ts | fail | pass | blocked |
| src/compiler/factory/baseNodeFactory.ts | fail | pass | blocked |
| src/compiler/factory/emitHelpers.ts | fail | pass | blocked |
| src/compiler/factory/emitNode.ts | fail | pass | blocked |
| src/compiler/factory/nodeChildren.ts | fail | pass | blocked |
| src/compiler/factory/nodeConverters.ts | fail | pass | blocked |
| src/compiler/factory/nodeFactory.ts | fail | pass | blocked |
| src/compiler/factory/nodeTests.ts | fail | pass | blocked |
| src/compiler/factory/parenthesizerRules.ts | fail | pass | blocked |
| src/compiler/factory/utilities.ts | fail | pass | blocked |
| src/compiler/factory/utilitiesPublic.ts | fail | pass | blocked |
| src/compiler/hostErrors.ts | pass | pass | fail |
| src/compiler/moduleNameResolver.ts | fail | fail | blocked |
| src/compiler/moduleSpecifiers.ts | fail | pass | blocked |
| src/compiler/parser.ts | fail | pass | blocked |
| src/compiler/path.ts | fail | pass | blocked |
| src/compiler/performance.ts | fail | pass | blocked |
| src/compiler/performanceCore.ts | fail | pass | blocked |
| src/compiler/program.ts | fail | fail | blocked |
| src/compiler/programDiagnostics.ts | fail | pass | blocked |
| src/compiler/resolutionCache.ts | fail | fail | blocked |
| src/compiler/scanner.ts | fail | pass | blocked |
| src/compiler/semver.ts | fail | pass | blocked |
| src/compiler/sourcemap.ts | fail | fail | blocked |
| src/compiler/symbolWalker.ts | fail | pass | blocked |
| src/compiler/sys.ts | fail | fail | blocked |
| src/compiler/tracing.ts | fail | pass | blocked |
| src/compiler/transformer.ts | fail | fail | blocked |
| src/compiler/transformers/classFields.ts | fail | fail | blocked |
| src/compiler/transformers/classThis.ts | fail | pass | blocked |
| src/compiler/transformers/declarations.ts | fail | fail | blocked |
| src/compiler/transformers/declarations/diagnostics.ts | fail | pass | blocked |
| src/compiler/transformers/destructuring.ts | fail | pass | blocked |
| src/compiler/transformers/es2015.ts | fail | fail | blocked |
| src/compiler/transformers/es2016.ts | fail | pass | blocked |
| src/compiler/transformers/es2017.ts | fail | fail | blocked |
| src/compiler/transformers/es2018.ts | fail | pass | blocked |
| src/compiler/transformers/es2019.ts | fail | pass | blocked |
| src/compiler/transformers/es2020.ts | fail | pass | blocked |
| src/compiler/transformers/es2021.ts | fail | pass | blocked |
| src/compiler/transformers/esDecorators.ts | fail | fail | blocked |
| src/compiler/transformers/esnext.ts | fail | fail | blocked |
| src/compiler/transformers/generators.ts | fail | fail | blocked |
| src/compiler/transformers/jsx.ts | fail | fail | blocked |
| src/compiler/transformers/legacyDecorators.ts | fail | pass | blocked |
| src/compiler/transformers/module/esnextAnd2015.ts | fail | fail | blocked |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | fail | pass | blocked |
| src/compiler/transformers/module/module.ts | fail | pass | blocked |
| src/compiler/transformers/module/system.ts | fail | pass | blocked |
| src/compiler/transformers/namedEvaluation.ts | fail | pass | blocked |
| src/compiler/transformers/taggedTemplate.ts | fail | pass | blocked |
| src/compiler/transformers/ts.ts | fail | pass | blocked |
| src/compiler/transformers/typeSerializer.ts | fail | pass | blocked |
| src/compiler/transformers/utilities.ts | fail | pass | blocked |
| src/compiler/tsbuild.ts | fail | pass | blocked |
| src/compiler/tsbuildPublic.ts | fail | fail | blocked |
| src/compiler/tsconfig.json | fail | fail | blocked |
| src/compiler/types.ts | fail | pass | blocked |
| src/compiler/utilities.ts | fail | pass | blocked |
| src/compiler/utilitiesPublic.ts | fail | pass | blocked |
| src/compiler/visitorPublic.ts | fail | pass | blocked |
| src/compiler/watch.ts | fail | fail | blocked |
| src/compiler/watchPublic.ts | fail | fail | blocked |
| src/compiler/watchUtilities.ts | fail | pass | blocked |

Checker: 2/79 source files. Lowering: 0/79 source files.
All files: 82; JSON inputs fail the source-extension gate.
Blocked means lowering was not attempted because the checker/input gate failed.

Latent lowering, area: measured on a checker-rejected program.
NotYet: 1084; Refused: 2746.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 451 | 451 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | 01a1143b-d691 | 0 | 388 | 388 |
| an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) | OWNER BLANK | 0 | 169 | 169 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 129 | 0 | 129 |
| an unproven predicate argument for parameter test (argument &quot;isTypeNode&quot;) | OWNER BLANK | 0 | 75 | 75 |
| reading Debug | 01a113e7-bfed | 69 | 0 | 69 |
| a function returning T &#124; undefined | 01a1143c | 53 | 0 | 53 |
| a BinaryExpression with a value and a boolean | OWNER BLANK | 50 | 0 | 50 |
| a function returning T | 01a1143c | 45 | 0 | 45 |
| a BinaryExpression with a value and a value | OWNER BLANK | 42 | 0 | 42 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.
