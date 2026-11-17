# Readonly object slots, adaptation 77

Adamic, task #5pstmvw. This one is type-only, so the emitted JavaScript doesn't
change. Every edit is an interface member, a new interface, a return type
annotation, a cast's type, a type assertion, or an `import type`, and the
emitter erases all of them.

## The edit

ObjectType declares five resolved slots, all optional, and ResolvedType
redeclares them required:

| Slot | ObjectType | ResolvedType |
|---|---|---|
| `members` | `SymbolTable \| undefined` | `SymbolTable` |
| `properties` | `Symbol[] \| undefined` | `Symbol[]` |
| `callSignatures` | `readonly Signature[] \| undefined` | `readonly Signature[]` |
| `constructSignatures` | `readonly Signature[] \| undefined` | `readonly Signature[]` |
| `indexInfos` | `readonly IndexInfo[] \| undefined` | `readonly IndexInfo[]` |

So a ResolvedType seen as ObjectType, or a `GenericType | ResolvedType` seen
that way, is a view that can write `undefined` where ResolvedType reads a
value. There are 21 refusals of that shape in checker.ts. With the five slots
readonly on ObjectType, the view can't write them.

`adapt.cjs` finds each edit through the stock AST. Each one is refused before
any write unless it finds exactly upstream's declaration or statements, and a
second run changes nothing.

**types.ts, ObjectType.** `readonly ` goes before each of the five. Each is
spelled `T | undefined` here because 20-optional-declarations, which runs
first, writes every optional member that way:

```ts
// dprint-ignore
export interface ObjectType extends Type {
    objectFlags: ObjectFlags;
    /** @internal */ readonly members?: SymbolTable | undefined;             // Properties by name
    /** @internal */ readonly properties?: Symbol[] | undefined;             // Properties
    /** @internal */ readonly callSignatures?: readonly Signature[] | undefined;      // Call signatures of type
    /** @internal */ readonly constructSignatures?: readonly Signature[] | undefined; // Construct signatures of type
    /** @internal */ readonly indexInfos?: readonly IndexInfo[] | undefined;  // Index signatures
    /** @internal */ objectTypeWithoutAbstractConstructSignatures?: ObjectType;
}
```

**types.ts, the construction view.** A new internal interface goes right after
ObjectType. It redeclares all five writable, with ObjectType's own optional
types:

```ts
/** @internal */
// What createObjectType makes: an ObjectType whose resolved slots are still being written
export interface ObjectTypeUnderConstruction extends ObjectType {
    members?: SymbolTable | undefined;
    properties?: Symbol[] | undefined;
    callSignatures?: readonly Signature[] | undefined;
    constructSignatures?: readonly Signature[] | undefined;
    indexInfos?: readonly IndexInfo[] | undefined;
}
```

**types.ts, the rest of the family.** ResolvedType already redeclares all
five, required and writable. It doesn't change, and `setStructuredTypeMembers`
writes through it as before. Every other interface in the family stays exactly
as upstream declares it: InterfaceType, TypeReference, GenericType and
AnonymousType declare none of the five, and `adapt.cjs` refuses if any of them
does. They all read the slots readonly from ObjectType, so no two of them
differ in readonly. The three writes that went through InterfaceType or
AnonymousType take a type assertion in checker.ts instead (below).

**The trap: subtype reduction.** Making a slot writable on one subtype of
ObjectType and not on its siblings changes which unions reduce. In tsc's
strict subtype relation, which subtype reduction uses, a readonly property is
never a subtype of a writable one (checker.ts:24348, "This ensures the subtype
relationship is ordered"). An earlier draft of this adaptation redeclared
AnonymousType's `callSignatures` and `constructSignatures` writable. In
getObjectTypeInstantiation, `target` is a conditional expression typed
`AnonymousType | DeferredTypeReference | GenericType`. With that draft,
GenericType, whose `callSignatures` stays readonly from ObjectType, stopped
reducing into AnonymousType. `target.instantiations` became
`Map<string, Type> | Map<string, TypeReference>`, and stock tsc refused
`target.instantiations.set(id, result)` with TS2345 (J-cpngv7y).

A second draft redeclared `members` writable on InterfaceType and on
GenericType together (GenericType has to follow InterfaceType, or it inherits
two `members` that differ in readonly and fails with TS2320). That built, and
it kept getBaseTypes' reset compiling, but it made the reset legal for every
InterfaceType everywhere. The meter showed what that cost: two sites of an
ObjectType seen as GenericType, "whose readonly field members becomes
writable" (J-mev2fpt). So the three writes take a type assertion at the write
instead. An assertion touches one statement and no declaration, so every union
over the family reduces as upstream's does, and whatever claim the assertion
makes is judged at that statement and nowhere else.

**The view only where a slot is written.** checker.ts names
ObjectTypeUnderConstruction at the six places a slot is written outside
ResolvedType, and nowhere else. createObjectType still returns ObjectType, so
every other caller sees what it saw upstream.

A third draft had createObjectType return the view. Its own writes and its two
writing callers compiled, but every caller that casts its result then cast
from the view instead of from ObjectType, and the meter judged those casts
anew (J-7zaedjy). Adamic accepts a downcast from ObjectType to a kind of
ObjectType. From the view, which claims writable slots, it refused six as "a
cast the runtime can't check": `createObjectType(...) as X` with X
InterfaceType (:13475), TypeReference (:16926), DeferredTypeReference
(:16949), MappedType (:19736), ReverseMappedType (:26499) and EvolvingArrayType
(:28642). Two more came from the return type too. In
getTypeFromTypeLiteralOrFunctionOrConstructorTypeNode, `let type`, typed by the
return (:20078), takes `createArrayType`'s plain ObjectType (:20082), which is
an ObjectType seen as the view. And instantiateAnonymousType's
`createObjectType(...) as AnonymousType` (:20967) became a relation from the
view to AnonymousType that Adamic couldn't prove (AnonymousType's optional
`target`). That's eight sites for a return type that no caller but two needed.
So the view moved to the two callers that write.

**checker.ts, createObjectType (checker.ts:5552).** It keeps upstream's return
annotation, `ObjectType`. Only its local's cast names the view, so its own
five `= undefined` writes go through it, and `return type;` is an upcast:

```ts
function createObjectType(objectFlags: ObjectFlags, symbol?: Symbol): ObjectType {
    const type = createTypeWithSymbol(TypeFlags.Object, symbol!) as ObjectTypeUnderConstruction;
```

**checker.ts, the two writing callers.** getOrCreateTypeFromSignature
(:16663) and getTypeWithoutSignatures (:21254) each write all five slots of
the type createObjectType gives them. Each declaration's initializer is cast
to the view at its end, and nothing else in them changes:

```ts
const type = createObjectType(ObjectFlags.Anonymous | ObjectFlags.SingleSignatureType, signature.declaration?.symbol) as ObjectTypeUnderConstruction;
const result = createObjectType(ObjectFlags.Anonymous, type.symbol) as ObjectTypeUnderConstruction;
```

They return ObjectType and Type, which are upcasts over the same field types.
`adapt.cjs` finds the one function of each name with its parameter list, and
the declaration once, upstream's text or already cast, and refuses anything
else.

**checker.ts, the asserted writes.** Three writes go through a type assertion
to the view. resolveBaseTypesOfClass (checker.ts:13317) resets an
InterfaceType's members (:13369), and resolveAnonymousTypeMembers
(checker.ts:14595) writes two slots after `setStructuredTypeMembers` (:14662
and :14683):

```ts
(type as ObjectTypeUnderConstruction).members = undefined;
(type as ObjectTypeUnderConstruction).callSignatures = getSignaturesOfSymbol(symbol);
(type as ObjectTypeUnderConstruction).constructSignatures = constructSignatures;
```

The emitter erases the assertion and its parentheses, so the JavaScript stays
`type.members = undefined` and so on. `adapt.cjs` finds the one
`resolveBaseTypesOfClass(type: InterfaceType)` and the one
`resolveAnonymousTypeMembers(type: AnonymousType)`, and each statement once,
upstream's text or already adapted, and refuses anything else.

**checker.ts, the import.** checker.ts gains one statement after its last
import:

```ts
import type { ObjectTypeUnderConstruction } from "./types.js";
```

The one import from `./_namespaces/ts.js` keeps its bytes. Later adaptations
(99-correctness-core-multimap among them) pin that declaration's whole text,
and they find it as the only import of its module. So the new name can't join
that list, and it can't come in through a second import of the barrel either.

All five slots are `@internal`, and so is the new interface. So the published
`typescript.d.ts` doesn't change and there's no API sanction.

The directory sorts after 76-truthful-casts and before 78-variance-marker.
Every later adaptation, and every consumer, reads apply's tree with the slots
readonly. Neither 75-optional-widening nor 76-truthful-casts edits ObjectType,
ResolvedType, InterfaceType, GenericType, AnonymousType, createObjectType,
getOrCreateTypeFromSignature, getTypeWithoutSignatures,
resolveAnonymousTypeMembers or resolveBaseTypesOfClass, so the match expects
upstream's shape.

## Certificate

As with 72 and 74, the compiled program's gate certifies it. Adamic compiles
`src/compiler` and judges every view at its relation site, so a write through
a wider view inside the compiled program is refused there, not trusted. The
count of writes below is checker-resolved over all of upstream's `src`, not
textual.

## The writers

J-jwgad54 counted the writes to the five slots in apply's tree at main
932a6344, checker-resolved over all 605 src roots. It found 24, all in
checker.ts. Upstream line numbers:

| Writer | Lines | Receiver | After 77 |
|---|---|---|---|
| createObjectType | :5555 to :5559 | ObjectType, fresh | ObjectTypeUnderConstruction, its local's cast |
| setStructuredTypeMembers | :5653 to :5659, six writes | ResolvedType | unchanged |
| getBaseTypes' reset, in resolveBaseTypesOfClass | :13369 | InterfaceType | ObjectTypeUnderConstruction, by assertion |
| resolveAnonymousTypeMembers | :14662, :14683 | AnonymousType | ObjectTypeUnderConstruction, by assertion |
| getOrCreateTypeFromSignature | :16664 to :16668 | createObjectType's return | ObjectTypeUnderConstruction, its declaration's cast |
| getTypeWithoutSignatures | :21255 to :21259 | createObjectType's return | ObjectTypeUnderConstruction, its declaration's cast; a regression, see below, closed by #2gf0k9w |

There were no writes with a key that isn't a literal, none through `any`, and
no `Object.assign` or other property road onto an ObjectType. After the edit,
every one of the 24 writes goes through an interface that declares the slot
writable: 18 through ObjectTypeUnderConstruction and 6 through ResolvedType.
None goes through ObjectType itself, an AnonymousType or an InterfaceType.

## Premises a reader can re-check

Readonly is a checker fact. Two roads could write a slot without the checker
seeing an assignment to it, and both are closed today by a stop in front of
them, not by this adaptation:

- `Object.assign`, which has a NotYet in front of it. Lifting it must judge
  readonly on the target's properties (#w9ysnpk).
- A generic key write `t[k] = v` with `K extends keyof T`, which is NotYet.
  Whoever lifts it judges readonly per property of K's constraint, under the
  ruling on writes through a key that isn't a literal (docs/rulings.md).

If either stop is lifted without that judgment, this certificate no longer
holds for the five slots.

A third premise is a refusal this adaptation makes on purpose: getBaseTypes'
reset, `(type as ObjectTypeUnderConstruction).members = undefined` in
resolveBaseTypesOfClass. It really does make an InterfaceType's `members`
undefined until resolution runs again, while other readers may hold the same
object as a ResolvedType, whose `members` promises a SymbolTable. tsc guards it
with `!members` in resolveStructuredTypeMembers, and no declared type says so.
Whoever lifts that refusal has to prove that no ResolvedType view of the type is
read between the reset and the re-resolution, or change the reset itself; a
type-level argument alone doesn't reach it.

## What stays named and refused

**getBaseTypes' reset.** At :13369, `resolveBaseTypesOfClass` (which
getBaseTypes calls) writes `type.members = undefined` through an InterfaceType
when it finds its base types resolving circularly. This really does make an
InterfaceType's members undefined until they re-resolve. The write now goes
through `(type as ObjectTypeUnderConstruction)`, which sees an InterfaceType,
whose `members` is readonly, as one whose `members` is writable. Where Adamic
refuses that, the refusal stands at the reset's own line. That's the honest
place for it: the hole is named where it is, and no declaration claims that
every InterfaceType's `members` may be written. This adaptation doesn't claim
the reset is safe.

**InterfaceType seen as ResolvedType.** Where checker.ts casts an
InterfaceType to a ResolvedType, the cast reads `members` as defined over a
type whose `members` can be reset. Those casts keep whatever refusal they have
today. Nothing here answers them.

**The view's own sites.** Each place the view is named sees a readonly slot as
a writable one, and the census judges each at its own line. createObjectType's
local and the two callers' declarations cast a fresh ObjectType, one no other
code has seen yet, to the view. The reset's assertion sees an InterfaceType as
ObjectTypeUnderConstruction, and the two assertions in
resolveAnonymousTypeMembers see an AnonymousType that way. The two anonymous
writes put defined signature lists into a type `setStructuredTypeMembers` has
just resolved, but the assertion itself claims more than that. What the view
moves at these relation sites is for the census to say, and the measure lists
every reason that grows.

## Measured

Measured on J-gb2pzgn (adaptation at 2c950a95, scratch measure `scratch-77/`
beside it, on a sealed box), with 77 present (after) and dropped (before):

| Measurement | Before | After |
|---|---:|---:|
| Writes to the five slots, all of src, checker-resolved | 24, 15 through plain ObjectType | 24: 18 through the view, 6 through ResolvedType, 0 through plain ObjectType |
| Writes with a key that isn't a literal, through `any`, or by `Object.assign` | 0 | 0 |
| Stock tsc diagnostics, all upstream src | the same set | the same set, 0 TS2540 new |
| The meter's latent total | 3,926 | 3,910 |
| The object family (a ResolvedType view of ObjectType writing undefined) | 21 | 0 |
| Lane, whole, on the after tree | | 398 units, 106,366 pass, the plan's 1 fail |

`stage3-lane-byte-comparisons` passed with `typescript.d.ts` byte-identical,
and `shard-plan.json`'s runtime holds. The emitted JavaScript doesn't change:
every edit is a declaration, an annotation, a cast's type or a type-only import.

**The mutant** leaves ObjectType's `members` writable and changes nothing else.
The object family comes back to 21.

## Reasons that grew

Three reasons grew, five sites, every one on a line this adaptation writes,
each a view that makes a readonly slot writable, which is what it is:

- **InterfaceType seen as ObjectTypeUnderConstruction, 1 site:** the reset in
  resolveBaseTypesOfClass. It really does make an InterfaceType's `members`
  undefined until it re-resolves, so its refusal is the hole named, at its own
  line.
- **AnonymousType seen as ObjectTypeUnderConstruction, 2 sites:**
  resolveAnonymousTypeMembers' writes of `callSignatures` and
  `constructSignatures`, after setStructuredTypeMembers has resolved the type.
- **ObjectType seen as ObjectTypeUnderConstruction, 2 sites:** the two writing
  callers' declarations, getOrCreateTypeFromSignature and
  getTypeWithoutSignatures, each a fresh object from createObjectType.

Net, 21 refusals go and 5 come, all five where a slot is written.

**Four are moves and one is a regression** (J-c9fy9cp, each new site placed in
its enclosing function and that function's refusals counted in the before
census). resolveBaseTypesOfClass held 11 before, resolveAnonymousTypeMembers 1
and getOrCreateTypeFromSignature 1, so those never lowered whole.
**getTypeWithoutSignatures held none and lowered whole before**; it now stops
at `createObjectType(ObjectFlags.Anonymous, type.symbol) as
ObjectTypeUnderConstruction`, a cast from the finished type back to a writable
one, which Adamic refuses, so the stop fails closed. Adamic ruled to land 77
with it named (#5pstmvw). Its follow-up is **#2gf0k9w** (Must, Memory's): a
creator whose own creation cast gives birth to the view, used only by the
writers, which leaves as the finished type, so getTypeWithoutSignatures lowers
whole again. When it lands, this paragraph says closed.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation. The measure compares each
tree's diagnostics and refusal reasons with the tree's root read as `<tree>`,
since a message naming a declaration prints its path.
