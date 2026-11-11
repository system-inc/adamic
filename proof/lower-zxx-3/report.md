# Lower zxx-3 proof report

All requested checks passed. No failing tests or first error lines were produced. `TestCallTargetReadersLower` passed.

## Checkout and environment

- Repository: `https://github.com/system-inc/adamic.git`.
- Tested detached HEAD: `e56ef2d059f4ee155c91b2adb70cbe47808e2cdd` (`compiler/forof-writes-3`).
- Parent: `16429a30452d027809f0ea85bc395479423e493e`; the tested HEAD is one commit above it.
- Ran `git fetch origin compiler/forof-writes-3`, detached checkout of the requested HEAD, and `git submodule update --init --recursive`; all succeeded.
- Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` before setup and checks. Created its missing `TMPDIR`, `/tmp/adamic-gate`, outside the repository before running Go.
- Provisioned upstream `https://github.com/microsoft/TypeScript.git` at `/workspace/typescript-proof-zxx-3`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Exported `ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-proof-zxx-3` for every proof command.
- Toolchain: `go version go1.27.1 linux/amd64`; clang 20.1.8.
- Raw command logs and exit records are outside the repository in `/tmp/adamic-proof-zxx-3`.
- The repository working tree was clean before and after testing. No compiler source was edited.

## Commands and results

`-json` was added to whole-package test commands to retain explicit per-test outcomes; package selection, count, and requested timeouts were preserved. Counts below include subtests.

| Command | Exit | Result |
| --- | ---: | --- |
| `gofmt -l internal/lower` | 0 | Empty output; formatting clean. |
| `go vet ./internal/lower ./internal/ir` | 0 | Empty output; no diagnostics. |
| `go test ./internal/ir -count=1 -timeout 30m -json` | 0 | PASS: 57 tests/subtests; 0 failures; 0 skips. |
| `go test ./internal/lower -run 'ForOf\|LocalWritesAreLocalFieldsWrites\|LocalFieldsClassifyEveryIntField\|Spread' -count=1 -v -timeout 40m` | 0 | PASS: 23 top-level tests, 75 results including subtests; 0 failures; 0 skips. |
| `go test ./internal/compilercensus -count=1 -timeout 40m -json` | 0 | PASS: 7 tests/subtests; 0 failures; 0 skips. |
| `go test ./internal/dry -count=1 -timeout 40m -json` | 0 | PASS: 12 tests/subtests; 0 failures; 0 skips. |
| `go test ./internal/flow -count=1 -timeout 40m -json` | 0 | PASS: 1597 tests/subtests; 0 failures; 0 skips. |
| `go test ./internal/lowered -count=1 -timeout 40m -json` | 0 | PASS: 19 tests/subtests; 0 failures; 0 skips. |

`TestCallTargetReadersLower`: **PASS**, 1.32 seconds. Every failing IR test and first error line: **none**. Every additional-package failure and first error line: **none**.

### Execution times (UTC)

| Check | Start | End |
| --- | --- | --- |
| gofmt | 2026-11-11T12:16:14Z | 2026-11-11T12:16:14Z |
| vet | 2026-11-11T12:16:14Z | 2026-11-11T12:19:34Z |
| ir | 2026-11-11T12:19:34Z | 2026-11-11T12:27:31Z |
| lower | 2026-11-11T12:27:31Z | 2026-11-11T12:28:15Z |
| compilercensus | 2026-11-11T12:28:15Z | 2026-11-11T12:28:24Z |
| dry | 2026-11-11T12:28:24Z | 2026-11-11T12:29:09Z |
| flow | 2026-11-11T12:29:09Z | 2026-11-11T12:36:20Z |
| lowered | 2026-11-11T12:36:20Z | 2026-11-11T12:37:14Z |

## Source-guard package discovery

Searched `internal/**/*_test.go` with `rg` for `../lower`, `internal/lower`, path components naming `lower`, source-reading calls (`ReadFile`, `ReadDir`, `Glob`, `ParseFile`), and repository-wide walks. Inspected constructed paths and broad scans as well as literal paths. The whole IR run already covers its lowering-source guards in `call_targets_guard_test.go`, `method_callee_test.go`, and `iterator_next_guard_test.go`.

Other packages whose guards read lowering Go source, including indirect directory scans, were run whole:

| Package | Guard evidence | Result |
| --- | --- | --- |
| `internal/compilercensus` | `census_test.go:82` (`TestEveryCCompilerRunsThroughNative`), `census_test.go:175` (`TestEveryRunnerTimeCompileNamesItsTarget`), `jobserver_test.go:30`, and `product_clang_test.go:46` scan compiler/repository Go source, including lowering. | PASS; exit 0; no failures. |
| `internal/dry` | `mirrors_test.go:31` (`TestCohereMirrors`) uses `goFiles` at line 57 to parse repository Go source, including lowering, for mirror-policy guards. | PASS; exit 0; no failures. |
| `internal/flow` | `source_line_test.go:17` (`TestEveryStatementIsClassified`) walks `..` at line 62 and parses compiler Go source to guard statement dispatchers, including lowering. | PASS; exit 0; no failures. |
| `internal/lowered` | `lowered_test.go:377` (`TestEveryEnvironmentReadOfTheLoweringIsKeyed`) enumerates the load/lower dependency closure on Linux and Darwin and parses its source for environment reads. | PASS; exit 0; no failures. |

Fixture reads under `lower/testdata`, imports of `lower`, cache-input declarations, and mutation inputs alone were distinguished from guards over lowering Go source. No other package with such a source guard was identified.

## Selected lowering results

Every result emitted by the requested selection is listed below, including nested subtests.

| Test | Result |
| --- | --- |
| `TestCallableContractPlainSpread` | PASS |
| `TestCallableContractSpreadRefused` | PASS |
| `TestClassFeaturesThrowingAccessorSpread` | PASS |
| `TestForOfConstBindingStillDispatches` | PASS |
| `TestForOfReassignedBindingIsNoPlainSpreadSource` | PASS |
| `TestForOfReassignedBindingIsUnknown` | PASS |
| `TestForOfReassignedPatternBindingIsUnknown` | PASS |
| `TestForOfReassignedSpreadSourceIsNotPlain` | PASS |
| `TestGeneratorVoidReturnIsNeverReadBySpreadOrArrayFrom` | PASS |
| `TestInputSpreadCoverage` | PASS |
| `TestInputSpreadCoverage/fileStatus_alias.a` | PASS |
| `TestInputSpreadCoverage/fileStatus_array.a` | PASS |
| `TestInputSpreadCoverage/fileStatus_local.a` | PASS |
| `TestInputSpreadCoverage/fileStatus_nested.a` | PASS |
| `TestInputSpreadCoverage/fileStatus_parentheses.a` | PASS |
| `TestInputSpreadCoverage/fileStatus_tuple.a` | PASS |
| `TestInputSpreadCoverage/programArguments_alias.a` | PASS |
| `TestInputSpreadCoverage/programArguments_array.a` | PASS |
| `TestInputSpreadCoverage/programArguments_local.a` | PASS |
| `TestInputSpreadCoverage/programArguments_nested.a` | PASS |
| `TestInputSpreadCoverage/programArguments_parentheses.a` | PASS |
| `TestInputSpreadCoverage/programArguments_tuple.a` | PASS |
| `TestInputSpreadCoverage/readDirectory_alias.a` | PASS |
| `TestInputSpreadCoverage/readDirectory_array.a` | PASS |
| `TestInputSpreadCoverage/readDirectory_local.a` | PASS |
| `TestInputSpreadCoverage/readDirectory_nested.a` | PASS |
| `TestInputSpreadCoverage/readDirectory_parentheses.a` | PASS |
| `TestInputSpreadCoverage/readDirectory_tuple.a` | PASS |
| `TestInputSpreadCoverage/readTextFile_alias.a` | PASS |
| `TestInputSpreadCoverage/readTextFile_array.a` | PASS |
| `TestInputSpreadCoverage/readTextFile_local.a` | PASS |
| `TestInputSpreadCoverage/readTextFile_nested.a` | PASS |
| `TestInputSpreadCoverage/readTextFile_parentheses.a` | PASS |
| `TestInputSpreadCoverage/readTextFile_tuple.a` | PASS |
| `TestInputSpreadCoverage/utf8At_alias.a` | PASS |
| `TestInputSpreadCoverage/utf8At_array.a` | PASS |
| `TestInputSpreadCoverage/utf8At_later.a` | PASS |
| `TestInputSpreadCoverage/utf8At_local.a` | PASS |
| `TestInputSpreadCoverage/utf8At_multiple.a` | PASS |
| `TestInputSpreadCoverage/utf8At_nested.a` | PASS |
| `TestInputSpreadCoverage/utf8At_parentheses.a` | PASS |
| `TestInputSpreadCoverage/utf8At_tuple.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_alias.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_array.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_empty.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_local.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_nested.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_parentheses.a` | PASS |
| `TestInputSpreadCoverage/utf8Length_tuple.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_alias.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_array.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_later.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_local.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_multiple.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_nested.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_parentheses.a` | PASS |
| `TestInputSpreadCoverage/writeTextFile_tuple.a` | PASS |
| `TestLibrarySpreadDateView` | PASS |
| `TestLibrarySpreadErrorView` | PASS |
| `TestLibrarySpreadMapView` | PASS |
| `TestLibrarySpreadPlainProducers` | PASS |
| `TestLibrarySpreadRegExpUnion` | PASS |
| `TestLibrarySpreadSetView` | PASS |
| `TestLibrarySpreadTypedArrayView` | PASS |
| `TestLocalFieldsClassifyEveryIntField` | PASS |
| `TestLocalWritesAreLocalFieldsWrites` | PASS |
| `TestMixedTupleSpreadIsNotYet` | PASS |
| `TestNullableObjectSpreadKeepsNarrowerLiteralRefused` | PASS |
| `TestOptionalWideningSpreadOverwrite` | PASS |
| `TestSpreadRecordsTheGettersItMayRun` | PASS |
| `TestSpreadRecordsTheGettersItMayRun/a_class's_getter` | PASS |
| `TestSpreadRecordsTheGettersItMayRun/a_getter_inherited_from_a_base` | PASS |
| `TestSpreadRecordsTheGettersItMayRun/a_getter_of_a_class_no_Tree_can_be` | PASS |
| `TestSpreadRecordsTheGettersItMayRun/an_object_literal's_getter` | PASS |
| `TestSpreadRecordsTheGettersItMayRun/no_getter_anywhere` | PASS |

## Report delivery

The report branch is `proof/lower-zxx-3-report`, created from the tested HEAD after checks. Its only added file is `proof/lower-zxx-3/report.md`. The report-only commit is published to `origin` as `refs/heads/proof/lower-zxx-3-report`; no source changes or other branches are included.
