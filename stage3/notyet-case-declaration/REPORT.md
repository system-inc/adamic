Built switch-wide case bindings with local readiness and capture cells; Node decides behavior.
Commits: compiler base b410340d, replay merge a3bed8f1, fixture reduction a66178c6; implementation follows this report.
Focused source/backend/release/sanitized oracle passed (1.419s); counts refresh passed (22.265s).
Nine independent mutants failed output or exit comparisons; no build-warning or lowering-refusal kills counted.
Coverage: 63 unique declaration sites from 72 CSV observations; full tsc lowering and unrelated refusals remain uncovered.

Assumption: start at origin/main, then merge the specifically requested current
origin/area/compiler. The compiler pin resolved to b410340dc8f889b5799c3bc519117c63def3aa24.
Census replay 9a1f14c5d994aa855625e7cfa295677060348fec was merged without rewriting history.

Direct let/const bindings have storage and a Boolean readiness local around the
whole switch, recreated at every entry. Their initializer remains at its source
case. The readiness local is captured with the binding. Existing Checked reads
and writes use it, and reuse the existing ReferenceError check in both backends.
Explicit blocks keep their own scope. The discriminant is evaluated once before
entering the case scope. Destructured bindings initialize separately in order.

The two binder reductions cover a discriminated export and parameter-index lookup.
The latter spells numeric concatenation as a template to avoid the independent
string-plus-number stop. Scope witnesses cover fallthrough, shadowing, forward
captures, destructuring, skipped declarations, RHS-before-write checking, fresh
closure cells on re-entry, continue, break, and declarations without initializers.
Finished fixtures are leak-checked. No designed refusal was relaxed.

The source oracle required a correction: Node transform mode rewrites a later
case binding to label1/value1 but leaves earlier captures using label/value.
Node strip mode leaves lexical identity intact. oracle/node.mjs now strips
 erasable source and falls back to transform only for ERR_UNSUPPORTED_TYPESCRIPT_SYNTAX.
The existing enum and parameter-property fixtures hold that fallback. Reverting
the strip choice fails the switch scope fixture on source output. This is an
observed source-transformation bug, not an Adamic semantic exception.

Commands ran with /workspace/adamic-tools/env.sh sourced; all test output went
to files. No whole-package test or full gate ran.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSwitchCaseSourceReductions|TestNativeAgreesWithNode/internal/oracle/testdata/(switch_case_|enums\.a|parameter_properties\.a)' -count=1 -v -timeout 10m
python3 internal/oracle/testdata/run-switch-case-mutants.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
```

Logs: /tmp/notyet-case-final.log, /tmp/notyet-case-mutants.log,
/tmp/adamic-switch-case-mutants/*.log, /tmp/notyet-case-counts.log.
Counts gained exactly seven rows; existing rows did not move.

| Mutant | Fixture | Catcher |
| --- | --- | --- |
| native-read | switch_case_read_tdz | Node exit-code comparison |
| native-write | switch_case_write_tdz | Node exit-code comparison, RHS still prints first |
| javascript-read | switch_case_read_tdz | source/backend exit-code comparison |
| javascript-write | switch_case_write_tdz | source/backend exit-code comparison |
| early-ready | switch_case_capture_tdz | premature access fails Node comparison |
| entry-ready | switch_case_reentry_tdz | second-entry access fails Node comparison |
| never-ready | switch_case_scope | valid initialized reads fail Node comparison |
| initializer-twice | switch_case_scope | extra initializer stdout |
| oracle-transform | switch_case_scope | source transformation changes lexical identity |

Replay command, used before and after with the same corpus:

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-case-adapted/src/tsc/tsc.ts -where /tmp/notyet-case-adapted/src/compiler/binder.ts:372:13 -kind NotYet -reason 'a declaration directly in a case (wrap the case in a block)'
# same command for binder.ts:731:17
```

372:13 reproduced before (exit 0). Afterward the declaration stop disappears;
its initializer reaches Refused 'a cast the runtime can't check' at 372:39,
then census rollback reaches NotYet 'a BinaryExpression with a boolean and a value'
at 373:17. The original earlier cast refusal at 359:29 also remains.
731:17 never reproduced independently on this latest base: both before and
after select getDeclarationName and stop at 661:14, NotYet
'a function returning __String | undefined'. The after commands exit 1 because
the old requested signature no longer reproduces; that exit is not a build failure.
Raw logs: /tmp/notyet-case-replay-{372,731}-{before,after}.log.

Coverage is a syntactic root-reason claim, not a claim that 63 whole units now
compile. coverage.json deduplicates the pinned roots/raw.csv by
(kind, where, reason, text), exactly as the table does. Unchecked casts, other
initializer operations, captured slotless values, and recursive closure cycles
retain their existing refusals or NotYet stops.

Setup first failed during a concurrent base merge with missing typedArrayWrite,
censusFieldSlotless, and parameterProperty symbols. Retried on the settled base:
Go ready 0.029s, submodules 0.069s, markdown ready 0.082s, clang ready 0.208s,
build ready 55.340s, cache warm 55.444s, done 55.470s. nproc=5, CPU quota=4.
Node v24.19.0, Go 1.27.1, clang 20.1.8. Setup logs are
/tmp/notyet-case-setup.log and /tmp/notyet-case-setup-retry.log.
