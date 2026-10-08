# Proven rest arguments

Admit boxed union rest arrays, callable/generic rest, named closure expressions and proven fixed tuple dispatch.
Commits: replay 8b0db31c, union acc99382, callable 6415c4cf, named 563fe969; tuple implementation carries this report.
Focused Node, JavaScript, release native and sanitized native oracles pass; seven new positive count rows.
Admission, packing, calling-convention, tuple length/shape/escape/forwarding mutants all fail their focused fixtures.
Captured-callable cycle refusal is preserved; nested declarations need their existing frame prerequisite; no full gate or whole-project compilation.

The unit-specific compiler-area base overrides the generic main-base instruction.
The replay merge includes 9a1f14c5d994aa855625e7cfa295677060348fec.
The table is origin/codex/stage3-notyet-table at 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7.
Its raw.csv and roots/raw.csv each contain 353 matching observations at 46 unique
(kind, where, reason, text) root sites. This is representation-rule coverage,
not a claim that all 46 bodies or their whole project now compile.

Only the rest admission guard changes. censusRestCall already fits scalar items
into ir.Union boxes and copies spreads into new arrays. Existing array operation
guards remain, so admitting the parameter does not accept unsupported operations.
The two .a fixtures reduce binder.ts:567:78 and binder.ts:2746:72 to their fixed
arguments and forwarding behavior. They cover empty rest, string/number/undefined
items, runtime-built strings, evaluation order, repeated forwarding, and cleanup.
Their observations are lengths and side effects; they do not observe union payload
reads, which still stop at the general union-array element guard.

## Replay evidence

Prepared the table's recorded stage3 snapshot 9d534d3a in /tmp/rest-exact-input,
then ran its apply.sh into /tmp/rest-exact-adapted. All 81 source hashes match
stage3/notyet-table/source-manifest.json. Earlier setup probes used a different
adaptation snapshot; only the hash-matched runs below are delivery evidence.

With the old admission guard restored temporarily, both exact commands reproduced
NotYet a rest array of DiagnosticArguments. On the implementation:

- binder.ts:567:78 reaches NotYet a BinaryExpression with a value and a value
  at binder.ts:568:52 (getSourceFileOfNode(node) || file).
- binder.ts:2746:72 has no findings in its selected census unit.

Commands, from the repository root, with the setup environment sourced:

```sh
go run ./stage3/census/latent/replay -project /tmp/rest-exact-adapted/src/tsc/tsc.ts -where /tmp/rest-exact-adapted/src/compiler/binder.ts:567:78 -kind NotYet -reason 'a rest array of DiagnosticArguments'
go run ./stage3/census/latent/replay -project /tmp/rest-exact-adapted/src/tsc/tsc.ts -where /tmp/rest-exact-adapted/src/compiler/binder.ts:2746:72 -kind NotYet -reason 'a rest array of DiagnosticArguments'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/rest_union_' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
```

Baseline replay exits 0. Post-change replay exits 1 because the requested old
signature is absent; stdout still records the selected unit and its new findings.
Focused oracle: ok in 15.777s, and restored final run ok in 0.559s.
Counts update: ok in 32.028s. New counts (alloc/free/retain/release/peak/region):
diagnostic 16/16/7/21/8/0; token 28/28/24/38/12/0.
Every test wrote directly to a log. /tmp/rest-mutant-reject-union.log records
both lowering failures; /tmp/rest-mutant-drop-rest-items.log records stdout
mismatches and successful compilation, so this mutant was not killed by warnings.
The original source was restored in a finally block after each mutant sequence.
/tmp/rest-exact-{before,after}-{567,2746}.{json,log} preserve replay output.
/tmp/rest-final.log and /tmp/rest-counts.log preserve validation.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh, then source
/workspace/adamic-tools/env.sh. No setup workaround was needed. Timing seconds:
Go 0.261, Node 0.308, clang 0.990, markdown 2.959, submodules 29.210,
Go build 330.333, cache warm 330.689, done 330.881. nproc=5,
cpu.max=400000 100000. /tmp/notyet-rest-setup.log contains the full output.
No PR was opened and no whole package or full gate was run.

## Function-value and generic rest step

The six table roots have six raw observations. The supported rule now packs array
rest arguments for arrows, ordinary function values, and specialized named generic
functions. Both call paths share packing; fixed overloads retain implementation
rest packing. Named closures remain reserved for the following step.

Fixtures rest_callable_generic.a and rest_callable_write.a run through the existing
Node, JavaScript, native release and sanitized native oracle. Source-only probes
pin the existing captured-callable cycle Refused and two adapter NotYet boundaries.
The reduced not wrapper is refused for a possible reference-counting cycle; changing
that ownership ruling requires a cycle-safe captured callable contract.

Exact replays use the same command above with core.ts:2480:13 and program.ts:582:43
and reason 'a rest parameter outside a nongeneric named function'. Both reproduced
the baseline. Afterwards core reaches 'a rest parameter other than an array' (its
uninstantiated generic tuple constraint); program reaches 'a rest callable seen
through a different calling convention' (optional fixed arguments in its contextual
signature). These are next stops, not claims that the source bodies compile.

Mutants, each restored: bypass closure packing makes writer forwarding stop at
SpreadElement; remove generic declaration discovery makes generic relay forwarding
stop at SpreadElement; bypass rest-convention equality accepts rest_callable_view.a
and fails its expected adapter boundary; bypass lexical contextual-convention check
makes rest_callable_context.a fail its expected named boundary. Logs are
/tmp/rest-mutant-{closure-pack,generic-discovery,callable-view,context-view}.log.

Focused commands: ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'TestRestCallableViewsNeedAdapters|TestRestNotKeepsCycleRefusal|TestNativeAgreesWithNode/internal/oracle/testdata/(rest_(union|callable)_|census_overload_contracts)'
-count=1 -timeout 10m; go test ./internal/lower -run
'TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestCensus' -count=1 -timeout 10m.
Counts: go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts.
Logs /tmp/rest-callable-{final,lower,counts}.log. The first counts attempt exposed
a fixed-overload packing regression; implementation-signature fallback corrected it.
No IR, backend or runtime C changes were needed for array rest callable support.

Callable restored validation: oracle ok 1.685s; lower ok 1.717s; counts ok 37.390s.

## Named closure step

Fixture commit ba766d51 reduces enqueue and reportIncompatibleError to existing
counted named function expressions. Removing the named-rest guard preserves the
same packed-array convention. Focused oracle ok 0.601s; named guard restoration
mutant fails at rest_named.a:4:45; counts refresh ok 26.456s, new row
20/20/22/41/12/0. Commands: ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
-run TestNativeAgreesWithNode/internal/oracle/testdata/rest_named.a -count=1
-timeout 10m; counts command as above.

The table has four unique named-closure roots and four observations. Both exact
examples already did not reproduce on the selected area/compiler base before this
step: checker.ts:22486:70 reaches 'an array of never' at 22489:36; core.ts:1577:22
reaches 'a rest array of T[]' at that parameter. After replays give the same stops.
No table unlock is claimed for those two examples. Full nested declarations still
stop at 'a function inside a function (a closure)' on this base. Their counted frame
environment prerequisite b15216da is on the table branch, outside this small rest
unit; the existing function-expression backend needs no change.
Logs /tmp/rest-named-{before,after,counts}.log and
/tmp/rest-mutant-named-closure.log; replay /tmp/rest-closure-{core,checker}-{before,after}.{json,log}.

## Fixed tuple rest step

Fixture commit c0f3182b reduces the two constructor-type dispatchers. Tuple rest
is packed into the existing numbered-field object storage with an explicit length
field. A source-use proof allows only length/constant-index reads and sole-spread
forwarding; it refuses whole-tuple escape, writes and methods. Consequently the
JavaScript object storage cannot expose a different array identity. Only distinct
fixed arities with represented field types are admitted. Named fixed calls can
forward a sole spread of a plain fixed tuple by reading its numbered fields.
Optional, variadic, overlapping-arity shapes and spreads into tuple rest stay NotYet.
No backend, IR or C runtime edits are needed.

Both exact baseline replays reproduced 'a rest parameter other than an array'.
Afterwards createConstructorTypeNode reaches 'a void call used as a value' at
nodeFactory.ts:2366:13; updateConstructorTypeNode reaches the same stop at
2402:13, both the Debug.fail fallback. Commands use the replay command above
with factory/nodeFactory.ts:2363:40 and :2399:40 and the original reason.
The raw CSV has two observations at two unique roots for this reason.

Mutants restored after each run: restore tuple admission refusal (both positives
fail lowering); increment stored length (JavaScript output/exit comparisons and
native ASan fail); disable fixed tuple forwarding (SpreadElement stop); bypass
source-use proof (mutation and escape boundary fixtures incorrectly lower); allow
duplicate arities (ambiguous fixture reaches a different unsafe representation
boundary); ignore fixed-length proof (optional fixture reaches a later tuple-length
stop instead of the parameter guard). All final mutant runs build successfully.
Logs /tmp/rest-mutant-tuple-{admission,length,forward,escape,shape,optional}.log.

Restored all-rest oracle ok 2.881s; lower checks ok 3.110s; counts refresh ok
44.506s. New rows create 12/12/2/12/6/0 and update 14/14/2/10/6/0.
Commands: ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'TestRestTupleStorageBoundaries|TestRestCallableViewsNeedAdapters|TestRestNotKeepsCycleRefusal|TestNativeAgreesWithNode/internal/oracle/testdata/(rest_|census_overload_contracts)'
-count=1 -timeout 10m; lower and counts commands as above. Logs
/tmp/rest-all-final.log, /tmp/rest-tuple-{lower,counts}.log; replays
/tmp/rest-tuple-{create,update}-{before,after}.{json,log}.

Final raw-table accounting: 46 DiagnosticArguments roots (353 observations),
six outside-nongeneric roots, four named-closure roots, two nonarray roots.
The last three reasons have respectively six, four and two observations.
These 58 listed roots describe the admission families; the named-closure baseline
was already past that stop, and next-stop advancement is not whole-project compilation.

Final extraction regression mutants also run: original union admission restored
fails both union fixtures; dropping scalar rest items fails stdout in JavaScript
and native after successful compilation; outside-nongeneric admission restored
fails writer and generic fixtures. Logs /tmp/rest-mutant-{reject-union-final,
drop-rest-items-final,outside-admission}.log. All sources restored.
Worker branch histories were checked across 20 codex/notyet-* refs before shared
lowering edits. Other recent expression changes touched enumNeverValue, prefix
updates, logical/equality combination and void-call handling, not the rest helpers,
sameKeeping, callClosure or callFunction edited here.
