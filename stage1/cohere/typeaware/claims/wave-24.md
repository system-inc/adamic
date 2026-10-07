# Type-aware wave 24

Branch: `codex/typeaware-wave-24`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed checker-dependent rules, positions 70, 71 and 72 after excluding
all 26 rules already ported on the base:

1. `@typescript-eslint/prefer-return-this-type`
2. `@typescript-eslint/return-await`
3. `@typescript-eslint/use-unknown-in-catch-callback-variable`

Selection uses the combined compiler and repository counts linked from
VOLUME_REPORT.md in `validation-volume/*-all.counts`, descending combined
volume with lexical ties. Each selected rule has zero recorded findings in
both populations. The inventory ranking includes non-checker rules and is
not the checker-only selection used here.

Checked 268 origin refs after fetching all branch refs on October 7, 2026.
No competing claim or named type-aware port for these rules was found.
This claim precedes implementation; it does not assert completed ports.

## Next batch after the original ports were pushed

Original three completed and pushed in implementation commit
`406e7546846aeb58fb918ad4668f09f131a78e1e`.

Next claimed rules, first eligible checker-dependent entries in descending
combined VOLUME_REPORT ranking after excluding ports on main and the bridge
branch and every claim on every fetched origin branch:

1. `nexus/correctness-no-process-exit-after-output`
2. `nexus/correctness-no-uncleared-race-timeout`
3. `nexus/correctness-require-blocking-standard-streams`

Fetched 325 origin refs on October 7, 2026; examined 33 distinct Markdown
claim blobs, the 26 existing ports and main's stage1/cohere implementation.
Main tip: `ef3d907e`; bridge tip: `5afbdb83`. These are full-ranking positions
122, 123 and 124, with zero findings in both frozen populations. None was
claimed or ported on the checked refs. This update precedes implementation.

No shared registration generator or existing shared test harness will be edited.
