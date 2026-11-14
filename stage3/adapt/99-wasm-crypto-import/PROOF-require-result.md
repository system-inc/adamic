# Require result annotation proof

Task #baobab item 9; base 8e3a784aa5a601f57342c07b68f145fe56252c21.
The only added source change is `(baseDir, moduleName): ModuleImportResult =>`.
Stock TypeScript 6.0.3 accepts both returns with zero diagnostics on all 79 roots,
using stage3/api's pinned Node types. Adamic types exits zero on the same roots.
All 604 source files emit byte-identical JavaScript (1,208 comparisons with
comments kept/removed). The full sys.ts CommonJSBundle census goes from 185 to
184 findings: only the 1610:22 refusal_scan finding disappears; zero new or
changed findings. `module: any` adds no finding under the existing computed-path
require host model. The restoring mutant reinstates the original source.

The evidence branch retains complete streams, delta records with phases,
commands, source pins, contract, restoring-mutant, upstream npm ci / hereby dts
and empty declaration diffs, full apply, and whole unfiltered lane receipts.
These receipts, rather than this document, establish the final acceptance gates.
No compiler rule, sanction, lane count or filter is changed.
