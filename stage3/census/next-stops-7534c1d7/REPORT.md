# Next stops on 7534c1d7

Measured commit: `7534c1d7575db02bb7b6512119e0685987b7e333`. Parent: `a358dba98e777bee7574f6e01cad4aaa29e4c4c3`. Baseline: typescript/next-stops-ff4ce03f at `2147e992cf42301c1e20c64c8d39c9ab027c65d7` (both formats).

## Stops and movement from ff4ce03f

### CommonJSBundle

- 73 roots: sys.ts:618:33; Refused; a cast the runtime can't check. 73 moved, 0 unchanged from ff4ce03f.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged from ff4ce03f.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from ff4ce03f.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from ff4ce03f.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged from ff4ce03f.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged from ff4ce03f.

### EcmaScript

- 73 roots: sys.ts:618:33; Refused; a cast the runtime can't check. 73 moved, 0 unchanged from ff4ce03f.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged from ff4ce03f.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from ff4ce03f.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged from ff4ce03f.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged from ff4ce03f.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged from ff4ce03f.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

## Adapted source at new stops

### sys.ts:618:33

```typescript
    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
```

Adapted context:

```typescript
615:     let timerToUpdateChildWatches: any;
616:
617:     const filePathComparer = getStringComparer(!useCaseSensitiveFileNames);
618:     const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
619:
620:     return (dirName, callback, recursive, options) =>
621:         recursive ?
```

Exact reason: a cast the runtime can't check.

Phase: refusal_scan. Owner: internal/lower/cast_proof.go: (*lowering).castProof (createDirectoryWatcherSupportingRecursive). Emitter: internal/lower/cast_proof.go:197.

Ownership evidence: stage3/adapt/99-correctness-timer-closures/owners.json pins src/compiler/sys.ts:createDirectoryWatcherSupportingRecursive; the line is unchanged by the single-watcher forwarder adaptation. lower.go runs refuse over modules before body lowering, so this cast preempts sys.ts:515:39 (#036ve6w) optionalCall in body lowering.

## CommonJSBundle


Measured compiler/apply commit: `7534c1d7575db02bb7b6512119e0685987b7e333`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

### Stops and movement from ff4ce03f

- 73 roots: sys.ts:618:33; Refused; a cast the runtime can't check. 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### Adapted source at new stops

#### sys.ts:618:33

```typescript
    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
```

Phase: refusal_scan. Owner: internal/lower/cast_proof.go: (*lowering).castProof (createDirectoryWatcherSupportingRecursive).

stage3/adapt/99-correctness-timer-closures/owners.json pins src/compiler/sys.ts:createDirectoryWatcherSupportingRecursive; the line is unchanged by the single-watcher forwarder adaptation. lower.go runs refuse over modules before body lowering, so this cast preempts sys.ts:515:39 (#036ve6w) optionalCall in body lowering

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:618:33 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go: (*lowering).castProof (createDirectoryWatcherSupportingRecursive) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

### Root membership

- sys.ts:618:33 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with ff4ce03f

Baseline: typescript/next-stops-ff4ce03f at `2147e992cf42301c1e20c64c8d39c9ab027c65d7`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:514:17 roots now go to:

- 73 to sys.ts:618:33: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| core.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| path.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| program.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |

### Adapted source at stops

#### sys.ts:618:33

```typescript
615:     let timerToUpdateChildWatches: any;
616:
617:     const filePathComparer = getStringComparer(!useCaseSensitiveFileNames);
618:     const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
619:
620:     return (dirName, callback, recursive, options) =>
621:         recursive ?
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


Measured compiler/apply commit: `7534c1d7575db02bb7b6512119e0685987b7e333`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

### Stops and movement from ff4ce03f

- 73 roots: sys.ts:618:33; Refused; a cast the runtime can't check. 73 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### Adapted source at new stops

#### sys.ts:618:33

```typescript
    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
```

Phase: refusal_scan. Owner: internal/lower/cast_proof.go: (*lowering).castProof (createDirectoryWatcherSupportingRecursive).

stage3/adapt/99-correctness-timer-closures/owners.json pins src/compiler/sys.ts:createDirectoryWatcherSupportingRecursive; the line is unchanged by the single-watcher forwarder adaptation. lower.go runs refuse over modules before body lowering, so this cast preempts sys.ts:515:39 (#036ve6w) optionalCall in body lowering

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:618:33 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go: (*lowering).castProof (createDirectoryWatcherSupportingRecursive) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

### Root membership

- sys.ts:618:33 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with ff4ce03f

Baseline: typescript/next-stops-ff4ce03f at `2147e992cf42301c1e20c64c8d39c9ab027c65d7`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:514:17 roots now go to:

- 73 to sys.ts:618:33: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| core.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| path.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| program.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") |
| tracing.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:514:17; Refused; a cast the runtime can't check | sys.ts:618:33; Refused; a cast the runtime can't check |

### Adapted source at stops

#### sys.ts:618:33

```typescript
615:     let timerToUpdateChildWatches: any;
616:
617:     const filePathComparer = getStringComparer(!useCaseSensitiveFileNames);
618:     const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
619:
620:     return (dirName, callback, recursive, options) =>
621:         recursive ?
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

A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. The successful build completed while apply was running; both were complete before measurement. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one successful ordinary build served both formats. A premature build before recursive submodule checkout finished exited 1: `go: cannot load module cohere/TypeScript/tsc listed in go.work file: open cohere/TypeScript/tsc/go.mod: no such file or directory`. That attempt is retained as build-before-submodules; the completed-submodule retry exited 0. The separate sys.ts single-file census was skipped.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | e97bbf2982457cd8e60386ca76d6608a23d842a9 |
| baobab: CommonJSBundle checkpoint 20/79 roots | 1aa1265301ff90c1f841bb07f28ba972b893ff6f |
| baobab: CommonJSBundle checkpoint 30/79 roots | 23cdf0f72a5db091778cdc9e926cea222b472451 |
| baobab: CommonJSBundle checkpoint 40/79 roots | eca9d9ebe17428d5eeb42633eaecd641f7d1eb91 |
| baobab: CommonJSBundle checkpoint 50/79 roots | c7f9c09fb9efb3770c90adfcee2ac4082ba5ffe0 |
| baobab: CommonJSBundle checkpoint 60/79 roots | 1d1e05cd5b9d91636cb13931a759796532b6c6c6 |
| baobab: CommonJSBundle checkpoint 70/79 roots | 2aaba4d4151a714b3226f5e3146810504d7f6dd6 |
| baobab: CommonJSBundle checkpoint 79/79 roots | fe65a169a191ed298c44881d6473e5defca313fa |
| baobab: EcmaScript checkpoint 10/79 roots | e82a5ab601f03c1623592fbd813540d24c9a4a4c |
| baobab: EcmaScript checkpoint 20/79 roots | 56422754e3238bd1a3c5e4f4cbbd660312345b28 |
| baobab: EcmaScript checkpoint 30/79 roots | fcf306a334f0ded148a2ae7813f296031ac33ee4 |
| baobab: EcmaScript checkpoint 40/79 roots | 243e2b1decfbb972bb8c30d1227e77500b66ce1f |
| baobab: EcmaScript checkpoint 50/79 roots | 1c6aec83bc3f3bec28c9e7d0eb5e8bac6eada7d1 |
| baobab: EcmaScript checkpoint 60/79 roots | af1aaa0b41c253b41d8b5896f6e844840b830152 |
| baobab: EcmaScript checkpoint 70/79 roots | c678ae89223bed4b40458a0a5190cfa38296eec9 |
| baobab: EcmaScript checkpoint 79/79 roots | 24f546671f22ce58962bc324088aad86fc51c04c |
