Temporary: comes out when bounded indexed reads and dense indexed compound assignments land

Slice-only plan, published before implementation. Require slice.json at the input
root. Never apply to the full adapted upstream tree. Files: core.ts, scanner.ts,
and utilities.ts in createScanner's reached slice only.

Resolve TS2532 and TS2345 with explicit checked reads at the reported bounded
array/string accesses. For compound typed-array assignments, read, combine,
and assign the same value, preserving the index and RHS evaluation order.
No non-null assertions or checker-option changes. Missing values must fail
explicitly rather than silently inventing a value. Document the density/bounds
invariant for each site and decline any site without a justified invariant.

Validate idempotence, stock Node token bytes against all 509,014 baseline tokens,
and the stage 3 baseline oracle before calling the adaptation complete. A mutant
must demonstrate that the checked-read path can reject a missing value. Any
unvalidated adaptation remains a scratch probe and is reported as such.
