# Step 2 dependency check and Source confirmation

Confirmed the Source representation change on compiler/hidden-boundaries-main f9e3c1bb. No overload commits were applied because their production prerequisites are absent from this main-based branch.

## Why the Source expectation moved

The production change is 702c3ecb507cab7199f7b3e60ee3802262e92eb7, cherry-picked as 46ee597d. In internal/lower/expression.go, lowering.representation now falls back from an unmapped type parameter to its base constraint. An object constraint uses ir.Union to retain the runtime brand instead of assuming an object header. Concrete mappings and substitutions still take precedence. TestRepresentationClockSourceCheckedTypes changed in d80b094e because Source extends CompilerType now has this tagged storage. The test-only correction in cb76d847 describes the same expectation; that commit was not taken.

The checked-types probe contains declarations rather than an executable program. Its companion executable is internal/oracle/testdata/representation_clock_source.a, which reads source.id and source.payload.label through describe<Source extends CompilerType>. It prints exactly `7:source\n` on Node. The oracle agrees byte for byte on stdout, stderr and exit status in JavaScript, ASan/UBSan native and release native, with leak and recorded-count checks. The hidden_boundary_generic_tnode_constraints.a fixture also passes; its holder.node access exercises an unmapped constrained binder, so it covers the constraint-backed path itself.

Disabling only the constraint fallback with the existing no-constraint Go overlay makes the Source probe fail: representation = (0, false), want (10, true). The unmodified product passes with ir.Union (10). This directly ties the expectation change to hidden-06. The existing clock-source-ignore-payload mutant also passes its detection test: both backends finish cleanly but their stdout differs from Node.

| Test | Result | Seconds |
| --- | --- | ---: |
| TestRepresentationClockSourceCheckedTypes | pass | 0.030 |
| TestNativeAgreesWithNode/representation_clock_source.a | Node agrees in both backends and native modes | 0.450 |
| TestNativeAgreesWithNode/hidden_boundary_generic_tnode_constraints.a | Node agrees in both backends and native modes | 0.450 |
| TestRepresentationClockSourceMutant/clock-source-ignore-payload | mutant detected by stdout comparison | 0.340 |
| TestRepresentationClockSourceCheckedTypes with no-constraint overlay | expected failure on missing representation | 0.040 |

Commands (all test output is in adjacent JSON logs):

```sh
ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 timeout 150 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(representation_clock_source|hidden_boundary_generic_tnode_constraints).a$' -count=1 -timeout 90s -json
GOMAXPROCS=4 timeout 150 go test ./internal/lower -run '^TestRepresentationClockSourceCheckedTypes$' -count=1 -timeout 90s -json
ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 timeout 150 go test ./internal/oracle -run '^TestRepresentationClockSourceMutant$' -count=1 -timeout 90s -json
GOMAXPROCS=4 timeout 150 go test -overlay "$PWD/review/compiler/hidden-boundaries-main/no-constraint-overlay.json" ./internal/lower -run '^TestRepresentationClockSourceCheckedTypes$' -count=1 -timeout 90s -json
node --input-type=module-typescript < internal/oracle/testdata/representation_clock_source.a
```

## Missing prerequisites

The brief's 3f656456 does not resolve. The matching listed commit is 3f65645870580e574b909f22b222151fd0dacdd1, Record overload value checks and the next hidden boundary. It contains reports and measurement evidence, not the overload-value implementation. Its report identifies fdb63c7bc8 as that implementation.

5a718e9356a53f6909ae6abc227bf50da1ce2952 adds visitor code that directly calls the missing helpers below and modifies a missing file. e9fd9bbbca290ce5dbde7952522eab91a9953917 contains the visitor evidence. Applying the three named commits alone cannot provide these dependencies.

| Missing file | Missing symbols used by the requested change | Introduction commit |
| --- | --- | --- |
| internal/lower/overload_values.go | overloadValueBinding, overloadValueTarget, callOverloadValue | fdb63c7bc8bf42976cfd5c0edd9186c78bf2a272 |
| internal/lower/overload_relations.go | overloadResultFailure | c09a96043e |
| internal/lower/overload_results.go | overloadDeferredParameter, overloadNullableParameter | c68bf26cb4225e1f24a84f1a412f10e7eaa3da9a |

Observed with git cat-file on HEAD and source inspection of the requested commits: these files and symbols are absent. The intervening own production commits touching the result implementation are 0c84ce9c (callback contracts), 8c7a0d84 (structural results), c09a9604 (fixed return contracts and relation diagnostics), 0cfe794b (checked result fields), and fdb63c7b (known closure promises), after c68bf26c. This is the earlier overload-result foundation from the old stack, not a dependency present on main.

Conservative choice: honor the parent's instruction to name a missing stack dependency and stop. Do not take the evidence-only commits and imply the checked behavior landed, and do not silently import the prerequisite production chain. Step 2 remains blocked on rebuilding that chain. Only this Source verification and dependency report are committed and pushed. No counts change, new fixture, product change or whole-package run.

## Setup and lane

Ran `GOPROXY=https://proxy.golang.org|direct timeout 180 bash cloud/setup.sh` and sourced /workspace/adamic-tools/env.sh. Go build ready 36.045 s; setup done 36.333 s. nproc=5; cpu.max=400000 100000 (four CPUs). Full setup timings are in setup.log.

After committing this evidence, run the required bounded lane command before pushing. Its output is reported in the final response.
