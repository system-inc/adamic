Re-cut the rejected hidden-boundaries change with an exact receiver-runtime witness.
Source commit 688e1cd84e5fa449714a7837c41a47d2d987c969; main base 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4; the final response supplies the delivery SHA.
Source Node and all three compiled lanes agree; runtime witness/products, focused controls, counts, reader guard and vet passed.
The original 45 mutant detections were rerun, plus one new stale-witness mutant reproducing Loom's failure.
Full packages, full gate, full census, historical replay and performance checks were not run.

Loom change: chg_9qbdqc9eer68bywzjmnrs2q9be. Branch: compiler/loom-hidden-boundaries-2. No PR or Loom submission was made. The earlier focused verification omitted TestRuntimeSetsAreEachUsed, allowing the stale witness to pass worker verification.

Why the old witness stopped needing receivers

census_optional_values_system.a contains System.write, System.readFile and System.exit as literal methods, plus process.stdout.write and process.exit. None reads this or super, including defaults and nested arrows. receiverIndependentMethod proves that syntactically; objectMethod marks their IR functions Receiver=false and creates no this parameter. Their calls then use the ordinary closure ABI. This removes fifteen receiver retain/release pairs, moving the fixture from 67/91 to 52/76 retains/releases while allocations/frees stay 29/29. Native emission finds no receiver closure and emits no ADAMIC_CLOSURE_RECEIVERS define. native.Features is exactly []. This is the intended compiler change, not lost source behavior.

The replacement

host_void_method.a has receiver.write reading this.prefix. Its system.write method is receiver-independent, but receiver.write still needs the receiver ABI. Native emission turns on exactly [ADAMIC_CLOSURE_RECEIVERS], with no convention or canonical-closure feature. Replaced only the receiver-only runtimeSets witness and added its reason as a comment. Neither fixture was edited.

The set is retained. No test or product stops using it. All eleven feature sets remain, and all four receiver-only archives remain owned by TestProduct_NativeRuntimesSanitized0, TestProduct_NativeRuntimesRelease0, TestProduct_NativeRuntimesCounted0 and TestProduct_NativeRuntimesWASI0. All twelve runtime-product test groups passed, building the eleven sets in each of four option modes. The initial native product run skipped WASI before its SDK was available; the subsequent enabled WASI run passed all three groups.

Four-lane observations

| Fixture | Source Node | JavaScript backend Node | Release native | ASan/UBSan native | Features |
|---|---|---|---|---|---|
| census_optional_values_system.a | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | [] |
| host_void_method.a | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | [ADAMIC_CLOSURE_RECEIVERS] |

LeakSanitizer also passes for both. receiver-four-lanes.log records every lane's stdout/stderr/exit explicitly. Source Node stdout, which every compiled lane reproduces byte for byte:

census_optional_values_system.a:
```text
written
exit:undefined
exit:undefined
exit:7
file:utf8
file:ascii
missing
exit:undefined
exit:2
```

host_void_method.a:
```text
hello
receiver: called
argument
last
```

Integration and scope

Fetched current main and squash-merged the previous branch's delivery commit. The squash had no conflicts. Git automatically combined expression.go and counts.md; main's null/undefined template interpolation is preserved. The new authored source change is in internal/oracle/runtimes_test.go only. The delivery retains the previous compiler net change, TypeScript hatch controls and historical evidence. No cohere code was copied or submodule changed. New evidence lives under review/compiler/loom-hidden-boundaries-2; previous evidence is historical provenance, not this run's results.

Production compiler files changed against main:

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

No native runtime C or header source changed. No new package, compiler consumer or product was added, so compiler-dependencies.json needs no change. Products retain main's portable keys; no test/product hard-coded build directory was introduced. Mutant source files are .go.txt and overlays only.

Setup

GOPROXY=https://proxy.golang.org|direct; nproc=5, cpu.max=400000 100000 (four CPUs). All verification shells use GOMAXPROCS=4; ordinary focused tests use -parallel 4, runtime products and the two four-lane witnesses use -parallel 1. Go 1.27.1, Node 24.19.0, native clang 20.1.8 and WASI SDK 27 with clang 20.1.8. Both setup runs succeeded. Timing lines:

```text
setup.log
setup: submodules ready (0.069s)
setup: go ready (0.074s)
setup: node ready (0.086s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.016s
setup: markdown dependencies ready (0.195s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.517s)
setup: shared cache ready (2.455s)
setup: go build ready (69.375s)
setup: test binaries deferred (use --warm-tests) (69.459s)
setup: build cache warm (69.460s)
setup: build-flags commit=87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/137 737 load-after=2.94 0.91 0.32 1/143 1459
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (69.487s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.SdhOAH
setup-wasi.log
setup: node ready (0.023s)
setup: go ready (0.025s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
setup: markdown dependencies ready (0.096s)
setup: submodules ready (0.099s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.255s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk, linking with /workspace/adamic-tools/bin/clang) (26.150s)
setup: shared cache ready (27.970s)
setup: go build ready (75.704s)
setup: test binaries deferred (use --warm-tests) (76.104s)
setup: build cache warm (76.114s)
setup: build-flags commit=87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=1.45 0.90 0.38 6/201 3599 load-after=9.99 3.35 1.26 16/316 9684
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (76.214s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.XZIdoR
```

Counts

One successful TestCountsAreRecorded regeneration ran on this base. It left the squashed counts table byte-identical. Against main the inherited change adds 33 fixture rows and changes nine existing rows; the witness fix moves no row. counts.diff contains every row. Current-ABI and main-ABI checks reproduce all nine sets of counts, and the wrong-ABI mutant fails all nine owning count assertions.

Commands and outputs

Every test wrote directly to a log; none was piped. Every long command has an outer timeout and every correctness Go test has -timeout 90s. Full count/product families use longer bounded process limits to run their already split top-level leaves; every measured leaf stayed below 60 seconds. No test was deleted or split.

| Run | Exit | Command seconds |
|---|---:|---:|
| call-target-export-warm | 0 | 21.154 |
| cmd-adamic | 0 | 56.996 |
| counts-audit-current | 0 | 36.413 |
| counts-audit-main-abi | 0 | 54.903 |
| counts-audit-wrong-abi-mutant | 1 | 16.616 |
| counts-regeneration | 0 | 143.968 |
| fixtures | 0 | 18.084 |
| internal-flow | 0 | 61.318 |
| internal-ir | 0 | 45.909 |
| internal-lower | 0 | 17.001 |
| internal-oracle | 0 | 18.215 |
| receiver-four-lanes | 0 | 34.6 |
| runtime-products-wasi | 0 | 74.344 |
| runtime-products | 0 | 150.208 |
| runtime-witness | 0 | 11.406 |
| stale-runtime-witness-mutant | 1 | 23.384 |
| vet | 0 | 5.438 |

Exact command arrays are in commands-results.json. logs.json preserves every raw log in the commit, since the repository intentionally ignores .log files; the named local logs remain in this review directory. Every leaf's seconds, including products, counts and expected mutant failures, are in test-seconds.csv. Go compilation is included in command seconds; test seconds include in-test setup.

Mutants

The first mutant batch reached its 600-second outer limit after 33 catches. Only unfinished checks were resumed under another hard limit; completed catches were not repeated. Compiler overlays are regenerated from the current merged tree, with original patch positions mapped through main's edits rather than using stale full-source overlays. Each caught mutant must exit 1 with an owning assertion; no build-warning or timeout failure counts. The four flow mutants additionally require their specific failure marker.

| Mutant run | Exit | Actual failing catchers |
|---|---:|---|
| counts-audit-wrong-abi-mutant | 1 | TestAuditClosureCountsHostVoidMethod, TestAuditClosureCountsUserIteratorsRestTdz, TestAuditClosureCountsCensusOptionalValuesSystem, TestAuditClosureCountsHostNeverBranches, TestAuditClosureCountsCensusUnaryNumeric, TestAuditClosureCountsUserIterators, TestAuditClosureCountsClosureConventionReceiverRest, TestAuditClosureCountsCensusOptionalValuesPadding, TestAuditClosureCountsCensusOptionalValuesForms |
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
| mutant-fields-drop-constraint-storage | 1 | TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintMutationRemainsNotYet, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintString |
| mutant-fields-drop-field-storage | 1 | TestOverloadStructuralRefusesMixedFieldStorage |
| mutant-fields-ignore-mapper | 1 | TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintArrayLayout |
| mutant-fields-ignore-substitution | 1 | TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintUnknown |
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
| mutant-results-trust-result | 1 | TestOverloadResultsLiarStopsLiar, TestOverloadResultsLiarStopsEvaluateLiar |
| mutant-results-trust-shared-result | 1 | TestOverloadResultsRefusesSharedReadonlyResult, TestOverloadResultsRefusesSharedWritableResult |
| mutant-step3-receiver-proof-mutant | 1 | TestMethodDestructuringSafetyStructuralReceiver, TestMethodDestructuringSafetyExplicitReceiver, TestMethodDestructuringSafetyReceiver, TestMethodDestructuringSafetyDefaultReceiver, TestMethodDestructuringSafetyArrowReceiver |
| mutant-structural-drop-proof | 1 | TestOverloadStructuralResults |
| mutant-structural-mutable-variance | 1 | TestOverloadStructuralRefusesWritableVariance |
| mutant-structural-reverse-covariance | 1 | TestOverloadStructuralRefusesReverseReadonlyVariance |
| mutant-structural-trust-shape | 1 | TestOverloadStructuralRefusesMissingBlockField |
| mutant-values-drop-indirect-check | 1 | TestOverloadValuesReturnedLiar |
| mutant-values-skip-single-signature-check | 1 | TestOverloadValuesNarrowLiar |
| mutant-visitors-admit-unproven-invocation | 1 | TestOverloadVisitorsOverloadedHelperLiar |
| mutant-visitors-erase-TIn-to-Node | 1 | TestOverloadVisitorsHelperLiar |
| stale-runtime-witness-mutant | 1 | TestRuntimeSetsAreEachUsed |

The two remaining original semantic mutants are in the green oracle controls: TestHiddenBoundary04PresentArrayMutant catches undefined returned for present input through Node stdout (0|0 versus 7|0); TestHiddenBoundary04EagerIndexMutant catches eager index evaluation through Node side-effect counts. Together with the 38 compiler overlays, four flow overlays and one wrong-ABI overlay, these are the original 45 detections. The new stale-runtime-witness overlay restores census_optional_values_system.a as the witness and fails TestRuntimeSetsAreEachUsed with the same "turns on []" message as Loom. It tests the witness check itself while preserving correct compiler and fixture behavior.

Limits

No whole-package test run or full gate ran. Historical census/replay/performance evidence inherited by the squash was not remeasured. WASI runtime archives were built, but WASI fixture execution was not requested or run. The old ABI attribution's TDZ control agrees on empty stdout and an explicit stop; it does not assert equality of raw Node/Adamic error text or exit status. No PR was opened and no Loom submission was made.
