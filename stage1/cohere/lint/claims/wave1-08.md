# Lint wave 1 slot 08

Branch: codex/lint-wave1-08.

Claimed rules, positions 22, 23 and 24 in the ordered measured handoff in
HELPERS.md (helpers/REPORT.md links to that list):

22. no-useless-computed-key
23. react-hooks/gating
24. react/forbid-foreign-prop-types

All origin branches were fetched and searched before this claim. No existing
stage1 implementation of these rules was found. None is skipped.

The helpers merge conflicted in six lint integration files. The registration
versions were retained to preserve directory registration. Helper additions were
retained. docs/parallel-work.md is absent from main and both foundation branches.

This claim is committed and pushed before any rule implementation.

## Next three rules, October 7

After pushing all existing work and fetching every origin branch, claim:

1. structure/tailwind-no-physical-direction (last available helper-ready rule,
   position 46 in HELPERS.md linked by helpers/REPORT.md).
2. @next/next/no-assign-module-variable (first available syntax-ready inventory rule).
3. @typescript-eslint/default-param-last (second available syntax-ready inventory rule).

Selection excludes rules ported on origin/main and rules named in claim files
under stage1/cohere/lint/claims on any fetched origin branch. Main snapshot:
ef3d907e. Inventory ordering is origin/codex/lint-inventory inventory.json's
syntax ready for AST/API adaptation cohort. Existing claims remain reserved.

This update is pushed before writing code for the next three rules. New Adamic
source uses .a. Shared compatibility work, if needed for verification, is applied
only to scratch overlays, with default integration limitations reported explicitly.

## Third claim, October 7

After pushing f81b353a and fetching all origin heads, claim the first three
available rules in the inventory syntax-ready cohort (positions 9, 10 and 11):

1. @typescript-eslint/no-unnecessary-type-constraint
2. @typescript-eslint/prefer-as-const
3. @typescript-eslint/prefer-enum-initializers

All 46 rules in the original helper-ready handoff linked by helpers/REPORT.md
are claimed. Scan covered 320 origin refs and all actual claim documents under
stage1/cohere/lint/claims, excluding evidence and reports from claim assertions.
Main ef3d907ecdc4c771b016f7d9c52372def057a340 has none of these three rules.
The first eight syntax-ready entries are claimed by slots 08 and 14. The list
ordering comes from origin/codex/lint-inventory inventory.json, not a fresh
readiness expansion. Existing reservations remain in place.

This claim is committed and pushed before implementation. Rule sources use .a.
Shared integration compatibility stays in scratch overlays and will be reported.
