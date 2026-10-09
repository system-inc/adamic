Whole program: 1/79
Own file: 26/79

Stage 3 meter 20261007T201631Z

| File | Whole program | Own file | Lowering |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.ts | fail | pass | blocked |
| src/compiler/binder.ts | fail | fail | blocked |
| src/compiler/builder.ts | fail | fail | blocked |
| src/compiler/builderPublic.ts | fail | pass | blocked |
| src/compiler/builderState.ts | fail | fail | blocked |
| src/compiler/builderStatePublic.ts | fail | pass | blocked |
| src/compiler/checker.ts | fail | fail | blocked |
| src/compiler/commandLineParser.ts | fail | fail | blocked |
| src/compiler/core.ts | fail | fail | blocked |
| src/compiler/corePublic.ts | fail | fail | blocked |
| src/compiler/debug.ts | fail | fail | blocked |
| src/compiler/diagnosticInformationMap.generated.ts | fail | pass | blocked |
| src/compiler/diagnosticMessages.generated.json | fail | fail | blocked |
| src/compiler/diagnosticMessages.json | fail | fail | blocked |
| src/compiler/emitter.ts | fail | fail | blocked |
| src/compiler/executeCommandLine.ts | fail | fail | blocked |
| src/compiler/expressionToTypeNode.ts | fail | fail | blocked |
| src/compiler/factory/baseNodeFactory.ts | fail | pass | blocked |
| src/compiler/factory/emitHelpers.ts | fail | fail | blocked |
| src/compiler/factory/emitNode.ts | fail | pass | blocked |
| src/compiler/factory/nodeChildren.ts | fail | pass | blocked |
| src/compiler/factory/nodeConverters.ts | fail | pass | blocked |
| src/compiler/factory/nodeFactory.ts | fail | fail | blocked |
| src/compiler/factory/nodeTests.ts | fail | pass | blocked |
| src/compiler/factory/parenthesizerRules.ts | fail | fail | blocked |
| src/compiler/factory/utilities.ts | fail | fail | blocked |
| src/compiler/factory/utilitiesPublic.ts | fail | pass | blocked |
| src/compiler/hostErrors.ts | pass | pass | fail |
| src/compiler/moduleNameResolver.ts | fail | fail | blocked |
| src/compiler/moduleSpecifiers.ts | fail | fail | blocked |
| src/compiler/parser.ts | fail | fail | blocked |
| src/compiler/path.ts | fail | fail | blocked |
| src/compiler/performance.ts | fail | pass | blocked |
| src/compiler/performanceCore.ts | fail | fail | blocked |
| src/compiler/program.ts | fail | fail | blocked |
| src/compiler/programDiagnostics.ts | fail | pass | blocked |
| src/compiler/resolutionCache.ts | fail | fail | blocked |
| src/compiler/scanner.ts | fail | fail | blocked |
| src/compiler/semver.ts | fail | pass | blocked |
| src/compiler/sourcemap.ts | fail | fail | blocked |
| src/compiler/symbolWalker.ts | fail | pass | blocked |
| src/compiler/sys.ts | fail | fail | blocked |
| src/compiler/tracing.ts | fail | fail | blocked |
| src/compiler/transformer.ts | fail | fail | blocked |
| src/compiler/transformers/classFields.ts | fail | fail | blocked |
| src/compiler/transformers/classThis.ts | fail | pass | blocked |
| src/compiler/transformers/declarations.ts | fail | fail | blocked |
| src/compiler/transformers/declarations/diagnostics.ts | fail | fail | blocked |
| src/compiler/transformers/destructuring.ts | fail | fail | blocked |
| src/compiler/transformers/es2015.ts | fail | fail | blocked |
| src/compiler/transformers/es2016.ts | fail | pass | blocked |
| src/compiler/transformers/es2017.ts | fail | fail | blocked |
| src/compiler/transformers/es2018.ts | fail | fail | blocked |
| src/compiler/transformers/es2019.ts | fail | pass | blocked |
| src/compiler/transformers/es2020.ts | fail | pass | blocked |
| src/compiler/transformers/es2021.ts | fail | pass | blocked |
| src/compiler/transformers/esDecorators.ts | fail | fail | blocked |
| src/compiler/transformers/esnext.ts | fail | fail | blocked |
| src/compiler/transformers/generators.ts | fail | fail | blocked |
| src/compiler/transformers/jsx.ts | fail | fail | blocked |
| src/compiler/transformers/legacyDecorators.ts | fail | fail | blocked |
| src/compiler/transformers/module/esnextAnd2015.ts | fail | fail | blocked |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | fail | pass | blocked |
| src/compiler/transformers/module/module.ts | fail | fail | blocked |
| src/compiler/transformers/module/system.ts | fail | fail | blocked |
| src/compiler/transformers/namedEvaluation.ts | fail | pass | blocked |
| src/compiler/transformers/taggedTemplate.ts | fail | fail | blocked |
| src/compiler/transformers/ts.ts | fail | fail | blocked |
| src/compiler/transformers/typeSerializer.ts | fail | pass | blocked |
| src/compiler/transformers/utilities.ts | fail | fail | blocked |
| src/compiler/tsbuild.ts | fail | fail | blocked |
| src/compiler/tsbuildPublic.ts | fail | fail | blocked |
| src/compiler/tsconfig.json | fail | fail | blocked |
| src/compiler/types.ts | fail | fail | blocked |
| src/compiler/utilities.ts | fail | fail | blocked |
| src/compiler/utilitiesPublic.ts | fail | fail | blocked |
| src/compiler/visitorPublic.ts | fail | pass | blocked |
| src/compiler/watch.ts | fail | fail | blocked |
| src/compiler/watchPublic.ts | fail | fail | blocked |
| src/compiler/watchUtilities.ts | fail | fail | blocked |

Checker: 1/79 source files. Lowering: 0/79 source files.
All files: 82; JSON inputs fail the source-extension gate.
Blocked means lowering was not attempted because the checker/input gate failed.

Latent lowering, main: measured on a checker-rejected program.
NotYet: 2081; Refused: 2136.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| the non-null assertion ! | 01a1130a | 0 | 521 | 521 |
| reading SyntaxKind | compiler/stage3-front | 484 | 0 | 484 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | 01a1143b-d691 | 0 | 384 | 384 |
| a function without a body | 01a1143b-f5d4 | 191 | 0 | 191 |
| enum | compiler/stage3-front | 0 | 162 | 162 |
| an EnumDeclaration | compiler/stage3-front | 154 | 0 | 154 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 88 | 0 | 88 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 84 | 0 | 84 |
| an ExportDeclaration | 01a113e3-a058 | 0 | 77 | 77 |
| a type predicate whose return is not proven (there is no body proving this parameter) | 01a1143b-d691 | 0 | 71 | 71 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.
