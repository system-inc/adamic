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

## Stream completion checkpoint

All three continuation algorithms are now written in owned .a files. The timer
rule remains covered by WAVE_19_TIMEOUT_REPORT.md; both stream rules now have
native control-flow analysis, independent unchanged-Go comparisons, per-rule
byte-only mutants and normal/sanitized corpus checks. Raw declaration ancestry
also supports explicit alias resolution and safely handles binding-pattern names.
Production integration still requires the three dispatch lines in the unapplied
wave_19_next_registration.patch. The unmodified archive explicitly refuses each
question. No shared harness, generator or dispatch file was edited in this
checkpoint, and no further rules have been claimed. See WAVE_19_STREAM_REPORT.md
in the parent directory for the final commands, evidence and timing.

## Second continuation claim

After pushing all three Nexus algorithms and their parity, sanitizer and mutant
proofs through 8845163bf64cad8c62a06ab5f83ab3141b91080a, fetched all origin heads.
The snapshot has 389 origin refs, 40 unique text claim blobs, 197 ranked rules,
26 named base/main ports and 142 rules named in claim files. Thirty candidates
remain after those exclusions. The first three by combined volume descending
and lexical ties are reserved for this branch before any implementation:

- react-hooks/set-state-in-effect (compiler 0, repository 0)
- react-hooks/set-state-in-render (compiler 0, repository 0)
- react-hooks/static-components (compiler 0, repository 0)

The previous Nexus production registrations remain pending under the shared-file
restriction; their algorithms and overlay-based checks are already pushed.

## React overlap discovered on refresh

The later all-head refresh includes wave 16 claim commit 58d204d3 (02:30:09 UTC), earlier than our f2e92f7b claim commit (02:31:23 UTC). Its origin branch claims the same three React rules. Skip react-hooks/set-state-in-effect, react-hooks/set-state-in-render and react-hooks/static-components under the instruction to skip rules claimed on another origin branch. Our second-continuation reservation is superseded; wave 19 has no React implementation to merge. No further rules were claimed. The current React HIR prerequisite status is recorded in wave_19_react/RESUME_REPORT.md.
