# Type-aware wave 19

Base: 0d540f413625f016f20fea39761c7b184f335de6.

Selected from VOLUME_REPORT.md's linked all-family count tables, sorted by
combined compiler and repository volume descending, with lexical ties, excluding
the 26 rules already ported on the base:

| Remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 55 | @typescript-eslint/consistent-generic-constructors | 0 |
| 56 | @typescript-eslint/dot-notation | 0 |
| 57 | @typescript-eslint/no-array-constructor | 0 |

These rules are reserved for codex/typeaware-wave-19. No matching claim files
or named ports were found on fetched origin branches before this claim.

## Continuation claim

The original three rules are complete and pushed through c5a19fe0.
Fetched all origin heads before this selection. VOLUME_REPORT.md's linked
all-family tables contain 197 checker-dependent rules, ranked by combined
compiler/repository findings descending with lexical ties. Excluded the base
ports, ports on origin/main and origin/codex/tsgo-c-library, and every rule named
in claim files on any of the 325 fetched origin refs. The first three available
rules are reserved for this branch:

- nexus/correctness-no-process-exit-after-output (combined findings 0)
- nexus/correctness-no-uncleared-race-timeout (combined findings 0)
- nexus/correctness-require-blocking-standard-streams (combined findings 0)

No implementation for this continuation precedes this claim commit and push.
Shared registration generator and test harness remain untouched.

Continuation status: claimed, not implemented. The existing bridge lacks raw
ancestry/global-augmentation, resolved callee declaration and program module-graph
facts; registering new question files requires the shared facts.go dispatch,
which Ahra's correction forbids editing outside owned rule files. Stopped without
editing shared files. See ../WAVE_19_NEXT_REPORT.md for the exact blocker.

Continuation checkpoint f8ef4017: no-uncleared-race-timeout is implemented and
verified using a scratch registration overlay. Its production bridge registration
is still pending. Isolated declaration-ancestry, resolved-callee and program-modules
facts and Adamic decoders are prepared and contract-tested. The other two rule
implementations remain unfinished; no additional rules have been claimed.
The actual production archive's unsupported-question refusal is now tested.
See ../WAVE_19_TIMEOUT_REPORT.md for evidence and the unapplied registration patch.
