# Latent census record facts: reviewed re-cut

**One commit on current-main pin `7e896ce4890d74568585b55973a3c982e7d2629a`, on `typescript/census-carry-facts-2`. Neither old carry commit is its parent. Byte identity is proven on the three-file subset only; an all-79-files comparison has not been completed.**

The unchanged baseline was built from its own detached worktree at the same pin, using that worktree's own overlay generator and templates. Both binaries loaded the identical 79-file adapted compiler tree; the existing exact `LATENT_ONLY_FILE` filter restricted attempts to utilities, binder and emitter in separate sequential processes. Input hashes were checked before the runs. No subset file was dropped; every run stayed well under the 12 GiB stop. A complete prior-tip comparison passed at `ae228d…`, but main advanced with a record-lowering change; both binaries and all six runs were rebuilt/repeated at the final pin. That superseded run is separate in [tip-advance-resources.csv](tip-advance-resources.csv).

## Requested revisions

1. **Only carried fields need a policy.** `loweringFieldPolicies` now lists only `entriesRecords`, `entriesProgramRecords`, and `entriesProgramRecordsKnown`, all `programFact`. The unused `unitSnapshot`, `sharedIdentity`, and `nilQuery` entries/constants are gone. Unlisted fields use ordinary typed snapshots. `TestLatentCarryFactsAuditedFields` audits those three fields against an independent three-field contract: declaration presence, exact audited type, table membership, generated capture/install and owned map copy. This independent contract keeps a deleted table entry red without requiring policies for unrelated fields. `TestLatentCarryFactsIgnoresUncarriedFields` adds a new scalar memo to a scratch lower.go and proves the carry guard accepts it.
2. **Replay is tied to the four main-source bodies.** `TestLatentCarryFactsReplayBodiesMatchMain`, beside `overlaytext/anchors_test.go`, fingerprints `entriesProgramHasRecords`, `entriesProgramRecordsWalk`, `entriesRecordLiteral`, and `entriesRecordLiteralWalk`. It hashes normalized AST bodies: parse without comments, then `go/format` with an empty FileSet so comments and source positions are excluded. [replay-body-shas.json](replay-body-shas.json) records the fingerprints. A change names the method and says the old replay must be re-proven by the subset byte-identity run in this report. A scratch no-op body edit turned it red.
3. **Facts use the registered base.** The one probe is now `base.latentSnapshot()` after `noteInheritance`, `noteAccessorNames`, `latentRegister`, and `latentRegisterNested`, rather than `fresh()`. These passes all receive the same ordered `program.Files()` and visit the program's declarations, not just the selected file: inheritance/getter names and global/nested declarations are program-wide registration; the selected file affects candidate enumeration and the refusal scan. The first registered base therefore has the declaration context supplied to every unit. The snapshot isolates the probe's module-read effects; only the proven map and two scalars are installed on bases, and the existing pending flag restores first-ask effect timing. No fresh-versus-registered equivalence is assumed.
4. **History is re-cut.** The two old commits were applied as a working patch on main, not retained as ancestors. The final branch is a single commit above the stated main pin, pushed plainly to the new name. The old branch remains untouched.

All three new tests call `t.Parallel()`, invoke no Go/compiler subprocess, and finish far under 30 s. The main-source fingerprint guard is not an overlay-only test. No file under `internal/` was changed.

## Current-tip byte identity and timings

| File | Baseline wall s | Carried wall s | Baseline peak KiB | Carried peak KiB |
|---|---:|---:|---:|---:|
| binder.ts | 98.973 | 66.182 | 1,594,476 | 1,633,392 |
| emitter.ts | 258.097 | 151.445 | 2,360,712 | 2,401,164 |
| utilities.ts | 210.191 | 146.142 | 2,099,832 | 2,139,880 |

Total census wall time: **567.261 → 363.769 s (1.559× faster)**. Summed unit lowering wall time: **516.779 → 304.264 s**; process CPU time: 664.087 → 445.864 s. Both runs attempted 1470 units. Peak RSS is the child's `wait4` rusage; `/proc` is sampled every 0.25 s for the safety stop, and wall time includes that completion-polling granularity. The box is `98b2b7d59531`: nproc 5, reported RAM 17.6 GB, CPU quota `400000 100000`, memory limit 16 GiB, no swap. These are single measured comparisons, not confidence intervals.

Both combined `latent.jsonl` files have SHA-256:

```text
aa793916377b0fa2297910a457d5afb4a01f41d90eca202f4f085061ff88410d
```

The combined file keeps one identical checker header and the binder/emitter/utilities records from the three independent filtered processes. [latent.sha256](latent.sha256) records every individual before/after file hash too; [binaries.sha256](binaries.sha256) identifies the two rebuilt executables. Raw ledgers are in `/workspace/scratch/carry2-subset/{baseline,carried}/`. No line differed. [timing-table.csv](timing-table.csv) contains all paired units; [top-20.csv](top-20.csv) contains the largest twenty drops.

| Unit (file:line) | Before s | After s | Drop s | Ratio |
|---|---:|---:|---:|---:|
| `emitLiteral` (emitter.ts:2118) | 5.980 | 0.089 | 5.891 | 67.04× |
| `emitFiles` (emitter.ts:752) | 15.586 | 9.774 | 5.811 | 1.59× |
| `diagnosticsEqualityComparer` (utilities.ts:8924) | 3.474 | 0.102 | 3.372 | 34.11× |
| `setCommonJsModuleIndicator` (binder.ts:3188) | 3.423 | 0.172 | 3.251 | 19.92× |
| `pipelineEmitWithHintWorker` (emitter.ts:1556) | 21.401 | 18.206 | 3.195 | 1.18× |
| `emitJSDocTypedefTag` (emitter.ts:4183) | 3.652 | 0.478 | 3.173 | 7.63× |
| `emitImportClause` (emitter.ts:3705) | 3.530 | 0.390 | 3.141 | 9.06× |
| `getDiagnosticCode` (utilities.ts:8936) | 3.190 | 0.051 | 3.138 | 62.18× |
| `createDetachedDiagnostic` (utilities.ts:8612) | 3.503 | 0.368 | 3.136 | 9.53× |
| `createFileDiagnostic` (utilities.ts:8685) | 3.445 | 0.323 | 3.122 | 10.68× |
| `emitJSDoc` (emitter.ts:4100) | 3.540 | 0.422 | 3.118 | 8.39× |
| `emitJSDocSimpleTypedTag` (emitter.ts:4128) | 3.189 | 0.084 | 3.104 | 37.79× |
| `getDiagnosticMessage` (utilities.ts:8940) | 3.146 | 0.050 | 3.095 | 62.80× |
| `emitLeadingComments` (emitter.ts:5976) | 3.291 | 0.195 | 3.095 | 16.85× |
| `bindExportAssignment` (binder.ts:3126) | 3.254 | 0.171 | 3.084 | 19.04× |
| `hasTabstop` (utilities.ts:11077) | 3.132 | 0.057 | 3.075 | 54.73× |
| `emitJSDocPropertyLikeTag` (emitter.ts:4250) | 3.401 | 0.327 | 3.074 | 10.42× |
| `emitTrailingSynthesizedComment` (emitter.ts:5900) | 3.277 | 0.216 | 3.061 | 15.14× |
| `compareMessageText` (utilities.ts:8818) | 3.430 | 0.378 | 3.052 | 9.08× |
| `emitVariableDeclaration` (emitter.ts:3402) | 3.386 | 0.336 | 3.050 | 10.08× |

## Tests and mutants

`go test -count=1 ./stage3/census/...` passed after the final-tip rerun and restored mutants. Vet passed for both changed test/generator packages, the non-test overlay lower source and the overlay driver. Logs: [tests](tests.log.txt), [package vet](statecopy-vet.log.txt), [source vet](overlay-source-vet.log.txt), [driver vet](tool-vet.log.txt). Expanded source-vet argv is retained in [overlay-source-vet-command.json](overlay-source-vet-command.json).

The three required mutants went red; each is restored or routed only through scratch test/source overlays. Exact lines:

```text
carry2-table-mutant: guard_test.go:246: carried lowering.entriesRecords is missing from the programFact table
carry2-type-mutant: guard_test.go:246: carried lowering.entriesProgramRecordsKnown type changed: got int, audited bool
carry2-body-mutant: carry_replay_test.go:77: entries_records.go: entriesProgramHasRecords body changed: carry replay was proven against the old bodies; re-prove it with the subset byte-identity run in stage3/census/latent/evidence/carry-facts/report.md (got 9dcfaeeb08a7609f4ae51c555b705b3a8f8ae966c8a4c7623809d155fa1e896f, want d08d63eabd0fb9f08c9750de9bce09b724ecbaf5ec5c3b52dfdd19117100157a)
```

[carry-facts-mutants.md](carry-facts-mutants.md) gives commands and individual logs. The unrelated-field control stays green. Tests and source/code scope checks are in [commands.md](commands.md).

## Retained earlier measurement evidence

The original profile and equality investigation remains explicitly historical, at pin `9c607b…`: 90 cold aggregate record walks, mean 3.067210 s; 89,229 unit/literal observations covering 1,256 literals with zero conflicting Boolean verdicts and eight consistently true literals. The profile chain was `enumNeverValue → property → entriesProgramHasRecords → entriesProgramRecordsWalk → entriesRecordLiteral → entriesRecordLiteralWalk`. The two named hotspots reached the same global Boolean and its literal-map discovery, not two separate global Booleans. Nested costs must not be added. The stored types produced different checker-type identity slices within runs and remain **uncarried**. The four fingerprinted query bodies are unchanged on the final main pin; the new registered-probe and current-tip output proof is the comparison above.

[Profile counts/costs](profile-walks.csv), [Boolean answers](program-record-answers.csv), [literal verdicts](literal-verdicts.csv), [all observations](literal-observations.jsonl.gz), [candidate equality counts](answer-equality.csv), [counterexample](storage-types-counterexample.json), [pprof](utilities-profile-top.txt), and [sampled stack](enum-record-stack.txt) preserve that investigation. Detailed old costs/counts and decisions for other candidates are in [prior-report.md](prior-report.md), clearly labelled as the previous version. The earlier 16 GiB kill remains in [full-baseline-blocker.md](full-baseline-blocker.md). Those old timings/hashes are not the new result.

TypeScript/adapted input pin is `050880ce59e30b356b686bd3144efe24f875ebc8`; [adapted-source-shas.csv](adapted-source-shas.csv) verifies all 79 files. Cohere is `06af2e399c6c1fb93fd570b774027942747400b4`, its TypeScript submodule `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go 1.27.1, Node 24.19.0, clang 20.1.8.

[commands.md](commands.md) gives the exact current-tip measurement commands and an **unfiltered all-79-files baseline-versus-carried command for a larger house box**. Byte identity remains proven on the subset only.
