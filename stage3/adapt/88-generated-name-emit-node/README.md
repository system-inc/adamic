# A generated name's emitNode, adaptation 88

Adamic, task #38kdd4d (family 4's `emitNode`, 58 sites), ruled Nov 18 as
road 3. It withdrew road (a), which would have made `Node.emitNode` readonly
and certified its writers by name. Nothing in lower can carry a certificate
by name, so the declaration that lies is fixed instead. `Node.emitNode` stays
writable.

## The lie

GeneratedIdentifier and GeneratedPrivateIdentifier (types.ts, both
`@internal`) redeclared their slot narrower than Node's:

```ts
readonly emitNode: EmitNode & { autoGenerate: AutoGenerateInfo; };
```

tsc's own constructors break that claim for a statement or two.
createBaseGeneratedIdentifier and createBaseGeneratedPrivateIdentifier cast a
fresh node to the generated type, then call setIdentifierAutoGenerate, whose
getOrCreateEmitNode writes `{}` before `autoGenerate` is set. The two clones,
cloneGeneratedIdentifier and cloneGeneratedPrivateIdentifier, call
`setOriginal(clone, node)` first. That writes a mergeEmitNode result, and
mergeEmitNode never copies `autoGenerate`. Every value of the two types seen
as a plain Identifier, PrivateIdentifier or Node was refused as a view that
could write the narrower slot.

## The edits

- **types.ts:** both interfaces drop the redeclaration and keep Node's
  `emitNode`. They're internal, so no line of the published `typescript.d.ts`
  moves and nothing is sanctioned.
- **The nine reads** of `x.emitNode.autoGenerate` as present now read
  `Debug.checkDefined(getIdentifierAutoGenerate(x))`. That's tsc's own getter
  beside setIdentifierAutoGenerate, with a stop that fails loudly.

| File | Reads |
|---|---|
| emitter.ts | generateName, makeName |
| utilitiesPublic.ts | 1 |
| transformers/classFields.ts | 1, which writes `autoGenerate.flags` through the checked read |
| transformers/utilities.ts | 1 |
| transformers/module/module.ts | 1 |
| factory/utilities.ts | getNodeForGeneratedName |
| factory/nodeFactory.ts | cloneGeneratedIdentifier, cloneGeneratedPrivateIdentifier |

Each file's import from `_namespaces/ts.js` gains `Debug` and
`getIdentifierAutoGenerate` where it lacks them.

Each edit is found through the stock AST by its owner and its exact upstream
text, refused before any write if anything has moved, and left as it is on a
second run. A read off anything but a plain name is refused, so `x` is
evaluated once either way.

## The premise

Every generated name has `emitNode.autoGenerate` set before its constructor
returns, and nothing clears it afterward. setIdentifierAutoGenerate is its
only setter, and its four callers are the four generated-name constructors.
So each checked read finds it present on tsc's own paths, and the check never
fires there. If a caller of the public API ever builds a generated name some
other way, the read stops loudly where it would have read a field of
undefined.

## Measured

MEASUREMENTS

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
