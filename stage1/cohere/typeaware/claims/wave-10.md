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

Continuation status: the timeout rule is ported and independently validated;
process-output and blocking-streams remain partial. The runtime-context bridge
question is now registered. The timeout oracle passes 37 controls, compiler77
and repository287, including sanitizer runs, a native decision mutant and
released-handle checks. Separate native state/index/timer helpers match Go;
the combined native probe still stalls in internal/fresh (45s, exit 124).
Process-output still lacks CFG/root/callee analysis; blocking-streams still lacks
CFG/load-time/call analysis and a complete runner. Neither is counted as a
completed rule. See ../wave_10_next/README.md for commands, evidence and limits.
These claims remain reserved; no additional rules were claimed.
