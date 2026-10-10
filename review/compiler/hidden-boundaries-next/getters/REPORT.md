Merged current main and repaired getters flow coverage toward step 30.
Main: 55347dce1537c0bd05d35eab0dc5d663c4823803; merge: 81dc9523ed13d949bb1157a4a96e980e9cf17d1d; branch: compiler/hidden-boundaries-next.
All 21 getters observations pass; three runnable shapes agree in all four lanes with leak checks; counts regenerated once with zero row changes.
All 40 existing mutant catchers and four new flow-harness mutants have their expected outcomes.
Full gate and census not run; no getters fixture body, oracle registration, or refusal expectation changed.

## Findings

The previous report's wording was imprecise. The getters observation tests already pass on this branch. The failures were in TestFlowCorpusRemainder, and untouched main 55347dce reproduces them. main-oracle.log passes in 1.835 s; main-flow.log fails in 0.850 s, admitting deliberate refusals and corrupting the throwing getter trace.

The corpus glob admitted all eighteen deliberately non-lowering getters witnesses, although each oracle observation correctly required its stop. The flow-only eligibility table now excludes exactly those eighteen shapes. TestFlowGettersCensusRefusalClassification verifies the actual compiler still stops at every excluded fixture's path and named reason, so a future lowering gain fails the classification check instead of being silently excluded. The three runnable shapes must remain in the corpus.

The trace snapshot walked value[key] and read value.code. That executes getters outside source evaluation, causing extra behavior, recursive tracing and throws. In getters_census_throwing_object.a, the extra getter execution left the tracer in another function's frame; liveness then indexed instruction 5 in a graph with one instruction. Snapshots now inspect own property descriptors, recursively print data properties and mark accessor properties without calling them. The regression witness proves getters stay uncalled while nested data mutations are still observed. Compiler sources are unchanged.

No newly lowered getter was found. No real compilation gain is claimed. The three already runnable shapes remain getters_census_class_map.a, getters_census_binary_inline.a and getters_census_throwing_object.a. Each compares source Node, emitted JavaScript on Node, release native and ASan/UBSan native, with leak checks; final-oracle.log records all four outputs.

## Each getter shape

No fixture or its oracle expectations was edited. The following eighteen retain their stops and leave only the runnable flow corpus. Their file, line, column and message are byte-for-byte equal between untouched main and the merged branch after normalizing the checkout root.

| Fixture | Unchanged location and diagnostic |
|---|---|
| internal/oracle/testdata/getters_census_binary_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_callback_pair_methods.a | 11:9: stage 0 can't lower a non-accessor method in an accessor literal yet |
| internal/oracle/testdata/getters_census_closure_number_methods.a | 5:14: stage 0 can't lower a non-accessor method in an accessor literal yet |
| internal/oracle/testdata/getters_census_closure_number_next.a | 5:41: stage 0 can't lower a non-accessor method in an accessor literal yet |
| internal/oracle/testdata/getters_census_closure_union_next.a | 5:45: stage 0 can't lower a non-accessor method in an accessor literal yet |
| internal/oracle/testdata/getters_census_higher_callback_pair_methods.a | 11:9: stage 0 can't lower a non-accessor method in an accessor literal yet |
| internal/oracle/testdata/getters_census_lazy_object.a | 9:57: stage 0 can't lower reading factory yet |
| internal/oracle/testdata/getters_census_optional_boolean_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_optional_comment_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_optional_type_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_overloaded_function.a | 18:46: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| internal/oracle/testdata/getters_census_primary_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_snapshot_next.a | 7:56: stage 0 can't lower a non-accessor method in an accessor literal yet |
| internal/oracle/testdata/getters_census_unary_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_unary_node_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_update_comment_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_update_node_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| internal/oracle/testdata/getters_census_update_type_function.a | 3:24: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |

The ten cycle-capable refusals still identify the callback path and offer removal of the captured back-reference, a noncapturing module function, or Weak with a strong owner. The unchecked-cast refusal still identifies 18:46 and offers a proven upcast, tagged discriminated-union cast, or nominal ancestry downcast. The seven NotYet messages retain their named paths and missing operation; they did not contain a suggested fix on main and none was removed or changed here. There is no from/to diagnostic change.

| Existing runnable fixture | Reason it remains agreement |
|---|---|
| internal/oracle/testdata/getters_census_class_map.a | class accessor, interface and union reads agree in all four lanes |
| internal/oracle/testdata/getters_census_binary_inline.a | inline closure memoization agrees, with initialization and result order preserved |
| internal/oracle/testdata/getters_census_throwing_object.a | getter throw and catch agree; descriptor snapshots now preserve source evaluation |

## Changed paths and checks

The main merge resolves stage3/census/latent/README.md once by retaining both hidden caller/interval replay and speculative continuation sections. Its other paths are automatic main merges.

The harness fix changes internal/flow/flow_test.go (explicit getter refusal eligibility), internal/flow/trace_test.go (descriptor snapshots), and adds internal/flow/getters_census_test.go (classification, three runnable traces, side-effect-free snapshot and nested mutation checks). No internal/oracle/getters_census_test.go edit was warranted. No compiler/backend source was edited. paths-from-main.txt lists every delivery path relative to main, including inherited hidden-boundaries work and this review evidence.

Five new top-level parallel leaves, measured on this 4-CPU cgroup (nproc 5):

| Test | Seconds |
|---|---:|
| TestFlowGettersCensusBinaryInline | 0.18 |
| TestFlowGettersCensusClassMap | 0.18 |
| TestFlowGettersCensusThrowingObject | 0.18 |
| TestFlowTraceSnapshotsDoNotInvokeGetters | 0.04 |
| TestFlowGettersCensusRefusalClassification | 0.65 |

flow-final.log also reruns the entire remainder and the existing selectable accessor/mutation witnesses. It passes in 2.978 s. The targeted oracle passes in 2.832 s; TestCallTargetReaders passes in 22.820 s. go build ./... and go vet ./internal/flow ./internal/oracle return 0 with empty logs.

## Mutants

compiler-mutants.json records every command, exit and failing catcher for the 38 existing compiler overlays; patches/ contains their exact zero-context mutations (apply to a scratch source with git apply --unidiff-zero, then map it with a Go overlay). All 38 exit 1 with test assertion failures. This includes the visitor-input mutant caught by the missing callback-contravariance diagnostic while a later invocation guard still refuses the program. That expected diagnostic catcher is recorded explicitly rather than the old stale got-nil marker.

TestHiddenBoundary04PresentArrayMutant and TestHiddenBoundary04EagerIndexMutant pass their mutant-detection assertions in final-oracle.log, completing the original forty. They respectively compare undefined-for-present output and eager-index side effects with source Node.

| New harness mutant | Catcher and observation |
|---|---|
| invoke-field-getters | TestFlowTraceSnapshotsDoNotInvokeGetters: snapshot invoked value getter |
| invoke-code-getter | TestFlowTraceSnapshotsDoNotInvokeGetters: snapshot invoked code getter |
| omit-data-mutation | TestFlowTraceSnapshotsDoNotInvokeGetters: snapshot lost nested data mutation |
| admit-refused-getters | TestFlowGettersCensusRefusalClassification: eighteen deliberate refusals entered the runnable corpus |

All four final runs exit 1 on their owning assertion. Their source overlays are .go.txt, never compilable Go evidence. The first field-getter runner expected a value-getter marker but the code getter appeared first in its witness. Reordering the two getter declarations made that witness independent of the code-getter guard; harness-mutants-final.json records the four corrected reruns. No production source was changed for any mutant, and no timeout or build-warning failure counts as a catch.

## Commands

All tests write to logs, use -count=1 and -timeout 90s unless the count recorder's specified -timeout 180s is shown. Every long process also carries an outer timeout. All verification shells source /workspace/adamic-tools/env.sh and use GOMAXPROCS=4; oracle comparisons and compiler mutants use ADAMIC_GATE_UNCACHED=1.

```text
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
git fetch origin
git checkout compiler/hidden-boundaries-next
git merge origin/main
git add stage3/census/latent/README.md
git commit -m 'Merge current main and preserve both latent census README sections'

# Identical command on untouched main and the merged branch:
timeout 180 go test ./internal/oracle -run '^TestGettersCensus|^TestNativeAgreesWithNode$/internal/oracle/testdata/getters_census_' -count=1 -timeout 90s -v
# Untouched main, reproduces the prior harness failure:
timeout 180 go test ./internal/flow -run '^TestFlowCorpusRemainder$/^\.\.$/^oracle$/^testdata$/^getters_census_' -count=1 -timeout 90s -v

# One count regeneration: 116.499 s, zero row changes:
timeout 240 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 180s -args -update-counts
# Original forty: per-overlay commands in compiler-mutants.json and these two semantic mutants:
timeout 180 go test ./internal/oracle -run '^TestGettersCensus|^TestHiddenBoundary04(PresentArrayMutant|EagerIndexMutant)$|^TestNativeAgreesWithNode$/internal/oracle/testdata/getters_census_' -count=1 -timeout 90s -v
# Four new harness overlays: exact commands in harness-mutants-final.json

gofmt -w internal/flow/getters_census_test.go internal/flow/flow_test.go internal/flow/trace_test.go
timeout 180 go test ./internal/flow -run '^TestFlowGetters|^TestFlowTraceSnapshots|^TestFlowCorpusRemainder$|^TestFlowCorpusHiddenBoundaryRefusalClassification$|^TestFlowProgram.*(accessor|mutations)' -count=1 -timeout 90s -v
timeout 180 go build ./...
timeout 180 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
timeout 180 go vet ./internal/flow ./internal/oracle
git diff --check

# After committing, before the single push:
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
git push origin compiler/hidden-boundaries-next
```

## Setup

nproc is 5; cgroup cpu.max is 400000 100000 (four CPUs). Setup succeeded. Timing lines:

```text
setup: go ready (0.025s)
setup: node ready (0.027s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.011s
setup: markdown dependencies ready (0.090s)
setup: submodules ready (0.126s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.252s)
setup: shared cache ready (1.201s)
setup: go build ready (18.591s)
setup: test binaries deferred (use --warm-tests) (18.715s)
setup: build cache warm (18.717s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (18.751s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.P1wpPl
```
