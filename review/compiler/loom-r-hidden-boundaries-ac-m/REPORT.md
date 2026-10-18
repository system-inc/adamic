Re-cut hidden boundaries and a-check headers onto current fetched main, preserving both merge intents.
Branch compiler/loom-r-hidden-boundaries-ac-m; base 5a94465370ea0172e25aae8c13478e59e40c2bb4; original tip 8c931b22c50f29466da3a1cbe3036784c9b447cf.
All requested recursive package suites, explicit reader guard, vet and gofmt pass functionally after the documented cache setup correction.
All 42 original compiler/flow mutants, both a-check header mutants, the old-generator parser mutant and the removed-wait mutant are caught.
Qualification: untouched TestSplitTSGoAgreesUnit00 cold product setup took 207.04 seconds; it was not rerun. No runtime C or header changes.

## Base, application and conflict resolutions

Fetched origin at the start and used its newest main, 5a94465370ea0172e25aae8c13478e59e40c2bb4, rather than the older 5179838b named in the request. Created compiler/loom-r-hidden-boundaries-ac-m from that commit. Used git merge-tree --write-tree --merge-base=8c931b22c50f29466da3a1cbe3036784c9b447cf~1 origin/main 8c931b22c50f29466da3a1cbe3036784c9b447cf, then git cherry-pick --no-commit 8c931b22c50f29466da3a1cbe3036784c9b447cf. Only the source commit's own delta is applied. merge-tree.log and apply.log retain the observations.

- cloud/fast-gate/product-order.json: the conflict was main's newly added moves waits versus this change's canonical-closures wait. Kept every one of main's 422 wait keys and their exact values, and appended only cmd/adamic TestProduct_CommandRuntimeCanonicalClosures -> internal/oracle TestProduct_NativeRuntimesSanitized2. The canonical set remains at index 2 and Build uses index modulo 3, so this owner is still correct.
- internal/oracle/counts.md: discarded the conflicted table, restored main's exact file, then regenerated it with the house Linux clang 20.1.8. No row was merged or edited manually. The generated 42 row deltas exactly equal the original change's rows; main's other rows and newer header remain intact.

One additional integration edit is necessary: main's cloud/fast-gate/command-runtime-order.py only accepted nil feature lists and refused the new explicit []string canonical-closures declaration. Extended that parser to accept either nil or a literal []string list. This extends main's generator; it does not revert its behavior or remove any entry. The old parser exits 1 before writing; the extended parser exits 0 and is byte-for-byte idempotent on the resolved JSON. order-generator-mutant.json records the old-parser mutant. No other main line was deliberately reverted outside the original source delta.

## Assumptions and limitations

Interpreted the source branch as the exact supplied one-commit tip, not a diff from new main. Retained its existing TypeScript control pairs and historical evidence unchanged. Treated opt-in and platform skips as skips, not coverage. The later request for complete recursive package testing took precedence over the earlier 30-minute first-push target; the first pushed commit includes the completed evidence rather than an incomplete claim.

The blanket 60-second cold-run requirement is not fully met: the untouched existing native TestSplitTSGoAgreesUnit00 took 207.04 seconds. In the full build-mode invocation this nonparallel consumer ran before its parallel product builders; the checker archive alone took 156.37 seconds. Main already declares TestProduct_TSGoArchive, TestProduct_SplitTSGoSource and TestProduct_SplitTSGoBaseline. The cause is the cold product preparation inside this developer-mode run, not a new source test. It was not rerun or hidden by removing coverage. cold-native-budget.json retains the setup phases. A separate Loom-style, products-first measurement of this consumer was not run. All test events in all Go test files added or modified by the original change were below 60 seconds, maximum 41.23 seconds; touched-test-grain.json records that superset. This distinction does not erase the observed full-run budget breach.

## Toolchain and recovered setup failures

Ran export GOPROXY='https://proxy.golang.org|direct', timeout 600 bash cloud/setup.sh and sourced /workspace/adamic-tools/env.sh. Setup passed. nproc=5, cgroup quota=4 CPUs, Go 1.27.1, Node v24.19.0, clang 20.1.8. Tests used GOMAXPROCS=4, ADAMIC_GOCACHE_STORE=off and TZ=UTC. Cumulative setup timings:

```
setup: node ready (0.068s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.022s
setup: markdown dependencies ready (0.146s)
setup: go ready (0.146s)
setup: submodules ready (0.249s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.422s)
setup: shared cache ready (2.143s)
setup: go build ready (74.017s)
setup: test binaries deferred (use --warm-tests) (74.181s)
setup: build cache warm (74.184s)
setup: build-flags commit=5a94465370ea0172e25aae8c13478e59e40c2bb4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.04 0.01 0.00 1/138 807 load-after=7.42 2.37 0.83 1/146 1900
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (74.351s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.m58WwI
```

The first counts attempt failed with no space left on device creating a local product directory. Root had zero available space, with 14 GB of disposable local product artifacts. Moved only that cache directory to the named scratch path and removed it, recovering 14 GB; source and review evidence were preserved. The failed attempt wrote no count table. Retried regeneration once successfully: counts-update-disk-full.log and results-disk-full.json retain the failure; counts-update.log retains the successful 129.911-second run.

The initial full command-package run could not find warmed archives because driver tests set temporary XDG_CACHE_HOME values. Retried the full package with ADAMIC_BUILD_CACHE_DIR=/home/agent/.cache/adamic-build explicitly inherited, as Loom supplies a fixed shared product directory. It passed without any test or compiler code change. Both attempts are retained, not counted as mutants.

## Exact validation commands and results

| Label | Command | Exit | Wall seconds | Max reported test seconds |
| --- | --- | ---: | ---: | ---: |
| counts-update | `timeout 840 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 13m -parallel 4 -json -args -update-counts` | 0 | 129.911 | 24.54 |
| command-runtime-products | `timeout 150 go test ./cmd/adamic -run ^TestProduct_CommandRuntime -count=1 -timeout 90s -parallel 4 -json` | 0 | 17.091 | 5.48 |
| cmd-adamic | `timeout 840 go test ./cmd/adamic/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 1 | 21.671 | 6.87 |
| internal-ir | `timeout 840 go test ./internal/ir/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 61.109 | 41.23 |
| internal-lower | `timeout 840 go test ./internal/lower/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 174.254 | 29.75 |
| internal-flow | `timeout 840 go test ./internal/flow/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 114.091 | 27.02 |
| internal-native | `timeout 840 go test ./internal/native/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 686.963 | 207.04 |
| internal-oracle | `timeout 840 go test ./internal/oracle/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 716.227 | 48.45 |
| stage3-fixtures | `timeout 840 go test ./stage3/fixtures/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 102.333 | 12.07 |
| stage3-census-latent-refusalrewrite | `timeout 840 go test ./stage3/census/latent/refusalrewrite/... -count=1 -timeout 13m -parallel 4 -p 1 -json` | 0 | 3.541 | 1.55 |
| vet | `timeout 180 go vet ./cmd/adamic/... ./internal/ir/... ./internal/lower/... ./internal/flow/... ./internal/native/... ./internal/oracle/... ./stage3/fixtures/... ./stage3/census/latent/refusalrewrite/...` | 0 | 6.146 | 0 |
| gofmt | `timeout 30 gofmt -l cmd/adamic/overload_field_checks_test.go cmd/adamic/overload_value_checks_test.go cmd/adamic/overload_visitor_checks_test.go cmd/adamic/runtimes_test.go internal/flow/corpus_coverage_test.go internal/flow/flow_test.go internal/flow/getters_census_test.go internal/flow/trace_test.go internal/ir/call_targets_guard_test.go internal/lower/census_overload_binders_test.go internal/lower/census_overload_proof.go internal/lower/census_small.go internal/lower/collections.go internal/lower/expression.go internal/lower/generic_constraint_storage.go internal/lower/hidden_boundary_generic_tnode_test.go internal/lower/hidden_never_array_test.go internal/lower/iteration.go internal/lower/iteration_origin.go internal/lower/iteration_test.go internal/lower/method_destructuring.go internal/lower/method_destructuring_test.go internal/lower/object.go internal/lower/overload_callback.go internal/lower/overload_callback_test.go internal/lower/overload_field_hatch.go internal/lower/overload_relations.go internal/lower/overload_result_fields.go internal/lower/overload_results.go internal/lower/overload_results_test.go internal/lower/overload_return_call.go internal/lower/overload_structural.go internal/lower/overload_structural_test.go internal/lower/overload_values.go internal/lower/overload_values_test.go internal/lower/overload_visitor_calls.go internal/lower/overload_visitor_domain.go internal/lower/overload_visitors_test.go internal/lower/representation_clock_source_test.go internal/native/runtimeproducts/runtimeproducts.go internal/oracle/census_overload_binders_test.go internal/oracle/hidden_boundary_04_test.go internal/oracle/hidden_boundary_generic_tnode_test.go internal/oracle/hidden_never_array_test.go internal/oracle/method_destructuring_test.go internal/oracle/overload_callback_test.go internal/oracle/overload_field_hatch_test.go internal/oracle/overload_results_test.go internal/oracle/overload_structural_test.go internal/oracle/overload_values_test.go internal/oracle/overload_visitors_test.go` | 0 | 0.065 | 0 |

Corrected full command run: env ADAMIC_BUILD_CACHE_DIR=/home/agent/.cache/adamic-build timeout 180 go test ./cmd/adamic/... -count=1 -timeout 90s -parallel 4 -p 1 -json. Exit 0; cmd-adamic-shared-cache.log records the complete run.

Explicit guard: timeout 120 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout 90s -json. Exit 0; reader-guard.log records it. gofmt -l output is empty. Source whitespace checks exclude historical review artifacts; inherited patch and log bytes are retained.

Final accepted runs contain 10015 pass events and 392 skip events, with zero failing events. Pass counts include nested test events and setup/product tests, not just independent leaves. Skips are listed in all-test-events.json; opt-in WASI and platform-only cases are not claimed as executed. The initial cache failures are excluded from this accepted-run total and described above. No complete repository gate, ahra invocation or Loom submission was run.

## Fresh a-check and mutant evidence

Built this branch's compiler with timeout 120 go build -o /workspace/scratch/hidden-ac-m-compiler ./cmd/adamic, then ran timeout 180 python3 review/compiler/loom-r-hidden-boundaries-ac-m/a_check.py. The unchanged pinned Gate.aCheck at fbac28c62493f27a788edc02a18bc8edb68de5da accepts all 60 changed programs: {'refused': 33, 'checked': 25, 'NotYet': 2}. Every program is byte-identical to the original supplied tip. Both removed-header and wrong-reason mutants fail, using real compiler diagnostics, and the header is restored. See a-check-statuses.json, a-check.json and a-check-mutants.json.

Ran timeout 840 python3 review/compiler/loom-r-hidden-boundaries-ac-m/mutants.py. The original 38 compiler overlays and four flow harness mutants all exit 1 at test assertions; the harness exits 0. No mutant is credited for a Go build failure, C compiler error, Werror or timeout. All copied Go mutant sources are .go.txt and cannot enter normal builds. The original patch sources remain referenced under review/compiler/hidden-boundaries-next/getters; this pass writes only new evidence.

| Mutant | Catching failed tests |
| --- | --- |
| mutant-results-trust-optional-property | TestOverloadResultsRefusesOptionalResultProperty |
| mutant-results-trust-result | TestOverloadResultsLiarStopsLiar, TestOverloadResultsLiarStopsEvaluateLiar |
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
| mutant-fields-drop-constraint-storage | TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintMutationRemainsNotYet, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintBoolean |
| mutant-fields-object-layout | TestHiddenTNodeConstraintMutationRemainsNotYet |
| mutant-fields-array-layout | TestHiddenTNodeConstraintArrayLayout |
| mutant-fields-unconstrained-layout | TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown |
| mutant-fields-ignore-substitution | TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintString |
| mutant-fields-ignore-mapper | TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintAny |
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
| mutant-step3-receiver-proof-mutant | TestMethodDestructuringSafetyReceiver, TestMethodDestructuringSafetyDefaultReceiver, TestMethodDestructuringSafetyStructuralReceiver, TestMethodDestructuringSafetyExplicitReceiver, TestMethodDestructuringSafetyArrowReceiver |
| mutant-invoke-field-getters | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-invoke-code-getter | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-omit-data-mutation | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-admit-refused-getters | TestFlowGettersCensusRefusalClassification |

## Every changed counts row

Each tuple is allocations, frees, retains, releases, peak live, in regions. New means there was no row on the fetched base. All other rows are unchanged.

| Fixture | Before | After |
| --- | --- | --- |
| internal/oracle/testdata/census_optional_values_forms.a | 102, 102, 52, 169, 23, 0 | 102, 102, 51, 168, 23, 0 |
| internal/oracle/testdata/census_optional_values_padding.a | 42, 42, 73, 108, 12, 0 | 42, 42, 69, 104, 12, 0 |
| internal/oracle/testdata/census_optional_values_system.a | 29, 29, 67, 91, 14, 0 | 29, 29, 52, 76, 14, 0 |
| internal/oracle/testdata/census_overload_binder_result_checked.a | new | 0, 0, 1, 0, 0, 0 |
| internal/oracle/testdata/census_unary_numeric.a | 187, 187, 178, 349, 14, 0 | 187, 187, 169, 340, 14, 0 |
| internal/oracle/testdata/closure_convention_receiver_rest.a | 11, 11, 12, 23, 7, 0 | 11, 11, 9, 20, 7, 0 |
| internal/oracle/testdata/hidden_boundary_generic_optional_array.a | new | 4, 4, 3, 7, 4, 0 |
| internal/oracle/testdata/hidden_boundary_generic_tnode.a | new | 2, 2, 0, 2, 2, 0 |
| internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a | new | 8, 8, 6, 14, 5, 0 |
| internal/oracle/testdata/hidden_boundary_never_array.a | new | 3, 3, 0, 4, 3, 0 |
| internal/oracle/testdata/hidden_boundary_never_array_observations.a | new | 17, 17, 8, 33, 5, 0 |
| internal/oracle/testdata/hidden_boundary_optional_array_order.a | new | 20, 20, 19, 40, 8, 0 |
| internal/oracle/testdata/hidden_wave2_method_destructuring.a | new | 26, 26, 23, 48, 10, 0 |
| internal/oracle/testdata/host_never_branches.a | 50, 50, 57, 92, 5, 0 | 50, 50, 56, 91, 5, 0 |
| internal/oracle/testdata/host_void_method.a | 5, 5, 14, 18, 5, 0 | 5, 5, 12, 16, 5, 0 |
| internal/oracle/testdata/overload_callback_served.a | new | 11, 11, 14, 20, 6, 0 |
| internal/oracle/testdata/overload_field_hatch/kind-valid.ts | new | 14, 14, 7, 20, 6, 0 |
| internal/oracle/testdata/overload_field_hatch/value-undefined.ts | new | 9, 9, 19, 23, 5, 0 |
| internal/oracle/testdata/overload_field_hatch/value-valid.ts | new | 9, 9, 17, 23, 5, 0 |
| internal/oracle/testdata/overload_results_binding.a | new | 6, 6, 14, 16, 5, 0 |
| internal/oracle/testdata/overload_results_evaluate.a | new | 9, 9, 13, 22, 5, 0 |
| internal/oracle/testdata/overload_results_parameter.a | new | 2, 2, 5, 6, 2, 0 |
| internal/oracle/testdata/overload_results_scalar.a | new | 3, 3, 11, 14, 2, 0 |
| internal/oracle/testdata/overload_results_transform.a | new | 8, 8, 9, 11, 5, 0 |
| internal/oracle/testdata/overload_structural_block.a | new | 6, 6, 3, 9, 4, 0 |
| internal/oracle/testdata/overload_structural_evaluator.a | new | 4, 4, 9, 14, 3, 0 |
| internal/oracle/testdata/overload_structural_factory.a | new | 9, 9, 8, 18, 5, 0 |
| internal/oracle/testdata/overload_structural_fields.a | new | 4, 4, 16, 18, 4, 0 |
| internal/oracle/testdata/overload_values/module-valid.ts | new | 4, 4, 8, 14, 3, 0 |
| internal/oracle/testdata/overload_values/narrow-valid.ts | new | 7, 7, 17, 23, 4, 0 |
| internal/oracle/testdata/overload_values/order-valid.ts | new | 6, 6, 12, 17, 5, 0 |
| internal/oracle/testdata/overload_values/proven.a | new | 7, 7, 12, 19, 4, 0 |
| internal/oracle/testdata/overload_values/returned-valid.ts | new | 13, 13, 26, 37, 6, 0 |
| internal/oracle/testdata/overload_visitors/helper.a | new | 5, 5, 11, 17, 4, 0 |
| internal/oracle/testdata/overload_visitors/helper.ts | new | 5, 5, 11, 17, 4, 0 |
| internal/oracle/testdata/overload_visitors/instantiations.ts | new | 11, 11, 11, 21, 9, 0 |
| internal/oracle/testdata/overload_visitors/original.a | new | 5, 5, 11, 17, 4, 0 |
| internal/oracle/testdata/overload_visitors/original.ts | new | 5, 5, 11, 17, 4, 0 |
| internal/oracle/testdata/overload_visitors/overloaded-helper.a | new | 5, 5, 11, 17, 4, 0 |
| internal/oracle/testdata/overload_visitors/overloaded-helper.ts | new | 5, 5, 11, 17, 4, 0 |
| internal/oracle/testdata/user_iterators.a | 663, 663, 455, 904, 64, 0 | 663, 663, 408, 857, 64, 0 |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5, 0, 5, 3, 5, 0 | 5, 0, 3, 2, 5, 0 |

## Runtime clearance

No line of internal/native/runtime/*.c or *.h changes in this re-cut. The original supplied commit's own runtime C/header diff is also empty, so there are no extra runtime lines to list for clearance. Main's newer runtime sources are preserved exactly. The host_void_method.a witness is unchanged, and its registration as the receiver-only witness is retained.

The missing-wait mutant removes only cmd/adamic TestProduct_CommandRuntimeCanonicalClosures from the resolved JSON. TestCommandRuntimeProductsWaitForOwners fails with the required Sanitized2 owner; the original JSON is restored byte for byte in finally. See removed-canonical-wait.json and removed-canonical-wait.log. No other wait is changed by this experiment.
