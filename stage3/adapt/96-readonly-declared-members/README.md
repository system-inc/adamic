# Readonly declared members, adaptation 96-readonly-declared-members

Adamic, task #r11wppy item 1, the public change. This is type-only: the
emitted JavaScript does not change. Its internal sibling is
95-readonly-empty-array-slots.

## The edit

`adapt.cjs` inserts `readonly ` before the array type of three members of
`InterfaceTypeWithDeclaredMembers` in `src/compiler/types.ts`:

```ts
export interface InterfaceTypeWithDeclaredMembers extends InterfaceType {
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

The directory sorts after 95-readonly-empty-array-slots, its internal sibling:
the two were type-checked together over tsc's whole program, and this one lands
after it. No adaptation reads or pins these members or the text of
`InterfaceTypeWithDeclaredMembers`.

## Why

Each member holds tsc's shared `emptyArray`, typed `never[]`, for a type with
no members of that kind. Seen as a writable `Signature[]` or `IndexInfo[]`,
that's a view that could write a value where never is read, a lie. A readonly
member is a read-only view of it, and Adamic's never[]-seen-as-T[] refusal
clears at the six sites the members hold: checker.ts:13762:81, 13763:86,
13764:77, 17922:39, 17923:44 and 17924:35.

`declaredProperties` stays as tsc writes it. The readonly fixpoint's third pass
put three sites under it, but measured with `emptyArray` left `never[]` it
clears none of them (J-gxkn10c): their values flow on into getNamedMembers'
mutable return, and that's where they stop. By Adamic's rule a public slot goes
readonly only when a listed site needs it, so it doesn't.

`emptyArray` stays `never[]` and no other slot moves: with the sentinel
mutable, each member clears its sites alone. The four-member set this came from
type-checks clean over tsc's whole program with stock tsc 6.0.3, together with
adaptation 95-readonly-empty-array-slots' internal change (J-n5a2k81).

## Public consequence

`InterfaceTypeWithDeclaredMembers` is public, so the published
`typescript.d.ts` gains `readonly` on the three members' array types. Code that
pushes onto one, or assigns it to a mutable array, stops type-checking; code
that reads them is unaffected.

`api-sanctions.json` names exactly those three members, each replacing only the
type, for `stage3/lane/generate-api.cjs`. The regenerated
`stage3/lane/sanctioned-api.json` gains three normalized changes,
`member:["ts","InterfaceTypeWithDeclaredMembers",<member>]#0`, one per member.
`shard-plan.json`'s runtime hash moves with the sanction, and
`stage3/lane/test_check.py` counts 279 sanctioned API changes, three more than
276.

## Proof

Sealed, by scan-all's instrumented probe over sys.ts's closure in apply's tree,
with stock TypeScript 6.0.3: TBD after this change's own run. With the fourth
member still in (J-gxkn10c): six never[] stops go, none is new, the mutant that
takes `declaredIndexInfos`' one `readonly ` back out brings back exactly its two
(13764:77 and 17924:35), stock tsc reads the same one error before and after,
and the adapted tree's product test and stage3/lane pass.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
