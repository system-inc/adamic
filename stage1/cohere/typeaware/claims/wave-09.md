# Type-aware wave 09

Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.
Branch: codex/typeaware-wave-09.

Claimed rules, positions 25 through 27 in descending combined compiler and
repository counts from VOLUME_REPORT.md's validation-volume tables, excluding
the 26 rules already ported on the base, with lexical ties:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 25 | prefer-const | 6 | 1 | 7 |
| 26 | radix | 4 | 3 | 7 |
| 27 | @typescript-eslint/non-nullable-type-assertion-style | 4 | 0 | 4 |

All fetched origin branches were inspected for ports and claims before this
commit. None of these three has an existing stage 1 port or claim. Inventory
entries and config set names are not implementations. No rules were skipped.

## Continuation claim, October 7

The original three rules are implemented, tested and pushed in 03ea146b, with
validation logs in d0e9ba7c. Fetched all origin heads again before selection:
325 remote refs, 33 distinct Markdown claim blobs under typeaware/claims.
Excluded native ports on origin/main and origin/codex/tsgo-c-library, including
abbreviated TypeScript rule names, and every ranked rule named in an origin
claim file. Inventory/config references are not implementations.

The first three remaining rules by combined compiler plus repository volume,
with lexical ties, are reserved on this same branch:

- `nexus/correctness-no-process-exit-after-output` (0 compiler, 0 repository)
- `nexus/correctness-no-uncleared-race-timeout` (0 compiler, 0 repository)
- `nexus/correctness-require-blocking-standard-streams` (0 compiler, 0 repository)

This claim update is pushed before code. New sources use .a. Changes stay in
this unit's rule directories; shared harness and registration generator are
owned by codex/lint-harness-dot-a and will not be edited.

## Continuation withdrawn after concurrent-claim refresh

A second all-head fetch found concurrent claims for all three continuation
rules. Wave 02's claim 958c5c2f was committed at 2026-10-07T00:59:43Z,
before this branch's 24667c85 at 2026-10-07T01:00:59Z. Wave 09 yields all
three continuation reservations and takes no replacements. The original
prefer-const, radix and non-nullable-type-assertion-style ports remain complete.

Work performed before the collision was observed is saved as an explicitly
withdrawn experiment in ../validation-wave-09-continuation/withdrawn-experiment.tar.gz, with no
shared registration edits.
The timeout port passed independent full-byte controls and both corpora using
a temporary checker registration overlay. The other two rules are partial,
not completed ports. See ../WAVE_09_CONTINUATION_REPORT.md for gaps and evidence.
