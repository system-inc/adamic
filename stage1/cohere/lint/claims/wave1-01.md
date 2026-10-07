# Lint wave 1, slot 01

Branch: codex/lint-wave1-01. Base: origin/main d090af5.

The helper report links the ordered 46-rule handoff in ../HELPERS.md.
Positions 1, 2 and 3 are:

1. @typescript-eslint/consistent-type-assertions: claimed for investigation.
2. @typescript-eslint/consistent-type-definitions: skipped, already implemented in
   origin/codex/stage1-lint-batch4-typescript at
   63782c53678711717f9fd4f12762ce3d103b9a3d,
   stage1/cohere/lint/consistent_type_definitions.ts.
3. @typescript-eslint/init-declarations: skipped, already implemented on that
   branch in stage1/cohere/lint/init_declarations.ts, also on
   origin/codex/stage1-lint-batch4 at d486b03a15f3202acc317dc81c40cc21818e21f7
   in stage1/cohere/lint/ts_init_declarations.ts.

No rule implementation has been written before this claim commit and push.
Existing branch implementations are observations, not parity certifications by
this worker. The remaining rule is not declared ported.

## Outcome

The original investigation report in wave1-01-report.md and wave1-01-evidence/ is historical. Consistent-type-assertions is now implemented and registered in its own .a directory. Its fixtures, structured suggestions, semantic mutant and projected compiler/stage1 corpus pass on all three backends. Independent JSX parsing and the shared harness import failure remain explicit limitations. Current evidence is in ../rules/typescript-consistent-type-assertions/testdata/evidence/REPORT.md.

## Next three rules, October 7

Previous work was already pushed at 4541490 before fetching the new queue.
Fetched main: ef3d907ecdc4c771b016f7d9c52372def057a340.
Every fetched origin branch's stage1/cohere/lint/claims/ Markdown was searched
for these public names; main's stage1 Adamic/TypeScript source was searched for
ports. None of the following names occurs in either search.

1. structure/tailwind-no-physical-direction, position 46 of the original
   helper-ready handoff referenced by helpers/REPORT.md.
2. @eslint-community/eslint-comments/require-description, first remaining
   syntax-only inventory entry (needs_type_information=false).
3. @next/next/google-font-display, next remaining syntax-only inventory entry.

The first 45 helper-ready names already occur in origin claim files. Once that
46-name handoff is exhausted, selection follows inventory.json array order on
origin/codex/lint-inventory. New helper packages are dependencies, not claims of
rule completion. This claim update is pushed before any new implementation.

## Next three after bf9675da

The completed claimed implementations and evidence were pushed at bf9675da before this selection. Main is ef3d907ecdc4c771b016f7d9c52372def057a340. The refreshed audit searched all 340 origin branches and 52 distinct claim blobs. No original helper-ready rule remains available. Following the syntax-only inventory array order, the next three unported-on-main and unclaimed names are:

1. better-tailwindcss/no-deprecated-classes
2. better-tailwindcss/no-duplicate-classes
3. better-tailwindcss/no-unknown-classes

This claim update is committed and pushed before writing any implementation for these three rules.

Outcome of the new selection: all three remain reserved and unported. Investigation stopped at the absent live design-system engine and program context, following Ahra's instruction for blockers beyond the shared harness. Each own rule directory contains a REPORT.md; no rule descriptor was added and no passing parity claim is made.
