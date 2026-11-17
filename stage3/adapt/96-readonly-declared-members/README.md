# Readonly declared members, adaptation 96-readonly-declared-members

Adamic, task #r11wppy item 1, the public change. This is type-only: the
emitted JavaScript does not change. Its internal sibling is
95-readonly-empty-array-slots.

## The edit

`adapt.cjs` inserts `readonly ` before the array type of four members of
`InterfaceTypeWithDeclaredMembers` in `src/compiler/types.ts`:

```ts
export interface InterfaceTypeWithDeclaredMembers extends InterfaceType {
    declaredProperties: readonly Symbol[];                   // Declared members
    declaredCallSignatures: readonly Signature[];            // Declared call signatures
    declaredConstructSignatures: readonly Signature[];       // Declared construct signatures
    declaredIndexInfos: readonly IndexInfo[];                // Declared index signatures
}
```

It parses types.ts with stock TypeScript 6.0.3 and refuses before any write
unless it finds exactly one `InterfaceTypeWithDeclaredMembers` with exactly one
property signature of each name, each with upstream's type and no question
mark or modifier. A member already readonly keeps its bytes, so a second run
changes nothing. No source coordinate is a rewrite address.

The directory sorts after 95-readonly-empty-array-slots. resolveObjectTypeMembers
passes `declaredProperties` to createInstantiatedSymbolTable, whose `symbols`
that adaptation makes readonly; before it, stock tsc refuses a readonly array
seen as a mutable one (TS4104). No adaptation reads or pins the four members or
the text of `InterfaceTypeWithDeclaredMembers`.

## Why

Each member holds tsc's shared `emptyArray`, typed `never[]`, for a type with
no members of that kind. Seen as a writable `Symbol[]`, `Signature[]` or
`IndexInfo[]`, that's a view that could write a value where never is read, a
lie. A readonly member is a read-only view of it, and Adamic's
never[]-seen-as-T[] refusal clears at the 11 sites the members hold.

`emptyArray` stays `never[]` and no other slot moves: with the sentinel
mutable, each member clears its sites alone. This set type-checks clean over
tsc's whole program with stock tsc 6.0.3 (J-n5a2k81, together with adaptation
95-readonly-empty-array-slots' internal change).

## Public consequence

`InterfaceTypeWithDeclaredMembers` is public, so the published
`typescript.d.ts` gains `readonly` on the four members' array types. Code that
pushes onto one, or assigns it to a mutable array, stops type-checking; code
that reads them is unaffected.

`api-sanctions.json` names exactly those four members, each replacing only the
type, for `stage3/lane/generate-api.cjs`. The regenerated
`stage3/lane/sanctioned-api.json` gains four normalized changes,
`member:["ts","InterfaceTypeWithDeclaredMembers",<member>]#0`, one per member.
`shard-plan.json`'s runtime hash moves with the sanction, and
`stage3/lane/test_check.py` counts 280 sanctioned API changes, four more than
276.

## Proof

| Measurement | Before | After | Mutant |
|---|---:|---:|---:|
| never[] stops at the 11 sites | TBD | TBD | TBD |
| Stops new against before | | TBD | |
| Stock tsc errors over src/compiler | TBD | TBD | TBD |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
