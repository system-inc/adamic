# Submit head `810ce7d9` against its moved base

**Zero control-to-submit lowering regressions. Distinct NotYet and Refused totals are unchanged.** Six already-stopped units move from overload-target stops to six target-specific **“generic claim TOut”** stops. Both first-error censuses completed all 79 files concurrently on the same box. This is **measured on a checker-rejected program**: a no-stop attempt is not a claim of ordinary accepted-program compilation or native output.

| Metric | A: base `54412bc0` | B: submit `810ce7d9` |
| --- | ---: | ---: |
| Distinct NotYet | 652 | 652 |
| Distinct Refused | 2,115 | 2,115 |
| Errors / panics | 0 / 0 | 0 / 0 |
| Attempted units without lowering finding | 2,930 | 2,930 |
| Timed attempted units | 4,017 | 4,017 |
| Unit lowering wall seconds (sum) | 527.742235 | 513.993167 |
| Unit lowering CPU seconds (sum) | 570.632129 | 547.790641 |
| Whole-run wall seconds | 936.443 | 928.681 |
| Peak RSS (KiB) | 5558364 | 5794972 |

The new base is **`54412bc063f205c452ef6839c19ec01662327f7c`**; the submit head is **`810ce7d9ced9105f188a6171cc384a61f62c70ac`**, exactly three commits above it. The requested [base diff](base-diff-stat.txt) from `141ad1d244d6a2d9c5403858086d505dfdef5fd5` shows **six conformance fixture files, 37 insertions**; there are **no changes under `stage3/adapt` or `stage3/census`**. Therefore this pair **reuses `/workspace/scratch/tree`** rather than rebuilding it. The adaptation driver, source pin, API dependencies and dependency pins also match. All 79 source hashes were checked before and after the pair ([manifest](input-sha256.json)). Both census binaries were built at their own requested SHAs using the meter overlay commands; the new control binary is byte-identical to the earlier control.

No unit that lowers on A is absent or stopped on B: the [regression list is empty](lowering-regressions.json). All 2,930 attempted units without lowering-phase findings are retained. The earlier two regressions, `computeLineOfPosition` (`scanner.ts:520:1`) and `indexOfNode` (`utilities.ts:1351:1`), remain no-stop attempts on both sides ([explicit check](prior-regressions-check.json)). The [six changed first stops](unit-outcome-changes.md) are all in already-stopped units; no stopped unit becomes a no-stop attempt.

The six requested overload-target rows each go **1 → 0**. For each target below, the exact replacement reason is **`a checked predicate target <Target> that the generic claim TOut may be instantiated at`**, each **0 → 1**, always at **`src/compiler/core.ts:1779:35`** (`test(value)` in `tryCast`). These are **six distinct sites under `(kind, where, reason, text)`, at one source location**, because the reasons name different targets. There are no “generic claim” reasons on A.

| Target | Overload A | Overload B | Generic claim A | Generic claim B | Old overload site |
| --- | ---: | ---: | ---: | ---: | --- |
| ClassElement | 1 | 0 | 0 | 1 | src/compiler/utilitiesPublic.ts:1732:1 |
| JSDocPropertyLikeTag | 1 | 0 | 0 | 1 | src/compiler/utilitiesPublic.ts:1449:1 |
| LateBoundDeclaration | 1 | 0 | 0 | 1 | src/compiler/emitter.ts:1158:18 |
| LeftHandSideExpression | 1 | 0 | 0 | 1 | src/compiler/core.ts:1784:32 |
| LiteralExpression | 1 | 0 | 0 | 1 | src/compiler/utilitiesPublic.ts:1498:1 |
| TypeElement | 1 | 0 | 0 | 1 | src/compiler/utilitiesPublic.ts:1782:1 |

All **12 changed reason rows** mention **predicates**: these six removals and six additions. The [changed-reason table](changed-reasons.md) gives every row, counts and up to five example locations per side. [Overload rows and sites](overload-target-rows.json), [all generic-claim reasons and all their sites/attempting units](generic-claim-reasons.json), and the [full generic-claim site table](generic-claim-sites.md) retain the complete details without truncation. The [view/predicate table](view-predicate-reasons.md) includes unchanged rows; no read/write-through-view row changes and no observed reason literally says “branded supertype”.

Observed unit lowering wall time is **-13.749068 seconds (-2.605%, 0.97395×)**; CPU is **-22.841488 seconds (-4.003%)**. Whole-run wall time is -7.762 seconds. The [20 largest B−A unit differences](timing-top20.md) give absolute seconds and B/A ratios; [all units](timing-all.md) and [JSON](unit-times.json) preserve every matched unit. Unit timing includes registration, lowering and recovery, excluding loading/checking and refusal scanning.

Noise floor: 1215 identical-outcome import/interface/type-alias units whose lowering does not execute a runtime branded-predicate call. Median signed B−A is -1.176 ms; median absolute difference **2.774 ms**, p90 absolute **21.433 ms**. Median B/A is **0.9757**, p10–p90 **0.8362–1.1097** ([baseline](noise-baseline.json)). The unaffected-unit median itself shifts toward B by about 2.43%; this single pair does not isolate a causal speedup from GC, cache and scheduling variation.

Both runs started at 2026-11-01T15:22:17.853585+00:00 (A) and 2026-11-01T15:22:17.853828+00:00 (B), within 0.3 ms. B ended at 2026-11-01T15:37:46.534457+00:00; A at 2026-11-01T15:37:54.296195+00:00. Both exited 0 without hitting the 5400-second bound. Same box `6923d492b4a0`: 5 CPUs (`nproc`), cgroup quota 400000/100000 (4 CPU equivalents), 18,882,662,400 bytes RAM, no swap ([box](box.json), [A resources](A-resources.json), [B resources](B-resources.json)).

Pins and reproduction:

- Control: `54412bc063f205c452ef6839c19ec01662327f7c`; submit: `810ce7d9ced9105f188a6171cc384a61f62c70ac`.
- Adapted source: TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`, originally adapted once using `141ad1d2`, reused because adaptations are unchanged.
- A binary SHA-256: `10371fb48fc2df1b6f414584ed0ec2c0f2320be156b2c11282c97b85561a4382`.
- B binary SHA-256: `22fc2f107f11c96a3bfb9b86fbfc6080d002f8fc2bfa3c118fa304b426686c0e`.
- Both compiler worktrees have identical census/recorder code, no recorder backport and no tracked edits. Nothing under `internal/` was changed. The initialized `cohere` pin is shared through the same dependency symlink arrangement. [Exact commands](commands.md), [provenance](provenance.json), [runner](run-censuses.py), [analysis](analyze.py), [packager](package.py) and [logs](logs/).
- [A ledger](A.jsonl.gz), [B ledger](B.jsonl.gz), [A unit times](A-times.jsonl.gz) and [B unit times](B-times.jsonl.gz) round-trip exactly to the raw streams. Existing meter [lowering_summary validation](validation.json) and the requested [census delta](census-delta.json) agree: **2,767 → 2,767** distinct NotYet/Refused sites over **79 files**. Three SkippedDependency sites on each side are excluded. Checker headers and all 4,543 unit eligibility records (4,017 attempted, 498 deferred, 28 skipped) are identical; the control raw ledger is byte-identical to the previous controls. `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1` throughout.
