Built: source e1d03f9700a51458833f12db1320d3d95374a025 applied as one commit on pinned main for compiler/loom-r-slice-2.
Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is in the final response and git log.
Commands and outputs: focused source tests, Node and sanitizers, flow corpus checks, guards, counts and changed-package vet; exact results below.
Mutants: original semantic checks plus extended guard probes and structural inline probe; every observed catcher is below.
Not covered: full package suites/full gate, ahra/Loom, macOS and WASI; cycle-ledger original-source corpus skipped for missing configured input.

## Base and conflict resolutions

Fetched origin, verified the full base SHA above, checked out compiler/loom-r-slice-2 from that exact commit, then used `git cherry-pick --no-commit e1d03f9700a51458833f12db1320d3d95374a025`. This applies only the source commit's own parent-to-commit diff. No diff from main to the source tip was applied. The initial conflict hunks are retained in conflicts.diff.

Every conflicted hunk:

1. internal/ir/call_targets_guard_test.go, callTargetExports: source's custom archive product conflicted with main's centralized buildcache.GoExports. Kept main's GoExports and its explicit leakcheck dependency, called separately for native/lower/fresh/flow. Kept all four reader leaves, four product leaves, and four serial preparation leaves. The serial leaves prepare the actual shared product instead of calculating unused legacy inputs. Removed the superseded custom archive extraction implementation. Retained source's external test import alias handling and per-package stale-reader checks.
2. internal/native/runtime/adamic.h, adamic_thrown declaration: main made the pending exception thread-local. Kept `extern _Thread_local adamic_object *adamic_thrown`, added source's omitted-message comment and field-name helper declaration.
3. internal/native/runtime/exceptions.c, after adamic_error_new: main added OrdinaryToPrimitive, including lazy method lookup, ownership, pending-exception propagation and TypeError creation. Kept that whole function and source's separate adamic_error_field_name function. Kept thread-local state and source's absent-message shape, empty fallback and representation bytes.
4. internal/native/runtime/object.c, adamic_object_optional_find: main publishes shape and slot together in an atomic packed cache. Kept its atomic load, search, cache-store helper and absence behavior; inserted only source's field-name remapping before lookup. The ordinary find remapping auto-merged while keeping main's packed cache store.

Auto-merged lowering/emission changes were retained together with main's newer infrastructure; auto-merges.patch records those diffs. No production lowering change beyond the source's own diff was necessary. Main's newer jobserver support in units.go and newer lowering infrastructure remain in the auto-merged result.

Other necessary adaptations: the native Error C harness now initializes main's one-word caches with `{0}` instead of `{NULL, 0}`. The source self-compare test uses main's shared `leakcheck.Unbalanced(leakRun(counted))` instead of removed local `unbalanced`; the first guard compile attempt exposed this before running tests (initial-build-failure.txt).

Sixteen refusal fixtures moved into miscompile_2a_refused/ and weak_miscompiles_refused/, with corresponding path/fixture registrations updated. refused-fixtures.json lists every move and verifies unchanged source contents. Accepted fixtures remain in their original directories. Conservative assumption: all rejected fixtures belong under *_refused even though this pinned base's flow collector globs only top-level oracle files. Eight separate TestSlice2 flow leaves explicitly exercise the accepted Error and Weak programs; the source's 35 callback flow leaves remain. No other package's tests were edited to conceal a refusal.

## Runtime clearance

Every runtime +/- line is IDENTICAL to the source commit's own runtime +/- lines. There are no differing runtime lines to list. runtime-comparison.json contains both complete line lists; runtime.patch is the re-cut patch against the pinned base. adamic.h +2/-0; exceptions.c +15/-2; object.c +2/-0. Newer main context (thread-local pending errors, OrdinaryToPrimitive, atomic cache publication) remains untouched in the own diff.

## Counts

Took base counts.md, then regenerated exactly once on Linux with house clang 20.1.8, after focused compilation/tests passed: `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts`. Exit 0. Added 46 owned rows: 35 callback, three miscompile controls/Error, five Weak, three self-compare. No existing row changed or disappeared. All 46 added rows also match the source commit numerically (counts-source-comparison.json). counts-attribution.json and counts-rows.md list every row and number. This exact base has counts-only counts.md: main already moved run digests into keyed measurements; its header directs the command used. No old schema or Mac table was introduced.

## Guards and execution conditions

SortedMap, exact runtime-set witnesses, literal native-agreement registration, every-test parallel/explanation, runtime field-layout and placeless split-build guards pass. No compiler-consuming package or runtime feature set was added, so existing compiler-dependencies and runtime-set declarations suffice. All new parallel tests call t.Parallel first. Four serial export preparation leaves carry direct Not parallel explanations. TestCallTargetReaders used -timeout=90s. Vet covered all seven changed Go packages.

GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh. nproc=5; cgroup cpu.max=400000 100000. Runners use four-CPU affinity and GOMAXPROCS=4. Setup passed, no workaround needed; its complete timing lines follow:

```text
setup: go ready (0.039s)
setup: submodules ready (0.084s)
setup: node ready (0.090s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.028s
setup: markdown dependencies ready (0.160s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.477s)
setup: shared cache ready (0.686s)
setup: go build ready (67.184s)
setup: test binaries deferred (use --warm-tests) (67.287s)
setup: build cache warm (67.289s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=1.16 0.25 0.08 1/139 854 load-after=3.11 0.84 0.28 2/173 2223
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (67.316s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.JYxaLB
```

## Focused commands

All test output was written to logs, never piped. Wall seconds include command compilation; individual leaf seconds are in *-seconds.json. Every long command had a hard timeout. No test over its 60-second share was rerun.

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
| parallel | 0 | 4.353 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| maps | 0 | 20.592 | `go test ./internal/ir -run ^TestEveryMapAProgramHoldsIsSorted$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 1 | 1.976 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 0 | 10.348 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| agreement-registry | 0 | 9.319 | `go test ./internal/lower -run ^TestNativeAgreementRowsNameRegisteredPrograms$ -count=1 -json -timeout=90s -parallel=4` |
| ir | 0 | 2.777 | `go test ./internal/ir -run ^Test(CallbackAdapterTargets|ClosureTargetsBoundOnlyProvenValues)$ -count=1 -json -timeout=90s -parallel=4` |
| lower | 0 | 2.931 | `go test ./internal/lower -run ^Test(LibraryRegexOffsetRequiresNoCaptures|ClassStaticInitializerCallIsEmitted)$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized0 | 0 | 13.199 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized1 | 0 | 13.904 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized2 | 0 | 13.325 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized2$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release0 | 0 | 9.706 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release1 | 0 | 9.833 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release2 | 0 | 10.067 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease2$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted0 | 0 | 11.092 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted1 | 0 | 10.268 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted2 | 0 | 10.439 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted2$ -count=1 -json -timeout=90s -parallel=4` |
| oracle | 0 | 14.733 | `go test ./internal/oracle -run ^Test(CallbackWidening|WeakDifferential|Miscompile2A|SelfCompare) -count=1 -json -timeout=90s -parallel=4` |
| native | 0 | 19.174 | `go test ./internal/native -run ^Test(ErrorUndefinedMessageHasNoOwnProperty|SplitSelfCompareAgreesWithNode)$ -count=1 -json -timeout=90s -parallel=4` |
| flow | 0 | 19.25 | `go test ./internal/flow -run ^Test(CallbackWidening|Slice2) -count=1 -json -timeout=90s -parallel=4` |
| flow-corpus | 0 | 5.405 | `go test ./internal/flow -run ^TestFlowCorpus -count=1 -json -timeout=90s -parallel=4` |
| fresh | 0 | 6.515 | `go test ./internal/fresh -run ^TestCallbackWidening -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 18.361 | `go test ./stage3/fixtures -run ^Test(FixturesSelfCompare|FixtureDirectoriesHaveTopLevelTests) -count=1 -json -timeout=90s -parallel=4` |
| native-guards | 0 | 3.748 | `go test ./internal/native -run ^Test(RuntimeFieldLayoutsAreIncluded|SplitBuildsNameNoPlaceOfTheirInputs)$ -count=1 -json -timeout=90s -parallel=4` |
| counts | 0 | 58.513 | `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts` |
| counts-table | 0 | 3.185 | `go test ./internal/oracle -run ^TestCountsAreRecordedTable$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Native | 0 | 11.443 | `go test ./internal/ir -run ^TestCallTargetExportInputsNative$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Lower | 0 | 6.684 | `go test ./internal/ir -run ^TestCallTargetExportInputsLower$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Fresh | 0 | 6.602 | `go test ./internal/ir -run ^TestCallTargetExportInputsFresh$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Flow | 0 | 6.968 | `go test ./internal/ir -run ^TestCallTargetExportInputsFlow$ -count=1 -json -timeout=90s -parallel=4` |
| readers | 0 | 5.578 | `go test ./internal/ir -run ^TestCallTargetReaders -count=1 -json -timeout=90s -parallel=4` |
| parallel | 0 | 1.067 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| lower | 0 | 3.389 | `go test ./internal/lower -run ^Test(LibraryRegexOffsetRequiresNoCaptures|ClassStaticInitializerCallIsEmitted|LibrarySpreadErrorView|NestedCallbackCycleIsRefused|ParameterPropertyCallbackReceiver|PredicateCallbackContracts|EveryNeedsCallbackEffects|ReadinessErrorIncludesReceiverExpression|InheritedLibraryReadsNeverLoadOwnFieldsError)$ -count=1 -json -timeout=90s -parallel=4` |
| vet | 0 | 4.354 | `go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/fresh ./internal/oracle ./stage3/fixtures` |
| primitive-main | 0 | 24.084 | `go test ./internal/native -run ^TestOrdinaryToPrimitive$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Native | 0 | 5.732 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesNative$ -count=1 -json -timeout=90s -parallel=4` |
| slot-pairs-main | 0 | 10.059 | `go test ./internal/native -run ^TestParallelSlotPairs$/^(asan|release)$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Lower | 0 | 4.865 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesLower$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Fresh | 0 | 4.522 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFresh$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Flow | 0 | 4.705 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFlow$ -count=1 -json -timeout=90s -parallel=4` |

## Direct compiler consumers

reader-packages.json records all 50 direct consumers discovered from Go imports, including test imports. Seven are covered by the focused checks above. The other 43 use these explicit selectors, avoiding whole suites. Packages without tests are compiled. TestOriginalCycleLedger skips without ADAMIC_CYCLE_LEDGER_ROOT; that corpus is not claimed as run. Setup also compiled the entire repository. No cohere code was copied.

| Package | Selector | Exit | Wall seconds | Observations |
|---|---|---:|---:|---|
| ./bench/regex | `^$` | 0 | 0.468 | compile only, no tests |
| ./bridge/tsgo | `^TestBridgeUnitsCoverEveryPiece$` | 0 | 8.388 | TestBridgeUnitsCoverEveryPiece: pass 0s |
| ./cloud/reports/decode-ascii | `^TestRuntimeFeaturesMatchNode$` | 0 | 12.834 | TestRuntimeFeaturesMatchNode: pass 5.97s |
| ./cmd/adamic | `^TestUsageExitStatus$` | 0 | 8.198 | TestUsageExitStatus: pass 0s |
| ./cmd/adamic-meter | `^TestOptionalPositionCountsUTF16Surrogates$` | 0 | 8.434 | TestOptionalPositionCountsUTF16Surrogates: pass 0.02s |
| ./cmd/adamic-test262 | `^TestAdaptSingletonArrayCallback$` | 0 | 8.278 | TestAdaptSingletonArrayCallback: pass 0s |
| ./internal/compilerinputs | `^TestLoweringIsACopy$` | 0 | 3.471 | TestLoweringIsACopy: pass 0s |
| ./internal/fuzz | `^TestOneSeedOneProgram$` | 0 | 7.571 | TestOneSeedOneProgram: pass 0.02s |
| ./internal/javascript | `^TestViewMixedUnionSelectsNumber$` | 0 | 1.529 | TestViewMixedUnionSelectsNumber: pass 0.06s |
| ./internal/leakcheck | `^TestCountedBalance$` | 0 | 0.766 | TestCountedBalance: pass 0s |
| ./internal/lower/cycleledger | `^TestOriginalCycleLedger$` | 0 | 53.035 | TestOriginalCycleLedger: skip 0s |
| ./internal/native/runtimeproducts | `^TestTheJobserverNeverReachesAnArchivesBytes$` | 0 | 12.647 | TestTheJobserverNeverReachesAnArchivesBytes: pass 11.03s |
| ./stage1/cohere/css | `^TestCompositionMatchesGoUnion$` | 0 | 61.285 | TestCompositionMatchesGoUnion: pass 53.6s |
| ./stage1/cohere/cssnumbers | `^TestCSSNumbersAnswerProductsRekey$` | 0 | 7.946 | TestCSSNumbersAnswerProductsRekey: pass 0.36s |
| ./stage1/cohere/cssstrings | `^TestMultiPushGap$` | 0 | 8.67 | TestMultiPushGap: pass 0.12s |
| ./stage1/cohere/estree | `^TestAcceptanceMutantsUnion$` | 0 | 8.413 | TestAcceptanceMutantsUnion: pass 0s |
| ./stage1/cohere/formatfiles | `^TestExtAgreesWithGo$` | 0 | 8.488 | TestExtAgreesWithGo/bare_dot_i: pass 0s, TestExtAgreesWithGo/single_dot: pass 0s, TestExtAgreesWithGo/double_dot: pass 0s, TestExtAgreesWithGo/no_dot: pass 0s, TestExtAgreesWithGo/trailing_dot: pass 0s, TestExtAgreesWithGo/last_of_multiple_dots: pass 0s, TestExtAgreesWithGo/absolute_nested_dotfile: pass 0s, TestExtAgreesWithGo/absolute_root_dotfile: pass 0s, TestExtAgreesWithGo/root: pass 0s, TestExtAgreesWithGo/empty: pass 0s, TestExtAgreesWithGo/relative_nested_dotfile: pass 0s, TestExtAgreesWithGo/dot_before_separator: pass 0s, TestExtAgreesWithGo/nested_trailing_dot: pass 0s, TestExtAgreesWithGo/bare_Unicode_dotfile: pass 0s, TestExtAgreesWithGo/Unicode_extension: pass 0s, TestExtAgreesWithGo/dot_at_zero_without_slash: pass 0s, TestExtAgreesWithGo/escaped_extension: pass 0s, TestExtAgreesWithGo: pass 0.33s |
| ./stage1/cohere/gitignore | `^TestPortAnswerPreparedUnitPlan$` | 0 | 7.92 | TestPortAnswerPreparedUnitPlan: pass 0s |
| ./stage1/cohere/graphql | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 8.518 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_error_made_elsewhere.ts: pass 0.12s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_throwing_function_value.ts: pass 1.26s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/graphql/printer | `^TestPrinterConstructorGap$` | 0 | 8.586 | TestPrinterConstructorGap: pass 0.22s |
| ./stage1/cohere/json | `^TestJSONClassOrderAnswerKey$` | 0 | 9.405 | TestJSONClassOrderAnswerKey: pass 0.24s |
| ./stage1/cohere/lint | `^TestDecodedOptionsAndMutantUnion$` | 0 | 11.428 | TestDecodedOptionsAndMutantUnion: pass 0s |
| ./stage1/cohere/lint/helpers | `^TestHelperMutantsUnion$` | 0 | 9.482 | TestHelperMutantsUnion: pass 0.98s |
| ./stage1/cohere/lint/helpers/comments | `^TestCommentMutantsUnion$` | 0 | 7.294 | TestCommentMutantsUnion: pass 0s |
| ./stage1/cohere/lint/regex | `^TestDynamicPatternGap$` | 0 | 8.799 | TestDynamicPatternGap: pass 0.21s |
| ./stage1/cohere/lint/rules/no-underscore-dangle | `^TestCompileProfiles$` | 0 | 37.251 | TestCompileProfiles: pass 29.52s |
| ./stage1/cohere/lint/rules/no-unsafe-negation | `^TestCompileProfiles$` | 0 | 35.175 | TestCompileProfiles: pass 25.83s |
| ./stage1/cohere/lint/rules/no-unsafe-optional-chaining | `^TestCompileProfiles$` | 0 | 28.943 | TestCompileProfiles: pass 22.63s |
| ./stage1/cohere/lint/rules/typescript-no-this-alias | `^TestCompileProfiles$` | 0 | 28.033 | TestCompileProfiles: pass 22.11s |
| ./stage1/cohere/markdownblocks | `^TestArrayGrowthWitness$` | 0 | 8.368 | TestArrayGrowthWitness: pass 1.05s |
| ./stage1/cohere/markdowninline | `^TestDelimiterExpressionMatchesNode$` | 0 | 7.866 | TestDelimiterExpressionMatchesNode: pass 1.12s |
| ./stage1/cohere/mediaquery | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 8.037 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_return_undefined.ts: pass 1.24s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_generic_return.ts: pass 1.24s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/5_local_children_array.ts: pass 1.28s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/3_optional_chaining_array.ts: pass 1.28s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_generic_parameter.ts: pass 0.27s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/4_mutable_child_array.ts: pass 0.3s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/selector | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 7.959 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/6_undefined_case.ts: pass 0.18s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/3_optional_boolean.ts: pass 0.18s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/4_array_from_iterator.ts: pass 0.18s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_multiple_push.ts: pass 0.12s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_mixed_field.ts: pass 0.14s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/5_conditional_panic.ts: pass 1.27s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/static_single_assignment | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 6.451 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_block_arrow_returning_class.ts: pass 0.14s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_call_through_optional_chain.ts: pass 0.15s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/suppression | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 7.615 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_return_undefined_array.ts: pass 1.19s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_mutual_recursion.ts: pass 1.2s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/tsprinter | `^TestClosedOptionalBooleanFunctionGap$` | 0 | 8.073 | TestClosedOptionalBooleanFunctionGap: pass 1.12s |
| ./stage1/cohere/typeaware | `^TestSixShardUnionRejectsLossAndDuplication$` | 0 | 6.853 | TestSixShardUnionRejectsLossAndDuplication: pass 0s |
| ./stage1/cohere/values | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 7.874 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_index_of_position.ts: pass 1.37s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/5_empty_array_default.ts: pass 1.37s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/3_case_constant.ts: pass 1.38s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/4_private_method.ts: pass 1.38s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_boolean_or_undefined_field.ts: pass 0.1s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/yaml | `^TestComposeMatchGoUnion$` | 0 | 7.066 | TestComposeMatchGoUnion: pass 0.47s |
| ./stage1/typescript/parser | `^TestCompilerShardsHoldEveryFileOnce$` | 0 | 6.656 | TestCompilerShardsHoldEveryFileOnce: skip 0s |
| ./stage1/typescript/scanner | `^TestGapStandsWhereGapsMdSays$` | 0 | 6.394 | TestGapStandsWhereGapsMdSays: pass 0.09s |
| ./stage3/census/tool | `^$` | 0 | 1.626 | compile only, no tests |
| ./stage3/drivers/parser/probe | `^$` | 0 | 1.219 | compile only, no tests |

## Mutants and catchers

External mutant commands, overlays and full logs are retained under mutants/. Go mutant sources have .go.txt names. A build failure, timeout or compiler warning does not count as a semantic kill. Original refusal-revert is expressed narrowly as the omitted adapter plus its recorded eight refusal assertions, rather than restoring five obsolete compiler files over main. The isolated Weak callback probe disables that stop and uses this pinned base's object.go to omit the adapter, retaining main's newer unrelated object lowering.

| Mutant | Caught | Wall seconds | Observed catcher markers |
|---|---|---:|---|
| error-spread | True | 15.478 | want path-bearing stop |
| maybe-setter | True | 15.655 | want path-bearing stop |
| long-name | True | 15.324 | want path-bearing stop |
| computed-name | True | 18.903 | Unhandled case in Node.Text |
| constructor-capture | True | 18.577 | want path-bearing stop |
| optional-error | True | 47.429 | member access within null pointer |
| error-own-message | True | 33.961 | native ; Node  |
| method-views | True | 20.205 | both backends unexpectedly admit the pinned view NotYet; release: exit=0 stdout="t1 none; release: stdout differs; AddressSanitizer |
| union-liveness | True | 22.406 | release: exit codes differ; sanitized: exit codes differ; runtime error: member access within null pointer |
| weak-callback | True | 18.972 | both backends must stop at |
| adaptation-omitted | True | 19.17 | native |
| nullable-token-omitted | True | 19.851 | exit |
| unrepresented-parameter-admitted | True | 19.941 | expected unrepresented parameter refusal |
| adapter-target-omitted | True | 15.399 | Unknown:true |
| split-numeric-fold | True | 18.022 | native: <nil>; Node  |
| refusal-revert | True | 13.848 | --- FAIL: |
| weak-callback-isolated | True | 19.695 | both backends unexpectedly admit the pinned callback NotYet; AddressSanitizer: heap-buffer-overflow |
| weak-counter-registration-omitted | True | 8.276 | want the row of |
| static-inline-definition | True | 17.224 | inline function 'adamic_unit_adamic_same_number' is not defined |
| fixture-owner-omitted | True | 6.772 | has no top-level test |
| reader-allowlist-omitted | True | 9.755 | unapproved call-target read |

TestSelfCompareConstantMutant is the embedded original semantic probe: the mutated NaN comparison compiles and executes successfully, and the oracle observes wrong stdout; the test passes only when that mutant is killed. All 17 original semantic probes and four extended semantic/guard probes were caught. Negative probe leaves also stayed below 60 seconds (maximum 36.64 seconds); mutant-test-seconds.json records them. The static-inline-definition probe is structural only, caught by clang, and receives no semantic credit. Refusal boundary probes distinguish conservative rejection from accidental admission; sanitizer probes independently establish the unsafe behavior where applicable.

## Individual affected test timings

| Test | Seconds | Check |
|---|---:|---|
| TestNativeAgreementRowsNameRegisteredPrograms | 0.01 | agreement-registry-seconds |
| TestCountsAreRecordedTable | 0 | counts-table-seconds |
| TestFixtureDirectoriesHaveTopLevelTests | 0 | fixtures-seconds |
| TestFixturesSelfCompareArrayFill | 10.36 | fixtures-seconds |
| TestFixturesSelfCompareSweep | 10.35 | fixtures-seconds |
| TestFixturesSelfCompareClosure | 10.36 | fixtures-seconds |
| TestFlowCorpusUnitsCoverEveryProgram | 0.01 | flow-corpus-seconds |
| TestFlowCorpusSetupIsShared | 0.05 | flow-corpus-seconds |
| TestFlowCorpusRemainder | 2.69 | flow-corpus-seconds |
| TestCallbackWideningNullableObjectMap | 0.2 | flow-seconds |
| TestCallbackWideningObjectGeneric | 0.21 | flow-seconds |
| TestCallbackWideningLibraryMap | 0.22 | flow-seconds |
| TestSlice2WeakSingle | 0.24 | flow-seconds |
| TestCallbackWideningObjectForEach | 0.19 | flow-seconds |
| TestCallbackWideningObjectFind | 0.2 | flow-seconds |
| TestCallbackWideningObjectEvery | 0.19 | flow-seconds |
| TestCallbackWideningObjectFilter | 0.23 | flow-seconds |
| TestCallbackWideningNumberToSorted | 0.18 | flow-seconds |
| TestCallbackWideningNumberThrow | 0.23 | flow-seconds |
| TestCallbackWideningNumberSort | 0.23 | flow-seconds |
| TestCallbackWideningNumberSome | 0.21 | flow-seconds |
| TestCallbackWideningNumberReduce | 0.17 | flow-seconds |
| TestCallbackWideningNumberReduceAccumulator | 0.22 | flow-seconds |
| TestCallbackWideningOptionalNumberMap | 0.2 | flow-seconds |
| TestCallbackWideningNumberNamedSort | 0.2 | flow-seconds |
| TestCallbackWideningBooleanMap | 0.18 | flow-seconds |
| TestCallbackWideningObjectGenericMap | 0.19 | flow-seconds |
| TestCallbackWideningNumberArguments | 0.21 | flow-seconds |
| TestCallbackWideningNumberGenericMap | 0.23 | flow-seconds |
| TestCallbackWideningObjectSort | 0.18 | flow-seconds |
| TestCallbackWideningObjectSome | 0.19 | flow-seconds |
| TestCallbackWideningObjectMap | 0.19 | flow-seconds |
| TestCallbackWideningObjectReduce | 0.23 | flow-seconds |
| TestCallbackWideningNumberFindIndex | 0.18 | flow-seconds |
| TestCallbackWideningNumberMap | 0.18 | flow-seconds |
| TestCallbackWideningNumberIndex | 0.21 | flow-seconds |
| TestCallbackWideningNumberForEach | 0.19 | flow-seconds |
| TestCallbackWideningNumberGeneric | 0.22 | flow-seconds |
| TestCallbackWideningNumberFindLast | 0.17 | flow-seconds |
| TestSlice2WeakBooleanFilter | 0.19 | flow-seconds |
| TestSlice2WeakHeldUnion | 0.18 | flow-seconds |
| TestSlice2WeakMatchingMethods | 0.21 | flow-seconds |
| TestSlice2WeakUnion | 0.18 | flow-seconds |
| TestCallbackWideningNumberEvery | 0.18 | flow-seconds |
| TestCallbackWideningNumberFind | 0.19 | flow-seconds |
| TestSlice2UndefinedError | 0.19 | flow-seconds |
| TestCallbackWideningNumberFilter | 0.19 | flow-seconds |
| TestSlice2MiscompileControls | 0.2 | flow-seconds |
| TestSlice2OptionalError | 0.18 | flow-seconds |
| TestCallbackWideningOptionalObjectMap | 0.19 | flow-seconds |
| TestCallbackWideningNumberCapture | 0.19 | flow-seconds |
| TestCallbackWideningStringMap | 0.14 | flow-seconds |
| TestCallbackWideningLibraryMap | 0.07 | fresh-seconds |
| TestCallbackWideningNumberIndex | 0.08 | fresh-seconds |
| TestCallbackWideningNumberGenericMap | 0.08 | fresh-seconds |
| TestCallbackWideningObjectEvery | 0.09 | fresh-seconds |
| TestCallbackWideningOptionalNumberMap | 0.05 | fresh-seconds |
| TestCallbackWideningNullableObjectMap | 0.06 | fresh-seconds |
| TestCallbackWideningStringMap | 0.06 | fresh-seconds |
| TestCallbackWideningOptionalObjectMap | 0.08 | fresh-seconds |
| TestCallbackWideningNumberArguments | 0.04 | fresh-seconds |
| TestCallbackWideningBooleanMap | 0.05 | fresh-seconds |
| TestCallbackWideningObjectGenericMap | 0.03 | fresh-seconds |
| TestCallbackWideningObjectMap | 0.04 | fresh-seconds |
| TestCallbackWideningObjectSome | 0.04 | fresh-seconds |
| TestCallbackWideningObjectReduce | 0.04 | fresh-seconds |
| TestCallbackWideningNumberFindIndex | 0.04 | fresh-seconds |
| TestCallbackWideningObjectSort | 0.06 | fresh-seconds |
| TestCallbackWideningNumberForEach | 0.04 | fresh-seconds |
| TestCallbackWideningNumberGeneric | 0.04 | fresh-seconds |
| TestCallbackWideningNumberFindLast | 0.06 | fresh-seconds |
| TestCallbackWideningNumberSome | 0.06 | fresh-seconds |
| TestCallbackWideningNumberThrow | 0.05 | fresh-seconds |
| TestCallbackWideningNumberToSorted | 0.05 | fresh-seconds |
| TestCallbackWideningNumberSort | 0.04 | fresh-seconds |
| TestCallbackWideningObjectForEach | 0.05 | fresh-seconds |
| TestCallbackWideningObjectFind | 0.04 | fresh-seconds |
| TestCallbackWideningObjectGeneric | 0.05 | fresh-seconds |
| TestCallbackWideningNumberFilter | 0.04 | fresh-seconds |
| TestCallbackWideningNumberFind | 0.04 | fresh-seconds |
| TestCallbackWideningNumberEvery | 0.04 | fresh-seconds |
| TestCallbackWideningNumberReduceAccumulator | 0.03 | fresh-seconds |
| TestCallbackWideningObjectFilter | 0.06 | fresh-seconds |
| TestCallbackWideningNumberCapture | 0.03 | fresh-seconds |
| TestCallbackWideningNumberMap | 0.04 | fresh-seconds |
| TestCallbackWideningNumberReduce | 0.05 | fresh-seconds |
| TestCallbackWideningNumberNamedSort | 0.05 | fresh-seconds |
| TestCallbackAdapterTargets | 0 | ir-seconds |
| TestClosureTargetsBoundOnlyProvenValues | 0 | ir-seconds |
| TestClassStaticInitializerCallIsEmitted | 0.07 | lower-seconds |
| TestEveryNeedsCallbackEffects | 0.1 | lower-seconds |
| TestNestedCallbackCycleIsRefused | 0.11 | lower-seconds |
| TestParameterPropertyCallbackReceiver | 0.08 | lower-seconds |
| TestLibrarySpreadErrorView | 0.07 | lower-seconds |
| TestReadinessErrorIncludesReceiverExpression | 0.13 | lower-seconds |
| TestLibraryRegexOffsetRequiresNoCaptures | 0.29 | lower-seconds |
| TestPredicateCallbackContracts | 0 | lower-seconds |
| TestInheritedLibraryReadsNeverLoadOwnFieldsError | 0.03 | lower-seconds |
| TestEveryMapAProgramHoldsIsSorted | 0 | maps-seconds |
| TestMiscompile2AConstructorArrow | 3.26 | mutant-test-seconds |
| TestMiscompile2AConstructorNumberArrow | 3.26 | mutant-test-seconds |
| TestCallbackWideningNullableObjectMapRelease | 3.42 | mutant-test-seconds |
| TestCallbackWideningNullableObjectMapSanitized | 3.53 | mutant-test-seconds |
| TestWeakDifferentialUnionNarrowed | 4.33 | mutant-test-seconds |
| TestWeakDifferentialSort | 3.21 | mutant-test-seconds |
| TestWeakDifferentialEach | 3.25 | mutant-test-seconds |
| TestWeakDifferentialReduce | 3.25 | mutant-test-seconds |
| TestWeakDifferentialSome | 3.3 | mutant-test-seconds |
| TestWeakDifferentialMap | 0.17 | mutant-test-seconds |
| TestWeakDifferentialFind | 0.19 | mutant-test-seconds |
| TestMiscompile2ALongName | 2.77 | mutant-test-seconds |
| TestCallbackWideningRefuseSort | 3.96 | mutant-test-seconds |
| TestCallbackWideningRefuseSome | 3.97 | mutant-test-seconds |
| TestCallbackWideningRefuseReduce | 4.0 | mutant-test-seconds |
| TestCallbackWideningRefuseMap | 4.03 | mutant-test-seconds |
| TestCallbackWideningRefuseEvery | 0.22 | mutant-test-seconds |
| TestCallbackWideningRefuseForEach | 0.23 | mutant-test-seconds |
| TestCallbackWideningRefuseFind | 0.22 | mutant-test-seconds |
| TestCallbackWideningRefuseFilter | 0.19 | mutant-test-seconds |
| TestMiscompile2AOverrideSource | 3.87 | mutant-test-seconds |
| TestCallbackWideningUnrepresentedParameter | 3.54 | mutant-test-seconds |
| TestSplitSelfCompareAgreesWithNode | 1.41 | mutant-test-seconds |
| TestMiscompile2AMaybeSetter | 2.81 | mutant-test-seconds |
| TestCallbackAdapterTargets | 0.0 | mutant-test-seconds |
| TestCountsAreRecordedTable | 0.0 | mutant-test-seconds |
| TestCallbackWideningNumberEveryRelease | 3.78 | mutant-test-seconds |
| TestCallbackWideningNumberEverySanitized | 3.86 | mutant-test-seconds |
| TestErrorUndefinedMessageHasNoOwnProperty | 9.05 | mutant-test-seconds |
| TestMiscompile2AErrorSpread | 2.91 | mutant-test-seconds |
| TestMiscompile2AErrorSubclassSpread | 2.91 | mutant-test-seconds |
| TestWeakDifferentialViewMethod | 3.75 | mutant-test-seconds |
| TestWeakDifferentialViewMethodParameter | 3.77 | mutant-test-seconds |
| TestWeakDifferentialSort | 3.34 | mutant-test-seconds |
| TestWeakDifferentialReduce | 3.39 | mutant-test-seconds |
| TestWeakDifferentialEach | 3.46 | mutant-test-seconds |
| TestWeakDifferentialSome | 3.48 | mutant-test-seconds |
| TestWeakDifferentialMap | 0.52 | mutant-test-seconds |
| TestWeakDifferentialFind | 0.53 | mutant-test-seconds |
| TestMiscompile2AOptionalError | 19.61 | mutant-test-seconds |
| TestMiscompile2AMarkerUndefined | 36.64 | mutant-test-seconds |
| TestSplitSelfCompareAgreesWithNode | 1.24 | mutant-test-seconds |
| TestRuntimeFieldLayoutsAreIncluded | 0 | native-guards-seconds |
| TestSplitBuildsNameNoPlaceOfTheirInputs | 1.06 | native-guards-seconds |
| TestSplitSelfCompareAgreesWithNode | 4.82 | native-seconds |
| TestErrorUndefinedMessageHasNoOwnProperty | 7.46 | native-seconds |
| TestSelfCompareConstantMutant | 2.78 | oracle-seconds |
| TestCallbackWideningNullableObjectMapJavaScript | 2.87 | oracle-seconds |
| TestCallbackWideningNumberEveryRelease | 2.92 | oracle-seconds |
| TestMiscompile2AUndefinedError | 0.38 | oracle-seconds |
| TestSelfCompareArrayFill | 0.54 | oracle-seconds |
| TestMiscompile2AMarkerUndefined | 1.34 | oracle-seconds |
| TestMiscompile2AMarker | 1.33 | oracle-seconds |
| TestMiscompile2AErrorSubclassSpread | 0.15 | oracle-seconds |
| TestMiscompile2AConstructorNumberArrow | 0.16 | oracle-seconds |
| TestSelfCompareClosure | 2.23 | oracle-seconds |
| TestSelfCompareSweep | 5.13 | oracle-seconds |
| TestMiscompile2AConstructorArrow | 0.19 | oracle-seconds |
| TestMiscompile2AOverrideSource | 0.21 | oracle-seconds |
| TestMiscompile2ADerivedSymbol | 0.23 | oracle-seconds |
| TestMiscompile2ALongName | 0.25 | oracle-seconds |
| TestMiscompile2AMaybeSetter | 0.18 | oracle-seconds |
| TestMiscompile2AErrorSpread | 0.17 | oracle-seconds |
| TestCallbackWideningUnrepresentedParameter | 0.16 | oracle-seconds |
| TestMiscompile2AOptionalError | 0.44 | oracle-seconds |
| TestWeakDifferentialSome | 0.19 | oracle-seconds |
| TestWeakDifferentialHeldUnionJavaScript | 0.23 | oracle-seconds |
| TestWeakDifferentialHeldUnionRelease | 0.13 | oracle-seconds |
| TestCallbackWideningOptionalObjectMap | 0.59 | oracle-seconds |
| TestWeakDifferentialHeldUnionSanitized | 0.35 | oracle-seconds |
| TestMiscompile2AControls | 1.62 | oracle-seconds |
| TestWeakDifferentialBooleanFilterJavaScript | 0.24 | oracle-seconds |
| TestWeakDifferentialBooleanFilterRelease | 0.31 | oracle-seconds |
| TestWeakDifferentialSort | 0.22 | oracle-seconds |
| TestWeakDifferentialBooleanFilterSanitized | 0.48 | oracle-seconds |
| TestCallbackWideningNullableObjectMapRelease | 0.16 | oracle-seconds |
| TestWeakDifferentialMatchingMethods | 0.63 | oracle-seconds |
| TestCallbackWideningNullableObjectMapSanitized | 0.29 | oracle-seconds |
| TestCallbackWideningObjectSort | 0.64 | oracle-seconds |
| TestCallbackWideningOptionalNumberMap | 0.56 | oracle-seconds |
| TestCallbackWideningStringMap | 0.48 | oracle-seconds |
| TestCallbackWideningBooleanMap | 0.54 | oracle-seconds |
| TestCallbackWideningObjectGenericMap | 0.53 | oracle-seconds |
| TestCallbackWideningNumberGenericMap | 0.53 | oracle-seconds |
| TestCallbackWideningNumberToSorted | 0.52 | oracle-seconds |
| TestCallbackWideningNumberFindIndex | 0.51 | oracle-seconds |
| TestCallbackWideningNumberFindLast | 0.58 | oracle-seconds |
| TestCallbackWideningLibraryMap | 0.55 | oracle-seconds |
| TestCallbackWideningNumberArguments | 1.61 | oracle-seconds |
| TestCallbackWideningNumberThrow | 0.62 | oracle-seconds |
| TestCallbackWideningNumberCapture | 0.6 | oracle-seconds |
| TestCallbackWideningNumberReduceAccumulator | 0.62 | oracle-seconds |
| TestWeakDifferentialEach | 0.21 | oracle-seconds |
| TestWeakDifferentialReduce | 0.19 | oracle-seconds |
| TestCallbackWideningNumberIndex | 0.58 | oracle-seconds |
| TestWeakDifferentialMap | 0.17 | oracle-seconds |
| TestWeakDifferentialFind | 0.2 | oracle-seconds |
| TestCallbackWideningNumberNamedSort | 0.59 | oracle-seconds |
| TestCallbackWideningNumberSort | 0.56 | oracle-seconds |
| TestCallbackWideningObjectSome | 0.51 | oracle-seconds |
| TestCallbackWideningObjectMap | 0.54 | oracle-seconds |
| TestCallbackWideningObjectReduce | 0.62 | oracle-seconds |
| TestCallbackWideningObjectGeneric | 0.59 | oracle-seconds |
| TestCallbackWideningObjectForEach | 0.57 | oracle-seconds |
| TestCallbackWideningObjectFind | 0.55 | oracle-seconds |
| TestCallbackWideningObjectFilter | 0.59 | oracle-seconds |
| TestWeakDifferentialViewMethod | 0.21 | oracle-seconds |
| TestCallbackWideningObjectEvery | 0.54 | oracle-seconds |
| TestWeakDifferentialUnionNarrowed | 0.52 | oracle-seconds |
| TestWeakDifferentialSingleNarrowed | 0.48 | oracle-seconds |
| TestCallbackWideningNumberForEach | 0.49 | oracle-seconds |
| TestCallbackWideningNumberSome | 0.53 | oracle-seconds |
| TestCallbackWideningNumberReduce | 0.5 | oracle-seconds |
| TestCallbackWideningNumberMap | 0.52 | oracle-seconds |
| TestWeakDifferentialViewMethodParameter | 0.21 | oracle-seconds |
| TestCallbackWideningNumberEveryJavaScript | 0.09 | oracle-seconds |
| TestCallbackWideningNumberGeneric | 0.57 | oracle-seconds |
| TestCallbackWideningNumberFilter | 0.58 | oracle-seconds |
| TestCallbackWideningNumberFind | 0.56 | oracle-seconds |
| TestCallbackWideningNumberEverySanitized | 0.23 | oracle-seconds |
| TestEveryTestIsParallelOrSaysWhy | 0.33 | parallel-seconds |
| TestOrdinaryToPrimitive | 20.4 | primitive-main-seconds |
| TestProduct_CallTargetReaderArchivesFlow | 1.74 | reader-archive-Flow-seconds |
| TestProduct_CallTargetReaderArchivesFresh | 1.72 | reader-archive-Fresh-seconds |
| TestProduct_CallTargetReaderArchivesLower | 1.75 | reader-archive-Lower-seconds |
| TestProduct_CallTargetReaderArchivesNative | 2.03 | reader-archive-Native-seconds |
| TestCallTargetExportInputsFlow | 4.23 | reader-key-Flow-seconds |
| TestCallTargetExportInputsFresh | 3.92 | reader-key-Fresh-seconds |
| TestCallTargetExportInputsLower | 4.14 | reader-key-Lower-seconds |
| TestCallTargetExportInputsNative | 4.46 | reader-key-Native-seconds |
| TestCallTargetReadersFresh | 2.61 | readers-seconds |
| TestCallTargetReadersNative | 2.74 | readers-seconds |
| TestCallTargetReadersFlow | 2.79 | readers-seconds |
| TestCallTargetReadersLower | 2.81 | readers-seconds |
| TestProduct_NativeRuntimesCounted0 | 8.25 | runtime-Counted0-seconds |
| TestProduct_NativeRuntimesCounted1 | 7.64 | runtime-Counted1-seconds |
| TestProduct_NativeRuntimesCounted2 | 7.87 | runtime-Counted2-seconds |
| TestProduct_NativeRuntimesRelease0 | 7.16 | runtime-Release0-seconds |
| TestProduct_NativeRuntimesRelease1 | 7.35 | runtime-Release1-seconds |
| TestProduct_NativeRuntimesRelease2 | 7.57 | runtime-Release2-seconds |
| TestProduct_NativeRuntimesSanitized0 | 10.62 | runtime-Sanitized0-seconds |
| TestProduct_NativeRuntimesSanitized1 | 11.25 | runtime-Sanitized1-seconds |
| TestProduct_NativeRuntimesSanitized2 | 10.56 | runtime-Sanitized2-seconds |
| TestRuntimeSetsAreEachUsed | 1.03 | runtime-sets-seconds |
| TestParallelSlotPairs | 6.61 | slot-pairs-main-seconds |
