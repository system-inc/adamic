# generic-optional-return

Source: `src/compiler/core.ts:1083` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

firstOrUndefined keeps its body and generic signature. Exercise a nonempty numeric array and an empty array. Concrete generic optional returns still hit the exact census reason on this main.

Historical main census: **66** `a function returning T | undefined` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a function returning T | undefined`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
