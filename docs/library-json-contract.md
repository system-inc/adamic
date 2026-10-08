# Library JSON encoder contract

Source anchors without a revision mean `f5d34f90`; `l2:` means `23538b4f`.
**New** requirements are interfaces to implement, not existing base capabilities.
Node **24.19.0** decides serialization, hook calls and throws.

## 1. Runtime array element descriptors

Arrays must carry descriptors of their **actual stored elements** across factories,
copies and writes. `references` controls ownership only (`internal/native/runtime/adamic.h:358`); it cannot distinguish
number from boolean, or string from object. Structural views cannot establish storage
(`internal/lower/library_json_stringify.go:190`).

The element reader returns a borrowed `(adamic_value, const adamic_json_schema *)`
for each index below `array->length`. A homogeneous array reads
`array->elements[index]` with `schema->element`, exactly as
`internal/native/runtime/json_stringify.c:176` does today. **New:** runtime factories
must carry that complete descriptor; heterogeneous arrays need a descriptor/tag
per stored element and an element reader exposing the same pair. They must not use
`adamic_json_tuple`: today's tuple reads `adamic_object` slots, not array storage.

| Resolved descriptor kind | Slot read | Array JSON |
|---|---|---|
| `number` | `.number` | Node number spelling; `-0` becomes `0`; NaN and infinities become `null` |
| `string` | `.reference` as `adamic_string *` | Quoted and escaped UTF-16 string |
| `boolean` | `.boolean` | `true` or `false` |
| `null` | No payload read | `null` |
| `object` | `.reference` as `adamic_object *`, complete actual field schemas | Recursively encoded object |
| `array` | `.reference` as `adamic_array *`, complete element descriptor | Recursively encoded array |
| `undefined` / absent index | No payload read | `null` |

Basis: `internal/native/runtime/json_stringify.c:155`, `:164`, `:167`, `:184`;
string escaping is at `:75`. Optional numbers must be unpacked first (`:108`);
optional booleans must use tagged boxes, not a two-word struct in one slot
(`internal/native/library_json_stringify.go:28`). **New:** an absent index needs
explicit presence metadata or an undefined tag; zero bits, false, NaN and null
references are not evidence of a hole. l2 handles direct literal holes separately
(`l2:internal/lower/library_json_stringify.go:227`).

Without a descriptor, refuse **before reading slots** with
`JSON.stringify runtime array without complete element descriptors`. Do not guess
from ownership or the first element, even for an empty array. Today's lowerer
already refuses an unproven element type (`internal/lower/library_json_stringify.go:184`).

## 2. Typed toJSON results

**New:** a runtime hook hands the encoder this logical tagged result:

```c
typedef struct adamic_json_result {
    adamic_value value;
    const adamic_json_schema *schema; /* non-NULL; schema->kind is the tag */
} adamic_json_result;
```

The schema describes actual storage and descendants and survives encoding/cleanup.
`null` and `undefined` have separate descriptor tags even when both payloads are NULL. A raw `adamic_value`
plus an ownership boolean is not an accepted result. Reference results transfer
one owned reference; release it once after use, including omission or failure.
Array element reads remain borrowed. Ownership is independent of the JSON tag.

Serialize number, string, boolean, null, and fully described objects/arrays as above.
Undefined and functions return undefined at the root, omit an object property,
and occupy `null` in an array (`internal/native/runtime/json_stringify.c:144`,
`:155`, `:178`, `:263`). Refuse untagged/unknown results or incomplete container
metadata; never substitute `{}` or `null` for unsupported data. Unsupported BigInt,
Symbol, exotics and cycles remain refused pending Node-correct behavior and throws.

l2 currently calls a proven zero-argument closure, reads its untagged return and
substitutes one static schema (`l2:internal/native/runtime/json_stringify.c:147`).
Its accepted return types are number/string/boolean/undefined or proven constant
arrays (`l2:internal/lower/library_json_stringify.go:206`, `:568`); null, object
and union hook returns are not already enabled. **New:** the returned pair replaces
that static-schema assumption and extends cleanup beyond string/array
(`l2:internal/native/runtime/json_stringify.c:163`). Call the hook once with its
actual receiver and key (root `""`, property name, decimal array index), before
replacer/encoding; do not call a returned object's own hook again at that same
position. Hooks observing receiver/key remain refused until that calling convention
exists. Propagate throws and stop encoding; l2's emitter already checks thrown state
(`l2:internal/native/library_json_stringify.go:58`).

## 3. Union-returning methods

For an existing boxed union, inspect `value.reference->kind`, never the payload's
truthiness or the first instantiated return type. Normalize to the tagged pair
above before serialization; preserve the original owned box until cleanup.
Today's dispatch is
`internal/native/runtime/json_stringify.c:112`:

| Runtime union tag | Decode and output |
|---|---|
| NULL | Undefined, unless the proven descriptor's `null_reference` says null (`:105`); new results must distinguish both explicitly |
| `adamic_kind_number` | Unbox `adamic_number_box.number`; Node number spelling, nonfinite `null` |
| `adamic_kind_boolean` | Unbox `adamic_boolean_box.boolean`; `true` / `false` |
| `adamic_kind_string` | Quoted/escaped string |
| `adamic_kind_map` | `{}` only for proven intrinsic Map/Set with no enumerable own additions |
| `adamic_kind_closure` | Root undefined / omitted property / array `null` |
| `adamic_kind_array` | **New:** fetch actual element metadata, then recursively print `[...]` |
| `adamic_kind_object` | **New:** fetch complete actual enumerable field/hook metadata, then recursively print `{...}` or apply a proven intrinsic descriptor |
| Any other tag | Refuse unsupported representation |

An object descriptor records actual enumerable own names, slot indexes and child
schemas in Node key order; a declared structural view cannot supply it
(`internal/native/runtime/json_stringify.h:11`,
`internal/native/runtime/json_stringify.c:184`). l2 added bounded literal-origin schemas
(`l2:internal/lower/library_json_stringify.go:510`), Date/RegExp and private
parsed carriers; these do not describe arbitrary object unions
(`l2:internal/native/runtime/json_stringify.c:194`, `:210`).

Missing array/object metadata refuses with `JSON union lacks proven container metadata`
(today `internal/native/runtime/json_stringify.c:123`). Container unions and shared
generic union bodies stay refused until this new metadata handoff exists
(`internal/lower/library_json_stringify.go:161`, `:199`). **New:** metadata follows
the actual method result across calls; it must never be cached from the first
return or reconstructed from an ownership bit.
