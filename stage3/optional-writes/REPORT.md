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
