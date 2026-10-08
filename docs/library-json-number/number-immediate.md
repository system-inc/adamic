# Immediate boxed Number receivers

An immediate intrinsic `new Number(...)` is used only for its Number internal
slot by `toString`, `toFixed`, `toExponential`, `toPrecision` and `valueOf`,
including an explicit prototype method `.call` receiver. Object identity never
escapes. General boxed objects, structural views, object coercion, spread and
extra constructor arguments stay refused. Conversion, radix output and decimal
formatting reuse the existing runtime without algorithm changes.

The separate validity runner's interim Number measurement is 190 pass, zero
disagreements, 32 refused, 44 not-typescript, zero crashed, 74 skipped. All 38
new passes execute against Node. The oracle fixture also compares binary64
boundaries, negative zero, missing/undefined/null/boolean/string constructor
arguments, radix 2/16/36, all decimal formatters and argument side effects.

The full uncached native/Node/backend/release/leak fixture passed in 0.380s.
The allocation-count gate passed in 9.629s: 234 alloc, 234 free, 43 retain,
271 release, peak 6, region 0. Both commands wrote their complete output to
`/tmp/library-json-number-number-final.log` and
`/tmp/library-json-number-counts-number.log`.

The required exponent-boundary mutant changes the existing formatter's two
`point <= 21` comparisons to `point <= 20`. Both programs compile, exit zero,
have empty stderr and remain leak-clean; only Node's stdout comparison catches
native printing `1e+20` instead of `100000000000000000000` (9.006s).
Log: `/tmp/library-json-number-exponent-mutant.log`. The runtime was restored.

An additional pinning wrapper was tried and its removal mutant survived the
fixture: the existing emitter already preserves the observable evaluation
order. That unnecessary wrapper and its claim were removed. This surviving
experiment is not counted as a successful mutant.
