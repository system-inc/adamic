# Type-aware lint wave 08

Branch: codex/typeaware-wave-08.
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Positions 22, 23 and 24 in the combined by-volume ranking referenced by
VOLUME_REPORT.md, excluding the 26 rules already ported on the base:

22. no-loop-func (compiler 9, repository 2, combined 11)
23. nexus/correctness-no-import-cycle-load-time-read (compiler 10, repository 0, combined 10)
24. @typescript-eslint/no-require-imports (compiler 9, repository 0, combined 9)

All fetched origin stage1 trees were checked for corresponding implementation
filenames and claim Markdown containing these names. No match was found.
No rule is skipped. This claim is committed and pushed before implementation.
New Adamic files use .a.

## Continuation claim

The original three rules are complete and pushed in d0d6310f. After fetching all
origin heads, checked all 33 claim documents across 325 refs and native rule
registrations on origin/main and origin/codex/tsgo-c-library. The combined
compiler/repository ranking has 197 checker-dependent entries; 25 of the base
ports appear in that ranking and 96 entries are named in origin claim files.

The first three neither ported on those two branches nor claimed on any origin
branch are, with lexical ordering for equal combined volume:

- nexus/correctness-no-process-exit-after-output (0)
- nexus/correctness-no-uncleared-race-timeout (0)
- nexus/correctness-require-blocking-standard-streams (0)

Claimed for wave 08 continuation. This update is committed and pushed before
implementation. New Adamic files remain .a; existing shared harness and generator
files will not be edited.

Continuation status: all three ports are complete in the owned native profile.
Normal and ASan/UBSan output matches unmodified Go cohere on 100 upstream process
programs, compiler 77, repository 287 and race controls 22. Each rule has a
compiling byte-only mutant; raw-fact, state, released-handle and full bridge
ownership/sanitizer gates pass. Shared registration/profile integration is left
to the assigned harness worker; an owned scratch overlay links the new raw
questions without editing shared files. See ../wave08-next/REPORT.md.

## Second continuation claim

All six earlier claims are complete and pushed through cfa7115d and a116e599.
Fetched all origin heads again and checked all claim Markdown blobs across 356
origin refs, plus full and shorthand native rule registrations on origin/main
and origin/codex/tsgo-c-library. Of 197 ranked checker-dependent rules, 25 are
already ported and 136 are named in claims; 36 remain eligible.

The first three eligible rules in the combined by-volume ranking, with lexical
ordering for tied zero totals, are:

- react-hooks/globals (0)
- react-hooks/immutability (0)
- react-hooks/no-deriving-state-in-effects (0)

Reserved for this branch. This claim update is committed and pushed before
implementation. Shared harness, generator and dispatcher files stay untouched;
new native implementation files will be .a.
