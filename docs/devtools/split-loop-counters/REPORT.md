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

Instrument: Python `time.perf_counter` around the whole command. Unit durations: Go's `go test -json` `Elapsed` for leaf pass/fail events. Parent setup is also a budget unit. For parents with direct parallel children, the parent `Elapsed` excludes child waits and measures its own setup/cleanup. The table below reports command wall; setup measurements are recorded separately in the follow-up section.

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

## Setup-budget follow-up

The gate budgets both leaf subtests and each parent's own setup/cleanup. The initial wall comparison did not report setup explicitly. On this same instance, fresh cold measurements of commit `3b720788` show that its setup was already below 30 s (not the whole counts wall). This follow-up removes native initialization from both parents altogether, using option (b): the existing independent case units perform it, with the existing `sync.Once` still building once per process. All time spent building or waiting for shared initialization remains charged to those leaves; it is not subtracted as a build product. The `internal/buildcache` dependency is therefore unnecessary for this change.

Before children, counts now only enumerates cases, reads/validates the recorded table, and allocates result slots and a CLI output path. Loop counters only generates its unchanged source declarations and allocates batch paths. Native runtime libraries, Node version discovery, Node batch observations, CLI builds, and result reads occur within selected leaves.

| Test | Setup before this follow-up, cold | Setup after, cold | Largest after leaf | Case coverage |
|---|---:|---:|---:|---:|
| TestCountsAreRecorded | 14.38 s | 0.02 s | 11.56 s | 819 / 819 |
| TestLoopCountersAgreeWithNode | 11.28 s | 0.03 s | 12.83 s | 1,193 / 1,193 |

Instrument: Go's testing timer, reported as the parent pass event's `Elapsed` by `go test -json`. Direct parallel children's waits are excluded from that parent duration; shared initialization and `sync.Once` waits are included in each leaf's `Elapsed`. Python `time.perf_counter` also records the full command wall. Both before and after use `GOMAXPROCS=4`, `-parallel=4`, `-count=1`, `ADAMIC_GATE_UNCACHED=1`, and a distinct initially empty XDG native/result cache; the preparatory Go build cache is reused. `setup-timings.json` and the four compressed JSON logs retain all measurements and exact commands/environments. This distinguishes own setup from total wall; it does not assume that an almost unchanged wall means unchanged setup.

For comparison with the original main-based harness, the previously retained cold logs have parent durations of 58.08 s for counts and 46.00 s for loops on this box. The original counts parent runs a synchronous fixture group plus interface/predicate helpers; the split moved those helper cases into their own leaves. The new follow-up moves its remaining shared runtime initialization into those measured leaf paths.

Every before/after run's leaf names exactly match the existing 819 / 1,193 inventories, with no omissions or extra case units. All after parents and all 2,012 after leaves are under 30 s and pass. A separate cold `^never$` child filter passes both parents without creating any native/result cache directory (`setup-no-leaves.json`, JSON log), verifying that filtering out leaves performs no shared native/result initialization. Full gate package checks and the exact requested loop command pass on the final harness (follow-up logs).

`implementation-before-setup.sha256` preserves the prior harness hashes; `implementation.sha256` now identifies this measured follow-up. Generated case and batch source hashes are unchanged.
