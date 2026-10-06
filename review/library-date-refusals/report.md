Newly passing: +52 ordinary; +100 with --adapt.
Disagreements: 0; crashes: 0 in both Date measurements.
Largest refusal: TS2345 (76); largest valid reason with adaptation: detached toString (13).
Built: Date descriptors, Number(Date), checked Date/scalar annotations and identity assertions.
Fixed: missing flow trace; the complete day sweep remains an oracle fixture.

Branch `codex/library-date-refusals` starts at `codex/library-date` tip `f7cbc8c`.
The first commit is the ownership claim, `9fe69aa`. Coverage `39fb2e7` was merged
in `6f77373`. No language change was made. The TS-invalid coercion cases remain
refused. The old diagnosis in review/library-date/GAPS.md now records the decision
that supersedes it. `verdict.go`, `classify.go`, `lower.go`, `emit.go` and
`native.go` were not edited.

## Measurements

Pinned test262: `3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd`.
Every newly passing test was actually run against Node 24.19.0 with the runner's
sanitized native build. No formerly passing test regressed. Optional adaptation
has its own before measurement; its extra coverage is not mixed with ordinary
mode. JSON files contain every directory table and refusal count; the two
new-passes text files contain complete path lists.

| Mode | Measurement | Pass | Disagreements | Refused | Crashes | Skipped | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| Ordinary | Before | 47 | 0 | 277 | 0 | 270 | 594 |
| Ordinary | After | 99 | 0 | 225 | 0 | 270 | 594 |
| --adapt | Before | 63 | 0 | 261 | 0 | 270 | 594 |
| --adapt | After | 163 | 0 | 161 | 0 | 270 | 594 |

```sh
source /workspace/adamic-tools/env.sh
TZ=UTC go run ./cmd/adamic-test262 -json -work /tmp/date-refusals-before -test262 /workspace/test262 built-ins/Date > /tmp/date-refusals-before.json 2> /tmp/date-refusals-before.log
TZ=UTC go run ./cmd/adamic-test262 -adapt -json -work /tmp/date-refusals-adapt-before -test262 /workspace/test262 built-ins/Date > /tmp/date-refusals-adapt-before.json 2> /tmp/date-refusals-adapt-before.log
TZ=UTC go run ./cmd/adamic-test262 -json -work /tmp/date-refusals-final2 -test262 /workspace/test262 built-ins/Date > /tmp/date-refusals-final2.json 2> /tmp/date-refusals-final2.log
TZ=UTC go run ./cmd/adamic-test262 -adapt -json -work /tmp/date-refusals-adapt-final3 -test262 /workspace/test262 built-ins/Date > /tmp/date-refusals-adapt-final3.json 2> /tmp/date-refusals-adapt-final3.log
```

All four commands exit 0. After the final extension to optional annotations,
the audit also verified all 324 ordinary attempted sources were byte-identical
to the measured ordinary sources. Compiler diagnostics and exact source hashes
for 648 before/after comparisons are in triage.json. Stock tsc logs are recorded
separately for each mode and phase.

The runner's runCommandWithLimit appends `TZ=UTC` to compiler, native and Node
child environments. Oracle and flow Date test initialization pins TZ before
parallel work; their child processes inherit it. This fixes the timezone on both
sides, including local getters/setters, construction, formatting and parsing.
Date.now(), Date(), and zero-argument new Date() retain their explicit refusal
explaining why the wall clock cannot be compared with Node.

## Missing trace reproduction and fix

Observed immediately after the coverage merge: the full library_date_days.a
passes the ordinary uncached oracle (108.608s), but its flow path check fails
(36.527s) because trace.txt is absent. A temporary diagnostic ran the exact
tracing path with Node stderr exposed and an uncaught-exception monitor.
trace-node-stack.log shows RangeError: Invalid array length at Array.push in
adamicLeave, followed by the exit callback also failing. Node exits 70 and cannot
write its trace. The original sweep is a complete, valid Date program; the trace
array is too large for V8. The oracle is right to require the trace file.

The full 4001-year day sweep is preserved byte-for-byte at
internal/oracle/testdata/date_sweeps/library_date_days.a and remains explicitly
registered in the oracle and counts table. Flow discovery scans root fixture
files, so it instead receives a bounded root library_date_days.a. It samples
26 years, all months, first/middle/last days and the original three times of day,
including negative years, year zero, leap centuries and 1970. Its hash and visit
functions are unchanged. No missing trace is treated as success and no generic
trace/oracle code was changed.

The former failing flow test now passes in 1.030s: 538,726 points and 738,263
events. All Date flow checks pass in 8.829s. The preserved full sweep passes the
uncached oracle on the final runtime in 67.092s, with source Node, JavaScript
backend, sanitized native, release native and leak checks.

One concurrent full-sweep attempt returned partial sanitized/backend outputs and
exit -1 at the oracle's one-minute child deadline. The same uncached check passed
in isolation without changing the fixture or implementation. This is evidence
of load-sensitive timing, not a behavior fix. The unrelated complete flow corpus
was stopped after 270.769s; its Date fixtures were then run explicitly. The full
repository gate was not run.

## Refusal audit and implementation

The largest valid ordinary reason was var (63). Existing --adapt provides its
checked style rewrite; it was measured before adding coverage. The largest next
reasons in that mode were prototype descriptor observations (45) and evolving
any locals (43). Date-specific files implement the observations and checked
annotations, with small shared dispatch hooks.

Descriptor support covers Date length/name, Date.parse and Date.UTC length/name,
prototype method length/name, and dynamic string hasOwnProperty keys on Date,
Date.prototype and observed intrinsic methods. Descriptor observation does not
materialize prototypes or detach methods. Number on a proven intrinsic Date
reads its numeric internal slot, including NaN and clipping boundaries; generic
object coercion remains unbuilt.

The optional adapter infers only one checked assignment type per uninitialized
local: number, string or the library's intrinsic Date. It adds undefined to keep
the initial value, refuses mixed/uncertain writes, and checks both the input and
the rewritten program. It does not add casts or change runtime expressions.

The harness rewrites SameValue/NotSameValue to identity conditions when one
operand is null, proven unshadowed undefined, or the unshadowed intrinsic
Date.prototype. Strict equality has precisely SameValue's behavior in these
cases. Operand and message evaluation order, single evaluation, default failure
text, NaN and signed-zero behavior are checked against Node. Scope information
is kept through templates. A helper keeps message fallback inside the function,
avoiding the TS2869 that a literal-message nullish expression would introduce.

16 of the original 17 assertion-domain restrictions are removed. Ten of those
tests pass with adaptation; six still require prototype identity. The remaining
restriction is toJSON/invoke-result.js, which also borrows toJSON onto an object
whose custom toISOString returns another object. Supporting that complete test
needs generic ToPrimitive and dynamic method lookup; it was not rewritten to
omit those observations.

| Mode / phase | Refused | First compiler TS code also rejected by stock tsc | Other first reasons |
|---|---:|---:|---:|
| Ordinary before | 277 | 123 | 154 |
| Ordinary after | 225 | 107 | 118 |
| --adapt before | 261 | 123 | 138 |
| --adapt after | 161 | 107 | 54 |

The 123 baseline matches include 17 harness-created restrictions, leaving 106
correct TS-invalid refusals. The 107 final matches include the one remaining
harness restriction. These are first-diagnostic counts, not assertions that an
entire test has only one blocker. Changes to diagnostic order can expose another
already-existing stock-tsc error; triage.json records the code and full output.

Stock TypeScript 6.0.3 was installed under /tmp/date-refusals-tsc. It checked all
324 exact attempted sources in each mode, with a console declaration matching
Adamic, noEmit, strict, noUncheckedIndexedAccess, exactOptionalPropertyTypes,
noImplicitReturns, noFallthroughCasesInSwitch, erasableSyntaxOnly,
verbatimModuleSyntax, allowImportingTsExtensions, moduleDetection force,
module esnext, moduleResolution bundler, target/lib es2024, and ignoreConfig.
Exit 2 is expected for the recorded TS-invalid cases. No matching checker
rejection was treated as a reason to widen Date's parameter declarations.

Remaining valid limitations and minimal programs are in GAPS.md. The largest
individual valid reason with adaptation is detached toString (13); prototype
value observations account for nine more. Captured ISO validation cannot become
a caught exception when its native failure is a panic. Wall-clock cases remain
refused by design. New outcomes for TS-invalid tests belong to the other worker.

## Validation and mutants

Every test output was directed to a log file. Setup succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (210s)
setup: done in 210s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. setup.log contains the full output.

| Check | Command / result |
|---|---|
| Ordinary Date fixtures | Uncached TestNativeAgreesWithNode filtered to library_date: 17 fixtures, PASS 11.362s; date-oracle.log |
| Full preserved sweep | Uncached TestNativeAgreesWithNode filtered to date_sweeps: PASS 67.092s; full-sweep-after.log |
| Date mutants | Uncached go test ./internal/oracle -run '^TestDateOracleCatchesMutants$' -count=1 -v: PASS 14.771s; mutants.log |
| Counts | go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts: PASS 90.963s |
| Lower/native/load | Complete package tests: PASS 76.635s / 416.061s / 7.436s; JavaScript backend compiles, no package tests |
| Date flow | Path, mutation-range and liveness checks plus TestDate checks: PASS 8.829s; flow.log |
| Runner | Complete go test ./cmd/adamic-test262 -count=1: PASS 37.450s; runner-tests.log |
| Final source audit | 324 ordinary sources byte-identical; 324 adapted sources recompiled: PASS 39.258s; audit.log |
| Static | go vet on runner/lower/native/javascript: exit 0, empty output; vet.log |

The counts update changes only the bounded fixture row and adds metadata,
Number(Date), and preserved-full-sweep rows. The full sweep's original allocation
counts are unchanged: 4,400,010 allocations and frees. Bounded sweep: 3,102;
metadata: 36; Number(Date): 43. Each new ordinary fixture balances allocations
and frees, with no values left in regions.

All 18 native Date mutants compiled valid C, exited 0 with empty sanitizer stderr,
and were caught solely by Node's differing stdout. Each is an actual subtest in
mutants.log:

| Mutant family | Wrong answer injected |
|---|---|
| number_date | Add one to Number(Date)'s valueOf result |
| metadata_name | Date name becomes Dote |
| metadata_length | Intrinsic length 7 becomes 8 |
| metadata_own | Negate own-property result |
| nullable_typeof | Null JSON result has undefined typeof |
| nullable_number | Null numeric conversion becomes NaN |
| nullable_stringify | Drop nullable string JSON behavior |
| date_stringify | Serialize a Date as a map |
| toJSON | Invalid Date returns a string instead of null |
| dynamic_parse | Add one millisecond to dynamic parsing |
| constructor_clip | Add one millisecond to construction |
| UTC | Add one millisecond to Date.UTC |
| invalid_NaN | Invalid getter returns zero |
| getters | Add one to getter results |
| setters | Add one to setter results |
| iso_format | Change an ISO millisecond digit |
| format | Shift formatted dates by a day |
| parse | Add one millisecond to parsing |

Three adapter mutants were run and restored. The wrong SameValue operator causes
the Node comparison regression to fail. Replacing a proven numeric annotation
with string is rejected by the strict post-check and makes the coverage test
fail. Reintroducing literal-message `??` fails the checked-source regression with
TS2869. Their expected failing logs are recorded alongside the passing checks.
