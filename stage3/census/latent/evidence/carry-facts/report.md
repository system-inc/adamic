# Census carry in first-error mode and across files

**Stopped at the final record-composition assertion in step 3.** The carry passes both requested one-process baseline-versus-changed comparisons, and all three fresh single-file pairs. But utilities-only and multi-file records already differ in unchanged main: nine deferred-generic `instantiated` witnesses are false alone and true with binder/emitter. The same difference is present on the changed side. It is not valid to claim those raw file records match; no witness fields were normalized or reset to force equality.

Branch: `typescript/census-carry-facts-3`, one re-cut commit above fetched main pin `7e896ce4890d74568585b55973a3c982e7d2629a`. The prior carry commits are not ancestors. No file under `internal/` changed. Remote main advanced to `50503bc6abc22da12e9f9a0caad28b7f4bb1b20d` during the measurements. The stopped result stays on its measured starting pin; it is **not** a proof or re-cut on that newer remote tip. I did not relabel the measurements or rebase onto an unmeasured compiler after the composition blocker.

## What changed

`latentRun` now has one facts pointer outside its file/unit loops. After every unit's original `fresh()`, `noteInheritance`, `noteAccessorNames`, and `latentRegister`, it calls the same `latentUseProgramFacts` helper as full mode. On the first call, the helper snapshots that registered lowering, computes the same three audited fields once, and captures only `entriesRecords`, `entriesProgramRecords`, and `entriesProgramRecordsKnown`. Every installation clones the literal map, applies the fields, and sets `latentProgramRecordsPending`. Fresh first-error units therefore own their maps even though they do not take full-mode unit snapshots. The guard now checks that installation clone as well as generated capture/install and snapshot copying.

Full mode calls the helper after its normal registration including `latentRegisterNested`. Its facts pointer also spans the whole census. Its `LATENT_ONLY_FILE` accepts newline-separated exact absolute names; none of the 79 measured paths contains a newline. This remains a membership filter in normal program file order, not a change to program loading. Speculative traversal and its existing selector are unchanged and unproven here.

## Completed identity proofs

1. **First-error, all 79 files:** `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, no file filter, one unchanged baseline process followed by one changed process. Both produced 79 file records and 4,026 timed units. This exercises reuse across file checkers in the actual requested meter mode.
2. **Full mode, binder + emitter + utilities:** `LATENT_FULL=1`, newline-list filter, one process each side. Both produced three records and 1,470 timed units. Binder is first in normal file order, so its facts are applied to emitter and utilities. The main baseline lacks a list selector; its full-mode binary has only a scratch selection shim, with no carry or lowering-state changes. The first-error baseline is strictly the unchanged main overlay.
3. **Three fresh full-mode single-file pairs:** each baseline/changed pair is byte-identical. These do not independently prove cross-file reuse; proof 2 does.

All inputs retained the complete same 79-file checked program and declarations. [adapted-source-shas.csv](adapted-source-shas.csv) was verified before running. All census processes ran sequentially, with a 14 GiB stop. No run crossed it, so no file was dropped. Peak RSS is child `wait4` rusage; `/proc` sampling every 0.2 s enforces the stop. Wall times include that completion polling. These are single measured comparisons, not confidence intervals.

| Proof | Baseline wall s | Changed wall s | Baseline peak KiB | Changed peak KiB | Unit wall sums before → after s |
|---|---:|---:|---:|---:|---:|
| first-error-all79 | 924.834 | 702.319 | 6,076,976 | 6,194,448 | 517.008 → 299.002 |
| full-multi | 572.727 | 359.848 | 4,714,032 | 4,779,780 | 526.059 → 310.839 |
| full-single-binder.ts | 98.475 | 65.780 | 1,612,988 | 1,652,348 | 84.497 → 48.804 |
| full-single-emitter.ts | 258.040 | 149.408 | 2,368,404 | 2,369,048 | 239.878 → 127.987 |
| full-single-utilities.ts | 207.955 | 146.839 | 2,113,256 | 2,135,288 | 190.040 → 125.870 |

The actual first-error wall time is **924.834 → 702.319 s (1.317× faster)**; summed lowering time is **517.008 → 299.002 s**. Full one-process wall time is **572.727 → 359.848 s (1.592× faster)**; summed lowering time is **526.059 → 310.839 s**. Exact values, CPU sums and counts are in [resources.csv](resources.csv); [timing-table.csv](timing-table.csv) is the requested summary. Per-unit ledgers and top-20 drops are committed for each case, including [first-error timings](first-error-all79-unit-timings.csv), [first-error top twenty](first-error-all79-top-20.csv), [full multi-file timings](full-multi-unit-timings.csv) and [full multi-file top twenty](full-multi-top-20.csv).

### Both hashes per one-process proof

first-error-all79:

```text
baseline d4844b59a75ab2fef73d2293337664358d1e85d38814138e083b0a07147b6d9a
changed  d4844b59a75ab2fef73d2293337664358d1e85d38814138e083b0a07147b6d9a
```

full-multi:

```text
baseline 0f08f80d0e1b82b47c3a4af9f7ed5dddd4811757258faed07b9a0e0e69da6906
changed  0f08f80d0e1b82b47c3a4af9f7ed5dddd4811757258faed07b9a0e0e69da6906
```

[latent.sha256](latent.sha256) records every run and the raw singles-combined ledgers; [binaries.sha256](binaries.sha256) records all three executables. Raw JSONL and per-unit timing streams remain under `/workspace/scratch/carry3-runs/`. Neither proof establishes full-mode identity over all 79 files, nor equivalence between different attempted-unit sets.

## Why the final composition assertion fails

`Latent` resets one census-wide `latentInstantiated` map, collects file records through `latentRun`, and only after all selected files have run fills each deferred unit's `Instantiated` field from that map. The generic call hook records concrete body-entry witnesses from callers in other files. This is the behavior already described in the deferred-generic README: witnesses survive callers in later files.

Therefore loading the same program is not the same as attempting the same units. Binder/emitter attempts contribute nine utilities witnesses absent from a utilities-only attempt set. [file-record-comparison.csv](file-record-comparison.csv) shows binder and emitter records match exactly on both sides; utilities differs. [composition-differences.csv](composition-differences.csv) shows the only parsed differences, on both unchanged and changed sides:

| Utilities deferred unit | Single-file | Multi-file |
|---|---:|---:|
| `changeExtension` | false | true |
| `setTextRangePos` | false | true |
| `setTextRangeEnd` | false | true |
| `setTextRangePosEnd` | false | true |
| `setParent` | false | true |
| `setParentRecursive` | false | true |
| `bindParentToChildIgnoringJSDoc` | false | true |
| `bindJSDoc` | false | true |
| `bindParentToChild` | false | true |

These witnesses are observations of the attempted census, not program-only lowering memos. Changing the carry cannot make the unchanged baseline's two records equal. Resetting witnesses per file, masking `instantiated`, or injecting unobserved witnesses into single-file results would change the existing ledger's meaning and invalidate the requested raw comparison. None was done.

Raw composition hashes (diagnostic assembly only, not a passed identity):

```text
baseline one process:     0f08f80d0e1b82b47c3a4af9f7ed5dddd4811757258faed07b9a0e0e69da6906
baseline singles combined: aa793916377b0fa2297910a457d5afb4a01f41d90eca202f4f085061ff88410d
```

```text
carried one process:     0f08f80d0e1b82b47c3a4af9f7ed5dddd4811757258faed07b9a0e0e69da6906
carried singles combined: aa793916377b0fa2297910a457d5afb4a01f41d90eca202f4f085061ff88410d
```

The runner exited 3 at this assertion; [composition-failure.json](composition-failure.json) retains its exact location. [proofs.json](proofs.json) distinguishes all passed comparisons from the failed composition condition. No claim is made that the whole requested proof suite passed.

## Guards, tests and provenance

The three mutants were rerun: removed `entriesRecords` table entry, changed carried-known flag type in a scratch lower.go, and a no-op inserted in a scratch query body. Each went red; source restoration was followed by `go test -count=1 ./stage3/census/...` whole, which passed. Vet passed for the generator/guard packages, overlay non-test lower source and driver. [Mutant red lines](mutant-red-lines.txt), [mutant procedure](carry-facts-mutants.md), [tests](tests.log.txt), [package vet](statecopy-vet.log.txt), [overlay vet](overlay-source-vet.log.txt), and [driver vet](tool-vet.log.txt) retain those results. New tests use `t.Parallel()`, in-process parsing/generation only, and finish under 30 s. The narrowed carried-field and normalized main-body guards remain in place.

Machine: `98b2b7d59531`, nproc 5, reported RAM 17.6 GB, cgroup CPU quota `400000 100000` (4 CPUs), 16 GiB memory limit, no swap. TypeScript/adapted pin `050880ce59e30b356b686bd3144efe24f875ebc8`; Cohere `06af2e399c6c1fb93fd570b774027942747400b4`, its TypeScript submodule `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go 1.27.1, Node 24.19.0, clang 20.1.8.

Earlier full-only, separate-process measurements remain historical in [full-only-report.md](full-only-report.md), alongside the older profile, literal-verdict equality data and storage-type counterexample. Those are not a first-error or cross-file proof. The original full all-files 16 GiB failure is also historical; the successful all-files proof here is first-error only.

[commands.md](commands.md) gives exact commands and scope, including the scratch list-selection shim. This result is stopped at the composition condition, with the implemented carry and the successful same-attempt-set proofs pushed for review.
