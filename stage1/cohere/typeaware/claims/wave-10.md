# Type-aware wave 10 claim

Branch: `codex/typeaware-wave-10`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Positions use the combined compiler and repository counts linked by
VOLUME_REPORT.md, sorted by descending count with lexical ties, excluding the
26 rules already ported on the base branch.

| Remaining position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 28 | no-unmodified-loop-condition | 4 | 0 | 4 |
| 29 | @typescript-eslint/no-redundant-type-constituents | 3 | 0 | 3 |
| 30 | @typescript-eslint/prefer-includes | 2 | 1 | 3 |

All origin branches were fetched and their distinct stage1 trees searched before
this claim. No implementation or claim for these three rules was found. Existing
inventory entries, count registrations, configuration and skip-types evidence do
not constitute ports. No rule was skipped.

## Continuation claim

The original three ports are completed and pushed in 8cfa5d02. After fetching all
origin heads, the next three unclaimed checker-dependent rules in the combined
VOLUME_REPORT.md full count tables are:

- nexus/correctness-no-process-exit-after-output (combined volume 0)
- nexus/correctness-no-uncleared-race-timeout (combined volume 0)
- nexus/correctness-require-blocking-standard-streams (combined volume 0)

Selection inspected production Adamic source on origin/codex/tsgo-c-library and
origin/main, and every distinct Markdown claim blob under
stage1/cohere/typeaware/claims on all origin branches. The ranking has 197 checker
rules; 25 base ports occur in it, 96 remaining names are mentioned in claim files,
and 76 candidates remain. These are the first three. No implementation precedes
this claim commit and push. Shared harness and registration generator remain
outside this continuation's edits.

Continuation status: timeout is fully ported and independently validated in
93916207, with a passing native/sanitized corpus rerun after the latest bridge
changes. Process-output and blocking-streams now have complete candidate source
implementations, CFG/callee/load-time logic and separate .a runners. Their source
execution matches Go bytes on 92 controls, with one clean-exit source mutant each.
Their full native builds both time out in internal/fresh at 45s, so neither is
counted as completed: native corpus parity, sanitizers and native timings are
blocked. No shared harness, generator or protected compiler files were edited.
See ../wave_10_next/README.md and evidence/source-ports for evidence and limits.
These claims remain reserved; no additional rules were claimed.
