# 99y per-file compiler census

Compiler and apply base: 322f80565edaa978d928ca559803c578ed67c105.
The ordinary census tool uses no overlay, refusal suppression, backend change,
or partial-root fallback. All 79 exact root paths use the completed full
/workspace/y-after tree. Groups contain 20, 20, 20, and 19 roots. Each was
committed and plainly pushed as soon as it finished. Each invocation retains
both raw ordinary rows and counts only the first; their non-timing fields match.

Compared against the grouped table in typescript/next-stops-7e3c0682,
retained as reference-report.txt. All 73 roots at core.ts:966:72 move past it.

| Roots | Location | Kind | Exact reason | Root names |
| ---: | --- | --- | --- | --- |
| 73 | `core.ts:1049:34` | Refused | a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `diagnosticInformationMap.generated.ts` |
| 1 | `debug.ts:243:59` | NotYet | a generic instantiation whose type argument for T, constrained to Node, can't be read | `core.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `types.ts` |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | `debug.ts` |

## Comparison with the reference

| Roots | Reference first stop | After first stop | Change |
| ---: | --- | --- | --- |
| 73 | `core.ts:966:72` | `core.ts:1049:34` | changed |
| 2 | `-` | `-` | unchanged |
| 1 | `debug.ts:243:59` | `debug.ts:243:59` | unchanged |
| 1 | `semver.ts:71:9` | `semver.ts:71:9` | unchanged |
| 1 | `commandLineParser.ts:293:35` | `commandLineParser.ts:293:35` | unchanged |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | `diagnosticInformationMap.generated.ts:13:45` | unchanged |

These are first lowering stops after checking. Accepted roots indicate ordinary
whole-root Lower acceptance only, with no native-emission or aggregate-program
claim. The reference's three readiness artifact classifications remain incomplete;
no new artifact or whole-root credit is inferred. census-results.json retains
every root's exact before and after row, including duplicate basenames.
