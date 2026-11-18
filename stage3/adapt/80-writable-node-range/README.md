# Writable node range, adaptation 80

Adamic, task #aryr7ea, ruled Nov 17 under rulings.md:391. This is type-only:
the emitted JavaScript doesn't change.

## The edit

`adapt.cjs` replaces the one `ReadonlyTextRange` in two heritage clauses of
`src/compiler/types.ts`:

```ts
export interface Node extends TextRange {                                   // was ReadonlyTextRange
export interface NodeArray<T extends Node> extends ReadonlyArray<T>, TextRange { // was ..., ReadonlyTextRange
```

It parses types.ts with stock TypeScript 6.0.3 and refuses before any write
unless it finds exactly one `Node` and one `NodeArray` interface, each with one
`extends` clause written exactly as upstream writes it. A clause already adapted
keeps its bytes, so a second run changes nothing. No source coordinate is a
rewrite address.

Nothing else changes. `TextRange`, `ReadonlyTextRange`, every member of `Node`
and `NodeArray`, and every rule stay as they are.

## The premise

tsc's own public `setTextRange` writes `pos` and `end` on a node after the
factory made it, through `setTextRangePos` and `setTextRangeEnd`
(`(range as TextRange).pos = pos`). So `readonly` on `Node`'s and `NodeArray`'s
`pos` and `end` was a promise to consumers, never a fact about the value.

The scout on #aryr7ea (J-k8h76p9) found every assignment to `.pos` or `.end` in
`src/compiler`. Those two are the only writes after creation. Every other write
is on an object its own function is still making: the Node, Token and
Identifier constructors' `this.pos`, `createNodeArray`'s fresh array,
`copyFileReferenceAsSynthetic`'s new reference, and a local range in
`emitDecoratorsAndModifiers`.

## Why this road

rulings.md:93 and :112 stand whole. A readonly field seen again as writable is
refused, because a readonly location is covariant and a subtype may narrow it.
Adaptation 72's certificate rests on that rule, and so does every `.a` file's
right to believe the word.

So the lie was in tsc. Once the declaration is true, a Node seen as a
TextRange is the same writable `number` on both sides. The 347 writers (setTextRange and
its worker, positioning a made node) and the 105 readers (setSourceMapRange,
setCommentRange, the generators' locations and the rest) are then ordinary
views. They need no `ReadonlyTextRange` slots and no change to setSourceMapRange,
setCommentRange or EmitNode.

This extends rulings.md:391, which lets an adaptation to tsc's public API only
add truth and sends anything else to a ruling. Removing a `readonly` that tsc's
own public function disproves is adding truth, for these two interfaces and
these two members, by name. Every other removal still goes to Adamic with its
count.

## Public consequence

`Node` and `NodeArray` are public, so the published `typescript.d.ts` changes
their two heritage clauses. Code that type-checked still does. The cost: a
consumer may now write `node.pos`, `node.end`, `array.pos` or `array.end` with no
cast, where upstream's types forbid it.

`api-sanctions.json` names exactly those two clauses (`"kind": "HeritageClause"`,
`"replace": "member"`), a kind `stage3/lane/generate-api.cjs` accepts for this
adaptation. The regenerated `stage3/lane/sanctioned-api.json` gains two
normalized changes, `InterfaceDeclaration:["ts","Node"]#0` and
`InterfaceDeclaration:["ts","NodeArray"]#0`, each its heritage with
`ReadonlyTextRange` replaced by `TextRange` and nothing else.
`shard-plan.json`'s runtime hash moves with the sanction, and
`stage3/lane/test_check.py` counts 278 sanctioned API changes, two more than 276.

## Measured

Measured on main 932a6344 with this adaptation as it is here, before its number
moved from 73 to 80. Each census is the latent census over the applied tree's
src/compiler in first-error mode, counting unique (kind, where, reason) sites as
the meter does.

- **The edit.** adapt.cjs makes exactly the two heritage edits, a second run
  changes nothing, and a moved heritage (`extends ReadonlyTextRange, Object`) is
  refused by name before any write (J-xandtwy).
- **Emitted JavaScript.** One program over all of upstream's src, emitted
  before and after: 604 JavaScript files each, 0 differing (J-kk2e2y0).
- **Narrower positions.** Of 20 `pos` or `end` members on types assignable to
  `TextRange`, 4 are on types assignable to `Node` or `NodeArray` (services'
  `NodeObject` and `TokenOrIdentifierObject`), all `number`. The two narrower ones are `SynthesizedComment`'s `pos: -1` and
  `end: -1`, which is neither, writable and narrower before this adaptation and
  after it (J-xandtwy, J-kk2e2y0).
- **Stock tsc.** 2,765 diagnostics over all of upstream's src before and after,
  0 new and 0 gone. The lane is green whole.
- **The family.** 550 sites before (at 547 places), 2 after. By the scout's
  classes (J-k8h76p9): the 347 writer places and 100 of the 102 reader places
  leave it. The 2 left are tsc's own generic setters, `setTextRangePos` and
  `setTextRangeEnd`, which cast `T extends ReadonlyTextRange` to `TextRange`:
  the shape of adaptation 71's `text-range.a`, still refused (J-gcd3cy5 and
  J-rbq1xks before, J-91r0nz0 after).
- **Where each class goes.** All 347 writer places now lower with no finding.
  Of the 102 reader places, 69 lower with no finding, 31 stop at optional
  `source` in `SourceMapRange` absent from the structural source, and 1 at an
  open numeric enum used as a literal tag (transformers/es2015.ts:2574). Each of
  the 98 `Mutable<X>` places stops at its next readonly field: `flags` at 76,
  `kind` at 22.
- **Reasons that grow.** 130 sites in all: those 98 next fields, and 32 of
  optional `source` in `SourceMapRange`. The `location` reasons stay at 34.
  Nothing is built for them here.
- **The mutant.** With only Node's clause put back (NodeArray's kept), 297 of
  the 347 writer places, 94 of the 102 reader places and all 98 `Mutable<X>`
  places come back (492 sites, J-4cwaw6n). The 50 writer places and 8 reader
  places that stay out have a `NodeArray` source (57 sites) or tsc's `Children`
  (1), which NodeArray's own clause covers. Both clauses put back is main's
  tree, with all 347.
- **SynthesizedComment.** One site names it before and after, the same:
  emitter.ts:3232, a `(node: CommentRange) => boolean` seen as one over
  `SynthesizedComment`, refused because a write of `number` would replace its
  `-1`. It stands, with its own reason, outside this ruling.

## Not built here

The 98 sites behind `Mutable<X>` views write other fields (flags, modifiers,
parent, type, the factory's update helpers). `pos` only rode along because
`Mutable` lifts readonly from every field, and with it gone they stop at `flags`
or `kind`, as measured above. rulings.md:253 keeps unproven writes through
`Mutable<T>` refused. Their table, by field and by whether each write is on an
object its maker still holds, is its own task.

`setTextRangePos` and `setTextRangeEnd` themselves are not changed: whether
their `T` is constrained to `TextRange` is a question for Adamic.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
