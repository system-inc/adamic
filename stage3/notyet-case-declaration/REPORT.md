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
to files. The final touched-package tests below supersede the initial focused-only run; no full gate ran.

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

## Optional case representations

The eight-site reason compares number | undefined (an optional enum, Map.get,
array element or optional discriminant) with a present numeric case. Present
case constants now carry the scrutinee's pair representation. The absent value
is retained and does not accidentally match zero or false. The same operation
also handles boolean | undefined. Other representation mismatches stay NotYet.
The production diff is four lines in switchStatement; no backend change.

Two reductions from checker.ts:19130 and :19144 use Map.get and string mapping,
including absent entries and a template/string-array mapping. Node source tests
pass. Both backends, release C, sanitized C and leaks pass (0.502s). The
optional-tag mutant makes a case absent; optional-presence forces the scrutinee
to present zero. Both fail stdout comparisons, with no build/refusal kill.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSwitchCaseOptionalSources|TestNativeAgreesWithNode/internal/oracle/testdata/switch_case_optional_' -count=1 -v -timeout 10m
python3 internal/oracle/testdata/run-switch-case-mutants.py optional-tag optional-presence
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
```

Counts pass in 30.033s and add only the two new rows. Fixture step b8cbb9e5 was
pushed separately. Logs: /tmp/notyet-case-optional-{before,source,after,mutants,final,counts}.log.

Exact requested replays for checker.ts:19130:13 and :19144:13, using the same
-project entry and reason "a case whose type differs from the switch's", each
exit 1 both before and after: Refused 'a cast the runtime can't check' at
19129:40 and 19143:40 respectively. The old reason is hidden behind this earlier
designed refusal on the latest compiler base. No claim of a reproduced before
signature or full checker compilation is made. Raw replay logs are
/tmp/notyet-case-replay-{19130,19144}-{before,after}.log. The original eight unique
CSV sites are counted as the optional numeric-case lesson, not eight compiled units.


## Undefined constant case

semver.ts:380:13 is literal undefined, not a dynamic expression. Its exact
before replay reproduced the named stop. switchStatement now fits the constant
to the scrutinee representation: null reference for optional strings, absent
pair for optional numbers and booleans. Dynamic expressions remain NotYet.
The standalone Node reduction tests grouped labels, default placement, empty
strings, zero, false and NaN. Source, JavaScript, release native, sanitized
native and leak checks pass in 0.400s.

undefined-string replaces the absent string constant with "="; undefined-pair
replaces numeric absence with present zero. Both fail Node stdout comparisons.
The first numeric mutant attempt used a nonexistent IR type and failed to build;
that was rejected as evidence, corrected to IsMaybe/Present, and rerun to an
actual oracle comparison failure. All mutants were restored.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSwitchCaseUndefinedSource|TestNativeAgreesWithNode/internal/oracle/testdata/switch_case_undefined' -count=1 -v -timeout 10m
python3 internal/oracle/testdata/run-switch-case-mutants.py undefined-string undefined-pair
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
go run ./stage3/census/latent/replay -project /tmp/notyet-case-adapted/src/tsc/tsc.ts -where /tmp/notyet-case-adapted/src/compiler/semver.ts:380:13 -kind NotYet -reason "a case that isn't a constant"
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -count=1 -timeout 30m
```

Counts refresh passes in 25.393s and adds exactly one fixture row. Raw CSV has
one observation and one unique site for this reason: total coverage across the
three requested reasons is 63 + 8 + 1 = 72 unique reason-sites. None was skipped
for territory; no designed refusal was relaxed and no runtime C file changed.
Files outside switchStatement in the overall implementation: internal/ir/ir.go,
internal/lower/locals.go, internal/lower/switch_declarations.go,
internal/native/emit_locals.go, internal/javascript/javascript.go, oracle/node.mjs,
the three independent oracle registration/source test files, ten .a fixtures,
the mutant runner, counts.md and this report/coverage.json. Other workers' lower
functions were not edited; remote notyet ownership logs were checked.
Logs: /tmp/notyet-case-undefined-{before,after,mutants,pair-mutant,counts}.log,
/tmp/notyet-case-replay-semver-{before,after}.log, /tmp/notyet-case-packages.log.

After replay: the parseComparator switch lowers; no NotYet remains in its source
body (lines 336-400). The requested signature therefore exits 1 as disappeared.
Earlier helper findings remain, including NotYet assigning to an
ObjectLiteralExpression at semver.ts:60:14 and reading major at 130:28.
