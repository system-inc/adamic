# Deterministic Date library slice

This is the historical first-pass report. See [finish-report.md](finish-report.md) for completed toJSON and dynamic parsing.

Branch `codex/library-date`, from current origin/main `5d4c801`.
`git fetch origin` and `git checkout -b codex/library-date origin/main` completed.
The runner branch was fetched and merged before measurement:
`git merge origin/cloud/grok-test262-runner` printed `Already up to date.`
Main already contains that runner. Its Go source and tests were read in full.
Test262 is pinned at `3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd`.

## Observations

46 tests newly pass, all agreeing with Node. Both runs use the original tests,
with adaptation disabled. No behavior disagreements or crashes were observed.
The largest remaining refusal is TS2345 argument type mismatch (100 cases),
followed by `var` (63) and inherited prototype reads (45). Skips are unchanged;
144 require propertyHelper.js, 48 require isConstructor.js, and many require Symbol.
These numbers measure what the runner can lower, not complete Date conformance.
The complete passing paths and refusal reasons are retained in after.json.

| Measurement | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Before, built-ins/Date | 0 | 0 | 324 | 0 | 270 | 594 |
| After, built-ins/Date | 46 | 0 | 278 | 0 | 270 | 594 |

## Implementation

New lowering is in `internal/lower/library_date.go`, runtime arithmetic in
`internal/native/runtime/date.c`, and the IR and both emitters in slice-named files.
Shared dispatch changes are small. Neither `internal/lower/lower.go`,
`internal/native/emit.go`, nor `internal/native/native.go` was edited.

Built: numeric and copying constructors; year/month/component construction;
Date.UTC; every getUTC*/setUTC* method; getTime and valueOf; toISOString;
invalid dates and their NaN; UTC-fixed local getters/setters and string methods;
Date.parse and new Date(string) for constant ISO forms, including incomplete ISO
dates, expanded years, offsets and fractional milliseconds. Parsing also accepts
V8's UTC toString and toUTCString forms, either as constants or direct Date method
results, covering test262's parse/zero.js invariants. Explicit
Date.prototype.method.call(actualDate, ...) is supported without inventing a
prototype object or allowing a non-Date receiver.

Dates retain a scalar internal slot without enumerable properties or reference
captures. Structural views that erase or forge this slot are refused. Setter
mutations are visible to flow analysis. Gregorian civil arithmetic uses floor
division for negative days, avoids libc time_t and supports the inclusive
plus/minus 8.64e15 millisecond limits, years before 1970 and beyond 9999, leap
centuries, normalization, invalid inputs, year 0..99 remapping and negative zero.
Floating-point evaluation order is held to V8's Date.UTC result to the millisecond.

Date.now, reading Date.now as a value, new Date() without arguments, and Date()
are Refused: the wall clock is nondeterministic and cannot be compared with Node;
an explicit timestamp is required. Ordinary unsupported operations remain NotYet.

### UTC contract

The compiler permits local Date operations only when its environment has exactly
`TZ=UTC`; otherwise it gives the explicit UTC requirement. The test262 runner's
runCommandWithLimit appends TZ=UTC to every child environment, including the
compiler, native binary and Node. The oracle sets TZ=UTC before parallel fixture
lowering and executeWith appends it for both native and Node children. The flow
trace tests use the same policy. This pins the timezone rather than depending on
the host timezone. Runtime local operations intentionally use the same civil UTC
arithmetic; builds and executions outside this contract are not supported.

### Requested coverage left unbuilt

Date.toJSON is explicitly NotYet. Invalid dates return null on Node, while Adamic
has no general null value representation. Returning a string or undefined would
silently miscompile it. Full toJSON also requires generic receiver, coercion and
method-lookup semantics that this nominal slice does not represent.

Arbitrary dynamic parse strings and the rest of V8's permissive legacy grammar
are explicitly NotYet. Negative-zero expanded date-only inputs are refused
because V8 falls back to legacy parsing. Object coercion callbacks, detached
methods, prototype mutation, locale methods, dynamic this, spread arguments and
catchable Date.toISOString RangeErrors remain outside this slice. Uncaught invalid
ISO formatting agrees with Node's RangeError observation; a catch around that
validation is refused. This is a partial implementation of the requested unit,
not a claim that all Date tests or all requested methods are covered.

## Toolchain

`bash cloud/setup.sh > /tmp/library-date-setup.log 2>&1` succeeded. Exact timings:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (246s)
setup: done in 246s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed `5`. Go 1.27.1, clang 20.1.8, Node v24.19.0. Every subsequent
Go command sourced `/workspace/adamic-tools/env.sh`. Full setup output is setup.log.

## Commands and outputs

All test output was redirected to logs and read afterwards; no test run was piped.
The two directory measurements used this command, saving separate JSON and logs:

```
TZ=UTC go run ./cmd/adamic-test262 -json -test262 /workspace/test262 built-ins/Date
```

Both exited 0. Before: `pass 0 fail 0 refused 324 crashed 0 skipped 270 total 594`.
After: `pass 46 fail 0 refused 278 crashed 0 skipped 270 total 594`.

The broad touched-package checks were:

```
TZ=UTC go test ./internal/lower ./internal/native ./internal/fresh ./cmd/adamic-test262 -count=1 -timeout 10m
```

PASS: lower 42.468s, native 374.344s, fresh 59.178s, runner 94.268s (packages.log).
A subsequent full lowering check passed in 58.249s. Full flow
`TZ=UTC go test ./internal/flow -count=1 -timeout 10m` passed in 303.065s (flow.log).
These broad runs preceded the final explicit .call fixture; final focused checks
below verify that addition and the final ordinary oracle checks every fixture.

```
TZ=UTC ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^Test(NativeAgreesWithNode|DateOracleCatchesMutants)$' -count=1 -timeout 10m -v
```

PASS in 209.658s (oracle.log). All ordinary registered fixtures and the eight Date
mutants ran. Cache counters: native hits=0 misses=752; node hits=0 misses=538;
probe hits=0 misses=0. Date fixtures agree across source Node, sanitized native,
release native and the JavaScript backend; terminating programs pass leak checks.

```
TZ=UTC go test ./internal/oracle -run '^TestCountsAreRecorded$' -update-counts -count=1 -timeout 4m
```

PASS in 91.053s (counts.log). Eight Date rows added; all existing rows unchanged.

```
TZ=UTC go test ./internal/flow -run '^TestDateMutationRanges$' -count=1 -timeout 2m -v
TZ=UTC go test ./internal/lower -run '^TestDate' -count=1 -timeout 2m -v
```

Restored flow PASS in 0.269s: `Node observed 92 Date mutations, all inside their
ranges` (flow-restored.log). Date refusal checks PASS in 0.668s, including a child
compiler with TZ=America/New_York (refusals.log).
`gofmt -l cmd internal`, `go vet ./...`, and `git diff --check` all exit 0 with empty
logs. The complete `go test ./...` integration gate was not run; the touched
packages, all ordinary oracle fixtures, counts, static checks and focused final
checks above are the worker validation.

One earlier full ordinary oracle run failed only on the setter fixture: that
fixture acquired explicit .call while the running test binary still had the old
lowering. The failure was an unbound-method refusal, not a behavior disagreement.
Its log is retained as oracle-before-call-rebuild.log. The final stable-source
rerun above passed. During development Node also rejected assumptions about
incomplete ISO datetimes and nonzero digits beyond millisecond precision at
24:00; those were corrected and are now fixture inputs.

## Mutant evidence

TestDateOracleCatchesMutants retains all eight behavior mutants. Each compiled
with -Werror, exited 0, and had empty stderr under ASan/UBSan. Only Node stdout
comparison caught the wrong answer; compiler or sanitizer rejection does not
count. Every family printed `Node caught the mutant: stdout differs; sanitizer clean`.

| Family | Deliberate wrong behavior | Caught by |
|---|---|---|
| Constructor / TimeClip | Add one millisecond before construction | Node stdout |
| Date.UTC / civil arithmetic | Add one millisecond to result | Node stdout |
| Invalid-date NaN | Replace NaN getters with zero | Node stdout |
| UTC / local getters | Add one to getter results | Node stdout |
| UTC / local setters | Add one millisecond to reported setter results | Node stdout |
| ISO formatting | Replace final millisecond digit with zero | Node stdout |
| UTC / local string formatting | Shift Date inputs by one day | Node stdout |
| ISO / rendered-string parsing | Add one millisecond to parsed result | Node stdout |

An additional flow mutant omitted the setter's n.store effect, while retaining
valid Go and normal Node execution. TestDateMutationRanges failed:
`Node observed Date mutation outside its inferred range: mutated 4 8`, exit 1,
in 0.513s (flow-mutant.log). The implementation was restored and the check passed,
observing all 92 Date mutations. This mutant proves the mutation-range check,
separately from the native behavior comparisons.

## Full directory tables

### Before

| Directory | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Date | 0 | 0 | 49 | 0 | 29 | 78 |
| built-ins/Date/UTC | 0 | 0 | 13 | 0 | 4 | 17 |
| built-ins/Date/now | 0 | 0 | 4 | 0 | 2 | 6 |
| built-ins/Date/parse | 0 | 0 | 4 | 0 | 4 | 8 |
| built-ins/Date/prototype | 0 | 0 | 43 | 0 | 1 | 44 |
| built-ins/Date/prototype/Symbol.toPrimitive | 0 | 0 | 0 | 0 | 18 | 18 |
| built-ins/Date/prototype/constructor | 0 | 0 | 0 | 0 | 1 | 1 |
| built-ins/Date/prototype/getDate | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getDay | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getFullYear | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getHours | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getMilliseconds | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getMinutes | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getMonth | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getSeconds | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getTime | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getTimezoneOffset | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCDate | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCDay | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCFullYear | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCHours | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCMilliseconds | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCMinutes | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCMonth | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCSeconds | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/Date/prototype/setDate | 0 | 0 | 9 | 0 | 5 | 14 |
| built-ins/Date/prototype/setFullYear | 0 | 0 | 15 | 0 | 5 | 20 |
| built-ins/Date/prototype/setHours | 0 | 0 | 18 | 0 | 5 | 23 |
| built-ins/Date/prototype/setMilliseconds | 0 | 0 | 9 | 0 | 5 | 14 |
| built-ins/Date/prototype/setMinutes | 0 | 0 | 13 | 0 | 5 | 18 |
| built-ins/Date/prototype/setMonth | 0 | 0 | 12 | 0 | 5 | 17 |
| built-ins/Date/prototype/setSeconds | 0 | 0 | 12 | 0 | 5 | 17 |
| built-ins/Date/prototype/setTime | 0 | 0 | 6 | 0 | 5 | 11 |
| built-ins/Date/prototype/setUTCDate | 0 | 0 | 3 | 0 | 4 | 7 |
| built-ins/Date/prototype/setUTCFullYear | 0 | 0 | 2 | 0 | 4 | 6 |
| built-ins/Date/prototype/setUTCHours | 0 | 0 | 7 | 0 | 4 | 11 |
| built-ins/Date/prototype/setUTCMilliseconds | 0 | 0 | 4 | 0 | 4 | 8 |
| built-ins/Date/prototype/setUTCMinutes | 0 | 0 | 4 | 0 | 4 | 8 |
| built-ins/Date/prototype/setUTCMonth | 0 | 0 | 5 | 0 | 4 | 9 |
| built-ins/Date/prototype/setUTCSeconds | 0 | 0 | 5 | 0 | 4 | 9 |
| built-ins/Date/prototype/toDateString | 0 | 0 | 3 | 0 | 4 | 7 |
| built-ins/Date/prototype/toISOString | 0 | 0 | 15 | 0 | 2 | 17 |
| built-ins/Date/prototype/toJSON | 0 | 0 | 2 | 0 | 11 | 13 |
| built-ins/Date/prototype/toLocaleDateString | 0 | 0 | 0 | 0 | 4 | 4 |
| built-ins/Date/prototype/toLocaleString | 0 | 0 | 0 | 0 | 4 | 4 |
| built-ins/Date/prototype/toLocaleTimeString | 0 | 0 | 0 | 0 | 4 | 4 |
| built-ins/Date/prototype/toString | 0 | 0 | 4 | 0 | 4 | 8 |
| built-ins/Date/prototype/toTemporalInstant | 0 | 0 | 0 | 0 | 8 | 8 |
| built-ins/Date/prototype/toTimeString | 0 | 0 | 2 | 0 | 4 | 6 |
| built-ins/Date/prototype/toUTCString | 0 | 0 | 5 | 0 | 4 | 9 |
| built-ins/Date/prototype/valueOf | 0 | 0 | 2 | 0 | 4 | 6 |

### After

| Directory | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Date | 0 | 0 | 49 | 0 | 29 | 78 |
| built-ins/Date/UTC | 10 | 0 | 3 | 0 | 4 | 17 |
| built-ins/Date/now | 0 | 0 | 4 | 0 | 2 | 6 |
| built-ins/Date/parse | 3 | 0 | 1 | 0 | 4 | 8 |
| built-ins/Date/prototype | 0 | 0 | 43 | 0 | 1 | 44 |
| built-ins/Date/prototype/Symbol.toPrimitive | 0 | 0 | 0 | 0 | 18 | 18 |
| built-ins/Date/prototype/constructor | 0 | 0 | 0 | 0 | 1 | 1 |
| built-ins/Date/prototype/getDate | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getDay | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getFullYear | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getHours | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getMilliseconds | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getMinutes | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getMonth | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getSeconds | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getTime | 2 | 0 | 1 | 0 | 5 | 8 |
| built-ins/Date/prototype/getTimezoneOffset | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCDate | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCDay | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCFullYear | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCHours | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCMilliseconds | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCMinutes | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCMonth | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/getUTCSeconds | 1 | 0 | 2 | 0 | 5 | 8 |
| built-ins/Date/prototype/setDate | 0 | 0 | 9 | 0 | 5 | 14 |
| built-ins/Date/prototype/setFullYear | 0 | 0 | 15 | 0 | 5 | 20 |
| built-ins/Date/prototype/setHours | 0 | 0 | 18 | 0 | 5 | 23 |
| built-ins/Date/prototype/setMilliseconds | 0 | 0 | 9 | 0 | 5 | 14 |
| built-ins/Date/prototype/setMinutes | 0 | 0 | 13 | 0 | 5 | 18 |
| built-ins/Date/prototype/setMonth | 0 | 0 | 12 | 0 | 5 | 17 |
| built-ins/Date/prototype/setSeconds | 0 | 0 | 12 | 0 | 5 | 17 |
| built-ins/Date/prototype/setTime | 0 | 0 | 6 | 0 | 5 | 11 |
| built-ins/Date/prototype/setUTCDate | 0 | 0 | 3 | 0 | 4 | 7 |
| built-ins/Date/prototype/setUTCFullYear | 0 | 0 | 2 | 0 | 4 | 6 |
| built-ins/Date/prototype/setUTCHours | 0 | 0 | 7 | 0 | 4 | 11 |
| built-ins/Date/prototype/setUTCMilliseconds | 0 | 0 | 4 | 0 | 4 | 8 |
| built-ins/Date/prototype/setUTCMinutes | 0 | 0 | 4 | 0 | 4 | 8 |
| built-ins/Date/prototype/setUTCMonth | 0 | 0 | 5 | 0 | 4 | 9 |
| built-ins/Date/prototype/setUTCSeconds | 0 | 0 | 5 | 0 | 4 | 9 |
| built-ins/Date/prototype/toDateString | 1 | 0 | 2 | 0 | 4 | 7 |
| built-ins/Date/prototype/toISOString | 0 | 0 | 15 | 0 | 2 | 17 |
| built-ins/Date/prototype/toJSON | 0 | 0 | 2 | 0 | 11 | 13 |
| built-ins/Date/prototype/toLocaleDateString | 0 | 0 | 0 | 0 | 4 | 4 |
| built-ins/Date/prototype/toLocaleString | 0 | 0 | 0 | 0 | 4 | 4 |
| built-ins/Date/prototype/toLocaleTimeString | 0 | 0 | 0 | 0 | 4 | 4 |
| built-ins/Date/prototype/toString | 1 | 0 | 3 | 0 | 4 | 8 |
| built-ins/Date/prototype/toTemporalInstant | 0 | 0 | 0 | 0 | 8 | 8 |
| built-ins/Date/prototype/toTimeString | 1 | 0 | 1 | 0 | 4 | 6 |
| built-ins/Date/prototype/toUTCString | 3 | 0 | 2 | 0 | 4 | 9 |
| built-ins/Date/prototype/valueOf | 0 | 0 | 2 | 0 | 4 | 6 |
