# Next stops on 37f8314b

73/73 roots moved off core.ts:966:72 and now stop at `core.ts:1049:34`, Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold.

Compiler and apply SHA: `37f8314b98577c8856a382f8641e7d8a2ca091e7`. TypeScript 6.0.3 pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Box: `1dcf084093ea`; 5 visible CPUs; cgroup cpu.max `400000 100000`; MemTotal:       18440100 kB.

## Grouped first stops

Roots are relative to `src/compiler/`. Every root is retained, including root-order artifact candidates.

| Roots | Location | Kind | Exact reason | Root names | Compared with 7e3c0682 |
| ---: | --- | --- | --- | --- | --- |
| 73 | `core.ts:1049:34` | Refused | a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations.ts`, `transformers/declarations/diagnostics.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` | new group; roots moved from core.ts:966:72 |
| 2 | `-` | accepted | - | `corePublic.ts`, `hostErrors.ts` | unchanged |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) | unchanged |
| 1 | `debug.ts:243:59` | NotYet | a generic instantiation whose type argument for T, constrained to Node, can't be read | `core.ts` | unchanged |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `types.ts` (artifact classification incomplete) | unchanged |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | `debug.ts` (artifact classification incomplete) | unchanged |

## Whole-root lowering

2/79 roots lower whole in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.
This credits ordinary Lower acceptance of the entire resolved root program. It does not claim native emission or acceptance of a multi-root aggregate.

## Comparison with 7e3c0682

| Previous group | Roots | Current result |
| --- | ---: | --- |
| `core.ts:966:72`, Refused: a value of type T[] seen as {}[], which can write {} where T is read | 73 | moved to `core.ts:1049:34`, Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold; this is a new first-stop group |
| `debug.ts:243:59`, NotYet: a generic instantiation whose type argument for T, constrained to Node, can't be read | 1 | unchanged |
| `-`, accepted: - | 2 | unchanged |
| `semver.ts:71:9`, NotYet: a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 | unchanged |
| `commandLineParser.ts:293:35`, NotYet: reading an enum before its runtime initialization; move the call after the enum declaration | 1 | unchanged |
| `diagnosticInformationMap.generated.ts:13:45`, NotYet: reading an enum before its runtime initialization; move the call after the enum declaration | 1 | unchanged |

`comparison.json` compares every exact root name against the prior report. The old core.ts:966:72 group is gone as a first stop.

## New group source and lowering decision

Exact adapted source at `src/compiler/core.ts:1048-1050`; this is the entire enclosing function, preserved byte-for-byte in `core.ts-1048-1050.txt`:

```ts
export function toSorted<T>(array: readonly T[], comparer?: Comparer<T>): SortedReadonlyArray<T> {
    return (array.length === 0 ? emptyArray : array.slice().sort(comparer)) as unknown as SortedReadonlyArray<T>;
}
```

Line 1049, column 34 is `emptyArray`. The observed exact reason is printed in the grouped table. At compiler SHA 37f8314b, `internal/lower/invariance.go:737` assembles the type-parameter reason in `wideningRefusal`; line 740 returns `Refused`. The refusal caller is `refuseWidening`, starting at line 449: its ordinary value branch obtains the slot or implied target at lines 563-576, calls `freshOrWidened` at line 576, and passes a non-nil finding to `wideningRefusal` at line 587. `freshOrWidened` starts at line 746; `widened` starts at line 61 and tests writable type-parameter constraints at lines 109-115; container argument checking is in `widenedElements` and `widenedArguments`, lines 252-281, with the mutable element incompatibility decision at lines 273-274. These functions decide the widening relation; the formatter prints the target type parameter and its constraint. `reason-grep.txt` retains the grep hits and `invariance.go.txt` retains the exact lowering source. No compiler instrumentation was used to infer a more specific dynamic call path.

## Root-order artifacts

STOPPING REASON: Meter entry census did not complete within its 1800-second bound; attempted-source and entry-finding evidence is incomplete. The ordinary 79-root census is complete. The entry stream contains only the records actually emitted before termination. No artifact label is inferred from a Node throw alone.

- `debug.ts`: artifact classification incomplete; Node exit 1, `TypeError: Cannot read properties of undefined (reading 'assert')`. Its ordinary first-stop row remains in the table.
- `diagnosticInformationMap.generated.ts`: artifact classification incomplete; Node exit 1, `TypeError: Cannot read properties of undefined (reading 'AbstractKeyword')`. Its ordinary first-stop row remains in the table.
- `types.ts`: artifact classification incomplete; Node exit 1, `TypeError: Cannot read properties of undefined (reading 'Error')`. Its ordinary first-stop row remains in the table.

The official meter artifact total is undetermined. Only ordinary accepted roots receive whole-root lowering credit.

## Reproduction and evidence

Setup: `seconds=1263 exit=0`. Full apply: `seconds=1448 exit=0`. Product key: `a1db6a54c8f66c31165d594e63e7d9c0327f51fcd236abb560316161d8f15cd2`. Manifest SHA-256: `6f0b343f0ea0d62d4604f072e198f6fdf1eb12e7fb93a5b04f8d2ccb279a5d70`. The exact manifest is `adapted-product.manifest.json`.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
stage3/apply.sh --build-product /workspace/next-stops-37f8314b-adapted > apply.log 2>&1
go build -o /workspace/scratch/next-stops-37f8314b/census ./stage3/census/tool > build.log 2>&1
python3 stage3/census/next-stops-37f8314b/run-groups.py > groups.log 2>&1
python3 stage3/census/next-stops-37f8314b/build-entry.py > entry-overlay.log 2>&1
go build -overlay /workspace/scratch/next-stops-37f8314b/entry-overlay/overlay.json -o /workspace/scratch/next-stops-37f8314b/entry-census ./stage3/census/latent/tool > entry-build.log 2>&1
python3 stage3/census/next-stops-37f8314b/run-entry.py > entry-runner.log 2>&1
```

The missing upstream source was cloned at v6.0.3 and verified against the exact pin. All 79 sorted compiler roots use one completed adapted tree and the unchanged ordinary census built from 37f8314b. No compiler change, refusal suppression, adaptation overlay, or partial-root fallback is used in the ordinary census.

Groups contain 20, 20, 20, and 19 roots. Each single-file CLI invocation emits an independent-root row and an identical single-root aggregate row; the runner verifies their non-timing fields agree, retains both in `group-N.jsonl`, and counts the first only in `group-N-rows.json`. `census.jsonl` concatenates the four raw streams. No root is counted twice. No non-TypeScript input or 79-root aggregate is mixed into the table.

The separate artifact probe uses the meter standard entry overlay. All physical scratch Go sources are named `.go.txt`. `binaries.sha256` and `census-buildinfo.txt` identify the binaries. The entry measurement has an 1800-second bound with process-group termination and reaping.

The completed-apply checkpoint and every completed group were committed and pushed promptly. Required checkpoint history is preserved; every commit changes only this report directory. The environment context interruption killed the original supervisors. The completed nineteenth root was recovered from its two intact raw rows and the orphaned entry probe was terminated; interruption.json records this. The ordinary runner resumed at root 20, and the bounded entry probe restarted. All census processes exited or were terminated before completion. No running census process remains. The report branch was pushed plainly, without pushing main or force-pushing.
