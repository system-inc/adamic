# React no-arrow-function-lifecycle

Roadmap step 13, unit s13-react-no-arrow-function-lifecycle. Based on
origin/area/stage1-lint at 9156bf5c579a44d687c9955d13e44f9ad8bbb6f8.
The all-origin-ref ownership check found no existing rule descriptor before porting.
The adapter uses unchanged Go cohere NoArrowFunctionLifecycle, with no options,
at cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359.

The component-class test uses the shared React projection, class recognizer,
component-base predicate and base-name predicate. Parameters and parenthesized
concise bodies use the existing rules/react helpers. The factory predicate and
comment scan are rule-local functions in the upstream rule itself. No private
copy of a shared helper or compiler change is included.

The [selected gate](testdata/selected_test.go), installed through a Go overlay,
captures 102 unique upstream cases, proves that all five witnesses fire, and
compares findings, messages, byte ranges, automatic edits, suggestions and
converged fixed source with Go on source Node, emitted JavaScript and sanitized
native. It includes the witnesses with all rules selected. Both new top-level
Go tests call t.Parallel. [run-selected.sh](testdata/run-selected.sh) reproduces
this gate and writes its output to a fresh log directory.

Witnesses cover static versus instance membership, concise object bodies,
comments before and after the arrow, strict factory recognition, plain parameters,
typed and optional parameters, destructuring, return annotations, quoted names,
class expressions, generic async arrows and Unicode offsets.

Go behavior deliberately retained: only createReactClass is an ES5 factory;
static render and instance getDerivedStateFromProps do not report; fixes decline
typed/defaulted/destructured/rest/optional parameters, return annotations and
head comments. The generic async render in testdata/unicode.ts.txt is fixed by
Go to a method without async or its type parameters; this port matches it.
Go's trailing-comment scan extends to the end of the line and is literal-blind.

The [mutant](mutant.json) removes the static lifecycle name from membership.
It compiles and disagrees with Go on source Node, emitted JavaScript and
ASan/UBSan native. The [selected log](testdata/selected.log) records all three
catches and 54,305 identical comparison bytes. The selected package reports
PASS in 315.584 seconds, including the cold sanitized checker archive and C
object builds. Structural object arrays for comment spans were refused during
development; storing each start/end pair in an ordinary number array expresses
the same algorithm and passes. No helper or language gap stopped this port.

Setup logs are /tmp/s13-react-no-arrow-function-lifecycle-setup.log and
/tmp/s13-react-no-arrow-function-lifecycle-wasi-setup.log. The first setup reports
Go ready 0.056s, Node ready 0.054s, submodules ready 0.144s, clang ready 0.455s,
go build ready 16.509s, cache warm 16.730s, and done 16.782s. The WASI setup
reports SDK ready 24.346s and done 128.259s, concurrent with the cold Go oracle
build. nproc is 5, with a four-core quota.

## Complete lint gate

The first whole-package run was interrupted when the environment reconnected.
Its retained partial log has 135 passes, no failures and one skip, but no final
package result. It is not counted as a completed gate. The required full gate
was rerun with warmed caches and a new profile directory.

The [complete log](testdata/whole.jsonl) and [machine and command record](testdata/whole-summary.json)
record a passing package: 154 passing test results, zero failures,
and one skip (48 passing top-level tests and one skipped top-level test).
The skip is TestCheckerBridgeRefusalPending: this baseline's prelude lacks
TSGoError, and the test explicitly awaits the checker error-as-value integration.
All optional input sets were supplied. Changing that unrelated compiler contract
is outside this rule's directory.

Wall time is 1570.65 seconds; nproc is 5, GOMAXPROCS=4,
and the test parallel limit is 4. Load before: 0.09 2.22 3.53 1/151 147465.
Load after: 4.96 4.99 4.79 3/158 282731.
The clean TypeScript checkout is at 050880ce59e30b356b686bd3144efe24f875ebc8;
WASI SDK 27 is installed, both benchmark switches are enabled, and both profile
variables point at the same fresh directory in the command record.

The gate exceeded ten minutes. Observed costs include 459.25 seconds
in the serial throughput, fixer and profile checks, TestRulesAgree at
516.22 seconds (including separate typed-case replays),
and the full 94-rule mutant batch at 0.04 seconds.
Parallel durations overlap; they must not be added as a wall-time decomposition.
In the first attempt, two serial native fixer rebuilds alone took about 85 seconds
each. No rule finding mismatch or rule-specific hang was observed.

The inherited compiler/stage1 comparison holds 812 manifest files, including
77 compiler files and 734 retained stage1 sources, plus its control, on Go,
Node, emitted JavaScript and sanitized native: 31,484,977 identical bytes.
The package retains its explicit malformed-parser recovery limits; those are
not claimed as fully equivalent recovered trees or converged fixes. No own
upstream case is excluded. Arbitrary uncaptured inputs are outside this proof.

Best-of-five compiler throughput, 28,338 findings: Go 13,410.83 findings/s,
native 1,803.79/s, Node 3,303.46/s. JSX throughput, 441 sources and 505 findings:
Go 5,401.21/s, native 2,987.94/s, Node 693.84/s. These measure the complete
registered suite, not this rule alone.
