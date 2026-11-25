# Make the require or import argument tuple readonly

Compiler accepts this internal, type-only rider to Adamic's Nov 24 2026
ruling 3 on #bbssdcy. The change marks the single StringLiteralLike
position in RequireOrImportCall.arguments as readonly. It preserves
the call expression, argument order, and inferred literal type.

The structural pin is the RequireOrImportCall type alias in
src/compiler/types.ts. The family 83 require-literal collector keeps
its own adaptation and snapshot.

From the repository root, after a full apply with the pinned API packages,
run the snapshot test.

```sh
adapter=stage3/adapt/99-seats-build22-104-require-readonly-tuple
NODE_PATH="$PWD/stage3/api/node_modules" node "$adapter/test.cjs" <applied-tree>
```

The test reads this adapter's before and after refs, checks the readonly
tuple independently of the replay, and verifies the single-edit replay.
