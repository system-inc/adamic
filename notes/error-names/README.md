# Error names and spreads

Base: origin/runtime/error-spread e73fd72. Merged origin/runtime/uncaught-name bc75558.

No output disagreements in the final programs. Twenty new oracle fixtures: six compiled
and compared on source Node, native (sanitized and release), and the JavaScript backend;
fourteen assert NotYet. Counts are recorded only for the six compiled fixtures, as the
existing counts harness requires. observations.json records every explicit CLI build,
Node output and successful native run. Unsupported builds have native=null.

## Cases and existing coverage

| Condition or value | Existing program before this work | New program or result |
| --- | --- | --- |
| Direct Error narrowed in a catch | error_spread.a | Already covered |
| Error through an interface parameter, plain and Error callers | error_spread_view.a | Already covered |
| Direct constructed Error stored in a const | None | coverage_error_construct.a, NotYet |
| Error called without new | None | coverage_error_call.a, earlier NotYet: reading Error |
| Error name/message assigned; spread with replacement field | None | coverage_error_mutated.a, NotYet |
| Error or undefined, both callers | None | coverage_error_optional.a, NotYet after nonnullable type lookup |
| Error hidden by interface return and called expression | None | coverage_error_view_return.a, NotYet |
| RegExp literal, including changed lastIndex | None | coverage_regexp_literal.a, earlier RegExp spread NotYet |
| RegExp constructor and ordinary call | None | coverage_regexp_new.a and coverage_regexp_call.a, NotYet |
| RegExp hidden by interface parameter | None | coverage_regexp_view.a, prototypeHazard NotYet |
| Other representations hidden by interface: array, number, string, boolean, closure | None for these object spreads | coverage_view_array.a, coverage_view_number.a, coverage_view_string.a, coverage_view_boolean.a, coverage_view_closure.a, NotYet |
| Plain object with no hazard, bare spread, replacement, retained source | reuse.a, spread_snapshot.a, library_object_freeze.a | coverage_plain_view.a also isolates an interface parameter with only plain callers |
| Possibly undefined plain source, present and absent, bare/replaced/copied/reused | spread_undefined.a | Already covered |
| Uncaught default name and nonempty message | exceptions_uncaught.a, closures_throw_uncaught.a | Already covered |
| Empty name and nonempty message; both empty | uncaught_names.a, uncaught_names_empty.a from merged branch | Already covered |
| Default name, explicitly empty message | None | coverage_uncaught_default_empty.a, prints Error |
| Default name, omitted constructor argument | None | coverage_uncaught_no_argument.a, prints Error |
| Custom name and message emptied after construction | None | coverage_uncaught_custom_empty.a, prints Custom |
| Custom name and custom message set after construction | None | coverage_uncaught_custom_message.a, prints Custom1: changed 2 |
| Empty name and message changed after construction | None | coverage_uncaught_empty_changed.a, prints changed 3 |

The changed objectLiteral code first propagates an expression-lowering error, rejects a
non-object representation, checks Error after removing nullability, checks a nonempty
prototypeHazard reason, then passes safe objects to the existing spread implementation.
RegExp direct spreads stop earlier in regexUnsupportedUse. The new view programs cover
Error, RegExp and other native representations; the plain program covers no hazard.
The uncaught formatter covers all four combinations of empty/nonempty name/message,
including custom values and post-construction writes.

## Cases without an executable program

Error(...) without new cannot reach the spread or uncaught formatter: reading Error is
NotYet. Throwing a stored or parameter Error is NotYet unless it is the narrowed binding
of the surrounding catch. Initial exploratory probes demonstrated both restrictions;
final mutation probes use catch/rethrow, which reaches the formatter.

prototypeHazard's moduleOrder failure is not a runtime value case: module loading fails
before a valid oracle program reaches this branch. Its user-defined member check for the
empty member name cannot be reached with an ordinary named member. Checks conditional on
valueOf, toLocaleString, hasOwnProperty or propertyIsEnumerable do not apply to this call,
which passes the empty string. No program was invented to claim those paths were covered.
Allocation failure in adamic_uncaught cannot be deterministically forced by a short closed
oracle program without a fault-injection harness.

## Mutation

Changed one line of internal/native/runtime/exceptions.c from:

```c
size_t separator = name->length > 0 && message->length > 0 ? 2 : 0;
```

to:

```c
size_t separator = message->length > 0 ? 2 : 0;
```

coverage_uncaught_empty_changed.a failed on stderr, after compiling successfully:
Node printed `adamic: panic: changed 3\n`, native printed
`adamic: panic: : changed 3\n`, both exited 70. mutant.log records the failure.
The line was restored before the final gate; no compiler mutation is committed.

## Commands

All Go and Node verification shells sourced /workspace/adamic-tools/env.sh.
Setup: bash cloud/setup.sh (first cache-warm run encountered the temporary merge
conflict; rerun succeeded), source /workspace/adamic-tools/env.sh, nproc.

```sh
git fetch origin main runtime/error-spread runtime/uncaught-name
git fetch origin runtime/error-spread:refs/remotes/origin/runtime/error-spread runtime/uncaught-name:refs/remotes/origin/runtime/uncaught-name
git log --oneline origin/main..origin/runtime/error-spread
git log --format=full origin/main..origin/runtime/error-spread
git diff origin/main...origin/runtime/error-spread
git log --oneline origin/main..origin/runtime/uncaught-name
git log --format=full origin/main..origin/runtime/uncaught-name
git diff origin/main...origin/runtime/uncaught-name
git switch -c coverage/error-names origin/runtime/error-spread
git merge origin/runtime/uncaught-name -m 'Merge uncaught error name reporting for coverage'
# Retained both fixture lists in the merge conflict, then:
git add internal/oracle/oracle_test.go
git commit --no-edit

go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_' -count=1 -v
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_view_' -count=1 -v
# The coverage_ command was also repeated after restoring the mutant.
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
# With the one-line mutant, then restored:
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_uncaught_empty_changed.a' -count=1 -v

gofmt -w internal/oracle/oracle_test.go
gofmt -w internal/flow/flow_test.go
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
```

The first coverage_ oracle invocation exposed multi-argument console.log (the library
accepts one argument) and unsupported stored/parameter throws. Those probes were corrected
and the same oracle command rerun. The explicit builds were repeated on corrected files.
For every coverage_*.a file, a Python subprocess loop ran these commands, capturing
stdout, stderr and exit independently (the final results are observations.json):

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs /workspace/adamic/internal/oracle/testdata/<stem>.a
go run ./cmd/adamic build internal/oracle/testdata/<stem>.a -o /tmp/adamic-gate/<stem>
/tmp/adamic-gate/<stem> # only after a successful build
```

Tests wrote logs under /tmp/adamic-gate/error-names-*.log and the completed logs were read.
Source inspection used cat, sed, rg and git status/diff; process inspection used ps.

The first full gate exposed flow's glob picking up the NotYet fixtures, including the
branch's existing error_spread.a and error_spread_view.a. internal/flow/flow_test.go now
explicitly excludes those sixteen files from its compiled-program discovery. The oracle
still asserts each NotYet result; other unexpected lowering failures remain failures.
After this harness correction, format, vet and the full uncached gate were rerun.

The superseded first gate reported flow failures and a passing oracle package, but was
stopped (exit 143) after the corrected rerun passed flow. A Python process-tree walk
checked the original go-test PID 8548, parent 7005 and original log filename, then sent
SIGTERM to that run and its thirteen remaining descendants. The final rerun was untouched.

## Final verification

The corrected full uncached gate exited 0. All packages passed, including flow,
native, oracle, Unicode properties and every stage-1 package. gate.log contains the
full package results. gofmt -l cmd internal and go vet ./... produced no output;
counts.log records the successful count update. The restored 20-fixture oracle run
is in oracle.log. No unmutated program produced a differing native output.

Commits and pushes used these commands (the first push succeeded; no retry was needed):

```sh
git commit -m 'Cover error spreads and uncaught names with oracle programs'
git push -u origin coverage/error-names
git commit -m 'Record the complete uncached verification results'
git push origin coverage/error-names
```
