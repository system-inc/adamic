Re-cut hidden-boundary compiler fixes and the getters flow harness fix as one squash on main.
Source 65392f86cc66a4f54f8651e9102713db0a2e37a2; base main c0fd8945b872e57a341ccee5973e3f4d1b75c34b; delivery SHA is supplied in the final report and git HEAD.
Focused compiler, flow, oracle and getters checks plus changed-package vet passed; command outputs and seconds are recorded below.
Compiler and harness mutants with their actual failing tests are recorded below, alongside two IR semantic mutants and the ABI attribution mutant.
Full packages, full gate, full census, historical replay/meter tools, WASI and performance were not run.

The branch is compiler/loom-hidden-boundaries. No PR or Loom submission was made. The source net change includes historical review artifacts and census templates; those are preserved, not treated as fresh observations. New evidence lives only in this directory. No cohere source was copied and its gitlink did not change.

Conflict resolutions

- internal/flow/flow_test.go had conflicts in the eligibility expression and trailing refusal helper. Kept main's shared lowered-program cache and existing constructor/non-null eligibility, added hidden-boundary eligibility, and kept the source's reason-checked getters table in getters_census_test.go. Removed the duplicate main helper because the source table owns the same eighteen refusals and checks their actual stops. Three runnable getters remain eligible.
- internal/flow/trace_test.go had one conflict in object field snapshots. Kept main's cached immutable trace preparation and chose descriptor.value for data fields and a constant marker for accessors. The auto-merged code-property guard also reads its descriptor. This preserves main's no-getter intent and closes its remaining code-getter read while keeping nested data mutation observations.

Rules and assumptions

Every completed test leaf is recorded in test-seconds.csv, including failures and mutant catches. No added or touched test exceeded 60 seconds, and none was deleted or split. nproc is 5; cpu.max is 400000 100000, a four-CPU quota. Verification uses GOMAXPROCS=4 and -parallel 4. Node is v24.19.0, Go is go1.27.1 and clang is 20.1.8.

The source adds no compiler-consuming package or product, so cloud/fast-gate/compiler-dependencies.json needs no entry. Changed tests have no hard-coded machine/build directory; temporary output names remain local test scratch. Main's export-data product is warmed by a bounded go list before TestCallTargetReaders. Source .ts files are retained as ordinary TypeScript hatch controls paired with .a refusal controls: renaming them would erase the tested distinction. No new Adamic source was authored outside .a. Historical evidence stays at its source paths to preserve the requested net change.

Setup and recovered failures

The initial timeout 600 bash cloud/setup.sh exited 124 during build-cache warming. Its submodules completed at 499.726 seconds and shared cache at 505.355 seconds. The first test attempt could not load cohere/TypeScript/tsc/go.mod while checkout was pending. A subsequent cold Go build reached its outer 180-second limit before any test leaf ran. Both logs are preserved. Setup retry completed successfully; its timing lines follow.

```text
setup: go ready (0.021s)
setup: node ready (0.022s)
setup: submodules ready (0.056s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: markdown dependencies ready (0.072s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.194s)
setup: shared cache ready (13.464s)
setup: go build ready (92.839s)
setup: test binaries deferred (use --warm-tests) (93.050s)
setup: build cache warm (93.055s)
setup: build-flags commit=c0fd8945b872e57a341ccee5973e3f4d1b75c34b nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=2.36 2.01 0.89 2/181 7815 load-after=6.73 3.92 1.69 6/226 13144
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (93.105s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.5Au7M1
```

The first counts attempt failed on main's missing @types/node 25.3.3 and did not rewrite counts.md. Installed the exact stage3/api/package-lock.json dependencies with npm ci --ignore-scripts. The following successful recorder invocation rewrote the table once with zero difference from the squashed source table. The mutant batch reached its outer 600-second limit after all 38 compiler catches were recorded; resumed only the four unfinished harness mutants under another hard limit. The verification evidence parser initially failed on pretty-printed go list JSON strings; fixed it to accept only object events and resumed without rerunning passed tests.

Production compiler files changed against main

- internal/lower/census_overload_proof.go
- internal/lower/census_small.go
- internal/lower/collections.go
- internal/lower/expression.go
- internal/lower/generic_constraint_storage.go
- internal/lower/iteration.go
- internal/lower/iteration_origin.go
- internal/lower/method_destructuring.go
- internal/lower/object.go
- internal/lower/overload_callback.go
- internal/lower/overload_field_hatch.go
- internal/lower/overload_relations.go
- internal/lower/overload_result_fields.go
- internal/lower/overload_results.go
- internal/lower/overload_return_call.go
- internal/lower/overload_structural.go
- internal/lower/overload_values.go
- internal/lower/overload_visitor_calls.go
- internal/lower/overload_visitor_domain.go

No internal/native runtime C or header file changed. The delivery also retains the source edits to stage3/census/latent/{full,replay,units}.go.txt and replay/main.go.txt, plus its documentation and cycles status metadata; these historical census paths were not remeasured.

Counts attribution

33 new fixture rows and 9 existing rows change against main; no rows were removed. The full diff is counts.diff. Existing allocations, frees, peak and regions stay identical. Restoring only main's collections.go, iteration.go and iteration_origin.go reproduces all nine old rows; current sources reproduce all nine new rows. The wrong-ABI mutant must fail every owning count assertion.

| Existing fixture | Main retains/releases | New retains/releases | Cause |
|---|---:|---:|---|
| internal/oracle/testdata/census_optional_values_system.a | 67/91 | 52/76 | 15 unused sys/process receivers |
| internal/oracle/testdata/census_optional_values_forms.a | 52/169 | 51/168 | 1 unused object.run receiver |
| internal/oracle/testdata/census_optional_values_padding.a | 73/108 | 69/104 | 4 unused literal callback receivers |
| internal/oracle/testdata/census_unary_numeric.a | 178/349 | 169/340 | 9 unused coercion receivers |
| internal/oracle/testdata/closure_convention_receiver_rest.a | 12/23 | 9/20 | 3 unused Runner.run receivers |
| internal/oracle/testdata/host_never_branches.a | 57/92 | 56/91 | 1 unused system.abort receiver |
| internal/oracle/testdata/host_void_method.a | 14/18 | 12/16 | 2 unused system.write receivers |
| internal/oracle/testdata/user_iterators.a | 455/904 | 408/857 | 47 unused iterator factory/next/return receivers |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5/3 | 3/2 | factory pair removed; next retain removed before TDZ stop |

New rows, each attributable to its registered source fixture:

```text
| internal/oracle/testdata/census_overload_binder_result_checked.a | 0 | 0 | 1 | 0 | 0 | 0 |
| internal/oracle/testdata/hidden_boundary_optional_array_order.a | 20 | 20 | 19 | 40 | 8 | 0 |
| internal/oracle/testdata/hidden_boundary_generic_optional_array.a | 4 | 4 | 3 | 7 | 4 | 0 |
| internal/oracle/testdata/hidden_boundary_generic_tnode.a | 2 | 2 | 0 | 2 | 2 | 0 |
| internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a | 8 | 8 | 6 | 14 | 5 | 0 |
| internal/oracle/testdata/hidden_boundary_never_array.a | 3 | 3 | 0 | 4 | 3 | 0 |
| internal/oracle/testdata/hidden_boundary_never_array_observations.a | 17 | 17 | 8 | 33 | 5 | 0 |
| internal/oracle/testdata/hidden_wave2_method_destructuring.a | 26 | 26 | 23 | 48 | 10 | 0 |
| internal/oracle/testdata/overload_callback_served.a | 11 | 11 | 14 | 20 | 6 | 0 |
| internal/oracle/testdata/overload_field_hatch/kind-valid.ts | 14 | 14 | 7 | 20 | 6 | 0 |
| internal/oracle/testdata/overload_field_hatch/value-valid.ts | 9 | 9 | 17 | 23 | 5 | 0 |
| internal/oracle/testdata/overload_field_hatch/value-undefined.ts | 9 | 9 | 19 | 23 | 5 | 0 |
| internal/oracle/testdata/overload_results_transform.a | 8 | 8 | 9 | 11 | 5 | 0 |
| internal/oracle/testdata/overload_results_evaluate.a | 9 | 9 | 13 | 22 | 5 | 0 |
| internal/oracle/testdata/overload_results_binding.a | 6 | 6 | 14 | 16 | 5 | 0 |
| internal/oracle/testdata/overload_results_parameter.a | 2 | 2 | 5 | 6 | 2 | 0 |
| internal/oracle/testdata/overload_results_scalar.a | 3 | 3 | 11 | 14 | 2 | 0 |
| internal/oracle/testdata/overload_structural_block.a | 6 | 6 | 3 | 9 | 4 | 0 |
| internal/oracle/testdata/overload_structural_evaluator.a | 4 | 4 | 9 | 14 | 3 | 0 |
| internal/oracle/testdata/overload_structural_fields.a | 4 | 4 | 16 | 18 | 4 | 0 |
| internal/oracle/testdata/overload_structural_factory.a | 9 | 9 | 8 | 18 | 5 | 0 |
| internal/oracle/testdata/overload_values/proven.a | 7 | 7 | 12 | 19 | 4 | 0 |
| internal/oracle/testdata/overload_values/returned-valid.ts | 13 | 13 | 26 | 37 | 6 | 0 |
| internal/oracle/testdata/overload_values/narrow-valid.ts | 7 | 7 | 17 | 23 | 4 | 0 |
| internal/oracle/testdata/overload_values/order-valid.ts | 6 | 6 | 12 | 17 | 5 | 0 |
| internal/oracle/testdata/overload_values/module-valid.ts | 4 | 4 | 8 | 14 | 3 | 0 |
| internal/oracle/testdata/overload_visitors/instantiations.ts | 11 | 11 | 11 | 21 | 9 | 0 |
| internal/oracle/testdata/overload_visitors/original.a | 5 | 5 | 11 | 17 | 4 | 0 |
| internal/oracle/testdata/overload_visitors/original.ts | 5 | 5 | 11 | 17 | 4 | 0 |
| internal/oracle/testdata/overload_visitors/helper.a | 5 | 5 | 11 | 17 | 4 | 0 |
| internal/oracle/testdata/overload_visitors/helper.ts | 5 | 5 | 11 | 17 | 4 | 0 |
| internal/oracle/testdata/overload_visitors/overloaded-helper.a | 5 | 5 | 11 | 17 | 4 | 0 |
| internal/oracle/testdata/overload_visitors/overloaded-helper.ts | 5 | 5 | 11 | 17 | 4 | 0 |
```

Commands and observed results

All test output went directly to a file; no test process was piped. Each JSON below records the exact arguments, exit, wall seconds and every test's seconds. Test times include their in-test setup; command times additionally include Go compilation. Source Node, JavaScript backend Node and native sanitizer/leak comparisons are exercised by the oracle and attribution checks.

| Run | Exit | Command seconds |
|---|---:|---:|
| call-target-export-warm | 0 | 3.975 |
| cmd-adamic | 0 | 17.708 |
| counts-audit-current | 0 | 25.349 |
| counts-audit-main-abi | 0 | 35.562 |
| counts-audit-wrong-abi-mutant | 1 | 9.82 |
| counts-missing-node-types | 1 | 136.449 |
| counts-regeneration | 0 | 76.783 |
| fixtures | 0 | 29.554 |
| getters-observations | 0 | 14.347 |
| initial-cold-build | 124 | 180.006 |
| internal-flow | 0 | 23.762 |
| internal-ir | 0 | 56.779 |
| internal-lower | 0 | 8.419 |
| internal-oracle | 0 | 22.554 |
| vet | 0 | 47.308 |

Every per-test measurement is in [test-seconds.csv](test-seconds.csv), and exact commands and outcomes are in [commands-results.json](commands-results.json). These include all recorder shards and fixture leaves.

Mutants and their actual catchers

| Mutant | Exit | Failing owning tests |
|---|---:|---|
| counts-audit-wrong-abi-mutant | 1 | TestAuditClosureCountsUserIteratorsRestTdz, TestAuditClosureCountsHostNeverBranches, TestAuditClosureCountsCensusOptionalValuesSystem, TestAuditClosureCountsCensusUnaryNumeric, TestAuditClosureCountsUserIterators, TestAuditClosureCountsHostVoidMethod, TestAuditClosureCountsClosureConventionReceiverRest, TestAuditClosureCountsCensusOptionalValuesPadding, TestAuditClosureCountsCensusOptionalValuesForms |
| mutant-admission-drop-fixed-return-proof | 1 | TestOverloadStructuralResults |
| mutant-admission-drop-relation-diagnostic | 1 | TestOverloadStructuralRefusesFactoryResultCovariance |
| mutant-admission-drop-return-storage | 1 | TestOverloadStructuralRefusesFactoryFieldStorage |
| mutant-admission-reverse-result-covariance | 1 | TestOverloadStructuralRefusesFactoryLiteralCovariance |
| mutant-admission-trust-visitor-input | 1 | TestOverloadStructuralRefusesVisitorDomainLiar |
| mutant-admission-trust-writable-result | 1 | TestOverloadStructuralRefusesFactoryWritableInvariance |
| mutant-admit-refused-getters | 1 | TestFlowGettersCensusRefusalClassification |
| mutant-binder-result-drop-check | 1 | TestNativeAgreesWithNode_045/internal/oracle/testdata/census_overload_binder_result_checked.a, TestNativeAgreesWithNode_045 |
| mutant-callback-field-storage | 1 | TestOverloadCallbackFieldStorage |
| mutant-callback-input | 1 | TestOverloadCallbackUnservedInput |
| mutant-callback-representation | 1 | TestOverloadCallbackUnservedRepresentation |
| mutant-callback-result | 1 | TestOverloadCallbackUnservedResult |
| mutant-default | 1 | TestNativeAgreesWithNode_000/internal/oracle/testdata/hidden_boundary_generic_tnode.a, TestNativeAgreesWithNode_000 |
| mutant-field-hatch-kind | 1 | TestOverloadFieldHatchKindLiar |
| mutant-field-hatch-value | 1 | TestOverloadFieldHatchValueLiar |
| mutant-fields-array-layout | 1 | TestHiddenTNodeConstraintArrayLayout |
| mutant-fields-drop-constraint-storage | 1 | TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintMutationRemainsNotYet, TestHiddenTNodeConstraintString |
| mutant-fields-drop-field-storage | 1 | TestOverloadStructuralRefusesMixedFieldStorage |
| mutant-fields-ignore-mapper | 1 | TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown |
| mutant-fields-ignore-substitution | 1 | TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintUnknown |
| mutant-fields-object-layout | 1 | TestHiddenTNodeConstraintMutationRemainsNotYet |
| mutant-fields-unconstrained-layout | 1 | TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown |
| mutant-generic-value | 1 | TestHiddenTNodeGenericValueRemainsNotYet |
| mutant-invoke-code-getter | 1 | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-invoke-field-getters | 1 | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-never-storage | 1 | TestNativeAgreesWithNode_062/internal/oracle/testdata/hidden_boundary_never_array.a, TestNativeAgreesWithNode_062 |
| mutant-never-writable | 1 | TestHiddenNeverArrayRejectsWritableWidening |
| mutant-omit-data-mutation | 1 | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-results-drop-parameter-check | 1 | TestOverloadResultsLiarStopsParameterLiar |
| mutant-results-drop-specialization | 1 | TestNativeAgreesWithNode_105/internal/oracle/testdata/overload_results_binding.a, TestNativeAgreesWithNode_105 |
| mutant-results-trust-optional-property | 1 | TestOverloadResultsRefusesOptionalResultProperty |
| mutant-results-trust-result | 1 | TestOverloadResultsLiarStopsEvaluateLiar, TestOverloadResultsLiarStopsLiar |
| mutant-results-trust-shared-result | 1 | TestOverloadResultsRefusesSharedReadonlyResult, TestOverloadResultsRefusesSharedWritableResult |
| mutant-step3-receiver-proof-mutant | 1 | TestMethodDestructuringSafetyStructuralReceiver, TestMethodDestructuringSafetyDefaultReceiver, TestMethodDestructuringSafetyReceiver, TestMethodDestructuringSafetyExplicitReceiver, TestMethodDestructuringSafetyArrowReceiver |
| mutant-structural-drop-proof | 1 | TestOverloadStructuralResults |
| mutant-structural-mutable-variance | 1 | TestOverloadStructuralRefusesWritableVariance |
| mutant-structural-reverse-covariance | 1 | TestOverloadStructuralRefusesReverseReadonlyVariance |
| mutant-structural-trust-shape | 1 | TestOverloadStructuralRefusesMissingBlockField |
| mutant-values-drop-indirect-check | 1 | TestOverloadValuesReturnedLiar |
| mutant-values-skip-single-signature-check | 1 | TestOverloadValuesNarrowLiar |
| mutant-visitors-admit-unproven-invocation | 1 | TestOverloadVisitorsOverloadedHelperLiar |
| mutant-visitors-erase-TIn-to-Node | 1 | TestOverloadVisitorsHelperLiar |

TestHiddenBoundary04PresentArrayMutant catches returning undefined for present input through Node stdout (0|0 versus 7|0). TestHiddenBoundary04EagerIndexMutant catches eager index evaluation through Node side-effect counts. Both test leaves pass their mutant-detection assertions. The four harness mutant logs additionally require the named getter invocation, lost nested data mutation, or refusal admission marker; compile errors and timeouts do not count as catches. Compiler overlays are freshly generated from the resolved tree using the source patches, not stale full-source overlays.

Limits

Only focused tests and fixtures were run, plus the required all-row count recorder and changed-package vet. No whole-package test command or full gate ran. Historical full-census, replay and performance claims in inherited evidence are provenance, not current validation. The TDZ attribution witness preserves empty stdout and an explicit refusal; it does not claim equal raw Node and Adamic exit codes or error text.
