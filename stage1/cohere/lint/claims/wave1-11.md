# Lint wave 1 slot 11

Branch: codex/lint-wave1-11.

The helper report points to the ordered list in ../HELPERS.md. Assigned positions:

31. `adamic/no-definite-assignment`: claimed for a new directory port.
32. `base/boundary-no-global-container`: skipped, already ported on origin/codex/stage1-lint-batch4 in base_boundary_no_global_container.ts.
33. `base/consistency-no-hand-built-declared-error`: skipped, already ported on origin/codex/stage1-lint-batch4 in base_consistency_no_hand_built_declared_error.ts.

Existing-port evidence: batch 4 implementation 7bd94a2f92cc17d12bb69bbd78bfe2c4ac4ca63d and its BATCH4.md report. This claim is pushed before new rule code.

Foundation merge note: helpers conflicts with registration in six shared files. Resolution retains directory registration and imports the helper package, inventory and HELPERS.md; it does not restore the helper branch's legacy shared dispatcher. docs/parallel-work.md is absent in main and both foundation branches; docs/lint-registration.md provides the available contract.

## Next three rules, October 7

After pushing all previous work and fetching every origin head, the first available
helper-ready rule is position 46. Positions 1 to 45 are named in claim files on
origin/codex/lint-wave1-01 through origin/codex/lint-wave1-15. The helper report
links the ordered names in HELPERS.md.

Claimed next, in queue order:

1. `structure/tailwind-no-physical-direction` (helper-ready position 46).
2. `@eslint-community/eslint-comments/require-description` (first syntax-only inventory entry).
3. `@next/next/google-font-display` (next syntax-only inventory entry).

Selection uses inventory.json on origin/codex/lint-inventory and excludes names
in claim files under stage1/cohere/lint/claims on every fetched origin branch.
None of these three has an executable selection site in origin/main's lint
sources. All three are reserved here before any new implementation code.

Existing .a registration limitations remain explicit. Any validation workaround
will be confined to owned files and scratch overlays; no shared production code
or lists will be edited.

Follow-up handoff: .a candidates exist for the three new claims. Their Go findings
and unchanged fixes match on the complete compiler and stage1 corpora with independent
Adamic parsing, plus all captured upstream tests through Go AST projection and 161
non-JSX upstream cases through Adamic parsing. Production .a registration, 18 native
JSX upstream cases, malformed options and suppression integration remain uncompleted.
See ../rules/structure-tailwind-no-physical-direction/REPORT.md. These names remain
reserved; this is not a declaration that the full requested port bar is met.

## Next three after finishing the claimed work

All owned implementation and validation for the four earlier claims was pushed at 2ec28ca6. The remaining shared JSX and suppression gaps are recorded in the owned report.

A fresh fetch covers 335 origin refs and 52 claim documents; main is ef3d907ecdc4c771b016f7d9c52372def057a340. All 46 helper-ready names are already claimed. The earlier tentative TypeScript-rule selection was taken by wave1-05 before this fetch and is skipped. The first three eligible syntax-only inventory names are now reserved here:

1. `base/security-require-context-access`
2. `better-tailwindcss/enforce-canonical-classes`
3. `better-tailwindcss/enforce-consistent-class-order`

This claim update is pushed before any implementation of these three rules. No additional rules are claimed.

## Next syntax-only batch after pushed shared-gap reports

Owned work is pushed at ca2a9d5a. The two Tailwind rules remain explicitly blocked by the shared project CSS provider and multiple-fix serialization; no complete-rule claim is made. The refreshed audit covers 356 origin refs and 54 claim documents. Helper-ready rules are exhausted. The next three available syntax-only inventory entries are reserved before implementation:

1. `nexus/consistency-no-screaming-snake-case`
2. `nexus/import-no-forbidden-source`
3. `nexus/import-require-path-alias`

This claim update is pushed before writing rule code.

## Dedup handoff, harness 41eb6eab2

The published ledger supersedes our three duplicate reservations: physical-direction and require-description belong to wave1-05; Google-font-display belongs to wave1-01. Our losing copies are removed. Five unique/winning rule ports remain, with owned comparison code and current evidence under ../rules/adamic-no-definite-assignment/DEDUP_REPORT.md. The moved LEGACY reports/evidence preserve the old scope without claiming the retired rules. Shared witness options and one parser refusal remain explicitly parked under #zmh9v36. No new batch rule or helper is claimed.
