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
