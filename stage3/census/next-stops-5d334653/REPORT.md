# Next stops on 5d334653 — CommonJSBundle complete

CommonJSBundle: all 79 roots measured twice and validated. EcmaScript has not started. Baseline: typescript/next-stops-9c322233 at 304ca436.

# Next stops: CommonJSBundle

Measured compiler/apply commit: `5d334653fa9ce301738e9eb2ef016c99da4589b8`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:290:19 | Refused | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | refusal_scan | adaptation 99-correctness-timer-closures (createDynamicPriorityPollingWatchFile) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:2934:41 | Refused | a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | refusal_scan | adaptation 70, stage 3 |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

## Root membership

- sys.ts:290:19 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:2934:41 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

## Comparison with 9c322233

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

## Adapted source at stops

### sys.ts:290:19

```typescript
287:
288:     function pollLowPollingIntervalQueue(_timeoutType: string, queue: PollingIntervalQueue) {
289:         // Always poll complete list of changedFilesInLastPoll
290:         pollQueue(changedFilesInLastPoll, PollingInterval.Low, /*pollIndex*/ 0, changedFilesInLastPoll.length);
291:
292:         // Finally do the actual polling of the queue
293:         pollPollingIntervalQueue(_timeoutType, queue);
```

### commandLineParser.ts:293:35

```typescript
290:     {
291:         name: "watchFile",
292:         type: new Map(Object.entries({
293:             fixedpollinginterval: WatchFileKind.FixedPollingInterval,
294:             prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
295:             dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
296:             fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

### diagnosticInformationMap.generated.ts:13:45

```typescript
10: /* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
11: /** @internal */
12: export const Diagnostics = {
13:     Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
14:     Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
15:     _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
16:     A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### scanner.ts:2934:41

```typescript
2931:             if (ch >= CharacterCodes._1 && ch <= CharacterCodes._9) {
2932:                 const start = pos;
2933:                 scanDigits();
2934:                 decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
2935:                 return true;
2936:             }
2937:             return false;
```

### semver.ts:71:9

```typescript
68:         const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;
69:
70:         Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
71:         Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
72:
73:         this.major = major;
74:         this.minor = minor;
```

## Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
