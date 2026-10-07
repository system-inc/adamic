# Date landing verification

Source: origin/codex/library-date at f7cbc8c.
Landing branch: codex/library-date-land.
Base verified: origin/main e8ba3d5 (devirtualization and call-target routing included).
Date-3 stopped. This is landing work only, not new library implementation.

Standing rule acknowledged: land means rebase into our own -land branch, re-green,
push that branch, and report its SHA for @system_adamic to merge into area/library.
Never push main. Never force-push. The original Date branches remain unchanged.

## Conflicts

Preserved main's accessor literal and user-method dispatch in object.go alongside
Date's nominal internal-slot guard and Date call dispatch. The Date guard precedes
accessor lowering so a structural accessor object cannot supply Date's internal slot.
User-method dispatch retains priority over library dispatch.

Main advanced from e011f8f to e8ba3d5 during the first gate. Rebased again before
publishing and reran the gate below. Kept both Date and main's five call-target
fixtures in oracle_test.go. Regenerated the conflicted counts table using the real
counted programs; measured rows match the combined table, with no further value
changes. Date-specific lowering, runtime and parser files match the original branch.

The general nullable-reference compiler path has not landed at the verified base.
Existing Date compatibility remains for this first rebase. No nullable special case
was added or extended; another rebase and gate will be needed once that path lands.
The old TS2345 review is historical: the owner ruled programs rejected by stock
tsc are correctly refused, not language gaps.

## Commands and outputs

Source /workspace/adamic-tools/env.sh before each command; stock tsc 6.0.3 is on
PATH under /tmp/date-refusals-tsc/node_modules/.bin. Output was redirected to logs.

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

All pass. Package times: runner 162.750s, flow 147.311s, fresh 102.498s,
ir 2.921s, load 4.040s, lower 71.270s, native 207.647s; JavaScript has no tests.
Date oracle: 5.081s, uncached, zero disagreements across native, release and JS.
Fourteen Date mutants: 8.276s, each valid C, exit 0, empty stderr, caught only by
Node stdout comparison. Counts regeneration: 73.442s, complete table.
No complete repository test gate was run; all touched packages and Date fixtures
were re-greened. Logs are beside this report.

Setup initially attempted its warming build while a conflict was unresolved and
failed with syntax error unexpected << in object.go at lines 17 and 579. After
resolving it, setup was rerun successfully:

```
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (2s)
setup: node ready (3s)
setup: submodules ready (3s)
setup: build cache warm (356s)
setup: done in 356s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc=5. The first oracle filter selected only mutants; the explicit fixture
filter above was then run. All published evidence is from the final base e8ba3d5.
