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
