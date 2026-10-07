# Type-aware wave 28 claim

Branch: `codex/typeaware-wave-28`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

The VOLUME_REPORT.md all-family counts in validation-volume are ranked by
combined compiler and repository volume descending, with lexical ties. Excluding
all 26 ports on the base leaves 172 checker-dependent candidates.

| Remaining position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 82 | base/correctness-require-verify-optional-parity | 0 | 0 |
| 83 | block-scoped-var | 0 | 0 |
| 84 | getter-return | 0 | 0 |

All three are claimed here before implementation. No matching rule source or
claim was found across 268 fetched origin refs (64 distinct stage1 trees).
Counts, catalogs and configuration references are not implementations.
No rule was skipped. Zero-volume corpora require positive controls and mutants.

## Continuation claim

After the original three ports were tested and pushed at `24c600f017352ef3dd8d7c80f104a26099bc4b85`, origin was fetched again. The first three by combined volume, excluding the 26 base ports, main ports, and canonical rule names in claims across all 325 origin refs, are:

1. `nexus/correctness-no-process-exit-after-output` (0 compiler, 0 repository).
2. `nexus/correctness-no-uncleared-race-timeout` (0 compiler, 0 repository).
3. `nexus/correctness-require-blocking-standard-streams` (0 compiler, 0 repository).

There were 33 distinct Markdown claim blobs naming 96 ranked rules at this snapshot. None of these three is claimed or implemented on main or the bridge branch. This continuation claim is pushed before any implementation. Shared registration generators and shared test harness files remain outside this worker's territory.
