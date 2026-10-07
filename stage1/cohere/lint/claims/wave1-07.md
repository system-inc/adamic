# Lint wave 1 slot 07 claim

Branch: `codex/lint-wave1-07`.

Rules, positions 19 through 21 in the ordered measured handoff in
`origin/codex/lint-helpers:stage1/cohere/lint/HELPERS.md`:

19. `no-underscore-dangle`
20. `no-unsafe-negation`
21. `no-unsafe-optional-chaining`

The helper REPORT.md refers to HELPERS.md for the ordered names; REPORT.md itself
contains no ordered rule list. No implementation was found for these names in
stage1 lint source on main or the 222 fetched origin refs. Frequency inventory
logs mention the names but are not ports. None is skipped as already ported.

This claim precedes implementation. Registration merged into current main
`d090af531216ddd3c25a0dede6b82d7c0a6edf76`. The subsequent helpers merge produced
six shared-file conflicts and was aborted, leaving the registration merge intact.
`docs/parallel-work.md` is absent on main and both foundation branches. The unit
is blocked on a compatible foundation and the missing parallel-work contract;
no rule implementation or parity claim is made by this reservation.

## October 7 continuation

The previous branch contents were pushed before selecting more work.
Fetched all origin heads without recursive submodule fetching: 297 remote refs.
Checked every claims-directory Markdown/JSON blob on those refs, including
reports, and main's executable rule selections. The first 45 helper-ready names
are already claimed; the remaining name is helper-ready position 46. After that,
the first eligible syntax-only entries in inventory.json are the two below.

New rules, claimed before implementation:

1. `structure/tailwind-no-physical-direction`
2. `@eslint-community/eslint-comments/require-description`
3. `@next/next/google-font-display`

Main snapshot: `ef3d907ecdc4c771b016f7d9c52372def057a340`.
Helper snapshot: `6769b88e` (REPORT.md delegates names to HELPERS.md/readiness.json).
Inventory snapshot: `73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf`.
Syntax-only means neither checker/type information nor binding required; the
inventory's original order is preserved, including entries waiting on helpers.
No new rule code precedes this claim push. Prior reservations remain unchanged.

## Next continuation, October 7

Pushed existing work (everything up to date), then fetched all 320 origin refs.
The original 46 helper-ready names are claimed. Scanning actual claim Markdown
on every origin branch and main's source leaves these first syntax-only entries:

1. `@typescript-eslint/no-non-null-asserted-optional-chain`
2. `@typescript-eslint/no-non-null-assertion`
3. `@typescript-eslint/no-this-alias`

Inventory evidence JSON that enumerates a rule with `claims: []` is not a claim;
these names occur only in such selection ledgers, not reservations. Main remains
`ef3d907ecdc4c771b016f7d9c52372def057a340`. This reservation is pushed before code.
