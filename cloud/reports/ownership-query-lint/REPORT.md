# Ownership query lowering demand cost

Lowering now records move demand as it emits `parallelMap`, instead of asking
`OwnershipTransfers` to reflectively traverse every statement and expression in
the completed IR. With no moved site, flag-on skips discovery and heap preparation
entirely. With demand, lowering calls `QueryOwnershipTransfers` directly. The
original automatic-discovery API remains available for independently constructed
IR and the agreement harness.

`parallelMoves` classifies only `parallelMap` arguments and caches the result by
source call and type mapper. Preflight, callback validation, and emission reuse
one classification per instantiated site. The cache is allocated on the first
site; programs with no sites allocate nothing for demand detection. Different
generic instantiations retain independent classifications. Demand detection is
O(number of instantiated parallelMap sites), with an O(1) final flag check. Full
demanded heap preparation still has its existing CFG/interprocedural fixed-point
costs; this change makes no linear-cost claim for that analysis.

## Timings

Base: `b10d3ca9b416225b710d7b5ab2ec4b068255ae0d`. Same generated lint registry and
`stage1/cohere/lint/main.ts` input, 760 functions, in both binaries. Each sample
loads a fresh checked program and enables TSGo outside the timed region. An
explicit GC is also outside the region. Timing and CPU profiling cover only
`LowerWithOptions`, including its checker work, normalization, cycle finding,
optional ownership analysis, borrow, and counters. Each mode uses three samples.

| Source | Flag off samples (ms) | Flag on samples (ms) | Best off | Best on | On vs off |
| --- | --- | --- | ---: | ---: | ---: |
| Baseline | 2334.693, 2232.880, 2177.347 | 2111.780, 2368.348, 2327.069 | 2177.347 | 2111.780 | -3.01% |
| Optimized | 2260.283, 2296.007, 2315.997 | 2260.115, 2330.859, 2222.697 | 2260.283 | 2222.697 | -1.66% |

The optimized flag-on best is within 2% of flag-off. The original +14.6% baseline
regression did **not** reproduce on this machine. Negative differences are sample
variability, not a claimed speedup; the stronger evidence is that lowering no
longer enters ownership discovery at all for this input. Initial exploratory
profiles used `lint.ts`; those were discarded and are excluded from this table
and all saved profiles. [Raw samples](timings.txt).

Machine: Linux/amd64, AMD EPYC 9V74 80-Core Processor; 5 visible CPUs,
`GOMAXPROCS` default 5, cgroup quota `400000 100000` (4 CPUs), 16 GiB memory
limit; Go 1.27.1, Node v24.19.0, clang 20.1.8. Setup completed, and no task build
or test ran concurrently with timing. Shared-machine 1/5/15-minute load:
`5.60/5.50/3.52` before baseline, `4.58/5.28/3.48` between baseline and optimized,
`3.79/5.07/3.44` after optimized. The elevated load averages include the preceding
setup/build and sanitizer work; instantaneous runnable count was 1 at each
snapshot. [Machine](machine.txt), [load snapshots](load.txt).

## CPU profiles

Each saved pprof text merges all three lowering-only CPU profiles for its mode.
The default Go CPU sampling period is 10 ms.

Before, flag-on has the call graph:
`LowerWithOptions → OwnershipTransfers → hasMoved → reflect.Value.Kind`.
`OwnershipTransfers`/`hasMoved` account for 10 ms cumulative across the three
profiles (0.11% of 8.85 s sampled CPU). It is demand detection, with no demanded
heap preparation. After, neither flag setting has an ownership discovery or
query frame in its profile. The source guard also proves zero entry into that
path when no moved `parallelMap` was emitted; sample absence alone would not.

| Flag-on hotspot | Before flat / cumulative | After flat / cumulative |
| --- | ---: | ---: |
| `erasedMethodField.func1` | 610 / 1710 ms | 560 / 1650 ms |
| AST `Node.ForEachChild` | 460 / 2600 ms | 560 / 2550 ms |
| `runtime.tryDeferToSpanScan` | 510 / 780 ms | 510 / 830 ms |
| `OwnershipTransfers` / `hasMoved` | 0 / 10 ms | no samples; not called |

Full top and call graphs, including the flag-off controls:

- [Before off top](before-off-top.txt), [call graph](before-off-callgraph.txt)
- [Before on top](before-on-top.txt), [call graph](before-on-callgraph.txt)
- [After off top](after-off-top.txt), [call graph](after-off-callgraph.txt)
- [After on top](after-on-top.txt), [call graph](after-on-callgraph.txt)

The binary `.cpu` files remain in `/tmp/ownership-query/main-{before,after}-{off,on}-{0,1,2}.cpu`.
The unchanged baseline binary was compiled in a detached checkout at the base
commit, sharing the pinned cohere dependency checkout. Both binaries lowered the
same absolute source entry in the working repository. VCS stamping was disabled
for the baseline's shared dependency checkout.

## Validation

- `go test ./internal/lower ./internal/fresh -run OwnershipQuery -count=1 -v`:
  passed. Agreement: all 74 fixtures, 1 admitted IR transfer, 4 independent query
  refusals; normalized/extracted-root and region tests passed. [Log](agreement.txt).
- `go test ./internal/oracle -run '^TestOwnershipQueryNodeShapes$' -count=1 -v`:
  passed. All five shapes preserve their answers: flat and array admitted;
  nested, ring and outside-ring refused by the existing runtime gate. Admitted
  shapes and move fixtures passed release, ASan, slab ASan, and TSan variants
  with one worker and default workers. [Log](shapes.txt).
- `go vet ./internal/lower ./internal/fresh`, gofmt, `git diff --check`: passed.

The full repository gate and other operating systems were not run.

## Reproduce

From the repository root, after the normal cloud setup:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
mkdir -p /tmp/ownership-query
go build -o /tmp/ownership-query/profile ./cloud/reports/ownership-query-lint/testdata
/tmp/ownership-query/profile off after
/tmp/ownership-query/profile on after
go tool pprof -top /tmp/ownership-query/after-on-{0,1,2}.cpu
go tool pprof -peek='LowerWithOptions|OwnershipTransfers|QueryOwnershipTransfers|hasMoved' \
  /tmp/ownership-query/after-on-{0,1,2}.cpu
```

[Harness](testdata/profile.go). To compare the base commit, copy the same harness
into a checkout at `b10d3ca9`, build there, and run from this repository root with
label `before`; generate the same registry before either measurement. Stop builds
and tests before collecting timings.
