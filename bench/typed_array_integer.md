# Typed-array integer indices, October 8, 2026

The native backend now uses integer-index read and write helpers when the existing
counter proof marks the index local. The helpers retain the signed bounds check,
missing-read behavior and write panic. Writes use the existing value conversions
without testing an already-proven index for integrality again.

## Method and machine

Before: `56d53e10bf0b83c953e156687f60531fb17945fe` (the requested
`runtime/area-take-leftovers` base). After: that base plus this change. Both logs
print the base SHA because the after measurement preceded the implementation commit.

Machine `16bbf745127b`: shared Linux/amd64 container, Intel Xeon Platinum 8573C,
5 logical CPUs, cgroup quota `400000 100000` (4 CPUs), Linux 6.18.44,
Go 1.27.1, clang 20.1.8, Node v24.19.0. Setup and tests were not running alongside
timing. The load averages still include the preceding setup compilation.

Both passes used:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go run ./bench -only primes,primes_5m -rounds 5
```

The existing bench runner builds native release binaries with clang `-O2`, excludes
builds from elapsed time, and runs five fresh-process rounds interleaving native
and Node for each program. Times are the minimum wall-clock sample for each
runtime; startup, source type stripping on Node, checksumming and output are included.
Native and Node give the same answer in each reported row. Counted builds run
separately and untimed. All samples and peak memory are in the
[before log](typed_array_integer.before.log) and [after log](typed_array_integer.after.log).

| workload | native before | Node before | native after | Node after |
|---|---:|---:|---:|---:|
| `primes` | 0.240 s | 0.210 s | 0.280 s | 0.231 s |
| `primes_5m` | 0.055 s | 0.123 s | 0.062 s | 0.180 s |

Load (1, 5, 15 minutes): before pass `7.63 5.28 2.28` → `6.39 5.12 2.27`;
after pass `3.28 4.48 2.18` → `2.88 4.32 2.17`.

These measurements show no speedup. Generated C for both workloads still has two
double-index reads and one double-index write, and no integer-index access.
Their function-parameter bounds, nonlinear outer condition and variable inner
step do not qualify for the existing counter proof. Extending that proof is
outside this change. Cloud timing varies substantially even for Node; these
numbers do not establish a performance regression from the new integer path.

## Workloads

`primes.a` is copied verbatim from `runtime/honest-benchmarks`, commit `fb714279`:
ten fresh Uint8Array sieves at limits 2,000,000 through 2,000,009, summing prime
counts and values (stdout `1429153778578`). `primes_5m.a` uses the existing
`internal/oracle/testdata/typed_arrays_primes_large.a` sieve function unchanged,
with a single call at 5,000,000 replacing the small-limit driver (stdout `348513`).

SHA-256:

```text
ab84dee93ab656378f51a8a5b6fe7707bc66fd7dfe4895bba67ff884fd83d518  primes.a
986352065d17326c4e300a3e866facdf4133938bece124b25df493fe7f3c7990  primes_5m.a
```

## Focused validation

The oracle fixtures cover all four kinds, direct stores and conversion stores
(including 300 → 44 in Uint8Array), negative integer reads, reads at the upper
bound, and an integer write at the bound with the existing exit-70 panic pinned
independently against Node. The generated-C test requires eight integer reads
and four integer writes in the all-kinds fixture.

The bounds mutant copies the shipping integer read helper into the generated
unit and removes its check. The at-bound read must fail with ASan's
`heap-buffer-overflow`; a compilation failure is not accepted as a kill.

Passed commands (no full package test run):

```sh
go test ./internal/oracle -count=1 -timeout 20m -v \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/typed_arrays_.*[.]a$'
go test ./internal/oracle -count=1 -timeout 20m -v \
  -run '^(TestTypedArray.*|TestLoopCountersAgreeWithNode)$'
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -timeout 20m -v \
  -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/typed_arrays_integer(_stop)?[.]a$'
```

All 18 typed-array fixtures passed in 10.244 s. The dedicated tests passed in
32.897 s, including the mutant, both write-panic pins, the platform statistics
reference and the 1,193-case counter sweep (492 counters kept in integers).
Both new WASI fixtures passed in 7.953 s.

`go test ./internal/oracle -run TestCountsAreRecorded -timeout 30m -args -update-counts`
also passed (119.697 s); the regenerated `internal/oracle/counts.md` is committed
separately from the implementation and fixtures.
