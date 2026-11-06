# Views V3 checked element price

**Exact latent closure totals: incomplete.** 161 of 10721 full-mode units paired across 6 observed files.

**Source upper bound (seven forms): 2,518; readonly subset: 1,216.** Broader named roads: 2,861 / 1,302. Neither is the admission rule.

| Latent measure (observed pieces only) | Main | Head |
|---|---:|---:|
| view_elements | 0 | 48 |
| view_element_checks | 0 | 28 |
| units | 161 | 161 |
| attempted | 160 | 160 |
| panicked_units | 0 | 0 |
| ir_observed | 160 | 160 |
| finding_free_attempted | 94 | 94 |
| finding_free_view_elements | 0 | 15 |
| finding_free_view_element_checks | 0 | 12 |

Task 6wfvfhm. Report only, not for landing. Box `4b7d656a5150`, Linux x86_64, 2026-11-06 UTC.
Head `9626d0113f08269e519ba216ff7f95ff84af16db`; main `2c3f322975b2dee68889e82eab0bbca324d417a9`.
Report base `12fec995dcadc7f1ae94e4ba63835884451378ae`; upstream TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`.
Both pinned compilers use their own apply.sh products; all 746 src files are byte-identical. Both complete tsc entry programs are checker-clean with zero diagnostics. Only the report directory is committed.

## Exact latent observation

The runtime overlay counts ir.ViewElement and ir.ViewElementCheck by reflection using exact package and type names. It walks returned unit statements and all function bodies lowered in that unit result. Non-nil IR pointers, including ElementCheck, are visited once by identity. Main compiles without the optional types and reports zero. Counts are per isolated unit, so dependencies independently lowered in multiple units count in each observation. These are static IR nodes, not runtime execution counts.

The meter retains partial IR after statement rollback and omits final whole-program validation and pruning. attempted excludes skips and deferred generics; ir_observed counts attempts whose snapshot reached the counter; finding_free_attempted counts attempted units with no lowering-phase finding. The finding-free columns supply a strict meaning of units lowered without claiming successful whole-program compilation.

Each piece selects one or several files in the full tsc entry closure with LATENT_FULL=1 and LATENT_ONLY_FILE. Deferred instantiated witnesses have that piece context. Small files may share a piece to amortize project loading; no file appears in two completed pieces. The unmodified meter and counted binaries run with the identical selection. Removing only view_elements, view_element_checks and ir_measured from the raw counted bytes must reproduce the entire meter output byte for byte on both compilers. The ordinary fields, with only absolute adapted roots canonicalized, are passed to stage3/census/pair. Source bytes must match. Each binary run has a 23-minute deadline and is committed and plainly pushed as its own raw piece. The group is paired after all four raw runs are complete; then its paired archive retains the raw records and leg timings, replacing the redundant standalone archives.

The head-owned fixtures prove (ViewElement, ViewElementCheck): per-read (2,0), map-good (1,1), for-of-good (1,1). Main gives (0,0) for all three. Both skip-kind mutants fail the same known-count assertions. A separate shared-pointer probe confirms one check reached three ways counts once, and its skip-check mutant fails. Every other latent field is byte-identical on both trees for the fixtures and mutants. counter-proof.json and counter-proof.tar.gz retain this proof.

## Per-file latent counts

| File | Units paired / planned | Main elements | Main checks | Head elements | Head checks | Main attempted / free / panic | Head attempted / free / panic |
|---|---:|---:|---:|---:|---:|---:|---:|
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 1 / 1 | 0 | 0 | 0 | 0 | 1 / 1 / 0 | 1 / 1 / 0 |
| src/compiler/semver.ts | 29 / 29 | 0 | 0 | 12 | 3 | 29 / 21 / 0 | 29 / 21 / 0 |
| src/tsc/_namespaces/ts.ts | 1 / 1 | 0 | 0 | 0 | 0 | 1 / 1 / 0 | 1 / 1 / 0 |
| src/tsc/tsc.ts | 6 / 6 | 0 | 0 | 0 | 0 | 6 / 5 / 0 | 6 / 5 / 0 |
| src/compiler/builder.ts | 55 / 127 | 0 | 0 | 28 | 14 | 55 / 25 / 0 | 55 / 25 / 0 |
| src/compiler/executeCommandLine.ts | 69 / 69 | 0 | 0 | 8 | 11 | 68 / 41 / 0 | 68 / 41 / 0 |

## Top ten observed latent files

main:

- src/compiler/_namespaces/ts.moduleSpecifiers.ts: 0 elements, 0 checks
- src/compiler/semver.ts: 0 elements, 0 checks
- src/tsc/_namespaces/ts.ts: 0 elements, 0 checks
- src/tsc/tsc.ts: 0 elements, 0 checks
- src/compiler/builder.ts: 0 elements, 0 checks
- src/compiler/executeCommandLine.ts: 0 elements, 0 checks
head:

- src/compiler/builder.ts: 28 elements, 14 checks
- src/compiler/executeCommandLine.ts: 8 elements, 11 checks
- src/compiler/semver.ts: 12 elements, 3 checks
- src/compiler/_namespaces/ts.moduleSpecifiers.ts: 0 elements, 0 checks
- src/tsc/_namespaces/ts.ts: 0 elements, 0 checks
- src/tsc/tsc.ts: 0 elements, 0 checks

## Broad source upper bound

Stock TypeScript 6.0.3, with the options from void-intersections/count.cjs, enumerates read sites in the tsc entry closure on the completed main adapted tree. Array and readonly-array receivers include tuple receivers, union/intersection arms containing arrays and interfaces inheriting array bases. The readonly subset contains a readonly array or readonly tuple arm, including mixed unions. Neither column ports the checked-view admission rule. These count source sites, not a numeric limit on the sum of repeated dependency IR across isolated units. Index assignments using plain = are excluded; compound assignments and increments read and count. Repeated loop executions count as one static site. The seven fixture forms are index, at, for-of, spread, slice, map and forEach; named array handout methods are also inventoried by their method name.

The source fixture includes mutable and readonly arrays, union, tuple, mutable tuple and intersection receivers. It expects 12 reads with 10 in the readonly subset. The index-skipping mutant fails the same assertion.

Seven fixture forms alone: 2,518 reads; readonly subset 1,216. Additional named array roads account for the remainder of the broader total.

### Per form

| Form | Upper bound | Readonly subset |
|---|---:|---:|
| index | 1327 | 695 |
| for-of | 623 | 312 |
| spread | 278 | 43 |
| slice | 117 | 76 |
| map | 113 | 59 |
| forEach | 60 | 31 |
| join | 55 | 3 |
| pop | 48 | 0 |
| indexOf | 44 | 34 |
| filter | 42 | 16 |
| some | 33 | 17 |
| find | 30 | 1 |
| sort | 23 | 0 |
| splice | 21 | 0 |
| includes | 18 | 6 |
| every | 12 | 6 |
| concat | 6 | 1 |
| shift | 6 | 0 |
| findIndex | 4 | 2 |
| copyWithin | 1 | 0 |
| at | 0 | 0 |
### Top ten seven-form files

| File | Upper bound | Readonly subset |
|---|---:|---:|
| src/compiler/checker.ts | 1106 | 530 |
| src/compiler/utilities.ts | 137 | 68 |
| src/compiler/core.ts | 119 | 82 |
| src/compiler/program.ts | 107 | 61 |
| src/compiler/builder.ts | 62 | 27 |
| src/compiler/commandLineParser.ts | 62 | 28 |
| src/compiler/transformers/es2015.ts | 57 | 40 |
| src/compiler/transformers/generators.ts | 56 | 15 |
| src/compiler/debug.ts | 49 | 6 |
| src/compiler/parser.ts | 49 | 20 |
### Top ten broader-road files

| File | Upper bound | Readonly subset |
|---|---:|---:|
| src/compiler/checker.ts | 1243 | 569 |
| src/compiler/utilities.ts | 155 | 74 |
| src/compiler/core.ts | 131 | 82 |
| src/compiler/program.ts | 115 | 63 |
| src/compiler/commandLineParser.ts | 87 | 31 |
| src/compiler/builder.ts | 75 | 27 |
| src/compiler/parser.ts | 60 | 20 |
| src/compiler/transformers/generators.ts | 59 | 15 |
| src/compiler/moduleNameResolver.ts | 58 | 12 |
| src/compiler/transformers/es2015.ts | 57 | 40 |
### Per file

| File | Seven forms | Seven readonly | Broader roads | Broader readonly |
|---|---:|---:|---:|---:|
| src/compiler/checker.ts | 1106 | 530 | 1243 | 569 |
| src/compiler/utilities.ts | 137 | 68 | 155 | 74 |
| src/compiler/core.ts | 119 | 82 | 131 | 82 |
| src/compiler/program.ts | 107 | 61 | 115 | 63 |
| src/compiler/builder.ts | 62 | 27 | 75 | 27 |
| src/compiler/commandLineParser.ts | 62 | 28 | 87 | 31 |
| src/compiler/transformers/es2015.ts | 57 | 40 | 57 | 40 |
| src/compiler/transformers/generators.ts | 56 | 15 | 59 | 15 |
| src/compiler/debug.ts | 49 | 6 | 54 | 6 |
| src/compiler/parser.ts | 49 | 20 | 60 | 20 |
| src/compiler/emitter.ts | 43 | 22 | 51 | 24 |
| src/compiler/executeCommandLine.ts | 43 | 6 | 50 | 8 |
| src/compiler/transformers/module/module.ts | 39 | 14 | 40 | 14 |
| src/compiler/moduleNameResolver.ts | 37 | 8 | 58 | 12 |
| src/compiler/factory/utilities.ts | 32 | 7 | 33 | 7 |
| src/compiler/tsbuildPublic.ts | 32 | 19 | 37 | 20 |
| src/compiler/factory/nodeFactory.ts | 28 | 15 | 32 | 15 |
| src/compiler/transformers/module/system.ts | 28 | 8 | 28 | 8 |
| src/compiler/binder.ts | 27 | 16 | 30 | 19 |
| src/compiler/watch.ts | 24 | 10 | 29 | 11 |
| src/compiler/path.ts | 23 | 8 | 27 | 8 |
| src/compiler/moduleSpecifiers.ts | 20 | 9 | 29 | 14 |
| src/compiler/programDiagnostics.ts | 20 | 5 | 20 | 5 |
| src/compiler/sourcemap.ts | 20 | 7 | 20 | 7 |
| src/compiler/transformer.ts | 20 | 4 | 20 | 4 |
| src/compiler/utilitiesPublic.ts | 19 | 15 | 30 | 24 |
| src/compiler/transformers/classFields.ts | 18 | 14 | 18 | 14 |
| src/compiler/scanner.ts | 15 | 9 | 17 | 9 |
| src/compiler/sys.ts | 15 | 4 | 17 | 4 |
| src/compiler/expressionToTypeNode.ts | 14 | 14 | 15 | 15 |
| src/compiler/transformers/es2018.ts | 13 | 9 | 13 | 9 |
| src/compiler/transformers/module/esnextAnd2015.ts | 13 | 9 | 14 | 9 |
| src/compiler/resolutionCache.ts | 12 | 7 | 15 | 7 |
| src/compiler/semver.ts | 12 | 4 | 16 | 6 |
| src/compiler/transformers/esnext.ts | 12 | 11 | 12 | 11 |
| src/compiler/builderState.ts | 11 | 7 | 13 | 7 |
| src/compiler/transformers/es2017.ts | 11 | 10 | 12 | 11 |
| src/compiler/transformers/jsx.ts | 11 | 9 | 13 | 11 |
| src/compiler/tracing.ts | 10 | 3 | 11 | 3 |
| src/compiler/transformers/declarations.ts | 10 | 4 | 14 | 4 |
| src/compiler/transformers/utilities.ts | 9 | 8 | 9 | 8 |
| src/compiler/transformers/esDecorators.ts | 8 | 4 | 9 | 5 |
| src/compiler/symbolWalker.ts | 7 | 4 | 7 | 4 |
| src/compiler/transformers/legacyDecorators.ts | 7 | 5 | 7 | 5 |
| src/compiler/transformers/namedEvaluation.ts | 7 | 5 | 7 | 5 |
| src/compiler/watchUtilities.ts | 7 | 0 | 13 | 0 |
| src/compiler/transformers/destructuring.ts | 6 | 4 | 6 | 4 |
| src/compiler/visitorPublic.ts | 6 | 6 | 6 | 6 |
| src/compiler/factory/emitHelpers.ts | 5 | 4 | 5 | 4 |
| src/compiler/factory/emitNode.ts | 4 | 0 | 4 | 0 |
| src/compiler/transformers/ts.ts | 4 | 4 | 4 | 4 |
| src/compiler/transformers/classThis.ts | 3 | 3 | 3 | 3 |
| src/compiler/watchPublic.ts | 3 | 1 | 3 | 1 |
| src/compiler/transformers/es2020.ts | 2 | 0 | 2 | 0 |
| src/compiler/transformers/typeSerializer.ts | 2 | 2 | 2 | 2 |
| src/compiler/transformers/declarations/diagnostics.ts | 1 | 1 | 1 | 1 |
| src/compiler/transformers/taggedTemplate.ts | 1 | 1 | 1 | 1 |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 0 | 0 | 0 | 0 |
| src/compiler/builderPublic.ts | 0 | 0 | 0 | 0 |
| src/compiler/builderStatePublic.ts | 0 | 0 | 0 | 0 |
| src/compiler/corePublic.ts | 0 | 0 | 0 | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | 0 | 0 | 0 | 0 |
| src/compiler/factory/baseNodeFactory.ts | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeChildren.ts | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeConverters.ts | 0 | 0 | 2 | 2 |
| src/compiler/factory/nodeTests.ts | 0 | 0 | 0 | 0 |
| src/compiler/factory/parenthesizerRules.ts | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilitiesPublic.ts | 0 | 0 | 0 | 0 |
| src/compiler/hostErrors.ts | 0 | 0 | 0 | 0 |
| src/compiler/performance.ts | 0 | 0 | 0 | 0 |
| src/compiler/performanceCore.ts | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2016.ts | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2019.ts | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2021.ts | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 | 0 |
| src/compiler/tsbuild.ts | 0 | 0 | 0 | 0 |
| src/compiler/types.ts | 0 | 0 | 0 | 0 |
| src/tsc/_namespaces/ts.ts | 0 | 0 | 0 | 0 |
| src/tsc/tsc.ts | 0 | 0 | 0 | 0 |

## Evidence and status

Completed raw streams, meter baselines, paired deltas, logs and inner SHA256SUMS are packed in pieces/*.tar.gz. SHA256SUMS in this directory identifies every archive. The earlier raw-checkpoint.tar.gz records the superseded admission investigation and interrupted first apply; it is not current census evidence. Reproduction sources live in this directory; the meter and production compiler remain unchanged.

Full closure totals remain unavailable until every planned file is complete.

## Declaration partition

The frozen shared plan has 4564 top-level declarations owning 10721 full-mode units, including nested declarations, in 153 initial pieces. Go inventories from both compilers match the independent stock TypeScript inventory. Each unit belongs to exactly one contiguous top-level interval. Audits reject lost, duplicated and moved units, including a move that preserves the global inventory.

The posted four whole-file observations are seed leaves and are never remeasured. New pieces use the scheduling-only selector in full mode; first-error mode is not substituted. Selection-off output is byte-identical to the previously proven counted and original meter fixture streams on both compilers, and disjoint selected fixture units and lowering findings exactly reconstruct the whole fixture. New chunks run main then head and are paired immediately. They do not independently rerun a third and fourth meter binary: the original field-identity proof and the new selector-off proof are retained in declaration-selector-proof.tar.gz. This distinction avoids claiming a per-piece four-run identity test that was not performed.

Each new paired piece has a combined ten-minute limit, with raw output and timers saved even on interruption. An interrupted interval is never rerun unchanged: its immutable refinement partitions its top-level atoms and all owned units between two contiguous children. A single atomic declaration that times out is reported blocked rather than silently omitted. The shared root INDEX stays frozen and active refinements must preserve its complete inventory. Raw piece payloads are packed into rolling declaration-data archives to keep the file count bounded.
