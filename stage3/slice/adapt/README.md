# Temporary adaptation scope

Full-tree adaptations remain under `stage3/adapt/`: 50 (implicit returns),
51 (deferred fallthrough plan, no-op), 60 (factory localSymbol), 61 (factory
 typeExpression), 62 (tracing write), 63 (tracing legend), 64 (parser range read).
These accept the complete upstream tree; 60–64 are the parser worker's edits.

Slice-only adaptations moved here, with names unchanged: 52–59 and 80–86.
52–57, 59, 80–82, 85, 87–88 are the active scanner profile. 58, 83 and 84 are
retired; 86 is a plan without an adapter. Each implemented scanner adapter
requires the declaration-slice manifest. No parser adaptation was moved.

Integration proof from area/stage3 `4ad53a4`: full `stage3/apply.sh` completes.
The default oracle runs all tests: 106,366 pass; its single API snapshot mismatch
is exactly reconstructed by `32-indexed-reads-program/public-api.cjs` from
sanctioned 20/40/70, handoff and host-owner edits. All 60,930 other reference
baselines are byte-identical. No reference was accepted or changed for this run.
Scanner's original 509,014-token comparison and parser's complete dump both have
empty full-tree/slice diffs. The expanded scanner reference is recorded under
`stage3/drivers/scanner/evidence/coverage.json`.
