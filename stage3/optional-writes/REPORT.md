Built Effects cleanup scopes and contracts for all 67 exact-only optional rows; 67 scheduled, zero emitted in production.
Commits: cleanup a593d04e; required views 6db50e57; propagated results 4b1922bd; optional storage a022b1e1; final four-row group recorded in branch history.
Focused final-family oracle passes in 9.028s; load/lower/commands in 3.796s, 0.704s, 0.010s; production retains 194 errors.
28 run mutants: guards catch bad writes, undefined results and bad storage; Node catches representation mismatches; leaks and counts catch outer cleanup.
Not covered: whole-program emitted checks, the 90 declaration contracts, the other worker's 104 flag contracts, and the full repository gate.

Historical reports below retain the earlier batch observations; the latest production evidence and remaining.json supersede their counts.

Built direct-write and fresh-construction presence guards and a recorded implements representation relation.
Commits: integration d335b9b1; direct-write batch 58e743bf; construction batch 9af86993; CLI verification commit recorded in git history on codex/stricter-optional-writes.
Focused loader/lower/command tests pass; runtime Node/native/JavaScript proofs pass; production schedules 50, emits 0.
Numeric, reference and class absent-write mutants stop at the guard, exit 70; fresh-construction absence mutant stops at the guard; required-field absence mutant disagrees with Node.
Not covered: 17 other optional-site contracts, whole-program lowering, full repository gate, indexed/catch/JSON work.

The loader admits exactOptionalPropertyTypes-only TS2412 diagnostics for direct
property assignments and exact-only TS2420 implements relations. Ordinary project
errors and all other option errors remain errors. Sites with joint indexed/optional
attribution remain outside this unit. Pending contracts are never reported as
emitted checks. The production probe distinguishes scheduled contracts from checks
in a successfully lowered program.

A direct store holds its receiver once, evaluates the value, stores it using the
existing presence representation, then checks own presence. A missing descriptor
still stops loudly in the existing write machinery. Numeric, reference and class
fixtures observe in, keys, hasOwn and undefined reads against independent Node.
Separate absent-store mutants for each representation compile with sanitizers and
stop at the new guard with exit 70. The readiness pass preserves this guard.

A class implements relation is recorded after the instance storage has lowered.
There is no additional value check: its required undefined field is present,
exactly as ruled. A required-field absence mutant compiles and exits 0, producing
false for in and hasOwn; independent Node catches that disagreement. Interface
alias Object.keys currently refuses during lowering, so the fixture observes keys
and hasOwn through the same class instance, and in/read through its optional view.
A separate exact interface assignment remains an unconverted TS2375 relation.

`c`, `js` and build input paths accept `--explain-checks`; emitted checks are
printed to stderr after successful lowering. Types-only checking does not claim
that a guard was emitted.

Production input is generated with stage3/apply.sh from ledger revision
3f0926c0a55a7b5f64f037b1745e0e984e08c8be, followed by pinned upstream npm ci.
Using current adaptations instead changes 14 input files and yields 168 sites;
that tree cannot substantiate claims about the original 67.
The exact historical input yields all 171 sites, including all 67 exact-only
optional sites. The direct batch schedules 24 TS2412 writes and one TS2420 relation. Fresh data literals add 25 scheduled sites: 13 TS2375, 7 TS2379, 3 TS2345, 2 TS2322. Spreads, accessor literals and other structural/callback contracts remain errors.
It retains 90 ordinary production diagnostics and 121 unconverted option errors.
Consequently **zero of the 67 have compiled into whole-program emitted checks**.
The successful runtime fixtures emit three direct checks, one construction check and record one relation.
This distinction is deliberate and is a block on the requested completion claim.

Commands (full output retained in evidence or named temporary logs):

```
go test ./internal/load ./internal/lower ./cmd/adamic -run 'TestProduction|TestProject|TestOptionalWidening|TestWASIRequestSelection|TestRequestNativeStaysCommand' -count=1
go test ./internal/oracle -run '^TestOptional(LiteralGuard|WriteGuard|ImplementsRepresentation)$' -v -count=1
go test ./internal/oracle -run 'TestOptionalWriteGuard|TestOptionalField' -count=1
go run ./stage3/optional-writes/probe.go /tmp/optional-writes-ledger-tree/src/compiler/*.ts
go run ./stage3/stricter-options/probe.go --production /tmp/optional-writes-ledger-tree/src/compiler/*.ts
python stage3/stricter-options/validate.py /tmp/optional-ledger-loader.json /tmp/optional-writes-ledger-tree --allow-project-errors
```

Planning estimate previously given: October 10 UTC for all 67, first half one
working day. This remains provisional; the whole-program checker blocks mean it
is not a verified delivery date. Do not count the scheduled batch as completed
whole-program checks.

Fresh construction checks hold the newly allocated object, then require every
explicit data property to be own-present before returning it to its target
contract. A sanitizer-enabled mutant marks the allocation's slots absent; the
construction guard catches it with exit 70. The loader test separately proves
that acceptance records a pending contract without claiming an emitted check.

Final verification: `go test ./internal/load ./internal/lower ./internal/native
./internal/javascript ./cmd/adamic -count=1 -timeout 10m` passes: loader 20.482s,
lowerer 37.417s, native 125.077s, commands 4.454s; JavaScript has no package tests.
After the trailing build flag fix, commands pass again in 1.232s. The focused
oracle passes in 4.412s. Native build with trailing --explain-checks prints the
TS2375 construction check and the sanitizer binary prints
`true|slot|true|undefined`. C emission prints the same explanation to stderr.
WASI explanations are explicitly unsupported; native and JavaScript lowering
explanations are implemented.

All run mutants, including merged presence regression controls:

| Mutant | Catcher |
| --- | --- |
| Numeric write stores absent | New presence guard, exit 70 |
| Reference write stores absent | New presence guard, exit 70 |
| Class write stores absent | New presence guard, exit 70 |
| Fresh construction stores absent | New presence guard, exit 70 |
| Required implements field stores absent | Node stdout disagreement |
| Drop absent slot reservation | Missing-field stop, exit 70; Node disagreement |
| Drop spread reservation | Missing-field stop, exit 70; Node disagreement |
| Treat initially absent fields as present | Node stdout disagreement |
| Use layout order instead of write order | Node stdout disagreement |
| Delete leaves the property present | Node stdout disagreement |
| Enumerate interface alias with static class keys | Node stdout disagreement, sanitizer off/on |
| Checked copy drops presence | Independent state assertions, sanitizer off/on |
| Checked copy drops readiness | Checked read stops, sanitizer off/on |
| Presence and readiness storage overlap | Independent state assertions, sanitizer off/on |

This remains an incomplete unit. The remaining 17 sites need checks for spreads,
nested callback/collection contracts, and diagnostics propagated from rejected
exact-optional relations. Merely checking outer object presence would miss nested
writes; those sites have not been waived. The 90 ordinary production errors are
an additional block on whole-program emission, outside this unit's requested
optional-write scope. There is no verified completion date for all 67.

Current main synchronization: merged origin/main at 8b388310 after the CLI commit
7bea27a8. The single gaps_test.go conflict was resolved hunk by hunk: retain main's
exact-refusal machinery and added probes, and the presence side's newer nested
function diagnosis. Re-running the unchanged probes exposed three stale
expectations: mixed-array join's current NotYet, string logical-or now succeeding,
and shift's current exact NotYet. The test preserves exact Refused checks and adds
exact NotYet checks; GAPS.md records the new observations while retaining history.
The affected suite passes in 18.022s, optional focused packages pass, and the
focused oracle passes in 4.104s. Main changed no internal or command files between
f4efdd23 and 8b388310. The full repository gate remains unrun.

The explicit remaining 17-site inventory is in remaining.json. No incomplete
contract is reported as trusted or emitted. The private branch is synchronized;
whole-program delivery remains blocked, with 50 scheduled and zero emitted.

Synchronization merge 00e6f367 is pushed. Its merge-wide diff --check reports
whitespace in main-imported CRLF proving input and retained validation logs.
Those upstream evidence bytes were preserved. The optional unit and the manually
reconciled markdownblocks files pass the scoped whitespace check. Full raw
merge-wide whitespace output is retained in evidence/main-whitespace.log.gz.

Spread batch (October 8 UTC): 53 of 67 contracts scheduled, zero emitted in
the whole-program probe. Added leading-spread snapshot checks and fixed-name
method construction checks. The source key snapshot precedes later initializers:
a fixture deletes a copied property after the spread, and Node, native and
JavaScript still agree that the copy is present with an undefined value. A mutant
that drops copied presence stops at the guard with exit 70. Unsupported spread
lowering stops with a named NotYet rather than omitting the copy check. There are
now 15 run mutants; the additional mutant is the lost copied-presence mutant.

The remaining inventory now contains 14 sites. The 90 previously called ordinary
production errors are declaration-contract diagnostics in the same adapted
compiler sources, outside all 171 stricter-option rows. Of these, 88 match ledger
rows classified other: 82 collection iterator inference, two nested JSX callback
unknowns, two JSON.stringify return contracts, one repeated tuple destructuring
diagnostic and one ES2025 Set declaration mismatch. Two additional diagnostics
are core.ts:1626:37 and core.ts:1731:21, where Generator<T, undefined, unknown>
fails the production prelude SetIterator contract because its yielded done may
be undefined while the prelude requires false. Full messages, ledger identities
and causes are in evidence/production-declaration-diagnostics.json. The project's
own tsconfig audit reports zero project errors. These 90 are therefore not
ordinary errors under the project's own tsconfig. They remain production loader
blocks; this unit does not change the other worker's JSON or indexed contracts.

Focused loader/lower/command tests pass (5.682s, 0.662s, 0.008s); the prior spread
oracle pass was 5.036s. The updated guard refusal is also rechecked before commit.
The provisional completion estimate remains October 10 UTC, with no verified
whole-program completion date while the declaration blocks remain.

Next conditional-contract attempt is blocked: admitting the literal on the right
of `resolved && { ... }` schedules 54 sites, but the native runtime fixture fails
to compile. Cleanup references `adamic_local_2_optionalConstruction` outside the
branch scope where the guard declared it. This is a clang error, not a successful
mutant catch. The attempted acceptance and reflection extensions were withdrawn;
the pushed implementation retains 53 scheduled contracts and 14 explicit errors.
The reproducible .a witness and raw clang log are saved in evidence. Resolving
this requires correct ownership and cleanup scope for Effects locals in lazy
branches; widening the accepted diagnostic without that fix is unsound.

Spread native package validation passed in 125.274s; JavaScript has no package
tests. Updated focused runtime proof passed in 5.341s. Batch commit 59ca7f85 is
pushed to codex/stricter-optional-writes. Provisional October 10 UTC remains an
estimate, contingent on the ownership-scope and production declaration blocks.

Each production declaration blocker now has its own file:line:column and required contract work in [PRODUCTION-ROUTING.md](PRODUCTION-ROUTING.md). Raw messages and ledger identities remain in the JSON evidence.

Cleanup-scope batch: Effects declares guard locals through declareLocal in
emit_locals.go, which holds each local in e.scopes. Previously effects in taste.go
did not push a scope, so releaseScopes emitted those releases in the enclosing
function after the lazy branch had ended. Every Effects expression now pushes
one private cleanup scope, holds its result before cleanup, releases its locals
at the expression site, and transfers the result count to the caller. This is
one rule for all guard expressions and other Effects uses. No protected emitter
assembly file was edited.

The saved conditional reproducer is now a passing fixture, tested with present
and undefined branches and repeated constructions. Native ASan/UBSan, explicit
LeakSanitizer, counted builds and JavaScript match Node. Counts are allocations
12, frees 12, retains 9, releases 21, peak 6, regions 0. A compiling mutant moves
guard-local cleanup to outer process scope, overwriting the held earlier value:
its output matches Node, but LeakSanitizer and allocation/free counts catch the
leak. Its failure is not a clang diagnostic. The original construction absence
mutant still stops at the presence guard. There are now 17 run mutants.

The pinned production probe schedules 54 of 67, including moduleNameResolver.ts
286, with zero emitted whole-program checks and 207 errors (90 production
declaration contracts, 104 other flag contracts, 13 remaining optional contracts).
Focused oracle and all 17 mutants pass in 14.446s. Setup finished in 29.964s with
nproc 5; Go ready 0.024s, Node 0.028s, submodules 0.064s, markdown 0.075s, clang
0.174s, Go build 29.819s. Full affected packages and count regeneration are
checked separately. Provisional all-67 date remains October 10 UTC.

Required-field view batch: checker.ts:43740 now schedules a presence check at
the interface argument. The optional parameter includes undefined, but the
actual source object has a required value field. The source is evaluated once,
its own presence is checked, and its value is preserved, including undefined.
The optional-parameter fixture matches Node for in, read, keys and hasOwn in
both backends with ASan/UBSan and explicit leak checks. An absent-allocation
mutant compiles and stops at the field-presence guard with exit 70. The ledger
now schedules 55 of 67, emits zero whole-program checks, and retains 206 errors:
90 production declaration diagnostics, 104 other flag contracts, 12 optional
contracts. There are now 18 run mutants. Fixture pass: 0.719s. Focused load/lower/
command tests: 3.429s, 0.942s, 0.009s.

The complete affected package gate for cleanup passed: native 146.446s, lower
56.967s, load 32.533s, commands 2.383s, JavaScript has no package tests. Count
regeneration passed in 46.454s and updated nine existing rows, preserving their
allocation/free balance. The shared Effects rule changes measured retain/release
counts and releases the optional-join holder earlier. Date remains October 10 UTC.

Propagated-result batch: six simple undefined-to-required diagnostics now insert
a required-object result check at the use: moduleNameResolver.ts:642, 644, 647,
648 and resolutionCache.ts:1282, 1297. Exact-optional failures prevent the strict
checker from narrowing earlier assignments; the owning project accepts those
assignments and narrows the result. Only the two-line undefined-to-required
diagnostic shape is admitted; nested callback/container errors remain errors.
The producer fixture preserves present-and-undefined fields, matches Node in both
backends, passes sanitizer/leak checks, and emits both its earlier presence check
and its later required-result check. A compiling mutant removes only the producer
result after the earlier guard; the later check catches it with exit 70. Its
absent-construction mutant still fails the earlier presence check. There are now
20 run mutants. Focused proof passes in 0.831s; loader/lower/commands in 3.389s,
0.779s, 0.008s. Production schedules 61/67, emits zero whole-program checks, and
retains 200 errors: 90 declarations, 104 other flags, six optional contracts.
The provisional all-67 date is October 10 UTC.

Optional-source view batch: checker.ts:21434 and 46159 now check the object
representation at the interface use. The native object ABI has insertion ranks
separate from values; the runtime check requires its heap-kind discriminator.
Other IR representations stop with a named NotYet. Legitimate absent fields and
undefined receivers remain allowed. This does not attempt to infer past writes
from a field that is currently absent. The fixture distinguishes an empty source,
a source written with undefined and an undefined receiver; Node, native and
JavaScript agree under sanitizers and explicit leak checks. A storage-kind mutant
compiles and stops at the representation check with exit 70. A lost-write-presence
mutant compiles and exits zero; independent Node observations catch its absent
property. These are separate proofs of the check and representation behavior.
There are now 22 run mutants. Focused oracle passes in 11.337s; load/lower/commands
in 3.682s, 0.719s, 0.007s. Production schedules 63/67, emits zero whole-program
checks, and retains 198 errors: 90 declarations, 104 other flags, four optional
contracts. Date remains October 10 UTC.

Final four-row contract group (October 8 UTC): moduleNameResolver.ts:2772 and
classFields.ts:2745 use recursive object contracts. The guard descends through
required fields, allows nullable inner values, and checks required source fields
that become optional through the target view. Fixtures cover readonly data,
readonly nullable value, and nullish assignment, with Node/native/JavaScript,
ASan/UBSan and explicit leaks. Each leaf-only absent mutant leaves its enclosing
container intact and stops at the inner own-presence guard with exit 70.

commandLineParser.ts:2677 uses an object-array contract: it checks the nullable
array storage, its reference-element ABI, each object's presence storage, and
required source fields through the optional element view. A mapped-array fixture
covers present and undefined arrays, with all four observations against Node.
Its missing-inner-field mutant stops at the presence check.

checker.ts:1683 uses an object-return callback representation contract. Lowering
first requires an object-return ABI, otherwise it stops with a named NotYet; the
runtime check requires closure storage. It does not call the callback early or
forbid legitimate absent optional properties in its return. Both compiled
backends preserve the same object-return presence representation. A wrong-closure
storage mutant stops at the ABI guard with exit 70. A lost returned-property
presence mutant exits zero and is caught by independent Node observations.
This is separate proof of the representation and of the runtime guard.

There are now 28 run mutants. The final family passes in 9.028s, focused loader/
lower/commands in 3.796s, 0.704s, 0.010s. All 171 option identities remain in the
pinned production probe: 67 exact-only optional contracts scheduled, zero emitted
whole-program checks, 194 errors (90 declaration contracts and 104 other flags).
remaining.json is empty because no optional contract is unhandled, not because
the whole program has compiled. Runtime fixtures compile checks of every kind.
All-67 contract implementation date: October 8 UTC. The provisional whole-program
date remains October 10 UTC, contingent on the routed declaration contracts and
the other worker's flag checks; it is not a verified delivery date.

Complete production declaration routing list follows. Each need is an inference
from its recorded diagnostic chain or ledger cause; no routed fix is claimed.

| Location | Code / ledger row | What it needs |
| --- | --- | --- |
| src/compiler/builder.ts:1246:69 | TS2345 / D001 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1258:65 | TS2488 / D002 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1292:61 | TS2488 / D003 | Checker diagnostic owner: reconcile repeated tuple-destructuring failure reporting with the stock project baseline. |
| src/compiler/builder.ts:1345:64 | TS2345 / D004 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1347:41 | TS2345 / D005 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1347:97 | TS2345 / D006 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1493:60 | TS2345 / D007 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1495:26 | TS2345 / D008 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1496:45 | TS2345 / D009 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:1562:64 | TS2345 / D010 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:2377:13 | TS2769 / D013 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:2379:45 | TS2345 / D014 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:395:118 | TS2345 / D015 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:395:85 | TS2345 / D016 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/builder.ts:991:17 | TS2345 / D017 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10022:63 | TS2345 / D018 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10022:86 | TS18048 / D019 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10026:97 | TS2345 / D020 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10033:67 | TS18048 / D021 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10033:79 | TS18048 / D022 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10053:53 | TS2345 / D023 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10061:90 | TS18048 / D024 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10062:69 | TS18048 / D025 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10063:69 | TS2345 / D026 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10064:47 | TS18048 / D027 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10064:93 | TS2345 / D028 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10065:122 | TS18048 / D029 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10066:103 | TS2345 / D030 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10070:50 | TS2345 / D031 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10319:83 | TS2345 / D032 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10320:51 | TS2345 / D033 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10324:76 | TS2769 / D034 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:10336:86 | TS2345 / D035 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:13986:47 | TS2345 / D038 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:13990:49 | TS2345 / D039 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:14643:66 | TS2345 / D040 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15145:9 | TS2322 / D041 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15695:41 | TS18048 / D042 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15697:22 | TS18048 / D043 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15697:47 | TS18048 / D044 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15700:51 | TS18048 / D045 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15701:42 | TS2345 / D046 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15704:43 | TS2345 / D047 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:15706:52 | TS2345 / D048 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:16689:70 | TS2345 / D049 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:16693:62 | TS2322 / D050 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:1779:9 | TS2322 / D052 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:18662:15 | TS2322 / D055 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:25972:13 | TS2322 / D062 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:25974:9 | TS2322 / D063 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:2900:36 | TS2488 / D064 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:35290:33 | TS2769 / D065 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:35291:79 | TS2345 / D066 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:46734:64 | TS2345 / D075 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:46736:167 | TS2345 / D076 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:51465:110 | TS2345 / D078 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:51530:21 | TS18048 / D079 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:51532:25 | TS18048 / D080 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:51533:41 | TS18048 / D081 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:5564:33 | TS2769 / D098 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:8566:17 | TS2322 / D102 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:8861:21 | TS2322 / D103 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:9363:17 | TS2322 / D104 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/checker.ts:9591:37 | TS2345 / D105 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/commandLineParser.ts:1874:96 | TS2345 / D106 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/commandLineParser.ts:2689:66 | TS2345 / D111 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/commandLineParser.ts:2952:13 | TS2322 / D112 | JSON worker: reconcile the production stringify return contract with the project, keeping the undefined-result runtime obligation. |
| src/compiler/commandLineParser.ts:3449:55 | TS2322 / D113 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/commandLineParser.ts:4014:5 | TS2322 / D114 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/core.ts:1626:37 | TS2322 / not in original ledger | Prelude/library owner: reconcile SetIterator.next result done with stock IteratorYieldResult (false or undefined), without treating undefined as done. |
| src/compiler/core.ts:1637:11 | TS2740 / D115 | Prelude/library owner: use the project ES2024 Set surface or adapt the implementation for all required ES2025 methods. |
| src/compiler/core.ts:1731:21 | TS2322 / not in original ledger | Prelude/library owner: reconcile SetIterator.next result done with stock IteratorYieldResult (false or undefined), without treating undefined as done. |
| src/compiler/program.ts:4988:28 | TS2345 / D136 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/program.ts:5038:14 | TS2488 / D137 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/resolutionCache.ts:1616:118 | TS2345 / D145 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/resolutionCache.ts:1618:39 | TS18048 / D146 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/resolutionCache.ts:1619:46 | TS2345 / D147 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/resolutionCache.ts:1619:74 | TS2345 / D148 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/resolutionCache.ts:1619:99 | TS18048 / D149 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/sys.ts:1708:56 | TS2345 / D154 | JSON worker: reconcile the production stringify return contract with the project, keeping the undefined-result runtime obligation. |
| src/compiler/transformers/es2017.ts:325:21 | TS2322 / D176 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/transformers/es2017.ts:327:17 | TS18048 / D177 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/transformers/esnext.ts:205:52 | TS2345 / D190 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/transformers/jsx.ts:174:24 | TS2488 / D222 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/transformers/jsx.ts:187:172 | TS18046 / D223 | Collection iterator declaration fix, then recheck the downstream JSX callback inference; no catch-variable change. |
| src/compiler/transformers/jsx.ts:187:188 | TS18046 / D224 | Collection iterator declaration fix, then recheck the downstream JSX callback inference; no catch-variable change. |
| src/compiler/tsbuildPublic.ts:1064:76 | TS18048 / D232 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/tsbuildPublic.ts:1065:62 | TS18048 / D233 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/tsbuildPublic.ts:1066:42 | TS18048 / D234 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
| src/compiler/tsbuildPublic.ts:1776:54 | TS2345 / D235 | Reconcile the production collection iterator next overload with the owning project library so generic Iterable inference does not inject undefined; preserve genuine end-of-iteration checks. |
