Built: rebased compiler slice 3 with optional reads after calls, spread calls, JSON shorthand, refusal rulings, assignment witnesses and TSGo caching.
Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source e3b5c979fffa8faafa2e179b2b071b489b8d0db9; delivery SHA is in the final response and git log.
Commands and outputs: all focused checks, reader checks, changed-package vet, one counts regeneration and counts attribution passed.
Mutants: all 14 production mutants failed their intended tests; all eight embedded mutants were caught by their comparisons.
Not covered: whole package suites, full gate, the pre-generated flow corpus outside the remainder, WASI, macOS execution and the real checker ABI.

## Application and conflicts

Created compiler/loom-r-slice-3 from exactly dbdf977f5b, after git fetch origin. Fetched the source SHA explicitly and applied git cherry-pick --no-commit e3b5c979fffa8faafa2e179b2b071b489b8d0db9. Only its own parent-to-commit change was applied. The base stays fixed even if origin/main advances.

Exactly two files conflicted, each with one marked hunk:

- internal/lower/refusals.go: main inserted parallelPreflight at the same preflight point where the slice inserted mergedFieldRulings and functionAnnotations. Kept parallelPreflight first, then mergedFieldRulings, then functionAnnotations, each returning its error. The auto-merged visitor retains main checks and the slice functionCallRefusal.
- internal/oracle/counts.md: main had five instantiation rows where the source added six TypeScript-mode ruling rows. Restored the entire base-main table, without hand merging any row, and regenerated it once after compilation and focused tests passed. The five instantiation rows remain unchanged; all six ruling rows are generated.

Conservative interpretation of the worker scope: run focused selectors in every changed package and affected IR/flow/freshness/JavaScript readers, rather than whole package suites. No new compiler-consuming package or runtime feature set was added. The dependency registry needs no addition. TestEveryMapAProgramHoldsIsSorted and TestRuntimeSetsAreEachUsed pass. The pre-existing SortedMap types in the IR are retained.

All new top-level oracle fixtures lower. FlowCorpusRemainder checked all 22 paths absent from the generated corpus, including this slice's seven new top-level oracle fixtures, with all four flow analyses. The refusal witnesses were already nested under refusal_rulings/, outside flow's top-level glob. No new top-level refused fixture remains to move. The nested computed optional-index witness now compiles and passes its dedicated oracle test.

## Counts

Exactly one -update-counts run on Linux with house clang 20.1.8: 22 additions, 11 existing numeric changes, zero removals. The command took 63.623 seconds wall time (62.094 seconds in the package); this is aggregate time, not a test leaf over 60 seconds. Its slowest test event was 5.07 seconds. The predicate-direction table is unchanged. Every changed numeric row matches the source commit's recorded value (counts-source-comparison.json).

An isolated Go overlay restores only internal/lower/narrowed.go, internal/lower/exceptions.go, internal/native/reuse.go and internal/flow/build.go to the fixed base. Direct measurements of the 11 changed existing rows all reproduce the base values exactly: TestSlice3CountsAttribution passes, package 2.912 seconds. This did not regenerate or write counts.md. The optional-read/throw change therefore owns those 11 deltas. Ten rows now construct a catchable TypeError rather than stopping at the old invariant panic; 09_tree has conservative ownership/reuse changes from the new exceptional edges. New rows belong to the registered witnesses.

Every moved row follows. Values are allocations/frees/retains/releases/peak/regions.

| Fixture | Base | Rebased |
|---|---|---|
| internal/load/testdata/0.1/compile/09_tree.ts | 19/19/28/69/16/0 | 33/33/66/69/16/0 |
| internal/oracle/testdata/047cb0d_n_arrayindex.a | 1/1/1/2/1/0 | 2/1/4/4/1/0 |
| internal/oracle/testdata/047cb0d_n_element.a | 4/3/6/8/4/0 | 5/3/9/10/4/0 |
| internal/oracle/testdata/047cb0d_n_element_method.a | 4/3/6/9/4/0 | 5/3/9/11/4/0 |
| internal/oracle/testdata/047cb0d_n_element_plain.a | 4/3/5/8/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/call_spread_array.a | new | 56/56/112/115/14/0 |
| internal/oracle/testdata/element_access_presence.a | 4/3/5/7/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/json_shorthand.a | new | 18/18/8/27/4/0 |
| internal/oracle/testdata/narrowed_fields.a | 4/3/4/6/4/0 | 5/3/7/8/4/0 |
| internal/oracle/testdata/narrowed_methods.a | 3/3/3/6/3/0 | 4/3/6/8/3/0 |
| internal/oracle/testdata/narrowed_reads.a | 5/5/3/7/4/0 | 6/5/6/9/4/0 |
| internal/oracle/testdata/optional_after_call_193.a | new | 3/3/6/9/3/0 |
| internal/oracle/testdata/optional_after_call_79.a | new | 4/4/6/10/4/0 |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | new | 3/3/1/6/2/0 |
| internal/oracle/testdata/optional_after_call_propagation.a | new | 18/18/13/20/7/0 |
| internal/oracle/testdata/optional_after_call_required.a | new | 3/3/5/7/2/0 |
| internal/oracle/testdata/optional_after_call_variants.a | new | 11/11/10/21/3/0 |
| internal/oracle/testdata/refusal_rulings/function_length.a (TypeScript mode) | new | 20/20/4/29/13/0 |
| internal/oracle/testdata/refusal_rulings/function_signature.a | new | 3/3/0/3/3/0 |
| internal/oracle/testdata/refusal_rulings/function_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 |
| internal/oracle/testdata/refusal_rulings/merge_missing.a (TypeScript mode) | new | 1/0/1/1/1/0 |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a (TypeScript mode) | new | 1/0/1/1/1/0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a | new | 3/3/1/4/3/0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a (TypeScript mode) | new | 3/3/1/4/3/0 |
| internal/oracle/testdata/refusal_rulings/merge_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 |
| internal/oracle/testdata/regexp_null_narrowed.a | 4/4/8/5/3/0 | 5/4/11/7/3/0 |
| internal/oracle/testdata/reuse_narrowed.a | 7/5/6/8/5/0 | 8/5/9/10/5/0 |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | new | 18/18/18/32/6/0 |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | new | 6/6/5/13/3/0 |
| stage3/fixtures/assignment-proofs/03_return_defined.a | new | 8/8/6/18/4/0 |
| stage3/fixtures/assignment-proofs/04_chained.a | new | 4/4/6/15/3/0 |
| stage3/fixtures/assignment-proofs/05_compound.a | new | 8/8/0/8/4/0 |
| stage3/fixtures/assignment-proofs/06_wider_target.a | new | 4/4/0/4/4/0 |

## Runtime lines

The runtime delta against the fixed base adds precisely the same five lines as the source commit's own runtime delta, with no removals. There are no additional or altered runtime patch lines for clearance. Main's intervening runtime work stays intact. runtime-comparison.json asserts equality of every added/removed line.

```c
// internal/native/runtime/adamic.h
	size_t source_length;
// internal/native/runtime/closure.c
	closure->source_length = 0;
// internal/native/runtime/union.c
	if (value != NULL && value->kind == adamic_kind_closure && strcmp(name, "length") == 0) {
		return adamic_box_number((double)((adamic_closure *)value)->source_length);
	}
```

## Validation

All Go test invocations use -count=1 and a hard -timeout; ordinary focused runs use -timeout=90s and -parallel=4. The runners set GOMAXPROCS=4 and four-CPU affinity. Oracle runs set ADAMIC_GATE_UNCACHED=1. Tests run source Node, JavaScript backend, release native, ASan/UBSan and Linux leak checks. The ruled missing-merged-field TypeScript cases pin their specified exit-70 checks independently of Node's undefined read. TSGo cache tests link inert ABI stubs: they test real cold/warm linking, Node output, placeless binary bytes and archive reuse, without claiming real checker ABI execution.

Exact commands, exits and wall seconds are in checks.jsonl; raw outputs are the matching .log files. No test output was piped. Every recorded normal test event is below 60 seconds; no test was deleted, split or rerun for exceeding its budget. Every test event, including subtests and all counts rows, is listed in test-seconds.json. gofmt on every changed Go file and git diff --check outside inherited review evidence both print nothing. Historical patches/logs carried by the source are retained byte for byte and have their original patch whitespace.

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
| parallel | 0 | 2.38 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| maps | 0 | 12.103 | `go test ./internal/ir -run ^TestEveryMapAProgramHoldsIsSorted$ -count=1 -json -timeout=90s -parallel=4` |
| agreement-registry | 0 | 7.494 | `go test ./internal/lower -run ^TestNativeAgreementRowsNameRegisteredPrograms$ -count=1 -json -timeout=90s -parallel=4` |
| native | 0 | 16.67 | `go test ./internal/native -run ^Test(TSGoCache(Release\|Sanitized)\|TSGoBuildSeesTheProgramsFeatures\|CheckerBuildsNameNoPlaceOfTheArchive\|NativeCachesWriteNothingUnderHome)$ -count=1 -json -timeout=90s -parallel=4` |
| optional | 0 | 18.848 | `go test ./internal/oracle -run ^TestOptionalAfterCall -count=1 -json -timeout=90s -parallel=4` |
| spread | 0 | 12.236 | `go test ./internal/oracle -run ^Test(CallSpread\|JSONStringifyShorthand) -count=1 -json -timeout=90s -parallel=4` |
| function | 0 | 5.311 | `go test ./internal/oracle -run ^TestFunction(Adamic\|TypeScript\|CallSignature) -count=1 -json -timeout=90s -parallel=4` |
| merged | 0 | 3.794 | `go test ./internal/oracle -run ^TestMergedField -count=1 -json -timeout=90s -parallel=4` |
| assignment | 0 | 6.762 | `go test ./internal/oracle -run ^TestAssignmentProof -count=1 -json -timeout=90s -parallel=4` |
| lower | 0 | 1.123 | `go test ./internal/lower -run ^TestOptionalAfterCallPreservesEarlierThrow$ -count=1 -json -timeout=90s -parallel=4` |
| parallel-preflight | 0 | 1.477 | `go test ./internal/lower -run ^TestParallel(ReadonlyCheckerRepresentation\|GlobalMarkingRoots)$ -count=1 -json -timeout=90s -parallel=4` |
| flow | 0 | 17.541 | `go test ./internal/flow -run ^Test(OptionalAfterCallFlowEdges\|FlowCorpusRemainder\|FlowCorpusUnitsCoverEveryProgram)$ -count=1 -json -timeout=90s -parallel=4` |
| fresh-corpus | 0 | 5.48 | `go test ./internal/fresh -run ^Test(FreshCorpusRemainder\|FreshCorpusUnitsCoverEveryProgram)$ -count=1 -json -timeout=90s -parallel=4` |
| javascript | 0 | 0.923 | `go test ./internal/javascript -run ^TestViewMixedUnion -count=1 -json -timeout=90s -parallel=4` |
| cssstrings | 0 | 10.173 | `go test ./stage1/cohere/cssstrings -run ^TestMultiPushGap$ -count=1 -json -timeout=90s -parallel=4` |
| json | 0 | 5.921 | `go test ./stage1/cohere/json -run ^TestClosedMultiplePushGap$ -count=1 -json -timeout=90s -parallel=4` |
| markdownblocks | 0 | 13.932 | `go test ./stage1/cohere/markdownblocks -run ^TestParserRepresentationProbes$ -count=1 -json -timeout=90s -parallel=4` |
| selector | 0 | 5.773 | `go test ./stage1/cohere/selector -run ^TestEachGapStandsWhereGapsMdSaysItDoes$ -count=1 -json -timeout=90s -parallel=4` |
| yaml | 0 | 7.073 | `go test ./stage1/cohere/yaml -run ^Test(LexerGaps\|Product_YAMLClosedGaps(Lowered\|Native))$ -count=1 -json -timeout=90s -parallel=4` |
| parser | 0 | 5.573 | `go test ./stage1/typescript/parser -run ^TestPushSpreadGap$ -count=1 -json -timeout=90s -parallel=4` |
| scanner | 0 | 5.767 | `go test ./stage1/typescript/scanner -run ^TestGapStandsWhereGapsMdSays$ -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 17.841 | `go test ./stage3/fixtures -run ^Test(FixturesAssignmentProofs\|FixtureDirectoriesHaveTopLevelTests)$ -count=1 -json -timeout=90s -parallel=4` |
| readers | 0 | 7.782 | `go test ./internal/ir -run ^TestCallTargetReaders -count=1 -json -timeout=90s -parallel=4` |
| jobserver | 0 | 1.471 | `go test ./internal/native -run ^Test(CompilesHoldAJobserverTokenEach\|AFailedCompileGivesItsTokenBack\|AJobserverThatCantBeOpenedFailsTheCompile\|TheJobserverMovesNoKey)$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 0 | 2.187 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| vet | 0 | 4.82 | `go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/javascript ./internal/oracle ./stage3/fixtures ./stage1/cohere/cssstrings ./stage1/cohere/json ./stage1/cohere/markdownblocks ./stage1/cohere/selector ./stage1/cohere/yaml ./stage1/typescript/parser ./stage1/typescript/scanner` |
| json-existing | 0 | 1.574 | `go test ./stage1/cohere/json -run ^TestDocumentedStageZeroGaps$ -count=1 -json -timeout=90s -parallel=4` |
| yaml-consumers | 0 | 2.641 | `go test ./stage1/cohere/yaml -run ^Test(ClosedLexerGaps\|ClosedStringPresenceGap\|ClosedValuePresenceGap)$ -count=1 -json -timeout=90s -parallel=4` |
| counts | 0 | 63.623 | `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts` |
| counts-table | 0 | 1.22 | `go test ./internal/oracle -run ^TestCountsAreRecordedTable$ -count=1 -json -timeout=90s -parallel=4` |

Counterfactual command: `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 timeout 180 go test ./internal/oracle -overlay=review/compiler/loom-r-slice-3/counts-counterfactual/overlay.json -run ^TestSlice3CountsAttribution$ -count=1 -json -timeout=90s -parallel=4`.

Each selected top-level test's seconds follow; subtest seconds are in test-seconds.json.

| Check | Test | Seconds |
|---|---|---:|
| agreement-registry | TestNativeAgreementRowsNameRegisteredPrograms | 0.01 |
| assignment | TestAssignmentProofChained | 3.3 |
| assignment | TestAssignmentProofScannerKeyword | 3.34 |
| assignment | TestAssignmentProofReturnDefinedMutant | 3.34 |
| assignment | TestAssignmentProofIfNarrowingMutant | 3.35 |
| assignment | TestAssignmentProofReturnDefined | 1.06 |
| assignment | TestAssignmentProofScannerKeywordMutant | 1.1 |
| assignment | TestAssignmentProofIfNarrowing | 1.1 |
| assignment | TestAssignmentProofCompoundMutant | 1.1 |
| assignment | TestAssignmentProofWiderTargetMutant | 1.1 |
| assignment | TestAssignmentProofCompound | 1.06 |
| assignment | TestAssignmentProofWiderTarget | 1.09 |
| assignment | TestAssignmentProofChainedMutant | 1.14 |
| counts-table | TestCountsAreRecordedTable | 0 |
| cssstrings | TestMultiPushGap | 5.45 |
| fixtures | TestFixtureDirectoriesHaveTopLevelTests | 0 |
| fixtures | TestFixturesAssignmentProofs | 10.18 |
| flow | TestFlowCorpusUnitsCoverEveryProgram | 0.03 |
| flow | TestOptionalAfterCallFlowEdges | 0.31 |
| flow | TestFlowCorpusRemainder | 2.96 |
| fresh-corpus | TestFreshCorpusUnitsCoverEveryProgram | 0 |
| fresh-corpus | TestFreshCorpusRemainder | 0.93 |
| function | TestFunctionAdamicCallAnnotationRefused | 2.52 |
| function | TestFunctionTypeScriptCallRefused | 2.52 |
| function | TestFunctionAdamicUnusedAnnotationRefused | 2.53 |
| function | TestFunctionCallSignatureAgreesWithNode | 2.81 |
| function | TestFunctionAdamicLengthAnnotationRefused | 0.29 |
| function | TestFunctionTypeScriptUnusedAccepted | 0.47 |
| function | TestFunctionTypeScriptLengthAgreesWithNode | 1.66 |
| javascript | TestViewMixedUnionSelectsNumber | 0.05 |
| javascript | TestViewMixedUnionUnknownAndUnavailable | 0.1 |
| jobserver | TestCompilesHoldAJobserverTokenEach | 0.08 |
| jobserver | TestAFailedCompileGivesItsTokenBack | 0 |
| jobserver | TestAJobserverThatCantBeOpenedFailsTheCompile | 0 |
| jobserver | TestTheJobserverMovesNoKey | 0 |
| json-existing | TestDocumentedStageZeroGaps | 0 |
| json | TestClosedMultiplePushGap | 0.78 |
| lower | TestOptionalAfterCallPreservesEarlierThrow | 0 |
| maps | TestEveryMapAProgramHoldsIsSorted | 0 |
| markdownblocks | TestParserRepresentationProbes | 8.4 |
| merged | TestMergedFieldAdamicComputedRefused | 1.98 |
| merged | TestMergedFieldAdamicUnusedRefused | 1.98 |
| merged | TestMergedFieldTypeScriptUnusedAccepted | 2.23 |
| merged | TestMergedFieldTypeScriptComputedChecked | 2.44 |
| merged | TestMergedFieldTypeScriptMissingChecked | 0.47 |
| merged | TestMergedFieldAdamicReadRefused | 0.08 |
| merged | TestMergedFieldTypeScriptInitializedAgreesWithNode | 0.66 |
| merged | TestMergedFieldAdamicInitializedAgreesWithNode | 0.41 |
| native | TestTSGoBuildSeesTheProgramsFeatures | 0.53 |
| native | TestNativeCachesWriteNothingUnderHome | 4.46 |
| native | TestCheckerBuildsNameNoPlaceOfTheArchive | 4.01 |
| native | TestTSGoCacheRelease | 4.59 |
| native | TestTSGoCacheSanitized | 6.71 |
| optional | TestOptionalAfterCallVariants | 12.53 |
| optional | TestOptionalAfterCallComputed | 12.55 |
| optional | TestOptionalAfterCall79 | 12.55 |
| optional | TestOptionalAfterCallRequired | 12.57 |
| optional | TestOptionalAfterCall193 | 0.32 |
| optional | TestOptionalAfterCallPropagation | 0.36 |
| parallel-preflight | TestParallelReadonlyCheckerRepresentation | 0.31 |
| parallel-preflight | TestParallelGlobalMarkingRoots | 0.31 |
| parallel | TestEveryTestIsParallelOrSaysWhy | 0.25 |
| parser | TestPushSpreadGap | 0.82 |
| readers | TestCallTargetReaders | 5.91 |
| runtime-sets | TestRuntimeSetsAreEachUsed | 0.66 |
| scanner | TestGapStandsWhereGapsMdSays | 0.82 |
| selector | TestEachGapStandsWhereGapsMdSaysItDoes | 0 |
| spread | TestJSONStringifyShorthandWrongBindingMutant | 2.63 |
| spread | TestJSONStringifyShorthand | 2.71 |
| spread | TestCallSpreadEvaluatedTwiceMutant | 7.82 |
| spread | TestCallSpreadArray | 11.1 |
| yaml-consumers | TestClosedValuePresenceGap | 0.51 |
| yaml-consumers | TestClosedStringPresenceGap | 0.54 |
| yaml-consumers | TestClosedLexerGaps | 1.09 |
| yaml | TestProduct_YAMLClosedGapsLowered | 0.43 |
| yaml | TestProduct_YAMLClosedGapsNative | 2.18 |
| yaml | TestLexerGaps | 2.26 |
| counts | TestCountsAreRecordedTable | 1.15 |
| counts | TestCountsAreRecorded_031 | 0.01 |
| counts | TestCountsAreRecorded_016 | 0.01 |
| counts | TestCountsAreRecorded_008 | 0.01 |
| counts | TestCountsAreRecorded_030 | 0.01 |
| counts | TestCountsAreRecorded_029 | 0.0 |
| counts | TestCountsAreRecorded_027 | 0.01 |
| counts | TestCountsAreRecorded_026 | 0.0 |
| counts | TestCountsAreRecorded_025 | 0.02 |
| counts | TestCountsAreRecorded_028 | 0.02 |
| counts | TestCountsAreRecorded_024 | 0.01 |
| counts | TestCountsAreRecorded_023 | 0.02 |
| counts | TestCountsAreRecorded_022 | 0.02 |
| counts | TestCountsAreRecorded_021 | 0.02 |
| counts | TestCountsAreRecorded_020 | 0.0 |
| counts | TestCountsAreRecorded_018 | 0.0 |
| counts | TestCountsAreRecorded_019 | 0.0 |
| counts | TestCountsAreRecorded_015 | 0.01 |
| counts | TestCountsAreRecorded_017 | 0.01 |
| counts | TestCountsAreRecorded_014 | 0.0 |
| counts | TestCountsAreRecorded_012 | 0.01 |
| counts | TestCountsAreRecorded_010 | 0.0 |
| counts | TestCountsAreRecorded_013 | 0.0 |
| counts | TestCountsAreRecorded_011 | 0.0 |
| counts | TestCountsAreRecorded_004 | 0.0 |
| counts | TestCountsAreRecorded_006 | 0.01 |
| counts | TestCountsAreRecorded_009 | 0.0 |
| counts | TestCountsAreRecorded_007 | 0.01 |
| counts | TestCountsAreRecorded_005 | 0.01 |
| counts | TestCountsAreRecorded_001 | 0.0 |
| counts | TestCountsAreRecorded_003 | 0.0 |
| counts | TestCountsAreRecorded_002 | 0.0 |
| counts | TestCountsAreRecorded_000 | 0.0 |
| counts-attribution | TestSlice3CountsAttribution | 0 |

## Mutants and catchers

Each production mutant ran independently through a Go overlay of the final source, with -timeout=80s and an outer timeout 90. All exited 1 at their intended test. Build failures, clang failures and timeouts were excluded from credit. Full mutated sources are .go.txt files under mutants/; commands and wall seconds are in mutants-results.json.

| Mutant | Catching test | Test seconds | Failure |
|---|---|---:|---|
| check-optional-receiver | TestOptionalAfterCall79 | 2.16 | behavior assertion |
| javascript-required-panic | TestOptionalAfterCallRequired | 2.09 | behavior assertion |
| native-required-panic | TestOptionalAfterCallRequired | 2.18 | behavior assertion |
| drop-flow-throw-edge | TestOptionalAfterCallFlowEdges | 0.04 | behavior assertion |
| drop-required-propagation | TestOptionalAfterCallPropagation | 2.05 | behavior assertion and LeakSanitizer (76 bytes in one allocation) |
| merged-computed-read | TestMergedFieldTypeScriptComputedChecked | 2.13 | behavior assertion |
| merged-declaration | TestMergedFieldAdamicUnusedRefused | 0.03 | behavior assertion |
| merged-read | TestMergedFieldTypeScriptMissingChecked | 2.07 | behavior assertion |
| function-annotation | TestFunctionAdamicUnusedAnnotationRefused | 0.04 | behavior assertion |
| function-call | TestFunctionTypeScriptCallRefused | 0.05 | behavior assertion |
| function-length | TestFunctionTypeScriptLengthAgreesWithNode | 3.24 | behavior assertion |
| erase-earlier-throw | TestOptionalAfterCallPreservesEarlierThrow | 0.00 | behavior assertion |
| tsgo-absolute-main | TestTSGoCacheSanitized | 1.04 | behavior assertion |
| tsgo-rewrite-runtime | TestTSGoCacheRelease | 0.90 | behavior assertion |

The eight embedded mutants also pass their catch requirements: TestAssignmentProofScannerKeywordMutant, TestAssignmentProofIfNarrowingMutant, TestAssignmentProofReturnDefinedMutant, TestAssignmentProofChainedMutant, TestAssignmentProofCompoundMutant, TestAssignmentProofWiderTargetMutant, TestCallSpreadEvaluatedTwiceMutant and TestJSONStringifyShorthandWrongBindingMutant. Assignment witnesses and mutants pass stock TypeScript 6.0.3 with zero diagnostics. The six source mutants change Node stdout from its independent golden; spread double evaluation and wrong shorthand binding disagree with source Node in native and JavaScript. Each mutant's normal test seconds is in the preceding table.

## Setup and limits

The first setup returned 1 because the checkout left by the earlier interrupted task had an empty cohere worktree: cohere/TypeScript/tsc/go.mod was absent, despite setup reporting submodules ready. Restored the recorded cohere commit and initialized its pinned nested TypeScript submodule; no code was copied from it. A second setup stalled in the shared Go cache helper and was terminated (143). Used the supported ADAMIC_GOCACHE_OFF=1 fallback, and sourced /workspace/adamic-tools/env.sh for builds/tests. The formatting command initially lacked the sourced tool PATH; it was then run successfully with that environment. All source tests passed without code changes beyond the conflict resolution.

Successful setup timing lines:

```text
setup: node ready (0.015s)
setup: go ready (0.014s)
setup: submodules ready (0.044s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.005s
setup: markdown dependencies ready (0.046s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.099s)
setup: shared cache off (ADAMIC_GOCACHE_OFF=1) (0.100s)
setup: go build ready (131.115s)
setup: test binaries deferred (use --warm-tests) (131.190s)
setup: build cache warm (131.192s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=off load-before=0.20 0.10 0.04 5/262 3132 load-after=8.13 3.09 1.15 1/258 6771
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (131.221s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.hgde2z
```

nproc=5; cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0. The pinned TypeScript dependency install (timeout 120 npm ci --prefix stage3/api --ignore-scripts) exited 0.

No ahra command, PR, Loom submission, main push or force push was performed. This unit does not claim a full gate, full package suites, execution of the older generated flow corpus, macOS/WASI validation, or real TSGo checker ABI behavior. The earlier unrelated review/compiler/loom-slice-3/BLOCKED.md remains untracked and is excluded from this commit. The delivery branch was absent remotely before work (remote-before.txt).
