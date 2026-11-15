# scan-all: one closure run

Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. Zero body-phase or tail-rule counts are never a pass.

Root: `/tmp/et209ny/adapted-final/src/compiler/sys.ts`

Rule overlays: scout 0d078a8d; bind ba619e28; records b529db90. These are independent rule heads, not a claim they are merged into main.

Pipeline completed: false (measurement only; never compiler acceptance).

Boundary: production abort at lowering.declareModule

Underlying first stop: lowering.declareModule at /tmp/et209ny/adapted-final/src/compiler/core.ts:261:1: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined

Recorded events: 5538; distinct (rule, site, words) triples: 4022. Retries may repeat sites.

Rule / boundary. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. Zero body-phase or tail-rule counts are never a pass.

| Rule / boundary | Events | Distinct (rule, site, words) triples |
| --- | ---: | ---: |
| lowering.declareModule | 1 | 1 |
| refusal_scan/boundObservation | 3 | 3 |
| refusal_scan/castRefusal | 2532 | 1021 |
| refusal_scan/classDeadZoneGap | 4 | 4 |
| refusal_scan/classViewRefusal | 18 | 18 |
| refusal_scan/constraintRelationRefusal | 452 | 452 |
| refusal_scan/deleteRefusal | 1 | 1 |
| refusal_scan/enumRefusal | 962 | 961 |
| refusal_scan/enumerationDescriptorRefusal | 3 | 3 |
| refusal_scan/enumerationIndexedWriteRefusal | 2 | 2 |
| refusal_scan/methodValueRefusal | 151 | 151 |
| refusal_scan/namespaceRefusal | 1 | 1 |
| refusal_scan/narrowedNeverReadGap | 2 | 2 |
| refusal_scan/narrowedNullishReadGap | 11 | 11 |
| refusal_scan/nodeLibraryRefusal | 2 | 2 |
| refusal_scan/phantomRefusal | 1 | 1 |
| refusal_scan/predicateArgumentsRefusal | 23 | 23 |
| refusal_scan/preflight-or-contract | 23 | 23 |
| refusal_scan/recordStorageView | 24 | 24 |
| refusal_scan/refuseAnyMember | 1 | 1 |
| refusal_scan/refuseOptionalWidening | 233 | 229 |
| refusal_scan/refusePromiseIntoVoid | 24 | 24 |
| refusal_scan/refuseVoidRead | 3 | 3 |
| refusal_scan/refuseVoidWidening | 3 | 3 |
| refusal_scan/stringWideningRefusal | 1053 | 1053 |
| refusal_scan/unionSlotWriteRefusal | 5 | 5 |

File derived from recorded site. Records refusal-scan stops only, plus the first declaration, body or tail stop reached; blind past that first stop. Zero body-phase or tail-rule counts are never a pass.

| File derived from recorded site | Events | Distinct (rule, site, words) triples |
| --- | ---: | ---: |
| /tmp/et209ny/adapted-final/src/compiler/binder.ts | 189 | 166 |
| /tmp/et209ny/adapted-final/src/compiler/builder.ts | 115 | 113 |
| /tmp/et209ny/adapted-final/src/compiler/builderState.ts | 2 | 2 |
| /tmp/et209ny/adapted-final/src/compiler/checker.ts | 1415 | 1414 |
| /tmp/et209ny/adapted-final/src/compiler/commandLineParser.ts | 89 | 89 |
| /tmp/et209ny/adapted-final/src/compiler/emitter.ts | 117 | 117 |
| /tmp/et209ny/adapted-final/src/compiler/executeCommandLine.ts | 13 | 13 |
| /tmp/et209ny/adapted-final/src/compiler/expressionToTypeNode.ts | 38 | 38 |
| /tmp/et209ny/adapted-final/src/compiler/factory/emitHelpers.ts | 7 | 7 |
| /tmp/et209ny/adapted-final/src/compiler/factory/emitNode.ts | 8 | 8 |
| /tmp/et209ny/adapted-final/src/compiler/factory/nodeConverters.ts | 21 | 21 |
| /tmp/et209ny/adapted-final/src/compiler/factory/nodeFactory.ts | 1722 | 243 |
| /tmp/et209ny/adapted-final/src/compiler/factory/parenthesizerRules.ts | 17 | 17 |
| /tmp/et209ny/adapted-final/src/compiler/factory/utilities.ts | 83 | 83 |
| /tmp/et209ny/adapted-final/src/compiler/moduleNameResolver.ts | 42 | 42 |
| /tmp/et209ny/adapted-final/src/compiler/moduleSpecifiers.ts | 40 | 40 |
| /tmp/et209ny/adapted-final/src/compiler/parser.ts | 144 | 144 |
| /tmp/et209ny/adapted-final/src/compiler/program.ts | 79 | 77 |
| /tmp/et209ny/adapted-final/src/compiler/programDiagnostics.ts | 10 | 10 |
| /tmp/et209ny/adapted-final/src/compiler/resolutionCache.ts | 22 | 22 |
| /tmp/et209ny/adapted-final/src/compiler/sourcemap.ts | 18 | 18 |
| /tmp/et209ny/adapted-final/src/compiler/symbolWalker.ts | 7 | 7 |
| /tmp/et209ny/adapted-final/src/compiler/sys.ts | 6 | 6 |
| /tmp/et209ny/adapted-final/src/compiler/transformer.ts | 7 | 6 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/classFields.ts | 69 | 69 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/classThis.ts | 3 | 3 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/declarations.ts | 55 | 55 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/declarations/diagnostics.ts | 54 | 54 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/destructuring.ts | 26 | 26 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/es2015.ts | 149 | 149 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/es2016.ts | 10 | 10 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/es2017.ts | 23 | 23 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/es2018.ts | 42 | 42 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/es2020.ts | 5 | 5 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/esDecorators.ts | 94 | 94 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/esnext.ts | 24 | 24 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/generators.ts | 80 | 80 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/jsx.ts | 14 | 14 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/legacyDecorators.ts | 15 | 15 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/module/esnextAnd2015.ts | 9 | 9 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/module/impliedNodeFormatDependent.ts | 1 | 1 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/module/module.ts | 69 | 69 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/module/system.ts | 38 | 38 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/namedEvaluation.ts | 28 | 28 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/taggedTemplate.ts | 1 | 1 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/ts.ts | 68 | 68 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/typeSerializer.ts | 8 | 8 |
| /tmp/et209ny/adapted-final/src/compiler/transformers/utilities.ts | 17 | 17 |
| /tmp/et209ny/adapted-final/src/compiler/tsbuildPublic.ts | 56 | 52 |
| /tmp/et209ny/adapted-final/src/compiler/utilities.ts | 233 | 233 |
| /tmp/et209ny/adapted-final/src/compiler/utilitiesPublic.ts | 33 | 33 |
| /tmp/et209ny/adapted-final/src/compiler/visitorPublic.ts | 42 | 42 |
| /tmp/et209ny/adapted-final/src/compiler/watch.ts | 13 | 12 |
| /tmp/et209ny/adapted-final/src/compiler/watchPublic.ts | 34 | 31 |
| /tmp/et209ny/adapted-final/src/compiler/watchUtilities.ts | 14 | 14 |

## Run provenance

Main base: `fe5a1a54f045e9bce4e5aa32ca74885a317fbebd`. One root import-closure run, using the pinned scout rule overlays named above.

TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`; adapted by this base's `stage3/apply.py`.

Adapted patch-set SHA-256: `60a33ba7c256ca4d15a5868d0a1e16677fd8dd90af9b4b318d1b564d0eef4dca`.

Wall time: **195.508535 s**; peak process RSS: **1,917,916 KiB** (1.829 GiB), Linux `wait4`. Timing excludes overlay build and source adaptation, includes load/check/measurement.
