# Container fit census (zrnkpdz): latent pieces in progress

Statecopy fixed-array support is pushed at `62ddd47969e5168e5a0e74a23498b85d144b2361`, one commit on 1197ee88. Its broad-array mutant fails the guard, the full latent test subtree passes, and all three original/selector meter binaries compile. The selector admission audit passes. containerFits remains an ordinary owned map in every snapshot, not a programFact.

The frozen plan in evidence/plan covers 79 files and 4,557 units exactly once. utilitiesPublic is one piece; utilities.ts uses disjoint unit chunks as in the reused tooling. Every other file belongs to exactly one file group. Each child has a 25 minute ceiling plus at most 10 seconds for termination, and each piece is committed and ordinarily pushed before another starts.

Base compiler: d4c6b07a. Head: 62ddd4796. Logging head: 29de73556, the same patch on f913c480. All use the fixed meter tooling. The single adapted tree and source-manifest.json are unchanged.

Partial observations are under evidence/latent/. First-stop classifications, regressions, cap conclusions, log distributions and common-unit time sums remain pending until the planned paired pieces finish.

The prior stopped report is preserved in STOPPED-REPORT.md. The completed ordinary 79-root runs from that report remain valid: the new commit changes meter state generation only. Their first stops all matched, with no ordinary regressions or cap rows; walls were 780.291357 seconds base and 726.362797 seconds head, one sample on this box.

## Public-file budget revision

The initial base utilitiesPublic whole-file p00 reached its 25 minute ceiling at 1,500.120056 seconds. It was intentionally stopped with SIGQUIT and committed/pushed at 258190b0c; this is not a measured compiler panic. Its header-only findings and partial timers are excluded from comparisons. evidence/public-budget-stop.json records replacement public chunks sized from observed timers. Each replacement public chunk will receive the head/log byte check. The requested single whole-file check could not finish inside the piece budget.

## Completed q00 checkpoint

Base, logging head and unlogged head q00 each completed 80 declarations and 74 timed attempts. Main pair reports five changed first stops, all moving from the Map/Set fit stop to checked predicate membership over an open tag domain. There are zero q00 regressions and zero cap rows. The 223,349-byte head/log record streams are byte-identical; evidence/byte-checks/q00.json records the SHA-256. These are partial findings, not a whole-census result.

## Live completed-pair index

These are partial paired results. Final cohort tables, full regression lists, walk distributions and common-unit sums wait for the entire partition. Each public chunk also has an exact head/log byte check.

| Piece | Changed first stops | Regressions | Cap rows | Public byte check |
| --- | ---: | ---: | ---: | --- |
| q00 | 5 | 0 | 0 | equal |
| q01 | 0 | 0 | 0 | equal |
| q02 | 1 | 0 | 0 | equal |
