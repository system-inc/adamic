# type-predicate

Source: `src/compiler/factory/nodeTests.ts:318` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

isIdentifier keeps its body. Node and Identifier are reduced to kind and escapedText; the open input still does not prove escapedText. The blanket census reason now has an explicit failed-proof suffix.

Historical main census: **588** `a type predicate` sites (checker-rejected measurement).
Current fixture: **Refused**, `a type predicate whose return is not proven (return expression is not a trusted check on node)`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
