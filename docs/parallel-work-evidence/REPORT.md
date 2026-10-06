# Removing shared fixture lists

Branch: `codex/no-shared-lists`, from current main
`5d4c8012a0877094134e6c6bac367ff68f9313e8`.
Scanner inspection only: `0090256e607c3f2de7d5b67cef680ec010f95c1d`.

## Built

Recursive sidecar discovery replaces the ordinary fixture list, its two init
appenders and the input fixture list. Each fixture owns its options and counts.
The human command is `go run ./cmd/oracle-fixtures -counts`; `-list` outputs paths
and all options, with byte-exact argv in hex. Filtered `-update-counts` writes only
changed files that were actually measured. Count comparisons still run on cached
observations. Input probe identities include their own options; generated-C,
Node, runtime and toolchain identities keep their existing cache behavior.

`../parallel-work.md` describes the scanner branch's registration and proposes
directory discovery with generated, untracked static dispatch. `CLAUDE.md` now
instructs fixture builders to add only their own source, sidecar and count file.

## Setup and baseline

```sh
git fetch origin && git checkout -b codex/no-shared-lists origin/main
bash cloud/setup.sh > /tmp/adamic-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

Setup exited 0; its timing lines were:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (159s)
setup: done in 160s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc`: 5. Go 1.27.1, clang 20.1.8, Node v24.19.0. The configured tools directory
was `/workspace/adamic-tools`. An initial snapshot attempt sourced the default
`/opt/adamic-tools/env.sh`, which did not exist, and the image's unrelated `go`
answered `Go: Unknown option: test`. Sourcing the configured path fixed it; setup
itself did not fail.

Before editing registration, a temporary `TestMigrationSnapshot` inspected the
actual in-memory ordinary and input registries after init, encoded argument bytes
as hex, and saved the existing table. It ran with:

```sh
go test ./internal/oracle -run '^TestMigrationSnapshot$' -count=1 > /tmp/oracle-snapshot.log 2>&1
```

Exit 0, oracle 0.130s. The temporary test was removed after migration. The captured
fixture set includes the seven class-inheritance registrations and the object
prototype init registration, not just the central literal.

## Equivalence

```sh
go run ./cmd/oracle-fixtures -list > docs/parallel-work-evidence/fixtures-after.json
go run ./cmd/oracle-fixtures -counts > docs/parallel-work-evidence/counts-after.md
diff -u docs/parallel-work-evidence/fixtures-before.json docs/parallel-work-evidence/fixtures-after.json
diff -u docs/parallel-work-evidence/counts-before.md docs/parallel-work-evidence/counts-after.md
```

Both diffs exited 0 and produced 0 bytes. Both lists have exactly 263 fixtures,
257 ordinary and 6 input, with identical booleans and every argument byte.
Both tables have 262 rows and identical headers, ordering and all six values per
row. `stack_overflow.a` remains executed but uncounted. These snapshots are
historical proof, not new registries to maintain when adding fixtures.

All measurements also passed:

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/oracle-counts-check.log 2>&1
```

Exit 0, oracle 41.815s. Metadata behavior tests passed in 0.008s.

## Mutants

Every test's output was written to its own log. Each source mutation was restored
in a `finally` block before the next one; no mutant is committed.

| Mutation | Exact test selection | Observed failure |
|---|---|---|
| Discovery returns `Checked: false`, ignoring the fixture's `checked: true` sidecar | `TestNativeAgreesWithNode/internal/oracle/testdata/writes_past_end.a` | Exit 1: source Node exits 0 and continues printing; native and backend exit 70 at the inserted bounds check. The oracle reports `exit codes differ`. No clang or sanitizer failure masks this check. |
| Remove `DisallowUnknownFields` | `TestDiscoveryRejectsBrokenOptions` | Exit 1: accepted the misspelled `cheked` option. |
| Return success for changed counts when update is off | `TestCountUpdatesAreIndependent` | Exit 1: `changed count was accepted`. |
| Rewrite a count file even when its measured row is unchanged | `TestCountUpdatesAreIndependent` | Exit 1: `unchanged counts file was rewritten`. |
| Change recorded allocations from 2 to 999 in writes_past_end.a's count file | `TestCountsAreRecorded/fixtures/internal/oracle/testdata/writes_past_end.a` | Exit 1: recorded 999, measured 2. All other values stayed equal. |

The source-mutant commands used:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/writes_past_end.a -count=1 -v -timeout 30m > /tmp/oracle-mutant-checked-option-ignored.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle/fixturedata -run TestDiscoveryRejectsBrokenOptions -count=1 -v -timeout 30m > /tmp/oracle-mutant-unknown-option-ignored.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle/fixturedata -run TestCountUpdatesAreIndependent -count=1 -v -timeout 30m > /tmp/oracle-mutant-count-change-ignored.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle/fixturedata -run TestCountUpdatesAreIndependent -count=1 -v -timeout 30m > /tmp/oracle-mutant-unchanged-count-rewritten.log 2>&1
```

The recorded-count mutant and real flag repair used:

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded$/fixtures/internal/oracle/testdata/writes_past_end.a$' -count=1 -v -timeout 30m > /tmp/oracle-recorded-count-mutant.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$/fixtures/internal/oracle/testdata/writes_past_end.a$' -count=1 -v -timeout 30m -args -update-counts > /tmp/oracle-filtered-count-update.log 2>&1
```

The repair exited 0 (1.923s) and changed exactly
`internal/oracle/testdata/writes_past_end.a.counts.json`. Comparing bytes and
nanosecond mtimes of all 262 count files proved every other file unchanged, and
the repaired file was byte-identical to its baseline. Running the update again
exited 0 (1.620s) and changed zero files, including their mtimes.

## Cache checks

```sh
go test ./internal/oracle -run '^TestGateCache' -count=1 -v -timeout 30m > /tmp/oracle-cache-tests.log 2>&1
```

Exit 0, 1.493s: generated C, runtime/Node/runner invalidation, Node dependency
inputs, atomic evidence and uncached bypass all passed.

Each of these commands ran twice, with separate `-1.log` and `-2.log` files:

```sh
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/writes_past_end.a$' -count=1 -v -timeout 30m > /tmp/oracle-cached-native-1.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$/internal/oracle/testdata/read_arguments.a$' -count=1 -v -timeout 30m > /tmp/oracle-cached-input-1.log 2>&1
```

All four exited 0. Native's first run: native misses 2, Node misses 2, 6.477s;
second: native hits 2, Node hits 2, no misses, 0.774s. Input's first run: probe
misses 1, 3.071s; second: probe hits 1, no misses, 1.303s.

The recorded-count mutant was also rerun after warming that exact count result.
`oracle-cached-count-warm.log` exited 0. `oracle-cached-count-mutant.log` exited 1
with `native hits=1 misses=0`, still reporting recorded allocations 999 against
measured 2. Thus cached observations do not cache the baseline comparison.
The original count file was restored byte for byte afterward.

## Gate and limits

`gofmt -l cmd internal` printed nothing, `git diff --check` printed nothing, and
`go vet ./... > /tmp/oracle-vet.log 2>&1` exited 0 with empty output.

The full uncached command is:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/oracle-full-uncached.log 2>&1
```

The full gate exited 0. Every package passed, including all stage 1 packages.
Oracle: 404.233s; native: 443.984s; JSON: 1009.382s; parser: 450.907s.
The longest package was unicodeproperties at 1407.295s. The full output is saved
in `oracle-full-uncached.log`. All oracle count comparisons passed with the
unchanged per-fixture baselines.

Only Linux was exercised. The scanner branch was inspected with `git show`, not
modified or tested with the proposed registry. That proposal needs implementation
and its own parity and omission mutants on the scanner worker's branch.
