# Guarded integer counters, October 8, 2026

## Why the 5M sieve was refused

At parent `5212b966104f839f4673d9b01b0b95b5c661abc0`, the three loops
in `primes_5m.a` failed the static proof in `internal/lower/counters.go`:

| loop | first rejecting condition | remaining obstacles |
|---|---|---|
| `prime = 2; prime * prime <= limit; prime++` | `condition.Left.(ir.Read)` fails: the left operand is `ir.Binary{Multiply}` | the numeric parameter `limit` has no statically known whole range |
| `multiple = prime * prime; multiple <= limit; multiple += prime` | `boundRange(start)` fails: `prime` is unmarked, and its update makes `known.assigned[prime]` true | `neverNegativeZero` accepts multiplication only when `constantOf` succeeds; `stepOf` requires `constantOf(prime)`; `boundRange(limit)` has no declaration/range for the parameter |
| `index = 2; index <= limit; index++` | `boundRange(condition.Right)` fails: `limit` is neither a known counter nor a local with a declared whole value | its runtime value can be fractional, nonfinite, or allow an update beyond the exact integer range |

## Proof extension

The existing static proof runs first. A separate extension recognizes the same
direct comparisons and also a nonnegative counter squared on the left of `<`
or `<=`. Squaring is monotone on that nonnegative range. The body counter's
upper bound is the floor of the square root of the largest allowed bound;
its last update must still fit within ±2^53. The condition continues to evaluate
the original double product, including the final failing check.

Starts must already be proven whole and within ±2^53. Sums retain the existing
proof. A product of two nonnegative whole counters cannot produce -0; the
formerly refused products outside this added shape retain their refusal.
Steps may be whole loop-invariant scalar values, including an enclosing
counter. Zero, negative and fractional magnitudes retain the double path.
Subtraction supports descending loops with a positive magnitude.

Unknown bounds or steps read from invariant numeric locals (including
parameters and values written before the loop) use a loop-entry guard. It checks
integrality, finite range, positive step, and the last update's range. For
ascending `<=`, the bound must be at most `2^53 - step`; strict `<` instead uses
`2^53 - (step - 1)`. Descending comparisons use the symmetric lower limits.
The squared shape uses its smaller counter range to check the largest step.
The fallback is the original double block. Known bad constants, globals,
capture cells, property/call step expressions and numeric bounds or steps
written in the loop do not gain a specialization. The original static range
proof for changing lengths is retained where no entry snapshot is needed.

`ir.Local.CounterGuard` records the conditional proof. Native emission uses
scoped counter modes; it does not mutate the original IR to choose a branch.
Facts inherited from a guarded outer counter carry that dependency, so an
inner integer proof cannot leak into the outer double path. Nested guarded
specializations are suppressed inside a double fallback, preventing exponential
code duplication for independent nested parameter loops. Existing unconditional
integer counters remain unconditional.

## Validation

The original counter sweep still has **1,193 cases and 492 unconditional integer
counters**, with every expected classification unchanged. The new
`counters_shapes.a` fixture covers squared conditions, product and sum starts,
invariant counter/parameter/local steps, inclusive parameter bounds, array
length bounds, descending loops, and dependent and independent nested guards.
It includes fractional/nonfinite bounds and steps, steps outside ±2^53 and the
int64 range, a whole outer-counter step whose update crosses 2^53, starts outside
the safe product range, square edges near √(2^53), pre-loop local assignments,
changing bounds, and direct/closure changes to a step. Instrumented generated C
pins which calls actually execute integer declarations and which use doubles.

The non-invariant-step mutant forges the exact guard that admitting a changing
step would create. The initial step is 1, but the body changes it to 0.5. It
compiles with sanitizers and runs cleanly; Node catches the changed stdout.
The earlier bounds-check mutant still fails with ASan's heap-buffer-overflow.

Focused commands:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go test ./internal/oracle -count=1 -timeout 20m -v \
  -run '^(TestCounterShapesProofAndFallbacks|TestCounterNonInvariantStepMutant|TestLoopCountersAgreeWithNode)$'
go test ./internal/oracle -count=1 -timeout 20m -v \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(counters_shapes|typed_arrays_.*)[.]a$'
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -timeout 20m -v \
  -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/counters_shapes[.]a$'
go test ./internal/oracle -count=1 -timeout 20m -v -run '^TestTypedArray.*$'
go test ./internal/oracle -run TestCountsAreRecorded -timeout 30m -args -update-counts
```

The proof/step-mutant/counter sweep passed in 43.714 s. All 19 selected native
fixtures passed in 15.045 s, including all 18 existing typed-array fixtures.
The final shape fixture passed on WASI in 1.037 s. Counts are regenerated and
committed separately. No whole package test suite was run.

## 5M sieve measurement

Same machine `16bbf745127b`: shared Linux/amd64 container, Intel Xeon Platinum
8573C, 5 logical CPUs, cgroup quota `400000 100000` (4 CPUs), Linux 6.18.44,
Go 1.27.1, clang 20.1.8, Node v24.19.0. Run at 06:37:35 UTC. This is
`5212b966` plus this proof extension; the log prints the parent SHA because
measurement preceded the new commit. No setup, tests or counts ran alongside
timing. Load averages include earlier work.

```sh
go run ./bench -only primes_5m -rounds 5
```

Five fresh-process rounds interleave native and Node. The existing runner
reports minimum wall time, including startup/type stripping/checksum/output;
builds are excluded, and counted runs are separate and untimed. Native is the
release build (`clang -O2`). [Every sample](typed_array_integer.guarded.log).

| workload | native best | native peak RSS | Node best | Node peak RSS | native / Node |
|---|---:|---:|---:|---:|---:|
| `primes_5m` | **0.038 s** | 5.5 MB | **0.125 s** | 54.7 MB | 0.31x |

Both print `348513`. All three sieve counters now have guarded integer paths;
the 5,000,000 call passes their guards and uses integer typed-array accesses.
Counts: 3 allocations, 3 frees, 0 retains, 3 releases, peak live 2, regions 0.
Load before **2.02 / 1.39 / 1.02**, after **2.01 / 1.41 / 1.03** (1, 5, 15 min).

For context, the previous pass on this branch reported native 0.062 s and Node
0.180 s under a different load. Both runtimes improved in this pass, so the
native before/after difference alone is not a controlled speedup claim.

Workload SHA-256 (unchanged):
`986352065d17326c4e300a3e866facdf4133938bece124b25df493fe7f3c7990`.

The final dedicated typed-array pins/mutant pass took 0.940 s. Final counts
regeneration passed in 123.233 s and changes only the new shape fixture's row.
