# Return a boolean from the scanner's some callback

Correctness's Nov 26 ruling on #c5b5bxh covers `scanGroupName` inside
`createScanner` at upstream `scanner.ts:3007` and adapted
`scanner.ts:3017` in the before snapshot and final tree. The callback
uses `group !== undefined && group.has(tokenValue)` in place of
`group?.has(tokenValue)`. `some` tests truthiness, and `undefined` and
`false` are both falsy, so every input gives the same answer.

The adapter pins the whole `createScanner` declaration by kind, name,
before and after text, and SHA-256. The change is JavaScript-changing
and inside the function body. The snapshot test compares stock public
declaration emit byte for byte.

From the repository root, run:

    NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-some-boolean-callback/test.cjs TREE

The test reads both snapshots and runs their callbacks through Node's
`some`. The witness checks tokens `"a"` and `"b"` against
`[undefined, new Set(["a"]), undefined]`, then covers empty and
undefined-only stacks. Loom runs the sealed lane, the unit census and
the restoring patch.
