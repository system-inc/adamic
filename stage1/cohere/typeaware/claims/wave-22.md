# Type-aware lint wave 22

Branch: `codex/typeaware-wave-22`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 64, 65 and 66 (one based):

1. `@typescript-eslint/no-non-null-asserted-nullish-coalescing`
2. `@typescript-eslint/no-unnecessary-qualifier`
3. `@typescript-eslint/no-unused-private-class-members`

Ranking uses VOLUME_REPORT.md's referenced compiler-all.counts and
repository-all.counts, combined descending volume with lexical ties, excluding
all 26 rules already ported on the base. The registry count table contains 197
checker-dependent rules; method-signature-style is one of the 26 ports but is
absent from that checker-dependent table. All three selected rules have zero
recorded findings on both corpora.

Before this claim, all origin branch refs were fetched and inspected for claims
and type-aware implementations of these names. None was claimed or ported; no
rule was skipped. New Adamic implementation files will use `.a`.
