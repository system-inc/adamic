Built: slice 3 re-cut as one commit, retaining assignment proofs, optional-read exceptions, call spread, JSON shorthand, frontend rulings and TSGo caching.
Commits: base d79302e50c022c997c00f7c68ddb7475f9ee65fc; source 9365dc4007838b7d1a6a317749a7e8af8474cb60; delivery SHA is in the final response and git log.
Commands and outputs: focused fixtures, seven stage-1 gap packages, all guards, readers and 14-package vet pass; counts add 22 rows and change 11 owned existing rows.
Mutants: all 14 production mutants and eight embedded mutants were caught; each result and catcher is below.
Not covered: full package suites, the full gate, WASI, macOS execution and real checker ABI execution. No PR or Loom submission was made.

## Main and the conflict

Exactly one textual conflict occurred, in internal/native/tsgo.go. Main's ae1d1c703eef648048dd94e7d53581dba8862b99 added a withCompilerToken wrapper around the checker link, enforcing ADAMIC_CC_JOBSERVER. The source selects compilerName(options) and caches the TSGo runtime through splitTSGoRuntime. Resolved once by retaining both the selected compiler and token wrapper. The source's separate Darwin compile also takes a token, preserving main's every-clang rule for that new path. Relative inputs, runtime caching, feature flags, prefix maps and stable Darwin object metadata remain. Linux TSGo cache, placeless archive and jobserver guards pass. The cache tests link inert ABI stubs; they do not exercise the real checker ABI.

Current main's counts API moved measurement into keyed products (1f07ce2d, followed by ac8ffaac and 9f8ce955) and removed the digest column again. Conservative assumption: use the fetched tip's counts-only schema and current preparation/measurement API. Assignment registrations return the new tuple. The six frontend-mode registrations prepare C once and return a measurement closure; the physical TypeScript copy's path is passed through main's countPortable, so a temporary directory cannot enter the digest.

The first oracle guard invocation did not build because the assignment registrations still returned only a string. No test ran or mutant credit was claimed. Converted those registrations and ran the actual guard successfully. All source artifacts remain .a; the source's existing temporary .ts frontend adapter is retained. Stock TypeScript 6.0.3 dependencies were installed with timeout 120 npm ci --prefix stage3/api --ignore-scripts (npm.log).

## Counts and runtime

Took main's complete counts.md, then regenerated exactly once on Linux with house clang 20.1.8. The 22 additions and 11 existing changes have no removals. counts-attribution.json records every full row. An isolated Go overlay reverting only internal/lower/narrowed.go, internal/lower/exceptions.go, internal/native/reuse.go and internal/flow/build.go to main reproduces all 11 old main rows exactly (counts-counterfactual.log, 2.862 seconds in the test process). This was a measurement proof, with no second regeneration. Exact command: GOMAXPROCS=4 timeout 180 go test ./internal/oracle -overlay=review/compiler/loom-slice-3-3/counts-counterfactual/overlay.json -run ^TestSlice3CountsAttribution$ -count=1 -json -timeout=90s -parallel=4; exit 0. New rows and their numeric changes agree with the source's recorded ownership.

Runtime +/- lines are identical to the source: adamic.h +1, closure.c +1, union.c +3, no removals or differences (runtime-comparison.json).

| Existing row | Main A/F/R/L/P/G | Re-cut A/F/R/L/P/G |
|---|---|---|
| internal/load/testdata/0.1/compile/09_tree.ts | 19/19/28/69/16/0 | 33/33/66/69/16/0 |
| internal/oracle/testdata/narrowed_reads.a | 5/5/3/7/4/0 | 6/5/6/9/4/0 |
| internal/oracle/testdata/narrowed_methods.a | 3/3/3/6/3/0 | 4/3/6/8/3/0 |
| internal/oracle/testdata/narrowed_fields.a | 4/3/4/6/4/0 | 5/3/7/8/4/0 |
| internal/oracle/testdata/reuse_narrowed.a | 7/5/6/8/5/0 | 8/5/9/10/5/0 |
| internal/oracle/testdata/regexp_null_narrowed.a | 4/4/8/5/3/0 | 5/4/11/7/3/0 |
| internal/oracle/testdata/element_access_presence.a | 4/3/5/7/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/047cb0d_n_element_plain.a | 4/3/5/8/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/047cb0d_n_element.a | 4/3/6/8/4/0 | 5/3/9/10/4/0 |
| internal/oracle/testdata/047cb0d_n_element_method.a | 4/3/6/9/4/0 | 5/3/9/11/4/0 |
| internal/oracle/testdata/047cb0d_n_arrayindex.a | 1/1/1/2/1/0 | 2/1/4/4/1/0 |

## Guards and grain

TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy and TestCallTargetReaders all pass. No new compiler-consuming Go package or runtime feature set was introduced, so the dependency registry and runtime witness declarations needed no additions. The source's existing serial markdown helper explanation satisfies the parallel rule. No test was deleted or split; every completed affected leaf is below 60 seconds. leaf-seconds.json lists every selected leaf's seconds. Tests include setup in their clock. Runtime products were prepared one group at a time.

The TSGo cold/warm release and sanitized builds agree with source Node, retain identical binary bytes, and do not replace their runtime archive. The emitted binaries name no build directory. The native jobserver tests pass. Changed-package vet includes internal/javascript and all seven changed stage-1 packages, 14 packages total. No code was copied from cohere.

Setup completed in 51.649 seconds. Every timing line is retained in setup.log. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc=5 and cgroup cpu.max=400000 100000. All builds and tests sourced the setup environment with GOPROXY=https://proxy.golang.org|direct; suites used GOMAXPROCS=4 and four-CPU affinity. Counterfactual measurement used GOMAXPROCS=4 on the same four-CPU quota. No whole package suite or full gate was run; setup compiled all repository packages.

## Commands

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
| parallel | 0 | 1.171 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| maps | 0 | 12.688 | `go test ./internal/ir -run ^TestEveryMapAProgramHoldsIsSorted$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 1 | 2.221 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 0 | 9.477 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| agreement-registry | 0 | 9.679 | `go test ./internal/lower -run ^TestNativeAgreementRowsNameRegisteredPrograms$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized0 | 0 | 12.836 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized1 | 0 | 12.247 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized2 | 0 | 13.268 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized2$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release0 | 0 | 10.076 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release1 | 0 | 10.004 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release2 | 0 | 10.396 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease2$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted0 | 0 | 10.143 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted1 | 0 | 10.059 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted2 | 0 | 9.779 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted2$ -count=1 -json -timeout=90s -parallel=4` |
| native | 0 | 16.469 | `go test ./internal/native -run ^Test(TSGoCache(Release\|Sanitized)\|TSGoBuildSeesTheProgramsFeatures\|CheckerBuildsNameNoPlaceOfTheArchive\|NativeCachesWriteNothingUnderHome)$ -count=1 -json -timeout=90s -parallel=4` |
| optional | 0 | 7.849 | `go test ./internal/oracle -run ^TestOptionalAfterCall -count=1 -json -timeout=90s -parallel=4` |
| spread | 0 | 8.263 | `go test ./internal/oracle -run ^Test(CallSpread\|JSONStringifyShorthand) -count=1 -json -timeout=90s -parallel=4` |
| function | 0 | 7.625 | `go test ./internal/oracle -run ^TestFunction(Adamic\|TypeScript\|CallSignature) -count=1 -json -timeout=90s -parallel=4` |
| merged | 0 | 6.929 | `go test ./internal/oracle -run ^TestMergedField -count=1 -json -timeout=90s -parallel=4` |
| assignment | 0 | 9.901 | `go test ./internal/oracle -run ^TestAssignmentProof -count=1 -json -timeout=90s -parallel=4` |
| flow | 0 | 17.081 | `go test ./internal/flow -run ^TestOptionalAfterCallFlowEdges$ -count=1 -json -timeout=90s -parallel=4` |
| lower | 0 | 2.532 | `go test ./internal/lower -run ^TestOptionalAfterCallPreservesEarlierThrow$ -count=1 -json -timeout=90s -parallel=4` |
| cssstrings | 0 | 13.692 | `go test ./stage1/cohere/cssstrings -run ^TestMultiPushGap$ -count=1 -json -timeout=90s -parallel=4` |
| json | 0 | 8.26 | `go test ./stage1/cohere/json -run ^TestClosedMultiplePushGap$ -count=1 -json -timeout=90s -parallel=4` |
| markdownblocks | 0 | 18.258 | `go test ./stage1/cohere/markdownblocks -run ^TestParserRepresentationProbes$ -count=1 -json -timeout=90s -parallel=4` |
| selector | 0 | 7.364 | `go test ./stage1/cohere/selector -run ^TestEachGapStandsWhereGapsMdSaysItDoes$ -count=1 -json -timeout=90s -parallel=4` |
| yaml | 0 | 10.082 | `go test ./stage1/cohere/yaml -run ^Test(LexerGaps\|Product_YAMLClosedGaps(Lowered\|Native))$ -count=1 -json -timeout=90s -parallel=4` |
| parser | 0 | 8.135 | `go test ./stage1/typescript/parser -run ^TestPushSpreadGap$ -count=1 -json -timeout=90s -parallel=4` |
| scanner | 0 | 8.816 | `go test ./stage1/typescript/scanner -run ^TestGapStandsWhereGapsMdSays$ -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 19.163 | `go test ./stage3/fixtures -run ^Test(FixturesAssignmentProofs\|FixtureDirectoriesHaveTopLevelTests)$ -count=1 -json -timeout=90s -parallel=4` |
| reader-products | 0 | 10.943 | `go test ./internal/ir -run ^TestProduct_CallTargetExports$ -count=1 -json -timeout=90s -parallel=4` |
| readers | 0 | 5.133 | `go test ./internal/ir -run ^TestCallTargetReaders -count=1 -json -timeout=90s -parallel=4` |
| jobserver | 0 | 2.936 | `go test ./internal/native -run ^Test(CompilesHoldAJobserverTokenEach\|AFailedCompileGivesItsTokenBack\|AJobserverThatCantBeOpenedFailsTheCompile\|TheJobserverMovesNoKey)$ -count=1 -json -timeout=90s -parallel=4` |
| vet | 0 | 4.109 | `go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/javascript ./internal/oracle ./stage3/fixtures ./stage1/cohere/cssstrings ./stage1/cohere/json ./stage1/cohere/markdownblocks ./stage1/cohere/selector ./stage1/cohere/yaml ./stage1/typescript/parser ./stage1/typescript/scanner` |
| counts | 0 | 53.568 | `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts` |
| counts-table | 0 | 3.131 | `go test ./internal/oracle -run ^TestCountsAreRecordedTable$ -count=1 -json -timeout=90s -parallel=4` |

## Production mutants

| Mutant | Caught | Seconds | Catching tests |
|---|---|---:|---|
| check-optional-receiver | True | 16.528 | TestOptionalAfterCall79 |
| javascript-required-panic | True | 10.421 | TestOptionalAfterCallRequired |
| native-required-panic | True | 11.611 | TestOptionalAfterCallRequired |
| drop-flow-throw-edge | True | 16.915 | TestOptionalAfterCallFlowEdges |
| drop-required-propagation | True | 16.698 | TestOptionalAfterCallPropagation |
| merged-computed-read | True | 15.997 | TestMergedFieldTypeScriptComputedChecked |
| merged-declaration | True | 13.542 | TestMergedFieldAdamicUnusedRefused |
| merged-read | True | 15.985 | TestMergedFieldTypeScriptMissingChecked |
| function-annotation | True | 15.009 | TestFunctionAdamicUnusedAnnotationRefused |
| function-call | True | 14.586 | TestFunctionTypeScriptCallRefused |
| function-length | True | 19.69 | TestFunctionTypeScriptLengthAgreesWithNode |
| erase-earlier-throw | True | 10.77 | TestOptionalAfterCallPreservesEarlierThrow |
| tsgo-absolute-main | True | 9.919 | TestTSGoCacheSanitized |
| tsgo-rewrite-runtime | True | 9.02 | TestTSGoCacheRelease |

Every production mutant exited 1 at a behavior assertion; build failures and timeouts are excluded. Exact commands and raw failures are in mutants-results.json and mutants/*/test.txt. Mutant Go sources use .go.txt.

The eight embedded mutants are TestAssignmentProofScannerKeywordMutant, TestAssignmentProofIfNarrowingMutant, TestAssignmentProofReturnDefinedMutant, TestAssignmentProofChainedMutant, TestAssignmentProofCompoundMutant, TestAssignmentProofWiderTargetMutant, TestCallSpreadEvaluatedTwiceMutant and TestJSONStringifyShorthandWrongBindingMutant. Each successfully builds/runs altered valid source or IR and requires a disagreement with independently captured Node bytes. Assignment witnesses also pass stock TypeScript with zero diagnostics; semantic execution uses release, JavaScript, ASan/UBSan and leak checks. TSGo's feature omission negative control is an include-order compile contract, not a semantic mutant credit.
