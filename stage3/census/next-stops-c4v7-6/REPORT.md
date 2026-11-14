# Next stops on c4v-7 task #p0fy047 (constraint-body-rigid-6)

CommonJSBundle: **sys.ts:304:13 cleared; scanner.ts:2934:41 cleared** (ordinary first-stop census).

EcmaScript: **sys.ts:304:13 cleared; scanner.ts:2934:41 cleared** (ordinary first-stop census).

Measured commit: `4f460a611b61a5b11c12ce61fde8ad5d0f11a912` (compiler/constraint-body-rigid-6, c4v-7 task #p0fy047). Parent: `a7d47283c73be5baa542965c1abbd3f5d5086fa4`. Baseline: `494781d17eb984d91f5db33d886ef41de1fd5004` on typescript/next-stops-5c8338c4.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

CommonJSBundle: 74 changed exact tuples; 5 unchanged from 5c8338c4.

Files lowering whole in CommonJSBundle: 2/79: corePublic.ts, hostErrors.ts.

EcmaScript: 74 changed exact tuples; 5 unchanged from 5c8338c4.

Files lowering whole in EcmaScript: 2/79: corePublic.ts, hostErrors.ts.

In each format, all 73 former sys.ts:304:13 roots now stop at sys.ts:514:17; the sys.ts root moves from scanner.ts:2934:41 to utilitiesPublic.ts:1280:52. The other five roots retain their exact baseline tuples. See each format below for every per-root comparison.


Main moved from 5c8338c4 to a7d47283 (#2t3h708). The exact internal/lower/expression.go diff is retained in [expression-change.diff.txt](expression-change.diff.txt). It changes sameKeeping and valueSlotKeeps to judge rest-element representations and reject hidden rest views recursively; it does not change castRefusal/castProof or initialization preflight. Reach assessments below are by emitting path, not a counterfactual compiler run.

## Adapted source at new stops

### sys.ts:514:17

```typescript
                (
```

Adapted context:

```typescript
511:     else {
512:         cache.set(path, {
513:             watcher: createWatcher(
514:                 (
515:                     // Cant infer types correctly so lets satisfy checker
516:                     (param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))
517:                 ) as T,
```

Exact reason: a cast the runtime can't check.

Phase: refusal_scan. Owner: internal/lower/cast_proof.go: (*lowering).castProof (createSingleWatcherPerName). Emitter: internal/lower/cast_proof.go: (*lowering).castRefusal -> castProof (rule 1900).

Ownership evidence: grep stage3/adapt/**/owners.json for createSingleWatcherPerName returns no match; no adaptation pins this function.

Reach of a7d47283 expression.go change: No: rule 1900 castRefusal/castProof refuses this as T cast during refusal_scan; castValue would call sameKeeping only after castProof succeeds. The a7d47283 expression.go rest-view checks are downstream and cannot emit this refusal.

### utilitiesPublic.ts:1280:52

```typescript
    return find(getJSDocTagsWorker(node, noCache), predicate);
```

Adapted context:

```typescript
1277:
1278: /** Get the first JSDoc tag of a specified kind, or undefined if not present. */
1279: function getFirstJSDocTag<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T, noCache?: boolean): T | undefined {
1280:     return find(getJSDocTagsWorker(node, noCache), predicate);
1281: }
1282:
1283: /** Gets all JSDoc tags that match a specified predicate */
```

Exact reason: an unproven predicate argument for parameter predicate (argument "predicate").

Phase: refusal_scan (predicate contracts). Owner: internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag). Emitter: internal/lower/refusals.go: (*lowering).refuse contract walk -> predicateArguments -> predicateArgumentProof; also registered rule 1700.

Ownership evidence: grep stage3/adapt/**/owners.json for getFirstJSDocTag returns no match; no adaptation pins this function.

Reach of a7d47283 expression.go change: No: predicateArguments/predicateArgumentProof rejects the forwarded predicate parameter in the refusal-scan contract pass. This is a type-predicate proof requirement, not a rest-element or hidden-rest representation check; the changed sameKeeping/valueSlotKeeps paths do not emit this stop.

## CommonJSBundle


Measured compiler/apply commit: `4f460a611b61a5b11c12ce61fde8ad5d0f11a912`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:514:17 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go: (*lowering).castProof (createSingleWatcherPerName) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

### Root membership

- sys.ts:514:17 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 5c8338c4

Baseline: typescript/next-stops-5c8338c4 at `494781d17eb984d91f5db33d886ef41de1fd5004`.

5 unchanged exact tuples; 74 changed.

The former 73 sys.ts:304:13 roots now go to:

- 73 to sys.ts:514:17: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| core.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| path.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| program.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |

### Adapted source at stops

#### sys.ts:514:17

```typescript
511:     else {
512:         cache.set(path, {
513:             watcher: createWatcher(
514:                 (
515:                     // Cant infer types correctly so lets satisfy checker
516:                     (param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))
517:                 ) as T,
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

#### expression.go reach: sys.ts:514:17

No: rule 1900 castRefusal/castProof refuses this as T cast during refusal_scan; castValue would call sameKeeping only after castProof succeeds. The a7d47283 expression.go rest-view checks are downstream and cannot emit this refusal.

#### expression.go reach: utilitiesPublic.ts:1280:52

No: predicateArguments/predicateArgumentProof rejects the forwarded predicate parameter in the refusal-scan contract pass. This is a type-predicate proof requirement, not a rest-element or hidden-rest representation check; the changed sameKeeping/valueSlotKeeps paths do not emit this stop.

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
## EcmaScript


Measured compiler/apply commit: `4f460a611b61a5b11c12ce61fde8ad5d0f11a912`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:514:17 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go: (*lowering).castProof (createSingleWatcherPerName) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

### Root membership

- sys.ts:514:17 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 5c8338c4

Baseline: typescript/next-stops-5c8338c4 at `494781d17eb984d91f5db33d886ef41de1fd5004`.

5 unchanged exact tuples; 74 changed.

The former 73 sys.ts:304:13 roots now go to:

- 73 to sys.ts:514:17: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| core.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| path.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| program.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:304:13; Refused; a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | sys.ts:514:17; Refused; a cast the runtime can't check |

### Adapted source at stops

#### sys.ts:514:17

```typescript
511:     else {
512:         cache.set(path, {
513:             watcher: createWatcher(
514:                 (
515:                     // Cant infer types correctly so lets satisfy checker
516:                     (param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))
517:                 ) as T,
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

#### expression.go reach: sys.ts:514:17

No: rule 1900 castRefusal/castProof refuses this as T cast during refusal_scan; castValue would call sameKeeping only after castProof succeeds. The a7d47283 expression.go rest-view checks are downstream and cannot emit this refusal.

#### expression.go reach: utilitiesPublic.ts:1280:52

No: predicateArguments/predicateArgumentProof rejects the forwarded predicate parameter in the refusal-scan contract pass. This is a type-predicate proof requirement, not a rest-element or hidden-rest representation check; the changed sameKeeping/valueSlotKeeps paths do not emit this stop.

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
## Setup and limits

A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one ordinary build served both formats. The separate sys.ts single-file census was skipped.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | 9e0b7c201f1c83d097332c5b234310cd97a0f98e |
| baobab: CommonJSBundle checkpoint 20/79 roots | 76be2f6e3eb2ddbedea760b4c7cdc8c16c93132f |
| baobab: CommonJSBundle checkpoint 30/79 roots | 5f2a74ad0a460bb264a0eafdb2f261b579a80b0a |
| baobab: CommonJSBundle checkpoint 40/79 roots | 467b655adfef910fa29991bcfb769f56f2347808 |
| baobab: CommonJSBundle checkpoint 50/79 roots | 2a76a8c2b3fe8c15ba3dbff7935e1d21f003db67 |
| baobab: CommonJSBundle checkpoint 60/79 roots | 94ef821b618dc851c2701a4db091f8656f924a55 |
| baobab: CommonJSBundle checkpoint 70/79 roots | 713039368c088ab2311900f9e3b741080487d1a5 |
| baobab: CommonJSBundle checkpoint 79/79 roots | c2522195d7f659793b3536aa428d61d9e6fb1b9f |
| baobab: EcmaScript checkpoint 10/79 roots | eec1737c1b778672d083a953685e796d2a319035 |
| baobab: EcmaScript checkpoint 20/79 roots | f1eab10d46d13b3079b31c63d6f521c5cbcd0054 |
| baobab: EcmaScript checkpoint 30/79 roots | ac53c0ac106f8cdd15d3ffe44d157b064ae43ed2 |
| baobab: EcmaScript checkpoint 40/79 roots | 435bfb1410d1c8c3ec6086df2b2c2584d1e0e335 |
| baobab: EcmaScript checkpoint 50/79 roots | d923709773b6526b38878a0f0bfadb8a6d51116d |
| baobab: EcmaScript checkpoint 60/79 roots | 8dabb9a7f76a0d79ee2ac9593d6cc95766086f2e |
| baobab: EcmaScript checkpoint 70/79 roots | 7460632d01f9af4434c74952936d535f4a4ac974 |
| baobab: EcmaScript checkpoint 79/79 roots | 9ae779e6fbe9777ae66c0bda19a77f611b672036 |
