# Readonly mapped positions, adaptation 82

Adamic, task #w743639 (its C). Type-only: the emitted JavaScript does not change.

## The lie

`sourcemap.ts`'s file-local `MappedPosition` declares its five fields writable.
`SourceMappedPosition` extends it and narrows `source`, `sourceIndex` and
`sourcePosition` to defined. So `sameMappedPosition(left: MappedPosition, right: MappedPosition)`,
handed to `sortAndDeduplicate<SourceMappedPosition>` as its equality comparer in
`getSourceMappings`, is a function that could write `undefined` into a slot its
caller reads as defined. Adamic refuses that view ("can write string | undefined
where string is read", sourcemap.ts:766). Nothing writes a MappedPosition's
fields: the decoder builds each one whole, and the comparers only read.

## The edit

`adapt.cjs` inserts `readonly ` before each of `MappedPosition`'s five fields,
found through the stock AST, each with upstream's type. It refuses on any other
shape, and a field already readonly keeps its bytes. `SourceMappedPosition`
redeclares its three narrower fields and stays writable, as adaptation 72 leaves
`DiagnosticWithLocation`.

## Certificate

The compiled program's gate, as for adaptation 72: a write through a wider view
inside `src/compiler` is refused at its relation site. The interface is
file-local, so nothing outside `sourcemap.ts` can see it.

## Measured

From J-4t94cm2 (sealed, scratch at main 0578cbd7 holding this and four sibling heads of #w743639), the tree whole against the same tree with this directory dropped:

- The meter's latent total: 3,690 without it, 3,689 with it.
- Gone: sourcemap.ts:766, the one site. Nothing new.
- Sources: sourcemap.ts alone, five lines, each by a `readonly ` insertion.
- Stock tsc over `src/compiler`, `src/services` and `src/server`: 0 new errors.
