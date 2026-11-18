# JSDoc type-like tags take their own typeExpression

Adamic, task #38kdd4d (family 4's `typeExpression`, 10 sites), ruled Nov 18 as
variant B. It's type-only and internal: the emitted JavaScript doesn't change,
and no line of the published `typescript.d.ts` moves.

## The lie

nodeFactory.ts makes and updates five JSDoc tags (type, return, this, throws,
satisfies) through two memoized arrows and two workers, generic in
`T extends JSDocTag & { typeExpression?: JSDocTypeExpression; }`. The
constraint's field is writable, while every tag declares its `typeExpression`
readonly, and three of them (type, this and satisfies) declare it required.
So each `getJSDocTypeLikeTag{Create,Update}Function<Tag>`
in the factory literal was refused as a view that could write the tag's
narrower field (nodeFactory.ts:878 to :953, 10 sites). The worker's write,
`node.typeExpression = typeExpression` of a `JSDocTypeExpression | undefined`,
needed 61-temporary-factory-type-expression's cast to compile at all.

## The edits

The JSDoc unary type workers already take their `type` as `T["type"]` under a
readonly constraint. These take `typeExpression` the same way:

| Owner | Upstream | Adapted |
|---|---|---|
| the two arrows' and the two workers' `T` | `JSDocTag & { typeExpression?: JSDocTypeExpression; }` | `JSDocTag & { readonly typeExpression?: JSDocTypeExpression; }` |
| getJSDocTypeLikeTagCreateFunction's arrow | `typeExpression?: JSDocTypeExpression` | `typeExpression: T["typeExpression"]` |
| getJSDocTypeLikeTagUpdateFunction's arrow | `typeExpression?: JSDocTypeExpression` | `typeExpression: T["typeExpression"]` |
| both arrows' calls | `createJSDocTypeLikeTagWorker(...)`, `updateJSDocTypeLikeTagWorker(...)` | the same, with `<T>` |
| createJSDocTypeLikeTagWorker | `typeExpression?: JSDocTypeExpression` | `typeExpression: T["typeExpression"]` |
| updateJSDocTypeLikeTagWorker | `typeExpression: JSDocTypeExpression \| undefined` | `typeExpression: T["typeExpression"]` |

Each tag's write is now the plain `node.typeExpression = typeExpression` of its
own field type, which is what 61's cast stood in for. **61 is taken out whole
in the same change.** Its cast was all it did, and its reason is gone, so
nothing temporary is left standing. This adapter refuses a tree that still
holds 61's cast.

91-patcher-discard pins createNodeFactory's whole text by sha256, and the
worker sits inside it, so 61's removal re-pins 91's `proofs/targets.json` for
createNodeFactory in the same change. The new pins are 91's own texts with
61's one cast undone. They were read from apply's
`refs/stage3/91-patcher-discard-before` and `-after` on main's tree, where
each text as it stands hashed to 91's stored pin first (J-rb5cdby).
setObjectAllocator's pin doesn't move.

Each edit is found through the stock AST by its owner and its exact upstream
text, refused before any write if anything has moved, and left as it is on a
second run. The directory sorts after 96-emit-jsdoc-tag-comment-parameters,
whose pin reads the two arrows' parameter lists exactly as upstream wrote
them.

## What stays refused

NodeFactory's public `updateJSDocThisTag(node, tagName, typeExpression:
JSDocTypeExpression | undefined, comment)` (types.ts:9184) lets a caller store
undefined into a JSDocThisTag, whose `typeExpression` is declared present.
Upstream, that lie hid inside the generic's writable view. Now the refusal
sits where it lives: the `updateJSDocThisTag` getter in the factory literal
(nodeFactory.ts:892) hands an arrow that takes a required
`JSDocTypeExpression` to the wider public signature. That's "a function taking
JSDocTypeExpression seen as one taking JSDocTypeExpression | undefined", 96's
reason. No caller inside tsc passes undefined there (the parser only creates
the tag, parser.ts:9601). It's a candidate for the upstream note list.

## Measured

Measured at main 375c551c, with 61 out and 91 re-pinned. Each census is the
latent census over the applied tree's src/compiler in first-error mode.

- **The edit.** adapt.cjs makes its ten edits on upstream's nodeFactory.ts,
  and a second run writes nothing. A moved parameter (the update worker's
  `typeExpression` said to take null) is refused by name with nothing written.
  A tree still holding 61's cast is refused by name too (J-3y7hq6r).
- **Stock tsc.** 2,765 diagnostics over all 713 of upstream's src files,
  before (61 and 91's old pins restored) and after, with 0 new and 0 gone
  (J-3y7hq6r).
- **Emitted JavaScript.** 604 files before and after, 0 differing
  (J-3y7hq6r).
- **The census.** 2,919 sites before and 2,909 after (J-d3p0hjg at main,
  J-mafz4q3). The 11 that go are the family's 10 (nodeFactory.ts:878 to :953)
  and 61's cast (nodeFactory.ts:5496). The one that appears is the kept
  refusal at nodeFactory.ts:892:9, the `get updateJSDocThisTag()` line: "a
  function taking JSDocTypeExpression seen as one taking JSDocTypeExpression |
  undefined". Each census prints that pin.
- **The mutant.** The four constraints made writable again, with the typed
  parameters kept, bring back all 10 family sites. The total is 2,919, and the
  refusal at 892:9 stays (J-t2wh6x0).
- **The lane.** All 398 units green on the adapted tree, counts as the plan
  holds, byte comparisons passing (J-3y7hq6r).

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
