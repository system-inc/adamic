Built string + number, number + string and string-slot += number for roadmap step 12.
Base 54cbc125; branch compiler/string-plus-number; scanner evidence 2a34651e.
Three new Node/backend/sanitizer/leak fixtures pass; four existing regressions pass; a-check checks 3 files; counts refreshed.
One %g formatter mutant completes cleanly and is caught by stdout for both addition orders and +=.
The complete scanner replay is not verified; the available full tree stops at unrelated checker errors.

# String plus number

`internal/lower/expression.go` wraps the numeric operand in the existing
`ir.NumberToString` before constructing `ir.Concat`. Templates already use that
IR node. Both backends and the native number formatter remain unchanged.
Number + number still uses numeric addition. String + string is unchanged.
The shared compound-assignment lowering uses the same conversion for string
locals, captured bindings and fields, including assignment expressions.

## Scanner reductions

Read `main-area-next-replay/ordered-stops.json` and the archived scanner source
in `private-placeholders.json.gz` on `2a34651e`. The recorded positions belong
to the private replay copies; throwing placeholders changed subsequent line numbers.

| Recorded site | Scanner expression | Minimal executable reduction |
| --- | --- | --- |
| scanner.ts:696:32, scanNumber | `mainFragment = "" + +tokenValue` | `console.log('' + +'08');` |
| scanner.ts:1221:26, checkBigIntSuffix | `tokenValue = "" + numericValue` | `console.log('' + 8);` |

The two fixtures are `scanner_string_number_696.a` and
`scanner_string_number_1221.a` under `internal/oracle/testdata`.
Both print `8` followed by a newline on Node and both backends.
Both were run through `adamic c` before the change and stopped with
`stage 0 can't lower a BinaryExpression with a string and a number yet`.
After minimizing them to one executable statement, the baseline check was
repeated using a Go overlay containing the original `54cbc125` expression.go.
Both again exit 1 at fixture line 3, column 13 with that exact reason.
The overlay does not change the working compiler sources.

## Acceptance fixture and mutant

`string_plus_number.a` sweeps 16 values: 0, -0, 7, -42, NaN, Infinity,
-Infinity, 1e21, 1e20, 1e-7, 1e-6, 0.5, -0.125,
1.23456789012345, 5e-324 and 1.7976931348623157e308.
Each is printed with string + number, number + string, string += number
and a template control. Thus -0 must print 0, special values retain
JavaScript spelling, exponent thresholds and full fractional precision come
from the existing formatter.

The fixture also checks numeric addition before concatenation, chained
concatenation, effectful operands in both orders, a captured string replaced
while its += right operand runs, the result of +=, and a string field update.
Runtime-built strings exercise ownership rather than only immortal literals.

`TestStringPlusNumberOracleCatchesFormatterMutant` replaces generated-C
`adamic_string_from_number` calls with a helper using `snprintf(..., "%g", ...)`.
The helper decodes its buffer through the existing runtime string constructor.
The mutated binary compiles under the ordinary strict C flags and ASan/UBSan,
exits 0 with empty stderr, and disagrees with source Node only in stdout.
The test independently requires the mutant to print `left:-0`, `-0:right`
and `slot:-0`, all absent from Node stdout. Template differences therefore
cannot be the only reason the mutant is caught.

## Commands and outputs

Every test command wrote its output to a log file. Setup's environment was
sourced in each Go shell; GOPROXY was `https://proxy.golang.org|direct`.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestStringPlusNumberOracleCatchesFormatterMutant$|^TestNativeAgreesWithNode$/internal/oracle/testdata/^(scanner_string_number_696|scanner_string_number_1221|string_plus_number|number_formats|string_append|undefined_strings|updates)[.]a$' -v -count=1 -timeout 30m > /tmp/string-plus-number-final-oracle.log 2>&1
# PASS: 7 fixtures and formatter mutant; oracle 1.239s.

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestStringPlusNumberOracleCatchesFormatterMutant$|^TestNativeAgreesWithNode$/internal/oracle/testdata/^(scanner_string_number_696|scanner_string_number_1221|string_plus_number)[.]a$' -v -count=1 -timeout 30m > /tmp/string-plus-number-minimal-oracle.log 2>&1
# PASS after final minimization: 3 fixtures and formatter mutant; oracle 0.592s.

go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/string-plus-number-final-counts.log 2>&1
# PASS: oracle 36.772s. Only the three new rows are added.

python3 /tmp/string-plus-number-a-check/run.py string-plus-number /tmp/string-plus-number-a-check/paths.json > /tmp/string-plus-number-a-check.log 2>&1
# PASS: 3 checked, a-check exit 0.
```

The a-check harness is the one documented in
`stage3/a-check-headers/REPORT.md`, calling unchanged `Gate.aCheck` from
`fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py`.
Its input list contains exactly the three new .a files. All three compiler
invocations exit 0, rather than passing as NotYet.

The normal fixture oracle compares original source on Node, generated
JavaScript on Node, sanitized native and release native. Successful native
programs also undergo LeakSanitizer. The final three counted rows are:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| internal/oracle/testdata/scanner_string_number_696.a | 2 | 2 | 2 | 4 | 2 | 0 |
| internal/oracle/testdata/scanner_string_number_1221.a | 2 | 2 | 0 | 2 | 2 | 0 |
| internal/oracle/testdata/string_plus_number.a | 171 | 171 | 10 | 184 | 7 | 0 |

## Scanner replay limit

The worker's `/workspace/scratch/scanner-main-next-records` replay tree is
absent in this workspace. Its archived source and stop evidence were read,
but its complete private tree was not reproduced.

The available adapted TypeScript tree was attempted explicitly:

```sh
/workspace/adamic-binaries/adamic-acheck c /tmp/scout-map-keys-adapted/src/compiler/scanner.ts > /tmp/string-plus-number-scanner.c 2> /tmp/string-plus-number-scanner-replay.log
```

Exit 1, before lowering: the first diagnostic is
`builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'`.
Further TS2345 and TS2488 diagnostics follow. This does not establish that
full scanner traversal reaches or clears either recorded stop. The two
reductions establish that their exact string/number operation now lowers
and agrees with Node. No native scanner token comparison is claimed.

No new union, boolean or nullish concatenation rule was added. Compound
updates of string array elements remain the existing NotYet; an attempted
tuple-element probe was removed from the acceptance sweep after reproducing
that separate limit. No whole package test or full gate was run.

## Toolchain setup

`bash cloud/setup.sh` exits 0. `nproc` prints 5.

```text
setup: go ready (0.026s)
setup: node ready (0.025s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: submodules ready (0.073s)
setup: markdown dependencies ready (0.074s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.171s)
setup: go build ready (45.355s)
setup: test binaries deferred (use --warm-tests) (45.519s)
setup: build cache warm (45.520s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (45.554s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.7vSI1d
```
