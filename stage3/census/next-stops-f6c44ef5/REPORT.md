# Next stops at f6c44ef5

The largest group is 73 roots at `src/compiler/core.ts:810:89`, Refused: a cast the runtime can't check. The old core.ts:736:36 first stop no longer appears.

Box: `c6856ea9a1af`, 5 visible CPUs, cgroup quota 4 CPUs, 18440100 kB host memory. Compiler and apply SHA: `f6c44ef5f5d3908e134e2f098bf5c9cabdff424f`. TypeScript source pin: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. Checker submodule: `d92d9bfee114c80be2c375d72edae966176e3a4f`; cohere: `06af2e399c6c1fb93fd570b774027942747400b4`.

## Grouped first stops

Root names below are relative to `src/compiler/`. Each of the 79 compiler roots occurs exactly once, including root-order artifacts.

| Roots | Location | Kind | Reason | Root names |
| ---: | --- | --- | --- | --- |
| 73 | `src/compiler/core.ts:810:89` | Refused | a cast the runtime can't check | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations.ts`, `transformers/declarations/diagnostics.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | whole | accepted | - | `corePublic.ts`, `hostErrors.ts` |
| 1 | `src/compiler/commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `diagnosticInformationMap.generated.ts` (root-order artifact) |
| 1 | `src/compiler/debug.ts:243:59` | NotYet | a generic instantiation whose type argument for T, constrained to Node, can't be read | `core.ts` |
| 1 | `src/compiler/diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `types.ts` (root-order artifact) |
| 1 | `src/compiler/semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | `debug.ts` (root-order artifact) |

## Roots that lower whole

2 roots lower whole: `corePublic.ts`, `hostErrors.ts`. This is acceptance of each root and its resolved program by the ordinary lowering gate, not a claim of native execution.

## Comparison with 99q

Baseline: `stage3/adapt/99-correctness-axolotl-q/CENSUS.md`, compiler `d40c2470e`. The baseline splits its two enum roots by exact location here.

| 99q group | Roots | Status at f6c44ef5 | Current distribution |
| --- | ---: | --- | --- |
| `src/compiler/core.ts:736:36` (Refused), a cast the runtime can't check | 73 | moved | 73 at `src/compiler/core.ts:810:89` (Refused): a cast the runtime can't check |
| `whole` (accepted), accepted | 2 | unchanged | 2 at `whole` (accepted): accepted |
| `src/compiler/commandLineParser.ts:293:35` (NotYet), reading an enum before its runtime initialization; move the call after the enum declaration | 1 | unchanged | 1 at `src/compiler/commandLineParser.ts:293:35` (NotYet): reading an enum before its runtime initialization; move the call after the enum declaration |
| `src/compiler/debug.ts:243:59` (NotYet), a generic instantiation whose type argument for T, constrained to Node, can't be read | 1 | unchanged | 1 at `src/compiler/debug.ts:243:59` (NotYet): a generic instantiation whose type argument for T, constrained to Node, can't be read |
| `src/compiler/diagnosticInformationMap.generated.ts:13:45` (NotYet), reading an enum before its runtime initialization; move the call after the enum declaration | 1 | unchanged | 1 at `src/compiler/diagnosticInformationMap.generated.ts:13:45` (NotYet): reading an enum before its runtime initialization; move the call after the enum declaration |
| `src/compiler/semver.ts:71:9` (NotYet), a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 | unchanged | 1 at `src/compiler/semver.ts:71:9` (NotYet): a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |

New exact groups: 73 at `src/compiler/core.ts:810:89` (Refused), a cast the runtime can't check.

## Root-order artifacts

The meter uses the `node["throws"]` flag to label a probed readiness root as a `root-order artifact` (`stage3/meter/report.py`, root-order reporting). These labels do not delete or replace any per-root first-stop row in this report. The same `stage3/meter/node_readiness.py` helper was run on all three baseline initialization-stop roots with the adapted tree, pinned TypeScript emitter and Node v24.19.0. Its raw logs and resolved program lists are in `node-readiness/`.

| Root | Node exit | Throws | Label | First error |
| --- | ---: | --- | --- | --- |
| `src/compiler/debug.ts` | 1 | True | root-order artifact | TypeError: Cannot read properties of undefined (reading 'assert') |
| `src/compiler/diagnosticInformationMap.generated.ts` | 1 | True | root-order artifact | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') |
| `src/compiler/types.ts` | 1 | True | root-order artifact | TypeError: Cannot read properties of undefined (reading 'Error') |

## Largest group source and lowering emitter

Exact adapted source, `src/compiler/core.ts`, lines 806 through 812. Stop is line 810, column 89.

```typescript
/** @internal */
export function sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: EqualityComparer<T>): SortedReadonlyArray<T>;
/** @internal */
export function sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: EqualityComparer<T>): SortedReadonlyArray<T> {
    return deduplicateSorted(toSorted(array, comparer), equalityComparer ?? comparer ?? compareStringsCaseSensitive as any as Comparer<T>);
}

```

Search: `rg -n -F "a cast the runtime can't check" internal/lower`

```text
internal/lower/cast_proof.go:180:	refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: l.castRepair(node)}
```

The full grep output is retained in `reason-search.txt`. This is the ordinary cast refusal emitted in `internal/lower/cast_proof.go:180`. The observed first stop is the comparator cast at core.ts:810, after the deduplicateSorted brand mint. Admission of the earlier mint does not imply that its value lowers: the refusal loop runs before declaration and body lowering (`internal/lower/lower.go:93` through the subsequent declareModule/moduleStatements calls). The value-specific NotYet is emitted separately by `internal/lower/brand_mint.go:135`. A later refused cast can therefore be the first stop before value lowering reaches that NotYet. This census establishes first stops, not which later value normalization remains necessary.

## Reproduction and evidence

The commands are based on `stage3/meter/twice-daily.sh`. The cited 99q CENSUS.md contains the baseline and evidence description but does not contain commands.

```sh
git switch -c typescript/next-stops-f6c44ef5-b f6c44ef5
git submodule update --init --recursive --depth 1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/next-stops-upstream
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-upstream bash stage3/apply.sh /workspace/scratch/next-stops-adapted
go build -o /workspace/scratch/next-stops-census ./stage3/census/tool
python3 stage3/census/next-stops-f6c44ef5/run-groups.py
```

One fresh adapted tree is shared by all roots. The binary is built from the unmodified f6c44ef5 census source; there is no compiler overlay. Each invocation takes one original root file and writes its per-file row using the ordinary load/check/lower path. The runner terminates and reaps that process after the first row to avoid repeating the same root in the tool's final whole-program row. Any extra row already written stays in the raw stream. The normalized table takes exactly the first row for each invocation. Compiler roots include nested namespace, factory and transformer directories. Non-TypeScript directory entries and the 79-entry whole-program limit are outside this per-root table.

The four groups have 20, 20, 20 and 19 roots. Each group was committed and plainly pushed on exit. The setup checkpoint was pushed before group 1. These checkpoint commits implement the later checkpoint instructions; the branch changes only this directory and is not for landing.

`SETUP.md` records setup/apply timing and the adapted source manifest hash. `adapted-src-manifest.json` identifies every adapted source file. `binary.sha256` identifies the census executable. `raw.jsonl` concatenates the unmodified group streams. `group-*/raw.jsonl`, `group-*/stderr.log`, `group-*/rows.jsonl` and `group-*/roots.txt` preserve batch evidence. `results.json` contains the complete normalized 79-root table, and `baseline-99q.json` preserves the exact comparison input.

Total measured per-root time: 734.546 seconds. No root reached the 30-minute no-output limit. All started census processes were terminated if still running and reaped.
