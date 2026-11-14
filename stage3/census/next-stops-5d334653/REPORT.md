# Next stops on 5d334653

Measured commit: `5d334653fa9ce301738e9eb2ef016c99da4589b8` (main, "Compose the polling queue with its files array so its layout lowers"). Parent: `8fe3dc44147e54180612ef4713196d40f02f384e`. Baseline: `304ca436e64c3b83119f0c6246af2259dbfe5402` on typescript/next-stops-9c322233.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

CommonJSBundle: 73 changed exact tuples; 6 unchanged from 9c322233.

Files lowering whole in CommonJSBundle: 2/79: corePublic.ts, hostErrors.ts.

EcmaScript: 73 changed exact tuples; 6 unchanged from 9c322233.

Files lowering whole in EcmaScript: 2/79: corePublic.ts, hostErrors.ts.

All 73 roots formerly stopped at the sys.ts:272:23 queue cast now stop at sys.ts:290:19 in both formats. The other six roots retain their exact baseline tuples.


## Adapted source at new stops

### sys.ts:290:19

```typescript
        pollQueue(changedFilesInLastPoll, PollingInterval.Low, /*pollIndex*/ 0, changedFilesInLastPoll.length);
```

Exact reason: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls | undefined)[], which can write WatchedFileWithUnchangedPolls | undefined where WatchedFileWithUnchangedPolls is read.

Phase: refusal_scan. Owner: adaptation 99-correctness-timer-closures (createDynamicPriorityPollingWatchFile). Emitter: internal/lower/invariance.go: (*lowering).wideningRefusal (via refuseWidening and stringWideningRefusal, rule 3100).

Ownership evidence: stage3/adapt/99-correctness-timer-closures/owners.json pins createDynamicPriorityPollingWatchFile; nested pollLowPollingIntervalQueue contains the stop.

## CommonJSBundle


Measured compiler/apply commit: `5d334653fa9ce301738e9eb2ef016c99da4589b8`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:290:19 | Refused | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | refusal_scan | adaptation 99-correctness-timer-closures (createDynamicPriorityPollingWatchFile) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:2934:41 | Refused | a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | refusal_scan | adaptation 70, stage 3 |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:290:19 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:2934:41 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 9c322233

Baseline: typescript/next-stops-9c322233 at `304ca436e64c3b83119f0c6246af2259dbfe5402`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:272:23 roots now go to:

- 73 to sys.ts:290:19: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls | undefined)[], which can write WatchedFileWithUnchangedPolls | undefined where WatchedFileWithUnchangedPolls is read

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| _namespaces/ts.performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| _namespaces/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| binder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builderPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builderState.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builderStatePublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| checker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| commandLineParser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| core.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| executeCommandLine.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| expressionToTypeNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/baseNodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/emitHelpers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/emitNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeChildren.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeConverters.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeTests.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/parenthesizerRules.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| parser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| path.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| performanceCore.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| program.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| programDiagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| resolutionCache.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| scanner.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| semver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| sourcemap.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| symbolWalker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| sys.ts | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read |
| tracing.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/classFields.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/classThis.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/declarations/diagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/declarations.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/destructuring.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2016.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2017.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2018.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2019.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2020.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2021.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/esDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/esnext.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/generators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/jsx.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/legacyDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/esnextAnd2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/module.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/system.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/namedEvaluation.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/taggedTemplate.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/typeSerializer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| tsbuild.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| tsbuildPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| visitorPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| watch.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| watchPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| watchUtilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |

### Adapted source at stops

#### sys.ts:290:19

```typescript
287:
288:     function pollLowPollingIntervalQueue(_timeoutType: string, queue: PollingIntervalQueue) {
289:         // Always poll complete list of changedFilesInLastPoll
290:         pollQueue(changedFilesInLastPoll, PollingInterval.Low, /*pollIndex*/ 0, changedFilesInLastPoll.length);
291:
292:         // Finally do the actual polling of the queue
293:         pollPollingIntervalQueue(_timeoutType, queue);
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

#### scanner.ts:2934:41

```typescript
2931:             if (ch >= CharacterCodes._1 && ch <= CharacterCodes._9) {
2932:                 const start = pos;
2933:                 scanDigits();
2934:                 decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
2935:                 return true;
2936:             }
2937:             return false;
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

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
## EcmaScript


Measured compiler/apply commit: `5d334653fa9ce301738e9eb2ef016c99da4589b8`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:290:19 | Refused | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | refusal_scan | adaptation 99-correctness-timer-closures (createDynamicPriorityPollingWatchFile) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:2934:41 | Refused | a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | refusal_scan | adaptation 70, stage 3 |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:290:19 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:2934:41 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 9c322233

Baseline: typescript/next-stops-9c322233 at `304ca436e64c3b83119f0c6246af2259dbfe5402`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:272:23 roots now go to:

- 73 to sys.ts:290:19: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls | undefined)[], which can write WatchedFileWithUnchangedPolls | undefined where WatchedFileWithUnchangedPolls is read

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| _namespaces/ts.performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| _namespaces/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| binder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builderPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builderState.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| builderStatePublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| checker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| commandLineParser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| core.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| executeCommandLine.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| expressionToTypeNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/baseNodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/emitHelpers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/emitNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeChildren.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeConverters.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/nodeTests.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/parenthesizerRules.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| factory/utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| parser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| path.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| performanceCore.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| program.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| programDiagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| resolutionCache.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| scanner.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| semver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| sourcemap.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| symbolWalker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| sys.ts | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read |
| tracing.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/classFields.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/classThis.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/declarations/diagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/declarations.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/destructuring.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2016.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2017.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2018.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2019.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2020.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/es2021.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/esDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/esnext.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/generators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/jsx.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/legacyDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/esnextAnd2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/module.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/module/system.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/namedEvaluation.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/taggedTemplate.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/typeSerializer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| transformers/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| tsbuild.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| tsbuildPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| visitorPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| watch.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| watchPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| watchUtilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:290:19; Refused; a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |

### Adapted source at stops

#### sys.ts:290:19

```typescript
287:
288:     function pollLowPollingIntervalQueue(_timeoutType: string, queue: PollingIntervalQueue) {
289:         // Always poll complete list of changedFilesInLastPoll
290:         pollQueue(changedFilesInLastPoll, PollingInterval.Low, /*pollIndex*/ 0, changedFilesInLastPoll.length);
291:
292:         // Finally do the actual polling of the queue
293:         pollPollingIntervalQueue(_timeoutType, queue);
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

#### scanner.ts:2934:41

```typescript
2931:             if (ch >= CharacterCodes._1 && ch <= CharacterCodes._9) {
2932:                 const start = pos;
2933:                 scanDigits();
2934:                 decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
2935:                 return true;
2936:             }
2937:             return false;
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

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
## Setup and limits

A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. All npm cache writes used scratch. No shared checkout writes were made. No compiler or adaptation changes were made. One completed apply and one ordinary build served both formats. The separate sys.ts single-file census was skipped.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | ff9894cabf97cb12b788d195ee21eb1d655edfa7 |
| baobab: CommonJSBundle checkpoint 20/79 roots | ea099299fab5af22d030637d7b685a972229ac6e |
| baobab: CommonJSBundle checkpoint 30/79 roots | 82a00f56dc908b5f4893452dc598ca335debf7ce |
| baobab: CommonJSBundle checkpoint 40/79 roots | b63b49cc0ff4413cb75b4316b9bdeba14aa4d410 |
| baobab: CommonJSBundle checkpoint 50/79 roots | 08c1934eedd4e41832efc98fb3cf8e3e87641a55 |
| baobab: CommonJSBundle checkpoint 60/79 roots | 7f97359e6c6efd5eb138d0fc6330d347377e359e |
| baobab: CommonJSBundle checkpoint 70/79 roots | 155ec0c5dee5e2b72f1dc0e0f68fa7ac8cb1d978 |
| baobab: CommonJSBundle checkpoint 79/79 roots | 640df5262357682a389683e5ee89297e8c25c318 |
| baobab: EcmaScript checkpoint 10/79 roots | 0da3a43326c6fe349ee86c4c2047146a368cdfae |
| baobab: EcmaScript checkpoint 20/79 roots | d6963394f23004204f559c5161e0da596821ef10 |
| baobab: EcmaScript checkpoint 30/79 roots | f926c0962eb27b9cd55aa40b590b22ba560c51f3 |
| baobab: EcmaScript checkpoint 40/79 roots | 81c7315fbf75ed16f869550c5d779e084bf45182 |
| baobab: EcmaScript checkpoint 50/79 roots | 537ebe96235be2ecb2475515db2cebae50e682ce |
| baobab: EcmaScript checkpoint 60/79 roots | 7faa0f8c1f74837e4822b2876d77e3b9ddd535af |
| baobab: EcmaScript checkpoint 70/79 roots | 0e6b1d462ad2cb3f8765f321fdfa9f6e2eef9c4c |
| baobab: EcmaScript checkpoint 79/79 roots | dea654777670f62050c35c6e1583ca8c0a4e3d50 |
