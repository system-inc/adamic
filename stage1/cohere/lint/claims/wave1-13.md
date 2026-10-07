# Lint wave 1, slot 13

Branch: codex/lint-wave1-13. Base: origin/main d090af5.
Positions follow the 30 option-ready rules followed by the 16 policy-ready rules
in HELPERS.md, referenced by helpers/REPORT.md.

37. nexus/consistency-no-for-in: skipped, already ported on
    origin/codex/stage1-lint-batch4 in nexus_consistency_no_for_in.ts.
38. nexus/consistency-no-hand-rolled-delay: claimed here.
39. nexus/consistency-no-return-void: claimed here.

This claim is pushed before rule implementation. New Adamic modules use .a.
The foundations conflicted in six lint entry points; directory registration
was retained while incoming helpers and inventory were merged.
The referenced docs/parallel-work.md is absent from main and both foundations;
docs/lint-registration.md provides the available directory ownership contract.

## Continuation claim

Fetched all origin heads on 2026-10-07; origin/main is ef3d907e.
Existing branch work was pushed before selection. Top-level claim Markdown
files on every fetched origin branch were checked, including skipped/reserved
names. All first 45 helper-ready positions are covered by those claims.

46. structure/tailwind-no-physical-direction: claimed here.

The remaining selections are the first unclaimed entries in inventory.json's
`syntax ready for AST/API adaptation` wave, in inventory order:

- @next/next/no-assign-module-variable: claimed here.
- @typescript-eslint/default-param-last: claimed here.

None is ported on fetched main or named by an existing origin claim.
This claim update is committed and pushed before implementation.

## Second continuation claim

Fetched all origin heads on 2026-10-07, main ef3d907e. All 46 helper-ready
names are already covered. Checked 39 distinct claims Markdown blobs on
all origin branches. The first remaining syntax-ready inventory entries,
absent from main selection sites and all branch claims, are:

- @typescript-eslint/no-unnecessary-type-constraint: claimed here.
- @typescript-eslint/prefer-as-const: claimed here.
- @typescript-eslint/prefer-enum-initializers: claimed here.

This claim is pushed before implementation. Full repair records will remain
rule-owned; the shared single-edit/single-suggestion record is insufficient.

## Third continuation claim

Prior candidates and complete repair evidence are pushed through df9f0451.
Fetched all origin heads on 2026-10-07; main remains ef3d907e. Checked 52
unique claims Markdown blobs across 330 origin refs, including reserved names.
All 46 helper-ready rules are covered. The first three remaining syntax-ready
inventory entries, absent from main and from all origin claims, are:

- default-case-last: claimed here.
- for-direction: claimed here.
- guard-for-in: claimed here.

This update is committed and pushed before any rule code is written.

## Fourth continuation claim

Previous three implementations and final evidence are pushed through 115987be.
Fetched 335 origin refs on 2026-10-07; main remains ef3d907e. Checked all
52 distinct recursive claim Markdown blobs and main's rule registrations.
All 46 helper-ready rules are covered. The first three unclaimed syntax-ready
entries are:

- no-constructor-return: claimed here.
- no-delete-var: claimed here.
- no-eq-null: claimed here.

This claim update is pushed before rule implementation.

## Fifth continuation claim

Prior ports and final evidence are pushed through a3f826a4. Fetched 341 origin
refs on 2026-10-07; main remains ef3d907e. All 52 distinct recursive claims
Markdown blobs and main registrations were checked. The helper-ready list is
exhausted. The first three unclaimed syntax-ready inventory entries are:

- no-multi-str: claimed here.
- no-nonoctal-decimal-escape: claimed here.
- no-octal: claimed here.

This update is committed and pushed before implementation.

## Current-main rebase and parking assessment

Replayed owned commits onto c01907a7 while retaining main shared files. All 17 rule witness comparisons and compiling mutants pass with the scratch compatibility harness; the broader supported corpus also passes. Shared registration remains blocked on #zmh9v36. Eleven parser/adapter exclusions remain beyond that harness gap, so no claim that the only-harness parking condition is met. No additional helper is claimed. See rules/nexus-consistency-no-return-void/PARKING_REPORT.md for commands, fresh logs, throughput and exclusions.

## Dedup withdrawals under harness 41eb6eab2

The complete shared DEDUP_LEDGER.md supersedes the historical continuation selections above. Withdrawn and removed here: @next/next/no-assign-module-variable and @typescript-eslint/default-param-last to wave1-15; @typescript-eslint/no-unnecessary-type-constraint, @typescript-eslint/prefer-as-const and @typescript-eslint/prefer-enum-initializers to wave1-08; no-multi-str, no-nonoctal-decimal-escape and no-octal to wave1-15; structure/tailwind-no-physical-direction to wave1-05. Keep only the eight retained rules listed in rules/nexus-consistency-no-return-void/DEDUP_REPORT.md. No new rule or helper is claimed. Two retained Nexus exclusions remain beyond missing shared registration. The older seventeen-rule and eleven-exclusion assessments are historical.
