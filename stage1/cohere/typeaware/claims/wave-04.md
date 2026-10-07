# Type-aware wave 04

Branch: `codex/typeaware-wave-04`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 10, 11 and 12 after excluding the 26 base ports
from the combined compiler/repository ranking referenced by VOLUME_REPORT.md:

| Position | Rule | Combined findings |
| --- | --- | ---: |
| 10 | nexus/consistency-require-constant-casing | 84 |
| 11 | nexus/consistency-require-matching-return-type | 74 |
| 12 | @typescript-eslint/strict-void-return | 56 |

Counts come from validation-volume/compiler-all.counts and repository-all.counts,
sorted by descending combined count with lexical ties. Both oracle suites define
the excluded base ports. All origin heads were fetched and inspected before
this claim. Matches in inventories and count logs are not implementations.
No existing port or claim was found for these three rules. No rules were skipped.

New Adamic source files use `.a`. Checker extensions, if needed, have separate
files per question, with only one-line registration changes in shared files.

## Continuation claim

Original three rules completed and pushed through `67b15e35`.
Fetched all origin heads (325 remote references) on October 7, 2026.
Excluded the 25 checker-dependent base ports (the 26th is not in the checker
ranking) and every rule named in claim Markdown on any origin branch (98).
Combined counts descend, with lexical ties; all positive-volume candidates
are ported or claimed. The first three available entries are:

- nexus/correctness-no-process-exit-after-output (0)
- nexus/correctness-no-uncleared-race-timeout (0)
- nexus/correctness-require-blocking-standard-streams (0)

These rules are claimed for the continuation. This claim is committed and
pushed before any implementation. New code stays in this worker's directories;
the shared registration generator and existing harness will not be edited.
