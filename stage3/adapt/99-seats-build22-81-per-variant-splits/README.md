# Per-variant array traversals for refusal family 81

Adamic's Nov 22, 2026 ruling at about 00:00Z on #bbssdcy approved the JavaScript-changing family 81 split after its audit. Adamic is the accepter after her review and Loom's unit census at main. Her Nov 23 answer at 07:21Z accepts the method split with its exposed depth-bound stop listed. Her Nov 24 answer at 00:08Z returned the split for a rewrite that avoids the four enum-tag sites. Her Nov 24 answer at 03:55Z asked for both readonly shapes to be measured before she picks. Ruling row 336 distinguishes a union of array types from an array of union elements. Rows 78 and 89 describe the fixed extra-field layout of `NodeArray` and the side record on an augmented array. Each traversal still uses the array stored on its original receiver.

## union-array-method

In `src/compiler/builder.ts`, `getBuildInfoFileVersionMap` reads `program.fileInfos` once into `versions: readonly (string | Pick<BuilderState.FileInfo, "version">)[]`. Both variants store strings or objects with a string version. The view keeps only that shared field and exposes a readonly array, so neither variant needs a copy. One `forEach` uses the original array as its receiver. The callback keeps the supplied index and computes the path before reading the version. The original bookkeeping still computes the file ID as `index + 1`. All reads of `program` use its unrefined type.

## union-array-iteration

In `src/compiler/transformers/utilities.ts`, `collectExportedVariableInfo` originally iterates `decl.name.elements` and recurses with each element. The two existing predicates test the same syntax kinds as `isBindingPattern`. The factory assigns `node.name = asName(name)` and stores each pattern's elements in `node.elements`. The two loops visit those original elements directly, in the original order. The array-pattern loop keeps `!isOmittedExpression(element)`. Object-pattern elements have only the `BindingElement` kind. There is no index callback at this site.

Recursion receives each element, with the original export collections. The identifier branch reads generated-name state, escaped text, emit flags, and original-node IDs, and updates the export collections. The body and its callees read no fields on the traversed `NodeArray` itself. Its `pos`, `end`, `hasTrailingComma` and `transformFlags` stay on the same object at every recursion level.

The adapter locates the traversals and predicate imports through stock TypeScript 6.0.3's AST. Each replacement requires its exact original or adapted snapshot. Both files are checked before writing.

Run `node stage3/adapt/99-seats-build22-81-per-variant-splits/test.cjs <applied tree>` on a complete Stage 3 apply. The test reads the adapter's two input files from its before snapshot, checks that repeat application preserves the output bytes, and runs the stock-transpiled methods and binding walks beside the originals. It checks sparse callback indices, version read order, roots, omissions, nested bindings and NodeArray identity and metadata. It checks that the original file-info array is `forEach`'s receiver and still `program.fileInfos` after the call. After the behavior checks, it compares the scratch output with the pipeline's `-after` snapshot. It is built to catch plants that pass `index + 1` to the root bookkeeping, recurse on omitted elements, replace the receiver before traversal, replace `program.fileInfos` after traversal, or change the output bytes with a comment. No gate at this pin runs the test yet.

## Exposed stops

Seats' full-mode census of `builder.ts` at `96dadc28` fills this section. Round 6's `readonly-view` first-error replay at `6b3b321f` recorded 118 builder sites against the base's 119, with zero findings in unit `builder.ts:2421:1`.

The targeted replay of `collectExportedVariableInfo`, unit `transformers/utilities.ts:357:1` on base and `transformers/utilities.ts:358:1` on head, cleared the iteration stop with 0 findings.
