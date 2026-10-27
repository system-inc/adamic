# Runtime invariants and GCC

The runtime remains C11. The checks below expose existing preconditions to
GCC without assuming undefined behavior or changing JavaScript arithmetic.
`adamic_invariant(what)` is a real panic with exit status 70, not a compiler
intrinsic promising that invalid state cannot happen. It uses the ordinary panic
path to flush pending output and prints exactly `adamic: panic: runtime invariant
broken: <what>` followed by a newline. The guards name `bignum digit count below
zero`, `argument reduction left no terms`, and `map iteration of a missing map`.
`adamic_unreachable` retains its existing function-return diagnostic.

## Decimal digit counts

`bignum_assign_bignum` follows V8 13.6.233.17's
[`AssignBignum`, `Zero`, and `Clamp`](https://github.com/v8/v8/blob/13.6.233.17/src/base/numbers/bignum.cc).
Counts start at zero. Assignment and growth preserve nonnegative counts, and
Clamp decrements only while the count is positive. Its clearing loop therefore
starts at a nonnegative index. An explicit negative-count stop makes that
precondition visible locally, including after inlining.

`TestGCCInvariantDTOA` exercises a valid copy before injecting a negative count.
Its guard-removal mutant reaches the negative array index under UBSan.
The four `TestToExponentialAndToPrecisionMatchNodeValues*` ranges check valid
conversions against Node; `TestToExponentialAndToPrecisionValueRangesTile` holds
their complete partition.
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
production optimization and stack-call flags. It removes blanket `-Werror` and
all inherited `-Werror=...` options, then promotes only array bounds,
maybe-uninitialized reads, destination overflow, and infinite recursion to errors.
Every other warning is retained verbatim in each unit's `diagnostics` product and
logged by its reader without failing the audit. Compiler-only regex bodies, the
checker bridge and its keyed ABI header, all emitter features, counted paths and
Program regions are enabled. There are 76 explicitly owned C files.
GCC's lack of a warning is compiler-version dependent; it is not proof that
all recursion terminates. No recursion change is justified without a
reproducing site. `TestGCCDiagnosticRecursionMutant` injects unconditional
self-recursion into the real sine implementation and requires GCC to reject
it with `-Winfinite-recursion`.

## Floating-point contraction

GCC ignores `#pragma STDC FP_CONTRACT OFF` in hypot.c, ieee754.c, and radix.c.
The audit records those ignored-pragma warnings without promoting them to
errors; it retains the pragmas for compilers that support them. Production `Flags` already supplies
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

## Platform and door policy

The shared GNU compiler helper skips only off Linux, with the fixed reason
`GNU gcc audits run on Linux; macOS gcc is clang`. On Linux it follows the
existing portability tests' executable lookup and version query, and fails hard
if gcc is absent or its version does not name the Free Software Foundation.
The GCC arithmetic witness reuses `fusingProcessor()` and `t.Skip(why)` exactly
as `TestArithmeticIsNeverFused`, including its existing capability reasons.
The exact new skip-census rows, generated by devtools/fast-gate's scanner,
are in [runtime-gcc-skips.json](runtime-gcc-skips.json) for classification.
They declare no classes here: Runtime routes the rows to the census owners.

`fusion_test.go` holds operation-specific fusion lists and oracle policy;
`TestArithmeticIsNeverFused` holds observable arithmetic, and
`TestReleaseFixtureKeepsArithmeticUnfused` holds ThinLTO link flags. None of them
inspects the effective contraction option for every runtime compilation, so the
three source-aware archive audits remain. The release profile uses main's
`Release: true` ThinLTO flags; the other profiles use ASan/UBSan and counting.
All optional runtime C files and the checker ABI header are included in each.
Archive compilation takes main's explicit jobserver and keeps its existing keys.

## Advisory warnings on GNU GCC 14.2

With all native feature bodies enabled, the full audit retains these six warnings:

| File and line | GCC warning |
| --- | --- |
| `hypot.c:19` | `-Wunknown-pragmas`: ignored `STDC FP_CONTRACT` |
| `ieee754.c:40` | `-Wunknown-pragmas`: ignored `STDC FP_CONTRACT` |
| `radix.c:27` | `-Wunknown-pragmas`: ignored `STDC FP_CONTRACT` |
| `regexp_compile_bytecode.c:155` | `-Wmisleading-indentation` |
| `regexp_compile_runtime.c:21` | `-Wmisleading-indentation` |
| `regexp_compile_runtime.c:26` | `-Wmisleading-indentation` |

The compact regex statements retain their original meaning and layout. These
advisories are retained for their owners; no production warning policy or runtime
code outside the three guards and the invariant stop is changed.

The existing GNU long-Unicode ASan portability test needs one current-toolchain
reconciliation: GCC 14 recognizes `__has_feature(address_sanitizer)` but does not
provide `sanitizer/allocator_interface.h`. Its test flags now add the installed
Clang resource headers with `-idirafter`, after GCC's own headers. Both allocator
APIs are exported by GNU libasan. This supplies declarations without replacing
GCC's standard or intrinsic headers, changing production flags, or editing
`host_runtime.c`. The include flag and read headers enter its existing runtime
cache key. The GCC audits themselves do not require that sanitizer include path.
