Built catchable RangeErrors for Number.toFixed, toExponential, toPrecision and toString(radix).
Branch codex/library-math-number, based on f5d34f904cd538a497688281ad3909bb580e7b04.
Math 118/0/38 unchanged; Number 190/0/32 becomes 195/0/27 (pass/disagreement/refused).
Four formatting mutants fail only Node comparison; removing the flow hook fails Node's trace comparison.
Array growth, boxing, coercion, detached methods and unsupported syntax remain refused; no whole packages run.

## Scope and port

The largest Math group is 29 tests constructing `new Array()` then growing it through indexed writes. This needs array growth and hole semantics beyond the library's existing bounds contract. I did not implement that compiler/array change here. The largest Number group is six unsupported `var` cases, also outside this library unit. The largest formatting gap is five catchable range-error tests, now passing without adapting their semantics.

Existing fdlibm and V8 digit-generation ports already implement the accepted numerical operations. This change ports V8's validation order from version 13.6.233.17 `src/builtins/builtins-number.cc` and `src/builtins/number.tq` into `internal/native/runtime/library_number_format.c` and its header, both named in THIRD_PARTY_NOTICES.md. Successful calls reuse the existing exact algorithms. Fixed and radix validate before nonfinite receivers; exponential and precision return nonfinite receivers before range rejection. Arguments use ToIntegerOrInfinity, including NaN and fractional values.

Pending RangeError objects own their message and carry the runtime error tag. Lowering's exception inference, native checks and flow edges share the same predicate. The fixture checks Error narrowing, RangeError names/messages, valid formatting, operand order and finally/cleanup. Source `instanceof RangeError` remains unsupported; this change does not claim that additional nominal-class support.

## Counts and refusal groups

Pinned test262: 5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81. Full machine summaries are before.json and after.json. Counts exclude skipped and stock-TypeScript-rejected cases.

| Directory | Before pass / disagreement / refused | After |
| --- | --- | --- |
| built-ins/Math | 118 / 0 / 38 | 118 / 0 / 38 |
| built-ins/Number | 190 / 0 / 32 | 195 / 0 / 27 |

Math also has 10 not-TypeScript and 161 skipped cases; Number has 44 not-TypeScript and 74 skipped. Neither run crashed.

Math refusals: array construction/growth 29; absent Math.sumPrecise 4; unsupported var 2; nondeterministic random 1; detached max/min 2. None was silently approximated.

Number remaining refusals: var 6; new identifiers 5 (three escaping Number boxes and two Object constructions used for coercion); isPrototypeOf 3; for-in without a fixed plain origin 2; isFinite member 2; detached toString 2; object Number coercion 1; string-plus-number coercion 1; explicit undefined toExponential argument 1; String box 1; loose equality 1; detached toFixed 1; void 1. These require syntax, object/prototype, coercion or intrinsic-call support beyond this formatting change.

New passes: toExponential/range.js, toFixed/range.js, toPrecision/range.js, toString/numeric-literal-tostring-radix-1.js and toString/numeric-literal-tostring-radix-37.js, all under built-ins/Number/prototype.

## Verification

All command output was written directly to log files, archived in logs.tar.gz. Source `/workspace/adamic-tools/env.sh`; add stage3/api/node_modules/.bin to PATH for test262.

Setup: GOPROXY=https://proxy.golang.org|direct bash cloud/setup.sh; npm ci --prefix stage3/api. Setup took 71.679 seconds: Node .035, Go .048, submodules .125, markdown .130, clang .390, Go build 71.340, cache warm 71.610. nproc=5, CPU quota=4; Node v24.19.0, Go 1.27.1, clang 20.1.8. npm installed three packages in 830ms.

Before and after:
```
go run ./cmd/adamic-test262 -adapt -json -test262 /tmp/library-math-number-test262 -work /tmp/library-math-number-before built-ins/Math built-ins/Number
go run ./cmd/adamic-test262 -adapt -json -test262 /tmp/library-math-number-test262 -work /tmp/library-math-number-after built-ins/Math built-ins/Number
```

Focused oracle final run (uncached):
```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(library_math_number_.*|precision_range[.]a|radix_range[.]a)|TestMathNumberOracleCatchesMutants' -count=1 -v -timeout=30m
```
PASS, 3.787s: five fixtures on Node, JavaScript, native release and sanitizer paths, and all eleven family mutants. Each new mutant offsets the checked formatter receiver by one; it compiles, exits cleanly and passes sanitizers, but Node's byte comparison fails. Four new families: fixed, exponential, precision, radix. Existing seven mutants remain green as kill tests.

```
go test ./internal/flow -run 'TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/library_math_number_prototype' -count=1 -v -timeout=10m
```
PASS .427s, 7510 points and 8973 events. `run-flow-mutant.py` removes only the formatting throw-edge hook through a Go overlay: Node's actual trace reports no edge from the formatting block, exit 1. The overlay leaves working sources untouched.

The updated existing lower refusal test (try around an indirectly failing repeat) passed .188s. Initial focused command's flow pattern selected no tests; the explicit command above subsequently ran the real trace check.

```
go test ./internal/native -run '^(TestNumbersFormatExactlyAsJavaScriptDoes|TestToExponentialAndToPrecisionMatchNode|TestToStringWithARadixMatchesNode|TestMathAndToFixedMatchJavaScript)$' -count=1 -v -timeout=30m
```
PASS 15.259s: shortest printing 255974 values; exponential/precision 778686 answers; radix 146452 answers; Math/toFixed 69132 answers. Zero Node differences.

```
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=30m -args -update-counts
go vet ./internal/ir ./internal/lower ./internal/native ./internal/flow ./internal/oracle
```
Counts PASS 217.707s; vet, gofmt and git diff --check clean. Only three recorded rows changed: expanded prototype fixture and two existing range-panic fixtures. The latter now allocate an owned RangeError before an uncaught stop, so their allocation counts intentionally change. No count values were edited by hand.

No whole-package test gate, WASI gate, new Math numerical implementation, array growth, general boxed-number semantics or unsupported syntax was covered. The fast gate runs broader checks.
