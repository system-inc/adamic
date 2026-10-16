# Runtime invariants and GCC

The runtime remains C11. The checks below expose existing preconditions to
GCC without assuming undefined behavior or changing JavaScript arithmetic.
`adamic_unreachable()` is a real panic with exit status 70, not a compiler
intrinsic promising that invalid state cannot happen.

## Decimal digit counts

`bignum_assign_bignum` follows V8 13.6.233.17's
[`AssignBignum`, `Zero`, and `Clamp`](https://github.com/v8/v8/blob/13.6.233.17/src/base/numbers/bignum.cc).
Counts start at zero. Assignment and growth preserve nonnegative counts, and
Clamp decrements only while the count is positive. Its clearing loop therefore
starts at a nonnegative index. An explicit negative-count stop makes that
precondition visible locally, including after inlining.

`TestGCCInvariantDTOA` exercises a valid copy before injecting a negative count.
Its guard-removal mutant reaches the negative array index under UBSan.
`TestToExponentialAndToPrecisionMatchNode` checks valid conversions against Node.
The independent GCC dtoa audit rejects the removed guard as an array-bounds
error.

## Trigonometric argument reduction

The reduction follows fdlibm's
[`__kernel_rem_pio2`](https://netlib.org/fdlibm/k_rem_pio2.c), via V8's port.
Initially `jz = jk`, with `jk` positive. If the distilled remainder is zero,
recomputation continues until some `iq` term at or above `jk` is nonzero.
Chopping zero terms consequently cannot remove every term. Otherwise `jz`
stays positive or grows. The conversion and product loops therefore initialize
`fq[0]`; testing `jz < 0` before them makes the implication explicit to GCC.

`TestGCCInvariantReduction` injects negative `jz` immediately before that stop
in a private copy. Removing the guard makes this copy return a result instead
of stopping. Only that fault-injection copy initializes `fq[0]` to a sentinel,
so the mutant's behavior is deterministic. Production still computes every
term. `TestIeee754MatchesNodeBitForBit` checks valid inputs, including huge
arguments and hard reduction cases, against Node. A separate GCC audit rejects
the guard's removal as a maybe-uninitialized error.

## Map iteration

Retaining NULL is defined for undefined values; iterating a NULL Map is not.
The null guard precedes allocation and the atomic iterator tally update.
`iterating` remains atomic and its ordering is unchanged. The GCC Map audit
rejects the removed guard's null-destination write. `TestGCCInvariantMap` first
creates, exhausts, and releases a real iterator, then exercises the stop; its
mutant reaches a UBSan null access.

## Recursion and compiler coverage

Each runtime C file, including optional-feature files, has its own
`TestProduct_GCCRuntime*` and `TestGCCRuntime*` owner. The inventory test rejects
an unowned addition. The audit enables all runtime features, preserves the
production optimization and stack-call flags, and treats array bounds,
uninitialized reads, destination overflow, and infinite recursion as errors.
GCC's lack of a warning is compiler-version dependent; it is not proof that
all recursion terminates. No recursion change is justified without a
reproducing site. `TestGCCDiagnosticRecursionMutant` injects unconditional
self-recursion into the real sine implementation and requires GCC to reject
it with `-Winfinite-recursion`.

## Floating-point contraction

GCC ignores `#pragma STDC FP_CONTRACT OFF` in hypot.c, ieee754.c, and radix.c.
The audit suppresses only the ignored-pragma warning; it retains the pragmas
for compilers that support them. Production `Flags` already supplies
`-ffp-contract=off`. `TestRuntimeCompilesEveryUnitUnfusedRelease`,
`TestRuntimeCompilesEveryUnitUnfusedSanitized`, and
`TestRuntimeCompilesEveryUnitUnfusedCounted` inspect the actual compiler
invocation for every runtime C file, including optional features. They require
the effective contraction option to be `off`, with no missing or duplicate
units. Each profile's archive is built once as a source-aware product.
`TestRuntimeUnfusedOmissionMutant` checks an actual dtoa compilation with that
option removed, reusing all other objects, and requires the same audit to
reject it. `TestArithmeticIsNeverFused` additionally checks observable arithmetic
on a processor supporting FMA. `TestGCCArithmeticUnfused` holds the same
arithmetic to Node using GCC itself; `TestGCCArithmeticContractionMutant`
requires allowing contraction to change the answer. Both use their own small
ahead-built product, independent of runtime archive construction.
