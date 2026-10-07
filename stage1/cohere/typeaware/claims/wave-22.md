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

## Continuation claim

After pushing all original wave 22 work, all origin heads were fetched again.
The first three available checker-dependent rules in the same combined-volume
ranking, excluding ports on origin/codex/tsgo-c-library and origin/main and
active claims on any origin branch, are:

1. `@typescript-eslint/no-misused-promises`
2. `@typescript-eslint/no-misused-spread`
3. `nexus/concurrency-no-lost-update`

The first two reservations were explicitly released in origin/codex/typeaware-wave-21's
claims/wave-21.md. The third was explicitly released in origin/codex/typeaware-wave-30's
claims/wave-30.md. Released reservations are available for reassignment; no
other active claims were found. All three have zero recorded corpus findings.
These three rules are now reserved for codex/typeaware-wave-22. This continuation
claim is pushed before writing any implementation for them.
