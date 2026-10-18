Validation command ledger

Environment: source /workspace/adamic-tools/env.sh; export GOPROXY='https://proxy.golang.org|direct' GOMAXPROCS=4 ADAMIC_GOCACHE_OFF=1; unset GOCACHEPROG.

Initial Git operations:
```sh
timeout 90s git fetch origin
git checkout -b compiler/loom-r-records dbdf977f5b8b0979d414ebe6f20e7aadd312e309
timeout 90s git fetch --recurse-submodules=no origin 9f9ccacf28d326e072b7adfbd9c10245e6901c38
git cherry-pick --no-commit 9f9ccacf28d326e072b7adfbd9c10245e6901c38
git checkout --ours internal/flow/flow_test.go
git add internal/flow/flow_test.go
git restore --source=dbdf977f5b --staged --worktree internal/oracle/counts.md
```

The eight git mv operations are enumerated exactly in fixture-moves.json. Oracle path edits and leakcheck adaptation are the only manual source changes outside the conflict resolution. Counts use the harness.

oracle-build: exit 1, 0.721s
```sh
timeout 150s go test -c -o /workspace/scratch/loom-r-records-oracle.test ./internal/oracle
```

lower: exit 0, 10.302s
```sh
timeout 150s go test ./internal/lower -run '^(TestRecord.*|TestNamedRecord.*|TestPartialRecord.*|TestDetachedOwn.*|TestEntries.*|TestOptionalIndexing.*)$' -count=1 -timeout 90s -v -parallel 4
```

oracle-build-adapted: exit 0, 7.562s
```sh
timeout 150s go test -c -o /workspace/scratch/loom-r-records-oracle.test ./internal/oracle
```

runtimes: exit 0, 91.31s
```sh
timeout 150s go test ./internal/oracle -run '^TestProduct_NativeRuntimes(Sanitized|Release|Counted)[0-2]$' -count=1 -timeout 90s -v -parallel 4
```

oracle-records: exit 0, 17.169s
```sh
timeout 150s go test ./internal/oracle -run '^(TestRecord.*|TestNamedRecord.*|TestPartialRecord.*|TestDetachedOwn.*|TestJSONReplacerDuplicates.*|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$' -count=1 -timeout 90s -v -parallel 4
```

oracle-node: exit 0, 33.449s
```sh
timeout 150s go test ./internal/oracle -run '^TestNativeAgreesWithNode_[0-9]+$/^internal$/^oracle$/^testdata$/^(records_.*|detached_own_.*|json_replacer_duplicates_object\.a)$' -count=1 -timeout 90s -v -parallel 4
```

native-records: exit 0, 99.369s
```sh
timeout 150s go test ./internal/native -run '^(TestProduct_RecordMutants|TestRecordsAgainstNode|TestRecordMutantsUnit0[0-5]|TestRecordReadMutantsUnit0[0-2]|TestJSONReplacerDuplicates.*Mutant)$' -count=1 -timeout 90s -v -parallel 4
```

stage3-products: exit 0, 61.738s
```sh
timeout 150s go test ./stage3/fixtures -run '^(TestProduct_FixtureOracleHook|TestProduct_FixtureRuntimesSanitized[0-2])$' -count=1 -timeout 90s -v -parallel 4
```

estree-product: exit 0, 68.12s
```sh
timeout 150s go test ./stage1/cohere/estree/oracles -run '^TestProduct_EstreeGoOracle$' -count=1 -timeout 90s -v -parallel 4
```

counts-update: exit 1, 181.871s
```sh
timeout 240s go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 180s -parallel 4 -args -update-counts
```

vet: exit 0, 74.82s
```sh
timeout 180s go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle
```

stage3: exit 0, 90.221s
```sh
timeout 150s go test ./stage3/fixtures -run '^TestFixtures(Records|Objects|Taste)$' -count=1 -timeout 90s -v -parallel 4
```

flow: exit 0, 123.422s
```sh
timeout 150s go test ./internal/flow -run '^(TestFlowCorpusRemainder|TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared)$' -count=1 -timeout 90s -v -parallel 4
```

guards: exit 1, 12.801s
```sh
timeout 150s go test ./cmd/adamic-gate ./internal/oracle -run '^(TestEveryTestIsParallelOrSaysWhy|TestEveryDuplicatedProductDeclarationAgrees|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestCountsAreRecordedTable)$' -count=1 -timeout 90s -v -parallel 4
```

readers: exit 0, 69.592s
```sh
timeout 150s go test ./internal/ir -run '^(TestProduct_CallTargetExports|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted)$' -count=1 -timeout 90s -v -parallel 4
```

stage3-api-setup: exit 0, 3.276s
```sh
timeout 120s npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
```

counts-mutant-product: exit 0, 97.581s
```sh
timeout 150s go test ./internal/oracle -run '^TestProduct_CountMeasurements13$' -count=1 -timeout 90s -v -parallel 4
```

command-consumer-cache-explicit: exit 0, 23.058s
```sh
timeout 150s go test ./cmd/adamic -run '^(TestProduct_CommandRuntimeSanitized|TestExplainChecksOutput)$' -count=1 -timeout 90s -v -parallel 4
```

stage3-typescript-setup: exit 0, 29.555s
```sh
timeout 120s git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/loom-r-records-typescript
```

counts-update-complete: exit 0, 281.203s
```sh
timeout 660s go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 600s -parallel 4 -args -update-counts
```

lane-with-pinned-inputs: exit 1, 88.498s
```sh
timeout 150s go test ./stage3/lane -run '^(TestShardsMatchManifest|TestApplyPython|TestLanePython)$' -count=1 -timeout 90s -v -parallel 4
```

counts-table-final: exit 0, 4.361s
```sh
timeout 150s go test ./internal/oracle -run '^TestCountsAreRecordedTable$' -count=1 -timeout 90s -v -parallel 4
```

source-mutants: exit 124, 480.039s
```sh
timeout 480s python3 review/compiler/loom-records-2/run-source-mutants.py
```

named-mutants: exit 0, 48.5s
```sh
timeout 240s python3 review/compiler/loom-records-2/run-named-mutants.py
```

extra-mutants: exit 0, 228.855s
```sh
timeout 480s python3 review/compiler/loom-r-records/extra-mutants.py
```

type-only-index-admission-final: exit 1, 19.339s
```sh
timeout 150s go test -overlay /workspace/adamic/review/compiler/loom-r-records/source-mutants/type-only-index-admission.json ./internal/lower -run '^TestRecordForms$' -count=1 -timeout 90s -v -parallel 4
```

consumer-oracle-owners: exit 0, 25.544s
```sh
timeout 150s go test ./stage1/cohere/lint/helpers/comments/oracles ./stage1/cohere/lint/regex/oracles -run '^(TestProduct_CommentsOracle|TestProduct_RegexOptionDialects)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-comments-regex-final: exit 0, 4.519s
```sh
timeout 150s go test ./stage1/cohere/lint/helpers/comments ./stage1/cohere/lint/regex -run '^(TestJsxParserGapIsExplicit|TestDynamicPatternGap|TestOptionDialectGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./bench/regex: exit 0, 1.482s
```sh
go test ./bench/regex -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./bridge/tsgo: exit 0, 37.901s
```sh
go test ./bridge/tsgo -run '^(TestTSGoRequiresLink)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./cloud/reports/decode-ascii: exit 0, 41.342s
```sh
go test ./cloud/reports/decode-ascii -run '^(TestRuntimeFeaturesMatchNode)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./cmd/adamic: exit 1, 37.641s
```sh
go test ./cmd/adamic -run '^(TestExplainChecksOutput)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./cmd/adamic-fuzz: exit 0, 2.234s
```sh
go test ./cmd/adamic-fuzz -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./cmd/adamic-meter: exit 0, 24.895s
```sh
go test ./cmd/adamic-meter -run '^(TestOptionalPositionCountsUTF16Surrogates)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./cmd/adamic-reduce: exit 0, 0.87s
```sh
go test ./cmd/adamic-reduce -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./cmd/adamic-test262: exit 0, 22.728s
```sh
go test ./cmd/adamic-test262 -run '^(TestAdaptSingletonArrayCallback)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/compilerinputs: exit 0, 6.62s
```sh
go test ./internal/compilerinputs -run '^(TestCheckKeyTellsTheCompilerFromTscWhole)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/fresh: exit 0, 19.88s
```sh
go test ./internal/fresh -run '^(TestFreshCorpusRemainder)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/fuzz: exit 0, 46.346s
```sh
go test ./internal/fuzz -run '^(TestGeneratedProgramsCheckAndLower)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/javascript: exit 0, 3.91s
```sh
go test ./internal/javascript -run '^(TestViewMixedUnionSelectsNumber)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/lower/cycleledger: exit 0, 106.834s
```sh
go test ./internal/lower/cycleledger -run '^(TestOriginalCycleLedger)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/native/runtimeproducts: exit 0, 1.596s
```sh
go test ./internal/native/runtimeproducts -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/css: exit 0, 40.157s
```sh
go test ./stage1/cohere/css -run '^(TestEachGapStandsWhereGapsMdSaysItDoes|TestClosedEmptyArrayUnionGap|TestClosedParserRegexGap|TestClosedOptionalBooleanConditionGap|TestClosedPrinterRegexGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/cssnumbers: exit 0, 13.454s
```sh
go test ./stage1/cohere/cssnumbers -run '^(TestCSSNumbersAnswerProductsRekey)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/cssstrings: exit 0, 14.503s
```sh
go test ./stage1/cohere/cssstrings -run '^(TestMultiPushGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/estree: exit 0, 32.929s
```sh
go test ./stage1/cohere/estree -run '^(TestPostfixValueGap|TestMethodReplacementGap|TestRawInputGap|TestParserRecoveryGap|TestInterfaceTypeMethodGap|TestRecoveryLibraryGaps|TestPinnedNumericGaps|TestTypeMemberLibraryGap|TestCookedSurrogateLibraryGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/formatfiles: exit 0, 5.406s
```sh
go test ./stage1/cohere/formatfiles -run '^(TestExtAgreesWithGo)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/gitignore: exit 0, 7.361s
```sh
go test ./stage1/cohere/gitignore -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/graphql: exit 0, 8.177s
```sh
go test ./stage1/cohere/graphql -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/graphql/printer: exit 0, 5.527s
```sh
go test ./stage1/cohere/graphql/printer -run '^(TestPrinterConstructorGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/json: exit 0, 9.36s
```sh
go test ./stage1/cohere/json -run '^(TestDocumentedStageZeroGaps|TestClosedEmptyFallbackGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint: exit 0, 8.702s
```sh
go test ./stage1/cohere/lint -run '^(TestClosedComparatorGaps|TestNestedConstructorGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint/helpers: exit 0, 8.81s
```sh
go test ./stage1/cohere/lint/helpers -run '^(TestKnownGapsAreExplicit)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint/helpers/comments: exit 1, 24.408s
```sh
go test ./stage1/cohere/lint/helpers/comments -run '^(TestJsxParserGapIsExplicit)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint/regex: exit 1, 10.042s
```sh
go test ./stage1/cohere/lint/regex -run '^(TestDynamicPatternGap|TestOptionDialectGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint/rules/no-underscore-dangle: exit 0, 45.421s
```sh
go test ./stage1/cohere/lint/rules/no-underscore-dangle -run '^(TestCompileProfiles)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint/rules/no-unsafe-negation: exit 0, 36.169s
```sh
go test ./stage1/cohere/lint/rules/no-unsafe-negation -run '^(TestCompileProfiles)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/lint/rules/no-unsafe-optional-chaining: exit 0, 30.598s
```sh
go test ./stage1/cohere/lint/rules/no-unsafe-optional-chaining -run '^(TestCompileProfiles)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/boundedrun: exit 0, 1.299s
```sh
go test ./internal/boundedrun -run '^(TestDeadlineKillsGrandchild)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./cmd/adamic-admission-delta: exit 0, 2.438s
```sh
go test ./cmd/adamic-admission-delta -run '^(TestClassification|TestDiffCoverage|TestShardPartitionKeepsDiff)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/compilercensus: exit 0, 9.165s
```sh
go test ./internal/compilercensus -run '^(TestEveryCCompilerRunsThroughNative|TestDependenciesFollowTheMarkers|TestDependenciesCoverTheDeclaredMap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./cmd/adamic-gocacheprog: exit 0, 3.252s
```sh
go test ./cmd/adamic-gocacheprog -run '^(TestNoToken)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/buildcache/cmd/productidentity: exit 0, 0.826s
```sh
go test ./internal/buildcache/cmd/productidentity -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/buildcache/cmd/traced: exit 0, 2.054s
```sh
go test ./internal/buildcache/cmd/traced -run '^(TestSettleParsesOnlyAskedCalls)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/buildcache/goplan: exit 0, 0.699s
```sh
go test ./internal/buildcache/goplan -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/tracked: exit 0, 5.249s
```sh
go test ./internal/tracked -run '^(TestChangedNamesEveryDifference)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./internal/toolchain: exit 0, 1.257s
```sh
go test ./internal/toolchain -run '^(TestAToolsCompilersAreClangUnlessTheCallerNamesOne)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./stage1/cohere/executable: exit 0, 4.007s
```sh
go test ./stage1/cohere/executable -run '^(TestWriteLeavesNoWritableDescriptorForSiblingForksToExec)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-extra-results.json ./stage3/lane: exit 1, 6.586s
```sh
go test ./stage3/lane -run '^(TestShardsMatchManifest|TestApplyPython|TestLanePython)$' -count=1 -timeout 90s -v -parallel 4
```

The source/named Python scripts enumerate the inner overlay commands and expected refusal witnesses. extra-mutants.json records each extra mutant command; its environment is explicit in extra-mutants.py. Raw exits of 1 are expected for caught mutants, not green baseline runs. The aggregate source runner timeout is completed by type-only-index-admission-final.

Additional preparation commands and environments: ADAMIC_BUILD_CACHE_DIR=/home/agent/.cache/adamic-build for command-consumer-cache-explicit; ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/loom-r-records-typescript for lane-with-pinned-inputs. No test output was piped.

vet-fixtures: exit 0, 0.316s
```sh
timeout 90s go vet ./stage3/fixtures
```

consumer-results.json ./stage1/cohere/lint/rules/typescript-no-this-alias: exit 0, 30.252s
```sh
go test ./stage1/cohere/lint/rules/typescript-no-this-alias -run '^(TestCompileProfiles)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/markdownblocks: exit 0, 7.939s
```sh
go test ./stage1/cohere/markdownblocks -run '^(TestArrayGrowthWitness)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/markdowninline: exit 0, 7.924s
```sh
go test ./stage1/cohere/markdowninline -run '^(TestDelimiterExpressionMatchesNode)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/mediaquery: exit 0, 8.415s
```sh
go test ./stage1/cohere/mediaquery -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/selector: exit 0, 9.062s
```sh
go test ./stage1/cohere/selector -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/static_single_assignment: exit 0, 6.99s
```sh
go test ./stage1/cohere/static_single_assignment -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/suppression: exit 0, 6.874s
```sh
go test ./stage1/cohere/suppression -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/tsprinter: exit 0, 8.214s
```sh
go test ./stage1/cohere/tsprinter -run '^(TestClassInterfaceMethodGap|TestClosedOptionalBooleanFunctionGap|TestCompilerGaps|TestClosedPrefixUpdateValueGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/typeaware: exit 0, 6.853s
```sh
go test ./stage1/cohere/typeaware -run '^(TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/values: exit 0, 7.123s
```sh
go test ./stage1/cohere/values -run '^(TestEachGapStandsWhereGapsMdSaysItDoes)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/cohere/yaml: exit 0, 10.386s
```sh
go test ./stage1/cohere/yaml -run '^(TestLexerGaps|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/typescript/parser: exit 0, 8.26s
```sh
go test ./stage1/typescript/parser -run '^(TestStrongAstParentGap|TestPushSpreadGap|TestClassMethodInterfaceGap|TestClosedOptionalFunctionValueGap|TestClosedConditionalEmptyArrayGap)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage1/typescript/scanner: exit 0, 5.546s
```sh
go test ./stage1/typescript/scanner -run '^(TestGapStandsWhereGapsMdSays|TestBigintGapStandsWhereGapsMdSays)$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage3/census/tool: exit 0, 0.164s
```sh
go test ./stage3/census/tool -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./stage3/drivers/parser/probe: exit 0, 0.265s
```sh
go test ./stage3/drivers/parser/probe -run '^$' -count=1 -timeout 90s -v -parallel 4
```

consumer-results.json ./internal/leakcheck: exit 0, 0.267s
```sh
go test ./internal/leakcheck -run '^(TestReportedLeakIsRejected|TestCountedBalance)$' -count=1 -timeout 90s -v -parallel 4
```

Final Git operations (executed after the evidence is staged):
```sh
git add -u
git add review/compiler/loom-r-records
git diff --cached --check
timeout 60s git commit -F review/compiler/loom-r-records/commit-message.txt
timeout 90s git push origin HEAD:refs/heads/compiler/loom-r-records
timeout 30s git ls-remote --heads origin refs/heads/compiler/loom-r-records
```
