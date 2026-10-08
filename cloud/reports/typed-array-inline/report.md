# Typed-array sieve baseline

Stopped at the user's instruction on October 8, 2026. Runtime owns the
replacement, runtime/typed-array-inline b53bed0d. This commit contains
measurement notes only. No emitter, runtime, fixture or dependency merge is
included. No measurements of b53bed0d were made.

## Observed release timings

Milliseconds per sieve, best of five batches of twenty sequential sieves.
Each batch starts a process; wall time includes startup, allocation,
initialization, marking, counting, cleanup and one checksum print. Types
alternate within each batch. Native release is the compiler's default clang
-O2, with -ffp-contract=off. These are Linux native results, not Wasm results.

| Limit | number[] | Uint8Array |
| --- | ---: | ---: |
| 1,000,000 | 13.148456 | 36.600217 |
| 5,000,000 | 99.830851 | 187.873517 |

Every baseline batch printed 1569960 at 1M and 6970260 at 5M, namely twenty
times 78498 and 348513 primes. These checksums agree between representations.
Node oracle tests were not run for this measurement-only work.

## Machine and baseline

Container 4af0c50d0cb2, Linux 6.18.44 x86_64, KVM, AMD EPYC 9V74 80-Core
Processor. nproc reports 5; cgroup cpu.max is 400000 100000 (four CPU quota).
Go 1.27.1, clang 20.1.8, Node 24.19.0.

Fetched origin/main f4efdd2369311d1420aa53fdf5c1a55bdda811d4 did not contain
typed arrays. The measured baseline was the local dependency merge
616283de of that main with origin/codex/typed-arrays-uint16
04f18a2a78ab8f7576e2c92deefa5a924099dad1. Its conflicts were resolved by
preserving main's cast proof and both sets of oracle fixtures. A missing
representation guard from the dependency was retained in castProof. That
merge is not included in this measurement branch.

Setup timings: Node 0.020s; Go 0.021s; markdown ready 0.062s; submodules
0.063s; clang 0.158s; go build 32.202s; warm build cache 32.380s; total
32.408s. Full setup output is in setup.txt.

## Sampled profile

A separate 5M batch was compiled with -O2 -pg and linked against the unchanged
release runtime archive. gprof's 10ms sampling attributes 44.15% to
adamic_typed_array_set, 11.11% to adamic_typed_array_get, 6.14% to
adamic_typed_array_length, and 1.17% to adamic_typed_array_check_write.
The remaining largest entry is main at 28.07%. The number[] profile attributes
47.99% to adamic_array_set, 33.67% to main, and 12.06% to
adamic_array_filled. Runtime routines were sampled but not instrumented for
call counts; blank call-count columns must not be read as zero calls.

Observed generated C calls the typed-array get, set and length functions.
Plain array writes also call adamic_array_set. The read/write/length call
cost and typed conversion are plausible contributors; this profile does not
isolate conversion cost from checking, dispatch, or call overhead.

## Reproducer and commands

The sieve below was written for this unit. For number[], replace the allocation
with new Array<number>(limit + 1).fill(0). For 5M, replace 1000000 with 5000000.

```typescript
function countPrimes(limit: number): number {
 const composite = new Uint8Array(limit + 1);
 for (let prime = 2; prime * prime <= limit; prime++) {
  if (composite[prime] !== 1) {
   for (let multiple = prime * prime; multiple < composite.length; multiple += prime) {
    composite[multiple] = 1;
   }
  }
 }
 let count = 0;
 for (let index = 2; index < composite.length; index++) {
  if (composite[index] !== 1) { count++; }
 }
 return count;
}
let total = 0;
for (let batch = 0; batch < 20; batch++) { total += countPrimes(1000000); }
console.log(`${total}`);
```

Build each form with:

```bash
source /workspace/adamic-tools/env.sh
go build ./cmd/adamic
./adamic build <sieve.a> -o <binary>
./adamic c <sieve.a> > <generated.c>
```

For five batches, alternate representations, run each binary once, assert its
exit code and checksum, and record (time.perf_counter() elapsed) * 1000 / 20.
The complete twenty observations are in baseline.json.

For the separate profile, compile generated C with clang -O2 -pg
-ffp-contract=off, -I <cached-runtime-directory>, link runtime.a using
-Wl,--whole-archive and -Wl,--no-whole-archive, and -lm. Run the binary in its
own directory and run gprof <binary> gmon.out > profile.txt. Complete sampled
profiles are included.

## Limits and stopped work

An emitter prototype was built locally before the stop. Its follow-up
measurement was interrupted; it is not a validated result and is omitted
from these baseline numbers. All uncommitted emitter changes were discarded.
No compiler changes are published. No new mutants, typed-array oracle gate,
Wasm benchmark, or measurement of runtime's inline implementation was run.
The performance unit is cancelled; await the next instruction.
