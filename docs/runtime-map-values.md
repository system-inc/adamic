# Map value representations

The 37 root sites below come from `stage3/notyet-table/roots/raw.csv` on
`codex/stage3-notyet-table` at `57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7`.
That census used compiler base `44583d3283fdd8674085a7ddcce040cf2a73a94e`.
These are value-storage findings, except the last row, whose type arguments
were not recovered. They are not Map key kinds. Counts are recorded sites,
not a measurement of how many tsc sites this change now compiles.

A Map entry holds an `adamic_value`. An `ir.Union` uses its reference member:
number members are existing number boxes, booleans are existing constant boxes,
strings and objects keep their existing counted references, and undefined is
NULL. `fit` already boxes members in Map construction and `set`; `get`, copying,
iteration, replacement, deletion and clearing already retain or release references.
This change allows that representation in `mapTypes`, without making Union
storable in every other compiler context and without changing Map keys.

All source sites in the table are relative to `src/compiler/`. Fixture names
are relative to `internal/oracle/testdata/`, with prefix `map_union_value_` and
suffix `.a`. Fixtures reduce tsc's shapes while preserving member representations;
they do not compile the complete tsc declarations or establish whole-program
checker acceptance.

| Recorded value type | Sites | Can the existing slot hold it? / compiler work still needed | Node fixture |
|---|---|---|---|
| `HostFileInfo` (7) | watchPublic.ts:718:21, 751:33, 759:32, 801:32, 814:32, 823:36, 1060:59 | Yes: object members plus false, as Union. | host_file_info |
| `false \| SourceFile \| undefined` (6) | program.ts:2557:13, 2563:9, 2775:16, 3570:13, 3726:9, 4408:21 | Yes: object, boolean box and NULL. | source_file |
| `false \| ResolvedProjectReference` (5) | program.ts:3757:16, 4007:27, 4018:17, 4031:17, 4042:9 | Yes: object or boolean box. | resolved_project |
| `string \| false` (5) | builder.ts:899:26; builderState.ts:456:9; program.ts:530:27, 537:23, 543:9 | Yes: string or boolean box. | string_false |
| `false \| MutableFileSystemEntries` (4) | watchUtilities.ts:145:16, 173:13, 233:16, 377:9 | Yes: object (including its reference fields) or boolean box. | filesystem_entries |
| `ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }` (2) | checker.ts:5239:46, 5297:21 | Physically an object reference fits. The existing intersection proof only accepts scalar fields; it needs proof for reference fields before `mapTypes` can accept this type. | Deferred |
| `T` (2) | core.ts:1908:17; moduleNameResolver.ts:1120:24 | Depends on specialization. A proven concrete instantiation can fit; an unresolved parameter needs specialization/type-substitution evidence, not an invented slot type. | Deferred |
| `VisitResult<ExportAssignment \| LateVisibilityPaintedStatement \| undefined>` (2) | transformers/declarations.ts:987:21, 1347:9 | Yes: this concrete alias is a Union of object references, an array reference and NULL. Unresolved generic VisitResult remains subject to specialization. | visit_result |
| `CompilerOptionsValue` (1) | commandLineParser.ts:2785:20 | No complete representation today: it includes both null and undefined, which would collide as NULL. It needs distinct null/undefined tagging and proven storage for its heterogeneous array members. | Deferred |
| `ResolvedConfigFilePath` (1) | tsbuildPublic.ts:663:5 | Physically a string reference fits. Branded primitive intersections need an erasure/representation proof; the existing object-intersection proof cannot supply it. | Deferred |
| `string \| number` (1) | commandLineParser.ts:3869:17 | Yes: string reference or number box. | string_number |
| Unknown key/value type arguments (1) | resolutionCache.ts:1478:9 | Cannot decide until the compiler recovers both arguments and proves their representations. | Deferred |

The seven enabled categories cover 30 recorded root sites. The remaining seven
sites are deferred. Each Node fixture exercises construction, `set`, `get`,
replacement, copy, deletion and reinsertion, iteration order, and clearing.
Object fixtures also keep a retrieved value alive across replacement. The
source-file fixture distinguishes a present undefined entry from a missing key
with `has`. The string/number fixture preserves NaN and negative zero as values.
The oracle compares native sanitizer, release, slab allocator and JavaScript
backend outputs against Node, and additionally checks leaks and balanced counts.
Boundary tests keep unsupported value representations and mixed Map key kinds
refused.
