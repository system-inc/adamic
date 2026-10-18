Built hidden-boundaries plus 15 required a-check headers, with every program body preserved.
Base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 6685711ba12b34e2e10b3c1fd08af31fa015b91e; branch compiler/loom-r-hidden-boundaries-ac.
All 60 real compiler checks and the unchanged pinned header predicate pass; focused readers, IR guard and vet pass.
Removed-header and wrong-reason mutants each fail the unchanged a-check predicate using real compiler diagnostics.
No full gate or ahra run; counts and runtime C are byte-identical to the source commit.

## Scope and assumptions

Created the branch at the exact requested base after fetching origin, then applied source commit 6685711 with git cherry-pick --no-commit. No conflict. The original compiler change and its review evidence are retained. This pass changes only 15 first-line headers outside this evidence directory. Existing matching headers are retained byte for byte, using the predicate's reason-substring contract. All 60 added or modified .a files are under internal; the source commit adds or modifies no review .a files. No new fixture or test leaf was added.

## Toolchain

Ran export GOPROXY='https://proxy.golang.org|direct' and timeout 600 bash cloud/setup.sh, then sourced /workspace/adamic-tools/env.sh. Setup succeeded without a workaround. nproc=5; CPU quota=4; Go 1.27.1; Node v24.19.0; clang 20.1.8. All timings below are cumulative setup seconds:

```
setup: go ready (0.097s)
setup: node ready (0.100s)
setup: submodules ready (0.121s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.024s
setup: markdown dependencies ready (0.283s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.555s)
setup: shared cache ready (1.964s)
setup: go build ready (24.526s)
setup: test binaries deferred (use --warm-tests) (24.726s)
setup: build cache warm (24.728s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.07 0.02 0.00 1/135 711 load-after=1.73 0.39 0.13 2/144 1366
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (24.773s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.6JUG8D
```

Tests ran with GOMAXPROCS=4, ADAMIC_GOCACHE_STORE=off and TZ=UTC. Built this branch's compiler with timeout 180 go build -o /workspace/scratch/hidden-ac-compiler ./cmd/adamic. classify.py invokes that binary as adamic c <path>, with a 45-second per-file timeout and four workers. statuses.json records full diagnostics, status, exact header and unchanged body SHA-256. Compiler SHA-256: f0bbd38938b5325041d6c387793fe5e4b101f3e95a2f15d5950f9b3f37eec56a.

## Per-file statuses

25 checked, 2 NotYet, 33 refused, 0 type errors. Eighteen refused files already had matching headers; fifteen gained headers. NotYet reasons are retained below and in statuses.json; no NotYet header is added.

| File | Status | Header change | Reason |
| --- | --- | --- | --- |
| internal/lower/testdata/method_destructuring_receiver.a | refused | added | a method in object destructuring; call it on its receiver or wrap that call in an arrow; destructuring would lose this |
| internal/lower/testdata/overload-results/evaluate-liar.a | checked | none |  |
| internal/lower/testdata/overload-results/liar.a | checked | none |  |
| internal/lower/testdata/overload-results/parameter-liar.a | checked | none |  |
| internal/lower/testdata/overload-structural/block-factory-liar.a | refused | added | overload 1 of transform result Block cannot be served by implementation result Body at result.kind; readonly covariance requires string to fit "Block"; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/block-liar.a | refused | added | overload 1 of transform result Block cannot be served by implementation result ConciseBody at result.statements; required field is absent from Expression; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/evaluator-reverse-liar.a | refused | added | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/evaluator-shared-liar.a | refused | added | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/evaluator-writable.a | refused | added | overload 1 of evaluate parameter input cannot be served by implementation parameter input; make the implementation accept every value admitted by this overload, without mutable widening or bivariance |
| internal/lower/testdata/overload-structural/factory-liar.a | refused | added | overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/factory-literal-liar.a | refused | added | overload 1 of evaluate result Result<"safe"> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit "safe"; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/factory-storage.a | refused | added | overload 1 of evaluate result Result<string &#124; number> cannot be served by implementation result Other &#124; Result<string &#124; number> at result.value; required field is absent from Other; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/factory-writable.a | refused | added | overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string> at result.value; writable invariance (reverse direction) requires string &#124; undefined to fit string; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/fields-mixed.a | refused | added | an overload result field at result.value without a single storage representation; narrow the result to a member before reading the field; matching storage does not replace the overload return proof |
| internal/lower/testdata/overload-structural/visitor-domain-liar.a | refused | added | overload 1 of visitNodes visitor argument path is unproven at /workspace/adamic/internal/lower/testdata/overload-structural/visitor-domain-liar.a:6:2: original-array input must fit TIn and exclude undefined; callback contravariance at callback.parameter1 requires Node to fit TIn; pass a present original input element at every visitor invocation, including helpers |
| internal/lower/testdata/overload-structural/visitor-domain.a | refused | added | overload 1 of visitNodes visitor argument path is unproven at /workspace/adamic/internal/lower/testdata/overload-structural/visitor-domain.a:6:30: original-array input must fit TIn and exclude undefined; callback contravariance at callback.parameter1 requires Node to fit TIn; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/census_overload_binder_result_checked.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_generic_optional_array.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_generic_tnode.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_never_array.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_never_array_observations.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_optional_array_order.a | checked | none |  |
| internal/oracle/testdata/hidden_boundary_refused/hidden_boundary_alias_cast_refused.a | refused | added | a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| internal/oracle/testdata/hidden_boundary_refused/hidden_boundary_optional_array_mutation_refused.a | refused | added | instantiating a generic function makes a value of type 7 written where string is read; make the collection readonly, or use a type parameter for the value being written (adamic/invariant-mutable) |
| internal/oracle/testdata/hidden_wave2_method_destructuring.a | checked | none |  |
| internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_mutation.a | NotYet | none | adamic: /workspace/adamic/internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_mutation.a:3:79: stage 0 can't lower assigning a field of a union of differently held members yet |
| internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_value.a | NotYet | none | adamic: /workspace/adamic/internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_value.a:3:36: stage 0 can't lower a generic function as a value yet |
| internal/oracle/testdata/overload_callback_served.a | checked | none |  |
| internal/oracle/testdata/overload_field_hatch/kind-liar.a | refused | retained | overload 1 of transform result Block cannot be served by implementation result ConciseBody at result.kind; readonly covariance requires SyntaxKind to fit SyntaxKind.Block; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_field_hatch/kind-valid.a | refused | retained | overload 1 of transform result Block cannot be served by implementation result ConciseBody at result.kind; readonly covariance requires SyntaxKind to fit SyntaxKind.Block; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_field_hatch/value-liar.a | refused | retained | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_field_hatch/value-undefined.a | refused | retained | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_field_hatch/value-valid.a | refused | retained | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_results_binding.a | checked | none |  |
| internal/oracle/testdata/overload_results_evaluate.a | checked | none |  |
| internal/oracle/testdata/overload_results_parameter.a | checked | none |  |
| internal/oracle/testdata/overload_results_scalar.a | checked | none |  |
| internal/oracle/testdata/overload_results_transform.a | checked | none |  |
| internal/oracle/testdata/overload_structural_block.a | checked | none |  |
| internal/oracle/testdata/overload_structural_evaluator.a | checked | none |  |
| internal/oracle/testdata/overload_structural_factory.a | checked | none |  |
| internal/oracle/testdata/overload_structural_fields.a | checked | none |  |
| internal/oracle/testdata/overload_values/module-liar.a | refused | retained | overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_values/module-valid.a | refused | retained | overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/oracle/testdata/overload_values/narrow-liar.a | refused | retained | an indirect value of an overload requires a per-overload proof: overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; prove every return for each exposed overload |
| internal/oracle/testdata/overload_values/narrow-valid.a | refused | retained | an indirect value of an overload requires a per-overload proof: overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; prove every return for each exposed overload |
| internal/oracle/testdata/overload_values/order-liar.a | refused | retained | an indirect value of an overload requires a per-overload proof: overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; prove every return for each exposed overload |
| internal/oracle/testdata/overload_values/order-valid.a | refused | retained | an indirect value of an overload requires a per-overload proof: overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; prove every return for each exposed overload |
| internal/oracle/testdata/overload_values/proven.a | checked | none |  |
| internal/oracle/testdata/overload_values/returned-liar.a | refused | retained | an indirect value of an overload requires a per-overload proof: overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; prove every return for each exposed overload |
| internal/oracle/testdata/overload_values/returned-valid.a | refused | retained | an indirect value of an overload requires a per-overload proof: overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; prove every return for each exposed overload |
| internal/oracle/testdata/overload_visitors/default-liar.a | refused | retained | overload 1 of visit visitor argument path is unproven at /workspace/adamic/internal/oracle/testdata/overload_visitors/default-liar.a:9:54: original-array input must fit T and exclude undefined; callback contravariance at callback.parameter1 requires Base to fit T; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/overload_visitors/helper-liar.a | refused | retained | overload 1 of visit visitor argument path is unproven at /workspace/adamic/internal/oracle/testdata/overload_visitors/helper-liar.a:8:2: original-array input must fit T and exclude undefined; callback contravariance at callback.parameter1 requires Base to fit T; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/overload_visitors/helper.a | checked | none |  |
| internal/oracle/testdata/overload_visitors/instantiations.a | refused | retained | overload 1 of visit visitor argument path is unproven at /workspace/adamic/internal/oracle/testdata/overload_visitors/instantiations.a:7:71: original-array input must fit T and exclude undefined; callback contravariance at callback.parameter1 requires Base to fit T; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/overload_visitors/original-liar.a | refused | retained | overload 1 of visit visitor argument path is unproven at /workspace/adamic/internal/oracle/testdata/overload_visitors/original-liar.a:9:2: original-array input must fit T and exclude undefined; callback contravariance at callback.parameter1 requires Base to fit T; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/overload_visitors/original.a | checked | none |  |
| internal/oracle/testdata/overload_visitors/overloaded-helper-liar.a | refused | retained | overload 1 of worker visitor argument path is unproven at /workspace/adamic/internal/oracle/testdata/overload_visitors/overloaded-helper-liar.a:9:2: original-array input must fit T and exclude undefined; callback contravariance at callback.parameter1 requires Base to fit T; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/overload_visitors/overloaded-helper.a | checked | none |  |

## Exact headers added

Each file below gained only this first line. All bytes below it equal source 6685711.

| File | First line |
| --- | --- |
| internal/lower/testdata/method_destructuring_receiver.a | // a-check: refused a method in object destructuring; call it on its receiver or wrap that call in an arrow; destructuring would lose this |
| internal/lower/testdata/overload-structural/block-factory-liar.a | // a-check: refused overload 1 of transform result Block cannot be served by implementation result Body at result.kind; readonly covariance requires string to fit "Block"; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/block-liar.a | // a-check: refused overload 1 of transform result Block cannot be served by implementation result ConciseBody at result.statements; required field is absent from Expression; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/evaluator-reverse-liar.a | // a-check: refused overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/evaluator-shared-liar.a | // a-check: refused overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/evaluator-writable.a | // a-check: refused overload 1 of evaluate parameter input cannot be served by implementation parameter input; make the implementation accept every value admitted by this overload, without mutable widening or bivariance |
| internal/lower/testdata/overload-structural/factory-liar.a | // a-check: refused overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit string &#124; undefined; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/factory-literal-liar.a | // a-check: refused overload 1 of evaluate result Result<"safe"> cannot be served by implementation result Result<string &#124; number &#124; undefined> at result.value; readonly covariance requires string &#124; number &#124; undefined to fit "safe"; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/factory-storage.a | // a-check: refused overload 1 of evaluate result Result<string &#124; number> cannot be served by implementation result Other &#124; Result<string &#124; number> at result.value; required field is absent from Other; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/factory-writable.a | // a-check: refused overload 1 of evaluate result Result<string &#124; undefined> cannot be served by implementation result Result<string> at result.value; writable invariance (reverse direction) requires string &#124; undefined to fit string; prove every return for this overload's admitted arguments, preserving result variance |
| internal/lower/testdata/overload-structural/fields-mixed.a | // a-check: refused an overload result field at result.value without a single storage representation; narrow the result to a member before reading the field; matching storage does not replace the overload return proof |
| internal/lower/testdata/overload-structural/visitor-domain-liar.a | // a-check: refused overload 1 of visitNodes visitor argument path is unproven at /workspace/adamic/internal/lower/testdata/overload-structural/visitor-domain-liar.a:6:2: original-array input must fit TIn and exclude undefined; callback contravariance at callback.parameter1 requires Node to fit TIn; pass a present original input element at every visitor invocation, including helpers |
| internal/lower/testdata/overload-structural/visitor-domain.a | // a-check: refused overload 1 of visitNodes visitor argument path is unproven at /workspace/adamic/internal/lower/testdata/overload-structural/visitor-domain.a:6:30: original-array input must fit TIn and exclude undefined; callback contravariance at callback.parameter1 requires Node to fit TIn; pass a present original input element at every visitor invocation, including helpers |
| internal/oracle/testdata/hidden_boundary_refused/hidden_boundary_alias_cast_refused.a | // a-check: refused a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| internal/oracle/testdata/hidden_boundary_refused/hidden_boundary_optional_array_mutation_refused.a | // a-check: refused instantiating a generic function makes a value of type 7 written where string is read; make the collection readonly, or use a type parameter for the value being written (adamic/invariant-mutable) |

## Predicate and mutants

Fetched the README's pinned fbac28c62493f27a788edc02a18bc8edb68de5da object explicitly from origin. gate.py is the unmodified cloud/fast-gate/run.py at that commit. check_headers.py invokes its unchanged Gate.aCheck method. The harness reuses the branch compiler already built above; all classification, header reading and comparison execute unchanged. No result replay is used: baseline and both mutants invoke the real compiler.

Command: timeout 180 python3 review/compiler/loom-r-hidden-boundaries-ac/check_headers.py. Result: all 60 accepted. For internal/lower/testdata/method_destructuring_receiver.a, removing its header fails as refused versus expected checked; replacing its reason with unrelated deliberate mutant reason fails as wrong expected refusal. Both exit 1 in the predicate. The original header is restored in finally. See header-check.log, a-check.json and header-mutants.json.

## File readers and positional expectations

path-references.json records exact filename, full path and specific directory matches in *_test.go and testdata. The five reader files are listed in reader-files.json. None pins a diagnostic line shifted by these 15 inserted headers, so no expectation or test source needed updating. Refusal checks match reason text, Node checks match output, and flow classification uses paths. All top-level tests in those five reader files were rerun, including hidden-boundary Node/backend comparisons and their embedded mutants.

## Validation commands and results

- `timeout 150 go test ./internal/flow -run ^(TestFlowCorpusHiddenBoundaryRefusalClassification|TestFlowCorpusRemainder|TestFlowCorpusSetupIsShared|TestFlowCorpusUnitsCoverEveryProgram)$ -count=1 -timeout 90s -parallel 4 -json`: exit 0, 9.76 seconds wall time.
- `timeout 150 go test ./internal/lower -run ^(TestMethodDestructuringSafetyArrowReceiver|TestMethodDestructuringSafetyClassOrigin|TestMethodDestructuringSafetyDefaultReceiver|TestMethodDestructuringSafetyExplicitReceiver|TestMethodDestructuringSafetyReceiver|TestMethodDestructuringSafetyStructuralClass|TestMethodDestructuringSafetyStructuralReceiver|TestOverloadStructuralRefusesBlockFactoryCovariance|TestOverloadStructuralRefusesFactoryFieldStorage|TestOverloadStructuralRefusesFactoryLiteralCovariance|TestOverloadStructuralRefusesFactoryResultCovariance|TestOverloadStructuralRefusesFactoryWritableInvariance|TestOverloadStructuralRefusesMissingBlockField|TestOverloadStructuralRefusesMixedFieldStorage|TestOverloadStructuralRefusesReverseReadonlyVariance|TestOverloadStructuralRefusesSharedWiderResult|TestOverloadStructuralRefusesVisitorDomainLiar|TestOverloadStructuralRefusesVisitorInputContravariance|TestOverloadStructuralRefusesWritableVariance|TestOverloadStructuralResults|TestReceiverIndependentMethodDestructuring)$ -count=1 -timeout 90s -parallel 4 -json`: exit 0, 3.65 seconds wall time.
- `timeout 150 go test ./internal/oracle -run ^(TestHiddenBoundary04AliasCastRefused|TestHiddenBoundary04EagerIndexMutant|TestHiddenBoundary04EvaluationOrder|TestHiddenBoundary04MutationRefused|TestHiddenBoundary04Node|TestHiddenBoundary04PresentArrayMutant|TestOverloadStructuralFieldRefusal)$ -count=1 -timeout 90s -parallel 4 -json`: exit 0, 7.363 seconds wall time.
- `timeout 150 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout 90s -json`: exit 0, 5.813 seconds wall time.
- `timeout 150 go vet ./cmd/adamic ./internal/flow ./internal/ir ./internal/lower ./internal/oracle ./internal/native/runtimeproducts`: exit 0, 1.733 seconds wall time.
- `timeout 30 git diff --check`: exit 0, 0.171 seconds wall time.

Test events: 66 passes; no failures or skips. Longest reported test: TestFlowCorpusRemainder 6.56 seconds. No leaf reached 60 seconds. tests.json retains every test name and duration. All command output is in adjacent log files.

Counts were not regenerated: git diff 6685711 -- internal/oracle/counts.md is empty. The source commit's Linux clang 20.1.8 table is preserved. git diff 6685711 -- internal/native/runtime is empty. Runtime witness host_void_method.a is preserved.

The compiler and runtime mutants from the original rebase remain documented in review/compiler/loom-r-hidden-boundaries/REPORT.md; this header pass reruns the two required header mutants and affected reader tests. It does not rerun the entire compiler mutant suite, full packages, full gate or ahra.

The staged whitespace check reports existing whitespace in 61 inherited review artifacts, all byte-identical to 6685711. Historical patch/log/CSV bytes were preserved. The product-source staged check and the entire incremental diff against 6685711 pass git diff --check.
