# Load-time imported-value ledger

Method: stock TypeScript 6.0.3 compiler API, with imported value references instrumented before emit. Node evaluates the emitted ES modules from `src/tsc/tsc.ts`. A separate uninstrumented emit has the same dependency graph and output. Independent depth-first traversal matches all 80 observed module body starts. Cohere was not run.

Observed on Linux, Node v24.19.0, `--version`: 57 original top-level statements, 914 imported-value reads, 658 distinct `(originating statement, read location, binding)` sites. All providers finished before the reads. The generated diagnostic initializer is separate in `trace.json.gz`: 2,130 reads, all after `types.ts` finished. There are zero observed reads before provider begin or end.

The API also identifies one direct conditional read not executed here: `sys.ts:1979` reads `Debug` when `sys` is absent. `debug.ts` runs third and `sys.ts` tenth, so that direct read is safe by order.

**Bound:** this certifies the exercised module-load path and inventories direct imported reads in other top-level branches. It does not certify unexecuted branches inside functions invoked at load, other operating systems, development hooks or optional host packages. I did not produce an exhaustive all-platform ledger or propose a source adaptation on this evidence.

Each row is accepted for cycle scheduling, represented by fixture 05; fixtures 09 and 10 separately cover copying imported function values and reading hoisted functions before provider body execution. This is a scheduling witness, not a claim to reproduce the full initializer in every row. Other features in those initializers remain their own workers' scope.

`ledger.json` gives each binding, its declaration and every observed read location, its provider end event, and the first/last read event. Provider end is strictly earlier than first read. `trace.json.gz` preserves every occurrence and originating top-level statement, including reads performed inside called functions; `order.json` preserves the graph and body order.

| Top-level statement | Imported bindings (provider body order) | Evidence / decision |
| --- | --- | --- |
| `src/compiler/semver.ts:46:1` | `Debug` (3), `isArray` (2), `emptyArray` (2), `every` (2) | reader body 4; accept by module order |
| `src/compiler/performanceCore.ts:89:1` | `isNodeLikeSystem` (2) | reader body 5; accept by module order |
| `src/compiler/performance.ts:59:1` | `noop` (2) | reader body 6; accept by module order |
| `src/compiler/performance.ts:62:1` | `timestamp` (5) | reader body 6; accept by module order |
| `src/compiler/sys.ts:564:1` | `noop` (2) | reader body 10; accept by module order |
| `src/compiler/sys.ts:1462:1` | `isNodeLikeSystem` (2), `memoize` (2), `some` (2) | reader body 10; accept by module order |
| `src/compiler/sys.ts:1971:1` | `Debug` (3) | reader body 10; accept by module order |
| `src/compiler/scanner.ts:300:1` | `LanguageFeatureMinimumTarget` (9) | reader body 13; accept by module order |
| `src/compiler/utilities.ts:668:1` | `noop` (2) | reader body 15; accept by module order |
| `src/compiler/utilities.ts:1384:1` | `memoize` (2) | reader body 15; accept by module order |
| `src/compiler/utilities.ts:9972:1` | `flatten` (2) | reader body 15; accept by module order |
| `src/compiler/utilities.ts:9978:1` | `flatten` (2) | reader body 15; accept by module order |
| `src/compiler/utilities.ts:10312:1` | `emptyArray` (2) | reader body 15; accept by module order |
| `src/compiler/factory/parenthesizerRules.ts:675:1` | `identity` (2) | reader body 17; accept by module order |
| `src/compiler/factory/nodeConverters.ts:183:1` | `notImplemented` (2) | reader body 18; accept by module order |
| `src/compiler/factory/nodeFactory.ts:7392:1` | `createBaseNodeFactory` (16) | reader body 19; accept by module order |
| `src/compiler/factory/nodeFactory.ts:7407:1` | `memoize` (2), `memoizeOne` (2), `forEach` (2) | reader body 19; accept by module order |
| `src/compiler/parser.ts:441:1` | `createNodeFactory` (19), `memoize` (2), `memoizeOne` (2), `forEach` (2) | reader body 26; accept by module order |
| `src/compiler/parser.ts:1437:1` | `createScanner` (13), `Debug` (3), `createNodeFactory` (19), `identity` (2), `memoize` (2), `memoizeOne` (2), `forEach` (2), `textToKeywordObj` (13) | reader body 26; accept by module order |
| `src/compiler/commandLineParser.ts:141:1` | `mapIterator` (2) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:288:1` | `WatchFileKind` (9), `Diagnostics` (12), `WatchDirectoryKind` (9), `PollingWatchKind` (9) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:363:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:572:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:605:1` | `ModuleKind` (9), `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:637:1` | `Diagnostics` (12), `ModuleResolutionKind` (9), `ModuleDetectionKind` (9) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:1716:1` | `hasProperty` (2) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:1727:1` | `isString` (2) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:1733:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:1745:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:1851:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:1858:1` | `ModuleKind` (9) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2119:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2162:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2167:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2310:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2320:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2341:1` | `Diagnostics` (12) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2351:1` | `arrayToMap` (2) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2357:1` | `arrayToMap` (2) | reader body 27; accept by module order |
| `src/compiler/commandLineParser.ts:2363:1` | `arrayToMap` (2) | reader body 27; accept by module order |
| `src/compiler/moduleNameResolver.ts:509:1` | `combinePaths` (11) | reader body 28; accept by module order |
| `src/compiler/binder.ts:499:1` | `Debug` (3), `createBinaryExpressionTrampoline` (24) | reader body 29; accept by module order |
| `src/compiler/moduleSpecifiers.ts:133:1` | `memoizeOne` (2) | reader body 31; accept by module order |
| `src/compiler/checker.ts:1410:1` | `and` (2) | reader body 33; accept by module order |
| `src/compiler/sourcemap.ts:820:1` | `identity` (2) | reader body 35; accept by module order |
| `src/compiler/transformer.ts:117:1` | `emptyArray` (2) | reader body 62; accept by module order |
| `src/compiler/transformer.ts:669:1` | `factory` (19), `notImplemented` (2), `noop` (2), `returnUndefined` (2) | reader body 62; accept by module order |
| `src/compiler/emitter.ts:1147:1` | `notImplemented` (2) | reader body 63; accept by module order |
| `src/compiler/emitter.ts:1200:1` | `memoize` (2) | reader body 63; accept by module order |
| `src/compiler/emitter.ts:1203:1` | `memoize` (2) | reader body 63; accept by module order |
| `src/compiler/emitter.ts:1206:1` | `memoize` (2) | reader body 63; accept by module order |
| `src/compiler/emitter.ts:1209:1` | `memoize` (2) | reader body 63; accept by module order |
| `src/compiler/program.ts:1370:1` | `Diagnostics` (12) | reader body 65; accept by module order |
| `src/compiler/program.ts:5056:1` | `emptyArray` (2) | reader body 65; accept by module order |
| `src/compiler/watch.ts:114:1` | `sys` (10), `createGetCanonicalFileName` (2) | reader body 72; accept by module order |
| `src/compiler/watch.ts:161:1` | `Diagnostics` (12) | reader body 72; accept by module order |
| `src/compiler/watch.ts:668:1` | `noop` (2) | reader body 72; accept by module order |
| `src/compiler/sys.ts:1979:1` | `Debug` (3) | reader body 10; accept by module order; conditional read not observed |

## Evaluation order

| Body order | Module |
| ---: | --- |
| 1 | `src/compiler/corePublic.ts` |
| 2 | `src/compiler/core.ts` |
| 3 | `src/compiler/debug.ts` |
| 4 | `src/compiler/semver.ts` |
| 5 | `src/compiler/performanceCore.ts` |
| 6 | `src/compiler/performance.ts` |
| 7 | `src/compiler/_namespaces/ts.performance.ts` |
| 8 | `src/compiler/tracing.ts` |
| 9 | `src/compiler/types.ts` |
| 10 | `src/compiler/sys.ts` |
| 11 | `src/compiler/path.ts` |
| 12 | `src/compiler/diagnosticInformationMap.generated.ts` |
| 13 | `src/compiler/scanner.ts` |
| 14 | `src/compiler/utilitiesPublic.ts` |
| 15 | `src/compiler/utilities.ts` |
| 16 | `src/compiler/factory/baseNodeFactory.ts` |
| 17 | `src/compiler/factory/parenthesizerRules.ts` |
| 18 | `src/compiler/factory/nodeConverters.ts` |
| 19 | `src/compiler/factory/nodeFactory.ts` |
| 20 | `src/compiler/factory/emitNode.ts` |
| 21 | `src/compiler/factory/emitHelpers.ts` |
| 22 | `src/compiler/factory/nodeTests.ts` |
| 23 | `src/compiler/factory/nodeChildren.ts` |
| 24 | `src/compiler/factory/utilities.ts` |
| 25 | `src/compiler/factory/utilitiesPublic.ts` |
| 26 | `src/compiler/parser.ts` |
| 27 | `src/compiler/commandLineParser.ts` |
| 28 | `src/compiler/moduleNameResolver.ts` |
| 29 | `src/compiler/binder.ts` |
| 30 | `src/compiler/symbolWalker.ts` |
| 31 | `src/compiler/moduleSpecifiers.ts` |
| 32 | `src/compiler/_namespaces/ts.moduleSpecifiers.ts` |
| 33 | `src/compiler/checker.ts` |
| 34 | `src/compiler/visitorPublic.ts` |
| 35 | `src/compiler/sourcemap.ts` |
| 36 | `src/compiler/transformers/utilities.ts` |
| 37 | `src/compiler/transformers/destructuring.ts` |
| 38 | `src/compiler/transformers/classThis.ts` |
| 39 | `src/compiler/transformers/namedEvaluation.ts` |
| 40 | `src/compiler/transformers/taggedTemplate.ts` |
| 41 | `src/compiler/transformers/ts.ts` |
| 42 | `src/compiler/transformers/classFields.ts` |
| 43 | `src/compiler/transformers/typeSerializer.ts` |
| 44 | `src/compiler/transformers/legacyDecorators.ts` |
| 45 | `src/compiler/transformers/esDecorators.ts` |
| 46 | `src/compiler/transformers/es2017.ts` |
| 47 | `src/compiler/transformers/es2018.ts` |
| 48 | `src/compiler/transformers/es2019.ts` |
| 49 | `src/compiler/transformers/es2020.ts` |
| 50 | `src/compiler/transformers/es2021.ts` |
| 51 | `src/compiler/transformers/esnext.ts` |
| 52 | `src/compiler/transformers/jsx.ts` |
| 53 | `src/compiler/transformers/es2016.ts` |
| 54 | `src/compiler/transformers/es2015.ts` |
| 55 | `src/compiler/transformers/generators.ts` |
| 56 | `src/compiler/transformers/module/module.ts` |
| 57 | `src/compiler/transformers/module/system.ts` |
| 58 | `src/compiler/transformers/module/esnextAnd2015.ts` |
| 59 | `src/compiler/transformers/module/impliedNodeFormatDependent.ts` |
| 60 | `src/compiler/transformers/declarations/diagnostics.ts` |
| 61 | `src/compiler/transformers/declarations.ts` |
| 62 | `src/compiler/transformer.ts` |
| 63 | `src/compiler/emitter.ts` |
| 64 | `src/compiler/watchUtilities.ts` |
| 65 | `src/compiler/program.ts` |
| 66 | `src/compiler/programDiagnostics.ts` |
| 67 | `src/compiler/builderStatePublic.ts` |
| 68 | `src/compiler/builderState.ts` |
| 69 | `src/compiler/builder.ts` |
| 70 | `src/compiler/builderPublic.ts` |
| 71 | `src/compiler/resolutionCache.ts` |
| 72 | `src/compiler/watch.ts` |
| 73 | `src/compiler/watchPublic.ts` |
| 74 | `src/compiler/tsbuild.ts` |
| 75 | `src/compiler/tsbuildPublic.ts` |
| 76 | `src/compiler/executeCommandLine.ts` |
| 77 | `src/compiler/expressionToTypeNode.ts` |
| 78 | `src/compiler/_namespaces/ts.ts` |
| 79 | `src/tsc/_namespaces/ts.ts` |
| 80 | `src/tsc/tsc.ts` |
