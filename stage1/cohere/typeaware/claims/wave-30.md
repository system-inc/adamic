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

## Continuation claim

Fetched every origin head after pushing 333b5ccd. Selected the first three in
VOLUME_REPORT.md's combined compiler/repository checker ranking that are neither
ported on origin/codex/tsgo-c-library or origin/main nor named in any origin
branch's stage1/cohere/typeaware/claims files. The ranking has 197 checker rules;
25 of the original 26 ports are checker-dependent, and claim files mention 92
ranked rules. All higher-ranked remaining candidates are claimed.

- nexus/correctness-no-collection-misuse (combined volume 0)
- nexus/correctness-no-discarded-outcome (combined volume 0)
- nexus/correctness-no-discarded-pure-result (combined volume 0)

These three are claimed for this continuation. No implementation precedes this
claim commit and push. The earlier released lost-update claim remains released.

Continuation completion: all three rules above are implemented in a74cf501,
with independent Go byte agreement, per-rule mutants, sanitizer runs and released
handles verified. Evidence is in ../WAVE_30_NEXT_REPORT.md and validation-wave-30-next.
No further rules were claimed after Ahra's correction.

## Third batch claim

Fetched all origin heads after verifying b64a21b3 was pushed. The first three
remaining rules in VOLUME_REPORT.md's combined checker ranking, excluding ports
on origin/main and origin/codex/tsgo-c-library and claims on every origin branch:

- nexus/correctness-no-process-exit-after-output (combined volume 0)
- nexus/correctness-no-uncleared-race-timeout (combined volume 0)
- nexus/correctness-require-blocking-standard-streams (combined volume 0)

The ranking contains 197 checker rules, 25 base ports and 96 distinct claim mentions at
this selection. No implementation precedes this claim commit and push.

Third batch status: uncleared-race-timeout is implemented in 80e56233 with Go
byte agreement, a caught range mutant, sanitizer checks and released-handle checks.
The process-exit and blocking-streams rules remain incomplete and claimed: only
their reusable catch-sensitive output state is ported and independently tested.
Their native control-flow event graph, resolved callee declarations and complete
program module-resolution facts are absent. No shared harness or registration
generator was edited. No subsequent batch was claimed. The reproducible selection
snapshot uses full rule-token boundaries (96 distinct claims; the initial loose
substring scan counted 98) and selects the same three rules.
See ../WAVE_30_THIRD_REPORT.md for exact evidence and remaining work.
