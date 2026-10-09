# optional-boolean-field

Source: `src/compiler/factory/emitNode.ts:166` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

getStartsOnNewLine keeps its complete body. Node retains emitNode and startsOnNewLine. Exercise a true field and an absent emitNode.

Historical main census: **25** `a field of type boolean | undefined` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a field of type boolean | undefined`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
