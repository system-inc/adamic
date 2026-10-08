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

The maintainer's requested fix is isolated in `06fc424` on
`codex/grok-date-coverage-flow`, directly based on `39fb2e7`. It adds
library_date_days.a beside size_class_churn.a and bitwise_sweep.a in the flow
program exclusion, explaining that the sweep's size is its purpose and the
other library_date_*.a fixtures cover its shapes. The full flow package on that
branch passes in 82.374s, including TestEveryPathNodeTakesIsInTheGraph,
TestEveryMutationIsInItsRange and TestLivenessHoldsOnEveryPath. Its output is in
grok-flow.log.

That fix is merged into the refusals branch in `50c029a`. The complete 4001-year
sweep is restored byte-for-byte from `39fb2e7` at its original path,
internal/oracle/testdata/library_date_days.a. Its ordinary oracle registration
and original counts remain. The earlier bounded root fixture and relocated
full-sweep copy are removed; no duplicate exhaustive fixture remains. No
missing trace is treated as success and no tracing implementation is changed.

The earlier bounded-fixture flow and relocated-sweep logs are retained as
historical evidence. Final integrated flow, original-path oracle and counts
checks pass in 87.138s, 63.461s and 31.026s, respectively, and are recorded in
flow-integrated.log, sweep-restored.log and counts-restored.log. The test262 measurements are unchanged because these
follow-up changes affect only fixture discovery, fixture location and review
records. Compiler and runner implementation are unchanged.

One earlier concurrent full-sweep attempt returned partial sanitized/backend
outputs and exit -1 at the oracle's one-minute child deadline. The same check
passed in isolation without an implementation change. The full repository gate
was not run.

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
| Initial ordinary Date fixtures | Uncached TestNativeAgreesWithNode filtered to library_date: 17 fixtures, PASS 11.362s; date-oracle.log |
| Historical relocated sweep | Uncached TestNativeAgreesWithNode filtered to date_sweeps: PASS 67.092s; full-sweep-after.log |
| Date mutants | Uncached go test ./internal/oracle -run '^TestDateOracleCatchesMutants$' -count=1 -v: PASS 14.771s; mutants.log |
| Counts | go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts: PASS 90.963s |
| Lower/native/load | Complete package tests: PASS 76.635s / 416.061s / 7.436s; JavaScript backend compiles, no package tests |
| Initial Date flow | Path, mutation-range and liveness checks plus TestDate checks: PASS 8.829s; flow.log |
| Final grok flow branch | Complete go test ./internal/flow -count=1 -timeout 30m -v: PASS 82.374s; grok-flow.log |
| Final integrated flow | Same full package command: PASS 87.138s; flow-integrated.log |
| Restored original sweep | Uncached TestNativeAgreesWithNode at library_date_days.a: PASS 63.461s; sweep-restored.log |
| Restored counts | Complete TestCountsAreRecorded without updating the table: PASS 31.026s; counts-restored.log |
| Runner | Complete go test ./cmd/adamic-test262 -count=1: PASS 37.450s; runner-tests.log |
| Final source audit | 324 ordinary sources byte-identical; 324 adapted sources recompiled: PASS 39.258s; audit.log |
| Static | go vet on runner/lower/native/javascript: exit 0, empty output; vet.log |

The final counts changes add metadata and Number(Date) rows. The full sweep's
original row and path are restored unchanged: 4,400,010 allocations and frees.
Metadata: 36; Number(Date): 43. Both new fixtures balance allocations and frees,
with no values left in regions.

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
