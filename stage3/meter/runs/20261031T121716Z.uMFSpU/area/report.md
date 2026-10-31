Whole program: 2/79
Own file: 58/79

Stage 3 meter 20261031T121716Z

| File | Whole program | Own file | Lowering |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | blocked |
| src/compiler/_namespaces/ts.ts | fail | pass | blocked |
| src/compiler/binder.ts | fail | pass | blocked |
| src/compiler/builder.ts | fail | pass | blocked |
| src/compiler/builderPublic.ts | fail | pass | blocked |
| src/compiler/builderState.ts | fail | pass | blocked |
| src/compiler/builderStatePublic.ts | fail | pass | blocked |
| src/compiler/checker.ts | fail | fail | blocked |
| src/compiler/commandLineParser.ts | fail | fail | blocked |
| src/compiler/core.ts | fail | fail | blocked |
| src/compiler/corePublic.ts | pass | pass | pass |
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
| src/compiler/hostErrors.ts | pass | pass | pass |
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

Checker: 2/79 source files. Lowering: 2/79 source files.
All files: 82; JSON inputs fail the source-extension gate.
Blocked means lowering was not attempted because the checker/input gate failed.

tsc entry lowering, area: blocked by checker diagnostics.

Latent lowering, area: measured on a checker-rejected program.
NotYet: 651; Refused: 2115.
Recovery boundaries: 0; named nested checker exclusions: 0.
Errors: 0; panics: 1.
- panic `src/compiler/factory/nodeFactory.ts:492:1`: runtime error: invalid memory address or nil pointer dereference, runtime error: invalid memory address or nil pointer dereference

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 534 | 534 |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 509 | 509 |
| a union or optional field read in a program with record storage | OWNER BLANK | 74 | 0 | 74 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 47 | 0 | 47 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 32 | 32 |
| a call through ?. (an optional call) | OWNER BLANK | 27 | 0 | 27 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 27 | 0 | 27 |
| a value of type any | adaptation 40, stage 3 | 25 | 0 | 25 |
| an ESM namespace object used as a value; use qualified live bindings | OWNER BLANK | 19 | 0 | 19 |
| a method read as a value (liftToBlock would lose its object, and this with it) | OWNER BLANK | 0 | 14 | 14 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

This box's times, not comparable across boxes: 1046.900 wall seconds; 1112.090 CPU seconds.

| Slowest files | Wall seconds | CPU seconds |
| --- | ---: | ---: |
| src/compiler/utilities.ts | 341.801 | 355.033 |
| src/compiler/types.ts | 159.342 | 154.193 |
| src/compiler/commandLineParser.ts | 86.386 | 89.111 |
| src/compiler/builder.ts | 66.943 | 69.150 |
| src/compiler/moduleNameResolver.ts | 57.328 | 61.715 |
