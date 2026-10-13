Re-cut optional-chain-after-call, spread/shorthand, refusal rulings, assignment proofs and tsgo caching on current main.
Source 46924252c3a0e9119acee809fb5ae276589739bc; base main c0fd8945b872e57a341ccee5973e3f4d1b75c34b; delivery commit is the single commit pushed to compiler/loom-slice-3.
Focused Node-held release and ASan/UBSan/LSan tests, counts regeneration, TestCallTargetReaders and vet pass; raw commands and outputs are linked below.
Production overlay mutants and eight embedded mutants are recorded below; no build-warning kill is counted as a behavioral kill.
Not covered: full packages, full gate, Loom submission, macOS execution or real checker ABI execution/performance.

The branch is a squash of origin/compiler/chain-slice-3 at its pinned source, rather than a cherry-pick of its stack. Main did not move at the final fetch. No pull request was opened and nothing was submitted to Loom.

Two content conflicts were resolved once:

- internal/native/tsgo.go: use the cached checker runtime, retaining main.c and runtime/checker archive links as relative compiler/linker inputs. Keep toolchain.CommandIn, prefix maps, absolute output normalization and checkerArchivePath. Reuse main's splitTSGoRuntime snapshot/key/location machinery, rather than the source's obsolete os.UserCacheDir cache. Keep Darwin's stable main.o timestamp and debug-map prefix handling. The selected checker archive remains a final-link input, outside the runtime key.
- stage1/cohere/yaml/gaps_test.go: multiplePush is now a closed gap held against source Node and both backends. Keep t.Parallel and main's product-based closedGap helper. Add multiplePush to the existing YAML closed-gaps product recipe in port_products_test.go; preserve all other gap checks.

An automatic merge was clean but the source's additionalFixtureCounts callbacks no longer matched main's countCase API. Adapted only the slice's own three registration files to individual keyed cases in main's bounded family. No test was deleted. All 56 selected top-level tests passed under 60 seconds. Go compilation is warmed separately, as Loom builds its test binaries ahead; setup within a test is included in the recorded test seconds. No new compiler-consuming package was added, so cloud/fast-gate/compiler-dependencies.json needs no entry.

Conservative assumption: main's shared checker runtime cache is authoritative for the whole and split builds. The cache witness links inert checker ABI stubs and executes a feature-bearing Adamic program that does not call those stubs. This validates caching, linking, binary identity, include-order features and program semantics, without claiming checker-library execution coverage. Existing source witnesses and their provenance were preserved; no new code was copied from cohere.

Runtime C/header paths changed against main:

- internal/native/runtime/adamic.h
- internal/native/runtime/closure.c
- internal/native/runtime/union.c

Compiler production paths changed against main:

- internal/flow/build.go
- internal/ir/defined.go
- internal/ir/ir.go
- internal/javascript/javascript.go
- internal/lower/array_call_arguments.go
- internal/lower/cycles.go
- internal/lower/exceptions.go
- internal/lower/expression.go
- internal/lower/function_rulings.go
- internal/lower/functions.go
- internal/lower/library_array.go
- internal/lower/library_json_stringify.go
- internal/lower/merged_fields.go
- internal/lower/narrowed.go
- internal/lower/refusals.go
- internal/native/emit_expressions.go
- internal/native/reuse.go
- internal/native/tsgo.go

Toolchain setup:

Exported GOPROXY=https://proxy.golang.org|direct and ran timeout 600 bash cloud/setup.sh with output in setup.log. It exited 124 during go list -deps -export -json ./..., before printing a final environment line or completing cache warming. Tools and submodules were ready. Sourced /workspace/adamic-tools/env.sh and completed a bounded focused compile-only warmup, then the focused checks. Initial attempts blocked on the missing submodule and an unused import are preserved, as are canceled duplicate cold compilations. No failed attempt is claimed as a pass.

Timing lines: Go ready 0.113 s; Node ready 0.119 s; clang ready 0.740 s; markdown dependencies ready 1.517 s; submodules ready 223.727 s; shared cache ready 233.953 s. nproc=5; cgroup cpu.max=400000 100000 (4 CPUs). Go 1.27.1, clang 20.1.8, Node 24.19.0. toolchain.txt records the outputs.

Exact validation commands and exit codes are in focused-results.json and gaps-green-results.json; each uses timeout 180, go test -timeout 90s -count=1 -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1, with output to a JSONL file. run-focused.py and run-gaps-green.py specify each selection. No whole-package or full-gate test selection was used. TestCallTargetReaders is readers-green.jsonl (40.04 s). Vet of all 14 changed Go packages exited 0 with no diagnostics in vet-final.log; formatting and non-evidence whitespace checks emitted no diagnostics.

Counts regenerated once with timeout 540 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 480s -parallel 4 -json -args -update-counts, output in counts-update.jsonl: package passes in 164.451 s. The existing bounded count family runs individual rows. No count value was hand edited. 22 new rows and 11 existing changed rows, no removals; all 33 resulting rows match the source attribution. counts-attribution.json contains exact before/after rows. A production overlay removing only the optional-read member reproduces all 11 current-main rows, observed in counts-counterfactual.jsonl (9.881 s), without regenerating the table again.

Every moved count row is below. A/F/R/L/P/G mean allocations/frees/retains/releases/peak/in regions.

| Row | Before | After | Attribution |
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

Every selected top-level test's observed Go elapsed seconds:

| Package | Test | Seconds |
|---|---|---:|
| internal/oracle | TestAssignmentProofWiderTargetMutant | 17.69 |
| internal/oracle | TestAssignmentProofScannerKeyword | 17.76 |
| internal/oracle | TestAssignmentProofCompoundMutant | 17.78 |
| internal/oracle | TestAssignmentProofChained | 17.98 |
| internal/oracle | TestAssignmentProofWiderTarget | 4.68 |
| internal/oracle | TestAssignmentProofChainedMutant | 4.71 |
| internal/oracle | TestAssignmentProofCompound | 5.31 |
| internal/oracle | TestAssignmentProofIfNarrowingMutant | 5.07 |
| internal/oracle | TestAssignmentProofReturnDefinedMutant | 4.48 |
| internal/oracle | TestAssignmentProofReturnDefined | 4.53 |
| internal/oracle | TestAssignmentProofIfNarrowing | 4.30 |
| internal/oracle | TestAssignmentProofScannerKeywordMutant | 4.45 |
| internal/flow | TestOptionalAfterCallFlowEdges | 0.18 |
| internal/oracle | TestFunctionAdamicUnusedAnnotationRefused | 9.86 |
| internal/oracle | TestFunctionTypeScriptCallRefused | 10.06 |
| internal/oracle | TestFunctionAdamicLengthAnnotationRefused | 10.13 |
| internal/oracle | TestFunctionAdamicCallAnnotationRefused | 0.59 |
| internal/oracle | TestFunctionTypeScriptUnusedAccepted | 11.02 |
| internal/oracle | TestFunctionCallSignatureAgreesWithNode | 1.87 |
| internal/oracle | TestFunctionTypeScriptLengthAgreesWithNode | 9.40 |
| stage1/cohere/cssstrings | TestMultiPushGap | 19.45 |
| stage3/fixtures | TestFixtureDirectoriesHaveTopLevelTests | 0.02 |
| stage3/fixtures | TestFixturesAssignmentProofs | 24.90 |
| stage1/cohere/json | TestClosedMultiplePushGap | 3.60 |
| stage1/cohere/markdownblocks | TestParserRepresentationProbes | 27.83 |
| stage1/cohere/selector | TestEachGapStandsWhereGapsMdSaysItDoes | 0.00 |
| stage1/typescript/parser | TestPushSpreadGap | 3.45 |
| stage1/typescript/scanner | TestGapStandsWhereGapsMdSays | 2.28 |
| stage1/cohere/yaml | TestProduct_YAMLClosedGapsLowered | 3.41 |
| stage1/cohere/yaml | TestProduct_YAMLClosedGapsNative | 13.45 |
| stage1/cohere/yaml | TestLexerGaps | 13.88 |
| internal/lower | TestOptionalAfterCallPreservesEarlierThrow | 0.00 |
| internal/oracle | TestMergedFieldAdamicComputedRefused | 7.40 |
| internal/oracle | TestMergedFieldAdamicUnusedRefused | 7.49 |
| internal/oracle | TestMergedFieldTypeScriptComputedChecked | 8.68 |
| internal/oracle | TestMergedFieldTypeScriptInitializedAgreesWithNode | 8.88 |
| internal/oracle | TestMergedFieldTypeScriptUnusedAccepted | 1.56 |
| internal/oracle | TestMergedFieldAdamicReadRefused | 0.47 |
| internal/oracle | TestMergedFieldTypeScriptMissingChecked | 1.95 |
| internal/oracle | TestMergedFieldAdamicInitializedAgreesWithNode | 1.20 |
| internal/native | TestTSGoBuildSeesTheProgramsFeatures | 0.51 |
| internal/native | TestCheckerBuildsNameNoPlaceOfTheArchive | 7.84 |
| internal/native | TestTSGoCacheRelease | 9.79 |
| internal/native | TestNativeCachesWriteNothingUnderHome | 10.01 |
| internal/native | TestTSGoCacheSanitized | 13.41 |
| internal/oracle | TestOptionalAfterCallComputed | 22.03 |
| internal/oracle | TestOptionalAfterCallRequired | 22.05 |
| internal/oracle | TestOptionalAfterCall79 | 22.08 |
| internal/oracle | TestOptionalAfterCallPropagation | 22.22 |
| internal/oracle | TestOptionalAfterCall193 | 0.73 |
| internal/oracle | TestOptionalAfterCallVariants | 0.74 |
| internal/ir | TestCallTargetReaders | 40.04 |
| internal/oracle | TestJSONStringifyShorthandWrongBindingMutant | 5.16 |
| internal/oracle | TestJSONStringifyShorthand | 5.38 |
| internal/oracle | TestCallSpreadEvaluatedTwiceMutant | 17.39 |
| internal/oracle | TestCallSpreadArray | 24.32 |

test-seconds.json records every selected subtest too. counts-update.jsonl records every counts family/row elapsed value. Embedded mutants are the six TestAssignmentProof*Mutant leaves, TestCallSpreadEvaluatedTwiceMutant and TestJSONStringifyShorthandWrongBindingMutant. Each passes by rejecting its changed source or IR against independently captured Node stdout, after successful compilation and sanitized/release execution. TestTSGoBuildSeesTheProgramsFeatures also executes its existing negative control: omit the program feature flags and syntax checking fails; this is an include-order compile-contract control, not a claimed runtime kill.

Production mutant outcomes:

| Mutant | Catcher | Invocation seconds |
|---|---|---:|
| check-optional-receiver | TestOptionalAfterCall79 | 67.439 |
| javascript-required-panic | TestOptionalAfterCallRequired | 31.417 |
| native-required-panic | TestOptionalAfterCallRequired | 21.242 |
| drop-flow-throw-edge | TestOptionalAfterCallFlowEdges | 62.359 |
| drop-required-propagation | TestOptionalAfterCallPropagation | 45.879 |
| merged-computed-read | TestMergedFieldTypeScriptComputedChecked | 35.432 |
| merged-declaration | TestMergedFieldAdamicUnusedRefused | 22.461 |
| merged-read | TestMergedFieldTypeScriptMissingChecked | 27.255 |
| function-annotation | TestFunctionAdamicUnusedAnnotationRefused | 26.517 |
| function-call | TestFunctionTypeScriptCallRefused | 23.372 |
| function-length | TestFunctionTypeScriptLengthAgreesWithNode | 28.807 |
| erase-earlier-throw | TestOptionalAfterCallPreservesEarlierThrow | 15.548 |
| tsgo-absolute-main | TestTSGoCacheSanitized | 23.968 |
| tsgo-rewrite-runtime | TestTSGoCacheRelease | 16.262 |

All 14 production overlays exit 1 at test assertions, without build failure, clang warning/error kills or timeouts. run-mutants.py records exact commands and rejects those other failure modes. All mutant Go sources are .go.txt. The final two cache replays use the corrected test logging; cache-mutants-final.log records both catches. The first cache test logged its success text even after a failed assertion; the final test prints that text only when it has not failed.

After that logging correction, the two cache tests passed again (cache-final-green.jsonl), TestCallTargetReaders passed again in 27.33 s (readers-final-green.jsonl), and vet on all 14 changed packages exited 0 with no diagnostics (vet-after-log-fix.log). This is the only post-validation test edit; it changes no assertion or compiler behavior.

Final replay TestTSGoCacheRelease: 2.65 s.

Final replay TestTSGoCacheSanitized: 2.98 s.

Final replay TestCallTargetReaders: 27.33 s.
