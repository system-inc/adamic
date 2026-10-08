Built: a general loud check for evaluated never values, before case (b) can rely on them.
Commits: implementation ae694d19 and follow-ups 7018919c/e45055dd; merged main 4e0bfda5 in 287c0567 and 71d7e491 in 2597e3ab.
Commands and outputs: lower PASS; focused uncached Node/native/backend probes PASS; final merged gate recorded below.
Mutants: lowering never as a no-op and skipping a never call both failed semantic assertions with valid C.
Not covered: the pending checked optional-view dependency, checked optional reads/writes and their erasure proofs, or unrelated unsupported syntax/signatures.

The October 7 05:50 ruling requires a runtime stop when the checker says never but
compiler reachability is not proven. No new unreachability erasure proof is used:
if such an expression is evaluated, it stops. An actually dead branch never runs
its check. The optional structural refusal remains active pending shared checked
views, and the write allowance for source never is not enabled separately.

## Existing never paths observed

- Identifier reads used their local's stored representation despite a flow type of
  never. Explicit never locals instead failed typeOf with NotYet.
- Function signatures treated a never return as void. Expression-bodied arrows
  returning a narrowed never had an explicit NotYet boundary. That boundary stays
  intact; assigning all never types a machine representation made a filter callback
  incorrectly reach the truthiness refusal during development, so that change was
  removed. Only declaration storage receives an inert placeholder.
- The prelude's panic already emitted ir.Panic, including return panic(...).
  It remains the original panic, with its effects and message.
- Native adamic_unreachable already stopped a function that unexpectedly fell off
  its end, but named a compiler bug rather than the source never expression.
- JSON schema handling treated never as undefined; its value lowering now reaches
  the same general expression check. Static predicate exhaustiveness and class
  relation checks also inspect Never flags, but are not runtime value checks.
- The open-numeric-enum check exists on codex/flag-enums at ec67b02f, in
  internal/lower/enum_never.go. It recovers enum identity through stored types and
  never aliases and emits an ordinary IR helper with Panic. Main and this branch
  still refuse enum declarations. The enum worker's check was inspected, not
  merged or replaced; enum integration remains to be checked when it lands.

## Added paths and independent observations

internal/lower/never.go intercepts evaluated never at the shared value entry,
including callers that use value directly. Expressions used as statements reach
it too. A never local has placeholder storage; no evaluated read can consume that
placeholder. Field and indexed reads evaluate their receiver/index in order, then
stop before loading an impossible physical representation. Other supported value
expressions retain their evaluation. Calls execute before the trap; a legitimate
panic or throw leaves the call normally and preserves its original behavior. Void
calls are replayed in the helper with argument locals rather than used as values.
A return of never from a void/never function evaluates the check without returning
a spurious machine value. Contextual, implied and cast targets choose a helper
result representation that is never observed after the panic.

never_reached.a uses a call to invalidate TypeScript's narrowing of a global
string-or-number variable. Node prints continued and exits 0. Sanitized native,
release native and the JavaScript backend all print nothing and exit 70 with this
complete pinned diagnostic:

```
adamic: panic: unreachable expression value at never_reached.a:5:31
```

Additional Node-held probes do the same through a boolean property, a boolean
array element, and identifier/property/element never values used as if conditions. A contextual result or the
identifier's stored representation keeps the unreachable check usable in that
condition without inventing a number-as-condition refusal. never_call.a prints called and panics with kept original on Node
and both backends, proving a real never call is not skipped. The two registered
fixture rows are new; no existing counts changed before the main merge.

The no-op mutant replaces the generated helper's Panic by return 0. All three
compiled runs then print continued and exit 0, failing TestNeverReached's pinned
exit/output assertion. The skip-call mutant removes the helper's effect evaluation;
the source still prints called and the original panic, while both backends lose
that output and produce the unreachable diagnostic. No clang rejection killed
either mutant. The source was restored after each run. The runnable mutations and
logs are beside this report under evidence/never-*.

## Commands, timing and limits

Every test wrote its output to a log. The environment file was
/workspace/adamic-tools/env.sh. Setup printed Go ready 1s, clang ready 2s, Node
ready 2s, submodules ready 3s, build cache warm 231s, done 231s. nproc was 5 and
cpu.max was 400000 100000 (four CPUs).

```
go test ./internal/lower -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNeverReached|TestNativeAgreesWithNode/internal/oracle/testdata/never_' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNeverReadKinds$' -count=1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 python3 docs/optional-widening/evidence/never-mutants.py
go vet ./internal/lower ./internal/oracle
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./stage3/fixtures ./internal/oracle -count=1 -timeout 30m
```

Observed before main merge: lower 16.989s, focused oracle 0.669s, property/index
probes 1.543s, counts 23.751s, restored focused oracle 1.604s, vet exit 0.
The first complete uncached oracle ran 216.998s and passed semantic tests but
failed for the missing new count row; the contemporaneous lower run exposed the
arrow boundary regression described above. Neither failed development run is
claimed green. Final results after merging main follow below.

This unit does not claim the full repository gate or stage-1 lint/corpus coverage.
Unsupported signatures and syntax remain explicit NotYet boundaries. No proof of
all 415 stage3 optional relation sites, nor a count of newly compiling optional
views, is claimed. The published interface-downcast dependency still lacks the
shared transitive optional-view entry point; (a), (c), checked reads, and the new
checked writes remain pending that dependency.

## Main merge and count evidence

Main 4e0bfda5 introduced field_access_paths.a, whose aliased absentExtra did not
declare amount. Keeping the prerequisite refusal made that old successful fixture
fail at its call to show. Its source now explicitly declares optional amount in
its object annotation. The same object fields and layout test remain; no field is
added at runtime. Stock Node's stripTypeScriptTypes ran main's exact source and the
annotated source independently: cmp passed for stdout (96 bytes) and stderr (empty).
The focused merged oracle then passed in 2.738s, including the field-layout test.

Count regeneration after main passed in 33.719s and restored canonical row order.
Two new never rows and the previous optional fixture rows are retained. Three
existing numerical rows changed relative to the merged records:

| Fixture | Before | Measured after main |
|---|---|---|
| user_iterators.a | 669 alloc, 669 free, 463 retain, 912 release, peak 67 | 663 alloc, 663 free, 455 retain, 904 release, peak 64 |
| optional_widening_boolean_as_number_array.a | 8 alloc, 8 free, 7 retain, 16 release, peak 6 | 8 alloc, 8 free, 6 retain, 15 release, peak 6 |
| optional_widening_boolean_as_number_return.a | 5 alloc, 5 free, 1 retain, 8 release, peak 4 | 5 alloc, 5 free, 0 retain, 7 release, peak 4 |

A scratch Go overlay replaced expression.go and statements.go by their exact main
4e0bfda5 versions, and replaced never.go by an empty package file. Counted native
runs of those three fixtures yielded precisely the measured-after-main rows above
(3.021s). Thus those shifts also occur with the new never checks disabled; they are
not attributed to this check. No allocation optimization is claimed here.

## Final restored checks

The complete merged uncached three-package run on 7018919c passed:
internal/lower 65.123s, stage3/fixtures 45.139s, internal/oracle 241.237s.
The final condition-context follow-up e45055dd then passed the complete lower
package (40.486s), all focused uncached never and field-layout oracle probes
(3.236s), and the complete counts check (48.661s). Final vet and whole cmd/internal
format logs are empty; git diff --check is clean. The last context extension was
verified with those focused probes, not a repeated complete Node oracle.

The final no-op mutant ran all TestNever probes. Identifier, property, element,
and condition paths all lost their required exit-70 stop; the primary fixture
printed continued and exited 0 in sanitized native, release native and the backend.
The final skip-call mutant still failed the independent Node comparison. Both
compiled valid C, and both changes were restored before the final checks.

Main 71d7e491 was merged as 2597e3ab. The final package checks passed: lower 41.326s, fixtures 32.554s, focused oracle 3.841s, counts 26.376s. Interface-downcasts remains at 6b50efe0.
Only codex/optional-widening-2 is pushed for this work; refusal-2 remains intact.
No PR is opened, no main/area branch is changed, and no cohere source is copied.

The October 7 06:55 ruling supersedes write refusals: optional wider-view writes must check the actual runtime shape and slot contract. This document reports only the independent never prerequisite; checked writes are not implemented here. Nested never constituents do not qualify as unreachable expressions.

After source reduction 6b3f285e, the complete uncached lower/fixtures/oracle gate was repeated and passed 49.747s/37.896s/236.513s. This supersedes the earlier focused-only validation limit.
