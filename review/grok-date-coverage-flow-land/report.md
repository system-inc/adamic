# Date coverage flow landing

Source: origin/codex/grok-date-coverage-flow at 06fc424.
Landing branch: codex/grok-date-coverage-flow-land.
Rebased onto origin/main e8ba3d5. Original branches stay unchanged.
Never push main; never force-push. @system_adamic merges into area/library.

Preserved main accessor and user-method dispatch alongside Date guards and dispatch.
Preserved both Date and main call-target oracle registrations. Regenerated the
conflicted counts table; all measured rows match, with no further value changes.
The exhaustive library_date_days.a is unchanged, retained in the oracle, excluded
only from flow tracing: its 1.46 million days at three times exceed the tracing limit,
while other library_date fixtures cover its shapes.

## Verification

Source /workspace/adamic-tools/env.sh; stock tsc on PATH. This isolated worktree
used GOFLAGS=-buildvcs=false and a temporary link to the installed cohere submodule,
removed after testing. All output was redirected to the accompanying logs.

```
TZ=UTC go test ./cmd/adamic-test262 ./internal/flow ./internal/fresh ./internal/ir \
  ./internal/javascript ./internal/load ./internal/lower ./internal/native \
  -count=1 -timeout 30m
TZ=UTC ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/library_date_' \
  -count=1 -timeout 30m -v
TZ=UTC ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestDateOracleCatchesMutants$' -count=1 -timeout 30m -v
TZ=UTC go test ./internal/oracle -run '^TestCountsAreRecorded$' \
  -count=1 -timeout 30m -args -update-counts
```

All pass. Runner 537.832s; flow 158.160s; fresh 84.586s; IR 2.478s;
load 3.284s; lower 57.001s; native 236.447s; JavaScript has no tests.
The full flow package includes TestEveryPathNodeTakesIsInTheGraph,
TestEveryMutationIsInItsRange and TestLivenessHoldsOnEveryPath: all passed.
Uncached Date oracle 190.390s, zero disagreements, including library_date_days.a
(125.15s). Eight mutants passed in 2.252s: constructor_clip, UTC, invalid_NaN,
getters, setters, iso_format, format, parse. Only Node stdout comparison catches
the valid, sanitizer-clean mutants. Full counts regeneration passed in 31.710s.
No full repository gate was run; all touched packages and Date fixtures passed.

Nullable compiler path is not at this base; another rebase and gate are required
when it lands. Date-3 remains stopped. This branch retains its original arithmetic
coverage scope; the subsequent Date JSON/parser additions are on the other lands.
