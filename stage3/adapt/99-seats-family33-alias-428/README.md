# Family 33 conditional-key measurement

Adamic's Nov 22 ruling on #bbssdcy authorizes measuring the checked-predicate site at `src/compiler/transformers/utilities.ts:428:34` on `0679623c2aa6`.

In `IdentifierNameMap.toKey`, initialize `nodeForId` from the existing `node` binding immediately before the ternary. Use it once in the false arm's `getNodeId` call. The existing `getNodeForGeneratedName(name)` call still runs once. The positive test, identity comparison, and recursive call retain their existing arguments. The false arm needs the `Node` accepted by `getNodeId`, which the return type of `getNodeForGeneratedName` supplies without narrowing by `isMemberName`.

The adapter locates the conditional by its AST owners and checks its exact before or after text. This is one measured site. The ledger entry records its scope. Its name guard walks identifiers, so a direct `eval` string naming `nodeForId` would pass it; `docs/rulings.md:290` refuses `eval` in Adamic's input and tsc 6.0.3's `toKey` holds none, so no such drift can compile through Adamic (Adamic, Nov 25, #g84cc6q).
`test.cjs` compares the adapter output with `-after` and runs `toKey` from both snapshots side by side on recording stubs to check that `getNodeId` receives the same node, verifies on both snapshots that the adapter refuses a `toKey` already holding `nodeForId`, and Adamic accepts the head, on her Nov 24 ruling on #bbssdcy, after its review and census.
