# Date completion: JSON and dynamic parsing

Continues `codex/library-date` from `5c021e552b00709448c0fd4e25be88a355db792b`.
The first slice and its historical measurements remain in report.md. This phase
builds the two behaviors that it explicitly left unfinished.

## Runner observations

| Measurement | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Before this phase | 46 | 0 | 278 | 0 | 270 | 594 |
| After this phase | 47 | 0 | 277 | 0 | 270 | 594 |
| Original main baseline | 0 | 0 | 324 | 0 | 270 | 594 |

One newly passing test: `built-ins/Date/parse/year-zero.js`, whose loop supplies
dynamic strings. All 47 passing tests agree with Node; no crashes were observed.
The biggest refusal remains TS2345 argument type mismatch (100), followed by
`var` (63). `toJSON` adds no test262 passes because its files use refused language
or harness patterns, including var, accessors, detached methods and generic
borrowed receivers. Its accepted typed behavior is held to Node in the new
oracle fixture. These counts are observations, not a claim of full test262 conformance.

Both measurements used the pinned test262 checkout
`3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd`, with adaptation disabled:

```
TZ=UTC go run ./cmd/adamic-test262 -json -test262 /workspace/test262 built-ins/Date
```

Both exit 0; complete JSON, directory tables, passing paths and refusal reasons
are in finish-before.json and finish-after.json. Logs are finish-before.log and
finish-after.log. Full directory tables are also below.

## What is built

Date.toJSON on actual Date receivers returns their exact ISO string or null for
an invalid/non-finite Date. Optional key arguments are evaluated, then ignored;
receiver binding and evaluation order are preserved. No wall clock is consulted.
The Date library declaration is corrected before checking: `string | null`, not
stock TypeScript's inaccurate `string` promise. The existing reference null
pointer now represents a nullable string; it is never conflated with undefined.
Null checks, typeof, String/Number conversion, templates, concatenation,
nullish coalescing, stored results and function returns are covered.

JSON.stringify recognizes Dates and nullable strings, including Date fields in
literal objects and array elements. Date serialization happens after all call
arguments have been evaluated; the space argument's Date mutation fixture proves
this order. Both backends preserve the Date instead of erasing it into a plain
object. Nullable generic instantiations use their concrete type and a distinct
body key: a string body cannot be reused for string | null. The fixture exercises
generic functions and a generic class with an invalid-date JSON result.

Dynamic Date.parse and new Date(string) share V8's ISO-first, legacy-fallback
scanner and composer algorithm. It accepts the ISO forms, expanded years,
fractional milliseconds, lowercase separators, compact offsets, and the
Date.toString/toUTCString forms checked by test262. It also handles month names
and prefixes, slash/dash/dot dates, two-digit years, AM/PM, named zones, numeric
zone offsets, ignored comments and leading words, Unicode whitespace and embedded
NUL behavior. Invalid input returns NaN; negative-zero expanded dates follow
V8's distinction between invalid datetime forms and date-only legacy fallback.
The private implementation is in date_parse_impl.h, included by date.c. V8's BSD
license is retained in date_v8_license.txt.

The new dynamic fixture includes explicit edge inputs and sweeps of month names,
years, offsets, separators and malformed suffixes, comparing both constructors
and Date.parse with Node. Existing millisecond arithmetic and inclusive plus/minus
8.64e15 limits remain held by the original Date fixtures.

### Determinism and boundaries

The existing oracle and runner pin TZ=UTC before lowering and append TZ=UTC to
both native and Node child environments. Dynamic parsing always requires that
compiler contract because the input may select a local-time legacy or ISO form.
The non-UTC child test covers both Date.parse and new Date(string).
Date.now, new Date() with no arguments, and Date() remain refused with the wall
clock reason.

Generic borrowed toJSON receivers, custom ToPrimitive hooks, dynamic prototype
replacement, accessors, locale formatting and other existing language restrictions
remain explicitly refused. A generic receiver is not falsely described as
requiring a Date internal slot: its refusal names the unbuilt coercion and method
lookup. Mixed null/undefined unions remain unbuilt. Object.is involving null is
explicitly NotYet rather than boxing null as undefined; strict equality works.
This phase completes the requested actual-Date toJSON and dynamic string parsing
behavior; it does not relax Adamic's other language restrictions.

## Validation and outputs

All commands sourced `/workspace/adamic-tools/env.sh`; test output went to logs.
Node v24.19.0, V8 13.6.233.17-node.51, Go 1.27.1 and clang 20.1.8.
The original successful setup timings still apply: Go 0s, clang 1s, Node 1s,
submodules 1s, build cache warm 246s, total 246s; nproc=5, CPU quota=4.
The toolchain was not reinstalled for this continuation.

```
TZ=UTC ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^Test(NativeAgreesWithNode|DateOracleCatchesMutants)$' -count=1 -timeout 10m -v
```

Final PASS in 185.993s (finish-oracle.log), including every ordinary registered
fixture and all 14 Date behavior mutants. Cache observations: native hits=0
misses=758; node hits=0 misses=548; probe hits=0 misses=0. Source Node, sanitized
native, release native and the JavaScript backend agree. Terminating Date fixtures
pass leak checks. A prior broader run before the generic addition also passed
in 376.081s; the final result above supersedes it.

```
TZ=UTC go test ./internal/oracle -run '^TestDateOracleCatchesMutants$' -count=1 -timeout 3m -v
```

PASS in 9.000s; all 14 mutants compiled and exited 0 with empty sanitizer stderr,
and only Node stdout comparison caught them (finish-mutants.log). The same 14
ran again in the final ordinary oracle command above.

```
TZ=UTC go test ./internal/oracle -run '^TestCountsAreRecorded$' -update-counts -count=1 -timeout 4m
```

Final PASS in 99.815s (finish-counts.log). Two rows added, every old row unchanged:
JSON fixture allocations/frees 260/260, peak 9; dynamic parsing 3293/3293, peak 9.

```
TZ=UTC go test ./internal/load ./internal/lower ./internal/fresh ./cmd/adamic-test262 -count=1 -timeout 5m
TZ=UTC go test ./internal/native -count=1 -timeout 10m
TZ=UTC go test ./internal/flow -count=1 -timeout 10m
```

PASS: load 11.254s, lower 86.615s, fresh 129.602s, runner 218.924s
(finish-packages.log); native 430.266s (finish-native.log); flow 213.612s
(finish-flow.log). These broad runs preceded the final generic-only integration.
The runtime is unchanged since them; the final oracle above covers that integration.
Final loading/lowering/ownership rerun:

```
TZ=UTC go test ./internal/lower ./internal/load ./internal/fresh -count=1 -timeout 5m
```

PASS: lower 69.098s, load 7.822s, fresh 80.578s (finish-final-lower.log).
The focused Date flow check passes in 0.648s, observing all 92 Date mutations
inside their inferred ranges (finish-flow-focused.log). Restored timezone and
clock refusal tests pass in 1.859s (finish-tz-restored.log).
`gofmt -l cmd internal`, `go vet ./...`, and `git diff --check` all exit 0 with empty
logs. The complete ./... integration gate was not run; touched packages, all
ordinary oracle fixtures, counts, static checks and final targeted checks are
reported above.

Two development fixture checks were rejected by the checker, not by a behavior
disagreement: raw JSON.stringify output needed undefined handling (TS2345), and
a constructor parameter property violated erasableSyntaxOnly (TS1294). Both
fixtures were corrected before the final successful run.

## Every mutant and what caught it

The original eight behavior mutants remain: constructor/TimeClip +1ms, Date.UTC
+1ms, invalid NaN replaced with zero, getters +1, setters +1ms, corrupted final ISO
millisecond digit, formatting shifted one day, and parsing +1ms. Node stdout
caught each, with normal exit and no sanitizer finding.

| New family | Mutant | Caught by |
|---|---|---|
| toJSON | Return Invalid Date text instead of invalid-date null | Node stdout only |
| Dynamic/legacy parse | Add 1ms to parsed results | Node stdout only |
| Nullable typeof | Report undefined instead of object for null | Node stdout only |
| Nullable Number conversion | Return NaN instead of zero for null | Node stdout only |
| Nullable JSON serialization | Treat null strings as undefined | Node stdout only |
| Date JSON serialization | Serialize Dates as empty Maps | Node stdout only |

Three additional checks were proved independently and restored:

- Restoring the false stock non-null toJSON declaration made
  TestDateJSONRequiresNullHandling fail, exit 1 in 0.566s, because the unsafe
  string use was accepted. The restored check passed in 0.527s
  (finish-type-mutant.log, finish-type-restored.log).
- Omitting dynamic parsing's UTC guard made TestDateLocalTZRequirement fail,
  exit 1 in 0.689s, with `want an explicit UTC requirement, got <nil>`.
  The restored tests passed (finish-tz-mutant.log, finish-tz-restored.log).
- Collapsing nullable and present string generic keys compiled and produced a
  clean native exit 0 with empty stderr. Generic typeof printed undefined and
  serialization printed missing for null; source Node printed object and null.
  Only Node stdout comparison caught it, exit 1 in 1.201s
  (finish-generic-mutant.log). The final restored oracle passes.

The first slice's separate omitted-setter flow mutant and its restored proof
remain recorded in report.md and flow-mutant.log.

## Full directory tables

### Before this phase

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

### After this phase

| Directory | Pass | Disagreement | Refused | Crash | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Date | 0 | 0 | 49 | 0 | 29 | 78 |
| built-ins/Date/UTC | 10 | 0 | 3 | 0 | 4 | 17 |
| built-ins/Date/now | 0 | 0 | 4 | 0 | 2 | 6 |
| built-ins/Date/parse | 4 | 0 | 0 | 0 | 4 | 8 |
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
