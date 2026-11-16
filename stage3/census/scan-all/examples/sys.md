# scan-all example

One root import-closure run. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. A zero for a body or tail rule is never a pass. Preparation failures can prevent scanning entirely.

Root: `/workspace/scratch/et209ny-state/adapted-example/src/compiler/sys.ts`

Wall: 366.8239 s; peak RSS: 1900544 KiB (Linux child maximum).

Completed: False. Boundary: production abort at lowering.declareModule

First declaration/body/tail stop: `lowering.declareModule` at `/workspace/scratch/et209ny-state/adapted-example/src/compiler/core.ts:261:1`: overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined

One root import-closure run. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. A zero for a body or tail rule is never a pass. Preparation failures can prevent scanning entirely.

| Observation kind | Rule | Events | Distinct (rule, site, words) triples |
| --- | --- | ---: | ---: |
| first declaration/body/tail stop | `lowering.declareModule` | 1 | 1 |
| scan stop | `refusal_scan/boundObservation` | 3 | 3 |
| scan stop | `refusal_scan/castRefusal` | 2532 | 1021 |
| scan stop | `refusal_scan/classDeadZoneGap` | 4 | 4 |
| scan stop | `refusal_scan/classViewRefusal` | 18 | 18 |
| scan stop | `refusal_scan/constraintRelationRefusal` | 452 | 452 |
| scan stop | `refusal_scan/deferred` | 7 | 7 |
| scan stop | `refusal_scan/deleteRefusal` | 1 | 1 |
| scan stop | `refusal_scan/enumRefusal` | 962 | 961 |
| scan stop | `refusal_scan/enumerationDescriptorRefusal` | 3 | 3 |
| scan stop | `refusal_scan/enumerationIndexedWriteRefusal` | 2 | 2 |
| scan stop | `refusal_scan/methodValueRefusal` | 151 | 151 |
| scan stop | `refusal_scan/namespaceRefusal` | 1 | 1 |
| scan stop | `refusal_scan/narrowedNeverReadGap` | 2 | 2 |
| scan stop | `refusal_scan/narrowedNullishReadGap` | 11 | 11 |
| scan stop | `refusal_scan/nodeLibraryRefusal` | 2 | 2 |
| scan stop | `refusal_scan/phantomRefusal` | 1 | 1 |
| scan stop | `refusal_scan/predicateArguments` | 23 | 23 |
| scan stop | `refusal_scan/predicateArgumentsRefusal` | 23 | 23 |
| scan stop | `refusal_scan/recordStorageView` | 24 | 24 |
| scan stop | `refusal_scan/refuseAnyMember` | 1 | 1 |
| scan stop | `refusal_scan/refuseOptionalWidening` | 233 | 229 |
| scan stop | `refusal_scan/refusePromiseIntoVoid` | 24 | 24 |
| scan stop | `refusal_scan/refuseVoidRead` | 3 | 3 |
| scan stop | `refusal_scan/refuseVoidWidening` | 3 | 3 |
| scan stop | `refusal_scan/stringWideningRefusal` | 1053 | 1053 |
| scan stop | `refusal_scan/unionSlotWriteRefusal` | 5 | 5 |

One root import-closure run. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. A zero for a body or tail rule is never a pass. Preparation failures can prevent scanning entirely.

| Observation kind | File (from site) | Events | Distinct (rule, site, words) triples |
| --- | --- | ---: | ---: |
| first declaration/body/tail stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/core.ts` | 1 | 1 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/binder.ts` | 189 | 166 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/builder.ts` | 116 | 114 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/builderState.ts` | 2 | 2 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/checker.ts` | 1415 | 1414 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/commandLineParser.ts` | 90 | 90 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/emitter.ts` | 118 | 118 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/executeCommandLine.ts` | 13 | 13 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/expressionToTypeNode.ts` | 38 | 38 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/factory/emitHelpers.ts` | 7 | 7 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/factory/emitNode.ts` | 8 | 8 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/factory/nodeConverters.ts` | 21 | 21 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/factory/nodeFactory.ts` | 1723 | 244 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/factory/parenthesizerRules.ts` | 17 | 17 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/factory/utilities.ts` | 83 | 83 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/moduleNameResolver.ts` | 41 | 41 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/moduleSpecifiers.ts` | 40 | 40 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/parser.ts` | 145 | 145 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/program.ts` | 79 | 77 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/programDiagnostics.ts` | 10 | 10 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/resolutionCache.ts` | 22 | 22 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/sourcemap.ts` | 18 | 18 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/symbolWalker.ts` | 7 | 7 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/sys.ts` | 6 | 6 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformer.ts` | 7 | 6 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/classFields.ts` | 69 | 69 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/classThis.ts` | 3 | 3 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/declarations.ts` | 55 | 55 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/declarations/diagnostics.ts` | 54 | 54 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/destructuring.ts` | 26 | 26 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/es2015.ts` | 149 | 149 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/es2016.ts` | 10 | 10 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/es2017.ts` | 23 | 23 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/es2018.ts` | 42 | 42 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/es2020.ts` | 5 | 5 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/esDecorators.ts` | 94 | 94 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/esnext.ts` | 24 | 24 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/generators.ts` | 80 | 80 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/jsx.ts` | 14 | 14 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/legacyDecorators.ts` | 15 | 15 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/module/esnextAnd2015.ts` | 9 | 9 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/module/impliedNodeFormatDependent.ts` | 1 | 1 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/module/module.ts` | 69 | 69 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/module/system.ts` | 38 | 38 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/namedEvaluation.ts` | 28 | 28 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/taggedTemplate.ts` | 1 | 1 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/ts.ts` | 68 | 68 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/typeSerializer.ts` | 8 | 8 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/transformers/utilities.ts` | 17 | 17 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/tsbuildPublic.ts` | 56 | 52 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/utilities.ts` | 234 | 234 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/utilitiesPublic.ts` | 33 | 33 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/visitorPublic.ts` | 43 | 43 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/watch.ts` | 13 | 12 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/watchPublic.ts` | 34 | 31 |
| scan stop | `/workspace/scratch/et209ny-state/adapted-example/src/compiler/watchUtilities.ts` | 14 | 14 |

Total recorded events: 5545; distinct (rule, site, words) triples: 4029.

Only the underlying first pipeline stop is counted; a substituted returned diagnosis is not another observation. Later pass results remain unknown.
