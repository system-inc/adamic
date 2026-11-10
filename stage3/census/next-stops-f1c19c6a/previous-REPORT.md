# Next stops on main 7910ab94

Measured for tasks baobab and b5w3ycg. Report only.

Compiler and apply SHA: `7910ab94a42b0c7f62a79329d7f250b1b2045d05`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `37ba255ddb18`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-10T03:19:37.102052+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Compared with b322b6ba | Root names |
| ---: | --- | --- | --- | --- | --- |
| 74 | `debug.ts:525:37` | Refused | a cast the runtime can't check | 1 unchanged, 73 moved | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 unchanged | `debug.ts` (artifact classification incomplete) |

All 73 roots that stopped at core.ts:2599:80 in b322b6ba now stop at debug.ts:525:37. The other six roots are unchanged. There are no new tuples.

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with b322b6ba

Baseline: b322b6ba2f12454443bbbbdaaa6844fdce22a440, from typescript/next-stops-b322b6ba. Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `debug.ts:525:37` Refused | 74 | 1 | 73 | 0 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |

Total: 6 unchanged, 73 moved, 0 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `_namespaces/ts.performance.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `_namespaces/ts.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `binder.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `builder.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `builderPublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `builderState.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `builderStatePublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `checker.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `commandLineParser.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `core.ts` | unchanged | `debug.ts:525:37` Refused | `debug.ts:525:37` Refused |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `executeCommandLine.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `expressionToTypeNode.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/baseNodeFactory.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/emitHelpers.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/emitNode.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/nodeChildren.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/nodeConverters.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/nodeFactory.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/nodeTests.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/parenthesizerRules.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/utilities.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `factory/utilitiesPublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `moduleSpecifiers.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `parser.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `path.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `performance.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `performanceCore.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `program.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `programDiagnostics.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `resolutionCache.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `scanner.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `semver.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `sourcemap.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `symbolWalker.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `sys.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `tracing.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformer.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/classFields.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/classThis.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/declarations/diagnostics.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/declarations.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/destructuring.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2015.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2016.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2017.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2018.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2019.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2020.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/es2021.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/esDecorators.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/esnext.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/generators.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/jsx.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/legacyDecorators.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/module/esnextAnd2015.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/module/impliedNodeFormatDependent.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/module/module.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/module/system.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/namedEvaluation.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/taggedTemplate.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/ts.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/typeSerializer.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `transformers/utilities.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `tsbuild.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `tsbuildPublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `utilitiesPublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `visitorPublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `watch.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `watchPublic.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |
| `watchUtilities.ts` | moved | `core.ts:2599:80` Refused | `debug.ts:525:37` Refused |

## Stop source and lowering emitters

### debug.ts:525:37

Refused: `a cast the runtime can't check`. internal/lower/cast_proof.go:195 (castRefusal).

```ts
                // if we're in es2015, attach the method to a shared prototype for `FlowNode`
                // so the method doesn't show up in the watch window.
                if (!flowNodeProto) {
                    flowNodeProto = Object.create(Object.prototype) as FlowNode;
                    attachFlowNodeDebugInfoWorker(flowNodeProto);
                }
                Object.setPrototypeOf(flowNode, flowNodeProto);
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:267 and :318 emit this NotYet reason.

```ts
    {
        name: "watchFile",
        type: new Map(Object.entries({
            fixedpollinginterval: WatchFileKind.FixedPollingInterval,
            prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
            dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
            fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

### diagnosticInformationMap.generated.ts:13:45

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:267 and :318 emit this NotYet reason.

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### semver.ts:71:9

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. internal/lower/namespaces.go:269 and :321 emit this NotYet reason.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

## Root-order evidence

Classification: incomplete. The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.

The baseline report.py criteria require attempted-source and readiness-finding evidence from the entry run plus a Node throw. An incomplete entry ledger cannot establish an artifact. Independent stops remain counted.

| Root | Independent stop | Node exit | First error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Latent totals

Guarded entry: LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, bound 1800 seconds. Child exit -15; elapsed 1800.267 seconds; stopping reason elapsed bound. Emitted findings: NotYet 0, Refused 0, panics 0. Completed sources 0/81; stream records 1; checker diagnostics 0.

Did not finish inside its bound. Totals cover emitted findings only. Unemitted findings are unknown, so these are not complete-program latent totals. No per-unit timing ledger was emitted.

Ordinary census, 79 counted observations: NotYet 3, Refused 74, panics 0, accepted 2. These are first-stop totals. The ordinary stream carries no latent counters, so ordinary full latent totals are unavailable.

## Method, checkpoints and cleanup

Read the b322b6ba method and scripts first and reused its recipe at 7910ab94. One adapted tree, an ordinary binary built from that SHA with no overlay, sorted roots in sequential groups of 20, 20, 20 and 19. Each root emits two observations (158 raw); only the first counts (79 in census.jsonl). run.py, analyze.py, entry-run.py, latent-totals.py, classify-artifacts.py, source-excerpts.py, validate.py and package-evidence.py reuse the prior scripts with the SHA and destination changed. measure.py runs the steps in the foreground. Its unintended 1800-second outer group timeout was paused while the group child continued, then resumed after group completion to collect the child exit. This repeats the baseline supervision workaround. supervise-groups.py and groups-supervision.json record the intervention; root order and per-root limits were unchanged.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The source checkout HEAD is the pin above. apply-status.json and build-status.json identify the single apply and ordinary binary build.

Guarded entry uses this SHA's stage3/census/latent/make_overlay.py and stage3/meter/entry_overlay.py, then go build with the resulting overlay. LATENT_ASSERT_NO_OUTPUT guards against usable IR output. All 82 adapted compiler-directory files match the baseline hashes (adapted-source-comparison.json). Binary hashes, Go build metadata, adapted-source hashes and child status files identify measured bytes.

A plain push to typescript/next-stops-7910ab94-checkpoint followed apply immediately, then each root group. checkpoint-push-*-status.json records each child exit. The final branch has one report-only commit directly on 7910ab94. No force-push and no push to main.

Every census child process group was killed and reaped, including after normal exits. validation.json records observation counts, duplicate agreement, source cleanliness and no surviving census processes. Files over 1 MB are gzip-compressed.
