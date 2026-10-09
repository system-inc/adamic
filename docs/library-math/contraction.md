# Math, parseInt and fused multiply-adds

Adamic never fuses a multiply and an add. A program prints the same bits on every platform, and those
bits are what JavaScript's rounding of each operation on its own gives: Node's answers on x86-64, and
fdlibm's as written. The flag is `-ffp-contract=off` in `native.Flags`, so it is in the runtime
library's cache key too, and `TestArithmeticIsNeverFused` fails on any processor with a fused
multiply-add (every arm64) if the flag goes.

## Node is not the same everywhere

Node on macOS arm64 (v24.14.1, and v24.19.0, the version Adamic pins; both V8 13.6.233.17) runs a V8 compiled with clang's default, `-ffp-contract=on`, which fuses a
multiply and an add within one expression. Checked on Darwin 27.0.0 with Apple clang 21 (#myatdyv); v24.19.0 gives v24.14.1's answer on every one of the 380,000 inputs below, and the sweeps, parseInt and navigation.a pass under it:

- V8's own `src/base/ieee754.cc` at 13.6.233.17, built alone with `-ffp-contract=on`, gives that
  Node's answers for 20,000 random inputs to each of sin, cos, tan, log, log10, log2, log1p, exp,
  expm1, asin, acos, atan, sinh, cosh, tanh and cbrt, every one; with `off`, it gives x86-64 Node's.
- `runtime/ieee754.c`, this port, built the same way with its `#pragma STDC FP_CONTRACT OFF` lifted,
  gives that Node's answers for all nineteen functions (asinh, acosh and atanh included), 380,000
  inputs, every one. Built with `fast`, which also fuses across statements, it misses a few (log2 in
  87 of 20,000), so `on` is what Node does, not just something close.
- parseInt("9007199254740993", 36), V8's HandleGenericCase loop as `runtime/parse.c` has it: with
  `fma()` 44fa555c722220b7, Node's; without, 44fa555c722220b5, Adamic's.

Across the bit-for-bit sweeps that is 1,910 of 1,281,787 Math answers and 4,598 of 108,023 parseInt
answers on that Mac, and none on Linux x86-64, which has no fused multiply-add to contract into.
Inference, unverified: Node on Linux arm64 contracts too, since GCC contracts by default outside strict
ISO mode.

## How the tests hold Adamic to such a Node

Linux x86-64 stays the gate of record and stays bit for bit. Elsewhere a difference from Node is
forgiven only when contraction alone explains it:

- `native.Options.FusedRuntime` (tests only) builds the runtime as that Node's V8 is built:
  `-ffp-contract=on`, with `ADAMIC_FUSED_RUNTIME` lifting the runtime's FP_CONTRACT OFF pragmas
  (ieee754.c, hypot.c, radix.c). The program is built without, since V8 never fuses JavaScript's own
  arithmetic.
- `TestIeee754MatchesNodeBitForBit` and `TestNumbersParseExactlyAsJavaScriptDoes` forgive an answer
  only when that build gives exactly Node's (`internal/native/fused_test.go`). The oracle forgives a
  fixture only when that build prints exactly Node's output (`internal/oracle/contraction_test.go`).
  On a machine that can't fuse the build is the unfused one, so nothing is forgiven there. A mistake
  in the runtime's source is in both builds, so it still fails.

With that model nothing on the Mac is left over: 1,910 Math and 4,598 parseInt answers differ from
Adamic's and all of them are that build's, and navigation.a prints that build's output. An earlier
version used `-ffp-contract=fast`, which left ten Math answers no build reproduced; they were the
difference between fusing within an expression and across statements, and `on` reproduces all ten.

The sweep in `TestIeee754MatchesNodeBitForBit` is reproducible: its branch points used to come out of a
map in a different order each run, so the seeded random draws, and the mismatches, moved between runs.

## Mutants run on macOS arm64

| Mutant | Caught by |
|---|---|
| `-ffp-contract=off` dropped from `Flags` | TestArithmeticIsNeverFused |
| FusedRuntime keeps the pragmas | the Math sweep, 1,910 unexplained |
| FusedRuntime with `fast` instead of `on` | the Math sweep, 10 unexplained |
| A forgiveness that ignores Node's answer, or compares with Adamic's | TestContractionForgivesOnlyWhatAFusedBuildReproduces |
| kernel_cos's C1 changed in its fifteenth significant digit | the Math sweep (3,014 unexplained) and navigation.a |
| parseInt chunked at 36 squared | the parseInt sweep (8,295 unexplained) |
| The oracle's explanation built without contraction | navigation.a |
