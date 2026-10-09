# Defense of the scanner agreement family

Starting origin/main: 76c59c81e8617cea1892a01841494895927a712c. Prior audit: 0622eb766202b361f3b063ae641ade8a036a9f54.

Verdict: not defended after three honest production mutation attempts. Each mutant made the agreement family fail, and each also made TestProfileSnapshotsAgree fail. The other six grouped rows passed. This is evidence of three measured shared catches, not a deletion recommendation or a claim that the rows are equivalent under every fault.

## Code under test and oracle

CODE UNDER TEST: the TypeScript scanner port, stage1/typescript/scanner/{characters,tokens,scanner,main}.ts, compiled as sanitized/release/profiled native products and executed directly through Node's TypeScript stripping loader. The file driver, scanner methods, character predicates and table interpretation are production code.

ORACLE: the pinned TypeScript-Go scanner executable, with full byte-for-byte answers for every case: token kind, byte positions, flags, values, diagnostic codes and spans, with manifest case framing. The checker, Go scanner, oracle adapter, corpus builder, source fixtures, harness and tests were untouched. Native compiler and runtime code were not mutated. This is an external-run oracle.

## Current scope and coverage

All 29 top-level tests from the audit still exist, with no newly named tests. All 29 top-level tests ran in each matrix with benchmark and profile options enabled; grouped as eight rows, as in the audit. The exact test names are in list.log. No row skipped, panicked or timed out. family members are listed in results.json.

The requested Go coverprofile command used -coverpkg=github.com/system-inc/adamic/stage1/typescript/scanner separately for the sixteen-shard family and TestProfileSnapshotsAgree. Both .cover files contain only mode: set, because this package has only Go test files and the code under test is TypeScript. Those profiles provide no port coverage and are not represented as proof of exclusive lines.

Actual port execution coverage was also collected with NODE_V8_COVERAGE for each row's Node side, without changing any harness. family-v8-port.json contains sixteen clean-port process observations; snapshot-v8-port.json contains one. Built-in mutant products were excluded by comparing all four files in each executed script's directory with the original port bytes. coverage_compare.py merges executed intervals, respecting nested zero-count ranges, then compares the union of the family's executions with the snapshot execution. coverage-diffs.json records no family-only executed intervals in any of the four port files. Offsets are UTF16 offsets in Node's transformed JavaScript, not origin/main TypeScript line numbers. Native execution coverage was not collected.

Both rows reached these port functions: contains, digit, isIdentifierPart, isIdentifierStart, isLineBreak, isSpace; Scanner constructor and field initialization, advance, bigint, code, decimalDigits, error, escape, fragment, identifier, number, numberValue, punctuation, rescanGreater, rescanSlash, rescanTemplate, scan, scanJsx, string, template and unicode; driver run and written, plus module-level table/driver execution. scanJsxIdentifier and scanJsxAttributeValue were not reached in either clean Node coverage observation.

Reading both bodies establishes the remaining execution differences: the family partitions the same full askedCorpus across sixteen manifests and compares a sanitized native build plus Node; snapshots compares release, profiled and Node products with the whole askedCorpus. The driver creates a fresh Scanner for every file, so its per-file scanner state does not carry between cases. Case numbers are manifest-local. The family additionally checks four built-in failure witnesses; those are comparisons in the harness, not production differences to mutate under this defense's rules.

## Three attempts

No exclusive clean-port input or coverage lead was found. These attempts therefore tested three distinct generated scanner modes under the different batch layouts and sanitized versus release build regimes. They deliberately change production semantics, preserving all built-in mutation sites and causing no preparation failures.

* D01 flips scanJsx's whitespace classification. Both rows reject JsxTextAllWhiteSpaces where Go says JsxText, including the generated Unicode JSX case. Count-only throughput is unchanged.
* D02 flips the condition allowing an unescaped slash to terminate a regex outside a character class. Both rows detect changed regular-expression boundaries/diagnostics. The full rows passed lists and first failing output lines are in matrix.json.
* D03 changes template rescan's diagnostic/report option from true to false. Both rows detect the invalid-template escape result. Its native builds compile and the normal report comparison fails, not a build precondition.

Each standalone diff is against starting origin/main, has no switch and passed git apply --check. Actual native builds, including TestProfileArtifacts' release/counted/profiled builds and every TestProduct_Scanner construction test, passed while each mutant was active. matrix.json records those build events and their test elapsed times. Each mutant used its own ADAMIC_BUILD_CACHE_DIR and fresh profile artifact directory. ProfileSnapshotsAgree therefore ran mutant products rather than cached clean executables.

## Name and assertion finding

TestScannerAgreesWithTypescriptGo family does what its name promises for the selected corpus: it compares full answer bytes to the pinned Go scanner on both sanitized native and Node sides. Its assertions are not limited to an exit code or token count. No missing named assertion was found. Its shard ownership and built-in failure witnesses are additional reasons to retain its structure, although no permitted production mutant in this defense established a package-unique semantic catch.

## Costs, ambiguities and replay

Warm toolchain worked, so setup was skipped. npm ci ran before the green baseline; separate npm elapsed time was not recorded. nproc: 5. Enabled baseline: 53.373 binary seconds. Family coverage run: 34.202 binary seconds; snapshot coverage run: 22.483 binary seconds. Three mutant commands, including compilation: 176.189 wall seconds. Individual build/test durations are retained in raw JSON; their parallel durations must not be summed as a package wall time.

The brief's Go coverage requirement cannot measure this TypeScript port. Empty Go profiles and actual V8 source coverage are both supplied, with the native-coverage limit explicit. The family must be selected by ^TestScannerAgreesWithTypescriptGo_[0-9]+$, not a literal family name. The audit report was spread across REPORT.md, rows.json, inventory.md and REPLAY.md; all were read and copied.

The repository corpus collector rejects dirty tracked TypeScript. Each mutant was therefore committed locally before its run and restored with a new commit afterward, without resetting or rewriting history. Variant hashes are recorded in matrix.json. Central replay should apply and commit a diff in a clean checkout before running the same enabled matrix. The TypeScript corpus used the already available git checkout at pin 050880ce59e30b356b686bd3144efe24f875ebc8. ProfileSnapshotsAgree requires a fresh profile directory built by TestProfileArtifacts before it runs; both environment variables must point to that directory.

No tests were deleted, rewritten or weakened. No oracle or preparation guards were altered. Only scanner.ts production code was mutated and restored. No other packages were tested and no PR was opened. All three mutants were caught, so there are no survivors. All source diffs are restored in the final branch tree; evidence and the clean variant/restore commit history remain.

Final restored whole-package baseline: 44.868 binary seconds, all 29 top-level tests passed with no skips.
