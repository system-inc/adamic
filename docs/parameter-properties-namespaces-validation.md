# Parameter properties and namespaces: validation

Base: `origin/main` at `5d4c801`. Branch: `codex/parameter-properties-namespaces`.
The parameter-property and namespace steps are committed separately.

Setup printed Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 71s,
total 71s; `nproc` is 5, with a four-CPU cgroup quota. The environment file is
`/workspace/adamic-tools/env.sh`.

## Parameter properties

The two registered fixtures are `parameter_properties.a` and
`parameter_properties_ownership.a`. Both pass source Node, generated JavaScript,
native release, ASan/UBSan and LeakSanitizer. Callback captures and labels are
built at runtime so the ownership test does not depend on immortal strings.

Commands, with every test's output sent directly to a log:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/load -count=1 > /tmp/adamic-pp-packages-final.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/parameter_properties|TestParameterProperty' -count=1 -v > /tmp/adamic-pp-green.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/adamic-pp-counts.log 2>&1
go vet ./... > /tmp/adamic-pp-vet.log 2>&1
```

Output: lower 7.665s, load 0.598s, filtered oracle/mutants 0.849s, counts 10.635s,
all `ok`; vet's log is empty. The gate is extended after the namespace step.

Seven parameter-property mutants were killed:

| Mutant | Catch |
| --- | --- |
| Omit a numeric parameter-property store | Node stdout; compiled cleanly, exit 0, clean sanitizers |
| Move the store after the constructor body | Node stdout; compiled cleanly, exit 0, clean sanitizers |
| Omit the callback slot's retain | ASan heap-use-after-free |
| Ignore readonly-to-writable views | `TestParameterPropertiesSoundness/readonly_view` |
| Ignore parameter properties in override invariance | `TestParameterPropertiesSoundness/mutable_override` |
| Skip field-initializer definite initialization | `TestParameterPropertiesSoundness/field_initializer` |
| Skip parameter-default definite initialization | `TestParameterPropertiesSoundness/early_default` |

The first three are permanent oracle tests. Guard mutants temporarily removed
the named check, ran the named regression, and restored the source after each
run. Each failed an assertion; none was killed by a build error. Direct readonly
writes and external private/protected reads are rejected by the TypeScript
checker, with a mutable/public control admitted.

An initial early-field-read probe declared the class without constructing it,
so it never reached field-initializer lowering. The corrected probe constructs
the class through a variable declaration; the guard mutant is then killed.

Counts added: parameter_properties.a 39 allocations/39 frees, 24 retains,
66 releases, peak 10; ownership 120/120, 192 retains, 232 releases, peak 12.
All prior rows are unchanged. Neither program disables regions or Perceus reuse.
