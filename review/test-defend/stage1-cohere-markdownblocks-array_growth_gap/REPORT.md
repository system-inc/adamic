# Identifier scalar defense

Starting origin/main: 2b1be38362046455e0e5454d8b0e674a8630e95d.
Prior audit: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.

## Code and oracle

Code under test: identifier.ts and its identifierCaseKeys.ts / identifierCaseValues.ts tables, executed as source TypeScript and compiled natively through Adamic. The unchanged external oracle is Go cohere NormalizeIdentifier, lowercased using Go strings.ToLower, over every 1,112,064 Unicode scalar. The test compares complete output bytes, not only an exit code or count. Its backend, Node and native executions are inside the same row; there is no separate executor twin among these rows.

## Defense

D1 changes the final case-table value from 125251 to 125250 at identifierCaseValues.ts:85. For U+1E921 the original result is U+1E943 and the mutant gives U+1E942. This is a constant-change mutant in production, with no oracle, harness or test edits. The standalone diff is diffs/D1.diff. It applies to the starting origin/main; apply-validation.json records the check. Actual sanitizer and release builds succeeded in the mutant matrix, as recorded in native-rebuild-evidence.json.

Only TestMdastIdentifierScalars failed, with identifier_independent_test.go:27: identifier scalar oracle first byte difference at 753221 (lengths 7789593/7789593). Fifteen other selected functions passed, none skipped. matrix.json names every pass. This is a defended result within the complete identified import-caller set, not a claim of package-wide or repository-wide uniqueness. A successful unique attempt made further mutants unnecessary. Production source was restored exactly before committing evidence.

## Coverage and scope

The required Go coverpkg profiles were collected individually for the target and subsumer over internal/lower and internal/native. They instrument compiler execution, not the TypeScript port. go-coverage-difference.json is supplemental. Actual production-source V8 coverage establishes that the target executes identifier 1,112,064 times, whereas the subsumer loads neither identifier nor its tables. exclusive-lines.json maps the executed final table literal to baseline line 85. Only identifier_probe.ts and mdast_probe.ts import the changed production modules; caller-map.json gives their import paths and every selected current caller, plus the audited subsumer as a control. Current inventory has 763 test names; none were added or vanished against the saved audit inventory.

The whole-package clean baseline timed out at 90.061 test-binary seconds, with no ordinary assertion failure beforehand. The timeout is not a production kill. A clean reached-caller baseline then passed, binary 39.951 seconds. The D1 reached-caller matrix completed in 39.171 binary seconds. Outside the identified caller set and outside this package, outcomes remain unknown.

## Costs and brief limitations

Warm tools worked, so setup was skipped; nproc was 5. npm ci in stage3/api completed in 0.475 wall seconds before baseline. Whole baseline wall time was 92.224 seconds; clean reached matrix 42.223; D1 matrix 41.286. Target coverage wall time was 18.884 seconds and subsumer coverage 3.885. Product rebuild logs give individual reported durations (identifier lowering 2.07 seconds; release native 0.37; sanitized native 0.58; malformed-event lowering 4.63 and native 6.27). Those products overlap and are not additive wall time.

The brief's Go coverage command cannot cover TypeScript, so collecting both compiler coverage and actual V8 source coverage cost extra work. Whole-package timeout required identifying all production import callers and narrowing the matrix. Prior subsumption rested on a broad compiler conditional mutant, which did not touch this rare case-table boundary. Missing prior failing-line text in the request was resolved by reading saved audit evidence. No additional discrepancy in the brief was observed. The row's name matches its all-scalar assertions; no name/assertion gap was found. No test was deleted, rewritten or weakened.

See rows.json for the verdict and exact command, D1-matrix.log for original test output, and prior-* files for the audit evidence read.
