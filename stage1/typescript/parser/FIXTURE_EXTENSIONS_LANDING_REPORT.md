All 60 parser lint fixtures now use .ts.txt or .tsx.txt. Their bytes are
unchanged; `validation/fixture-extensions/renames.json` records every old and
new repository path and SHA-256. No tracked .ts or .tsx files remain under
parser/testdata. The compiler-and-stage1 corpus retains its ordinary source
selection and no new exclusion list.

`lint_cases_test.go` discovers the new extensions and copies each input to a
temporary .ts or .tsx filename before running Go, Node and sanitized native.
Both parsers select JSX from the filename, so passing .tsx.txt directly would
change the grammar. Expected-sidecar stems remove both the text and source
extensions. A bare source fixture fails the test by name rather than silently
falling out of discovery. `stage1/cohere/lint/lint_test.go` reads its three
await/new fixtures as text and uses temporary source paths for its explicit
no-new and all-rule rows.

All 60 renamed cases passed TestLintCases. Restoring
`testdata/lint_cases/decorated_async_promise_executor.ts.txt` to a bare .ts
filename planted a missed-rename mutant; TestLintCases rejected it by name
before building or comparing. The fixture was restored. Its log and metadata
are in `validation/fixture-extensions/missed-rename-mutant.log` and `mutant.json`.

`validation/fixture-extensions/run_gates.py` runs the parser package,
TestCompilerAndStage1Agree, TestRulesAgree, and TestProfileSnapshotsAgree
sequentially. The pinned clean TypeScript checkout is set, parser benchmarks
are enabled, and TestProfileArtifacts creates fresh release and profiling
snapshots in /tmp/adamic-fixture-extensions-profile.

All requested gates passed, with zero failures and zero skips. The parser
package took 950.524 seconds (178 passes including all 60 fixtures and all
22,497 incomplete comparisons). TestCompilerAndStage1Agree took 453.727
seconds and matched all 875 source files on Go, native, Node and emitted
JavaScript. TestRulesAgree took 86.797 seconds. TestProfileSnapshotsAgree
took 194.380 seconds against fresh release and profiling builds; the combined
profile package, including artifact preparation, took 234.299 seconds.
`validation/fixture-extensions/report.json` records counts, commands, command
wall times and load readings. Vet and the pinned checkout cleanliness check
passed. This unit is pushed once after its fixtures, gates and mutant passed.
