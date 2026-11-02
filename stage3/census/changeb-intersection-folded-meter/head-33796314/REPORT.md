# New optimizer head 33796314: same-tree latent census

**C completed all 79 files**, alone on the original box and unchanged adapted tree. The original A ledger remains the control; the original B ledger supplies the incremental comparison. All three pass the meter’s existing `report.py.lowering_summary` coverage and distinct-count validator; headers agree exactly. C exited 0.

| Distinct sites | A (head 1) | B (5e5f57d3) | C (33796314) |
|---|---|---|---|
| NotYet | 555 | 554 | 554 |
| Refused | 2652 | 2652 | 2652 |
| error | 0 | 0 | 0 |
| panic | 0 | 0 | 0 |
| SkippedDependency | 2 | 2 | 2 |

A-lowers → C-stops: **0 outside the record guard**, **0 at the record guard**. All successful-A units absent or otherwise non-lowering on C: **0**. [Separate record-guard exceptions with their sites](vs-a/record-guard-regressions.md), [other regressions](vs-a/other-regressions.md), [full absent/non-lowering list](vs-a/lowers-to-absent-or-nonlowering.json).

C has **7 record-guard first stops**, across **7 distinct guard sites**. These seven stops are unchanged from both A and B; none is a new A-lowers/C-stops exception. The exact reason is **a union or optional field read in a program with record storage**. [Every guard first stop with A/B outcomes](record-guard-first-stops.md) and [every guard site and owner](record-guard-sites.md) include cases that were already stopped on A or B.

## First stops and sites

A → C unit first-stop/outcome changes: **1**; changed reason rows: **1**. [Every named unit that moves](vs-a/unit-changes.md), [every changed reason with up to five locations on each side](vs-a/changed-reasons.md), and [every A lowering-phase site absent from C](vs-a/lowering-sites-absent-c.md). Reasons mentioning a view, branded supertype, intersection or predicate are marked `*` by their full wording in the reason table.

B → C: **zero site, reason-row, unit-outcome or first-stop changes**. The complete C ledger is **byte-for-byte identical to B**, raw SHA256 `d222ff253f037e6d1d1180d2c4ee59f845060600c71ed6c56b75ad1f9c48dead`. This first-error census reaches no additional stop from the record-guard change. [Exhaustive site list](vs-5e5f57d3/site-movements.md), [unit list](vs-5e5f57d3/unit-changes.md) and [reason counts](vs-5e5f57d3/changed-reasons.md) are empty; JSON companions retain the empty additions and removals.

`nullParenthesizerRules`, `factory/parenthesizerRules.ts:675:1`:

- A: NotYet: a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} | null), read on the true side as Expression & LeftHandSideExpression @ `src/compiler/core.ts:1799:32`.
- B: NotYet: a template interpolating an object, an array, a map, a function or undefined @ `src/compiler/core.ts:1801:59`.
- C: NotYet: a template interpolating an object, an array, a map, a function or undefined @ `src/compiler/core.ts:1801:59`.

The only changed reason row against A is the checked-predicate row quoted above, **1 → 0** at `core.ts:1799:32`. The template-interpolation row remains **2 → 2 distinct sites**, because the reached stop already appears in A through another caller. No view, branded-supertype, intersection or write-through-view reason row changes.

All **78 direct cast/tryCast calls in 22 units** use the original source inventory, whose source hashes still match. [Every caller’s A/B/C outcome](caller-outcomes.md) names the unit and first stop; checker-skipped bodies are identified explicitly, with the first checker diagnostic rather than a claimed lowering result. [Full source calls and outcomes](caller-outcomes.json). Deferred-generic witness metadata changes are saved separately from first-stop movements in each pair’s `unit-metadata-changes.json`.

## Time and resources

**Timing design differs:** A and B were the original concurrent pair; C ran alone, as requested. These are observed per-unit differences on the same box, not an isolated cost estimate for the record guard or intersection rule. Unit time includes registration, lowering and panic recovery; whole-run time also includes loading, refusal scanning and output. All attempted units have exactly one timing row, and the timing unit populations match.

| Measurement | A | B | C |
|---|---|---|---|
| Timed units | 4032 | 4032 | 4032 |
| Summed unit wall seconds | 981.694305 | 971.800246 | 931.511542 |
| Summed unit CPU seconds | 1048.981864 | 1031.210375 | 1039.195381 |
| Whole-run wall seconds | 1779.622594 | 1772.418276 | 1632.365312 |
| Peak RSS (KiB) | 6220244 | 6473348 | 6704252 |

C−A summed unit wall time: **-50.182763s**, ratio **0.948881**. [Requested 20 largest C−A differences, seconds and ratios](vs-a/timing-top20.md); [all units TSV](vs-a/unit-times.tsv); [all units including CPU JSON](vs-a/unit-times.json). The [C−B top 20](vs-5e5f57d3/timing-top20.md) and full table are also included.

The 91 unaffected import-declaration controls have median absolute C−A difference **0.011718s**, p95 **0.042960s**, maximum **0.092356s**. Median C/A **0.947**; 5th–95th percentile **0.703–1.266**. They execute no lowered receiver read, destructure or predicate test, but still include registration cost. [Every control](vs-a/timing-controls.md), [A/C and B/C control summaries](timing-noise.json). The original simultaneous pair’s [noise floor](../timing-noise.json) remains available; shifts in these unaffected controls must be considered when reading C’s timing ratios.

Box: same kernel/hostname as the original run, nproc 5, cgroup CPU quota `400000 100000` (four cores), memory limit 16 GiB. [Preflight](preflight.json) records the same-box and all-79-file hash checks. C had the same 5,400-second timeout and 60-second kill grace, and did not reach it. [C wall time, peak RSS, exact command and UTC bounds](c-resource.json).

## Reproduction

- A: `3aec8230abb53ac93b92b24f5e3ae916f9cd36a8` (original ledger and times reused byte-for-byte).
- B: `5e5f57d3b530264cb09eba5712a8a362375aede2` (original ledger and times reused byte-for-byte).
- C: `33796314e5bd17ec159c0c4a92dd633552bf27ab`.
- Intermediate test-only commit: `148e6df3563ca142e6bda215f12ce5ae3265708b`.
- TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`; A’s original adapted tree is reused, without rebuilding.
- A binary SHA256: `04c72fc31d9fc1a9e3e0b26e1a154465d5d2078892a99b1106e7ad3057339004`.
- B binary SHA256: `1c184a82c8964821b06aeafb789f2df88f0601f19f145f7e2d9ecf9f78761843`.
- C binary SHA256: `3c262e22d64d75d80d5d3d863a1335fab4ef4ca13341e4e36449e640ec6476c6`.

Exact commands: [preparation](preparation.sh), [prescribed overlay/build](build.sh), [alone run](run.sh), [both census-delta comparisons and validator](compare.sh). [Compiler history](compiler-history.txt) and [B→C compiler diff](compiler-diff.patch) document the tested revisions. A’s and C’s stage3 trees are identical. The worktree shared the same pinned setup dependencies; no internal/ source was edited. Only this evidence subdirectory is added to the branch.

`c.jsonl.gz` and `c-times.jsonl.gz` archive the raw C streams; A/B streams remain in the parent directory. All gzip streams are CRC/roundtrip checked and their raw hashes recorded in [provenance.json](provenance.json); `SHA256SUMS` covers the new artifacts. To rerun compare.sh after packaging, first `gzip -dk c.jsonl.gz`; analyze.py can read compressed streams directly. [Validation](validation.json), [A→C delta](vs-a/census-delta.json), [B→C delta](vs-5e5f57d3/census-delta.json), logs and all derived tables are retained.
