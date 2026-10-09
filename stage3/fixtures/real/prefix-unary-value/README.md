# prefix-unary-value

Source: `src/compiler/utilities.ts:997` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

isPlainJsFile keeps its body. SourceFile retains only scriptKind and checkJsDirective; ScriptKind supplies the two used constants. Exercise both a present JS file and an absent file.

Historical main census: **85** `a PrefixUnaryExpression on a value` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a PrefixUnaryExpression on a value`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
