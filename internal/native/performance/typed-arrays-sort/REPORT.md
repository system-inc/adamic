# Typed array numeric sort for Workers

The measurement compares sorting one million deterministic numbers in a native
Float64Array, a native number[] with `(left, right) => left - right`, and Node's
Float64Array sort on the same AMD EPYC 9V74 box, Linux 6.18.44.
The native builds use ordinary Adamic release
options, clang 20.1.8 `-O2 -flto=thin`, without sanitizers or counting. Node is 24.19.0.

Input: LCG seed 42, `state = (state * 1664525 + 1013904223) >>> 0`, value
`state % 1000001 - 500000`. Every driver independently constructs the same input.
Native generated C is instrumented with `getrusage(RUSAGE_SELF)` immediately
around its emitted sort call. Node uses `process.cpuUsage()` around its sort.
Reported times are user CPU seconds for sorting, including its workspace, and
exclude process startup, input construction, validation and printing. No comparator
or sorting algorithm is replaced by the measurement driver.

All 21 runs are pinned with taskset to CPU 0, alternating forward/reverse order
for seven rounds. No package tests or builds ran alongside the timed rounds.
`nproc` is 5, the cgroup quota is four CPUs. The final load averages were
1.178, 1.026 and 0.658. Results describe this input and machine, not a guarantee
for Workers' route or every distribution.

| Driver | Median user seconds | Best user seconds |
|---|---:|---:|
| Native Float64Array.sort() | 0.118944 | 0.115743 |
| Native number[].sort(numeric comparator) | 0.206716 | 0.203938 |
| Node Float64Array.sort() | 0.112739 | 0.109607 |

The observed median native typed sort is 1.738 times faster than native number[]
sort and takes 1.055 times Node's user time. Native typed takes about 5.5 percent more median user time than Node on this
input; this is not a general parity claim.

Every run validates ascending order and prints the same length, endpoints and
integer checksum: `1000000 -500000 499997 319534362 ordered true`.

Run after sourcing the setup environment, with no competing test workload:

```sh
python3 internal/native/performance/typed-arrays-sort/measure.py /tmp/adamic-typed-sort-release-bench > /tmp/typed-sort-release-bench.log 2>&1
```

The script saves instrumented generated C, release binaries, the build log and
each round's stdout/stderr in that output directory. Add `--build-only` after
the directory to prepare artifacts without timing. The complete seven-round
JSON observations are copied into `observations.json` beside this report.

This benchmark contains finite integer-valued doubles with duplicates; NaN,
signed zero, infinities and view ranges are tested separately by the Node-held
oracle fixtures and C harness. Other sizes, sorted inputs, other distributions,
end-to-end Workers latency and peak memory are not measured here.
