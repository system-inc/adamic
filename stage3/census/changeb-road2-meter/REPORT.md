# Change B road 2 against A: paired latent census

**B introduces two lowering regressions despite three fewer distinct NotYet sites.** Both first-error censuses completed all 79 files on the same adapted tree, concurrently on one box. This is **measured on a checker-rejected program**; “lowers” here means an attempted unit has no lowering-phase finding, not ordinary accepted-program compilation or native output.

| Metric | A | B |
| --- | ---: | ---: |
| Distinct NotYet | 652 | 649 |
| Distinct Refused | 2,115 | 2,115 |
| Errors / panics | 0 / 0 | 0 / 0 |
| Attempted units without lowering finding | 2,930 | 2,928 |
| Attempted units | 4,017 | 4,017 |
| Unit lowering wall seconds (sum) | 536.176348 | 540.628953 |
| Unit lowering CPU seconds (sum) | 573.804803 | 577.673705 |
| Whole-run wall seconds | 960.309 | 967.821 |
| Whole-run peak RSS (KiB) | 5745076 | 6200580 |

The two A-successful units stopped by B are:

- `computeLineOfPosition`, `src/compiler/scanner.ts:520:1`: NotYet at `src/compiler/core.ts:1203:10`, “a checked predicate target readonly number[] over a generic argument readonly T[], which narrows to an intersection no view covers”.
- `indexOfNode`, `src/compiler/utilities.ts:1351:1`: NotYet at `src/compiler/core.ts:1203:10`, the same wording for `readonly Node[]` over `readonly T[]`.

Both call `binarySearch`, which reaches the `some(array)` predicate in `binarySearchKey`. No A unit is absent in B. Checker headers, all 4,543 unit eligibility records (including 498 deferred generics and 28 skipped checker bodies), and source hashes are identical. There are nine changed unit first stops: these two regressions, six existing blockers moving to the already-counted “a getter in a checked field contract” site at `sourcemap.ts:474:9`, and `getAllAccessorDeclarations` moving from that getter to a new predicate stop. No stopped unit becomes a no-finding attempt. See [all changed first stops](unit-outcome-changes.md) and [regression ledger](lowering-regressions.json).

All nine changed reason rows mention **predicates**; the two new generic-array rows also explicitly mention **“no view covers”**. The third new row is “a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which isn’t its supertype” (0 → 1), at `utilities.ts:6838:9`. Six “checked predicate overload target …” rows go 1 → 0: **ClassElement, JSDocPropertyLikeTag, LateBoundDeclaration, LeftHandSideExpression, LiteralExpression, TypeElement**. The [complete changed-reason table](changed-reasons.md) includes A/B counts and up to five locations per side. The [view/predicate table](view-predicate-reasons.md) names every matching row, including unchanged ones. No observed reason literally says “branded supertype”; no read/write-through-view row changes. The existing checked-view-field and callable-view rows remain 1 → 1.

Total unit lowering wall time is **+4.452605 seconds (0.830%, 1.00830×)**; CPU is +3.868902 seconds. Whole-run wall time is +7.512 seconds. The [20 largest B−A unit differences](timing-top20.md) give absolute seconds and ratios; [all units](timing-all.md) and [raw timing streams](A-times.jsonl.gz) retain the full pairing. Unit timing includes registration, lowering and panic recovery; it excludes loading/checking and refusal scanning, so its sum is below whole-run time.

Noise baseline: 1215 identical-outcome import/interface/type-alias units, whose lowering does not execute a runtime branded-predicate call. Median signed B−A is 0.514 ms; median absolute difference **3.046 ms**, p90 absolute **29.213 ms**. Median B/A is 1.0106; p10–p90 is **0.7418–1.3508**. Shared registration, GC and scheduling still contribute. This single concurrent pair measures an observed cost, not a repeat-run confidence interval or an isolated causal overhead estimate. [Baseline details](noise-baseline.json).

Box: 5 CPUs (`nproc`), CPU quota 400000/100000 (4 CPU equivalents), 18,882,662,400 bytes RAM, no swap. Both runs started at 2026-11-01 14:26:36 UTC, within 0.3 ms of each other; both exited 0 around 16 minutes later. Exact resource observations are in [A](A-resources.json), [B](B-resources.json) and [box](box.json).

Pins and reproducibility:

- A: `141ad1d244d6a2d9c5403858086d505dfdef5fd5`.
- B: `9e4f2184fd0d887bb582294b2ff91b1683c9e2f0` (`optimizer/branded-supertype-road2`).
- TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`; adapted once using A’s `apply.sh`.
- A binary SHA-256: `10371fb48fc2df1b6f414584ed0ec2c0f2320be156b2c11282c97b85561a4382`.
- B binary SHA-256: `b2b59ca4c19a0e50b513c01152485de8bf310303eaa871e7f721538a96469138`.
- Both SHAs already had the identical unit-times recorder, so no `1bab347531e8c238d97940ea843ad3c166509e18` backport was necessary. No tracked compiler files or anything under `internal/` was edited.
- [Exact commands](commands.md), [runner](run-censuses.py), [analysis](analyze.py), [source hashes](input-sha256.json), [provenance](provenance.json), and [logs](logs/). GNU time was unavailable in an initial wrapper; neither census executed in that failed launch. The successful pair used Python `wait4` peak RSS and monotonic wall time.
- [A ledger](A.jsonl.gz) and [B ledger](B.jsonl.gz) preserve raw events; distinct counts use `(kind, where, reason, text)`. The requested [census delta](census-delta.json) agrees with [the existing meter’s lowering_summary coverage/count validator](validation.json): 2,767 → 2,764 total NotYet/Refused sites. Each has 3 excluded SkippedDependency sites. No usable IR/output was allowed (`LATENT_ASSERT_NO_OUTPUT=1`); `LATENT_FULL=0`.
