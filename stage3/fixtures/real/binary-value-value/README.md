# binary-value-value

Source: `src/compiler/commandLineParser.ts:2317` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

getWatchOptionsNameMap keeps its complete cache expression. Reduce OptionsNameMap to size and supply a counted map factory. Two calls exercise cache creation and cache reuse; the factory runs once.

Historical main census: **32** `a BinaryExpression with a value and a value` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a BinaryExpression with a value and a value`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
