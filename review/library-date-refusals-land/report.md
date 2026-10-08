# Date refusals landing

Source: origin/codex/library-date-refusals at 4ec07a3.
Landing branch: codex/library-date-refusals-land.
Rebased onto origin/main e8ba3d5. Original branches remain unchanged.
Never push main; never force-push. @system_adamic merges into area/library.
Date-3 stopped; no new library or nullable-special-case code was added.

## Conflicts and retained intent

Preserved main accessor literal and user-method dispatch alongside Date nominal
internal-slot guards and Date dispatch. Preserved both main call-target fixtures
and Date oracle registrations. Regenerated the conflicted counts table from the
complete fixture set: all measured rows match, no further value changes.
The original exhaustive library_date_days.a is retained unchanged in the oracle;
its existing exclusion from flow tracing remains, alongside smaller shape fixtures.
The branch retains JSON, dynamic/legacy parsing, metadata, Number(Date), and the
runner's Date adaptations. No changes to the other worker's verdict/classify files.

## Verification

Source /workspace/adamic-tools/env.sh and put stock tsc on PATH. All output was
redirected to the accompanying logs. Tests ran in the warmed primary checkout.

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

All pass. Runner 141.302s; flow 160.053s; fresh 89.968s; IR 4.022s;
load 3.832s; lower 56.488s; native 250.198s; JavaScript has no tests.
The full flow package includes all three tracing checks.
Uncached Date oracle 181.713s, zero disagreements, including library_date_days.a
(115.62s). Eighteen mutants passed in 6.903s, each caught only by Node stdout
comparison after valid compilation, successful exit and sanitizer-clean execution:
number_date, metadata_name, metadata_length, metadata_own, nullable_typeof,
nullable_number, nullable_stringify, date_stringify, toJSON, dynamic_parse,
constructor_clip, UTC, invalid_NaN, getters, setters, iso_format, format, parse.
Full counts regeneration passed in 43.833s. No full repository gate was run;
all touched packages and Date oracle fixtures passed.

The compiler's general nullable path is not at this base. These landing branches
need another rebase and gate when it lands. New library work remains stopped
pending the backlog decision. Old TS2345 coercion reports are historical; stock-tsc
rejection is correct refusal under the owner's ruling, not a compiler gap.
