# Type-aware lint wave 30

Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Assigned positions in the combined compiler plus repository ranking linked by
VOLUME_REPORT.md, descending count with lexical ties, excluding the 26 existing
ports before counting positions:

| Position | Rule | Combined findings |
| --- | --- | ---: |
| 88 | nexus/concurrency-no-lost-update | 0 |
| 89 | nexus/consistency-no-iso-string-date-cut | 0 |
| 90 | nexus/correctness-no-callback-in-parse-try | 0 |

The volume counter has 197 entries. Twenty-five of the existing ports occur in
that counter; method-signature-style is not checker-dependent in the registry
and does not occur there. This leaves 172 ranked candidates.

Fetched all origin heads and checked stage1 source paths and claim documents
before this claim. No conflicting port or claim was found for these rules.
No implementation was written before this claim commit.

Completion status: the ISO-date-cut and callback-in-parse-try rules are implemented.
The lost-update rule is unimplemented and its claim is released for reassignment.
See ../WAVE_30_REPORT.md for evidence and limits.
