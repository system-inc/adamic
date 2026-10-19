Fixed deterministic gob encoding for checked-write field summaries and contract diagnostic names, preserving emitted C.
SHAs: source c63c16caf4fc2614cf6b4b63679a01b5b32d0997; base main 6935e5933fbb74b28825a2fe221af3e0a65c1066; delivery SHA is supplied in the final push response.
Commands and outputs: 195 focused leaves PASS (maximum 9.11s), gob regressions PASS .017s, 120 C outputs byte-identical, reader guard PASS 44.604s, vet PASS 32.011s.
Mutants: all prior 63 caught again, plus two plain-map regressions caught by both the ownership guard and differing serialized bytes.
Not covered: whole packages/full gate, WASI, stock compiler census and benchmarks. No PR or Loom submission is made.

This fixes Loom red chg_2sjgbmagn3jfj21rwey4v73gcv. Branch compiler/loom-checked-writes-2 starts from current origin/main and squashes origin/compiler/loom-checked-writes, with the correction in that same single commit. Main was fetched again immediately before preparing the commit and had not moved. The squash had no conflicts; main's newer expression, emitter, runtime and count-table changes merged automatically.

Program.CheckedWrites now has type SortedMap[string, bool]; Program.ContractTypeNames has type SortedMap[int, string]. Four lowering allocation sites now explicitly initialize ir.SortedMap: checkedOverloadResult, checkedWidening, finishWriteContracts and contractTypeID. The native metadata test's CheckedWrites literal is also explicitly sorted. Every reader/writer was audited in map-access-audit.txt. Indexing, len, range, nil tests and nil assignments retain exactly their Go-map semantics because SortedMap is a defined map type; native and JavaScript lookup readers therefore need no syntactic change. Only gob encoding uses its canonical encoder.

The existing TestAProgramIsTheSameBytesEveryTimeGobWritesIt now populates 16 entries in each of these maps, including false summary entries and key zero, encodes the complete program 32 times, and checks exact round-trip equality of both fields. TestEveryMapAProgramHoldsIsSorted reproduced both original errors before the fix, and passes afterward. The existing nil-map regression also passes. No serialization assertion or guard was weakened.

C identity was checked on the same current-main squash before and after only the sorted-map conversion. A scratch-only Go overlay adds TestCheckedWritesEmissionProbe to the existing oracle package, calling its real checkedWriteFixture, checkedRuntimeFieldProgram and checkedWriteMessageProgram adapters and native.C. It captures all 105 TypeScript witnesses, four proven Adamic fixtures, six runtime-field adapters and five exact-message adapters: 120 outputs, 771,905 bytes in each inventory. Python compared every raw byte string, not normalized text and not just hashes. All 120 match. emission-comparison.json gives lengths and both SHA-256 values; emission-before/ and emission-after/ retain the actual C as .c.txt files. The 11 refusal-only Adamic controls intentionally emit no C and were checked by their existing refusal tests. The scratch probe itself took 3.390s before and 2.840s after. Neither native/runtime nor counts.md changed as part of the sorted-map fix, and no count-table regeneration was needed.

The full previous focused inventory was rerun, including all checked-write fixtures against source Node, native release, ASan/UBSan native, JavaScript and leak checks where applicable. It also includes native metadata/runtime focused tests, both lower signature tests, main's five representation mutants and all three gob tests. All 195 top-level leaves passed; maximum leaf 9.11s and maximum entire focused invocation including build/setup 19.669s. No touched or added test exceeded 60s on this four-CPU quota. The reader guard passed in 44.604s; vet passed on all ten changed packages in 32.011s.

All 63 original mutants were rerun conservatively rather than narrowing the set by inferred map reachability: 40 embedded IR/diagnostic mutants, 17 standalone overlay mutants, five main representation mutants and the count-registration mutant. The latter fails precisely TestCountsAreRecorded_026/stage3/checked-writes/number-fit.ts (2.330s) and its owner shard, with the expected changed-count diagnostic. Every standalone overlay was re-cut onto this current-main implementation, retaining the SortedMap allocations, so a build/type failure is never counted as a kill. All 17 were caught by their intended assertions.

Two additional isolated mutations restore one field at a time to its plain map type. plain-checked-writes fails the map ownership guard specifically at Program.CheckedWrites and produces 12 encodings in 32. plain-contract-type-names fails specifically at Program.ContractTypeNames and produces 13 encodings in 32. Both compile successfully and fail the repeated-encoding test as well as the architectural check, proving the serialization check detects the actual nondeterministic bytes. Their test leaves took .030s and .040s; the full Go invocations were 16.239s and 16.469s.

Setup: export GOPROXY=https://proxy.golang.org|direct, then timeout 240 bash cloud/setup.sh. It completed without a workaround. Timing lines: Go .058s, submodules .058s, Node .076s, markdown .164s, clang .339s, shared cache 1.123s, Go build 53.614s, warm cache 53.676s, done 53.726s. nproc=5; cgroup cpu.max=400000 100000 (four CPU quota). Go 1.27.1, clang 20.1.8, Node 24.19.0. Sourced /workspace/adamic-tools/env.sh and set GOMAXPROCS=4 for final validation. No cohere code was copied.

Commands and full logs are recorded beside this report. All test output went directly to log files. The focused runner selects at most 20 named top-level leaves per invocation and uses -count=1 -parallel=4 -json -timeout=90s with an outer timeout 120. Standalone mutants use the corresponding -overlay file and reject build failures or timeouts as catchers. The emission probe is an explicitly selected scratch Go overlay ending in .go.txt; it is not a permanent test that writes the repository on Loom. No new package consumes the compiler, so no compiler-dependency declaration changes were needed. gofmt and git diff --check on changed source are clean.

```text
timeout 120 go test ./internal/ir -run '^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt|ASortedMapIsNilWhereAPlainMapIs)$' -count=1 -json -timeout 90s
timeout 600 python3 review/compiler/loom-checked-writes-2/run-focused.py
timeout 600 python3 review/compiler/loom-checked-writes-2/run-mutants.py
timeout 120 go test -overlay review/compiler/loom-checked-writes-2/emission-overlay.json ./internal/oracle -run '^TestCheckedWritesEmissionProbe$' -count=1 -json -timeout 90s
timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s
timeout 180 go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle ./internal/regexp ./stage1/cohere/values
```

Every focused test's seconds (all subcase and auxiliary timings are also in test-seconds.json):

| Package | Test | Seconds |
|---|---|---:|
| ./cmd/adamic | TestExplainChecksDriver | 0 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_container_fill_misfit | 8.36 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_container_number_misfit | 8.77 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_emit_node_misfit | 8.97 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_overload_callback_misfit | 9 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_overload_result_misfit | 1.51 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_flow_node_misfit | 1.36 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_optional_boolean_misfit | 1.59 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_number_misfit | 1.45 |
| ./cmd/adamic | TestExplainCheckedWritesOutput_container_map_misfit | 1.04 |
| ./internal/lower | TestCheckedSignatureEveryReceivingOverload | 0.15 |
| ./internal/lower | TestCheckedSignatureInstantiation | 0.15 |
| ./internal/native | TestRuntimeFieldLayoutsAreIncluded | 0.01 |
| ./internal/native | TestRuntimeCheckedReferenceArrayRejectsLiteral | 1.14 |
| ./internal/native | TestRuntimeCheckedReferenceArrayRejectsWeakerContract | 1.22 |
| ./internal/native | TestRegExpIteratorResultShape | 1.26 |
| ./internal/native | TestRuntimeCheckedReferenceArrayRejectsWrongElement | 0.17 |
| ./internal/native | TestObjectMetadataCheckedWritesIsPerEmission | 0 |
| ./internal/native | TestObjectMetadataCacheIsPerEmission | 0 |
| ./internal/native | TestRuntimeCheckedReferenceArrayKeepsContract | 0.14 |
| ./internal/native | TestRuntimeCheckedLibraryPushRejectsLiteral | 0.15 |
| ./internal/native | TestRuntimeCheckedLibraryPush | 9.11 |
| ./internal/native | TestRuntimeReleasePaths | 7.83 |
| ./internal/oracle | TestCheckedWriteMessage16 | 3.72 |
| ./internal/oracle | TestCheckedWriteMessageMutant13 | 3.73 |
| ./internal/oracle | TestCheckedWriteMessage03 | 3.9 |
| ./internal/oracle | TestCheckedRuntimeRegexFields | 4.04 |
| ./internal/oracle | TestCheckedWriteMessage13 | 0.75 |
| ./internal/oracle | TestCheckedWriteMessage15 | 0.98 |
| ./internal/oracle | TestCheckedRuntimeStatusFields | 0.81 |
| ./internal/oracle | TestCheckedWriteMessage02 | 0.71 |
| ./internal/oracle | TestCheckedViewsNextAdamicRefusals | 0.36 |
| ./internal/oracle | TestCheckedRuntimeReferenceArrayLaterWriter | 0.65 |
| ./internal/oracle | TestCheckedWriteMessageMutant03 | 0.61 |
| ./internal/oracle | TestCheckedViewsNextProvenAdamic | 0.9 |
| ./internal/oracle | TestCheckedWriteMessageMutant16 | 0.36 |
| ./internal/oracle | TestCheckedWiderWritesAdamicRefuses | 0.73 |
| ./internal/oracle | TestCheckedWriteMessageMutant15 | 0.83 |
| ./internal/oracle | TestCheckedWiderWritesProvenAdamic | 0.86 |
| ./internal/oracle | TestCheckedRuntimeFSFields | 0.92 |
| ./internal/oracle | TestCheckedRuntimeDirectoryFields | 0.48 |
| ./internal/oracle | TestCheckedRuntimeMapFields | 0.54 |
| ./internal/oracle | TestCheckedRuntimeSetFields | 0.57 |
| ./internal/oracle | TestCheckedWiderWrites_branch_bottom_number_misfit | 4.7 |
| ./internal/oracle | TestCheckedWiderWrites_required_boolean_misfit | 4.94 |
| ./internal/oracle | TestCheckedWiderWrites_overload_callback_fit | 5.13 |
| ./internal/oracle | TestCheckedWiderWrites_branch_start_fit | 5.14 |
| ./internal/oracle | TestCheckedWiderWrites_branch_length_misfit | 1.08 |
| ./internal/oracle | TestCheckedWiderWrites_branch_start_misfit | 1.11 |
| ./internal/oracle | TestCheckedWiderWrites_branch_bottom_number_fit | 1.69 |
| ./internal/oracle | TestCheckedWiderWrites_branch_length_fit | 1.92 |
| ./internal/oracle | TestCheckedWiderWrites_branch_file_misfit | 1.26 |
| ./internal/oracle | TestCheckedWiderWrites_optional_false_misfit | 1.19 |
| ./internal/oracle | TestCheckedWiderWrites_branch_file_fit | 1.64 |
| ./internal/oracle | TestCheckedWiderWritesKeepSpreadOverrideRefusal | 0.07 |
| ./internal/oracle | TestCheckedWiderWrites_optional_false_fit | 1.34 |
| ./internal/oracle | TestCheckedWiderWritesProvenContainers | 1.63 |
| ./internal/oracle | TestCheckedEmitProvenAdamic | 1.35 |
| ./internal/oracle | TestCheckedNeverContractMutant | 1.26 |
| ./internal/oracle | TestCheckedWiderWrites_optional_boolean_misfit | 1.09 |
| ./internal/oracle | TestCheckedWiderWrites_optional_boolean_fit | 1.73 |
| ./internal/oracle | TestCheckedWiderWrites_overload_callback_misfit | 1.22 |
| ./internal/oracle | TestCheckedWiderWrites_required_boolean_fit | 1.58 |
| ./internal/oracle | TestCheckedWiderWrites_emit_auto_misfit | 6.81 |
| ./internal/oracle | TestCheckedWiderWrites_overload_result_misfit | 6.94 |
| ./internal/oracle | TestCheckedWiderWrites_branch_bottom_string_fit | 6.94 |
| ./internal/oracle | TestCheckedWiderWrites_emit_callback_fit | 7.08 |
| ./internal/oracle | TestCheckedWiderWrites_emit_comment_misfit | 0.9 |
| ./internal/oracle | TestCheckedWiderWrites_emit_resolution_misfit | 1.11 |
| ./internal/oracle | TestCheckedWiderWrites_emit_comment_fit | 1.14 |
| ./internal/oracle | TestCheckedWiderWrites_emit_resolution_fit | 1.4 |
| ./internal/oracle | TestCheckedWiderWrites_generic_site_misfit | 0.99 |
| ./internal/oracle | TestCheckedWiderWrites_overload_result_fit | 1.09 |
| ./internal/oracle | TestCheckedWiderWrites_branch_bottom_string_misfit | 0.96 |
| ./internal/oracle | TestCheckedWiderWrites_generic_site_fit | 1.2 |
| ./internal/oracle | TestCheckedWiderWrites_emit_node_misfit | 0.97 |
| ./internal/oracle | TestCheckedWiderWrites_emit_auto_fit | 1.29 |
| ./internal/oracle | TestCheckedWiderWrites_emit_callback_misfit | 0.89 |
| ./internal/oracle | TestCheckedWiderWrites_container_slice_misfit | 0.78 |
| ./internal/oracle | TestCheckedWiderWrites_emit_node_fit | 1.45 |
| ./internal/oracle | TestCheckedWiderWrites_container_map_values_misfit | 1.08 |
| ./internal/oracle | TestCheckedWiderWrites_container_map_values_fit | 1.42 |
| ./internal/oracle | TestCheckedWiderWrites_container_slice_fit | 1.15 |
| ./internal/oracle | TestCheckedWiderWrites_container_nested_misfit | 4.59 |
| ./internal/oracle | TestCheckedWiderWrites_container_splice_misfit | 4.62 |
| ./internal/oracle | TestCheckedWiderWrites_container_string_fit | 4.83 |
| ./internal/oracle | TestCheckedWiderWrites_container_boolean_fit | 4.92 |
| ./internal/oracle | TestCheckedWiderWrites_container_number_misfit | 0.64 |
| ./internal/oracle | TestCheckedWiderWrites_container_number_fit | 0.76 |
| ./internal/oracle | TestCheckedWiderWrites_container_alias_misfit | 0.58 |
| ./internal/oracle | TestCheckedWiderWrites_container_alias_fit | 0.88 |
| ./internal/oracle | TestCheckedWiderWrites_container_map_object_misfit | 0.65 |
| ./internal/oracle | TestCheckedWiderWrites_container_fill_misfit | 0.52 |
| ./internal/oracle | TestCheckedWiderWrites_container_map_object_fit | 0.98 |
| ./internal/oracle | TestCheckedWiderWrites_container_boolean_misfit | 0.61 |
| ./internal/oracle | TestCheckedWiderWrites_container_object_misfit | 0.59 |
| ./internal/oracle | TestCheckedWiderWrites_container_fill_fit | 0.82 |
| ./internal/oracle | TestCheckedWiderWrites_container_nested_fit | 0.94 |
| ./internal/oracle | TestCheckedWiderWrites_container_map_misfit | 0.88 |
| ./internal/oracle | TestCheckedWiderWrites_container_splice_fit | 1.07 |
| ./internal/oracle | TestCheckedWiderWrites_container_object_fit | 1.07 |
| ./internal/oracle | TestCheckedWiderWrites_container_map_fit | 0.64 |
| ./internal/oracle | TestCheckedWiderWrites_container_string_misfit | 0.57 |
| ./internal/oracle | TestCheckedWiderWrites_flow_array_misfit | 2.95 |
| ./internal/oracle | TestCheckedWiderWrites_never_wide_fit | 3.09 |
| ./internal/oracle | TestCheckedWiderWrites_flow_node_fit | 3.1 |
| ./internal/oracle | TestCheckedWiderWrites_flow_array_fit | 3.13 |
| ./internal/oracle | TestCheckedWiderWrites_never_splice_misfit | 0.3 |
| ./internal/oracle | TestCheckedWiderWrites_never_index_misfit | 0.31 |
| ./internal/oracle | TestCheckedWiderWrites_never_nullable_misfit | 0.36 |
| ./internal/oracle | TestCheckedWiderWrites_never_nullable_fit | 0.51 |
| ./internal/oracle | TestCheckedWiderWrites_flow_undefined_misfit | 0.36 |
| ./internal/oracle | TestCheckedWiderWrites_flow_undefined_fit | 0.61 |
| ./internal/oracle | TestCheckedWiderWrites_never_object_fit | 0.52 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_alias_misfit | 0.38 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_proof_fit | 0.56 |
| ./internal/oracle | TestCheckedWiderWrites_never_object_misfit | 0.4 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_alias_fit | 0.57 |
| ./internal/oracle | TestCheckedWiderWrites_never_string_misfit | 0.42 |
| ./internal/oracle | TestCheckedWiderWrites_never_number_misfit | 0.3 |
| ./internal/oracle | TestCheckedWiderWrites_never_string_fit | 0.56 |
| ./internal/oracle | TestCheckedWiderWrites_flow_node_misfit | 0.38 |
| ./internal/oracle | TestCheckedWiderWrites_never_number_fit | 0.43 |
| ./internal/oracle | TestCheckedWiderWrites_boolean_misfit | 2.86 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_proof_misfit | 2.89 |
| ./internal/oracle | TestCheckedWiderWrites_number_fit | 3.04 |
| ./internal/oracle | TestCheckedWiderWrites_flags_fit | 3.07 |
| ./internal/oracle | TestCheckedWiderWrites_string_misfit | 0.39 |
| ./internal/oracle | TestCheckedWiderWrites_boolean_fit | 0.57 |
| ./internal/oracle | TestCheckedWiderWrites_flags_misfit | 0.42 |
| ./internal/oracle | TestCheckedWiderWrites_string_fit | 0.55 |
| ./internal/oracle | TestCheckedWiderWrites_reference_structure | 0.37 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_inner_misfit | 0.37 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_nested_fit | 0.66 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_inner_fit | 0.68 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_nested_misfit | 0.5 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_misfit | 0.53 |
| ./internal/oracle | TestCheckedWiderWrites_number_misfit | 0.54 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_related_misfit | 0.59 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_rich_misfit | 0.53 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_related_fit | 0.79 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_rich_fit | 0.67 |
| ./internal/oracle | TestCheckedWiderWrites_diagnostic_fit | 0.44 |
| ./internal/oracle | TestCheckedWiderWrites_range_misfit | 2.88 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_flow_undefined_misfit | 2.93 |
| ./internal/oracle | TestCheckedWiderWriteMutants_drop_literal_set | 2.99 |
| ./internal/oracle | TestCheckedWiderWrites_compound_fit | 3.09 |
| ./internal/oracle | TestCheckedWiderWrites_parent_misfit | 0.45 |
| ./internal/oracle | TestCheckedWiderWriteMutants_drop_check | 0.55 |
| ./internal/oracle | TestCheckedWiderWrites_parent_fit | 0.63 |
| ./internal/oracle | TestCheckedWiderWrites_receiver_fit | 0.61 |
| ./internal/oracle | TestCheckedWiderWrites_nullable_number_misfit | 0.41 |
| ./internal/oracle | TestCheckedWiderWrites_range_fit | 0.59 |
| ./internal/oracle | TestCheckedDiagnosticReferenceMutants_drop_nested_contract | 0.38 |
| ./internal/oracle | TestCheckedWiderWrites_nullable_number_fit | 0.5 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_flow_node_misfit | 0.36 |
| ./internal/oracle | TestCheckedDiagnosticReferenceMutants_drop_allocation_proof | 0.39 |
| ./internal/oracle | TestCheckedDiagnosticReferenceMutants_drop_alias_check | 0.36 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_string_misfit | 0.36 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_map_misfit | 0.36 |
| ./internal/oracle | TestCheckedWiderWrites_compound_misfit | 0.34 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_number_misfit | 0.34 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_flow_array_misfit | 0.34 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_splice_misfit | 2.8 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_map_object_misfit | 2.86 |
| ./internal/oracle | TestCheckedEmitContractsMutants_emit_node_misfit | 2.87 |
| ./internal/oracle | TestCheckedEmitContractsMutants_emit_callback_misfit | 2.89 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_fill_misfit | 0.45 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_boolean_misfit | 0.59 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_nested_misfit | 0.59 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_object_misfit | 0.61 |
| ./internal/oracle | TestCheckedViewsNextMutants_required_boolean_misfit | 0.52 |
| ./internal/oracle | TestCheckedViewsNextMutants_branch_bottom_number_misfit | 0.53 |
| ./internal/oracle | TestCheckedViewsNextMutants_branch_length_misfit | 0.6 |
| ./internal/oracle | TestCheckedViewsNextMutants_branch_start_misfit | 0.69 |
| ./internal/oracle | TestCheckedViewsNextMutants_branch_file_misfit | 0.54 |
| ./internal/oracle | TestCheckedViewsNextMutants_optional_false_misfit | 0.53 |
| ./internal/oracle | TestCheckedViewsNextMutants_optional_boolean_misfit | 0.62 |
| ./internal/oracle | TestCheckedEmitContractsMutants_emit_drop_check | 0.69 |
| ./internal/oracle | TestCheckedFlowContainerContractMutants_container_alias_misfit | 0.63 |
| ./internal/oracle | TestCheckedEmitContractsMutants_emit_resolution_misfit | 0.52 |
| ./internal/oracle | TestCheckedEmitContractsMutants_emit_auto_misfit | 0.55 |
| ./internal/oracle | TestCheckedEmitContractsMutants_emit_comment_misfit | 0.37 |
| ./internal/oracle | TestCountsAreRecordedTable | 0.02 |
| ./internal/oracle | TestFamilyWrappersAreCurrent | 0.04 |
| ./internal/oracle | TestFamiliesOwnEveryCaseOnce | 0.09 |
| ./internal/oracle | TestCheckedViewsNextMutants_overload_callback_misfit | 3.68 |
| ./internal/oracle | TestCheckedViewsNextMutants_generic_site_misfit | 3.62 |
| ./internal/oracle | TestCheckedViewsNextMutants_overload_result_misfit | 3.7 |
| ./internal/oracle | TestCheckedViewsNextMutants_branch_bottom_string_misfit | 3.77 |
| ./internal/oracle | TestCheckedViewV2RepresentationMutants | 0 |
| ./stage1/cohere/values | TestEachGapStandsWhereGapsMdSaysItDoes | 0 |
| ./internal/ir | TestEveryMapAProgramHoldsIsSorted | 0 |
| ./internal/ir | TestASortedMapIsNilWhereAPlainMapIs | 0.01 |
| ./internal/ir | TestAProgramIsTheSameBytesEveryTimeGobWritesIt | 0.02 |

Every standalone mutant and its catcher (all exit 1, caught=true):

| Mutant | Catcher | Invocation seconds |
|---|---|---:|
| revert-native-reference-fields | ^TestCheckedRuntime | 27.644 |
| revert-js-reference-fields | ^TestCheckedRuntime | 24.536 |
| revert-array-field-schema | ^TestCheckedRuntime(Regex|Directory)Fields$ | 28.802 |
| drop-runtime-array-contract | TestRuntimeCheckedReferenceArrayKeepsContract | 26.577 |
| drop-string-heap-kind | TestRuntimeCheckedReferenceArrayKeepsContract | 17.877 |
| accept-runtime-array-literal | TestRuntimeCheckedReferenceArrayRejectsLiteral | 17.985 |
| accept-weaker-array-contract | TestRuntimeCheckedReferenceArrayRejectsWeakerContract | 17.875 |
| restore-number-view-native | ^TestCheckedWriteMessage02$ | 21.563 |
| restore-number-view-js | ^TestCheckedWriteMessage02$ | 11.272 |
| drop-js-runtime-array-contract | ^TestCheckedRuntimeReferenceArrayLaterWriter$ | 12.053 |
| drop-js-string-heap-kind | ^TestCheckedRuntimeReferenceArrayLaterWriter$ | 11.541 |
| prior-literal | ^TestRuntimeCheckedLibraryPushRejectsLiteral$ | 14.573 |
| drop-checked-write-metadata | ^TestObjectMetadataCheckedWritesIsPerEmission$ | 9.933 |
| skip-signature-instantiation | ^TestCheckedSignatureInstantiation$ | 11.59 |
| disable-mapped-identity | ^TestCheckedSignatureInstantiation$ | 11.09 |
| accept-optional-mapped-identity | ^TestCheckedSignatureInstantiation$ | 10.757 |
| skip-receiving-overloads | ^TestCheckedSignatureEveryReceivingOverload$ | 11.371 |
| plain-checked-writes | ^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt)$ | 16.239 |
| plain-contract-type-names | ^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt)$ | 16.469 |
