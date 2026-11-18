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

Measured at main 150b7b5e.

- **The edit.** adapt.cjs makes its four edits on upstream's files, and a
  second run writes nothing. A moved write (the awaiter's with one flag
  fewer) is refused by name, and neither file is written (J-709jyaj).
- **Stock tsc.** 2,765 diagnostics over all 713 of upstream's src files,
  before and after, with 0 new and 0 gone (J-709jyaj).
- **Emitted JavaScript.** It differs in factory/emitHelpers.js and
  expressionToTypeNode.js alone, the two edited files (J-709jyaj).
- **The witness.** The whole lane on 87's tree, with
  `Debug.assert(!isParseTreeNode(generatorFunc))` before both
  getOrCreateEmitNode calls: 398 units, 0 failing, counts {pass 106,366,
  fail 1} where the plan holds the same, and the byte comparisons pass
  (J-709jyaj). On main's tree, the same assert before both upstream writes
  is green too (J-g98psn6). Its reach control, the opposite assert, turns
  165 units red, so both asserts run across the suite (J-yehfz5s).
- **The census.** 87 moves no census site by itself, and it unblocks none:
  road 3 for family 4's `emitNode` (99-cli-generated-name-emit-node) leaves
  `Node.emitNode` writable and doesn't rest on it. Adamic kept it (item 3)
  so that tsc writes `emitNode` in only two kinds of place:
  - emitNode.ts's own setters and setOriginalNode;
  - the code making a node: the Node, Token and Identifier constructors and
    the two source-file clones.

  The emitHelpers writes and the dead tuple's were the only hand writes
  outside both. Any later ruling about who writes `emitNode` reads that set
  whole.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
