# Historical blockers resolved on wave 2

This note supersedes the parking status from the earlier nine-rule landing.
The recovered parser on base 95968dd93ad0876245f181af63931c3134e14b3d
restores `@typescript-eslint/no-unnecessary-parameter-property-assignment` to
this landing branch. Its original reproducer remains at
`parked/constructor-parser.ts.txt`; the active regression witness is
`../typescript-eslint-no-unnecessary-parameter-property-assignment/testdata/modified-object-binding.ts.txt`.
All 92 upstream cases are included, without filtering the recovery case.

The fix-budget rejection for `no-lonely-if` was resolved by the shared harness.
Its original reproducer remains at `parked/fix-budget.ts.txt`, and its unchanged
eleven-level active witness remains registered.

Neither reproducer was deleted or weakened. This branch registers all ten
owned rule descriptors. Earlier reports record earlier commits and are
historical; WAVE2_LANDING_REPORT.md records this landing's checks.
