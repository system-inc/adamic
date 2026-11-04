# Slice 4 (#b5w3ycg) price: frequency on Node x per-read cost native

Measured on: Intel Xeon @ 2.10GHz, 4 vCPU (shared VM), Node v22.23.4, clang 18.1.3, tsc v6.0.3 (050880ce).
The scout branch `evidence/compiler/slice4-scout` does not exist, so the selection was rebuilt with the TS API
(`scripts/select-and-instrument.cjs`) on the PLAIN v6.0.3 tree, not the adapted tree.

Selection: AsExpression/TypeAssertion in src/compiler, source = non-union interface deriving Node whose `kind` is
the open `SyntaxKind`; target = interface (or union of interfaces) whose `kind` is wholly enum literals.
1679 sites (1421 interface downcast, 258 union member; 607 in checker.ts; 1263 with target tags claimed by no other
selected target). Lower's ~370 was NOT reproduced; per-site data lets any subset be priced.

Files: scripts/ (selector+instrumenter, aggregate, price), data/ (sites.csv, read-sites.csv, counts-*.json,
per-site-*.csv, price-*.csv, summary.json, price.json, node-timings.txt, instrumented.diff), bench/ (generator,
sources, timings, emitted C for the checked read).
