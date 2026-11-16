# scan-all example

One root import-closure run. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. A zero for a body or tail rule is never a pass. Preparation failures can prevent scanning entirely.

Root: `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/sys.ts`

Wall: 322.8682 s; peak RSS: 1847564 KiB (Linux child maximum).

Completed: False. Boundary: production abort at lowering.declareModule

First declaration/body/tail stop: `lowering.declareModule` at `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/core.ts:261:1`: overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined

One root import-closure run. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. A zero for a body or tail rule is never a pass. Preparation failures can prevent scanning entirely.

| Observation kind | Rule | Events | Distinct (rule, site, words) triples |
| --- | --- | ---: | ---: |
| first declaration/body/tail stop | `lowering.declareModule` | 1 | 1 |
| scan stop | `refusal_scan/boundObservation` | 3 | 3 |
| scan stop | `refusal_scan/castRefusal` | 2532 | 1021 |
| scan stop | `refusal_scan/classDeadZoneGap` | 4 | 4 |
| scan stop | `refusal_scan/classViewRefusal` | 18 | 18 |
| scan stop | `refusal_scan/constraintRelationRefusal` | 438 | 438 |
| scan stop | `refusal_scan/deferred` | 7 | 7 |
| scan stop | `refusal_scan/deleteRefusal` | 1 | 1 |
| scan stop | `refusal_scan/enumRefusal` | 962 | 961 |
| scan stop | `refusal_scan/enumerationDescriptorRefusal` | 3 | 3 |
| scan stop | `refusal_scan/enumerationIndexedWriteRefusal` | 2 | 2 |
| scan stop | `refusal_scan/methodValueRefusal` | 22 | 22 |
| scan stop | `refusal_scan/narrowedNeverReadGap` | 2 | 2 |
| scan stop | `refusal_scan/narrowedNullishReadGap` | 11 | 11 |
| scan stop | `refusal_scan/phantomRefusal` | 1 | 1 |
| scan stop | `refusal_scan/predicateArguments` | 22 | 22 |
| scan stop | `refusal_scan/predicateArgumentsRefusal` | 22 | 22 |
| scan stop | `refusal_scan/recordStorageView` | 24 | 24 |
| scan stop | `refusal_scan/refuseAnyMember` | 1 | 1 |
| scan stop | `refusal_scan/refuseOptionalWidening` | 105 | 101 |
| scan stop | `refusal_scan/refusePromiseIntoVoid` | 24 | 24 |
| scan stop | `refusal_scan/refuseVoidRead` | 3 | 3 |
| scan stop | `refusal_scan/refuseVoidWidening` | 3 | 3 |
| scan stop | `refusal_scan/stringWideningRefusal` | 1051 | 1051 |
| scan stop | `refusal_scan/unionSlotWriteRefusal` | 5 | 5 |

One root import-closure run. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. A zero for a body or tail rule is never a pass. Preparation failures can prevent scanning entirely.

| Observation kind | File (from site) | Events | Distinct (rule, site, words) triples |
| --- | --- | ---: | ---: |
| first declaration/body/tail stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/core.ts` | 1 | 1 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/binder.ts` | 181 | 158 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/builder.ts` | 111 | 109 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/builderState.ts` | 2 | 2 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/checker.ts` | 1349 | 1348 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/commandLineParser.ts` | 90 | 90 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/emitter.ts` | 66 | 66 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/executeCommandLine.ts` | 11 | 11 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/expressionToTypeNode.ts` | 35 | 35 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/factory/emitHelpers.ts` | 7 | 7 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/factory/emitNode.ts` | 8 | 8 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/factory/nodeConverters.ts` | 21 | 21 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/factory/nodeFactory.ts` | 1716 | 237 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/factory/parenthesizerRules.ts` | 17 | 17 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/factory/utilities.ts` | 81 | 81 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/moduleNameResolver.ts` | 41 | 41 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/moduleSpecifiers.ts` | 39 | 39 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/parser.ts` | 133 | 133 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/program.ts` | 66 | 64 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/programDiagnostics.ts` | 10 | 10 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/resolutionCache.ts` | 22 | 22 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/sourcemap.ts` | 17 | 17 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/symbolWalker.ts` | 7 | 7 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformer.ts` | 7 | 6 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/classFields.ts` | 63 | 63 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/classThis.ts` | 3 | 3 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/declarations.ts` | 52 | 52 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/declarations/diagnostics.ts` | 53 | 53 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/destructuring.ts` | 24 | 24 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/es2015.ts` | 144 | 144 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/es2016.ts` | 10 | 10 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/es2017.ts` | 23 | 23 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/es2018.ts` | 41 | 41 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/es2020.ts` | 5 | 5 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/esDecorators.ts` | 89 | 89 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/esnext.ts` | 23 | 23 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/generators.ts` | 79 | 79 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/jsx.ts` | 11 | 11 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/legacyDecorators.ts` | 15 | 15 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/module/esnextAnd2015.ts` | 7 | 7 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/module/impliedNodeFormatDependent.ts` | 1 | 1 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/module/module.ts` | 62 | 62 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/module/system.ts` | 32 | 32 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/namedEvaluation.ts` | 13 | 13 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/taggedTemplate.ts` | 1 | 1 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/ts.ts` | 54 | 54 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/typeSerializer.ts` | 8 | 8 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/transformers/utilities.ts` | 17 | 17 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/tsbuildPublic.ts` | 50 | 46 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/utilities.ts` | 223 | 223 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/utilitiesPublic.ts` | 30 | 30 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/visitorPublic.ts` | 38 | 38 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/watch.ts` | 12 | 11 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/watchPublic.ts` | 32 | 29 |
| scan stop | `/workspace/scratch/scan-all-sys-03c140ec/run/adapted/src/compiler/watchUtilities.ts` | 14 | 14 |

Total recorded events: 5267; distinct (rule, site, words) triples: 3751.

Only the underlying first pipeline stop is counted; a substituted returned diagnosis is not another observation. Later pass results remain unknown.
