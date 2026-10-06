# Math, parseInt and fused multiply-adds

Adamic never fuses a multiply and an add. A program prints the same bits on every platform, and those
bits are what JavaScript's rounding of each operation on its own gives: Node's answers on x86-64, and
fdlibm's as written. The flag is `-ffp-contract=off` in `native.Flags`, so it is in the runtime
library's cache key too, and `TestArithmeticIsNeverFused` fails on any processor with a fused
multiply-add (every arm64) if the flag goes.

## Node is not the same everywhere

Node v24.14.1 on macOS arm64 runs a V8 compiled with multiply-adds contracted. Checked on Darwin 27.0.0
with Apple clang 21 (#myatdyv):

- `runtime/ieee754.c` (the port of V8's ieee754.cc) built alone with `-ffp-contract=fast` gives that
  Node's Math.cos, Math.tan and Math.acosh bit for bit where the unfused build, and Adamic, differ from
  it in the last bit (cos of 0x7feffffffffffffe: Node 3fd7ffdfb4c5308f, Adamic 3fd7ffdfb4c53090).
- parseInt("9007199254740993", 36), V8's HandleGenericCase loop as `runtime/parse.c` has it: with
  `fma()` 44fa555c722220b7, Node's; without, 44fa555c722220b5, Adamic's.

Across the bit-for-bit sweeps that is 1,910 of 1,281,787 Math answers and 4,598 of 108,023 parseInt
answers on that Mac, and none on Linux x86-64, which has no fused multiply-add to contract into.
Inference, unverified: Node on Linux arm64 contracts too, since GCC contracts by default outside strict
ISO mode.

## How the tests hold Adamic to such a Node

Linux x86-64 stays the gate of record and stays bit for bit. Elsewhere a difference from Node is
forgiven only when contraction alone explains it:

- `native.Options.FusedRuntime` (tests only) builds the runtime with `-ffp-contract=fast` and the
  program without, since V8 never fuses JavaScript's own arithmetic.
- `TestIeee754MatchesNodeBitForBit` and `TestNumbersParseExactlyAsJavaScriptDoes` forgive an answer
  only when that build gives exactly Node's (`internal/native/fused_test.go`). The oracle forgives a
  fixture only when that build prints exactly Node's output (`internal/oracle/contraction_test.go`).
  On a machine that can't fuse the build is the unfused one, so nothing is forgiven there. A mistake
  in the runtime's source is in both builds, so it still fails.

On that Mac, ten Math answers (log, log10, asinh and tan, each one ulp from Adamic's) differ in a way
no build of the port reproduces, with on or fast, -O1 to -O3, or any -mcpu. They are left failing,
not recorded as exceptions: a list of forgiven answers would be a rule of its own. Inference,
unverified: Node's clang, a different version from this one, chose to fuse the other product in an
expression with two (log's `s * (hfsq + R) + dk * ln2_lo`). So the Math sweep is clean on Linux
x86-64, the gate of record, and reports exactly these ten on macOS arm64.

The sweep in `TestIeee754MatchesNodeBitForBit` is reproducible: its branch points used to come out of a
map in a different order each run, so the seeded random draws, and the mismatches, moved between runs.

## Mutants run on macOS arm64

| Mutant | Caught by |
|---|---|
| `-ffp-contract=off` dropped from `Flags` | TestArithmeticIsNeverFused |
| FusedRuntime builds without contraction | both sweeps, 1,900 and 4,598 unexplained |
| A forgiveness that ignores Node's answer, or compares with Adamic's | TestContractionForgivesOnlyWhatAFusedBuildReproduces |
| kernel_cos's C1 changed in its fifteenth significant digit | the Math sweep (3,014 unexplained) and navigation.a |
| parseInt chunked at 36 squared | the parseInt sweep (8,295 unexplained) |
| The oracle's explanation built without contraction | navigation.a |
