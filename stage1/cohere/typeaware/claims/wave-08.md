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
