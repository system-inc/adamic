Re-cut checked writes on current main, preserving its infrastructure and the source lowering.
Commit: delivery SHA is in the final response; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; source 27f67fc0d86d753ab7a0d2aa5143c0b4e15ea106.
Commands: 193 focused top-level tests PASS, maximum event 16.31s; counts PASS 63.672s; four guards PASS; readers PASS 23.598s; changed-package vet PASS.
Mutants: 19 standalone overlays, 45 embedded witnesses and one planted count row caught, 65 total.
Not covered: whole-package/full-repository gate, WASI, compiler census, benchmarks, or every unrelated consumer test. No PR or Loom submission.

The branch is one squash commit on the fetched main tip. All test output is in this directory. Source evidence is retained as inherited history. No cohere code was copied.

Conflicts and integration:

- Text conflict in internal/native/runtime_profile_test.go: keep main c06cf486, which deleted duplicate TestRuntimeReleasePaths after proving TestRuntimeStringViews catches both release mutants. No test was removed to meet a time limit.
- Semantic stale expectation in stage3/fixtures/objects/status.json: keep main 6bfe1389, which records the earlier line-79 NotYet for the optional ModuleSpecifierResolutionHost intersection. The source expected a later line-87 prototype-origin refusal. The first directory run caught this; the corrected object directory passes in 6.230s. Cycle and namespace directories passed in the first run.
- Main 1f07ce2d moved count digests into measurement-product keys and removed the digest column again. Its current counts.md and complete counts_test.go infrastructure are retained. Adapt only the three source count-registration files to prepare emitted C before returning a lazy measure closure. countPortable removes repository spelling from preparation inputs. No new compiler-consuming package or new runtime feature set is introduced.

Counts: reset to main, then run exactly one successful -update-counts on Linux with clang 20.1.8. There are 91 changed existing rows and 120 new rows, with none removed. Every numeric delta and every added row matches the source exactly. Existing rows add 230570 retains and releases each, with allocations, frees, peak and regions unchanged: receiver holds in ArrayPush and spliceArguments. Full before/after values and source-delta comparison: counts-row-changes.json. Counts update has 1178 passing test events, maximum 10.33s. The table structure check passes afterward.

internal/native/runtime: all 261 added/removed lines are identical to the source delta against merge base 6935e5933fbb74b28825a2fe221af3e0a65c1066. Differences: none. Machine-readable comparison: runtime-delta-comparison.json.

Guards and grain:

- TestEveryMapAProgramHoldsIsSorted passes with the source SortedMap fields, including CheckedWrites and ContractTypeNames. Both plain-map mutants fail this guard and the repeated-gob-bytes check.
- TestRuntimeSetsAreEachUsed passes. No runtime-set table change is needed.
- TestNativeAgreementRowsNameRegisteredPrograms passes. No added call bypasses the literal registry.
- TestEveryTestIsParallelOrSaysWhy passes. The inherited driver has its explicit Not parallel reason; the other added tests call t.Parallel first.
- TestCallTargetReaders passes with -timeout 90s.
- No new consumer package needs a compiler-dependencies.json entry. Existing consumers are exercised by focused driver, native, oracle, stage1 values, flow/fresh/JavaScript/regexp and three stage3 directories.
- No new product key names a build directory. Counts registrations use normalized emitted C as preparation inputs.
- All completed own top-level tests are under 60s on GOMAXPROCS=4, -parallel=4, cgroup cpu.max=400000 100000. No leaf needed another split. test-seconds.json records focused events; all other JSON logs preserve individual times.

Toolchain setup passed: Go .081s; clang .592s; Node .062s; markdown .984s; submodules 108.883s; shared cache 114.470s; Go build 413.475s; cache warm 413.556s; done 413.586s. nproc=5, quota 4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. Sourced /workspace/adamic-tools/env.sh and set GOPROXY=https://proxy.golang.org|direct. Early bounded commands expired compiling before any test started; their logs are retained and are not validation. A temporary exec-server disconnect recovered after wait_for_environment; mutant jobs continued.

Assumptions: preserve current main counts architecture even though the request predates its removal of the digest column. Inherited .ts files are TypeScript frontend probes and retain their stronger .a refusal controls; no new Adamic .ts program was authored.

Exact focused commands and outcomes: focused-results.json and run-focused.py. Standalone commands, seconds, selectors and outcomes: mutant-results.json and run-mutants.py. Final additional commands used -count=1, -parallel=4 where relevant and -timeout 90s, with outer timeout 120 or 150:

```text
go test ./internal/ir ./cmd/adamic-gate ./internal/lower ./internal/oracle -run "^(TestEveryMapAProgramHoldsIsSorted|TestEveryTestIsParallelOrSaysWhy|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$"
go test ./internal/ir -run ^TestCallTargetReaders$
go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle ./internal/regexp ./stage1/cohere/values ./stage3/fixtures
go test ./internal/flow ./internal/fresh ./internal/javascript ./internal/regexp -run "^(TestDebuggerHasNoFlowInstruction|TestMethodKeepsArgument|TestRegexOperationsDoNotPoisonTreeWrites|TestRegexOperandsStillJudgeCycleWrites|TestViewMixedUnionSelectsNumber|TestCaptureFactsCompilerTable|TestParse)$"
go test ./stage3/fixtures -run ^TestFixtures(Cycles|Namespaces|Objects)$
go test ./stage3/fixtures -run ^TestFixturesObjects$
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout 570s -args -update-counts
ADAMIC_FAMILY_PROBE=TestCountsAreRecorded=stage3/checked-writes/number-fit.ts go test ./internal/oracle -run ^TestCountsAreRecorded_026$/stage3/checked-writes/number-fit.ts$
go test ./internal/oracle -run ^TestCountsAreRecordedTable$
```

The planted row adds one recorded allocation through main's family probe. The real row measures 6/6/2/10/6/0, the planted row says 7/6/2/10/6/0; TestCountsAreRecorded_026 and its number-fit subtest fail with the expected changed-count diagnostic. The table is not modified by the probe.

Every standalone mutant and its catcher:

| Mutant | Catcher | Invocation seconds |
|---|---|---:|
| revert-native-reference-fields | ^TestCheckedRuntime | 23.978 |
| revert-js-reference-fields | ^TestCheckedRuntime | 10.887 |
| revert-array-field-schema | ^TestCheckedRuntime(Regex|Directory)Fields$ | 16.52 |
| drop-runtime-array-contract | TestRuntimeCheckedReferenceArrayKeepsContract | 14.501 |
| drop-string-heap-kind | TestRuntimeCheckedReferenceArrayKeepsContract | 14.21 |
| accept-runtime-array-literal | TestRuntimeCheckedReferenceArrayRejectsLiteral | 14.866 |
| accept-weaker-array-contract | TestRuntimeCheckedReferenceArrayRejectsWeakerContract | 14.881 |
| restore-number-view-native | ^TestCheckedWriteMessage02$ | 24.27 |
| restore-number-view-js | ^TestCheckedWriteMessage02$ | 10.41 |
| drop-js-runtime-array-contract | ^TestCheckedRuntimeReferenceArrayLaterWriter$ | 10.139 |
| drop-js-string-heap-kind | ^TestCheckedRuntimeReferenceArrayLaterWriter$ | 10.354 |
| prior-literal | ^TestRuntimeCheckedLibraryPushRejectsLiteral$ | 13.796 |
| drop-checked-write-metadata | ^TestObjectMetadataCheckedWritesIsPerEmission$ | 7.706 |
| skip-signature-instantiation | ^TestCheckedSignatureInstantiation$ | 9.258 |
| disable-mapped-identity | ^TestCheckedSignatureInstantiation$ | 9.043 |
| accept-optional-mapped-identity | ^TestCheckedSignatureInstantiation$ | 9.015 |
| skip-receiving-overloads | ^TestCheckedSignatureEveryReceivingOverload$ | 9.394 |
| plain-checked-writes | ^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt)$ | 14.666 |
| plain-contract-type-names | ^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt)$ | 13.717 |

Every embedded mutation witness (the named test asserts the mutated behavior differs from the pinned checked behavior or catches the changed representation):

| Witness | Seconds |
|---|---:|
| TestCheckedWriteMessageMutant13 | 0.5 |
| TestCheckedWriteMessageMutant03 | 0.45 |
| TestCheckedWriteMessageMutant16 | 0.42 |
| TestCheckedWriteMessageMutant15 | 0.47 |
| TestCheckedNeverContractMutant | 0.52 |
| TestCheckedWiderWriteMutants_drop_literal_set | 3.1 |
| TestCheckedFlowContainerContractMutants_flow_undefined_misfit | 0.49 |
| TestCheckedFlowContainerContractMutants_container_map_misfit | 0.47 |
| TestCheckedFlowContainerContractMutants_container_string_misfit | 0.46 |
| TestCheckedFlowContainerContractMutants_container_number_misfit | 0.42 |
| TestCheckedFlowContainerContractMutants_flow_array_misfit | 0.49 |
| TestCheckedWiderWriteMutants_drop_check | 0.51 |
| TestCheckedDiagnosticReferenceMutants_drop_nested_contract | 0.5 |
| TestCheckedFlowContainerContractMutants_flow_node_misfit | 0.51 |
| TestCheckedDiagnosticReferenceMutants_drop_allocation_proof | 0.52 |
| TestCheckedDiagnosticReferenceMutants_drop_alias_check | 0.51 |
| TestCheckedFlowContainerContractMutants_container_map_object_misfit | 3.16 |
| TestCheckedFlowContainerContractMutants_container_splice_misfit | 3.16 |
| TestCheckedEmitContractsMutants_emit_node_misfit | 3.22 |
| TestCheckedEmitContractsMutants_emit_callback_misfit | 3.25 |
| TestCheckedFlowContainerContractMutants_container_fill_misfit | 0.47 |
| TestCheckedFlowContainerContractMutants_container_boolean_misfit | 0.48 |
| TestCheckedFlowContainerContractMutants_container_nested_misfit | 0.44 |
| TestCheckedFlowContainerContractMutants_container_object_misfit | 0.49 |
| TestCheckedEmitContractsMutants_emit_drop_check | 0.52 |
| TestCheckedEmitContractsMutants_emit_auto_misfit | 0.51 |
| TestCheckedViewsNextMutants_required_boolean_misfit | 0.56 |
| TestCheckedViewsNextMutants_branch_bottom_number_misfit | 0.53 |
| TestCheckedViewsNextMutants_branch_length_misfit | 0.56 |
| TestCheckedFlowContainerContractMutants_container_alias_misfit | 0.49 |
| TestCheckedViewsNextMutants_branch_start_misfit | 0.62 |
| TestCheckedViewsNextMutants_branch_file_misfit | 0.58 |
| TestCheckedViewsNextMutants_optional_boolean_misfit | 0.51 |
| TestCheckedEmitContractsMutants_emit_resolution_misfit | 0.47 |
| TestCheckedEmitContractsMutants_emit_comment_misfit | 0.47 |
| TestCheckedViewsNextMutants_optional_false_misfit | 0.54 |
| TestCheckedViewsNextMutants_generic_site_misfit | 2.85 |
| TestCheckedViewsNextMutants_branch_bottom_string_misfit | 2.91 |
| TestCheckedViewsNextMutants_overload_result_misfit | 2.91 |
| TestCheckedViewsNextMutants_overload_callback_misfit | 2.91 |
| TestCheckedViewV2RepresentationMutants/float64 | 0.36 |
| TestCheckedViewV2RepresentationMutants/null | 0.37 |
| TestCheckedViewV2RepresentationMutants/uint8 | 0.36 |
| TestCheckedViewV2RepresentationMutants/int32 | 0.38 |
| TestCheckedViewV2RepresentationMutants/undefined | 0.31 |
