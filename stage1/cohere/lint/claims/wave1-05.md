# Lint wave 1 slot 05

Branch: codex/lint-wave1-05.

Positions 13, 14 and 15 in the helper handoff list in HELPERS.md, linked from helpers/REPORT.md:

- max-depth: skipped, already ported as max_depth.ts on origin/codex/stage1-lint-batch4.
- max-nested-callbacks: skipped, already ported as max_nested_callbacks.ts on origin/codex/stage1-lint-batch4.
- no-cond-assign: claimed for this worker, under rules/no-cond-assign/.

The registration foundation merged cleanly. The helper foundation conflicted in six shared lint files; the directory registration versions were retained, with helper-owned files brought in. docs/parallel-work.md is absent; docs/lint-registration.md supplies the registration contract.

## Continuation claim, October 7

Fetched all origin heads without recursive submodule fetching. Checked direct
claim Markdown files on all 297 origin refs and main's stage1 lint tree.
The first eligible helper-ready rule, then the first two eligible syntax-only
inventory entries in inventory order, are reserved for this branch:

1. structure/tailwind-no-physical-direction
2. @eslint-community/eslint-comments/require-description
3. @next/next/google-font-display

The inventory syntax-only queue includes both syntax-ready and syntax-waiting-on-
helpers rules; type-aware and binding rules are excluded. These names were not
found in any claim file and no corresponding port was found on origin/main.
This update is pushed before implementation. The earlier no-cond-assign draft
was saved in 85dd293 at the user's request; it is not a certified port.

Read docs/parallel-work.md from origin/codex/no-shared-lists. Shared infrastructure
remains its owner's territory. Compatibility overlays may be used in scratch
for evidence, with any default integration blocker stated separately.

## Second continuation claim, October 7

After pushing all earlier work, fetched all 320 origin refs. Searched all Markdown
under stage1/cohere/lint/claims/ on every origin ref (39 unique blobs), plus main's
lint modules. No helper-ready rule remains eligible. The first three eligible
syntax-only entries in inventory order are reserved for this branch:

1. @typescript-eslint/no-non-null-asserted-optional-chain
2. @typescript-eslint/no-non-null-assertion
3. @typescript-eslint/no-this-alias

These names are neither ported on main nor present in any origin claim Markdown.
This update is committed and pushed before implementation. Earlier implementation
and evidence through 4570161b were already pushed; no shared source is claimed.

## Third continuation claim, October 7

All earlier candidates and blocked-rule work are pushed through 3b58afe4. The
Google-font URL decision and entity decoder match Go on 276 cases and have a
compiling semantic mutant caught on all three runtimes; full rule execution is
blocked by shared JSX parsing. Profile compilation and non-null suggestion
serialization remain shared-harness gaps; no shared source is edited.

Fetched all 329 origin refs. Checked 52 unique claim Markdown blobs recursively,
main's descriptors and source modules. No eligible helper-ready entry remains.
The first three eligible syntax-only inventory entries are now reserved here:

1. @typescript-eslint/prefer-function-type
2. @typescript-eslint/prefer-namespace-keyword
3. @typescript-eslint/triple-slash-reference

This claim is pushed before implementation. The .ts fallback explicitly granted
by Ahra remains in effect while the shared .a harness branch is unavailable.

## Fourth continuation claim, October 7

All earlier owned implementations and evidence are pushed through 022fcef4.
Fetched every origin head explicitly with refs/heads/*:refs/remotes/origin/*,
then checked all 341 origin refs and 52 unique recursive claim Markdown blobs,
as well as main's lint sources and descriptors. No helper-ready rule remains
eligible. The first three eligible syntax-only inventory entries are reserved:

1. better-tailwindcss/no-unnecessary-whitespace
2. boundaries/dependencies
3. complexity

This claim is committed and pushed before implementation. The current checkout's
shared harness still needs its owner's published .a/profile/edit integration;
Adamic sources may use Ahra's explicit .ts fallback while that is pending.
