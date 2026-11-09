# Next stops on main c3b7baf0

Measured for #b5w3ycg and #baobab. Report only.

Compiler and apply SHA: `c3b7baf0b3438bf5ac7aa9fcf6eb2ba385e2f9bd`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `bb0f9cc0be97`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-09T07:59:54Z.

One completed apply at `/workspace/scratch/next-stops-adapted`, exit 0. Ordinary census binary built from the same main SHA, with no overlay. All 82 compiler-directory files match the c3b22184 adapted-source hashes. `binaries.json` and the source manifests identify the measured bytes.

## Grouped first stops

Root names and stop locations are relative to `src/compiler/`. Groups are sorted largest first. Every root appears once; repeated single-root observations agree in every non-timing field.

| Roots | Location | Kind | Exact reason | Compared with c3b22184 | Root names |
| ---: | --- | --- | --- | --- | --- |
| 73 | `core.ts:2599:15` | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | unchanged | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | unchanged | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `debug.ts:525:37` | Refused | a cast the runtime can't check | unchanged | `core.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | unchanged | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | unchanged | `debug.ts` (artifact classification incomplete) |

## Roots lowering whole

2/79 in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.

This is Lower acceptance of each complete resolved root program. No native output is measured. No additional acceptance credit is assigned to readiness roots.

## Comparison with c3b22184

Baseline compiler/apply: `c3b22184745d809cb281a37ef1989c895596223f`, from `origin/typescript/next-stops-c3b22184`. Its report and grouped data are retained as `previous-REPORT.md` and `previous-grouped-first-stops.json`. `comparison.json` records before and after tuples for all 79 roots.

Unchanged means the same location, kind and reason for that root. Moved means a changed root tuple that already appeared in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged roots | Moved roots | Classification |
| --- | ---: | ---: | ---: | --- |
| `core.ts:2599:15` NotYet | 73 | 73 | 0 | unchanged |
| `-` accepted | 2 | 2 | 0 | unchanged |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | unchanged |
| `debug.ts:525:37` Refused | 1 | 1 | 0 | unchanged |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | unchanged |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | unchanged |

New groups: 0. Source excerpts for every stop are retained in `source-excerpts.json` and the adjacent `.txt` files.

## hostErrors.ts

**Yes, hostErrors.ts lowers whole under the loader's options.** Its counted group-2 row is accepted, and both observations agree. A separate focused production-census invocation also emitted two accepted observations (`hostErrors-probe.jsonl`).

`stage3/census/tool/main.go` calls `load.Load([]string{root})`, then `lower.Lower`. `internal/load/load.go:59` defines compilerOptions directly: strict checking, noUncheckedIndexedAccess, exactOptionalPropertyTypes, verbatimModuleSyntax, forced modules, ESNext/bundler, ES2024 with es2024 libraries and no ambient package types. It does not parse the compiler tsconfig. The helper has no imports, so this independent root does not load the compiler-wide module graph.

There is no measured stop to attribute or bisect in the requested interval. The quoted alternative, `in when a possibly absent spread has no property presence descriptors`, is constructed by `internal/lower/unknown.go:314` and emitted with the `in when ` prefix at `internal/lower/unknown.go:56`. `presenceHazard` scans the root's module order, so a different program graph can expose spread hazards that this helper root does not contain. That explains why the tsconfig-based observation cannot replace this measurement; no reproduction of that separate run is claimed.

Measured source around line 3:

```ts
/** @internal */
export function errorCode(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string" ? error.code : undefined;
}

/** @internal */
export function errorMessage(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "message" in error && typeof error.message === "string" ? error.message : undefined;
}
```

## Root-order classification

The current `stage3/meter/report.py:271` requires the readiness stop source to have attempted entry units and no entry readiness finding at that source; Node throws then determine the artifact label. The ordinary first stops remain visible above. The same current-main Node helper and its mirrored loader options were used; full results and child exit codes are retained under `node-readiness/`.

**Stopping reason:** The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.

Entry child exit `-15` after 1800.321 seconds; reason: elapsed bound. Its process group was killed and reaped. Entry stream: 1 records, 0/81 completed source records. These roots retain the explicit incomplete label.

The baseline report also marked these three roots artifact classification incomplete after its bounded entry probe. This report preserves that evidence limitation.

| Root | Independent stop | Node exit | First error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Stop source and lowering emitters

No new first-stop group was observed. Existing reasons map to:

| Stop | Lowering emitter at measured SHA |
| --- | --- |
| `core.ts:2599:15` | `internal/lower/host_bindings.go:107`, ambient declarator stop; reason assembled by `hostBindingStop` at line 139 |
| `debug.ts:525:37` | `internal/lower/cast_proof.go:195`, `castRefusal` |
| enum readiness stops | `internal/lower/namespaces.go:267` and `:318`, `namespaceInitialization` |
| `semver.ts:71:9` | `internal/lower/namespaces.go:269` and `:321`, `namespaceInitialization` |

### core.ts:2599:15

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

### commandLineParser.ts:293:35

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

### debug.ts:525:37

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

### diagnosticInformationMap.generated.ts:13:45

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

### semver.ts:71:9

Source lines 68 through 74:

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

## Reproduction, checkpoints and cleanup

Setup: `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, and `export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source`, whose HEAD is the pin above. Setup and dependency commands exited 0. The setup final record names the measured main SHA.

Apply: `bash stage3/apply.sh /workspace/scratch/next-stops-adapted`. Build: `go build -o /workspace/scratch/next-stops-census ./stage3/census/tool`. `run.py`, `pipeline.py`, and the command/status ledgers retain actual invocations.

The ordinary per-file recipe follows c3b22184: sorted roots, sequential groups of 20, 20, 20 and 19. Each child receives a single file path and emits the same observation twice, once through the per-file loop and once through its single-root whole-program call. Only the first counts. `group-*.jsonl` retains both raw observations; `census.jsonl` contains the 79 counted observations.

The separate guarded entry binary uses the current meter overlay with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, bounded to 1800 seconds. It cannot emit usable IR. Its output is used only for artifact evidence. The ordinary census uses no overlay or latent measurement options.

Every census child exit code and process-group cleanup is recorded. `validation.json` verifies coverage, agreement, source cleanliness and no surviving measurement processes. Logs are files; anything over 1 MB is gzip-compressed. The raw streams remain adjacent to this report.

The apply and each root group were pushed without force to `typescript/next-stops-c3b7baf0-checkpoint`. `checkpoint-*-status.json` records push outcomes. The final branch `typescript/next-stops-c3b7baf0` contains one report-only commit directly on the measured main SHA. Main was never pushed.
