# For-of object roots

Base `b410340dc8f889b5799c3bc519117c63def3aa24`; replay merge
`c68b6ceb0bd43283c6b919f8bc2c823084bd69e2` (includes `9a1f14c5`).
**2 of the 118 original object-iteration gates now advance**, after the fixed
primitive tuple and private library-iterator view changes below. The named
NodeArray examples still reproduce.

## Observed scope

The raw `roots/raw.csv` on `origin/codex/stage3-notyet-table` has 128 attempt rows
with exact reason `for...of over an object`, deduplicating to 118 diagnostic sites.
Stock TypeScript 6.0.3's checker on the complete adapted tsc entry identifies:

| Type | Sites |
| --- | ---: |
| NodeArray | 103 |
| JSDocArray | 6 |
| MutableNodeArray | 1 |
| Iterable<T> | 4 |
| Iterable<Symbol> | 1 |
| ElaborationIterator | 1 |
| readonly string tuple | 1 |
| SortedReadonlyArray<EmitHelper> | 1 |

[Per-site checker observations](site-types.csv). All 81 source SHA-256 values match
`stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json`.
These are source-type observations, not proof of native compilation.

Both examples are NodeArray: binder:432:37 is `NodeArray<Statement>`;
binder:1296:36 is `NodeArray<Expression>`. They are not Map or Set values.
The current compiler already lowers concrete custom iterables through iteration.go.
Its receiver, hidden-return and generic-origin guards were left intact.

## Why the two examples stay stopped

NodeArray extends ReadonlyArray and carries pos, end, hasTrailingComma and
transformFlags. Adamic represents that interface as an object, but the library
array iterator is inherited, not an own closure field. Treating the object pointer
as an array pointer would invent a native layout. Treating the inherited iterator
as an own field would invent a runtime method. Neither is sound.

The reductions keep that interface shape and the source loop, using an arrow to
make the loop lower before the unrelated metadata construction. Node prints `2`
and `true`, respectively. TestForOfObjectNodeArrayStops pins the exact NotYet and
nil IR. Its independent mutant replaces the object stop with an empty statement
list. Both cases fail the stop assertion; the later array/object intersection
construction becomes the next stop. Production code is restored after the run.

The requested release/sanitized native and JavaScript comparisons cannot run for
these reductions: lowering returns no IR. They are explicitly registered as
stopped fixtures, with separate source-Node checks. No green backend comparison
or completed object lowering is claimed.

The design question is whether NodeArray's array metadata construction and writes
are admitted as fixed array storage, or adapted to an object containing an array.
`docs/0.1.md` refuses expandos and prototype mutation by design. A ruling admitting
fixed metadata on arrays would need a representation and ownership design, plus
all metadata accesses and construction held to Node. This unit does not weaken
that refusal or assume that every structural Iterable has an array layout.
The five structural Iterable roots additionally need protocol dispatch that
preserves built-in, class and closure receiver conventions and optional return.
The tuple and SortedReadonlyArray roots need their own storage proofs.

## Commands and results

Every test writes to a log; no whole package or full gate was run.
Environment: Node 24.19.0, Go 1.27.1, clang 20.1.8; nproc 5, CPU quota 4.
`export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh` succeeded.
Timing lines: node 0.059s, go 0.071s, clang 0.462s, markdown 1.068s,
submodules 15.407s, go build 292.079s, cache warm 292.188s, done 292.221s.
Source `/workspace/adamic-tools/env.sh` in each build/test shell.

```
go run ./stage3/census/latent/replay -project /tmp/for-of-adapted/src/tsc/tsc.ts -where /tmp/for-of-adapted/src/compiler/binder.ts:432:37 -kind NotYet -reason 'for...of over an object'
go run ./stage3/census/latent/replay -project /tmp/for-of-adapted/src/tsc/tsc.ts -where /tmp/for-of-adapted/src/compiler/binder.ts:1296:36 -kind NotYet -reason 'for...of over an object'
```

Both exit 0 and print `reproduced NotYet: for...of over an object`.
Total times 125.841s and 120.009s include cold overlay compilation overlapping setup;
load/register/lower times are 4.879s and 4.341s.
The first unit also stops at binder:423:18 `a BinaryExpression with a value and a value`
and binder:424:9 `reading name`. The second also refuses the unchecked cast at 1304:42.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestForOfObjectNodeArrayStops|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_stopped' -count=1 -timeout 10m
```

Exit 0, oracle 0.230s. The omit-object-stop mutant exits 1 with the intended
`want the NodeArray representation stop and no IR` assertion.
Logs: `/tmp/for-of-setup.log`, `/tmp/for-of-apply.log`,
`/tmp/for-of-before-{432,1296}.log`, `/tmp/for-of-stopped.log`,
`/tmp/for-of-stop-mutant.log`, `/tmp/for-of-counts-step2.log`.

## Object bindings in forOf

Object binding patterns now lower for represented object elements in arrays and
through the existing custom iteration protocol. Plain fields and aliases use the
existing destructureFrom checks. Bindings execute within the iteration body;
source evaluation stays once, captured bindings stay fresh, and held string
fields survive source removal. Descriptor dispatch is inherited from finishAccessors.
No IR, backend or runtime change was necessary.

Three new .a oracle fixtures are registered from
internal/oracle/for_of_object_destructure_test.go. Source Node agrees with the
JavaScript backend, release native, and sanitized native (ASan/UBSan and leak
checks). The iterable fixture exercises receiver-dependent next and return,
continue, break, return, throw and exhaustion. The lower tests retain explicit
stops for defaults, rest, computed fields, map values and stored collection
iterators, and retain the method-extraction refusal.

The raw CSV has four distinct object-binding gate sites. **4/4 original gate
stops advance; this does not claim four fully lowered functions.** Replays:

| Original site | Next relevant stop |
| --- | --- |
| checker.ts:26333:18 | Refused: a cast the runtime can't check, 26335:61 |
| checker.ts:42067:22 | NotYet: for...of over an object, 42067:40 (NodeArray) |
| transformers/destructuring.ts:301:10 | NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class, 302:26 |
| program.ts:798:18 | NotYet: a destructured name held otherwise than its field, 798:47 |

For each, the replay command uses the complete adapted tsc project, the original
where, -kind NotYet and -reason 'a for...of destructuring an object'. Before
lowering the named two examples reproduce (exit 0). After lowering all four
original signatures are absent (replay exit 1), with the named findings above.
These are latent measurements on a checker-rejected entry-root program, not
whole-program compilation. Other earlier stops in the selected units remain.

Validation commands (all redirected to logs):

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestForOfObjectBindingChecks|TestForOfObjectNodeArrayStops|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure' -count=1 -timeout 10m
python3 internal/lower/testdata/run-for-of-object-mutants.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
```

Focused tests pass: lower 0.160s, oracle 0.690s. Counts refresh passes in 22.652s.
Mutants are isolated and restored by the committed runner; every mutant must
fail its named fixture for the specified reason, rather than a build error:

| Mutant | Catcher |
| --- | --- |
| Reject object bindings | sites fixture: original object-binding NotYet |
| Bind numeric fields to zero | sites fixture: stdout differs |
| Evaluate source each iteration | sites fixture: stdout differs |
| Omit return on early exit | iterable fixture: stdout differs |
| Call return with a fresh receiver | iterable fixture: stdout differs |
| Omit getter dispatch for weight | sites fixture: stdout differs |

The earlier omit-object-stop mutant is also caught by both NodeArray stop tests.
Logs: /tmp/for-of-destructure-focus.log, /tmp/for-of-destructure-mutants.log,
/tmp/adamic-for-of-object-mutants/*.log, /tmp/for-of-destructure-counts.log,
/tmp/for-of-destructure-before-{26333,42067}.log and after replay logs.

The expanded territory permission does not itself resolve the NodeArray design
question. No reason was skipped solely because an IR/backend file was outside
territory. General structural Iterable dispatch remains unimplemented: its
library signatures erase built-in/class/closure receiver conventions and allow
optional return. Preserving those origins needs a separate protocol representation;
this change neither assumes an own closure nor removes existing origin checks.
Coverage for the original 118 object roots remains **0/118**. No runtime C files
were changed. The array metadata and structural protocol limitations are explicit.

## Landing verification

Implementation commit: 88207be7. Merged current origin/main
(ef3141e9b1152ab51b51497f8ce3a2799449c8a3) in 24abd055, without conflicts.
The same focused lower/oracle command passes after merge: lower 0.261s,
oracle 0.712s (/tmp/for-of-landing-focus.log). The two original binder replays
both still exit 0 and reproduce their exact original object stops, recorded in
/tmp/for-of-after-432.log and /tmp/for-of-after-1296.log. No backend comparison
is claimed for those stopped reductions. Counts were refreshed again after
merge (/tmp/for-of-landing-counts.log). No full package tests or full gate ran.

## October 8 correction and fixed tuple group

The pushed views merge 40c696f2 is reverted by 472923b6, without rewriting
history. The revert tree exactly matches c59c040b. The incomplete non-null merge
was aborted first; cab81a59 then merges only c41c0e06. The October 8 ruling keeps
non-null assertions refused in .a. No views-dependent certification is claimed
on this branch. Compiler area still resolves to b410340d at 11:23 UTC.

On this permitted base the fourteen recorded replays were rerun. Both binder
examples still reproduce the exact object stop; all four original object-binding
gates still advance to the stops already recorded above. The three generator
units still reproduce their latent object stops, while normal lowering refuses
generator functions by design. They need ownership and cancellation rules for
suspended frames, not removal of the generator refusal. core.ts:2173:29 is
masked by a function returning T | undefined at 2169:17. The SortedReadonlyArray
site is masked by reading bundle at emitter.ts:2051:13, with an unchecked cast
also refused. Neither masked site is counted as covered.

The one fixed readonly string tuple at moduleSpecifiers.ts:800:25 now lowers
through the tuple's existing numeric object fields. The new helper accepts only
nonempty required tuples with homogeneous string, number or boolean storage and
a plain binding. It holds the source once, selects a field for each index, and
declares a fresh loop binding. Existing Loop IR handles continue, break and
return; builtin array iterators have no return method to call. Custom iterables
still take iteration.go's protocol path before tuple handling. No IR, backend or
runtime changes were required.

The .a fixture covers built strings, source reassignment during iteration,
captured bindings, continue, break, return, and primitive tuple storage. Tuple
element writes remain an existing separate stop and are not certified. Optional,
rest, mixed and object-element tuple storage remains stopped. Eight isolated
mutants were run and restored:

| Mutant | Catcher |
| --- | --- |
| Admit empty storage | lower storage-stop assertion |
| Admit binding patterns | lower storage-stop assertion |
| Admit optional/rest storage | lower storage-stop assertion |
| Admit mixed/object storage | lower storage-stop assertion |
| Shorten length by one | oracle stdout differs |
| Read field zero for every element | oracle stdout differs |
| Evaluate source each iteration | oracle stdout differs |
| Reuse one captured binding | oracle stdout differs |

Commands, all redirected to logs:

```
ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/flow ./internal/ir ./internal/lower ./internal/oracle -run 'TestNonNull|TestExplainChecksDriver|TestAdamicNullishAssertionsAreRefused|TestImpossibleNonNullFixturesAreRefused|TestPossibleNonNullAdamicAssertionsAreRefused|TestCheckedNonNull|TestForOfObjectBindingChecks|TestForOfObjectNodeArrayStops|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure' -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestForOfTupleStorageChecks|TestForOfObjectBindingChecks|TestForOfObjectNodeArrayStops|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_(tuple|object_destructure)' -count=1 -timeout 10m
python3 internal/lower/testdata/run-for-of-tuple-mutants.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
go run ./stage3/census/latent/replay -project /tmp/for-of-adapted/src/tsc/tsc.ts -where /tmp/for-of-adapted/src/compiler/moduleSpecifiers.ts:800:25 -kind NotYet -reason 'for...of over an object'
```

c41 focus passes: CLI 4.354s, lower 2.329s, oracle 3.808s; flow and IR
compile with no tests selected. Tuple final focus passes: lower 0.406s, oracle
0.957s. Node agrees with release native, sanitized native and JavaScript.
Counts update passes in 26.062s, recording 29 allocations and 29 frees.
The tuple replay exits 1 because its original signature no longer reproduces;
its next loop-body stop is an ElementAccessExpression at 801:22. The earlier
any at 799:9 and later unbound deps/result remain. This is one advanced gate,
not a fully compiled compiler function. Raw CSV coverage is **1/118** for the
object reason and **4/4** for the original object-binding gate.

Logs: /tmp/for-of-c41-focus.log, /tmp/for-of-c41-recorded.log,
/tmp/for-of-night-replays/recorded.json, /tmp/for-of-tuple-final.log,
/tmp/for-of-tuple-mutants.log, /tmp/adamic-for-of-tuple-mutants/*.log,
/tmp/for-of-tuple-counts.log and /tmp/for-of-tuple-after-800.log.

## Full root audit and preserved generator refusal

The finished tuple unit is 46659743. All 118 deduplicated original object sites
were replayed at that commit with the same official command and immutable
adapted sources. [Compact per-site evidence](root-replay.csv) records the SHA,
source checker type, result and first remaining stop. **109 exact original
signatures reproduce, one tuple gate is covered, and eight sites are masked.**
All 118 yielded a latent measurement; none was a replay infrastructure error.
Masked sites are binder.ts:1823:29 and 1833:29; checker.ts:13449:40,
13551:46, 29248:47 and 48275:36; core.ts:2173:29; emitter.ts:2062:38.
The source type table alone does not count a masked site as lowered. Total
object coverage stays 1/118. Complete logs: /tmp/for-of-root-audit.log and
/tmp/for-of-root-audit/recorded.json, plus each site's log in that directory.

Three .a reductions correspond to the generator loops at core.ts:333:21,
438:21 and 510:21. They are registered by their dedicated oracle test, outside
the shared runnable-fixture list, whose stopped cases require NotYet rather
than Refused. Source Node prints 2/3 for map, 1/2/2/3 for flat map, and 2 for
filtering. Ordinary lowering returns nil IR with `a generator function` in all
three cases. The reductions remain under a separate testdata directory because
the flow tests discover runnable .a fixtures at the testdata root.

No production generator change is made. Omitting the existing generator refusal
is a temporary, restored mutant; all three fixtures catch it with
`want generator ownership refusal and no IR`. It fails semantically, not through
Go compilation or clang warnings. The ruling needed to build these units is the
owned suspended-frame, resume, cancellation and completion design described in
docs/user-iterators.md. Eagerly materializing the yields would change laziness
and early-exit behavior. Native/JavaScript certification cannot run while the
required refusal holds, and is not claimed.

Commands for this unit, redirected to logs:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestForOfGeneratorRootsRemainRefused -count=1 -timeout 10m
python3 internal/lower/testdata/run-for-of-generator-mutant.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestGeneratorsAreRefusedEvenWithoutYield|TestIteratorViewsCannotHideReturn|TestIteratorViewsCannotEraseReceivers|TestForOfGeneratorRootsRemainRefused|TestForOfTupleStorageChecks|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_tuple' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
```

Logs: /tmp/for-of-generator-focus.log, /tmp/for-of-generator-mutant-run.log,
/tmp/for-of-generator-mutant.log, /tmp/for-of-generator-final.log and
/tmp/for-of-generator-counts.log. Refused fixtures have no allocation rows.
Final focused results: lower 0.365s, oracle 0.809s. Counts refresh passes in
25.008s with no row change for the refused fixtures.
No runtime C files or backends were changed.

Remaining dependencies: NodeArray, JSDocArray, MutableNodeArray and the branded
SortedReadonlyArray need the array storage work to land through area/compiler.
The unlanded codex/views-integration branch is named as that dependency and will
not be merged again. General structural Iterable/ElaborationIterator dispatch
still needs preserved receiver conventions, optional return presence and a
represented protocol result; these remain NotYet, rather than design refusals.
The three generator units remain Refused by design. The standing push rule is
one push per completed locally verified unit, with another only to fix a red.

## Private library iterator view

The remaining Iterable<Symbol> root at checker.ts:15779:30 has one private
consumer whose caller passes propSet.values(). A bounded proof now admits a
plain iteration binding when the parameter belongs to an unexported nongeneric
function in a module, every function reference is a direct call, every argument
at that position is a fresh library Map/Set keys/values/entries iterator, and the
parameter is used only as a for-of source. Export aliases, indirect calls,
parameter aliases/reassignment, spread argument positions, defaults, unknown
callers, unsupported elements and custom producers remain stopped. Physical
element storage must match; a present Weak read is an object but keeps a handle.
General structural iterables still need their own protocol representation.

The helper inspects the owning module, not just the entry import graph. The
first real-site replay still reproduced because that graph omitted the selected
source. A private module function cannot be called from another file without an
export or an escaping reference, both rejected. Global scripts are not admitted;
ordinary .a loading uses forced module detection. No other lower function,
backend, IR or runtime production file was changed.

The existing stored collection iterator lowering supplies the next closure and
its receiver convention. These proven builtin iterators have no return method;
custom iterators continue through iteration.go with their close rules. The .a
oracle fixture holds built strings and objects, live Set deletion/insertion,
captured bindings, early return, exhaustion, and Map/Set origins for one consumer
to Node in JavaScript, release native and sanitized native.

All thirteen final mutants fail their intended tests:

| Mutant | Catcher |
| --- | --- |
| Omit view admission | oracle: original object NotYet |
| Accept unchecked producer | origin-stop assertion |
| Ignore escaped function reference | origin-stop assertion |
| Ignore parameter use/reassignment | origin-stop assertion |
| Admit exported consumer | origin-stop assertion |
| Admit default parameter | origin-stop assertion |
| Admit generic consumer | origin-stop assertion |
| Admit unsupported maybe element | origin-stop assertion |
| Admit consumer without callers | origin-stop assertion |
| Ignore shifted spread position | origin-stop assertion |
| Use present read type as physical storage | origin-stop assertion |
| Share one captured iterator binding | oracle: stdout differs |
| Stop at the first iterator step | oracle: stdout differs |

Three early variants survived and were corrected rather than counted as proof:
the first escape mutant stopped AST traversal before reaching callers; the first
generic probe still failed a separate type proof; the first Weak-union probe did
not isolate present-handle storage. The final runner continues traversal after
ignored references, and the probes include an unused generic parameter and a
present Weak intersection. The complete final thirteen-mutant run passes,
excluding compilation failures, clang warnings and panics as catches. Named
export alias coverage was added and its escape mutant rerun successfully.

The official replay now omits checker.ts:15779:30's original signature and
reaches core.ts:220:112, `a generic function as a value`. The selected compiler
function is not yet fully lowered. [Replay and source hashes](library-view-replay.json)
record the measured production sources and the raw CSV count: **2/118** original
object gates and **4/4** original object-binding gates. Both binder examples still
reproduce their object stop. The full root-replay.csv remains the historical
46659743 audit; this JSON records the additional covered site. The next generic
callable kind is outside forOf. The published function-values worker report
certifies callable union slots, not this generic kind; no worker branch is merged.

Commands, all redirected to logs:

```
python3 internal/lower/testdata/run-for-of-library-view-mutants.py
python3 internal/lower/testdata/run-for-of-library-view-mutants.py ignore-function-escape
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestForOfLibraryViewOriginChecks|TestForOfTupleStorageChecks|TestForOfObjectBindingChecks|TestForOfObjectNodeArrayStops|TestForOfGeneratorRootsRemainRefused|TestIteratorViewsCannotHideReturn|TestIteratorViewsCannotEraseReceivers|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_(library_view|tuple|object_destructure)' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
go run ./stage3/census/latent/replay -project /tmp/for-of-adapted/src/tsc/tsc.ts -where /tmp/for-of-adapted/src/compiler/checker.ts:15779:30 -kind NotYet -reason 'for...of over an object'
```

Final focus passes: lower 0.737s, oracle 0.949s. Counts refresh passes in 24.887s;
the new fixture records 56 allocations and 56 frees. Logs are
/tmp/for-of-library-view-final.log, /tmp/for-of-library-view-counts.log,
/tmp/for-of-library-view-mutants-final.log,
/tmp/adamic-for-of-library-view-mutants/*.log,
/tmp/for-of-library-view-export-focus.log,
/tmp/for-of-library-view-export-mutant.log and
/tmp/for-of-library-view-after-{15779,432,1296}.log.

The live area/compiler was resolved by an explicit tracking refspec, because the
checkout's default fetch refspec updates only main. It advanced through
b68b2fe1 to 84e7f8f6, with no changes in the relevant compiler packages and no
views dependency landing. Earlier b410340d observations described the cached
tracking ref, not a fresh remote resolution. No additional area merge or partial
push was made. Views-dependent certifications remain deferred.
