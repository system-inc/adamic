# Next stops on 7e3c0682

73/73 roots moved off core.ts:810:89. The new biggest group is 73 roots at `core.ts:966:72`, Refused: a value of type T[] seen as {}[], which can write {} where T is read.

Compiler and apply source: `7e3c0682da96cbe01f7ae9b9820f1bd8f1993671`. TypeScript 6.0.3 pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Box: `ccb125c8c87a`; 5 visible CPUs; cgroup quota 400000/100000 microseconds; memory 18440100 kB. Load start: 3.24 8.38 4.88 1/148 10188; load end: 1.78 2.87 3.19 1/151 11769.

## Grouped first stops

All root names below are relative to `src/compiler/`. Root-order artifacts retain their ordinary first-stop rows; they are never folded into accepted or omitted.

| Roots | Location | Kind | Exact reason | Root names | Comparison with 99w |
| ---: | --- | --- | --- | --- | --- |
| 73 | `core.ts:966:72` | Refused | a value of type T[] seen as {}[], which can write {} where T is read | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations.ts`, `transformers/declarations/diagnostics.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` | new stop; roots moved from 99w |
| 2 | `-` | accepted | - | `corePublic.ts`, `hostErrors.ts` | unchanged |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) | unchanged |
| 1 | `debug.ts:243:59` | NotYet | a generic instantiation whose type argument for T, constrained to Node, can't be read | `core.ts` | unchanged |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `types.ts` (artifact classification incomplete) | unchanged |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | `debug.ts` (artifact classification incomplete) | unchanged |

## Whole-root lowering

2/79 roots lower whole in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.
This means ordinary Lower accepted each entire root program. It does not claim native emission or that the aggregate 79-root program lowers.

## Comparison with 99w

| 99w group | Roots | Status now |
| --- | ---: | --- |
| `core.ts:810:89`, Refused: a cast the runtime can't check | 73 | moved to `core.ts:966:72`, Refused: a value of type T[] seen as {}[], which can write {} where T is read |
| `debug.ts:243:59`, NotYet: a generic instantiation whose type argument for T, constrained to Node, can't be read | 1 | unchanged |
| `-`, accepted: - | 2 | unchanged |
| `semver.ts:71:9`, NotYet: a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 | unchanged |
| `commandLineParser.ts:293:35`, NotYet: reading an enum before its runtime initialization; move the call after the enum declaration | 1 | unchanged |
| `diagnosticInformationMap.generated.ts:13:45`, NotYet: reading an enum before its runtime initialization; move the call after the enum declaration | 1 | unchanged |

The old core.ts:810 group is gone as a first stop. The core.ts:966:72 group is new. The two accepted roots and the four singleton stop groups are compared above using every exact root path from 99w's census-results.json, including duplicate basenames in different subdirectories.

## Root-order artifacts

STOPPING REASON: The meter entry census was terminated at its 1800-second bound after emitting only its header. The 79-root ordinary census is complete. Full meter artifact classification is incomplete because attempted-source and entry-finding evidence was not emitted. No artifact is inferred from Node throws alone. The three Node-throwing readiness roots remain separate ordinary first-stop rows, with incomplete labels. `root-order.json` retains the gap and Node results.

- `debug.ts`: artifact classification incomplete; Node exit 1, `TypeError: Cannot read properties of undefined (reading 'assert')`. No whole-root lowering credit is assigned.
- `diagnosticInformationMap.generated.ts`: artifact classification incomplete; Node exit 1, `TypeError: Cannot read properties of undefined (reading 'AbstractKeyword')`. No whole-root lowering credit is assigned.
- `types.ts`: artifact classification incomplete; Node exit 1, `TypeError: Cannot read properties of undefined (reading 'Error')`. No whole-root lowering credit is assigned.

The official meter artifact total is undetermined. Only the two ordinary accepted roots are credited as whole-root lowerings. Artifact roots cannot gain whole-root credit without complete entry-program evidence.

## New biggest stop

Exact adapted source, `src/compiler/core.ts:958-971` (the retained `core.ts-958-971.txt` preserves source line endings):

```ts
/** @internal */
export function combine<T extends {}>(xs: T | readonly T[] | undefined, ys: T | readonly T[] | undefined): T | readonly T[] | undefined;
/** @internal */
export function combine<T extends {}>(xs: T | T[] | undefined, ys: T | T[] | undefined): T | T[] | undefined;
/** @internal */
export function combine<T extends {}>(xs: T | T[] | undefined, ys: T | T[] | undefined) {
    if (xs === undefined) return ys;
    if (ys === undefined) return xs;
    if (isArray(xs)) return isArray(ys) ? concatenate(xs, ys) : append(xs, ys);
    if (isArray(ys)) return append(ys, xs);
    return [xs, ys];
}

/**
```

The stop is line 966, column 72: the first argument `xs` in `append(xs, ys)`. The exact reason is assembled by `wideningRefusal` in `internal/lower/invariance.go:715`, with `Refused` returned at line 729. `reason-grep.txt` retains the grep hits and `invariance.go.txt` retains lines 699-730. These are the compiler files at 7e3c0682.

## Reproduction and evidence

Setup took 498 seconds; apply took 518 seconds on this box. Product key: `f6c20f45fa0538d699f7d9d09943825f1afe19c92234f11ba161fd91792db801`. Manifest SHA-256: `94e776fcf1aca8fa63b776d0aab0323ff40ae17dbfe27ddf89a3cc15e33abd12`. `adapted-product.manifest.json` retains the exact manifest. `binaries.sha256` identifies the unchanged ordinary census and the separate meter entry-census binaries.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/next-stops/setup.log 2>&1
npm ci --prefix stage3/api > /workspace/scratch/next-stops/npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
stage3/apply.sh --build-product /workspace/next-stops-adapted > /workspace/scratch/next-stops/apply.log 2>&1
go build -o /workspace/scratch/next-stops/census ./stage3/census/tool > /workspace/scratch/next-stops/build.log 2>&1
python3 stage3/census/next-stops-7e3c0682/run-groups.py > /workspace/scratch/next-stops/groups.log 2>&1
```

The missing upstream source checkout was cloned at tag v6.0.3 and its HEAD verified against the exact pin. The 79 sorted TypeScript roots use one completed adapted tree. Groups contain 20, 20, 20, and 19 roots. The unchanged census CLI accepts one file as its source argument and emits a per-file row plus a single-root aggregate row. Both raw rows are retained in each group JSONL; the runner checks their non-timing fields are identical and counts only the first. Thus no root is counted twice. No non-TypeScript inputs or multi-root aggregate are mixed into the 79-root table.

The ordinary census has no overlay, suppression, compiler edits, adaptation edits, or partial-root fallback. The separate meter entry census uses the meter's standard measurement overlay for artifact classification only. `build-entry.py` reproduces it with physical scratch Go filenames ending in .go.txt; virtual Go paths and overlay contents are unchanged. Entry raw evidence is retained compressed beside this report, along with Node probes and source excerpts.

Each completed group was committed and pushed immediately. All census invocations exited or were reaped before completion; no census process remains. The branch is report only, with required checkpoint history preserved and no push to main.
