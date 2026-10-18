Rebased #wvhe1jz step 30 hidden boundaries onto the exact requested main, with refusal placement and current product ownership fixed.
One commit on base dbdf977f5b8b0979d414ebe6f20e7aadd312e309, carrying source 566a5d637570b9449a120e1d8d001179aab7aca1's own change; delivery SHA is the branch tip.
Focused compiler/reader tests and vet pass; one Linux counts regeneration passes in 234.225 s, adding 33 rows and changing nine.
All 42 original overlay mutants, the ABI restoration, runtime-witness and canonical-owner mutants are caught; both in-test IR mutants are caught against Node.
Not covered: full gate, whole packages, full census/replay/meter, WASI, performance, or ahra loom check. No runtime C line changed in this unit.

Started with git fetch origin and git checkout -b compiler/loom-r-hidden-boundaries dbdf977f5b. Applied git cherry-pick --no-commit 566a5d637570b9449a120e1d8d001179aab7aca1, so this is the commit's own parent-to-child change, not a main-to-old-tip diff. Main's newer work remains the base. Historical source review evidence is carried as provenance; this directory contains fresh results. No cohere code was copied and its gitlink is unchanged.

There was one conflicted hunk, in internal/flow/flow_test.go's programs filter. Main added uninstrumentedAsyncFixture(path); the source added refusedHiddenBoundaryFixture(path). Resolution keeps main's filter unchanged, including its async exclusion, and moves both new refusal controls into internal/oracle/testdata/hidden_boundary_refused/. The explicit oracle refusal tests now load that directory. No new top-level refused fixture remains, and no unused hidden-boundary exception helper remains. The getter census helper still moves into its own test file as the source change intended. Exact competing and resolved lines are in conflicts.json; apply.log records the single conflict.

A clean merge also needed one current-main integration fix. The source command tests emit ADAMIC_CANONICAL_CLOSURES, but main's commandRuntimes initially declared only the empty sanitized set. First they failed for an unbuilt base archive; after warming that product, two tests still failed with no runtime product declared for ADAMIC_CANONICAL_CLOSURES. Added sanitized-canonical-closures and TestProduct_CommandRuntimeCanonicalClosures as a reader of internal/oracle TestProduct_NativeRuntimesSanitized2, with its wait in product-order.json. Oracle remains its sole builder. Current ownership, wait and missing-owner guards pass. This changes Go product declarations only, not C.

The conservative choices are to preserve main's async exclusions and Lean runtime implementation, move only this change's new refusal controls, and retain the source's paired TypeScript frontend controls and historical census templates unchanged. The 21 inherited .ts controls are byte-identical to the source commit; no new Adamic program was written as .ts. They are the original explicit TypeScript counterparts, not an invented replacement for .a programs. Provenance checks are in provenance.json.

Runtime C clearance: zero lines to list. The source commit's own internal/native/runtime diff is empty (runtime-source-own.diff), and this unit's diff from the exact base is empty too (runtime-c.diff). Newer main runtime code is retained, not replaced from the old source tree. runtime-versus-source-tree.diff is only a historical comparison against the old full tree, not this unit's runtime delta. host_void_method.a is unchanged from the source, including receiver.write reading this.prefix; TestRuntimeSetsAreEachUsed passes with it as the receiver-only witness. Main's LeanGet/LeanProduct wiring is preserved.

Toolchain: initial setup overlapped applying the source tree and failed with undefined overload lowering methods during go build. The resolved tree then compiled, and rerunning timeout 600 bash cloud/setup.sh succeeded. The inferred cause of the first failure is the changing source tree; the log records the observation. GOPROXY was https://proxy.golang.org|direct and the printed /workspace/adamic-tools/env.sh was sourced. Successful timing lines: Node 0.044 s, Go 0.041 s, submodules 0.095 s, markdown skipped step-duration=0.009 s / ready 0.117 s, clang 0.301 s, shared cache 0.875 s, go build 46.316 s, build cache warm 46.477 s, done 46.519 s. nproc=5, cpu.max=400000 100000; GOMAXPROCS=4 and test.parallel=4. House clang is 20.1.8, Go 1.27.1, Node 24.19.0. Setup logs preserve both attempts.

Counts were never hand-merged. The auto-merged table was discarded in favor of git show dbdf977f5b:internal/oracle/counts.md, then regenerated exactly once after the resolved oracle tests compiled. Read the current recorder and its table header; main's actual table has seven columns, with digests in keyed measurement products. Command: ADAMIC_GATE_UNCACHED=1 timeout 840 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 13m -parallel 4 -json -args -update-counts. Exit 0, 234.225 s, Linux with house clang 20.1.8. No second -update-counts run occurred. Later audits only measured individual rows and read the table. counts-regeneration.jsonl and counts-toolchain.log are the complete evidence.

The nine existing rows change only retains/releases; allocations, frees, peak and regions stay fixed:

| Fixture | Base retains/releases | Rebased retains/releases |
|---|---:|---:|
| internal/oracle/testdata/census_optional_values_forms.a | 52/169 | 51/168 |
| internal/oracle/testdata/census_optional_values_padding.a | 73/108 | 69/104 |
| internal/oracle/testdata/census_optional_values_system.a | 67/91 | 52/76 |
| internal/oracle/testdata/census_unary_numeric.a | 178/349 | 169/340 |
| internal/oracle/testdata/closure_convention_receiver_rest.a | 12/23 | 9/20 |
| internal/oracle/testdata/host_never_branches.a | 57/92 | 56/91 |
| internal/oracle/testdata/host_void_method.a | 14/18 | 12/16 |
| internal/oracle/testdata/user_iterators.a | 455/904 | 408/857 |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5/3 | 3/2 |

All 33 added rows are listed below. Full numeric rows and the exact diff are in counts-changes.json and counts.diff. No row was removed.

- internal/oracle/testdata/census_overload_binder_result_checked.a
- internal/oracle/testdata/hidden_boundary_generic_optional_array.a
- internal/oracle/testdata/hidden_boundary_generic_tnode.a
- internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a
- internal/oracle/testdata/hidden_boundary_never_array.a
- internal/oracle/testdata/hidden_boundary_never_array_observations.a
- internal/oracle/testdata/hidden_boundary_optional_array_order.a
- internal/oracle/testdata/hidden_wave2_method_destructuring.a
- internal/oracle/testdata/overload_callback_served.a
- internal/oracle/testdata/overload_field_hatch/kind-valid.ts
- internal/oracle/testdata/overload_field_hatch/value-undefined.ts
- internal/oracle/testdata/overload_field_hatch/value-valid.ts
- internal/oracle/testdata/overload_results_binding.a
- internal/oracle/testdata/overload_results_evaluate.a
- internal/oracle/testdata/overload_results_parameter.a
- internal/oracle/testdata/overload_results_scalar.a
- internal/oracle/testdata/overload_results_transform.a
- internal/oracle/testdata/overload_structural_block.a
- internal/oracle/testdata/overload_structural_evaluator.a
- internal/oracle/testdata/overload_structural_factory.a
- internal/oracle/testdata/overload_structural_fields.a
- internal/oracle/testdata/overload_values/module-valid.ts
- internal/oracle/testdata/overload_values/narrow-valid.ts
- internal/oracle/testdata/overload_values/order-valid.ts
- internal/oracle/testdata/overload_values/proven.a
- internal/oracle/testdata/overload_values/returned-valid.ts
- internal/oracle/testdata/overload_visitors/helper.a
- internal/oracle/testdata/overload_visitors/helper.ts
- internal/oracle/testdata/overload_visitors/instantiations.ts
- internal/oracle/testdata/overload_visitors/original.a
- internal/oracle/testdata/overload_visitors/original.ts
- internal/oracle/testdata/overload_visitors/overloaded-helper.a
- internal/oracle/testdata/overload_visitors/overloaded-helper.ts

Commands and results: every test command wrote directly to a log, never through a pipe. verify.py records exact arguments and results in the per-package JSON files and commands-results.json. Its bounded go test commands select named tests, use -count=1 -timeout 90s -parallel 4 -json, and have a 180 s command wall limit. Earlier command-test labels were reused by the original harness; their final files are the green results, while the integration errors above are the observed reasons for the corrections.

- timeout 600 python3 review/compiler/loom-r-hidden-boundaries/verify.py: final exit 0. cmd/adamic 25.182 s; flow 83.939 s including compilation (maximum leaf 19.34 s); export-data warmup 38.105 s; IR 64.944 s including compilation, TestCallTargetReaders 16.76 s; lower 13.257 s; focused oracle 27.615 s; selected native-oracle fixture families 53.593 s; changed-package vet 5.955 s. The oracle run includes all added hidden-boundary, overload, binder and destructuring tests, plus both IR mutation checks. Flow includes the new refusal classification, getter census checks, shared setup and all newly added top-level runnable fixtures through TestFlowCorpusRemainder. Oracle comparisons use source Node, JavaScript, native sanitizers and leak checks.
- timeout 180 go test ./cmd/adamic -run '^TestProduct_CommandRuntime(Sanitized|CanonicalClosures)$' -count=1 -timeout 90s -parallel 4 -json: exit 0, both needed products readied before ordinary command tests. command-runtime-products.log.
- timeout 180 go test ./cmd/adamic -run '^TestCommandRuntime(ArchivesHaveOneBuilder|ProductsWaitForOwners)$|^TestSharedCommandRuntimeCannotBuildItsOwnersArchive$' -count=1 -timeout 90s -parallel 4 -json: exit 0. Maximum leaf 10.26 s. command-runtime-guards.log.
- timeout 180 go test ./internal/native/runtimeproducts -run '^(TestProduct_LeanCensusExports|TestTheRuntimeBuildReachesOnlyItsDeclaredFiles)$' -count=1 -timeout 90s -parallel 4 -json: exit 0; runtime source census leaf 25.32 s. Its go vet also exits 0.
- timeout 180 go test ./stage3/fixtures -run '^TestProduct_FixtureOracleHook$' -count=1 -timeout 90s -parallel 4 -json: exit 0, product 25.79 s. Then the source host reader command with -run '^TestFixturesHost$/^host$/^13_createDirectory[.]a$': exit 0, selected fixture 22.51 s. Then -run '^TestFixturesCycles$/^cycles$/^(05_safe_load_read|06_import_order_mutant)$/^main[.]a$': exit 0, parent 3.53 s. Logs: stage3-hook-product.log, stage3-host.log, stage3-cycles.log.
- timeout 180 go test ./internal/native -run '^(TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestClosureConventionOwnerRuntimeSites|TestNoReaderCallingConvention|TestLibraryMapSetIteratorResources|TestRegExpIteratorResultShape)$' -count=1 -timeout 90s -parallel 4 -json: exit 0, maximum 7.34 s. native-readers.log.
- timeout 180 go test ./internal/javascript -run '^(TestViewMixedUnionUnknownAndUnavailable|TestViewMixedUnionSelectsNumber)$' -count=1 -timeout 90s -parallel 4 -json: exit 0, maximum 0.17 s. javascript-readers.log.
- timeout 180 go test -overlay review/compiler/loom-r-hidden-boundaries/counts-audit-overlay.json ./internal/oracle -run '^TestAuditClosureCounts' -count=1 -timeout 90s -parallel 4 -json: exit 0. All nine existing-row ABI counts and Node/backend observations pass; maximum 32.60 s. The original evidence-only audit is adapted only to exact-base and regenerated expected rows. counts-audit.log.
- Native agreement registration: timeout 180 go test ./internal/lower -run '^TestNativeAgreementRowsNameRegisteredPrograms$' -count=1 -timeout 90s, exit 0. Oracle registration/table: same bounds, -run '^(TestFamilyWrappersAreCurrent|TestFamiliesOwnEveryCaseOnce|TestCountsAreRecordedTable)$', exit 0. No counts remeasurement or rewrite in this static table check.
- gofmt -l of every changed Go file and git diff --check -- cmd internal cloud docs: clean. Supplemental vet ./internal/native/runtimeproducts exits 0. Every added/touched selected test event is under 60 s; the maximum positive event is 32.60 s. Detailed events are in test-seconds.csv. Count-recorder events are separately retained in the regeneration JSON log.

The original 42 mutant overlays were rerun from the source report's patches. timeout 600 python3 review/compiler/loom-r-hidden-boundaries/mutants.py reached its wall limit after 26 completed catches, during the next compilation. Resumed the same command, skipping recorded catches; final exit 0, all 42 caught. The interrupted compilation is not a catch. Each test command carries its own 90 s test timeout and bounded command wall time. There were no build-failure, C warning-only or test-timeout kills. Exact commands, exits, seconds and failing leaves are in mutants-results.json. Original mutants and current catchers:

| Mutant | Catcher |
|---|---|
| mutant-results-trust-optional-property | TestOverloadResultsRefusesOptionalResultProperty |
| mutant-results-trust-result | TestOverloadResultsLiarStopsEvaluateLiar, TestOverloadResultsLiarStopsLiar |
| mutant-results-drop-specialization | TestNativeAgreesWithNode_105/internal/oracle/testdata/overload_results_binding.a, TestNativeAgreesWithNode_105 |
| mutant-results-drop-parameter-check | TestOverloadResultsLiarStopsParameterLiar |
| mutant-results-trust-shared-result | TestOverloadResultsRefusesSharedReadonlyResult, TestOverloadResultsRefusesSharedWritableResult |
| mutant-callback-input | TestOverloadCallbackUnservedInput |
| mutant-callback-result | TestOverloadCallbackUnservedResult |
| mutant-callback-representation | TestOverloadCallbackUnservedRepresentation |
| mutant-callback-field-storage | TestOverloadCallbackFieldStorage |
| mutant-structural-drop-proof | TestOverloadStructuralResults |
| mutant-structural-trust-shape | TestOverloadStructuralRefusesMissingBlockField |
| mutant-structural-reverse-covariance | TestOverloadStructuralRefusesReverseReadonlyVariance |
| mutant-structural-mutable-variance | TestOverloadStructuralRefusesWritableVariance |
| mutant-fields-drop-field-storage | TestOverloadStructuralRefusesMixedFieldStorage |
| mutant-fields-drop-constraint-storage | TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintMutationRemainsNotYet |
| mutant-fields-object-layout | TestHiddenTNodeConstraintMutationRemainsNotYet |
| mutant-fields-array-layout | TestHiddenTNodeConstraintArrayLayout |
| mutant-fields-unconstrained-layout | TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown |
| mutant-fields-ignore-substitution | TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintString |
| mutant-fields-ignore-mapper | TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintObject |
| mutant-admission-drop-fixed-return-proof | TestOverloadStructuralResults |
| mutant-admission-reverse-result-covariance | TestOverloadStructuralRefusesFactoryLiteralCovariance |
| mutant-admission-trust-visitor-input | TestOverloadStructuralRefusesVisitorDomainLiar |
| mutant-admission-trust-writable-result | TestOverloadStructuralRefusesFactoryWritableInvariance |
| mutant-admission-drop-return-storage | TestOverloadStructuralRefusesFactoryFieldStorage |
| mutant-admission-drop-relation-diagnostic | TestOverloadStructuralRefusesFactoryResultCovariance |
| mutant-field-hatch-kind | TestOverloadFieldHatchKindLiar |
| mutant-field-hatch-value | TestOverloadFieldHatchValueLiar |
| mutant-values-drop-indirect-check | TestOverloadValuesReturnedLiar |
| mutant-values-skip-single-signature-check | TestOverloadValuesNarrowLiar |
| mutant-visitors-erase-TIn-to-Node | TestOverloadVisitorsHelperLiar |
| mutant-visitors-admit-unproven-invocation | TestOverloadVisitorsOverloadedHelperLiar |
| mutant-binder-result-drop-check | TestNativeAgreesWithNode_045/internal/oracle/testdata/census_overload_binder_result_checked.a, TestNativeAgreesWithNode_045 |
| mutant-default | TestNativeAgreesWithNode_000/internal/oracle/testdata/hidden_boundary_generic_tnode.a, TestNativeAgreesWithNode_000 |
| mutant-generic-value | TestHiddenTNodeGenericValueRemainsNotYet |
| mutant-never-storage | TestNativeAgreesWithNode_062/internal/oracle/testdata/hidden_boundary_never_array.a, TestNativeAgreesWithNode_062 |
| mutant-never-writable | TestHiddenNeverArrayRejectsWritableWidening |
| mutant-step3-receiver-proof-mutant | TestMethodDestructuringSafetyStructuralReceiver, TestMethodDestructuringSafetyDefaultReceiver, TestMethodDestructuringSafetyReceiver, TestMethodDestructuringSafetyArrowReceiver, TestMethodDestructuringSafetyExplicitReceiver |
| mutant-invoke-field-getters | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-invoke-code-getter | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-omit-data-mutation | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-admit-refused-getters | TestFlowGettersCensusRefusalClassification |

Supplemental mutants from the source proof, plus the new product integration:

- Restore only collections.go, iteration.go and iteration_origin.go from the exact dbdf977f5b base through counts-abi-mutant-overlay.json. The same nine TestAuditClosureCounts tests exit 1, all at numeric count attribution assertions. Node/backend code still compiles; this is the receiver ABI delta's catcher. Positive counts-audit.log and negative counts-abi-mutant.log retain all nine observations. Current-main files, not stale source-parent files, supply this mutation.
- Change only the receiver-only witness back to census_optional_values_system.a in an overlay of the current runtimeproducts file, retaining LeanGet and the declared-file implementation. TestRuntimeSetsAreEachUsed exits 1 in 3.73 s: the witness turns on [] rather than ADAMIC_CLOSURE_RECEIVERS. runtime-witness-mutant.log.
- Clear only the new canonical command runtime's owner. TestCommandRuntimeArchivesHaveOneBuilder exits 1 in 9.01 s: two builder packages [internal/oracle cmd/adamic]. canonical-runtime-owner-mutant.log.
- TestHiddenBoundary04PresentArrayMutant and TestHiddenBoundary04EagerIndexMutant pass by observing the deliberately wrong IR's outputs/side effects disagree with source Node in both backends, as the source report requires. Oracle lowered() still returns fresh IR per call, so these do not share mutated IR with parallel positive cases.

All overlay source copies are .go.txt, never compilable Go files in review. No full package/gate, historical census/replay/meter or WASI run is claimed. Inherited replay templates are reference artifacts rather than new live Go packages. The ahra CLI was not invoked, and no PR or Loom submission was made.
