# Fresh emitNode writes, adaptation 87

Adamic, task #38kdd4d (family 4's `emitNode`), ruled Nov 18 as item 3, and
kept when she ruled road 3 for the family, which leaves `Node.emitNode`
writable. This adaptation removes the three `emitNode` writes tsc makes by
hand outside its emitNode.ts setters. Two of them go through
getOrCreateEmitNode, the setter tsc uses everywhere else. The third is on a
node nothing ever reads. Every edit is internal, so no line of the published
`typescript.d.ts` moves.

## The edits

| File | Owner | Upstream | Adapted |
|---|---|---|---|
| factory/emitHelpers.ts | createAsyncGeneratorHelper | `(generatorFunc.emitNode \|\| (generatorFunc.emitNode = {} as EmitNode)).flags \|= ...` | `getOrCreateEmitNode(generatorFunc).flags \|= ...` |
| factory/emitHelpers.ts | createAwaiterHelper | the same write | the same call |
| factory/emitHelpers.ts | the import list | `EmitNode` | `getOrCreateEmitNode`, in its place |
| expressionToTypeNode.ts | typeFromArrayLiteral | `const tupleType = factory.createTupleTypeNode(elementTypesInfo);` and `tupleType.emitNode = { ... };` | `factory.createTupleTypeNode(elementTypesInfo);` |

Each is found by its owner and its exact upstream text, refused before any
write if it has moved, and left as it is on a second run. Nothing is written
until every edit in both files has been found.

## The premises

- **generatorFunc is never a parse-tree node.** getOrCreateEmitNode differs
  from the inline `||` in two ways only. For a parse-tree node, it records the
  node in its source file's `annotatedNodes`. For a node whose emitNode is
  already Immutable, it asserts. createAwaiterHelper writes onto the
  createFunctionExpression it makes six lines above. createAsyncGeneratorHelper's
  one caller in src/compiler, es2018.ts:1178, hands it a createFunctionExpression
  it has just made. Neither is a parse-tree node, and neither has an emitNode
  yet, so the two writes are value-identical.
- **The tuple is dead.** typeFromArrayLiteral makes the tuple, writes its
  emitNode, restores the context and returns `notImplemented`. `tupleType` is
  named only by its declaration and the write. The create stays as a
  statement, so every call tsc makes still happens.
- **EmitNode leaves the import list.** The two writes were the only places
  emitHelpers.ts named it.

## Measured

MEASUREMENTS

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
