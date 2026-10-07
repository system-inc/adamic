- Built: opt-in stage measurements and a four-worker replay; existing gate flags and tests are unchanged.
- Commits: compiler/runtime base `e8ba3d5d81de4d3773c723914fccd4c76248b965`; sample `e74e313203a020646cce1c83d95a5362f167c5f1`; replay `b61bdf56f140bdb0329823a45dd6eab80b0d967b`.
- Commands and outputs: three uncached oracle package baselines passed; the 50-fixture measurement and all 307 fixture replays passed; hash checks and repository vet passed.
- Mutants: omit stdout, stderr, exit code, or stream boundaries from the hash; each failed `TestO0HashCatchesMutants` with exit 1 at the intended assertion.
- Not covered: an O0 run of every other oracle probe/input test, the separate O2 verification lane, other repository packages, or sanitized O0 beyond the selected 50 fixtures.

**The proposed fixture correctness lane saves 1.215 seconds (2.86%) on this four-core quota: 42.547 to 41.331 seconds, best of three.** Sanitized compilation and execution stay at O1; release and counted builds use O0, including their runtime libraries. All three paired runs improved, by 0.522s, 1.228s, 1.136s.

The unmodified complete oracle package took 66.975s, best of three. A planning extrapolation would be about 65.8s, or roughly **one second saved**, if the remaining work and scheduling are unchanged. That is an inference, not a measured modified-package gate: the replay combines each fixture's counted work with its comparisons, uses four workers, separates clang compilation and linking, and excludes other oracle tests. The original package uses its existing parallel test structure and defaults to five test workers under a four-core quota.

Keeping a separate O2 verification lane adds builds, links and executions. It does not remove that O2 work from the whole gate. No reduction in total gate work has been demonstrated for the two-lane proposal. The saving above is for the correctness lane with O2 work removed from that lane.

**Clang is the big stage here.** In the full fixture replay, main.c compilation occupies 30.5% of the four-worker wall budget and linking another 30.4%. Source and backend Node together occupy 21.6%; sanitized execution plus leak checks occupy 7.8%. Neither Node nor sanitized execution dominates this population. However, O0 only reduces compilation, and its slower release/count execution consumes much of the gain.

## Timing conditions and commands

All times below are from the same box and unchanged compiler/runtime source at the base commit. Each before/after pair is interleaved on the same harness commit; order reverses in loop 1. The 50-fixture run took more than five minutes (415.445s), but all three repetitions were completed anyway. The six-lane replay completed in 259.475s. Those aggregate test durations include preparation and measurement overhead and are not optimization comparisons.

Build-flags line (Sample, covering all its timing tables): commit `e74e313203a020646cce1c83d95a5362f167c5f1`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; load before `5.45 4.96 2.79 1/168 59096`; load after `1.52 2.21 2.23 1/161 76240`; `uncached observations; fresh runtime artifacts once per exact flag set`. Replay additionally sets `GOMAXPROCS=4` and `-parallel 4`.

Build-flags line (Replay, covering all its timing tables): commit `b61bdf56f140bdb0329823a45dd6eab80b0d967b`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; load before `0.72 1.65 2.02 1/156 76910`; load after `4.58 3.54 2.74 1/175 128037`; `uncached observations; warm runtime artifacts; sanitized O1 in both lanes`. Replay additionally sets `GOMAXPROCS=4` and `-parallel 4`.

Common C flags:

```text
-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls
```

Sanitized adds `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`; release adds `-O2`; counted adds `-DADAMIC_COUNT -O2`. The sample replaces each O1/O2 token with O0, preserving every other flag. The replay preserves sanitized O1 and replaces only release/count O2 with O0. Every build's complete flags are in the raw observations; a separate verification of those records checked optimization, sanitizer, count, FP contraction and sibling-call flags.

Exact commands, run from the repository root after `source /workspace/adamic-tools/env.sh`:

```sh
# Repeat for loop=1,2,3; record /proc/loadavg immediately before and after.
ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle > /tmp/adamic-o0-baseline-$loop.json 2>&1

ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1

GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_WALL_OUTPUT=/tmp/adamic-o0-wall.jsonl go test -v -parallel 4 -count=1 -timeout 30m ./internal/oracle -run '^TestO0MeasureWall$' > /tmp/adamic-o0-wall.log 2>&1
```

## Unmodified package calibration

Build-flags line: sample harness commit and tool versions above; sanitized O1, release/count O2; uncached oracle observations, warm existing runtime artifacts; default Go test parallelism. Each row gives its load before and after (one-, five-, fifteen-minute averages). Setup and vet had finished before these three runs. An earlier run overlapping setup cache warming was discarded.

| Loop | Before, current flags | After | Load before / after | Instrument |
|---|---:|---|---|---|
| 1 | 67.847s | Not changed | 3.59 4.32 2.07 / 5.07 4.70 2.37 | `ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle > /tmp/adamic-o0-baseline-1.json 2>&1` |
| 2 | 68.161s | Not changed | 5.07 4.70 2.37 / 5.61 4.84 2.58 | `ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle > /tmp/adamic-o0-baseline-2.json 2>&1` |
| 3 | 66.975s | Not changed | 5.61 4.84 2.58 / 5.45 4.96 2.79 | `ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle > /tmp/adamic-o0-baseline-3.json 2>&1` |

## Full population wall and stage measurements

Replay build-flags line and exact replay command above apply. Runtime artifacts are already built before each timed lane; all fixture observations run uncached. This measures 307 lowered fixtures, including counted builds except the existing stack_overflow exclusion, with the original checked-fixture comparison and leak policy. It performs one source Node and one backend Node run per fixture, not one per native variant.

| Loop | Before | After | Instrument |
|---|---:|---:|---|
| 0 | 44.904s | 44.383s | `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_WALL_OUTPUT=/tmp/adamic-o0-wall.jsonl go test -v -parallel 4 -count=1 -timeout 30m ./internal/oracle -run '^TestO0MeasureWall$' > /tmp/adamic-o0-wall.log 2>&1` |
| 1 | 42.559s | 41.331s | `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_WALL_OUTPUT=/tmp/adamic-o0-wall.jsonl go test -v -parallel 4 -count=1 -timeout 30m ./internal/oracle -run '^TestO0MeasureWall$' > /tmp/adamic-o0-wall.log 2>&1` |
| 2 | 42.547s | 41.411s | `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_WALL_OUTPUT=/tmp/adamic-o0-wall.jsonl go test -v -parallel 4 -count=1 -timeout 30m ./internal/oracle -run '^TestO0MeasureWall$' > /tmp/adamic-o0-wall.log 2>&1` |

| Loop/mode | Load before | Load after |
|---|---|---|
| 0/current | 0.72 1.65 2.02 | 3.20 2.18 2.18 |
| 0/o0 | 3.20 2.18 2.18 | 3.88 2.50 2.29 |
| 1/current | 4.58 2.85 2.42 | 4.90 3.21 2.56 |
| 1/o0 | 3.88 2.50 2.29 | 4.58 2.85 2.42 |
| 2/current | 4.90 3.21 2.56 | 4.96 3.44 2.67 |
| 2/o0 | 4.96 3.44 2.67 | 4.58 3.54 2.74 |

Stage totals below use the best whole lane in each mode (before loop 2, after loop 1), rather than choosing a different favorable loop for each stage. Seconds are summed across workers. Wall shares are `stage seconds / (4 * lane wall seconds)`: average occupation of the four-worker wall budget. Parallel stages do not have a unique additive critical-path attribution; these shares are not CPU profiles or claims that deleting a stage saves its full occupancy. Instrument: the exact replay command above.

| Stage | Before worker seconds | After worker seconds | Before wall share | After wall share |
|---|---:|---:|---:|---:|
| Load and lower | 8.793 | 8.850 | 5.17% | 5.35% |
| C emission | 0.499 | 0.492 | 0.29% | 0.30% |
| clang on main.c | 51.910 | 41.646 | 30.50% | 25.19% |
| Link | 51.663 | 51.938 | 30.36% | 31.42% |
| Sanitized binary run | 7.631 | 7.000 | 4.48% | 4.23% |
| Release binary run | 2.476 | 5.604 | 1.45% | 3.39% |
| Counted binary run | 2.443 | 6.069 | 1.44% | 3.67% |
| Leak-check binary run | 5.685 | 5.687 | 3.34% | 3.44% |
| Source Node run | 24.182 | 23.740 | 14.21% | 14.36% |
| Backend Node run | 12.505 | 12.480 | 7.35% | 7.55% |
| Scheduling and unassigned overhead | 2.400 | 1.820 | 1.41% | 1.10% |

Runtime build contributes zero to the timed warm-artifact lane; its fresh-build timings are measured separately below. Main compilation saves 10.264 worker seconds, while release and counted execution together increase by 6.754 worker seconds in these best lanes. Linking barely changes. An optimized release-only verification build/link/run represents another 25.19 measured worker seconds (about 6.30 four-worker capacity seconds); including optimized counted verification makes that 50.43 worker seconds (12.61 capacity seconds). These are work estimates from the current replay, not separately measured O2 lane elapsed times.

## Selected 50 fixtures: all variants at O0

Sample build-flags line above applies. Each loop aggregates 50 fixtures and all three native variants; before is sanitized O1 and release/count O2, after all three are O0. Node is deliberately repeated beside each variant and each mode to detect drift. Thus these totals describe the diagnostic workload, not the production gate. The first sample revision also reran checked fixtures with leak detection; those panic observations are retained in the evidence, and the final harness follows the original policy of skipping their leak run.

| Loop | Before stage-total seconds | After stage-total seconds | Instrument |
|---|---:|---:|---|
| 0 | 59.769 | 71.276 | `ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1` |
| 1 | 57.961 | 71.326 | `ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1` |
| 2 | 59.723 | 71.461 | `ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1` |

Best diagnostic total: before loop 1, after loop 0. Shares here are of the summed sequential diagnostic stages, excluding one-time runtime preparation and shared lowering; they are not whole-oracle wall shares. Instrument: the exact sample command above.

| Stage | Before seconds | After seconds | Before share | After share |
|---|---:|---:|---:|---:|
| C emission | 0.184 | 0.188 | 0.32% | 0.26% |
| clang on main.c | 13.061 | 5.373 | 22.53% | 7.54% |
| Link | 7.489 | 7.657 | 12.92% | 10.74% |
| Binary run, all variants | 9.407 | 23.992 | 16.23% | 33.66% |
| Diagnostic leak rerun | 3.719 | 8.509 | 6.42% | 11.94% |
| Source Node run | 13.768 | 14.776 | 23.75% | 20.73% |
| Backend Node run | 10.334 | 10.782 | 17.83% | 15.13% |

Per-variant totals in those same loops:

| Variant | Stage | Before seconds | After seconds |
|---|---|---:|---:|
| sanitized | emit | 0.059 | 0.064 |
| sanitized | clang | 7.341 | 2.545 |
| sanitized | link | 4.776 | 4.955 |
| sanitized | run | 5.772 | 14.501 |
| release | emit | 0.058 | 0.060 |
| release | clang | 2.857 | 1.402 |
| release | link | 1.346 | 1.341 |
| release | run | 1.786 | 4.244 |
| counted | emit | 0.067 | 0.064 |
| counted | clang | 2.862 | 1.426 |
| counted | link | 1.368 | 1.360 |
| counted | run | 1.849 | 5.247 |

The total for all-O0 diagnostic runs increases by 13.315s in the best loops. In bitwise_sweep alone, sanitized main.c compilation at O0 is about 0.05s, but the binary comparison takes about 4.7s and its leak rerun about 4.9s. The proposed lane correctly avoids this sanitized slowdown.

Fresh runtime compilation and archiving, measured once per exact flag set, from an immutable source/header snapshot. These artifacts are private to the experiment and reused by its fixtures; no result cache is added. Instrument: the exact sample command above. Sample build-flags line applies; runtime artifacts are fresh, not loaded from the normal runtime cache.

| Flag-set loop | Before seconds | After seconds | Instrument |
|---|---:|---:|---|
| sanitized | 3.846 | 1.652 | `ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1` |
| release | 2.285 | 1.067 | `ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1` |
| counted | 2.282 | 1.041 | `ADAMIC_GATE_UNCACHED=1 ADAMIC_O0_BASELINE=/tmp/adamic-o0-baseline-1.json ADAMIC_O0_OUTPUT=/tmp/adamic-o0-results.jsonl go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestO0Measure$' > /tmp/adamic-o0-measure.log 2>&1` |

The runtime compiler cache is not an added cache in this unit. Measurement builds snapshot and build their O0 artifacts directly, and every fixture execution is uncached. Shared load/lower in the selected sample was 1.714s, measured once per fixture and excluded from the flag-dependent totals.

## Fixture selection

Baseline loop 1 was an uncached `go test -json` run of the complete internal/oracle package. Rank the 307 TestNativeAgreesWithNode fixture observations by elapsed descending, breaking ties by path ascending. Take 25 slowest. Sort the other 282 paths and take indices `0, 11, 22, ..., 264` (fixed stride 11), yielding 25 distinct additional fixtures. The measurement regenerates this choice from the baseline log on every rerun; no random seed or hand selection is involved.

| Slowest 25 fixture | Baseline subtest seconds |
|---|---:|
| `internal/oracle/testdata/bitwise_sweep.a` | 8.40 |
| `internal/oracle/testdata/concat_too_long.a` | 4.65 |
| `internal/oracle/testdata/user_iterators.a` | 2.93 |
| `internal/oracle/testdata/sweeps/regexp_methods.a` | 2.67 |
| `internal/oracle/testdata/size_class_churn.a` | 2.55 |
| `internal/oracle/testdata/regexp.a` | 1.49 |
| `internal/oracle/testdata/long_chain.a` | 1.15 |
| `internal/oracle/testdata/integer_format.a` | 1.09 |
| `internal/oracle/testdata/collections.a` | 1.05 |
| `internal/oracle/testdata/class_as_interface.a` | 0.97 |
| `internal/oracle/testdata/closures_throw.a` | 0.94 |
| `internal/oracle/testdata/visits.a` | 0.91 |
| `internal/oracle/testdata/class_inheritance_generic.a` | 0.89 |
| `internal/oracle/testdata/library_map_set.a` | 0.88 |
| `internal/oracle/testdata/sets.a` | 0.86 |
| `internal/oracle/testdata/library_map_set_next.a` | 0.85 |
| `internal/oracle/testdata/class_features_accessors.a` | 0.84 |
| `internal/oracle/testdata/large_output.a` | 0.84 |
| `internal/oracle/testdata/named_function_values.a` | 0.84 |
| `internal/oracle/testdata/replace_all_large.a` | 0.84 |
| `internal/oracle/testdata/maybe_number_slots.a` | 0.83 |
| `internal/oracle/testdata/finally_leaves.a` | 0.82 |
| `internal/oracle/testdata/library_fnexpr_loops.a` | 0.81 |
| `internal/oracle/testdata/shared_slices.a` | 0.78 |
| `internal/oracle/testdata/library_function_expressions.a` | 0.77 |

| Fixed-stride 25 fixture | Baseline subtest seconds |
|---|---:|
| `dedication/dedication.a` | 0.45 |
| `internal/load/testdata/0.1/compile/10_unicode.ts` | 0.49 |
| `internal/oracle/testdata/borrow_defined_lent_field.a` | 0.53 |
| `internal/oracle/testdata/call_targets_reuse.a` | 0.52 |
| `internal/oracle/testdata/class_identity.a` | 0.45 |
| `internal/oracle/testdata/class_oct6_subclass_holder.a` | 0.54 |
| `internal/oracle/testdata/e4eec87_f1_field_narrowed.a` | 0.47 |
| `internal/oracle/testdata/find_index_shrinks.a` | 0.47 |
| `internal/oracle/testdata/generic_values.a` | 0.51 |
| `internal/oracle/testdata/json_stringify_scalars.a` | 0.50 |
| `internal/oracle/testdata/library_array_metadata.a` | 0.44 |
| `internal/oracle/testdata/library_map_set_construct.a` | 0.73 |
| `internal/oracle/testdata/library_object_assign.a` | 0.48 |
| `internal/oracle/testdata/library_object_same.a` | 0.43 |
| `internal/oracle/testdata/map_iteration.a` | 0.57 |
| `internal/oracle/testdata/narrowed_compared.a` | 0.53 |
| `internal/oracle/testdata/number_formats.a` | 0.54 |
| `internal/oracle/testdata/panic_in_try.a` | 0.46 |
| `internal/oracle/testdata/regexp_cycle_weak.a` | 0.46 |
| `internal/oracle/testdata/regions.a` | 0.55 |
| `internal/oracle/testdata/reuse_narrowed.a` | 0.45 |
| `internal/oracle/testdata/set_maybe_numbers.a` | 0.49 |
| `internal/oracle/testdata/spread_snapshot.a` | 0.59 |
| `internal/oracle/testdata/strings.a` | 0.55 |
| `internal/oracle/testdata/undefined_elements.a` | 0.49 |

## Output agreement and mutants

**Differences: none.** All 2,982 selected-sample executions and all 10,770 full-population executions carry stdout and stderr SHA256 values, raw exit code, and a combined tuple SHA256. Every matching before/after pair agrees. The hash manifest also gives a separate SHA256 of the decimal exit code. This is evidence for these fixtures and flag sets, not a proof that arbitrary programs have no undefined behavior.

The combined hash is SHA256 over JSON containing the exit integer and base64-encoded stdout/stderr bytes. JSON preserves boundaries and arbitrary bytes; nothing trims, normalizes or ignores stderr. A difference record would retain the full before/after byte arrays (base64) and exit codes so a byte diff can be reproduced. There are no difference records to list.

| Code mutant | Intended failing assertion | Command and result |
|---|---|---|
| drop_stdout | `mutant 0 escaped` in TestO0HashCatchesMutants | `go test -count=1 ./internal/oracle -run '^TestO0HashCatchesMutants$' > /tmp/adamic-o0-mutant-drop_stdout.log 2>&1`: exit 1 |
| drop_stderr | `mutant 1 escaped` in TestO0HashCatchesMutants | `go test -count=1 ./internal/oracle -run '^TestO0HashCatchesMutants$' > /tmp/adamic-o0-mutant-drop_stderr.log 2>&1`: exit 1 |
| drop_exit | `mutant 2 escaped` in TestO0HashCatchesMutants | `go test -count=1 ./internal/oracle -run '^TestO0HashCatchesMutants$' > /tmp/adamic-o0-mutant-drop_exit.log 2>&1`: exit 1 |
| drop_boundaries | `mutant 3 escaped` in TestO0HashCatchesMutants | `go test -count=1 ./internal/oracle -run '^TestO0HashCatchesMutants$' > /tmp/adamic-o0-mutant-drop_boundaries.log 2>&1`: exit 1 |

Each mutant compiled successfully and failed the hash assertion, not clang or a compiler refusal. Source was restored between mutants. The restored command passed; the preexisting TestTheOracleCatchesOneByte also ran in all three complete package baselines and caught its changed dedication byte as stdout disagreement. No cache was added, so there is no new cache key-component mutant to run.

## Toolchain and remaining checks

`bash cloud/setup.sh > /tmp/adamic-o0-setup.log 2>&1` succeeded. Its printed timing lines: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 128s; done in 128s on 5 processors, cgroup cpu.max 400000 100000, 17.6 GB. The printed environment is `/workspace/adamic-tools/env.sh`, sourced for each build/test shell. Setup timing is printed setup evidence, not a comparative benchmark; its load before was not recorded. Setup was excluded from all comparative measurements.

`go vet ./... > /tmp/adamic-o0-final-vet.log 2>&1` passed. `gofmt -l cmd internal` and `git diff --check` passed with no diagnostics. Complete repository `go test ./...` was not run; the complete oracle package passed three times uncached, and the opt-in measurements exercised the touched package directly. A diff against the base confirms no changes to internal/native, internal/lower, oracle_test.go, or counts_test.go.

## Evidence and rerunning

Checked-in gzip artifacts preserve the uncached selection baseline, raw stage observations, full replay observations, hash manifest and mutant logs. Decompress with `gzip -dc`; use the baseline JSON as ADAMIC_O0_BASELINE, or generate a fresh uncached baseline with the command above. ADAMIC_O0_ALL=1 expands the stage experiment to the full population; ADAMIC_O0_ONCE=1 opts into one repetition for a long run. The wall replay always runs three interleaved pairs. All subprocesses use the oracle's existing bounded execution helper.

- [o0_measure_baseline.gz](o0_measure_baseline.gz), decompressed SHA256 `a822fdf2c05d827b63cf7b51e49b37e73f8b79c695dd03a5acc07f6d61e66e45`.
- [o0_measure_sample.gz](o0_measure_sample.gz), decompressed SHA256 `da3c4e993debc6aec23d2e074c11c5bbd056ec58a4c11ff1cd8b5dab2d65d813`.
- [o0_measure_wall.gz](o0_measure_wall.gz), decompressed SHA256 `569cfe5a52e73a409c0ec6e3baee09ef76f33e66e504fbfe4c6cf97b6dc537e2`.
- [o0_measure_hashes.gz](o0_measure_hashes.gz), decompressed SHA256 `c1d951181dacaf4d41139ce4dbb9122f7e2a27dc576292ffc1f0be3acf085dfb`.
- [o0_measure_mutants.gz](o0_measure_mutants.gz), decompressed SHA256 `b977a20472adc5a74458f6215928b954e73629e19a79aab0cb237f1f5a006346`.
