# isPropertyName claims what its body accepts (#ad161ms)

Adamic ruled on task #ad161ms. This is type-only.

`isPropertyName` in `src/compiler/utilitiesPublic.ts` returns true for five
kinds: `Identifier`, `PrivateIdentifier`, `StringLiteral`, `NumericLiteral`
and `ComputedPropertyName`. Its predicate claimed `node is PropertyName`, and
`PropertyName` also holds `NoSubstitutionTemplateLiteral` and `BigIntLiteral`,
two kinds the body never accepts. A node the function rejected could still be
one of those two, so the false branch dropped kinds the value might hold, and
the true branch admitted kinds it never holds. The predicate now names the five:

```ts
export function isPropertyName(node: Node): node is Identifier | PrivateIdentifier | StringLiteral | NumericLiteral | ComputedPropertyName {
```

The file's namespace import gains `type ComputedPropertyName` and
`type NumericLiteral`, each in the list's own order, and drops `PropertyName`
when nothing else in the file names it, as nothing does at the pin.

The function is public, so the published `typescript.d.ts` line moves with it.
`api-sanctions.json` sanctions that one line, the predicate of the
`FunctionDeclaration` `isPropertyName`, and `stage3/lane/sanctioned-api.json`
is regenerated from it.

## The adapter

`adapt.cjs` finds `isPropertyName` and the namespace import through the stock
TypeScript 6.0.3 AST and pins the function's whole text, upstream or adapted.
It refuses any other text, a parse error, and a partly adapted file: the claim
adapted without the import's two names, or the import's names without the
claim. Before it writes, it compares the file's emitted JavaScript, with
comments kept and removed. A second run writes nothing.

The directory sorts after 99-correctness-binding-views, which pins the
import's whole text. 99-truthful-generated-name-emit-node sorts after this one
and inserts into the same import by name, pinning none of its text.

## What is proven

The proof job at this commit applies the adaptations twice into fresh trees
and compares them byte for byte, runs the pin's stock `tsc -b src --force`
over the adapted tree for zero diagnostics, builds the public `typescript.d.ts`
with and without this adaptation and prints the whole difference, which is the
one `isPropertyName` line, compares the emitted JavaScript with and without it,
and runs the lane's sanction, API and shard tests and the adapted tree product.
The mutant keeps the import edit and drops the claim's: the public line stays
upstream's and the lane's API check refuses the tree.
