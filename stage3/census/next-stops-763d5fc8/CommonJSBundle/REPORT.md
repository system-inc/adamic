# Next stops: CommonJSBundle

Measured compiler/apply commit: `763d5fc83558ae8ceb0202189b18430dff499ab8`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

## Stops and movement from 55c375d9

- 73 roots: sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it). 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

## Adapted source at new stops

### sys.ts:1387:5

```typescript
    sys.writeFile = (path, data, writeBom) =>
```

Phase: lowering (deferred host-member-write refusal). Owner: 99-correctness-write-file-view: patchWriteFileEnsuringDirectory; internal/lower/refusals.go: methodValueRefusal / methodValueStop.

stage3/adapt/99-correctness-write-file-view/owners.json explicitly pins sys.ts:patchWriteFileEnsuringDirectory, including this assignment. The timer-closures textual match in TestServerHost does not own this compiler function. methodValueRefusal records the method-value error and defers a writable host-member assignment (refusals.go:262-268). lowerWithElementKey defers hostMemberPending (lower.go:114); hostMemberPending returns the first deferred write if lowering stops before judgement, while checkHostMemberWrites returns the recorded stop if it cannot admit the write. The ordinary record does not distinguish these return paths. The exact diagnostic says method read, but the measured source is an assignment target.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:1387:5 | Refused | a method read as a value (writeFile would lose its object, and this with it) | lowering (deferred host-member-write refusal) | 99-correctness-write-file-view: patchWriteFileEnsuringDirectory; internal/lower/refusals.go: methodValueRefusal / methodValueStop |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

## Root membership

- sys.ts:1387:5 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

## Comparison with 55c375d9

Baseline: typescript/next-stops-55c375d9 at `78a6c9dad71aeec7da329de10aea5387f75392b7`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:1385:31 roots now go to:

- 73 to sys.ts:1387:5: a method read as a value (writeFile would lose its object, and this with it). Forecast did not hold: the writeFile assignment is before all forecast sites.

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| _namespaces/ts.performance.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| _namespaces/ts.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| binder.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| builder.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| builderPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| builderState.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| builderStatePublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| checker.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| commandLineParser.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| core.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| executeCommandLine.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| expressionToTypeNode.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/baseNodeFactory.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/emitHelpers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/emitNode.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/nodeChildren.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/nodeConverters.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/nodeFactory.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/nodeTests.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/parenthesizerRules.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| factory/utilitiesPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| moduleSpecifiers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| parser.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| path.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| performance.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| performanceCore.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| program.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| programDiagnostics.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| resolutionCache.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| scanner.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| semver.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| sourcemap.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| symbolWalker.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformer.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/classFields.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/classThis.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/declarations/diagnostics.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/declarations.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/destructuring.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2015.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2016.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2017.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2018.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2019.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2020.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/es2021.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/esDecorators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/esnext.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/generators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/jsx.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/legacyDecorators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/module/esnextAnd2015.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/module/module.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/module/system.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/namedEvaluation.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/taggedTemplate.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/ts.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/typeSerializer.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| transformers/utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| tsbuild.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| tsbuildPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| utilitiesPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| visitorPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| watch.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| watchPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |
| watchUtilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) |

## Adapted source at stops

### sys.ts:1387:5

```typescript
1384:     // patch writefile to create folder before writing the file
1385:     const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
1386:     const originalWriteFile = writeFileView.writeFile;
1387:     sys.writeFile = (path, data, writeBom) =>
1388:         writeFileEnsuringDirectories(
1389:             path,
1390:             data,
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

### utilitiesPublic.ts:1280:52

```typescript
1277:
1278: /** Get the first JSDoc tag of a specified kind, or undefined if not present. */
1279: function getFirstJSDocTag<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T, noCache?: boolean): T | undefined {
1280:     return find(getJSDocTagsWorker(node, noCache), predicate);
1281: }
1282:
1283: /** Gets all JSDoc tags that match a specified predicate */
```

## Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from emitting source and stage3/adapt/**/owners.json searches retained in ownership-search.json.
