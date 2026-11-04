Whole program: 79/79
Own file: 79/79

Stage 3 meter 20261104T131717Z

| File | Whole program | Own file | Lowering |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | pass | pass | fail |
| src/compiler/_namespaces/ts.performance.ts | pass | pass | fail |
| src/compiler/_namespaces/ts.ts | pass | pass | fail |
| src/compiler/binder.ts | pass | pass | fail |
| src/compiler/builder.ts | pass | pass | fail |
| src/compiler/builderPublic.ts | pass | pass | fail |
| src/compiler/builderState.ts | pass | pass | fail |
| src/compiler/builderStatePublic.ts | pass | pass | fail |
| src/compiler/checker.ts | pass | pass | fail |
| src/compiler/commandLineParser.ts | pass | pass | fail |
| src/compiler/core.ts | pass | pass | fail |
| src/compiler/corePublic.ts | pass | pass | pass |
| src/compiler/debug.ts | pass | pass | fail |
| src/compiler/diagnosticInformationMap.generated.ts | pass | pass | fail |
| src/compiler/diagnosticMessages.generated.json | fail | fail | blocked |
| src/compiler/diagnosticMessages.json | fail | fail | blocked |
| src/compiler/emitter.ts | pass | pass | fail |
| src/compiler/executeCommandLine.ts | pass | pass | fail |
| src/compiler/expressionToTypeNode.ts | pass | pass | fail |
| src/compiler/factory/baseNodeFactory.ts | pass | pass | fail |
| src/compiler/factory/emitHelpers.ts | pass | pass | fail |
| src/compiler/factory/emitNode.ts | pass | pass | fail |
| src/compiler/factory/nodeChildren.ts | pass | pass | fail |
| src/compiler/factory/nodeConverters.ts | pass | pass | fail |
| src/compiler/factory/nodeFactory.ts | pass | pass | fail |
| src/compiler/factory/nodeTests.ts | pass | pass | fail |
| src/compiler/factory/parenthesizerRules.ts | pass | pass | fail |
| src/compiler/factory/utilities.ts | pass | pass | fail |
| src/compiler/factory/utilitiesPublic.ts | pass | pass | fail |
| src/compiler/hostErrors.ts | pass | pass | pass |
| src/compiler/moduleNameResolver.ts | pass | pass | fail |
| src/compiler/moduleSpecifiers.ts | pass | pass | fail |
| src/compiler/parser.ts | pass | pass | fail |
| src/compiler/path.ts | pass | pass | fail |
| src/compiler/performance.ts | pass | pass | fail |
| src/compiler/performanceCore.ts | pass | pass | fail |
| src/compiler/program.ts | pass | pass | fail |
| src/compiler/programDiagnostics.ts | pass | pass | fail |
| src/compiler/resolutionCache.ts | pass | pass | fail |
| src/compiler/scanner.ts | pass | pass | fail |
| src/compiler/semver.ts | pass | pass | fail |
| src/compiler/sourcemap.ts | pass | pass | fail |
| src/compiler/symbolWalker.ts | pass | pass | fail |
| src/compiler/sys.ts | pass | pass | fail |
| src/compiler/tracing.ts | pass | pass | fail |
| src/compiler/transformer.ts | pass | pass | fail |
| src/compiler/transformers/classFields.ts | pass | pass | fail |
| src/compiler/transformers/classThis.ts | pass | pass | fail |
| src/compiler/transformers/declarations.ts | pass | pass | fail |
| src/compiler/transformers/declarations/diagnostics.ts | pass | pass | fail |
| src/compiler/transformers/destructuring.ts | pass | pass | fail |
| src/compiler/transformers/es2015.ts | pass | pass | fail |
| src/compiler/transformers/es2016.ts | pass | pass | fail |
| src/compiler/transformers/es2017.ts | pass | pass | fail |
| src/compiler/transformers/es2018.ts | pass | pass | fail |
| src/compiler/transformers/es2019.ts | pass | pass | fail |
| src/compiler/transformers/es2020.ts | pass | pass | fail |
| src/compiler/transformers/es2021.ts | pass | pass | fail |
| src/compiler/transformers/esDecorators.ts | pass | pass | fail |
| src/compiler/transformers/esnext.ts | pass | pass | fail |
| src/compiler/transformers/generators.ts | pass | pass | fail |
| src/compiler/transformers/jsx.ts | pass | pass | fail |
| src/compiler/transformers/legacyDecorators.ts | pass | pass | fail |
| src/compiler/transformers/module/esnextAnd2015.ts | pass | pass | fail |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | pass | pass | fail |
| src/compiler/transformers/module/module.ts | pass | pass | fail |
| src/compiler/transformers/module/system.ts | pass | pass | fail |
| src/compiler/transformers/namedEvaluation.ts | pass | pass | fail |
| src/compiler/transformers/taggedTemplate.ts | pass | pass | fail |
| src/compiler/transformers/ts.ts | pass | pass | fail |
| src/compiler/transformers/typeSerializer.ts | pass | pass | fail |
| src/compiler/transformers/utilities.ts | pass | pass | fail |
| src/compiler/tsbuild.ts | pass | pass | fail |
| src/compiler/tsbuildPublic.ts | pass | pass | fail |
| src/compiler/tsconfig.json | fail | fail | blocked |
| src/compiler/types.ts | pass | pass | fail |
| src/compiler/utilities.ts | pass | pass | fail |
| src/compiler/utilitiesPublic.ts | pass | pass | fail |
| src/compiler/visitorPublic.ts | pass | pass | fail |
| src/compiler/watch.ts | pass | pass | fail |
| src/compiler/watchPublic.ts | pass | pass | fail |
| src/compiler/watchUtilities.ts | pass | pass | fail |

Checker: 79/79 source files. Lowering: 2/79 source files.
All files: 82; JSON inputs fail the source-extension gate.
Blocked means lowering was not attempted because the checker/input gate failed.

tsc entry lowering, main: measured on a checker-clean entry-root program.
Resolved source files: 81; NotYet: 594; Refused: 4267.
Errors: 0; panics: 0; skipped dependencies: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 1251 | 1251 |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 984 | 984 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 86 | 86 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 69 | 69 |
| optional property id in Node absent from structural source never, which can hide fields | OWNER BLANK | 0 | 56 | 56 |
| a Map or Set interface implemented by a value that isn&#x27;t a Map or Set | OWNER BLANK | 50 | 0 | 50 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 36 | 36 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 34 | 0 | 34 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 32 | 0 | 32 |
| a call through ?. (an optional call) | OWNER BLANK | 28 | 0 | 28 |

Counts cover the entry root and its resolved implementation dependencies, excluding declarations.
Errors and panics are retained separately in JSON. This census does not establish native output.

Latent lowering, main: measured on a checker-rejected program.
NotYet: 593; Refused: 4265.
Recovery boundaries: 0; named nested checker exclusions: 0.
Errors: 0; panics: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 1251 | 1251 |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 984 | 984 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 86 | 86 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 69 | 69 |
| optional property id in Node absent from structural source never, which can hide fields | OWNER BLANK | 0 | 56 | 56 |
| a Map or Set interface implemented by a value that isn&#x27;t a Map or Set | OWNER BLANK | 50 | 0 | 50 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 36 | 36 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 34 | 0 | 34 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 32 | 0 | 32 |
| a call through ?. (an optional call) | OWNER BLANK | 28 | 0 | 28 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

This box's times, not comparable across boxes: 498.760 wall seconds; 590.323 CPU seconds.

| Slowest files | Wall seconds | CPU seconds |
| --- | ---: | ---: |
| src/compiler/utilities.ts | 146.073 | 167.928 |
| src/compiler/types.ts | 63.869 | 73.647 |
| src/compiler/utilitiesPublic.ts | 58.199 | 71.142 |
| src/compiler/factory/utilities.ts | 19.028 | 21.457 |
| src/compiler/commandLineParser.ts | 15.436 | 17.903 |
