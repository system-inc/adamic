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

## Third batch claim

After completing and pushing the six existing claims (590f4fa6), all origin
heads were fetched and every origin claim document and base/main port inventory
was inspected again. The first three unclaimed, unported entries of the
combined by-volume ranking are:

1. `no-global-assign`
2. `no-implicit-globals`
3. `no-implied-eval`

All three have zero recorded compiler and repository findings. They are now
reserved for this branch. This claim is pushed before any implementation code.
The fetched ref SHAs and selection inventory are in
validation-wave-22-third/selection.json.

## Fourth batch claim

After completing, testing and pushing all nine earlier claims through a7e33d0c,
all origin heads were fetched and every origin claim and base/main port
inventory was scanned again. The first three available entries in the combined
by-volume ranking are:

1. `prefer-numeric-literals`
2. `prefer-object-has-own`
3. `prefer-object-spread`

All have zero recorded findings on both frozen corpora. These three are now
reserved for this branch. The claim is pushed before implementation; fetched
ref SHAs and the full selection are in validation-wave-22-fourth/selection.json.

## Fourth batch completion

All three fourth-batch rules are ported as .a, compared byte for byte with Go
on positive controls and both frozen corpora, tested with per-rule repair
mutants and sanitizers, and committed with WAVE_22_FOURTH_REPORT.md. All twelve
claims on this branch are complete. No additional rules were claimed.

## Fifth batch claim

All twelve earlier rules were completed, tested and pushed through 7f3b9262
before fetching every origin head again. After excluding base/main ports and
names in every origin Markdown claim, the first available rules are:

1. `react-hooks/set-state-in-effect`
2. `react-hooks/set-state-in-render`
3. `react-hooks/static-components`

These zero-volume rules are reserved for this branch. This claim is pushed
before implementation. The fetched refs and inventory are recorded in
validation-wave-22-fifth/selection.json.
