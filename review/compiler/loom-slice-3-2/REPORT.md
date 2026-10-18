Re-cut parked Loom change chg_09tfwyx05s0q5stp30p6rk41k6 on current main as compiler/loom-slice-3-2.
Source ee4686b8bfa14799a468376535fcf9f9e042c5f2; base main 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4; delivery SHA is reported with the push.
56 focused top-level tests, Node-held release and sanitizer fixtures, TestCallTargetReaders, counts regeneration and vet passed.
Fourteen production mutants and eight embedded mutants are recorded below; behavioral catches exclude build errors and timeouts.
Not covered: full packages, full gate, macOS execution, real checker ABI execution, PR creation or Loom submission.

The squash was exactly origin/compiler/loom-slice-3 onto origin/main. The actual merge conflicted only in internal/oracle/counts.md. Took the complete current-main table with git checkout --ours, then regenerated once. The eight rows landed ahead of slice 3 remain unchanged. The resulting diff adds 22 slice-3 rows and changes 11 existing rows, with no removals. Every before and after value matches the parked slice's attribution. counts-attribution.json records all rows. Reverting only the optional-read member's four production files to current main in a Go overlay reproduces all 11 existing main rows (counts-counterfactual.jsonl, 5.162 s). No second counts regeneration or hand edit was used.

The reported tsgo conflict did not recur on fetched main. git diff c0fd8945 HEAD -- internal/native/tsgo.go and the corresponding path log are empty: current main's file is byte-identical to the parked base. The parked source's combined resolution therefore applies cleanly. tsgo-main-since-base.patch and tsgo-new-landings.txt record this observation; no newly landed tsgo change is invented.

The landed tsgo infrastructure preserved by the source is:

- 6f970c1e3f902bc472650c0e1624524576b401db: resolve clang/ar through toolchain and run builds through CommandIn.
- 5192cecf70f59a4e67479a4113e6af31da0ca452: native caches under the build cache, and checker archives linked by relative name.
- cbf6502eb2ccac97446a6234cb473751e2bbbc47: compile checker/runtime units with relative input names so binaries name no scratch directory.

The resulting BuildTSGo keeps slice 3's runtime caching through main's splitTSGoRuntime machinery, program feature flags, relative main/runtime/checker inputs, prefix maps, absolute output normalization, and stable Darwin object metadata. Cache tests hold first/warm release and sanitized execution to source Node, identical binary bytes and an unchanged runtime archive. They link inert ABI stubs and do not claim checker execution coverage. Conservative assumption: main's shared checker runtime cache remains authoritative for whole and split builds.

Runtime changes are exactly five added lines, zero removed lines:

- internal/native/runtime/adamic.h: +1 (source_length field).
- internal/native/runtime/closure.c: +1 (source_length initialized to zero).
- internal/native/runtime/union.c: +3 (dynamic closure length observation).

runtime-proof.json verifies byte-for-byte that removing only those five known lines yields current main for all three files. runtime-diff.patch contains the full diff. No other runtime path changes.

Toolchain setup first exited 1: it sampled the squash while new files were arriving and reported missing Throws, SourceLength and ruling helpers. setup.txt preserves the failure. Retried on the fully squashed tree; setup-stable.txt records success. Exported GOPROXY=https://proxy.golang.org|direct and sourced /workspace/adamic-tools/env.sh for all validation. Timings: Node 0.035 s; Go 0.038 s; submodules 0.093 s; markdown 0.097 s; clang 0.194 s; shared cache 0.593 s; Go build 57.923 s; done 58.134 s. nproc=5 and cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands and complete outputs are versioned in this directory:

- run-focused.py and focused-results.json: nine selected invocations, go test -count=1 -timeout 90s -parallel 4 under timeout 180, GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1. No entire package selection. Each output is its *-green.jsonl.
- run-gaps-green.py and gaps-green-results.json: seven changed stage-1 gap packages and the assignment-proofs owner, with the same limits and uncached Node/sanitizer checks.
- timeout 540 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 480s -parallel 4 -json -args -update-counts: 201.374 s, exit 0, counts-update.jsonl. Main's existing bounded count family measures every row.
- TestCallTargetReaders: -run ^TestCallTargetReaders$ -count=1 -timeout 90s -parallel 4, 24.57 s, exit 0, readers-green.jsonl.
- timeout 180 go vet over all 14 changed Go packages: exit 0 with no diagnostics, vet.txt. Formatting emitted no diagnostics, formatting.txt.
- run-mutants.py: each production overlay under timeout 90 with go test -timeout 80s -count=1 -v, uncached; exact commands and invocation seconds in mutants-results.json, complete individual output in mutants/<name>/test.txt. Mutant Go sources use .go.txt.

Every selected top-level test's seconds are below. test-seconds.json includes every selected subtest; counts-update.jsonl includes every count-family row. All added/touched test leaves observed under 60 s; no tests deleted or split. No new compiler-consuming package was added, so no dependency-manifest entry was necessary. No code was copied from cohere.

| Package | Test | Seconds |
|---|---|---:|
| internal/oracle | TestAssignmentProofScannerKeyword | 15.52 |
| internal/oracle | TestAssignmentProofCompoundMutant | 15.58 |
| internal/oracle | TestAssignmentProofWiderTargetMutant | 15.62 |
| internal/oracle | TestAssignmentProofChained | 15.94 |
| internal/oracle | TestAssignmentProofWiderTarget | 5.25 |
| internal/oracle | TestAssignmentProofCompound | 5.35 |
| internal/oracle | TestAssignmentProofIfNarrowingMutant | 5.38 |
| internal/oracle | TestAssignmentProofReturnDefinedMutant | 5.18 |
| internal/oracle | TestAssignmentProofIfNarrowing | 4.98 |
| internal/oracle | TestAssignmentProofReturnDefined | 5.20 |
| internal/oracle | TestAssignmentProofChainedMutant | 5.36 |
| internal/oracle | TestAssignmentProofScannerKeywordMutant | 5.27 |
| internal/flow | TestOptionalAfterCallFlowEdges | 0.05 |
| internal/oracle | TestFunctionTypeScriptCallRefused | 8.85 |
| internal/oracle | TestFunctionAdamicUnusedAnnotationRefused | 8.86 |
| internal/oracle | TestFunctionAdamicCallAnnotationRefused | 8.92 |
| internal/oracle | TestFunctionAdamicLengthAnnotationRefused | 0.54 |
| internal/oracle | TestFunctionTypeScriptUnusedAccepted | 9.81 |
| internal/oracle | TestFunctionCallSignatureAgreesWithNode | 1.37 |
| internal/oracle | TestFunctionTypeScriptLengthAgreesWithNode | 9.16 |
| stage1/cohere/cssstrings | TestMultiPushGap | 23.89 |
| stage3/fixtures | TestFixtureDirectoriesHaveTopLevelTests | 0.02 |
| stage3/fixtures | TestFixturesAssignmentProofs | 17.64 |
| stage1/cohere/json | TestClosedMultiplePushGap | 3.21 |
| stage1/cohere/markdownblocks | TestParserRepresentationProbes | 45.73 |
| stage1/cohere/selector | TestEachGapStandsWhereGapsMdSaysItDoes | 0.00 |
| stage1/typescript/parser | TestPushSpreadGap | 1.39 |
| stage1/typescript/scanner | TestGapStandsWhereGapsMdSays | 1.36 |
| stage1/cohere/yaml | TestProduct_YAMLClosedGapsLowered | 0.99 |
| stage1/cohere/yaml | TestProduct_YAMLClosedGapsNative | 8.15 |
| stage1/cohere/yaml | TestLexerGaps | 8.36 |
| internal/lower | TestOptionalAfterCallPreservesEarlierThrow | 0.00 |
| internal/oracle | TestMergedFieldAdamicUnusedRefused | 9.38 |
| internal/oracle | TestMergedFieldTypeScriptUnusedAccepted | 10.93 |
| internal/oracle | TestMergedFieldTypeScriptMissingChecked | 11.61 |
| internal/oracle | TestMergedFieldAdamicComputedRefused | 0.72 |
| internal/oracle | TestMergedFieldTypeScriptComputedChecked | 11.70 |
| internal/oracle | TestMergedFieldTypeScriptInitializedAgreesWithNode | 2.45 |
| internal/oracle | TestMergedFieldAdamicReadRefused | 0.41 |
| internal/oracle | TestMergedFieldAdamicInitializedAgreesWithNode | 1.57 |
| internal/native | TestTSGoBuildSeesTheProgramsFeatures | 0.45 |
| internal/native | TestCheckerBuildsNameNoPlaceOfTheArchive | 9.39 |
| internal/native | TestTSGoCacheRelease | 10.82 |
| internal/native | TestNativeCachesWriteNothingUnderHome | 10.88 |
| internal/native | TestTSGoCacheSanitized | 13.42 |
| internal/oracle | TestOptionalAfterCall79 | 29.16 |
| internal/oracle | TestOptionalAfterCallRequired | 29.41 |
| internal/oracle | TestOptionalAfterCall193 | 29.46 |
| internal/oracle | TestOptionalAfterCallVariants | 29.92 |
| internal/oracle | TestOptionalAfterCallPropagation | 1.33 |
| internal/oracle | TestOptionalAfterCallComputed | 1.17 |
| internal/ir | TestCallTargetReaders | 24.57 |
| internal/oracle | TestJSONStringifyShorthandWrongBindingMutant | 4.99 |
| internal/oracle | TestJSONStringifyShorthand | 5.10 |
| internal/oracle | TestCallSpreadEvaluatedTwiceMutant | 26.85 |
| internal/oracle | TestCallSpreadArray | 41.95 |

Every moved counts row is below. A/F/R/L/P/G mean allocations/frees/retains/releases/peak/in regions.

| Row | Before | After | Member |
|---|---|---|---|
| internal/load/testdata/0.1/compile/09_tree.ts | 19/19/28/69/16/0 | 33/33/66/69/16/0 | optional-chain-after-call-main |
| internal/oracle/testdata/narrowed_reads.a | 5/5/3/7/4/0 | 6/5/6/9/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/narrowed_methods.a | 3/3/3/6/3/0 | 4/3/6/8/3/0 | optional-chain-after-call-main |
| internal/oracle/testdata/narrowed_fields.a | 4/3/4/6/4/0 | 5/3/7/8/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/reuse_narrowed.a | 7/5/6/8/5/0 | 8/5/9/10/5/0 | optional-chain-after-call-main |
| internal/oracle/testdata/regexp_null_narrowed.a | 4/4/8/5/3/0 | 5/4/11/7/3/0 | optional-chain-after-call-main |
| internal/oracle/testdata/call_spread_array.a | new | 56/56/112/115/14/0 | spread-shorthand |
| internal/oracle/testdata/element_access_presence.a | 4/3/5/7/4/0 | 5/3/8/10/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/refusal_rulings/function_signature.a | new | 3/3/0/3/3/0 | refusal-rulings-main |
| internal/oracle/testdata/json_shorthand.a | new | 18/18/8/27/4/0 | spread-shorthand |
| internal/oracle/testdata/refusal_rulings/merge_present.a | new | 3/3/1/4/3/0 | refusal-rulings-main |
| internal/oracle/testdata/047cb0d_n_element_plain.a | 4/3/5/8/4/0 | 5/3/8/10/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/047cb0d_n_element.a | 4/3/6/8/4/0 | 5/3/9/10/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/047cb0d_n_element_method.a | 4/3/6/9/4/0 | 5/3/9/11/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/047cb0d_n_arrayindex.a | 1/1/1/2/1/0 | 2/1/4/4/1/0 | optional-chain-after-call-main |
| internal/oracle/testdata/optional_after_call_79.a | new | 4/4/6/10/4/0 | optional-chain-after-call-main |
| internal/oracle/testdata/optional_after_call_193.a | new | 3/3/6/9/3/0 | optional-chain-after-call-main |
| internal/oracle/testdata/optional_after_call_variants.a | new | 11/11/10/21/3/0 | optional-chain-after-call-main |
| internal/oracle/testdata/optional_after_call_required.a | new | 3/3/5/7/2/0 | optional-chain-after-call-main |
| internal/oracle/testdata/optional_after_call_propagation.a | new | 18/18/13/20/7/0 | optional-chain-after-call-main |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | new | 3/3/1/6/2/0 | optional-chain-after-call-main |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | new | 18/18/18/32/6/0 | assignment-proofs-main |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | new | 6/6/5/13/3/0 | assignment-proofs-main |
| stage3/fixtures/assignment-proofs/03_return_defined.a | new | 8/8/6/18/4/0 | assignment-proofs-main |
| stage3/fixtures/assignment-proofs/04_chained.a | new | 4/4/6/15/3/0 | assignment-proofs-main |
| stage3/fixtures/assignment-proofs/05_compound.a | new | 8/8/0/8/4/0 | assignment-proofs-main |
| stage3/fixtures/assignment-proofs/06_wider_target.a | new | 4/4/0/4/4/0 | assignment-proofs-main |
| internal/oracle/testdata/refusal_rulings/function_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 | refusal-rulings-main |
| internal/oracle/testdata/refusal_rulings/function_length.a (TypeScript mode) | new | 20/20/4/29/13/0 | refusal-rulings-main |
| internal/oracle/testdata/refusal_rulings/merge_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 | refusal-rulings-main |
| internal/oracle/testdata/refusal_rulings/merge_missing.a (TypeScript mode) | new | 1/0/1/1/1/0 | refusal-rulings-main |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a (TypeScript mode) | new | 1/0/1/1/1/0 | refusal-rulings-main |
| internal/oracle/testdata/refusal_rulings/merge_present.a (TypeScript mode) | new | 3/3/1/4/3/0 | refusal-rulings-main |

Eight embedded mutants are the six TestAssignmentProof*Mutant tests, TestCallSpreadEvaluatedTwiceMutant and TestJSONStringifyShorthandWrongBindingMutant. Each passes by rejecting changed valid source or IR against independently captured Node stdout after successful stock TypeScript and/or backend compilation, sanitized execution and leak checks. The existing TestTSGoBuildSeesTheProgramsFeatures also executes its feature-flag omission negative control; that is an include-order compile-contract check, not a claimed runtime kill.

Production mutant outcomes:

| Mutant | Catching test | Invocation seconds |
|---|---|---:|
| check-optional-receiver | TestOptionalAfterCall79 | 54.428 |
| javascript-required-panic | TestOptionalAfterCallRequired | 43.278 |
| native-required-panic | TestOptionalAfterCallRequired | 50.572 |
| drop-flow-throw-edge | TestOptionalAfterCallFlowEdges | 42.806 |
| drop-required-propagation | TestOptionalAfterCallPropagation | 32.816 |
| merged-computed-read | TestMergedFieldTypeScriptComputedChecked | 24.315 |
| merged-declaration | TestMergedFieldAdamicUnusedRefused | 20.094 |
| merged-read | TestMergedFieldTypeScriptMissingChecked | 25.261 |
| function-annotation | TestFunctionAdamicUnusedAnnotationRefused | 17.624 |
| function-call | TestFunctionTypeScriptCallRefused | 15.286 |
| function-length | TestFunctionTypeScriptLengthAgreesWithNode | 20.496 |
| erase-earlier-throw | TestOptionalAfterCallPreservesEarlierThrow | 11.151 |
| tsgo-absolute-main | TestTSGoCacheSanitized | 11.473 |
| tsgo-rewrite-runtime | TestTSGoCacheRelease | 11.864 |

Every production mutant exited 1 at a test assertion, not a build failure, warning/error kill or timeout. All 14 raw failures are versioned as test.txt files. Main remained 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4 at the final fetch. Nothing was submitted to Loom and no PR was opened.
