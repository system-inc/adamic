# Next stops after brand-mint admission

All 73 roots that stopped at `core.ts:736:36` in 99q now stop at `src/compiler/core.ts:810:89`, Refused: "a cast the runtime can't check". No root still stops at core.ts:736. The next refusal is the comparator fallback in `sortAndDeduplicate`, before ordinary lowering reaches the admitted brand values. This census does not observe a NotYet at those values.

Compiler and adaptation SHA: `f6c44ef5f5d3908e134e2f098bf5c9cabdff424f`, fetched as origin/main. Report-only branch: `typescript/next-stops-f6c44ef5`.

TypeScript: `v6.0.3`, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Both measurement tools and the single adapted tree come from the measured Adamic SHA. The ordinary tool uses no overlay, suppression, backend change, or partial-root fallback.

Box: `56aa137099a9`; 5 available CPUs; quota 400000/100000 microseconds; memory 18440100 kB. Full load metadata is in box.json. Times describe this box only.

The complete ordinary run took 816.834 recorded seconds (13.61 minutes). It finished before the 30-minute timeout; no split was needed. All 79 roots have exactly one row. The three JSON-input rows and the aggregate 79-root single-entry error remain in census.jsonl, outside the table.

## First stops

Root names below are relative to `src/compiler/`. Groups use the exact (location, kind, reason). Artifact labels are supplemental; no root is folded away.

| Count | Location | Kind | Reason | Root names |
| ---: | --- | --- | --- | --- |
| 73 | `src/compiler/core.ts:810:89` | Refused | a cast the runtime can't check | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations.ts`, `transformers/declarations/diagnostics.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | - | accepted | - | `corePublic.ts`, `hostErrors.ts` |
| 1 | `src/compiler/commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `diagnosticInformationMap.generated.ts` (root-order artifact) |
| 1 | `src/compiler/debug.ts:243:59` | NotYet | a generic instantiation whose type argument for T, constrained to Node, can't be read | `core.ts` |
| 1 | `src/compiler/diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `types.ts` (root-order artifact) |
| 1 | `src/compiler/semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | `debug.ts` (root-order artifact) |

**2/79 roots lower whole:** `corePublic.ts`, `hostErrors.ts`. This means the ordinary Lower call succeeded; no native output is claimed.

## Comparison with 99q

The per-root comparison in results.json uses 99q's census-results.json as well as its CENSUS.md table. Compiler and adaptation changes are measured together; this is not a causal isolation of one commit.

| 99q group | Current group | Classification |
| --- | --- | --- |
| 73 Refused at core.ts:736:36 | The same 73 roots, Refused at core.ts:810:89, same reason | Moved; the core.ts:810:89 location group is new |
| 2 lower whole | corePublic.ts and hostErrors.ts | Unchanged |
| 2 enum-before-init | diagnosticInformationMap.generated.ts at commandLineParser.ts:293:35; types.ts at diagnosticInformationMap.generated.ts:13:45 | Unchanged, kept as two exact-location groups |
| 1 at semver.ts:71:9 | debug.ts, same NotYet and reason | Unchanged |
| 1 at debug.ts:243:59 | core.ts, same NotYet and reason | Unchanged |

No additional root or reason group is new. In total: 73 moved, 6 unchanged.

## Biggest group source and emitter

The ordinary path in `internal/lower/lower.go` runs module refusal scans before `declareModule` and function-body lowering. Recognition admits the mint casts to that scan, but the later comparator refusal stops this program before branded-value lowering. The tests of isolated mint examples and these full-root first stops therefore observe different phases.

Exact adapted `src/compiler/core.ts:807-811` (line 810 is the stop):

```ts
export function sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: EqualityComparer<T>): SortedReadonlyArray<T>;
/** @internal */
export function sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: EqualityComparer<T>): SortedReadonlyArray<T> {
    return deduplicateSorted(toSorted(array, comparer), equalityComparer ?? comparer ?? compareStringsCaseSensitive as any as Comparer<T>);
}
```

Column 89 starts `compareStringsCaseSensitive as any as Comparer<T>`. The full return line is retained above.

The reason is constructed in `internal/lower/cast_proof.go:180`, inside `castProof`. The cast scan calls it through `castRefusal` at line 150; the scan rule is registered at line 164. The `any` source/target branch returns this refusal at lines 185-186. Brand-mint recognition is checked earlier, at lines 176-178.

```go
    refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: l.castRepair(node)}
```

Located with `rg -n -F "a cast the runtime can" internal/lower`; the source file and its SHA are pinned by the compiler commit.

## Evidence

The ordinary binary build metadata records Go 1.27.1 and f6c44ef5. Its vcs.modified flag is true with the untracked report directory present; source-diff.txt is empty and confirms the measured tracked sources are unchanged.

`census.jsonl` and `census.log.gz` retain the complete ordinary stream and stderr. `results.json` retains normalized roots, groups, timings, and the per-root 99q comparison. `provenance.json` records compiler, TypeScript pin, binary hashes, and key adapted source hashes. `COMMANDS.md` records preparation and invocation commands. Build and apply logs are retained alongside these files.

## Root-order artifact labels

The exact candidate expression from `stage3/meter/report.py:root_order_checks` was evaluated with proven inputs, then its Node-throws classification was applied. All three readiness stops are **root-order artifacts**. Their original first stops and root counts remain in the table. This does not certify whole-program lowering for their entry reaches.

The proof avoids an unrelated complete body census. Native full-mode unit timing rows witness attempted units in every stop file used by a readiness candidate: semver.ts, commandLineParser.ts, and diagnosticInformationMap.generated.ts. The timer runs only after the unit is marked attempted; the only later status change is panic, which the meter also accepts. Additional attempted files cannot change candidate membership because every ordinary readiness stop file is already witnessed.

The readiness-finding set is provably empty in this guarded entry binary. Both readiness reasons are emitted only inside `namespaceInitialization` in `internal/lower/namespaces.go` (lines 255, 257, 306, and 309). Its sole production call is in lower.go:90, removed by the meter overlay when LowerWithOptions is replaced with its no-output error. The active overlay has only the function definition, no call or method-value reference. The guarded entry binary contains neither the function symbol nor either reason literal; the ordinary binary contains all of them. `artifact-proof.json` records these checks, binary hashes are in provenance.json, and the symbol-table evidence is retained. No finding or first-stop record is synthesized.

| Root | Original first stop | Node result | Meter label |
| --- | --- | --- | --- |
| `debug.ts` | `src/compiler/semver.ts:71:9` | exit 1: TypeError: Cannot read properties of undefined (reading 'assert') | root-order artifact |
| `diagnosticInformationMap.generated.ts` | `src/compiler/commandLineParser.ts:293:35` | exit 1: TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | root-order artifact |
| `types.ts` | `src/compiler/diagnosticInformationMap.generated.ts:13:45` | exit 1: TypeError: Cannot read properties of undefined (reading 'Error') | root-order artifact |

Node v24.19.0 and TypeScript 6.0.3 were used through the unchanged meter node_readiness.py helper. Native attempted-unit witnesses, the exact candidate expression, all probe results, and the reachability certificate are retained in artifact-proof.json. The incomplete supplementary body scans were stopped and are not represented as complete entry censuses. Their original data, including the timing witnesses, are in auxiliary-entry-attempts.tar.gz. Symbol tables are entry-symbols.txt.gz and census-symbols.txt.gz. The 79-root ordinary census is complete.
