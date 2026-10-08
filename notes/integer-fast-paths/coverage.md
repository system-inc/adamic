# Integer fast path coverage

Base: origin/codex/integer-fast-paths at a183e50. Reviewed both commit messages,
the merge-base diff, changed implementation and test files, benchmark sources,
and performance reports. Numeric semantics remain float64 between operations.

The table distinguishes existing source programs from the new focused programs.
The branch's integer_fast_paths.a has a dedicated oracle test, not a fixtures row.
Other existing files below are ordinary oracle fixtures. Names omit testdata/.

| Code case | Existing source coverage | Added coverage |
|---|---|---|
| ToUint32: magnitude strictly below 2^31, including signed zero and fractions of either sign | bitwise.a, bitwise_sweep.a, integer_fast_paths.a | calls, nested, loops, slots |
| ToUint32: boundary +/-2^31 and larger values through inclusive +/-2^53, including fractional truncation | same three sweeps | calls, nested, loops, slots |
| ToUint32 slow: finite magnitudes above 2^53, positive/negative modulo and zero modulo | same three sweeps | nested includes 2^53+2 |
| ToUint32 slow: NaN and +/-Infinity map to zero | same three sweeps | calls, nested, loops, slots |
| Remainder: both magnitudes <=2^53, divisor nonzero, both integer round trips succeed | integer_fast_paths.a, numbers.a; timsort.a has small recurrences | loops crosses int32 and safe-integer boundaries; remainder has compound edge sweep |
| Remainder: nonzero result takes dividend sign, independent of divisor sign | integer_fast_paths.a, numbers.a | calls and loops, remainder |
| Remainder: zero result takes dividend sign, including -0 | integer_fast_paths.a uses Object.is | remainder tests reciprocal as well as Object.is on compound slots |
| Remainder fallback: left/right outside range or nonfinite, divisor +/-0 | integer_fast_paths.a pairs all values | calls and loops; remainder compound +/-0, NaN and Infinity divisors |
| Remainder fallback: left fractional, right fractional, or both | integer_fast_paths.a; numbers.a only fractional dividend | loops mixes integers and fractions in both operands |
| Signed result: bits below, equal to, and above 0x80000000 | bitwise.a, bitwise_sweep.a, integer_fast_paths.a | nested, calls, loops |
| Shift count: low five bits, including 0, 31, 32, negative/fractional/nonfinite and huge counts | bitwise.a, bitwise_sweep.a | nested right expressions and loop-generated counts |
| Arithmetic right shift: sign unset/set, count zero/nonzero | all three sweeps | nested right operands and calls |
| Unsigned right shift: unsigned final double, including >=2^31 | all three sweeps | arithmetic after unsigned results, nested, calls |
| Pure unary ~ and pure binary &, \|, ^, <<, >>, >>> | all three sweeps | nested on both sides and right-nested shift counts |
| Direct bitwise children stay uint32; other children use ordinary double evaluation | integer_fast_paths.a has direct left nesting; bitwise.a has a few arithmetic and ~ barriers | nested has addition, subtraction, multiplication, division, remainder and conditional barriers |
| Impure unary ~ and all six binary operators use runtime helpers, with left-to-right snapshots | integer_fast_paths.a only has impure outer \|; bitwise_sweep.a has compound >>> side effects | calls exercises every operator with direct calls, plus captured closure and sibling mutations |
| Pure fallback leaves: locals/parameters/global reads, field/index/map/optional reads, Math/Number/char-code expressions | sweeps have variables; bitwise_sweep.a has field/index compound updates | slots combines reads as fused operands, present/absent optional numbers |
| Compound operators: local, field and array targets | bitwise_sweep.a covers all bitwise updates; integer_fast_paths.a local %=; updates.a field %= | remainder adds int32-edge field/array %= with mutating target/divisor calls; calls captured %= |
| Loop recurrences and integer/float arithmetic crossing representation boundaries | bitwise_sweep.a generates powers but no feedback from bitwise results; timsort.a has small integer remainder feedback | loops feeds bitwise and remainder results back through arithmetic in for and while loops |

Added files are integer_coverage_calls.a, integer_coverage_nested.a,
integer_coverage_loops.a, integer_coverage_slots.a and integer_coverage_remainder.a.
All print their individual results; signed zero is observed with Object.is and,
for remainder, reciprocal division. No checksum substitutes for result comparison.

No branch semantic condition was impossible to express. These programs do not
claim exhaustive combinations of arbitrary float64 values or every unrelated IR
leaf type. Non-number operands are rejected by the language, not handled by this
branch. Benchmarks and measurement script behavior are not compiler semantics.
