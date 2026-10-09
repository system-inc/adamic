# Callable and accessor storage for optional calls

Proposal for @system_adamic, task #tvq1eqm, roadmap step 18. This document
requests a ruling; it does not implement or approve the pending cases.
Implementation baseline: compiler/optional-calls-main 28bf5bfbea924356e4e82c7de5848c053f4707f6.

## Proposed ruling

Admit an optional invocation when the selected value has a proven callable
signature, its storage and result are represented, and selection preserves Node's
receiver and evaluation order. A function-valued data field and a getter returning
a function are different storage shapes. Both may support direct optional calls,
but a getter must execute as a read, never become a cached function field.
Keep the existing refusal for detached methods; do not silently bind them.
Weak and unchecked callable storage stay NotYet until their own contracts apply.
Approval of this proposal alone does not make any pending fixture pass.

## Evidence from tsc

Locations below are in the adapted TypeScript 6.0.3 source pinned at
050880ce59e30b356b686bd3144efe24f875ebc8, used by the rebuild. These are source
observations, not proof that the whole entry compiles.

| Shape | Observed source and limits |
|---|---|
| Field holding a function | `program.ts:1882` initializes `getModuleResolutionCache` with an arrow; `program.ts:1605` and `1624` invoke `host.getModuleResolutionCache?.()`. The declared host member is an optional method at `types.ts:8166`, so declaration syntax alone cannot decide whether the allocation stores a callback or a prototype method. |
| Getter returning a function | The pending reduced class-getter witness is in `TestStep18OptionalCallableBoundaries/chain_getter_selection`. The historical `chain-roots.json` records zero roots for callable accessor selection. No production tsc optional-call getter occurrence was established; this is a storage boundary to rule, not a measured tsc stop. |
| Member read and called later | `program.ts:525-529` saves `host.readFile`, `fileExists`, `directoryExists`, `createDirectory` and `writeFile`; `program.ts:1734` explicitly binds `host.readFile` to `host`. These establish that host functions escape their member expression, not that every saved function is later called with `?.`. The reduced optional-local witness below pins that combination. |
| Optional method | `checker.ts:6910` calls `context.tracker.reportInaccessibleThisError?.()`; `types.ts:10060` declares that optional method. `checker.ts:7093` also guards the tracker itself. `program.ts:1605` guards only the callable, not the host. |

The entry still first stops at `src/compiler/builder.ts:1246:69`, TS2345
(`Path | undefined` passed as `string`). No new entry run or census retirement is
claimed for this documentation unit. [The rebuild audit](main-rebuild.md) records
the executed measurement and pending boundaries.

## Meaning and proposed lowering by shape

| Shape | What Node does | Proposed Adamic storage and invocation | Refused or pending |
|---|---|---|---|
| Function-valued field: `object.callback?.(arg())` | Read the field once, test its value for null/undefined, then evaluate arguments and invoke that saved value with `object` as `this`. Arrows keep their lexical `this`; ordinary functions receive the object. Argument writes cannot change which function this call uses. | An owned closure slot with nullable presence; save and retain the object and selected closure before arguments. Carry the call-site receiver separately from the closure's captured environment. Use the selected function's real parameter/result ABI, including discarded reference results. | Invalid writes and unsound callable variance remain Refused under existing rules. Checked or uninitialized callable fields remain NotYet until callable readiness and signature validation are represented. Weak slots remain NotYet; absence is not a substitute for an approved Weak upgrade. |
| Getter returning a function: `object.callback?.(arg())` or `object?.callback?.(arg())` | A live object invokes its getter once with the object as `this`, before the callable check and arguments. If the getter returns undefined, arguments do not run. The second spelling skips the getter when the object is null/undefined. A returned ordinary function is then called with the original object as `this`; a returned arrow keeps its lexical `this`. | Store a represented accessor entry, with receiver-aware getter dispatch and an explicit returned closure ABI, separate from a data slot. Guard an optional object first; execute the getter once, retain its returned closure, guard that closure, then evaluate arguments. Preserve getter effects on every read and retain the original object across them. | Currently NotYet: callable accessor selection. Propose admission only for known represented accessors with proven signatures and ownership. Dynamic descriptor mutation, unknown accessor dispatch or unrepresented return storage remain NotYet; forbidden reflection stays Refused. This does not approve setter behavior. |
| Read then call: `const callback = object.callback; callback?.(arg())` | Perform the member read, including any getter, at assignment time. The later call uses the saved value and has no object receiver (`this` is undefined in strict code). It does not re-read the field or getter. A lexical arrow still keeps its captures. | Copy/retain a receiver-independent closure into the local and release it at the local's lifetime end. The later optional call snapshots that local and passes no object receiver. A getter-produced ordinary function that requires a receiver cannot be converted into a bound closure implicitly. | A true method read as a value remains Refused (`unbound-method`), including a structural function-property view hiding a method. Proposal: admit detached data callbacks only when receiver independence is proved across their origins. Unknown origins stay NotYet; a proven receiver-dependent origin is Refused. The existing explicit arrow wrapper keeps its object. Broader `bind` support requires its own proven lowering. |
| Optional method: `object.method?.(arg())` and `object?.method?.(arg())` | Select the own property or inherited method once, guard the selected value, and invoke it with the saved object as `this`. An absent own property may expose an inherited method; an own undefined value shadows it. Only null/undefined skip: a non-callable present value throws after evaluating arguments. | Represent method selection as a saved pair of object and callable entry. Keep own-property presence distinct from its value and from prototype fallback; do not synthesize a callable for an absent optional method. Resolve checked structural views through their existing validation/readiness path, then snapshot the selected entry before arguments. | Represented direct methods already work. Method detachment and unsound method variance remain Refused. Checked/uninitialized callable members remain NotYet until represented validation can prove the selected signature. Overloads or other unsupported signatures remain NotYet; an optional check is no permission for an unsafe call. |

`object?.callback(args)` guards the object only: a live object's missing callback
is still a failed ordinary call. `object.callback?.(args)` guards the callable
only: a missing object still fails on the read. Parentheses around a member
reference, `(object.callback)?.(args)`, preserve its receiver; assigning it to a
local or selecting it with a comma expression does not. Optional absence yields
undefined, not zero or an empty string. A present function returning undefined
still ran, and must not be confused with an absent invocation in a continuation.

## Storage checks and ownership offered for approval

A typed callable slot must prove its argument/result contract on initialization,
assignment and checked-view access. Mutable aliases retain existing invariance;
method syntax must not bypass the audited parameter and result judgments. A
readonly structural view does not prove that a hidden getter is a data field.
Selection therefore needs the allocation's member kind, or represented runtime
dispatch that distinguishes data, accessor and method entries.

For an admitted call, keep the receiver and selected callable alive across getter
and argument effects, including writes that replace the field or last owning
reference. Release saved references after preserving the result. A getter's
returned closure has its own owned lifetime; its capture is not the call-site
receiver. Optional results use a genuine undefined representation, and invocation
presence remains separate from returned undefined. Observed void results,
erased never-rest markers, unsupported overloads and narrowed optional-result
representations remain the rebuild's NotYet boundaries, outside this ruling.

A check for null/undefined is not a callable type check. Unsafe `any`, unchecked
casts and unbound methods keep their refusals. Represented checked storage must
use the already ruled failure contract for invalid values before invoking them;
this proposal adds no exception handling, new panic text or unchecked signature
coercion. Until that check exists, the case is NotYet.

## Reduced Node observations

The following strict JavaScript probe is reproducible directly on Node. It is a
semantics witness inside this proposal, not an Adamic acceptance fixture. Run the
single `javascript` block with `node --input-type=module`; the assertions below
passed on Node v24.19.0. Output was written to
`/tmp/optional-storage-contract-node.log`. No compiler or runtime was changed,
no counts or stage 3 records were regenerated, and no compiler mutants were
rerun for a documentation-only proposal.

```javascript
import assert from 'node:assert/strict';
const events = [];
const arg = () => { events.push('arg'); return 7; };
const owner = {
  label: 'owner',
  callback(value) { events.push(`old:${this.label}:${value}`); return value; },
};
const replace = () => {
  events.push('replace');
  owner.callback = () => events.push('new');
  return 7;
};
owner.callback?.(replace());
const getter = {
  label: 'getter-owner',
  get callback() {
    events.push('get');
    return function(value) { events.push(`call:${this.label}:${value}`); };
  },
};
getter.callback?.(arg());
const absentGetter = { get callback() { events.push('get-absent'); return undefined; } };
absentGetter.callback?.(arg());
const absentObject = undefined;
absentObject?.callback?.(arg());
const saved = getter.callback;
events.push('saved');
// Use a receiver-observing function whose body is safe when detached.
const independent = { callback() { events.push(this === undefined ? 'local:no-this' : 'member:this'); } };
independent.callback?.();
const local = independent.callback;
local?.();
assert.throws(() => saved?.(7), TypeError);
const prototype = { method() { events.push(`prototype:${this.label}`); } };
const inherited = Object.create(prototype);
inherited.label = 'child';
inherited.method?.();
inherited.method = undefined;
inherited.method?.(arg());
const lexical = { callback: () => events.push(this === undefined ? 'arrow:no-this' : 'arrow:this') };
lexical.callback?.();
const nonCallable = { callback: 0 };
assert.throws(() => nonCallable.callback?.(arg()), TypeError);
assert.deepEqual(events, [
  'replace', 'old:owner:7',
  'get', 'arg', 'call:getter-owner:7', 'get-absent',
  'get', 'saved', 'member:this', 'local:no-this',
  'prototype:child', 'arrow:no-this', 'arg',
]);
console.log(JSON.stringify(events));
```

These observations pin getter-before-check ordering, absence skipping arguments,
snapshotting across replacement, inherited selection and undefined shadowing,
lexical versus call-site receivers, and receiver loss on detachment. The
non-callable probe illustrates Node's behavior outside sound callable storage;
it is not a request to admit that program to Adamic.

## Decision requested

Approve direct optional calls through proven data callbacks, represented callable
getters and represented optional methods with the selection rules above. Preserve
Refused for detached methods; permit copying data callbacks only with a proof of
receiver independence. Keep unknown/Weak/checked-unrepresented storage NotYet.
A ruling would authorize subsequent implementation and Node fixtures, including
mutants for repeated getter reads, checking before the getter, eager arguments,
callee reselection, accidental binding and dropped receiver/callable ownership.
