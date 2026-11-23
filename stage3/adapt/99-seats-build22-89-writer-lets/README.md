# Writer locals use let

Adamic's Nov 22, 2026, 21:49Z ruling on #bbssdcy admits this family 89
adaptation under row 400. Compiler is the accepter after review and census,
applying that family ruling as ruled.
The two sites are the declaration at `src/compiler/utilities.ts:676:5`
in `createSingleLineStringWriter` and the declaration block beginning at
`src/compiler/utilities.ts:6301:5` in `createTextWriter`.

Row 202 refuses these function-local `var` declarations. They use neither
an asserted initializer (rows 60 and 202) nor direct private namespace storage
(rows 202 and 275). Row 275's readiness rule matters for locals whose types
exclude undefined. The single-line writer initializes `str` before
constructing its callbacks. In the multiline writer, only function
declarations appear between the locals and the first `reset()`. Before
the return exposes any callback, that call assigns all six locals. Their
first reads therefore follow initialization under every call order. The
replacement initializers use the exact values set by the existing reset.

```ts
function reset(): void {
    output = "";
    indent = 0;
    lineStart = true;
    lineCount = 0;
    linePos = 0;
    hasTrailingComment = false;
}
```

The initial call and the returned `clear: reset` stay in place. Each factory
invocation creates fresh cells for the callbacks it returns. The TypeScript
AST locates the two functions and exact declaration spans. The adapter checks
the original reset text and its execution order and preserves source trivia.
The adjacent upstream comments record the original reason for choosing
`var`: avoiding TDZ-check overhead.

The scope check checks parameters, hoisted `var` declarations, and direct
body declarations for collisions with the seven writer locals. Bindings
inside nested functions may shadow these names in their own scopes.

Run the adapter's guard and interleaved writer checks with the applied tree,
whose before snapshot supplies both input files and whose after snapshot
supplies this adapter's recorded output. Later edits to utilities.ts leave
these checks valid:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" \
  node stage3/adapt/99-seats-build22-89-writer-lets/test.cjs /tmp/adapted
```

The test compares both writers' text and line state against the original
writers in Node, including two clears with intervening writes.

Temporary under row 400. Remove when a future Adamic ruling admits these
function-local `var` declarations with per-invocation captured cells and
the same initialization and reset behavior, and that support lands.

Incremental scoreboard: **1 file, 7 lines added, 7 removed**.
