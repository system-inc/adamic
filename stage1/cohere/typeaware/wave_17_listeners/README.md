Built: numeric rule.json listener declarations for the 15 existing wave-17 claims; no new claims.
Commits: base and validated source remain b825c83b on main e8ba3d5d; this commit changes metadata only.
Commands and outputs: python3 wave_17_listeners/verify.py PASS for all 15 declarations; git diff --check clean.
Mutants: replacing SourceFile 307 with Unknown 0 is caught by the independent production-listener comparison.
Not covered: node-handler conversion, shared kind-indexed dispatch, complete BooleanPropNaming and production JSX.

Each rule directory contains a rule.json with its production name and numeric
kinds. Values come from the pinned typescript-go SyntaxKind enumeration. The
verification script checks each declaration against the unchanged Go production
listener keys, rather than treating the handwritten manifest as its own truth.
No shared generator, harness, parser, driver or compiler source changed.

These are listener declarations for integration, not a claim that the existing
implementations meet the new node-handler contract. Existing rule entry points
still scan nodes or receive table indexes, and their parser exposes string
kinds. Converting them to accept an already loaded numeric-kind node needs the
shared node and driver interface. SourceFile listeners intentionally preserve
Go's whole-file entry points rather than guessing more selective semantics.
The current driver does not consume these files, so findings, fixes and
suggestions are unaffected and the prior landing byte-oracle results apply to
exactly the same executable sources. No rule oracle or sanitizer rerun is
represented as a fresh result for this metadata-only change.

All existing implementation and validation work is already pushed. The remaining
BooleanPropNaming scope includes an unwired component/props detector and a
configurable RegExp matcher. The shared parser still lacks production JSX and
stage 0 refuses nonconstant RegExp patterns. No full parity is claimed.
The next unclaimed candidates found in the previous scan are JSX rules, beginning
with react/jsx-fragments; this unit adds declarations to existing claims rather
than increasing work in progress while their integration remains unfinished.

Verification output is preserved in verification.log. The numeric-key mutation
is applied to an in-memory copy only; it does not alter a committed declaration.
