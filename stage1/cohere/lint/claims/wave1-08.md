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
