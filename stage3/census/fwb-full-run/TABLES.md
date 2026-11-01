# Full latent entry census tables

| Metric | Main | Fix |
| --- | ---: | ---: |
| files | 81 | 81 |
| units | 10705 | 10705 |
| NotYet findings | 224 | 317 |
| Refused findings | 4049 | 4178 |
| panic findings | 9818 | 7929 |
| error findings | 0 | 109 |
| Boundary findings | 9818 | 8151 |
| SkippedDependency findings | 0 | 0 |

## Main panic reasons

| Reason | Findings | Example unit |
| --- | ---: | --- |
| latent state copy: unexported IR field methodTables | 9818 | /workspace/scratch/tree/src/compiler/corePublic.ts:3:1 |

## Fix panic reasons

| Reason | Findings | Example unit |
| --- | ---: | --- |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/checker.ts | 2390 | /workspace/scratch/tree/src/compiler/checker.ts:1455:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/utilities.ts | 710 | /workspace/scratch/tree/src/compiler/utilities.ts:644:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/parser.ts | 509 | /workspace/scratch/tree/src/compiler/parser.ts:462:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts | 473 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:481:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/emitter.ts | 403 | /workspace/scratch/tree/src/compiler/emitter.ts:431:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/utilitiesPublic.ts | 228 | /workspace/scratch/tree/src/compiler/utilitiesPublic.ts:296:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/nodeTests.ts | 227 | /workspace/scratch/tree/src/compiler/factory/nodeTests.ts:235:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/program.ts | 199 | /workspace/scratch/tree/src/compiler/program.ts:328:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2015.ts | 164 | /workspace/scratch/tree/src/compiler/transformers/es2015.ts:483:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/binder.ts | 162 | /workspace/scratch/tree/src/compiler/binder.ts:330:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/moduleNameResolver.ts | 131 | /workspace/scratch/tree/src/compiler/moduleNameResolver.ts:118:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/scanner.ts | 123 | /workspace/scratch/tree/src/compiler/scanner.ts:42:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/generators.ts | 119 | /workspace/scratch/tree/src/compiler/transformers/generators.ts:318:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/ts.ts | 116 | /workspace/scratch/tree/src/compiler/transformers/ts.ts:235:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/classFields.ts | 108 | /workspace/scratch/tree/src/compiler/transformers/classFields.ts:354:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/sys.ts | 94 | /workspace/scratch/tree/src/compiler/sys.ts:51:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/utilities.ts | 93 | /workspace/scratch/tree/src/compiler/factory/utilities.ts:183:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/commandLineParser.ts | 90 | /workspace/scratch/tree/src/compiler/commandLineParser.ts:1728:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/builder.ts | 85 | /workspace/scratch/tree/src/compiler/builder.ts:268:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/esDecorators.ts | 76 | /workspace/scratch/tree/src/compiler/transformers/esDecorators.ts:295:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/module/module.ts | 74 | /workspace/scratch/tree/src/compiler/transformers/module/module.ts:174:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/debug.ts | 72 | /workspace/scratch/tree/src/compiler/debug.ts:123:5 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/module/system.ts | 70 | /workspace/scratch/tree/src/compiler/transformers/module/system.ts:137:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/path.ts | 67 | /workspace/scratch/tree/src/compiler/path.ts:41:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/declarations.ts | 61 | /workspace/scratch/tree/src/compiler/transformers/declarations.ts:221:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/executeCommandLine.ts | 60 | /workspace/scratch/tree/src/compiler/executeCommandLine.ts:105:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/resolutionCache.ts | 59 | /workspace/scratch/tree/src/compiler/resolutionCache.ts:258:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2018.ts | 56 | /workspace/scratch/tree/src/compiler/transformers/es2018.ts:154:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/expressionToTypeNode.ts | 55 | /workspace/scratch/tree/src/compiler/expressionToTypeNode.ts:164:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/core.ts | 46 | /workspace/scratch/tree/src/compiler/core.ts:232:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/moduleSpecifiers.ts | 44 | /workspace/scratch/tree/src/compiler/moduleSpecifiers.ts:182:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2017.ts | 44 | /workspace/scratch/tree/src/compiler/transformers/es2017.ts:119:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/emitHelpers.ts | 41 | /workspace/scratch/tree/src/compiler/factory/emitHelpers.ts:148:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/utilities.ts | 39 | /workspace/scratch/tree/src/compiler/transformers/utilities.ts:107:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/legacyDecorators.ts | 39 | /workspace/scratch/tree/src/compiler/transformers/legacyDecorators.ts:88:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/parenthesizerRules.ts | 37 | /workspace/scratch/tree/src/compiler/factory/parenthesizerRules.ts:54:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/sourcemap.ts | 37 | /workspace/scratch/tree/src/compiler/sourcemap.ts:36:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/jsx.ts | 36 | /workspace/scratch/tree/src/compiler/transformers/jsx.ts:92:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/declarations/diagnostics.ts | 31 | /workspace/scratch/tree/src/compiler/transformers/declarations/diagnostics.ts:138:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/builderState.ts | 29 | /workspace/scratch/tree/src/compiler/builderState.ts:39:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/watchUtilities.ts | 28 | /workspace/scratch/tree/src/compiler/watchUtilities.ts:118:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/esnext.ts | 24 | /workspace/scratch/tree/src/compiler/transformers/esnext.ts:71:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/destructuring.ts | 23 | /workspace/scratch/tree/src/compiler/transformers/destructuring.ts:99:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/watch.ts | 23 | /workspace/scratch/tree/src/compiler/watch.ts:125:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/commandLineParser.ts | 20 | /workspace/scratch/tree/src/compiler/commandLineParser.ts:274:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/emitNode.ts | 17 | /workspace/scratch/tree/src/compiler/factory/emitNode.ts:36:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/typeSerializer.ts | 17 | /workspace/scratch/tree/src/compiler/transformers/typeSerializer.ts:132:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/core.ts | 16 | /workspace/scratch/tree/src/compiler/core.ts:2086:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/performance.ts | 15 | /workspace/scratch/tree/src/compiler/performance.ts:29:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/semver.ts | 14 | /workspace/scratch/tree/src/compiler/semver.ts:141:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/namedEvaluation.ts | 14 | /workspace/scratch/tree/src/compiler/transformers/namedEvaluation.ts:56:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/tracing.ts | 13 | /workspace/scratch/tree/src/compiler/tracing.ts:59:5 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/symbolWalker.ts | 13 | /workspace/scratch/tree/src/compiler/symbolWalker.ts:30:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2020.ts | 12 | /workspace/scratch/tree/src/compiler/transformers/es2020.ts:42:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/types.ts | 10 | /workspace/scratch/tree/src/compiler/types.ts:4599:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/nodeConverters.ts | 10 | /workspace/scratch/tree/src/compiler/factory/nodeConverters.ts:41:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/visitorPublic.ts | 10 | /workspace/scratch/tree/src/compiler/visitorPublic.ts:382:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/module/esnextAnd2015.ts | 10 | /workspace/scratch/tree/src/compiler/transformers/module/esnextAnd2015.ts:67:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformer.ts | 9 | /workspace/scratch/tree/src/compiler/transformer.ts:78:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/builderPublic.ts | 9 | /workspace/scratch/tree/src/compiler/builderPublic.ts:162:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/programDiagnostics.ts | 8 | /workspace/scratch/tree/src/compiler/programDiagnostics.ts:89:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/utilities.ts | 7 | /workspace/scratch/tree/src/compiler/utilities.ts:668:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/module/impliedNodeFormatDependent.ts | 7 | /workspace/scratch/tree/src/compiler/transformers/module/impliedNodeFormatDependent.ts:18:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/tsbuildPublic.ts | 7 | /workspace/scratch/tree/src/compiler/tsbuildPublic.ts:205:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/sys.ts | 6 | /workspace/scratch/tree/src/compiler/sys.ts:81:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/baseNodeFactory.ts | 6 | /workspace/scratch/tree/src/compiler/factory/baseNodeFactory.ts:26:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2016.ts | 6 | /workspace/scratch/tree/src/compiler/transformers/es2016.ts:21:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/debug.ts | 4 | /workspace/scratch/tree/src/compiler/debug.ts:99:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts | 4 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:7399:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/nodeChildren.ts | 4 | /workspace/scratch/tree/src/compiler/factory/nodeChildren.ts:14:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2019.ts | 4 | /workspace/scratch/tree/src/compiler/transformers/es2019.ts:17:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/es2021.ts | 4 | /workspace/scratch/tree/src/compiler/transformers/es2021.ts:25:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/performanceCore.ts | 3 | /workspace/scratch/tree/src/compiler/performanceCore.ts:30:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/scanner.ts | 3 | /workspace/scratch/tree/src/compiler/scanner.ts:409:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/parser.ts | 3 | /workspace/scratch/tree/src/compiler/parser.ts:432:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/taggedTemplate.ts | 3 | /workspace/scratch/tree/src/compiler/transformers/taggedTemplate.ts:35:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/program.ts | 3 | /workspace/scratch/tree/src/compiler/program.ts:680:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/tsbuildPublic.ts | 3 | /workspace/scratch/tree/src/compiler/tsbuildPublic.ts:169:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/expressionToTypeNode.ts | 3 | /workspace/scratch/tree/src/compiler/expressionToTypeNode.ts:169:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/semver.ts | 2 | /workspace/scratch/tree/src/compiler/semver.ts:46:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/performanceCore.ts | 2 | /workspace/scratch/tree/src/compiler/performanceCore.ts:89:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/hostErrors.ts | 2 | /workspace/scratch/tree/src/compiler/hostErrors.ts:2:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/tracing.ts | 2 | /workspace/scratch/tree/src/compiler/tracing.ts:369:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/utilitiesPublic.ts | 2 | /workspace/scratch/tree/src/compiler/factory/utilitiesPublic.ts:14:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/utilities.ts | 2 | /workspace/scratch/tree/src/compiler/transformers/utilities.ts:389:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/classThis.ts | 2 | /workspace/scratch/tree/src/compiler/transformers/classThis.ts:72:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/watchUtilities.ts | 2 | /workspace/scratch/tree/src/compiler/watchUtilities.ts:381:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/watch.ts | 2 | /workspace/scratch/tree/src/compiler/watch.ts:114:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/executeCommandLine.ts | 2 | /workspace/scratch/tree/src/compiler/executeCommandLine.ts:99:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/types.ts | 1 | /workspace/scratch/tree/src/compiler/types.ts:7331:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/diagnosticInformationMap.generated.ts | 1 | /workspace/scratch/tree/src/compiler/diagnosticInformationMap.generated.ts:6:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/diagnosticInformationMap.generated.ts | 1 | /workspace/scratch/tree/src/compiler/diagnosticInformationMap.generated.ts:12:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/utilitiesPublic.ts | 1 | /workspace/scratch/tree/src/compiler/utilitiesPublic.ts:477:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/parenthesizerRules.ts | 1 | /workspace/scratch/tree/src/compiler/factory/parenthesizerRules.ts:675:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/factory/baseNodeFactory.ts | 1 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:7392:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/path.ts | 1 | /workspace/scratch/tree/src/compiler/moduleNameResolver.ts:509:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/moduleNameResolver.ts | 1 | /workspace/scratch/tree/src/compiler/moduleNameResolver.ts:1687:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/binder.ts | 1 | /workspace/scratch/tree/src/compiler/binder.ts:499:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformers/taggedTemplate.ts | 1 | /workspace/scratch/tree/src/compiler/transformers/taggedTemplate.ts:29:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/transformer.ts | 1 | /workspace/scratch/tree/src/compiler/transformer.ts:669:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/emitter.ts | 1 | /workspace/scratch/tree/src/compiler/emitter.ts:428:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/builderState.ts | 1 | /workspace/scratch/tree/src/compiler/builderState.ts:56:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/builder.ts | 1 | /workspace/scratch/tree/src/compiler/builder.ts:1570:1 |
| lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/watchPublic.ts | 1 | /workspace/scratch/tree/src/compiler/watchPublic.ts:106:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/compiler/tsbuild.ts | 1 | /workspace/scratch/tree/src/compiler/tsbuild.ts:10:1 |
| statement panic: lower: source module has no import-declaration route: /workspace/scratch/tree/src/tsc/tsc.ts | 1 | /workspace/scratch/tree/src/tsc/tsc.ts:6:1 |

## Fix unit outcomes

| Outcome | Units |
| --- | ---: |
| attempted with no finding | 1639 |
| deferred_generic_instantiation | 823 |
| panic | 7925 |
| NotYet or Refused | 254 |
| split_checker_body | 64 |

Units are keyed by (file, where). Only attempted units with no findings owned by that key count as lowering with no finding. Deferred generics and checker splits remain separate. NotYet or Refused outcomes exclude panics. File refusal scan findings without a matching unit are counted in finding totals, not assigned to a unit.

## Fix top 20 NotYet or Refused reasons among units without a panic

| Kind | Reason | Distinct units | Example unit | Example site |
| --- | --- | ---: | --- | --- |
| Refused | a cast the runtime can't check | 24 | /workspace/scratch/tree/src/compiler/core.ts:1066:1 | /workspace/scratch/tree/src/compiler/core.ts:1066:100 |
| Refused | an object refinement using an open numeric enum as a literal tag | 13 | /workspace/scratch/tree/src/compiler/types.ts:8742:1 | /workspace/scratch/tree/src/compiler/types.ts:8932:52 |
| NotYet | a value of type any | 11 | /workspace/scratch/tree/src/compiler/utilities.ts:8147:1 | /workspace/scratch/tree/src/compiler/utilities.ts:8147:36 |
| Refused | an ambient value declaration | 7 | /workspace/scratch/tree/src/compiler/core.ts:2589:1 | /workspace/scratch/tree/src/compiler/core.ts:2589:1 |
| Refused | overload 1 of createToken result SuperExpression cannot be served by implementation result Mutable<Token<TKind>> | 7 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:1441:5 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:1441:5 |
| Refused | overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined | 7 | /workspace/scratch/tree/src/compiler/visitorPublic.ts:508:1 | /workspace/scratch/tree/src/compiler/visitorPublic.ts:508:1 |
| Refused | optional property id in Node absent from structural source never, which can hide fields | 5 | /workspace/scratch/tree/src/compiler/utilities.ts:3479:1 | /workspace/scratch/tree/src/compiler/utilities.ts:3481:80 |
| Refused | overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer | 5 | /workspace/scratch/tree/src/compiler/utilities.ts:3182:1 | /workspace/scratch/tree/src/compiler/utilities.ts:3182:1 |
| Refused | overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean | 5 | /workspace/scratch/tree/src/compiler/utilities.ts:3471:1 | /workspace/scratch/tree/src/compiler/utilities.ts:3471:1 |
| Refused | overload 3 of getGlobalType result GenericType cannot be served by implementation result ObjectType \| undefined | 5 | /workspace/scratch/tree/src/compiler/checker.ts:17481:5 | /workspace/scratch/tree/src/compiler/checker.ts:17483:5 |
| NotYet | a value of type BindableStaticNameExpression | 4 | /workspace/scratch/tree/src/compiler/binder.ts:3438:5 | /workspace/scratch/tree/src/compiler/binder.ts:3438:43 |
| NotYet | new an Identifier | 4 | /workspace/scratch/tree/src/compiler/utilities.ts:10225:1 | /workspace/scratch/tree/src/compiler/utilities.ts:10225:29 |
| NotYet | a non-object representation passed through an opaque object parameter | 3 | /workspace/scratch/tree/src/compiler/debug.ts:273:5 | /workspace/scratch/tree/src/compiler/debug.ts:274:66 |
| NotYet | a value of type BindableObjectDefinePropertyCall | 3 | /workspace/scratch/tree/src/compiler/binder.ts:3201:5 | /workspace/scratch/tree/src/compiler/binder.ts:3201:45 |
| NotYet | a value of type InitializedVariableDeclaration | 3 | /workspace/scratch/tree/src/compiler/transformers/ts.ts:1654:5 | /workspace/scratch/tree/src/compiler/transformers/ts.ts:1654:43 |
| NotYet | a value of type Readonly<PathPathComponents> | 3 | /workspace/scratch/tree/src/compiler/resolutionCache.ts:269:1 | /workspace/scratch/tree/src/compiler/resolutionCache.ts:269:43 |
| Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | 3 | /workspace/scratch/tree/src/compiler/debug.ts:283:5 | /workspace/scratch/tree/src/compiler/debug.ts:288:72 |
| Refused | overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping | 3 | /workspace/scratch/tree/src/compiler/sourcemap.ts:544:5 | /workspace/scratch/tree/src/compiler/sourcemap.ts:544:5 |
| Refused | overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 3 | /workspace/scratch/tree/src/compiler/builder.ts:1663:1 | /workspace/scratch/tree/src/compiler/builder.ts:1663:1 |
| Refused | overload 1 of createImmediatelyInvokedArrowFunction result ImmediatelyInvokedArrowFunction cannot be served by implementation result Mutable<CallExpression> | 3 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:6429:5 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:6429:5 |

These reasons are counted among units in the NotYet or Refused outcome, excluding every unit with a panic finding. Each reason counts distinct units; a unit can have several reasons. Full reason lists and examples are in summary.json.

## Fix top 20 NotYet or Refused reasons across all units

This additional table includes units that also panic. It does not describe the 254 units that stop without a panic.

| Kind | Reason | Distinct units | Example unit |
| --- | --- | ---: | --- |
| Refused | a cast the runtime can't check | 541 | /workspace/scratch/tree/src/compiler/core.ts:350:1 |
| Refused | an object refinement using an open numeric enum as a literal tag | 288 | /workspace/scratch/tree/src/compiler/types.ts:8742:1 |
| Refused | a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 58 | /workspace/scratch/tree/src/compiler/utilities.ts:8644:1 |
| Refused | a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 54 | /workspace/scratch/tree/src/compiler/commandLineParser.ts:3572:1 |
| Refused | optional property id in Node absent from structural source never, which can hide fields | 28 | /workspace/scratch/tree/src/compiler/utilitiesPublic.ts:931:1 |
| Refused | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 27 | /workspace/scratch/tree/src/compiler/tracing.ts:200:5 |
| NotYet | a non-object representation passed through an opaque object parameter | 24 | /workspace/scratch/tree/src/compiler/core.ts:1350:1 |
| NotYet | an ESM namespace object used as a value; use qualified live bindings | 19 | /workspace/scratch/tree/src/compiler/debug.ts:444:5 |
| Refused | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 18 | /workspace/scratch/tree/src/compiler/utilities.ts:10957:1 |
| Refused | optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 17 | /workspace/scratch/tree/src/compiler/factory/nodeFactory.ts:7431:1 |
| NotYet | a structural relation deeper than 64 type pairs while checking void function results | 16 | /workspace/scratch/tree/src/compiler/binder.ts:571:5 |
| NotYet | a CompilerOptions seen as Pick<CompilerOptions, "module" \| "target"> (fixed objects and records have different storage; copy explicitly) | 13 | /workspace/scratch/tree/src/compiler/utilities.ts:2157:1 |
| Refused | a method read as a value (liftToBlock would lose its object, and this with it) | 13 | /workspace/scratch/tree/src/compiler/visitorPublic.ts:543:1 |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 13 | /workspace/scratch/tree/src/compiler/utilitiesPublic.ts:1204:1 |
| Refused | optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 13 | /workspace/scratch/tree/src/compiler/binder.ts:737:5 |
| Refused | a generator function | 12 | /workspace/scratch/tree/src/compiler/core.ts:332:1 |
| Refused | a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 | /workspace/scratch/tree/src/compiler/emitter.ts:2221:5 |
| Refused | a value of type ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 12 | /workspace/scratch/tree/src/compiler/checker.ts:5664:5 |
| Refused | yield (generators) | 12 | /workspace/scratch/tree/src/compiler/core.ts:332:1 |
| NotYet | a value of type any | 11 | /workspace/scratch/tree/src/compiler/utilities.ts:8147:1 |
