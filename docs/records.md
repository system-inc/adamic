# String records

Status: stage-3 design, October 7, 2026. This document defines the dictionary
representation being built by `codex/runtime-records` and its compiler lowering
on `codex/records-lowering`. It does not claim that records already compile.
Index signatures remain `NotYet` until their representation and operations are
held by the oracle. This stage-3 decision supersedes the index-signature policy
in [0.1](0.1.md) for the forms below; fixed-shape objects retain their own rules.

## Evidence and scope

The census at `origin/codex/tsc-census`, `stage3/census/REPORT.md`, section
"String records and post-creation properties", reports 11 index signatures in
7 original compiler files, 5 explicit `Record<string, T>` references in 4 files,
and 68 element-access sites with a string-index receiver. The last number counts
uses, including aliases, not independent dictionary declarations. The pinned
source is TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`.

`stage3/census/repro/r24/main.a` declares `interface Values { [key: string]:
number; }`, initializes `{}`, and prints `String(values)`. Its census observation
is `Refused: an index signature`. The new policy is `NotYet` until implementation,
not a claim that the census observed that spelling. No original corpus entry
reached lowering in the census; its checker diagnostics precede these operations.

`stage3/census/data/string_lookups.json` contains concrete examples:

- `core.ts:1279`: `hasOwnProperty.call(map, key) ? map[key] : undefined`, on
  `MapLike<T>`, distinguishes own entries from prototype properties.
- `core.ts:1470`: `result[key] ??= []`, on `Record<string, T[]>`, needs one
  evaluation of the receiver and key and keeps the array written back.
- `commandLineParser.ts:4142,4144,4159`: lookup, assignment and deletion on
  `MapLike<WatchDirectoryFlags>`.
- `parser.ts:10734,10741,10794`: writes to pure string-index objects, including
  a value union and a possibly missing array read.
- `commandLineParser.ts` and `utilities.ts`: many `CompilerOptions` lookups,
  which do not establish that a mixed named-property shape is a pure record.

Inference: dictionary lowering addresses a real source form, but it does not
close all 68 sites. Mixed shapes, unchecked presence assumptions, host values,
casts and unrelated syntax need their own decisions and proofs.

## Admitted type forms

A mutable record type has exactly one unrestricted string index signature and
no named properties, methods, call signatures, construct signatures, numeric
index signature or symbol index signature. Recognition uses the resolved checker
type, including inherited members, not the name of an alias or interface.

```typescript
type Values<T> = { [key: string]: T };
type ByName<T> = Record<string, T>;
interface MapLike<T> { [key: string]: T; }
```

These three spellings have the same representation after `T` is instantiated.
`T` must itself have a proven supported representation. Finite key unions such
as `Record<"left" | "right", T>` are fixed shapes, not unrestricted records.
Numeric, symbol, template-pattern, multiple and readonly index signatures remain
`NotYet` in this unit. Readonly views need a separate variance and mutation design.

An index signature beside named properties remains `NotYet`, with a diagnostic
explaining that the dictionary does not yet preserve named-property contracts.
For example, `{ required: number; [key: string]: number }` promises `required`
is present, whereas a dynamic delete could remove it. A narrower named field
also has a stronger value type than an arbitrary index write. Optional fields,
readonly fields and inherited named members do not remove this problem. Do not
erase the fields, silently route them to another store, or accept a mixed shape
through an alias. Conversions between fixed objects and records require a proven
copy or a representation-aware relation; an existing object cannot simply be
reinterpreted as a dictionary.

## Representation and planned IR

Add a distinct counted `Record` representation with string keys and a proven
homogeneous value representation. It is neither a fixed-shape `Object` nor a
JavaScript `Map`. Record expressions carry their instantiated element type;
write sites carry source/type information for invariance and cycle analysis.
The following are semantic operations, with proposed names rather than an
already implemented Go API:

| Source | IR operation and meaning |
| --- | --- |
| `{}` or a contextually typed record literal | `RecordLiteral`: allocate a record; define own data entries in source evaluation order |
| `r[k]` | `RecordRead`: own value, missing `undefined`, or an inherited-property result requiring the boundary below |
| an own-key-proven `r[k]` | `RecordReadOwn`: read only the dictionary table |
| `r[k] = value` | `RecordSet`: assignment semantics, one receiver/key/value evaluation, stores `T` and returns the assigned value where used |
| `delete r[k]` | `RecordDelete`: delete an own entry; JavaScript's delete expression returns true even when absent |
| `k in r` | `RecordHas`: own or inherited presence, not value truthiness |
| own-property test | `RecordHasOwn`: own presence, including a present undefined value |
| `Object.keys(r)` | `RecordKeys`: a new owned `string[]` of own enumerable keys |
| `Object.values(r)` | `RecordValues`: a new owned `T[]` of present entries, including undefined when `T` permits it |
| `Object.entries(r)` | `RecordEntries`: a new owned `[string, T][]`, preserving each pair and key order |
| `for (const k in r)` | `RecordForIn`: own enumerable string keys for the supported ordinary prototype; mutation semantics must be held to Node |
| `{ ...r, key: value }` into a record | `RecordSpread`: copy own enumerable entries before evaluating following fields, then define them; obey the existing single leading spread restriction |
| `JSON.stringify(r)` | record JSON schema/traversal: own enumerable entries in key order, serialized by the existing JSON rules |

Under `noUncheckedIndexedAccess`, a general `r[k]` is `T | undefined`, not `T`.
Missing numbers and booleans need their optional representations; references use
their missing representation. Presence is independent of the value: an own entry
holding undefined appears in keys, values, entries and `in`. A presence proof must
remain valid across any call or write that could delete or replace the entry.
Compound writes and `??=` preserve JavaScript evaluation order and do not perform
an unconditional write. Dot access on a record, if admitted, uses the same
record-key operation rather than a fixed object slot.

JSON omits object entries whose values are undefined or functions; nested values,
non-finite numbers and escaping retain the existing serializer's Node semantics.
A record schema must establish all possible value behavior, including `toJSON`:
an own callable `toJSON` can change serialization. Unsupported callable hooks,
replacers or value representations stay `NotYet`, never silently omitted under
a schema that claims otherwise. A record itself is not iterable by `for...of`.
Runtime iteration is an implementation tool, not a new source-level protocol.

## Key order and prototype boundary

Own array-index keys come first in ascending numeric order: canonical decimal
strings from `"0"` through `"4294967294"`. Other strings follow in first insertion
order. `"01"`, `"-0"`, `"1.0"` and `"4294967295"` are ordinary strings.
Overwriting keeps an existing key's place. Deleting and re-adding an ordinary
string moves it to the end; a re-added array-index key resumes numeric order.
Keys compare as JavaScript strings, including lone surrogates. Numeric keys,
where supported, undergo JavaScript property-key conversion exactly once.

A plain `{}` inherits Object.prototype. `"toString" in r` is true without an own
entry, and an unguarded read can return a function outside `T | undefined`.
`get_own` alone is therefore insufficient for general reads. A null-prototype
table cannot silently replace `{}` either. Prototype mutation and arbitrary
prototype-bearing objects remain outside this representation.

The forwarded C proposal has `adamic_record_new(bool reference_values)`,
`get_own`, consuming `set`, `define`, `delete`, `has_own`, `has`, `size`, owned
`keys`, and borrowed iterator results. General `get` additionally needs a tagged
result distinguishing `Missing`, `Own` and `Inherited`, with a value for `Own`.
An own undefined entry must still be tagged `Own`. The exact tagged declaration
and iterator lifetime/release convention are awaiting runtime confirmation.

Assignment and definition are different: `r["__proto__"] = value` may invoke the
inherited setter; a computed literal key or spread defines an own data entry.
The colon spelling `{ __proto__: value }` is a prototype initializer, not a data
entry, and remains `NotYet`. `define` must not invoke that setter. Assignment to
an existing own data entry named `"__proto__"` must be considered separately
from assignment to the inherited setter.

An unsupported statically known prototype operation can report compile-time
`NotYet`. A runtime cannot issue that compile-time diagnostic for a dynamic key.
Before general dynamic reads and writes land, the lead must settle either a
proven own-key restriction or explicit matching runtime guards in both backends
for unsupported inherited reads and prototype-setting assignments. Neither may
pretend the result is undefined or define a data entry in place of the setter.
The JavaScript backend may use a plain object or a Map plus explicit ordering
and prototype behavior; raw Map insertion order is insufficient.

## Ownership and soundness

The record owns counts on its stored keys and reference-bearing values, as a Map
does. `set` and `define` consume the key and value: lowering hands over an owned
count, retaining a borrowed input first. Replacing or deleting an entry lets go
of its old owned contents exactly once. The new count must be secured before the
old value is released, including self-assignment through an alias. Freeing the
record releases every remaining entry without recursive-stack growth.

`get_own`, tagged `get` and iterator outputs borrow. A read kept beyond the
immediate operation takes its own count before any subsequent evaluation can
mutate the table. Keys/values/entries arrays and spread copies own their contents,
so they survive deletion and record destruction. Iterators must keep the record
alive and define when borrowed outputs cease to be valid; table reallocation
cannot leave a saved pointer dangling. Loop exit, break, return and exception
cleanup release iterator ownership exactly once.

Record values are mutable slots in the cycle finder, just like Map values.
Reaching follows the record's values, including records nested through objects,
arrays, unions or closures. Every set, define, literal and spread write must be
recorded and judged by the fresh-write proof; an unknown operation stays
conservative. A strong value that can reach its holder cannot bypass the cycle
finder merely because its field name is dynamic. Weak values need their existing
handle representation and cannot be treated as strong references.

Records are invariant mutable containers. `Record<string, Dog>` is not a
`Record<string, Animal>`: the wider alias could store an Animal without `bark`,
then the narrower alias would read it as a Dog. Apply this rule wherever a value
enters a typed slot, including arguments, returns, fields, captures, unions,
assertions and inferred widening. A separately allocated copy can be judged by
its contents; aliases of the same mutable record cannot be widened.

## Acceptance evidence required for lowering

Hold each operation to original source on Node, the JavaScript backend, and native
under ASan, UBSan and the leak check. Use `.a` fixtures with runtime-built keys
and values for ownership probes. Cover numeric-looking key boundaries, overwrite,
delete/re-add, missing reads, present undefined, prototype names, both kinds of
`__proto__` write, evaluation order, spread snapshots, JSON and mutation during
for-in. Use the census snippets above with their adaptations stated explicitly;
do not claim the untouched tsc corpus compiles.

Run a mutant per operation that changes its actual result or ownership, and say
which oracle observation or sanitizer catches it. Independently prove refusal
checks can fail by allowing a mixed shape, mutable widening or cycle-closing
write and running its failing probe. Counts-table changes must be recorded.
This design-only step introduces no compiler checks and runs no compiler mutants;
the list is a requirement for implementation, not completed evidence.
