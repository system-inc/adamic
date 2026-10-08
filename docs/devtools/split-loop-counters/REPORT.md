# devtools-split-loop-counters

Branch base: `origin/main` at `54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8`. Budget: **30 seconds per shardable unit** on this instance's 4-CPU quota (cpu.max `400000 100000`).

## Change and coverage

`TestLoopCountersAgreeWithNode` now has 1,193 parallel leaves, `c1` through `c1193`. The original 1,193 case IDs are unchanged. There are still 1,205 counter-kind assertions: the last 12 cases also check their nested counter. `coverage.json` lists old and new IDs and counter names. The independently extracted generators produce exactly the same function bodies, call order, and expected counter kinds; removing the duplicated read-only prelude reassembles the original source byte for byte (`old-cases.json.gz`, `new-cases.json.gz`, generator sources alongside them).

Native sanitized/release binaries are shared in 19 groups of at most 64 functions, each built once through `sync.Once`. A command argument executes only the selected function. Node observes each group's original sweep once; every subtest owns its counter-kind assertions, expected output line, sanitized execution, and release execution. Functions only read shared globals. The batch input hashes include the added argument dispatcher; the per-case hashes describe the unchanged original bodies and prelude.

`TestCountsAreRecorded` now has 819 parallel leaves: the 790 original fixture leaves, 14 interface helper cases, and 15 predicate helper cases. `counts-coverage.json` lists both complete inventories. Mapping: original `TestCountsAreRecorded/fixtures/<path>` becomes `TestCountsAreRecorded/<path>`; formerly parent-run `interfaceCastCounts` and `predicateCountsTable` rows become subtests named by that row's source path. Runtime counting, predicate direction validation, fixture order, headers, and full `-update-counts` output remain covered. The two default-interface fixtures share one CLI build through `sync.Once`; runtime libraries use the existing shared identity/runtime setup.

The gate enumerates the initialized test inventory using `-oracle-unit-list`, including fixtures appended by init helpers. Both parents are expanded into leaves rather than assigned as aggregate shard units. The actual gate enumeration probe and complete gate package tests pass (logs included).

## Planted failures

`mutants.patch.gz` records both temporary changes, removed before final validation:

* Flip only c1's integer expectation: `TestLoopCountersAgreeWithNode/c1` fails with `kept in an integer true, want false`; c2 passes.
* Change only string_views_lifetime.a's recorded allocations from 19 to 20: its own counts leaf fails, showing recorded 20 versus measured 19; string_views_holders.a passes.

Each test command exited 1. Parent failures are propagation from those leaves. The compressed JSON logs retain the exact failures and neighboring passes. The final loops and counts runs pass after removing both changes.

## Cold measurements on this same box

Instrument: Python `time.perf_counter` around the whole command. Unit durations: Go's `go test -json` `Elapsed` for leaf pass/fail events. Aggregate parents with parallel children are not shard units; their `Elapsed` excludes child waits, so command wall time is reported below.

Pinned submodules, tool versions/binary hashes, quota, and cache policy are in `machine.json`. Repository Node dependencies were installed from `stage3/api/package-lock.json` before paired runs. Each measured run has a new empty native/result cache via XDG_CACHE_HOME, `ADAMIC_GATE_UNCACHED=1`, and `-count=1`. Go compilation is preparatory and its build cache is reused for paired measurements; this is cold test execution, not a fresh toolchain bootstrap. Sources and Node declaration hashes are included; `implementation.sha256` identifies the measured final harness, `generated-inputs.sha256` each original loop input, and `batch-inputs.json` the actual grouped dispatcher inputs.

| Command selection | Before wall | After wall | New leaves | Slowest after leaf |
|---|---:|---:|---:|---:|
| Loop counters, four workers | 47.87 s | 31.96 s | 1,193 | 1.84 s |
| Counts, four workers | 61.15 s | 60.53 s | 819 | 4.30 s |

Counts wall changed little; the substantive change is that all of its pieces can be sharded and mismatches fail inside their own leaf. The new loop test also scales with worker slots: cold wall 41.15 s at `-parallel=1` versus 31.96 s at `-parallel=4`, both on the same 4-CPU instance.

The final combined cold run passes all 2,012 leaves; its largest leaf is 4.55 s. Every enumerated leaf appears exactly once and passes. Combined command wall is 85.78 s versus 85.75 s before: running both parents together makes their children share four slots, and the combined wall is essentially unchanged. This is not a claim that the full repository gate was run or is below five minutes.

`timings.json` contains exact commands, environments, instrument readings, exits, and leaf maxima. Corresponding compressed JSON logs are retained.

## Required final command

`go test ./internal/oracle -run TestLoopCountersAgreeWithNode` exits 0 with both plants removed; `requested-command.log` retains its output. The full cold counts command also exits 0. The complete `cmd/adamic-gate` test package passes.
