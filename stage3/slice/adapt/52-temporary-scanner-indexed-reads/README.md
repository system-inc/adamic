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

## Baseline rejection and revised plan

The first full baseline rejected the guarded generic forEach read: other compiler
callers legitimately pass undefined array entries. Keep their callbacks and
indices exactly as before. Revise that one loop to array.entries() iteration,
which visits undefined entries and holes rather than rejecting them, obtains
the same increasing indices, and observes dynamic array length. The scanner
arrays are append-built, while the full baseline covers undefined entries.
The other bounded guards remain separate. Re-run the complete baseline.

## Revised validation

Full unfiltered upstream baseline passes: 106,367 tests, zero failures, zero
pending, zero baseline differences. Exact slice edits were mapped by recorded
source spans onto complete upstream declarations for validation only; the
adaptation scripts themselves still reject full-tree input. Install, build and
tests all exit 0. Combined wall time 259.105 seconds, nproc 5, four workers.
Node scanner output matches all 509,014 full-tree tokens. Idempotence passes.
The earlier rejected implementations are historical probes, not this result.
See scanner/evidence/member-baseline-revised-report.json and its test log.

Read invariants: Unicode range maps are nonempty even-length static arrays and
binary-search mid is an even range start; mid+1 is its end. Levenshtein loops
bound i and j by the respective string lengths. Uint16 segments are zero-filled
and allocated to hold all input digit bits; nonzero residuals have a following
allocated segment. forEach uses entries to preserve legitimate undefined
values and holes. Clearing the reached Unicode table in a scratch mutant is
caught by the explicit guard (Node exits 1). The token end-offset mutant is
caught by diff. No generic callback read rejects undefined in this revision.

The full-tree indexed-read adaptation may spell a compound assignment target
with a non-null marker. This adapter unwraps that marker when recognizing `|=`
and rewrites the complete assignment, preserving a valid assignment target.
The integrated full-tree-to-slice Node comparison caught and validates this case.
