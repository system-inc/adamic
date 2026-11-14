# Next stops on 8e3a784a

Measured commit: `8e3a784aa5a601f57342c07b68f145fe56252c21`. Parent: `d7e350fca62de13d64b8e91f742b53312a917762`. Baseline: typescript/next-stops-7534c1d7 at `4eb05ced35dd2e90e3c688fdf55b0aab7d79e649` (both formats).

## Stops and movement from 7534c1d7

### CommonJSBundle

- 73 roots: sys.ts:790:87; Refused; a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read. 73 moved, 0 unchanged from 7534c1d7.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged from 7534c1d7.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 7534c1d7.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 7534c1d7.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged from 7534c1d7.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged from 7534c1d7.

### EcmaScript

- 73 roots: sys.ts:790:87; Refused; a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read. 73 moved, 0 unchanged from 7534c1d7.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged from 7534c1d7.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 7534c1d7.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 7534c1d7.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged from 7534c1d7.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged from 7534c1d7.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

Neither sys.ts:515:39 (#036ve6w) nor sys.ts:594:5 (#trxxkge) is first for the former 73 roots: all stop at sys.ts:790:87 during refusal_scan, before body lowering.

## Adapted source at new stops

### sys.ts:790:87

```typescript
            if (!cache.get(dirPath)?.targetWatcher) invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
```

Adapted context:

```typescript
787:             // Because the child refresh is fresh, we would need to invalidate whole root directory being watched
788:             // to ensure that all the changes are reflected at this time
789:             const hasChanges = updateChildWatches(dirName, dirPath, options);
790:             if (!cache.get(dirPath)?.targetWatcher) invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
791:         }
792:
793:         sysLog(`sysLog:: invokingWatchers:: Elapsed:: ${timestamp() - start}ms:: ${cacheToUpdateChildWatches.size}`);
```

Exact reason: a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read.

Phase: refusal_scan (mutable invariance). Owner: internal/lower/invariance.go: (*lowering).wideningRefusal (onTimerToUpdateChildWatches, nested in createDirectoryWatcherSupportingRecursive). Emitter: internal/lower/invariance.go:1178.

Ownership evidence: stage3/adapt/99-correctness-timer-closures/owners.json pins src/compiler/sys.ts:createDirectoryWatcherSupportingRecursive (ownerAfterSha256 fc2cddfdae88032897bb5bf01255c880b34685e5e37cb899caf4a7430b9cbdf2); onTimerToUpdateChildWatches is nested inside it. Mutable widening is checked in refusal_scan before declareModule or body lowering (internal/lower/lower.go:141-152), so sys.ts:790:87 preempts both sys.ts:515:39 (#036ve6w) optionalCall and sys.ts:594:5 (#trxxkge) destructuredField. Neither candidate is reached first.

## CommonJSBundle


Measured compiler/apply commit: `8e3a784aa5a601f57342c07b68f145fe56252c21`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

### Stops and movement from 7534c1d7

- 73 roots: sys.ts:790:87; Refused; a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read. 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### Adapted source at new stops

#### sys.ts:790:87

```typescript
            if (!cache.get(dirPath)?.targetWatcher) invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
```

Phase: refusal_scan (mutable invariance). Owner: internal/lower/invariance.go: (*lowering).wideningRefusal (onTimerToUpdateChildWatches, nested in createDirectoryWatcherSupportingRecursive).

stage3/adapt/99-correctness-timer-closures/owners.json pins src/compiler/sys.ts:createDirectoryWatcherSupportingRecursive (ownerAfterSha256 fc2cddfdae88032897bb5bf01255c880b34685e5e37cb899caf4a7430b9cbdf2); onTimerToUpdateChildWatches is nested inside it. Mutable widening is checked in refusal_scan before declareModule or body lowering (internal/lower/lower.go:141-152), so sys.ts:790:87 preempts both sys.ts:515:39 (#036ve6w) optionalCall and sys.ts:594:5 (#trxxkge) destructuredField. Neither candidate is reached first

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:790:87 | Refused | a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read | refusal_scan (mutable invariance) | internal/lower/invariance.go: (*lowering).wideningRefusal (onTimerToUpdateChildWatches, nested in createDirectoryWatcherSupportingRecursive) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

### Root membership

- sys.ts:790:87 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 7534c1d7

Baseline: typescript/next-stops-7534c1d7 at `4eb05ced35dd2e90e3c688fdf55b0aab7d79e649`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:618:33 roots now go to:

- 73 to sys.ts:790:87: a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| _namespaces/ts.performance.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| _namespaces/ts.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| binder.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builder.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builderPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builderState.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builderStatePublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| checker.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| commandLineParser.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| core.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| executeCommandLine.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| expressionToTypeNode.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/baseNodeFactory.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/emitHelpers.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/emitNode.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeChildren.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeConverters.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeFactory.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeTests.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/parenthesizerRules.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/utilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/utilitiesPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| moduleSpecifiers.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| parser.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| path.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| performance.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| performanceCore.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| program.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| programDiagnostics.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| resolutionCache.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| scanner.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| semver.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| sourcemap.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| symbolWalker.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformer.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/classFields.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/classThis.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/declarations/diagnostics.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/declarations.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/destructuring.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2015.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2016.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2017.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2018.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2019.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2020.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2021.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/esDecorators.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/esnext.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/generators.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/jsx.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/legacyDecorators.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/esnextAnd2015.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/module.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/system.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/namedEvaluation.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/taggedTemplate.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/ts.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/typeSerializer.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/utilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| tsbuild.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| tsbuildPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| utilitiesPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| visitorPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| watch.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| watchPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| watchUtilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |

### Adapted source at stops

#### sys.ts:790:87

```typescript
787:             // Because the child refresh is fresh, we would need to invalidate whole root directory being watched
788:             // to ensure that all the changes are reflected at this time
789:             const hasChanges = updateChildWatches(dirName, dirPath, options);
790:             if (!cache.get(dirPath)?.targetWatcher) invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
791:         }
792:
793:         sysLog(`sysLog:: invokingWatchers:: Elapsed:: ${timestamp() - start}ms:: ${cacheToUpdateChildWatches.size}`);
```

#### commandLineParser.ts:293:35

```typescript
290:     {
291:         name: "watchFile",
292:         type: new Map(Object.entries({
293:             fixedpollinginterval: WatchFileKind.FixedPollingInterval,
294:             prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
295:             dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
296:             fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

#### diagnosticInformationMap.generated.ts:13:45

```typescript
10: /* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
11: /** @internal */
12: export const Diagnostics = {
13:     Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
14:     Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
15:     _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
16:     A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

#### semver.ts:71:9

```typescript
68:         const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;
69:
70:         Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
71:         Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
72:
73:         this.major = major;
74:         this.minor = minor;
```

#### utilitiesPublic.ts:1280:52

```typescript
1277:
1278: /** Get the first JSDoc tag of a specified kind, or undefined if not present. */
1279: function getFirstJSDocTag<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T, noCache?: boolean): T | undefined {
1280:     return find(getJSDocTagsWorker(node, noCache), predicate);
1281: }
1282:
1283: /** Gets all JSDoc tags that match a specified predicate */
```

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from emitting source and stage3/adapt/**/owners.json searches retained in ownership-search.json.
## EcmaScript


Measured compiler/apply commit: `8e3a784aa5a601f57342c07b68f145fe56252c21`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

### Stops and movement from 7534c1d7

- 73 roots: sys.ts:790:87; Refused; a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read. 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### Adapted source at new stops

#### sys.ts:790:87

```typescript
            if (!cache.get(dirPath)?.targetWatcher) invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
```

Phase: refusal_scan (mutable invariance). Owner: internal/lower/invariance.go: (*lowering).wideningRefusal (onTimerToUpdateChildWatches, nested in createDirectoryWatcherSupportingRecursive).

stage3/adapt/99-correctness-timer-closures/owners.json pins src/compiler/sys.ts:createDirectoryWatcherSupportingRecursive (ownerAfterSha256 fc2cddfdae88032897bb5bf01255c880b34685e5e37cb899caf4a7430b9cbdf2); onTimerToUpdateChildWatches is nested inside it. Mutable widening is checked in refusal_scan before declareModule or body lowering (internal/lower/lower.go:141-152), so sys.ts:790:87 preempts both sys.ts:515:39 (#036ve6w) optionalCall and sys.ts:594:5 (#trxxkge) destructuredField. Neither candidate is reached first

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:790:87 | Refused | a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read | refusal_scan (mutable invariance) | internal/lower/invariance.go: (*lowering).wideningRefusal (onTimerToUpdateChildWatches, nested in createDirectoryWatcherSupportingRecursive) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

### Root membership

- sys.ts:790:87 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 7534c1d7

Baseline: typescript/next-stops-7534c1d7 at `4eb05ced35dd2e90e3c688fdf55b0aab7d79e649`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:618:33 roots now go to:

- 73 to sys.ts:790:87: a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| _namespaces/ts.performance.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| _namespaces/ts.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| binder.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builder.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builderPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builderState.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| builderStatePublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| checker.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| commandLineParser.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| core.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| executeCommandLine.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| expressionToTypeNode.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/baseNodeFactory.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/emitHelpers.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/emitNode.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeChildren.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeConverters.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeFactory.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/nodeTests.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/parenthesizerRules.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/utilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| factory/utilitiesPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| moduleSpecifiers.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| parser.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| path.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| performance.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| performanceCore.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| program.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| programDiagnostics.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| resolutionCache.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| scanner.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| semver.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| sourcemap.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| symbolWalker.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformer.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/classFields.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/classThis.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/declarations/diagnostics.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/declarations.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/destructuring.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2015.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2016.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2017.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2018.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2019.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2020.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/es2021.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/esDecorators.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/esnext.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/generators.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/jsx.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/legacyDecorators.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/esnextAnd2015.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/module.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/module/system.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/namedEvaluation.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/taggedTemplate.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/ts.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/typeSerializer.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| transformers/utilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| tsbuild.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| tsbuildPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| utilitiesPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| visitorPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| watch.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| watchPublic.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |
| watchUtilities.ts | sys.ts:618:33; Refused; a cast the runtime can't check | sys.ts:790:87; Refused; a value of type Map&lt;Path, string[]&gt; seen as InvokeMap, which can write true \| string[] where string[] is read |

### Adapted source at stops

#### sys.ts:790:87

```typescript
787:             // Because the child refresh is fresh, we would need to invalidate whole root directory being watched
788:             // to ensure that all the changes are reflected at this time
789:             const hasChanges = updateChildWatches(dirName, dirPath, options);
790:             if (!cache.get(dirPath)?.targetWatcher) invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
791:         }
792:
793:         sysLog(`sysLog:: invokingWatchers:: Elapsed:: ${timestamp() - start}ms:: ${cacheToUpdateChildWatches.size}`);
```

#### commandLineParser.ts:293:35

```typescript
290:     {
291:         name: "watchFile",
292:         type: new Map(Object.entries({
293:             fixedpollinginterval: WatchFileKind.FixedPollingInterval,
294:             prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
295:             dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
296:             fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

#### diagnosticInformationMap.generated.ts:13:45

```typescript
10: /* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
11: /** @internal */
12: export const Diagnostics = {
13:     Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
14:     Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
15:     _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
16:     A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

#### semver.ts:71:9

```typescript
68:         const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;
69:
70:         Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
71:         Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
72:
73:         this.major = major;
74:         this.minor = minor;
```

#### utilitiesPublic.ts:1280:52

```typescript
1277:
1278: /** Get the first JSDoc tag of a specified kind, or undefined if not present. */
1279: function getFirstJSDocTag<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T, noCache?: boolean): T | undefined {
1280:     return find(getJSDocTagsWorker(node, noCache), predicate);
1281: }
1282:
1283: /** Gets all JSDoc tags that match a specified predicate */
```

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from emitting source and stage3/adapt/**/owners.json searches retained in ownership-search.json.

## Setup and limits

A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. Both apply and build completed before measurement. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one successful ordinary build served both formats. The separate sys.ts single-file census was skipped.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

Git HTTPS credentials stopped working after the EcmaScript 60-root checkpoint. The 70- and 79-root Git pushes exited 128: `fatal: could not read Username for 'https://github.com': No such device or address`. The connected GitHub tool published the remaining objects and advanced refs without force; each published checkpoint tree was verified identical to its local tree. publication.json maps the local and published checkpoint commits. The measured SHA includes enum-never commit 1d3b46dc9e85f9c1a555fcae39f1c125c8a181f9 (verified ancestry).


## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | f4d59b439482388eaa9463f3a58d81928b93a700 |
| baobab: CommonJSBundle checkpoint 20/79 roots | b5c6210c93d8c269748e63fe23afb40ed705daed |
| baobab: CommonJSBundle checkpoint 30/79 roots | c664743dfa4ab4cdced875f063420889dcdf5c52 |
| baobab: CommonJSBundle checkpoint 40/79 roots | f85ea83bbcc81d804260c65b6614fcdedbef7a26 |
| baobab: CommonJSBundle checkpoint 50/79 roots | b17728a6c4cddcde72689afb1dc2d5187d60a0ae |
| baobab: CommonJSBundle checkpoint 60/79 roots | bcadb324906c006e235a2ac02699a93c8ef29149 |
| baobab: CommonJSBundle checkpoint 70/79 roots | f123c4541a0aee066999e756de377f3e638b8823 |
| baobab: CommonJSBundle checkpoint 79/79 roots | 055b4f50daab9a797b7f2cf56f19c9d550c6b062 |
| baobab: EcmaScript checkpoint 10/79 roots | 5347807fd21cda10823c1f33040aede00035ca0d |
| baobab: EcmaScript checkpoint 20/79 roots | d6d8f67b94a843454e181ddfb150bd6521b40af3 |
| baobab: EcmaScript checkpoint 30/79 roots | b846276daac75a4caa55cf8302e6956e8f566f78 |
| baobab: EcmaScript checkpoint 40/79 roots | 0fb8539ebdd7780dc3f9f0379d8fdf221745a9c2 |
| baobab: EcmaScript checkpoint 50/79 roots | e7e8f110beb798427721676d63e02a27f879199f |
| baobab: EcmaScript checkpoint 60/79 roots | d8aaead4b59b45f1b2b50430ac798170be73f103 |
| baobab: EcmaScript checkpoint 70/79 roots | 3fe53edd1f8d2143df2b8009d4046d49919713b3 |
| baobab: EcmaScript checkpoint 79/79 roots | b30e44b103fe0a987b979b4f74683a7b28aed154 |
