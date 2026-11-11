# Paired unit-end view guard census

Task 3zex15c. Report only. Parent `16429a304` against head `769b0e534aa9106cec787a670d0243cd3734c9c0`. The source branch still pointed at that head at the start of this continuation.

**First-error census complete: the head adds 289 raw and 289 distinct guard findings.** These are 627 refusal observations over 635 recorded views, in 340 units. The parent guard accepts every recorded view. All reported first-error refusal paths have source-based readings in [the complete case catalogue](guard-findings.md).

Full-mode status: **12/26 affected files fully verified**. The remaining checks are in progress.

## Paired counts

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Cleared with no remaining finding | 0 | 0 |
| Changed to another reason | 0 | 0 |
| New guard findings | 289 | 289 |
| Pair-tool new-site rows | 289 | 289 |
| Guard first-stop changes | 340 | 340 |
| Unique new guard source identities | 233 | 233 |
| Unique sites previously free of normal and guard findings | 229 | 229 |

The existing pair tool counts a new-site row for each new reason at a previously guard-finding-free identity. Multiple reasons can share that identity; the unique identity counts above prevent those rows from being mistaken for different physical sites. Raw observations retain repeated view registrations and repeated views reached in different units. The pair tool deduplicates by its documented raw and access-chain identities.

Of the 340 newly guard-stopped units, **114** had no ordinary lowering finding and **226** were already ordinarily blocked. Changed guard first stops in already guard-blocked units: **0**. New guard probe panics/errors: **0**. Ordinary first stops and outcomes changed: **0**, verified byte for byte. [Every changed unit guard first stop and its reading](unit-stops.md) is listed separately.

These are unit-local counterfactual guard stops. Whole-program graph membership is unmodeled. The seam does not claim that these units would reach a production graph guard, or that the entire tsc program would compile. The production reason text still says "in a graph program" because it is the real guard's text.

## Every new refusal, grouped by road

| First new road in the actual failure path | Observations | New raw | New distinct |
| --- | ---: | ---: | ---: |
| elements | 2 | 2 | 2 |
| index values | 462 | 211 | 211 |
| union arms | 163 | 76 | 76 |
| base pair check | 0 | 0 | 0 |

A path can use several roads. The table assigns it to the first newly followed mechanism; every later step, source type and target type is retained in the exact reason and raw steps. The target "an arm" route also handles intersections in production. The catalogue quotes every new finding's reason, source text, root source/target names, owner units and reading.

The dominant defect is the combined target-arm and source-arm traversal. At `Modifier as Modifier`, the target loop follows `Modifier as AbstractKeyword` because one direction of assignability relates the union and its member. The subsequent source-arm loop then demands `AccessorKeyword as AbstractKeyword`. Those leaf kinds are genuinely disjoint. The guard is wrong to manufacture that pair under an unchanged `Modifier` slot. The same issue appears after the elements and index-value roads expose shared AST metadata.

The independent [identity fixture](identity.ts.txt) has `value: Choice` on both interfaces, with `Choice = A | B`. The parent accepts; the head refuses through `Choice as Choice`, `Choice as A`, and `B as A`. A write to that shared slot admits no value excluded by its source type. This uses the real guard on both sides, not a reimplementation. The fixture's normal file records are byte-identical; only cache telemetry differs.

The other first-error paths are checked predicate alternatives: the code tests node kinds or calls a checked tryCast predicate before admitting the corresponding target alternative. A disjoint source kind does not become that concrete target alternative. The guard has lost that correlation. Some predicates admit several target alternatives; the readings distinguish those alternatives rather than incorrectly claiming the whole predicate excludes all of them.

Reading categories: {'guard bug: identity/union cross-product': 278, 'guard bug: checked tryCast alternative': 2, 'guard bug: checked pattern alternatives': 4, 'guard bug: excluded discriminator arm': 5}. Unread first-error cases: 0. A false reported path does not prove that a different, later path through the same view would also be harmless; the real guard returns its first failure per view.

## Seam, state and noninterference

Only scratch overlays change. Production internal/lower files remain untouched. [instrument-guard.py](instrument-guard.py) applies the identical hook edits and [guard-hook.go.txt](guard-hook.go.txt) to both original meter entry overlays. The observer is deferred immediately after the first-error fresh lowering and the full-mode unit snapshot, in the same unit-end position as the earlier boundary/counting observer. It never calls findCycles or graphTypes.

For each unit with views, the observer takes the existing typed `latentSnapshot`, which copies lowering-owned maps and IR while retaining checker/AST identities. It evaluates each saved view in a singleton viewSources slice, using a fresh `&cycleFinder{l: probe, where: map[*checker.Type]*ast.Node{}}`. This exposes every recorded view refusal instead of hiding later views behind the guard's first refusal. It never calls latentRecord or assigns the guard error to the unit outcome.

No extra graph fields are populated. The guard reads the unit's checker/program, concrete type mapper/substitution, viewSources, staticGlobals and representation metadata. fields/deriveFields and isFunction/deriveIsFunction lazily fill finder-local fieldCache and functionCache; deriveRelated uses the checker. GraphTypes, selected graph nodes, shapes, reachability, allocation flows and cycle seeds are not prerequisites read by this guard. Existing ordinary registration/prepass metadata is copied from the unit; no whole-program cycle state is invented. All observed unit GraphTypes maps are empty because that pass is omitted, not because acyclicity was established.

The real guard acceptance decision is called unchanged with the mutant disabled. After a refusal, its real viewPairFailure is queried to extract the exact path; road classification does not decide acceptance. GuardRefusal, GuardUnit and GuardError records are separate side streams. Every attempted unit, including zero-view units, has an observation. Per-unit view counts agree on both sides. Probe errors are explicit and zero.

The scratch mutant adds an early refusal inside the real viewPairGuard for every singleton pair. On both compilers, in first-error and full modes, the small control records one view and one forced refusal with zero errors. The real control accepts the view. Original-binary comparisons preserve normal findings in both modes, proving that zero refusals cannot result from a silent seam. See raw/control-verification.json and raw/full-control-verification.json.

Both complete original first-error baseline streams are byte-identical. Each seam run preserves all header, file, unit and finding bytes against its own original binary. The exact serialized finding arrays contain 5767 observations and have SHA-256 `da230b61cdc7d4e729e15cb6a0e71d78dc80eb895f734b00900f30cd9ea84463` on all four runs. The only permitted differences are checker relation-cache telemetry counters caused by the additional queries; no such telemetry is mistaken for a lowering finding. [normal-findings-verification.json](normal-findings-verification.json) preserves the proof hashes and full-stream equality flags.

## Coverage and full-mode checks

The stock 6.0.3 parser independently counts 4562 top-level units in 81 resolved entry-program files: 4058 attempted units and 504 standalone generic functions deferred until a concrete call. Each side observed all 4058 attempted unit states and the same 635 views. The entry is checker-clean. First-error stops retain only the state reached before the normal failure, even when that unit body would be discarded. Deferred/unreached bodies, paths after a first error and whole-program graph selection remain outside this count.

The 26 affected full-mode files are the union of the new refusal owner files and source-site files. They differ because a caller unit can record a view at an instantiated predicate's declaration in another file. Each full run uses the same tsc entry with `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, and exactly one adapted `LATENT_ONLY_FILE`. Each verified file has an original-parent baseline plus parent and head seam runs; both ordinary streams are byte-checked against that common baseline. The complete original source-path proof and first-error baseline equality justify sharing it.

Full mode probes the successful state retained after normal failed-statement rollback. It can therefore have fewer views than a first-error state containing a failed statement. Full and first-error populations are reported separately. [FULL.md](FULL.md) lists every verified file, its counts, every quoted parent/head guard finding, and all pending files. No incomplete file is presented as a completed paired check.

## Pins, replay and resources

One adapted tree came from parent stage3/apply.sh and TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`, product key `e7b65b95e17a4f32a70bb0cb987f3c27056029d053a3eed60eb571c3cc70af25`. source-manifest.json hashes all 747 adapted src files. adapted-entry-sources.tar.gz preserves the exact 81 entry-program files, including type declarations used for the readings. Both compiler builds use their own source SHA and pinned cohere `5360753e19e4417675a90e92a1823bea66b13918`, TypeScript-go `d92d9bfee114c80be2c375d72edae966176e3a4f`. The report-only commits do not change those compiler source bytes. Binary hashes and original/guard overlays are archived.

Required setup was completed in the preceding report attempt: source /workspace/adamic-tools/env.sh, cloud/setup.sh, npm ci --prefix stage3/api, and explicit ADAMIC_TYPESCRIPT_SOURCE at the verified pin. The earlier 287b32b3 method was read; it explicitly added its counting hook and never ran findCycles. Its depth-limited pair probe is not substituted for the real guard here. The old intentionally stopped first-parent run is historical and excluded from these counts.

Box: four-CPU quota and 16 GiB memory limit. The four first-error processes overlapped with GOMAXPROCS=2. Full children use GOMAXPROCS=2 and a recorded GOMEMLIMIT=2560MiB soft target, with up to four independent children. [RUNS.md](RUNS.md) lists every actual child exit, elapsed seconds and peak RSS. Timings are overlapping observations, not isolated compiler performance comparisons. Logs are files; large finished raw streams are gzipped. Go evidence files end in .go.txt.

Checkpoints are pushed after preparation/controls and each completed run to typescript/view-guard-roads-census-checkpoint. The continuation is a new report-only commit on typescript/view-guard-roads-census, extending the preceding blocked-report commit. No main/source branch or production lowering change is pushed; no force push is used.
