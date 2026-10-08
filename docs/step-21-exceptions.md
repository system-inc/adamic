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
   Fixed points must compare set contents, not just cardinality. The current
   `LiveOut` updates a block only when its live-set size changes; a same-size
   replacement can hide a later iteration's change. This is a compiler correctness
   issue within the existing language, suitable for the implementation unit.
8. **Other analyses:** initialization, freshness, aliasing and region escape must
   see throws and finalizers too. SSA assignment definitions must not be mistaken
   for stores that happened on the exceptional edge.
9. **Panics:** invariant checks remain terminal and flush stdout; catch/finally
   must not execute after them. Resource failures and engine stack text cannot be
   claimed byte-identical without a defined contract and external oracle evidence.

The highest source exposure is already-admitted try/catch. The first build will
hold that shape to Node and repair liveness convergence without admitting a new
throw kind. It cannot truthfully retire the eight TS18046 diagnostics, saved-Error
NotYet or arbitrary-value refusal. Their measured exception-root retirement is
zero until a reviewed admission change and a census replay establish otherwise.
