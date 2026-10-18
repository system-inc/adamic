Built the hidden-boundary source squash with main's runtime products and Linux-regenerated counts.
One commit on base d79302e50c022c997c00f7c68ddb7475f9ee65fc, branch compiler/loom-hidden-boundaries-3; git supplies its delivery SHA.
Focused checks, fixture readers, registration guards and vet pass; counts recorder PASS 255.904s.
38 compiler plus 4 flow-harness mutants caught, plus ABI and runtime-witness mutants; two in-test IR mutants also caught.
Not covered: full packages/full gate, full census, historical replay/meter tools, WASI or performance.

Source: b4ff4677947b81df13793e0dd949bbee2ed7d724. Historical source review artifacts, census templates and paired .ts hatch
controls are retained as provenance, not presented as fresh observations. No cohere code was
copied and its gitlink is unchanged. New evidence is confined to this directory.

Conflict resolution and conservative assumptions

Main moved runtime feature sets from internal/oracle/runtimes_test.go into
internal/native/runtimeproducts/runtimeproducts.go. Kept main's oracle wiring in full and
applied the witness change at its new owner. host_void_method.a is preserved unchanged,
including receiver.write reading this.prefix; it now witnesses exactly ADAMIC_CLOSURE_RECEIVERS.
The existing receivers-plus-node-host set's sole witness, stage3/fixtures/host/13_createDirectory.a,
now enables only ADAMIC_NODE_HOST. The exact-witness guard failed on that stale set before it
was removed. The retained eleven sets pass. The stage3 host reader also passes Node and its
native sanitizer/leak hook after removal.

Took origin/main's counts.md before regeneration. Observed main at d79302e50c022c997c00f7c68ddb7475f9ee65fc has seven table
columns, with digests in keyed measurement products, not a digest column in counts.md.
Conservative choice: retain main's actual recorder/schema rather than invent a column or copy
the parked table. The recorder rewrote the complete table once on Linux, house clang 20.1.8.
There are 1058 runtime fixture rows: 33 new rows, nine existing changes, none removed.
Only retains/releases change in those nine rows. Their positive ABI audits also compare to
source Node, JavaScript and native sanitizer/leak behavior. Full diff: counts.diff.

| Existing fixture | Main retains/releases | Current retains/releases |
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

Commands and outputs

Every command ran from the repository root after sourcing /workspace/adamic-tools/env.sh.
All test output was redirected directly to logs, never piped. GOMAXPROCS=4 and -parallel 4
bound test concurrency on this four-CPU quota. Exact source-verifier and mutant arguments,
exits, command seconds and failing/passing tests are in commands-results.json,
mutants-results.json and test-seconds.csv. The verifier's output is verify.log.

- export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh
  exit 0. Node ready 0.040s, Go ready 0.044s, submodules ready 0.102s, Markdown ready 0.122s,
  clang ready 0.280s, shared cache ready 0.702s, go build ready 73.894s,
  build cache warm 74.055s, done 74.123s. nproc=5, cpu.max=400000 100000.
  Go 1.27.1, Node 24.19.0, clang 20.1.8. Full timing lines: setup.log.
- timeout 600 python3 review/compiler/loom-hidden-boundaries-3/verify.py
  exit 0. Its bounded go test commands use -count=1 -timeout 90s -parallel 4 -json.
  Command totals: cmd/adamic 28.223s; flow 40.727s; reader 35.042s (test leaf 16.93s);
  lower 8.367s; oracle 41.030s; fixture-family selection 44.053s; changed-package vet 9.134s.
  No whole-package test command was run. All selected test events are under 60s.
- ADAMIC_GATE_UNCACHED=1 timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 9m -parallel 4 -json -args -update-counts
  PASS 255.904s. Complete output: counts-regeneration.log.gz. Toolchain: counts-toolchain.log.
- timeout 600 python3 review/compiler/loom-hidden-boundaries-3/mutants.py
  first batch reached its 600s outer limit after 24 completed catches. Resumed with the same
  command and ADAMIC_GOCACHE_STORE=off, skipping recorded catches, to avoid remote Go-cache
  lookups. Resumed exit 0; all 42 source compiler/harness mutants caught. A partial command
  interrupted in compilation is not claimed as a catch. No source Go file was mutated.
- timeout 180 go test -overlay review/compiler/loom-hidden-boundaries-3/counts-audit-overlay.json ./internal/oracle -run '^TestAuditClosureCounts' -count=1 -timeout 90s -parallel 4 -json
  PASS 28.650s; nine explicit existing-row assertions plus Node/backend comparisons.
  Evidence-only historical helper adapted to main's counted API returning a measurement thunk;
  the initial compile failure is retained in counts-audit-current.log and is not a test verdict.
- timeout 180 go test -overlay review/compiler/loom-hidden-boundaries-3/counts-audit-main-abi-overlay.json ./internal/oracle -run '^TestAuditClosureCounts' -count=1 -timeout 90s -parallel 4 -json
  exit 1; all nine count assertions fail when only collections.go, iteration.go and
  iteration_origin.go are restored from actual current main. Non-compilable source snapshots
  and overlay are retained. No warning or unrelated build failure counts as a kill.
- timeout 180 go test ./internal/oracle -run '^TestRuntimeSetsAreEachUsed$' -count=1 -v -timeout 90s -parallel 4
  initial stale combined set rejected in 1.59s; final registry PASS 1.13s.
  Reverting just the receiver-only witness to census_optional_values_system.a through
  runtime-witness-revert-overlay.json fails the guard in 1.54s: it turns on no features.
- timeout 180 go test ./stage3/fixtures -run '^TestFixturesHost$/^host$/^13_createDirectory[.]a$' -count=1 -timeout 90s -parallel 4 -json
  PASS; host parent 47.77s, selected fixture 25.72s, including native/sanitizer/leak hook.
- timeout 180 go test ./stage3/fixtures -run '^TestFixturesCycles$/^cycles$/^(05_safe_load_read|06_import_order_mutant)$/^main[.]a$' -count=1 -timeout 90s -parallel 4 -json
  PASS; parent 5.41s. Covers both source status.json updates.
- timeout 180 go test ./internal/lower -run '^TestNativeAgreementRowsNameRegisteredPrograms$' -count=1 -timeout 90s
  PASS 0.042s.
- timeout 180 go test ./internal/oracle -run '^(TestFamilyWrappersAreCurrent|TestFamiliesOwnEveryCaseOnce|TestCountsAreRecordedTable)$' -count=1 -timeout 90s
  PASS 0.035s; validates registration/table after regeneration, without measuring it again.
- timeout 180 go vet ./internal/native/runtimeproducts
  exit 0, no output; source-verifier vet covered cmd/adamic, flow, ir, lower and oracle.
- git diff --check origin/main -- cmd internal cloud docs
  exit 0, no output. Exact inherited mutant patches remain literal evidence.

Mutants and actual failing checks

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
| mutant-fields-drop-constraint-storage | TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintMutationRemainsNotYet |
| mutant-fields-object-layout | TestHiddenTNodeConstraintMutationRemainsNotYet |
| mutant-fields-array-layout | TestHiddenTNodeConstraintArrayLayout |
| mutant-fields-unconstrained-layout | TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown |
| mutant-fields-ignore-substitution | TestHiddenTNodeConstraintNumber, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintObjectUnion, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintString |
| mutant-fields-ignore-mapper | TestHiddenTNodeConstraintObject, TestHiddenTNodeConstraintUnconstrained, TestHiddenTNodeConstraintBoolean, TestHiddenTNodeConstraintStructuralLength, TestHiddenTNodeConstraintArrayLayout, TestHiddenTNodeConstraintString, TestHiddenTNodeConstraintAny, TestHiddenTNodeConstraintUnknown, TestHiddenTNodeConstraintObjectUnion |
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
| mutant-step3-receiver-proof-mutant | TestMethodDestructuringSafetyReceiver, TestMethodDestructuringSafetyDefaultReceiver, TestMethodDestructuringSafetyStructuralReceiver, TestMethodDestructuringSafetyArrowReceiver, TestMethodDestructuringSafetyExplicitReceiver |
| mutant-invoke-field-getters | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-invoke-code-getter | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-omit-data-mutation | TestFlowTraceSnapshotsDoNotInvokeGetters |
| mutant-admit-refused-getters | TestFlowGettersCensusRefusalClassification |
| receiver runtime witness reverted | TestRuntimeSetsAreEachUsed |
| main receiver ABI restored | all nine TestAuditClosureCounts tests |

TestHiddenBoundary04PresentArrayMutant and TestHiddenBoundary04EagerIndexMutant also pass
in the ordinary focused run: they deliberately change IR and require Node-observable wrong
stdout/side effects from both backends. The four flow-harness mutations require their named
getter-invocation, lost-data-mutation or refused-fixture-admission marker. Compiler mutations
fail selected owning tests; no recorded catch is a build failure, timeout or warning-only kill.

Detailed seconds for every selected source/overlay test event are in test-seconds.csv.
All added/touched test events and the evidence-only ABI audit leaves stayed under 60 seconds.
The unchanged all-row recorder infrastructure is recorded separately as well.
