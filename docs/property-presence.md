# Property presence, enumeration and primitive conversion

Design only. No runtime, compiler or fixture implementation is proposed in this
commit. This follows [the stage 3 runtime root census](stage3-runtime-roots.md).
The scope is the ten general-object `for...in` sites, one array `for...in` site
and two template-interpolation sites in the frozen morning census. Counts are
historical observations, not a promise that this design alone retires them.

## Contract and current boundary

A property's existence is independent of its value and its static type.
`{x: undefined}` owns `x`; `{}` does not. An absent optional field synthesized
for a structural view must not become an own property. Likewise `[undefined]`
has index `0`, whereas an array hole does not. A non-enumerable own property can
hide an enumerable property of the same name on a prototype. Checking the value
for undefined, NULL or zero cannot implement any of these distinctions.

At the task base, `adamic_shape` in `internal/native/runtime/adamic.h` describes
names, a reference bitmap and method tables. Object slots are fixed, arrays are
dense and `adamic_object_keys` enumerates a known own shape. The packed slot
cache assumes immutable shape identity. The fetched `area/runtime` base
`c1c6073021e8e5cd6005fcdee59eb3b377395fd8` still has no field-kind array.
The **new shape field kinds** below are a companion metadata proposal, not a
claim that an unmerged implementation already exists on this branch. If that
work lands first, reuse its tags and ownership rules rather than adding a second
independent type system.

The proposed supported subset consists of runtime-created ordinary objects and
explicitly described built-in objects. Arbitrary JS host objects, proxies and
an unrestricted mutable prototype system remain outside the initial contract.
An interface's static properties never stand in for the actual object's keys.

## Shape descriptors and value metadata

Keep immutable, interned storage shapes. Each own property descriptor identifies:

- A length-delimited UTF-16 string key or supported symbol identity, a slot or
  accessor entry, and enumerable/configurable/writable bits. String keys must
  represent lone surrogates and embedded NUL; C-string names alone are inadequate
  for general property keys. Private class fields are not property descriptors.
- Its field kind: number, boolean, string, undefined, null, optional scalar,
  tagged union, object, array, map or callable. A kind describes how to read a
  slot, not whether the property exists. Container kinds point to element/value
  metadata when needed. Reference retention remains separately derivable and
  auditable; do not reinterpret a boolean's unspecified high bits as a number.
- Stable creation order and whether the descriptor is data or accessor. An
  accessor needs a receiver-aware invocation thunk and result-kind metadata;
  enumeration reads its descriptor without invoking its getter.

The shape kind table is shared across instances. Homogeneous arrays/maps can
share a payload kind/schema; mixed containers need a tagged element schema.
Dynamic properties store a tagged value or a descriptor-kind reference, not an
untyped `adamic_value` that a generic consumer guesses how to decode. Optional
scalar packing and the present-undefined distinction require explicit adapters.
A tagged runtime value includes the primitive/reference discriminant needed to
interpret an arbitrary getter or conversion-hook result. Heap kind alone does
not distinguish all scalar values or each field of a compound value.

Use a shape flag for “all own slots always present”. Such objects need no
presence bitmap. Shapes with conditionally present storage use a per-instance
bitmap; initializing an optional slot to undefined does not set its presence bit.
Ordinary writes set presence only when a JS property is actually created.
Deleting a configurable property clears presence and releases/zeros its stored
value, with kind-correct ownership. Deleting a non-configurable property follows
the supported strict/sloppy language mode; it must not silently clear the bit.

Retain immutable descriptors across mutations. Adding properties or changing
attributes either transitions to a new interned shape or promotes the object to
a descriptor dictionary. Deletion followed by recreation gets a new creation
ordinal for non-index string keys. Reusing an old slot must not reuse its old
enumeration order. The dictionary also stores complete descriptor kinds.

## Per-kind storage

Add one nullable `property_state` pointer to each property-capable heap layout.
A null pointer selects an immutable default prototype/descriptor recipe based on
heap kind and, for class instances, class identity. Allocate the sidecar only
for optional presence, sparse storage, extra properties or a supported prototype
override. The sidecar owns its bitmap, descriptor dictionary, insertion ordinals,
prototype override and structural generation. It does not copy structural views.

| Kind | Own descriptors and presence | Prototype and conversion information |
| --- | --- | --- |
| Plain object | Storage shape, descriptor attributes and optional bitmap; dictionary after supported additions/recreation. Literal properties are enumerable; absent structural padding is not an own property. | Explicit ordinary-object or null prototype recipe. Optional prototype override is a retained object reference. Runtime lookup sees actual own descriptors and prototype descriptors. |
| Class instance | Actual emitted instance shape; initializer/constructor execution establishes which public fields exist. Optional declared fields have storage but no presence until created. Private fields and internal slots are excluded. | Class identity selects a shared prototype descriptor object containing non-enumerable methods/accessors and a base-prototype link. Public fields may shadow methods. Conversion hooks use normal lookup with the instance as receiver. |
| Array | Existing length/capacity/elements plus payload kind; null hole bitmap means dense. A bitmap is allocated on the first hole/deletion. Extra own string/symbol properties live in the sidecar; regex result properties must migrate into this same descriptor model. | Shared Array prototype recipe and optional override. `length` is an own non-enumerable, non-configurable property with its special writable semantics. Array-specific ToString delegates to a looked-up `join`, not unconditionally to a private helper. |
| Map | Entries retain their existing hashing/iteration storage and kind metadata. Extra JS own properties have a separate property sidecar. Map entries and `size` are not own enumerable properties. | Shared Map prototype recipe: methods/accessors and Symbol.toStringTag are non-enumerable. Default object conversion is `[object Map]`; it never enumerates or stringifies entries to implement ToString. |
| Closure/function object | Existing code/environment plus property-state pointer, callable ABI descriptor and source-text identity. Described `name`/`length` are non-enumerable own properties. Constructor-capable functions additionally describe their own `prototype` property; arrow functions must not acquire one. | Shared Function prototype recipe, with supported callable `toString` and `valueOf`. Function source metadata preserves JS source spelling or a defined native/bound-function form. The address of emitted C code is not JS function text. |

Prototype descriptor recipes can be shared immutable runtime objects, lazily
materialized if identity is observed. Class prototypes need consistent identity
across instances and inheritance. A built-in prototype has real property
attributes and callable descriptors, not a string-only list of method names.
`Object.create(null)` needs a distinct null-prototype recipe: it cannot inherit
`valueOf` or `toString` by accident.

A dense array with a non-null sidecar may still have a null hole bitmap. Growing
length creates holes; writing beyond length creates intervening holes and grows
length. A present undefined element retains its presence bit. Shrinking length
deletes indices above the new bound under the array descriptor rules. Index
keys are canonical strings in `0..2^32-2`; `4294967295`, `-0` and `01` are ordinary
string keys, not array indices. Runtime-only length/capacity bookkeeping is not
a user-visible enumerable field.

Property ownership must participate in release, graph/region analysis and any
cycle policy. A retained prototype link or hook closure can introduce a cycle.
A loop iterator retains its receiver/prototype objects until closed, including
on break, return and exception. None of those edges may be invisible to the
existing reference-count/cycle machinery.

## Descriptor lookup and for...in

Introduce logical operations for own-descriptor lookup, descriptor-only own key
listing, and prototype lookup. They apply to all property-capable heap kinds.
A generic Get returns the actual value or invokes a getter with the original
receiver, including when the getter was found on a prototype. An own-descriptor
query never evaluates a getter. Callable data fields, class methods and built-in
methods use receiver-aware adapters rather than guessing a closure signature.

An enumeration iterator carries the retained receiver/current prototype, a
snapshot of that object's own string keys, a cursor and a visited-key set.
For the initial subset, membership, attributes and prototype links remain stable
through the loop. Values can change. At each object in the chain:

1. Obtain its own string keys: canonical index keys ascending, then remaining
   string keys in creation order. Symbols are excluded from for...in.
2. For each unvisited key, look up its own descriptor without Get. If it exists,
   mark the key visited **even if non-enumerable**; yield only enumerable keys.
   A non-enumerable shadow suppresses the inherited property. Absent padding and
   array holes have no descriptor and are not marked visited.
3. Advance to the next prototype and repeat, stopping at null. Built-in methods,
   class methods and array length are normally non-enumerable but still matter
   for shadowing. Map entries and closure capture slots never enter the key list.

The implementation should recheck presence before yielding, both as defensive
validation and to enable a later deletion-aware extension. This alone does not
promise arbitrary mutation-during-enumeration semantics. ECMAScript permits
variation for several mutation cases; choose and test an explicit supported
policy before admitting them, rather than claiming Node's incidental behavior
is a universal rule. The first version refuses loops whose effects may add/delete
properties, change attributes/prototypes or invoke such behavior through unknown
calls. It may support precise deletion rules later: an absent descriptor is not
marked visited, so the same name on a prototype remains eligible.

Property-graph stability is an effect contract over aliases, calls and getters;
“the loop body does not syntactically write the iterable” is insufficient.
Version checks detect violated native invariants, but must not substitute a new
exception for an otherwise valid supported JS program. When the compiler cannot
establish the contract, the source remains refused. Prototype traversal itself
invokes no getters and runs no user code.

## Generic ToPrimitive and ToString

Template substitution performs ToString on each evaluated expression in source
order. It is not JSON.stringify, and an object's structural view cannot erase
conversion hooks. Evaluate the operand once; retain it throughout lookup/calls;
preserve hook side effects, exceptions and changes visible to later lookups.

For objects, arrays, maps and functions, ToPrimitive with hint `string` performs:

1. Get the supported well-known `Symbol.toPrimitive` property through the actual
   descriptor/prototype chain. Undefined/null means no exotic hook. Otherwise
   require a callable value, invoke it with the original receiver and the string
   `"string"`, and require a primitive result; non-callable/non-primitive results
   raise the appropriate JS TypeError.
2. Without that hook, perform OrdinaryToPrimitive in string-hint order:
   look up `toString`, call it if callable, accept a primitive result; otherwise
   look up `valueOf`, call if callable, accept a primitive result. A missing or
   non-callable ordinary method is skipped. If neither produces a primitive,
   raise TypeError. Do not prefetch both methods: the first can modify the second.
3. Convert the primitive result with the normal JS ToString rules: undefined,
   null and booleans have their standard spellings; numbers use the existing JS
   formatter, including `-0` becoming `0`; strings preserve UTF-16 contents.
   Abstract ToString on a symbol throws (template substitution does not use the
   special `String(symbol)` exception). BigInt support needs its own decimal
   conversion and representation; it remains refused in this initial subset.

Default Object.prototype.toString consults the supported Symbol.toStringTag
property and a correct built-in fallback tag. Custom string tags and getters
are observable. Ordinary valueOf returns the object itself and is not a
primitive success. A null-prototype object with no usable hooks therefore fails
ToPrimitive; do not manufacture `[object Object]` for it.

Array.prototype.toString gets `join` from the actual array and calls it if
callable; otherwise it invokes Object.prototype.toString. The default join
visits positions up to its captured length, using ordinary indexed Get: holes
may resolve to inherited indexed properties. Missing, undefined and null values
contribute empty strings; other values use their string conversion with its
side effects. This is separate from for...in membership. A nested array must not
be flattened using an untyped slot cast. Cyclic arrays and side-effectful joins
require an explicit, tested recursion policy; refuse unproved cycles initially.

Function.prototype.toString needs preserved source slices for ordinary source
functions and the specified native-style form for supported built-ins/bound
functions. Preserve original JS text through lowering, including arrows and
method syntax; the emitted function name is not an acceptable replacement.
Do not claim generic function interpolation until this source/ABI metadata exists.
Unsupported dynamic/bound constructors remain refused rather than emitting an
invented source string. Captures are not enumerated or printed as properties.

Conversions need an emitter-generated callable ABI adapter that packages the
receiver, hint/arguments and tagged primitive result. A runtime hook must check
callability and result kind; the compiler cannot certify either from a structural
annotation. Normal cleanup applies on both ordinary return and exceptions.

## Compiler proofs and runtime checks

| Compiler responsibility | Runtime responsibility |
| --- | --- |
| Emit complete actual shapes, kinds, own descriptors, prototype recipes and initializer-presence transitions; ensure all constructors/aliases preserve them. | Query actual presence and attributes; traverse actual supported prototypes; never replace descriptors with the checker-visible field list. |
| Preserve storage invariance, reference ownership, callable receiver ABI and typed result adapters. Reject unrepresented exotic/host kinds and effects outside the supported subset. | Decode fields using shape kinds, discriminate generic hook results, check callability, throw JS conversion errors and perform correct retention/release. |
| Prove stable enumeration membership/attributes/prototypes, or prove a narrower fixed-origin path. Account for aliases and transitive effects. | Produce key order, suppress shadowed names and omit holes; validate the approved descriptor model. Invariant validation is not an implementation of arbitrary JS mutations. |
| Prove hooks/prototype slots cannot change before eliding generic lookup; avoid converting object references from static interface information alone. | Perform observable Gets/calls in order with the original receiver; do not cache a getter's value or user conversion result as a shape property. |
| Preserve function source slices, container schemas and well-known symbol identities. | Format primitive results and apply each built-in object's conversion behavior, including errors. |

A structural cast/view preserves the object's identity, own presence and prototype;
it neither initializes absent fields nor constructs a new property list. Exported
native APIs must obey the same descriptor contract as compiler-emitted allocations.

## Cost and existing fast paths

Choose an explicit tradeoff: one nullable sidecar pointer per property-capable
allocation, with lazy sidecar contents. On a conventional 64-bit ABI that is an
additional eight header bytes, before allocator-size-class effects, for objects,
arrays, maps and closures. Flexible-array offsets and all allocation/GC consumers
must be updated together. No exact RSS or performance percentage is claimed.
An alternative heap side table avoids the header word but adds lookup,
synchronization and lifetime costs; it is not the chosen baseline.

For F potentially absent fields, bitmap storage is `ceil(F/8)` bytes plus sidecar
allocation/alignment. Sparse arrays pay `ceil(capacity/8)` for the initial bitmap;
very sparse arrays should promote to an index dictionary rather than allocate a
bitmap proportional to a huge logical length. Descriptor dictionaries, order
indices, enumeration snapshots and visited sets add costs only to generic paths.
Runtime prototype objects are shared, and the default recipe requires no per-instance
prototype allocation. An explicit prototype override introduces a retained edge.

New shape field kinds add immutable per-shape metadata, not a kind word to every
existing scalar slot. The exact descriptor size is an ABI choice; a compact
kind tag plus shared schema pointer/index is sufficient in principle. Keep the
existing references table during migration, or derive it from kinds without
putting a kind switch into every release loop. Release remains correct for tagged
unions whose live payload may be a reference. Presence-aware objects additionally
skip absent slots; initialized absent reference storage stays zero for safety.

The present packed cache is one atomic 64-bit immutable-shape identity/slot word.
Preserve it for fixed own-slot reads and writes when the compiler proves presence
and representation. Shared descriptor kinds require no per-read lookup beyond
the already known shape/slot, and no atomic per-slot kind mutation.

A cached slot is **not a cached existence result**. If a shape can have absent
slots, presence must be tested after a cache hit unless flow/effect proof makes
it unnecessary. An absent own slot may fall back to a prototype; it must not be
read as an owned undefined value. Existing `hasOwnProperty`, `in`, Object.keys
and object copy/spread must adopt descriptor presence too. Optional packing's
`present` flag describes a value, not own-property existence.

Generic prototype/dictionary lookup uses a separate cache structure with shape
and structural-generation guards for the receiver and traversed chain, or takes
an uncached slow path initially. Do not cram mutable generation/prototype state
into the existing 48-bit-shape/16-bit-slot encoding. Dictionary reallocation,
shape transitions, deletion/recreation, attribute changes and prototype changes
invalidate affected plans. Negative lookups need equivalent chain guards.
Value-only writes need not invalidate descriptor caches, but a conversion cache
must fetch the latest hook value and must never memoize arbitrary hook results.

The dense-array fast path keeps its current indexed storage with a null hole
bitmap. Present-index proofs can keep direct reads; generic indexed Get checks
holes and inherited properties. Array join with default prototypes and a dense
primitive payload can use a specialized helper only after proving hook/prototype
and element-conversion behavior. Reference arrays cannot acquire a generic
string interpretation merely because their `references` flag is true.

Direct class dispatch can remain fast under a proof that relevant method/hook
properties cannot be overridden; otherwise emitted calls must observe own
shadowing and receiver semantics. This design does not quietly solve the separate
17 method-replacement roots. Method-table and sidecar dispatch integration is a
separate compiler/runtime contract.

## What remains refused

The initial design admits ordinary represented objects with complete descriptors,
known supported prototypes and an established enumeration effect contract. It
continues to refuse:

- Proxies and their ownKeys/getOwnPropertyDescriptor/getPrototypeOf/Get traps;
  arbitrary host objects and exotic descriptor algorithms not represented here.
- Unknown prototype mutation, descriptor mutation during active enumeration,
  unrestricted dictionary effects through unknown aliases/calls, and prototype
  cycles. Supported prototype creation must prove/check acyclicity and integrate
  retained links with the cycle policy.
- Arbitrary user symbols beyond explicitly supported well-known identities,
  unrepresented getters/setters, unrepresented callable ABI results and unsupported
  function source forms. Symbol keys are never stringified into enumeration keys.
- BigInt conversions until the runtime has its representation, and unproved cyclic
  array conversion. JSON.stringify/toJSON is a separate protocol; these changes
  do not implement it or dynamic RegExp construction.
- Generic/brand/type-proof gaps elsewhere in lowering. Removing these 13 guards
  without complete descriptors, kinds, hooks and effect proofs is unsound.

A future implementation should cover each admitted feature with Node comparisons
and mutants for absence versus undefined, non-enumerable shadows, hole versus
inherited index, key ordering, wrong receiver, wrong hint, lookup ordering, field
kind decoding, symbol TypeError and function source preservation. These are a
validation plan, not tests claimed to have run in this design-only task.

## Requirements for each of the 13 measured roots

The locations below are exact entries from
`e8c283b5ed32477805357b652a170b85a04b2469`:
`stage3/notyet-table/rerun-0730/after/roots.csv`. The listed needs follow their
reason category; without replaying their complete binding/effect context they
are not claims about each site's exact runtime shape or successful retirement.

| Root location | Category | What admitting this site would need |
| --- | --- | --- |
| src/compiler/commandLineParser.ts:2788:24 | Object for...in | Actual receiver descriptors/presence and prototype key traversal; establish stable enumeration effects for this receiver's aliases and calls. |
| src/compiler/commandLineParser.ts:2975:24 | Object for...in | Enumerate runtime own keys rather than the structural field list; preserve key order and non-enumerable shadowing, with a stable-chain proof. |
| src/compiler/commandLineParser.ts:4265:23 | Object for...in | Complete possible receiver-kind metadata, including arrays if the structural view permits them; actual presence plus an enumeration effect contract. |
| src/compiler/core.ts:1289:23 | Object for...in | Runtime key descriptors for each possible object, absent optional-field handling and supported prototype traversal; prove membership/attributes remain stable. |
| src/compiler/core.ts:1314:23 | Object for...in | Identity-preserving structural views and actual own presence; distinguish enumerable inherited keys from own non-enumerable shadows and verify effects. |
| src/compiler/debug.ts:432:28 | Object for...in | Supported concrete object/host boundary, actual descriptors and chain; refuse an unrepresented host receiver rather than inventing keys. |
| src/compiler/factory/nodeFactory.ts:6137:25 | Object for...in | Instance/shape field kinds and initializer presence, excluding synthetic optional slots; own/prototype descriptors plus stable enumeration effects. |
| src/compiler/moduleNameResolver.ts:2428:23 | Object for...in | Actual receiver identity and own creation order, optional-field presence and prototype duplicate suppression; establish the loop effect contract. |
| src/compiler/moduleNameResolver.ts:432:27 | Object for...in | Full receiver-kind/prototype descriptors instead of static-map assumptions; Map entries must not become own properties if a Map is possible. |
| src/compiler/moduleSpecifiers.ts:927:23 | Object for...in | Presence-aware own descriptors and prototype traversal for every admitted receiver kind; prove no unmodeled membership or attribute mutation. |
| src/compiler/factory/nodeFactory.ts:7538:23 | Array for...in | Hole/index presence, ascending canonical indices, extra string-property insertion order, non-enumerable length and prototype shadowing; stable enumeration effects. |
| src/compiler/core.ts:1786:59 | Template conversion | Tagged operand/result kinds; string-hint ToPrimitive with observable receiver-aware hook Gets/calls, primitive ToString and propagated errors; source metadata if a callable is possible. |
| src/compiler/semver.ts:452:37 | Template conversion | Concrete/tagged conversion of each interpolation in source order, with correct object/array/function defaults and hooks for the actual operand; no untyped reference formatting. |

## Design validation

Documentation only on `runtime/stage3-runtime-roots`. Read the runtime layouts,
slot cache/lookup and emitter layout evidence on the task branch; fetched the
branch and census by name and inspected current `area/runtime` shape metadata.
The 13 location/category lines are checked against the pinned CSV, including
10 object, one array and two conversion sites, without counting attempts or
rollback echoes. No semantic tests, Node fixtures, runtime changes or benchmarks
are claimed. Implementation decisions that still need prototypes are descriptor
packing, sparse-array promotion thresholds, cycle treatment and generic callable
adapter details; none may be settled by weakening the existing refusal guards.
