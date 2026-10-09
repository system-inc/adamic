# Defense of internal/load export collision audit

Starting origin/main: cf735d9fba9e38de6368575e5630e44375a86eaf. Prior audit: 4ae96142a65eff81725403519a69754ea4e7d83c.

All 17 current package tests ran in the clean baseline and every mutant matrix. The test names are unchanged from the audit. No test skipped or timed out. No production or test changes remain. Every standalone diff passed git apply --check and go vet ./internal/load/ when applied alone. Each matrix used its own ADAMIC_BUILD_CACHE_DIR. nproc is 5.

Five rows defended; the explicit-export row was not defended in three attempts. This is evidence to keep the five rows, not evidence to delete the sixth. results.json has verdicts, matrix.json has every failed and passed row, and coverage-diffs.json has target-exclusive covered blocks relative to each named subsumer. All line numbers refer to the starting commit.

## Code under test and oracle, per row

* Explicit export: Adamic load, compiler options and .a source aliasing. Oracle: handwritten expectation that loading an explicit re-export succeeds. It does not inspect the selected binding.
* Type error: Adamic diagnostic collection, position and localization. Oracle: handwritten exact TypeScript TS2322 string and position. The upstream checker and localization implementation were not mutated; the loader's assignment of its localized result was dropped.
* Non-Adamic inputs: Adamic Load root validation. Oracle: handwritten error substrings for JavaScript, missing and shadowed roots, plus error presence for nil roots.
* Spec programs: Adamic loader configuration and TypeScript root resolution. Oracle: acceptance of checked-in docs/0.1.md examples. This row runs no external comparator.
* Different Node version: Adamic nodeTypesIndex metadata validation. Oracle: handwritten requirement that rejecting 24.0.0 mentions the repository's pin 25.3.3.
* Official console signatures: Adamic Node import recognition and prelude selection. Oracle: acceptance of console.log(true), console.log(undefined), and console.log('a','b') under official @types/node 25.3.3 declarations. Its unique catch is rejection of a type-only Node import, not direct proof of each signature's shape.

## Coverage and semantic differences

* Explicit export reaches successful-load root collection at load.go:152-163, unlike the collision refusal row. Many other successful loaders reach it. Its unique input is an explicit export clause resolving two star exports, a checker responsibility. Only Adamic configuration and aliasing were mutated.
* Type error reaches default diagnostic localization at load.go:253; the collision row instead uses Adamic's custom TS2308 message. D01 removes that assignment. Other rows only inspect diagnostic codes, counts or positions and pass.
* Root refusal reaches load.go:93, 110, 212, 215 and 219 that its subsumer never reaches. D02 excludes the empty list from the existing refusal bound. Only this row calls Load(nil).
* Spec programs reach successful root loading and .ts source reading, unlike the collision row. On shared compiler options they import './geometry.ts'; the star collision and explicit export rows use .a aliases. D06 changes that option and uniquely rejects compile/07_modules with TS5097. D06 was originally selected for explicit re-export loading, and its observed unique catch instead defends this row.
* Wrong Node version uniquely covers node_library.go:32 relative to its positive-load subsumer. D03 drops only the version comparison, preserving package-name validation and successful loads.
* Console has no exclusive covered blocks relative to its subsumer. Its import is type-only, while its subsumer imports values. D05 excludes type-only imports from the existing Node recognition condition and uniquely fails this row with TS2591. It changes a condition, not an inserted statement.

## Attempt limits and owner findings

D04 used ES5 and D07 used Classic resolution. Current TypeScript has removed those options, producing broad TS5108 configuration failures. These attempts do not show a narrow behavioral defense. D08 uses supported CommonJS with verbatim syntax and rejects exported values in several rows. D06 leaves explicit .a re-export loading green but rejects the spec's .ts extension import. The explicit-export row therefore has three observed non-unique attempts, with two relevant limitations: one removed option and no owned resolver branch exclusive to explicit binding resolution.

TestExplicitExportResolvesStarCollision promises resolution in its name, but asserts only error-free loading. It does not inspect that 'shared' resolves to the left module or that a nonnil program is returned. The previous empty-answer result remains prior-session evidence only; no new empty-answer probe was run in this defense.

The wrong-version row checks an error substring, so another error mentioning the pin could satisfy it. The console row likewise accepts any nil error without inspecting the program. Defense establishes unique catches for the selected real production mutations, not complete oracle strength.

## Costs and brief friction

Warm toolchain worked, so setup was skipped. npm ci ran successfully before the baseline; its separate elapsed time was not measured. Baseline test binary: 1.460 seconds. Eight full-package mutant invocations, including compilation: 49.887 seconds. Eight go vet validations: 1.313 seconds. Coverage ran once for each of the eight distinct target/subsumer tests; logs retain the package-reported times. No separate performance medians were requested.

The brief refers to an audit report but does not name its file. report.md did not exist; results.json, summary.json, code-and-oracle.md, menu.json and matrix.json were read instead, and relevant evidence copied here. Fetching by name initially populated FETCH_HEAD only, so an explicit remote-tracking ref was fetched before reading the rest. The defense branch starts at current origin/main, not the old audit commit. Current pinned TypeScript rejects two old compiler options, which cost two unhelpful matrix runs. A mutant aimed at one row can uniquely defend a different row; D06 is recorded transparently in both attempt histories.

No other packages were run. Upstream TypeScript checker code, installed declarations, fixtures, harnesses and tests were untouched. No tests were removed or rewritten. All eight mutants were caught; there are no survivors. Full passing-row lists for each unique mutant are recorded in matrix.json.
