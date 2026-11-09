# Next stops on main c3b22184

Adamic, #baobab. Measured census, report only, not for landing.

Compiler/apply SHA: `c3b22184745d809cb281a37ef1989c895596223f`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `c820e8078203`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB.

All 79 compiler roots use one completed full apply at `/workspace/c3b22184-adapted` and the ordinary census tool built from the pinned main. Compiler and adapter source diffs are clean. Hash manifests identify the tool and adapted source bytes. The ordinary census uses no overlay.

## Files lowering whole

2/79 roots lower whole in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.
This is Lower acceptance of each entire resolved root program. No native output is measured. Artifact roots receive no additional credit without the meter evidence.

## Grouped first stops

Root names are relative to `src/compiler/`. All roots are retained, including readiness candidates. Groups are sorted largest first. The accepted row is retained for coverage.

| Roots | Stop | Kind | Construct | Exact reason | Compared with bcbe5d12 | Forecast | Root names |
| ---: | --- | --- | --- | --- | --- | --- | --- |
| 73 | `core.ts:2599:15` | NotYet | `ambient process host binding declaration: declare const process: { nextTick?: unknown; browser?: unknown; } \| undefined; declare const require: unknown;` | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | new | none, new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | `lowers whole` | - | unchanged | not applicable, lowers whole | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | `fixedpollinginterval: WatchFileKind.FixedPollingInterval,` | reading an enum before its runtime initialization; move the call after the enum declaration | unchanged | none, new | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `debug.ts:525:37` | Refused | `attachFlowNodeDebugInfo: Object.create(Object.prototype) as FlowNode` | a cast the runtime can't check | moved | debug.ts:537 cast | `core.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | `diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage` | reading an enum before its runtime initialization; move the call after the enum declaration | unchanged | none, new | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | `Version.constructor: Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | unchanged | none, new | `debug.ts` (artifact classification incomplete) |

## Comparison with bcbe5d12

Baseline compiler/apply: `bcbe5d1244ee8a218d860415dea19cca114bae46`, report commit `30057ecaa2d71ee183e20f32ad7f3e4def6b5bb2`. The exact prior report and grouped JSON are retained. `comparison.json` contains old and current location/kind/reason tuples, constructs, and forecasts for every root.

The comparison column says unchanged when every root retains its exact tuple, moved when roots change to another forecasted or already known stop, and new for a previously unseen tuple outside the forecasts. The separate forecast column labels every forecast miss new, including unchanged baseline groups.

| Current group | Roots | Exact tuple unchanged | Roots moved | New tuple in this table |
| --- | ---: | ---: | ---: | --- |
| `core.ts:2599:15` NotYet | 73 | 0 | 73 | yes |
| `-` accepted | 2 | 2 | 0 | no |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | no |
| `debug.ts:525:37` Refused | 1 | 0 | 1 | yes |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | no |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | no |

Where the prior 73 collator roots stop now:

| Current stop | Roots from the old collator group |
| --- | ---: |
| `core.ts:2599:15` NotYet | 73 |

## Forecasts from #b5w3ycg

Matching identifies the source construct when coordinates shift. `forecast-anchors.json` records the old source lines and verifies their files against the bcbe5d12 source manifest. Named overload walls match the enclosing declaration and the overload-result reason. Every stop matching none is new under the requested forecast rule.

| Forecast | Matching roots | Measured stop |
| --- | ---: | --- |
| debug.ts:537 cast | 1 | `debug.ts:525:37` |
| debug.ts:577 cast | 0 | not reached as a first stop |
| semver.ts:287 | 0 | not reached as a first stop |
| performanceCore.ts:29 | 0 | not reached as a first stop |
| performance.ts:189 | 0 | not reached as a first stop |
| tryGetPerformanceHooks performance.now presence test (Lower a3) | 0 | not reached as a first stop |
| findAncestor overload-result wall | 0 | not reached as a first stop |
| serializePropertySymbolsForClassOrInterface overload-result wall | 0 | not reached as a first stop |

## Every group outside the forecasts: source and lowering decision

### core.ts:2599:15

Construct: `ambient process host binding declaration: declare const process: { nextTick?: unknown; browser?: unknown; } | undefined; declare const require: unknown;`

Exact reason: `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,`.

Exact source excerpt: `core.ts-2599-15.txt`, lines around the stop. Display line endings are normalized; the text file retains the extracted bytes.

```ts
        return array.slice(index);
    }
}
declare const process: { nextTick?: unknown; browser?: unknown; } | undefined; declare const require: unknown;
/** @internal */
export function isNodeLikeSystem(): boolean {
    // This is defined here rather than in sys.ts to prevent a cycle from its
```

Emitter: `internal/lower/host_bindings.go:139`, `hostBindingStop` (function starts at line 124). Caller: `internal/lower/host_bindings.go:106`, `(*lowering).hostBindingStatement`. hostBindingStop walks row.Targets() at lines 126-129, collecting Pending columns and their owners. It builds the exact reason at line 139. hostBindingStatement (starts at line 96) matches the ambient declarator to the table row, calls hostBindingStop at line 106, and emits l.notYet(declaration, reason) at line 107. ambientRefusal (internal/lower/ambient.go:61) calls hostBindingStatement at line 76; the rule is registered at ordinal 2200, line 94. The embedded process row has native Pending (Runtime), javascript Provided, wasm32 Pending (Runtime and Platforms), and model node:process. The decision uses the compiler-embedded host-binding rows.

### semver.ts:71:9

Construct: `Version.constructor: Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");`

Exact reason: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`.

Exact source excerpt: `semver.ts-71-9.txt`, whole enclosing function. Display line endings are normalized; the text file retains the extracted bytes.

```ts
    constructor(major: number | string, minor = 0, patch = 0, prerelease: string | readonly string[] = "", build: string | readonly string[] = "") {
        if (typeof major === "string") {
            const result = Debug.checkDefined(tryParseComponents(major), "Invalid version");
            ({ major, minor, patch, prerelease, build } = result);
        }

        Debug.assert(major >= 0, "Invalid argument: major");
        Debug.assert(minor >= 0, "Invalid argument: minor");
        Debug.assert(patch >= 0, "Invalid argument: patch");

        const prereleaseArray = prerelease ? isArray(prerelease) ? prerelease : prerelease.split(".") : emptyArray;
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
        this.patch = patch;
        this.prerelease = prereleaseArray;
        this.build = buildArray;
    }
```

Emitter: `internal/lower/namespaces.go:269`, `(*lowering).namespaceInitialization` (function starts at line 260). Caller: `internal/lower/lower.go:113`, `lowerWithElementKey`. The checkCall closure follows graph.reach(target). A reached namespace declaration absent from initialized emits line 269 at the reached read. The direct-read alternative emits the same reason at line 321. lowerWithElementKey calls this walk before refusal and function-body lowering.

### commandLineParser.ts:293:35

Construct: `fixedpollinginterval: WatchFileKind.FixedPollingInterval,`

Exact reason: `reading an enum before its runtime initialization; move the call after the enum declaration`.

Exact source excerpt: `commandLineParser.ts-293-35.txt`, lines around the stop. Display line endings are normalized; the text file retains the extracted bytes.

```ts
    {
        name: "watchFile",
        type: new Map(Object.entries({
            fixedpollinginterval: WatchFileKind.FixedPollingInterval,
            prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
            dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
            fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

Emitter: `internal/lower/namespaces.go:318`, `(*lowering).namespaceInitialization` (function starts at line 260). Caller: `internal/lower/lower.go:113`, `lowerWithElementKey`. The visit closure finds a namespace value node whose namespaceRuntimeEnum declaration is not initialized and emits line 318. The reached-call alternative emits the same reason at line 267. lowerWithElementKey calls this walk before refusal and function-body lowering.

### diagnosticInformationMap.generated.ts:13:45

Construct: `diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage`

Exact reason: `reading an enum before its runtime initialization; move the call after the enum declaration`.

Exact source excerpt: `diagnosticInformationMap.generated.ts-13-45.txt`, lines around the stop. Display line endings are normalized; the text file retains the extracted bytes.

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

Emitter: `internal/lower/namespaces.go:318`, `(*lowering).namespaceInitialization` (function starts at line 260). Caller: `internal/lower/lower.go:113`, `lowerWithElementKey`. The visit closure finds a namespace value node whose namespaceRuntimeEnum declaration is not initialized and emits line 318. The reached-call alternative emits the same reason at line 267. lowerWithElementKey calls this walk before refusal and function-body lowering.

## Root-order artifacts

Classification follows the current main revision of `stage3/meter/report.py`. Node throws alone are insufficient: the stop source must have attempted entry units and no entry readiness finding there. Candidate roots stay visible in the grouped table.

STOPPING REASON: The bounded entry census did not complete. report.py requires its attempted-source and finding ledger before classifying root-order artifacts.

Entry child exit -15 after 1801.008 seconds: configured elapsed-time bound. The ordinary 79-root census completed. Entry stream records: 1. Last completed timing: `/workspace/c3b22184-adapted/src/compiler/factory/nodeConverters.ts:183:1`. The interrupted active unit was not recorded.

| Root | Independent stop | Node exit | Node first error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Method and evidence

The same recipe as next-stops-bcbe5d12: source the tools environment, run cloud/setup.sh, npm ci --prefix stage3/api, and export ADAMIC_TYPESCRIPT_SOURCE to the pinned pristine source. Complete a fresh full apply. Run the ordinary per-file census sequentially in groups of 20, 20, 20 and 19, using the same Path sorting as the reference runner. Each invocation emits two single-root observations; all non-timing fields must agree, and only the first counts.

The separate entry probe uses the meter guarded overlay, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and the prior 1800-second elapsed bound. A supervisor also stops any child after 1800 seconds without output, records the actual child exit code, kills its process group, and reaps it. Physical scratch Go files use .go.txt. Node readiness probes use the meter helper.

The completed apply and each completed root group were pushed to `typescript/next-stops-c3b22184-checkpoint`. The requested branch contains one report commit directly on c3b22184. No force push or main push is used.

Raw streams, timing, progress, setup, apply, tool-build, probe logs, and the guarded overlay are packed in `raw-streams.tar.gz`, with `raw-streams.tar.gz.sha256` and a member hash manifest. Verify with `sha256sum -c raw-streams.tar.gz.sha256`. Direct JSON retains the complete root ledger, comparison, forecasts, provenance, process cleanup, and final verification.

Reproduction commands and actual invocations are in prepare.py, measure.py, run-groups.py, prepare-entry.py, and run-entry.py. All log destinations and child exit codes are retained.
