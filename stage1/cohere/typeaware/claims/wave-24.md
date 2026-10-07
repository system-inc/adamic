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

## Third batch after all six prior ports were pushed

Previous batch completed, tested and pushed in `cce6c361a94ed730f5a743ce84deab9f91142b58`.

Next three eligible checker-dependent rules:

1. `prefer-promise-reject-errors`
2. `prefer-regex-literals`
3. `prefer-rest-params`

Fetched 356 origin refs on October 7, 2026, inspected 33 distinct Markdown
claim blobs, and excluded ports on main and the bridge branch and all claims
on every origin branch. Full-ranking positions 158, 159 and 160, each zero
findings in both frozen populations. Main: `ef3d907ecdc4c771b016f7d9c52372def057a340`.
Bridge: `5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`. This update precedes implementation.

No shared registration generator or existing shared test harness will be edited.

## Fourth batch after all nine prior ports were pushed

Previous batch completed, tested and pushed in `09bf947bb040786384d3ef055feb376ea4c0832c`.

Next three eligible checker-dependent rules:

1. `react-hooks/set-state-in-effect`
2. `react-hooks/set-state-in-render`
3. `react-hooks/static-components`

Fetched 389 origin refs on October 7, 2026; inspected 33 distinct Markdown
claim blobs and excluded ports on main/bridge and all claims on all origin
branches. Full-ranking positions 168, 169 and 170, each zero findings.
Main: `e011f8f60899586d6373a5ccb07335ad82cfbf3c`.
Bridge: `5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`.
This update precedes implementation. No shared harness or generator will be edited.

## Fourth batch parked

Status: parked, counted finished for the landing-first cap by Ahra's instruction.

- `react-hooks/set-state-in-effect`: blocked on native high-level IR and capture analysis.
- `react-hooks/set-state-in-render`: blocked on native high-level IR, single-assignment construction and control dominance.
- `react-hooks/static-components`: blocked on native high-level IR and JSX parser integration.

Cohere analysis modules are being ported to Adamic on #dnv6f2c; JSX support is landing on `area/stage1-lint`. These are not completed native rule ports. The Go oracle, native JSX refusal probe, listener metadata and evidence already pushed remain in `wave-24-fourth/`. No shared source changes are made to bypass these dependencies. Nine earlier native ports remain oracle-green on current main `f8013f0baac41ddc340d76f83bddde38536a8f07`.
