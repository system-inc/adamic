# Hidden boundaries on main: step 1

Rebuilt hidden-08 and hidden-06 on origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0. Only these two storage changes are included.

## Commits and resolution

- f81387f242baabae0bca55801cc68a801ca672bf becomes 455b7932: never[] uses ordinary non-reference storage while retaining the checker element type. Writable widening stays refused.
- 702c3ecb507cab7199f7b3e60ee3802262e92eb7 becomes 46ee597d: concrete substitutions and mappings take precedence; scalar constraints retain scalar storage; object constraints retain runtime tags. Unsupported constraints and unsafe mutation stay refused.
- Both cherry-picks conflicted only in internal/oracle/counts.md. Preserve every main row and regenerate the four new rows. No existing row moves or changes.
- Main already has internal/lower/instantiate.go:newTypeMapper and the tagged representation helpers. No dependency only available on stack c3 was needed.
- Main's TestRepresentationClockSourceCheckedTypes expected unresolved Source to remain unsupported. Update that expectation to tagged Union, keeping the concrete subtype probes and unsupported length-only constraint. This is the same test expectation addressed by cb76d847, without taking that commit or its other evidence.
- Move the copied historical hidden-06 census under historical-hidden-06/. Those measurements belong to its original base and are not a census of this branch.

## Setup

GOPROXY=https://proxy.golang.org|direct; timeout 180 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh.
Go ready 0.026 s; Node ready 0.027 s; markdown ready 0.084 s; submodules ready 0.086 s; clang ready 0.216 s; Go build ready 36.345 s; done 36.643 s. nproc=5; cgroup cpu.max=400000 100000, four CPUs; GOMAXPROCS=4 for verification.

## Targeted lower tests

Command: `timeout 150 go test ./internal/lower -run "^(TestHidden|TestRepresentationClockSourceCheckedTypes)" -count=1 -timeout 90s -json`. No package-wide run. Constraint probes are separate top-level parallel tests.

| Test | Seconds |
| --- | ---: |
| TestHiddenTNodeConstraintObjectUnion | 0.070 |
| TestHiddenTNodeConstraintUnconstrained | 0.080 |
| TestHiddenTNodeConstraintNumber | 0.080 |
| TestHiddenTNodeConstraintUnknown | 0.100 |
| TestHiddenTNodeConstraintMutationRemainsNotYet | 0.060 |
| TestRepresentationClockSourceCheckedTypes | 0.090 |
| TestHiddenTNodeGenericValueRemainsNotYet | 0.060 |
| TestHiddenTNodeConstraintArrayLayout | 0.050 |
| TestHiddenTNodeConstraintAny | 0.040 |
| TestHiddenTNodeConstraintStructuralLength | 0.050 |
| TestHiddenNeverArrayRejectsWritableWidening | 0.130 |
| TestHiddenTNodeConstraintBoolean | 0.030 |
| TestHiddenTNodeConstraintObject | 0.040 |
| TestHiddenTNodeConstraintString | 0.040 |

## Node-held fixtures

Command: `ADAMIC_GATE_UNCACHED=1 timeout 150 go test ./internal/oracle -run "^TestNativeAgreesWithNode$/internal/oracle/testdata/hidden_boundary_(generic_tnode|never_array).*" -count=1 -timeout 90s -json`. Positive fixtures compare stdout, stderr and exit status with Node in JavaScript, ASan/UBSan native and release native; the harness checks leaks and recorded ownership counts.

| Fixture | Result | Seconds |
| --- | --- | ---: |
| hidden_boundary_generic_tnode_value.a | NotYet retained | 0.180 |
| hidden_boundary_generic_tnode_mutation.a | NotYet retained | 0.200 |
| hidden_boundary_generic_tnode_constraints.a | Node agrees in both backends, sanitized and release native | 1.060 |
| hidden_boundary_generic_tnode.a | Node agrees in both backends, sanitized and release native | 1.360 |
| hidden_boundary_never_array_observations.a | Node agrees in both backends, sanitized and release native | 1.360 |
| hidden_boundary_never_array.a | Node agrees in both backends, sanitized and release native | 1.190 |

## Mutants

Go overlays leave production files untouched. Each command uses `-count=1 -timeout 90s`; logs, overlay maps and .diff/.go.txt sources are beside this report. All eleven targeted runs fail on their intended assertion.

| Mutant | Caught by | Seconds |
| --- | --- | ---: |
| array-layout | TestHiddenTNodeConstraintArrayLayout | 0.030 |
| default | TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_generic_tnode.a | 0.540 |
| generic-value | TestHiddenTNodeGenericValueRemainsNotYet | 0.060 |
| ignore-mapper | TestHiddenTNodeConstraintString | 0.100 |
| ignore-substitution | TestHiddenTNodeConstraintNumber | 0.090 |
| no-constraint | TestHiddenTNodeConstraintNumber | 0.060 |
| object-layout | TestHiddenTNodeConstraintObject | 0.030 |
| unconstrained | TestHiddenTNodeConstraintUnconstrained | 0.040 |
| never-storage | TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_never_array.a | 0.030 |
| never-writable | TestHiddenNeverArrayRejectsWritableWidening | 0.050 |
| source-object-layout | TestRepresentationClockSourceCheckedTypes | 0.030 |

## Counts and scope

`timeout 150 go test ./internal/oracle -run "^TestCountsAreRecorded$" -count=1 -timeout 90s -args -update-counts` passes in 78.876 s. This is the existing repository-wide count recorder, not a new or touched test leaf. It adds only hidden_boundary_generic_tnode.a, hidden_boundary_generic_tnode_constraints.a, hidden_boundary_never_array.a and hidden_boundary_never_array_observations.a. All new or touched lower and fixture leaves are below 60 s.

No overload-result, visitor, optional array access or method destructuring changes were taken. No current census was run; later steps remain outside this turn. No whole packages or full gate were run.
