# JSX option map storage, family 101

Adamic ruled on family 101 on Nov 22, 2026 at about 20:50 UTC in #bbssdcy.
Her settled fallback takes Compiler's option (a) when the shared
enum-valued pair list stops at the union-valued Map constructor. Both
backends stopped there. Adamic confirmed on Nov 23 at 07:21 UTC that
Compiler is the accepter for this explicit literal-pair enumerated
table under the widened split. Review-i read rounds 5 and 6. Loom's
unit census at main 0c9b1201 (J-16fzg89) measured round 5's adapted
bytes, which later rounds keep. Compiler's acceptance is pending.

The site is the JSX option descriptor's `type: jsxOptionMap` in
`src/compiler/commandLineParser.ts`. The readonly option slot expects a
boxed number-or-string value. The mutable Map allocation fits the
readonly view under `docs/rulings.md:95`. The forward Map receives
literal pairs in the slot's value type.

The adapter derives the inverse pairs in forward order. For each forward
key, it resolves the named member in the pinned `JsxEmit` AST, reads its
numeric initializer, and stringifies that value. The emitted inverse
initializer contains the derived string key and forward key.
The ruling at `docs/rulings.md:335` governs number formatting and Map
iteration order. Both declarations and the option slot are checked
against their expected source shape before either initializer is edited.

The test checks the adapter's two edits and zero-edit replay, that its
output equals the pipeline's after snapshot byte for byte, both tables'
insertion order and typed reads, and CLI and JSON JSX handling including
the invalid-value diagnostic. It builds both sides with stock TypeScript.
A planted swap of the first two output pairs goes red in all three
behavior cases.

Run `test.cjs APPLIED_TREE` with Node and the pinned
`stage3/api/node_modules` on `NODE_PATH`, where `APPLIED_TREE` is a tree
`stage3/apply.sh` wrote. The test reads this adapter's before and after
snapshots from the tree's `refs/stage3/` refs, never its working files,
so a later adapter that edits `commandLineParser.ts` leaves it valid.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" \
  node stage3/adapt/99-seats-build22-101-jsx-option-map/test.cjs "$applied_tree"
```
