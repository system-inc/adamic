Whole program: 79/79
Own file: 79/79
Module format: CommonJSBundle

Stage 3 meter 20261118T131049Z

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


Lowering, file roots: 2/79; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
Relation budget: ok; assignable 314,384 of 1,000,000 pairs; assignable cache 314,384 pairs, within 1,000,000.
tsc entry body lowering, main: measured on a checker-clean entry-root program.
Module format: CommonJSBundle.
3227 units lower without a body stop; 0 have their whole-program passes measured
Scoped tail: 2 units stopped (checkRecordGuardedReads: 2).
Resolved source files: 81; NotYet: 436; Refused: 2513.
Errors: 0; panics: 0; skipped dependencies: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: | ---: |
| a cast the runtime can&#x27;t check | #b5w3ycg (unchecked) | 0 | 878 | 878 |
| an object refinement using an open numeric enum as a literal tag | #x2cjvy9 (unchecked) | 0 | 768 | 768 |
| a generic comparison deeper than 4 positions that can reach a Map | OWNER BLANK | 35 | 0 | 35 |
| writing a possibly absent optional own field | OWNER BLANK | 27 | 0 | 27 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | OWNER BLANK | 0 | 26 | 26 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | OWNER BLANK | 0 | 25 | 25 |
| a value of type any: any has no proven type | OWNER BLANK | 0 | 20 | 20 |
| a value of type FlowNode seen as FlowNode &#124; undefined, whose readonly field antecedent becomes writable: a readonly field may hold something narrower than undefined, which a write of FlowNode[] &#124; undefined would replace | OWNER BLANK | 0 | 19 | 19 |
| a value without nominal ancestry seen as BinaryExpressionStateMachine&lt;TOuterState, TState, TResult&gt; | OWNER BLANK | 0 | 18 | 18 |
| a generic instantiation whose type argument for TIn, constrained to Node &#124; undefined, can&#x27;t be read | OWNER BLANK | 17 | 0 | 17 |

Counts cover the entry root and its resolved implementation dependencies, excluding declarations.
Errors and panics are retained separately in JSON. This census does not establish native output.

Latent body lowering, main: measured on a checker-rejected program. Module format: CommonJSBundle.
3220 units lower without a body stop; 0 have their whole-program passes measured
Scoped tail: 2 units stopped (checkRecordGuardedReads: 2).
NotYet: 436; Refused: 2513.
Recovery boundaries: 0; named nested checker exclusions: 0.
Errors: 0; panics: 0.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a cast the runtime can&#x27;t check | #b5w3ycg (unchecked) | 0 | 878 | 878 |
| an object refinement using an open numeric enum as a literal tag | #x2cjvy9 (unchecked) | 0 | 768 | 768 |
| a generic comparison deeper than 4 positions that can reach a Map | OWNER BLANK | 35 | 0 | 35 |
| writing a possibly absent optional own field | OWNER BLANK | 27 | 0 | 27 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | OWNER BLANK | 0 | 26 | 26 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | OWNER BLANK | 0 | 25 | 25 |
| a value of type any: any has no proven type | OWNER BLANK | 0 | 20 | 20 |
| a value of type FlowNode seen as FlowNode &#124; undefined, whose readonly field antecedent becomes writable: a readonly field may hold something narrower than undefined, which a write of FlowNode[] &#124; undefined would replace | OWNER BLANK | 0 | 19 | 19 |
| a value without nominal ancestry seen as BinaryExpressionStateMachine&lt;TOuterState, TState, TResult&gt; | OWNER BLANK | 0 | 18 | 18 |
| a generic instantiation whose type argument for TIn, constrained to Node &#124; undefined, can&#x27;t be read | OWNER BLANK | 17 | 0 | 17 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

This box's times, not comparable across boxes: 942.892 wall seconds; 1122.707 CPU seconds.

| Slowest files | Wall seconds | CPU seconds |
| --- | ---: | ---: |
| src/compiler/utilities.ts | 262.617 | 307.278 |
| src/compiler/utilitiesPublic.ts | 115.406 | 134.409 |
| src/compiler/types.ts | 103.465 | 130.556 |
| src/compiler/factory/nodeTests.ts | 39.662 | 46.686 |
| src/compiler/factory/utilities.ts | 36.759 | 44.017 |
