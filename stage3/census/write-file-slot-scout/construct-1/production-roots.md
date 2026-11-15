# All 79 production roots

Control: a77ca4a82271f18c142d8843ea76604b30d3833d merged with main 89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825. Candidate: frozen implementation merged with the same main.

Real production Load/Lower; CommonJSBundle. Driver selects entries only; inspect implementation is byte-identical to production.

| Root under src/compiler | Control | Candidate |
|---|---|---|
| `_namespaces/ts.moduleSpecifiers.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `_namespaces/ts.performance.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `_namespaces/ts.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `binder.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `builder.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `builderPublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `builderState.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `builderStatePublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `checker.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `commandLineParser.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `core.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `corePublic.ts` | accepted | accepted |
| `debug.ts` | NotYet src/compiler/semver.ts:71:9 | NotYet src/compiler/semver.ts:71:9 |
| `diagnosticInformationMap.generated.ts` | NotYet src/compiler/commandLineParser.ts:293:35 | NotYet src/compiler/commandLineParser.ts:293:35 |
| `emitter.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `executeCommandLine.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `expressionToTypeNode.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/baseNodeFactory.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/emitHelpers.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/emitNode.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/nodeChildren.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/nodeConverters.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/nodeFactory.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/nodeTests.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/parenthesizerRules.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/utilities.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `factory/utilitiesPublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `hostErrors.ts` | accepted | accepted |
| `moduleNameResolver.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `moduleSpecifiers.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `parser.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `path.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `performance.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `performanceCore.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `program.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `programDiagnostics.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `resolutionCache.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `scanner.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `semver.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `sourcemap.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `symbolWalker.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `sys.ts` | Refused src/compiler/utilitiesPublic.ts:1284:52 | Refused src/compiler/utilitiesPublic.ts:1284:52 |
| `tracing.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformer.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/classFields.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/classThis.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/declarations.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/declarations/diagnostics.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/destructuring.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2015.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2016.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2017.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2018.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2019.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2020.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/es2021.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/esDecorators.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/esnext.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/generators.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/jsx.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/legacyDecorators.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/module/esnextAnd2015.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/module/impliedNodeFormatDependent.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/module/module.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/module/system.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/namedEvaluation.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/taggedTemplate.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/ts.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/typeSerializer.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `transformers/utilities.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `tsbuild.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `tsbuildPublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `types.ts` | NotYet src/compiler/diagnosticInformationMap.generated.ts:13:45 | NotYet src/compiler/diagnosticInformationMap.generated.ts:13:45 |
| `utilities.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `utilitiesPublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `visitorPublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `watch.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `watchPublic.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |
| `watchUtilities.ts` | Refused src/compiler/sys.ts:1387:5 | NotYet src/compiler/sys.ts:1481:57 |

All 73 former target roots move to the same clearTimeout boundary; all other six outcomes are unchanged.
