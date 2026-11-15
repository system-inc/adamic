# Next stops on 763d5fc8 — partial, blocked at checkpoint push

Measured commit: `763d5fc83558ae8ceb0202189b18430dff499ab8`. Baseline: `typescript/next-stops-55c375d9` at `78a6c9dad71aeec7da329de10aea5387f75392b7` (both formats).

## Stops and movement from 55c375d9

### CommonJSBundle — 50/79 measured

- 45 roots: sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it). 45 moved, 0 unchanged.
- 2 roots: -; accepted; -. 0 moved, 2 unchanged.
- 1 root: commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration. 0 moved, 1 unchanged.
- 1 root: semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. 0 moved, 1 unchanged.
- 1 root: utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate"). 0 moved, 1 unchanged.

### New stop: sys.ts:1387:5

```typescript
    sys.writeFile = (path, data, writeBom) =>
```

Owning adaptation: `stage3/adapt/99-correctness-write-file-view/owners.json` explicitly pins `patchWriteFileEnsuringDirectory`. The stop is the assignment target. The exact diagnostic remains “a method read as a value (writeFile would lose its object, and this with it)”.

45 of the former 73 roots were measured: all 45 meet this assignment first, before the forecast clearTimeout (#7yz6rtp), watchFile/watchDirectory (#cyjqp5f), or process.recordreplay (#b7phwab) sites. The forecast did not hold for these 45. The remaining 28 of that former group are unmeasured, so this is not an all-73 conclusion.

### EcmaScript — 0/79 measured

Not started: the recipe requires CommonJSBundle to be complete, validated and pushed first. No second apply or build occurred.

## Blocker and publication state

The 50-root CommonJSBundle checkpoint committed locally as `c61309f47f08a0bcf5599ced090d225dc77a4bb3`, then its push exited 128. An identical push retry and a read-only `git ls-remote origin` check failed with the same error:

```text
fatal: could not read Username for 'https://github.com': No such device or address
```

Last confirmed successful remote checkpoint: `a40780236` (40/79 roots), on `typescript/next-stops-763d5fc8-checkpoint`. The requested final branch `typescript/next-stops-763d5fc8` has not been pushed. This partial report is local. Restore repository credentials before resuming at the failed checkpoint push; do not rerun completed measurements.

## Grouped table — CommonJSBundle measured subset

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 45 | sys.ts:1387:5 | Refused | a method read as a value (writeFile would lose its object, and this with it) | lowering (deferred host-member-write refusal) | 99-correctness-write-file-view: patchWriteFileEnsuringDirectory; internal/lower/refusals.go: methodValueRefusal / methodValueStop |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

## Root membership

- sys.ts:1387:5 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- semver.ts:71:9 NotYet: `debug.ts`
- utilitiesPublic.ts:1280:52 Refused: `sys.ts`

Files lowering whole: 2/50 measured: `corePublic.ts`, `hostErrors.ts`. Outcomes for 29 roots remain unknown. No native-output success is claimed.

## Per-root comparison with 55c375d9

All 79 roots are listed. CommonJSBundle has 50 measured outcomes; every EcmaScript outcome is unmeasured.

| Root | 55c375d9 (location; kind; exact reason) | CommonJSBundle at 763d5fc8 | EcmaScript at 763d5fc8 |
| --- | --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| _namespaces/ts.performance.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| _namespaces/ts.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| binder.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| builder.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| builderPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| builderState.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| builderStatePublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| checker.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| commandLineParser.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| core.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| corePublic.ts | -; accepted; - | -; accepted; - | UNMEASURED |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | UNMEASURED |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | UNMEASURED |
| emitter.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| executeCommandLine.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| expressionToTypeNode.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/baseNodeFactory.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/emitHelpers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/emitNode.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/nodeChildren.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/nodeConverters.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/nodeFactory.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/nodeTests.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/parenthesizerRules.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| factory/utilitiesPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| hostErrors.ts | -; accepted; - | -; accepted; - | UNMEASURED |
| moduleNameResolver.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| moduleSpecifiers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| parser.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| path.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| performance.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| performanceCore.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| program.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| programDiagnostics.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| resolutionCache.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| scanner.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| semver.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| sourcemap.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| symbolWalker.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | UNMEASURED |
| tracing.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformer.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/classFields.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/classThis.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/declarations.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/declarations/diagnostics.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/destructuring.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/es2015.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1387:5; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED |
| transformers/es2016.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/es2017.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/es2018.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/es2019.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/es2020.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/es2021.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/esDecorators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/esnext.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/generators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/jsx.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/legacyDecorators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/module/esnextAnd2015.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/module/module.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/module/system.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/namedEvaluation.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/taggedTemplate.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/ts.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/typeSerializer.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| transformers/utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| tsbuild.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| tsbuildPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | UNMEASURED | UNMEASURED |
| utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| utilitiesPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| visitorPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| watch.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| watchPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |
| watchUtilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | UNMEASURED | UNMEASURED |

## Adapted source at measured stops

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

Phase: lowering (deferred host-member-write refusal). Owner: 99-correctness-write-file-view: patchWriteFileEnsuringDirectory; internal/lower/refusals.go: methodValueRefusal / methodValueStop. Emitter: internal/lower/refusals.go:305; internal/lower/host_member_write.go:112-117,142-178; internal/lower/lower.go:114.

stage3/adapt/99-correctness-write-file-view/owners.json explicitly pins sys.ts:patchWriteFileEnsuringDirectory, including this assignment. The timer-closures textual match in TestServerHost does not own this compiler function. methodValueRefusal records the method-value error and defers a writable host-member assignment (refusals.go:262-268). lowerWithElementKey defers hostMemberPending (lower.go:114); hostMemberPending returns the first deferred write if lowering stops before judgement, while checkHostMemberWrites returns the recorded stop if it cannot admit the write. The ordinary record does not distinguish these return paths. The exact diagnostic says method read, but the measured source is an assignment target.

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

Phase: lowering (initialization preflight). Owner: internal/lower/namespaces.go: (*lowering).namespaceInitialization. Emitter: internal/lower/namespaces.go: (*lowering).namespaceInitialization.



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

Phase: lowering (initialization preflight). Owner: internal/lower/namespaces.go: (*lowering).namespaceInitialization. Emitter: internal/lower/namespaces.go: (*lowering).namespaceInitialization.



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

Phase: refusal_scan (predicate contracts). Owner: internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag). Emitter: internal/lower/refusals.go: (*lowering).refuse contract walk -> predicateArguments -> predicateArgumentProof; also registered rule 1700.

grep stage3/adapt/**/owners.json for getFirstJSDocTag returns no match; no adaptation pins this function

## Setup, method and limits

A fresh detached worktree was created from an isolated scratch clone; the shared checkout was not modified. Recursive submodules initialized successfully. Microsoft TypeScript was cloned at `050880ce59e30b356b686bd3144efe24f875ebc8`; package version `6.0.3` confirmed. `stage3/api` npm ci exited 0. All generated products and caches used scratch.

An initial build before recursive checkout completed exited 1 with `go: cannot load module cohere/TypeScript/tsc listed in go.work file: open cohere/TypeScript/tsc/go.mod: no such file or directory`. Its log and exit are retained separately. The retry after submodule initialization exited 0. One successful build and one apply (exit 0) produced the measured binary and tree. No adaptation or compiler changes were made. The separate sys.ts single-file census was skipped.

Each completed root ran in its own child twice, observations equal except timing. Groups contain at most ten roots and run three children at a time. Child and group bounds are 1780 seconds, with exits recorded. All 50 measured roots and five groups exited 0 and stayed below 1800 seconds. The checkpoint push failure prevents continuation under the requested checkpoint order.

`partial-validation.json` verifies only the measured subset, source hashes, binary hash, logs and scope. No full-census or cross-format validation is claimed. `setup.json`, `checkpoint-push-failure.json`, paired root JSONL files, status files, source evidence and compressed logs preserve the evidence. The copied nine scripts target the full 79-root census and should be resumed after credentials are restored.
