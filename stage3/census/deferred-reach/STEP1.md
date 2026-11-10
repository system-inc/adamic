# Step 1: deferred generic reach

Task abrq61p. Pinned main: `bdbc73864f82dbc0c51434decf0ffe96d4782767`.

Upstream TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Setup and API npm ci exited 0. Adaptation exited 0. No census instrumentation was changed for these measurements.

The directory and entry passes run in separate processes. Each boolean is finalized only after every file in that process has been processed. It records concrete generic body entry, not successful lowering.

## Compiler directory

Scope: `src/compiler`, 79 implementation files, first-error mode. Header checker_rejected: `false`. The overlay retains its fixed measurement label.

Raw SHA-256: `5ac570c066c7adc8131e99da3634f9d510eeadfac03f3ef7b8007a63c6fea312`. Child exit: `0`.

| File | All units | Deferred, instantiated true | Deferred, instantiated false |
| --- | ---: | ---: | ---: |
| src/compiler/core.ts | 264 | 61 | 146 |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 1 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 1 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 78 | 0 | 0 |
| src/compiler/binder.ts | 16 | 0 | 0 |
| src/compiler/builder.ts | 86 | 1 | 1 |
| src/compiler/builderPublic.ts | 16 | 0 | 0 |
| src/compiler/builderState.ts | 5 | 0 | 0 |
| src/compiler/builderStatePublic.ts | 3 | 0 | 0 |
| src/compiler/checker.ts | 51 | 0 | 0 |
| src/compiler/commandLineParser.ts | 171 | 6 | 6 |
| src/compiler/corePublic.ts | 9 | 0 | 0 |
| src/compiler/debug.ts | 5 | 0 | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | 3 | 0 | 0 |
| src/compiler/emitter.ts | 50 | 0 | 2 |
| src/compiler/executeCommandLine.ts | 50 | 0 | 1 |
| src/compiler/expressionToTypeNode.ts | 9 | 0 | 0 |
| src/compiler/factory/baseNodeFactory.ts | 3 | 0 | 0 |
| src/compiler/factory/emitHelpers.ts | 45 | 0 | 0 |
| src/compiler/factory/emitNode.ts | 41 | 3 | 20 |
| src/compiler/factory/nodeChildren.ts | 6 | 0 | 0 |
| src/compiler/factory/nodeConverters.ts | 3 | 0 | 0 |
| src/compiler/factory/nodeFactory.ts | 26 | 1 | 0 |
| src/compiler/factory/nodeTests.ts | 228 | 0 | 0 |
| src/compiler/factory/parenthesizerRules.ts | 3 | 0 | 0 |
| src/compiler/factory/utilities.ts | 108 | 2 | 6 |
| src/compiler/factory/utilitiesPublic.ts | 4 | 1 | 0 |
| src/compiler/hostErrors.ts | 3 | 0 | 0 |
| src/compiler/moduleNameResolver.ts | 163 | 4 | 11 |
| src/compiler/moduleSpecifiers.ts | 49 | 0 | 1 |
| src/compiler/parser.ts | 75 | 1 | 21 |
| src/compiler/path.ts | 81 | 2 | 2 |
| src/compiler/performance.ts | 24 | 0 | 0 |
| src/compiler/performanceCore.ts | 12 | 0 | 0 |
| src/compiler/program.ts | 85 | 1 | 4 |
| src/compiler/programDiagnostics.ts | 5 | 0 | 0 |
| src/compiler/resolutionCache.ts | 31 | 0 | 0 |
| src/compiler/scanner.ts | 84 | 4 | 6 |
| src/compiler/semver.ts | 28 | 0 | 0 |
| src/compiler/sourcemap.ts | 31 | 0 | 1 |
| src/compiler/symbolWalker.ts | 2 | 0 | 0 |
| src/compiler/sys.ts | 62 | 0 | 2 |
| src/compiler/tracing.ts | 10 | 0 | 0 |
| src/compiler/transformer.ts | 17 | 1 | 1 |
| src/compiler/transformers/classFields.ts | 23 | 0 | 0 |
| src/compiler/transformers/classThis.ts | 7 | 0 | 2 |
| src/compiler/transformers/declarations.ts | 14 | 0 | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | 8 | 0 | 0 |
| src/compiler/transformers/destructuring.ts | 22 | 0 | 0 |
| src/compiler/transformers/es2015.ts | 13 | 0 | 0 |
| src/compiler/transformers/es2016.ts | 2 | 0 | 0 |
| src/compiler/transformers/es2017.ts | 6 | 0 | 0 |
| src/compiler/transformers/es2018.ts | 4 | 0 | 0 |
| src/compiler/transformers/es2019.ts | 2 | 0 | 0 |
| src/compiler/transformers/es2020.ts | 2 | 0 | 0 |
| src/compiler/transformers/es2021.ts | 2 | 0 | 0 |
| src/compiler/transformers/esDecorators.ts | 9 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 9 | 0 | 0 |
| src/compiler/transformers/generators.ts | 16 | 0 | 0 |
| src/compiler/transformers/jsx.ts | 3 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 2 | 0 | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | 2 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 2 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 3 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 2 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 21 | 0 | 2 |
| src/compiler/transformers/taggedTemplate.ts | 5 | 0 | 0 |
| src/compiler/transformers/ts.ts | 5 | 0 | 0 |
| src/compiler/transformers/typeSerializer.ts | 6 | 0 | 0 |
| src/compiler/transformers/utilities.ts | 50 | 0 | 6 |
| src/compiler/tsbuild.ts | 5 | 0 | 0 |
| src/compiler/tsbuildPublic.ts | 106 | 2 | 70 |
| src/compiler/types.ts | 875 | 0 | 0 |
| src/compiler/utilities.ts | 893 | 13 | 46 |
| src/compiler/utilitiesPublic.ts | 251 | 6 | 3 |
| src/compiler/visitorPublic.ts | 41 | 2 | 11 |
| src/compiler/watch.ts | 46 | 4 | 6 |
| src/compiler/watchPublic.ts | 25 | 0 | 7 |
| src/compiler/watchUtilities.ts | 26 | 0 | 5 |
| **All files** | **4555** | **115** | **389** |

## tsc entry

Scope: `src/tsc/tsc.ts`, 81 implementation files, first-error mode. Header checker_rejected: `false`. The overlay retains its fixed measurement label.

Raw SHA-256: `da5c6eeb845cd1f773dd3e24ad7a515ab0a7f36d4859bcb6cd756748a44fd430`. Child exit: `0`.

| File | All units | Deferred, instantiated true | Deferred, instantiated false |
| --- | ---: | ---: | ---: |
| src/compiler/core.ts | 264 | 61 | 146 |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 1 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 1 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 78 | 0 | 0 |
| src/compiler/binder.ts | 16 | 0 | 0 |
| src/compiler/builder.ts | 86 | 1 | 1 |
| src/compiler/builderPublic.ts | 16 | 0 | 0 |
| src/compiler/builderState.ts | 5 | 0 | 0 |
| src/compiler/builderStatePublic.ts | 3 | 0 | 0 |
| src/compiler/checker.ts | 51 | 0 | 0 |
| src/compiler/commandLineParser.ts | 171 | 6 | 6 |
| src/compiler/corePublic.ts | 9 | 0 | 0 |
| src/compiler/debug.ts | 5 | 0 | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | 3 | 0 | 0 |
| src/compiler/emitter.ts | 50 | 0 | 2 |
| src/compiler/executeCommandLine.ts | 50 | 0 | 1 |
| src/compiler/expressionToTypeNode.ts | 9 | 0 | 0 |
| src/compiler/factory/baseNodeFactory.ts | 3 | 0 | 0 |
| src/compiler/factory/emitHelpers.ts | 45 | 0 | 0 |
| src/compiler/factory/emitNode.ts | 41 | 3 | 20 |
| src/compiler/factory/nodeChildren.ts | 6 | 0 | 0 |
| src/compiler/factory/nodeConverters.ts | 3 | 0 | 0 |
| src/compiler/factory/nodeFactory.ts | 26 | 1 | 0 |
| src/compiler/factory/nodeTests.ts | 228 | 0 | 0 |
| src/compiler/factory/parenthesizerRules.ts | 3 | 0 | 0 |
| src/compiler/factory/utilities.ts | 108 | 2 | 6 |
| src/compiler/factory/utilitiesPublic.ts | 4 | 1 | 0 |
| src/compiler/hostErrors.ts | 3 | 0 | 0 |
| src/compiler/moduleNameResolver.ts | 163 | 4 | 11 |
| src/compiler/moduleSpecifiers.ts | 49 | 0 | 1 |
| src/compiler/parser.ts | 75 | 1 | 21 |
| src/compiler/path.ts | 81 | 2 | 2 |
| src/compiler/performance.ts | 24 | 0 | 0 |
| src/compiler/performanceCore.ts | 12 | 0 | 0 |
| src/compiler/program.ts | 85 | 1 | 4 |
| src/compiler/programDiagnostics.ts | 5 | 0 | 0 |
| src/compiler/resolutionCache.ts | 31 | 0 | 0 |
| src/compiler/scanner.ts | 84 | 4 | 6 |
| src/compiler/semver.ts | 28 | 0 | 0 |
| src/compiler/sourcemap.ts | 31 | 0 | 1 |
| src/compiler/symbolWalker.ts | 2 | 0 | 0 |
| src/compiler/sys.ts | 62 | 0 | 2 |
| src/compiler/tracing.ts | 10 | 0 | 0 |
| src/compiler/transformer.ts | 17 | 1 | 1 |
| src/compiler/transformers/classFields.ts | 23 | 0 | 0 |
| src/compiler/transformers/classThis.ts | 7 | 0 | 2 |
| src/compiler/transformers/declarations.ts | 14 | 0 | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | 8 | 0 | 0 |
| src/compiler/transformers/destructuring.ts | 22 | 0 | 0 |
| src/compiler/transformers/es2015.ts | 13 | 0 | 0 |
| src/compiler/transformers/es2016.ts | 2 | 0 | 0 |
| src/compiler/transformers/es2017.ts | 6 | 0 | 0 |
| src/compiler/transformers/es2018.ts | 4 | 0 | 0 |
| src/compiler/transformers/es2019.ts | 2 | 0 | 0 |
| src/compiler/transformers/es2020.ts | 2 | 0 | 0 |
| src/compiler/transformers/es2021.ts | 2 | 0 | 0 |
| src/compiler/transformers/esDecorators.ts | 9 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 9 | 0 | 0 |
| src/compiler/transformers/generators.ts | 16 | 0 | 0 |
| src/compiler/transformers/jsx.ts | 3 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 2 | 0 | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | 2 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 2 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 3 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 2 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 21 | 0 | 2 |
| src/compiler/transformers/taggedTemplate.ts | 5 | 0 | 0 |
| src/compiler/transformers/ts.ts | 5 | 0 | 0 |
| src/compiler/transformers/typeSerializer.ts | 6 | 0 | 0 |
| src/compiler/transformers/utilities.ts | 50 | 0 | 6 |
| src/compiler/tsbuild.ts | 5 | 0 | 0 |
| src/compiler/tsbuildPublic.ts | 106 | 2 | 70 |
| src/compiler/types.ts | 875 | 0 | 0 |
| src/compiler/utilities.ts | 893 | 13 | 46 |
| src/compiler/utilitiesPublic.ts | 251 | 6 | 3 |
| src/compiler/visitorPublic.ts | 41 | 2 | 11 |
| src/compiler/watch.ts | 46 | 4 | 6 |
| src/compiler/watchPublic.ts | 25 | 0 | 7 |
| src/compiler/watchUtilities.ts | 26 | 0 | 5 |
| src/tsc/_namespaces/ts.ts | 1 | 0 | 0 |
| src/tsc/tsc.ts | 6 | 0 | 0 |
| **All files** | **4562** | **115** | **389** |

## Reproduction

The baseline overlay was generated from the pinned main checkout with `make_overlay.py`, formatted with gofmt, and built with `go build -buildvcs=false -overlay=OVERLAY/overlay.json ./stage3/census/latent/tool`. The entry overlay extends it with `stage3/meter/entry_overlay.py`, exactly as the daily meter does.

Both invocations set `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=TIMES.jsonl`. Arguments were the completed adapted root shown above and a fresh output JSONL path. The ordinary meter entry and 79-file checker passes are retained separately.

Scratch evidence is under `/workspace/scratch/abrq61p/`; adaptation and setup logs are `/tmp/abrq61p-apply.log`, `/tmp/abrq61p-setup.log`, and `/tmp/abrq61p-main-setup.log`. All 712 adapted source hashes are in `sources.json`.

Environment transitions killed the wrappers while their census children survived. Replacement monitors recover each orphan child's own exit from `/proc/PID/stat`; no census child was restarted. A monitor replacement exited 143; this is not a census exit. The guard stops a child after 30 minutes without log or timing/output growth.

## Status and classification limits

Only Step 1 is delivered. Step 2 was not started; Step 3 has preliminary reference and declaration evidence only. Cold setup/adaptation took about 34 minutes; the complete fresh census runs consumed most of the two-hour turn window. A second complete census needed to verify a tooling change would not fit the remaining window. No witness field, mutant, reader compatibility claim or final classification is claimed.

The stock TypeScript inventory finds 137 bodyless generic overload signatures among the 504 directory deferrals, including 84 of core.ts's 207. A bodyless signature has no executable body to enter. It is not dead code merely because its inventory unit lacks a body-entry flag.

The TypeScript 6.0.3 language service searched all 712 TypeScript files under src/. All 207 core generics have semantic references and direct calls; 175 have calls outside src/compiler/, including services and harness code that these censuses do not lower. Those callers have no census first finding. The requested three classes therefore need an additional explicit treatment for bodyless signatures and callers outside the measured scope. No unreached unit is labeled hidden from references alone.

STEP1.json retains one row per deferred unit for each scope, with its unchanged boolean and independent body/reference counts. These are measurement facts, not the requested caller-stop classification. The compressed raw streams and source hashes in evidence/ preserve the exact baseline for a future comparison.

No production compiler or census tooling source changed. The requested final tool and report branches were not created.


On unchanged main, 4 bodyless signatures nevertheless have instantiated true. These are entry-hook flags, not proof of executable body coverage:

- `skipOuterExpressions`, `src/compiler/factory/utilities.ts:654:1`
- `findAncestor`, `src/compiler/utilitiesPublic.ts:786:1`
- `getParseTreeNode`, `src/compiler/utilitiesPublic.ts:826:1`
- `visitNode`, `src/compiler/visitorPublic.ts:142:1`
