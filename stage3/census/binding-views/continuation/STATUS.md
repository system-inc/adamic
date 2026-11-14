# Declaration gate cleared; source proofs pass; whole lane pending

Resolution-only reproduces the original export-modifier diff. Program-only
leaves both declarations identical. Internal exported ResolutionReadView with
both halves applied leaves both declarations byte-identical.

Full resolution census: 133 -> 129 unique, 0 new, 4 gone, 0 changed.
Program census: scan only, 81 -> 79, 0 new, 2 gone, 0 changed.
The restoring mutant returns all six exact refusals. Stock tsc's 79 roots with
scout Node ambient inputs report 0 -> 0; Adamic types exits 0 with no diagnostics.
All 1,208 JavaScript emits across 604 files match baseline. Adapter controls pass.
Generated input checks pass with no regenerated files. Full lane is running.
