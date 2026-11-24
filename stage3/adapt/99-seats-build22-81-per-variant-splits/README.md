# Per-variant array traversals for refusal family 81

Adamic's Nov 22, 2026 ruling at about 00:00Z on #bbssdcy approved the JavaScript-changing family 81 split after its audit. Adamic is the accepter after her review and Loom's unit census at main. Her Nov 23 answer at 07:21Z accepts the method split with the exposed stop listed below. Her Nov 24 answer at 00:08Z returned the split for a rewrite that avoids the four enum-tag sites. Ruling row 336 distinguishes a union of array types from an array of union elements. Rows 78 and 89 describe the fixed extra-field layout of `NodeArray` and the side record on an augmented array. Each traversal still uses the array stored on its original receiver.

## union-array-method

In `src/compiler/builder.ts`, `getBuildInfoFileVersionMap` evaluates `isIncrementalBundleEmitBuildInfo(program)` as a boolean, then captures the original `program.fileInfos` array. The local predicate checks the argument against that captured array and uses the variant result to narrow the array alone. Its sole caller passes the captured array. The bundle and multifile callbacks each keep the `forEach` index, read `program.fileNames[index]!` through the unrefined program, and compute the path before reading the version. They pass the path, version and index to the unchanged map and root bookkeeping, which still computes the file ID as `index + 1`. Neither traversal replaces the array.

## union-array-iteration

In `src/compiler/transformers/utilities.ts`, `collectExportedVariableInfo` originally iterates `decl.name.elements` and recurses with each element. The two existing predicates test the same syntax kinds as `isBindingPattern`. The factory assigns `node.name = asName(name)` and stores each pattern's elements in `node.elements`. The two loops visit those original elements directly, in the original order. The array-pattern loop keeps `!isOmittedExpression(element)`. Object-pattern elements have only the `BindingElement` kind. There is no index callback at this site.

Recursion receives each element, with the original export collections. The identifier branch reads generated-name state, escaped text, emit flags, and original-node IDs, and updates the export collections. The body and its callees read no fields on the traversed `NodeArray` itself. Its `pos`, `end`, `hasTrailingComma` and `transformFlags` stay on the same object at every recursion level.

The adapter locates the traversals and predicate imports through stock TypeScript 6.0.3's AST. Each replacement requires its exact original or adapted snapshot. Both files are checked before writing.

Run `node stage3/adapt/99-seats-build22-81-per-variant-splits/test.cjs <applied tree>` on a complete Stage 3 apply. The test reads the adapter's two input files from its before snapshot, checks that repeat application preserves the output bytes, and runs the stock-transpiled methods and binding walks beside the originals. It checks sparse callback indices, version read order, roots, omissions, nested bindings and NodeArray identity and metadata. It checks that the original file-info array is `forEach`'s receiver and still `program.fileInfos` after the call. After the behavior checks, it compares the scratch output with the pipeline's `-after` snapshot. It is built to catch plants that pass `index + 1` to the root bookkeeping, recurse on omitted elements, replace the receiver before traversal, replace `program.fileInfos` after traversal, or change the output bytes with a comment. No gate at this pin runs the test yet.

## Exposed stops

The compiler-directory targeted replay is the source of this list. The replay of the rewritten method at 6b3b321f found none of the four open numeric enum-tag sites introduced by round 5. Loom's unit census at main checks the result in both scopes.

- `getBuildInfoFileVersionMap`, unit `builder.ts:2421:1`. The array stop at `builder.ts:2433:5` cleared. The exposed NotYet at `builder.ts:1199:1` reads “a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo)”. Its family is the checked-predicate target depth bound. The Nov 22 meter's exact `per_reason` key is `NotYet: a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo)`, with count 1 in each scope. The family's number was not found. The unit had already stopped during lowering, and the meter lists the exposed reason, so the stop is not new under Adamic's rule.

The method unit won't lower until the depth-bound family lands. Adamic's Nov 23 answer explains why the split goes in:

> **5. Accept the method split** after its read and census: the depth-bound stop it exposes is listed, so it's not new by the rule. That's the line's rule, and loading the line is what we want.

The targeted replay of `collectExportedVariableInfo`, unit `transformers/utilities.ts:357:1` on base and `transformers/utilities.ts:358:1` on head, cleared the iteration stop with 0 findings.
