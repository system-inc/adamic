# no-extra-bind port

Base: 6bf7bcec73a04b85900c5a69a257b3cf34b35a40.
Go cohere oracle: 7945d102a6c18dd36adf9114a758ce646e8b2359.
TypeScript corpus: 050880ce59e30b356b686bd3144efe24f875ebc8, clean checkout.

The source-file Go walk is expressed as listeners on function expressions and arrows.
Function scope scans include parameter defaults, cross arrows, and stop at other
functions, methods, accessors, constructors and classes. Static bind properties,
parenthesis climbs, callee identity, exact argument count and spread refusal match Go.

The shared scanner supplies removal token starts; shared comments.forFile supplies
comment ranges. Byte bounds for the comment check use the runtime utf8Length helper.
Two independent automatic edits preserve parentheses between access and call. The
rule has no options or suggestions. No upstream code or shared module is changed.

Observed selected parity: 76 unique captured upstream source/file/options cases,
plus six owned witnesses, and the inherited corpus. Go, source Node, emitted
JavaScript and ASan/UBSan native compare byte-for-byte. Selected tests passed in
382.563s; log: /tmp/s13-no-extra-bind-selected-final.log. The initial selected run
failed on Go dependency download stderr; rerunning after download passed.

The arrow-binding-is-observable mutant removes reporting on arrows, compiles and
finishes on Node and emitted JavaScript, and disagrees with Go on arrows.ts.txt.
Log: /tmp/s13-no-extra-bind-mutant.log. The existing shared mutant harness runs
ordinary syntax mutants on those two runtimes; native parity is tested separately.

Go's intentionally narrow side-effect-free argument set is retained: bare template,
negative number and arrow arguments report but receive no fix, despite their inert
appearance. Tests pin these cases. No upstream behavior was corrected in the port.

Whole gate inputs: /tmp/s13-no-extra-bind-gate-inputs.log.
Whole gate output: /tmp/s13-no-extra-bind-whole.log.
Whole gate metrics: /tmp/s13-no-extra-bind-whole-metrics.log.
Setup output and timings: /tmp/s13-no-extra-bind-setup.log and
/tmp/s13-no-extra-bind-wasi-setup.log (nproc=5).
Remote duplicate check: /tmp/s13-no-extra-bind-existing.log (empty).

Whole lint package: exit 0, 1081.249s reported by Go, 1083s shell wall time.
152 passing result records including direct subtests, 0 failures, 1 skip;
top-level totals are 46 passes, 0 failures, 1 skip. The intentional failing
subprocess logged by TestEmittedJavaScriptMismatch is comparator evidence, not
an outer-suite failure. The only skip is TestCheckerBridgeRefusalPending,
which explicitly awaits codex/tsgo-errors-as-values. No missing input caused a
skip. Profile snapshots agree with Go. End load: 5.50 / 5.67 / 4.55; nproc=5.

Commands (from repository root, setup environment sourced and GOPROXY set):
- go run ./cmd/lint-registry
- go test -count=1 -v -timeout 30m ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses'
- go test -count=1 -v -timeout 30m ./stage1/cohere/lint -run '^TestMutants$/^arrow-binding-is-observable$'
- go test -count=1 -v -timeout 30m ./stage1/cohere/lint

There is no rule-specific blocker. No new top-level Go tests were introduced.
