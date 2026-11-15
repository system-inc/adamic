**0 roots move to the sameKeepingWalk stop** ("one keeps something weakly that the other keeps strongly").

0/79 roots change first-stop tuples between main `89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825` and head `37d96306133ed2782184bee1bd8406d61b83a858`; 79 are unchanged. CommonJSBundle only. Complete observations (including diagnostics and fix text) are also compared across revisions, excluding timing. The baseline is freshly built and measured at main itself, beside the head.

No root has a different first stop.

## Where the 73 go

These are the same 73 roots grouped at sys.ts:1387:5 in the copied census. Membership is pinned in [former-73.json](former-73.json). Both current revisions were measured independently:

| Revision | Roots | First stop | Exact reason |
| --- | ---: | --- | --- |
| base | 73 | sys.ts:1387:5 (Refused) | a method read as a value (writeFile would lose its object, and this with it) |
| head | 73 | sys.ts:1387:5 (Refused) | a method read as a value (writeFile would lose its object, and this with it) |

## Head grouped first-stop table

Locations and root names are relative to the shared adapted `src/compiler/` tree.

| Roots | Location | Kind | Exact reason | Phase |
| ---: | --- | --- | --- | --- |
| 73 | sys.ts:1387:5 | Refused | a method read as a value (writeFile would lose its object, and this with it) | lowering exit / post-lowering host-write judgement (refusal captured during scan) |
| 2 | - | accepted | - | accepted |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | initialization preflight (before refusal scan and expression lowering) |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | initialization preflight (before refusal scan and expression lowering) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | initialization preflight (before refusal scan and expression lowering) |
| 1 | utilitiesPublic.ts:1284:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) |

## Root membership

- `sys.ts:1387:5` (Refused): `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`.
- `-` (accepted): `corePublic.ts`, `hostErrors.ts`.
- `commandLineParser.ts:293:35` (NotYet): `diagnosticInformationMap.generated.ts`.
- `diagnosticInformationMap.generated.ts:13:45` (NotYet): `types.ts`.
- `semver.ts:71:9` (NotYet): `debug.ts`.
- `utilitiesPublic.ts:1284:52` (Refused): `sys.ts`.

## Adapted source at changed stops

There are no changed stops between these revisions, so there are no changed-stop source lines to quote. All observed stop excerpts are retained below.

## Stop phases and source evidence

The new `sameKeepingWalk` guard is a lowering-phase check, before `present()`. A lowering-phase stop can become first only where a root has no earlier scan stop. `lowerWithElementKey` runs initialization checks and the refusal scan before declaration and expression lowering. The table reports the phase of the actual first diagnostic; it does not establish whether the new guard would be encountered after that diagnostic was removed. No stop-removal experiment was performed.

At head, `internal/lower/expression.go:413-414` returns true on `visited[t]` or the 64-type limit. Lines 467-485 add the union guard, which returns false. The caller at lines 2224-2231 emits `a <own type> seen as a <slot type> (one keeps something weakly that the other keeps strongly)`. This wording already existed on main. The census measures changes to the first-stop tuple; it does not instrument the guard.

### sys.ts:1387:5 (Refused)

Phase: lowering exit / post-lowering host-write judgement (refusal captured during scan). Emitter: internal/lower/refusals.go:305; internal/lower/host_member_write.go:112-117,142-178; internal/lower/lower.go:114.

The adapted line is an assignment target. methodValueRefusal captures the refusal in the scan and defers writable host-method assignments. lowerWithElementKey defers hostMemberPending, which returns the first pending write stop if lowering exits before judgement. checkHostMemberWrites can also return that recorded stop. Ordinary census output does not distinguish these paths, and therefore does not reveal any later stop masked by this one. The adaptation owner pins patchWriteFileEnsuringDirectory.

```typescript
1384:     // patch writefile to create folder before writing the file
1385:     const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
1386:     const originalWriteFile = writeFileView.writeFile;
1387:     sys.writeFile = (path, data, writeBom) =>
1388:         writeFileEnsuringDirectories(
1389:             path,
1390:             data,
```

### - (accepted)

Phase: accepted. Emitter: lower.LowerWithOptions (accepted; no stop).

### commandLineParser.ts:293:35 (NotYet)

Phase: initialization preflight (before refusal scan and expression lowering). Emitter: internal/lower/namespaces.go: (*lowering).namespaceInitialization.

```typescript
290:     {
291:         name: "watchFile",
292:         type: new Map(Object.entries({
293:             fixedpollinginterval: WatchFileKind.FixedPollingInterval,
294:             prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
295:             dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
296:             fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

### diagnosticInformationMap.generated.ts:13:45 (NotYet)

Phase: initialization preflight (before refusal scan and expression lowering). Emitter: internal/lower/namespaces.go: (*lowering).namespaceInitialization.

```typescript
10: /* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
11: /** @internal */
12: export const Diagnostics = {
13:     Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
14:     Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
15:     _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
16:     A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### semver.ts:71:9 (NotYet)

Phase: initialization preflight (before refusal scan and expression lowering). Emitter: internal/lower/namespaces.go: (*lowering).namespaceInitialization.

```typescript
68:         const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;
69:
70:         Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
71:         Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
72:
73:         this.major = major;
74:         this.minor = minor;
```

### utilitiesPublic.ts:1284:52 (Refused)

Phase: refusal_scan (predicate contracts). Emitter: internal/lower/lower.go:142 -> internal/lower/refusals.go:83 -> internal/lower/predicates.go:711,760-784 (predicateArguments / predicateArgumentProof).

The measured adapted line 1284 is the predicate argument in getFirstJSDocTag. The pinned refusal scan checks predicate arguments before expression lowering. No owners.json object explicitly owns getFirstJSDocTag (ownership-search.json).

```typescript
1281:
1282: /** Get the first JSDoc tag of a specified kind, or undefined if not present. */
1283: function getFirstJSDocTag<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T, noCache?: boolean): T | undefined {
1284:     return find(getJSDocTagsWorker(node, noCache), predicate);
1285: }
1286:
1287: /** Gets all JSDoc tags that match a specified predicate */
```

## Setup and validation

The two measurement worktrees were created fresh with `git worktree add --detach` under `/workspace/scratch/next-stops-37d96306/`, separately at each requested SHA. Both ran `git submodule update --init --recursive --depth 1`. TypeScript v6.0.3 was pinned at `050880ce59e30b356b686bd3144efe24f875ebc8`; `npm ci` in head/stage3/api exited 0. `git diff 89614cc0 37d96306 -- stage3` is empty: stage3 is identical. The adapted source was built once and reused for both ordinary census binaries, each built from its own compiler worktree. No compiler or adaptation source was changed. The shared checkout has no tracked changes.

GOTMPDIR, TMPDIR, GOCACHE, GOMODCACHE, npm cache and adaptation product storage are in scratch. Binary paths, hashes and sizes are recorded in binaries.json; binaries themselves are excluded. Any captured Go source has a `.go.txt` suffix. The adapted source manifest was frozen before measurement and checked afterward.

Each revision ran all 79 ordinary file roots, one root per child, in groups of at most ten, with three children at a time. Each child and group has a 1780-second timeout, allowing at most ten seconds for termination before kill, and actual exit codes and elapsed times recorded. All root children produce two observations equal in every field except timing: 316 raw observations in total. CommonJSBundle only; no separate sys.ts single-file run. This is a first-stop measurement and makes no claim about native execution.

The first setup checkpoint compressed two setup logs while their processes were active. Both complete logs were recovered from still-open file descriptors and replaced at the next checkpoint; checkpoint.py now compresses only logs with completed status records. No measurement was rerun for this.

## Evidence

[comparison.json](comparison.json) lists both stops for every root. `base/CommonJSBundle/` and `head/CommonJSBundle/` retain all paired observations, root and group statuses, grouped tables and source evidence. [validation.json](validation.json) records validation; [artifact-manifest.json](artifact-manifest.json) hashes artifacts; [compression.json](compression.json) hashes raw and compressed logs. [provenance.json](provenance.json) identifies the input commits. Checkpoint history is recorded in checkpoints.json.
