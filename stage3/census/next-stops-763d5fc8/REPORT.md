# Next stops on 763d5fc8

Measured commit: `763d5fc83558ae8ceb0202189b18430dff499ab8`. Parent: `936dfd7be5c8096d35d4320444f86a64d26fb058`. Baseline: typescript/next-stops-55c375d9 at `78a6c9dad71aeec7da329de10aea5387f75392b7` (both formats).

## Stops and movement from 55c375d9

### CommonJSBundle

- 73 roots: sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it). 73 moved, 0 unchanged from 55c375d9.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged from 55c375d9.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 55c375d9.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 55c375d9.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged from 55c375d9.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged from 55c375d9.

### EcmaScript

- 73 roots: sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it). 73 moved, 0 unchanged from 55c375d9.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged from 55c375d9.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 55c375d9.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from 55c375d9.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged from 55c375d9.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged from 55c375d9.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

The forecast was clearTimeout (task #7yz6rtp), watchFile/watchDirectory (task #cyjqp5f), or process.recordreplay (task #b7phwab), whichever was first in source order and still on main. Forecast locations in the 55c375d9 tree: sys.ts:1500:57, sys.ts:1598:57, sys.ts:1532:13, sys.ts:1533:13, sys.ts:1588:192.

CommonJSBundle: 73 of the former 73 roots meet sys.ts:1387:5 first. The forecast did not hold: the writeFile assignment in patchWriteFileEnsuringDirectory stops them before any of the forecast sites.

EcmaScript: 73 of the former 73 roots meet sys.ts:1387:5 first. The forecast did not hold: the writeFile assignment in patchWriteFileEnsuringDirectory stops them before any of the forecast sites.

## Adapted source at new stops

### sys.ts:1387:5

```typescript
    sys.writeFile = (path, data, writeBom) =>
```

Adapted context:

```typescript
1384:     // patch writefile to create folder before writing the file
1385:     const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
1386:     const originalWriteFile = writeFileView.writeFile;
1387:     sys.writeFile = (path, data, writeBom) =>
1388:         writeFileEnsuringDirectories(
1389:             path,
1390:             data,
```

Exact reason: a method read as a value (writeFile would lose its object, and this with it).

Phase: lowering (deferred host-member-write refusal). Owner: 99-correctness-write-file-view: patchWriteFileEnsuringDirectory; internal/lower/refusals.go: methodValueRefusal / methodValueStop. Emitter: internal/lower/refusals.go:305; internal/lower/host_member_write.go:112-117,142-178; internal/lower/lower.go:114.

Ownership evidence: stage3/adapt/99-correctness-write-file-view/owners.json explicitly pins sys.ts:patchWriteFileEnsuringDirectory, including this assignment. The timer-closures textual match in TestServerHost does not own this compiler function. methodValueRefusal records the method-value error and defers a writable host-member assignment (refusals.go:262-268). lowerWithElementKey defers hostMemberPending (lower.go:114); hostMemberPending returns the first deferred write if lowering stops before judgement, while checkHostMemberWrites returns the recorded stop if it cannot admit the write. The ordinary record does not distinguish these return paths. The exact diagnostic says method read, but the measured source is an assignment target..

## CommonJSBundle


Measured compiler/apply commit: `763d5fc83558ae8ceb0202189b18430dff499ab8`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

### Stops and movement from 55c375d9

- 73 roots: sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it). 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### Adapted source at new stops

#### sys.ts:1387:5

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

### Root membership

- sys.ts:1387:5 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 55c375d9

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

### Adapted source at stops

#### sys.ts:1387:5

```typescript
1384:     // patch writefile to create folder before writing the file
1385:     const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
1386:     const originalWriteFile = writeFileView.writeFile;
1387:     sys.writeFile = (path, data, writeBom) =>
1388:         writeFileEnsuringDirectories(
1389:             path,
1390:             data,
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


Measured compiler/apply commit: `763d5fc83558ae8ceb0202189b18430dff499ab8`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

### Stops and movement from 55c375d9

- 73 roots: sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it). 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### Adapted source at new stops

#### sys.ts:1387:5

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

### Root membership

- sys.ts:1387:5 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 55c375d9

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

### Adapted source at stops

#### sys.ts:1387:5

```typescript
1384:     // patch writefile to create folder before writing the file
1385:     const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
1386:     const originalWriteFile = writeFileView.writeFile;
1387:     sys.writeFile = (path, data, writeBom) =>
1388:         writeFileEnsuringDirectories(
1389:             path,
1390:             data,
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

The first 50-root checkpoint push lost GitHub credentials (exit 128); the exact failure and historical partial report are retained in checkpoint-push-failure.json and BLOCKED-AT-50.md. The user-requested retry succeeded, after which measurement resumed without repeating completed roots.

## Setup and limits

A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. An initial build attempted before nested submodule checkout completed exited 1 because cohere/TypeScript/tsc/go.mod was absent; its exact log and exit are retained as build-before-submodules. The retry after submodule initialization exited 0. The upstream clone confirmed v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. Both apply and build completed before measurement. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one successful ordinary build served both formats. The separate sys.ts single-file census was skipped.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | 65989b9c64969423b67ed59f3bd9c1969905227c |
| baobab: CommonJSBundle checkpoint 20/79 roots | 7ebd721a97b9716d13a683f0a7ee525cf50234ae |
| baobab: CommonJSBundle checkpoint 30/79 roots | be0a38bbc4fab5790168d8f8eb67370171d74e26 |
| baobab: CommonJSBundle checkpoint 40/79 roots | a40780236f0610ab684e35af18552f197a475dd5 |
| baobab: CommonJSBundle checkpoint 50/79 roots | c61309f47f08a0bcf5599ced090d225dc77a4bb3 |
| baobab: CommonJSBundle checkpoint 60/79 roots | 76a0bf1882a29d5e58a797d577dd58610f45b62d |
| baobab: CommonJSBundle checkpoint 70/79 roots | 40282801814006cd95cba247d740d97c8bebf089 |
| baobab: CommonJSBundle checkpoint 79/79 roots | c60f6de456f5a8d912738447acd4fb57233e6f6a |
| baobab: EcmaScript checkpoint 10/79 roots | 150dd76175c432fe5cd98e4dfe542b6245c09ee8 |
| baobab: EcmaScript checkpoint 20/79 roots | b2548b1b9216c98debbcae585a425a580fe8f8f3 |
| baobab: EcmaScript checkpoint 30/79 roots | 637b0101ae0dc3d7461d61bb9cab22db44e9a55a |
| baobab: EcmaScript checkpoint 40/79 roots | bde0bb0d4e140a9faa5353179c16570820868095 |
| baobab: EcmaScript checkpoint 50/79 roots | d2fe75d890a97816983e28387bc8d29d28b9538c |
| baobab: EcmaScript checkpoint 60/79 roots | 24e055d210a0c302ba55a5ba32cff23305000720 |
| baobab: EcmaScript checkpoint 70/79 roots | b771ef9d23ecc67a719c57637422e67b8b3d4a3c |
| baobab: EcmaScript checkpoint 79/79 roots | 1187072ec58f2abc3baea3354d9f86982f5be1ca |
