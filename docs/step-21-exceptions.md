# Step 21: native exceptions

## Census: observed evidence

Delivery base: `8cb5e7c189fc431ab3cab7e0be2c0530e722a598`, the specified
`origin/compiler/area-next-fixtures` candidate. The delivery branch is
`codex/scout-exceptions`. No other worker branch is merged.

The full compiler-scope census committed on this base is
`stage3/meter/runs/20261008T035244Z.latent-full/compiler/full.jsonl.gz`.
It is measured on a checker-rejected, adapted program. A root here means a
unique attempted unit (`unit`), not a repeated finding or an entire imported file.
The hidden ranking at `6c4fc1af` instead measures compiler `ed6e2975`:
its bytes are outermost-cause credit, an estimate of bytes exposed by removing
one blocker, not successful compilation or dynamic heat. Those snapshots cannot
be added or presented as one current-base measurement.

**Observation:** neither snapshot records an exception-specific Refused or NotYet
reason. The exact-reason census has zero roots for the production exception
reasons below. The historical ranking has no matching reason entry, so its hidden
credit is zero, not a measurement that these constructs have no hidden source.
There cannot be three diagnostic witnesses for a reason with no observations.

| Production diagnostic shape | Base roots | Historical hidden credit | Witnesses |
| --- | ---: | ---: | --- |
| Refused: `throwing a <type>` | 0 | 0 | none recorded |
| NotYet: `throwing an Error that isn't made where it's thrown or caught by the catch around it` | 0 | 0 | none recorded |
| NotYet: `a catch that destructures what it caught` | 0 | 0 | none recorded |
| NotYet: `assigning to what a catch caught` | 0 | 0 | none recorded |
| NotYet: `new Error with options` | 0 | 0 | none recorded |
| NotYet: `new Error with a message that isn't a string` | 0 | 0 | none recorded |
| NotYet: `a try around <operation>, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md)` | 0 | 0 | none recorded |

The last template covers repeat, normalization, number formatting, array and
 typed-array ranges, RegExp flag checks, frozen-object writes and Hash
finalization. It is a representation/behavior barrier, not a request to silently
turn a catchable failure into a terminal panic.

The refusal table supplied inside the ranking has no exception-specific row.
`reading exception` is **not** catch narrowing: it names the generator transform's
`ExceptionBlock` local. Its 12 historical boundaries receive zero hidden credit;
three witnesses are `transformers/generators.ts:2250`, `:2252`, `:2257`.
`reading error` names diagnostic-message locals in `checker.ts:36462`, `:36467`,
`:36489`, also zero credit. CatchClause/ThrowStatement type names denote compiler
AST data; their checked-view and writable-variance barriers belong to those steps.
Detached `throwIfCancellationRequested` belongs to method binding. Counting these
as retired exception roots would be incorrect.

The original stock census does expose the catch boundary through **checker**
TS18046: eight diagnostics, seven distinct file:lines, five file roots. Witnesses:
`commandLineParser.ts:2301`, `program.ts:406`, `sys.ts:1553`.
The other lines are `program.ts:441`, `sys.ts:1280`, `tracing.ts:66`, and
`transformers/jsx.ts:187` (two accesses). Hidden bytes are unavailable for this
exact checker code: the ranking attributes checker bodies as one cause, not by
TS diagnostic. Do not substitute that bucket's 1,730,511 bytes for TS18046.
Disabling unknown catch variables would change the proof contract and is not built.

## Source exposure, distinct from refusal roots

A pinned TypeScript 6.0.3 AST walk over `src/compiler/**/*.ts` finds 34 executable
try statements: 26 try/catch and eight try/finally. There are nine `throw new Error`
and three identifier throws. There are no executable try/catch/finally combinations,
other throw expressions or Error subclasses in this corpus. Helper template strings
and comments are not executable compiler-source AST and are excluded.

| Shape | Source sites | Three original file:line witnesses |
| --- | ---: | --- |
| try/catch | 26 | `commandLineParser.ts:2297`, `moduleSpecifiers.ts:134`, `sys.ts:1549` |
| try/finally | 8 | `checker.ts:1937`, `symbolWalker.ts:50`, `utilities.ts:787` |
| throw new Error | 9 | `core.ts:1583`, `moduleNameResolver.ts:1681`, `utilitiesPublic.ts:448` |
| throw identifier | 3 | `debug.ts:203`, `program.ts:2854`, `sys.ts:1555` |

Two identifier throws rethrow catch bindings; debug.ts constructs an Error in a
local before throwing it, currently NotYet when reduced. None of these counts is
a count of independently compilable roots. Earlier checker, host, signature and
body barriers can prevent exception lowering from being attempted.

## Reproduction and validation

`python3 docs/step-21-exceptions/census.py` reads the two pinned artifacts and
checks the distinction between actual exceptions and compiler AST names.
`census.json` retains its output. Its root count deduplicates attempted units.
A selector mutant that admits `reading exception` must fail the exclusion
assertion. An accounting mutant that drops one unknown-catch diagnostic must fail
the independent eight-diagnostic assertion. No native fixture is added in this unit.

Setup succeeded with `GOPROXY='https://proxy.golang.org|direct'` and selected
`/workspace/adamic-tools/env.sh`. Cumulative timing lines: Node 0.229s,
Go 0.235s, clang 0.950s, markdown ready 1.485s, submodules 23.056s,
Go build 211.122s, test binaries deferred 211.235s, cache warm 211.237s,
done 211.267s. `nproc` printed 5; cgroup quota is four CPUs.

## Design: existing contract and proposals for @system_adamic

The approved 0.1 refusal table is historical; docs/memory.md records the existing
0.2 Error-only extension. docs/escape-hatches.md requires proof at every dynamic
boundary and keeps inserted checks terminal. This unit does not amend either
language document. Every admission change below is a **proposal**, including
changing Refused to NotYet. No proposal is enabled by a flag or built here.

### Representation and lowering by shape

| Shape | Existing representation/lowering | Proposed extension and obligation |
| --- | --- | --- |
| fresh Error throw | `ir.MakeError` makes a counted object with name/message; `ir.Throw` retains it into `adamic_thrown` | Keep single evaluation and take ownership before dropping expression temporaries |
| arbitrary value throw | Refused; no universal catch payload | An owned tagged dynamic value plus a separate pending boolean; undefined and null are valid thrown payloads, never absence of an exception |
| saved Error | NotYet unless the expression is a catch identifier | Admit only after general Error identity/representation is proven; preserve identity and owned payload across aliases |
| catch unknown | Catch binding is physically Error in the current closed throw world; `instanceof Error` folds to true | Bind the tagged payload as unknown; implement actual typeof, literal and nominal checks before extraction. Unknown never authorizes an interface or callable-signature assertion |
| rethrow | Catch takes the pending word, clears it, then a throw retains the same object | Move/retain the same tagged value, including nullish/primitive values; never manufacture an Error wrapper |
| finally normal/return/break | Cleanup paths save the result or pending Error, clear the pending word while finally runs, and route every outgoing completion through finallies | Model completion as Normal/Return/Throw/Break/Continue with target and owned payload. Finally's abrupt completion replaces and releases the old completion exactly once |
| Error subclasses | Ordinary class inheritance exists, but builtin Error ancestry is not represented by that path | Share nominal Error ancestry and subclass field layout, constructor ordering and name/message semantics; `instanceof` must test real ancestry. Native stack text remains unsupported |
| closures/nested functions | `MayThrow` and `ClosuresMayThrow` propagate to a fixed point; calls test pending state before reading their result | Keep exceptional exits in nested closure environments, bound methods and every callback loop. A narrower target-set optimization requires an independently proven complete target set |
| library exceptions | Selected library calls under try are NotYet where native panics and Node throws | Separately propose catchable RangeError/TypeError and host-error contracts, with names, messages and fields preserved. Do not silently make compiler invariant panics catchable |

The existing native Error object has two counted string fields. A universal
exception payload should reuse the existing dynamic-value metadata and ownership
helpers wherever they can carry every supported type; this is not permission to
pretend all references are objects. Tags must distinguish number, boolean, string,
null, undefined, object and callable references; further admitted kinds need their
actual representations. Payloads stored in regions must escape to a durable owner
before unwinding frees that region. Cycle analysis must follow strong contents of
thrown values and captured cells; no collector is available to repair a cycle.

Native cleanup remains explicit C control flow. Throwing callees return only a
placeholder; callers inspect pending state before extracting or using a result.
No longjmp may skip an owning Adamic frame. Sort's internal non-owning frames have
a separate existing cleanup protocol and need a dedicated oracle when changed.
Generated JavaScript uses JavaScript's own completion semantics. The source Node
run remains independent of that backend; both generated backends must agree with it.
Uncaught arbitrary values require a separately specified String(value) and exit
contract, including user conversion effects. This design does not invent one.

### Refused programs retained for the proposal

These minimal programs state exactly which admission changes need review. They
are examples, not claims of successful lowering. Reductions and observations
follow in the fixture unit.

```typescript
// Arbitrary values, including an undefined payload that must still be pending.
try { throw undefined; } catch (e: unknown) { console.log(typeof e); }
try { throw 7; } catch (e: unknown) {
    if (typeof e === 'number') console.log(`${e + 1}`);
}
// A saved Error, reduced from debug.fail's construction then throw.
const saved = new Error('failure');
throw saved;
// Builtin Error ancestry is different from an ordinary declared base class.
class Cancelled extends Error {}
try { throw new Cancelled('cancelled'); } catch (e: unknown) {
    console.log(`${e instanceof Cancelled}`);
}
// The checker must continue refusing an unproven read on unknown.
try { throw new Error('failure'); } catch (e: unknown) { console.log(e.message); }
// Library failure is catchable on Node; a terminal native panic is not equivalent.
try { console.log('x'.repeat(-1)); } catch { console.log('caught'); }
```

Unchecked unknown reads, any, unrelated double casts, forged predicate contracts,
unproven callable signatures and unsafe mutable widening stay refused. An Error
subclass must not bypass nominal identity, initialization or cycle proofs.
Error options/cause, stack, non-string message conversion, catch destructuring,
universal payloads, builtin subclass constructors and unimplemented library
failures remain at their current diagnostics pending review and implementation.

### Silent-miscompile audit

1. **Absence versus payload:** a pending flag must distinguish throwing undefined,
   null, zero, false, NaN and empty string from normal completion.
2. **Order and identity:** evaluate throw operands once; preserve the same object
   on catch/rethrow; take ownership before releasing temporaries or a source slot.
3. **Narrowing:** replacing the closed Error world invalidates the current true
   `instanceof Error` fold and Error-shaped catch local. Both must change together.
   A tag test must precede payload extraction; checked views stay transitive.
4. **Cleanup:** parameters, partial arguments, arrays under construction, iterator
   holds, callback results, capture cells and regions need exactly one release on
   every exceptional exit. A later argument can throw before a consumed call runs.
5. **Finally:** save a return result before finally mutates its source; nested
   return/throw/break/continue must replace completions and retain their targets.
   Finally can call and catch another throwing callback without losing the outer
   pending completion. Throwing from catch does not re-enter the same catch.
6. **Effects:** every callback ABI and runtime loop needs a pending test before
   processing its result. MayThrow summaries include virtual targets, nested
   functions, recursive calls and catch/finally bodies. Unknown effects are broad.
7. **Liveness (#zm26nev):** exceptional edges preserve the pre-assignment value
   when evaluation throws. Reuse and moves must retain values read by catch or
   finally; globals cannot be nulled before a throw leaves them observable.
   The current `LiveOut` starts from empty sets and grows them monotonically,
   so cardinality is sufficient to detect a change. The earlier design commit
   suspected a convergence bug; review of the transfer function disproved that
   inference. No convergence change is justified by this observation.
8. **Other analyses:** initialization, freshness, aliasing and region escape must
   see throws and finalizers too. SSA assignment definitions must not be mistaken
   for stores that happened on the exceptional edge.
9. **Panics:** invariant checks remain terminal and flush stdout; catch/finally
   must not execute after them. Resource failures and engine stack text cannot be
   claimed byte-identical without a defined contract and external oracle evidence.

The highest source exposure is already-admitted try/catch. The build unit will
hold that existing shape to Node and prove throw-path liveness with a compiler
mutant, without admitting a new throw kind. It cannot truthfully retire the eight TS18046 diagnostics, saved-Error
NotYet or arbitrary-value refusal. Their measured exception-root retirement is
zero until a reviewed admission change and a census replay establish otherwise.

## Fixtures: reductions and observed outcomes

All source references below are pinned original TypeScript, fetched from its own
repository, not copied from cohere. The positive reductions replace host/compiler
objects with small strings, arrays and callbacks while preserving the source's
exception/control-flow structure. Additional finally-return/break and throw-path
consumption cases are explicitly stress controls, not claimed original tsc sites.
The source AST census is reproducible with `syntax.cjs` and TypeScript 6.0.3;
`syntax.json` records every site. Error subclasses and arbitrary origin payloads
are scope extensions because the executable tsc AST contains neither.

| Acceptance fixture | Original source | Observed stdout (lines separated by /) |
| --- | --- | --- |
| `step21_catch_callback.a` | `program.ts:398`, `:2844`; `utilitiesPublic.ts:738` | text0 / read1 / text2 / fallback |
| `step21_finally_callback.a` | `utilities.ts:787`, `checker.ts:1937`, `symbolWalker.ts:50` | inner2 / outer1 / failed3 / outer1 / 0 / visit4 |
| `step21_rethrow.a` | `program.ts:2844`, `sys.ts:1549`; added identity control | ok0 / recovered / cancel2 / true / identity3 |
| `step21_finally_completion.a` | return-through-finally from `utilities.ts:787`; added override/break controls | try0 / finally1 / finally2 / cleanup0 / after |
| `step21_liveness.a` | callback assignment from `commandLineParser.ts:2297`; added consumed-array control | original1,added2 / original1 / failure3 / original1 |

Every positive fixture exits zero with empty stderr. Each agrees with independent
source Node, generated JavaScript, sanitized native and release native; each passes
the separate leak check. Strings are built at runtime, so lifetime failures cannot
hide behind immortal literals. The test registers these fixtures from
`internal/oracle/step21_exceptions_test.go`, without editing the protected shared
oracle file. They participate in the ordinary oracle and counts gate.

| Proposal probe under `docs/step-21-exceptions/proposals` | Original reduction / scope extension | Observed compiler outcome | Source Node stdout |
| --- | --- | --- | --- |
| `any_value.a` | Generalizes `program.ts:2854` rethrow to undefined origin | Refused: throwing a undefined | undefined |
| `saved_error.a` | `debug.ts:197-203`, removing stack capture/debugger | NotYet: throwing an Error that isn't made where it's thrown or caught by the catch around it | saved1 |
| `error_subclass.a` | Generalizes `program.ts:2848` cancellation class to Error ancestry | NotYet: a computed class base; name the base class directly | true |
| `unknown_read.a` | `commandLineParser.ts:2301` e.message | Checker TS18046: error is of type unknown | unknown |
| `library_failure.a` | Recovery from `utilities.ts:7810`; known repeat failure replaces JSON parsing | NotYet: a try around repeat, whose failure is a panic natively but a throw a catch can take on Node | caught |

All five source probes finish on Node with exit zero and empty stderr. Their
negative outcomes are acceptance tests too: no proposal is admitted accidentally.
These files are outside the repository's ordinary Adamic include paths. The
unknown-read probe invokes Node directly because the cached source helper first
loads source through Adamic and intentionally stops on this checker diagnostic.
The first two test runs exposed that harness limitation; the corrected run passes.

For each positive fixture, an IR mutant adds one line inside an executed catch or
finally. Both generated backends and release native disagree with source stdout;
all five mutants finish cleanly, with no sanitizer finding or leak. These prove
the external comparison can fail; they are semantic IR mutants, not five claimed
production compiler defects. `TestStep21FixtureMutants` returns success only when
every mutated artifact is rejected by that comparison.

Focused command, output in `/tmp/scout-fixtures-proof.log`:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep21|TestNativeAgreesWithNode/internal/oracle/testdata/step21_' -count=1 -v -timeout 10m
```

Result: exit zero, oracle 1.595s, five positive programs, five proposal probes,
five killed semantic mutants, native cache hits 0/misses 30, Node hits 0/misses 20.
The complete output is retained in `evidence/fixtures.log.txt`.

The required counts refresh initially failed only because the cloud setup had not
installed `@types/node` for existing Node host fixtures. The workaround is
`npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund`, which installs
the repository's pinned dependencies without changing its manifests. The setup
itself succeeded; this was a separate counts prerequisite. The full counts
refresh was rerun after that installation, rather than hiding unrelated rows.

Counts command: `go test ./internal/oracle -run TestCountsAreRecorded -count=1
-timeout 30m -args -update-counts`, output `/tmp/scout-counts-final.log`, exit zero,
37.015s. New allocation/free totals are 17/17 (catch callback), 19/19 (finally
callback), 13/13 (rethrow), 17/17 (completion), and 12/12 (liveness); none uses a
region. All existing numeric rows are unchanged. Regeneration also moves the
existing logical_and_reference_maybe row to its fixture-list position and removes
a stale taste/17_binder_flow row which the base no longer counts. These are
inherited registry/table reconciliation, not exception lifetime changes.

## Unit four: throw-path implementation held to external behavior

The most common shape is already implemented on the requested base: Error-only
try/catch, including calls through closures. No new lowering is justified by this
census. This unit adds a regression assertion for the existing throw-path
liveness implementation and proves the production cleanup paths with source
mutants. It does not claim a new exception representation or an admission change.

`TestStep21ThrowPathLiveness` checks that the original `text` array remains live
before the consuming grow call when a later callback assignment can throw into a
catch reading that original value. The associated acceptance fixture independently
matches Node, so this assertion pins an observable lifetime requirement rather
than merely restating the dataflow transfer implementation.

`run-mutants.py` temporarily changes one production source anchor, runs a focused
oracle check, requires the intended failure, and restores the original bytes in a
finally clause. Run it serially, with no concurrent compiler tests. No production
source changes remain in this delivery.

| Production mutant | Fixture / assertion that catches it | Observed failure |
| --- | --- | --- |
| Remove exceptional flow edges | step21_liveness | UBSan runtime error |
| Kill the old assignment value on the exceptional edge | TestStep21ThrowPathLiveness | old text is dead |
| Retain statement temporaries instead of releasing them on a throw | existing exceptions fixture | leak check |
| Remove the pending exception's finally-scope owner | step21_finally_completion | LeakSanitizer, 262 bytes / four allocations |
| Store an Error message without retaining it | step21_catch_callback | ASan heap-use-after-free |

All five exit nonzero with their intended catcher; none is credited for a compile
failure. Individual complete outputs are in `evidence/*-leak.log.txt`,
`evidence/missing-exception-edge.log.txt`,
`evidence/assignment-kills-handler-value.log.txt` and
`evidence/error-message-not-retained.log.txt`. The restoring runner's summary is
`evidence/production-mutants.log.txt`.

Two preliminary attempts are explicitly not credited: removing `e.end()` from
throw emission produced invalid C; changing temporary releases survived the small
callback reduction because it did not exercise that ownership path. The final
cleanup mutation uses the existing deeper exception fixture and fails its leak
check. The liveness control passes before mutation. These are test observations,
not newly discovered production defects.

Final focused verification, logged in `evidence/final-checks.log.txt`:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep21|TestNativeAgreesWithNode/internal/oracle/testdata/(step21_|exceptions[.]a$|closures_throw[.]a$|finally_leaves[.]a$|regions_throw[.]a$|reuse_throw[.]a$)' -count=1 -v -timeout 10m
go vet ./internal/oracle
```

This covers both generated backends, release native, sanitizers and leaks for the
five reductions and the existing exception, closure, finally, region and reuse
regressions. The counts refresh for the new fixtures is recorded above. No whole
package test or full gate was run.

**Roots retired: zero; hidden bytes retired: zero.** The actual current census has
no exception-specific refusal/NotYet roots to retire. Arbitrary thrown values,
saved-Error origins, real unknown narrowing, Error subclasses and recoverable
library failures remain proposals for @system_adamic. Their refused programs and
current outcomes are recorded above; implementing them now would change admission
without the requested language ruling.

Final result: exit zero, oracle 12.232s, native cache hits 0/misses 51, Node
hits 0/misses 34. Go vet exits zero with empty output. The oracle selector also
matches existing class_inheritance_exceptions and fallthrough_exceptions, for
twelve ordinary oracle fixtures in total. Five proposal probes, five semantic IR
mutants and the direct throw-path liveness assertion all pass. The five production
mutants are separate expected-failure runs recorded in their individual logs.

## Adopted items 1 and 3: arbitrary payloads and unknown catches

The ruling adopts the design above and supersedes its proposal status. The branch
merges area-next's own `dcdbb909` line. Items 1 and 3 are implemented together;
items 5, 4 and 2 follow in that order. The extension applies to both .a and .ts.

Throws now use the existing owned union representation. Scalars have boxes,
references preserve their identity and heap brand, null has an explicit sentinel,
and undefined is a null payload pointer. `adamic_exception_pending` is independent
of that pointer, so both null and undefined propagate through callbacks and
finally. Typed nullable references convert their null pointer to the explicit
null sentinel before entering the payload. The operand is evaluated once.

A catch now holds Union, the runtime representation of unknown. typeof and literal
tests observe the actual payload. instanceof Error tests a nominal runtime class;
a plain object with name/message fields fails it. Narrowed Error and user-class
reads recheck nominal identity before interpreting fields. Unknown interface and
call-signature assertions remain Refused; unguarded message reads remain TS18046.
Reassigning a catch binding stores a new dynamic value and subsequent tests see it.

The completion paths retain or move the same payload. A finalizer's return, break
or replacement throw releases the suspended payload through its owning scope.
Boxing now preserves reachability in region escape analysis. A fresh object which
can be thrown is allocated under a durable heap owner, rather than in the
consuming statement's region. No owning frame is skipped by longjmp.

Uncaught exceptions flush stdout and exit 1. Neither runtime invokes String or
prints a stack. The Node oracle runtime no longer intercepts all exceptions and
converts them into exit-70 panics. The shared oracle comparison is changed only as
required by the adopted rule: equal exit-1 exceptions compare stdout and exit code;
sanitizer reports still fail, and exit-70 panics keep their stderr comparison.
The shared fixture registry is retained. Inserted checks and compiler invariants
remain terminal panics and cannot be caught.

The new dynamic and region-payload reductions generalize program.ts:2844-2854 and
commandLineParser.ts:2297. Arbitrary origin tags are scope extensions because tsc
itself originates only Error values. Additional uncaught controls cover undefined
and an object whose conversion method would print if invoked. A temporary .ts
copy of the .a dynamic fixture proves equal admission without adding a new .ts
program to the repository.

Rule mutants are reproducible with `run-adopted-mutants.py`, run serially:

| Mutant | Catcher |
| --- | --- |
| Disable both unknown-assertion proof barriers | admission test requires Refused |
| Convert an uncaught object to a string | uncaught test rejects exit 70 |
| Use payload presence as pending state | Node stdout comparison |
| Fold instanceof Error to true | Node exit comparison |
| Classify a boolean as a number | Node stdout comparison |
| Box null as undefined | Node stdout comparison |
| Give finally's pending payload two owners | ASan heap-use-after-free |
| Ignore boxed reachability, allocating the payload in the statement region | ASan heap-use-after-free when catch reads it |
| Exit 70 for an uncaught exception | uncaught exit check |

All nine mutants fail for their intended reason and restore source bytes. One
preliminary unknown-assertion mutation survived because the up-front proof still
refused it; the credited mutant disables both independent barriers. An earlier
uncaught mutant was rejected by the runner's expected-message assertion despite
correctly failing the test; the final runner names the actual exit check.

Measured census retirement for items 1 and 3: **zero roots, zero hidden bytes**.
The pinned census contains no corresponding actual refusal roots. The generalized
undefined proposal now compiles; the TS18046 diagnostics remain correct refusals
of unguarded unknown reads and are not claimed as retired. Saved Error origin,
Error subclasses and library recovery retain their current barriers for the later
numbered deliveries.

Commands and observed results for this delivery:

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep21|TestNativeAgreesWithNode/internal/oracle/testdata/(step21_|exceptions[.]a$|exceptions_uncaught[.]a$|closures_throw[.]a$|closures_throw_uncaught[.]a$|regions_throw[.]a$|reuse_throw[.]a$)' -count=1 -v -timeout 10m`: PASS, 12.336s, native misses 70, Node misses 52. The selector also includes class-inheritance and fallthrough exception regressions.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep21TypeScriptExtension|TestNativeAgreesWithNode/internal/oracle/testdata/closure_convention_host24[.]a$' -count=1 -v -timeout 10m`: PASS, 9.802s. This follows the host local-name collision exposed by the first counts run.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts`: PASS, 39.368s. The initial run failed on that host C name collision and is retained as evidence.
- `python3 docs/step-21-exceptions/run-adopted-mutants.py`: nine intended failures, runner exit zero.
- `python3 docs/step-21-exceptions/prove-uncaught-sanitizer.py`: intended ASan failure, runner exit zero. The shared oracle reports sanitizer failure even though stdout and exit 1 match the source.
- `go test ./internal/lower -run 'TestWhatZeroOneRefusesIsRefusedWithAFix|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat' -count=1`: PASS, 0.802s.
- `go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle`: exit zero, empty output.

Logs are retained in `evidence/items13-*`, `evidence/adopted-*` and the individual
mutant files. New allocation/free rows are 65/65 (dynamic), 3/3 (region payload),
0/0 (uncaught undefined) and 3/3 (uncaught object). Existing catches gain counted
borrows for checked dynamic narrowing; the refreshed rows record those actual
retains/releases. No full package test or full gate was run.

## Adopted item 5: represented library failures

Library argument validation now lowers to ordinary owned Throw IR before native primitives run. RangeError and TypeError have nominal identities below Error; filesystem errors retain their actual nominal ancestry and code, and Hash update/digest after finalization now throws Error with ERR_CRYPTO_HASH_FINALIZED. Names and messages agree with source Node. Existing compiler-inserted narrowing checks still call terminal panic and exit 70; no catch or finally executes after that exit.

The added tsc recovery reductions extend utilities.ts:7810, program.ts:2844 and sys.ts:1549 with range validation, frozen writes, regexp global-flag validation and host callbacks. These extensions test represented library failures; they are not claims that tsc contains each invalid argument. Both backends, release native, sanitizers and leak checks pass. The pinned census retires **0 roots and 0 hidden bytes**: its exception buckets remain empty; the independent source witnesses justify acceptance coverage rather than invented census savings.

Commands and observations:

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep21|TestNativeAgreesWithNode/internal/oracle/testdata/step21_' -count=1 -v -timeout=10m`: PASS, 15.296s.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(step21_|number_formats|precision_range|radix_range|node_crypto|fs_file)' -count=1 -v -timeout=10m`: PASS, 2.478s.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`: PASS, 50.110s; counts refreshed.
- Focused lower admission tests and `go vet ./internal/lower ./internal/ir ./internal/native ./internal/javascript`: PASS. No full package tests or full gate ran.

`run-library-mutants.py` restores every mutation. Wrong RangeError identity, wrong range message, wrong TypeError identity and wrong filesystem TypeError ancestry each fail stdout comparison. Hash failures changed to panic fail exit-code comparison. A soundness check changed to a catchable throw fails the explicit exit-70 assertion (native 0). The runner initially used the wrong expected diagnostic string for that final mutant; the mutant itself failed correctly, and its rerun matches the assertion. Logs are in `docs/step-21-exceptions/evidence/library-*.log.txt` and each named mutant log.

Coverage is the represented argument and state failures named above, including repeat, normalize, numeric formats, typed arrays, Array construction/from, frozen properties, regexp flags and filesystem/Hash errors. Allocation exhaustion, oversized concatenation/builders and unrepresented library APIs are not established by these fixtures. Error subclass construction and saved Error origins remain staged for the following deliveries; stack remains NotYet.

## Adopted item 4: Error subclasses

Declared Error subclasses now share the owned name/message prefix with their builtin ancestor. Construction fixes the actual derived identity before the builtin initializer writes that prefix, then runs derived field initializers after super and before the remaining constructor statements. Error, RangeError and TypeError ancestry use actual nominal tables in C and actual builtin prototypes in the JavaScript backend; unrelated Error subclasses stay distinct. A default subclass name remains Error, rather than becoming the subclass's spelling. Rethrow retains the same object.

The program.ts:2848 cancellation reduction extends the original class identity shape with builtin Error ancestry, a second derived level, RangeError, default construction, ordered field effects and a live message alias. Source Node, both backends, release native, sanitizers and leaks agree. Stack access is explicitly NotYet. Existing prototype/descriptor boundaries remain: Error own-property reflection cannot use the prefix as evidence for Node's inherited name or non-enumerable message; Error constructor options/cause remain NotYet.

Four production mutants establish the checks: missing builtin ancestry and a wrong default name fail stdout comparison; moving derived fields before the base initializer fails stdout comparison; releasing the shared prefix twice fails ASan with a heap use-after-free when a live message alias is read after catch cleanup. The first release mutant survived an immortal message, and also survived a sole heap owner because iterative destruction skips a child whose count is already zero. The final fixture keeps a second live owner, so the duplicate release consumes an observable alias. Its restored control passes.

Commands: the focused `TestStep21` and `TestNativeAgreesWithNode/.../(step21_|inheritance_)` oracle selection, `go test ./internal/lower -run 'TestInheritance|TestPrototype' -count=1`, four serial mutants in `run-subclass-mutants.py`, narrowed compiler-package vet and the required counts update. Logs record their results under `docs/step-21-exceptions/evidence/subclass-*.log.txt`. No full package tests or full gate ran. This delivery retires **0 pinned census roots and 0 hidden bytes**; the cancellation witness is source coverage, not an exception refusal bucket in the census. Saved Error origins are the remaining ordered item.

## Adopted item 2: saved Error identity

The saved-origin NotYet is removed. Debug.fail's construct/keep/throw shape uses the same owned Error object as immediate throw. Alias writes stay visible, callbacks keep the Error alive, Error-typed views preserve a subclass's dynamic identity, and catch/rethrow through finally retains that object rather than wrapping it. The debug.ts:197-203 reduction exercises those observations with runtime-built messages.

The JavaScript backend now exits 1 for an uncaught exception without invoking a renderer, implicit conversion or stack access. Source Node remains the external oracle on stdout and exit code. Both emitted backends stay quiet on stderr for ordinary uncaught payloads; sanitizer failures remain independently fatal. Exact builtin nominal ancestry is also rechecked after a call invalidates an unknown's TypeError narrowing. The inserted check exits 70 before catch/finally, as the adopted terminal-check rule requires.

Production mutants:

- `saved-error-wrapped`: reconstruct every thrown Error. Node stdout rejects false alias identity.
- `rethrow-wrapped`: reconstruct only the catch binding's payload. Node stdout rejects false rethrow identity, while the first throw still preserves aliases.
- `builtin-narrow-forgets-subtype`: accept any Error where TypeError was proven. The explicit terminal-check assertion fails: native exits 0 instead of 70. Its first draft did not compile because identity was unused; that draft is not counted as a semantic mutant. The final compiling mutation fails the intended assertion.
- `backend-renders-stack`: remove quiet uncaught handling. The emitted backend's stderr assertion fails even though stdout and exit 1 still match source Node.

The required counts update exposed a regression from item 4's last bracket-access edit: a binary array index reached Node.Text while checking for the literal stack key. The check now requires a string-literal index before reading its text. The existing node_buffer_bom fixture passes under both backends, sanitizers and leaks. Counts were rerun after that correction. A captured unknown slot remains an existing NotYet; the stale-narrowing witness uses a supported global unknown slot rather than broadening that cell representation.

Final commands are recorded with full output in `evidence/saved-final-restored.log.txt`, `saved-lower-final.log.txt`, `saved-vet-final.log.txt` and `saved-counts-restored.log.txt`: focused exception/closure/inheritance/Buffer oracle selection, changed lower admission and prototype/inheritance tests, narrowed compiler-package vet, and the required counts refresh. No full package tests or full gate ran. Census retirement remains **0 roots and 0 hidden bytes** for this delivery: the recorded exception refusal buckets are empty, and TS18046 remains the intended unknown-read refusal. This closes the ordered saved-origin work toward step 21 (#cxr5x2v), within the represented payload/library/class surfaces documented above.

The authorized base merge is 258a8f2c, incorporating area-next dcdbb909. The adopted toolchain setup passed after exporting GOPROXY=https://proxy.golang.org|direct and sourcing /workspace/adamic-tools/env.sh. Its cumulative timing lines were Node 0.021s, Go 0.024s, submodules 0.061s, markdown dependencies 0.071s, clang 0.145s, build 38.875s, deferred test binaries 39.005s, warm cache 39.007s and done 39.036s. nproc printed 5; the CPU quota is four. The raw setup log is `evidence/adopted-setup.log.txt`.

Final restored results: the oracle selection passed in 7.068s (native cache 0 hits/106 misses; Node 0 hits/77 misses), the selected lower tests passed in 1.883s, narrowed vet exited 0, and counts update passed in 37.943s. The saved Error fixture records 20 allocations/20 frees. The terminal narrowing fixture deliberately exits 70 with live owners; that terminal boundary neither unwinds nor runs catch/finally.

Exact final oracle command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep21|TestNativeAgreesWithNode/internal/oracle/testdata/(step21_|exceptions|closures_throw|inheritance_|node_buffer_bom)' -count=1 -v -timeout=10m`. Exact lower selection: `go test ./internal/lower -run 'TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestWhatZeroOneRefusesIsRefusedWithAFix|TestPrototype|TestInheritance' -count=1`. Vet: `go vet ./internal/lower ./internal/ir ./internal/native ./internal/javascript`. Counts: `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`.
