# Type-aware wave 21

Base: 0d540f413625f016f20fea39761c7b184f335de6.

Selected from VOLUME_REPORT.md's linked all-family count tables, sorted by
combined compiler and repository volume descending, with lexical ties, excluding
the 26 rules already ported on the base:

| Remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 61 | @typescript-eslint/no-misused-promises | 0 |
| 62 | @typescript-eslint/no-misused-spread | 0 |
| 63 | @typescript-eslint/no-mixed-enums | 0 |

These rules are reserved for codex/typeaware-wave-21. No claim files or named
ports were found on fetched origin branches before this claim.

## Final status

`@typescript-eslint/no-mixed-enums` is implemented and verified in 63746b1e.
`@typescript-eslint/no-misused-promises` and `@typescript-eslint/no-misused-spread`
are unimplemented. Their reservations are released for reassignment.
This wave is incomplete; see ../WAVE_21_REPORT.md.

## Continuation claim

After fetching all origin heads on October 7, 2026, the next three rules never
named in any origin claim file and not ported on main (ef3d907e) or the bridge
branch (5afbdb83) are reserved for this branch:

| Original remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 91 | nexus/correctness-no-collection-misuse | 0 |
| 92 | nexus/correctness-no-discarded-outcome | 0 |
| 93 | nexus/correctness-no-discarded-pure-result | 0 |

Previously named rules, including explicitly released reservations, were skipped.
This claim is pushed before implementation.

Continuation completion: all three continuation rules are implemented and
verified in c565946d. See ../WAVE_21_NEXT_REPORT.md and
../validation-wave-21-next/ for byte comparisons, mutants, sanitizer results
and timings. No further claims were taken after Ahra's correction.

## Released-reservation-aware continuation

Fetched all 325 origin refs after the instruction to include explicit releases.
Wave 22 now actively reserves no-misused-promises, no-misused-spread and
concurrency-no-lost-update, so those released reservations are skipped.
The first three available rules in the same combined-volume ranking are:

| Original remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 97 | nexus/correctness-no-process-exit-after-output | 0 |
| 98 | nexus/correctness-no-uncleared-race-timeout | 0 |
| 99 | nexus/correctness-require-blocking-standard-streams | 0 |

No matching native implementation on origin/main or origin/codex/tsgo-c-library
and no active origin claim was found. These three are reserved for this branch.
This claim update is pushed before writing implementation code.

Released-reservation-aware continuation status: all three new reservations
remain unimplemented. Work stopped at missing whole-program source/module
resolution facts under Ahra's restriction on shared-file edits. See
../WAVE_21_RELEASED_REPORT.md and ../validation-wave-21-released/ for the live
checker control, explicit refusals, probe mutants and Go production tests.
No further rules were claimed.

Released-reservation-aware continuation completion: all three rules are now
implemented and validated. See ../WAVE_21_PROCESS_REPORT.md and
../validation-wave-21-process/. The previous missing-program-facts blocker was
resolved with isolated raw bridge questions and three dispatch arms.

## Fourth batch claim

All previously reserved rules are implemented, tested and pushed through
880f6f6f. Fetched every origin head, audited 347 origin refs and 33 claim
Markdown files, and checked native sources on main ef3d907e and bridge 5afbdb83.
Explicitly released promises, spread and lost-update reservations now have
active wave-22 claims and were skipped. The first three available rules in
combined descending volume, with lexical ties, are reserved here:

| Full ranking position | Rule | Combined findings |
| --- | --- | ---: |
| 146 | no-obj-calls | 0 |
| 147 | no-object-constructor | 0 |
| 148 | no-promise-executor-return | 0 |

None is ported on either base branch or named in any fetched origin claim.
This claim update is pushed before implementation.

Fourth batch implementation and validation: all three native rule classes are
ported with complete finding/fix/suggestion agreement on supported controls and
both frozen corpora, per-rule mutants and released/sanitizer checks. The object-
constructor rule remains blocked on two JSX inputs and one keyword-label input
in the shared parser; its rule and suggestion logic are complete. See
../WAVE_21_CORE_REPORT.md and ../validation-wave-21-core/ for exact refusals.
These reservations are retained, not released. No subsequent batch was claimed.


## Fifth batch claim

Previous native rule implementations and validation are pushed through 6af00212.
The three object-constructor inputs blocked by the shared parser are explicitly
recorded in WAVE_21_CORE_REPORT.md; all remaining rule logic is ported.
Fetched all 389 origin refs and inspected 33 Markdown claim documents and
native source ports on main e011f8f6 and bridge 5afbdb83. Explicit releases
were inspected: exhaustive-deps has active wave-12 and wave-27 claims;
core prefer-promise-reject-errors has active continuation claims;
promises, spread and lost-update have active wave-22 claims.
The first three available entries in the combined descending-volume ranking,
with lexical ties, are reserved here before writing code:

| Full ranking position | Rule | Combined findings |
| --- | --- | ---: |
| 168 | react-hooks/set-state-in-effect | 0 |
| 169 | react-hooks/set-state-in-render | 0 |
| 170 | react-hooks/static-components | 0 |

None is ported on either base branch or named in any fetched origin claim.
This claim update is pushed before implementation.
