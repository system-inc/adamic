# Next stops on main bcbe5d12

Task b5w3ycg. Measured, report only, not for landing.

Compiler and apply SHA: `bcbe5d1244ee8a218d860415dea19cca114bae46`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `2ee68327757d`; 5 visible CPUs; cgroup CPU quota `400000 100000` (4 CPU equivalents); MemTotal:       18440100 kB. Managed environment: `ccarenv_b64_Y2NhcmVudl82MWJjNmEwMjE3Yzg4MTkxODEyY2I3ODNkM2JiMGQ2ZQ`.

One completed full apply produced `/workspace/bcbe5d12-adapted`. All 79 ordinary independent compiler roots use this tree and the ordinary census tool built from bcbe5d12. The compiler and adapter input diff is clean; hashes are retained in the source and binary manifests. The ordinary census uses no overlay.

## Grouped first stops

Roots are relative to `src/compiler/`. Counts include every root, including readiness candidates. Groups are sorted largest first.

| Roots | Location | Kind | Exact reason | Forecast match | Root names |
| ---: | --- | --- | --- | --- | --- |
| 73 | `core.ts:2103:26` | Refused | a method read as a value (compare would lose its object, and this with it) | core.ts:2116 .compare value read (shifted coordinates) | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | not applicable, lowers whole | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | none, new | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `debug.ts:355:26` | Refused | a method read as a value (toString would lose its object, and this with it) | none, new | `core.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | none, new | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | none, new | `debug.ts` (artifact classification incomplete) |

## Whole-root lowering

2/79 roots lower whole in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.
This is ordinary Lower acceptance of each entire resolved root program. No native output is measured. Readiness candidates receive no additional lowering credit without the meter evidence.

## Comparison with af42d3dc

Baseline compiler/apply: `af42d3dc1261a3db536064c3dfbb5eaa7e8b6405`. The prior report and grouped JSON are retained beside this report. `comparison.json` records both exact tuples and the forecast for every root.

| Current group | Roots | Unchanged exact tuple | Moved exact tuple | New tuple in this table |
| --- | ---: | ---: | ---: | --- |
| `core.ts:2103:26` Refused | 73 | 0 | 73 | yes |
| `-` accepted | 2 | 2 | 0 | no |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | no |
| `debug.ts:355:26` Refused | 1 | 0 | 1 | yes |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | no |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | no |

The collator operation moved from `core.ts:2116:26` to `core.ts:2103:26`; its kind and exact reason are unchanged. A moved coordinate alone is not a newly exposed compiler limitation. The `core.ts` root moved from the old generic-instantiation stop at `debug.ts:243:59` to the new method-value stop at `debug.ts:355:26`. The accepted group and the three readiness groups are unchanged.

## The forecasts

The forecast column applies to every root in its row; `comparison.json` gives one record per root. Matching uses the source operation when adaptations shift coordinates. Under the requested rule, every stopped group matching none of the six forecasts is called new below, even if its exact tuple was already present in the prior table. Accepted roots have no stop and no forecast match.

| Forecast | Matching roots | Measured location |
| --- | ---: | --- |
| `core.ts:2116` | 73 | `core.ts:2103:26` |
| `debug.ts:537` | 0 | not reached as a first stop |
| `debug.ts:577` | 0 | not reached as a first stop |
| `semver.ts:287` | 0 | not reached as a first stop |
| `performanceCore.ts:29` | 0 | not reached as a first stop |
| `performance.ts:189` | 0 | not reached as a first stop |

## Every group outside the forecasts: source and decision

### commandLineParser.ts:293:35

Exact reason: reading an enum before its runtime initialization; move the call after the enum declaration.

Source: `commandLineParser.ts-293-35.txt`, lines around the stop. The text file preserves the source bytes; line endings are normalized only in this display.

```ts
    {
        name: "watchFile",
        type: new Map(Object.entries({
            fixedpollinginterval: WatchFileKind.FixedPollingInterval,
            prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
            dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
            fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

Emitter: `internal/lower/namespaces.go:318` in `(*lowering).namespaceInitialization` (function begins at line 260). Caller: `internal/lower/lower.go:104`, `lowerWithElementKey`. The visit closure sees a namespace value node whose namespaceRuntimeEnum declaration is not in initialized and emits line 318. The reached-call alternative emits the same text at line 267. lowerWithElementKey invokes the initialization walk before refusal/body lowering.

### debug.ts:355:26

Exact reason: a method read as a value (toString would lose its object, and this with it).

Source: `debug.ts-355-26.txt`, whole enclosing function. The text file preserves the source bytes; line endings are normalized only in this display.

```ts
    export function getFunctionName(func: AnyFunction): string {
        if (typeof func !== "function") {
            return "";
        }
        else if (hasProperty(func, "name")) {
            return func.name;
        }
        else {
            const text = Function.prototype.toString.call(func);
            const match = /^function\s+([\w$]+)\s*\(/.exec(text);
            return match ? match[1]! : "";
        }
    }
```

Emitter: `internal/lower/refusals.go:285` in `(*lowering).methodValueRefusal` (function begins at line 260). Caller: `internal/lower/refusals.go:105`, `(*lowering).refuse`. The direct-bind exemption is at line 261. The guard at line 264 selects a property-access read outside a direct call without an allowed method-read proof. The method symbol at line 272 selects the refusal emitted at line 285. The refusal registry registers this rule at ordinal 2900, line 291. The visit closure in refuse (function begins at line 35) invokes rule.check at line 105, returning the first rule error. The outer call is Function.prototype.toString.call(func); the toString property is the receiver of .call rather than the direct call expression, so called(node) does not exempt it.

### diagnosticInformationMap.generated.ts:13:45

Exact reason: reading an enum before its runtime initialization; move the call after the enum declaration.

Source: `diagnosticInformationMap.generated.ts-13-45.txt`, lines around the stop. The text file preserves the source bytes; line endings are normalized only in this display.

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

Emitter: `internal/lower/namespaces.go:318` in `(*lowering).namespaceInitialization` (function begins at line 260). Caller: `internal/lower/lower.go:104`, `lowerWithElementKey`. The visit closure sees a namespace value node whose namespaceRuntimeEnum declaration is not in initialized and emits line 318. The reached-call alternative emits the same text at line 267. lowerWithElementKey invokes the initialization walk before refusal/body lowering.

### semver.ts:71:9

Exact reason: a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration.

Source: `semver.ts-71-9.txt`, whole enclosing function. The text file preserves the source bytes; line endings are normalized only in this display.

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

Emitter: `internal/lower/namespaces.go:269` in `(*lowering).namespaceInitialization` (function begins at line 260). Caller: `internal/lower/lower.go:104`, `lowerWithElementKey`. The checkCall closure follows graph.reach(target). A reached namespace declaration absent from initialized emits line 269 at the reached read. The direct-read alternative emits the same text at line 321. lowerWithElementKey invokes the initialization walk before refusal/body lowering.

## Root-order artifacts

The classification follows `stage3/meter/report.py`: Node throws alone do not establish eligibility. The stop source must have attempted entry units and no entry readiness finding there. Artifact roots remain explicit in the first-stop table.

STOPPING REASON: The bounded entry census did not complete. report.py requires its attempted-source and finding ledger before classifying root-order artifacts.

Entry child exit -15 after 1800.979 seconds: configured elapsed-time bound. The ordinary 79-root census completed. The entry stream has 1 record(s). Last completed timing: `/workspace/bcbe5d12-adapted/src/compiler/factory/utilities.ts:786:1`. The interrupted active unit was not recorded.

Artifact classification is incomplete for the following roots; none is folded away or credited as lowering whole.

| Root | Independent stop | Node child exit | Node first error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Method and evidence

The roots are sorted and run sequentially in groups of 20, 20, 20 and 19. Each root invocation emits an independent observation and a single-root aggregate. The runner verifies all non-timing fields agree, retains both raw observations, and counts the first once. Non-source rows and a multi-root aggregate do not enter the denominator.

The separate entry probe follows the prior report: guarded meter overlay, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and a 1800-second elapsed bound. The supervisor also stops a child after 1800 seconds without output, records its actual exit code, kills its process group, and reaps it. Scratch Go sources use .go.txt. Node probes use the meter helper unchanged.

Preparation failures before the dependency checkout completed are retained in the raw archive. Their completed replacement builds exited 0. The complete apply and every completed root group were pushed to `typescript/next-stops-bcbe5d12-checkpoint`. The final requested branch contains one report commit on the pinned main; no force push or main push is used.

The first verification checked string ordering rather than the reference runner's Path ordering. That order-only assertion was corrected; the same 79 roots and raw streams passed. The archived initial verification failure is retained, and validation.json records the final successful checks.

Raw streams, progress, timing, setup, apply, probe and build logs, plus the guarded overlay, are in `raw-streams.tar.gz`. Verify with `sha256sum -c raw-streams.tar.gz.sha256`. Direct JSON files retain per-root outcomes, comparisons, provenance, cleanup, and validation. Lowering source evidence is retained as .go.txt.

## Reproduction

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
# The source checkout must be at 050880ce59e30b356b686bd3144efe24f875ebc8.
python3 stage3/census/next-stops-bcbe5d12/supervise.py apply bash stage3/apply.sh --build-product /workspace/bcbe5d12-adapted > apply-runner.log 2>&1
go build -o /workspace/scratch/b5w3ycg/census ./stage3/census/tool > census-build.log 2>&1
python3 stage3/census/next-stops-bcbe5d12/run-groups.py > groups-runner.log 2>&1
```
