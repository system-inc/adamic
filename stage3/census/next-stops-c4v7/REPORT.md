# Next stops on c4v-7

sys.ts:304:13: **cleared**. scanner.ts:2934:41: **cleared**. **12 new refusal-scan findings.**

Full combined report: [c4v7-census/REPORT.md](../c4v7-census/REPORT.md).

## Part 1 — next stops on c4v-7

All 79 roots ran twice in CommonJSBundle, then twice in EcmaScript: 316 observations. Duplicate observations agree except timing; 79/79 first-stop tuples agree across formats. Each completed group was checkpoint-pushed. CommonJSBundle was validated and pushed before EcmaScript started. One ordinary binary and the same adapted head tree served both formats.

The before column is the measured 5c8338c4 CommonJSBundle table from `typescript/next-stops-5c8338c4` at `d4f3ac9023013b7eb3d9dac4e8f61e1d64ecb76c`. That branch has no measured EcmaScript baseline. EcmaScript after values below are measured; its before reference is explicitly CommonJSBundle.

### Exact stop tuples

- **B01**: `-`; `accepted`; -. Phase: accepted.
- **B02**: `commandLineParser.ts:293:35`; `NotYet`; reading an enum before its runtime initialization; move the call after the enum declaration. Phase: lowering (initialization preflight).
- **B03**: `diagnosticInformationMap.generated.ts:13:45`; `NotYet`; reading an enum before its runtime initialization; move the call after the enum declaration. Phase: lowering (initialization preflight).
- **B04**: `scanner.ts:2934:41`; `Refused`; a value of type (TextRange & { value: number; })[] | undefined seen as {}[] | undefined, which can write {} where TextRange & { value: number; } is read. Phase: refusal_scan.
- **B05**: `semver.ts:71:9`; `NotYet`; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. Phase: lowering (initialization preflight).
- **B06**: `sys.ts:304:13`; `Refused`; a value of type (WatchedFileWithUnchangedPolls | undefined)[] seen as (WatchedFileWithIsClosed | undefined)[], which can write WatchedFileWithIsClosed | undefined where WatchedFileWithUnchangedPolls | undefined is read. Phase: refusal_scan.
- **A01**: `-`; `accepted`; -. Phase: accepted.
- **A02**: `commandLineParser.ts:293:35`; `NotYet`; reading an enum before its runtime initialization; move the call after the enum declaration. Phase: lowering (initialization preflight).
- **A03**: `diagnosticInformationMap.generated.ts:13:45`; `NotYet`; reading an enum before its runtime initialization; move the call after the enum declaration. Phase: lowering (initialization preflight).
- **A04**: `semver.ts:71:9`; `NotYet`; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. Phase: lowering (initialization preflight).
- **A05**: `sys.ts:514:17`; `Refused`; a cast the runtime can't check. Phase: refusal_scan.
- **A06**: `utilitiesPublic.ts:1280:52`; `Refused`; an unproven predicate argument for parameter predicate (argument "predicate"). Phase: refusal_scan.

### Every root, before and after

Each identifier expands to the exact location, kind and reason above. The two format reports also retain full tuple text per root: [CommonJSBundle](../next-stops-c4v7/CommonJSBundle/REPORT.md), [EcmaScript](../next-stops-c4v7/EcmaScript/REPORT.md).

| Root | Before: main CJS | After: head CJS | After: head ES |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | B06 | A05 | A05 |
| `_namespaces/ts.performance.ts` | B06 | A05 | A05 |
| `_namespaces/ts.ts` | B06 | A05 | A05 |
| `binder.ts` | B06 | A05 | A05 |
| `builder.ts` | B06 | A05 | A05 |
| `builderPublic.ts` | B06 | A05 | A05 |
| `builderState.ts` | B06 | A05 | A05 |
| `builderStatePublic.ts` | B06 | A05 | A05 |
| `checker.ts` | B06 | A05 | A05 |
| `commandLineParser.ts` | B06 | A05 | A05 |
| `core.ts` | B06 | A05 | A05 |
| `corePublic.ts` | B01 | A01 | A01 |
| `debug.ts` | B05 | A04 | A04 |
| `diagnosticInformationMap.generated.ts` | B02 | A02 | A02 |
| `emitter.ts` | B06 | A05 | A05 |
| `executeCommandLine.ts` | B06 | A05 | A05 |
| `expressionToTypeNode.ts` | B06 | A05 | A05 |
| `factory/baseNodeFactory.ts` | B06 | A05 | A05 |
| `factory/emitHelpers.ts` | B06 | A05 | A05 |
| `factory/emitNode.ts` | B06 | A05 | A05 |
| `factory/nodeChildren.ts` | B06 | A05 | A05 |
| `factory/nodeConverters.ts` | B06 | A05 | A05 |
| `factory/nodeFactory.ts` | B06 | A05 | A05 |
| `factory/nodeTests.ts` | B06 | A05 | A05 |
| `factory/parenthesizerRules.ts` | B06 | A05 | A05 |
| `factory/utilities.ts` | B06 | A05 | A05 |
| `factory/utilitiesPublic.ts` | B06 | A05 | A05 |
| `hostErrors.ts` | B01 | A01 | A01 |
| `moduleNameResolver.ts` | B06 | A05 | A05 |
| `moduleSpecifiers.ts` | B06 | A05 | A05 |
| `parser.ts` | B06 | A05 | A05 |
| `path.ts` | B06 | A05 | A05 |
| `performance.ts` | B06 | A05 | A05 |
| `performanceCore.ts` | B06 | A05 | A05 |
| `program.ts` | B06 | A05 | A05 |
| `programDiagnostics.ts` | B06 | A05 | A05 |
| `resolutionCache.ts` | B06 | A05 | A05 |
| `scanner.ts` | B06 | A05 | A05 |
| `semver.ts` | B06 | A05 | A05 |
| `sourcemap.ts` | B06 | A05 | A05 |
| `symbolWalker.ts` | B06 | A05 | A05 |
| `sys.ts` | B04 | A06 | A06 |
| `tracing.ts` | B06 | A05 | A05 |
| `transformer.ts` | B06 | A05 | A05 |
| `transformers/classFields.ts` | B06 | A05 | A05 |
| `transformers/classThis.ts` | B06 | A05 | A05 |
| `transformers/declarations/diagnostics.ts` | B06 | A05 | A05 |
| `transformers/declarations.ts` | B06 | A05 | A05 |
| `transformers/destructuring.ts` | B06 | A05 | A05 |
| `transformers/es2015.ts` | B06 | A05 | A05 |
| `transformers/es2016.ts` | B06 | A05 | A05 |
| `transformers/es2017.ts` | B06 | A05 | A05 |
| `transformers/es2018.ts` | B06 | A05 | A05 |
| `transformers/es2019.ts` | B06 | A05 | A05 |
| `transformers/es2020.ts` | B06 | A05 | A05 |
| `transformers/es2021.ts` | B06 | A05 | A05 |
| `transformers/esDecorators.ts` | B06 | A05 | A05 |
| `transformers/esnext.ts` | B06 | A05 | A05 |
| `transformers/generators.ts` | B06 | A05 | A05 |
| `transformers/jsx.ts` | B06 | A05 | A05 |
| `transformers/legacyDecorators.ts` | B06 | A05 | A05 |
| `transformers/module/esnextAnd2015.ts` | B06 | A05 | A05 |
| `transformers/module/impliedNodeFormatDependent.ts` | B06 | A05 | A05 |
| `transformers/module/module.ts` | B06 | A05 | A05 |
| `transformers/module/system.ts` | B06 | A05 | A05 |
| `transformers/namedEvaluation.ts` | B06 | A05 | A05 |
| `transformers/taggedTemplate.ts` | B06 | A05 | A05 |
| `transformers/ts.ts` | B06 | A05 | A05 |
| `transformers/typeSerializer.ts` | B06 | A05 | A05 |
| `transformers/utilities.ts` | B06 | A05 | A05 |
| `tsbuild.ts` | B06 | A05 | A05 |
| `tsbuildPublic.ts` | B06 | A05 | A05 |
| `types.ts` | B03 | A03 | A03 |
| `utilities.ts` | B06 | A05 | A05 |
| `utilitiesPublic.ts` | B06 | A05 | A05 |
| `visitorPublic.ts` | B06 | A05 | A05 |
| `watch.ts` | B06 | A05 | A05 |
| `watchPublic.ts` | B06 | A05 | A05 |
| `watchUtilities.ts` | B06 | A05 | A05 |

### Every new first-stop tuple and its adapted source line

#### sys.ts:514:17

```typescript
                (
```

Context:

```typescript
511:     else {
512:         cache.set(path, {
513:             watcher: createWatcher(
514:                 (
515:                     // Cant infer types correctly so lets satisfy checker
516:                     (param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))
517:                 ) as T,
```

Kind: `Refused`. Reason: a cast the runtime can't check. Phase: refusal_scan. Emitter: internal/lower/cast_proof.go:155 castRefusal (rule 1900), castProof refusal at line 195. Owner: createSingleWatcherPerName in adapted upstream sys.ts.

Source attribution: Inside generic body createSingleWatcherPerName<T extends FileWatcherCallback | FsWatchCallback> at sys.ts:498. The callback expression is cast as T in createWatcher argument 1 (whose parameter is callback: T). The same finding is already present in both main and head scan-only streams.


This is an existing main refusal-scan finding newly reached as a first stop.

#### utilitiesPublic.ts:1280:52

```typescript
    return find(getJSDocTagsWorker(node, noCache), predicate);
```

Context:

```typescript
1277: 
1278: /** Get the first JSDoc tag of a specified kind, or undefined if not present. */
1279: function getFirstJSDocTag<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T, noCache?: boolean): T | undefined {
1280:     return find(getJSDocTagsWorker(node, noCache), predicate);
1281: }
1282: 
1283: /** Gets all JSDoc tags that match a specified predicate */
```

Kind: `Refused`. Reason: an unproven predicate argument for parameter predicate (argument "predicate"). Phase: refusal_scan. Emitter: internal/lower/predicates.go:708 predicateArgumentsRefusal (rule 1700), refusal at line 779. Owner: getFirstJSDocTag in adapted upstream utilitiesPublic.ts.

Source attribution: Inside generic body getFirstJSDocTag<T extends JSDocTag> at utilitiesPublic.ts:1279. The forwarded predicate parameter is argument 2 of find, in its predicate slot. The same finding is already present in both main and head scan-only streams.


This is an existing main refusal-scan finding newly reached as a first stop.

### Lowering-phase first stops

- CommonJSBundle: `diagnosticInformationMap.generated.ts`, `types.ts`, `debug.ts`.
- EcmaScript: `diagnosticInformationMap.generated.ts`, `types.ts`, `debug.ts`.

Scan-only does not see these lowering-phase first stops. Complete source excerpts, paired root observations, process bounds, actual exits, grouped tuples and validations are under `../next-stops-c4v7/`. This is first-stop measurement, not successful native compilation or execution.

### Validation and provenance

Both formats passed the copied and retargeted validator. All child exits were zero; durations were below 1,800 seconds (process limits 1,780 seconds). Adapted source hashes and the ordinary binary hash were rechecked. All tracked changes relative to c4v-7 are in the two census evidence directories; no compiler or adaptation source changed. The requested branch was created under a new name and all pushes were non-forced.
