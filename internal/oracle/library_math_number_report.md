# Math and Number library slice

Branch: `codex/library-math-number`. Main: `fe3b9f2`. Runner merged by fast-forward: `bb766d42ba415def680c8223bc7825e228141a9a`.

## Observed test262 results

Corpus: `tc39/test262` at `7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`. Every pass compares the native result with Node 24.19.0; the runner uses ASan and UBSan.

| Run | Directory | Pass | Disagree | Refused | Crash | Skip | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| Before | built-ins/Math | 83 | 0 | 83 | 0 | 161 | 327 |
| Before | built-ins/Number | 61 | 0 | 205 | 0 | 74 | 340 |
| After | built-ins/Math | 100 | 0 | 66 | 0 | 161 | 327 |
| After | built-ins/Number | 152 | 0 | 114 | 0 | 74 | 340 |

108 newly passing tests; no previous pass was lost. Zero disagreements and crashes before and after.

## Built

- Math.clz32, Math.imul and Math.fround: V8 13.6.233.17 conversion algorithms, with unsigned shifts and multiplication to avoid C signed overflow. Float32 overflow follows V8's exact threshold guards. These Math routines do not call libm.
- LN10, LN2, LOG10E, LOG2E, SQRT1_2 and SQRT2 as exact binary64 hexadecimal constants.
- Number() and Number(value) for proven primitive numbers, booleans, strings and undefined; literal null; primitive unions and optional primitive values. Unary plus uses the same conversion. Strings require complete numeric syntax, including Unicode trimming, unsigned binary/octal/hexadecimal literals, Infinity and signed zero. Decimal and power-of-two rounding reuse the existing Node-tested parsers.
- Number.prototype valueOf, toString, toFixed, toExponential and toPrecision directly, and through .call with a numeric receiver. Number.prototype's own numeric data is +0.
- Number and Number.prototype hasOwnProperty with string keys, Number.length, Number.name and typeof Number.

New code is isolated in `internal/lower/library_math_number.go`, `internal/native/library_math_number.go` and `internal/native/runtime/library_math_number.c`. Existing dispatch, IR return typing and the JavaScript backend have small hooks. V8 attribution is in THIRD_PARTY_NOTICES.md.

Math.sign, Math.cbrt and the other transcendental functions were already implemented on main. The measured remaining Math refusals contain no missing supported numeric Math operation. Existing V8 ports remain unchanged.

## Remaining refusals and coverage limits

**built-ins/Math**

| Reason | Count |
|---|---:|
| refuses var | 50 |
| error TS2345: Argument of type '…' is not assignable to parameter of type '…'. | 9 |
| error TS2550: Property '…' does not exist on type '…'. Do you need to change your target library? Try changing the '…' compiler option to '…' or later. | 4 |
| error TS18048: '…' is possibly '…'. | 1 |
| refuses a method read as a value (max would lose its object, and this with it) | 1 |
| refuses a method read as a value (min would lose its object, and this with it) | 1 |

**built-ins/Number**

| Reason | Count |
|---|---:|
| not yet: new an Identifier | 39 |
| error TS2345: Argument of type '…' is not assignable to parameter of type '…'. | 24 |
| refuses var | 12 |
| error TS2551: Property '…' does not exist on type '…'. Did you mean '…'? | 5 |
| error TS7006: Parameter '…' implicitly has an '…' type. | 5 |
| error TS2339: Property '…' does not exist on type '…'. | 3 |
| error TS2790: The operand of a '…' operator must be optional. | 3 |
| error TS2322: Type '…' is not assignable to type '…'. | 2 |
| not yet: a try around toString, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | 2 |
| not yet: reading Function | 2 |
| not yet: reading isFinite | 2 |
| refuses a method read as a value (toString would lose its object, and this with it) | 2 |
| refuses for...in | 2 |
| error TS2403: Subsequent variable declarations must have the same type.  Variable '…' must be of type '…', but here has type '…'. | 1 |
| error TS2554: Expected 0 arguments, but got 1. | 1 |
| not yet: a BinaryExpression with a string and a number | 1 |
| not yet: a try around toExponential, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | 1 |
| not yet: a try around toFixed, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | 1 |
| not yet: a try around toPrecision, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | 1 |
| not yet: a value argument to toExponential | 1 |
| not yet: reading Object | 1 |
| refuses == | 1 |
| refuses a method read as a value (toFixed would lose its object, and this with it) | 1 |
| refuses the void operator | 1 |

Across both directories, `var` is the biggest refusal reason (62 tests). The largest remaining library gap is boxed `new Number(...)` (39 tests). A boxed Number requires object identity, [[NumberData]], prototype behavior and memory ownership; lowering it to a primitive would silently change behavior. It remains refused.

Not covered: arbitrary first-class Number constructor aliases/identity, boxed Number objects, object coercion through user valueOf/toString, nonnumeric prototype receivers, detached methods, prototype toLocaleString, spread/extra arguments, dynamic mutation of built-in objects, and catchable native range errors in numeric formatting. Checker and language refusals were preserved. Property-helper, constructor-harness and other runner skips are not conformance claims.

## Oracle fixtures and counts

Registered in internal/oracle/oracle_test.go:

- library_math_number_math.a: integer wrapping, large exponents, NaN/infinities, float32 overflow/underflow, signed zero, all six constants; 24 input values and their pairwise imul products.
- library_math_number_convert.a: strict string grammar, Unicode space, embedded NUL, invalid suffixes and separators, base literals, optional values and primitive unions, metadata and own properties.
- library_math_number_prototype.a: direct prototype calls, numeric .call receivers, formatting edges and signed zero.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/library_math_number_math.a | 1280 | 1280 | 50 | 1331 | 8 | 0 |
| internal/oracle/testdata/library_math_number_convert.a | 199 | 199 | 197 | 365 | 7 | 0 |
| internal/oracle/testdata/library_math_number_prototype.a | 71 | 71 | 2 | 74 | 8 | 0 |

## Mutants run

`TestMathNumberOracleCatchesMutants` builds each mutant under ASan/UBSan and requires exit 0 with empty stderr. Each was caught only by comparing its stdout with the original source on Node.

| Family | Mutation | Caught by |
|---|---|---|
| clz32 | Emit fround instead | Node stdout differs |
| fround | Emit sign instead | Node stdout differs |
| imul | Emit pow instead | Node stdout differs |
| constants | Change LN10 by one binary64 ULP | Node stdout differs |
| conversion | Parse only the decimal prefix instead of the complete string | Node stdout differs |
| own properties | Invert hasOwnProperty | Node stdout differs |
| prototype formatting | Add one to the toFixed receiver | Node stdout differs |

## Commands and outputs

All test output was redirected to log files. Commands below were run after sourcing `/workspace/adamic-tools/env.sh`.

```text
bash cloud/setup.sh > /tmp/adamic-setup-math-number.log 2>&1
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
v24.19.0
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (22s)
setup: done in 22s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

The normal origin fetch configuration did not include the runner branch. It was fetched explicitly:

```text
git fetch origin
git checkout -b codex/library-math-number origin/main
git fetch origin cloud/grok-test262-runner:refs/remotes/origin/cloud/grok-test262-runner
git merge origin/cloud/grok-test262-runner
Result: fast-forward to bb766d4
```

```sh
go run ./cmd/adamic-test262 -json -test262 /tmp/adamic-test262-corpus -work /tmp/adamic-math-number-before built-ins/Math built-ins/Number > /tmp/adamic-math-number-before.json 2> /tmp/adamic-math-number-before.log
go run ./cmd/adamic-test262 -json -test262 /tmp/adamic-test262-corpus -work /tmp/adamic-math-number-final built-ins/Math built-ins/Number > /tmp/adamic-math-number-final.json 2> /tmp/adamic-math-number-final.log
```

Outputs are the before/after tables above. An intermediate run before metadata additions measured Math 100 pass, Number 144 pass, with zero disagreements.

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_math_number|TestMathNumberOracleCatchesMutants' -count=1 -v > /tmp/adamic-math-number-verified.log 2>&1
```

Output: PASS, all 3 fixtures and all 7 mutants; oracle package 71.762s. Fixtures compare source Node, JavaScript backend, sanitized native and release native, and check leaks.

```sh
go vet ./... > /tmp/adamic-math-number-vet-final.log 2>&1
go test ./internal/lower ./internal/native ./internal/ir ./internal/javascript ./cmd/adamic-test262 -run 'TestIeee754MatchesNodeBitForBit|TestMathAndToFixedMatchJavaScript|Test' -count=1 -timeout=30m > /tmp/adamic-math-number-packages-verified.log 2>&1
```

```text
go vet ./...: exit 0, no diagnostics
ok  	github.com/system-inc/adamic/internal/lower	7.042s
ok  	github.com/system-inc/adamic/internal/native	116.181s
?   	github.com/system-inc/adamic/internal/ir	[no test files]
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/cmd/adamic-test262	34.608s
```

The final package command's `|Test` alternative selects all tests in those packages. The whole-repository go test gate was not run; the native package alone took 267.480s in the preceding package run. Final go vet covers all packages.

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts > /tmp/adamic-math-number-counts.log 2>&1
go test ./internal/oracle -run 'TestCountsAreRecorded/fixtures/internal/oracle/testdata/library_math_number' -count=1 -v -timeout=30m -args -update-counts > /tmp/adamic-math-number-counts-final.log 2>&1
```

The first broad count run used a compiler built before the metadata addition, then read the updated fixture source. It refused Number.length and failed after 449.174s without writing counts. The final filtered generator passed all 3 fixtures (12.078s); its measured rows were merged into the preserved full table in fixture registration order. The final conversion row was also checked with `adamic build ... --count` and an 8 MiB stack: allocations 199, frees 199, retains 197, releases 365, peak 7, regions 0.

During implementation, an initial native build caught a missing fourth string-slice argument; it was corrected. A later oracle run caught the omitted JavaScript-backend prototypeHasOwnProperty dispatch; it was corrected before the final passing run. An initial prototype mutant used an incompatible C signature; the final mutant changes the receiver and must execute cleanly. An initial fixture filter matched no subtests; the final filter explicitly ran all 3 fixtures.

## Newly passing test262 tests

```text
built-ins/Math/LN10/value.js
built-ins/Math/LN2/value.js
built-ins/Math/LOG10E/value.js
built-ins/Math/LOG2E/value.js
built-ins/Math/SQRT1_2/value.js
built-ins/Math/SQRT2/value.js
built-ins/Math/clz32/Math.clz32.js
built-ins/Math/clz32/Math.clz32_1.js
built-ins/Math/clz32/Math.clz32_2.js
built-ins/Math/clz32/infinity.js
built-ins/Math/clz32/int32bit.js
built-ins/Math/clz32/nan.js
built-ins/Math/fround/Math.fround_Infinity.js
built-ins/Math/fround/Math.fround_NaN.js
built-ins/Math/fround/Math.fround_Zero.js
built-ins/Math/fround/value-convertion.js
built-ins/Math/imul/results.js
built-ins/Number/S15.7.1.1_A2.js
built-ins/Number/S15.7.3_A1.js
built-ins/Number/S15.7.3_A2.js
built-ins/Number/S15.7.3_A3.js
built-ins/Number/S15.7.3_A4.js
built-ins/Number/S15.7.3_A5.js
built-ins/Number/S15.7.3_A6.js
built-ins/Number/S15.7.3_A8.js
built-ins/Number/S9.3.1_A1.js
built-ins/Number/S9.3.1_A10.js
built-ins/Number/S9.3.1_A11.js
built-ins/Number/S9.3.1_A12.js
built-ins/Number/S9.3.1_A13.js
built-ins/Number/S9.3.1_A14.js
built-ins/Number/S9.3.1_A15.js
built-ins/Number/S9.3.1_A16.js
built-ins/Number/S9.3.1_A17.js
built-ins/Number/S9.3.1_A18.js
built-ins/Number/S9.3.1_A19.js
built-ins/Number/S9.3.1_A2.js
built-ins/Number/S9.3.1_A20.js
built-ins/Number/S9.3.1_A21.js
built-ins/Number/S9.3.1_A22.js
built-ins/Number/S9.3.1_A23.js
built-ins/Number/S9.3.1_A24.js
built-ins/Number/S9.3.1_A25.js
built-ins/Number/S9.3.1_A26.js
built-ins/Number/S9.3.1_A27.js
built-ins/Number/S9.3.1_A28.js
built-ins/Number/S9.3.1_A29.js
built-ins/Number/S9.3.1_A30.js
built-ins/Number/S9.3.1_A31.js
built-ins/Number/S9.3.1_A32.js
built-ins/Number/S9.3.1_A3_T1.js
built-ins/Number/S9.3.1_A3_T1_U180E.js
built-ins/Number/S9.3.1_A4_T1.js
built-ins/Number/S9.3.1_A5_T1.js
built-ins/Number/S9.3.1_A5_T2.js
built-ins/Number/S9.3.1_A6_T1.js
built-ins/Number/S9.3.1_A7.js
built-ins/Number/S9.3.1_A8.js
built-ins/Number/S9.3.1_A9.js
built-ins/Number/S9.3_A1_T1.js
built-ins/Number/S9.3_A2_T1.js
built-ins/Number/S9.3_A3_T1.js
built-ins/Number/S9.3_A4.1_T1.js
built-ins/Number/S9.3_A4.2_T1.js
built-ins/Number/prototype/S15.7.4_A3.1.js
built-ins/Number/prototype/S15.7.4_A3.2.js
built-ins/Number/prototype/S15.7.4_A3.3.js
built-ins/Number/prototype/S15.7.4_A3.4.js
built-ins/Number/prototype/S15.7.4_A3.5.js
built-ins/Number/prototype/S15.7.4_A3.6.js
built-ins/Number/prototype/S15.7.4_A3.7.js
built-ins/Number/prototype/toExponential/this-is-0-fractiondigits-is-0.js
built-ins/Number/prototype/toPrecision/this-is-0-precision-is-1.js
built-ins/Number/string-binary-literal-invalid.js
built-ins/Number/string-binary-literal.js
built-ins/Number/string-hex-literal-invalid.js
built-ins/Number/string-numeric-separator-literal-bil-bd-nsl-bd.js
built-ins/Number/string-numeric-separator-literal-bil-bd-nsl-bds.js
built-ins/Number/string-numeric-separator-literal-bil-bds-nsl-bd.js
built-ins/Number/string-numeric-separator-literal-bil-bds-nsl-bds.js
built-ins/Number/string-numeric-separator-literal-dd-dot-dd-ep-sign-minus-dd-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-dd-dot-dd-ep-sign-minus-dds-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-dd-dot-dd-ep-sign-plus-dd-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-dd-dot-dd-ep-sign-plus-dds-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-dd-nsl-dd-one-of.js
built-ins/Number/string-numeric-separator-literal-dds-dot-dd-nsl-dd-ep-dd.js
built-ins/Number/string-numeric-separator-literal-dds-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-dot-dd-nsl-dd-ep.js
built-ins/Number/string-numeric-separator-literal-dot-dd-nsl-dds-ep.js
built-ins/Number/string-numeric-separator-literal-dot-dds-nsl-dd-ep.js
built-ins/Number/string-numeric-separator-literal-dot-dds-nsl-dds-ep.js
built-ins/Number/string-numeric-separator-literal-hil-hd-nsl-hd.js
built-ins/Number/string-numeric-separator-literal-hil-hd-nsl-hds.js
built-ins/Number/string-numeric-separator-literal-hil-hds-nsl-hd.js
built-ins/Number/string-numeric-separator-literal-hil-hds-nsl-hds.js
built-ins/Number/string-numeric-separator-literal-hil-od-nsl-od-one-of.js
built-ins/Number/string-numeric-separator-literal-nzd-nsl-dd-one-of.js
built-ins/Number/string-numeric-separator-literal-nzd-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-nzd-nsl-dds.js
built-ins/Number/string-numeric-separator-literal-oil-od-nsl-od-one-of.js
built-ins/Number/string-numeric-separator-literal-oil-od-nsl-od.js
built-ins/Number/string-numeric-separator-literal-oil-od-nsl-ods.js
built-ins/Number/string-numeric-separator-literal-oil-ods-nsl-od.js
built-ins/Number/string-numeric-separator-literal-oil-ods-nsl-ods.js
built-ins/Number/string-numeric-separator-literal-sign-minus-dds-nsl-dd.js
built-ins/Number/string-numeric-separator-literal-sign-plus-dds-nsl-dd.js
built-ins/Number/string-octal-literal-invald.js
built-ins/Number/string-octal-literal.js
```
