# Step 09 counts

Pinned main: 45487a809f89885a3fc651cd590e7dabf31362dc.

- 3 witness `.a` programs; 3 Node baselines match recorded stdout/stderr/exit.
- 3 expected main refusals; 0 native binaries or native output comparisons.
- 3 source mutants caught by Node baseline comparison; 0 native mutant runs.
- 3 ledger mutants caught by independent API/main-proof reconciliation.
- 79 TypeScript roots, 82 regular closure files; 18 bridges (5 unknown, 13 any).
- 651 predicate annotations: 580 with bodies (3 Proven, 577 Refused), 71 without bodies (Refused).
- 18 virtual direct-cast replacements: all Refused.
- 3,011 property stores preserved; 49 candidates, incomplete runtime expando classification.
- 9 descriptor calls / 24 names; 1 existing-object assign; 1 Function-valued expression.

Native statement/IR/allocation counts do not exist for these refused witnesses.
