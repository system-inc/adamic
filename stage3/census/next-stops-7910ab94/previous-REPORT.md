# Next stops on main b322b6ba

Measured for #baobab and #b5w3ycg. Report only.

Compiler and apply SHA: `b322b6ba2f12454443bbbbdaaa6844fdce22a440`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `55fe9204d368`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-09T18:51:55.574230+00:00.

All 73 roots that previously stopped at the `process` binding now stop at the adjacent `require` declaration, `core.ts:2599:80`, with `Refused: an ambient value declaration`. The other six roots are unchanged.

## Grouped first stops

Locations and root names are relative to `src/compiler/`. Sorted largest first. Only the first of each root's two observations counts. All duplicate observations agree in every non-timing field.

| Roots | Location | Kind | Exact reason | Compared with c3b7baf0 | Root names |
| ---: | --- | --- | --- | --- | --- |
| 73 | `core.ts:2599:80` | Refused | an ambient value declaration | new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | unchanged | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `debug.ts:525:37` | Refused | a cast the runtime can't check | unchanged | `core.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | unchanged | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | unchanged | `debug.ts` (artifact classification incomplete) |

## Roots lowering whole

2/79 in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.

This measures Lower acceptance of each complete resolved root program. Native output is not measured. Readiness roots receive no additional acceptance credit.

## Comparison with c3b7baf0

Baseline compiler/apply SHA: `c3b7baf0b3438bf5ac7aa9fcf6eb2ba385e2f9bd`, from `typescript/next-stops-c3b7baf0`. Its report and grouped table are retained in `previous-REPORT.md` and `previous-grouped-first-stops.json`.

Unchanged means the same location, kind and exact reason for that root. Moved means a different tuple already present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged roots | Moved roots | New roots |
| --- | ---: | ---: | ---: | ---: |
| `core.ts:2599:80` Refused | 73 | 0 | 0 | 73 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `debug.ts:525:37` Refused | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |

Across 79 roots: 6 unchanged, 0 moved to an existing tuple, 73 at a new tuple.

Each root, with its before and after tuple retained in `comparison.json`:

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `_namespaces/ts.performance.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `_namespaces/ts.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `binder.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `builder.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `builderPublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `builderState.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `builderStatePublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `checker.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `commandLineParser.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `core.ts` | unchanged | `debug.ts:525:37` Refused | `debug.ts:525:37` Refused |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `executeCommandLine.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `expressionToTypeNode.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/baseNodeFactory.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/emitHelpers.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/emitNode.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/nodeChildren.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/nodeConverters.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/nodeFactory.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/nodeTests.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/parenthesizerRules.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/utilities.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `factory/utilitiesPublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `moduleSpecifiers.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `parser.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `path.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `performance.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `performanceCore.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `program.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `programDiagnostics.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `resolutionCache.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `scanner.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `semver.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `sourcemap.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `symbolWalker.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `sys.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `tracing.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformer.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/classFields.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/classThis.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/declarations/diagnostics.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/declarations.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/destructuring.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2015.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2016.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2017.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2018.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2019.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2020.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/es2021.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/esDecorators.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/esnext.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/generators.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/jsx.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/legacyDecorators.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/module/esnextAnd2015.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/module/impliedNodeFormatDependent.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/module/module.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/module/system.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/namedEvaluation.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/taggedTemplate.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/ts.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/typeSerializer.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `transformers/utilities.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `tsbuild.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `tsbuildPublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `utilitiesPublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `visitorPublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `watch.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `watchPublic.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |
| `watchUtilities.ts` | new | `core.ts:2599:15` NotYet | `core.ts:2599:80` Refused |

## Stop source and lowering emitters

### core.ts:2599:80 (new)

Refused: `an ambient value declaration`. `internal/lower/ambient.go:79` assigns the exact reason in `ambientRefusal`; line 91 emits the `Refused` result.

Source lines 2596 through 2602:

```ts
        return array.slice(index);
    }
}
declare const process: { nextTick?: unknown; browser?: unknown; } | undefined; declare const require: unknown;
/** @internal */
export function isNodeLikeSystem(): boolean {
    // This is defined here rather than in sys.ts to prevent a cycle from its
```

### commandLineParser.ts:293:35 (unchanged)

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. `internal/lower/namespaces.go:267` and `:318`, `namespaceInitialization`.

Source lines 290 through 296:

```ts
    {
        name: "watchFile",
        type: new Map(Object.entries({
            fixedpollinginterval: WatchFileKind.FixedPollingInterval,
            prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
            dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
            fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

### debug.ts:525:37 (unchanged)

Refused: `a cast the runtime can't check`. `internal/lower/cast_proof.go:195`, `castRefusal`.

Source lines 522 through 528:

```ts
                // if we're in es2015, attach the method to a shared prototype for `FlowNode`
                // so the method doesn't show up in the watch window.
                if (!flowNodeProto) {
                    flowNodeProto = Object.create(Object.prototype) as FlowNode;
                    attachFlowNodeDebugInfoWorker(flowNodeProto);
                }
                Object.setPrototypeOf(flowNode, flowNodeProto);
```

### diagnosticInformationMap.generated.ts:13:45 (unchanged)

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. `internal/lower/namespaces.go:267` and `:318`, `namespaceInitialization`.

Source lines 10 through 16:

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### semver.ts:71:9 (unchanged)

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. `internal/lower/namespaces.go:269` and `:321`, `namespaceInitialization`.

Source lines 68 through 74:

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

## Root-order classification

The same `stage3/meter/report.py` criteria and `node_readiness.py` helper as the baseline are used. Classification requires entry attempted-source and readiness-finding evidence, plus a Node throw. The ordinary first stops remain counted above.

Entry classification: **incomplete**. The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.

Entry emitted 1 stream record(s) and completed 0/81 source records.

| Root | Independent stop | Node exit | First error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Latent totals

Guarded entry run, `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, bound 1800 seconds: child exit `-15`, elapsed 1800.272 seconds, stopping reason `elapsed bound`. Checker diagnostics: 0.

Emitted guarded-entry findings: **NotYet 0, Refused 0, panics 0**. Completed sources: 0/81.

The entry run did not finish inside its bound. These counts cover emitted records only. Unemitted findings are unknown; these are not complete-program latent totals. The stream carries no completed source ledger beyond the coverage stated above.

Ordinary census, summed over the 79 counted first observations: **NotYet 3, Refused 74, panics 0**, plus 2 accepted. The ordinary stream carries no latent totals. These are first-stop counts only; full latent totals for the ordinary census are unavailable. `latent-totals.py` and `latent-totals.json` retain the calculation and labels.

## Method, checkpoints and cleanup

The baseline REPORT.md method and scripts were read before execution. The measurement recipe is reused at b322b6ba: one apply tree, an ordinary binary from the same SHA with no overlay, sorted roots in sequential groups of 20, 20, 20 and 19, two raw observations per root, first observation counted. `group-*.jsonl` contains 158 observations; `census.jsonl` contains the 79 counted observations.

Setup: `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, and `export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source`, whose HEAD is the pin above. Setup, dependency installation, apply and both binary builds exited 0.

Apply: `bash stage3/apply.sh /workspace/scratch/next-stops-adapted`. Ordinary build: `go build -o /workspace/scratch/next-stops-census ./stage3/census/tool`. Guarded build uses the selected SHA's `stage3/census/latent/make_overlay.py` and `stage3/meter/entry_overlay.py`, then `go build -overlay=/workspace/scratch/next-stops-entry-overlay/overlay.json -o /workspace/scratch/next-stops-entry ./stage3/census/latent/tool`. The guarded binary cannot emit usable IR.

All 82 compiler-directory source files match the c3b7baf0 adapted-source hashes (`adapted-source-comparison.json`). `binaries.json`, build information, source manifests, scripts, command/status ledgers and raw streams identify the measured bytes and child exits. Each census process group was killed and reaped after the child exit. `validation.json` verifies 79 roots, 158 observations, duplicate agreement, source cleanliness and no surviving census processes. Anything over 1 MB is gzip-compressed.

The initial extra preparation checkpoint found the previous run's remote checkpoint and was rejected. The existing tip was fetched, then the fresh evidence was appended with a plain fast-forward push. The required apply checkpoint was pushed immediately after apply, followed by a checkpoint after every root group. `checkpoint-*-status.json` records the results.

An unintended outer preparation-plus-pipeline timeout wrapper was paused while its child continued in the original foreground session, then resumed after pipeline completion to collect the child's actual exit code. This did not change root order, root limits, binaries or observations; `pipeline-supervision.json` records the intervention.

The final branch `typescript/next-stops-b322b6ba` contains one report-only commit directly on b322b6ba. All changes are under this directory. No force-push was used and main was never pushed.
