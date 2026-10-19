Built: slice 2 plus naming, static-inventory and a-check fixes as one commit on compiler/loom-r-slice-2-2.
Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 4ca997c7a13f67e01591ab1722e9cc02ba90a4ee; delivery SHA in final response/git log.
Commands and outputs: both reported reds reproduced, corrected guards pass; focused tests, compiler consumers, counts, reader guards and vet rerun below.
Mutants: original semantic/guard probes and structural inline probe rerun; three new fix probes and six live header probes recorded below.
Not covered: full package suites/full gate, ahra or Loom submission, macOS/WASI/TSan; cycle-ledger original corpus skipped without its configured source tree.

## Reproduction and fixes

Fetched origin (main advanced to 99c369e0), but used exactly dbdf977f5b as explicitly requested. Created compiler/loom-r-slice-2-2 from that commit and applied only 4ca997c7's own diff with cherry-pick --no-commit. No application conflicts. The earlier slice's lowering, runtime and infrastructure resolutions are retained; the historical report is review/compiler/loom-r-slice-2/REPORT.md.

The exact reported tests were run before editing. TestEveryTestHasOneDescriptiveName failed with 35 colliding callback identifier families across internal/flow, internal/fresh and internal/oracle. TestRuntimeStaticsAreListed failed on exceptions.c:18, empty_message. names-before.log and statics-before.log retain the raw observations.

Renamed 35 slice-owned flow tests to TestFlowCallbackWidening... and 35 fresh tests to TestFreshCallbackWidening... . Kept oracle names, all checks and the naming rule/pinned exceptions unchanged. renames.json lists all 70 exact old/new identifiers. No tests were deleted, and no other package's tests were edited to make room.

Added only exceptions.c:empty_message:1 to internal/native/testdata/runtime-statics/storage.txt, next to main's existing error_name entry, with the protection audit. Its C type is not const, so the scanner requires review, but it has no mutable shared state: ADAMIC_STRING initializes references=0, slab=0, units=0, index=ADAMIC_LITERAL_INDEX. heap.c retain/drop paths leave a zero count unchanged; share.c skips zero-count non-region identities; string_index.c excludes immortal literals from units/index writes; string_append.c requires reference_count==1 for in-place writes. This follows main's existing immortal-literal convention and requires neither a lock nor thread-local storage. The unsafe-shared-state stop condition does not apply.

Strengthened the slice's existing TestErrorUndefinedMessageHasNoOwnProperty C harness to call string-unit lookup and share, then verify its literal's references, slab, units and index remain unchanged. The immutable-units mutant changes the real runtime cache-write guard and is caught by this assertion, not a compiler warning. An initial negative run returned through the assertion without freeing the temporary Error, so LeakSanitizer masked its expected exit 4; that attempt is retained and NOT credited. The harness now releases the Error before its intentional failure exit, and the final probe observes exit 4 with no sanitizer masking (fix-probes-final.log).

## Runtime clearance

NO runtime source line changes relative to 4ca997c7. internal/native/runtime/{adamic.h,exceptions.c,object.c} are byte-identical; all other runtime files are untouched. runtime-clearance.json records this. The only new runtime-related production-tree edit is the inventory key and its audit comment; the other edit is test-only C embedded in the slice's existing Go test. The original slice's runtime +/- lines remain exactly the original +19/-2 over three files, unchanged from the previous report.

## A-check

Built this branch's cmd/adamic and invoked `adamic c <path>` on every one of the 63 added/modified .a programs. Final results: 46 clean, 14 NotYet, three Refused, zero checker errors. The 46 clean programs also have counted native builds in the full counts refresh; source fixtures are held to Node and sanitizers by the focused tests. Clean and NotYet programs carry no a-check error header. All source program bodies are byte-identical to 4ca997c7 after removing an old or new first-line a-check header.

Exactly five header changes, fully recorded with unchanged-body SHA256 in header-changes.json:
- Add `// a-check: refused this escaping a constructor through a closure before every field is set` to miscompile_2a_refused/d7054e9_this_arrow.a and d7054e9_this_arrow_number.a.
- Remove obsolete `// a-check: compile` from stage3/fixtures/self-compare/array-fill.a, closure.a and sweep.a.

The existing array_structural.a refusal header already matches its measured reason and remains unchanged. No type-error headers were needed. acheck-final.json holds every command outcome, reason/code, body hash and header; per-file compiler stderr is retained. No a-check classification relies on a Mac table or a source-file exemption.

For each of the three refusal programs, physically removed and independently corrupted its header, compiled it again with the real branch compiler, applied README's documented reason/code predicate to the fresh diagnostic, and restored the bytes in finally blocks. All six fail the header predicate as expected (header-mutants.json). These are documented-header-policy checks using fresh compiler diagnostics, not a claim of running ahra/Loom or an unavailable historical Gate.aCheck implementation. The historical README report's pinned run.py path is absent from that object in this checkout; no gate implementation was modified.

## Counts, guards and execution conditions

Regenerated counts once after the header corrections with the Linux house clang 20.1.8. counts.md is byte-identical to 4ca997c7: no rows added, removed or changed in this fix. The complete slice still adds its original 46 owned rows versus dbdf977f5b. Counts leaf maximum is 3.53 seconds; counts-leaf-seconds.json records each row.

SortedMap, exact runtime witnesses, literal native-agreement registration, parallel-test explanations, descriptive unique names, static inventory, runtime field layouts and placeless split-build guards pass. No compiler-consuming package or runtime feature set was added. Reader guards use -timeout=90s; vet covers all seven changed Go packages. Four serial export preparation leaves keep direct Not parallel explanations; renamed tests call t.Parallel first.

All test output went to log files. Every long command had a hard timeout. Tests ran with GOMAXPROCS=4 and four-CPU affinity. nproc=5; cgroup cpu.max=400000 100000. Setup passed without workaround, using GOPROXY=https://proxy.golang.org|direct and the printed env.sh. Complete setup timing lines:

```text
setup: node ready (0.036s)
setup: go ready (0.042s)
setup: submodules ready (0.054s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.011s
setup: markdown dependencies ready (0.092s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.210s)
setup: shared cache ready (0.968s)
setup: go build ready (15.545s)
setup: test binaries deferred (use --warm-tests) (15.642s)
setup: build cache warm (15.643s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.08 0.02 0.01 1/134 694 load-after=0.90 0.20 0.06 1/139 1348
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (15.671s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.sj022b
```

## Focused commands and results

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
| names | 0 | 1.071 | `go test ./cmd/adamic-gate -run ^TestEveryTestHasOneDescriptiveName$ -count=1 -json -timeout=90s -parallel=4` |
| statics | 0 | 12.797 | `go test ./internal/native -run ^TestRuntimeStaticsAreListed$ -count=1 -json -timeout=90s -parallel=4` |
| parallel | 0 | 1.028 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| maps | 0 | 2.532 | `go test ./internal/ir -run ^TestEveryMapAProgramHoldsIsSorted$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 0 | 3.569 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| agreement-registry | 0 | 2.528 | `go test ./internal/lower -run ^TestNativeAgreementRowsNameRegisteredPrograms$ -count=1 -json -timeout=90s -parallel=4` |
| oracle | 0 | 15.061 | `go test ./internal/oracle -run ^Test(CallbackWidening|WeakDifferential|Miscompile2A|SelfCompare) -count=1 -json -timeout=90s -parallel=4` |
| native | 0 | 3.894 | `go test ./internal/native -run ^Test(ErrorUndefinedMessageHasNoOwnProperty|SplitSelfCompareAgreesWithNode)$ -count=1 -json -timeout=90s -parallel=4` |
| ir | 0 | 2.952 | `go test ./internal/ir -run ^Test(CallbackAdapterTargets|ClosureTargetsBoundOnlyProvenValues)$ -count=1 -json -timeout=90s -parallel=4` |
| flow | 0 | 19.091 | `go test ./internal/flow -run ^Test(FlowCallbackWidening|Slice2) -count=1 -json -timeout=90s -parallel=4` |
| flow-corpus | 0 | 5.245 | `go test ./internal/flow -run ^TestFlowCorpus -count=1 -json -timeout=90s -parallel=4` |
| fresh | 0 | 6.813 | `go test ./internal/fresh -run ^TestFreshCallbackWidening -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 7.443 | `go test ./stage3/fixtures -run ^Test(FixturesSelfCompare|FixtureDirectoriesHaveTopLevelTests) -count=1 -json -timeout=90s -parallel=4` |
| native-guards | 0 | 3.886 | `go test ./internal/native -run ^Test(RuntimeFieldLayoutsAreIncluded|SplitBuildsNameNoPlaceOfTheirInputs)$ -count=1 -json -timeout=90s -parallel=4` |
| counts | 0 | 86.561 | `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts` |
| counts-table | 0 | 3.948 | `go test ./internal/oracle -run ^TestCountsAreRecordedTable$ -count=1 -json -timeout=90s -parallel=4` |
| lower | 0 | 4.839 | `go test ./internal/lower -run ^Test(LibraryRegexOffsetRequiresNoCaptures|ClassStaticInitializerCallIsEmitted|LibrarySpreadErrorView|NestedCallbackCycleIsRefused|ParameterPropertyCallbackReceiver|PredicateCallbackContracts|EveryNeedsCallbackEffects|ReadinessErrorIncludesReceiverExpression|InheritedLibraryReadsNeverLoadOwnFieldsError)$ -count=1 -json -timeout=90s -parallel=4` |
| primitive-main | 0 | 7.45 | `go test ./internal/native -run ^TestOrdinaryToPrimitive$ -count=1 -json -timeout=90s -parallel=4` |
| slot-pairs-main | 0 | 5.931 | `go test ./internal/native -run ^TestParallelSlotPairs$/^(asan|release)$ -count=1 -json -timeout=90s -parallel=4` |
| native | 0 | 14.148 | `go test ./internal/native -run ^Test(ErrorUndefinedMessageHasNoOwnProperty|SplitSelfCompareAgreesWithNode)$ -count=1 -json -timeout=90s -parallel=4` |
| names | 0 | 1.225 | `go test ./cmd/adamic-gate -run ^TestEveryTestHasOneDescriptiveName$ -count=1 -json -timeout=90s -parallel=4` |
| statics | 0 | 3.135 | `go test ./internal/native -run ^TestRuntimeStaticsAreListed$ -count=1 -json -timeout=90s -parallel=4` |
| oracle | 0 | 16.242 | `go test ./internal/oracle -run ^Test(CallbackWidening|WeakDifferential|Miscompile2A|SelfCompare) -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 7.708 | `go test ./stage3/fixtures -run ^Test(FixturesSelfCompare|FixtureDirectoriesHaveTopLevelTests) -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Native | 0 | 7.48 | `go test ./internal/ir -run ^TestCallTargetExportInputsNative$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Lower | 0 | 4.477 | `go test ./internal/ir -run ^TestCallTargetExportInputsLower$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Fresh | 0 | 7.075 | `go test ./internal/ir -run ^TestCallTargetExportInputsFresh$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Flow | 0 | 7.42 | `go test ./internal/ir -run ^TestCallTargetExportInputsFlow$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Native | 0 | 4.155 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesNative$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Lower | 0 | 4.065 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesLower$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Fresh | 0 | 4.044 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFresh$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Flow | 0 | 4.122 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFlow$ -count=1 -json -timeout=90s -parallel=4` |
| readers | 0 | 5.911 | `go test ./internal/ir -run ^TestCallTargetReaders -count=1 -json -timeout=90s -parallel=4` |
| parallel | 0 | 0.917 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| vet | 0 | 2.429 | `go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/fresh ./internal/oracle ./stage3/fixtures` |

## Compiler consumers

The seven compiler/fixture packages use the focused checks above; the other 43 direct consumers use the prior explicit focused plan again. Packages with no tests get compilation checks. The only skip is TestOriginalCycleLedger for absent ADAMIC_CYCLE_LEDGER_ROOT. This is not original compiler-corpus coverage. No cohere code was copied.

| Package | Selector | Exit | Wall seconds | Leaf outcomes |
|---|---|---:|---:|---|
| ./bench/regex | `^$` | 0 | 0.216 | compile only |
| ./bridge/tsgo | `^TestBridgeUnitsCoverEveryPiece$` | 0 | 2.956 | TestBridgeUnitsCoverEveryPiece: pass 0s |
| ./cloud/reports/decode-ascii | `^TestRuntimeFeaturesMatchNode$` | 0 | 3.757 | TestRuntimeFeaturesMatchNode: pass 1.09s |
| ./cmd/adamic | `^TestUsageExitStatus$` | 0 | 2.985 | TestUsageExitStatus: pass 0s |
| ./cmd/adamic-meter | `^TestOptionalPositionCountsUTF16Surrogates$` | 0 | 2.792 | TestOptionalPositionCountsUTF16Surrogates: pass 0.01s |
| ./cmd/adamic-test262 | `^TestAdaptSingletonArrayCallback$` | 0 | 2.841 | TestAdaptSingletonArrayCallback: pass 0s |
| ./internal/compilerinputs | `^TestLoweringIsACopy$` | 0 | 2.693 | TestLoweringIsACopy: pass 0s |
| ./internal/fuzz | `^TestOneSeedOneProgram$` | 0 | 2.963 | TestOneSeedOneProgram: pass 0.02s |
| ./internal/javascript | `^TestViewMixedUnionSelectsNumber$` | 0 | 0.718 | TestViewMixedUnionSelectsNumber: pass 0.06s |
| ./internal/leakcheck | `^TestCountedBalance$` | 0 | 0.57 | TestCountedBalance: pass 0s |
| ./internal/lower/cycleledger | `^TestOriginalCycleLedger$` | 0 | 3.404 | TestOriginalCycleLedger: skip 0s |
| ./internal/native/runtimeproducts | `^TestTheJobserverNeverReachesAnArchivesBytes$` | 0 | 10.068 | TestTheJobserverNeverReachesAnArchivesBytes: pass 8.99s |
| ./stage1/cohere/css | `^TestCompositionMatchesGoUnion$` | 0 | 25.137 | TestCompositionMatchesGoUnion: pass 21.14s |
| ./stage1/cohere/cssnumbers | `^TestCSSNumbersAnswerProductsRekey$` | 0 | 6.107 | TestCSSNumbersAnswerProductsRekey: pass 0.45s |
| ./stage1/cohere/cssstrings | `^TestMultiPushGap$` | 0 | 4.49 | TestMultiPushGap: pass 0.18s |
| ./stage1/cohere/estree | `^TestAcceptanceMutantsUnion$` | 0 | 4.706 | TestAcceptanceMutantsUnion: pass 0s |
| ./stage1/cohere/formatfiles | `^TestExtAgreesWithGo$` | 0 | 4.502 | TestExtAgreesWithGo/bare_dot_i: pass 0s, TestExtAgreesWithGo/single_dot: pass 0s, TestExtAgreesWithGo/double_dot: pass 0s, TestExtAgreesWithGo/no_dot: pass 0s, TestExtAgreesWithGo/trailing_dot: pass 0s, TestExtAgreesWithGo/last_of_multiple_dots: pass 0s, TestExtAgreesWithGo/absolute_nested_dotfile: pass 0s, TestExtAgreesWithGo/absolute_root_dotfile: pass 0s, TestExtAgreesWithGo/root: pass 0s, TestExtAgreesWithGo/empty: pass 0s, TestExtAgreesWithGo/relative_nested_dotfile: pass 0s, TestExtAgreesWithGo/dot_before_separator: pass 0s, TestExtAgreesWithGo/nested_trailing_dot: pass 0s, TestExtAgreesWithGo/bare_Unicode_dotfile: pass 0s, TestExtAgreesWithGo/Unicode_extension: pass 0s, TestExtAgreesWithGo/dot_at_zero_without_slash: pass 0s, TestExtAgreesWithGo/escaped_extension: pass 0s, TestExtAgreesWithGo: pass 0.25s |
| ./stage1/cohere/gitignore | `^TestPortAnswerPreparedUnitPlan$` | 0 | 5.101 | TestPortAnswerPreparedUnitPlan: pass 0s |
| ./stage1/cohere/graphql | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 6.263 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_error_made_elsewhere.ts: pass 0.29s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_throwing_function_value.ts: pass 1.9s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/graphql/printer | `^TestPrinterConstructorGap$` | 0 | 5.04 | TestPrinterConstructorGap: pass 0.2s |
| ./stage1/cohere/json | `^TestJSONClassOrderAnswerKey$` | 0 | 4.439 | TestJSONClassOrderAnswerKey: pass 0.19s |
| ./stage1/cohere/lint | `^TestDecodedOptionsAndMutantUnion$` | 0 | 5.308 | TestDecodedOptionsAndMutantUnion: pass 0s |
| ./stage1/cohere/lint/helpers | `^TestHelperMutantsUnion$` | 0 | 6.027 | TestHelperMutantsUnion: pass 1.82s |
| ./stage1/cohere/lint/helpers/comments | `^TestCommentMutantsUnion$` | 0 | 4.933 | TestCommentMutantsUnion: pass 0s |
| ./stage1/cohere/lint/regex | `^TestDynamicPatternGap$` | 0 | 5.104 | TestDynamicPatternGap: pass 0.22s |
| ./stage1/cohere/lint/rules/no-underscore-dangle | `^TestCompileProfiles$` | 0 | 33.957 | TestCompileProfiles: pass 28.77s |
| ./stage1/cohere/lint/rules/no-unsafe-negation | `^TestCompileProfiles$` | 0 | 27.476 | TestCompileProfiles: pass 24.66s |
| ./stage1/cohere/lint/rules/no-unsafe-optional-chaining | `^TestCompileProfiles$` | 0 | 23.855 | TestCompileProfiles: pass 21.53s |
| ./stage1/cohere/lint/rules/typescript-no-this-alias | `^TestCompileProfiles$` | 0 | 23.322 | TestCompileProfiles: pass 20.94s |
| ./stage1/cohere/markdownblocks | `^TestArrayGrowthWitness$` | 0 | 4.35 | TestArrayGrowthWitness: pass 1.21s |
| ./stage1/cohere/markdowninline | `^TestDelimiterExpressionMatchesNode$` | 0 | 4.996 | TestDelimiterExpressionMatchesNode: pass 1.85s |
| ./stage1/cohere/mediaquery | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 4.777 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_generic_return.ts: pass 1.52s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/5_local_children_array.ts: pass 1.54s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/4_mutable_child_array.ts: pass 1.56s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/3_optional_chaining_array.ts: pass 1.58s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_generic_parameter.ts: pass 0.27s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_return_undefined.ts: pass 0.3s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/selector | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 3.958 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_mixed_field.ts: pass 0.17s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/6_undefined_case.ts: pass 0.18s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/3_optional_boolean.ts: pass 0.18s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/4_array_from_iterator.ts: pass 0.12s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_multiple_push.ts: pass 0.12s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/5_conditional_panic.ts: pass 1.23s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/static_single_assignment | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 2.88 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_call_through_optional_chain.ts: pass 0.13s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_block_arrow_returning_class.ts: pass 0.14s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/suppression | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 3.797 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_return_undefined_array.ts: pass 1.03s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_mutual_recursion.ts: pass 1.03s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/tsprinter | `^TestClosedOptionalBooleanFunctionGap$` | 0 | 3.538 | TestClosedOptionalBooleanFunctionGap: pass 1.02s |
| ./stage1/cohere/typeaware | `^TestSixShardUnionRejectsLossAndDuplication$` | 0 | 2.579 | TestSixShardUnionRejectsLossAndDuplication: pass 0s |
| ./stage1/cohere/values | `^TestEachGapStandsWhereGapsMdSaysItDoes$` | 0 | 3.791 | TestEachGapStandsWhereGapsMdSaysItDoes/gaps/2_boolean_or_undefined_field.ts: pass 0.15s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/3_case_constant.ts: pass 1.02s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/1_index_of_position.ts: pass 1.2s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/5_empty_array_default.ts: pass 1.21s, TestEachGapStandsWhereGapsMdSaysItDoes/gaps/4_private_method.ts: pass 1.22s, TestEachGapStandsWhereGapsMdSaysItDoes: pass 0s |
| ./stage1/cohere/yaml | `^TestComposeMatchGoUnion$` | 0 | 3.033 | TestComposeMatchGoUnion: pass 0.49s |
| ./stage1/typescript/parser | `^TestCompilerShardsHoldEveryFileOnce$` | 0 | 2.433 | TestCompilerShardsHoldEveryFileOnce: skip 0s |
| ./stage1/typescript/scanner | `^TestGapStandsWhereGapsMdSays$` | 0 | 2.83 | TestGapStandsWhereGapsMdSays: pass 0.12s |
| ./stage3/census/tool | `^$` | 0 | 0.924 | compile only |
| ./stage3/drivers/parser/probe | `^$` | 0 | 0.922 | compile only |

## Mutants

All original semantic probes are rerun on this branch, including the embedded TestSelfCompareConstantMutant, which compiles/runs its NaN mutant and requires wrong stdout against Node. Static inline placement remains a structural-only clang catch and receives no semantic credit. No build failure or timeout is credited as a semantic catch. Full commands, overlays and logs are under this review directory; mutant Go sources have .go.txt extensions. The earlier narrow refusal-revert and isolated Weak probes retain main infrastructure instead of restoring obsolete compiler files.

| Mutant | Caught | Wall seconds | Catcher |
|---|---|---:|---|
| error-spread | True | 5.948 | want path-bearing stop |
| maybe-setter | True | 5.821 | want path-bearing stop |
| long-name | True | 6.054 | want path-bearing stop |
| computed-name | True | 5.925 | Unhandled case in Node.Text |
| constructor-capture | True | 6.223 | want path-bearing stop |
| optional-error | True | 7.282 | member access within null pointer |
| error-own-message | True | 16.869 | native ; Node  |
| method-views | True | 11.703 | both backends unexpectedly admit the pinned view NotYet; release: exit=0 stdout="t1 none; release: stdout differs; AddressSanitizer |
| union-liveness | True | 11.298 | release: exit codes differ; sanitized: exit codes differ; runtime error: member access within null pointer |
| weak-callback | True | 10.809 | both backends must stop at |
| adaptation-omitted | True | 10.617 | native |
| nullable-token-omitted | True | 10.497 | exit |
| unrepresented-parameter-admitted | True | 10.14 | expected unrepresented parameter refusal |
| adapter-target-omitted | True | 4.868 | Unknown:true |
| split-numeric-fold | True | 18.816 | native: <nil>; Node  |
| refusal-revert | True | 6.942 | --- FAIL: |
| weak-callback-isolated | True | 7.875 | both backends unexpectedly admit the pinned callback NotYet; AddressSanitizer: heap-buffer-overflow |
| weak-counter-registration-omitted | True | 3.296 | want the row of |
| static-inline-definition | True | 14.069 | inline function 'adamic_unit_adamic_same_number' is not defined |
| fixture-owner-omitted | True | 2.994 | has no top-level test |
| reader-allowlist-omitted | True | 7.497 | unapproved call-target read |
| colliding-flow-name | True | 1.017 | lives in two or more packages |
| unlisted-empty-message | True | 3.541 | unlisted mutable storage empty_message |
| immortal-units-write | True | 13.387 | exit status 4 |

Six additional live header-policy catches are detailed in header-mutants.json. The first masked immutability probe is recorded separately in fix-mutants-initial.json and fix-mutants/immortal-units-write-initial/test.log; final proof is the exit-4 assertion after cleanup.

## Per-test timing evidence

Every added/touched leaf stays below 60 seconds; no over-budget leaf was rerun. *-seconds.json files hold individual measurements, consumers.jsonl holds consumer leaves, and mutant-test-seconds.json holds negative probes. Wall seconds above include Go compilation and may exceed a leaf's measured duration.

| Test | Seconds | Check |
|---|---:|---|
| TestNativeAgreementRowsNameRegisteredPrograms | 0.01 | agreement-registry-seconds |
| TestCountsAreRecordedTable | 0 | counts-table-seconds |
| TestFixtureDirectoriesHaveTopLevelTests | 0.01 | fixtures-seconds |
| TestFixturesSelfCompareClosure | 3.01 | fixtures-seconds |
| TestFixturesSelfCompareSweep | 2.95 | fixtures-seconds |
| TestFixturesSelfCompareArrayFill | 2.99 | fixtures-seconds |
| TestFlowCorpusUnitsCoverEveryProgram | 0.01 | flow-corpus-seconds |
| TestFlowCorpusSetupIsShared | 0.05 | flow-corpus-seconds |
| TestFlowCorpusRemainder | 2.64 | flow-corpus-seconds |
| TestFlowCallbackWideningNumberNamedSort | 0.24 | flow-seconds |
| TestFlowCallbackWideningNumberFindLast | 0.24 | flow-seconds |
| TestFlowCallbackWideningLibraryMap | 0.27 | flow-seconds |
| TestFlowCallbackWideningObjectGeneric | 0.28 | flow-seconds |
| TestSlice2WeakHeldUnion | 0.22 | flow-seconds |
| TestSlice2WeakUnion | 0.2 | flow-seconds |
| TestSlice2WeakMatchingMethods | 0.25 | flow-seconds |
| TestSlice2WeakBooleanFilter | 0.23 | flow-seconds |
| TestSlice2WeakSingle | 0.21 | flow-seconds |
| TestSlice2OptionalError | 0.2 | flow-seconds |
| TestSlice2UndefinedError | 0.23 | flow-seconds |
| TestSlice2MiscompileControls | 0.23 | flow-seconds |
| TestFlowCallbackWideningOptionalObjectMap | 0.22 | flow-seconds |
| TestFlowCallbackWideningNullableObjectMap | 0.21 | flow-seconds |
| TestFlowCallbackWideningOptionalNumberMap | 0.21 | flow-seconds |
| TestFlowCallbackWideningStringMap | 0.23 | flow-seconds |
| TestFlowCallbackWideningBooleanMap | 0.2 | flow-seconds |
| TestFlowCallbackWideningNumberArguments | 0.21 | flow-seconds |
| TestFlowCallbackWideningNumberGenericMap | 0.2 | flow-seconds |
| TestFlowCallbackWideningObjectGenericMap | 0.22 | flow-seconds |
| TestFlowCallbackWideningObjectSort | 0.21 | flow-seconds |
| TestFlowCallbackWideningObjectSome | 0.19 | flow-seconds |
| TestFlowCallbackWideningObjectReduce | 0.19 | flow-seconds |
| TestFlowCallbackWideningObjectMap | 0.2 | flow-seconds |
| TestFlowCallbackWideningNumberToSorted | 0.2 | flow-seconds |
| TestFlowCallbackWideningObjectForEach | 0.2 | flow-seconds |
| TestFlowCallbackWideningObjectFind | 0.19 | flow-seconds |
| TestFlowCallbackWideningObjectFilter | 0.23 | flow-seconds |
| TestFlowCallbackWideningObjectEvery | 0.18 | flow-seconds |
| TestFlowCallbackWideningNumberThrow | 0.18 | flow-seconds |
| TestFlowCallbackWideningNumberSome | 0.21 | flow-seconds |
| TestFlowCallbackWideningNumberSort | 0.18 | flow-seconds |
| TestFlowCallbackWideningNumberFilter | 0.21 | flow-seconds |
| TestFlowCallbackWideningNumberReduce | 0.23 | flow-seconds |
| TestFlowCallbackWideningNumberFindIndex | 0.21 | flow-seconds |
| TestFlowCallbackWideningNumberFind | 0.19 | flow-seconds |
| TestFlowCallbackWideningNumberGeneric | 0.21 | flow-seconds |
| TestFlowCallbackWideningNumberIndex | 0.2 | flow-seconds |
| TestFlowCallbackWideningNumberMap | 0.23 | flow-seconds |
| TestFlowCallbackWideningNumberEvery | 0.21 | flow-seconds |
| TestFlowCallbackWideningNumberCapture | 0.17 | flow-seconds |
| TestFlowCallbackWideningNumberForEach | 0.16 | flow-seconds |
| TestFlowCallbackWideningNumberReduceAccumulator | 0.2 | flow-seconds |
| TestFreshCallbackWideningNumberIndex | 0.06 | fresh-seconds |
| TestFreshCallbackWideningLibraryMap | 0.07 | fresh-seconds |
| TestFreshCallbackWideningObjectEvery | 0.07 | fresh-seconds |
| TestFreshCallbackWideningNumberGenericMap | 0.08 | fresh-seconds |
| TestFreshCallbackWideningNumberForEach | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberFindLast | 0.05 | fresh-seconds |
| TestFreshCallbackWideningNumberGeneric | 0.07 | fresh-seconds |
| TestFreshCallbackWideningNumberFindIndex | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberFind | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberFilter | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberEvery | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberCapture | 0.05 | fresh-seconds |
| TestFreshCallbackWideningNumberSome | 0.05 | fresh-seconds |
| TestFreshCallbackWideningNumberToSorted | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberThrow | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberSort | 0.04 | fresh-seconds |
| TestFreshCallbackWideningObjectMap | 0.05 | fresh-seconds |
| TestFreshCallbackWideningObjectSome | 0.04 | fresh-seconds |
| TestFreshCallbackWideningObjectReduce | 0.04 | fresh-seconds |
| TestFreshCallbackWideningObjectSort | 0.07 | fresh-seconds |
| TestFreshCallbackWideningObjectForEach | 0.04 | fresh-seconds |
| TestFreshCallbackWideningObjectGeneric | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberReduceAccumulator | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberReduce | 0.04 | fresh-seconds |
| TestFreshCallbackWideningObjectFind | 0.04 | fresh-seconds |
| TestFreshCallbackWideningNumberNamedSort | 0.03 | fresh-seconds |
| TestFreshCallbackWideningNullableObjectMap | 0.04 | fresh-seconds |
| TestFreshCallbackWideningOptionalNumberMap | 0.04 | fresh-seconds |
| TestFreshCallbackWideningStringMap | 0.04 | fresh-seconds |
| TestFreshCallbackWideningOptionalObjectMap | 0.06 | fresh-seconds |
| TestFreshCallbackWideningBooleanMap | 0.04 | fresh-seconds |
| TestFreshCallbackWideningObjectGenericMap | 0.03 | fresh-seconds |
| TestFreshCallbackWideningNumberMap | 0.03 | fresh-seconds |
| TestFreshCallbackWideningNumberArguments | 0.03 | fresh-seconds |
| TestFreshCallbackWideningObjectFilter | 0.02 | fresh-seconds |
| TestCallbackAdapterTargets | 0 | ir-seconds |
| TestClosureTargetsBoundOnlyProvenValues | 0 | ir-seconds |
| TestClassStaticInitializerCallIsEmitted | 0.12 | lower-seconds |
| TestParameterPropertyCallbackReceiver | 0.1 | lower-seconds |
| TestNestedCallbackCycleIsRefused | 0.07 | lower-seconds |
| TestLibrarySpreadErrorView | 0.11 | lower-seconds |
| TestLibraryRegexOffsetRequiresNoCaptures | 0.29 | lower-seconds |
| TestReadinessErrorIncludesReceiverExpression | 0.14 | lower-seconds |
| TestEveryNeedsCallbackEffects | 0.08 | lower-seconds |
| TestPredicateCallbackContracts | 0 | lower-seconds |
| TestInheritedLibraryReadsNeverLoadOwnFieldsError | 0.06 | lower-seconds |
| TestEveryMapAProgramHoldsIsSorted | 0 | maps-seconds |
| TestEveryTestHasOneDescriptiveName | 0.29 | names-seconds |
| TestRuntimeFieldLayoutsAreIncluded | 0 | native-guards-seconds |
| TestSplitBuildsNameNoPlaceOfTheirInputs | 1 | native-guards-seconds |
| TestSplitSelfCompareAgreesWithNode | 1.11 | native-seconds |
| TestErrorUndefinedMessageHasNoOwnProperty | 1.13 | native-seconds |
| TestCallbackWideningNullableObjectMapJavaScript | 3.28 | oracle-seconds |
| TestWeakDifferentialHeldUnionJavaScript | 3.31 | oracle-seconds |
| TestCallbackWideningNumberEveryRelease | 3.32 | oracle-seconds |
| TestWeakDifferentialHeldUnionRelease | 0.15 | oracle-seconds |
| TestWeakDifferentialHeldUnionSanitized | 3.47 | oracle-seconds |
| TestWeakDifferentialBooleanFilterJavaScript | 0.25 | oracle-seconds |
| TestWeakDifferentialBooleanFilterRelease | 0.21 | oracle-seconds |
| TestWeakDifferentialSort | 0.18 | oracle-seconds |
| TestWeakDifferentialBooleanFilterSanitized | 0.47 | oracle-seconds |
| TestWeakDifferentialSome | 0.21 | oracle-seconds |
| TestWeakDifferentialReduce | 0.24 | oracle-seconds |
| TestWeakDifferentialMap | 0.24 | oracle-seconds |
| TestWeakDifferentialFind | 0.21 | oracle-seconds |
| TestWeakDifferentialEach | 0.2 | oracle-seconds |
| TestWeakDifferentialMatchingMethods | 0.79 | oracle-seconds |
| TestWeakDifferentialViewMethod | 0.2 | oracle-seconds |
| TestWeakDifferentialViewMethodParameter | 0.24 | oracle-seconds |
| TestWeakDifferentialSingleNarrowed | 0.61 | oracle-seconds |
| TestWeakDifferentialUnionNarrowed | 0.66 | oracle-seconds |
| TestSelfCompareArrayFill | 0.55 | oracle-seconds |
| TestMiscompile2AUndefinedError | 0.37 | oracle-seconds |
| TestSelfCompareConstantMutant | 1.37 | oracle-seconds |
| TestMiscompile2AMarker | 1.23 | oracle-seconds |
| TestMiscompile2AMarkerUndefined | 1.44 | oracle-seconds |
| TestSelfCompareClosure | 2.37 | oracle-seconds |
| TestSelfCompareSweep | 2.61 | oracle-seconds |
| TestMiscompile2AErrorSubclassSpread | 0.22 | oracle-seconds |
| TestMiscompile2AConstructorNumberArrow | 0.23 | oracle-seconds |
| TestMiscompile2AConstructorArrow | 0.22 | oracle-seconds |
| TestMiscompile2AOverrideSource | 0.21 | oracle-seconds |
| TestMiscompile2ADerivedSymbol | 0.18 | oracle-seconds |
| TestMiscompile2ALongName | 0.25 | oracle-seconds |
| TestMiscompile2AMaybeSetter | 0.2 | oracle-seconds |
| TestMiscompile2AErrorSpread | 0.17 | oracle-seconds |
| TestCallbackWideningUnrepresentedParameter | 0.19 | oracle-seconds |
| TestMiscompile2AOptionalError | 0.49 | oracle-seconds |
| TestCallbackWideningNullableObjectMapSanitized | 0.29 | oracle-seconds |
| TestCallbackWideningOptionalObjectMap | 0.54 | oracle-seconds |
| TestCallbackWideningObjectSort | 0.54 | oracle-seconds |
| TestCallbackWideningNullableObjectMapRelease | 0.17 | oracle-seconds |
| TestMiscompile2AControls | 1.87 | oracle-seconds |
| TestCallbackWideningNumberFindLast | 0.6 | oracle-seconds |
| TestCallbackWideningLibraryMap | 0.57 | oracle-seconds |
| TestCallbackWideningNumberThrow | 0.6 | oracle-seconds |
| TestCallbackWideningNumberCapture | 0.63 | oracle-seconds |
| TestCallbackWideningNumberReduceAccumulator | 0.61 | oracle-seconds |
| TestCallbackWideningNumberIndex | 0.63 | oracle-seconds |
| TestCallbackWideningNumberNamedSort | 0.57 | oracle-seconds |
| TestCallbackWideningOptionalNumberMap | 0.6 | oracle-seconds |
| TestCallbackWideningStringMap | 0.55 | oracle-seconds |
| TestCallbackWideningBooleanMap | 0.54 | oracle-seconds |
| TestCallbackWideningNumberGenericMap | 0.54 | oracle-seconds |
| TestCallbackWideningNumberToSorted | 0.52 | oracle-seconds |
| TestCallbackWideningObjectGenericMap | 0.57 | oracle-seconds |
| TestCallbackWideningNumberArguments | 1.65 | oracle-seconds |
| TestCallbackWideningNumberSort | 0.55 | oracle-seconds |
| TestCallbackWideningObjectSome | 0.53 | oracle-seconds |
| TestCallbackWideningObjectReduce | 0.62 | oracle-seconds |
| TestCallbackWideningObjectMap | 0.57 | oracle-seconds |
| TestCallbackWideningObjectGeneric | 0.58 | oracle-seconds |
| TestCallbackWideningObjectForEach | 0.62 | oracle-seconds |
| TestCallbackWideningObjectFind | 0.59 | oracle-seconds |
| TestCallbackWideningObjectFilter | 0.57 | oracle-seconds |
| TestCallbackWideningObjectEvery | 0.59 | oracle-seconds |
| TestCallbackWideningNumberForEach | 0.59 | oracle-seconds |
| TestCallbackWideningNumberSome | 0.58 | oracle-seconds |
| TestCallbackWideningNumberReduce | 0.57 | oracle-seconds |
| TestCallbackWideningNumberMap | 0.56 | oracle-seconds |
| TestCallbackWideningNumberGeneric | 0.58 | oracle-seconds |
| TestCallbackWideningNumberEveryJavaScript | 0.13 | oracle-seconds |
| TestCallbackWideningNumberFilter | 0.64 | oracle-seconds |
| TestCallbackWideningNumberEverySanitized | 0.28 | oracle-seconds |
| TestCallbackWideningNumberFind | 0.54 | oracle-seconds |
| TestCallbackWideningNumberFindIndex | 0.48 | oracle-seconds |
| TestEveryTestIsParallelOrSaysWhy | 0.23 | parallel-seconds |
| TestOrdinaryToPrimitive | 3.95 | primitive-main-seconds |
| TestProduct_CallTargetReaderArchivesFlow | 1.51 | reader-archive-Flow-seconds |
| TestProduct_CallTargetReaderArchivesFresh | 1.46 | reader-archive-Fresh-seconds |
| TestProduct_CallTargetReaderArchivesLower | 1.42 | reader-archive-Lower-seconds |
| TestProduct_CallTargetReaderArchivesNative | 1.61 | reader-archive-Native-seconds |
| TestCallTargetExportInputsFlow | 4.46 | reader-key-Flow-seconds |
| TestCallTargetExportInputsFresh | 4.21 | reader-key-Fresh-seconds |
| TestCallTargetExportInputsLower | 1.53 | reader-key-Lower-seconds |
| TestCallTargetExportInputsNative | 4.65 | reader-key-Native-seconds |
| TestCallTargetReadersFresh | 3.09 | readers-seconds |
| TestCallTargetReadersFlow | 3.11 | readers-seconds |
| TestCallTargetReadersNative | 3.21 | readers-seconds |
| TestCallTargetReadersLower | 3.27 | readers-seconds |
| TestRuntimeSetsAreEachUsed | 0.94 | runtime-sets-seconds |
| TestParallelSlotPairs | 2.33 | slot-pairs-main-seconds |
| TestRuntimeStaticsAreListed | 0.15 | statics-seconds |
