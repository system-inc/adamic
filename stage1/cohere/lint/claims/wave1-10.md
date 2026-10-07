# Lint wave 1 slot 10

Branch: codex/lint-wave1-10. Owner: wave1-10.

The helper report delegates its ordered rule list to stage1/cohere/lint/HELPERS.md.
Positions 28, 29 and 30 in that list are claimed here:

28. react/no-unsafe
29. react/self-closing-comp
30. sort-vars

No implementation of these three names was found in main or the fetched origin
branch source trees on October 6, 2026. Helper descriptors and inventory rows
are readiness metadata, not ports. No rule is skipped as already ported.

This claim is committed and pushed before any rule implementation.
New Adamic source files for this unit must use .a, never .ts.

Foundation integration: the registration merge was clean. The helpers merge
conflicted in six shared lint driver files. The registration versions were
retained, with the standalone helper package and inventory added by the merge.

## Continuation claim, October 7

Fetched all origin heads without recursive submodule fetching. origin/main is
 ef3d907ecdc4c771b016f7d9c52372def057a340. Previous branch work is fully pushed.

The helper report points to HELPERS.md. Its measured comment handoff now expands
the original 46 to 62 helper-ready rules; option-gap bullets are not ready rules.
After excluding ports on main and claims on every fetched origin branch, the
first three candidates in measured handoff order are reserved by this unit:

- structure/tailwind-no-physical-direction (original handoff position 46)
- @typescript-eslint/ban-tslint-comment (comment handoff position 1)
- @typescript-eslint/no-invalid-this (comment handoff position 2)

The latest selection criterion excludes ports on main, rather than all branch
ports. ban-tslint-comment has older origin-branch implementations outside main;
these are reference material, not a main port or a claim in claims/.
No new implementation is written before this claim is committed and pushed.
The original three reservations remain as recorded above.
